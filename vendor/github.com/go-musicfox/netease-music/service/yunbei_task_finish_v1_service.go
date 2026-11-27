package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type YunbeiTaskFinishV1Service struct {
	YunbeiAmount string `json:"yunbeiAmount" form:"yunbeiAmount"`
}

func (service *YunbeiTaskFinishV1Service) YunbeiTaskFinishV1() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.YunbeiAmount == "" {
		data["yunbeiAmount"] = "150"
	} else {
		data["yunbeiAmount"] = service.YunbeiAmount
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/ad/power/yunbei/distribution/create", data, options)

	return code, reBody
}
