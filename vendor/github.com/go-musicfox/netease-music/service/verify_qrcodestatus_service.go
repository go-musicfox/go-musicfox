package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type VerifyQrcodestatusService struct {
	Qr string `json:"qr" form:"qr"`
}

func (service *VerifyQrcodestatusService) VerifyQrcodestatus() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["qrCode"] = service.Qr

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/frontrisk/verify/qrcodestatus", data, options)

	return code, reBody
}
