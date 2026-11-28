package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type CaptchaSafeSentService struct {
	Ctcode string `json:"ctcode" form:"ctcode"`
}

func (service *CaptchaSafeSentService) CaptchaSafeSent() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/sms/captcha/safe/sent",
	}
	data := make(map[string]string)
	if service.Ctcode == "" {
		data["ctcode"] = "86"
	} else {
		data["ctcode"] = service.Ctcode
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/sms/captcha/safe/sent", data, options)

	return code, reBody
}
