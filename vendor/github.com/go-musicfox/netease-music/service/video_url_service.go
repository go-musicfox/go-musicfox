package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type VideoUrlService struct {
	ID  string `json:"id" form:"id"`
	Res string `json:"resolution" form:"resolution"`
}

func (service *VideoUrlService) VideoUrl() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.Res == "" {
		data["resolution"] = "1080"
	} else {
		data["resolution"] = service.Res
	}
	data["ids"] = `["` + service.ID + `"]`
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/cloudvideo/playurl`, data, options)

	return code, reBody
}
