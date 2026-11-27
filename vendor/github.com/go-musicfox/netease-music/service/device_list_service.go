package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type DeviceListService struct {
}

func (service *DeviceListService) DeviceList() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/middle/user/device/list",
	}
	data := make(map[string]string)
	data["excStatus"] = "9"

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/middle/user/device/list", data, options)

	return code, reBody
}
