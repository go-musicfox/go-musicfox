package service

import (
	"net/http"

	"github.com/go-musicfox/netease-music/util"
)

type CommentVideoService struct {
	ID     string `json:"id" form:"id"`
	Limit  string `json:"limit" form:"limit"`
	Offset string `json:"offset" form:"offset"`
	Before string `json:"before" form:"before"`
}

func (service *CommentVideoService) CommentVideo() (float64, []byte) {

	cookiesOS := &http.Cookie{Name: "os", Value: "pc"}

	options := &util.Options{
		Crypto:  "weapi",
		Cookies: []*http.Cookie{cookiesOS},
	}
	data := make(map[string]string)
	data["rid"] = service.ID
	if service.Limit == "" {
		data["limit"] = "20"
	} else {
		data["limit"] = service.Limit
	}
	if service.Offset == "" {
		data["offset"] = "0"
	} else {
		data["offset"] = service.Offset
	}
	if service.Before == "" {
		data["beforeTime"] = "0"
	} else {
		data["beforeTime"] = service.Before
	}
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/v1/resource/comments/R_VI_62_`+service.ID, data, options)

	return code, reBody
}
