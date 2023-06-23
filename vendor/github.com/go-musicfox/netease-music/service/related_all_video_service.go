package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type RelatedAllVideoService struct {
	ID string `json:"id" form:"id"`
}

func (service *RelatedAllVideoService) RelatedAllVideo() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["id"] = service.ID
	data["type"] = "1"
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/cloudvideo/v1/allvideo/rcmd`, data, options)

	return code, reBody
}
