package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type PlaylistImportNameTaskCreateService struct {
	ImportStarPlaylist string `json:"importStarPlaylist" form:"importStarPlaylist"`
}

func (service *PlaylistImportNameTaskCreateService) PlaylistImportNameTaskCreate() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/playlist/import/name/task/create",
	}
	data := make(map[string]string)
	if service.ImportStarPlaylist == "" {
		data["importStarPlaylist"] = "false"
	} else {
		data["importStarPlaylist"] = service.ImportStarPlaylist
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/playlist/import/name/task/create", data, options)

	return code, reBody
}
