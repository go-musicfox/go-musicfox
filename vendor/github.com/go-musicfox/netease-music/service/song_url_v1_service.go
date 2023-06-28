package service

import (
	"net/http"

	"github.com/go-musicfox/netease-music/util"
)

type SongQualityLevel string

const (
	Standard SongQualityLevel = "standard"
	Higher   SongQualityLevel = "higher"
	Exhigh   SongQualityLevel = "exhigh"
	Lossless SongQualityLevel = "lossless"
	Hires    SongQualityLevel = "hires"
)

func (level SongQualityLevel) IsValid() bool {
	switch level {
	case Standard, Higher, Exhigh, Lossless, Hires:
		return true
	default:
		return false
	}
}

type SongUrlV1Service struct {
	ID         string           `json:"id" form:"id"`
	Level      SongQualityLevel `json:"level" form:"level"` // standard,higher,exhigh,lossless,hires
	EncodeType string           `json:"encodeType" form:"encodeType"`
	SkipUNM    bool
}

func (service *SongUrlV1Service) SongUrl() (float64, []byte) {

	cookiesOS := &http.Cookie{Name: "os", Value: "pc"}

	options := &util.Options{
		Crypto:  "eapi",
		Cookies: []*http.Cookie{cookiesOS},
		Url:     "/api/song/enhance/player/url/v1",
		SkipUNM: service.SkipUNM,
	}
	data := make(map[string]string)
	data["ids"] = "[" + service.ID + "]"
	if service.Level == "" {
		service.Level = Higher
	}
	data["level"] = string(service.Level)
	data["encodeType"] = service.EncodeType

	code, reBody, _ := util.CreateRequest("POST", `https://interface.music.163.com/eapi/song/enhance/player/url/v1`, data, options)

	return code, reBody
}
