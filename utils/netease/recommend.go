// Package netease provides helpers for NetEase Cloud Music APIs.
package netease

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/util"
)

// FetchDailyRecommendPlaylists requests the mobile home page's playlist block.
func FetchDailyRecommendPlaylists() (float64, []byte) {
	const path = "/api/link/page/rcmd/resource/show"
	// Preserve the JSON string header and boolean flags instead of letting
	// CreateRequest replace the header and stringify every parameter.
	params := util.Eapi(path, map[string]interface{}{
		"pageCode":           "HOME_RECOMMEND_PAGE",
		"cursor":             1,
		"refresh":            true,
		"blockCodeOrderList": `["PAGE_RECOMMEND_SPECIAL_CLOUD_VILLAGE_PLAYLIST"]`,
		"header":             "{}",
		"e_r":                false,
	})
	form := url.Values{"params": {params["params"]}}
	request, err := http.NewRequest(http.MethodPost, "https://interface.music.163.com/eapi/link/page/rcmd/resource/show", strings.NewReader(form.Encode()))
	if err != nil {
		return 0, nil
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client, err := newDailyRecommendClient()
	if err != nil {
		return 0, nil
	}
	if util.UNMProxyURL != "" {
		defer client.CloseIdleConnections()
	}
	response, err := client.Do(request)
	if err != nil {
		return 0, nil
	}
	defer func() { _ = response.Body.Close() }()
	const maxBody = 4 << 20
	body, err := io.ReadAll(io.LimitReader(response.Body, maxBody+1))
	if err != nil || len(body) > maxBody {
		return 0, nil
	}
	if response.StatusCode != http.StatusOK {
		return float64(response.StatusCode), body
	}
	if !json.Valid(body) {
		return 0, nil
	}
	code, err := jsonparser.GetFloat(body, "code")
	if err != nil {
		return 0, nil
	}
	return code, body
}

func newDailyRecommendClient() (*http.Client, error) {
	transport := http.DefaultTransport
	if util.UNMProxyURL != "" {
		proxyURL, err := url.Parse(util.UNMProxyURL)
		if err != nil || proxyURL.Host == "" {
			return nil, errors.New("invalid recommendation proxy URL")
		}
		base, ok := http.DefaultTransport.(*http.Transport)
		if !ok {
			return nil, errors.New("unsupported recommendation proxy transport")
		}
		proxyTransport := base.Clone()
		proxyTransport.Proxy = http.ProxyURL(proxyURL)
		transport = proxyTransport
	}
	return &http.Client{
		Transport: transport,
		Jar:       util.GetGlobalCookieJar(),
		Timeout:   util.HTTPClientTimeout,
	}, nil
}
