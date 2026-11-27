package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ListentogetherRoomCheckService struct {
	RoomId string `json:"roomId" form:"roomId"`
}

func (service *ListentogetherRoomCheckService) ListentogetherRoomCheck() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/listen/together/room/check",
	}
	data := make(map[string]string)
	data["roomId"] = service.RoomId

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/listen/together/room/check", data, options)

	return code, reBody
}
