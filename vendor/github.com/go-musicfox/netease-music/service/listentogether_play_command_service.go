package service

import (
	"encoding/json"

	"github.com/go-musicfox/netease-music/util"
)

type ListentogetherPlayCommandService struct {
	RoomId       string `json:"roomId" form:"roomId"`
	CommandType  string `json:"commandType" form:"commandType"`
	Progress     string `json:"progress" form:"progress"`
	PlayStatus   string `json:"playStatus" form:"playStatus"`
	FormerSongId string `json:"formerSongId" form:"formerSongId"`
	TargetSongId string `json:"targetSongId" form:"targetSongId"`
	ClientSeq    string `json:"clientSeq" form:"clientSeq"`
}

func (service *ListentogetherPlayCommandService) ListentogetherPlayCommand() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/listen/together/play/command/report",
	}
	commandInfo, _ := json.Marshal(map[string]string{
		"commandType":  service.CommandType,
		"progress":     service.Progress,
		"playStatus":   service.PlayStatus,
		"formerSongId": service.FormerSongId,
		"targetSongId": service.TargetSongId,
		"clientSeq":    service.ClientSeq,
	})
	data := make(map[string]string)
	data["roomId"] = service.RoomId
	data["commandInfo"] = string(commandInfo)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/listen/together/play/command/report", data, options)

	return code, reBody
}
