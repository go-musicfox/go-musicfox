package service

import (
	"encoding/json"
	"regexp"

	"github.com/go-musicfox/netease-music/util"
)

type RelatedPlaylistService struct {
	ID string `json:"id" form:"id"`
}

func (service *RelatedPlaylistService) RelatedPlaylist() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
		Ua:     "pc",
	}
	data := make(map[string]string)

	code, reBody, _ := util.CreateRequest("GET", `https://music.163.com/playlist?id=`+service.ID, data, options)

	reg := regexp.MustCompile("<div class=\"cver u-cover u-cover-3\">[\\s\\S]*?<img src=\"([^\"]+)\">[\\s\\S]*?<a class=\"sname f-fs1 s-fc0\" href=\"([^\"]+)\"[^>]*>([^<]+?)<\\/a>[\\s\\S]*?<a class=\"nm nm f-thide s-fc3\" href=\"([^\"]+)\"[^>]*>([^<]+?)<\\/a>")
	results := reg.FindAllSubmatch(reBody, -1)

	type Creator struct {
		UserId   string `json:"userId"`
		Nickname string `json:"nickname"`
	}

	type Result struct {
		Id          string  `json:"id"`
		Name        string  `json:"name"`
		CoverImgUrl string  `json:"coverImgUrl"`
		Creator     Creator `json:"creator"`
	}
	var Results []Result
	for _, result := range results {
		var item Result
		item.Id = string(result[2][len("/playlist?id="):])
		item.Name = string(result[3])
		item.CoverImgUrl = string(result[1][0 : len(result[1])-len("?param=50y50")])
		item.Creator.UserId = string(result[4][len("/user/home?id="):])
		item.Creator.Nickname = string(result[5])
		Results = append(Results, item)
	}

	res, _ := json.Marshal(Results)
	return code, res
}
