package netease

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/forgoer/openssl"
	neteaseutil "github.com/go-musicfox/netease-music/util"
)

type recommendRoundTripper func(*http.Request) (*http.Response, error)

func (f recommendRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func setupRecommendTestState(t *testing.T) http.CookieJar {
	t.Helper()
	previousTransport := http.DefaultTransport
	previousJar := neteaseutil.GetGlobalCookieJar()
	previousProxy := neteaseutil.UNMProxyURL
	previousTimeout := neteaseutil.HTTPClientTimeout
	t.Cleanup(func() {
		http.DefaultTransport = previousTransport
		neteaseutil.SetGlobalCookieJar(previousJar)
		neteaseutil.UNMProxyURL = previousProxy
		neteaseutil.HTTPClientTimeout = previousTimeout
	})
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	neteaseutil.SetGlobalCookieJar(jar)
	neteaseutil.UNMProxyURL = ""
	neteaseutil.HTTPClientTimeout = time.Second
	return jar
}

func TestFetchDailyRecommendPlaylistsUsesMobileFeedWithoutChangingProfile(t *testing.T) {
	jar := setupRecommendTestState(t)
	endpoint, err := url.Parse("https://interface.music.163.com/eapi/link/page/rcmd/resource/show")
	if err != nil {
		t.Fatal(err)
	}
	jar.SetCookies(endpoint, []*http.Cookie{
		{Name: "os", Value: "ios", Path: "/"},
		{Name: "appver", Value: "9.0.65", Path: "/"},
		{Name: "MUSIC_U", Value: "fake-login-token", Path: "/"},
	})
	previousCookies := jar.Cookies(endpoint)
	const response = `{"code":200,"data":{"blocks":[{"positionCode":"PAGE_RECOMMEND_SPECIAL_CLOUD_VILLAGE_PLAYLIST","dslData":{"blockResource":{"resources":[{"resourceId":"7","title":"Fresh recommendations"}]}}}]}}`
	requests := 0
	http.DefaultTransport = recommendRoundTripper(func(request *http.Request) (*http.Response, error) {
		requests++
		if request.Method != http.MethodPost || request.URL.String() != endpoint.String() {
			t.Errorf("request = %s %s, want POST %s", request.Method, request.URL, endpoint)
		}
		for name, want := range map[string]string{"os": "ios", "appver": "9.0.65", "MUSIC_U": "fake-login-token"} {
			cookie, err := request.Cookie(name)
			if err != nil {
				t.Errorf("missing %s cookie: %v", name, err)
			} else if cookie.Value != want {
				t.Errorf("%s cookie = %q, want %q", name, cookie.Value, want)
			}
		}
		if request.Header.Get("os") != "" || request.Header.Get("appver") != "" {
			t.Error("recommendation request injected a device profile")
		}
		if err := request.ParseForm(); err != nil {
			t.Fatal(err)
		}
		ciphertext, err := hex.DecodeString(request.PostForm.Get("params"))
		if err != nil || len(ciphertext) == 0 || len(ciphertext)%16 != 0 {
			t.Fatalf("invalid EAPI params: bytes=%d, err=%v", len(ciphertext), err)
		}
		plaintext, err := openssl.AesECBDecrypt(ciphertext, []byte("e82ckenh8dichen8"), openssl.PKCS7_PADDING)
		if err != nil {
			t.Fatal(err)
		}
		parts := strings.SplitN(string(plaintext), "-36cd479b6b5-", 3)
		if len(parts) != 3 || parts[0] != "/api/link/page/rcmd/resource/show" {
			t.Fatalf("unexpected EAPI envelope: %q", plaintext)
		}
		var payload map[string]any
		if err := json.Unmarshal([]byte(parts[1]), &payload); err != nil {
			t.Fatal(err)
		}
		for name, want := range map[string]any{
			"pageCode":           "HOME_RECOMMEND_PAGE",
			"cursor":             float64(1),
			"refresh":            true,
			"blockCodeOrderList": `["PAGE_RECOMMEND_SPECIAL_CLOUD_VILLAGE_PLAYLIST"]`,
			"header":             "{}",
			"e_r":                false,
		} {
			if got := payload[name]; !reflect.DeepEqual(got, want) {
				t.Errorf("payload[%s] = %#v, want %#v", name, got, want)
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
	if got := jar.Cookies(endpoint); !reflect.DeepEqual(got, previousCookies) {
		t.Errorf("recommendation request changed shared cookies: got %v, want %v", got, previousCookies)
	}
}

func TestFetchDailyRecommendPlaylistsRejectsFailedResponses(t *testing.T) {
	for _, tc := range []struct {
		name      string
		status    int
		body      string
		err       error
		wantLogin bool
	}{
		{name: "invalid JSON", status: http.StatusOK, body: "not JSON"},
		{name: "missing code", status: http.StatusOK, body: `{"data":{}}`},
		{name: "invalid code type", status: http.StatusOK, body: `{"code":"200"}`},
		{name: "HTTP failure", status: http.StatusInternalServerError, body: `{"code":200}`},
		{name: "login expired", status: http.StatusOK, body: `{"code":301}`, wantLogin: true},
		{name: "transport failure", err: errors.New("simulated transport failure")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			setupRecommendTestState(t)
			requests := 0
			http.DefaultTransport = recommendRoundTripper(func(request *http.Request) (*http.Response, error) {
				requests++
				if tc.err != nil {
					return nil, tc.err
				}
				return &http.Response{
					StatusCode: tc.status,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(tc.body)),
					Request:    request,
				}, nil
			})
			code, _ := FetchDailyRecommendPlaylists()
			if code == 200 {
				t.Fatal("failed response was reported as successful")
			}
			if tc.wantLogin && code != 301 {
				t.Errorf("expired login code = %v, want 301", code)
			}
			if requests != 1 {
				t.Errorf("made %d requests, want 1 without falling back to another feed", requests)
			}
		})
	}
}

func TestNewDailyRecommendClientHonorsProxyAndTimeout(t *testing.T) {
	jar := setupRecommendTestState(t)
	neteaseutil.UNMProxyURL = "http://proxy.invalid:7890"
	neteaseutil.HTTPClientTimeout = 37 * time.Millisecond
	client, err := newDailyRecommendClient()
	if err != nil {
		t.Fatal(err)
	}
	if client.Timeout != neteaseutil.HTTPClientTimeout {
		t.Errorf("timeout = %v, want %v", client.Timeout, neteaseutil.HTTPClientTimeout)
	}
	if client.Jar != jar {
		t.Error("client does not use the current login jar")
	}
	transport, ok := client.Transport.(*http.Transport)
	if !ok || transport.Proxy == nil {
		t.Fatalf("proxy transport = %T, want *http.Transport with a proxy", client.Transport)
	}
	defer transport.CloseIdleConnections()
	request, err := http.NewRequest(http.MethodPost, "https://interface.music.163.com/eapi/link/page/rcmd/resource/show", nil)
	if err != nil {
		t.Fatal(err)
	}
	proxy, err := transport.Proxy(request)
	if err != nil || proxy == nil || proxy.String() != neteaseutil.UNMProxyURL {
		t.Errorf("proxy = %v, err = %v, want %s", proxy, err, neteaseutil.UNMProxyURL)
	}
	if transport.TLSClientConfig != nil && transport.TLSClientConfig.InsecureSkipVerify {
		t.Error("proxy transport disabled TLS certificate verification")
	}
}

func TestNewDailyRecommendClientRejectsInvalidProxy(t *testing.T) {
	setupRecommendTestState(t)
	neteaseutil.UNMProxyURL = "://invalid-proxy"
	if _, err := newDailyRecommendClient(); err == nil {
		t.Fatal("invalid proxy was silently accepted")
	}
}
