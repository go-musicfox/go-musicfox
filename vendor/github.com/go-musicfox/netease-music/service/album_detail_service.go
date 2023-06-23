package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type AlbumDetailService struct {
	ID string `json:"id" form:"id"`
}

func (service *AlbumDetailService) AlbumDetail() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["id"] = service.ID
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/vipmall/albumproduct/detail`, data, options)

	return code, reBody
}
