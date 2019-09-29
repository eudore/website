package oauth2

import (
	"encoding/json"
	"fmt"
	"github.com/eudore/eudore"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"io/ioutil"
	"net/http"
	"net/url"
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

func (o *Oauth2GithubHandle) Callback(ctx eudore.Context) (map[string]interface{}, error) {
	// get code
	code := ctx.GetQuery("code")
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
		return nil, err
	}
	defer resp.Body.Close()

	var data = make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&data)
	return data, nil
}

func (o *Oauth2GithubHandle) GetUserId(data map[string]interface{}) string {
	return fmt.Sprint(int64(data["id"].(float64)))
}

/*
{
	"avatar_url": "https://avatars0.githubusercontent.com/u/30709860?v=4",
	"bio": null,
	"blog": "",
	"company": null,
	"created_at": "2017-08-04T01:36:35Z",
	"email": null,
	"events_url": "https://api.github.com/users/eudore/events{/privacy}",
	"followers": 3,
	"followers_url": "https://api.github.com/users/eudore/followers",
	"following": 0,
	"following_url": "https://api.github.com/users/eudore/following{/other_user}",
	"gists_url": "https://api.github.com/users/eudore/gists{/gist_id}",
	"gravatar_id": "",
	"hireable": null,
	"html_url": "https://github.com/eudore",
	"id": 30709860,
	"location": null,
	"login": "eudore",
	"name": null,
	"node_id": "MDQ6VXNlcjMwNzA5ODYw",
	"organizations_url": "https://api.github.com/users/eudore/orgs",
	"public_gists": 0,
	"public_repos": 4,
	"received_events_url": "https://api.github.com/users/eudore/received_events",
	"repos_url": "https://api.github.com/users/eudore/repos",
	"site_admin": false,
	"starred_url": "https://api.github.com/users/eudore/starred{/owner}{/repo}",
	"subscriptions_url": "https://api.github.com/users/eudore/subscriptions",
	"type": "User",
	"updated_at": "2019-06-28T11:12:43Z",
	"url": "https://api.github.com/users/eudore"
}
*/
