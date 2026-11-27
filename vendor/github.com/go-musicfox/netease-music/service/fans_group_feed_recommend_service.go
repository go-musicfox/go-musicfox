package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type FansGroupFeedRecommendService struct {
	ArtistSelf  string `json:"artistSelf" form:"artistSelf"`
	FansGroupId string `json:"fansGroupId" form:"fansGroupId"`
	GroupId     string `json:"groupId" form:"groupId"`
	Cursor      string `json:"cursor" form:"cursor"`
	Size        string `json:"size" form:"size"`
}

func (service *FansGroupFeedRecommendService) FansGroupFeedRecommend() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/fans/group/feed/recommend/get",
	}
	data := make(map[string]string)
	if service.ArtistSelf == "" {
		data["artistSelf"] = "0"
	} else {
		data["artistSelf"] = service.ArtistSelf
	}
	if service.FansGroupId == "" {
		data["fansGroupId"] = service.GroupId
	} else {
		data["fansGroupId"] = service.FansGroupId
	}
	if service.Cursor == "" {
		data["cursor"] = "0"
	} else {
		data["cursor"] = service.Cursor
	}
	if service.Size == "" {
		data["size"] = "10"
	} else {
		data["size"] = service.Size
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/fans/group/feed/recommend/get", data, options)

	return code, reBody
}
