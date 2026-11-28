package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type MusicianCloudbeanService struct {
}

func (service *MusicianCloudbeanService) MusicianCloudbean() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/cloudbean/get", data, options)

	return code, reBody
}
