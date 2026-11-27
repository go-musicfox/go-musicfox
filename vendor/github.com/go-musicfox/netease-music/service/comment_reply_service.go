package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type CommentReplyService struct {
	Type    string `json:"type" form:"type"`
	Id      string `json:"id" form:"id"`
	Cid     string `json:"cid" form:"cid"`
	Content string `json:"content" form:"content"`
}

func (service *CommentReplyService) CommentReply() (float64, []byte) {

	options := &util.Options{
		Crypto:     "xeapi",
		Url:        "/api/v1/resource/comments/reply",
		CheckToken: "v3",
	}
	prefix, ok := commentResourceThreadPrefixes[service.Type]
	if !ok {
		prefix = commentResourceThreadPrefixes["0"]
	}
	data := make(map[string]string)
	data["threadId"] = prefix + service.Id
	data["commentId"] = service.Cid
	data["content"] = service.Content
	data["resourceType"] = "0"

	code, reBody, _ := util.CreateRequest("POST", "https://interface3.music.163.com/xeapi/v1/resource/comments/reply", data, options)

	return code, reBody
}
