package chat

import (
	// "context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/website/internal/controller"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

/*
PostgreSQL Begin
CREATE TABLE tb_chat_message(
	"sendid" INTEGER,
	"receid" INTEGER,
	"status" INTEGER DEFAULT 0,
	"message" TEXT,
	"time" TIMESTAMP  DEFAULT (now())
);

PostgreSQL End
*/

type (
	// MessageController 定义消息控制器。
	MessageController struct {
		Logger eudore.Logger
		Hub    *Hub
		userid int
		controller.ControllerWebsite
	}
	// Message 定义一条消息格式
	Message struct {
		Sendid  int       `json:"sendid"`
		Receid  int       `json:"receid"`
		Message string    `json:"message"`
		Time    time.Time `json:"time"`
	}
	// Hub 定义客户端中心。
	Hub struct {
		Clients map[int]*Client
		app     *eudore.App
		db      *sql.DB
		pool    sync.Pool
	}
	// Client 定义一个用户客户端
	Client struct {
		hub    *Hub
		Conns  []net.Conn
		userid int
		mu     sync.Mutex
		send   chan []byte
	}
)

// NewMessageController 方法创建一个消息控制器。
func NewMessageController(app *eudore.App, db *sql.DB) *MessageController {
	ctl := &MessageController{
		Logger: app.Logger,
		Hub: &Hub{
			Clients: make(map[int]*Client),
			app:     app,
			db:      db,
		},
		ControllerWebsite: *controller.NewControllerWejass(db),
	}
	ctl.Hub.pool.New = func() interface{} {
		return &Client{
			hub: ctl.Hub,
		}
	}
	return ctl
}

// Init 方法初始化消息控制器，检查用户id。
func (ctl *MessageController) Init(ctx eudore.Context) error {
	ctl.userid = eudore.GetStringInt(ctx.GetParam("UID"))
	if ctl.userid == 0 {
		return fmt.Errorf("user not login")
	}
	return ctl.ControllerWebsite.Init(ctx)
}

// GetList 方法返回用户最近100条消息。
func (ctl *MessageController) GetList() (interface{}, error) {
	return ctl.QueryRows("SELECT * FROM tb_chat_message WHERE sendid=$1 or receid=$2 limit 100", ctl.userid, ctl.userid)
}

// GetListUsesr 方法返回全部用户id。
func (ctl *MessageController) GetListUsesr() interface{} {
	users := make([]int, 0, len(ctl.Hub.Clients))
	for k := range ctl.Hub.Clients {
		fmt.Println(k)
		users = append(users, k)
	}
	return users
}

// GetListClients 方法返回全部客户端的消息。
func (ctl *MessageController) GetListClients() interface{} {
	return ctl.Hub.Clients
}

// PutNew 方法让用户发送一条消息。
func (ctl *MessageController) PutNew() error {
	var msg Message
	if err := ctl.Bind(&msg); err != nil {
		return err
	}
	msg.Sendid = ctl.userid

	return ctl.Hub.AddMessage(msg)
}

// GetConnect 方法使控制器hub连接一个新的websocket连接。
func (ctl *MessageController) GetConnect() error {
	conn, _, _, err := ws.UpgradeHTTP(ctl.Request(), ctl.Response())
	if err != nil {
		ctl.Error(err)
		return err
	}

	client := ctl.Hub.GetClient(ctl.userid)
	go client.HandleConn(conn)
	return nil
}

// GetClient 方法使hub根据userid创建一个客户端，如果存在返回存在的客户端。
func (hub *Hub) GetClient(userid int) *Client {
	client, ok := hub.Clients[userid]
	if !ok {
		client = hub.pool.Get().(*Client)
		client.userid = userid
		client.send = make(chan []byte)
		go client.Run()
		hub.Clients[userid] = client
	}
	return client
}

// PutClient 方法关闭一个客户端并回收。
func (hub *Hub) PutClient(client *Client) {
	close(client.send)
	delete(hub.Clients, client.userid)
	hub.pool.Put(client)
}

// AddMessage 方法使hub处理一个消息，数据库记录并发送给客户端。
func (hub *Hub) AddMessage(msg Message) error {
	_, err := hub.db.Exec("INSERT INTO tb_chat_message(sendid, receid, message) VALUES ($1,$2,$3)", msg.Sendid, msg.Receid, msg.Message)
	hub.Clients[msg.Sendid].WriteJSON(msg)
	hub.Clients[msg.Receid].WriteJSON(msg)
	return err
}

// Run 方法启动一个客户端，处理发送文本消息和连接ping桢。
func (client *Client) Run() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case msg, ok := <-client.send:
			if !ok {
				return
			}
			client.WriteMessage(ws.OpText, msg)
		case <-ticker.C:
			client.WriteMessage(ws.OpPing, nil)
		}
	}
}

