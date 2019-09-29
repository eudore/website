package auth

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	// "github.com/eudore/eudore"
	"github.com/eudore/website/internal/controller"
	"github.com/eudore/website/internal/middleware"
)

/*
PostgreSQL Begin

-- 用户信息表
CREATE SEQUENCE seq_auth_user_info_id INCREMENT by 1 MINVALUE 1 START 1;
CREATE TABLE tb_auth_user_info(
	"id" INTEGER PRIMARY KEY DEFAULT nextval('seq_auth_user_info_id'),
	"name" VARCHAR(32) NOT NULL,
	"status" INTEGER DEFAULT 0,
	"level" INTEGER DEFAULT 0,
	"mail" VARCHAR(48),
	"tel" VARCHAR(16),
	"icon" INTEGER DEFAULT 0,
	"loginip" INTEGER DEFAULT 0,
	"logintime" TIMESTAMP,
	"sigintime" TIMESTAMP DEFAULT (now())
);
COMMENT ON TABLE "public"."tb_auth_user_info" IS '用户信息表';
COMMENT ON COLUMN "tb_auth_user_info"."icon" IS '图标ID，0使用gravatar';
COMMENT ON COLUMN "tb_auth_user_info"."loginip" IS '登录IP';
COMMENT ON COLUMN "tb_auth_user_info"."logintime" IS '上次登录时间';
COMMENT ON COLUMN "tb_auth_user_info"."sigintime" IS '注册时间';


-- 用户图标
CREATE SEQUENCE seq_auth_user_icon_id INCREMENT by 1 MINVALUE 1 START 1;
CREATE TABLE tb_auth_user_icon(
	"id" INTEGER PRIMARY KEY DEFAULT nextval('seq_auth_user_icon_id'),
	"user" INTEGER NOT NULL,
	"data" bytea
);
COMMENT ON TABLE "public"."tb_auth_user_icon" IS '用户图标';
COMMENT ON COLUMN "tb_auth_user_icon"."user" IS '用户ID';
COMMENT ON COLUMN "tb_auth_user_icon"."data" IS '图标二进制文件数据';



-- 用户绑定权限列表
CREATE TABLE tb_auth_user_permission(
	"userid" INTEGER,
	"permissionid" INTEGER,
	"effect" bool,
	"time" TIMESTAMP DEFAULT (now()),
	PRIMARY KEY("userid", "permissionid")
);
COMMENT ON TABLE "public"."tb_auth_user_permission" IS 'ACL用户绑定权限列表';
COMMENT ON COLUMN "tb_auth_user_permission"."userid" IS '用户id';
COMMENT ON COLUMN "tb_auth_user_permission"."permissionid" IS '权限id';

-- 用户绑定角色关系
CREATE TABLE tb_auth_user_role(
	"userid" INTEGER,
	"roleid" INTEGER,
	"time" TIMESTAMP  DEFAULT (now()),
	PRIMARY KEY("userid", "roleid")
);
COMMENT ON TABLE "public"."tb_auth_user_role" IS 'RBAC用户绑定角色关系';
COMMENT ON COLUMN "tb_auth_user_role"."userid" IS '用户id';
COMMENT ON COLUMN "tb_auth_user_role"."roleid" IS '角色id';

-- 用户绑定策略
CREATE TABLE tb_auth_user_policy(
	"userid" INTEGER,
	"policyid" INTEGER,
	"index" INTEGER DEFAULT 0,
	"time" TIMESTAMP DEFAULT (now()),
	PRIMARY KEY("userid", "policyid")
);
COMMENT ON TABLE "public"."tb_auth_user_policy" IS 'PBAC用户绑定策略';
COMMENT ON COLUMN "tb_auth_user_policy"."userid" IS 'User ID';
COMMENT ON COLUMN "tb_auth_user_policy"."policyid" IS 'Polic ID';
COMMENT ON COLUMN "tb_auth_user_policy"."index" IS '策略优先级';


INSERT INTO "public"."tb_auth_user_info"("name", "status", "level", "mail", "tel", "icon", "loginip", "logintime", "sigintime") VALUES ('root', 1, 0, 'eudore@eudore.cn', NULL, 0, 0, '2019-02-07 22:57:59', '2019-02-07 09:03:18.124699');
INSERT INTO "public"."tb_auth_user_info"("name", "status", "level", "mail", "tel", "icon", "loginip", "logintime", "sigintime") VALUES ('guest', 0, 0, 'guest@eudore.cn', '', 0, 0, '2019-01-01 00:00:00', '2019-04-27 07:41:38.974911');

-- 绑定默认权限 任意用户可以获得用户图标
INSERT INTO "tb_auth_user_permission"("userid", "permissionid", "effect") VALUES (0, (SELECT id FROM "tb_auth_permission" WHERE "name"='auth:User:GetIconNameByName'), 't');

INSERT INTO "tb_auth_user_policy"("userid", "policyid", "index") VALUES((SELECT id FROM tb_auth_user_info WHERE "name"='root'), (SELECT id FROM tb_auth_policy WHERE "name"='AdministratorAccess'), 100);
INSERT INTO "tb_auth_user_policy"("userid", "policyid", "index") VALUES((SELECT id FROM tb_auth_user_info WHERE "name"='guest'), (SELECT id FROM tb_auth_policy WHERE "name"='guest'), 100);

PostgreSQL End
*/
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Status    int       `json:"status"`
	Level     int       `json:"level"`
	Mail      string    `json:"mail"`
	Tel       string    `json:"tel"`
	Icon      int       `json:"icon"`
	Loginip   int64     `json:"loginip"`
	Logintime time.Time `json:"logintime"`
	Sigintime time.Time `json:"sigintime"`
}

