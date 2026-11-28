package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type PlaylistDetailDynamicService struct {
	Id string `json:"id" form:"id"`
	S  string `json:"s" form:"s"`
}

func (service *PlaylistDetailDynamicService) PlaylistDetailDynamic() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/playlist/detail/dynamic",
	}
	data := make(map[string]string)
	data["id"] = service.Id
	data["n"] = "100000"
	if service.S == "" {
		data["s"] = "8"
	} else {
		data["s"] = service.S
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/playlist/detail/dynamic", data, options)

	return code, reBody
}
