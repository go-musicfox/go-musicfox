package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type MsgCommentsService struct {
	UID        string `json:"uid" form:"uid"`
	Limit      string `json:"limit" form:"limit"`
	BeforeTime string `json:"before" form:"before"`
}

func (service *MsgCommentsService) MsgComments() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["uid"] = service.UID
	if service.Limit == "" {
		data["limit"] = "30"
	} else {
		data["limit"] = service.Limit
	}
	if service.BeforeTime == "" {
		data["beforeTime"] = "-1"
	} else {
		data["beforeTime"] = service.BeforeTime
	}
	data["order"] = "true"
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/v1/user/comments/`+service.UID, data, options)

	return code, reBody
}