type UserController struct {
	controller.ControllerWebsite
	Ram *middleware.Ram
}

func NewUserController(db *sql.DB, ram *middleware.Ram) *UserController {
	return &UserController{
		ControllerWebsite: controller.ControllerWebsite{
			DB: db,
		},
		Ram: ram,
	}
}

// Release 方法刷新用户绑定ram资源信息。
func (ctl *UserController) Release() error {
	// 如果修改策略信息超过，则刷新ram策略信息。
	if ctl.Response().Status() == 200 && ctl.GetParam("bind") != "" {
		switch ctl.GetParam("bind") {
		case "permission":
			ctl.Ram.InitUserBindPermission(ctl.DB)
		case "role":
			ctl.Ram.InitUserBindRole(ctl.DB)
		case "policy":
			ctl.Ram.InitUserBindPolicy(ctl.DB)
		}
	}
	return nil
}

// GetRouteParam 方法额外添加bind路由参数信息，用于Release刷新ram。
func (ctl *UserController) GetRouteParam(pkg, name, method string) string {
	params := ctl.ControllerWebsite.GetRouteParam(pkg, name, method)
	// 添加bind参数
	if strings.HasPrefix(method, "PutBind") || strings.HasPrefix(method, "DeleteBind") {
		method = strings.TrimPrefix(method, "PutBind")
		method = strings.TrimPrefix(method, "DeleteBind")
		switch method[0:4] {
		case "Perm":
			method = "permission"
		case "Poli":
			method = "policy"
		case "Role":
			method = "role"
		}
		params += fmt.Sprintf(" bind=%s", method)
	}
	return params
}

func (ctl *UserController) GetIndex() (interface{}, error) {
	return ctl.QueryPages("SELECT id,name,status,mail,loginip,logintime FROM tb_auth_user_info ORDER BY id")
}

func (ctl *UserController) GetCount() interface{} {
	return map[string]int{"count": ctl.QueryCount("SELECT count(1) FROM tb_auth_user_info")}
}

/*
用户对象
*/

func (ctl *UserController) GetInfoIdById() (interface{}, error) {
	return ctl.QueryJSON("SELECT id,name,status,level,COALESCE(mail, '') AS mail,COALESCE(tel, '') AS tel,loginip,logintime,sigintime FROM tb_auth_user_info WHERE id=$1;", ctl.GetParam("id"))
}
func (ctl *UserController) GetInfoNameByName() (interface{}, error) {
	return ctl.QueryJSON("SELECT id,name,status,level,COALESCE(mail, '') AS mail,COALESCE(tel, '') AS tel,loginip,logintime,sigintime FROM tb_auth_user_info WHERE name=$1;", ctl.GetParam("name"))
}

// PutNew 方法创建一个用户.
func (ctl *UserController) PutNew() {}

