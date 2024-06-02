package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type SongQualityLevel string

const (
	Standard SongQualityLevel = "standard"
	Higher   SongQualityLevel = "higher"
	Exhigh   SongQualityLevel = "exhigh"
	Lossless SongQualityLevel = "lossless"
	Hires    SongQualityLevel = "hires"
	JYEffect SongQualityLevel = "jyeffect"
	Sky      SongQualityLevel = "sky"
	JYMaster SongQualityLevel = "jymaster"
)

func (level SongQualityLevel) IsValid() bool {
	switch level {
	case Standard, Higher, Exhigh, Lossless, Hires, JYEffect, Sky, JYMaster:
		return true
	default:
		return false
	}
}

type SongUrlV1Service struct {
	ID         string           `json:"id" form:"id"`
	Level      SongQualityLevel `json:"level" form:"level"` // standard, exhigh, lossless, hires, jyeffect(高清环绕声), sky(沉浸环绕声), jymaster(超清母带) 进行音质判断
	EncodeType string           `json:"encodeType" form:"encodeType"`
	SkipUNM    bool
}

func (service *SongUrlV1Service) SongUrl() (float64, []byte) {
	options := &util.Options{
		Crypto:  "eapi",
		Url:     "/api/song/enhance/player/url/v1",
		SkipUNM: service.SkipUNM,
	}
	data := make(map[string]string)
	data["ids"] = "[" + service.ID + "]"
	if service.Level == "" {
		service.Level = Higher
	}
	data["level"] = string(service.Level)
	if service.Level == Sky {
		data["immerseType"] = "c51"
	}
	if service.EncodeType == "" {
		service.EncodeType = "flac"
	}
	data["encodeType"] = service.EncodeType

	code, reBody, _ := util.CreateRequest("POST", `https://interface.music.163.com/eapi/song/enhance/player/url/v1`, data, options)

	return code, reBody
}
