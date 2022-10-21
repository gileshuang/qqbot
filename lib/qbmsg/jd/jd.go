package jd

import (
	"errors"
	"io"
	"net/http"

	jsoniter "github.com/json-iterator/go"
)

func GetItemInfo(id string) (string, error) {
	jdUrl := "https://cd.jd.com/recommend?methods=accessories&cat=670%2C671%2C672&sku=" + id
	resp, err := http.Get(jdUrl)
	if err != nil {
		return "", errors.New("request jq api failed")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("get item info from jd failed")
	}
	out := "https://item.jd.com/" + id + ".html\n" +
		jsoniter.Get(body, "accessories", "data", "wName").ToString() + "\n" +
		"品牌：" + jsoniter.Get(body, "accessories", "data", "chBrand").ToString() + "\n" +
		"型号：" + jsoniter.Get(body, "accessories", "data", "model").ToString() + "\n" +
		"价格（不含优惠）：" + jsoniter.Get(body, "accessories", "data", "wMaprice").ToString() + "\n" +
		"[CQ:image,file=https://img12.360buyimg.com/n1/s450x450_" + jsoniter.Get(body, "accessories", "data", "imageUrl").ToString() + "]\n"
	return out, nil
}
