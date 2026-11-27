package service

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/go-musicfox/netease-music/util"
)

type AidjContentRcmdService struct {
	Latitude  string `json:"latitude" form:"latitude"`
	Longitude string `json:"longitude" form:"longitude"`
}

func (service *AidjContentRcmdService) AidjContentRcmd() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/aidj/content/rcmd/info",
	}
	extInfo := map[string]any{
		"noAidjToAidj":         false,
		"lastRequestTimestamp": time.Now().UnixMilli(),
		"listenedTs":           false,
	}
	if service.Latitude != "" {
		lat, _ := strconv.ParseFloat(service.Latitude, 64)
		lon, _ := strconv.ParseFloat(service.Longitude, 64)
		extInfo["lbsInfoList"] = []map[string]any{{
			"lat":  lat,
			"lon":  lon,
			"time": time.Now().Unix(),
		}}
	}
	ext, _ := json.Marshal(extInfo)

	data := make(map[string]string)
	data["extInfo"] = string(ext)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/aidj/content/rcmd/info", data, options)

	return code, reBody
}
