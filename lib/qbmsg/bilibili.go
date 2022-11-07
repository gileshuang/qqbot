package qbmsg

import (
	"qqbot/lib/qblog"
	"qqbot/lib/qbmsg/bili"
	"qqbot/lib/qbsql"
	"regexp"
	"time"

	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
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

func BliveMonitor(stopBlive chan bool, conn *websocket.Conn) {
	qblog.Log.Debug("start BliveMonitor")
	for {
		select {
		case <-stopBlive:
			return
		case <-time.After(time.Second * 30):
			BliveCheckStatus(conn)
		}
	}
}

func BliveCheckStatus(conn *websocket.Conn) {
	var (
		json      = jsoniter.ConfigCompatibleWithStandardLibrary
		bliveData string
		roomSuber map[string]([]int64) = make(map[string][]int64)
	)
	qblog.Log.Debug("Start BliveCheckStatus")
	qbsql.InitDB()
	sqlquery := "SELECT id,blive FROM channel WHERE blive != 'NULL' AND blive != '{}'"
	rows, err := qbsql.Db.Query(sqlquery)
	if err != nil {
		qblog.Log.Error("mysql query error:", err, "; SQL:", sqlquery)
	}
	for rows.Next() {
		var groupId int64
		rows.Scan(&groupId, &bliveData)
		// qblog.Log.Debug(groupId, bliveData)
		if bliveData == "" || bliveData == "{}" {
			continue
		}
		var rooms map[string]bili.RoomData
		json.Unmarshal([]byte(bliveData), &rooms)
		for ri, rd := range rooms {
			// 记录该直播间被哪些群订阅了
			roomSuber[ri] = append(roomSuber[ri], groupId)
			// 把本群的 roomData 信息写入全部订阅直播间列表
			if _, ok := bili.NotifyStat[ri]; ok {
				rd.LiveStatus = bili.NotifyStat[ri].LiveStatus
			}
			bili.NotifyStat[ri] = rd
		}
	}
	qblog.Log.Debug("当前各群订阅的直播间有：", roomSuber)
	// 准备 cq API
	apiReq := API{}
	apiReq.Params.AutoEscape = false
	apiReq.Action = "send_group_msg"
	for ri := range bili.NotifyStat {
		// 检查直播间状态并向订阅了该直播间的群发送通知
		// 每个直播间之间检查暂停一秒，避免过于频繁请求接口
		time.After(time.Second * 1)
		out, err := bili.LiveStatus(ri)
		if err != nil {
			// 报错或者直播状态无变化
			continue
		}
		for _, gid := range roomSuber[ri] {
			apiReq.Params.GroupId = gid
			apiReq.Params.Message = out
			err = apiReq.Send(conn)
			if err != nil {
				continue
			}
			// 整活，可删除
			apiReq.Params.Message = "这个主播不会真的觉得有人看吧，不会吧不会吧"
			err = apiReq.Send(conn)
			if err != nil {
				continue
			}
		}
	}
}
