package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type RecordRecentSongsService struct {
	Limit string `json:"limit" form:"limit"`
}

func (service *RecordRecentSongsService) RecordRecentSongs() (float64, []byte, error) {
	data := make(map[string]any)
	if service.Limit == "" {
		data["limit"] = "100"
	} else {
		data["limit"] = service.Limit
	}
	api := "https://music.163.com/api/play-record/song/list"
	code, reBody, err := util.CallWeapi(api, data)

	return code, reBody, err
}
