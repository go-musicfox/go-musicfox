package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type AdListeningRightsService struct {
}

func (service *AdListeningRightsService) AdListeningRights() (float64, []byte) {

	options := &util.Options{
		Crypto: "xeapi",
		Url:    "/api/ad/homepage/free/tab/extend/v2",
	}
	data := map[string]string{"entrance": "FREE_LISTEN_RN"}

	code, reBody, _ := util.CreateRequest("POST", "https://interface3.music.163.com/xeapi/ad/homepage/free/tab/extend/v2", data, options)

	return code, reBody
}
