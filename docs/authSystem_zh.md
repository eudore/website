本文大致记录了[eudore-website](https://github.com/eudore/website)认证鉴权体系的实现，实现了acl、rbac、pbac鉴权和ak、token、bearer认证，完整细节请查看[源码](https://github.com/eudore/website)。

[在线demo](https://www.eudore.cn/auth/),用户密码均为guest。

# 认证设计

eudore-website使用ak、token、bearer认证三种综合认证，原理通过web请求中间件使用请求信息获得用户信息，保存到请求上下文中然后供后续使用。

## bearer认证

**bearer认证原理是利用jwt非对称签名防止数据篡改。**

最初就初始化jwt解析对象，然后处理请求Authorization Header,解析出jwt的数据，从中提取到userid和username信息，然后设置请求上下文的参数中。

[源码](https://github.com/eudore/website/blob/master/internal/middleware/user.go#L48-L55)

```golang
func(ctx eudore.Context) {
	data, err := jwtParse.ParseBearer(ctx.GetHeader(eudore.HeaderAuthorization))
	if err == nil {
		ctx.SetParam("UID", eudore.GetString(data["userid"]))
		ctx.SetParam("UNAME", eudore.GetString(data["name"]))
		return
	}
	...
}
```

然后客户端请求添加Authorization Header.

例如基于mithriljs封装ajax添加Header：

```js
if(typeof m.request !== 'undefined') {
	var oldrequest = m.request
	m.request = function(args) {
		// add header
		if(!("headers" in args)) {
			args["headers"] = {}
		}
		if(Base.lang != "") {
			args["headers"]["Accept-Language"] = Base.lang
		}
		if(Base.bearer != "") {
			args["headers"]["Authorization"] = Base.bearer
		}
		if(requestid!="") {
			args["headers"]["X-Parent-Id"] = requestid
		}
		
		return oldrequest(args)
	}
}
```

具有一个全局遍历Base保存用户相关信息，例如Base.bearer,在使用m.request方法时，就自动给参数添加bearer信息。

curl直接-H指定header即可。

## token认证

**token认证原理使用token加载到对应的用户信息**

token供api使用。

创建数据表tb_auth_access_token，里面保存token对应的用户信息。

```sql
CREATE TABLE tb_auth_access_token(
	"userid" INTEGER PRIMARY KEY,
	"token" VARCHAR(32),
	"expires" TIMESTAMP,
	"createtime" TIMESTAMP DEFAULT (now())
);
```

从请求中提取到token参数，然后数据库查询tb_auth_access_token表，找到用户信息，设置到请求上下文中。

[源码](https://github.com/eudore/website/blob/master/internal/middleware/user.go#L57-L68)

```golang
stmtQueryAccessToken, err := db.Prepare("SELECT userid,(SELECT name FROM tb_auth_user_info WHERE id = userid) FROM tb_auth_access_token WHERE token=$1 and expires > now()")
func(ctx eudore.Context) {
	...

	token := ctx.GetQuery("token")
	if token != "" {
		var userid string
		var username string
		err := stmtQueryAccessToken.QueryRow(token).Scan(&userid, &username)
		if err == nil {
			ctx.SetParam("UID", userid)
			ctx.SetParam("UNAME", username)
			return
		}
		ctx.Error(err)
	}
	...
}
```

## ak认证

**ak认证的原理例如非对称加密实现，用户有效校验**

ak和token一样用于api使用，但是ak更加复杂和安全。

accesskey表明是那个ak，accesssecrect是签名使用的私钥，然后客户端和服务端使用accesssecrect签名一个数据得到签名结果signature，如果signature相同就是表示accesssecrect相同，那么用户使用的ak就是有效的。

ak认证创建tb_auth_access_key表，保存ak和用户信息和token表相识。

```sql
CREATE TABLE tb_auth_access_key(
	"userid" INTEGER PRIMARY KEY,
	"accesskey" VARCHAR(32),
	"accesssecrect" VARCHAR(32),
	"expires" TIMESTAMP,
	"createtime" TIMESTAMP DEFAULT (now())
);
```

ak认证先提取到accesskey、signature、expires三个参数，用于ak认证使用，accesskey对应ak记录、signature是ak签名结果、expires是签名过期时间。

先检查下有效时间是否有效，且有效时间不大于60分钟。

然后数据库查询一下accesskey对应的accesssecrect和用户数据。

再计算一下签名结果，如果结果和signature一样那么就是通过，然后设置用户数据。

当前签名格式是accesskey-expires,过于简单，但是也可以用。

[源码](https://github.com/eudore/website/blob/master/internal/middleware/user.go#L70-L99)

```golang
stmtQueryAccessKey, err := db.Prepare("SELECT userid,(SELECT name FROM tb_auth_user_info WHERE id = userid),accesssecrect FROM tb_auth_access_key WHERE accesskey=$1 and expires > $2")
func(ctx eudore.Context) {
	...
	key, signature, expires := ctx.GetQuery("accesskey"), ctx.GetQuery("signature"), ctx.GetQuery("expires")
	if key != "" && signature != "" && expires != "" {
		tunix, err := strconv.ParseInt(expires, 10, 64)
		if err != nil {
			ctx.Error(err)
			return
		}
		ttime := time.Unix(tunix, 0)
		if ttime.After(time.Now().Add(60 * time.Minute)) {
			ctx.Errorf("accesskey expires is to long, max 60 min")
			return
		}
		//
		var userid, username, scerect string
		err = stmtQueryAccessKey.QueryRow(key, ttime).Scan(&userid, &username, &scerect)
		if err != nil {
			ctx.Error(err)
			return
		}

		h := hmac.New(sha1.New, []byte(scerect))
		fmt.Fprintf(h, "%s-%s", key, expires)
		if signature != base64.StdEncoding.EncodeToString(h.Sum(nil)) {
			ctx.Errorf("signature is invalid")
			return
		}
		ctx.SetParam("UID", userid)
		ctx.SetParam("UNAME", username)
	}
}
```

# eudore RAM 鉴权设计

eudore-website使用acl、rbac、pbac三种复合鉴权设计，按顺序依次处理，某个对象可以处理就返回结果。

## RAM

eudore定义了一个ram接口,ram接口会传递用户id、用户行为信息给ram鉴权使用，然后ram对象返回处理结果和是否处理。

[源码定义](https://github.com/eudore/eudore/blob/master/middleware/ram/ram.go#L22-L27)

```golang
// RamHandler 定义Ram处理接口
type RamHandler interface {
	RamHandle(int, string, eudore.Context) (bool, bool)
	// return1 验证结果 return2 是否验证
}
```

RamHttp对象会处理http相关内容，获取到用户id和行为传递给多Ram对象依次处理，然后根据Ram结果处理，一般RAM对象如果处理了请求会设置ram参数为处理者，例如acl处理的请求，获得ram参数的值就是acl。

**ram需要两个参数用户id和action，用户id由认证体系提供的UID获得，action参数由路过提供的静态值**

例如路由指定的action参数为Get

```golang
app.GetFunc("/* action=Get", func(ctx eudore.Context){})
```

或者控制器指定的路由参数，例如website控制器使用的action参数，由包名称、控制器名称、控制器方法组成。

[源码](https://github.com/eudore/website/blob/master/internal/controller/controller.go#L46-L56)
```golang
// GetRouteParam 方法添加路由参数信息。
func (ctl *ControllerWebsite) GetRouteParam(pkg, name, method string) string {
	pos := strings.LastIndexByte(pkg, '/') + 1
	if pos != 0 {
		pkg = pkg[pos:]
	}
	if strings.HasSuffix(name, "Controller") {
		name = name[:len(name)-len("Controller")]
	}
	return fmt.Sprintf("action=%s:%s:%s", pkg, name, method)
```

例如一条路由注册日志

github.com/eudore/website/handlers/auth.PolicyController.GetIdById就是处理函数，对应的包是auth、控制器名称是Policy(移除应用控制器的Controller后缀)、控制器方法是GetIdById，组合的action为 auth:Policy:GetIdById

`{"time":"2019-10-02 18:57:23","level":"INFO","message":"RegisterHandler: GET /api/v1/auth/policy/id/:id prefix=/api/v1/auth action=auth:Policy:GetIdById [github.com/eudore/website/handlers/auth.PolicyController.GetIdById]"}`

## 数据库设计

website使用pgsql数据库，然后建立user_info、user_permisson、user_role、user_policy表，记录用户基本信息和绑定的权限、角(jue)色、策略信息，对应是是acl、rbac、pbac三种鉴权数据。

数据库唯一约束未添加

```sql

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
```

然后创建权限、角色、策略相关的表,创建permission、role、policy三种权限对象的信息，和role绑定的权限信息。

```sql
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
```


## Acl

acl(access control list)访问控制列表，记录用户对某个权限是允许和拒绝。

Permissions记录权限对应的id，对应tb_auth_permission表。

AllowBinds和DenyBinds记录用户绑定的信息(为何不定义成map[int]map[int]bool忘记了)，对应tb_auth_user_permission表。

RamHandle先使用权限行为转换成权限id，然后map查找用户id和权限id对应的结果，如果查找到就返回结果。

[源码](https://github.com/eudore/eudore/blob/master/middleware/ram/acl.go)

```golang
type Acl struct {
	AllowBinds  map[int]map[int]struct{}
	DenyBinds   map[int]map[int]struct{}
	Permissions map[string]int
}

// RamHandle 方法实现ram.RamHandler接口，匹配一个请求。
func (acl *Acl) RamHandle(id int, perm string, ctx eudore.Context) (bool, bool) {
	permid, ok := acl.Permissions[perm]
	// 存在这个权限
	if ok {
		// 绑定Allow
		_, ok = acl.AllowBinds[id][permid]
		if ok {
			ctx.SetParam(eudore.ParamRAM, "acl")
			return true, true
		}

		_, ok = acl.DenyBinds[id][permid]
		if ok {
			ctx.SetParam(eudore.ParamRAM, "acl")
			return false, true
		}
	}

	return false, false
}
```

## RBAC

基于角色的权限访问控制（Role-Based Access Control），判断一个用户的角色是否拥有对应的权限，由于用户绑定角色、角色绑定权限，所以只需要遍历用户的全部角色的全部权限判断即可。

三表对应关系：

RoleBinds => tb_auth_user_role
PermissionBinds => tb_auth_role_permission
Permissions => tb_auth_permission

RamHandle方法先转换权限成id，然后遍历用户id对应的全部角色，再遍历角色对应的全部权限id，检查用户是否拥有这个角色，如果某个角色拥有这个权限id，那么就是用户绑定的拥有这个权限(优化：未使用二分，匹配性能可提升4倍)。

[源码](https://github.com/eudore/eudore/blob/master/middleware/ram/rbac.go)

```golang
type (
	// Rbac 定义rbac对象。
	Rbac struct {
		RoleBinds       map[int][]int
		PermissionBinds map[int][]int
		Permissions     map[string]int
	}
)
// RamHandle 方法实现ram.RamHandler接口。
func (r *Rbac) RamHandle(id int, name string, ctx eudore.Context) (bool, bool) {
	permid, ok := r.Permissions[name]
	if !ok {
		return false, false
	}
	// 遍历角色
	for _, roles := range r.RoleBinds[id] {
		// 遍历权限
		for _, perm := range r.PermissionBinds[roles] {
			// 匹配权限
			if perm == permid {
				ctx.SetParam(eudore.ParamRAM, "rbac")
				return true, true
			}
		}
	}
	return false, false
}
```

## PBAC

PBAC基于策略的权限控制，一个用户有多个策略，依次判断策略匹配结果，pbac也是eudore-website主要使用的鉴权方式。

[源码](https://github.com/eudore/eudore/blob/master/middleware/ram/pbac.go)

```golang
type (
	// Pbac 定义PBAC鉴权对象。
	Pbac struct {
		PolicyBinds map[int][]int   `json:"-" key:"-"`
		Policys     map[int]*Policy `json:"-" key:"-"`
	}
)

// RamHandle 方法实现ram.RamHandler接口，匹配一个请求。
func (p *Pbac) RamHandle(id int, action string, ctx eudore.Context) (bool, bool) {
	// 获得资源resource
	resource := getResource(ctx)
	bs, ok := p.PolicyBinds[id]
	if ok {
		// 遍历全部策略
		for _, b := range bs {
			// 检查策略id是否存在
			ps, ok := p.Policys[b]
			if !ok {
				continue
			}
			// 匹配策略描述
			for _, s := range ps.Statement {
				if s.MatchAction(action) && s.MatchResource(resource) && s.MatchCondition(ctx) {
					ctx.SetParam(eudore.ParamRAM, "pbac")
					return s.Effect, true
				}
			}
		}
	}
	return false, false
}

// getResource 函数未更新
func getResource(ctx eudore.Context) string {
	path := ctx.Path()
	prefix := ctx.GetParam("prefix")
	if prefix != "" {
		path = path[len(prefix):]
	}
	ctx.SetParam("resource", path)
	return path
}

```

### 策略

eudore pbac的策略对象会绑定多个描述对象，每个描述对象具有鉴权结果(effect)、行为、资源和多项条件。

例如一个策略：

定义了一个描述对象，如果行为是auth和status任意对象的的Get方法就会通过，同时限制了请求时间是2021年前和http请求方法是GET方法(browser限制ua未实现)。

```json
{
    "version": "1",
    "description": "全部文档只读权限",
    "statement": [
        {
            "effect": true,
            "action": [
                "auth:*:Get*",
                "status:*:Get*"
            ],
            "resource": [
                "*"
            ],
            "conditions": {
                "time": {
                    "befor": "2020-12-31"
                },
                "method": [
                    "GET"
                ],
                "browser": [
                    "Chrome/60+",
                    "Chromium/0-90",
                    "Firefox"
                ]
            }
        }
    ]
}
```

go定义的Policy对象，其中Conditions作为接口，允许扩展多种条件限制，当前允许or、and、sourceip、time、method这些条件。

[源码](https://github.com/eudore/eudore/blob/master/middleware/ram/policy.go)

```golang
type (
	// Policy 定义一个策略。
	Policy struct {
		Description string      `json:"description"`
		Version     string      `json:"version"`
		Statement   []Statement `json:"statement"`
	}
	// Statement 定义一条策略内容。
	Statement struct {
		Effect     bool
		Action     []string
		Resource   []string
		Conditions *Conditions `json:"conditions,omitempty"`
	}
	// Conditions 定义PBAC使用的条件对象。
	Conditions struct {
		Conditions []Condition
	}

	// Condition 定义策略条件
	Condition interface {
		Name() string
		Match(ctx eudore.Context) bool
	}
	ConditionOr      struct {
		Conditions []Condition
	}
	ConditionAnd struct {
		Conditions []Condition
	}
	ConditionSourceIp struct {
		SourceIp []*net.IPNet
	}
	ConditionTime struct {
		Befor time.Time `json:"befor"`
		After time.Time `json:"after"`
	}
	ConditionMethod struct {
		Methods []string
	}
)
```
PBAC存在问题：

RegisterCondition函数忘记写了

获取resource对象未更新

browser限制ua未实现

# Website Ram 封装

website需要对ram数据与数据库同步，没有直接使用eudore-RAM，而是进行了简单封装。

## RAM

website-ram重新实现了eudore.RamHttp对象，同时额外添加用户访问自己资源通过，如果路由参数中具有username和userid就是访问属于用户自己的资源。

init系列函数是初始化7张权限表数据到ram对象

[源码](https://github.com/eudore/website/blob/master/internal/middleware/ram.go)

```golang
import eram "github.com/eudore/eudore/middleware/ram"

type Ram struct {
	Acl  *eram.Acl
	Rbac *eram.Rbac
	Pbac *eram.Pbac
}

func NewRam(app *eudore.App) *Ram {
	db, ok := app.Config.Get("keys.db").(*sql.DB)
	if !ok {
		panic("init middleware check config 'keys.db' not find database.")
	}
	ram := &Ram{
		Acl:  eram.NewAcl(),
		Rbac: eram.NewRbac(),
		Pbac: eram.NewPbac(),
	}

	errs := eudore.NewErrors()
	// 初始化: 权限、策略
	// TODO: 数据修改并发问题
	errs.HandleError(ram.InitPermissionInfo(db))
	errs.HandleError(ram.InitPolicyInfo(db))
	// 初始化: 用户绑定权限、用户绑定教师、角色绑定权限、用户绑定策略
	errs.HandleError(ram.InitUserBindPermission(db))
	errs.HandleError(ram.InitUserBindRole(db))
	errs.HandleError(ram.InitRoleBindPermission(db))
	errs.HandleError(ram.InitUserBindPolicy(db))
	if errs.GetError() != nil {
		panic(errs.GetError())
	}

	// 传递ram对象
	app.Set("keys.ram", ram)
	return ram
}

func (ram *Ram) NewRamFunc() eudore.HandlerFunc {
	handler := eram.NewRamAny(
		ram.Acl,
		ram.Rbac,
		ram.Pbac,
		eram.DenyHander,
	).RamHandle
	return func(ctx eudore.Context) {
		// 如果请求用户资源是用户本身的直接通过，UID、UNAME由用户信息中间件加载，userid、username由路由参数加载。
		if ctx.GetParam("userid") == ctx.GetParam("UID") && ctx.GetParam("userid") != "" {
			return
		}
		if ctx.GetParam("username") == ctx.GetParam("UNAME") && ctx.GetParam("username") != "" {
			return
		}

		// 执行ram鉴权逻辑
		action := ctx.GetParam("action")
		if len(action) > 0 && !eram.HandleDefaultRam(eudore.GetInt(ctx.GetParam("UID")), action, ctx, handler) {
			ctx.WriteHeader(403)
			ctx.Render(map[string]interface{}{
				eudore.ParamRAM:    ctx.GetParam("ram"),
				eudore.ParamAction: action,
			})
			ctx.End()
		}
	}
}
```

## Init

InitPermissionInfo方法就从数据库查询权限信息，然后赋值给ACL和RBAC对象。

其他init函数类似。

```golang
func (ram *Ram) InitPermissionInfo(db *sql.DB) error {
	rows, err := db.Query("SELECT id,name FROM tb_auth_permission")
	if err != nil {
		return err
	}
	defer rows.Close()

	var Permissions = make(map[string]int)
	var id int
	var name string
	for rows.Next() {
		err = rows.Scan(&id, &name)
		if err != nil {
			return err
		}
		Permissions[name] = id
	}
	// 共享权限信息
	ram.Acl.Permissions = Permissions
	ram.Rbac.Permissions = Permissions
	return nil
}
```

## Controller

例如[策略控制器](https://github.com/eudore/website/blob/master/handlers/auth/policy.go)赋值策略的管理，实现策略CURD和RAM信息同步，其他User、Permission、Role三个控制器行为类型。

```godoc
type PolicyController
    func NewPolicyController(db *sql.DB, ram *middleware.Ram) *PolicyController
    func (ctl *PolicyController) DeleteIdById() (err error)
    func (ctl *PolicyController) DeleteNameByName() (err error)
    func (ctl *PolicyController) GetCount() interface{}
    func (ctl *PolicyController) GetIdById() (interface{}, error)
    func (ctl *PolicyController) GetIndex() (interface{}, error)
    func (ctl *PolicyController) GetList() (interface{}, error)
    func (ctl *PolicyController) GetNameByName() (interface{}, error)
    func (ctl *PolicyController) GetSearchByKey() (interface{}, error)
    func (ctl *PolicyController) GetUserIdById() (interface{}, error)
    func (ctl *PolicyController) GetUserNameByName() (interface{}, error)
    func (ctl *PolicyController) PostIdById() (err error)
    func (ctl *PolicyController) PostNameByName() (err error)
    func (ctl *PolicyController) PutNew() (err error)
    func (ctl *PolicyController) Release() error
```

如果请求方法是POST、PUT、DELETE就是对策略信息有所修改，就调用RAM重新初始化策略数据，实现鉴权信息同步。

```golang
type PolicyController struct {
	controller.ControllerWebsite
	Ram *middleware.Ram
}

// Release 方法用于刷新ram策略信息。
func (ctl *PolicyController) Release() error {
	// 如果修改策略信息成功，则刷新ram策略信息。
	if ctl.Response().Status() == 200 && (ctl.Method() == "POST" || ctl.Method() == "PUT" || ctl.Method() == "DELETE") {
		ctl.Ram.InitPolicyInfo(ctl.DB)
	}
	return nil
}

```

[用户控制器实现同步用户绑定权限](https://github.com/eudore/website/blob/master/handlers/auth/user.go#L113-L161)

```golang
type UserController struct {
	controller.ControllerWebsite
	Ram *middleware.Ram
}

// Release 方法刷新用户绑定ram资源信息。
func (ctl *UserController) Release() error {
	// 如果修改策略信息成功，则刷新ram策略信息。
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

```

[用户绑定策略CURD](https://github.com/eudore/website/blob/master/handlers/auth/user.go#L319-L359)

```golang
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
```

# 访问日志

访问日志记录了请求信息，可以清晰的看到权限相关的行为。

```
{"time":"2019-10-02 19:27:14","level":"INFO","fields":{"time":"1.066777ms","route":"/auth/","method":"GET","path":"/auth/","proto":"HTTP/1.1","status":200,"remote":"59.63.178.92","host":"47.52.173.119:8082","size":582,"x-request-id":"294459f1b000000"}}
{"time":"2019-10-02 19:27:14","level":"INFO","fields":{"ram":"ram-pbac","remote":"59.63.178.92","proto":"HTTP/1.1","action":"auth:Permission:GetCount","resource":"/permission/count","x-request-id":"294459f67c00000","method":"GET","status":200,"route":"/api/v1/auth/permission/count","size":36,"x-parent-id":"294459f1b000000","path":"/api/v1/auth/permission/count","host":"47.52.173.119:8082","time":"1.337828ms"}}
{"time":"2019-10-02 19:27:14","level":"INFO","fields":{"remote":"59.63.178.92","proto":"HTTP/1.1","time":"1.695997ms","x-parent-id":"294459f1b000000","method":"GET","action":"auth:Permission:GetIndex","ram":"ram-pbac","x-request-id":"294459f68000000","host":"47.52.173.119:8082","status":200,"route":"/api/v1/auth/permission/index","path":"/api/v1/auth/permission/index","size":225,"resource":"/permission/index"}}
{"time":"2019-10-02 19:27:14","level":"INFO","fields":{"method":"GET","path":"/api/v1/auth/user/icon/name/root","action":"auth:User:GetIconNameByName","route":"/api/v1/auth/user/icon/name/:name","x-request-id":"294459f76c00000","remote":"59.63.178.92","proto":"HTTP/1.1","host":"47.52.173.119:8082","status":200,"time":"1.029588ms","size":12164,"ram":"ram-acl"}}
```


例如第二条格式化结果：

```json
{
  "time": "2019-10-02 19:27:14",
  "level": "INFO",
  "fields": {
    "ram": "ram-pbac",
    "remote": "59.63.178.92",
    "proto": "HTTP/1.1",
    "action": "auth:Permission:GetCount",
    "resource": "/permission/count",
    "x-request-id": "294459f67c00000",
    "method": "GET",
    "status": 200,
    "route": "/api/v1/auth/permission/count",
    "size": 36,
    "x-parent-id": "294459f1b000000",
    "path": "/api/v1/auth/permission/count",
    "host": "47.52.173.119:8082",
    "time": "1.337828ms"
  }
}
```

其中部分参数含义：

|  参数  | 值  | 含义  |
| ------------ | ------------ | ------------ |
|  path |  /api/v1/auth/permission/count | http请求路径  |
|  route | /api/v1/auth/permission/count  | 路由匹配规则  |
|  action | auth:Permission:GetCount  | 处理行为  |
|  ram | ram-pbac  |  ram执行者，如果status非403执行结果为通过 |
|  resoure | /permission/count  |  资源值，仅pbac存在 |

