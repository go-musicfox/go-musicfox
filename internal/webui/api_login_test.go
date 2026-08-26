package webui

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	cookiejar "github.com/juju/persistent-cookiejar"

	"github.com/go-musicfox/go-musicfox/internal/core"
)

// TestLoginQRKey mocks qrGetKey and verifies the unikey/qrcodeUrl response.
func TestLoginQRKey(t *testing.T) {
	prev := qrGetKey
	qrGetKey = func(_ http.CookieJar) (string, string, error) {
		return "uniKey-abc", "http://music.163.com/login?codekey=uniKey-abc&chainId=v1_x", nil
	}
	t.Cleanup(func() { qrGetKey = prev })

	s, ts := newWSServer(t)
	resp := doAuthedRequest(t, ts, "/api/login/qr/key", s.token, "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var body struct {
		UniKey    string `json:"uniKey"`
		QrcodeUrl string `json:"qrcodeUrl"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.UniKey != "uniKey-abc" {
		t.Fatalf("uniKey = %q, want uniKey-abc", body.UniKey)
	}
	if body.QrcodeUrl == "" {
		t.Fatal("qrcodeUrl empty")
	}
}

// TestLoginQRKeyError verifies an upstream failure answers error JSON instead
// of a fake unikey (the frontend must show 「获取二维码失败」, not dead-loop).
func TestLoginQRKeyError(t *testing.T) {
	prev := qrGetKey
	qrGetKey = func(_ http.CookieJar) (string, string, error) {
		return "", "", errors.New("net unreachable")
	}
	t.Cleanup(func() { qrGetKey = prev })

	s, ts := newWSServer(t)
	resp := doAuthedRequest(t, ts, "/api/login/qr/key", s.token, "", "")
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", resp.StatusCode)
	}
	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if body["ok"] != false || body["error"] == nil {
		t.Fatalf("error body = %v, want {ok:false, error}", body)
	}
}

// TestLoginQRImage renders a PNG for a unikey (pure local encode, no mocks).
func TestLoginQRImage(t *testing.T) {
	s, ts := newWSServer(t)
	resp := doAuthedRequest(t, ts, "/api/login/qr/image?key=uniKey-abc", s.token, "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "image/png" {
		t.Fatalf("content-type = %q, want image/png", ct)
	}
	body, _ := io.ReadAll(resp.Body)
	if len(body) < 8 || body[0] != 0x89 || body[1] != 'P' || body[2] != 'N' || body[3] != 'G' {
		t.Fatalf("body is not a PNG (len=%d)", len(body))
	}
}

// TestLoginQRImageMissingKey verifies a missing key answers 400.
func TestLoginQRImageMissingKey(t *testing.T) {
	s, ts := newWSServer(t)
	resp := doAuthedRequest(t, ts, "/api/login/qr/image", s.token, "", "")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

// TestLoginQRStatus801 verifies the pending-scan code passes through.
func TestLoginQRStatus801(t *testing.T) {
	prev := qrCheckStatus
	qrCheckStatus = func(_ string, _ http.CookieJar) (float64, []byte, error) {
		return 801, []byte(`{"code":801}`), nil
	}
	t.Cleanup(func() { qrCheckStatus = prev })

	s, ts := newWSServer(t)
	resp := doAuthedRequest(t, ts, "/api/login/qr/status?key=uniKey-abc", s.token, "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var body struct {
		Code float64 `json:"code"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if body.Code != 801 {
		t.Fatalf("code = %v, want 801", body.Code)
	}
}

// TestLoginQRStatusError verifies a check failure answers error JSON.
func TestLoginQRStatusError(t *testing.T) {
	prev := qrCheckStatus
	qrCheckStatus = func(_ string, _ http.CookieJar) (float64, []byte, error) {
		return 0, nil, errors.New("request blocked")
	}
	t.Cleanup(func() { qrCheckStatus = prev })

	s, ts := newWSServer(t)
	resp := doAuthedRequest(t, ts, "/api/login/qr/status?key=uniKey-abc", s.token, "", "")
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", resp.StatusCode)
	}
}

// TestLoginQRStatus803 verifies the confirmed-scan path: the login completion
// (stubbed) runs and a login event is broadcast over the WS connection.
func TestLoginQRStatus803(t *testing.T) {
	prevStatus := qrCheckStatus
	qrCheckStatus = func(_ string, _ http.CookieJar) (float64, []byte, error) {
		return 803, []byte(`{"code":803}`), nil
	}
	t.Cleanup(func() { qrCheckStatus = prevStatus })

	prevComplete := completeQRLogin
	completeQRLogin = func(_ *core.Engine, _ *cookiejar.Jar) error { return nil }
	t.Cleanup(func() { completeQRLogin = prevComplete })

	s, ts := newWSServer(t)
	c, _ := wsDialAuthed(t, s, ts)

	resp := doAuthedRequest(t, ts, "/api/login/qr/status?key=uniKey-abc", s.token, "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var body struct {
		Code float64 `json:"code"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if body.Code != 803 {
		t.Fatalf("code = %v, want 803", body.Code)
	}

	event, data := readEventFrame(t, c)
	if event != "login" {
		t.Fatalf("event = %q, want login", event)
	}
	if _, ok := data["user"]; !ok {
		t.Fatalf("login event missing user: %v", data)
	}
}
