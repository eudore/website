package auth

import (
	"github.com/eudore/website/framework"
	"time"
)

/*
PostgreSQL Begin


CREATE SEQUENCE seq_auth_user_info_id INCREMENT by 1 MINVALUE 1 START 1;
CREATE TABLE tb_auth_user_info(
	"id" INTEGER PRIMARY KEY DEFAULT nextval('seq_auth_user_info_id'),
	"name" VARCHAR(32) NOT NULL,
	"status" INTEGER DEFAULT 0,
	"level" INTEGER DEFAULT 0,
	"mail" VARCHAR(48) DEFAULT "",
	"tel" VARCHAR(16) DEFAULT "",
	"icon" bytea DEFAULT "";,
	"lang" VARCHAR(16) DEFAULT "",
	"loginip" VARCHAR(16) DEFAULT "",
	"logintime" TIMESTAMP,
	"sigintime" TIMESTAMP DEFAULT (now())
);
COMMENT ON TABLE "public"."tb_auth_user_info" IS '用户信息表';
COMMENT ON COLUMN "tb_auth_user_info"."icon" IS '图标二进制，为空使用gravatar';
COMMENT ON COLUMN "tb_auth_user_info"."loginip" IS '登录IP';
COMMENT ON COLUMN "tb_auth_user_info"."logintime" IS '上次登录时间';
COMMENT ON COLUMN "tb_auth_user_info"."sigintime" IS '注册时间';

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

CREATE VIEW vi_auth_user_permission AS
	SELECT userid, permissionid, u."name" AS username, P."name" AS permissionname, up.effect, P.description, up.TIME AS granttime
	FROM tb_auth_user_info u JOIN tb_auth_user_permission up ON u.ID = up.userid JOIN tb_auth_permission P ON P.ID = up.permissionid

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

CREATE VIEW vi_auth_user_role AS
	SELECT userid, roleid, u."name" AS username, r."name" AS rolename, r.description, ur.TIME AS granttime
	FROM tb_auth_user_info u JOIN tb_auth_user_role ur ON u.ID = ur.userid JOIN tb_auth_role r ON r.ID = ur.roleid

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

CREATE VIEW vi_auth_role_permission AS
	SELECT p.id as "permissionid", p.name as "permissionname", r.id AS "roleid", r.name AS "rolename",rp.time AS "granttime"
	FROM tb_auth_permission AS P JOIN tb_auth_role_permission rp ON P.ID = rp.permissionid JOIN tb_auth_role r ON r.ID = rp.roleid



-- PBAC策略信息表
CREATE SEQUENCE seq_auth_policy_id INCREMENT by 1 MINVALUE 1 START 1;
CREATE TABLE tb_auth_policy(
	"id" INTEGER PRIMARY KEY DEFAULT nextval('seq_auth_policy_id'),
	"name" VARCHAR(64),
	"version" VARCHAR(32) DEFAULT('v1'),
	"description" VARCHAR(512),
	"policy" VARCHAR(4096),
	"time" TIMESTAMP DEFAULT (now())
);
COMMENT ON TABLE "public"."tb_auth_policy" IS 'PBAC策略信息表';
COMMENT ON COLUMN "tb_auth_policy"."id" IS 'Polic ID';
COMMENT ON COLUMN "tb_auth_policy"."policy" IS '策略内容';

INSERT INTO "tb_auth_policy"("name", "description", "policy", "time") VALUES ('AdministratorAccess', '管理所有资源的权限', '{"version":"1","description":"22","statement":[{"effect":true,"action":["*"],"resource":["*"]}]}', '2019-09-14 09:15:12.129781');
INSERT INTO "tb_auth_policy"("name", "description", "policy", "time") VALUES ('NotePublicReadOnlyAccess', 'public文档只读权限', '{"version":"1","description":"public文档只读权限","statement":[{"effect":true,"action":["Get*"],"resource":["/note/content/public/*"]}]}', '2019-09-14 09:15:12.129781');
INSERT INTO "tb_auth_policy"("name", "description", "policy", "time") VALUES ('guest', 'guest使用', '{"version":"1","description":"全部文档只读权限","statement":[{"effect":true,"action":["auth:*:Get*","status:*:Get*"],"resource":["*"],"conditions":{"time":{"befor":"2020-12-31"},"method":["GET"],"browser":["Chrome/60+","Chromium/0-90","Firefox"]}}]}', '2019-09-28 08:12:35.656327');

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



CREATE VIEW vi_auth_user_policy AS
	SELECT userid, policyid, u."name" AS username, P."name" AS policyname, P.description, up.TIME AS granttime
	FROM tb_auth_user_info u JOIN tb_auth_user_policy up ON u.ID = up.userid JOIN tb_auth_policy P ON P.ID = up.policyid

-- 策略版本
CREATE TABLE tb_auth_policy_version(
	"id" INTEGER,
	"version" VARCHAR(32),
	"policy" VARCHAR(4096),
	"time" TIMESTAMP  DEFAULT (now())
	PRIMARY KEY("id", "version")
)


CREATE TABLE tb_auth_access_token(
	"userid" INTEGER PRIMARY KEY,
	"token" VARCHAR(32),
	"expires" TIMESTAMP,
	"createtime" TIMESTAMP DEFAULT (now())
);

CREATE TABLE tb_auth_access_key(
	"userid" INTEGER PRIMARY KEY,
	"accesskey" VARCHAR(32),
	"accesssecrect" VARCHAR(32),
	"expires" TIMESTAMP,
	"createtime" TIMESTAMP DEFAULT (now())
);

PostgreSQL End
*/

type User struct {
	ID        int       `alias:"id" json:"id"`
	Name      string    `alias:"name" json:"name"`
	Status    int       `alias:"status" json:"status"`
	Level     int       `alias:"level" json:"level"`
	Mail      string    `alias:"mail" json:"mail"`
	Tel       string    `alias:"tel" json:"tel"`
	Lang      string    `alias:"lang" json:"lang"`
	Loginip   int64     `alias:"loginip" json:"loginip"`
	Logintime time.Time `alias:"logintime" json:"logintime"`
	Sigintime time.Time `alias:"sigintime" json:"sigintime"`
}

func Init(app *framework.App) error {
	return app.Group("/api/v1/auth").AddController(
		framework.NewTableController("auth", "user", "tb_auth_user_info", new(User), app.DB),
		framework.NewTableController("auth", "permission", "tb_auth_permission", nil, app.DB),
		framework.NewTableController("auth", "role", "tb_auth_role", nil, app.DB),
		framework.NewTableController("auth", "policy", "tb_auth_policy", nil, app.DB),
		framework.NewTableController("auth", "policyversion", "tb_auth_policy_version", nil, app.DB),
		framework.NewTableController("auth", "userpermission", "tb_auth_user_permission", nil, app.DB),
		framework.NewTableController("auth", "userrole", "tb_auth_user_role", nil, app.DB),
		framework.NewTableController("auth", "userpolicy", "tb_auth_user_policy", nil, app.DB),
		framework.NewTableController("auth", "rolepermission", "tb_auth_role_permission", nil, app.DB),
		framework.NewViewController("auth", "userpermission", "vi_auth_user_permission", nil, app.DB),
		framework.NewViewController("auth", "userrole", "vi_auth_user_role", nil, app.DB),
		framework.NewViewController("auth", "userpolicy", "vi_auth_user_policy", nil, app.DB),
		framework.NewViewController("auth", "rolepermission", "vi_auth_role_permission", nil, app.DB),
		NewLoginController(app),
		NewIconController(app),
	)
}
