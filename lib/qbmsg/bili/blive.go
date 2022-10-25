package bili

import (
	"errors"
	"io"
	"net/http"

	jsoniter "github.com/json-iterator/go"
)

// https://github.com/SocialSisterYi/bilibili-API-collect/blob/master/live/info.md

var (
	NotifyStat map[string]RoomData = make(map[string]RoomData)
)

type RoomData struct {
	UID              int64       `json:"uid"`                // 主播mid
	UserName         string      `json:"username"`           // 本地数据库中记录的主播名字
	RoomId           int64       `json:"room_id"`            // 直播间长号
	ShortId          int64       `json:"short_id"`           // 直播间短号，为0是无短号
	Attention        int64       `json:"attention"`          // 关注数量
	Online           int64       `json:"online"`             // 观看人数
	IsPortrait       bool        `json:"is_portrait"`        // 是否竖屏
	LiveStatus       int         `json:"live_status"`        // 直播状态。0：未开播；1：直播中；2：轮播中；
	Description      string      `json:"description"`        // 描述
	AreaId           int64       `json:"area_id"`            // 分区id
	AreaName         string      `json:"area_name"`          // 分区名称
	ParentAreaId     int64       `json:"parent_area_id"`     // 父分区id
	ParentAreaName   string      `json:"parent_area_name"`   // 父分区名称
	OldAreaId        int64       `json:"old_area_id"`        // 旧版父分区id
	Background       string      `json:"background"`         // 背景图片链接
	Title            string      `json:"title"`              // 标题
	UserCover        string      `json:"user_cover"`         // 封面
	Keyframe         string      `json:"keyframe"`           // 关键帧 	用于网页端悬浮展示
	IsStrictRoom     bool        `json:"is_strict_room"`     // 未知 	未知
	LiveTime         string      `json:"live_time"`          // 直播开始时间 	YYYY-MM-DD HH:mm:ss
	Tags             string      `json:"tags"`               // 标签 ','分隔
	IsAnchor         int64       `json:"is_anchor"`          // 未知 未知
	RoomSilentType   string      `json:"room_silent_type"`   // 禁言状态
	RoomSilentLevel  int64       `json:"room_silent_level"`  // 禁言等级
	RoomSilentSecond int64       `json:"room_silent_second"` // 禁言时间 	单位是秒
	Pardants         string      `json:"pardants"`           // 未知 	未知
	AreaPardants     string      `json:"area_pardants"`      // 未知 	未知
	HotWords         []string    `json:"hot_words"`          // 热词
	HotWordsStatus   int64       `json:"hot_words_status"`   // 热词状态
	Verify           string      `json:"verify"`             // 未知 	未知
	NewPendants      interface{} `json:"new_pendants"`       // 头像框\大v
	UpSession        string      `json:"up_session"`         // 未知
	PkStatus         int64       `json:"pk_status"`          // pk状态
	PkId             int64       `json:"pk_id"`              // pk id
	BattleId         int64       `json:"battle_id"`          // 未知
}

func LiveStatus(roomId string) (string, error) {
	var (
		biliInfoAPI = "https://api.live.bilibili.com/room/v1/Room/get_info"
		biliStatAPI = "http://api.live.bilibili.com/room/v1/Room/get_status_info_by_uids"
		resp        *http.Response
		// out 为输出给QQ的信息
		// 当本次查询状态与上一次相同时，或者查询直播间状态异常时，返回 err != nil
		out string = ""
		err error
	)
	// 获取直播间信息，包括在线人数等
	resp, err = http.Get(biliInfoAPI + "?room_id=" + roomId)
	if err != nil {
		return "", errors.New("request bilibili api failed")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("read from bilibili api failed")
	}
	// 通过另一个接口获取主播用户名
	uid := jsoniter.Get(body, "data", "uid").ToString()
	resp, err = http.Get(biliStatAPI + "?uids[]=" + uid)
	if err != nil {
		return "", errors.New("request bilibili api failed")
	}
	defer resp.Body.Close()
	body2, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("read from bilibili api failed")
	}
	// 更新内存中的直播间信息
	live_status := jsoniter.Get(body, "data", "live_status").ToInt()
	last_room_data := NotifyStat[roomId]
	last_room_data.UID = jsoniter.Get(body, "data", "uid").ToInt64()
	last_room_data.UserName = jsoniter.Get(body2, "data", uid, "uname").ToString()
	// qblog.Log.Debug("room_id", jsoniter.Get(body, "data", "room_id").ToString())
	// qblog.Log.Debug("last_room_data.LiveStatus", last_room_data.LiveStatus)
	// qblog.Log.Debug("live_status", live_status)
	// 判断当前直播间状态
	if last_room_data.LiveStatus != 1 && live_status == 1 {
		// 直播已开始
		// 拼装输出信息
		out = "https://live.bilibili.com/" + jsoniter.Get(body, "data", "room_id").ToString() + "\n" +
			jsoniter.Get(body, "data", "title").ToString() + "\n" +
			"分区：" + jsoniter.Get(body, "data", "parent_area_name").ToString() + " - " +
			jsoniter.Get(body, "data", "area_name").ToString() + "\n" +
			"主播：" + last_room_data.UserName + "\n" +
			"关注人数：" + jsoniter.Get(body, "data", "attention").ToString() + " | " +
			"观看人数：" + jsoniter.Get(body, "data", "online").ToString() + "\n" +
			"封面：\n" +
			"[CQ:image,file=" + jsoniter.Get(body, "data", "user_cover").ToString() + "]\n"
		last_room_data.LiveStatus = live_status
	} else if last_room_data.LiveStatus == 1 && live_status == 0 {
		// 直播已关闭
		// 拼装输出信息
		out = last_room_data.UserName + " 的直播已结束。\n" +
			"关注人数：" + jsoniter.Get(body, "data", "attention").ToString() + " | " +
			"观看人数：" + jsoniter.Get(body, "data", "online").ToString() + "\n"
		last_room_data.LiveStatus = live_status
	} else if last_room_data.LiveStatus == 1 && live_status == 2 {
		// 直播已关闭，但主播开启了轮播
		// 拼装输出信息
		out = last_room_data.UserName + " 的直播已结束。\n" +
			"关注人数：" + jsoniter.Get(body, "data", "attention").ToString() + " | " +
			"观看人数：" + jsoniter.Get(body, "data", "online").ToString() + "\n" +
			"主播开启了视频轮播：\n" +
			"https://live.bilibili.com/" + jsoniter.Get(body, "data", "room_id").ToString() + "\n"
		last_room_data.LiveStatus = live_status
	} else {
		// 直播状态没变化，不需要通知
		last_room_data.LiveStatus = live_status
		NotifyStat[roomId] = last_room_data
		return "", errors.New("live status not changed")
	}
	NotifyStat[roomId] = last_room_data

	return out, nil
}
