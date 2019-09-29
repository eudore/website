package config

import (
	"github.com/eudore/eudore"
	// "github.com/eudore/eudore/component/eslogger"
	eudoreserver "github.com/eudore/eudore/component/server/eudore"
	//	fastserver "github.com/eudore/eudore/component/server/fasthttp"

	"github.com/eudore/website/handlers/auth"
)

type (
	Config struct {
		Command   string                       `set:"command"`
		Pidfile   string                       `set:"pidfile"`
		Workdir   string                       `set:"workdir"`
		Enable    []string                     `set:"enable"`
		Mods      map[string]*Config           `set:"mods"`
		Keys      map[string]interface{}       `set:"keys" description:"keys"`
		Listeners []*eudore.ServerListenConfig `set:"listeners" json:"listeners"`
		Component *ComponentConfig             `set:"component"`
		Proxy     *ProxyConfig                 `set:"proxy" json:"proxy"`

		Auth *auth.Config `set:"auth"`
		File *ConfigFile
		Note *ConfigNote
		Api  *ConfigApi
	}
	ProxyConfig struct {
		Backend map[string]string `set:"backend" json:"backend"`
		Routes  map[string]string `set:"routes" json:"routes"`
	}
	// logger router server
	ComponentConfig struct {
		// Svclog		eudore.ConfigMap
		// MultiLogger		*eudore.LoggerMultiConfig
		Logger *eudore.LoggerStdConfig `set:"logger"`
		// Eslogger		*eslogger.LoggerConfig
		// Multiserver		*eudore.ServerMultiConfig
		Server       *eudore.ServerConfigStd   `set:"server"`
		Eudoreserver eudoreserver.ServerConfig `set:"eudoreserver"`
		Notify       map[string]string         `set:"notify"`
		//		Fastserver		*fastserver.ServerConfig	`set:"fastserver"`
	}
	// ConfigMiddleware struct {

	// }
	ConfigAuth struct {
	}
	ConfigFile struct {
	}
	ConfigNote struct {
	}
	ConfigApi struct {
		Default string
		List    []string
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
				"C:\\Users\\Administrator\\go\\src\\github.com\\eudore\\website\\config\\config.json",
				"config.json",
			},
		},
	}
}

func GetConfig() *Config {
	return globalConfig
}
