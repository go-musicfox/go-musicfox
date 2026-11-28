package service

import (
	"strconv"

	"github.com/go-musicfox/netease-music/util"
)

type CommentNewService struct {
	Type      string `json:"type" form:"type"`
	Id        string `json:"id" form:"id"`
	PageSize  string `json:"pageSize" form:"pageSize"`
	PageNo    string `json:"pageNo" form:"pageNo"`
	SortType  string `json:"sortType" form:"sortType"`
	Cursor    string `json:"cursor" form:"cursor"`
	ShowInner string `json:"showInner" form:"showInner"`
}

func (service *CommentNewService) CommentNew() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/v2/resource/comments",
	}
	prefix, ok := commentResourceThreadPrefixes[service.Type]
	if !ok {
		prefix = commentResourceThreadPrefixes["0"]
	}
	pageSize, _ := strconv.Atoi(service.PageSize)
	if pageSize == 0 {
		pageSize = 20
	}
	pageNo, _ := strconv.Atoi(service.PageNo)
	if pageNo == 0 {
		pageNo = 1
	}
	sortType := service.SortType
	if sortType == "" || sortType == "1" {
		sortType = "99"
	}
	cursor := ""
	switch sortType {
	case "99":
		cursor = strconv.Itoa((pageNo - 1) * pageSize)
	case "2":
		cursor = "normalHot#" + strconv.Itoa((pageNo-1)*pageSize)
	case "3":
		cursor = service.Cursor
		if cursor == "" {
			cursor = "0"
		}
	}
	showInner := service.ShowInner
	if showInner == "" {
		showInner = "true"
	}

	data := make(map[string]string)
	data["threadId"] = prefix + service.Id
	data["pageNo"] = strconv.Itoa(pageNo)
	data["showInner"] = showInner
	data["pageSize"] = strconv.Itoa(pageSize)
	data["cursor"] = cursor
	data["sortType"] = sortType

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/v2/resource/comments", data, options)

	return code, reBody
}
