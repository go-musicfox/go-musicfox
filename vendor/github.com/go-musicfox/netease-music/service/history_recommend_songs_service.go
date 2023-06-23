package service

import (
	"net/http"

	"github.com/go-musicfox/netease-music/util"
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
