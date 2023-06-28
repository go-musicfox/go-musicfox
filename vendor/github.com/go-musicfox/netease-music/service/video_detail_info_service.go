package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type VideoDetailInfoService struct {
	ID string `json:"vid" form:"vid"`
}

func (service *VideoDetailInfoService) VideoDetailInfo() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["threadid"] = "R_VI_62_" + service.ID
	data["composeliked"] = "true"
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/comment/commentthread/info`, data, options)

	return code, reBody
}
