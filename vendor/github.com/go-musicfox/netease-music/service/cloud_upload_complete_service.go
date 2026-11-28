package service

import (
	"encoding/json"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/util"
)

type CloudUploadCompleteService struct {
	SongId     string `json:"songId" form:"songId"`
	ResourceId string `json:"resourceId" form:"resourceId"`
	Md5        string `json:"md5" form:"md5"`
	Filename   string `json:"filename" form:"filename"`
	Song       string `json:"song" form:"song"`
	Artist     string `json:"artist" form:"artist"`
	Album      string `json:"album" form:"album"`
	Bitrate    string `json:"bitrate" form:"bitrate"`
}

func (service *CloudUploadCompleteService) CloudUploadComplete() (float64, []byte) {

	if service.Bitrate == "" {
		service.Bitrate = "999000"
	}
	songName := service.Song
	if songName == "" {
		songName = service.Filename
	}
	artist := service.Artist
	if artist == "" {
		artist = "未知艺术家"
	}
	album := service.Album
	if album == "" {
		album = "未知专辑"
	}

	infoData := map[string]string{
		"md5":        service.Md5,
		"songid":     service.SongId,
		"filename":   service.Filename,
		"song":       songName,
		"album":      album,
		"artist":     artist,
		"bitrate":    service.Bitrate,
		"resourceId": service.ResourceId,
	}
	code, body, _ := util.CreateRequest(
		"POST",
		"https://music.163.com/eapi/upload/cloud/info/v2",
		infoData,
		&util.Options{Crypto: "eapi", Url: "/api/upload/cloud/info/v2"},
	)
	if code != 200 {
		return code, body
	}
	songID, _ := jsonparser.GetString(body, "songId")
	if songID == "" {
		songID = service.SongId
	}
	pubCode, pubBody, _ := util.CreateRequest(
		"POST",
		"https://music.163.com/eapi/cloud/pub/v2",
		map[string]string{"songid": songID},
		&util.Options{Crypto: "eapi", Url: "/api/cloud/pub/v2"},
	)
	result, _ := json.Marshal(map[string]any{
		"code": 200,
		"data": map[string]any{
			"songId": songID,
			"pub":    json.RawMessage(pubBody),
		},
	})
	return pubCode, result
}
