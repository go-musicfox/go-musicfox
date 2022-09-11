package service

import (
	"github.com/anhoder/netease-music/util"
	"net/http"
)

type HistoryRecommendSongsService struct {
}

func (service *HistoryRecommendSongsService) HistoryRecommendSongs() (float64, []byte) {

	cookiesOS := &http.Cookie{Name: "os", Value: "ios"}

	options := &util.Options{
		Crypto:  "weapi",
		Cookies: []*http.Cookie{cookiesOS},
	}
	data := make(map[string]string)
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/discovery/recommend/songs/history/recent`, data, options)

	return code, reBody
}
