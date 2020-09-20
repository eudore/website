package note

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/eudore/eudore"
	"github.com/eudore/website/framework"
)

/*


PostgreSQL Begin
CREATE SEQUENCE seq_note_content_id INCREMENT by 1 MINVALUE 1 START 1;
CREATE TABLE tb_note_content(
	"id" INTEGER PRIMARY KEY DEFAULT nextval('seq_note_content_id'),
	"nextid" INTEGER DEFAULT 0,
	"userid" INTEGER DEFAULT 0,
	"spaceid" INTEGER DEFAULT 0,
	"title" VARCHAR(50) NOT NULL,
	"directory" VARCHAR(128) NOT NULL,
	"format" VARCHAR(8) DEFAULT 'md',
	"tags" VARCHAR(128) DEFAULT '',
	"content" TEXT DEFAULT '',
	"createtime" TIMESTAMP DEFAULT (now()),
	"edittime" TIMESTAMP DEFAULT (now())
);

CREATE SEQUENCE seq_note_spaces_id INCREMENT by 1 MINVALUE 1 START 1;
CREATE TABLE tb_note_spaces(
	"id" INTEGER PRIMARY KEY DEFAULT nextval('seq_note_spaces_id'),
	"userid" INTEGER DEFAULT 0,
	"name" VARCHAR(32) NOT NULL,
	"gitpath" VARCHAR(128) DEFAULT '',
	"public" boolean DEFAULT 'f'
)


CREATE SEQUENCE seq_note_comment_id INCREMENT by 1 MINVALUE 1 START 1;
CREATE TABLE tb_note_comment(
	"id" INTEGER PRIMARY KEY DEFAULT nextval('seq_note_comment_id'),
	"userid" INTEGER DEFAULT 0,
	"noteid" INTEGER DEFAULT 0,
	"replyid" INTEGER DEFAULT 0,
	"format" VARCHAR(8),
	"content" TEXT,
	"createtime" TIMESTAMP DEFAULT (now()),
	"edittime" TIMESTAMP
);

*/


// Init 函数定义初始化内容。
func Init(app *framework.App) error {
	go initGitwork(app)
	api := app.Group("/api/v1/note")
	api.AddController(new(SpacesController))

	// content
	api.GetFunc("/content/:username/:spacename/* action=note:Content:GetContent", GetContent)
	api.PostFunc("/content/:username/:spacename/* action=note:Content:PostContent", PostContent)
	api.PutFunc("/content/:username/:spacename/* action=note:Content:PutContent", PutContent)
	api.DeleteFunc("/content/:username/:spacename/* action=note:Content:DeleteContent", DeleteContent)
	return nil
}

func initGitwork(app *framework.App) {
	workdir := eudore.GetString(app.Config.Note.Workdir, "note/workdir")
	gitpath := eudore.GetString(app.Config.Note.Workdir, "git")
	rows, err := app.Query("SELECT name,path FROM tb_note_gitnote")
	if err != nil {
		app.Error(err)
		return
	}
	var name, path string
	for rows.Next() {
		rows.Scan(&name, &path)
		name = filepath.Join(workdir, name)
		_, err := os.Stat(name)
		if err != nil {
			app.Warning(err)
			break
		}
		exec.CommandContext(app.Context, fmt.Sprintf("%s --git-dir=%s git clone %s", gitpath, name, path)).Run()
		exec.CommandContext(app.Context, fmt.Sprintf("%s --git-dir=%s git pull", gitpath, name)).Run()
	}
}
