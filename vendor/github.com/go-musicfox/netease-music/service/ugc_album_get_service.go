package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type UgcAlbumGetService struct {
	Id string `json:"id" form:"id"`
}

func (service *UgcAlbumGetService) UgcAlbumGet() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/rep/ugc/album/get",
	}
	data := make(map[string]string)
	data["albumId"] = service.Id

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/rep/ugc/album/get", data, options)

	return code, reBody
}
