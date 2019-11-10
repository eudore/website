package note

import (
	"database/sql"
	"github.com/eudore/website/internal/controller"
)

/*
PostgreSQL Begin
CREATE SEQUENCE seq_note_comment_id INCREMENT by 1 MINVALUE 1 START 1;
CREATE TABLE tb_note_comment(
	"id" INTEGER PRIMARY KEY DEFAULT nextval('seq_note_comment_id'),
	"path" VARCHAR(128),
	"userid" INTEGER DEFAULT 0,
	"format" VARCHAR(8),
	"content" TEXT,
	"createtime" TIMESTAMP DEFAULT (now()),
	"edittime" TIMESTAMP
);

PostgreSQL End
*/
type (
	CommentController struct {
		controller.ControllerWebsite
	}
)

func NewCommentController(db *sql.DB) *CommentController {
	return &CommentController{
		ControllerWebsite: controller.ControllerWebsite{
			DB: db,
		},
	}
}

func (ctl *CommentController) ControllerRoute() map[string]string {
	return map[string]string{
		"Get":    "/:username/*path",
		"Put":    "/:username/*path",
		"Post":   "/:username/*path",
		"Delete": "/:username/*path",
	}
}
func (ctl *CommentController) Get() (interface{}, error) {
	return ctl.QueryRows("SELECT C.*,U.name FROM tb_note_comment as C JOIN tb_auth_user_info as U ON C.userid=U.id WHERE path=$1;", ctl.GetParam("username")+"/"+ctl.GetParam("path"))
}
func (ctl *CommentController) Put()    {}
func (ctl *CommentController) Post()   {}
func (ctl *CommentController) Delete() {}
