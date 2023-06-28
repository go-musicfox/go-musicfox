package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type MvExclusiveRcmdService struct {
	Limit  string `json:"limit" form:"limit"`
	Offset string `json:"offset" form:"offset"`
}

func (service *MvExclusiveRcmdService) MvExclusiveRcmd() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.Limit == "" {
		data["limit"] = "30"
	} else {
		data["limit"] = service.Limit
	}
	if service.Offset == "" {
		data["offset"] = "0"
	} else {
		data["offset"] = service.Offset
	}
	data["order"] = "true"
	code, reBody, _ := util.CreateRequest("POST", `https://interface.music.163.com/api/mv/exclusive/rcmd`, data, options)

	return code, reBody
}
