package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type MusicFirstListenInfoService struct {
	Id string `json:"id" form:"id"`
}

func (service *MusicFirstListenInfoService) MusicFirstListenInfo() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/content/activity/music/first/listen/info",
	}
	data := make(map[string]string)
	data["songId"] = service.Id

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/content/activity/music/first/listen/info", data, options)

	return code, reBody
}
