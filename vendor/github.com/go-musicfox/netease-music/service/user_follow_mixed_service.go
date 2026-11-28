package service

import (
	"encoding/json"
	"strconv"

	"github.com/go-musicfox/netease-music/util"
)

type UserFollowMixedService struct {
	Size   string `json:"size" form:"size"`
	Cursor string `json:"cursor" form:"cursor"`
	Scene  string `json:"scene" form:"scene"`
}

func (service *UserFollowMixedService) UserFollowMixed() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/user/follow/users/mixed/get/v2",
	}
	size, _ := strconv.Atoi(service.Size)
	if size == 0 {
		size = 30
	}
	cursor, _ := strconv.Atoi(service.Cursor)
	page, _ := json.Marshal(map[string]int{
		"size":   size,
		"cursor": cursor,
	})

	data := make(map[string]string)
	data["authority"] = "false"
	data["page"] = string(page)
	if service.Scene == "" {
		data["scene"] = "0"
	} else {
		data["scene"] = service.Scene
	}
	data["size"] = strconv.Itoa(size)
	data["sortType"] = "0"

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/user/follow/users/mixed/get/v2", data, options)

	return code, reBody
}
