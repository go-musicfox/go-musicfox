package util

import (
	"regexp"
	"strings"
	"time"

	"github.com/go-musicfox/requests"
)

var checkTokenPattern = regexp.MustCompile(`null\(\[(\d+),\d+,"([^"]+)"\]\)`)

// fetchCheckToken 获取易盾反作弊 token。v2 依赖浏览器 JS SDK，当前实现仅支持 v3。
func fetchCheckToken(version string) string {
	if version != "v3" {
		return ""
	}
	req := requests.Requests()
	req.Client.Timeout = 10 * time.Second
	req.Header.Set("User-Agent", chooseUserAgent("mobile"))
	resp, err := req.Get("https://ac.dun.163yun.com/v3/b?pn=YD00000558929251")
	if err != nil {
		return ""
	}
	body := resp.Text()
	m := checkTokenPattern.FindStringSubmatch(body)
	if len(m) == 3 && m[1] == "200" {
		return strings.TrimSpace(m[2])
	}
	return ""
}
