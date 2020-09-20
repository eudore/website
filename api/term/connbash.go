package term

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
)

type (
	BashClientConn struct {
		cannel    func()
		envs      []string
		pty       *os.File
		cond      *sync.Cond
		Message   chan Message
		ptyTerm   string
		ptyColums uint32
		ptyRows   uint32
	}
)

func handlebash(w http.ResponseWriter, r *http.Request) {
	sconn, err := NewWebscoketServerConn(r, w)
	if err != nil {
		return
	}
	cconn := NewBashClientConn()

	ProxyConn(sconn, cconn)
}

func NewBashClientConn() *BashClientConn {
	return &BashClientConn{
		cond:    sync.NewCond(new(sync.Mutex)),
		Message: make(chan Message),
	}
}

func (conn *BashClientConn) Read(data []byte) (n int, err error) {
	conn.cond.L.Lock()
	if conn.pty == nil {
		conn.cond.Wait()
	}
	conn.cond.L.Unlock()
	return conn.pty.Read(data)
}
func (conn *BashClientConn) Write(data []byte) (int, error) {
	conn.cond.L.Lock()
	if conn.pty == nil {
		conn.cond.Wait()
	}
	conn.cond.L.Unlock()
	return conn.pty.Write(data)
}
func (conn *BashClientConn) Close() error {
	if conn.Message != nil {
		close(conn.Message)
		conn.Message = nil
	}
	if conn.pty != nil {
		conn.cannel()
		return conn.pty.Close()
	}
	return nil
}
func (conn *BashClientConn) SendMessage(msg Message) error {
	fmt.Println("BashClientConn", msg)
	switch val := msg.(type) {
	case *envMessage:
		conn.envs = append(conn.envs, val.Name+"="+val.Value)
	case *ptyMessage:
		conn.ptyTerm = val.Term
		conn.ptyColums = val.Columns
		conn.ptyRows = val.Rows
	case *shellMessage:
		return conn.Shell()
	case *windowChangeMessage:
		return conn.ResizeTerminal(val.Rows, val.Columns)
	case *exitStatusMessage:
		return conn.Close()
	case *pingMessage:
		conn.Message <- pongMessage{}
	case *errorMessage:
		return val.error
	}
	return nil
}

func (conn *BashClientConn) RecoMessage() <-chan Message {
	return conn.Message
}

func (conn *BashClientConn) Shell() (err error) {
	if conn.pty != nil {
		return nil
	}
	runtime.LockOSThread()
	ctx, cannel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "bash")
	conn.pty, err = pty.Start(cmd)
	if err != nil {
		return err
	}

	conn.cond.Broadcast()
	conn.cannel = cannel
	conn.ResizeTerminal(conn.ptyRows, conn.ptyColums)
	return
}

func (conn *BashClientConn) ResizeTerminal(row uint32, col uint32) error {
	if conn.pty == nil || row < 1 || col < 1 {
		return nil
	}
	window := struct {
		row uint16
		col uint16
		x   uint16
		y   uint16
	}{
		uint16(row),
		uint16(col),
		0,
		0,
	}
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		conn.pty.Fd(),
		syscall.TIOCSWINSZ,
		uintptr(unsafe.Pointer(&window)),
	)
	if errno != 0 {
		return errno
	}
	return nil
}
