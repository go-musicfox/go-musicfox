package service

import (
	"net/http"

	"github.com/go-musicfox/netease-music/util"
)

type ArtistSongsService struct {
	ID     string `json:"id" form:"id"`
	Limit  string `json:"limit" form:"limit"`
	Offset string `json:"offset" form:"offset"`
	Order  string `json:"order" form:"order"`
}

func (service *ArtistSongsService) ArtistSongs() (float64, []byte) {

	cookiesOS := &http.Cookie{Name: "os", Value: "pc"}

	options := &util.Options{
		Crypto:  "weapi",
		Cookies: []*http.Cookie{cookiesOS},
	}
	data := make(map[string]string)
	data["id"] = service.ID
	if service.Limit == "" {
		data["limit"] = "100"
	} else {
		data["limit"] = service.Limit
	}
	if service.Offset == "" {
		data["offset"] = "0"
	} else {
		data["offset"] = service.Offset
	}
	if service.Order == "" {
		data["order"] = "hot"
	} else {
		data["order"] = service.Order
	}
	data["work_type"] = "1"
	data["private_cloud"] = "true"
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/v1/artist/songs`, data, options)

	return code, reBody
}
