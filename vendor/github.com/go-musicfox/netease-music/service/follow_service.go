package service

import (
	"net/http"

	"github.com/go-musicfox/netease-music/util"
)

type FollowService struct {
	T  string `json:"t" form:"t"`
	Id string `json:"id" form:"id"`
}

func (service *FollowService) Follow() (float64, []byte) {

	cookiesOS := &http.Cookie{Name: "os", Value: "pc"}

	options := &util.Options{
		Crypto:  "weapi",
		Cookies: []*http.Cookie{cookiesOS},
	}
	data := make(map[string]string)
	if service.T == "1" {
		service.T = "follow"
	} else {
		service.T = "delfollow"
	}

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/user/`+service.T+`/`+service.Id, data, options)

	return code, reBody
}
