package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type MusicianTasksNewService struct {
}

func (service *MusicianTasksNewService) MusicianTasksNew() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/nmusician/workbench/mission/stage/list ", data, options)

	return code, reBody
}
