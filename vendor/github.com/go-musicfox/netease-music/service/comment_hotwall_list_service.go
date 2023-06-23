package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type CommentHotwallListService struct {
}

func (service *CommentHotwallListService) CommentHotwallList() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/comment/hotwall/list/get`, data, options)

	return code, reBody
}
