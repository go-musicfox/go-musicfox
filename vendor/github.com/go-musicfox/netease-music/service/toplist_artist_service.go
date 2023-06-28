package service

import (
	"net/http"

	"github.com/go-musicfox/netease-music/util"
)

type ToplistArtistService struct {
	Type   string `json:"type" form:"type"`
	Limit  string `json:"limit" form:"limit"`
	Offset string `json:"offset" form:"offset"`
}

func (service *ToplistArtistService) ToplistArtist() (float64, []byte) {

	cookiesOS := &http.Cookie{Name: "os", Value: "pc"}

	options := &util.Options{
		Crypto:  "weapi",
		Cookies: []*http.Cookie{cookiesOS},
	}
	data := make(map[string]string)
	if service.Type == "" {
		data["type"] = "1"
	} else {
		data["type"] = service.Type
	}
	if service.Limit == "" {
		data["limit"] = "100"
	} else {
		data["limit"] = service.Limit
	}
	if service.Offset == "" {
		data["offset"] = "0"
	} else {
		data["offset"] = service.Offset
	}
	data["order"] = "true"
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/toplist/artist`, data, options)

	return code, reBody
}
