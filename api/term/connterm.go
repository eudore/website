package term

import (
	"bytes"
	"fmt"
	"log"
)

type (
	ConnTermServer struct {
		Conn
		buffer   *bytes.Buffer
		name     string
		pos      int
		hosts    []Host
		messages []Message
		mod      int
		input    string
		Columns  uint32
		Rows     uint32
	}
)

var (
	keysUp    = []byte{27, 91, 65}
	keysDown  = []byte{27, 91, 66}
	keysEnter = []byte{13}
)

func NewTermServer(rw Conn, name string, hosts []Host) *ConnTermServer {
	return &ConnTermServer{
		Conn:     rw,
		buffer:   bytes.NewBuffer(nil),
		name:     name,
		pos:      0,
		hosts:    hosts,
		messages: make([]Message, 0, 5),
	}
}

func (conn *ConnTermServer) Println(args ...interface{}) {
	fmt.Fprint(conn.buffer, args...)
	fmt.Fprint(conn.buffer, "\r\n")
}
func (conn *ConnTermServer) Printf(format string, args ...interface{}) {
	fmt.Fprintf(conn.buffer, format, args...)
	fmt.Fprint(conn.buffer, "\r\n")
}
func (conn *ConnTermServer) Set(cmd string) {
	fmt.Fprint(conn.buffer, cmd)
}
func (conn *ConnTermServer) Clean() {
	fmt.Fprint(conn.buffer, "\033\143")
	// fmt.Fprint(conn.buffer, "\033[2J")
}

func (conn *ConnTermServer) Flush() {
	conn.Write(conn.buffer.Bytes())
	conn.buffer.Reset()
}

func (conn *ConnTermServer) Run() (Host, []Message) {
	done := make(chan struct{})
	defer close(done)
	go func(done chan struct{}) {
		ch := conn.RecoMessage()
		for {
			select {
			case msg := <-ch:
				conn.messages = append(conn.messages, msg)
				switch val := msg.(type){
				case *windowChangeMessage:
					conn.Columns = val.Columns
					conn.Rows = val.Rows
				case *ptyMessage:
					conn.Columns = val.Columns
					conn.Rows = val.Rows
				}
			case <-done:
				return
			}
		}
	}(done)

	body := make([]byte, 1024)
	for {
		conn.draw()
		n, err := conn.Conn.Read(body)
		if err != nil {
			log.Println(err)
			return Host{}, conn.messages
		}

		log.Println(body[:n], string(body[:n]))
		switch conn.mod {
		case 1:
			conn.handleCommand(body[:n])
		default:
			conn.handleNormal(body[:n])
		}
		if conn.mod == -1 {
			return conn.hosts[conn.pos], conn.messages
		}
	}
	return Host{}, conn.messages
}

func (conn *ConnTermServer) draw() {
	conn.Clean()
	conn.Printf("Welcome %s to eudore web shell !", conn.name)
	conn.Println()
	conn.Println("please select login host:")
	for i, host := range conn.hosts {
		if i == conn.pos {
			conn.Set("\033[4m")
			conn.Println(host.Format("%s\t %s\t %s"))
			conn.Set("\033[0m")
		} else {
			conn.Println(host.Format("%s\t %s\t %s"))
		}
	}

	if conn.mod == 1 {
		conn.Set(fmt.Sprintf("\033[%d;0H", conn.Rows))
		conn.Set(conn.input)
	} else {
		conn.Set("\033[?25l")
	}
	conn.Flush()
}

func (conn *ConnTermServer) handleNormal(data []byte) {
	switch {
	case bytes.Equal(data, keysUp) && conn.pos > 0:
		conn.pos--
	case bytes.Equal(data, keysDown) && conn.pos < len(conn.hosts)-1:
		conn.pos++
	case bytes.Equal(data, keysEnter):
		fmt.Println(conn.hosts[conn.pos], conn.messages)
		// return conn.hosts[conn.pos], conn.messages
		conn.mod = -1
	case bytes.Equal(data, []byte{58}):
		conn.mod = 1
		conn.handleCommand(data)
	}
}

func (conn *ConnTermServer) handleCommand(data []byte) {
	switch {
	case bytes.Equal(data, []byte{27}):
		conn.mod = 0
		conn.input = ""
	case bytes.Equal(data, []byte{127}):
		if conn.input != "" {
			conn.input = conn.input[:len(conn.input)-1]
		}
	case bytes.Equal(data, []byte{13}):
		fmt.Println("cmd:", conn.input)
		conn.input = ""
		conn.mod = 0
	default:
		conn.input += string(data)
	}
}
