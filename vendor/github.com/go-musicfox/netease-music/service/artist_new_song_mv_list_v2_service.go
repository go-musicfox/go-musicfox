package service

import (
	"strconv"
	"time"

	"github.com/go-musicfox/netease-music/util"
)

type ArtistNewSongMvListV2Service struct {
	StartTimestamp string `json:"startTimestamp" form:"startTimestamp"`
	Before         string `json:"before" form:"before"`
	SourceType     string `json:"sourceType" form:"sourceType"`
	Limit          string `json:"limit" form:"limit"`
	FirstRequest   string `json:"firstRequest" form:"firstRequest"`
}

func (service *ArtistNewSongMvListV2Service) ArtistNewSongMvListV2() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/sub/artist/new/works/song-mv/list/v2",
	}
	data := make(map[string]string)
	if service.StartTimestamp == "" && service.Before == "" {
		data["startTimestamp"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	} else if service.StartTimestamp != "" {
		data["startTimestamp"] = service.StartTimestamp
	} else {
		data["startTimestamp"] = service.Before
	}
	if service.SourceType == "" {
		data["sourceType"] = "1"
	} else {
		data["sourceType"] = service.SourceType
	}
	if service.Limit == "" {
		data["limit"] = "10"
	} else {
		data["limit"] = service.Limit
	}
	if service.FirstRequest == "" {
		data["firstRequest"] = "true"
	} else {
		data["firstRequest"] = service.FirstRequest
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/sub/artist/new/works/song-mv/list/v2", data, options)

	return code, reBody
}
