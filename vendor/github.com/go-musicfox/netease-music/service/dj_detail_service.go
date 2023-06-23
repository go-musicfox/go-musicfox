package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type DjDetailService struct {
	ID string `json:"rid" form:"rid"`
}

func (service *DjDetailService) DjDetail() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["id"] = service.ID
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/djradio/get`, data, options)

	return code, reBody
}
