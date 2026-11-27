package service

import (
	"encoding/json"
	"strings"

	"github.com/go-musicfox/netease-music/util"
)

type CommentInfoListService struct {
	Type string `json:"type" form:"type"`
	Ids  string `json:"ids" form:"ids"`
}

var commentResourceTypeIDs = map[string]string{
	"0": "4",
	"1": "5",
	"2": "0",
	"3": "3",
	"4": "1",
	"5": "62",
	"6": "2",
	"7": "14",
}

func (service *CommentInfoListService) CommentInfoList() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	resourceType, ok := commentResourceTypeIDs[service.Type]
	if !ok {
		resourceType = commentResourceTypeIDs["0"]
	}
	data["resourceType"] = resourceType

	ids := make([]string, 0)
	for _, id := range strings.Split(service.Ids, ",") {
		if id = strings.TrimSpace(id); id != "" {
			ids = append(ids, id)
		}
	}
	resourceIDs, _ := json.Marshal(ids)
	data["resourceIds"] = string(resourceIDs)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/resource/commentInfo/list", data, options)

	return code, reBody
}
