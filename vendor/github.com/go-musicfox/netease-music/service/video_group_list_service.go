package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type VideoGroupListService struct {
}

func (service *VideoGroupListService) VideoGroupList() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/cloudvideo/group/list`, data, options)

	return code, reBody
}
