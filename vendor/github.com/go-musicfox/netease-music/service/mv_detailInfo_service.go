package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type MvDetailInfoService struct {
	ID string `json:"mvid" form:"mvid"`
}

func (service *MvDetailInfoService) MvDetailInfo() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["threadid"] = "R_MV_5_" + service.ID
	data["composeliked"] = "true"
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/comment/commentthread/info`, data, options)

	return code, reBody
}
