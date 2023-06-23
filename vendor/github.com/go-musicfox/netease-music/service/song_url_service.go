package service

import (
	"net/http"

	"github.com/go-musicfox/netease-music/util"
)

type SongUrlService struct {
	ID      string `json:"id" form:"id"`
	Br      string `json:"br" form:"br"`
	SkipUNM bool
}

func (service *SongUrlService) SongUrl() (float64, []byte) {

	cookiesOS := &http.Cookie{Name: "os", Value: "pc"}

	options := &util.Options{
		Crypto:  "linuxapi",
		Cookies: []*http.Cookie{cookiesOS},
		SkipUNM: service.SkipUNM,
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
