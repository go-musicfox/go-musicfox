package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type MvSublistService struct {
	Limit  string `json:"limit" form:"limit"`
	Offset string `json:"offset" form:"offset"`
}

func (service *MvSublistService) MvSublist() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	if service.Limit == "" {
		service.Limit = "25"
	}
	if service.Offset == "" {
		service.Offset = "0"
	}
	data["limit"] = service.Limit
	data["offset"] = service.Offset
	data["total"] = "true"

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/cloudvideo/allvideo/sublist`, data, options)

	return code, reBody
}
