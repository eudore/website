package auth

import (
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/dchest/captcha"
	"github.com/eudore/eudore"
	"github.com/eudore/website/handlers/auth/oauth2"
	"github.com/eudore/website/internal/controller"
	"github.com/eudore/website/util/jwt"
	"golang.org/x/crypto/scrypt"
)

/*
PostgreSQL Begin

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


INSERT INTO "tb_auth_user_pass"("name", "pass", "salt", "id") VALUES ('root', 'd1fbb03f8a717f3d9cd2cf3e59d39fd1a227b7fc5ee2cea831b4050a1ae4dbe4', '0123456789', (SELECT id FROM tb_auth_user_info WHERE "name"='root'));

PostgreSQL End
*/

type (
	LoginController struct {
		controller.ControllerWebsite
		Oauth2 map[string]oauth2.Oauth2
	}
	captchaSigned struct {
		Verify  string `json:"verify"`
		Expires int64  `json:"expires"`
	}
	requestLoginAuth struct {
		Name    string `bind:"name" json:"name"`
		Pass    string `bind:"pass" json:"pass"`
		Captcha string `bind:"captcha" json:"captcha"`
		Verify  string `bind:"verify" json:"verify"`
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

var (
	letterBytes         = []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	captchaSecret       = []byte("secret")
	captchaSalf         = "9999"
	tokenVerify         jwt.VerifyFunc
	errCaptchaExpires   = errors.New("captcha is expires")
	errCaptchaInvalid   = errors.New("captcha is invalid")
	errLoginPassInvalid = errors.New("login website password is invalid")
)

func NewLoginController(app *eudore.Eudore, db *sql.DB) *LoginController {
	tokenVerify = jwt.NewVerifyHS256(app.GetBytes("auth.secrets.jwt"))
	captchaSecret = app.GetBytes("auth.secrets.captchakey")
	captchaSalf = app.GetString("auth.secrets.captchasalf", "9999")

	ctl := &LoginController{
		ControllerWebsite: controller.ControllerWebsite{
			DB: db,
		},
	}
	ctl.loadOauth2(app.App, db)
	return ctl
}

func (ctl *LoginController) loadOauth2(app *eudore.App, db *sql.DB) {
	rows, err := db.Query("SELECT name,server,clientid,clientsecret FROM tb_auth_oauth2_source")
	if err != nil {
		app.Error(err)
		return
	}

	oauths := make(map[string]oauth2.Oauth2)
	for rows.Next() {
		var name, server, id, secret string
		rows.Scan(&name, &server, &id, &secret)
		o, err := oauth2.NewOuath2(name)
		if err != nil {
			app.Error(err)
			continue
		}
		o.Config(&oauth2.Config{
			ClientID:     id,
			ClientSecret: secret,
			RedirectURL:  fmt.Sprintf("https://%s/api/v1/auth/login/callback/%s", server, name),
		})
		app.Infof("create %s oauth2 %s clientid: %s", server, name, id)
		oauths[server+" "+name] = o
	}
	ctl.Oauth2 = oauths
}

func (ctl *LoginController) GetRouteParam(pkg, name, method string) string {
	return ""
}

func (ctl *LoginController) createBase(userid int) map[string]interface{} {
	var name, lang string
	ctl.QueryRow("SELECT name,lang FROM tb_auth_user_info WHERE id=$1", userid).Scan(&name, &lang)
	return map[string]interface{}{
		"userid":  userid,
		"name":    name,
		"expires": time.Now().Add(72 * time.Hour).Unix(),
		"lang":    lang,
		"bearer": "Bearer " + tokenVerify.Signed(map[string]interface{}{
			"userid":  userid,
			"name":    name,
			"expires": time.Now().Add(72 * time.Hour).Unix(),
		}),
	}
}
func (ctl *LoginController) GetCaptcha() {
	h := ctl.Response().Header()
	h.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	h.Set("Pragma", "no-cache")
	h.Set("Expires", "0")

	var data = make([]byte, 6)
	for i := 0; i < 6; i++ {
		data[i] = letterBytes[rand.Intn(10)]
	}

	val, _ := scrypt.Key(data, captchaSecret, 16384, 8, 1, 32)
	h.Set("Captcha", tokenVerify.Signed(&captchaSigned{
		Verify:  hex.EncodeToString(val),
		Expires: time.Now().Add(5 * time.Minute).Unix(),
	}))

	img := captcha.NewImage(captchaSalf, data, 120, 40)
	// 兼容base64图片
	if ctl.GetHeader("Accept") == "application/base64" {
		h.Set("Content-Type", "application/base64")
		ctl.WriteString("data:image/png;base64,")
		encoder := base64.NewEncoder(base64.StdEncoding, ctl)
		img.WriteTo(encoder)
		encoder.Close()
		return
	}
	h.Set("Content-Type", "image/png")
	img.WriteTo(ctl)
}

func (ctl *LoginController) PostWebsite() error {
	var auth requestLoginAuth
	ctl.Bind(&auth)

	// 检测验证码
	var captcha captchaSigned
	err := tokenVerify.Parse(auth.Verify, &captcha)
	if err != nil {
		return err
	}
	if captcha.Expires < time.Now().Unix() {
		return errCaptchaExpires
	}

	if !captcha.checkVerify(auth.Captcha) {
		return errCaptchaInvalid
	}

	// 读取用户密码信息
	var userid int
	var pass string
	var salt []byte
	err = ctl.QueryRow("SELECT id,pass,salt FROM tb_auth_user_pass WHERE name=$1;", auth.Name).Scan(&userid, &pass, &salt)
	if err != nil {
		return err
	}

	// 计算hash
	keys, err := scrypt.Key([]byte(auth.Pass), salt, 16384, 8, 1, 32)
	if err != nil {
		return err
	}

	// 验证密码
	if hex.EncodeToString(keys) != pass {
		return errLoginPassInvalid
	}
	ctl.Infof("login user %s", auth.Name)

	return ctl.Render(ctl.createBase(userid))
}

func (ctl *LoginController) GetOauth2ByName() {
	o, ok := ctl.Oauth2[ctl.Host()+" "+ctl.GetParam("name")]
	if ok {
		oauthState := oauth2.GetRandomString()
		ctl.Redirect(302, o.Redirect(oauthState))
	} else {
		ctl.WriteString(fmt.Sprintf("%s oauth2 %s is invalid", ctl.Host(), ctl.GetParam("name")))
	}
}

func (ctl *LoginController) GetCallbackByName() error {
	name := ctl.GetParam("name")
	o, ok := ctl.Oauth2[ctl.Host()+" "+name]
	if !ok {
		return fmt.Errorf("%s oauth2 %s is invalid", ctl.Host(), name)
	}

	_, oauthid, err := o.Callback(ctl.Request())
	if err != nil {
		return err
	}

	var userid int
	err = ctl.QueryRow("SELECT userid FROM tb_auth_oauth2_login WHERE source=$1 AND originid=$2 AND state=0", name, oauthid).Scan(&userid)
	if err != nil {
		ctl.SetHeader("Content-Type", "text/plain; charset=utf-8")
		ctl.WriteString("signup or bind\r\n")
		ctl.WriteString(fmt.Sprintf("source %s userid: %s", name, oauthid))
		return nil
	}

	body, _ := json.Marshal(ctl.createBase(userid))
	ctl.SetHeader("Content-Security-Policy", "")
	ctl.SetHeader("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(ctl, `<!DOCTYPE html>
<html>
<script>
opener.localStorage.setItem('user', '%s')
opener.window.location = opener.redirect
window.close();
</script>
</html>`, body)
	return nil
}

func (cs captchaSigned) checkVerify(key string) bool {
	var data = make([]byte, 6)
	for i := 0; i < 6; i++ {
		data[i] = key[i] - '0'
	}

	val, _ := scrypt.Key(data, captchaSecret, 16384, 8, 1, 32)
	return hex.EncodeToString(val) == cs.Verify
}

/*
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
*/
