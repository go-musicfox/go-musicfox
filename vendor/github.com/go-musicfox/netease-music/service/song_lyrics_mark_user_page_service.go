package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SongLyricsMarkUserPageService struct {
	Limit  string `json:"limit" form:"limit"`
	Offset string `json:"offset" form:"offset"`
}

func (service *SongLyricsMarkUserPageService) SongLyricsMarkUserPage() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/song/play/lyrics/mark/user/page",
	}
	data := make(map[string]string)
	if service.Limit == "" {
		data["limit"] = "10"
	} else {
		data["limit"] = service.Limit
	}
	if service.Offset == "" {
		data["offset"] = "0"
	} else {
		data["offset"] = service.Offset
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/song/play/lyrics/mark/user/page", data, options)

	return code, reBody
}
