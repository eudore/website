package auth

import (
	"database/sql"
	"fmt"

	"github.com/eudore/eudore"
	"github.com/eudore/website/internal/controller"
	"github.com/eudore/website/internal/middleware"
)

// Init 函数初始化auth部分内容。
func Init(app *eudore.Eudore) error {
	// 获取 *sql.DB
	db, ok := app.Config.Get("keys.db").(*sql.DB)
	if !ok {
		return fmt.Errorf("keys.db not find database.")
	}

	// 获取 *middleware.Ram
	ram, ok := app.Config.Get("keys.ram").(*middleware.Ram)
	if !ok {
		return fmt.Errorf("auth.NewUserController require *middleware.Ram")
	}

	// 创建静态控制器处理页面渲染
	staticController := controller.NewControllerStatic()
	auth := app.Group("/auth")
	auth.GetFunc("/login/website", staticController.NewHTMLHandlerFunc("static/html/auth/login.html"))
	auth.GetFunc("/user/setting", staticController.NewHTMLHandlerFunc("static/html/auth/setting.html"))
	auth.AnyFunc("/", staticController.NewHTMLHandlerFunc("static/html/auth/index.html"))
	auth.AnyFunc("/*path", authOtherHandler)
	// auth.GetFunc("/signup")

	// 注册api控制器
	api := app.Group("/api/v1/auth")
	api.AddController(NewUserController(app, db, ram))
	api.AddController(NewLoginController(app, db))
	api.AddController(NewPermissionController(db, ram))
	api.AddController(NewRoleController(db, ram))
	api.AddController(NewPolicyController(db, ram))

	return nil
}

func authOtherHandler(ctx eudore.Context) {
	ctx.Redirect(302, "/auth/#!/"+ctx.GetParam("path"))
}
