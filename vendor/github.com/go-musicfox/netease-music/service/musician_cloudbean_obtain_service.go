package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type MusicianCloudbeanObtainService struct {
	Id     string `json:"id" form:"id"`
	Period string `json:"period" form:"period"`
}

func (service *MusicianCloudbeanObtainService) MusicianCloudbeanObtain() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["userMissionId"] = service.Id
	data["period"] = service.Period

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/nmusician/workbench/mission/reward/obtain/new", data, options)

	return code, reBody
}
