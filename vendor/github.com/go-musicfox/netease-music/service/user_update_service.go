package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type UserUpdateService struct {
	AvatarImgId string
	Birthday    string `json:"birthday" form:"birthday"`
	City        string `json:"city" form:"city"`
	Gender      string `json:"gender" form:"gender"`
	Nickname    string `json:"nickname" form:"nickname"`
	Province    string `json:"province" form:"province"`
	Signature   string `json:"signature" form:"signature"`
}

func (service *UserUpdateService) UserUpdate() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["avatarImgId"] = "0"
	data["birthday"] = service.Birthday
	data["city"] = service.City
	data["gender"] = service.Gender
	data["nickname"] = service.Nickname
	data["province"] = service.Province
	data["signature"] = service.Signature

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/user/profile/update`, data, options)

	return code, reBody
}
