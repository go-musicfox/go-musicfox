package webui

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// newTestServer builds a Server (nil engine: the auth surface exercised here
// never touches the dispatcher) and mounts a protected test endpoint behind
// authMiddleware on top of the real mux, then serves it over httptest.
func newTestServer(t *testing.T) (*Server, *httptest.Server) {
	t.Helper()
	s := NewServer(nil)
	s.mux.HandleFunc("GET /api/test", s.authMiddleware(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	ts := httptest.NewServer(s.mux)
	t.Cleanup(ts.Close)
	return s, ts
}

// doAuthedRequest issues a GET against path with an optional token cookie and
// optional Host/Origin overrides, never following redirects.
func doAuthedRequest(t *testing.T, ts *httptest.Server, path, cookie, host, origin string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, ts.URL+path, nil)
	if err != nil {
		t.Fatalf("NewRequest(%s): %v", path, err)
	}
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: tokenCookieName, Value: cookie})
	}
	if host != "" {
		req.Host = host
	}
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

// TestSecurityAuthMiddleware drives the four auth layers through a real
// authMiddleware-wrapped route.
func TestSecurityAuthMiddleware(t *testing.T) {
	tests := []struct {
		name   string
		cookie string // "valid" = real server token, "wrong" = garbage, "" = absent
		host   string // empty = keep the request URL host
		origin string // empty = no Origin header
		want   int
	}{
		{name: "no cookie", cookie: "", want: http.StatusUnauthorized},
		{name: "wrong cookie", cookie: "wrong", want: http.StatusUnauthorized},
		{name: "evil host", cookie: "valid", host: "evil.com", want: http.StatusForbidden},
		{name: "non-loopback ip host", cookie: "valid", host: "10.0.0.1", want: http.StatusForbidden},
		{name: "null origin", cookie: "valid", origin: "null", want: http.StatusForbidden},
		{name: "cross origin", cookie: "valid", origin: "http://evil.com", want: http.StatusForbidden},
		{name: "https origin", cookie: "valid", origin: "https://127.0.0.1", want: http.StatusForbidden},
		{name: "different port origin", cookie: "valid", origin: "http://127.0.0.1:1", want: http.StatusForbidden},
		{name: "malformed origin", cookie: "valid", origin: "://bad", want: http.StatusForbidden},
		{name: "no origin", cookie: "valid", want: http.StatusOK},
		{name: "same origin", cookie: "valid", origin: "same", want: http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, ts := newTestServer(t)
			cookie := s.token
			switch tt.cookie {
			case "valid":
			case "wrong":
				cookie = "deadbeefdeadbeef"
			case "":
				cookie = ""
			}
			origin := tt.origin
			if origin == "same" {
				origin = ts.URL
			}
			resp := doAuthedRequest(t, ts, "/api/test", cookie, tt.host, origin)
			if resp.StatusCode != tt.want {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.want)
			}
			if tt.want != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				// http.Error always appends a trailing newline; require the
				// body to carry no real content so no detail leaks.
				if strings.TrimSpace(string(body)) != "" {
					t.Fatalf("auth failure body = %q, want empty (no detail leak)", body)
				}
			}
		})
	}
}

// TestSecurityLocalhostHost ensures the whitelist matches localhost with a
// port (the httptest server only listens on 127.0.0.1, so this proves the
// port-stripping path rather than the raw host value).
func TestSecurityLocalhostHost(t *testing.T) {
	s, ts := newTestServer(t)
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("parse ts.URL: %v", err)
	}
	resp := doAuthedRequest(t, ts, "/api/test", s.token, "localhost:"+u.Port(), "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("localhost status = %d, want 200", resp.StatusCode)
	}
}

// TestSecurityTokenExchange checks the first-visit handshake: correct URL
// token → 302 with the HttpOnly session cookie; the static root is reachable
// with that cookie.
func TestSecurityTokenExchange(t *testing.T) {
	s, ts := newTestServer(t)
	resp := doAuthedRequest(t, ts, "/token?token="+s.token, "", "", "")
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("exchange status = %d, want 302", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/" {
		t.Fatalf("Location = %q, want %q", loc, "/")
	}
	var got *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == tokenCookieName {
			got = c
			break
		}
	}
	if got == nil {
		t.Fatal("Set-Cookie missing musicfox_webui_token")
	}
	if got.Value != s.token {
		t.Fatalf("cookie value = %q, want server token", got.Value)
	}
	if !got.HttpOnly {
		t.Error("cookie not HttpOnly")
	}
	if got.Path != "/" {
		t.Errorf("cookie Path = %q, want /", got.Path)
	}
	if got.SameSite != http.SameSiteStrictMode {
		t.Errorf("cookie SameSite = %v, want Strict", got.SameSite)
	}
	if got.MaxAge != 0 {
		t.Errorf("cookie MaxAge = %d, want 0 (session cookie)", got.MaxAge)
	}
}

// TestSecurityTokenExchangeRejectsBadToken ensures a wrong token answers 403
// and never sets the auth cookie.
func TestSecurityTokenExchangeRejectsBadToken(t *testing.T) {
	_, ts := newTestServer(t)
	resp := doAuthedRequest(t, ts, "/token?token=wrong", "", "", "")
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("bad-token status = %d, want 403", resp.StatusCode)
	}
	for _, c := range resp.Cookies() {
		if c.Name == tokenCookieName {
			t.Fatal("bad token must not set the auth cookie")
		}
	}
}

// TestSecurityVerifyWSRequest drives the shared cookie+host+origin validation
// used by the T4 WebSocket route before the upgrade Accept.
func TestSecurityVerifyWSRequest(t *testing.T) {
	s := NewServer(nil)
	tests := []struct {
		name    string
		cookie  string // empty = absent
		host    string
		origin  string
		wantErr bool
	}{
		{name: "valid", cookie: "valid", wantErr: false},
		{name: "no cookie", wantErr: true},
		{name: "evil host", cookie: "valid", host: "evil.com", wantErr: true},
		{name: "null origin", cookie: "valid", origin: "null", wantErr: true},
		{name: "cross origin", cookie: "valid", origin: "http://evil.com", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/ws", nil)
			if tt.cookie == "valid" {
				req.AddCookie(&http.Cookie{Name: tokenCookieName, Value: s.token})
			}
			if tt.host != "" {
				req.Host = tt.host
			}
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			err := s.verifyWSRequest(req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("verifyWSRequest err = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
