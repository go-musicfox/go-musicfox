package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type VipGrowthpointService struct {
}

func (service *VipGrowthpointService) VipGrowthpoint() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/vipnewcenter/app/level/growhpoint/basic", data, options)

	return code, reBody
}
