package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"sync"
	"time"
)

type websocketClientMgr struct {
	conn        *websocket.Conn
	addr        *string
	sendMsgChan chan string
	recvMsgChan chan string
	isAlive     bool
	timeout     int
}

// NewWsClientManager 构造函数
func NewWsClientManager(addrIp, addrPort string, timeout int) *websocketClientMgr {
	addrString := fmt.Sprint(addrIp, ":", addrPort)
	var sendChan = make(chan string, 10)
	var recvChan = make(chan string, 10)
	var conn *websocket.Conn
	return &websocketClientMgr{
		addr:        &addrString,
		conn:        conn,
		sendMsgChan: sendChan,
		recvMsgChan: recvChan,
		isAlive:     false,
		timeout:     timeout,
	}
}

// 链接服务端
func (wsc *websocketClientMgr) dail() {
	var err error
	u := url.URL{Scheme: "ws", Host: *wsc.addr}
	log.Printf("Connecting to %s", u.String())
	wsc.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	wsc.isAlive = true
	log.Printf("Successfully connected to %s", u.String())
}

// 发送消息
func (wsc *websocketClientMgr) sendMsgThread() {
	go func() {
		for {
			// msg := <-wsc.sendMsgChan
			msg := "{\"action\":\"send_group_msg\",\"params\":{\"group_id\":\"798131705\",\"message\":\"Hello World!\"}}"
			err := wsc.conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Println("SendError:", err)
				continue
			}
		}
	}()
}

// 读取消息
func (wsc *websocketClientMgr) readMsgThread() {
	go func() {
		for {
			if wsc.conn != nil {
				_, message, err := wsc.conn.ReadMessage()
				if err != nil {
					log.Println("ReceivingError:", err)
					wsc.isAlive = false
					// 出现错误，退出读取，尝试重连
					break
				}
				log.Printf("Received: %s", message)
				// 需要读取数据，不然会阻塞
				wsc.recvMsgChan <- string(message)
			}
		}
	}()
}

// 开启服务并重连
func (wsc *websocketClientMgr) start() {
	for {
		if wsc.isAlive == false {
			wsc.dail()
			wsc.sendMsgThread()
			wsc.readMsgThread()
		}
		time.Sleep(time.Second * time.Duration(wsc.timeout))
	}
}

func main() {
	wsc := NewWsClientManager("127.0.0.1", "6700", 10)
	wsc.start()
	var w1 sync.WaitGroup
	w1.Add(1)
	w1.Wait()
}