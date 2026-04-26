package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type CaptchaSentService struct {
	Ctcode    string `json:"ctcode" form:"ctcode"`
	Cellphone string `json:"phone" form:"phone"`
}

func (service *CaptchaSentService) CaptchaSent() (float64, []byte) {
	data := make(map[string]interface{})
	if service.Ctcode == "" {
		data["ctcode"] = "86"
	} else {
		data["ctcode"] = service.Ctcode
	}
	data["secrete"] = "music_middleuser_pclogin"
	data["cellphone"] = service.Cellphone

	code, reBody, _ := util.CallApi("https://music.163.com/api/sms/captcha/sent", data)

	return code, reBody
}
