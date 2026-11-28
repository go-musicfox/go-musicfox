package service

import (
	"encoding/json"
	"net/http"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/util"
)

type SongUrlV1302Service struct {
	Id    string `json:"id" form:"id"`
	Level string `json:"level" form:"level"`
}

func (service *SongUrlV1302Service) SongUrlV1302() (float64, []byte) {

	cookies := []*http.Cookie{{Name: "os", Value: "pc"}}
	if service.Level == "vivid" {
		cookies = []*http.Cookie{
			{Name: "os", Value: "android"},
			{Name: "appver", Value: "9.5.61"},
		}
	}
	data := map[string]string{
		"id":          service.Id,
		"immerseType": "c51",
		"level":       service.Level,
	}
	code, body, _ := util.CreateRequest(
		"POST",
		"https://music.163.com/eapi/song/enhance/download/url/v1",
		data,
		&util.Options{Crypto: "eapi", Url: "/api/song/enhance/download/url/v1", Cookies: cookies},
	)
	url := extractSongURL(body)
	if url != "" {
		result, _ := json.Marshal(map[string]any{"code": 302, "redirectUrl": url})
		return 302, result
	}

	fallback := map[string]string{
		"ids":        "[" + service.Id + "]",
		"level":      service.Level,
		"encodeType": "flac",
	}
	if service.Level == "sky" {
		fallback["immerseType"] = "c51"
	}
	if service.Level == "vivid" {
		fallback["encodeType"] = "mp3"
	}
	code, body, _ = util.CreateRequest(
		"POST",
		"https://music.163.com/eapi/song/enhance/player/url/v1",
		fallback,
		&util.Options{Crypto: "eapi", Url: "/api/song/enhance/player/url/v1", Cookies: cookies},
	)
	url = extractSongURL(body)
	if url == "" {
		return code, body
	}
	result, _ := json.Marshal(map[string]any{"code": 302, "redirectUrl": url})
	return 302, result
}

func extractSongURL(body []byte) string {
	if url, err := jsonparser.GetString(body, "data", "url"); err == nil && url != "" {
		return url
	}
	url, _ := jsonparser.GetString(body, "data", "[0]", "url")
	return url
}
