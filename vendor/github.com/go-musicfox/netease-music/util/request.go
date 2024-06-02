package util

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	urlpkg "net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/buger/jsonparser"
	"github.com/cnsilvan/UnblockNeteaseMusic/processor"
	"github.com/go-musicfox/requests"
)

const (
	iosAppVersion = "9.0.65"
)

var (
	globalDeviceId string
	deviceIdOnce   sync.Once
)

type Options struct {
	Crypto  string
	Ua      string
	Cookies []*http.Cookie
	Token   string
	Url     string
	SkipUNM bool
}

func chooseUserAgent(ua string) string {
	userAgentList := map[string]string{
		"mobile": "Mozilla/5.0 (iPhone; CPU iPhone OS 17_4_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4.1 Mobile/15E148 Safari/604.1",
		"pc":     "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Edg/124.0.0.0",
	}
	if ua == "mobile" {
		return userAgentList["mobile"]
	}
	return userAgentList["pc"]
}

var cookieJar http.CookieJar

func SetGlobalCookieJar(jar http.CookieJar) {
	cookieJar = jar
}

func CookieValueByName(cookies []*http.Cookie, name string, fallback string) string {
	var cookie *http.Cookie
	for _, v := range cookies {
		if v.Name == name {
			cookie = v
			break
		}
	}
	if cookie == nil || cookie.Value == "" {
		return fallback
	}
	return cookie.Value
}

