package task

import (
	"database/sql"
	"fmt"
	"github.com/eudore/eudore"
	"github.com/eudore/website/internal/controller"
)

/*

task

trigger
scheduler
executor
*/
func Init(app *eudore.Eudore) error {
	db, ok := app.Config.Get("keys.db").(*sql.DB)
	if !ok {
		return fmt.Errorf("keys.db not find database.")
	}

	/*	sc := NewScheduler(app.App)
		go sc.Run()

		sc.tasks[9] = &Task{id: 9, executer: 5, params: map[string]interface{}{"method": "GET", "url": "127.0.0.1", "body": ""}}
		sc.tasks[10] = &Task{id: 10, executer: 5, params: map[string]interface{}{"method": "GET", "url": "http://127.0.0.1", "body": ""}}
		sc.executers[5] = NewHttpExecutor()*/

	// httpTrigger := NewHttpTrigger(sc.Event)
	// httpTrigger.AddTrigger("PUT", "9", 9)
	// httpTrigger.AddTrigger("PUT", "10", 10)
	// app.AnyFunc("/task/trigger/http/*path", httpTrigger.Handle)

	app.AnyFunc("/task/", controller.NewControllerStatic().NewHTMLHandlerFunc("static/html/task/index.html"))
	app.AnyFunc("/task/*path", handleOtherTask)

	api := app.Group("/api/v1/task")
	api.AddController(NewScheduleController(app.App, db))
	api.AddController(NewTriggerController(db))
	api.AddController(NewExecutorController(db))
	api.AddController(NewTaskController(db))
	return nil
}

func handleOtherTask(ctx eudore.Context) {
	ctx.Redirect(302, "/task/#!/"+ctx.GetParam("path"))
}
