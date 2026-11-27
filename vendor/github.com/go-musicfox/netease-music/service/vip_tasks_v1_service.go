package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type VipTasksV1Service struct {
	Id string `json:"id" form:"id"`
}

func (service *VipTasksV1Service) VipTasksV1() (float64, []byte) {

	options := &util.Options{
		Crypto: "xeapi",
		Url:    "/api/middle/vip/mission/user/progress/list",
	}
	data := map[string]string{
		"taskType": "app_vip_task_center",
		"userId":   service.Id,
	}

	code, reBody, _ := util.CreateRequest("POST", "https://interface3.music.163.com/xeapi/middle/vip/mission/user/progress/list", data, options)

	return code, reBody
}
