package service

import (
	"encoding/json"

	"github.com/go-musicfox/netease-music/util"
)

type MlogMusicRcmdService struct {
	Mvid   string `json:"mvid" form:"mvid"`
	Songid string `json:"songid" form:"songid"`
	Limit  string `json:"limit" form:"limit"`
}

func (service *MlogMusicRcmdService) MlogMusicRcmd() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/mlog/rcmd/feed/list",
	}
	extInfo, _ := json.Marshal(map[string]string{"songId": service.Songid})
	data := make(map[string]string)
	if service.Mvid == "" {
		data["id"] = "0"
	} else {
		data["id"] = service.Mvid
	}
	data["type"] = "2"
	data["rcmdType"] = "20"
	if service.Limit == "" {
		data["limit"] = "10"
	} else {
		data["limit"] = service.Limit
	}
	data["extInfo"] = string(extInfo)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/mlog/rcmd/feed/list", data, options)

	return code, reBody
}
