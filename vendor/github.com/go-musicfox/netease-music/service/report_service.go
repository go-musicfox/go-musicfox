package service

import (
	"encoding/json"
	"strconv"

	"github.com/go-musicfox/netease-music/util"
)

type ReportService struct {
	ID         int64  `json:"id" form:"id"`
	Type       string `json:"type" form:"type"`             // song:歌曲, dj:播客
	SourceType string `json:"sourceType" form:"sourceType"` // list, album, dailySongRecommend, userfm
	SourceId   string `json:"sourceId" form:"sourceId"`
	Time       int64  `json:"time" form:"time"`
	Alg        string `json:"alg" form:"alg"`
	EndType    string `json:"endType" form:"endType"` // playend：正常结束；interrupt：第三方APP打断： exception: 错误； ui: 用户切歌
}

func (service *ReportService) Playend() (float64, []byte) {
	options := &util.Options{
		Crypto: "weapi",
	}

	if service.EndType == "" {
		service.EndType = "playend"
	}
	if service.Type == "" {
		service.Type = "song"
	}

	jsonData := map[string]interface{}{
		"type":     service.Type,
		"wifi":     0,
		"download": 0,
		"id":       service.ID,
		"time":     service.Time,
		"end":      service.EndType,
		"source":   service.SourceType,
		"mainsite": "1",
		"content":  "",
	}

	if service.SourceId != "" {
		jsonData["sourceId"] = service.SourceId
		if _, err := strconv.ParseInt(service.SourceId, 10, 64); err == nil {
			jsonData["content"] = "id=" + service.SourceId
		}
	}

	if service.Alg != "" {
		jsonData["alg"] = service.Alg
	}

	logs := []map[string]interface{}{
		{
			"action": "play",
			"json":   jsonData,
		},
	}

	data := make(map[string]string)
	if str, err := json.Marshal(logs); err == nil {
		data["logs"] = string(str)
	}

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/feedback/weblog`, data, options)

	return code, reBody
}

func (service *ReportService) Playstart() (float64, []byte) {
	options := &util.Options{
		Crypto: "weapi",
	}

	if service.Type == "" {
		service.Type = "song"
	}

	jsonData := map[string]interface{}{
		"id":       service.ID,
		"type":     service.Type,
		"content":  "",
		"mainsite": "1",
	}

	if _, err := strconv.ParseInt(service.SourceId, 10, 64); err == nil {
		jsonData["content"] = "id=" + service.SourceId
	}

	if service.Alg != "" {
		jsonData["alg"] = service.Alg
	}

	logs := []map[string]interface{}{
		{
			"action": "startplay",
			"json":   jsonData,
		},
	}

	data := make(map[string]string)
	if str, err := json.Marshal(logs); err == nil {
		data["logs"] = string(str)
	}

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/feedback/weblog`, data, options)

	return code, reBody
}
