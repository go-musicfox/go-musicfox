package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/go-musicfox/netease-music/util"
)

type VoiceUploadService struct {
	FileName         string `json:"fileName" form:"fileName"`
	MimeType         string `json:"mimeType" form:"mimeType"`
	SongName         string `json:"songName" form:"songName"`
	Description      string `json:"description" form:"description"`
	VoiceListId      string `json:"voiceListId" form:"voiceListId"`
	CoverImgId       string `json:"coverImgId" form:"coverImgId"`
	CategoryId       string `json:"categoryId" form:"categoryId"`
	SecondCategoryId string `json:"secondCategoryId" form:"secondCategoryId"`
	ComposedSongs    string `json:"composedSongs" form:"composedSongs"`
	Privacy          string `json:"privacy" form:"privacy"`
	PublishTime      string `json:"publishTime" form:"publishTime"`
	OrderNo          string `json:"orderNo" form:"orderNo"`
	AutoPublish      string `json:"autoPublish" form:"autoPublish"`
	AutoPublishText  string `json:"autoPublishText" form:"autoPublishText"`
	Data             []byte `json:"-" form:"-"`
}

type initiateMultipartUploadResult struct {
	UploadID string `xml:"UploadId"`
}

func voiceDupkey() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	h := hex.EncodeToString(buf)
	return fmt.Sprintf("%s-%s-%s-%s-%s", h[0:8], h[8:12], h[12:16], h[16:20], h[20:32])
}

func (service *VoiceUploadService) VoiceUpload() (float64, []byte) {

	file := &util.NosUploadFile{Name: service.FileName, MimeType: service.MimeType, Data: service.Data}
	file.Prepare()
	ext := util.FileExtension(service.FileName)
	filename := service.SongName
	if filename == "" {
		filename = util.SanitizeFilename(service.FileName)
	}
	if service.MimeType == "" {
		service.MimeType = "audio/mpeg"
	}

	token, body, err := util.AllocNosToken(&util.Options{Crypto: "weapi"}, map[string]string{
		"bucket":      "ymusic",
		"ext":         ext,
		"filename":    filename,
		"local":       "false",
		"nos_product": "0",
		"type":        "other",
	})
	if err != nil {
		return 500, body
	}
	objectKey := strings.ReplaceAll(token.Result.ObjectKey, "/", "%2F")
	baseURL := "https://ymusic.nos-hz.163yun.com/" + objectKey

	initBody, _, err := util.NosPost(baseURL+"?uploads", nil, map[string]string{
		"x-nos-token":             token.Result.Token,
		"X-Nos-Meta-Content-Type": service.MimeType,
	})
	if err != nil {
		return 500, []byte(err.Error())
	}
	var initResult initiateMultipartUploadResult
	if err := xml.Unmarshal(initBody, &initResult); err != nil || initResult.UploadID == "" {
		return 500, append([]byte("initiate multipart failed: "), initBody...)
	}

	const blockSize = 10 * 1024 * 1024
	etags := make([]string, 0)
	for offset, part := 0, 1; offset < len(file.Data); offset, part = offset+blockSize, part+1 {
		end := offset + blockSize
		if end > len(file.Data) {
			end = len(file.Data)
		}
		_, headers, err := util.NosPut(
			fmt.Sprintf("%s?partNumber=%d&uploadId=%s", baseURL, part, initResult.UploadID),
			file.Data[offset:end],
			map[string]string{
				"x-nos-token":  token.Result.Token,
				"Content-Type": service.MimeType,
			},
		)
		if err != nil {
			return 500, []byte(err.Error())
		}
		etags = append(etags, headers.Get("ETag"))
	}

	var complete strings.Builder
	complete.WriteString("<CompleteMultipartUpload>")
	for i, etag := range etags {
		complete.WriteString(fmt.Sprintf("<Part><PartNumber>%d</PartNumber><ETag>%s</ETag></Part>", i+1, etag))
	}
	complete.WriteString("</CompleteMultipartUpload>")
	if _, _, err := util.NosPost(baseURL+"?uploadId="+initResult.UploadID, []byte(complete.String()), map[string]string{
		"Content-Type":            "text/plain;charset=UTF-8",
		"X-Nos-Meta-Content-Type": service.MimeType,
		"x-nos-token":             token.Result.Token,
	}); err != nil {
		return 500, []byte(err.Error())
	}

	composed := make([]string, 0)
	if service.ComposedSongs != "" {
		composed = strings.Split(service.ComposedSongs, ",")
	}
	publishTime := service.PublishTime
	if publishTime == "" {
		publishTime = "0"
	}
	orderNo := service.OrderNo
	if orderNo == "" {
		orderNo = "1"
	}
	voiceData, _ := json.Marshal([]map[string]any{{
		"name":             filename,
		"autoPublish":      service.AutoPublish == "1",
		"autoPublishText":  service.AutoPublishText,
		"description":      service.Description,
		"voiceListId":      service.VoiceListId,
		"coverImgId":       service.CoverImgId,
		"dfsId":            token.Result.DocID,
		"categoryId":       service.CategoryId,
		"secondCategoryId": service.SecondCategoryId,
		"composedSongs":    composed,
		"privacy":          service.Privacy == "1",
		"publishTime":      publishTime,
		"orderNo":          orderNo,
	}})

	options := &util.Options{
		Crypto:  "eapi",
		Url:     "/api/voice/workbench/voice/batch/upload/preCheck",
		Headers: map[string]string{"x-nos-token": token.Result.Token},
	}
	preCheckData := map[string]string{"dupkey": voiceDupkey(), "voiceData": string(voiceData)}
	if code, preBody, _ := util.CreateRequest(
		"POST",
		"https://music.163.com/eapi/voice/workbench/voice/batch/upload/preCheck",
		preCheckData,
		options,
	); code != 200 {
		return code, preBody
	}

	uploadData := map[string]string{"dupkey": voiceDupkey(), "voiceData": string(voiceData)}
	code, uploadBody, _ := util.CreateRequest(
		"POST",
		"https://music.163.com/eapi/voice/workbench/voice/batch/upload/v2",
		uploadData,
		&util.Options{
			Crypto:  "eapi",
			Url:     "/api/voice/workbench/voice/batch/upload/v2",
			Headers: map[string]string{"x-nos-token": token.Result.Token},
		},
	)
	result, _ := json.Marshal(map[string]any{
		"code": 200,
		"data": json.RawMessage(uploadBody),
	})
	return code, result
}
