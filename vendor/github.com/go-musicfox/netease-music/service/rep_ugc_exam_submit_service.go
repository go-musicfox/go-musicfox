package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type RepUgcExamSubmitService struct {
	Answer     string `json:"answer" form:"answer"`
	ExamType   string `json:"examType" form:"examType"`
	QuestionId string `json:"questionId" form:"questionId"`
	TaskId     string `json:"taskId" form:"taskId"`
}

func (service *RepUgcExamSubmitService) RepUgcExamSubmit() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/rep/ugc/exam/submit",
	}
	data := make(map[string]string)
	data["examType"] = service.ExamType
	data["taskId"] = service.TaskId
	data["questionId"] = service.QuestionId
	data["answer"] = service.Answer

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/rep/ugc/exam/submit", data, options)

	return code, reBody
}
