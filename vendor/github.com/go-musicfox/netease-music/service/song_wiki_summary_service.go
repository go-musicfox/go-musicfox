package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SongWikiSummaryService struct {
	Id string `json:"id" form:"id"`
}

func (service *SongWikiSummaryService) SongWikiSummary() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/song/play/about/block/page",
	}
	data := make(map[string]string)
	data["songId"] = service.Id

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/song/play/about/block/page", data, options)

	return code, reBody
}
