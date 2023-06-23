package service

import (
	"net/http"

	"github.com/go-musicfox/netease-music/util"
)

type LikeService struct {
	ID string `json:"id" form:"id"`
	L  string `json:"like" form:"like"`
}

func (service *LikeService) Like() (float64, []byte) {
	options := &util.Options{
		Crypto: "weapi",
		Cookies: []*http.Cookie{
			{Name: "os", Value: "pc"},
			{Name: "appver", Value: "2.7.1.198277"},
		},
	}
	data := make(map[string]string)
	data["trackId"] = service.ID
	data["alg"] = "itembased"
	data["time"] = "3"
	if service.L == "" {
		data["like"] = "true"
	} else {
		data["like"] = service.L
	}

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/radio/like`, data, options)

	return code, reBody
}
