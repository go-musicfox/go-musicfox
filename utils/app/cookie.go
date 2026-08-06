package app

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-musicfox/netease-music/service"
	neteaseutil "github.com/go-musicfox/netease-music/util"
	cookiejar "github.com/juju/persistent-cookiejar"
)

var ParseCookieError = errors.New("解析cookie失败")

const ExpiresDuration = 365 * 24 * time.Hour

// ParseCookieFromStr 从字符串解析 Cookie 并存入 CookieJar
func ParseCookieFromStr(cookieStr string, jar http.CookieJar) error {
	cookies, err := http.ParseCookie(cookieStr)
	if err != nil {
		return fmt.Errorf("%w: %w", ParseCookieError, err)
	}
	// 补全持久化所需的元数据
	targetURL := "https://music.163.com"
	u, _ := url.Parse(targetURL)
	var finalCookies []*http.Cookie

	for _, c := range cookies {
		c.Path = "/"
		c.Expires = time.Now().Add(ExpiresDuration)
		finalCookies = append(finalCookies, c)
	}
	if jar != nil {
		jar.SetCookies(
			u,
			finalCookies,
		)
	}

	return nil
}

// RefreshCookieJar 刷新 CookieJar 并返回新的实例
func RefreshCookieJar() (jar *cookiejar.Jar, err error) {
	refreshLoginService := service.LoginRefreshService{}
	code, _, err := refreshLoginService.LoginRefresh()

	if err != nil {
		return jar, fmt.Errorf("Token 刷新网络请求失败: %w", err)
	} else if code == 200 {
		globalJar := neteaseutil.GetGlobalCookieJar()
		if jar, ok := globalJar.(*cookiejar.Jar); ok {
			return jar, nil
		} else {
			return jar, fmt.Errorf("Token 刷新成功但类型转换失败")
		}
	} else {
		return jar, fmt.Errorf("Token 刷新失败, Code: %d", int(code))
	}
}

// ApplyLoginStrategy 向全局 CookieJar 注入网易云反风控参数（os=pc + 随机 NMTID）。
// 与 SDK 的 util.ApplyRequestStrategy 对齐，并将其写死的固定 NMTID 覆盖为随机值，
// 避免 qrcode/unikey、cellphone、email 等登录接口在 weapi 裸请求下被服务端风控拦截。
func ApplyLoginStrategy() {
	jar := neteaseutil.GetGlobalCookieJar()
	neteaseutil.ApplyRequestStrategy(jar)
	u, _ := url.Parse("https://music.163.com")
	jar.SetCookies(u, []*http.Cookie{
		{Name: "NMTID", Value: randomHexString(32)},
	})
}

// randomHexString 返回长度为 n 的随机十六进制字符串。
func randomHexString(n int) string {
	b := make([]byte, n/2)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
