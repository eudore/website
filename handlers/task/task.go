package task

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/website/internal/controller"
)

/*
PostgreSQL Begin
-- task执行日志
CREATE TABLE tb_task_logger(
	"status" INTEGER,
	"params" VARCHAR(512),
	"message" TEXT,
	"starttime" TIMESTAMP,
	"endtime" TIMESTAMP,

	"eventid" INTEGER,
	"executorid" INTEGER
);


PostgreSQL End
*/

type (

	// Event 定义静态事件
	Event struct {
		Id          int
		Name        string
		Description string
		Params      map[string]interface{}
		ExecutorId  int
	}
	// Task 定义运行的一个任务
	Task struct {
		Event
		Context   context.Context
		Starttime time.Time
		Endtime   time.Time
		Err       error
		Message   []byte
	}
	TaskController struct {
		controller.ControllerWebsite
	}
)

func NewEvent(data map[string]interface{}) *Event {
	m := eudore.StringMap(data)
	e := &Event{
		Id:          m.GetInt("id"),
		Name:        m.GetString("name"),
		Description: m.GetString("description"),
		ExecutorId:  m.GetInt("executorid"),
		Params:      make(map[string]interface{}),
	}
	json.Unmarshal([]byte(m.GetString("params")), &e.Params)
	return e
}

func NewTaskController(db *sql.DB) *TaskController {
	return &TaskController{
		ControllerWebsite: *controller.NewControllerWejass(db),
	}
}

func (ctl *TaskController) GetIndex() (interface{}, error) {
	return ctl.QueryPages(`SELECT * FROM "tb_task_logger" ORDER BY "starttime" desc`)
}

func (ctl *TaskController) GetCount() interface{} {
	return map[string]int{"count": ctl.QueryCount("SELECT count(1) FROM tb_task_logger")}
}

func (ctl *TaskController) GetList() (interface{}, error) {
	return ctl.QueryRows(`SELECT * FROM "tb_task_logger" ORDER BY "starttime" desc`)
}

func (ctl *TaskController) GetInfoEventIdById() (interface{}, error) {
	return ctl.QueryJSON(`SELECT * FROM "tb_task_logger" WHERE eventid=$1 ORDER BY "starttime" DESC`, ctl.GetParam("id"))
}
func (ctl *TaskController) GetInfoExecutorById() (interface{}, error) {
	return ctl.QueryJSON(`SELECT * FROM "tb_task_logger" WHERE executorid=$1 ORDER BY "starttime" DESC`, ctl.GetParam("id"))
}
