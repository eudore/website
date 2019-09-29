package middleware

import (
	"database/sql"

	"github.com/eudore/eudore"
	eram "github.com/eudore/eudore/middleware/ram"
)

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

func (ram *Ram) InitPolicyInfo(db *sql.DB) error {
	rows, err := db.Query("SELECT id,policy FROM tb_auth_policy")
	if err != nil {
		return err
	}
	defer rows.Close()

	var Policys = make(map[int]*eram.Policy)
	var id int
	var policy string
	for rows.Next() {
		err = rows.Scan(&id, &policy)
		if err != nil {
			return err
		}
		Policys[id] = eram.NewPolicyStringJSON(policy)
	}
	ram.Pbac.Policys = Policys
	return nil
}

func (ram *Ram) InitUserBindPermission(db *sql.DB) error {
	rows, err := db.Query("SELECT userid,permissionid,effect FROM tb_auth_user_permission")
	if err != nil {
		return err
	}
	defer rows.Close()

	var AllowBinds = make(map[int]map[int]struct{})
	var DenyBinds = make(map[int]map[int]struct{})
	var userid int
	var permissionid int
	var effect bool
	for rows.Next() {
		err = rows.Scan(&userid, &permissionid, &effect)
		if err != nil {
			return err
		}
		if effect {
			as, ok := AllowBinds[userid]
			if !ok {
				as = make(map[int]struct{})
				AllowBinds[userid] = as
			}
			as[permissionid] = struct{}{}
		} else {
			ds, ok := DenyBinds[userid]
			if !ok {
				ds = make(map[int]struct{})
				DenyBinds[userid] = ds
			}
			ds[permissionid] = struct{}{}

		}
	}
	ram.Acl.AllowBinds = AllowBinds
	ram.Acl.DenyBinds = DenyBinds
	return nil
}

func (ram *Ram) InitUserBindRole(db *sql.DB) error {
	rows, err := db.Query("SELECT userid,roleid FROM tb_auth_user_role")
	if err != nil {
		return err
	}
	defer rows.Close()

	var RoleBinds = make(map[int][]int)
	var userid, roleid int
	for rows.Next() {
		err = rows.Scan(&userid, &roleid)
		if err != nil {
			return err
		}
		RoleBinds[userid] = append(RoleBinds[userid], roleid)
	}
	ram.Rbac.RoleBinds = RoleBinds
	return nil
}

func (ram *Ram) InitRoleBindPermission(db *sql.DB) error {
	rows, err := db.Query("SELECT roleid,permissionid FROM tb_auth_role_permission")
	if err != nil {
		return err
	}
	defer rows.Close()

	var PermissionBinds = make(map[int][]int)
	var roleid, permid int
	for rows.Next() {
		err = rows.Scan(&roleid, &permid)
		if err != nil {
			return err
		}
		PermissionBinds[roleid] = append(PermissionBinds[roleid], permid)
	}
	ram.Rbac.PermissionBinds = PermissionBinds
	return nil
}

func (ram *Ram) InitUserBindPolicy(db *sql.DB) error {
	rows, err := db.Query("SELECT userid,policyid FROM tb_auth_user_policy")
	if err != nil {
		return err
	}
	defer rows.Close()

	var PolicyBinds = make(map[int][]int)
	var userid, policyid int
	for rows.Next() {
		err = rows.Scan(&userid, &policyid)
		if err != nil {
			return err
		}
		PolicyBinds[userid] = append(PolicyBinds[userid], policyid)
	}
	ram.Pbac.PolicyBinds = PolicyBinds
	return nil
}
