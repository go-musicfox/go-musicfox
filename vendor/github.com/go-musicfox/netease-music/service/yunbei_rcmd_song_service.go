package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type YunbeiRcmdSongService struct {
	Id        string `json:"id" form:"id"`
	Reason    string `json:"reason" form:"reason"`
	YunbeiNum string `json:"yunbeiNum" form:"yunbeiNum"`
}

func (service *YunbeiRcmdSongService) YunbeiRcmdSong() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["songId"] = service.Id
	if service.Reason == "" {
		data["reason"] = "好歌献给你"
	} else {
		data["reason"] = service.Reason
	}
	data["scene"] = ""
	data["fromUserId"] = "-1"
	if service.YunbeiNum == "" {
		data["yunbeiNum"] = "10"
	} else {
		data["yunbeiNum"] = service.YunbeiNum
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/yunbei/rcmd/song/submit", data, options)

	return code, reBody
}
