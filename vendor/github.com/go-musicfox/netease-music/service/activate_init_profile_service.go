package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ActivateInitProfileService struct {
	Nickname string `json:"nickname" form:"nickname"`
}

func (service *ActivateInitProfileService) ActivateInitProfile() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/activate/initProfile",
	}
	data := make(map[string]string)
	data["nickname"] = service.Nickname

	code, reBody, _ := util.CreateRequest("POST", `http://music.163.com/eapi/activate/initProfile`, data, options)

	return code, reBody
}
