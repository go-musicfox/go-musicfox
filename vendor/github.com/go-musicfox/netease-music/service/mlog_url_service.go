package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type MlogUrlService struct {
	Id  string `json:"id" form:"id"`
	Res string `json:"res" form:"res"`
}

func (service *MlogUrlService) MlogUrl() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["id"] = service.Id
	if service.Res == "" {
		data["resolution"] = "1080"
	} else {
		data["resolution"] = service.Res
	}
	data["type"] = "1"

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/mlog/detail/v1", data, options)

	return code, reBody
}
