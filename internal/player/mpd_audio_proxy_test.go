package player

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

// TestMPDAudioProxy 验证代理核心行为:
//   - 无 Range: 200 全量透传, 且上游收到浏览器 UA
//   - 上游忽略 Range 返回 200 全量: 代理按 206 对齐返回正确字节区间
//   - wrapURL/isOurs 往返正确
func TestMPDAudioProxy(t *testing.T) {
	content := []byte("0123456789abcdefghijklmnopqrstuvwxyz") // 36 bytes

	var gotUA, gotRange string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		gotRange = r.Header.Get("Range")
		w.Header().Set("Content-Type", "audio/flac")
		w.Header().Set("Accept-Ranges", "bytes")
		// 模拟网易云 CDN: 无论是否带 Range 都返回 200 全量数据
		w.Header().Set("Content-Length", strconv.Itoa(len(content)))
		_, _ = w.Write(content)
	}))
	defer upstream.Close()

	proxy, err := newMPDAudioProxy()
	if err != nil {
		t.Fatal(err)
	}
	defer proxy.Close()

	proxied := proxy.wrapURL(upstream.URL + "/x.flac?authSecret=abc&d=1")
	if !proxy.isOurs(proxied) {
		t.Fatal("wrapURL result should be recognized as ours")
	}

	// 1. 无 Range → 200 全量透传, 且注入浏览器 UA
	req, _ := http.NewRequest(http.MethodGet, proxied, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if string(body) != string(content) {
		t.Fatal("passthrough body mismatch")
	}
	if gotUA == "" || gotUA == "Go-http-client/1.1" {
		t.Fatalf("expected browser user agent on upstream request, got %q", gotUA)
	}

	// 2. Range: bytes=10- → 上游返回 200 全量 → 代理应返回 206 且数据从偏移 10 对齐
	req2, _ := http.NewRequest(http.MethodGet, proxied, nil)
	req2.Header.Set("Range", "bytes=10-")
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatal(err)
	}
	body2, _ := io.ReadAll(resp2.Body)
	_ = resp2.Body.Close()
	if resp2.StatusCode != http.StatusPartialContent {
		t.Fatalf("expected 206, got %d", resp2.StatusCode)
	}
	if string(body2) != string(content[10:]) {
		t.Fatalf("range body mismatch: got %q want %q", body2, content[10:])
	}
	if cr := resp2.Header.Get("Content-Range"); cr != "bytes 10-35/36" {
		t.Fatalf("unexpected content-range: %q", cr)
	}
	if gotRange != "bytes=10-" {
		t.Fatalf("expected range forwarded upstream, got %q", gotRange)
	}
}

// TestMPDAudioProxyUpstream206 验证上游正确处理 Range 时 206 原样透传。
func TestMPDAudioProxyUpstream206(t *testing.T) {
	content := []byte("0123456789abcdefghijklmnopqrstuvwxyz")
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "audio/flac")
		w.Header().Set("Accept-Ranges", "bytes")
		if r.Header.Get("Range") == "bytes=5-" {
			w.Header().Set("Content-Range", "bytes 5-35/36")
			w.Header().Set("Content-Length", strconv.Itoa(len(content)-5))
			w.WriteHeader(http.StatusPartialContent)
			_, _ = w.Write(content[5:])
			return
		}
		w.Header().Set("Content-Length", strconv.Itoa(len(content)))
		_, _ = w.Write(content)
	}))
	defer upstream.Close()

	proxy, err := newMPDAudioProxy()
	if err != nil {
		t.Fatal(err)
	}
	defer proxy.Close()

	req, _ := http.NewRequest(http.MethodGet, proxy.wrapURL(upstream.URL+"/a.flac"), nil)
	req.Header.Set("Range", "bytes=5-")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusPartialContent {
		t.Fatalf("expected 206, got %d", resp.StatusCode)
	}
	if string(body) != string(content[5:]) {
		t.Fatalf("body mismatch: got %q want %q", body, content[5:])
	}
}

func TestIsNetEaseHost(t *testing.T) {
	cases := map[string]bool{
		"https://m801.music.126.net/obj/flac?authSecret=1":  true,
		"http://music.163.com/package/xxx":                  true,
		"https://example.com/a.flac":                        false,
		"https://m801.music.126.net.evil.com/a.flac":        false,
		":not-a-url":                                        false,
	}
	for u, want := range cases {
		if got := isNetEaseHost(u); got != want {
			t.Errorf("isNetEaseHost(%q) = %v, want %v", u, got, want)
		}
	}
}

func TestParseRangeStart(t *testing.T) {
	for in, want := range map[string]int64{
		"bytes=0-":    0,
		"bytes=1024-": 1024,
	} {
		got, err := parseRangeStart(in)
		if err != nil || got != want {
			t.Errorf("parseRangeStart(%q) = %d, %v; want %d", in, got, err, want)
		}
	}
	for _, in := range []string{"", "bytes=-100", "bytes=1-2", "items=1-", "bytes=abc-"} {
		if _, err := parseRangeStart(in); err == nil {
			t.Errorf("parseRangeStart(%q) expected error", in)
		}
	}
}
