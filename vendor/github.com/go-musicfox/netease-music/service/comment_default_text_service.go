package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type CommentDefaultTextService struct {
	ThreadId string `json:"threadId" form:"threadId"`
}

func (service *CommentDefaultTextService) CommentDefaultText() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["threadId"] = service.ThreadId

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/comment/default/text", data, options)

	return code, reBody
}
