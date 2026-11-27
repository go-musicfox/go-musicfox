package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ListentogetherAcceptService struct {
	InviterId string `json:"inviterId" form:"inviterId"`
	RoomId    string `json:"roomId" form:"roomId"`
}

func (service *ListentogetherAcceptService) ListentogetherAccept() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/listen/together/play/invitation/accept",
	}
	data := make(map[string]string)
	data["refer"] = "inbox_invite"
	data["roomId"] = service.RoomId
	data["inviterId"] = service.InviterId

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/listen/together/play/invitation/accept", data, options)

	return code, reBody
}
