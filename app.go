package main

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"

	"github.com/eudore/eudore"
	// _ "github.com/eudore/eudore/component/expvar"
	eserver "github.com/eudore/eudore/component/server/eudore"
	//	_ "github.com/eudore/eudore/component/server/fasthttp"
	"github.com/eudore/eudore/component/command"
	"github.com/eudore/eudore/component/expvar"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/component/notify"
	"github.com/eudore/eudore/component/pprof"
	"github.com/eudore/eudore/component/router/debug"
	"github.com/eudore/eudore/component/show"
	"github.com/eudore/eudore/middleware"

	"github.com/eudore/website/config"
	"github.com/eudore/website/handlers/auth"
	"github.com/eudore/website/handlers/status"
	// appcontext "github.com/eudore/website/internal/context"
	appcontroller "github.com/eudore/website/internal/controller"
	appmiddleware "github.com/eudore/website/internal/middleware"
)

func main() {
	app := eudore.NewEudore(
		eudore.NewConfigEudore(config.GetConfig()),
		debug.NewRouterDebug(),
		eserver.NewServer(nil),
	)
	// app.SetLevel(eudore.LogError)

	app.RegisterInit("eudore-command", 0x007, command.InitCommand)
	app.RegisterInit("eudore-notify", 0x00e, notify.Init)

	app.RegisterInit("init-data", 0x101, InitData)
	app.RegisterInit("init-static", 0x103, InitStatic)
	app.RegisterInit("init-middleware", 0x105, InitMidd)

	app.RegisterInit("auth", 0x310, auth.Reload)
	app.RegisterInit("status", 0x360, status.Init)
	app.RegisterInit("http-test", 0x400, InitTest)
	app.Run()
}

func InitData(app *eudore.Eudore) error {
	dbtype := app.Config.Get("keys.dbtype").(string)
	db, err := sql.Open(dbtype, app.Config.Get("keys.dbconfig").(string))
	if err != nil {
		return err
	}

	app.Config.Set("keys.dbconfig", "")
	app.Config.Set("keys.db", db)
	return nil
}

func InitStatic(app *eudore.Eudore) error {
	// static
	mergeHandler := appcontroller.NewControllerStatic().NewMergeFileHandlerFunc("static")
	staticHandler := appcontroller.NewControllerStatic().NewStaticHandlerFunc("")
	middlewares := eudore.HandlerFuncs{appmiddleware.NewAddHeaderFunc(),
		middleware.NewGzipFunc(5),
		middleware.NewTimeoutFunc(5 * time.Second),
		// 捕捉panic
		middleware.NewRecoverFunc(),
	}
	app.GetFunc("/js/:path dir=static/js/", middlewares, mergeHandler)
	app.GetFunc("/css/:path dir=static/css/", middlewares, mergeHandler)
	app.AddStatic("/css/lib/* action=GetStaticCss", "static")
	app.AddStatic("/js/lib/* action=GetStaticJs", "static")
	app.AddStatic("/favicon.ico dir=/data/web/static", "static")
	app.GetFunc("/static/* action=GetStatic", middlewares, staticHandler)
	return nil
}

func InitMidd(app *eudore.Eudore) error {
	// 检查数据库连接池
	_, ok := app.Config.Get("keys.db").(*sql.DB)
	if !ok {
		panic("init middleware check config 'keys.db' not find database/sql.DB.")
	}

	// 增加全局中间件
	cb := middleware.NewCircuitBreaker()
	cb.InjectRoutes(app.Group("/eudore/debug/breaker"))
	app.AddMiddleware(
		// add logger middleware
		middleware.NewLoggerFunc(app.App, "action", "ram", "route", "controllername", "controllermethod"),
		cb.Handle,
		// 附加额外header
		appmiddleware.NewAddHeaderFunc(),
		// cors
		middleware.NewCorsFunc(nil, map[string]string{
			"Access-Control-Allow-Credentials": "true",
			"Access-Control-Allow-Headers":     "Authorization,DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,X-Parent-Id",
			"Access-Control-Expose-Headers":    "X-Request-Id",
			"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, HEAD",
			"Access-Control-Max-Age":           "1000",
		}),
		// 流控
		middleware.NewRateFunc(10, 30),
		// gzip压缩
		middleware.NewGzipFunc(5),
		middleware.NewTimeoutFunc(5*time.Second),
		// 捕捉panic
		middleware.NewRecoverFunc(),
	)

	apiv1 := eudore.HandlerFuncs{appmiddleware.NewUserInfoFunc(app.App), appmiddleware.NewRam(app.App).NewRamFunc()}
	// 增加/api/v1使用的中间件
	app.Group("/api/v1/").AddMiddleware(apiv1...)
	// debug
	app.Group("/eudore/debug/").AddMiddleware(apiv1...)
	app.GetFunc("/eudore/debug/vars action=eudore:debug:vars", expvar.Expvar)
	pprof.RoutesInject(app.Group("/eudore/debug action=eudore:debug:pprof"))
	show.RoutesInject(app.Group("/eudore/debug action=eudore:debug:show"))
	show.RegisterObject("app", app.App)

	app.PostFunc("/eudore/csp", func(ctx eudore.Context) {
		eudore.JSON(string(ctx.Body()))
	})

	return nil
}

func InitTest(app *eudore.Eudore) error {
	client := httptest.NewClient(app)
	client.NewRequest("PUT", "/api/v1/auth/user/bind/permission/3").WithHeaderValue("Content-Type", "application/json").WithBodyString(`[{"id":4,"effect":"deny"},{"id":6,"effect":"allow"}]`) //.Do()
	if client.Next() {
		return client
	}
	return nil
}
