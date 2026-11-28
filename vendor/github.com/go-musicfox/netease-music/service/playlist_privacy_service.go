package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type PlaylistPrivacyService struct {
	Id string `json:"id" form:"id"`
}

func (service *PlaylistPrivacyService) PlaylistPrivacy() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/playlist/update/privacy",
	}
	data := make(map[string]string)
	data["id"] = service.Id
	data["privacy"] = "0"

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/playlist/update/privacy", data, options)

	return code, reBody
}
