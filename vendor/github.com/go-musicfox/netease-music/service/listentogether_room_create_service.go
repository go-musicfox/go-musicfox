package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ListentogetherRoomCreateService struct {
}

func (service *ListentogetherRoomCreateService) ListentogetherRoomCreate() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/listen/together/room/create",
	}
	data := make(map[string]string)
	data["refer"] = "songplay_more"

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/listen/together/room/create", data, options)

	return code, reBody
}
