package auth

type Config struct {
	Secrets  map[string]string `set:"secrets"`
	IconTemp string            `set:"icontemp"`
}

var iconTmp = "/tmp/wejass/icon/"