func (ctl *UserController) DeleteNameByName() (err error) {
	_, err = ctl.Exec("UPDATE tb_auth_user_info SET status=1 WHERE name=$1", ctl.GetParam("name"))
	return
}

func (ctl *UserController) DeleteIdById() (err error) {
	_, err = ctl.Exec("UPDATE tb_auth_user_info SET status=1 WHERE id=$1", ctl.GetParam("id"))
	return
}

func (ctl *UserController) PostInfoById()    {}
func (ctl *UserController) PostSettingById() {}

/*
用户图标
*/

func (ctl *UserController) GetIconIdById() {
	ctl.WriteFile("static/favicon.ico")
}
func (ctl *UserController) GetIconNameByName() {
	ctl.WriteFile("static/favicon.ico")
	/*	path := iconTmp + user.Name
		if PathExist(path) {
			return nil
		}
			if user.Icon == 0 {
				// get gravatar file
				hash := md5.Sum([]byte(user.Mail))
				resp, err := eudore.NewRequest("GET", fmt.Sprintf("https://www.gravatar.com/avatar/%x?s=%d&d=identicon", hash, 64), nil).Do()
				if err != nil {
					return err
				}

				newFile, err := os.Create(path)
				if err != nil {
					return err
				}

				defer newFile.Close()
				defer resp.Close()
				_, err = io.Copy(newFile, resp)
				return err
			} else {

			}*/
}
func (ctl *UserController) PostIconById() {}

/*
用户权限
*/

// GetPermissionIdById 方法根据用户id获取用户全部权限信息
func (ctl *UserController) GetPermissionIdById() (interface{}, error) {
	return ctl.QueryRows("SELECT * FROM tb_auth_user_permission AS u JOIN tb_auth_permission AS p ON u.permissionid = p.id WHERE userid=$1", ctl.GetParam("id"))
}

// GetPermissionIdById 方法根据用户name获取用户全部权限信息
func (ctl *UserController) GetPermissionNameByName() (interface{}, error) {
	return ctl.QueryRows("SELECT * FROM tb_auth_user_permission AS u JOIN tb_auth_permission AS p ON u.permissionid = p.id WHERE userid=(SELECT id FROM tb_auth_user_info WHERE name=$1)", ctl.GetParam("name"))
}

// PutBindPermissionById 方法给用户批量绑定多条权限。
//
// body: [{"id":4,"effect":"deny"},{"id":6,"effect":"allow"}]
func (ctl *UserController) PutBindPermissionById() error {
	err := ctl.ExecBodyWithJSON(fmt.Sprintf("INSERT INTO tb_auth_user_permission(userid,permissionid,effect) VALUES(%d,$1,$2='allow');", ctl.GetParamInt("id")), "id", "effect")
	return err
}

// PutBindPermissionDenyByUidById 方法给用户绑定一个拒绝权限
func (ctl *UserController) PutBindPermissionDenyByUidById() error {
	_, err := ctl.Exec("INSERT INTO tb_auth_user_permission(userid,permissionid, effect) VALUES($1,$2,false)", ctl.GetParam("uid"), ctl.GetParam("id"))
	return err
}

// PutBindPermissionAllowByUidById 方法给用户绑定一个允许权限
func (ctl *UserController) PutBindPermissionAllowByUidById() (err error) {
	_, err = ctl.Exec("INSERT INTO tb_auth_user_permission(userid,permissionid, effect) VALUES($1,$2,true)", ctl.GetParam("uid"), ctl.GetParam("id"))
	return
}

// DeleteBindPermissionById 方法批量移除用户权限。
//
// body: [{"id":1},{"id":2}]
func (ctl *UserController) DeleteBindPermissionById() error {
	err := ctl.ExecBodyWithJSON(fmt.Sprintf("DELETE FROM tb_auth_user_permission WHERE userid=%d AND permissionid=$1", ctl.GetParamInt("id")), "id")
	return err
}

// DeleteBindPermissionByUidById 方法移除用户的一个权限
func (ctl *UserController) DeleteBindPermissionByUidById() (err error) {
	_, err = ctl.Exec("DELETE FROM tb_auth_user_permission WHERE userid=$1 AND permissionid=$2", ctl.GetParam("uid"), ctl.GetParam("id"))
	return
}

/*
用户角色
*/

