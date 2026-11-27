package service

import (
	"encoding/json"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/util"
)

type CloudService struct {
	FileName string `json:"fileName" form:"fileName"`
	MimeType string `json:"mimeType" form:"mimeType"`
	Song     string `json:"song" form:"song"`
	Album    string `json:"album" form:"album"`
	Artist   string `json:"artist" form:"artist"`
	Bitrate  string `json:"bitrate" form:"bitrate"`
	Md5      string `json:"md5" form:"md5"`
	Size     string `json:"size" form:"size"`
	Data     []byte `json:"-" form:"-"`
}

func (service *CloudService) Cloud() (float64, []byte) {

	file := &util.NosUploadFile{Name: service.FileName, MimeType: service.MimeType, Data: service.Data, MD5: service.Md5}
	file.Prepare()
	if service.Size == "" {
		service.Size = strconv.FormatInt(file.Size, 10)
	}
	if service.Bitrate == "" {
		service.Bitrate = "999000"
	}
	songName := service.Song
	if songName == "" {
		songName = util.SanitizeFilename(service.FileName)
	}
	artist := service.Artist
	if artist == "" {
		artist = "未知艺术家"
	}
	album := service.Album
	if album == "" {
		album = "未知专辑"
	}
	ext := util.FileExtension(service.FileName)
	filename := util.SanitizeFilename(service.FileName)

	checkData := map[string]string{
		"bitrate": service.Bitrate,
		"ext":     "",
		"length":  service.Size,
		"md5":     file.MD5,
		"songId":  "0",
		"version": "1",
	}
	_, checkBody, _ := util.CreateRequest(
		"POST",
		"https://music.163.com/eapi/cloud/upload/check",
		checkData,
		&util.Options{Crypto: "eapi", Url: "/api/cloud/upload/check"},
	)
	needUpload, _ := jsonparser.GetBoolean(checkBody, "needUpload")
	songID, _ := jsonparser.GetString(checkBody, "songId")

	resourceToken, tokenBody, err := util.AllocNosToken(&util.Options{Crypto: "weapi"}, map[string]string{
		"bucket":      "",
		"ext":         ext,
		"filename":    filename,
		"local":       "false",
		"nos_product": "3",
		"type":        "audio",
		"md5":         file.MD5,
	})
	if err != nil {
		return 500, tokenBody
	}

	if needUpload {
		bucket := "jd-musicrep-privatecloud-audio-public"
		uploadToken, uploadBody, err := util.AllocNosToken(&util.Options{Crypto: "weapi"}, map[string]string{
			"bucket":      bucket,
			"ext":         ext,
			"filename":    filename,
			"local":       "false",
			"nos_product": "3",
			"type":        "audio",
			"md5":         file.MD5,
		})
		if err != nil {
			return 500, uploadBody
		}
		uploadServer, err := util.NosUploadServer(bucket)
		if err != nil {
			return 500, []byte(err.Error())
		}
		uploadURL := uploadServer + "/" + bucket + "/" + uploadToken.Result.ObjectKey + "?offset=0&complete=true&version=1.0"
		if _, _, err := util.NosPost(uploadURL, file.Data, map[string]string{
			"x-nos-token":    uploadToken.Result.Token,
			"Content-MD5":    file.MD5,
			"Content-Type":   service.MimeType,
			"Content-Length": service.Size,
		}); err != nil {
			return 500, []byte(err.Error())
		}
	}

	infoData := map[string]string{
		"md5":        file.MD5,
		"songid":     songID,
		"filename":   service.FileName,
		"song":       songName,
		"album":      album,
		"artist":     artist,
		"bitrate":    service.Bitrate,
		"resourceId": resourceToken.Result.ResourceID,
	}
	infoCode, infoBody, _ := util.CreateRequest(
		"POST",
		"https://music.163.com/eapi/upload/cloud/info/v2",
		infoData,
		&util.Options{Crypto: "eapi", Url: "/api/upload/cloud/info/v2"},
	)
	if infoCode != 200 {
		return infoCode, infoBody
	}
	infoSongID, _ := jsonparser.GetString(infoBody, "songId")
	if infoSongID == "" {
		infoSongID = songID
	}
	pubCode, pubBody, _ := util.CreateRequest(
		"POST",
		"https://music.163.com/eapi/cloud/pub/v2",
		map[string]string{"songid": infoSongID},
		&util.Options{Crypto: "eapi", Url: "/api/cloud/pub/v2"},
	)
	result, _ := json.Marshal(map[string]any{
		"code":    200,
		"data":    map[string]any{"songId": infoSongID, "check": json.RawMessage(checkBody)},
		"info":    json.RawMessage(infoBody),
		"pub":     json.RawMessage(pubBody),
		"pubCode": pubCode,
	})
	return 200, result
}
