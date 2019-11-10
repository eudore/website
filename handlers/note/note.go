package note

import (
	// "html/template"
	"database/sql"
	"github.com/eudore/website/internal/controller"
)

/*
PostgreSQL Begin
CREATE SEQUENCE seq_note_content_id INCREMENT by 1 MINVALUE 1 START 1;
CREATE TABLE tb_note_content(
	"id" INTEGER PRIMARY KEY DEFAULT nextval('seq_note_content_id'),
	"status" INTEGER DEFAULT 0,
	"userid" INTEGER DEFAULT 0,
	"path" VARCHAR(128),
	"ppath" VARCHAR(128) DEFAULT "",
	"format" VARCHAR(8),
	"title" VARCHAR(50),
	"topics" VARCHAR(128) DEFAULT "",
	"content" TEXT,
	"createtime" TIMESTAMP DEFAULT (now()),
	"edittime" TIMESTAMP
);

PostgreSQL End
*/
type (
	NoteController struct {
		controller.ControllerWebsite
	}
	ContentController struct {
		controller.ControllerWebsite
	}
)

func NewNoteController(db *sql.DB) *NoteController {
	return &NoteController{
		ControllerWebsite: controller.ControllerWebsite{
			DB: db,
		},
	}
}

func (ctl *NoteController) GetIndex() (interface{}, error) {
	return ctl.QueryPages("SELECT * FROM tb_note_content ORDER BY id")
}

func (ctl *NoteController) GetCount() interface{} {
	return map[string]int{"count": ctl.QueryCount("SELECT count(1) FROM tb_note_content")}
}

func (ctl *NoteController) GetList() (interface{}, error) {
	return ctl.QueryRows("SELECT * FROM tb_note_content")
}

func (ctl *NoteController) GetIndexNameByName() (interface{}, error) {
	return ctl.QueryPages("SELECT * FROM tb_note_content WHERE userid=(SELECT id FROM tb_auth_user_info WHERE namr=$1) ORDER BY id", ctl.GetParam("name"))
}

func NewContentController(db *sql.DB) *ContentController {
	return &ContentController{
		ControllerWebsite: controller.ControllerWebsite{
			DB: db,
		},
	}
}

func (ctl *ContentController) ControllerRoute() map[string]string {
	return map[string]string{
		"Get":    "/:username/*path",
		"Put":    "/:username/*path",
		"Post":   "/:username/*path",
		"Delete": "/:username/*path",
	}
}
func (ctl *ContentController) Get() (interface{}, error) {
	return ctl.QueryJSON("SELECT * FROM tb_note_content WHERE path=$1;", ctl.GetParam("username")+"/"+ctl.GetParam("path"))
}
func (ctl *ContentController) Put()    {}
func (ctl *ContentController) Post()   {}
func (ctl *ContentController) Delete() {}
