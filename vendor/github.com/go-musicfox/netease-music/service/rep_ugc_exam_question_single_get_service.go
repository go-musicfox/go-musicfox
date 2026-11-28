package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type RepUgcExamQuestionSingleGetService struct {
	ExamType string `json:"examType" form:"examType"`
	TaskId   string `json:"taskId" form:"taskId"`
}

func (service *RepUgcExamQuestionSingleGetService) RepUgcExamQuestionSingleGet() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/rep/ugc/exam/question/single/get",
	}
	data := make(map[string]string)
	data["examType"] = service.ExamType
	data["taskId"] = service.TaskId

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/rep/ugc/exam/question/single/get", data, options)

	return code, reBody
}
