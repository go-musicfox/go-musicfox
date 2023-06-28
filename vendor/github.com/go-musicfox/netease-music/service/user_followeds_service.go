package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type UserFollowedsService struct {
	Uid   string `json:"uid" form:"uid"`
	Limit string `json:"limit" form:"limit"`
	Time  string `json:"lasttime " form:"lasttime "`
}

func (service *UserFollowedsService) UserFolloweds() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/user/getfolloweds",
	}
	data := make(map[string]string)
	data["userId"] = service.Uid
	if service.Limit == "" {
		data["limit"] = "30"
	} else {
		data["limit"] = service.Limit
	}
	if service.Time == "" {
		data["time"] = "-1"
	} else {
		data["time"] = service.Time
	}
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/eapi/user/getfolloweds/`+service.Uid, data, options)

	return code, reBody
}
