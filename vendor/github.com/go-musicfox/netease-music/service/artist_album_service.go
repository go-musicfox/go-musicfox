package service

import (
	"net/http"

	"github.com/go-musicfox/netease-music/util"
)

type ArtistAlbumService struct {
	ID     string `json:"id" form:"id"`
	Limit  string `json:"limit" form:"limit"`
	Offset string `json:"offset" form:"offset"`
}

func (service *ArtistAlbumService) ArtistAlbum() (float64, []byte) {

	cookiesOS := &http.Cookie{Name: "os", Value: "pc"}

	options := &util.Options{
		Crypto:  "weapi",
		Cookies: []*http.Cookie{cookiesOS},
	}
	data := make(map[string]string)
	if service.Limit == "" {
		data["limit"] = "30"
	} else {
		data["limit"] = service.Limit
	}
	if service.Offset == "" {
		data["offset"] = "0"
	} else {
		data["offset"] = service.Offset
	}
	data["total"] = "true"
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/artist/albums/`+service.ID, data, options)

	return code, reBody
}
