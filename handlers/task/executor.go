package task

import (
	// "context"
	"database/sql"
	"github.com/eudore/website/internal/controller"
)

/*
PostgreSQL Begin
CREATE SEQUENCE seq_task_executor_id INCREMENT by 1 MINVALUE 1 START 1;
CREATE TABLE tb_task_executor(
	"id" INTEGER PRIMARY KEY DEFAULT nextval('seq_task_executor_id'),
	"name" VARCHAR(64),
	"description" VARCHAR(512),
	"type" VARCHAR(64),
	"config" VARCHAR(512),
	"time" TIMESTAMP  DEFAULT (now())
);

PostgreSQL End
*/
type (
	Executor interface {
		Run(*Task) error
	}
	ExecutorController struct {
		controller.ControllerWebsite
	}
)

func NewExecutorController(db *sql.DB) *ExecutorController {
	return &ExecutorController{
		ControllerWebsite: *controller.NewControllerWejass(db),
	}
}

func (ctl *ExecutorController) GetIndex() (interface{}, error) {
	return ctl.QueryPages("SELECT * FROM tb_task_executor ORDER BY id")
}

func (ctl *ExecutorController) GetCount() interface{} {
	return map[string]int{"count": ctl.QueryCount("SELECT count(1) FROM tb_task_executor")}
}

func (ctl *ExecutorController) GetList() (interface{}, error) {
	return ctl.QueryRows("SELECT * FROM tb_task_executor")
}

func (ctl *ExecutorController) GetInfoIdById() (interface{}, error) {
	return ctl.QueryJSON("SELECT * FROM tb_task_executor WHERE id=$1;", ctl.GetParam("id"))
}
func (ctl *ExecutorController) GetInfoNameByName() (interface{}, error) {
	return ctl.QueryJSON("SELECT * FROM tb_task_executor WHERE name=$1;", ctl.GetParam("name"))
}

func (ctl *ExecutorController) PutAgentRegister() {
	var ac AgentConfig
	ctl.Bind(&ac)
	ac.Addr = ctl.RealIP()
	ctl.Debug(ctl.GetQuery("server"), ac)
	ctl.Exec("UPDATE tb_task_executor SET name=$1,config=$2,time=now() WHERE type='agent'", ac.Name, ac.EncodeConfig())
}
