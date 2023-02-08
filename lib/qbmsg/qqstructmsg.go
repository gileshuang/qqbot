package qbmsg

import (
	"qqbot/lib/cqcode"
	"qqbot/lib/qblog"
	"regexp"

	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
)

type structMsg struct {
	App    string          `json:"app"`
	Desc   string          `json:"desc"`
	View   string          `json:"view"`
	Ver    string          `json:"ver"`
	Prompt string          `json:"prompt"`
	Meta   structMsgMeta   `json:"meta"`
	Config structMsgConfig `json:"config"`
}
type structMsgNews struct {
	Action         string `json:"action"`
	AndroidPkgName string `json:"android_pkg_name"`
	AppType        int    `json:"app_type"`
	Appid          int    `json:"appid"`
	Ctime          int    `json:"ctime"`
	Desc           string `json:"desc"`
	JumpURL        string `json:"jumpUrl"`
	Preview        string `json:"preview"`
	SourceIcon     string `json:"source_icon"`
	SourceURL      string `json:"source_url"`
	Tag            string `json:"tag"`
	Title          string `json:"title"`
	Uin            int    `json:"uin"`
}
type structMsgMeta struct {
	News structMsgNews `json:"news"`
}
type structMsgConfig struct {
	Ctime   int    `json:"ctime"`
	Forward bool   `json:"forward"`
	Token   string `json:"token"`
	Type    string `json:"type"`
}

func qqStructMsg(msg *Event, conn *websocket.Conn) (error, bool) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	strMsg := msg.Message
	re := regexp.MustCompile(`.+\[CQ:json,data=(.+)\]$`)
	jsonDatas := re.FindStringSubmatch(strMsg)
	if jsonDatas == nil || len(jsonDatas) < 2 {
		// 未匹配到 CQ code json 数据段，不是结构化消息
		return nil, false
	}
	qblog.Log.Info("收到了 CQ json 数据")
	jsonData := cqcode.UnescapeValue(jsonDatas[1])
	qblog.Log.Debug("jsonData", jsonData)
	data := new(structMsg)
	json.Unmarshal([]byte(jsonData), &data)
	if data.App != "com.tencent.structmsg" {
		// 不是结构化消息，直接退出
		return nil, false
	}
	qblog.Log.Info("收到了结构化消息")
	// 构建要发送消息的结构体
	apiReq := API{}
	apiReq.Params.AutoEscape = false
	if msg.MessageType == "group" {
		apiReq.Action = "send_group_msg"
		apiReq.Params.GroupId = msg.GroupId
	} else if msg.MessageType == "private" {
		apiReq.Action = "send_private_msg"
		apiReq.Params.GroupId = msg.UserId
	}
	// 判断结构化消息类型
	if data.Meta.News.Appid == 101492711 {
		// Acfun 分享消息
		apiReq.Params.Message = data.Meta.News.JumpURL + "\n" +
			data.Meta.News.Title + "\n" +
			"分享来源：" + data.Meta.News.Tag + "\n" +
			"简介：" + data.Meta.News.Desc + "\n" +
			"[CQ:image,file=" + data.Meta.News.Preview + "]\n"
		err := apiReq.Send(conn)
		if err != nil {
			return err, true
		}
		return nil, true
	}
	// 不是已知的结构化消息类型
	return nil, false
}
