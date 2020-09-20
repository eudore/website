package term

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/eudore/eudore"
	"golang.org/x/crypto/scrypt"
	"golang.org/x/crypto/ssh"
)

type (
	Server struct {
		context.Context
		DB        *sql.DB
		SSHConfig *ssh.ServerConfig
	}
)

func NewServer(ctx context.Context, db *sql.DB) *Server {
	return &Server{
		Context: ctx,
		DB:      db,
		SSHConfig: &ssh.ServerConfig{
			NoClientAuth: false,
			PasswordCallback: func(conn ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
				// no check password
				return getMetadata(conn), nil
			},
			PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
				publicBytes, err := ioutil.ReadFile("/root/.ssh/id_rsa.pub")
				if err != nil {
					return nil, err
				}
				out, _, _, _, err := ssh.ParseAuthorizedKey(publicBytes)
				if err != nil {
					return nil, err
				}

				if out.Type() != key.Type() || !bytes.Equal(out.Marshal(), key.Marshal()) {
					return nil, fmt.Errorf("ssh: host key mismatch")
				}

				return getMetadata(conn), nil
			},
		},
	}
}

func getMetadata(conn ssh.ConnMetadata) *ssh.Permissions {
	return &ssh.Permissions{
		Extensions: map[string]string{
			"user":       conn.User(),
			"sessionid":  string(conn.SessionID()),
			"remoteAddr": conn.RemoteAddr().String(),
			"localAddr":  conn.LocalAddr().String(),
		},
	}
}

func (srv *Server) getHostById(name, id string) (host Host) {
	srv.DB.QueryRowContext(srv.Context, `SELECT name,protocol,addr,user,password,privatekey 
		FROM tb_term_host AS h JOIN tb_term_user_host AS u ON h.id=u.hostid 
		WHERE id=$1 AND u.userid = (SELECT id FROM tb_term_user WHERE name=$2)`, id, name).Scan(&host.Name, &host.Protocol, &host.Addr, &host.User, &host.Password, &host.PrivateKey)
	return
}

func (srv *Server) getHostListByName(name string) []Host {
	rows, err := srv.DB.QueryContext(srv.Context, "SELECT h.* FROM tb_term_host AS h JOIN tb_term_user_host AS u on h.id=u.hostid WHERE u.userid=(SELECT id FROM tb_term_user WHERE name=$1)", name)
	if err != nil {
		return nil
	}
	var hs []Host
	for rows.Next() {
		var host Host
		rows.Scan(&host.ID, &host.Name, &host.Protocol, &host.Addr, &host.User, &host.Password, &host.PrivateKey)
		hs = append(hs, host)
	}
	return hs
}

func (srv *Server) NewHandleHTTP() func(ctx eudore.Context) error {
	return func(ctx eudore.Context) error {
		sconn, err := NewWebscoketServerConn(ctx.Request(), ctx.Response())
		if err != nil {
			return err
		}

		cconn, err := srv.CreateClientConn(srv.getHostById(ctx.GetParam("UNAME"), ctx.GetQuery("hostid")))
		if err != nil {
			cconn, err = srv.SelectClientConn(sconn, ctx.GetParam("UNAME"))
		}
		if err != nil {
			return err
		}

		ProxyConn(sconn, cconn)
		return nil
	}
}

func (srv *Server) ListenAndServe(port string) error {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Printf("net.Listen failed: %v", err)
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("listen.Accept failed: %v", err)
			return err
		}
		go srv.serveConn(conn)
	}
}

