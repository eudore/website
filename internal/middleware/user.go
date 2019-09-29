package middleware

import (
	"crypto/hmac"
	"crypto/sha1"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/website/util/jwt"
)

/*
PostgreSQL Begin

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

func NewUserInfoFunc(app *eudore.App) eudore.HandlerFunc {
	db, ok := app.Config.Get("keys.db").(*sql.DB)
	if !ok {
		panic(fmt.Errorf("keys.db not find database."))
	}
	stmtQueryAccessToken, err := db.Prepare("SELECT userid,(SELECT name FROM tb_auth_user_info WHERE id = userid) FROM tb_auth_access_token WHERE token=$1 and expires > now()")
	if err != nil {
		panic(err)
	}
	stmtQueryAccessKey, err := db.Prepare("SELECT userid,(SELECT name FROM tb_auth_user_info WHERE id = userid),accesssecrect FROM tb_auth_access_key WHERE accesskey=$1 and expires > $2")

	jwtParse := jwt.NewVerifyHS256([]byte("secret"))
	return func(ctx eudore.Context) {
		data, err := jwtParse.ParseBearer(ctx.GetHeader(eudore.HeaderAuthorization))
		if err == nil {
			ctx.SetParam("UID", eudore.GetString(data["userid"]))
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
			fmt.Println(time.Now().Add(50 * time.Minute).Unix())
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
}
