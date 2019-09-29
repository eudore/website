package auth

import (
	"database/sql"

	"github.com/eudore/website/internal/controller"
	"github.com/eudore/website/internal/middleware"
)

/*
PostgreSQL Begin

-- 角色信息表
CREATE SEQUENCE seq_auth_role_id INCREMENT by 1 MINVALUE 1 START 1;
CREATE TABLE tb_auth_role(
	"id" INTEGER PRIMARY KEY DEFAULT nextval('seq_auth_role_id'),
	"name" VARCHAR(32),
	"description" VARCHAR(64),
	"time" TIMESTAMP  DEFAULT (now())
);
COMMENT ON TABLE "public"."tb_auth_role" IS 'RBAC角色信息表';
COMMENT ON COLUMN "tb_auth_role"."id" IS '角色id';
COMMENT ON COLUMN "tb_auth_role"."name" IS '角色名称';

-- 角色绑定权限
CREATE TABLE tb_auth_role_permission(
	"roleid" INTEGER,
	"permissionid" INTEGER,
	"time" TIMESTAMP DEFAULT (now()),
	PRIMARY KEY("roleid", "permissionid")
);
COMMENT ON TABLE "public"."tb_auth_role_permission" IS 'RBAC角色绑定权限';
COMMENT ON COLUMN "tb_auth_role_permission"."roleid" IS '角色id';
COMMENT ON COLUMN "tb_auth_role_permission"."permissionid" IS '权限id';

PostgreSQL End
*/

type (
	Role struct {
		Id          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	RoleController struct {
		controller.ControllerWebsite
		Ram *middleware.Ram
	}
)

func NewRoleController(db *sql.DB, ram *middleware.Ram) *RoleController {
	ctl := &RoleController{}
	ctl.DB = db
	ctl.Ram = ram
	return ctl
}

func (ctl *RoleController) GetIndex() (interface{}, error) {
	var size = ctl.GetQueryDefaultInt("size", 20)
	var page = ctl.GetQueryInt("page") * size
	return ctl.QueryRows("SELECT id,name,description FROM tb_auth_role ORDER BY id limit $1 OFFSET $2", size, page)
}

func (ctl *RoleController) GetList() (interface{}, error) {
	return ctl.QueryRows("SELECT * FROM tb_auth_role")
}

func (ctl *RoleController) GetSearch() (interface{}, error) {
	return nil, nil
}

func (ctl *RoleController) GetIdById() (interface{}, error) {
	return ctl.QueryJSON("SELECT id,name,description,time FROM tb_auth_role WHERE id=$1;", ctl.GetParam("id"))
}

func (ctl *RoleController) GetNameByName() (interface{}, error) {
	return ctl.QueryJSON("SELECT id,name,description,time FROM tb_auth_role WHERE name=$1;", ctl.GetParam("name"))
}

// GetUserIdById 方法根据角色id获取用户全部权限信息
func (ctl *RoleController) GetPermissionIdById() (interface{}, error) {
	return ctl.QueryRows("SELECT p.*,r.time FROM tb_auth_permission AS p JOIN tb_auth_role_permission AS r ON p.id=r.permissionid WHERE r.roleid=$1", ctl.GetParam("id"))
}

// GetUserNameByName 方法根据角色name获取用户全部权限信息
func (ctl *RoleController) GetPermissionNameByName() (interface{}, error) {
	return ctl.QueryRows("SELECT p.*,r.time FROM tb_auth_permission AS p JOIN tb_auth_role_permission AS r ON p.id=r.permissionid WHERE r.roleid=(SELECT id FROM tb_auth_role WHERE name=$1)", ctl.GetParam("name"))
}

// GetUserIdById 方法根据角色id获取用户全部用户信息
func (ctl *RoleController) GetUserIdById() (interface{}, error) {
	return ctl.QueryRows("SELECT u.name AS username,r.userid,r.time FROM tb_auth_user_role AS r JOIN tb_auth_user_info AS u ON r.userid=u.id WHERE roleid=$1", ctl.GetParam("id"))
}

// GetUserNameByName 方法根据角色name获取用户全部用户信息
func (ctl *RoleController) GetUserNameByName() (interface{}, error) {
	return ctl.QueryRows("SELECT u.name AS username,r.userid,r.time FROM tb_auth_user_role AS r JOIN tb_auth_user_info AS u ON r.userid=u.id WHERE roleid=(SELECT id FROM tb_auth_role WHERE name=$1)", ctl.GetParam("name"))
}

func (ctl *RoleController) PostIdById() (err error) {
	var role Role
	err = ctl.Bind(&role)
	if err != nil {
		return err
	}
	_, err = ctl.Exec("UPDATE tb_auth_role SET name=$1,description=$2,time=now() WHERE id=$3;", role.Name, role.Description, ctl.GetParam("id"))
	return
}
func (ctl *RoleController) PostNameByName() (err error) {
	var role Role
	err = ctl.Bind(&role)
	if err != nil {
		return err
	}
	_, err = ctl.Exec("UPDATE tb_auth_role SET name=$1,description=$2,time=now() WHERE name=$3;", role.Name, role.Description, ctl.GetParam("name"))
	return
}

func (ctl *RoleController) PutNew() (err error) {
	var role Role
	err = ctl.Bind(&role)
	if err != nil {
		return err
	}
	_, err = ctl.Exec("INSERT INTO tb_auth_role(name,description) VALUES($1,$2);", role.Name, role.Description)
	return
}
func (ctl *RoleController) DeleteIdById() (err error) {
	_, err = ctl.Exec("DELETE FROM tb_auth_role WHERE id=$1", ctl.GetParam("id"))
	return
}
func (ctl *RoleController) DeleteNameByName() (err error) {
	_, err = ctl.Exec("DELETE FROM tb_auth_role WHERE name=$1", ctl.GetParam("name"))
	return
}
