package service

import (
	"encoding/json"

	"github.com/go-musicfox/netease-music/util"
)

type UserSocialStatusEditService struct {
	Type      string `json:"type" form:"type"`
	IconUrl   string `json:"iconUrl" form:"iconUrl"`
	Content   string `json:"content" form:"content"`
	ActionUrl string `json:"actionUrl" form:"actionUrl"`
}

func (service *UserSocialStatusEditService) UserSocialStatusEdit() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/social/user/status/edit",
	}
	data := make(map[string]string)
	content, _ := json.Marshal(map[string]string{
		"type":      service.Type,
		"iconUrl":   service.IconUrl,
		"content":   service.Content,
		"actionUrl": service.ActionUrl,
	})
	data["content"] = string(content)

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/social/user/status/edit", data, options)

	return code, reBody
}
