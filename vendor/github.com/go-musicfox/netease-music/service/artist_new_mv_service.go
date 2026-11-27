package service

import (
	"strconv"
	"time"

	"github.com/go-musicfox/netease-music/util"
)

type ArtistNewMvService struct {
	Limit  string `json:"limit" form:"limit"`
	Before string `json:"before" form:"before"`
}

func (service *ArtistNewMvService) ArtistNewMv() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.Limit == "" {
		data["limit"] = "20"
	} else {
		data["limit"] = service.Limit
	}
	if service.Before == "" {
		data["startTimestamp"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	} else {
		data["startTimestamp"] = service.Before
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/sub/artist/new/works/mv/list", data, options)

	return code, reBody
}
