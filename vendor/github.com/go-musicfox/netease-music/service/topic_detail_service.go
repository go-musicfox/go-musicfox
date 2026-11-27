package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type TopicDetailService struct {
	Actid string `json:"actid" form:"actid"`
}

func (service *TopicDetailService) TopicDetail() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["actid"] = service.Actid

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/act/detail", data, options)

	return code, reBody
}
