package service

import (
	"encoding/json"

	"github.com/go-musicfox/netease-music/util"
)

type LoginQRService struct {
	UniKey string `json:"unikey"`
}

func (service *LoginQRService) GetKey() (float64, []byte, string) {
	data := map[string]string{
		"type": "1",
	}

	options := &util.Options{
		Crypto: "weapi",
	}

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/login/qrcode/unikey`, data, options)
	if code != 200 || len(reBody) == 0 {
		return code, reBody, ""
	}

	_ = json.Unmarshal(reBody, service)

	return code, reBody, "http://music.163.com/login?codekey=" + service.UniKey
}

func (service *LoginQRService) CheckQR() (float64, []byte) {
	if service.UniKey == "" {
		return 0, nil
	}
	data := map[string]string{
		"type": "1",
		"key":  service.UniKey,
	}

	options := &util.Options{
		Crypto: "weapi",
	}

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/login/qrcode/client/login`, data, options)
	return code, reBody
}
