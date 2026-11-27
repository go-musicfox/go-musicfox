package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type AlbumPrivilegeService struct {
	Id string `json:"id" form:"id"`
}

func (service *AlbumPrivilegeService) AlbumPrivilege() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/album/privilege",
	}
	data := make(map[string]string)
	data["id"] = service.Id

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/album/privilege", data, options)

	return code, reBody
}
