package app

import (
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
		c.Expires = time.Now().Add(365 * 24 * time.Hour)
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
