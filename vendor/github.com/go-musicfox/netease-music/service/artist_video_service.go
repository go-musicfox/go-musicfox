package service

import (
	"encoding/json"
	"strconv"

	"github.com/go-musicfox/netease-music/util"
)

type ArtistVideoService struct {
	Id     string `json:"id" form:"id"`
	Size   string `json:"size" form:"size"`
	Cursor string `json:"cursor" form:"cursor"`
	Order  string `json:"order" form:"order"`
}

func (service *ArtistVideoService) ArtistVideo() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	size, _ := strconv.Atoi(service.Size)
	if size == 0 {
		size = 10
	}
	cursor, _ := strconv.Atoi(service.Cursor)
	page, _ := json.Marshal(map[string]int{
		"size":   size,
		"cursor": cursor,
	})

	data := make(map[string]string)
	data["artistId"] = service.Id
	data["page"] = string(page)
	data["tab"] = "0"
	if service.Order == "" {
		data["order"] = "0"
	} else {
		data["order"] = service.Order
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/mlog/artist/video", data, options)

	return code, reBody
}
