package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type VipGrowthpointGetService struct {
	Ids string `json:"ids" form:"ids"`
}

func (service *VipGrowthpointGetService) VipGrowthpointGet() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["taskIds"] = service.Ids

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/vipnewcenter/app/level/task/reward/get", data, options)

	return code, reBody
}
