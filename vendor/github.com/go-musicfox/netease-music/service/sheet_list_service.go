package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SheetListService struct {
	Ab string `json:"ab" form:"ab"`
	Id string `json:"id" form:"id"`
}

func (service *SheetListService) SheetList() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/music/sheet/list/v1",
	}
	data := make(map[string]string)
	data["id"] = service.Id
	if service.Ab == "" {
		data["abTest"] = "b"
	} else {
		data["abTest"] = service.Ab
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/music/sheet/list/v1", data, options)

	return code, reBody
}
