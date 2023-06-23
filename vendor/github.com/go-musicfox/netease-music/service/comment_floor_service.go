package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type CommentFloorService struct {
	ParentCommentId string `json:"parentCommentId" form:"parentCommentId"`
	Limit           string `json:"limit" form:"limit"`
	Type            string `json:"type" form:"type"`
	Id              string `json:"id" form:"id"`
	Time            string `json:"time" form:"time"`
}

func (service *CommentFloorService) CommentFloor() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	Type := map[string]string{
		"0": "R_SO_4_",  //  歌曲
		"1": "R_MV_5_",  //  MV
		"2": "A_PL_0_",  //  歌单
		"3": "R_AL_3_",  //  专辑
		"4": "A_DJ_1_",  //  电台,
		"5": "R_VI_62_", //  视频
	}
	if service.Limit == "" {
		data["limit"] = "20"
	} else {
		data["limit"] = service.Limit
	}
	if service.Time == "" {
		data["time"] = "0"
	} else {
		data["time"] = service.Time
	}
	data["parentCommentId"] = service.ParentCommentId
	data["threadId"] = Type[service.Type] + service.Id
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/resource/comment/floor/get`, data, options)

	return code, reBody
}
