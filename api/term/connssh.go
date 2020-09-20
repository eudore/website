package term

import (
	// "errors"
	"time"

	"golang.org/x/crypto/ssh"
)

type (
	SSHServerConn struct {
		*SSHConn
	}
	SSHClientConn struct {
		*SSHConn
		client *ssh.Client
	}
	SSHConn struct {
		ssh.Channel
		Requests <-chan *ssh.Request
		Message  chan Message
	}
)

func NewSSHServerConn(newch ssh.NewChannel) (*SSHServerConn, error) {
	ch, reqs, err := newch.Accept()
	if err != nil {
		return nil, err
	}
	return &SSHServerConn{
		SSHConn: NewSSHConn(ch, reqs),
	}, nil
}

func NewSSHClientConn(host Host) (*SSHClientConn, error) {
	config := &ssh.ClientConfig{
		User:            host.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 30,
	}
	if host.Password != "" {
		config.Auth = []ssh.AuthMethod{ssh.Password(host.Password)}
	}
	if host.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(host.PrivateKey))
		if err != nil {
			return nil, err
		}
		config.Auth = append(config.Auth, ssh.PublicKeys(signer))
	}

	client, err := ssh.Dial("tcp", host.Addr, config)
	if err != nil {
		return nil, err
	}

	cchan, reqs, cerr := client.OpenChannel("session", nil)
	if cerr != nil {
		return nil, cerr
	}

	return &SSHClientConn{
		client:  client,
		SSHConn: NewSSHConn(cchan, reqs),
	}, nil
}

func (conn *SSHClientConn) Close() error {
	defer conn.client.Close()
	return conn.SSHConn.Close()
}

func NewSSHConn(ch ssh.Channel, reqs <-chan *ssh.Request) *SSHConn {
	conn := &SSHConn{
		Channel:  ch,
		Requests: reqs,
		Message:  make(chan Message),
	}
	go conn.runMessage()
	return conn
}

func (conn *SSHConn) runMessage() {
	for {
		req := <-conn.Requests
		if req == nil {
			close(conn.Message)
			return
		}
		conn.Message <- sshUnmarshal(req)
	}
}

func (conn *SSHConn) SendMessage(msg Message) error {
	if msg == nil {
		return nil
	}
	name := msg.MessageType()
	wantReply := true
	switch name {
	case "window-change", "signal":
		wantReply = false
	case "start":
		return nil
	case "error":
		return msg.(errorMessage).error
	}
	_, err := conn.Channel.SendRequest(name, wantReply, sshMarshal(msg))
	return err
}

func (conn *SSHConn) RecoMessage() <-chan Message {
	return conn.Message
}
