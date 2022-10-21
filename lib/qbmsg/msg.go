package qbmsg

import (
	"github.com/gorilla/websocket"
)

func Private(msg *Event, conn *websocket.Conn) error {
	apiReq := API{}
	if msg.UserId == 123456789 {
		apiReq.Action = "send_private_msg"
		apiReq.Params.UserId = msg.UserId
		apiReq.Params.Message = msg.Sender.NickName + " 说：" + msg.RawMessage
		apiReq.Params.AutoEscape = false
		err := apiReq.Send(conn)
		if err != nil {
			return err
		}
	}
	return nil
}

func Group(msg *Event, conn *websocket.Conn) error {
	qqMiniApp(msg, conn)
	qqStructMsg(msg, conn)
	biliLink(msg, conn)
	jdLink(msg, conn)
	return nil
}
