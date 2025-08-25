package service

import (
	"encoding/json"

	"github.com/go-musicfox/netease-music/util"
)

type ScrobbleService struct {
	ID       string `json:"id" form:"id"`
	Sourceid string `json:"sourceid" form:"sourceid"`
	Time     int64  `json:"time" form:"time"`
}

func (service *ScrobbleService) Scrobble() (float64, []byte, error) {

	var logs = []map[string]interface{}{
		{
			"action": "play",
			"json": map[string]interface{}{
				"download": 0,
				"end":      "playend",
				"id":       service.ID,
				"sourceId": service.Sourceid,
				"time":     service.Time,
				"type":     "song",
				"wifi":     1,
				"source":   "list",
			},
		},
	}

	var data = make(map[string]interface{})
	if str, err := json.Marshal(logs); err == nil {
		data["logs"] = string(str)
	}

	api := "https://clientlogusf.music.163.com/weapi/feedback/weblog"
	cookiejar := util.GetGlobalCookieJar()
	csrfToken := util.GetCsrfToken(cookiejar)
	data["csrf_token"] = csrfToken
	code, bodyBytes, err := util.CallWeapi(api+"?csrf_token="+csrfToken, data)
	if err != nil {
		return code, bodyBytes, err
	}
	return code, bodyBytes, nil
}
