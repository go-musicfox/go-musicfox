package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type MvDetailService struct {
	ID string `json:"mvid" form:"mvid"`
}

func (service *MvDetailService) MvDetail() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["id"] = service.ID

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/v1/mv/detail`, data, options)

	return code, reBody
}
