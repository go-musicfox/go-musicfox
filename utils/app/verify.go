package app

import (
	"encoding/json"
	"fmt"

	neteaseutil "github.com/go-musicfox/netease-music/util"
)

// VerifyData 登录接口返回 -462 时携带的风控验证会话数据
type VerifyData struct {
	VerifyID    float64      `json:"verifyId"`
	VerifyType  float64      `json:"verifyType"`
	VerifyToken string       `json:"verifyToken"`
	Params      VerifyParams `json:"params"`
}

// VerifyParams 验证会话签名参数
type VerifyParams struct {
	EventID string `json:"event_id"`
	Sign    string `json:"sign"`
}

// GetVerifyQRCode 获取人机验证二维码
// Verify endpoints run as bare weapi requests: make sure os=pc + random
// NMTID are present instead of relying on the login flow to inject them.
func GetVerifyQRCode(verify VerifyData) (qrCode string, err error) {
	ApplyLoginStrategy()
	params, _ := json.Marshal(verify.Params)
	data := map[string]interface{}{
		"verifyConfigId": verify.VerifyID,
		"verifyType":     verify.VerifyType,
		"token":          verify.VerifyToken,
		"params":         string(params),
		"size":           150,
	}
	code, bodyBytes, err := neteaseutil.CallWeapi("https://music.163.com/api/frontrisk/verify/getqrcode", data)
	if err != nil {
		return "", err
	}
	if code != 200 {
		return "", fmt.Errorf("获取验证二维码失败, code: %.0f", code)
	}
	var resp struct {
		Data struct {
			QRCode string `json:"qrCode"`
		} `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &resp); err != nil {
		return "", err
	}
	if resp.Data.QRCode == "" {
		return "", fmt.Errorf("验证二维码为空")
	}
	return resp.Data.QRCode, nil
}

// CheckVerifyQRCodeStatus 轮询验证二维码状态
// qrCodeStatus: 0=已生成待扫码 10=已扫码待确认 20=验证成功 21=二维码已失效
// Verify endpoints run as bare weapi requests: make sure os=pc + random
// NMTID are present instead of relying on the login flow to inject them.
func CheckVerifyQRCodeStatus(qrCode string) (status float64, err error) {
	ApplyLoginStrategy()
	data := map[string]interface{}{"qrCode": qrCode}
	code, bodyBytes, err := neteaseutil.CallWeapi("https://music.163.com/api/frontrisk/verify/qrcodestatus", data)
	if err != nil {
		return 0, err
	}
	if code != 200 {
		return 0, fmt.Errorf("查询验证状态失败, code: %.0f", code)
	}
	var resp struct {
		Data struct {
			QRCodeStatus float64 `json:"qrCodeStatus"`
		} `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &resp); err != nil {
		return 0, err
	}
	return resp.Data.QRCodeStatus, nil
}
