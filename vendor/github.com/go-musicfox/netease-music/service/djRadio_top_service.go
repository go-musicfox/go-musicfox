package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type DjRadioTopService struct {
	DataGapDays string `json:"dataGapDays" form:"dataGapDays"`
	DataType    string `json:"dataType" form:"dataType"`
	DjRadioId   string `json:"djRadioId" form:"djRadioId"`
	SortIndex   string `json:"sortIndex" form:"sortIndex"`
}

func (service *DjRadioTopService) DjRadioTop() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/expert/worksdata/works/top/get",
	}
	data := make(map[string]string)
	if service.DjRadioId == "" {
		data["djRadioId"] = ""
	} else {
		data["djRadioId"] = service.DjRadioId
	}
	if service.SortIndex == "" {
		data["sortIndex"] = "1"
	} else {
		data["sortIndex"] = service.SortIndex
	}
	if service.DataGapDays == "" {
		data["dataGapDays"] = "7"
	} else {
		data["dataGapDays"] = service.DataGapDays
	}
	if service.DataType == "" {
		data["dataType"] = "3"
	} else {
		data["dataType"] = service.DataType
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/expert/worksdata/works/top/get", data, options)

	return code, reBody
}
