package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type RadioSportGetService struct {
	Bpm string `json:"bpm" form:"bpm"`
}

func (service *RadioSportGetService) RadioSportGet() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/radio/sport/get",
	}
	data := make(map[string]string)
	if service.Bpm == "" {
		data["bpm"] = "50"
	} else {
		data["bpm"] = service.Bpm
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/radio/sport/get", data, options)

	return code, reBody
}
