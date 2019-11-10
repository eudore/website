package auth

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

type (
	MailSenderConfig struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Addr     string `json:"addr"`
		Subject  string `json:"subject"`
	}
	MailSender struct {
		auth    smtp.Auth
		from    string
		host    string
		subject string
		server  string
		addr    string
	}
)

// NewMailSender().Send("xxxxx@qq.com", "iiiiiiii")
func NewMailSender(config *MailSenderConfig) *MailSender {
	addr := strings.Split(config.Addr, ":")
	if config.Subject == "" {
		config.Subject = "eudore website"
	}
	return &MailSender{
		auth:    smtp.PlainAuth("", config.Username, config.Password, addr[0]),
		from:    config.Username,
		subject: config.Subject,
		server:  addr[0],
		addr:    config.Addr,
	}
}

func (sender *MailSender) Send(to string, message string) error {
	conn, err := tls.Dial("tcp", sender.addr, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, sender.server)
	if err != nil {
		return err
	}
	defer client.Close()

	if err = client.Auth(sender.auth); err != nil {
		return err
	}

	err = client.Mail(sender.from)
	if err != nil {
		return err
	}
	for _, addr := range strings.Split(to, ";") {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "To: %s\r\nFrom: eudore <%s>\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", to, sender.from, sender.subject, message)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return client.Quit()
}
