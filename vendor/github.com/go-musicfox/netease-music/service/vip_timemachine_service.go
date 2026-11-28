package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type VipTimemachineService struct {
	StartTime string `json:"startTime" form:"startTime"`
	EndTime   string `json:"endTime" form:"endTime"`
	Limit     string `json:"limit" form:"limit"`
}

func (service *VipTimemachineService) VipTimemachine() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.StartTime != "" && service.EndTime != "" {
		data["startTime"] = service.StartTime
		data["endTime"] = service.EndTime
		data["type"] = "1"
		if service.Limit == "" {
			data["limit"] = "60"
		} else {
			data["limit"] = service.Limit
		}
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/vipmusic/newrecord/weekflow", data, options)

	return code, reBody
}
