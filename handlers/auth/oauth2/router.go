package oauth2

import (
	/*	"fmt"
		"net/http"
		"net/url"
		"time"*/
	"github.com/eudore/eudore"
)

const (
	// state cookie name
	CookieState    = "oauth2state"
	CookieRedirect = "oauth2redirect"
	CookiePath     = "/auth/oauth2/"
	CookieDomain   = "www.wejass.com"
	TokenOauth2    = "authtoken"
	TokenRediect   = "redirect"
	CallbackUrl    = "https://www.wejass.com/auth/oauth2/callback/"
)

func loadrouter(router eudore.Router) error {
	/*	rows, err := stmtQueryOauth2Source.Query()
		if err != nil {
			return err
		}
		var name, id, secret string
		for rows.Next() {
			rows.Scan(&name, &id, &secret)
			o, err := NewOuath2(name)
			if err != nil {
				fmt.Println(err, name)
				continue
			}
			fmt.Println("init oauth2", name)
			// use default config
			cf := o.Config(nil)
			cf.RedirectURL = CallbackUrl + name
			cf.ClientID = id
			cf.ClientSecret = secret
			router.GetFunc("/login/"+name, redirectfunc(o))
			router.GetFunc("/callback/"+name, callbackfunc(o))
		}*/
	return nil
}

/*
// Oauth2 login redirect func
func redirectfunc(o Oauth2) eudore.HandleFunc {
	return func(ctx eudore.Context) {
		// random stats
		oauthState := getRandomString()
		cookie := &http.Cookie{
			Name:     CookieState,
			Value:    oauthState,
			Path:     CookiePath,
			HttpOnly: true,
			Secure:   true,
			Domain:   CookieDomain,
			Expires:  time.Now().Add(100 * time.Second),
		}
		ctx.SetCookie(cookie)
		// save redirect
		redirect := ctx.GetParam("redirect").GetString()
		cookie = &http.Cookie{
			Name:     CookieRedirect,
			Value:    redirect,
			Path:     CookiePath,
			HttpOnly: true,
			Secure:   true,
			Domain:   CookieDomain,
			Expires:  time.Now().Add(100 * time.Second),
		}
		ctx.SetCookie(cookie)
		// redirect oauth2
		url := o.Redirect(oauthState)
		ctx.Redirect(http.StatusFound, url)
		fmt.Println("redirect:", url)
	}
}

// Oauth2 callback func
func callbackfunc(o Oauth2) eudore.HandleFunc {
	return func(ctx eudore.Context) {
		// check stats
		state := ctx.GetParam("state").GetString()
		cookie, err := ctx.Cookie(CookieState)
		if state != cookie.Value || err != nil {
			return
		}
		fmt.Println("state ", state)
		// load uid
		au, err := o.Callback(ctx.Request())
		err = au.getuid()
		oauth2_token, err := au.GetJwt()
		if err != nil {
			//http.Error(w,http.StatusText(403),http.StatusUnauthorized)
			ctx.WriteHeader(http.StatusUnauthorized)
		}
		fmt.Println("oauth2_token ", oauth2_token)

		// redirect
		data := url.Values{
			TokenOauth2:  {oauth2_token},
			TokenRediect: {ctx.GetCookie(CookieRedirect)},
			"format":     {"redirect"},
		}
		uri := "/auth/user/login?"
		if au.Uid == -1 {
			// sign up
			uri = "/auth/user/signup?"
		}
		ctx.Redirect(http.StatusFound, uri+data.Encode())
	}
}
*/
