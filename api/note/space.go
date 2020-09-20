package note

import (
	"github.com/eudore/eudore"
	"github.com/eudore/website/framework"
)

type SpaceInfo struct {
	ID     int    `alias:"id" json:"id"`
	UserID int    `alias:"userid" json:"userid"`
	Name   string `alias:"name" json:"name" validate:"nozero"`
	Public bool   `alias:"public" json:"public"`
}

type SpacesController struct {
}

func (ctl *SpacesController) GetPublic(ctx framework.Context) (interface{}, error) {
	return ctx.QueryRows("SELECT * FROM tb_note_spaces WHERE public='t'")
}

func (ctl *SpacesController) PutNew(ctx framework.Context) error {
	var space SpaceInfo
	err := ctx.Bind(&space)
	if err != nil {
		return err
	}
	_, err = ctx.Exec(`INSERT INTO tb_note_spaces(userid,name,public) VALUES($1,$2,$3)`, ctx.GetParam("UID"), space.Name, space.Public)
	return err
}

func (ctl *SpacesController) GetInfoByUsername(ctx framework.Context) (interface{}, error) {
	return ctx.QueryRows("SELECT * FROM tb_note_spaces WHERE userid=(SELECT id FROM tb_auth_user_info WHERE name=$1)", ctx.GetParam("username"))
}

func (ctl *SpacesController) GetInfoByUsernameBySpacename(ctx framework.Context) (interface{}, error) {
	return ctx.QueryMap(`SELECT * FROM tb_note_spaces WHERE userid=(SELECT id FROM tb_auth_user_info WHERE name=$1)
			AND name=$2`, ctx.GetParam("username"), ctx.GetParam("spacename"))
}

func (ctl *SpacesController) DeleteInfoByUsernameBySpacename(ctx framework.Context) (interface{}, error) {
	return ctx.QueryMap(`DELETE tb_note_spaces WHERE userid=(SELECT id FROM tb_auth_user_info WHERE name=$1)
			AND name=$2`, ctx.GetParam("username"), ctx.GetParam("spacename"))
}

func (ctl *SpacesController) GetListByUsernameBySpacename(ctx framework.Context) (interface{}, error) {
	return ctx.QueryRows(`SELECT "id","nextid","userid", "spaceid", "title", "directory", "format", "tags", "createtime", "edittime" 
			FROM tb_note_content WHERE spaceid=(SELECT id FROM tb_note_spaces WHERE userid=(SELECT id FROM tb_auth_user_info WHERE name=$1)
			AND name=$2)`, ctx.GetParam("username"), ctx.GetParam("spacename"))
}

// Init 实现控制器初始方法。
func (ctl *SpacesController) Init(eudore.Context) error {
	return nil
}

// Release 实现控制器释放方法。
func (ctl *SpacesController) Release(eudore.Context) error {
	return nil
}

// ControllerRoute 方法返回默认路由信息。
func (ctl *SpacesController) ControllerRoute() map[string]string {
	return map[string]string{
		"PutNew":          " valid=true",
		"Init":            "",
		"Release":         "",
		"Inject":          "",
		"ControllerRoute": "",
	}
}

// Inject 方法实现控制器注入到路由器的方法，ControllerSingleton控制器调用ControllerInjectSingleton方法注入。
func (ctl *SpacesController) Inject(controller eudore.Controller, router eudore.Router) error {
	router = router.Group("")
	params := router.Params()
	params.Set("enable-route-extend", "true")
	params.Set("ignore-init", "true")
	params.Set("ignore-release", "true")
	return eudore.ControllerInjectSingleton(controller, router)
}
