package note

import (
	"database/sql"
	"github.com/eudore/website/internal/controller"
)

/*
PostgreSQL Begin
CREATE TABLE tb_note_consent(
	"path" VARCHAR(128),
	"userid" INTEGER DEFAULT 0,
	"createtime" TIMESTAMP DEFAULT (now()),
	PRIMARY KEY("path", "userid")
);

PostgreSQL End
*/

type (
	ConsentController struct {
		controller.ControllerWebsite
	}
)

func NewConsentController(db *sql.DB) *ConsentController {
	return &ConsentController{
		ControllerWebsite: controller.ControllerWebsite{
			DB: db,
		},
	}
}

func (ctl *ConsentController) ControllerRoute() map[string]string {
	return map[string]string{
		"Get":    "/:username/*path",
		"Put":    "/:username/*path",
		"Delete": "/:username/*path",
	}
}
func (ctl *ConsentController) Get() (interface{}, error) {
	return ctl.QueryJSON("SELECT * FROM tb_note_content WHERE path=$1;", ctl.GetParam("username")+"/"+ctl.GetParam("path"))
}
func (ctl *ConsentController) Put()    {}
func (ctl *ConsentController) Delete() {}
