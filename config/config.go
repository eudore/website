package config

import (
	"github.com/eudore/eudore"
	// eudoreserver "github.com/eudore/eudore/component/server/eudore"

	"github.com/eudore/website/handlers/auth"
)

type (
	// Config 定义website的全局配置。
	Config struct {
		Command   string                       `set:"command"`
		Pidfile   string                       `set:"pidfile"`
		Workdir   string                       `set:"workdir"`
		Enable    []string                     `set:"enable"`
		Mods      map[string]*Config           `set:"mods"`
		Keys      map[string]interface{}       `set:"keys" description:"keys"`
		Listeners []*eudore.ServerListenConfig `set:"listeners" json:"listeners"`
		Component *ComponentConfig             `set:"component"`

		Auth *auth.Config `set:"auth"`
	}
	// ComponentConfig 定义website使用的组件的配置。
	ComponentConfig struct {
		Logger *eudore.LoggerStdConfig `set:"logger"`
		Server *eudore.ServerConfigStd `set:"server"`
		// Eudoreserver eudoreserver.ServerConfig `set:"eudoreserver"`
		Notify map[string]string `set:"notify"`
	}
)

var globalConfig *Config

func init() {
	globalConfig = &Config{
		Keys: map[string]interface{}{
			// linux、win、docker下默认配置文件路径。
			"config": []string{
				"/root/go/src/github.com/eudore/website/config/config-eudore.json",
				"/root/go/src/github.com/eudore/website/config/config.json",
				"C:\\Users\\Administrator\\go\\src\\github.com\\eudore\\website\\config\\config-eudore.json",
				"C:\\Users\\Administrator\\go\\src\\github.com\\eudore\\website\\config\\config.json",
				"config.json",
			},
		},
	}
}

// GetConfig 获得全局单例配置。
func GetConfig() *Config {
	return globalConfig
}
