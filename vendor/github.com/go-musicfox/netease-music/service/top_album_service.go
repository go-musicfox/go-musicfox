package service

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-musicfox/netease-music/util"
)

type TopAlbumService struct {
	Area   string `json:"area" form:"area"` //ALL:全部,ZH:华语,EA:欧美,KR:韩国,JP:日本
	Limit  string `json:"limit" form:"limit"`
	Offset string `json:"offset" form:"offset"`
	Type   string `json:"type" form:"type"`
	Year   string `json:"year" form:"year"`
	Month  string `json:"month" form:"month"`
}

func (service *TopAlbumService) TopAlbum() (float64, []byte) {

	cookiesOS := &http.Cookie{Name: "os", Value: "pc"}

	options := &util.Options{
		Crypto:  "weapi",
		Cookies: []*http.Cookie{cookiesOS},
	}
	data := make(map[string]string)

	if service.Area == "" {
		service.Area = "ALL"
	}
	if service.Limit == "" {
		service.Limit = "50"
	}
	if service.Offset == "" {
		service.Offset = "0"
	}
	if service.Type == "" {
		service.Type = "new"
	}
	if service.Year == "" {
		service.Year = strconv.Itoa(time.Now().Year())
	}
	if service.Month == "" {
		service.Month = strconv.Itoa(int(time.Now().Month()))
	}
	data["area"] = service.Area
	data["limit"] = service.Limit
	data["offset"] = service.Offset
	data["type"] = service.Type
	data["year"] = service.Year
	data["month"] = service.Month
	data["total"] = "true"
	data["rcmd"] = "false"

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/discovery/new/albums/area`, data, options)

	return code, reBody
}
