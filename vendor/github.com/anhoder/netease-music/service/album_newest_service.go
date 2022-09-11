package service

import (
	"github.com/anhoder/netease-music/util"
	"net/http"
)

type AlbumNewestService struct {}

func (service *AlbumNewestService) AlbumNewest() (float64, []byte) {

	cookiesOS := &http.Cookie{Name: "os", Value: "pc"}

	options := &util.Options{
		Crypto:  "weapi",
		Cookies: []*http.Cookie{cookiesOS},
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/discovery/newAlbum`, data, options)

	return code, reBody
}
