package qbmsg

import (
	"qqbot/lib/qblog"
	"qqbot/lib/qbmsg/bili"
	"regexp"

	"github.com/gorilla/websocket"
)

// biliLink parse bilibili link to video info
func biliLink(msg *Event, conn *websocket.Conn) error {
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
	// b23.tv
	re := regexp.MustCompile(`b23\.tv\/([0-9a-zA-Z]+)`)
	linkids := re.FindStringSubmatch(strMsg)
	if len(linkids) > 1 {
		// 匹配到 b23.tv 链接
		linkids = linkids[1:]
		for _, linkid := range linkids {
			bv, err := bili.B23ToBvid("https://b23.tv/" + linkid)
			if err != nil {
				continue
			}
			qblog.Log.Debug("获取到BV链接:", bv)
			apiReq.Params.Message, err = bili.GetVideoInfo(bv, "")
			if err != nil {
				continue
			}
			err = apiReq.Send(conn)
			if err != nil {
				continue
			}
		}
		return nil
	}
	// bvid
	re = regexp.MustCompile(`(BV[0-9a-zA-Z]+)`)
	linkids = re.FindStringSubmatch(strMsg)
	if len(linkids) > 1 {
		// 匹配到 bvid 链接
		linkids = linkids[1:]
		for _, linkid := range linkids {
			qblog.Log.Debug("获取到BV链接:", linkid)
			apiReq.Params.Message, err = bili.GetVideoInfo(linkid, "")
			if err != nil {
				continue
			}
			err = apiReq.Send(conn)
			if err != nil {
				continue
			}
		}
		return nil
	}
	// aid
	re = regexp.MustCompile(`av([0-9]+)`)
	linkids = re.FindStringSubmatch(strMsg)
	if len(linkids) > 1 {
		// 匹配到 bvid 链接
		linkids = linkids[1:]
		for _, linkid := range linkids {
			qblog.Log.Debug("获取到AV链接:", linkid)
			apiReq.Params.Message, err = bili.GetVideoInfo("", linkid)
			if err != nil {
				continue
			}
			err = apiReq.Send(conn)
			if err != nil {
				continue
			}
		}
		return nil
	}
	return nil
}
