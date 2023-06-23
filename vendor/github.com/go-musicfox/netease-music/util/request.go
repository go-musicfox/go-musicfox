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
	"time"

	"github.com/buger/jsonparser"
	"github.com/cnsilvan/UnblockNeteaseMusic/processor"
	"github.com/go-musicfox/requests"
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
	userAgentList := []string{
		"Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1",
		"Mozilla/5.0 (Linux; Android 5.0; SM-G900P Build/LRX21T) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 5.1.1; Nexus 6 Build/LYZ28E) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Mobile Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 10_3_2 like Mac OS X) AppleWebKit/603.2.4 (KHTML, like Gecko) Mobile/14F89;GameHelper",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 10_0 like Mac OS X) AppleWebKit/602.1.38 (KHTML, like Gecko) Version/10.0 Mobile/14A300 Safari/602.1",
		"Mozilla/5.0 (iPad; CPU OS 10_0 like Mac OS X) AppleWebKit/602.1.38 (KHTML, like Gecko) Version/10.0 Mobile/14A300 Safari/602.1",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.12; rv:46.0) Gecko/20100101 Firefox/46.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_5) AppleWebKit/603.2.4 (KHTML, like Gecko) Version/10.1.1 Safari/603.2.4",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:46.0) Gecko/20100101 Firefox/46.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36 Edge/13.10586",
	}

	rand.Seed(time.Now().UnixNano())
	index := 0
	if ua == "" {
		index = rand.Intn(len(userAgentList))
	} else if ua == "mobile" {
		index = rand.Intn(8)
	} else {
		index = rand.Intn(7) + 7
	}
	return userAgentList[index]
}

var cookieJar http.CookieJar

func SetGlobalCookieJar(jar http.CookieJar) {
	cookieJar = jar
}

func CreateRequest(method, url string, data map[string]string, options *Options) (resCode float64, resResp []byte, resCookies []*http.Cookie) {
	defer func() {
		if resCode != 200 {
			log.Printf("url: %s, method: %s, reqData: %#v, reqOptions: %+v, resCode: %f, resResp: %s, resCookies: %#v", url, method, data, options, resCode, resResp, resCookies)
		}
	}()

	if cookieJar == nil {
		cookieJar, _ = cookiejar.New(&cookiejar.Options{})
	}

	if u, err := urlpkg.Parse(url); err == nil {
		options.Cookies = append(options.Cookies, cookieJar.Cookies(u)...)
	}
	req := requests.Requests()

	req.Client.Jar = cookieJar
	req.Header.Set("User-Agent", chooseUserAgent(options.Ua))
	csrfToken := ""
	musicU := ""

	if method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if strings.Contains(url, "music.163.com") {
		req.Header.Set("Referer", "https://music.163.com")
	}
	if options.Cookies != nil {
		for _, cookie := range options.Cookies {
			req.SetCookie(cookie)
			if cookie.Name == "__csrf" {
				csrfToken = cookie.Value
			}
			if cookie.Name == "MUSIC_U" {
				musicU = cookie.Value
				cookieNuid := &http.Cookie{Name: "_ntes_nuid", Value: hex.EncodeToString([]byte(RandStringRunes(16)))}
				req.SetCookie(cookieNuid)
			}
		}
	}

	if musicU == "" {
		req.SetCookie(&http.Cookie{Name: "MUSIC_A", Value: ""})
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
			"osver":       "",
			"deviceId":    "",
			"mobilename":  "",
			"appver":      "6.1.1",
			"versioncode": "140",
			"buildver":    strconv.FormatInt(time.Now().Unix(), 10),
			"resolution":  "1920x1080",
			"os":          "android",
			"channel":     "",
			"requestId":   strconv.FormatInt(time.Now().Unix()*1000, 10) + strconv.Itoa(rand.Intn(1000)),
			"MUSIC_U":     musicU,
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
