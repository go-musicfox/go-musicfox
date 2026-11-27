package service

import (
	"encoding/json"
	"strconv"

	"github.com/go-musicfox/netease-music/util"
)

type YunbeiRcmdSongHistoryService struct {
	Size   string `json:"size" form:"size"`
	Cursor string `json:"cursor" form:"cursor"`
}

func (service *YunbeiRcmdSongHistoryService) YunbeiRcmdSongHistory() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	size, _ := strconv.Atoi(service.Size)
	if size == 0 {
		size = 20
	}
	page, _ := json.Marshal(map[string]any{
		"size":   size,
		"cursor": service.Cursor,
	})

	data := make(map[string]string)
	data["page"] = string(page)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/yunbei/rcmd/song/history/list", data, options)

	return code, reBody
}
