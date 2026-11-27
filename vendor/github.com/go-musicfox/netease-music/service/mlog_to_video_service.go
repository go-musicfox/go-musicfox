package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type MlogToVideoService struct {
	Id string `json:"id" form:"id"`
}

func (service *MlogToVideoService) MlogToVideo() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["mlogId"] = service.Id

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/mlog/video/convert/id", data, options)

	return code, reBody
}
