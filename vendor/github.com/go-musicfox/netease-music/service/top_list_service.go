package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type TopListService struct {
	Id string `json:"id" form:"id"`
}

func (service *TopListService) TopList() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/playlist/v4/detail",
	}
	data := make(map[string]string)
	data["id"] = service.Id
	data["n"] = "500"
	data["s"] = "0"

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/playlist/v4/detail", data, options)

	return code, reBody
}
