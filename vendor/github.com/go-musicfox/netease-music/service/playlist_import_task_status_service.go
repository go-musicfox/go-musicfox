package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type PlaylistImportTaskStatusService struct {
	Id string `json:"id" form:"id"`
}

func (service *PlaylistImportTaskStatusService) PlaylistImportTaskStatus() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/playlist/import/task/status/v2",
	}
	data := make(map[string]string)
	data["taskIds"] = "[\"" + service.Id + "\"]"

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/playlist/import/task/status/v2", data, options)

	return code, reBody
}
