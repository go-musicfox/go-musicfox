package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type VideoGroupService struct {
	GroupID string `json:"id" form:"id"`
	Offset  string `json:"offset" form:"offset"`
}

func (service *VideoGroupService) VideoGroup() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["groupId"] = service.GroupID
	if service.Offset == "" {
		data["offset"] = "0"
	} else {
		data["offset"] = service.Offset
	}
	data["order"] = "true"
	data["need_preview_url"] = "true"
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/videotimeline/videogroup/otherclient/get`, data, options)

	return code, reBody
}
