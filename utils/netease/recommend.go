// Package netease provides helpers for NetEase Cloud Music APIs.
package netease

import (
	"net/http"

	"github.com/go-musicfox/netease-music/util"
)

// FetchDailyRecommendPlaylists requests the current PC recommendation feed.
func FetchDailyRecommendPlaylists() (float64, []byte) {
	// Without a desktop profile this endpoint can return the legacy mobile feed.
	options := &util.Options{
		Crypto:  "weapi",
		Cookies: []*http.Cookie{{Name: "os", Value: "pc"}, {Name: "appver", Value: "3.1.41"}},
	}
	code, response, _ := util.CreateRequest("POST", "https://music.163.com/weapi/pc/personal/page/playlist/rcmd", map[string]string{}, options)
	return code, response
}
