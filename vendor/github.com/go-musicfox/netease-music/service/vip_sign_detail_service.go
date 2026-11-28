package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type VipSignDetailService struct {
	Timestamp string `json:"timestamp" form:"timestamp"`
}

func (service *VipSignDetailService) VipSignDetail() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/vipnewcenter/app/level/user/checkin/history/detail",
	}
	data := make(map[string]string)
	data["signDayTime"] = service.Timestamp
	data["type"] = "1"

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/vipnewcenter/app/level/user/checkin/history/detail", data, options)

	return code, reBody
}
