package service

import (
	"encoding/json"

	"github.com/go-musicfox/netease-music/util"
)

type PlaylistCoverUpdateService struct {
	Id       string `json:"id" form:"id"`
	FileName string `json:"fileName" form:"fileName"`
	MimeType string `json:"mimeType" form:"mimeType"`
	Data     []byte `json:"-" form:"-"`
}

func (service *PlaylistCoverUpdateService) PlaylistCoverUpdate() (float64, []byte) {

	if service.FileName == "" {
		service.FileName = "cover.jpg"
	}
	if service.MimeType == "" {
		service.MimeType = "image/jpeg"
	}
	token, body, err := util.AllocNosToken(&util.Options{Crypto: "weapi"}, map[string]string{
		"bucket":      "yyimgs",
		"ext":         "jpg",
		"filename":    service.FileName,
		"local":       "false",
		"nos_product": "0",
		"return_body": `{"code":200,"size":"$(ObjectSize)"}`,
		"type":        "other",
	})
	if err != nil {
		return 500, body
	}
	uploadURL := "https://nosup-hz1.127.net/yyimgs/" + token.Result.ObjectKey + "?offset=0&complete=true&version=1.0"
	if _, _, err := util.NosPost(uploadURL, service.Data, map[string]string{
		"x-nos-token":  token.Result.Token,
		"Content-Type": service.MimeType,
	}); err != nil {
		return 500, []byte(err.Error())
	}

	code, updateBody, _ := util.CreateRequest(
		"POST",
		"https://music.163.com/api/playlist/cover/update",
		map[string]string{"id": service.Id, "coverImgId": token.Result.DocID},
		&util.Options{Crypto: "weapi"},
	)
	result, _ := json.Marshal(map[string]any{
		"code": 200,
		"data": map[string]any{
			"url_pre": "https://p1.music.126.net/" + token.Result.ObjectKey,
			"imgId":   token.Result.DocID,
			"update":  json.RawMessage(updateBody),
		},
	})
	if code != 200 {
		return code, updateBody
	}
	return 200, result
}
