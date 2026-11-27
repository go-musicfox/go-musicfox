package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type ThinktankAuditResourceUpdateService struct {
	Judgement string `json:"judgement" form:"judgement"`
	TaskId    string `json:"taskId" form:"taskId"`
	Type      string `json:"type" form:"type"`
}

func (service *ThinktankAuditResourceUpdateService) ThinktankAuditResourceUpdate() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/thinktank/audit/resource/update",
	}
	data := make(map[string]string)
	if service.Type == "" {
		data["type"] = "4"
	} else {
		data["type"] = service.Type
	}
	data["taskId"] = service.TaskId
	data["judgement"] = service.Judgement

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/thinktank/audit/resource/update", data, options)

	return code, reBody
}
