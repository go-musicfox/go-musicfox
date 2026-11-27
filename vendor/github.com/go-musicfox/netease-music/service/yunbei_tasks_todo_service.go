package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type YunbeiTasksTodoService struct {
}

func (service *YunbeiTasksTodoService) YunbeiTasksTodo() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/usertool/task/todo/query", data, options)

	return code, reBody
}
