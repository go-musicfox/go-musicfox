package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type VoicelistListSearchService struct {
	DisplayStatus string `json:"displayStatus" form:"displayStatus"`
	Limit         string `json:"limit" form:"limit"`
	Name          string `json:"name" form:"name"`
	Offset        string `json:"offset" form:"offset"`
	Type          string `json:"type" form:"type"`
	VoiceFeeType  string `json:"voiceFeeType" form:"voiceFeeType"`
	VoiceListId   string `json:"voiceListId" form:"voiceListId"`
}

func (service *VoicelistListSearchService) VoicelistListSearch() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/voice/workbench/voice/list",
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
	if service.Name == "" {
		data["name"] = ""
	} else {
		data["name"] = service.Name
	}
	if service.DisplayStatus == "" {
		data["displayStatus"] = ""
	} else {
		data["displayStatus"] = service.DisplayStatus
	}
	if service.Type == "" {
		data["type"] = ""
	} else {
		data["type"] = service.Type
	}
	if service.VoiceFeeType == "" {
		data["voiceFeeType"] = ""
	} else {
		data["voiceFeeType"] = service.VoiceFeeType
	}
	data["radioId"] = service.VoiceListId

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/voice/workbench/voice/list", data, options)

	return code, reBody
}
