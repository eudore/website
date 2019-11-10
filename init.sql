\c website;
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

CREATE TABLE tb_note_consent(
	"path" VARCHAR(128),
	"userid" INTEGER DEFAULT 0,
	"createtime" TIMESTAMP DEFAULT (now()),
	PRIMARY KEY("path", "userid")
);


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



-- 用户信息表
CREATE SEQUENCE seq_auth_user_info_id INCREMENT by 1 MINVALUE 1 START 1;
CREATE TABLE tb_auth_user_info(
	"id" INTEGER PRIMARY KEY DEFAULT nextval('seq_auth_user_info_id'),
	"name" VARCHAR(32) NOT NULL,
	"status" INTEGER DEFAULT 0,
	"level" INTEGER DEFAULT 0,
	"mail" VARCHAR(48) DEFAULT "",
	"tel" VARCHAR(16) DEFAULT "",
	"icon" INTEGER DEFAULT 0,
	"lang" VARCHAR(16) DEFAULT "",
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



-- 绑定默认权限 任意用户可以获得用户图标



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





CREATE TABLE tb_auth_oauth2_source(
	"name" VARCHAR(16) PRIMARY KEY,
	"server" VARCHAR(32),
	"clientid" VARCHAR(80) NOT NULL,
	"clientsecret" VARCHAR(80) NOT NULL,
	PRIMARY KEY("name", "server")
);
COMMENT ON TABLE "public"."tb_auth_oauth2_source" IS 'oauth2 方式';
COMMENT ON COLUMN "tb_auth_oauth2_source"."name" IS '认证名称';
COMMENT ON COLUMN "tb_auth_oauth2_source"."name" IS '认证id';
COMMENT ON COLUMN "tb_auth_oauth2_source"."name" IS '认证secret';

CREATE TABLE tb_auth_oauth2_login(
	"source" VARCHAR(16),
	"originid" VARCHAR(80),
	"userid" INTEGER,
	"state" INTEGER,
	PRIMARY KEY("source", "originid")
);
COMMENT ON TABLE "public"."tb_auth_oauth2_login" IS 'oauth2 方式';
COMMENT ON COLUMN "tb_auth_oauth2_login"."originid" IS '认证源唯一id';
COMMENT ON COLUMN "tb_auth_oauth2_login"."userid" IS '认证后用户id';
COMMENT ON COLUMN "tb_auth_oauth2_login"."stats" IS '认证状态 0正常 1注册';

CREATE TABLE tb_auth_user_pass(
	"name" VARCHAR(32) PRIMARY KEY,
	"pass" VARCHAR(64),
	"salt" VARCHAR(10),
	"id" INTEGER
);
COMMENT ON TABLE "public"."tb_auth_user_pass" IS 'oauth2 用户登录密码';
COMMENT ON COLUMN "tb_auth_user_pass"."name" IS '登录用户';
COMMENT ON COLUMN "tb_auth_user_pass"."pass" IS '登录密码Hash';
COMMENT ON COLUMN "tb_auth_user_pass"."salt" IS '私钥';
COMMENT ON COLUMN "tb_auth_user_pass"."id" IS '认证ID';



CREATE TABLE tb_chat_message(
	"sendid" INTEGER,
	"receid" INTEGER,
	"status" INTEGER DEFAULT 0,
	"message" TEXT,
	"time" TIMESTAMP  DEFAULT (now())
);

-- task执行日志
CREATE TABLE tb_task_logger(
	"status" INTEGER,
	"params" VARCHAR(512),
	"message" TEXT,
	"starttime" TIMESTAMP,
	"endtime" TIMESTAMP,

	"eventid" INTEGER,
	"executorid" INTEGER
);


-- task触发器
CREATE SEQUENCE seq_task_trigger_id INCREMENT by 1 MINVALUE 1 START 1;
CREATE TABLE tb_task_trigger(
	"id" INTEGER PRIMARY KEY DEFAULT nextval('seq_task_trigger_id'),
	"name" VARCHAR(64),
	"description" VARCHAR(512),
	"event" VARCHAR(64),
	"params" VARCHAR(512),
	"schedule" VARCHAR(64),
	"executorid" INTEGER,
	"time" TIMESTAMP  DEFAULT (now())
);
COMMENT ON TABLE "public"."tb_task_trigger" IS 'task触发器';
COMMENT ON COLUMN "tb_task_trigger"."id" IS '触发器id';
COMMENT ON COLUMN "tb_task_trigger"."name" IS '触发器名称';
COMMENT ON COLUMN "tb_task_trigger"."event" IS '触发器事件';
COMMENT ON COLUMN "tb_task_trigger"."executorid" IS '触发器绑定的执行executor';


CREATE SEQUENCE seq_task_executor_id INCREMENT by 1 MINVALUE 1 START 1;
CREATE TABLE tb_task_executor(
	"id" INTEGER PRIMARY KEY DEFAULT nextval('seq_task_executor_id'),
	"name" VARCHAR(64),
	"description" VARCHAR(512),
	"type" VARCHAR(64),
	"config" VARCHAR(512),
	"time" TIMESTAMP  DEFAULT (now())
);


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

GRANT ALL PRIVILEGES ON DATABASE website to website;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO website;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO website;

INSERT INTO "tb_auth_policy"("name", "description", "policy", "time") VALUES ('AdministratorAccess', '管理所有资源的权限', '{"version":"1","description":"22","statement":[{"effect":true,"action":["*"],"resource":["*"]}]}', '2019-09-14 09:15:12.129781');
INSERT INTO "tb_auth_policy"("name", "description", "policy", "time") VALUES ('NotePublicReadOnlyAccess', 'public文档只读权限', '{"version":"1","description":"public文档只读权限","statement":[{"effect":true,"action":["Get*"],"resource":["/note/content/public/*"]}]}', '2019-09-14 09:15:12.129781');
INSERT INTO "tb_auth_policy"("name", "description", "policy", "time") VALUES ('guest', 'guest使用', '{"version":"1","description":"全部文档只读权限","statement":[{"effect":true,"action":["auth:*:Get*","status:*:Get*"],"resource":["*"],"conditions":{"time":{"befor":"2020-12-31"},"method":["GET"],"browser":["Chrome/60+","Chromium/0-90","Firefox"]}}]}', '2019-09-28 08:12:35.656327');
INSERT INTO "public"."tb_auth_user_info"("name", "status", "level", "mail", "tel", "icon", "loginip", "logintime", "sigintime") VALUES ('root', 1, 0, 'eudore@eudore.cn', NULL, 0, 0, '2019-02-07 22:57:59', '2019-02-07 09:03:18.124699');
INSERT INTO "public"."tb_auth_user_info"("name", "status", "level", "mail", "tel", "icon", "loginip", "logintime", "sigintime") VALUES ('guest', 0, 0, 'guest@eudore.cn', '', 0, 0, '2019-01-01 00:00:00', '2019-04-27 07:41:38.974911');
INSERT INTO "public"."tb_auth_permission"("name", "description") VALUES ('auth:User:GetIconNameByName', '获取用户图标权限');
INSERT INTO "public"."tb_task_trigger"("name", "description", "event", "params", "executorid") VALUES ('http-test', ' ', 'http', '{"method":"GET","url":"9","route":"/:num|isnum", "async": true}', 9);
INSERT INTO "tb_auth_user_permission"("userid", "permissionid", "effect") VALUES (0, (SELECT id FROM "tb_auth_permission" WHERE "name"='auth:User:GetIconNameByName'), 't');
INSERT INTO "tb_auth_user_policy"("userid", "policyid", "index") VALUES((SELECT id FROM tb_auth_user_info WHERE "name"='root'), (SELECT id FROM tb_auth_policy WHERE "name"='AdministratorAccess'), 100);
INSERT INTO "tb_auth_user_policy"("userid", "policyid", "index") VALUES((SELECT id FROM tb_auth_user_info WHERE "name"='guest'), (SELECT id FROM tb_auth_policy WHERE "name"='guest'), 100);
INSERT INTO "tb_auth_user_pass"("name", "pass", "salt", "id") VALUES ('root', 'd1fbb03f8a717f3d9cd2cf3e59d39fd1a227b7fc5ee2cea831b4050a1ae4dbe4', '0123456789', (SELECT id FROM tb_auth_user_info WHERE "name"='root'));
