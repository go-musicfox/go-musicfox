package service

import (
	"fmt"
	"strings"

	"github.com/go-musicfox/netease-music/util"
)

/*
   type 取值
   1:男歌手
   2:女歌手
   3:乐队

   area 取值
   -1:全部
   7华语
   96欧美
   8:日本
   16韩国
   0:其他

   initial 取值 a-z/A-Z
*/

type ArtistListService struct {
	Type    string `json:"type" form:"type"`
	Limit   string `json:"limit" form:"limit"`
	Offset  string `json:"offset" form:"offset"`
	Area    string `json:"area" form:"area"`
	Initial string `json:"initial" form:"initial"`
}

func (service *ArtistListService) ArtistList() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	if service.Limit == "" {
		data["limit"] = "30"
	} else {
		data["limit"] = service.Limit
	}
	if service.Offset == "" {
		data["offset"] = "0"
	} else {
		data["offset"] = service.Offset
	}
	if service.Type == "" {
		data["type"] = "1"
	} else {
		data["type"] = service.Type
	}
	data["total"] = "true"
	data["area"] = service.Area
	if service.Initial == "" {
		data["initial"] = ""
	} else {
		data["initial"] = fmt.Sprintf("%v", strings.ToUpper(service.Initial)[0])
	}

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/v1/artist/list`, data, options)

	return code, reBody
}
