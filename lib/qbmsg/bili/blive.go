package bili

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	jsoniter "github.com/json-iterator/go"
)

// https://github.com/SocialSisterYi/bilibili-API-collect/blob/master/live/info.md

var (
	NotifyStat map[string]RoomData = make(map[string]RoomData)
)

type RoomData struct {
	// 每次重新获取
	RoomId          int64  `json:"room_id"`          // 直播间长号
	UID             int64  `json:"uid"`              // 主播mid
	UserName        string `json:"username"`         // 本地数据库中记录的主播名字
	LiveStatus      int    `json:"live_status"`      // 上一次检查时的直播状态。0：未开播；1：直播中；2：轮播中；
	Attention       int64  `json:"attention"`        // 关注数量（粉丝数）
	AttentionBefore int64  `json:"attention_before"` // 直播开始时的关注数量
	Online          int64  `json:"online"`           // 观看人数
	OnlineMax       int64  `json:"online_max"`       // 最高观看人数
	// 以下仅开播需要，无需持久化
	Title          string `json:"title"`            // 标题
	ParentAreaName string `json:"parent_area_name"` // 父分区名称
	AreaName       string `json:"area_name"`        // 分区名称
	UserCover      string `json:"user_cover"`       // 封面
	Keyframe       string `json:"keyframe"`         // 关键帧 	用于网页端悬浮展示
	// ShortId          int64       `json:"short_id"`           // 直播间短号，为0是无短号
	// IsPortrait       bool        `json:"is_portrait"`        // 是否竖屏
	// Description      string      `json:"description"`        // 描述
	// AreaId           int64       `json:"area_id"`            // 分区id
	// ParentAreaId     int64       `json:"parent_area_id"`     // 父分区id
	// OldAreaId        int64       `json:"old_area_id"`        // 旧版父分区id
	// Background       string      `json:"background"`         // 背景图片链接
	// IsStrictRoom     bool        `json:"is_strict_room"`     // 未知 	未知
	// LiveTime         string      `json:"live_time"`          // 直播开始时间 	YYYY-MM-DD HH:mm:ss
	// Tags             string      `json:"tags"`               // 标签 ','分隔
	// IsAnchor         int64       `json:"is_anchor"`          // 未知 未知
	// RoomSilentType   string      `json:"room_silent_type"`   // 禁言状态
	// RoomSilentLevel  int64       `json:"room_silent_level"`  // 禁言等级
	// RoomSilentSecond int64       `json:"room_silent_second"` // 禁言时间 	单位是秒
	// Pardants         string      `json:"pardants"`           // 未知 	未知
	// AreaPardants     string      `json:"area_pardants"`      // 未知 	未知
	// HotWords         []string    `json:"hot_words"`          // 热词
	// HotWordsStatus   int64       `json:"hot_words_status"`   // 热词状态
	// Verify           string      `json:"verify"`             // 未知 	未知
	// NewPendants      interface{} `json:"new_pendants"`       // 头像框\大v
	// UpSession        string      `json:"up_session"`         // 未知
	// PkStatus         int64       `json:"pk_status"`          // pk状态
	// PkId             int64       `json:"pk_id"`              // pk id
	// BattleId         int64       `json:"battle_id"`          // 未知
}

func getBliveUsernameByUID(uid int64) (string, error) {
	var (
		biliStatAPI = "https://api.live.bilibili.com/room/v1/Room/get_status_info_by_uids"
	)
	// 通过另一个接口获取主播用户名
	uid_s := strconv.FormatInt(uid, 10)
	resp, err := http.Get(biliStatAPI + "?uids[]=" + uid_s)
	if err != nil {
		return "", errors.New("request bilibili api failed")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("read from bilibili api failed")
	}
	username := jsoniter.Get(body, "data", uid_s, "uname").ToString()
	if username == "" {
		return "", errors.New("get username from bilibili uid failed")
	}

	return username, nil
}

func getBliveLockedStatusByRoomID(roomId string) (bool, string) {
	var (
		biliLockAPI = "https://api.live.bilibili.com/room/v1/Room/room_init"
		err         error
		time_layout = "2006-01-02 15:04:05 MST"
		cstZone     = time.FixedZone("CST", 8*3600)
	)
	// 获取直播间的封禁状态
	resp, err := http.Get(biliLockAPI + "id=" + roomId)
	if err != nil {
		return false, ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, ""
	}
	is_locked := jsoniter.Get(body, "data", "is_locked").ToBool()
	lock_till := jsoniter.Get(body, "data", "is_locked").ToInt64()
	t_lock_till := time.Unix(lock_till, 0)

	return is_locked, t_lock_till.In(cstZone).Format(time_layout)
}

