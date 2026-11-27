package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type PlaylistDetailRcmdGetService struct {
	Id string `json:"id" form:"id"`
}

func (service *PlaylistDetailRcmdGetService) PlaylistDetailRcmdGet() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/playlist/detail/rcmd/get",
	}
	data := make(map[string]string)
	data["scene"] = "playlist_head"
	data["playlistId"] = service.Id
	data["newStyle"] = "true"

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/playlist/detail/rcmd/get", data, options)

	return code, reBody
}
