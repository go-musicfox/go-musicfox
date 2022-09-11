package service

import (
	"github.com/anhoder/netease-music/util"
	"net/http"
)

type PlaylistDeleteService struct {
	ID string `json:"id" form:"id"`
}

func (service *PlaylistDeleteService) PlaylistDelete() (float64, []byte) {

	cookiesOS := &http.Cookie{Name: "os", Value: "pc"}

	options := &util.Options{
		Crypto:  "weapi",
		Cookies: []*http.Cookie{cookiesOS},
	}
	data := make(map[string]string)
	data["ids"] = "[" + service.ID + "]"

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/playlist/remove`, data, options)

	return code, reBody
}
