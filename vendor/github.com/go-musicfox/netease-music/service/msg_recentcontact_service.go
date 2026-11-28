package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type MsgRecentcontactService struct {
}

func (service *MsgRecentcontactService) MsgRecentcontact() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/msg/recentcontact/get", data, options)

	return code, reBody
}
