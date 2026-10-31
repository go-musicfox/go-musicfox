// Package netease provides helpers for NetEase Cloud Music APIs.
package netease

import (
	"net/http"

	"github.com/go-musicfox/netease-music/util"
)

// FetchDailyRecommendPlaylists requests the mobile recommendation batches.
func FetchDailyRecommendPlaylists() (float64, []byte) {
	// Login refresh sets os=pc in the shared jar. Explicitly select mobile mode:
	// PC-mode recommendations can stay unchanged while mobile batches rotate.
	options := &util.Options{
		Crypto:  "weapi",
		Cookies: []*http.Cookie{{Name: "os", Value: "ios"}},
	}
	code, response, _ := util.CreateRequest("POST", "https://music.163.com/weapi/v1/discovery/recommend/resource", map[string]string{}, options)
	return code, response
}
