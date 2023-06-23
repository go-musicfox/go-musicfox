package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type DjSubService struct {
	RID string `json:"rid" form:"rid"`
	T   string `json:"t" form:"t"`
}

func (service *DjSubService) DjSub() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["id"] = service.RID
	if service.T == "1" {
		service.T = "sub"
	} else {
		service.T = "unsub"
	}
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/djradio/`+service.T, data, options)

	return code, reBody
}
