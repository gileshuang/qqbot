package ws

// https://www.jianshu.com/p/11ded5e80cdf

import (
	"net"
	"net/http"
	"qqbot/lib/qblog"
	"time"

	"github.com/gorilla/websocket"
)

type WsServer struct {
	listener net.Listener
	addr     string
	upgrade  *websocket.Upgrader
}

func NewWsServer(wsaddr string) *WsServer {
	ws := new(WsServer)
	ws.addr = wsaddr
	ws.upgrade = &websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			if r.Method != "GET" {
				qblog.Log.Warning("method is not GET")
				return false
			}
			return true
		},
	}
	return ws
}

func (slf *WsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	qblog.Log.Debug("r.URL.Path:", r.URL.Path)
	conn, err := slf.upgrade.Upgrade(w, r, nil)
	if err != nil {
		qblog.Log.Error("websocket error:", err)
		return
	}
	qblog.Log.Info("client connect:", conn.RemoteAddr())
	go slf.connHandle(conn)
}

func (slf *WsServer) connHandle(conn *websocket.Conn) {
	hand := NewWsHander(conn)
	defer func() {
		hand.Close()
		conn.Close()
	}()
	go hand.Run()
	for {
		conn.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(10000)))
		_, msg, err := conn.ReadMessage()
		// 判断是不是超时
		if netErr, ok := err.(net.Error); ok {
			if netErr.Timeout() {
				continue
			}
		}
		if err != nil {
			qblog.Log.Warning("ReadMessage from websocket error:", err)
			qblog.Log.Info("close client connect:", conn.RemoteAddr())
			return
		}
		go hand.Read(msg)
	}
}

func (w *WsServer) Start() (err error) {
	w.listener, err = net.Listen("tcp", w.addr)
	if err != nil {
		qblog.Log.Warning("net listen error:", err)
		return
	}
	err = http.Serve(w.listener, w)
	if err != nil {
		qblog.Log.Warning("http serve error:", err)
		return
	}
	return nil
}
