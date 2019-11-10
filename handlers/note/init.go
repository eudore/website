package note

import (
	"database/sql"
	"fmt"
	"github.com/eudore/eudore"
	"github.com/eudore/website/internal/controller"
)

func Init(app *eudore.Eudore) error {
	db, ok := app.Config.Get("keys.db").(*sql.DB)
	if !ok {
		return fmt.Errorf("keys.db not find database.")
	}

	app.AnyFunc("/note/", func(ctx eudore.Context) {
		ctx.SetHeader("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' data: 'unsafe-inline'; font-src 'self' data:; report-uri /eudore/csp")
	}, controller.NewControllerStatic().NewHTMLHandlerFunc("static/html/note/index.html"))
	app.AnyFunc("/note/*path", handleOtherNote)

	api := app.Group("/api/v1/note")
	api.AddController(NewNoteController(db))
	api.AddController(NewIndexController(db))
	api.AddController(NewContentController(db))
	api.AddController(NewCommentController(db))
	api.AddController(NewConsentController(db))

	return nil
}

func handleOtherNote(ctx eudore.Context) {
	ctx.Redirect(302, "/note/#!/"+ctx.GetParam("path"))
}

func checkOwner(db *sql.DB) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
	}
}
