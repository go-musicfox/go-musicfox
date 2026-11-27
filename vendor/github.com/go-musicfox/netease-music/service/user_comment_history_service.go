package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type UserCommentHistoryService struct {
	Limit string `json:"limit" form:"limit"`
	Time  string `json:"time" form:"time"`
	Uid   string `json:"uid" form:"uid"`
}

func (service *UserCommentHistoryService) UserCommentHistory() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["compose_reminder"] = "true"
	data["compose_hot_comment"] = "true"
	if service.Limit == "" {
		data["limit"] = "10"
	} else {
		data["limit"] = service.Limit
	}
	data["user_id"] = service.Uid
	if service.Time == "" {
		data["time"] = "0"
	} else {
		data["time"] = service.Time
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/comment/user/comment/history", data, options)

	return code, reBody
}
