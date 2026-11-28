package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type YunbeiExpenseService struct {
	Limit  string `json:"limit" form:"limit"`
	Offset string `json:"offset" form:"offset"`
}

func (service *YunbeiExpenseService) YunbeiExpense() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/point/expense",
	}
	data := make(map[string]string)
	if service.Limit == "" {
		data["limit"] = "10"
	} else {
		data["limit"] = service.Limit
	}
	if service.Offset == "" {
		data["offset"] = "0"
	} else {
		data["offset"] = service.Offset
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/point/expense", data, options)

	return code, reBody
}
