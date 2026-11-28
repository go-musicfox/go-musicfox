package service

import (
	"encoding/json"

	"github.com/go-musicfox/netease-music/util"
)

type SongWikiInfoService struct {
	Id string `json:"id" form:"id"`
}

func (service *SongWikiInfoService) SongWikiInfo() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/link/page/parent/relation/construct/info",
	}
	extJson, _ := json.Marshal(map[string]any{
		"states": map[string]any{
			"playingResource": map[string]string{
				"current": service.Id,
				"scene":   "songWiki",
			},
		},
	})
	data := make(map[string]string)
	data["extJson"] = string(extJson)
	data["positionCode"] = "songWikiMainPosition"

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/link/page/parent/relation/construct/info", data, options)

	return code, reBody
}
