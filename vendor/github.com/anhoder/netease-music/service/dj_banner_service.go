package service

import (
	"github.com/anhoder/netease-music/util"
	"net/http"
)

type DjBannerService struct {
}

func (service *DjBannerService) DjBanner() (float64, []byte) {

	cookiesOS := &http.Cookie{Name: "os", Value: "pc"}

	options := &util.Options{
		Crypto:  "weapi",
		Cookies: []*http.Cookie{cookiesOS},
	}
	data := make(map[string]string)
	code, reBody, _ := util.CreateRequest("POST", `http://music.163.com/weapi/djradio/banner/get`, data, options)

	return code, reBody
}
