package note

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/eudore/website/framework"
)

type Content struct {
	ID         int       `alias:"id" json:"id"`
	NextID     int       `alias:"nextid" json:"nextid"`
	SpaceID    int       `alias:"spaceid" json:"spaceid"`
	Title      string    `alias:"title" json:"title"`
	Directory  string    `alias:"directory" json:"directory"`
	Format     string    `alias:"format" json:"format"`
	Tags       []string  `alias:"tags" json:"tags" splitchar:","`
	Content    string    `alias:"content" json:"content"`
	ContentURL string    `alias:"contenturl" json:"contenturl"`
	CreateTime time.Time `alias:"createtime" json:"createtime"`
	EditTime   time.Time `alias:"edittime" json:"edittime"`
}

func getSpaces(ctx framework.Context) (spaceid int, err error) {
	user := ctx.GetParam("username")
	space := ctx.GetParam("spacename")
	err = ctx.QueryRow("SELECT id FROM tb_note_spaces WHERE userid=(SELECT id FROM tb_auth_user_info WHERE name=$1) AND name=$2", user, space).Scan(&spaceid)
	if err != nil {
		err = fmt.Errorf("not found user '%s' space '%s'", user, space)
	}
	return spaceid, err
}

func getTitleDirector(path string) (string, string) {
	title := ""
	pos := strings.LastIndexByte(path, '/')
	if pos == -1 {
		path, title = title, path
	} else {
		title = path[pos+1:]
		path = path[:pos]
	}
	return title, path
}

func GetContent(ctx framework.Context) (interface{}, error) {
	// spaceid
	spaceid, err := getSpaces(ctx)
	if err != nil {
		return nil, err
	}
	// directory title
	title, path := getTitleDirector(ctx.GetParam("*"))

	// content
	content := new(Content)
	err = ctx.QueryBind(content, "SELECT * FROM tb_note_content WHERE spaceid=$1 AND directory=$2 AND title=$3", spaceid, path, title)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(content.Format, "url-") && ctx.GetQuery("parseformat") != "false" {
		resp, err := http.Get(content.Content)
		if err != nil {
			return content, nil
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return content, nil
		}
		content.Format = content.Format[4:]
		content.ContentURL = content.Content
		content.Content = string(body)
	}
	return content, nil
}

// post fields: title directory format tags content
func PostContent(ctx framework.Context) (sql.Result, error) {
	var content Content
	err := ctx.Bind(&content)
	if err != nil {
		return nil, err
	}
	spaceid, err := getSpaces(ctx)
	if err != nil {
		return nil, err
	}
	title, path := getTitleDirector(ctx.GetParam("*"))
	return ctx.Exec(`UPDATE tb_note_content SET title=$1,Directory=$2,Format=$3,Tags=$4,Content=$5,EditTime=now() 
		WHERE title=$6 AND directory=$7 AND spaceid=$8`, content.Title, content.Directory, content.Format,
		strings.Join(content.Tags, ","), content.Content, title, path, spaceid)
}

func PutContent(ctx framework.Context) (sql.Result, error) {
	var content Content
	err := ctx.Bind(&content)
	if err != nil {
		return nil, err
	}
	spaceid, err := getSpaces(ctx)
	if err != nil {
		return nil, err
	}
	title, path := getTitleDirector(ctx.GetParam("*"))
	return ctx.Exec(`INSERT INTO tb_note_content(nextid,spaceid,title,directory,format,tags,content,edittime,createtime) 
		VAlUES($1,$2,$3,$4,$5,$6,$7,now(),now())`, content.NextID, spaceid, title, path, content.Format,
		strings.Join(content.Tags, ","), content.Content)
}

func DeleteContent(ctx framework.Context) (sql.Result, error) {
	spaceid, err := getSpaces(ctx)
	if err != nil {
		return nil, err
	}
	title, path := getTitleDirector(ctx.GetParam("*"))
	return ctx.Exec(`DELETE tb_note_content WHERE spaceid=$1 AND title=$2 AND=directory=$3`, spaceid, title, path)
}
