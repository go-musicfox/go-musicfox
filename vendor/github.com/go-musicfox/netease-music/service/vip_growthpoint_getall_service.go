package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type VipGrowthpointGetallService struct {
}

func (service *VipGrowthpointGetallService) VipGrowthpointGetall() (float64, []byte) {

	options := &util.Options{
		Crypto: "xeapi",
		Url:    "/api/vipnewcenter/app/level/task/reward/getall",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://interface3.music.163.com/xeapi/vipnewcenter/app/level/task/reward/getall", data, options)

	return code, reBody
}
