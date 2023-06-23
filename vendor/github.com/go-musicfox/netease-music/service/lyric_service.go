package service

import (
	"net/http"

	"github.com/go-musicfox/netease-music/util"
)

type LyricService struct {
	ID string `json:"id" form:"id"`
}

func (service *LyricService) Lyric() (float64, []byte) {

	cookiesOS := &http.Cookie{Name: "os", Value: "pc"}

	options := &util.Options{
		Crypto:  "linuxapi",
		Cookies: []*http.Cookie{cookiesOS},
	}
	data := make(map[string]string)
	data["id"] = service.ID
	data["lv"] = "-1"
	data["kv"] = "-1"
	data["tv"] = "-1"

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/song/lyric`, data, options)

	return code, reBody
}
