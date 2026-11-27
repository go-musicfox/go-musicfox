package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type GetUseridsService struct {
	Nicknames string `json:"nicknames" form:"nicknames"`
}

func (service *GetUseridsService) GetUserids() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["nicknames"] = service.Nicknames

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/user/getUserIds", data, options)

	return code, reBody
}
