package qbmsg

import (
	"qqbot/lib/qbmsg/jd"
	"regexp"

	"github.com/gorilla/websocket"
)

func jdLink(msg *Event, conn *websocket.Conn) error {
	var err error
	strMsg := msg.Message
	apiReq := API{}
	apiReq.Params.AutoEscape = false
	if msg.MessageType == "group" {
		apiReq.Action = "send_group_msg"
		apiReq.Params.GroupId = msg.GroupId
	} else if msg.MessageType == "private" {
		apiReq.Action = "send_private_msg"
		apiReq.Params.GroupId = msg.UserId
	}
	// jd.com
	re := regexp.MustCompile(`jd\.com.*\/([0-9]+).htm.*`)
	linkids := re.FindStringSubmatch(strMsg)
	if len(linkids) < 1 {
		return nil
	}
	linkid := linkids[1]
	apiReq.Params.Message, err = jd.GetItemInfo(linkid)
	if err != nil {
		return err
	}
	err = apiReq.Send(conn)
	if err != nil {
		return err
	}
	return nil
}
