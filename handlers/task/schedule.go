package task

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/website/internal/controller"
)

/*
trigger -> scheduler -> Executor

tb_task_trigger_cron id,name,corn,taskid
tb_task_trigger_cycle id,name,cycle,taskid
tb_task_trigger_http id,name,method,url,taskid
tb_task_task id executerid params
tb_task_executor id,type,params


tb_task_scheduler id, name, addr

*/

/*
func (sc *Scheduler) execute(id int) {
	ctx, cancel := context.WithTimeout(sc.ctx, 30*time.Second)
	defer cancel()
	task, ok := sc.tasks[id]
	if !ok {
		sc.Errorf("executer taskid %d not found", id)
		return
	}
	executer, ok := sc.executers[task.executer]
	if !ok {
		sc.Errorf("executer executerid %d not found", task.executer)
		return
	}
	err := executer.Run(context.WithValue(ctx, "params", task.params))
	if err != nil {
		sc.Errorf("executer %d fatal error: %v,params: %v", task.id, err, task.params)
	} else {
		sc.Infof("executer %d succss", task.id)
	}
}
*/
type ScheduleController struct {
	controller.ControllerWebsite
	Context   context.Context
	Logger    eudore.Logger
	Name      *string
	Event     chan *Event
	Triggers  map[string]Trigger
	Executors map[int]Executor
}

func NewScheduleController(app *eudore.App, db *sql.DB) *ScheduleController {
	name := "master"
	ctl := &ScheduleController{
		ControllerWebsite: *controller.NewControllerWejass(db),
		Context:           app.Context,
		Logger:            app.Logger,
		Name:              &name,
		Event:             make(chan *Event),
		Triggers:          make(map[string]Trigger),
		Executors:         make(map[int]Executor),
	}
	ctl.Triggers["http"] = NewHttpTrigger(ctl.Event)
	ctl.Triggers["cron"] = NewCronTrigger(ctl.Event)
	ctl.InitTriggers()
	ctl.InitExecutors()
	go ctl.Run()
	return ctl
}

// ControllerRoute 方法返回默认路由信息。
func (ctl *ScheduleController) ControllerRoute() map[string]string {
	return map[string]string{
		"InitTriggers": "",
		"Run":          "",
	}
}

func (ctl *ScheduleController) InitTriggers() error {
	data, err := ctl.QueryRowsContext(context.Background(), "SELECT * FROM tb_task_trigger WHERE schedule=$1 or schedule=''", *ctl.Name)
	if err != nil {
		panic(err)
	}
	for _, i := range data {
		err = ctl.Triggers[i["event"].(string)].AddEntry(i)
		if err != nil {
			ctl.Logger.Error("new entry error:", err)
		}
	}
	return nil
}

func (ctl *ScheduleController) InitExecutors() error {
	data, err := ctl.QueryRowsContext(context.Background(), "SELECT * FROM tb_task_executor")
	if err != nil {
		panic(err)
	}
	for _, i := range data {
		ctl.Logger.Info("new entry", i)
		switch i["type"] {
		case "http":
			ctl.Executors[int(i["id"].(int64))] = NewHttpExecutor()
		case "agent":
			ctl.Executors[int(i["id"].(int64))] = NewAgentExecutor(i["config"].(string))
		}
	}
	return nil
}

func (ctl *ScheduleController) Run() {
	for {
		select {
		case event := <-ctl.Event:
			task := &Task{
				Event:     *event,
				Context:   context.Background(),
				Starttime: time.Now(),
			}
			// ctx := context.WithValue(ctl.Context, "params", event.Params)
			if err := ctl.Executors[event.ExecutorId].Run(task); err != nil {
				task.Err = err
			}
			task.Endtime = time.Now()
			ctl.writeTask(task)

			ctl.Logger.Info("schedule run", event, ctl.Executors)
		case <-ctl.Context.Done():
			return
		}
	}
}

func (ctl *ScheduleController) writeTask(task *Task) {
	var status int
	if task.Err != nil {
		if task.Err == context.Canceled || task.Err == context.DeadlineExceeded {
			status = 2
		} else {
			status = 1
			task.Message = []byte(task.Err.Error())
		}
	}
	_, err := ctl.ExecContext(ctl.Context, `INSERT INTO tb_task_logger(status,params,message,starttime,endtime,eventid,executorid)
	 VALUES($1,$2,$3,$4,$5,$6,$7)`, status, "", task.Message, task.Starttime, task.Endtime, task.Event.Id, task.Event.ExecutorId)
	if err != nil {
		fmt.Println(task, err)
	}
}

// Inject 方法实现控制器注入到路由器的方法,调用ControllerBaseInject方法注入。
func (ctl *ScheduleController) Inject(controller eudore.Controller, router eudore.RouterMethod) error {
	router.Group("").SetParam("route", "").AnyFunc("/task/trigger/http/*path", ctl.Triggers["http"].(*HttpTrigger).Handle)
	return ctl.ControllerWebsite.Inject(controller, router)
}

func (ctl *ScheduleController) PutNewTask() {}
func (ctl *ScheduleController) PostById()   {}
