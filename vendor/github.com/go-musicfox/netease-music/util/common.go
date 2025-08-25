package util

import (
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

// RandStringRunes 返回随机字符串
func RandStringRunes(n int) string {
	var letterRunes = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[r.Intn(len(letterRunes))]
	}
	return string(b)
}

func StringOr(a, b string, others ...string) string {
	if a != "" {
		return a
	}
	if b != "" {
		return b
	}
	for _, v := range others {
		if v != "" {
			return v
		}
	}
	return ""
}

func Ternary[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

// 获取cookiejar中存储的csrf_token
func GetCsrfToken(cookieJar http.CookieJar) string {
	musicURL, err := url.Parse("https://music.163.com")
	if err != nil {
		log.Fatalf("Failed to parse music URL: %v", err)
	}
	csrfToken := ""
	if cookieJar != nil {
		for _, cookie := range cookieJar.Cookies(musicURL) {
			if cookie.Name == "__csrf" {
				csrfToken = cookie.Value
				break
			}
		}
	}
	return csrfToken
}

// 将 cookies 添加到指定的 CookieJar 中
func AddCookiesToJar(jar http.CookieJar, cookies map[string]string, targetURLStr string) {
	if jar == nil {
		log.Fatal("CookieJar的值不能为 nil")
	}

	targetURL, err := url.Parse(targetURLStr)
	if err != nil {
		log.Fatalf("无法解析 URL '%s': %v", targetURLStr, err)
	}

	var cookiesToSet []*http.Cookie
	for name, value := range cookies {
		cookie := &http.Cookie{
			Name:   name,
			Value:  value,
			Path:   "/",
			Domain: targetURL.Hostname(),
		}
		cookiesToSet = append(cookiesToSet, cookie)
	}

	if len(cookiesToSet) > 0 {
		jar.SetCookies(targetURL, cookiesToSet)
	}
}
