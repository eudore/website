package term

import (
	// "database/sql"
	// "github.com/eudore/eudore"
	"github.com/eudore/website/framework"
	"net/http"
)

func Init(app *framework.App) error {
	InitServer(app)

	api := app.Group("/api/v1/term")

	api.AddController(framework.NewTableController("term", "user", "tb_term_user", new(User), app.DB))
	api.AddController(framework.NewTableController("term", "host", "tb_term_host", new(Host), app.DB))
	api.AddController(framework.NewTableController("term", "hostgroup", "tb_term_hostgroup", new(Host), app.DB))
	api.AddController(framework.NewTableController("term", "video", "tb_term_video", nil, app.DB))

	// eudore.Set(app.Router, "print", eudore.NewPrintFunc(app.App))
	api2 := app.Group("/api/v1/term")
	api2.AddController(framework.NewTableController("term", "userhost", "tb_term_user_host", nil, app.DB))
	api2.AddController(framework.NewViewController("term", "userhost", "vi_term_user_host", new(ViewUserHost), app.DB))
	api2.AddController(framework.NewViewController("term", "userhostgroup", "vi_term_user_hostgroup", new(ViewUserHostgroup), app.DB))
	api2.AddController(framework.NewViewController("term", "hosthostgroup", "vi_term_host_hostgroup", new(ViewHostHostgroup), app.DB))

	// controller.NewTable2Struct(db, "vi_term_host_hostgroup")
	return nil
}

func InitServer(app *framework.App) {
	srv := NewServer(app.Context, app.DB)
	srv.AddHostKey()
	srv.SSHConfig.PasswordCallback = newCheckPassword(app.DB)
	srv.SSHConfig.PublicKeyCallback = newCheckPublicKey(app.DB)
	app.GetFunc("/api/v1/term/connect", srv.NewHandleHTTP())
	go srv.ListenAndServe(":8088")

	s := &http.Server{
		Addr:    app.Config.Term.Addr,
		Handler: http.HandlerFunc(handlebash),
	}
	go s.ListenAndServe()
}
