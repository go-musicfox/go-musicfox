package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type LikeListService struct {
	UID string `json:"uid" form:"uid"`
}

func (service *LikeListService) LikeList() (float64, []byte) {
	options := &util.Options{
		Crypto: "weapi",
	}

	data := make(map[string]string)
	data["uid"] = service.UID

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/song/like/get`, data, options)

	return code, reBody
}
