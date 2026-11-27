package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type RepUgcExamStartService struct {
	ExamType string `json:"examType" form:"examType"`
}

func (service *RepUgcExamStartService) RepUgcExamStart() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/rep/ugc/exam/start",
	}
	data := make(map[string]string)
	data["examType"] = service.ExamType

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/rep/ugc/exam/start", data, options)

	return code, reBody
}
