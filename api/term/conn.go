package term

import (
	"fmt"
	"io"
	"log"
	"reflect"
)

type (
	Conn interface {
		io.ReadWriteCloser
		SendMessage(Message) error
		RecoMessage() <-chan Message
	}
)

func ProxyConn(sconn, cconn Conn) {
	name := fmt.Sprintf("%v %v", reflect.TypeOf(sconn).Elem(), reflect.TypeOf(cconn).Elem())
	go func() {
		log.Println(name+"ProxyConn start go", 1)
		n, err := io.Copy(sconn, cconn)
		log.Println(name+"ProxyConn close go", 1, n, err)
		sconn.Close()
		cconn.Close()
	}()
	go func() {
		log.Println(name+"ProxyConn start go", 2)
		n, err := io.Copy(cconn, sconn)
		log.Println(name+"ProxyConn close go", 2, n, err)
		sconn.Close()
		cconn.Close()
	}()
	go func() {
		log.Println(name+"ProxyConn start go", 3)
		defer log.Println(name+"ProxyConn close go", 3)
		var msg Message
		var err error
		sch := sconn.RecoMessage()
		cch := cconn.RecoMessage()
		for {
			select {
			case msg = <-sch:
				err = cconn.SendMessage(msg)
			case msg = <-cch:
				err = sconn.SendMessage(msg)
			}
			if msg == nil {
				return
			}
			log.Println("ProxyConn cconn Message:", msg.MessageType(), msg)
			if err != nil {
				log.Println(err)
			}
		}
	}()
}
