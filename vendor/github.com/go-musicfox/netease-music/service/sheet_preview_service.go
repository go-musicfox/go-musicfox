package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SheetPreviewService struct {
	Id string `json:"id" form:"id"`
}

func (service *SheetPreviewService) SheetPreview() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/music/sheet/preview/info",
	}
	data := make(map[string]string)
	data["id"] = service.Id

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/music/sheet/preview/info", data, options)

	return code, reBody
}
