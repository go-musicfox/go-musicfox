package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type FmTrashService struct {
	SongID string `json:"id" form:"id"`
}

func (service *FmTrashService) FmTrash() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["songId"] = service.SongID

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/radio/trash/add?alg=RT&songId=`+service.SongID+`&time=25`, data, options)

	return code, reBody
}
