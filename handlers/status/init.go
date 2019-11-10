package status

import (
	"github.com/eudore/eudore"
	"github.com/eudore/website/internal/controller"
)

// Init 函数定义status初始化内容。
func Init(app *eudore.Eudore) error {
	auth := app.Group("/auth")
	auth.GetFunc("/login/wejass")
	app.AnyFunc("/status/*path", handleOtherStatus)
	app.AnyFunc("/status/", controller.NewControllerStatic().NewHTMLHandlerFunc("static/html/status/index.html"))

	api := app.Group("/api/v1/status")
	api.GetFunc("/app", getSystem)
	api.GetFunc("/build", getBuild)
	api.GetFunc("/system", getSystem)
	api.GetFunc("/config", getConfig(app.App))
	api.GetFunc("/web", getSystem)
	return nil
}

func handleOtherStatus(ctx eudore.Context) {
	ctx.Redirect(302, "/status/#!/"+ctx.GetParam("path"))
}
