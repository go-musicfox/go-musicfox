package service

import (
	"net/http"

	"github.com/anhoder/netease-music/util"
)

type SongUrlService struct {
	ID string `json:"id" form:"id"`
	Br string `json:"br" form:"br"`
}

func (service *SongUrlService) SongUrl() (float64, []byte) {

	cookiesOS := &http.Cookie{Name: "os", Value: "pc"}

	options := &util.Options{
		Crypto:  "linuxapi",
		Cookies: []*http.Cookie{cookiesOS},
	}
	data := make(map[string]string)
	data["ids"] = "[" + service.ID + "]"
	if service.Br == "" {
		service.Br = "320000"
	}
	data["br"] = service.Br

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/song/enhance/player/url`, data, options)

	return code, reBody
}