// HandleConn 方法使客户端处理一个连接。
func (client *Client) HandleConn(conn net.Conn) {
	client.mu.Lock()
	client.Conns = append(client.Conns, conn)
	client.mu.Unlock()
	client.Infof("client %d add new conn %s", client.userid, conn.RemoteAddr().String())

	var msg Message
	for {
		body, _, err := wsutil.ReadClientData(conn)
		if err != nil {
			break
		}

		err = json.Unmarshal(body, &msg)
		if err != nil {
			client.Errorf("wsutil json unmarshal message err: %v", err)
			break
		}
		msg.Sendid = client.userid
		msg.Time = time.Now()

		client.hub.AddMessage(msg)
	}
	client.Remove(conn)
}

// Remove 方法异常客户端的一个连接并关闭，如果客户端连接数量为零，则关闭客户端。
func (client *Client) Remove(conn net.Conn) {
	conn.Close()
	client.mu.Lock()
	defer client.mu.Unlock()
	client.Infof("client %d remove conn %s", client.userid, conn.RemoteAddr().String())

	for i, con := range client.Conns {
		if con == conn {
			client.Conns = append(client.Conns[:i], client.Conns[i+1:]...)

			if len(client.Conns) == 0 {
				client.hub.PutClient(client)
			}
			return
		}
	}
}

// WriteMessage 方法给客户端发送指定消息。
func (client *Client) WriteMessage(op ws.OpCode, p []byte) (err error) {
	for _, conn := range client.Conns {
		err := wsutil.WriteServerMessage(conn, op, p)
		if err != nil {
			client.Remove(conn)
			return err
		}
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	}
	return nil
}

// WriteJSON 方法给客户端发送指定json。
func (client *Client) WriteJSON(v interface{}) error {
	if client == nil {
		return nil
	}

	msg, err := json.Marshal(v)
	if err != nil {
		return err
	}

	client.send <- msg
	return nil
}

// Info 方法输出Info级别日志。
func (client *Client) Info(args ...interface{}) {
	client.hub.app.Logger.WithField("userid", client.userid).Info(args...)
}

// Infof 方法格式化输出Info级别日志。
func (client *Client) Infof(format string, args ...interface{}) {
	client.hub.app.Logger.WithField("userid", client.userid).Infof(format, args...)
}

// Error 方法输出Error级别日志。
func (client *Client) Error(args ...interface{}) {
	client.hub.app.Logger.WithField("userid", client.userid).Error(args...)
}

// Errorf 方法格式化输出Error级别日志。
func (client *Client) Errorf(format string, args ...interface{}) {
	client.hub.app.Logger.WithField("userid", client.userid).Errorf(format, args...)
}

// MarshalJSON 方法返回json序列化内容。
func (client *Client) MarshalJSON() ([]byte, error) {
	conns := make([]string, len(client.Conns))
	for i, conn := range client.Conns {
		conns[i] = conn.RemoteAddr().String()
	}
	return json.Marshal(conns)
}
