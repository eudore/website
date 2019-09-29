package auth

import (
	"database/sql"
	"fmt"

	// "github.com/eudore/eudore"
	"github.com/eudore/website/internal/controller"
	"github.com/eudore/website/internal/middleware"
)

/*
PostgreSQL Begin

-- 资源权限列表
CREATE SEQUENCE seq_auth_permission_id INCREMENT by 1 MINVALUE 1 START 1;
CREATE TABLE tb_auth_permission(
	"id" INTEGER PRIMARY KEY DEFAULT nextval('seq_auth_permission_id'),
	"name" VARCHAR(64) NOT NULL,
	"description" VARCHAR(512),
	"time" TIMESTAMP  DEFAULT (now())
);
COMMENT ON TABLE "public"."tb_auth_permission" IS '资源权限列表';
COMMENT ON COLUMN "tb_auth_permission"."id" IS '权限id';
COMMENT ON COLUMN "tb_auth_permission"."name" IS '权限行为';

INSERT INTO "public"."tb_auth_permission"("name", "description") VALUES ('auth:User:GetIconNameByName', '获取用户图标权限');



PostgreSQL End
*/

type (
	Permission struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	PermissionController struct {
		controller.ControllerWebsite
		Ram *middleware.Ram
	}
)

func NewPermissionController(db *sql.DB, ram *middleware.Ram) *PermissionController {
	ctl := &PermissionController{}
	ctl.DB = db
	ctl.Ram = ram
	return ctl
}

// Release 方法用于刷新ram权限信息。
func (ctl *PermissionController) Release() error {
	// 如果修改权限信息超过，则刷新ram权限信息。
	if ctl.Response().Status() == 200 && (ctl.Method() == "POST" || ctl.Method() == "PUT" || ctl.Method() == "DELETE") {
		ctl.Ram.InitPermissionInfo(ctl.DB)
	}
	return nil
}

func (ctl *PermissionController) GetIndex() (interface{}, error) {
	return ctl.QueryPages("SELECT * FROM tb_auth_permission ORDER BY id")
}

func (ctl *PermissionController) GetCount() interface{} {
	return map[string]int{"count": ctl.QueryCount("SELECT count(1) FROM tb_auth_permission")}
}

func (ctl *PermissionController) GetList() (interface{}, error) {
	return ctl.QueryRows("SELECT * FROM tb_auth_permission")
}

func (ctl *PermissionController) GetSearchByKey() (interface{}, error) {
	return ctl.QueryPages(`SELECT * FROM tb_auth_permission WHERE name ~ $1 OR description ~ $2 ORDER BY id`, ctl.GetParam("key"), ctl.GetParam("key"))
}

func (ctl *PermissionController) GetIdById() (interface{}, error) {
	id := ctl.GetParamInt("id")
	if id == 0 {
		return nil, fmt.Errorf("id is invalid %d", id)
	}
	return ctl.QueryJSON("SELECT id,name,description FROM tb_auth_permission WHERE id=$1", id)
}

func (ctl *PermissionController) GetNameByName() (interface{}, error) {
	return ctl.QueryJSON("SELECT id,name,description FROM tb_auth_permission WHERE name=$1", ctl.GetParam("name"))
}

// GetUserIdById 方法根据权限id获取用户全部用户信息
func (ctl *PermissionController) GetUserIdById() (interface{}, error) {
	return ctl.QueryRows("SELECT * FROM tb_auth_user_permission WHERE permissionid=$1", ctl.GetParam("id"))
}

// GetPermissionIdById 方法根据用户name获取用户全部权限信息
func (ctl *PermissionController) GetUserNameByName() (interface{}, error) {
	return ctl.QueryRows("SELECT * FROM tb_auth_user_permission WHERE permissionid=(SELECT id FROM tb_auth_permission WHERE name=$1)", ctl.GetParam("name"))
}

func (ctl *PermissionController) PostId() (err error) {
	var perm Permission
	err = ctl.Bind(&perm)
	if err != nil {
		return err
	}
	_, err = ctl.Exec("UPDATE tb_auth_permission SET name=$2, description=$3 WHERE id=$1", perm.ID, perm.Name, perm.Description)
	return
}

func (ctl *PermissionController) PostName() (err error) {
	var perm Permission
	err = ctl.Bind(&perm)
	if err != nil {
		return err
	}
	_, err = ctl.Exec("UPDATE tb_auth_permission SET name=$1, description=$2 WHERE name=$1", perm.Name, perm.Description)
	return
}

func (ctl *PermissionController) PutNew() (err error) {
	var perm Permission
	err = ctl.Bind(&perm)
	if err != nil {
		return err
	}
	if perm.Name == "" {
		return fmt.Errorf("put new Permission name is empty")
	}
	_, err = ctl.Exec("INSERT INTO tb_auth_permission(name,description) VALUES($1,$2)", perm.Name, perm.Description)
	return err
}

func (ctl *PermissionController) DeleteIdById() (err error) {
	_, err = ctl.Exec("DELETE FROM tb_auth_permission WHERE id=$1", ctl.GetParam("id"))
	return
}

func (ctl *PermissionController) DeleteNameByName() (err error) {
	_, err = ctl.Exec("DELETE FROM tb_auth_permission WHERE name=$1", ctl.GetParam("name"))
	return
}
