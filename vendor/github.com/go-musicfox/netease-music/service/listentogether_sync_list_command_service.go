package service

import (
	"encoding/json"
	"strings"

	"github.com/go-musicfox/netease-music/util"
)

type ListentogetherSyncListCommandService struct {
	RoomId      string `json:"roomId" form:"roomId"`
	CommandType string `json:"commandType" form:"commandType"`
	UserId      string `json:"userId" form:"userId"`
	Version     string `json:"version" form:"version"`
	RandomList  string `json:"randomList" form:"randomList"`
	DisplayList string `json:"displayList" form:"displayList"`
}

func (service *ListentogetherSyncListCommandService) ListentogetherSyncListCommand() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/listen/together/sync/list/command/report",
	}
	randomList := make([]string, 0)
	if service.RandomList != "" {
		randomList = strings.Split(service.RandomList, ",")
	}
	displayList := make([]string, 0)
	if service.DisplayList != "" {
		displayList = strings.Split(service.DisplayList, ",")
	}
	playlistParam, _ := json.Marshal(map[string]any{
		"commandType": service.CommandType,
		"version": []map[string]string{{
			"userId":  service.UserId,
			"version": service.Version,
		}},
		"anchorSongId":   "",
		"anchorPosition": -1,
		"randomList":     randomList,
		"displayList":    displayList,
	})
	data := make(map[string]string)
	data["roomId"] = service.RoomId
	data["playlistParam"] = string(playlistParam)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/listen/together/sync/list/command/report", data, options)

	return code, reBody
}
