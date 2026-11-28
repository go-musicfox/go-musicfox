package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type DeviceKickoffService struct {
	Captcha   string `json:"captcha" form:"captcha"`
	DeviceKey string `json:"deviceKey" form:"deviceKey"`
}

func (service *DeviceKickoffService) DeviceKickoff() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/middle/user/security/device/kickoff",
	}
	data := make(map[string]string)
	data["key"] = service.DeviceKey
	if service.Captcha == "" {
		data["captcha"] = ""
	} else {
		data["captcha"] = service.Captcha
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/middle/user/security/device/kickoff", data, options)

	return code, reBody
}
