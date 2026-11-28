package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type MusicianDataOverviewService struct {
}

func (service *MusicianDataOverviewService) MusicianDataOverview() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/creator/musician/statistic/data/overview/get", data, options)

	return code, reBody
}
