package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type CaptchaVerifyService struct {
	Ctcode    string `json:"ctcode" form:"ctcode"`
	Cellphone string `json:"phone" form:"phone"`
	Captcha   string `json:"captcha" form:"captcha"`
}

func (service *CaptchaVerifyService) CaptchaVerify() (float64, []byte) {
	data := make(map[string]interface{})
	if service.Ctcode == "" {
		data["ctcode"] = "86"
	} else {
		data["ctcode"] = service.Ctcode
	}
	data["cellphone"] = service.Cellphone
	data["captcha"] = service.Captcha

	code, reBody, _ := util.CallApi("https://music.163.com/api/sms/captcha/verify", data)

	return code, reBody
}
