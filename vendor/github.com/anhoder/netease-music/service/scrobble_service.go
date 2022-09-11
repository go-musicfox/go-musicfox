package service

import (
	"encoding/json"
	"github.com/anhoder/netease-music/util"
)

type ScrobbleService struct {
	ID       string `json:"id" form:"id"`
	Sourceid string `json:"sourceid" form:"sourceid"`
	Time     string `json:"time" form:"time"`
}

func (service *ScrobbleService) Scrobble() (float64, []byte) {

	//errBody:=make(map[string]interface{})
	//errBody["code"]=500
	//errBody["err"]="此接口本后台暂未实现，，，欢迎pr"
	//return errBody

	options := &util.Options{
		Crypto:  "weapi",
	}
	data := make(map[string]string)

	jsonn := make(map[string]interface{})
	jsonn["download"] = 0
	jsonn["end"] = "playend"
	jsonn["id"] = service.ID
	jsonn["sourceId"] = service.Sourceid
	jsonn["time"] = service.Time
	jsonn["type"] = "song"
	jsonn["wifi"] = 0

	long := make(map[string]interface{})
	long["action"] = "play"
	long["json"] = jsonn

	var longs []map[string]interface{}
	longs = append(longs, long)

	if str, err := json.Marshal(longs); err != nil {
		return 502, []byte("参数错误")
	} else {
		data["long"] = string(str)
	}

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/feedback/weblog`, data, options)

	return code, reBody
}
