package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ListentogetherSyncPlaylistGetService struct {
	RoomId string `json:"roomId" form:"roomId"`
}

func (service *ListentogetherSyncPlaylistGetService) ListentogetherSyncPlaylistGet() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/listen/together/sync/playlist/get",
	}
	data := make(map[string]string)
	data["roomId"] = service.RoomId

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/listen/together/sync/playlist/get", data, options)

	return code, reBody
}
