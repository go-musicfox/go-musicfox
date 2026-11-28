package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type VipInfoService struct {
	Uid string `json:"uid" form:"uid"`
}

func (service *VipInfoService) VipInfo() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["userId"] = service.Uid

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/music-vip-membership/front/vip/info", data, options)

	return code, reBody
}
