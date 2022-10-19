package ws

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
			// if r.URL.Path != "/ws" {
			// 	fmt.Println("client request path error")
			// 	log.Println("DEBUG: client request path error")
			// 	return false
			// }
			return true
		},
	}
	return ws
}

func (slf *WsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	qblog.Log.Debug("r.URL.Path:", r.URL.Path)
	// if r.URL.Path != "/ws" {
	// 	httpCode := http.StatusInternalServerError
	// 	reasePhrase := http.StatusText(httpCode)
	// 	fmt.Println("client request path error ", reasePhrase)
	// 	http.Error(w, reasePhrase, httpCode)
	// 	return
	// }
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
				// warninglog("ReadMessage timeout remote: %v\n", conn.RemoteAddr())
				// return
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

// func (slf *WsServer) connHandleTest(conn *websocket.Conn) {
// 	defer func() {
// 		conn.Close()
// 	}()
// 	stopCh := make(chan int)
// 	go slf.send(conn, stopCh)
// 	for {
// 		conn.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(5000)))
// 		_, msg, err := conn.ReadMessage()
// 		if err != nil {
// 			close(stopCh)
// 			// 判断是不是超时
// 			if netErr, ok := err.(net.Error); ok {
// 				if netErr.Timeout() {
// 					fmt.Printf("ReadMessage timeout remote: %v\n", conn.RemoteAddr())
// 					return
// 				}
// 			}
// 			// 其他错误，如果是 1001 和 1000 就不打印日志
// 			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
// 				fmt.Printf("ReadMessage other remote:%v error: %v \n", conn.RemoteAddr(), err)
// 			}
// 			return
// 		}
// 		fmt.Println("收到消息：", string(msg))
// 	}
// }

// // send10() 由 send 调用，测试一次性发送 10万条数据给 client, 如果不使用 time.Sleep browser 过了超时时间会断开
// func (slf *WsServer) send10(conn *websocket.Conn) {
// 	for i := 0; i < 100000; i++ {
// 		data := fmt.Sprintf("hello websocket test from server %v", time.Now().UnixNano())
// 		err := conn.WriteMessage(1, []byte(data))
// 		if err != nil {
// 			fmt.Println("send msg faild ", err)
// 			return
// 		}
// 		// time.Sleep(time.Millisecond * 1)
// 	}
// }

// // send() 函数测试发送数据
// func (slf *WsServer) send(conn *websocket.Conn, stopCh chan int) {
// 	slf.send10(conn)
// 	for {
// 		select {
// 		case <-stopCh:
// 			fmt.Println("connect closed")
// 			return
// 		case <-time.After(time.Second * 1):
// 			data := fmt.Sprintf("hello websocket test from server %v", time.Now().UnixNano())
// 			err := conn.WriteMessage(1, []byte(data))
// 			fmt.Println("sending....")
// 			if err != nil {
// 				fmt.Println("send msg faild ", err)
// 				return
// 			}
// 		}
// 	}
// }
