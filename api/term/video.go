package term

/*
PostgreSQL Begin
CREATE TABLE tb_term_video(
	"name" VARCHAR(64) NOT NULL,
	"user" VARCHAR(32) NOT NULL,
	"remoteaddr"VARCHAR(64) NOT NULL,
	"localaddr" VARCHAR(64) NOT NULL,
	"startstamp" TIMESTAMP  NOT NULL,
	"endstamp" TIMESTAMP NOT NULL,
	"savedir" VARCHAR(128) NOT NULL,
	"indexs" TEXT NOT NULL
);

PostgreSQL End
*/

import (
	"encoding/binary"
	// "encoding/json"
	// "database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
	// "github.com/eudore/eudore"
	// "github.com/eudore/website/internal/controller"
)

type (
	Video struct {
		mu         sync.Mutex
		length     int
		Name       string
		User       string
		RemoteAddr string
		LocalAddr  string
		Startstamp time.Time
		EndTime    time.Time
		SaveDir    string
		data       []VideoData
		Indexs     []VideoPart
	}
	VideoPart struct {
		Interval time.Duration
		Name     string
		Size     int
	}
	VideoData struct {
		Interval time.Duration
		Data     []byte
	}
	ConnVideoSave struct {
		Conn
		*Video
		EndHook func(*Video)
	}
	ConnVideoPlay struct {
	}
)

func (v *Video) AddData(data []byte) {
	v.mu.Lock()
	v.data = append(v.data, VideoData{
		Interval: time.Now().Sub(v.Startstamp),
		Data:     data,
	})
	v.length = v.length + len(data) + 8
	v.mu.Unlock()
	if v.length >= 32<<20 {
		v.SavePart()
	}
}

func (v *Video) SavePart() error {
	if len(v.data) == 0 {
		return nil
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	name := fmt.Sprintf("video-%s-%d-%d.bin", v.User, v.Startstamp.UnixNano(), len(v.Indexs))
	file, err := os.Create(filepath.Join(v.SaveDir, name))
	if err != nil {
		return err
	}
	go v.savePartWrite(file, v.data)
	v.Indexs = append(v.Indexs, VideoPart{
		Interval: v.data[0].Interval,
		Name:     name,
		Size:     v.length,
	})
	v.length = 0
	v.data = nil
	return nil
}

func (v *Video) savePartWrite(file io.WriteCloser, data []VideoData) {
	for _, i := range data {
		binary.Write(file, binary.LittleEndian, i.Interval)
		binary.Write(file, binary.LittleEndian, len(i.Data))
		file.Write(i.Data)
	}
	file.Close()
}

func NewConnVideoSave(conn Conn, fn func(*Video), video *Video) Conn {
	if video.Name == "" {
		video.Name = fmt.Sprintf("video-%s-%d", video.User, video.Startstamp.UnixNano())
	}
	return &ConnVideoSave{
		Conn:    conn,
		Video:   video,
		EndHook: fn,
	}
}

func (conn *ConnVideoSave) Write(data []byte) (n int, err error) {
	n, err = conn.Conn.Write(data)
	if err == nil {
		conn.Video.AddData(data[0:n])
	}
	return n, err
}

func (conn *ConnVideoSave) Close() error {
	fmt.Println("conn close", conn)
	conn.Video.SavePart()
	if conn.EndHook != nil && conn.Video.EndTime.IsZero() {
		conn.Video.EndTime = time.Now()
		conn.EndHook(conn.Video)
	}
	return conn.Conn.Close()
}

func NewConnVideoPlay() *ConnVideoPlay {
	return &ConnVideoPlay{}
}

func (conn *ConnVideoPlay) Run(sconn Conn) {

}
