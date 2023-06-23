package service

import (
	"net/http"

	"github.com/go-musicfox/netease-music/util"
)

type LoginRefreshService struct {
}

func (service *LoginRefreshService) LoginRefresh() (float64, []byte) {

	cookiesOS := &http.Cookie{Name: "os", Value: "pc"}

	options := &util.Options{
		Crypto:  "weapi",
		Ua:      "pc",
		Cookies: []*http.Cookie{cookiesOS},
	}
	data := make(map[string]string)
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/login/token/refresh`, data, options)

	return code, reBody
}
