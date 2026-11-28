package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ListentogetherEndService struct {
	RoomId string `json:"roomId" form:"roomId"`
}

func (service *ListentogetherEndService) ListentogetherEnd() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/listen/together/end/v2",
	}
	data := make(map[string]string)
	data["roomId"] = service.RoomId

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/listen/together/end/v2", data, options)

	return code, reBody
}
