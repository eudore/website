package oauth2

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/gitlab"
)

type Oauth2GitlabHandle struct {
	config *oauth2.Config
}

func newOauth2Gitlab() Oauth2 {
	return &Oauth2GitlabHandle{
		config: &oauth2.Config{
			Scopes:   []string{"read_user"},
			Endpoint: gitlab.Endpoint,
		},
	}
}

func (o *Oauth2GitlabHandle) Config(config *oauth2.Config) {
	joinConfig(o.config, config)
}

func (o *Oauth2GitlabHandle) Redirect(stats string) string {
	return o.config.AuthCodeURL(stats)
}

func (o *Oauth2GitlabHandle) Callback(req *http.Request) (map[string]interface{}, string, error) {
	// get code
	code := req.FormValue("code")
	token, err := o.config.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, "", ErrOauthCode
	}
	// get user info
	response, err := http.Get("https://gitlab.com/api/v4/user?access_token=" + token.AccessToken)
	defer response.Body.Close()

	var data = make(map[string]interface{})
	err = json.NewDecoder(response.Body).Decode(&data)
	return data, fmt.Sprint(int64(data["id"].(float64))), nil
}
