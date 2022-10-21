package bili

// doc: https://github.com/SocialSisterYi/bilibili-API-collect

import (
	"errors"
	"io"
	"net/http"
	"qqbot/lib/qblog"
	"regexp"

	jsoniter "github.com/json-iterator/go"
)

// Get302Location 获取一个 URL 的 301 或者 302 跳转地址
func Get302Location(url string) (string, error) {
	c := &http.Client{}
	c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		// 在 CheckRedirect 时返回错误，以禁止 go 自动跟随重定向
		return errors.New("jump out 302 redirect")
	}
	defer c.CloseIdleConnections()
	req, _ := http.NewRequest("GET", url, nil)
	resp, _ := c.Do(req)
	if resp.StatusCode != 301 && resp.StatusCode != 302 {
		qblog.Log.Warning("Url链接未返回302:", url)
		return "", errors.New("http status code error, not 302 status")
	}

	location := resp.Header.Get("Location")
	if location == "" {
		qblog.Log.Warning("解析302链接失败:", url)
		return "", errors.New("failed to get 302 location")
	}
	return location, nil
}

// AVLinkToAvid cut avid from av link
func AVLinkToAvid(location string) (string, error) {
	re := regexp.MustCompile(`[\/|^|\s]av([0-9]+)`)
	m := re.FindStringSubmatch(location)
	if m == nil || len(m) < 2 {
		// 未匹配到消息中stringstring
		qblog.Log.Info("未匹配到消息中的AVID:", location)
		return "", errors.New("failed to get AVID from string")
	}
	return m[1], nil
}

// BVLinkToBvid cut bvid from bv link
func BVLinkToBvid(location string) (string, error) {
	re := regexp.MustCompile(`[\/|^|\s](BV[0-9A-Za-z]+)`)
	m := re.FindStringSubmatch(location)
	if m == nil || len(m) < 2 {
		// 未匹配到消息中的 BVID
		qblog.Log.Info("未匹配到消息中的BVID:", location)
		return "", errors.New("failed to get BVID from string")
	}
	return m[1], nil
}

// B23ToBvid 从 b23.tv 的链接中获取 bvid
func B23ToBvid(url string) (string, error) {
	location, err := Get302Location(url)
	if err != nil {
		return "", err
	}
	bvid, err := BVLinkToBvid(location)
	if err != nil {
		return "", err
	}
	return bvid, nil
}

func GetVideoInfo(bvid, avid string) (string, error) {
	var (
		biliAPI = "http://api.bilibili.com/x/web-interface/view"
		resp    *http.Response
		err     error
	)
	if bvid != "" {
		resp, err = http.Get(biliAPI + "?bvid=" + bvid)
	} else if avid != "" {
		resp, err = http.Get(biliAPI + "?aid=" + avid)
	} else {
		return "", errors.New("one of bvid or avid should be given")
	}
	if err != nil {
		return "", errors.New("request bilibili api failed")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("read from bilibili api failed")
	}
	vBvid := jsoniter.Get(body, "data", "bvid")
	vDesc := jsoniter.Get(body, "data", "desc")
	qblog.Log.Debug("bvid:", vBvid)
	qblog.Log.Debug("desc:", vDesc)
	out := "https://www.bilibili.com/video/" + jsoniter.Get(body, "data", "bvid").ToString() + "\n" +
		"标题：" + jsoniter.Get(body, "data", "title").ToString() + "\n" +
		"up主：" + jsoniter.Get(body, "data", "owner", "name").ToString() + "\n" +
		"分区：" + jsoniter.Get(body, "data", "tname").ToString() + "\n" +
		"播放数：" + jsoniter.Get(body, "data", "stat", "view").ToString() + " | " +
		"弹幕数：" + jsoniter.Get(body, "data", "stat", "danmaku").ToString() + " | " +
		"评论数：" + jsoniter.Get(body, "data", "stat", "reply").ToString() + "\n" +
		"收藏数：" + jsoniter.Get(body, "data", "stat", "favorite").ToString() + " | " +
		"投币数：" + jsoniter.Get(body, "data", "stat", "coin").ToString() + " | " +
		"分享数：" + jsoniter.Get(body, "data", "stat", "share").ToString() + "\n" +
		"简介：" + jsoniter.Get(body, "data", "desc").ToString() + "\n" +
		"[CQ:image,file=" + jsoniter.Get(body, "data", "pic").ToString() + "]\n"
	return out, nil
}
