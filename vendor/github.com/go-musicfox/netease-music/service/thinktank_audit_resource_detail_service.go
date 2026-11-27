package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ThinktankAuditResourceDetailService struct {
	Type string `json:"type" form:"type"`
}

func (service *ThinktankAuditResourceDetailService) ThinktankAuditResourceDetail() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/thinktank/audit/resource/detail",
	}
	data := make(map[string]string)
	if service.Type == "" {
		data["type"] = "4"
	} else {
		data["type"] = service.Type
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/thinktank/audit/resource/detail", data, options)

	return code, reBody
}
