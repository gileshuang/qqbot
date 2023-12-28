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
	// error: return error info
	// bool: return if msg type matched
	var (
		err     error
		matched bool
	)
	err, matched = qqMiniApp(msg, conn)
	if err == nil && matched {
		return nil
	}
	err, matched = qqStructMsg(msg, conn)
	if err == nil && matched {
		return nil
	}
	biliLink(msg, conn)
	// 因比价宝接口异常，暂时不解析京东商品详情
	// jdLink(msg, conn)
	return nil
}
