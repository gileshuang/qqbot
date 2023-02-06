package qbmsg

import (
	"qqbot/lib/cqcode"
	"qqbot/lib/qblog"
	"qqbot/lib/qbmsg/bili"
	"regexp"

	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
)

type miniApp struct {
	App    string        `json:"app"`
	Desc   string        `json:"desc"`
	View   string        `json:"view"`
	Ver    string        `json:"ver"`
	Prompt string        `json:"prompt"`
	Meta   miniAppMeta   `json:"meta"`
	Config miniAppConfig `json:"config"`
}

type miniAppHost struct {
	Nick string `json:"nick"`
	Uin  int    `json:"uin"`
}

type miniAppShareTemplateData struct {
}

type miniAppDetail1 struct {
	AppType           int                      `json:"appType"`
	Appid             string                   `json:"appid"`
	Desc              string                   `json:"desc"`
	GamePoints        string                   `json:"gamePoints"`
	GamePointsURL     string                   `json:"gamePointsUrl"`
	Host              miniAppHost              `json:"host"`
	Icon              string                   `json:"icon"`
	Preview           string                   `json:"preview"`
	Qqdocurl          string                   `json:"qqdocurl"`
	Scene             int                      `json:"scene"`
	ShareTemplateData miniAppShareTemplateData `json:"shareTemplateData"`
	ShareTemplateID   string                   `json:"shareTemplateId"`
	ShowLittleTail    string                   `json:"showLittleTail"`
	Title             string                   `json:"title"`
	URL               string                   `json:"url"`
}

type miniAppMeta struct {
	Detail1 miniAppDetail1 `json:"detail_1"`
}

type miniAppConfig struct {
	AutoSize int    `json:"autoSize"`
	Ctime    int    `json:"ctime"`
	Forward  int    `json:"forward"`
	Height   int    `json:"height"`
	Token    string `json:"token"`
	Type     string `json:"type"`
	Width    int    `json:"width"`
}

func qqMiniApp(msg *Event, conn *websocket.Conn) error {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	strMsg := msg.Message
	re := regexp.MustCompile(`.*\[CQ:json,data=(.+)\]$`)
	jsonDatas := re.FindStringSubmatch(strMsg)
	if jsonDatas == nil || len(jsonDatas) < 2 {
		// 未匹配到 CQ code json 数据段，不是小程序
		return nil
	}
	qblog.Log.Info("收到了 CQ json 数据")
	jsonData := cqcode.UnescapeValue(jsonDatas[1])
	qblog.Log.Debug("jsonData", jsonData)
	data := new(miniApp)
	json.Unmarshal([]byte(jsonData), &data)
	if data.App != "com.tencent.miniapp_01" {
		// 不是QQ小程序，直接退出
		return nil
	}
	qblog.Log.Info("收到了QQ小程序")
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
	// 判断小程序类型
	if data.Meta.Detail1.Appid == "1109937557" {
		// B 站小程序
		qqDocUrl := data.Meta.Detail1.Qqdocurl
		qblog.Log.Debug("qqdocurl:", qqDocUrl)
		bv, err := bili.B23ToBvid(qqDocUrl)
		if err != nil {
			return nil
		}
		qblog.Log.Debug("获取到BV链接:", bv)
		apiReq.Params.Message, err = bili.GetVideoInfo(bv, "")
		if err != nil {
			return err
		}
		err = apiReq.Send(conn)
		if err != nil {
			return err
		}
	}
	if data.Meta.Detail1.Appid == "1108735743" {
		// 快手小程序
		apiReq.Params.Message = data.Meta.Detail1.Qqdocurl + "\n" +
			data.Meta.Detail1.Desc + "\n" +
			"[CQ:image,file=" + data.Meta.Detail1.Preview + "]\n"
		err := apiReq.Send(conn)
		if err != nil {
			return err
		}
	}
	return nil
}
