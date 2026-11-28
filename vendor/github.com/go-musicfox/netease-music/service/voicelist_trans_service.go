package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type VoicelistTransService struct {
	Limit     string `json:"limit" form:"limit"`
	Offset    string `json:"offset" form:"offset"`
	Position  string `json:"position" form:"position"`
	ProgramId string `json:"programId" form:"programId"`
	RadioId   string `json:"radioId" form:"radioId"`
}

func (service *VoicelistTransService) VoicelistTrans() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/voice/workbench/radio/program/trans",
	}
	data := make(map[string]string)
	if service.Limit == "" {
		data["limit"] = "200"
	} else {
		data["limit"] = service.Limit
	}
	if service.Offset == "" {
		data["offset"] = "0"
	} else {
		data["offset"] = service.Offset
	}
	if service.RadioId == "" {
		data["radioId"] = ""
	} else {
		data["radioId"] = service.RadioId
	}
	if service.ProgramId == "" {
		data["programId"] = "0"
	} else {
		data["programId"] = service.ProgramId
	}
	if service.Position == "" {
		data["position"] = "1"
	} else {
		data["position"] = service.Position
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/voice/workbench/radio/program/trans", data, options)

	return code, reBody
}
