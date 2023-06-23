package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type UserRecordService struct {
	UId  string `json:"uid" form:"uid"`
	Type string `json:"type" form:"type"`
}

func (service *UserRecordService) UserRecord() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["uid"] = service.UId

	if service.Type == "1" {
		data["type"] = "1"
	} else {
		data["type"] = "0"
	}
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/v1/play/record`, data, options)

	return code, reBody
}
