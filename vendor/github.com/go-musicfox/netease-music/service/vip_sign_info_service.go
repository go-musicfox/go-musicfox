package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type VipSignInfoService struct {
}

func (service *VipSignInfoService) VipSignInfo() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/vipnewcenter/app/user/sign/info", data, options)

	return code, reBody
}
