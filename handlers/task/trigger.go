package task

import (
	"database/sql"

	"github.com/eudore/website/internal/controller"
)

/*
PostgreSQL Begin
-- task触发器
CREATE SEQUENCE seq_task_trigger_id INCREMENT by 1 MINVALUE 1 START 1;
CREATE TABLE tb_task_trigger(
	"id" INTEGER PRIMARY KEY DEFAULT nextval('seq_task_trigger_id'),
	"name" VARCHAR(64),
	"description" VARCHAR(512),
	"event" VARCHAR(64),
	"params" VARCHAR(512),
	"schedule" VARCHAR(64),
	"executorid" INTEGER,
	"time" TIMESTAMP  DEFAULT (now())
);
COMMENT ON TABLE "public"."tb_task_trigger" IS 'task触发器';
COMMENT ON COLUMN "tb_task_trigger"."id" IS '触发器id';
COMMENT ON COLUMN "tb_task_trigger"."name" IS '触发器名称';
COMMENT ON COLUMN "tb_task_trigger"."event" IS '触发器事件';
COMMENT ON COLUMN "tb_task_trigger"."executorid" IS '触发器绑定的执行executor';

INSERT INTO "public"."tb_task_trigger"("name", "description", "event", "params", "executorid") VALUES ('http-test', ' ', 'http', '{"method":"GET","url":"9","route":"/:num|isnum", "async": true}', 9);

PostgreSQL End
*/
type (
	Trigger interface {
		AddEntry(map[string]interface{}) error
	}
	TriggerController struct {
		controller.ControllerWebsite
	}
)

func NewTriggerController(db *sql.DB) *TriggerController {
	return &TriggerController{
		ControllerWebsite: controller.ControllerWebsite{
			DB: db,
		},
	}
}

func (ctl *TriggerController) GetIndex() (interface{}, error) {
	return ctl.QueryPages("SELECT * FROM tb_task_trigger ORDER BY id")
}

func (ctl *TriggerController) GetCount() interface{} {
	return map[string]int{"count": ctl.QueryCount("SELECT count(1) FROM tb_task_trigger")}
}

func (ctl *TriggerController) GetList() (interface{}, error) {
	return ctl.QueryRows("SELECT * FROM tb_task_trigger")
}

func (ctl *TriggerController) GetInfoIdById() (interface{}, error) {
	return ctl.QueryJSON("SELECT * FROM tb_task_trigger WHERE id=$1;", ctl.GetParam("id"))
}
func (ctl *TriggerController) GetInfoNameByName() (interface{}, error) {
	return ctl.QueryJSON("SELECT * FROM tb_task_trigger WHERE name=$1;", ctl.GetParam("name"))
}
