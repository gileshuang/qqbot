package jd

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	jsoniter "github.com/json-iterator/go"
)

func GetItemInfo(id string) (string, error) {
	jdUrl := "https://www.bijiago.com/bjg/api/dplist?dpids=" + id + "-3"
	resp, err := http.Get(jdUrl)
	if err != nil {
		return "", errors.New("request jq api failed")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("get item info from jd failed")
	}
	out := ""
	for i := 0; i < jsoniter.Get(body).Size(); i++ {
		var price float32 = jsoniter.Get(body, 0, "price").ToFloat32() / 100
		out = out + jsoniter.Get(body, 0, "url").ToString() + "\n" +
			jsoniter.Get(body, 0, "title").ToString() + "\n" +
			"电商：" + jsoniter.Get(body, 0, "e_site_name").ToString() + "\n" +
			"店铺：" + jsoniter.Get(body, 0, "site_name").ToString() + "\n" +
			"价格（不含优惠）：" + fmt.Sprint(price) + "\n" +
			"[CQ:image,file=" + jsoniter.Get(body, 0, "img").ToString() + "]\n"
	}
	return out, nil
}
