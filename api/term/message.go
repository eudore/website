package term

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh"
)

type (
	Message interface {
		MessageType() string
	}

	messageType struct {
		Type    string          `json:"type"`
		Message json.RawMessage `json:"message,omitempty"`
	}
	envMessage struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	ptyMessage struct {
		Term     string `json:"term"`
		Columns  uint32 `json:"columns"`
		Rows     uint32 `json:"rows"`
		Width    uint32 `json:"width"`
		Height   uint32 `json:"height"`
		Modelist string `json:"modelist"`
	}
	subsystemMessage struct {
		Subsystem string `json:"subsystem"`
	}
	windowChangeMessage struct {
		Columns uint32 `json:"columns"`
		Rows    uint32 `json:"rows"`
		Width   uint32 `json:"width"`
		Height  uint32 `json:"height"`
	}
	signalMessage struct {
		Signal string `json:"signal"`
	}
	execMessage struct {
		Command string `json:"command"`
	}
	shellMessage      struct{}
	exitStatusMessage struct{}
	errorMessage      struct {
		error
	}
	pingMessage  struct{}
	pongMessage  struct{}
)

func webUnmarshal(data []byte) Message {
	var req messageType
	err := json.Unmarshal(data, &req)
	if err != nil {
		return errorMessage{error: err}
	}

	var msg Message
	switch req.Type {
	case "env":
		msg = new(envMessage)
	case "pty-req":
		msg = new(ptyMessage)
	case "subsystem":
		msg = new(subsystemMessage)
	case "window-change":
		msg = new(windowChangeMessage)
	case "signal":
		msg = new(signalMessage)
	case "exec":
		msg = new(execMessage)
	case "shell":
		return new(shellMessage)
	case "exit-status":
		return new(exitStatusMessage)
	case "ping":
		return new(pingMessage)
	case "pong":
		return new(pongMessage)
	default:
		return errorMessage{
			error: fmt.Errorf("undinet request type %s", req.Type),
		}
	}
	err = json.Unmarshal(req.Message, msg)
	if err != nil {
		return errorMessage{error: err}
	}
	return msg
}

func webMarshal(msg Message) (data []byte, err error) {
	if msg.MessageType() == "error" {
		return nil, msg.(errorMessage).error
	}
	var req messageType
	req.Message, err = json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	req.Type = msg.MessageType()
	data, err = json.Marshal(req)
	return
}

type x11request struct {
	SingleConnection bool
	AuthProtocol     string
	AuthCookie       string
	ScreenNumber     uint32
}

func sshUnmarshal(req *ssh.Request) Message {
	if req == nil {
		return nil
	}
	var msg Message
	switch req.Type {
	case "env":
		msg = new(envMessage)
	case "x11-req":
		msg = new(x11request)
	case "pty-req":
		msg = new(ptyMessage)
	case "subsystem":
		msg = new(subsystemMessage)
	case "window-change":
		msg = new(windowChangeMessage)
	case "signal":
		msg = new(signalMessage)
	case "exec":
		msg = new(execMessage)
	case "shell":
		return new(shellMessage)
	case "exit-status":
		return new(exitStatusMessage)
	case "ping":
		return new(pingMessage)
	default:
		return errorMessage{
			error: fmt.Errorf("undinet request type %s", req.Type),
		}
	}
	err := ssh.Unmarshal(req.Payload, msg)
	if err != nil {
		return errorMessage{error: err}
	}
	if req.Type == "x11-req" {
		fmt.Printf("%#v\n", msg)
	}
	return msg
}

func sshMarshal(msg Message) []byte {
	switch msg.MessageType() {
	case "shell", "exit-status", "ping", "pong":
		return nil
	default:
		return ssh.Marshal(msg)
	}
}

func init() {
	ssh.Marshal(&x11request{
		SingleConnection: false,
		AuthProtocol:     "MIT-MAGIC-COOKIE-1",
		AuthCookie:       "76a00de971aeb7523d7e6ac95c793185",
		ScreenNumber:     0,
	})
}

func (msg envMessage) MessageType() string          { return "env" }
func (msg x11request) MessageType() string          { return "x11-req" }
func (msg ptyMessage) MessageType() string          { return "pty-req" }
func (msg subsystemMessage) MessageType() string    { return "subsystem" }
func (msg windowChangeMessage) MessageType() string { return "window-change" }
func (msg signalMessage) MessageType() string       { return "signal" }
func (msg execMessage) MessageType() string         { return "exec" }
func (msg shellMessage) MessageType() string        { return "shell" }
func (msg exitStatusMessage) MessageType() string   { return "exit-status" }
func (msg pingMessage) MessageType() string         { return "ping" }
func (msg pongMessage) MessageType() string         { return "pong" }
func (msg errorMessage) MessageType() string        { return "error" }
