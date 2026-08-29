// Package qrlogin provides the UI-free QR login client for Netease Cloud
// Music, reusable by the TUI / WebUI / headless frontends.
package qrlogin

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-musicfox/netease-music/util"
	"github.com/imroc/req/v3"
)

// qrLoginClient 使用 Chrome TLS 指纹模拟的二维码登录专用 HTTP 客户端。
//
// 背景：网易云音乐的二维码登录接口对 TLS 指纹有检测，标准 Go http.Client
// 的 TLS 握手特征与浏览器差异较大，容易被服务端识别并拒绝（返回 462）。
// 通过 imroc/req/v3 的 SetTLSFingerprintChrome() 模拟 Chrome 浏览器的
// TLS ClientHello 指纹，让服务端认为请求来自真实的 Chrome 浏览器。
const qrLoginUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"

// The unikey endpoint must not receive cookies from the shared login jar.
var qrKeyClient = req.C().
	SetUserAgent(qrLoginUserAgent).
	SetCookieJar(nil).
	SetTLSFingerprintChrome()

// newQRLoginClient binds the current persistent cookie jar when polling starts.
// The global jar can be replaced after package initialization, so it must not be
// captured by a package-level client. Otherwise login cookies such as MUSIC_U
// would be written to a stale jar and unavailable to subsequent account requests.
func newQRLoginClient(cookieJar http.CookieJar) *req.Client {
	return req.C().
		SetUserAgent(qrLoginUserAgent).
		SetCookieJar(cookieJar).
		SetTLSFingerprintChrome()
}

// GetKey 使用 Chrome TLS 指纹请求二维码登录的 unikey 和 qrcodeUrl。
//
// 替代 service.LoginQRService.GetKey()，底层逻辑保持一致：
// 1. 对请求参数进行 weapi 加密
// 2. POST 到 /weapi/login/qrcode/unikey
// 3. 解析响应获取 unikey，拼接 qrcodeUrl
func GetKey(cookieJar http.CookieJar) (uniKey string, qrcodeUrl string, err error) {
	data := map[string]interface{}{
		"type":         1,
		"noCheckToken": true,
	}

	encodedParams, err := util.ApiParamsEncode(data)
	if err != nil {
		return "", "", fmt.Errorf("failed to encode api params: %w", err)
	}

	api := "https://music.163.com/weapi/login/qrcode/unikey"
	resp, err := qrKeyClient.R().
		SetHeaders(map[string]string{
			"Referer":            "https://music.163.com/",
			"Origin":             "https://music.163.com",
			"Accept-Language":    "zh-CN,zh;q=0.9,en;q=0.8",
			"Cache-Control":      "no-cache",
			"Pragma":             "no-cache",
			"Nm-Gcore-Status":    "1",
			"Priority":           "u=1, i",
			"Sec-Ch-Ua":          `"Not;A=Brand";v="8", "Chromium";v="150", "Google Chrome";v="150"`,
			"Sec-Ch-Ua-Mobile":   "?0",
			"Sec-Ch-Ua-Platform": `"Windows"`,
			"Sec-Fetch-Dest":     "empty",
			"Sec-Fetch-Mode":     "cors",
			"Sec-Fetch-Site":     "same-origin",
			"X-Channelsource":    "undefined",
			"X-Os":               "web",
		}).
		SetFormData(encodedParams).
		Post(api)
	if err != nil {
		return "", "", fmt.Errorf("failed to send qr key request: %w", err)
	}

	bodyBytes := resp.Bytes()
	var result struct {
		Code   float64 `json:"code"`
		UniKey string  `json:"unikey"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", "", fmt.Errorf("failed to unmarshal qr key response: %w", err)
	}
	if result.Code != 200 || result.UniKey == "" {
		return "", "", fmt.Errorf("无法获取二维码, code: %.0f", result.Code)
	}

	chainID := util.GenerateChainID(cookieJar)
	qrcodeUrl = "http://music.163.com/login?codekey=" + result.UniKey + "&chainId=" + chainID

	return result.UniKey, qrcodeUrl, nil
}

// CheckStatus 使用 Chrome TLS 指纹轮询二维码扫码状态。
//
// 替代 service.LoginQRService.CheckQR()，底层逻辑保持一致：
// 1. 注入反风控 cookie
// 2. 对请求参数进行 weapi 加密
// 3. POST 到 /weapi/login/qrcode/client/login
// 4. 解析响应中的 code 字段
func CheckStatus(uniKey string, cookieJar http.CookieJar) (code float64, respBytes []byte, err error) {
	if uniKey == "" {
		return 0, nil, nil
	}

	util.ApplyRequestStrategy(cookieJar)

	data := map[string]interface{}{
		"type":         1,
		"noCheckToken": true,
		"key":          uniKey,
	}

	encodedParams, err := util.ApiParamsEncode(data)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to encode api params: %w", err)
	}

	api := "https://music.163.com/weapi/login/qrcode/client/login"
	resp, err := newQRLoginClient(cookieJar).R().
		SetHeaders(map[string]string{
			"Referer": "https://music.163.com/",
			"Origin":  "https://music.163.com",
		}).
		SetFormData(encodedParams).
		Post(api)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to send qr check request: %w", err)
	}

	respBytes = resp.Bytes()
	var result struct {
		Code float64 `json:"code"`
	}
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return 0, respBytes, fmt.Errorf("failed to unmarshal qr check response: %w", err)
	}

	return result.Code, respBytes, nil
}
