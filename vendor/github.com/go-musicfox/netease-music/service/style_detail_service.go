package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type StyleDetailService struct {
	TagId string `json:"tagId" form:"tagId"`
}

func (service *StyleDetailService) StyleDetail() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["tagId"] = service.TagId

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/style-tag/home/head", data, options)

	return code, reBody
}
