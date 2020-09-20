package term

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type (
	WebscoketServerConn struct {
		*WebscoketConn
		// pty chan Pty
	}
	WebscoketClientConn struct {
		*WebscoketConn
	}
	WebscoketConn struct {
		net.Conn
		State      ws.State
		LastOpCode ws.OpCode
		Message    chan Message
		TextData   *bytes.Buffer
		reader     *wsutil.Reader
		writer     *wsutil.Writer
	}
)

func NewWebscoketServerConn(r *http.Request, w http.ResponseWriter) (*WebscoketServerConn, error) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return nil, err
	}
	return &WebscoketServerConn{
		WebscoketConn: NewWebscoketConn(conn, ws.StateServerSide),
	}, nil
}


func NewWebscoketClientConn(host Host) (*WebscoketClientConn, error) {
	conn, _, _, err := ws.Dial(context.Background(), fmt.Sprintf("ws://%s/bash", host.Addr))
	if err != nil {
		return nil, err
	}
	return &WebscoketClientConn{
		WebscoketConn: NewWebscoketConn(conn, ws.StateClientSide),
	}, nil
}

func NewWebscoketConn(conn net.Conn, state ws.State) *WebscoketConn {
	reader := wsutil.NewReader(conn, state)
	writer := wsutil.NewWriter(conn, state, ws.OpBinary)
	return &WebscoketConn{
		Conn:     conn,
		State:    state,
		Message:  make(chan Message),
		TextData: bytes.NewBuffer(nil),
		reader:   reader,
		writer:   writer,
	}
}

func (conn *WebscoketConn) Read(data []byte) (n int, err error) {
	for {
		header, err := conn.reader.NextFrame()
		// fmt.Println(header, err)
		if err == io.EOF {
			return 0, err
		}

		if header.OpCode == ws.OpContinuation {
			header.OpCode = conn.LastOpCode
		} else {
			if !header.Fin {
				conn.LastOpCode = header.OpCode
			}
		}
		switch header.OpCode {
		case ws.OpText:
			n, err := conn.reader.Read(data)
			conn.TextData.Write(data[0:n])
			if err == io.ErrUnexpectedEOF || err == io.EOF {
				err = nil
			}
			if header.Fin {
				msg := webUnmarshal(conn.TextData.Bytes())
				conn.TextData.Reset()
				// fmt.Println("WebscoketConn read", conn.State,msg)
				conn.Message <- msg
			}
			return 0, err
		case ws.OpBinary:
			n, err := conn.reader.Read(data)
			if err == io.ErrUnexpectedEOF || err == io.EOF {
				err = nil
			}
			if err == wsutil.ErrNoFrameAdvance {
				err = io.EOF
			}
			return n, err
		case ws.OpClose:
			conn.Close()
			return 0, io.EOF
		case ws.OpPing:
			wsutil.WriteMessage(conn.Conn, conn.State, ws.OpPong, nil)
			continue
		case ws.OpPong:
			io.CopyN(ioutil.Discard, conn.reader, header.Length)
			continue
		default:
			return 0, ws.ErrProtocolOpCodeReserved
		}
	}
}

func (conn *WebscoketConn) Write(data []byte) (int, error) {
	defer conn.writer.Flush()
	return conn.writer.Write(data)
}

func (conn *WebscoketConn) Close() error {
	if conn.Message != nil {
		close(conn.Message)
		conn.Message = nil
	}
	return conn.Conn.Close()
}

func (conn *WebscoketConn) SendMessage(msg Message) error {
	if msg == nil {
		return nil
	}
	// fmt.Println("WebscoketConn SendMessage", conn.State, msg)
	data, err := webMarshal(msg)
	if err != nil {
		return err
	}
	return wsutil.WriteMessage(conn.Conn, conn.State, ws.OpText, data)
}

func (conn *WebscoketConn) RecoMessage() <- chan Message {
	return conn.Message
}
