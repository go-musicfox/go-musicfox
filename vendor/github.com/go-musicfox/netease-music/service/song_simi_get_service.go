package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SongSimiGetService struct {
	Id string `json:"id" form:"id"`
}

func (service *SongSimiGetService) SongSimiGet() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/link/position/show/resource",
	}
	data := make(map[string]string)
	data["positionCode"] = "toolBarRcmdSong"
	data["resourceId"] = service.Id
	data["resourceType"] = "song"

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/link/position/show/resource", data, options)

	return code, reBody
}
