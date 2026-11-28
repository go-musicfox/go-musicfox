package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type MusicianVipTasksService struct {
}

func (service *MusicianVipTasksService) MusicianVipTasks() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/nmusician/workbench/special/right/vip/info",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/nmusician/workbench/special/right/vip/info", data, options)

	return code, reBody
}
