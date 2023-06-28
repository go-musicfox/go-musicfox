package service

import (
	"net/http"

	"github.com/go-musicfox/netease-music/util"
)

type EventDelService struct {
	EvId string `json:"evId" form:"evId"`
}

func (service *EventDelService) EventDel() (float64, []byte) {

	cookiesOS := &http.Cookie{Name: "os", Value: "pc"}

	options := &util.Options{
		Crypto:  "weapi",
		Cookies: []*http.Cookie{cookiesOS},
	}
	data := make(map[string]string)
	data["id"] = service.EvId

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/eapi/event/delete`, data, options)

	return code, reBody
}
