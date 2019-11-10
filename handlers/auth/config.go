package auth

type (
	Config struct {
		Secrets  map[string]string `json:"secrets" set:"secrets"`
		IconTemp string            `json:"icontemp" set:"icontemp"`
		Sender   SenderConfig      `json:"sender" set:"sender"`
	}
	SenderConfig struct {
		Mail MailSenderConfig `json:"mail" set:"mail"`
	}
)
