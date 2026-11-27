package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type CommentReportService struct {
	Cid    string `json:"cid" form:"cid"`
	Id     string `json:"id" form:"id"`
	Reason string `json:"reason" form:"reason"`
}

func (service *CommentReportService) CommentReport() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/report/reportcomment",
	}
	data := make(map[string]string)
	data["threadId"] = "R_SO_4_" + service.Id
	data["commentId"] = service.Cid
	data["reason"] = service.Reason

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/report/reportcomment", data, options)

	return code, reBody
}
