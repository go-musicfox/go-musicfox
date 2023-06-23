package service

import (
	"net/http"

	"github.com/go-musicfox/netease-music/util"
)

type CommentService struct {
	ID        string `json:"id" form:"id"`
	ThreadId  string `json:"threadId" form:"threadId"`
	Content   string `json:"content" form:"content"`
	T         string `json:"t" form:"t"`
	Type      string `json:"type" form:"type"`
	CommentId string `json:"commentId" form:"commentId"`
}

func (service *CommentService) Comment() (float64, []byte) {

	cookiesOS := &http.Cookie{Name: "os", Value: "pc"}

	options := &util.Options{
		Crypto:  "weapi",
		Cookies: []*http.Cookie{cookiesOS},
	}
	TYPE := make(map[string]string, 6)
	TYPE["0"] = "R_SO_4_"
	TYPE["1"] = "R_MV_5_"
	TYPE["2"] = "A_PL_0_"
	TYPE["3"] = "R_AL_3_"
	TYPE["4"] = "A_DJ_1_"
	TYPE["5"] = "R_VI_62_"
	TYPE["6"] = "A_EV_2_"

	T := make(map[string]string, 3)
	T["0"] = "delete"
	T["1"] = "add"
	T["2"] = "reply"

	if _, ok := TYPE[service.Type]; ok {
		service.Type = TYPE[service.Type]
	} else {
		service.Type = TYPE["0"]
	}

	if _, ok := T[service.T]; ok {
		service.T = T[service.T]
	} else {
		service.T = T["1"]
	}

	data := make(map[string]string)
	data["threadId"] = service.Type + service.ID

	if service.Type == "A_EV_2_" {
		data["threadId"] = service.ThreadId
	}

	if service.T == "add" {
		data["content"] = service.Content
	} else if service.T == "delete" {
		data["commentId"] = service.CommentId
	} else if service.T == "reply" {
		data["commentId"] = service.CommentId
		data["content"] = service.Content
	}
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/resource/comments/`+service.T, data, options)

	return code, reBody
}
