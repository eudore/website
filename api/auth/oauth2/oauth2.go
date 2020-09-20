// Package oauth2 定义oauth2处理第三方登录
//
package oauth2

import (
	"errors"
	"math/rand"
	"net/http"

	"golang.org/x/oauth2"
)

// Ouath2 handle
type (
	// Config 定义oauth2配置对象
	Config = oauth2.Config
	// Oauth2 定义http Oauth2接口
	Oauth2 interface {
		// Set Ouath2 config
		Config(*oauth2.Config)
		// Get redirect Addr
		Redirect(string) string
		// Handle callback request
		Callback(*http.Request) (map[string]interface{}, string, error)
	}
)

var (
	// ErrOauthCode code exchange failed error
	ErrOauthCode = errors.New("Code exchange failed")
	// ErrUnknownOauth2 unknown oauth2 error
	ErrUnknownOauth2 = errors.New("unknow oauth2")
)

// 定义ouath2方式
const (
	Oauth2Eudore = "eudore"
	Oauth2Github = "github"
	Oauth2Google = "google"
	Oauth2Gitlab = "gitlab"
)

// Define auth source
var oauth2source [4]string = [4]string{"eudore", "github", "google", "gitlab"}

// NewOuath2 func Ouath2 factory,return oauth2 handle
func NewOuath2(name string) (Oauth2, error) {
	switch name {
	case Oauth2Github:
		return newOauth2Github(), nil
	case Oauth2Google:
		return newOauth2Google(), nil
	case Oauth2Gitlab:
		return newOauth2Gitlab(), nil
	default:
		return nil, ErrUnknownOauth2
	}
}

// GetRandomString 创建一个16为随机数
func GetRandomString() string {
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
