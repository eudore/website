package chat

import (
	"database/sql"
	"fmt"
	"github.com/eudore/eudore"
	"github.com/eudore/website/internal/controller"
)

// Init 函数定义chat初始化内容。
func Init(app *eudore.Eudore) error {
	db, ok := app.Config.Get("keys.db").(*sql.DB)
	if !ok {
		return fmt.Errorf("keys.db not find database.")
	}

	app.AnyFunc("/chat/", controller.NewControllerStatic().NewHTMLHandlerFunc("static/html/chat.html"))
	app.AnyFunc("/chat/*path", handleOtherChat)

	api := app.Group("/api/v1/chat")
	api.AddController(NewMessageController(app.App, db))

	return nil
}

func handleOtherChat(ctx eudore.Context) {
	ctx.Redirect(302, "/chat/#!/"+ctx.GetParam("path"))
}
