package service

import (
	"encoding/base64"
	"encoding/json"
	"net/url"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/util"
	qrcode "github.com/skip2/go-qrcode"
)

type VerifyGetQrService struct {
	Vid   string `json:"vid" form:"vid"`
	Type  string `json:"type" form:"type"`
	Token string `json:"token" form:"token"`
	Evid  string `json:"evid" form:"evid"`
	Sign  string `json:"sign" form:"sign"`
}

func (service *VerifyGetQrService) VerifyGetQr() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	params, _ := json.Marshal(map[string]string{
		"event_id": service.Evid,
		"sign":     service.Sign,
	})
	data := make(map[string]string)
	data["verifyConfigId"] = service.Vid
	data["verifyType"] = service.Type
	data["token"] = service.Token
	data["params"] = string(params)
	data["size"] = "150"

	code, body, _ := util.CreateRequest("POST", "https://music.163.com/api/frontrisk/verify/getqrcode", data, options)
	if code != 200 {
		return code, body
	}
	qrCode, _ := jsonparser.GetString(body, "data", "qrCode")
	qrurl := "https://st.music.163.com/encrypt-pages?qrCode=" + url.QueryEscape(qrCode) +
		"&verifyToken=" + url.QueryEscape(service.Token) +
		"&verifyId=" + url.QueryEscape(service.Vid) +
		"&verifyType=" + url.QueryEscape(service.Type) +
		"&params=" + url.QueryEscape(string(params))
	qrimg := ""
	if png, err := qrcode.Encode(qrurl, qrcode.Medium, 256); err == nil {
		qrimg = "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
	}
	result, _ := json.Marshal(map[string]any{
		"code": 200,
		"data": map[string]string{
			"qrCode": qrCode,
			"qrurl":  qrurl,
			"qrimg":  qrimg,
		},
	})
	return 200, result
}
