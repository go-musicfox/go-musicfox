package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type EventPrivacyService struct {
	EvId    string `json:"evId" form:"evId"`
	Privacy string `json:"privacy" form:"privacy"`
}

func (service *EventPrivacyService) EventPrivacy() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/event/privacy/op",
	}
	data := make(map[string]string)
	data["eventId"] = service.EvId
	data["privacy"] = service.Privacy

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/event/privacy/op", data, options)

	return code, reBody
}
