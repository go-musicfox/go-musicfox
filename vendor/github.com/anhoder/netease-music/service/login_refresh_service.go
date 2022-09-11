package service

import (
	"github.com/anhoder/netease-music/util"
	"net/http"
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
