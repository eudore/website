package oauth2

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Oauth2GoogleHandle struct {
	config *oauth2.Config
}

func newOauth2Google() Oauth2 {
	return &Oauth2GoogleHandle{
		config: &oauth2.Config{
			Scopes: []string{"https://www.googleapis.com/auth/userinfo.profile",
				"https://www.googleapis.com/auth/userinfo.email"},
			Endpoint: google.Endpoint,
		},
	}
}

func (o *Oauth2GoogleHandle) Config(config *oauth2.Config) {
	joinConfig(o.config, config)
}

func (o *Oauth2GoogleHandle) Redirect(stats string) string {
	return o.config.AuthCodeURL(stats)
}

func (o *Oauth2GoogleHandle) Callback(req *http.Request) (map[string]interface{}, string, error) {
	code := req.FormValue("code")
	token, err := o.config.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, "", ErrOauthCode
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	defer response.Body.Close()

	var data = make(map[string]interface{})
	err = json.NewDecoder(response.Body).Decode(&data)
	return data, fmt.Sprint(int64(data["id"].(float64))), nil
}
