package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type HomepageDragonBallService struct {
}

func (service *HomepageDragonBallService) HomepageDragonBall() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/homepage/dragon/ball/static",
	}
	data := make(map[string]string)
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/eapi/homepage/dragon/ball/static`, data, options)

	return code, reBody
}
