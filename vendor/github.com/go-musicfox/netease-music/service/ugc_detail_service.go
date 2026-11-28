package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type UgcDetailService struct {
	AuditStatus string `json:"auditStatus" form:"auditStatus"`
	Limit       string `json:"limit" form:"limit"`
	Offset      string `json:"offset" form:"offset"`
	Order       string `json:"order" form:"order"`
	SortBy      string `json:"sortBy" form:"sortBy"`
	Type        string `json:"type" form:"type"`
}

func (service *UgcDetailService) UgcDetail() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.AuditStatus == "" {
		data["auditStatus"] = ""
	} else {
		data["auditStatus"] = service.AuditStatus
	}
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
	if service.Order == "" {
		data["order"] = "desc"
	} else {
		data["order"] = service.Order
	}
	if service.SortBy == "" {
		data["sortBy"] = "createTime"
	} else {
		data["sortBy"] = service.SortBy
	}
	if service.Type == "" {
		data["type"] = "1"
	} else {
		data["type"] = service.Type
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/rep/ugc/detail", data, options)

	return code, reBody
}