func CreateRequest(method, url string, data map[string]string, options *Options) (resCode float64, resResp []byte, resCookies []*http.Cookie) {
	defer func() {
		if resCode != 200 {
			log.Printf("url: %s, method: %s, reqData: %#v, reqOptions: %+v, resCode: %f, resResp: %s, resCookies: %#v", url, method, data, options, resCode, resResp, resCookies)
		}
	}()

	deviceIdOnce.Do(func() {
		idx := rand.Intn(len(deviceIds) - 1)
		globalDeviceId = deviceIds[idx]
	})

	if cookieJar == nil {
		cookieJar, _ = cookiejar.New(&cookiejar.Options{})
	}

	if u, err := urlpkg.Parse(url); err == nil {
		options.Cookies = append(options.Cookies, cookieJar.Cookies(u)...)
	}
	req := requests.Requests()

	var (
		os          = CookieValueByName(options.Cookies, "os", "ios")
		appver      = CookieValueByName(options.Cookies, "appver", Ternary(os != "pc", iosAppVersion, ""))
		osver       = CookieValueByName(options.Cookies, "osver", "17.4.1")
		deviceId    = CookieValueByName(options.Cookies, "deviceId", globalDeviceId)
		versionCode = CookieValueByName(options.Cookies, "versioncode", "140")
		mobileName  = CookieValueByName(options.Cookies, "mobilename", "")
		buildver    = CookieValueByName(options.Cookies, "buildver", strconv.FormatInt(time.Now().Unix(), 10))
		resolution  = CookieValueByName(options.Cookies, "resolution", "1920x1080")
		channel     = CookieValueByName(options.Cookies, "channel", "")
		musicU      = CookieValueByName(options.Cookies, "MUSIC_U", "")
		musicA      = CookieValueByName(options.Cookies, "MUSIC_A", "")
		csrfToken   = CookieValueByName(options.Cookies, "__csrf", "")
	)

	req.Client.Jar = cookieJar
	req.Header.Set("User-Agent", chooseUserAgent(options.Ua))
	req.Header.Set("os", os)
	req.Header.Set("appver", appver)

	options.Cookies = append(options.Cookies,
		&http.Cookie{Name: "__remember_me", Value: "true"},
		&http.Cookie{Name: "os", Value: os},
		&http.Cookie{Name: "appver", Value: appver},
	)
	if musicU != "" {
		options.Cookies = append(options.Cookies, &http.Cookie{Name: "_ntes_nuid", Value: hex.EncodeToString([]byte(RandStringRunes(16)))})
	}

	if method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if strings.Contains(url, "music.163.com") {
		req.Header.Set("Referer", "https://music.163.com")
	}
	for _, cookie := range options.Cookies {
		req.SetCookie(cookie)
	}

	if strings.Contains(url, "login") {
		options.Cookies = append(options.Cookies, &http.Cookie{Name: "NMTID", Value: hex.EncodeToString([]byte(RandStringRunes(16)))})
	}

	if options.Crypto == "weapi" {
		data["csrf_token"] = csrfToken
		data = Weapi(data)
		reg, _ := regexp.Compile(`/\w*api/`)
		url = reg.ReplaceAllString(url, "/weapi/")
	} else if options.Crypto == "linuxapi" {
		linuxApiData := make(map[string]interface{}, 3)
		linuxApiData["method"] = method
		reg, _ := regexp.Compile(`/\w*api/`)
		linuxApiData["url"] = reg.ReplaceAllString(url, "/api/")
		linuxApiData["params"] = data
		data = Linuxapi(linuxApiData)
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.90 Safari/537.36")
		url = "https://music.163.com/api/linux/forward"
	} else if options.Crypto == "eapi" {
		eapiData := make(map[string]interface{})
		for key, value := range data {
			eapiData[key] = value
		}
		rand.Seed(time.Now().UnixNano())
		header := map[string]string{
			"osver":       osver,
			"deviceId":    deviceId,
			"appver":      appver,
			"versioncode": versionCode,
			"mobilename":  mobileName,
			"buildver":    buildver,
			"resolution":  resolution,
			"__csrf":      csrfToken,
			"os":          os,
			"channel":     channel,
			"requestId":   strconv.FormatInt(time.Now().Unix()*1000, 10) + strconv.Itoa(rand.Intn(1000)),
		}

		if musicU != "" {
			header["MUSIC_U"] = musicU
		}
		if musicA != "" {
			header["MUSIC_A"] = musicA
		}

		for key, value := range header {
			req.SetCookie(&http.Cookie{Name: key, Value: value, Path: "/"})
		}
		eapiData["header"] = header
		data = Eapi(options.Url, eapiData)
		reg, _ := regexp.Compile(`/\w*api/`)
		url = reg.ReplaceAllString(url, "/eapi/")
	}

	var (
		err     error
		resp    *requests.Response
		UNMFlag = UNMSwitch && !options.SkipUNM
	)
	if method == "POST" {
		var form requests.Datas = data
		resp, err = req.Post(url, requests.DryRun(UNMFlag), form)
	} else {
		resp, err = req.Get(url, requests.DryRun(UNMFlag))
	}
	if err != nil {
		resCode, resResp, resCookies = 520, []byte(err.Error()), nil
		return
	}

	if UNMFlag {
		ConfigInit()

		request := req.HttpRequest()
		netease := processor.RequestBefore(request)
		if netease == nil {
			resCode, resResp, resCookies = 520, []byte("Request Blocked:"+url), nil
			return
		}

		if method == "POST" {
			var form requests.Datas = data
			resp, err = req.Post(url, form)
		} else {
			resp, err = req.Get(url)
		}
		if err != nil {
			resCode, resResp, resCookies = 520, []byte("Request Error:"+url), nil
			return
		}
		response := resp.R
		defer response.Body.Close()

		processor.RequestAfter(request, response, netease)
		resp.ReloadContent()
	}

	resCookies = resp.Cookies()

	resResp = resp.Content()
	//fmt.Println(string(body))
	b := bytes.NewReader(resResp)
	var out bytes.Buffer
	r, err := zlib.NewReader(b)
	// 数据被压缩 进行解码
	if err == nil {
		_, _ = io.Copy(&out, r)
		resResp = out.Bytes()
	}

	resCode, err = jsonparser.GetFloat(resResp, "code")
	if err != nil {
		resCode = 200
	}
	return
}
