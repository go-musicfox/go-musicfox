package service

import (
	"github.com/anhoder/netease-music/util"
	"net/http"
)

type RecommendSongsService struct {
	ID string `json:"id" form:"id"`
}

func (service *RecommendSongsService) RecommendSongs() (float64, []byte) {

	cookiesOS := &http.Cookie{Name: "os", Value: "ios"}

	options := &util.Options{
		Crypto:  "weapi",
		Cookies: []*http.Cookie{cookiesOS},
	}
	data := make(map[string]string)
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/v3/discovery/recommend/songs`, data, options)

	return code, reBody
}
