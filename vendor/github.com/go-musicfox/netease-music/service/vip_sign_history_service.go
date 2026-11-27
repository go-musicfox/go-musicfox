package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type VipSignHistoryService struct {
	Type string `json:"type" form:"type"`
}

func (service *VipSignHistoryService) VipSignHistory() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/vipnewcenter/app/minidesk/music/sign/pc",
	}
	data := make(map[string]string)
	if service.Type == "" {
		data["type"] = "0"
	} else {
		data["type"] = service.Type
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/vipnewcenter/app/minidesk/music/sign/pc", data, options)

	return code, reBody
}
