package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SigninProgressService struct {
	ModuleId string `json:"moduleId" form:"moduleId"`
}

func (service *SigninProgressService) SigninProgress() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.ModuleId == "" {
		data["moduleId"] = "1207signin-1207signin"
	} else {
		data["moduleId"] = service.ModuleId
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/act/modules/signin/v2/progress", data, options)

	return code, reBody
}