func (ctl *UserController) GetRoleIdById() (interface{}, error) {
	return ctl.QueryRows("SELECT * FROM tb_auth_user_role AS u JOIN tb_auth_role AS r ON u.roleid = r.id WHERE userid=$1", ctl.GetParam("id"))
}
func (ctl *UserController) GetRoleNameByName() (interface{}, error) {
	return ctl.QueryRows("SELECT * FROM tb_auth_user_role AS u JOIN tb_auth_role AS r ON u.roleid = r.id WHERE userid=(SELECT id FROM tb_auth_user_info WHERE name=$1)", ctl.GetParam("name"))
}

// PutBindRoleById 方法给用户批量绑定多个角(jue)色。
//
// body: [{"id":4},{"id":6}]
func (ctl *UserController) PutBindRoleById() error {
	err := ctl.ExecBodyWithJSON(fmt.Sprintf("INSERT INTO tb_auth_user_role(userid,roleid) VALUES(%d,$1);", ctl.GetParamInt("id")), "id")
	return err
}

func (ctl *UserController) PutBindRoleByUidById() (err error) {
	_, err = ctl.Exec("INSERT INTO tb_auth_user_role(userid,roleid) VALUES($1,$2)", ctl.GetParam("uid"), ctl.GetParam("id"))
	return
}

// DeleteBindRoleById 方法给用户删除绑定多个角(jue)色。
//
// body: [{"id":4},{"id":6}]
func (ctl *UserController) DeleteBindRoleById() error {
	err := ctl.ExecBodyWithJSON(fmt.Sprintf("DELETE FROM tb_auth_user_role WHERE userid=%s AND roleid=$1", ctl.GetParamInt("id")), "id")
	return err
}

func (ctl *UserController) DeleteBindRoleByUidById() (err error) {
	_, err = ctl.Exec("DELETE FROM tb_auth_user_role WHERE userid=$1 AND roleid=$2", ctl.GetParam("uid"), ctl.GetParam("id"))
	return
}

/*
用户策略
*/

// GetPolicyNameByName  方法根据策略id获得绑定的用户。
func (ctl *UserController) GetPolicyIdById() (interface{}, error) {
	return ctl.QueryRows("SELECT * FROM tb_auth_user_policy AS u JOIN tb_auth_policy AS p ON u.policyid = p.id WHERE userid=$1", ctl.GetParam("id"))
}

// GetPolicyNameByName  方法根据策略name获得绑定的用户。
func (ctl *UserController) GetPolicyNameByName() (interface{}, error) {
	return ctl.QueryRows("SELECT * FROM tb_auth_user_policy AS u JOIN tb_auth_policy AS p ON u.policyid = p.id WHERE userid=(SELECT id FROM tb_auth_user_info WHERE name=$1)", ctl.GetParam("name"))
}

// PutBindPolicyById 方法给用户批量绑定多条策略。
//
// body: [{"id":4},{"id":6}]
func (ctl *UserController) PutBindPolicyById() error {
	err := ctl.ExecBodyWithJSON(fmt.Sprintf("INSERT INTO tb_auth_user_policy(userid,policyid) VALUES(%d,$1);", ctl.GetParamInt("id")), "id")
	return err
}

// PutBindPolicyByUidById 方法给指定用户绑定指定权限。
func (ctl *UserController) PutBindPolicyByUidById() (err error) {
	_, err = ctl.Exec("INSERT INTO tb_auth_user_policy(userid,policyid) VALUES($1,$2)", ctl.GetParam("uid"), ctl.GetParam("id"))
	return
}

// DeleteBindPolicyById 方法给用户批量删除多条策略。
//
// body: [{"id":4},{"id":6}]
func (ctl *UserController) DeleteBindPolicyById() error {
	err := ctl.ExecBodyWithJSON(fmt.Sprintf("DELETE FROM tb_auth_user_policy WHERE userid=%s AND policyid=$1", ctl.GetParamInt("id")), "id")
	return err
}

// DeleteBindPolicyByUidById 方法给指定用户删除指定权限。
func (ctl *UserController) DeleteBindPolicyByUidById() (err error) {
	_, err = ctl.Exec("DELETE FROM tb_auth_user_policy WHERE userid=$1 AND policyid=$2", ctl.GetParam("uid"), ctl.GetParam("id"))
	return
}

func PathExist(_path string) bool {
	_, err := os.Stat(_path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