func (srv *Server) serveConn(conn net.Conn) {
	srvconn, chans, reqs, err := ssh.NewServerConn(conn, srv.SSHConfig)
	if err != nil {
		log.Println("ssh srver failed to handshake", err)
		return
	}
	defer srvconn.Close()

	name := srvconn.Permissions.Extensions["user"]

	go ssh.DiscardRequests(reqs)

	for newch := range chans {
		fmt.Println(newch.ChannelType(), newch.ExtraData())
		sconn, err := NewSSHServerConn(newch)
		if err != nil {
			log.Println(err)
			continue
		}

		cconn, err := srv.SelectClientConn(sconn, name)
		if err != nil {
			log.Println(err)
			continue
		}

		vconn := NewConnVideoSave(sconn, srv.saveVideo, &Video{
			User:       srvconn.Permissions.Extensions["user"],
			RemoteAddr: srvconn.Permissions.Extensions["remoteAddr"],
			LocalAddr:  srvconn.Permissions.Extensions["remoteAddr"],
			Startstamp: time.Now(),
			SaveDir:    "/tmp",
		})
		ProxyConn(vconn, cconn)
	}
}

func (srv *Server) SelectClientConn(sconn Conn, name string) (Conn, error) {
	host, msgs := NewTermServer(sconn, name, srv.getHostListByName(name)).Run()
	if host.ID == 0 {
		return nil, fmt.Errorf("not select host")
	}

	cconn, err := srv.CreateClientConn(host)
	if err != nil {
		return nil, err
	}
	for _, msg := range msgs {
		cconn.SendMessage(msg)
	}
	return cconn, err
}

func (srv *Server) CreateClientConn(host Host) (Conn, error) {
	switch host.Protocol {
	case "websocket":
		return NewWebscoketClientConn(host)
	case "ssh":
		return NewSSHClientConn(host)
	case "docker-image":
		return NewDockerClientConn(host.Addr)
	case "local-bash":
		return NewBashClientConn(), nil
	default:
		return nil, fmt.Errorf("undeinfe host protocol %s", host.Protocol)
	}
}

func (srv *Server) AddHostKey(paths ...string) error {
	if len(paths) == 0 {
		paths = []string{"/root/.ssh/id_rsa"}
	}
	for _, path := range paths {
		err := srv.addHostKey(path)
		if err != nil {
			return err
		}
	}
	return nil
}

func (srv *Server) addHostKey(path string) error {
	privateBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return err
	}

	srv.SSHConfig.AddHostKey(private)
	return nil
}

func (srv *Server) saveVideo(v *Video) {
	indexs, _ := json.Marshal(v.Indexs)
	_, err := srv.DB.Exec(`INSERT INTO tb_term_video("name","user","remoteaddr","localaddr","startstamp","endstamp","savedir","indexs") VALUES($1,$2,$3,$4,$5,$6,$7,$8);`, v.Name, v.User, v.RemoteAddr, v.LocalAddr, v.Startstamp, v.EndTime, v.SaveDir, string(indexs))
	fmt.Println(err)
}

// newCheckPassword 函数调用数据库验证密码。
func newCheckPassword(db *sql.DB) func(conn ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
	return func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
		var pass string
		var salt []byte
		err := db.QueryRow("SELECT pass,salt FROM tb_auth_user_pass WHERE name=$1;", conn.User()).Scan(&pass, &salt)
		if err != nil {
			return nil, err
		}

		// 计算hash
		keys, err := scrypt.Key(password, salt, 16384, 8, 1, 32)
		if err != nil {
			return nil, err
		}

		// 验证密码
		if hex.EncodeToString(keys) != pass {
			return nil, fmt.Errorf("ssh auth password invalid")
		}

		return getMetadata(conn), nil
	}
}

func newCheckPublicKey(db *sql.DB) func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	return func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
		var publicBytes []byte
		err := db.QueryRow("SELECT publickey FROM tb_term_user WHERE name=$1;", conn.User()).Scan(&publicBytes)
		if err != nil {
			return nil, err
		}

		out, _, _, _, err := ssh.ParseAuthorizedKey(publicBytes)
		if err != nil {
			return nil, err
		}

		if out.Type() != key.Type() || !bytes.Equal(out.Marshal(), key.Marshal()) {
			return nil, fmt.Errorf("ssh: host key mismatch")
		}

		return getMetadata(conn), nil
	}
}