func LiveStatus(roomId string) (string, error) {
	var (
		biliInfoAPI = "https://api.live.bilibili.com/room/v1/Room/get_info"
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
	// 实时获取的直播状态，以与记录的当前直播状态区分开
	live_status := jsoniter.Get(body, "data", "live_status").ToInt()
	// 获取内存中的直播间信息
	last_room_data := NotifyStat[roomId]
	defer func() {
		NotifyStat[roomId] = last_room_data
	}()
	// 更新内存中的直播间信息
	last_room_data.RoomId = jsoniter.Get(body, "data", "room_id").ToInt64()
	last_room_data.UID = jsoniter.Get(body, "data", "uid").ToInt64()
	last_room_data.Attention = jsoniter.Get(body, "data", "attention").ToInt64()
	last_room_data.Online = jsoniter.Get(body, "data", "online").ToInt64()
	if last_room_data.OnlineMax < last_room_data.Online {
		// 更新最高观看人数计数
		last_room_data.OnlineMax = last_room_data.Online
	}
	// 判断当前直播间状态
	if last_room_data.LiveStatus != 1 && live_status == 1 {
		// 直播已开始
		last_room_data.AttentionBefore = last_room_data.Attention
		last_room_data.Title = jsoniter.Get(body, "data", "title").ToString()
		last_room_data.UserName, _ = getBliveUsernameByUID(last_room_data.UID)
		last_room_data.ParentAreaName = jsoniter.Get(body, "data", "parent_area_name").ToString()
		last_room_data.AreaName = jsoniter.Get(body, "data", "area_name").ToString()
		last_room_data.Keyframe = jsoniter.Get(body, "data", "keyframe").ToString()
		// 拼装输出信息
		out = last_room_data.UserName + " 的直播已开始，但他似乎真的以为有人会看。\n" +
			"https://live.bilibili.com/" + strconv.FormatInt(last_room_data.RoomId, 10) + "\n" +
			last_room_data.Title + "\n" +
			"分区：" + last_room_data.ParentAreaName + " - " + last_room_data.AreaName + "\n" +
			"粉丝数：" + strconv.FormatInt(last_room_data.AttentionBefore, 10) + " | " +
			"当前观看人数：" + strconv.FormatInt(last_room_data.Online, 10) + "\n" +
			"当前直播画面：\n" +
			"[CQ:image,file=" + last_room_data.Keyframe + "]\n"
		last_room_data.LiveStatus = live_status
	} else if last_room_data.LiveStatus == 1 && live_status != 1 {
		// 直播已结束
		attention_new := last_room_data.Attention - last_room_data.AttentionBefore
		last_room_data.UserCover = jsoniter.Get(body, "data", "user_cover").ToString()
		// 拼装输出信息
		is_locked, s_lock_till := getBliveLockedStatusByRoomID(roomId)
		if !is_locked {
			out = last_room_data.UserName + " 的直播已结束。\n" +
				"当前粉丝数：" + strconv.FormatInt(last_room_data.Attention, 10) + " | " +
				"新增粉丝数：" + strconv.FormatInt(attention_new, 10) + "\n" +
				"当前观看人数：" + strconv.FormatInt(last_room_data.Online, 10) + " | " +
				"最高观看人数：" + strconv.FormatInt(last_room_data.OnlineMax, 10) + "\n" +
				"[CQ:image,file=" + last_room_data.UserCover + "]\n"
		} else {
			out = last_room_data.UserName + " 的直播间被封禁啦哈哈哈哈哈嗝~~~\n" +
				"当前粉丝数：" + strconv.FormatInt(last_room_data.Attention, 10) + " | " +
				"新增粉丝数：" + strconv.FormatInt(attention_new, 10) + "\n" +
				"最高观看人数：" + strconv.FormatInt(last_room_data.OnlineMax, 10) + "\n" +
				"解封时间：" + s_lock_till + "\n" +
				"[CQ:image,file=" + last_room_data.UserCover + "]\n"
		}
		if live_status == 2 {
			out = out + "主播开启了视频轮播：\n" +
				"https://live.bilibili.com/" + strconv.FormatInt(last_room_data.RoomId, 10) + "\n"
		}
		last_room_data.LiveStatus = live_status
		last_room_data.OnlineMax = 0 // 重置最高观看人数计数
	} else {
		// 直播状态没变化，不需要通知，仅更新状态
		last_room_data.UserName, _ = getBliveUsernameByUID(last_room_data.UID)
		last_room_data.LiveStatus = live_status
		// NotifyStat[roomId] = last_room_data
		return "", errors.New("live status not changed")
	}
	// NotifyStat[roomId] = last_room_data

	return out, nil
}
