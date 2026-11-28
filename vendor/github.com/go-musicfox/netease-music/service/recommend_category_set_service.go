package service

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/go-musicfox/netease-music/util"
)

type RecommendCategorySetService struct {
	Tags     string `json:"tags" form:"tags"`
	Category string `json:"category" form:"category"`
}

func (service *RecommendCategorySetService) RecommendCategorySet() (float64, []byte) {

	options := &util.Options{
		Crypto: "xeapi",
		Url:    "/api/homepage/daily/song/tag/save",
	}
	tags := make([]int, 0)
	for _, tag := range strings.Split(service.Tags, ",") {
		if tag = strings.TrimSpace(tag); tag == "" {
			continue
		}
		if v, err := strconv.Atoi(tag); err == nil {
			tags = append(tags, v)
		}
	}
	category, _ := strconv.Atoi(service.Category)
	payload, _ := json.Marshal(map[string]any{
		"tagIds":     tags,
		"categoryId": category,
	})
	data := map[string]string{"tags": string(payload)}

	code, reBody, _ := util.CreateRequest("POST", "https://interface3.music.163.com/xeapi/homepage/daily/song/tag/save", data, options)

	return code, reBody
}
