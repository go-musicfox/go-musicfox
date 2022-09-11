package service

import (
	"github.com/anhoder/netease-music/util"
	"net/http"
)

type HistoryRecommendDongsDetailService struct {
	Date string `json:"date" form:"date"`
}

func (service *HistoryRecommendDongsDetailService) HistoryRecommendDongsDetail() (float64, []byte) {

	cookiesOS := &http.Cookie{Name: "os", Value: "ios"}

	options := &util.Options{
		Crypto:  "weapi",
		Cookies: []*http.Cookie{cookiesOS},
	}
	data := make(map[string]string)
	data["date"] = service.Date

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/discovery/recommend/songs/history/detail`, data, options)

	return code, reBody
}
