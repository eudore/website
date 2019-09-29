package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/eudore/eudore"
)

const (
	BearerStar = "Bearer "
)

type (
	VerifyFunc func([]byte) string
)

func NewVerifyHS256(secret []byte) VerifyFunc {
	return func(b []byte) string {
		h := hmac.New(sha256.New, secret)
		h.Write(b)
		return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	}
}

func NewJwt(fn VerifyFunc) eudore.HandlerFunc {
	if fn == nil {
		fn = NewVerifyHS256([]byte("secret"))
	}
	return func(ctx eudore.Context) {
		jwtstr := ctx.GetHeader(eudore.HeaderAuthorization)
		if len(jwtstr) == 0 {
			return
		}
		if strings.HasPrefix(jwtstr, BearerStar) {
			jwt, err := fn.ParseMap(jwtstr[7:])
			if err != nil {
				ctx.WithField("error", "jwt invalid").Warning(err)
				return
			}
			if int64(jwt["expires"].(float64)) < time.Now().Unix() {
				ctx.Warning("jwt expirese")
				return
			}
			// ctx.SetValue(eudore.ValueJwt, jwt)
		} else {
			ctx.WithField("error", "bearer invalid").Warning("")
		}
	}
}

func (fn VerifyFunc) Signed(claims interface{}) string {
	payload, _ := json.Marshal(claims)
	var unsigned string = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.` + base64.RawURLEncoding.EncodeToString(payload)
	return fmt.Sprintf("%s.%s", unsigned, fn([]byte(unsigned)))
}

func (fn VerifyFunc) ParseMap(token string) (dst map[string]interface{}, err error) {
	dst = make(map[string]interface{})
	err = fn.Parse(token, &dst)
	return
}

func (fn VerifyFunc) Parse(token string, dst interface{}) error {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return errors.New("Error: incorrect # of results from string parsing.")
	}

	if fn([]byte(parts[0]+"."+parts[1])) != parts[2] {
		return errors.New("Error：jwt validation error.")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return err
	}
	//
	err = json.Unmarshal(payload, &dst)
	if err != nil {
		return err
	}
	return nil
}

func (fn VerifyFunc) ParseBearer(jwtstr string) (map[string]interface{}, error) {
	if strings.HasPrefix(jwtstr, BearerStar) {
		jwt, err := fn.ParseMap(jwtstr[7:])
		if err != nil {
			return nil, err
		}
		if int64(jwt["expires"].(float64)) < time.Now().Unix() {
			return nil, errors.New("jwt expires")
		}
		return jwt, nil
	}
	return nil, errors.New("Bearer invalid")
}
