package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ListentogetherStatusService struct {
}

func (service *ListentogetherStatusService) ListentogetherStatus() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/listen/together/status/get", data, options)

	return code, reBody
}
