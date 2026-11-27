package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type HugCommentService struct {
	Type string `json:"type" form:"type"`
	Sid  string `json:"sid" form:"sid"`
	Uid  string `json:"uid" form:"uid"`
	Cid  string `json:"cid" form:"cid"`
}

func (service *HugCommentService) HugComment() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/v2/resource/comments/hug/listener",
	}
	prefix, ok := commentResourceThreadPrefixes[service.Type]
	if !ok {
		prefix = commentResourceThreadPrefixes["0"]
	}
	data := make(map[string]string)
	data["targetUserId"] = service.Uid
	data["commentId"] = service.Cid
	data["threadId"] = prefix + service.Sid

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/v2/resource/comments/hug/listener", data, options)

	return code, reBody
}
