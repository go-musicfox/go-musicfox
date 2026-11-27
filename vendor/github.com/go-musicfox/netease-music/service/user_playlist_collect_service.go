package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type UserPlaylistCollectService struct {
	Limit  string `json:"limit" form:"limit"`
	Offset string `json:"offset" form:"offset"`
	Uid    string `json:"uid" form:"uid"`
}

func (service *UserPlaylistCollectService) UserPlaylistCollect() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/user/playlist/collect",
	}
	data := make(map[string]string)
	if service.Limit == "" {
		data["limit"] = "100"
	} else {
		data["limit"] = service.Limit
	}
	if service.Offset == "" {
		data["offset"] = "0"
	} else {
		data["offset"] = service.Offset
	}
	data["userId"] = service.Uid
	data["isWebview"] = "true"
	data["includeRedHeart"] = "true"
	data["includeTop"] = "true"

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/user/playlist/collect", data, options)

	return code, reBody
}
