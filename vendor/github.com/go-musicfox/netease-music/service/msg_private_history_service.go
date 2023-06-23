package service

import (
	"net/http"

	"github.com/go-musicfox/netease-music/util"
)

type MsgPrivateHistoryService struct {
	UID   string `json:"uid" form:"uid"`
	Limit string `json:"limit" form:"limit"`
	Time  string `json:"before" form:"before"`
}

func (service *MsgPrivateHistoryService) MsgPrivateHistory() (float64, []byte) {

	cookiesOS := &http.Cookie{Name: "os", Value: "pc"}

	options := &util.Options{
		Crypto:  "weapi",
		Cookies: []*http.Cookie{cookiesOS},
	}
	data := make(map[string]string)
	data["userId"] = service.UID
	if service.Limit == "" {
		data["limit"] = "30"
	} else {
		data["limit"] = service.Limit
	}
	if service.Time == "" {
		data["offset"] = "0"
	} else {
		data["offset"] = service.Time
	}
	data["order"] = "true"
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/msg/private/history`, data, options)

	return code, reBody
}
