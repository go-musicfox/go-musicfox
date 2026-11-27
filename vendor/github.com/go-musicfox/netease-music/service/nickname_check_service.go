package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type NicknameCheckService struct {
	Nickname string `json:"nickname" form:"nickname"`
}

func (service *NicknameCheckService) NicknameCheck() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["nickname"] = service.Nickname

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/nickname/duplicated", data, options)

	return code, reBody
}
