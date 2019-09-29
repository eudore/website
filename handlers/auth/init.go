package auth

import (
	"database/sql"
	"fmt"

	"github.com/eudore/eudore"

	// "wejass/handlers/auth/oauth2"
	"github.com/eudore/website/internal/controller"
	"github.com/eudore/website/internal/middleware"
	"github.com/eudore/website/util/jwt"
	"github.com/eudore/website/util/tool"
)

var (
	stmtQueryOauth2Pass  *sql.Stmt
	stmtUpdateOauth2Pass *sql.Stmt

	stmtQueryUserIdByName   *sql.Stmt
	stmtQueryUserInfo       *sql.Stmt
	stmtQueryUserInfoByName *sql.Stmt
	stmtQueryUserList       *sql.Stmt
	stmtQueryUserCount      *sql.Stmt

	stmtQueryUserIconData  *sql.Stmt
	stmtInsertUserIconData *sql.Stmt
	stmtInsertUser         *sql.Stmt
	// stmtUpdateSignUp			*sql.Stmt

	tokenVerify jwt.VerifyFunc
)

func init() {
	tokenVerify = jwt.NewVerifyHS256([]byte("secret"))
}

func Reload(app *eudore.Eudore) error {
	db, ok := app.Config.Get("keys.db").(*sql.DB)
	if !ok {
		return fmt.Errorf("keys.db not find database.")
	}

	ram, ok := app.Config.Get("keys.ram").(*middleware.Ram)
	if !ok {
		return fmt.Errorf("auth.NewUserController require *middleware.Ram")
	}

	var sqls = map[**sql.Stmt]string{
		&stmtQueryOauth2Pass:   "SELECT pass, salt, id FROM tb_auth_user_pass WHERE name=$1;",
		&stmtUpdateOauth2Pass:  "UPDATE tb_auth_user_pass SET pass=$1 WHERE name=$2",
		&stmtQueryUserIdByName: "SELECT id FROM tb_auth_user_info WHERE name=$1;",
		&stmtInsertUser:        "INSERT INTO tb_auth_user_info(name,loginip,sigintime) VALUES($1,$2,$3);",

		// user
		&stmtQueryUserInfo:       `SELECT name,status,level,COALESCE(mail, ''),COALESCE(tel, ''),loginip,logintime,sigintime FROM tb_auth_user_info WHERE id=$1;`,
		&stmtQueryUserInfoByName: `SELECT id,status,level,COALESCE(mail, ''),COALESCE(tel, ''),loginip,logintime,sigintime FROM tb_auth_user_info WHERE name=$1;`,
		&stmtQueryUserList:       "SELECT id,name,status,mail,loginip,logintime FROM tb_auth_user_info ORDER BY id limit $1 OFFSET $2",
		&stmtQueryUserCount:      "SELECT count(1) FROM tb_auth_user_info;",
		// user icon
		&stmtQueryUserIconData:  "SELECT data FROM tb_auth_user_icon WHERE id=$1;",
		&stmtInsertUserIconData: `INSERT INTO tb_auth_user_icon("user","data") VALUES($1,$2);`,
	}
	err := tool.InitStmt(db, sqls)
	if err != nil {
		return err
	}

	staticController := controller.NewControllerStatic()

	auth := app.Group("/auth")
	auth.GetFunc("/login/wejass", staticController.NewHTMLHandlerFunc("static/html/auth/login.html"))
	// auth.GetFunc("/login/:source", getLoginWejass)
	auth.GetFunc("/signup", authHandler)
	// auth.GetFunc("/icon/:name", getIconByName)
	// auth.GetFunc("/user/setting", userSetting)
	// auth.AnyFunc("/", authHandler)
	auth.AnyFunc("/", staticController.NewHTMLHandlerFunc("static/html/auth/index.html"))
	auth.AnyFunc("/*path", authOtherHandler)

	api := app.Group("/api/v1/auth")
	// api.GetFunc("/login/wejass", getLoginWejass)
	api.PostFunc("/login/wejass", postLoginWejass)
	api.GetFunc("/login/:source", authHandler)
	api.GetFunc("/callback/:source", authHandler)
	api.HeadFunc("/logout", authHandler)
	api.GetFunc("/refresh")

	api.GetFunc("/captcha", getCaptcha)

	api.AddController(NewUserController(db, ram))
	api.AddController(NewPermissionController(db, ram))
	api.AddController(NewRoleController(db, ram))
	api.AddController(NewPolicyController(db, ram))

	// setUserPass("root", "123")
	return nil
}

func authHandler(ctx eudore.Context) {
	// ctx.WriteHTMLWithPush("static/html/auth/index.html")
}

func authOtherHandler(ctx eudore.Context) {
	ctx.Redirect(302, "/auth/#!/"+ctx.GetParam("path"))
}
