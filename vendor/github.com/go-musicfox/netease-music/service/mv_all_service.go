package service

import (
	"encoding/json"

	"github.com/go-musicfox/netease-music/util"
)

type MvAllService struct {
	Limit  string `json:"limit" form:"limit"`
	Offset string `json:"offset" form:"offset"`
	Area   string `json:"area" form:"area"`
	Type   string `json:"type" form:"type"`
	Order  string `json:"order" form:"order"`
}

func (service *MvAllService) MvAll() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	tag := make(map[string]string)
	if service.Area == "" {
		service.Area = "全部"
	}
	if service.Type == "" {
		service.Type = "全部"
	}
	if service.Order == "" {
		service.Order = "上升最快"
	}
	tag["地区"] = service.Area
	tag["类型"] = service.Type
	tag["排序"] = service.Order

	tags, _ := json.Marshal(tag)
	data["tags"] = string(tags)
	if service.Limit == "" {
		data["limit"] = "30"
	} else {
		data["limit"] = service.Limit
	}
	if service.Offset == "" {
		data["offset"] = "0"
	} else {
		data["offset"] = service.Offset
	}
	data["order"] = "true"
	code, reBody, _ := util.CreateRequest("POST", `https://interface.music.163.com/api/mv/all`, data, options)

	return code, reBody
}
