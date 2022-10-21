package ws

import (
	"qqbot/lib/qblog"
	"qqbot/lib/qbmsg"
	"time"

	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
)

type WsHandler struct {
	conn   *websocket.Conn
	stopCh chan bool
}

func NewWsHander(wc *websocket.Conn) *WsHandler {
	wh := new(WsHandler)
	wh.conn = wc
	wh.stopCh = make(chan bool)
	return wh
}

func (slf *WsHandler) Close() {
	close(slf.stopCh)
}

func (slf *WsHandler) Run() {
	qblog.Log.Debug("Run a new WsHandler.")
	for {
		select {
		case <-slf.stopCh:
			return
		case <-time.After(time.Second * 3):
			// debuglog("wsHander running, RemoteAddr:", slf.conn.RemoteAddr())
		}
	}
}

func (slf *WsHandler) Read(msg []byte) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	msgData := qbmsg.Event{}
	err := json.Unmarshal(msg, &msgData)
	if err != nil {
		qblog.Log.Error("Decode json msg from go-cqhttp failed:", err)
		return
	}
	// 处理读取到的请求
	if msgData.PortType == "meta_event" {
		if msgData.MetaEventType == "lifecycle" {
			slf.conn.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(5000)))
		} else if msgData.MetaEventType == "heartbeat" {
			// 接收到 heartbeat 元事件，延长 websocket 读超时。
			slf.conn.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(msgData.Interval+3000)))
		}
	} else if msgData.PortType == "message" {
		qblog.Log.Debug("msgData", msgData)
		if msgData.MessageType == "group" {
			// 群聊消息
			qblog.Log.Debug("Get group message from", msgData.GroupId, ":", msgData.Message)
			qbmsg.Group(&msgData, slf.conn)
		} else if msgData.MessageType == "private" {
			// 私聊消息
			qblog.Log.Debug("Get new message from", msgData.UserId, ":", msgData.Message)
			qbmsg.Private(&msgData, slf.conn)
		}
	}
}
