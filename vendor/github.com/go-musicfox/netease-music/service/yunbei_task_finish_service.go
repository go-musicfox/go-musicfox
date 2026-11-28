package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type YunbeiTaskFinishService struct {
	DepositCode string `json:"depositCode" form:"depositCode"`
	UserTaskId  string `json:"userTaskId" form:"userTaskId"`
}

func (service *YunbeiTaskFinishService) YunbeiTaskFinish() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["userTaskId"] = service.UserTaskId
	if service.DepositCode == "" {
		data["depositCode"] = "0"
	} else {
		data["depositCode"] = service.DepositCode
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/usertool/task/point/receive", data, options)

	return code, reBody
}
