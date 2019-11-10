package auth

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/eudore/website/internal/controller"
	"github.com/eudore/website/internal/middleware"
)

/*
PostgreSQL Begin

-- PBAC策略信息表
CREATE SEQUENCE seq_auth_policy_id INCREMENT by 1 MINVALUE 1 START 1;
CREATE TABLE tb_auth_policy(
	"id" INTEGER PRIMARY KEY DEFAULT nextval('seq_auth_policy_id'),
	"name" VARCHAR(64),
	"description" VARCHAR(512),
	"policy" VARCHAR(4096),
	"time" TIMESTAMP  DEFAULT (now())
);
COMMENT ON TABLE "public"."tb_auth_policy" IS 'PBAC策略信息表';
COMMENT ON COLUMN "tb_auth_policy"."id" IS 'Polic ID';
COMMENT ON COLUMN "tb_auth_policy"."policy" IS '策略内容';

INSERT INTO "tb_auth_policy"("name", "description", "policy", "time") VALUES ('AdministratorAccess', '管理所有资源的权限', '{"version":"1","description":"22","statement":[{"effect":true,"action":["*"],"resource":["*"]}]}', '2019-09-14 09:15:12.129781');
INSERT INTO "tb_auth_policy"("name", "description", "policy", "time") VALUES ('NotePublicReadOnlyAccess', 'public文档只读权限', '{"version":"1","description":"public文档只读权限","statement":[{"effect":true,"action":["Get*"],"resource":["/note/content/public/*"]}]}', '2019-09-14 09:15:12.129781');
INSERT INTO "tb_auth_policy"("name", "description", "policy", "time") VALUES ('guest', 'guest使用', '{"version":"1","description":"全部文档只读权限","statement":[{"effect":true,"action":["auth:*:Get*","status:*:Get*"],"resource":["*"],"conditions":{"time":{"befor":"2020-12-31"},"method":["GET"],"browser":["Chrome/60+","Chromium/0-90","Firefox"]}}]}', '2019-09-28 08:12:35.656327');

PostgreSQL End
*/
type (
	Policy struct {
		Id          int       `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Policy      string    `json:"policy"`
		Time        time.Time `json:"time"`
	}
	PolicyController struct {
		controller.ControllerWebsite
		Ram *middleware.Ram
	}
)

func NewPolicyController(db *sql.DB, ram *middleware.Ram) *PolicyController {
	ctl := &PolicyController{}
	ctl.DB = db
	ctl.Ram = ram
	return ctl
}

// Release 方法用于刷新ram策略信息。
func (ctl *PolicyController) Release() error {
	// 如果修改策略信息成功，则刷新ram策略信息。
	if ctl.Response().Status() == 200 && (ctl.Method() == "POST" || ctl.Method() == "PUT" || ctl.Method() == "DELETE") {
		ctl.Ram.InitPolicyInfo(ctl.DB)
	}
	return ctl.ControllerWebsite.Release()
}

func (ctl *PolicyController) GetIndex() (interface{}, error) {
	return ctl.QueryPages("SELECT id,name,description,policy,time FROM tb_auth_policy ORDER BY id")
}

func (ctl *PolicyController) GetCount() interface{} {
	return map[string]int{"count": ctl.QueryCount("SELECT count(1) FROM tb_auth_policy")}
}

func (ctl *PolicyController) GetList() (interface{}, error) {
	return ctl.QueryRows("SELECT * FROM tb_auth_policy")
}

func (ctl *PolicyController) GetSearchByKey() (interface{}, error) {
	key := ctl.GetParam("key")
	return ctl.QueryPages("SELECT * FROM tb_auth_policy WHERE name ~ $1 OR description ~ $2 OR policy ~ $3 ORDER BY id", key, key, key)
}

func (ctl *PolicyController) GetIdById() (interface{}, error) {
	return ctl.QueryJSON("SELECT id,name,description,policy,time FROM tb_auth_policy WHERE id=$1;", ctl.GetParam("id"))
}

func (ctl *PolicyController) GetNameByName() (interface{}, error) {
	return ctl.QueryJSON("SELECT id,name,description,policy,time FROM tb_auth_policy WHERE name=$1;", ctl.GetParam("name"))
}

// GetUserIdById 方法根据权限id获取用户全部用户信息
func (ctl *PolicyController) GetUserIdById() (interface{}, error) {
	return ctl.QueryRows("SELECT u.name AS username,p.* FROM tb_auth_user_policy AS p JOIN tb_auth_user_info AS u ON p.userid=u.id WHERE policyid=$1", ctl.GetParam("id"))
}

// GetUserNameByName 方法根据权限name获取用户全部用户信息
func (ctl *PolicyController) GetUserNameByName() (interface{}, error) {
	return ctl.QueryRows("SELECT u.name AS username,p.* FROM tb_auth_user_policy AS p JOIN tb_auth_user_info AS u ON p.userid=u.id WHERE policyid=(SELECT id FROM tb_auth_policy WHERE name=$1)", ctl.GetParam("name"))
}

// PostIdById 方法根据id修改一个策略的信息。
func (ctl *PolicyController) PostIdById() (err error) {
	var policy Policy
	err = ctl.Bind(&policy)
	if err != nil {
		return err
	}
	_, err = ctl.Exec("UPDATE tb_auth_policy SET name=$1,description=$2,policy=$3,time=now() WHERE id=$4;", policy.Name, policy.Description, policy.Policy, ctl.GetParam("id"))
	return
}

// PostNameByName 方法根据名称修改一个策略的信息。
func (ctl *PolicyController) PostNameByName() (err error) {
	var policy Policy
	err = ctl.Bind(&policy)
	if err != nil {
		return err
	}
	_, err = ctl.Exec("UPDATE tb_auth_policy SET name=$1,description=$2,policy=$3,time=now() WHERE name=$4;", policy.Name, policy.Description, policy.Policy, ctl.GetParam("name"))
	return
}

// PutNew 方法新建一个策略信息，策略的policy毕竟是一个json。
func (ctl *PolicyController) PutNew() (err error) {
	var policy Policy
	err = ctl.Bind(&policy)
	if err != nil {
		return err
	}
	{
		data := make(map[string]interface{})
		err = json.Unmarshal([]byte(policy.Policy), &data)
		if err != nil {
			ctl.Fatal("policy body not is json.")
			return
		}
	}
	_, err = ctl.Exec("INSERT INTO tb_auth_policy(name,description,policy) VALUES($1,$2,$3);", policy.Name, policy.Description, policy.Policy)
	return
}

// DeleteIdById 方法根据id删除策略。
func (ctl *PolicyController) DeleteIdById() (err error) {
	_, err = ctl.Exec("DELETE FROM tb_auth_policy WHERE id=$1", ctl.GetParam("id"))
	return
}

// DeleteNameByName 方法根据名称删除策略。
func (ctl *PolicyController) DeleteNameByName() (err error) {
	_, err = ctl.Exec("DELETE FROM tb_auth_policy WHERE name=$1", ctl.GetParam("name"))
	return
}
