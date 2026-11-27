package service

import (
	"strconv"
	"time"

	"github.com/go-musicfox/netease-music/util"
)

type CalendarService struct {
	StartTime string `json:"startTime" form:"startTime"`
	EndTime   string `json:"endTime" form:"endTime"`
}

func (service *CalendarService) Calendar() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.StartTime == "" {
		data["startTime"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	} else {
		data["startTime"] = service.StartTime
	}
	if service.EndTime == "" {
		data["endTime"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	} else {
		data["endTime"] = service.EndTime
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/mcalendar/detail", data, options)

	return code, reBody
}
