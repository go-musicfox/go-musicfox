package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type PlaylistNameUpdateService struct {
	Id   string `json:"id" form:"id"`
	Name string `json:"desc" form:"name"`
}

func (service *PlaylistNameUpdateService) PlaylistNameUpdate() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/playlist/update/name",
	}
	data := make(map[string]string)
	data["id"] = service.Id
	data["name"] = service.Name
	code, reBody, _ := util.CreateRequest("POST", `http://interface3.music.163.com/eapi/playlist/update/name`, data, options)

	return code, reBody
}
