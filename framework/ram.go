package framework

import (
	"crypto/hmac"
	"crypto/sha1"
	"database/sql"
	"encoding/base64"
	"fmt"
	eram "github.com/eudore/eudore/component/ram"
	"strconv"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/website/util/jwt"
)

func NewUserInfoFunc(app *App) eudore.HandlerFunc {
	db := app.DB
	stmtQueryAccessToken, err := db.Prepare("SELECT userid,(SELECT name FROM tb_auth_user_info WHERE id = userid) FROM tb_auth_access_token WHERE token=$1 and expires > now()")
	if err != nil {
		panic(err)
	}

	stmtQueryAccessKey, err := db.Prepare("SELECT userid,(SELECT name FROM tb_auth_user_info WHERE id = userid),accesssecrect FROM tb_auth_access_key WHERE accesskey=$1 and expires > $2")
	if err != nil {
		panic(err)
	}

	jwtParse := jwt.NewVerifyHS256([]byte(app.Get("auth.secrets.jwt").(string)))
	return func(ctx eudore.Context) {
		data, err := jwtParse.ParseBearer(ctx.GetHeader(eudore.HeaderAuthorization))
		if err == nil {
			ctx.SetParam("UID", eudore.GetString(data["userid"], fmt.Sprint(eudore.GetFloat64(data["userid"]))))
			ctx.SetParam("UNAME", eudore.GetString(data["name"]))
			return
		}

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

		// websockt
		if ctx.GetHeader("Upgrade") != "" {
			bearer := ctx.GetQuery("bearer")
			if bearer != "" {
				data, err := jwtParse.ParseBearer(bearer)
				if err == nil {
					ctx.SetParam("UID", eudore.GetString(data["userid"]))
					ctx.SetParam("UNAME", eudore.GetString(data["name"]))
					return
				}
			}
		}
	}
}

type RAM struct {
	Acl  *eram.Acl
	Rbac *eram.Rbac
	Pbac *eram.Pbac
}

func NewRAM(app *App) *RAM {
	db := app.DB
	ram := &RAM{
		Acl:  eram.NewAcl(),
		Rbac: eram.NewRbac(),
		Pbac: eram.NewPbac(),
	}

	errs := NewErrors()
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

	return ram
}

func (ram *RAM) NewRAMFunc() eudore.HandlerFunc {
	return eram.NewMiddleware(ram.Acl, ram.Rbac, ram.Pbac)
}

func (ram *RAM) MatchAction(ctx eudore.Context, action string) (string, bool) {
	rams := []eram.Handler{ram.Acl, ram.Rbac, ram.Pbac}
	return eram.MatchAction(rams, ctx, action)
}

func (ram *RAM) InitPermissionInfo(db *sql.DB) error {
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
	ram.Acl.Lock()
	ram.Acl.Permissions = Permissions
	ram.Acl.Unlock()
	ram.Rbac.Lock()
	ram.Rbac.Permissions = Permissions
	ram.Rbac.Unlock()
	return nil
}

func (ram *RAM) InitPolicyInfo(db *sql.DB) error {
	rows, err := db.Query("SELECT id,name,policy FROM tb_auth_policy")
	if err != nil {
		return err
	}
	defer rows.Close()

	var Policys = make(map[int]*eram.Policy)
	var id int
	var name string
	var policy string
	for rows.Next() {
		err = rows.Scan(&id, &name, &policy)
		if err != nil {
			return err
		}
		policy, err := eram.ParsePolicyString(policy)
		policy.Name = name
		if err == nil {
			Policys[id] = policy
		}
	}
	ram.Pbac.Lock()
	ram.Pbac.Policys = Policys
	ram.Pbac.Unlock()
	return nil
}

func (ram *RAM) InitUserBindPermission(db *sql.DB) error {
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
	ram.Acl.Lock()
	ram.Acl.AllowBinds = AllowBinds
	ram.Acl.DenyBinds = DenyBinds
	ram.Acl.Unlock()
	return nil
}

func (ram *RAM) InitUserBindRole(db *sql.DB) error {
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
	ram.Rbac.Lock()
	ram.Rbac.RoleBinds = RoleBinds
	ram.Rbac.Unlock()
	return nil
}

func (ram *RAM) InitRoleBindPermission(db *sql.DB) error {
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
	ram.Rbac.Lock()
	ram.Rbac.PermissionBinds = PermissionBinds
	ram.Rbac.Unlock()
	return nil
}

func (ram *RAM) InitUserBindPolicy(db *sql.DB) error {
	rows, err := db.Query("SELECT userid,policyid FROM tb_auth_user_policy ORDER BY index")
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
	ram.Pbac.Lock()
	ram.Pbac.PolicyBinds = PolicyBinds
	ram.Pbac.Unlock()
	return nil
}
