package service

import (
	"encoding/json"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/util"
)

type CloudUploadTokenService struct {
	Md5      string `json:"md5" form:"md5"`
	FileSize string `json:"fileSize" form:"fileSize"`
	Filename string `json:"filename" form:"filename"`
	Bitrate  string `json:"bitrate" form:"bitrate"`
}

func (service *CloudUploadTokenService) CloudUploadToken() (float64, []byte) {

	if service.Bitrate == "" {
		service.Bitrate = "999000"
	}
	ext := util.FileExtension(service.Filename)
	checkData := map[string]string{
		"bitrate": service.Bitrate,
		"ext":     "",
		"length":  service.FileSize,
		"md5":     service.Md5,
		"songId":  "0",
		"version": "1",
	}
	code, checkBody, _ := util.CreateRequest(
		"POST",
		"https://music.163.com/eapi/cloud/upload/check",
		checkData,
		&util.Options{Crypto: "eapi", Url: "/api/cloud/upload/check"},
	)
	if code != 200 {
		return code, checkBody
	}
	needUpload, _ := jsonparser.GetBoolean(checkBody, "needUpload")
	songID, _ := jsonparser.GetString(checkBody, "songId")

	bucket := "jd-musicrep-privatecloud-audio-public"
	tokenData := map[string]string{
		"bucket":      bucket,
		"ext":         ext,
		"filename":    util.SanitizeFilename(service.Filename),
		"local":       "false",
		"nos_product": "3",
		"type":        "audio",
		"md5":         service.Md5,
	}
	token, tokenBody, err := util.AllocNosToken(&util.Options{Crypto: "weapi"}, tokenData)
	if err != nil {
		return 500, tokenBody
	}
	uploadServer, err := util.NosUploadServer(bucket)
	if err != nil {
		return 500, []byte(err.Error())
	}
	uploadURL := uploadServer + "/" + bucket + "/" + token.Result.ObjectKey + "?offset=0&complete=true&version=1.0"
	result, _ := json.Marshal(map[string]any{
		"code": 200,
		"data": map[string]any{
			"needUpload":  needUpload,
			"songId":      songID,
			"uploadToken": token.Result.Token,
			"objectKey":   token.Result.ObjectKey,
			"resourceId":  token.Result.ResourceID,
			"uploadUrl":   uploadURL,
			"bucket":      bucket,
			"md5":         service.Md5,
			"fileSize":    service.FileSize,
			"filename":    service.Filename,
		},
	})
	return 200, result
}
