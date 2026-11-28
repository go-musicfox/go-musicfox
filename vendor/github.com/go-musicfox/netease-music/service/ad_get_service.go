package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type AdGetService struct {
	TypeIds string `json:"type_ids" form:"type_ids"`
}

func (service *AdGetService) AdGet() (float64, []byte) {

	options := &util.Options{
		Crypto:     "xeapi",
		Url:        "/api/ad/get",
		CheckToken: "v3",
	}
	data := make(map[string]string)
	if service.TypeIds == "" {
		data["type_ids"] = `["400002_0"]`
	} else {
		data["type_ids"] = service.TypeIds
	}

	code, reBody, _ := util.CreateRequest("POST", "https://interface3.music.163.com/xeapi/ad/get", data, options)

	return code, reBody
}
