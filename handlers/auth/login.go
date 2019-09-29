package auth

import (
	"encoding/hex"
	"fmt"
	"github.com/eudore/eudore"
	"golang.org/x/crypto/scrypt"
	"time"
)


/*
PostgreSQL Begin

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


INSERT INTO "tb_auth_user_pass"("name", "pass", "salt", "id") VALUES ('root', 'd1fbb03f8a717f3d9cd2cf3e59d39fd1a227b7fc5ee2cea831b4050a1ae4dbe4', '0123456789', (SELECT id FROM tb_auth_user_info WHERE "name"='root'));

PostgreSQL End
*/

type (
	requestLoginAuth struct {
		Name    string `bind:"name" json:"name"`
		Pass    string `bind:"pass" json:"pass"`
		Captcha string `bind:"captcha" json:"captcha"`
		Verify  string `bind:"verify" json:"verify"`
	}
	responseLoginAuth struct {
		UserId string `json:"userid"`
		Name   string `json:"name"`
		// GrantId int    `json:"grantid,omitempty"`
		Bearer  string `json:"bearer,omitempty"`
		Expires int64  `json:"expires,omitempty"`
	}
	oauth2Pass struct {
		Name string
		Pass string
		Salt []byte
		Id   int
	}
	oauth2Authorize struct {
		Name          string
		Response_type string
		Client_id     string
		Redirect_uri  string
		Scope         []string
		State         string
	}
)

func postLoginWejass(ctx eudore.Context) {
	var auth requestLoginAuth
	ctx.Bind(&auth)

	// 检测验证码
	var captcha captchaSigned
	err := tokenVerify.Parse(auth.Verify, &captcha)
	if err != nil {
		ctx.Fatal(err)
	}
	eudore.JSON(captcha)
	if captcha.Expires < time.Now().Unix() {
		return
	}

	if !captcha.CheckVerify(auth.Captcha) {
		return
	}

	// 读取用户密码信息
	eudore.JSON(auth)
	var pass oauth2Pass
	err = stmtQueryOauth2Pass.QueryRow(auth.Name).Scan(&pass.Pass, &pass.Salt, &pass.Id)
	if err != nil {
		ctx.Fatal(err)
		return
	}

	// 计算hash
	keys, err := scrypt.Key([]byte(auth.Pass), pass.Salt, 16384, 8, 1, 32)
	if err != nil {
		ctx.Fatal(err)
		return
	}

	// 验证密码
	if hex.EncodeToString(keys) != pass.Pass {
		ctx.Fatal("pass not equal")
		return
	}
	ctx.Infof("login user %s", auth.Name)

	var resp = &responseLoginAuth{
		UserId:  fmt.Sprint(pass.Id),
		Name:    auth.Name,
		Expires: time.Now().Add(72 * time.Hour).Unix(),
	}
	resp.Bearer = "Bearer " + tokenVerify.Signed(resp)
	ctx.Render(resp)
}

// setUserPass("root", "qwer")
func setUserPass(name, newpass string) error {
	var pass oauth2Pass
	// fmt.Println("setUserPass", name, newpass)
	err := stmtQueryOauth2Pass.QueryRow(name).Scan(&pass.Pass, &pass.Salt, &pass.Id)
	if err != nil {
		return err
	}

	keys, err := scrypt.Key([]byte(newpass), pass.Salt, 16384, 8, 1, 32)
	if err != nil {
		return err
	}

	_, err = stmtUpdateOauth2Pass.Exec(hex.EncodeToString(keys), name)
	if err != nil {
		return err
	}
	return nil
}
