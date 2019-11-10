package note

import (
	"database/sql"
	"github.com/eudore/website/internal/controller"
)

type (
	IndexController struct {
		controller.ControllerWebsite
	}
)

func NewIndexController(db *sql.DB) *IndexController {
	return &IndexController{
		ControllerWebsite: controller.ControllerWebsite{
			DB: db,
		},
	}
}

func (ctl *IndexController) Get() (interface{}, error) {
	return ctl.QueryPages("SELECT * FROM tb_note_content ORDER BY id")
}

func (ctl *IndexController) GetCount() interface{} {
	return map[string]int{"count": ctl.QueryCount("SELECT count(1) FROM tb_note_content")}
}

func (ctl *IndexController) GetList() (interface{}, error) {
	return ctl.QueryRows("SELECT * FROM tb_note_content")
}

func (ctl *IndexController) GetNameByName() (interface{}, error) {
	return ctl.QueryPages("SELECT * FROM tb_note_content WHERE ownerid=(SELECT id FROM tb_auth_user_info WHERE namr=$1) ORDER BY id", ctl.GetParam("name"))
}
