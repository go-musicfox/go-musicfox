package service

import (
	"encoding/json"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/util"
)

type CloudImportService struct {
	Md5      string `json:"md5" form:"md5"`
	Id       string `json:"id" form:"id"`
	Bitrate  string `json:"bitrate" form:"bitrate"`
	FileSize string `json:"fileSize" form:"fileSize"`
	Song     string `json:"song" form:"song"`
	Artist   string `json:"artist" form:"artist"`
	Album    string `json:"album" form:"album"`
	FileType string `json:"fileType" form:"fileType"`
}

func (service *CloudImportService) CloudImport() (float64, []byte) {

	if service.Id == "" {
		service.Id = "-2"
	}
	if service.Artist == "" {
		service.Artist = "未知"
	}
	if service.Album == "" {
		service.Album = "未知"
	}
	checkSongs, _ := json.Marshal([]map[string]string{{
		"md5":      service.Md5,
		"songId":   service.Id,
		"bitrate":  service.Bitrate,
		"fileSize": service.FileSize,
	}})
	checkData := map[string]string{
		"uploadType": "0",
		"songs":      string(checkSongs),
	}
	code, body, _ := util.CreateRequest(
		"POST",
		"https://music.163.com/eapi/cloud/upload/check/v2",
		checkData,
		&util.Options{Crypto: "eapi", Url: "/api/cloud/upload/check/v2"},
	)
	if code != 200 {
		return code, body
	}
	songID, _ := jsonparser.GetString(body, "data", "[0]", "songId")
	if songID == "" {
		songID, _ = jsonparser.GetString(body, "data", "[0]", "songid")
	}

	importSongs, _ := json.Marshal([]map[string]string{{
		"songId":   songID,
		"bitrate":  service.Bitrate,
		"song":     service.Song,
		"artist":   service.Artist,
		"album":    service.Album,
		"fileName": service.Song + "." + service.FileType,
	}})
	importData := map[string]string{
		"uploadType": "0",
		"songs":      string(importSongs),
	}
	code, body, _ = util.CreateRequest(
		"POST",
		"https://music.163.com/eapi/cloud/user/song/import",
		importData,
		&util.Options{Crypto: "eapi", Url: "/api/cloud/user/song/import"},
	)
	return code, body
}
