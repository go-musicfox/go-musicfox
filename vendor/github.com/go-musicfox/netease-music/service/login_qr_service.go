package service

import (
	"encoding/json"
	"log"

	"github.com/go-musicfox/netease-music/util"
)

type LoginQRService struct {
	UniKey string `json:"unikey"`
}

// GetKey 获取要生成二维码的QrcodeUrl
//
// 返回：
//
//   - code: 状态码
//   - bodyByte：返回的响应体
//   - 获取到的Unikey
//   - error
func (service *LoginQRService) GetKey() (float64, []byte, string, error) {
	data := map[string]interface{}{
		"type":         1,
		"noCheckToken": true,
	}

	api := "https://music.163.com/weapi/login/qrcode/unikey"
	code, bodyBytes, err := util.CallWeapi(api, data)
	if err != nil {
		return code, bodyBytes, "", err
	}
	if code != 200 || len(bodyBytes) == 0 {
		return code, bodyBytes, "", err
	}
	err = json.Unmarshal(bodyBytes, service)
	if err != nil {
		log.Fatalf("Error unmarshalling bodybytes: %v", err)
	}

	// 生成 chainId，这个是新版本新加的参数
	cookieJar := util.GetGlobalCookieJar()
	chainID := util.GenerateChainID(cookieJar)
	qrcodeUrl := ("http://music.163.com/login?codekey=" +
		service.UniKey + "&chainId=" + chainID)
	return code, bodyBytes, qrcodeUrl, nil
}

func (service *LoginQRService) CheckQR() (float64, []byte, error) {
	if service.UniKey == "" {
		return 0, nil, nil
	}
	data := map[string]interface{}{
		"type":         1,
		"noCheckToken": true,
		"key":          service.UniKey,
	}

	api := "https://music.163.com/weapi/login/qrcode/client/login"
	cookieJar := util.GetGlobalCookieJar()
	util.ApplyRequestStrategy(cookieJar)
	code, bodyBytes, err := util.CallWeapi(api, data, cookieJar)
	if err != nil {
		return code, bodyBytes, err
	}
	return code, bodyBytes, nil
}
