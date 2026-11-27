package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type YunbeiTaskListV1Service struct {
}

func (service *YunbeiTaskListV1Service) YunbeiTaskListV1() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/ad/power/yunbei/distribution/list", data, options)

	return code, reBody
}
