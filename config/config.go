package config

import (
	"github.com/eudore/eudore"
	"sync"
)

type (
	// Config 定义website的全局配置。
	Config struct {
		sync.RWMutex
		Command   string                      `json:"command" alias:"command"`
		Pidfile   string                      `json:"pidfile" alias:"pidfile"`
		Workdir   string                      `json:"workdir" alias:"workdir"`
		Enable    []string                    `json:"enable" alias:"enable"`
		Mods      map[string]*Config          `json:"mods" alias:"mods"`
		Keys      map[string]interface{}      `json:"keys" alias:"keys"`
		Listeners []eudore.ServerListenConfig `json:"listeners" alias:"listeners"`
		Component *ComponentConfig            `json:"component" alias:"component"`

		Auth *AuthConfig `json:"auth" alias:"auth"`
		Note *NoteConfig `json:"note" alias:"note"`
		Term *TermConfig `json:"term" alias:"term"`
	}
	// ComponentConfig 定义website使用的组件的配置。
	ComponentConfig struct {
		DB     DBConfig                `json:"db" alias:"db"`
		Logger *eudore.LoggerStdConfig `json:"logger" alias:"logger"`
		Server *eudore.ServerStdConfig `json:"server" alias:"server"`
		Notify map[string]string       `json:"notify" alias:"notify"`
		Pprof  *PprofConfig            `json:"pprof" alias:"pprof"`
		Black  map[string]bool         `json:"black" alias:"black"`
	}
	DBConfig struct {
		Driver string `json:"driver" alias:"driver"`
		Config string `json:"config" alias:"config"`
	}
	PprofConfig struct {
		Godoc     string            `json:"godoc" alias:"godoc"`
		BasicAuth map[string]string `json:"basicauth" alias:"basicauth"`
	}

	AuthConfig struct {
		Secrets  map[string]string `json:"secrets" alias:"secrets"`
		IconTemp string            `json:"icontemp" alias:"icontemp"`
		Sender   MailSenderConfig  `json:"sender" alias:"sender"`
	}
	MailSenderConfig struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Addr     string `json:"addr"`
		Subject  string `json:"subject"`
	}

	NoteConfig struct {
		Gitpath string `json:"gitpath" alias:"gitpath"`
		Workdir string `json:"workdir" alias:"workdir"`
	}

	TermConfig struct {
		Addr string `alias:"addr" json:"addr"`
	}
)

func New() *Config {
	return &Config{
		Keys: map[string]interface{}{
			// linux、win、docker下默认配置文件路径。
			"config": []string{
				"config/config-eudore.json", // 使用的配置
				"config/config.json",        // 示范的配置
			},
		},
	}

}
