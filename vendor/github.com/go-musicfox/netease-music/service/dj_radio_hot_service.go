package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type DjRadioHotService struct {
	CateId string `json:"cateId" form:"cateId"`
	Limit  string `json:"limit" form:"limit"`
	Offset string `json:"offset" form:"offset"`
}

func (service *DjRadioHotService) DjRadioHot() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["cateId"] = service.CateId
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
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/djradio/hot`, data, options)

	return code, reBody
}
