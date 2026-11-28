package service

import (
	"github.com/go-musicfox/netease-music/util"
)

var commentResourceThreadPrefixes = map[string]string{
	"0": "R_SO_4_",
	"1": "R_MV_5_",
	"2": "A_PL_0_",
	"3": "R_AL_3_",
	"4": "A_DJ_1_",
	"5": "R_VI_62_",
	"6": "A_EV_2_",
	"7": "A_DR_14_",
}

type CommentHugListService struct {
	Type     string `json:"type" form:"type"`
	Sid      string `json:"sid" form:"sid"`
	Uid      string `json:"uid" form:"uid"`
	Cid      string `json:"cid" form:"cid"`
	Cursor   string `json:"cursor" form:"cursor"`
	Page     string `json:"page" form:"page"`
	IdCursor string `json:"idCursor" form:"idCursor"`
	PageSize string `json:"pageSize" form:"pageSize"`
}

func (service *CommentHugListService) CommentHugList() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/v2/resource/comments/hug/list",
	}
	prefix, ok := commentResourceThreadPrefixes[service.Type]
	if !ok {
		prefix = commentResourceThreadPrefixes["0"]
	}

	data := make(map[string]string)
	data["targetUserId"] = service.Uid
	data["commentId"] = service.Cid
	if service.Cursor == "" {
		data["cursor"] = "-1"
	} else {
		data["cursor"] = service.Cursor
	}
	data["threadId"] = prefix + service.Sid
	if service.Page == "" {
		data["pageNo"] = "1"
	} else {
		data["pageNo"] = service.Page
	}
	if service.IdCursor == "" {
		data["idCursor"] = "-1"
	} else {
		data["idCursor"] = service.IdCursor
	}
	if service.PageSize == "" {
		data["pageSize"] = "100"
	} else {
		data["pageSize"] = service.PageSize
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/v2/resource/comments/hug/list", data, options)

	return code, reBody
}
