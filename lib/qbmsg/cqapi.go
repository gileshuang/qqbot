package qbmsg

import (
	"qqbot/lib/qblog"

	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
)

type ApiParams struct {
	UserId     int64  `json:"user_id"`
	GroupId    int64  `json:"group_id"`
	Message    string `json:"message"`
	AutoEscape bool   `json:"auto_escape"`
}

type API struct {
	Action string    `json:"action"`
	Params ApiParams `json:"params"`
}

func (slf *API) Send(conn *websocket.Conn) error {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	req, err := json.Marshal(slf)
	if err != nil {
		qblog.Log.Error("Marshal request json error:", err)
		return err
	}
	err = conn.WriteMessage(1, req)
	if err != nil {
		qblog.Log.Error("Write message to websocket error:", err)
		return err
	}
	return nil
}
