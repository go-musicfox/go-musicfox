package service

import (
	"strconv"
	"time"

	"github.com/go-musicfox/netease-music/util"
)

type FanscenterTrendListService struct {
	StartTime string `json:"startTime" form:"startTime"`
	EndTime   string `json:"endTime" form:"endTime"`
	Type      string `json:"type" form:"type"`
}

func (service *FanscenterTrendListService) FanscenterTrendList() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/fanscenter/trend/list",
	}
	data := make(map[string]string)
	if service.StartTime == "" {
		data["startTime"] = strconv.FormatInt(time.Now().Add(-7*24*time.Hour).UnixMilli(), 10)
	} else {
		data["startTime"] = service.StartTime
	}
	if service.EndTime == "" {
		data["endTime"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	} else {
		data["endTime"] = service.EndTime
	}
	if service.Type == "" {
		data["type"] = "0"
	} else {
		data["type"] = service.Type
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/fanscenter/trend/list", data, options)

	return code, reBody
}
