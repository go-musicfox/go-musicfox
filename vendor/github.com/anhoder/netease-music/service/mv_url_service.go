package service

import (
	"github.com/anhoder/netease-music/util"
	"net/http"
)

type MvUrlService struct {
	ID string `json:"id" form:"id"`
	R  string `json:"r" form:"r"`
}

func (service *MvUrlService) MvUrl() (float64, []byte) {

	cookiesOS := &http.Cookie{Name: "os", Value: "pc"}

	options := &util.Options{
		Crypto:  "weapi",
		Cookies: []*http.Cookie{cookiesOS},
	}
	data := make(map[string]string)
	data["id"] = service.ID
	if service.R == "" {
		data["r"] = "1080"
	} else {
		data["r"] = service.R
	}

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/song/enhance/play/mv/url`, data, options)

	return code, reBody
}
