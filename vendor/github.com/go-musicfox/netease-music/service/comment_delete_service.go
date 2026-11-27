package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type CommentDeleteService struct {
	Type string `json:"type" form:"type"`
	Id   string `json:"id" form:"id"`
	Cid  string `json:"cid" form:"cid"`
}

func (service *CommentDeleteService) CommentDelete() (float64, []byte) {

	options := &util.Options{
		Crypto: "xeapi",
		Url:    "/api/resource/comments/delete",
	}
	prefix, ok := commentResourceThreadPrefixes[service.Type]
	if !ok {
		prefix = commentResourceThreadPrefixes["0"]
	}
	data := make(map[string]string)
	data["commentId"] = service.Cid
	data["threadId"] = prefix + service.Id

	code, reBody, _ := util.CreateRequest("POST", "https://interface3.music.163.com/xeapi/resource/comments/delete", data, options)

	return code, reBody
}
