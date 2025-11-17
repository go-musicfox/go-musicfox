package util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	math_rand "math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

func stringToEnt(s string) string {
	var builder strings.Builder
	for _, r := range s {
		if r > 255 {
			builder.WriteString(fmt.Sprintf("&#%d;", r))
		} else {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

// 生成_ntes_nuid
func GenerateNtesUID(
	timestamp int64,
	location string,
	screenWidth, screenHeight int,
	userAgent string,
	random float64,
	clientWidth, clientHeight int,
) string {
	clientDimensions := fmt.Sprintf("%d:%d", clientWidth, clientHeight)
	rawString := strconv.FormatInt(timestamp, 10) +
		location +
		strconv.Itoa(screenWidth) +
		strconv.Itoa(screenHeight) +
		userAgent +
		strconv.FormatFloat(random, 'f', -1, 64) +
		clientDimensions
	encodedString := stringToEnt(rawString)
	hasher := md5.New()
	hasher.Write([]byte(encodedString))
	hashBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashBytes)
}

func generateBrowserClientDimensions(screenWidth, screenHeight int) (clientWidth, clientHeight int) {
	verticalOffset := math_rand.N(150-90+1) + 90
	clientHeight = screenHeight - verticalOffset

	horizontalOffset := 0
	if math_rand.N(2) == 1 {
		horizontalOffset = 17
	}
	clientWidth = screenWidth - horizontalOffset

	return clientWidth, clientHeight
}
