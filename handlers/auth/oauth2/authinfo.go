package oauth2

import (
	"encoding/gob"
	"time"
	// "github.com/dgrijalva/jwt-go"
)

// Save callback get user info
type AuthInfo struct {
	Source int
	Id     string
	Name   string
	Email  string
	Uid    int
}

func init() {
	gob.Register(AuthInfo{})
}

func (au *AuthInfo) getuid() (err error) {
	/*	var stats int = -1
		err = sql.GetDB().QueryRow("SELECT UID,Stats FROM tb_auth_oauth2_login WHERE Source=? AND OID=?;", au.Source, au.Id).Scan(&au.Uid, &stats)
		if err != nil {
			if stmt, err := sql.GetDB().Prepare("INSERT tb_auth_oauth2_login(Source,Name,Email,OID,Stats) VALUES(?,?,?,?,?);"); err == nil {
				_, err = stmt.Exec(au.Source, au.Name, au.Email, au.Id, 1)
			}
			return
		}*/
	return
}

// User info to jwt
func (au *AuthInfo) GetJwt() string {
	return tokenVerify.Signed(map[string]interface{}{
		"source":     au.Source,
		"sourcename": au.GetSource(),
		"id":         au.Id,
		"name":       au.Name,
		"uid":        au.Uid,
		"expires":    time.Now().Add(1000 * time.Second).Unix(),
	})
}

// Get user auth type
func (au *AuthInfo) GetSource() string {
	return oauth2source[au.Source]
}
