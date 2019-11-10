package oauth2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type Oauth2GithubHandle struct {
	config *oauth2.Config
}

func newOauth2Github() Oauth2 {
	return &Oauth2GithubHandle{
		config: &oauth2.Config{
			Scopes:   []string{"user:email"},
			Endpoint: github.Endpoint,
		},
	}
}

func (o *Oauth2GithubHandle) Config(config *oauth2.Config) {
	joinConfig(o.config, config)
}

func (o *Oauth2GithubHandle) Redirect(stats string) string {
	return o.config.AuthCodeURL(stats)
}

func (o *Oauth2GithubHandle) Callback(req *http.Request) (map[string]interface{}, string, error) {
	errstr := req.FormValue("error")
	if errstr != "" {
		return nil, "", fmt.Errorf(errstr)
	}
	// get code
	code := req.FormValue("code")
	response, _ := http.PostForm("https://github.com/login/oauth/access_token", url.Values{
		"client_id":     {o.config.ClientID},
		"client_secret": {o.config.ClientSecret},
		"code":          {code},
	})
	defer response.Body.Close()

	// get user info
	contents, _ := ioutil.ReadAll(response.Body)
	resp, err := http.Get("https://api.github.com/user?" + string(contents))
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	var data = make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&data)
	return data, fmt.Sprint(int64(data["id"].(float64))), err
}
