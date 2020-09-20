package framework

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"strings"

	_ "github.com/lib/pq"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/command"
	"github.com/eudore/eudore/component/notify"
	"github.com/eudore/eudore/component/pprof"
	"github.com/eudore/eudore/middleware"
	"github.com/eudore/website/config"
)

var (
	BuildTime string
	CommitID  string
)

type App struct {
	*config.Config `alias:"config"`
	*eudore.App    `alias:"app"`
	*sql.DB        `alias:"db"`
	*RAM           `alias:"ram"`
}

func New() *App {
	conf := config.New()
	app := &App{
		Config: conf,
		App: eudore.NewApp(
			eudore.NewLoggerInit(),
			eudore.NewConfigEudore(conf),
		),
	}
	app.Options(context.WithValue(context.Background(), eudore.AppContextKey, app))
	return app
}

func (app *App) Run(inits ...func(*App) error) error {
	// 使用了LoggerInit在启动失败时将临时日志吐出来。
	defer func() {
		_, ok := app.Logger.(*eudore.LoggerInit)
		if ok {
			app.Options(eudore.NewLoggerStd(nil))
			app.Sync()
		}
	}()

	inits = append([]func(*App) error{
		InitConfig,
		InitCommand,
		InitLogger,
		InitNotify,
		InitData,
		InitStatic,
		InitMiddleware,
	}, inits...)
	inits = append(inits, InitServer)

	go func() {
		defer func() {
			r := recover()
			if r != nil {
				err := fmt.Errorf("website app recover error: %v", r)
				app.WithField("depth", "enable").WithField("depth", 1).WithField("stack", eudore.GetPanicStack(5)).Error(err)
				app.Options(err)
			}
		}()
		for i := range inits {
			err := inits[i](app)
			if err != nil {
				app.Options(err)
				break
			}
		}
	}()
	return app.App.Run()
}

func InitConfig(app *App) error {
	return app.Parse()
}

func InitCommand(app *App) error {
	return command.Init(app.App)
}

func InitLogger(app *App) error {
	_, ok := app.Logger.(interface {
		NextHandler(eudore.Logger)
	})
	if ok {
		app.Options(eudore.NewLoggerStd(app.Get("component.logger")))
	}
	return nil
}

func InitNotify(app *App) error {
	return notify.NewNotify(app.App).Run()
}

// InitServer 函数启动配置并启动服务。
func InitServer(app *App) error {
	// 设置server处理者。
	if h, ok := app.Get("keys.handler").(http.Handler); ok {
		app.Server.SetHandler(h)
	}
	// 设置go 1.13 net/htpp.Server生命周期上下文。
	eudore.Set(app.Server, "BaseContext", func(net.Listener) context.Context {
		return app.Context
	})

	// 监听全部配置
	// var lns []eudore.ServerListenConfig
	// ConvertTo(app.Config.Get("listeners"), &lns)
	for _, i := range app.Config.Listeners {
		ln, err := i.Listen()
		if err != nil {
			app.Error(err)
			continue
		}
		if i.HTTPS {
			app.Logger.Infof("listen https in %s %s,host name: %v", ln.Addr().Network(), ln.Addr().String(), i.Certificate.DNSNames)
		} else {
			app.Logger.Infof("listen https in %s %s", ln.Addr().Network(), ln.Addr().String())
		}
		app.Serve(ln)
	}
	return nil
}

// InitData 函数初始化数据库。
func InitData(app *App) error {
	db, err := sql.Open(app.Config.Component.DB.Driver, app.Config.Component.DB.Config)
	if err != nil {
		return err
	}

	var version string
	err = db.QueryRow("SELECT version()").Scan(&version)
	if err != nil {
		return err
	}
	db.SetMaxIdleConns(0)
	app.Infof("database version: %s", version)

	params := strings.Split(app.Config.Component.DB.Config, " ")
	for i := range params {
		if strings.HasPrefix(params[i], "password=") {
			params[i] = "password=********"
			break
		}
	}
	app.Config.Component.DB.Config = strings.Join(params, " ")
	app.DB = db
	app.RAM = NewRAM(app)
	app.GetFunc("/ram/match/:action", func(ctx eudore.Context) interface{} {
		name, result := app.RAM.MatchAction(ctx, ctx.GetParam("action"))
		return map[string]interface{}{
			"userid": ctx.GetParam("UID"),
			"action": ctx.GetParam("action"),
			"ram":    name,
			"result": result,
		}
	})
	return nil
}

// InitStatic 函数初始化静态文件
func InitStatic(app *App) error {
	api := app.Group("")
	api.AddMiddleware(
		NewAddHeaderFunc(),
		middleware.NewGzipFunc(5),
	)

	// static
	api.GetFunc("/static/*", NewStaticHandlerFunc(""))
	api.GetFunc("/js/:path", NewMergeFileHandlerFunc("static/js"))
	api.GetFunc("/js/lib/:path", NewMergeFileHandlerFunc("static/js/lib"))
	api.GetFunc("/css/:path", NewMergeFileHandlerFunc("static/css"))
	api.GetFunc("/css/lib/:path", NewMergeFileHandlerFunc("static/css/lib"))
	api.GetFunc("/favicon.ico", eudore.NewStaticHandler("static"))
	api.GetFunc("/version", func(eudore.Context) interface{} {
		return map[string]string{
			"BuildTime": BuildTime,
			"CommitID":  CommitID,
		}
	})
	return AutoInjectHTML(api, "static/html")
}

// InitMiddleware 函数初始化中间件和debug。
func InitMiddleware(app *App) error {
	app.AddHandlerExtend(NewContextExtend(app.DB)...)
	// app.AddHandlerExtend(NewExtendFuncMapStringError)

	// admin
	admin := app.Group("/eudore/debug godoc=" + app.Config.Component.Pprof.Godoc)
	admin.AddMiddleware(middleware.NewBasicAuthFunc(app.Config.Component.Pprof.BasicAuth))
	pprof.Init(admin)
	admin.AnyFunc("/pprof/look/*", pprof.NewLook(app))
	admin.AnyFunc("/admin/ui", middleware.HandlerAdmin)

	// 增加全局中间件
	app.AddMiddleware(
		// NewTracerFunc(),
		middleware.NewLoggerFunc(app.App, "route", "action", "ram", "policy", "basicauth", "controllername", "controllermethod", "browser", "sql"),
		middleware.NewDumpFunc(admin),
		middleware.NewBlackFunc(app.Config.Component.Black, admin),
		middleware.NewRateFunc(10, 100, app),
		// middleware.NewBreaker().InjectRoutes(admin).NewBreakFunc(),
		NewAddHeaderFunc(),
		// cors
		middleware.NewCorsFunc(nil, map[string]string{
			"Access-Control-Allow-Credentials": "true",
			"Access-Control-Allow-Headers":     "Authorization,DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,X-Parent-Id",
			"Access-Control-Expose-Headers":    "X-Request-Id",
			"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, HEAD",
			"Access-Control-Max-Age":           "1000",
		}),
		middleware.NewGzipFunc(5),
		middleware.NewRecoverFunc(),
	)
	// /api/v1/
	app.AnyFunc("/api/v1/*", eudore.HandlerRouter404)
	app.AddMiddleware(
		"/api/v1/",
		NewUserInfoFunc(app),
		app.RAM.NewRAMFunc(),
	)
	// 404 405
	app.AddHandler("404", "", eudore.HandlerRouter404)
	app.AddHandler("405", "", eudore.HandlerRouter405)
	return nil
}
