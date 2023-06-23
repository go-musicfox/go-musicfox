package service

import (
	"github.com/go-musicfox/netease-music/util"
)

/*
   有声书 10001
   知识技能 453050
   商业财经 453051
   人文历史 11
   外语世界 13
   亲子宝贝 14
   创作|翻唱 2001
   音乐故事 2
   3D|电子 10002
   相声曲艺 8
   情感调频 3
   美文读物 6
   脱口秀 5
   广播剧 7
   二次元 3001
   明星做主播 1
   娱乐|影视 4
   科技科学 453052
   校园|教育 4001
   旅途|城市 12
*/

type DjRecommendTypeService struct {
	CateId string `json:"type" form:"type"`
}

func (service *DjRecommendTypeService) DjRecommendType() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["cateId"] = service.CateId
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/djradio/recommend`, data, options)

	return code, reBody
}
