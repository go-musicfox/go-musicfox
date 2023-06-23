package service

import (
	"net/http"

	"github.com/go-musicfox/netease-music/util"
)

type PlaylistOrderUpdateService struct {
	Ids string `json:"ids" form:"ids"`
}

func (service *PlaylistOrderUpdateService) PlaylistOrderUpdate() (float64, []byte) {

	cookiesOS := &http.Cookie{Name: "os", Value: "pc"}

	options := &util.Options{
		Crypto:  "weapi",
		Cookies: []*http.Cookie{cookiesOS},
	}
	data := make(map[string]string)
	data["id"] = service.Ids
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/playlist/order/update`, data, options)

	return code, reBody
}
