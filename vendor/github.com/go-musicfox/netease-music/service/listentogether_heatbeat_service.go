package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ListentogetherHeatbeatService struct {
	PlayStatus string `json:"playStatus" form:"playStatus"`
	Progress   string `json:"progress" form:"progress"`
	RoomId     string `json:"roomId" form:"roomId"`
	SongId     string `json:"songId" form:"songId"`
}

func (service *ListentogetherHeatbeatService) ListentogetherHeatbeat() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/listen/together/heartbeat",
	}
	data := make(map[string]string)
	data["roomId"] = service.RoomId
	data["songId"] = service.SongId
	data["playStatus"] = service.PlayStatus
	data["progress"] = service.Progress

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/listen/together/heartbeat", data, options)

	return code, reBody
}
