package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type CommentHotwallSquareInfoService struct {
}

func (service *CommentHotwallSquareInfoService) CommentHotwallSquareInfo() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/comment/hotwall/today/squareinfo/get", data, options)

	return code, reBody
}
