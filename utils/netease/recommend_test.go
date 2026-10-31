package netease

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"

	neteaseutil "github.com/go-musicfox/netease-music/util"
)

type recommendRoundTripper func(*http.Request) (*http.Response, error)

func (f recommendRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestFetchDailyRecommendPlaylistsOverridesPCWithoutLosingLogin(t *testing.T) {
	previousTransport := http.DefaultTransport
	previousJar := neteaseutil.GetGlobalCookieJar()
	previousProxy := neteaseutil.UNMProxyURL
	previousUNM := neteaseutil.UNMSwitch
	t.Cleanup(func() {
		http.DefaultTransport = previousTransport
		neteaseutil.SetGlobalCookieJar(previousJar)
		neteaseutil.UNMProxyURL = previousProxy
		neteaseutil.UNMSwitch = previousUNM
	})

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	endpoint, err := url.Parse("https://music.163.com/weapi/v1/discovery/recommend/resource")
	if err != nil {
		t.Fatal(err)
	}
	jar.SetCookies(endpoint, []*http.Cookie{
		{Name: "os", Value: "pc", Path: "/"},
		{Name: "MUSIC_U", Value: "fake-login-token", Path: "/"},
	})
	neteaseutil.SetGlobalCookieJar(jar)
	neteaseutil.UNMProxyURL = ""
	neteaseutil.UNMSwitch = false

	const response = `{"code":200,"recommend":[{"id":7,"name":"Fresh recommendations"}]}`
	requests := 0
	http.DefaultTransport = recommendRoundTripper(func(request *http.Request) (*http.Response, error) {
		requests++
		if request.Method != http.MethodPost || request.URL.String() != endpoint.String() {
			t.Errorf("request = %s %s, want POST %s", request.Method, request.URL, endpoint)
		}
		if got := request.Header.Get("os"); got != "ios" {
			t.Errorf("os header = %q, want ios", got)
		}
		for name, want := range map[string]string{"os": "ios", "MUSIC_U": "fake-login-token"} {
			cookie, err := request.Cookie(name)
			if err != nil {
				t.Errorf("missing %s cookie: %v", name, err)
			} else if cookie.Value != want {
				t.Errorf("effective %s cookie = %q, want %q", name, cookie.Value, want)
			}
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(response)),
			Request:    request,
		}, nil
	})

	code, body := FetchDailyRecommendPlaylists()
	if requests != 1 {
		t.Errorf("made %d requests, want 1", requests)
	}
	if code != 200 || string(body) != response {
		t.Errorf("response = (%v, %q), want (200, %q)", code, body, response)
	}
}
