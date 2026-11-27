package service

import (
	"crypto/rand"
	"encoding/json"
	"strconv"

	"github.com/go-musicfox/netease-music/util"
)

const relaySessionChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type RelayPlayStateSubmitService struct {
	Id        string `json:"id" form:"id"`
	SessionId string `json:"sessionId" form:"sessionId"`
	Progress  string `json:"progress" form:"progress"`
	PlayMode  string `json:"playMode" form:"playMode"`
	Type      string `json:"type" form:"type"`
}

func (service *RelayPlayStateSubmitService) RelayPlayStateSubmit() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	sessionID := service.SessionId
	if sessionID == "" {
		buf := make([]byte, 12)
		_, _ = rand.Read(buf)
		for i := range buf {
			buf[i] = relaySessionChars[int(buf[i])%len(relaySessionChars)]
		}
		sessionID = string(buf)
	}
	progress, _ := strconv.Atoi(service.Progress)
	resourceType := service.Type
	if resourceType == "" {
		resourceType = "song"
	}
	playMode := service.PlayMode
	if playMode == "" {
		playMode = "list_loop"
	}
	state, _ := json.Marshal(map[string]any{
		"resource": map[string]string{
			"id":   service.Id,
			"type": resourceType,
		},
		"progress":  progress,
		"sessionId": sessionID,
		"playMode":  playMode,
	})

	data := make(map[string]string)
	data["playStateSubmitReq"] = string(state)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/relay/play/state/submit", data, options)

	return code, reBody
}
