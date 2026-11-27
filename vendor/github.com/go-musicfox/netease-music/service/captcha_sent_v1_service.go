package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type CaptchaSentV1Service struct {
	Ctcode string `json:"ctcode" form:"ctcode"`
	Phone  string `json:"phone" form:"phone"`
}

func (service *CaptchaSentV1Service) CaptchaSentV1() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/middle/captcha/sent/v1",
	}
	data := make(map[string]string)
	if service.Ctcode == "" {
		data["ctcode"] = "86"
	} else {
		data["ctcode"] = service.Ctcode
	}
	data["secrete"] = "music_middleuser_pclogin"
	data["cellphone"] = service.Phone
	data["scene"] = "0"

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/middle/captcha/sent/v1", data, options)

	return code, reBody
}
