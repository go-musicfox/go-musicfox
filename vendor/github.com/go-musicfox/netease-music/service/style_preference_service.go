package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type StylePreferenceService struct {
}

func (service *StylePreferenceService) StylePreference() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/tag/my/preference/get", data, options)

	return code, reBody
}
