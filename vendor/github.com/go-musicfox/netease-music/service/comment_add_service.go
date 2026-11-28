package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type CommentAddService struct {
	Type    string `json:"type" form:"type"`
	Id      string `json:"id" form:"id"`
	Content string `json:"content" form:"content"`
}

func (service *CommentAddService) CommentAdd() (float64, []byte) {

	options := &util.Options{
		Crypto:     "xeapi",
		Url:        "/api/resource/comments/add",
		CheckToken: "v3",
	}
	prefix, ok := commentResourceThreadPrefixes[service.Type]
	if !ok {
		prefix = commentResourceThreadPrefixes["0"]
	}
	data := make(map[string]string)
	data["threadId"] = prefix + service.Id
	data["content"] = service.Content
	data["resourceType"] = "0"
	data["expressionPicId"] = "-1"
	data["bubbleId"] = "-1"

	code, reBody, _ := util.CreateRequest("POST", "https://interface3.music.163.com/xeapi/resource/comments/add", data, options)

	return code, reBody
}
