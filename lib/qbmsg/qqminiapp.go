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
	re := regexp.MustCompile(`.+\[CQ:json,data=(.+)\]$`)
	jsonDatas := re.FindStringSubmatch(strMsg)
	if jsonDatas == nil || len(jsonDatas) < 2 {
		// 未匹配到 CQ code json 数据段，不是小程序
		return nil
	}
	qblog.Log.Info("收到了QQ小程序")
	apiReq := API{}
	jsonDatas = jsonDatas[1:]
	qblog.Log.Debug("jsonDatas", jsonDatas)
	for _, jsonData := range jsonDatas {
		data := new(miniApp)
		jsonData = cqcode.UnescapeValue(jsonData)
		json.Unmarshal([]byte(jsonData), &data)
		// 判断是不是 B 站小程序
		if data.Meta.Detail1.Appid == "1109937557" {
			qqDocUrl := data.Meta.Detail1.Qqdocurl
			qblog.Log.Debug("qqdocurl:", qqDocUrl)
			bv, err := bili.B23ToBvid(qqDocUrl)
			if err != nil {
				continue
			}
			qblog.Log.Debug("获取到BV链接:", bv)
			out, err := bili.GetVideoInfo(bv, "")
			if err != nil {
				return err
			}
			apiReq.Action = "send_group_msg"
			apiReq.Params.GroupId = msg.GroupId
			apiReq.Params.Message = out
			apiReq.Params.AutoEscape = false
			err = apiReq.Send(conn)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
