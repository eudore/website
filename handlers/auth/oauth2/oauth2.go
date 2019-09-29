// golang oauth2 define.
//
package oauth2

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/eudore/website/util/jwt"
	"github.com/eudore/website/util/tool"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/eudore/eudore"
	"golang.org/x/oauth2"
)

// Ouath2 handle
type (
	Oauth2 interface {
		// Set Ouath2 config
		Config(*oauth2.Config)
		// Get redirect Addr
		Redirect(string) string
		// Handle callback request
		Callback(eudore.Context) (map[string]interface{}, error)
		GetUserId(data map[string]interface{}) string
	}
	//
	responseLoginAuth struct {
		UserId string `json:"userid"`
		Name   string `json:"name"`
		// GrantId int    `json:"grantid,omitempty"`
		Bearer  string `json:"bearer,omitempty"`
		Expires int64  `json:"expires,omitempty"`
	}
)

var (
	stmtQueryOauth2Source *sql.Stmt
	stmtQueryOauth2Login  *sql.Stmt
	tokenVerify           jwt.VerifyFunc
	ErrOauthState         = errors.New("invalid oauth state")
	ErrOauthCode          = errors.New("Code exchange failed")
)

// Reload oauth2 config
func Init(app *eudore.Eudore) error {
	// init db
	db, ok := app.Config.Get("keys.db").(*sql.DB)
	if !ok {
		fmt.Errorf("keys.db not find database.")
	}

	// init stmt
	var sqls = map[**sql.Stmt]string{
		&stmtQueryOauth2Login: `SELECT userid FROM tb_auth_oauth2_login WHERE source=$1 AND originid=$2 AND state=0`,
	}
	err := tool.InitStmt(db, sqls)
	if err != nil {
		return err
	}

	// init source
	rows, err := db.Query("SELECT Name,ClientID,ClientSecret FROM tb_auth_oauth2_source;")
	if err != nil {
		return err
	}

	api := app.Group("/auth")
	var name, id, secret string
	for rows.Next() {
		rows.Scan(&name, &id, &secret)
		o, err := NewOuath2(name)
		if err != nil {
			app.Error(err)
			continue
		}

		o.Config(&oauth2.Config{
			ClientID:     id,
			ClientSecret: secret,
			RedirectURL:  "https://www.wejass.com:8083/auth/callback/" + name,
		})

		api.GetFunc("/login/"+name, redirectfunc(name, o, app.Cache))
		api.GetFunc("/callback/"+name, callbackfunc(name, o, app.Cache))
	}

	tokenVerify = jwt.NewVerifyHS256([]byte("secret"))
	return nil
}

func getRandomString() string {
	letters := []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY")
	result := make([]rune, 16)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}

func joinConfig(dst, src *oauth2.Config) {
	if src.ClientID != "" {
		dst.ClientID = src.ClientID
	}
	if src.ClientSecret != "" {
		dst.ClientSecret = src.ClientSecret
	}
	if src.Endpoint.AuthURL != "" && src.Endpoint.TokenURL != "" {
		dst.Endpoint = src.Endpoint
	}
	if src.RedirectURL != "" {
		dst.RedirectURL = src.RedirectURL
	}
	if src.Scopes != nil {
		dst.Scopes = src.Scopes
	}
}

func redirectfunc(name string, o Oauth2, cache eudore.Cache) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		oauthState := getRandomString()
		location := ctx.GetQuery("location")
		cache.Set(oauthState, location, 15*time.Minute)
		fmt.Println(oauthState, location)
		url := o.Redirect(oauthState)
		ctx.Redirect(http.StatusFound, url)
	}
}

func callbackfunc(name string, o Oauth2, cache eudore.Cache) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		state := ctx.GetQuery("state")
		//
		if !cache.IsExist(state) {
			// expire
		}

		data, err := o.Callback(ctx)
		if err != nil {
			ctx.Fatal(err)
			return
		}

		var userid int
		id := o.GetUserId(data)
		err = stmtQueryOauth2Login.QueryRow(name, id).Scan(&userid)
		if err != nil {
			// sign up
			ctx.Infof("%s register %s", name, id)
			return
		}

		// sign in
		ctx.Infof("%s sign in %s", name, id)
		var resp = &responseLoginAuth{
			UserId: fmt.Sprint(userid),
			// Name:    auth.Name,
			Expires: time.Now().Add(time.Hour).Unix(),
		}

		// redirect
		location := eudore.GetDefaultString(cache.Get(state), "/")
		if strings.IndexByte(location, '?') == -1 {
			location += "?authorization="
		} else {
			location += "&authorization="
		}
		location += tokenVerify.Signed(resp)
		ctx.Redirect(302, location)
	}
}
