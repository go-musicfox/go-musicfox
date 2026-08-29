package webui

import (
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"
	"net/url"
)

// tokenCookieName is the HttpOnly session cookie set by the token exchange
// endpoint and required by every protected route.
const tokenCookieName = "musicfox_webui_token"

// allowedHosts is the loopback host whitelist. r.Host is attacker-controlled
// via the Host header even on a loopback listener, so it is validated
// explicitly (ports are stripped before matching).
var allowedHosts = map[string]struct{}{
	"127.0.0.1": {},
	"localhost": {},
	"::1":       {},
}

// verifyToken compares got against want in constant time so a remote attacker
// cannot time the comparison to leak the token byte by byte. Empty values are
// rejected outright (ConstantTimeCompare would otherwise match empty vs empty).
func verifyToken(got, want string) bool {
	if got == "" || want == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}

// validHost reports whether host (with any port stripped) is a whitelisted
// loopback hostname. Hosts without a port compare directly (e.g. HTTP/1.0
// requests with a bare "localhost" Host).
func validHost(host string) bool {
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	_, ok := allowedHosts[host]
	return ok
}

// originAllowed implements the Origin layer: an absent Origin is allowed
// (plain navigation / non-CORS requests), "null" and any cross-origin Origin
// are rejected, and the response never reflects Access-Control-Allow-Origin.
func originAllowed(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	if origin == "null" {
		return false
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	// The loopback listener has no TLS, so an https Origin can never be
	// same-origin with us.
	if u.Scheme != "http" {
		return false
	}
	return u.Host == r.Host
}

// authStatus runs the security layers in order and returns the HTTP status to
// send on failure (0 = all layers passed):
//
//	① token cookie — 401 Unauthorized
//	② Host whitelist — 403 Forbidden
//	③ Origin check — 403 Forbidden
func (s *Server) authStatus(r *http.Request) int {
	switch {
	case !s.hasValidCookie(r):
		return http.StatusUnauthorized
	case !validHost(r.Host):
		return http.StatusForbidden
	case !originAllowed(r):
		return http.StatusForbidden
	default:
		return 0
	}
}

// hasValidCookie reports whether the request carries a cookie matching the
// server token.
func (s *Server) hasValidCookie(r *http.Request) bool {
	c, err := r.Cookie(tokenCookieName)
	return err == nil && verifyToken(c.Value, s.token)
}

// authMiddleware wraps an API handler with the token-cookie + Host + Origin
// layers. Any failure answers an empty 401/403 body (no detail leaked). It is
// reserved for the /api/* and /ws routes (T4/T6/T7); the static root and the
// token exchange endpoint stay unauthenticated.
func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if status := s.authStatus(r); status != 0 {
			http.Error(w, "", status)
			return
		}
		next(w, r)
	}
}

// handleTokenExchange is the first-visit handshake: the browser URL carries
// ?token=<server token>. On success the token is exchanged for an HttpOnly
// session cookie (Path=/, SameSite=Strict) and the client is redirected to the
// app root. A wrong or missing token answers 403 without setting a cookie.
func (s *Server) handleTokenExchange(w http.ResponseWriter, r *http.Request) {
	if !verifyToken(r.URL.Query().Get("token"), s.token) {
		http.Error(w, "", http.StatusForbidden)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     tokenCookieName,
		Value:    s.token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   0, // session cookie: no Max-Age/Expires attribute
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

// verifyWSRequest applies the same cookie + host + origin layers used by
// authMiddleware so the WebSocket route (T4) can reject an Accept before
// upgrading. Unlike authMiddleware it does not write a response: a failed WS
// upgrade has no HTTP response body path and the caller decides how to surface
// the rejection.
func (s *Server) verifyWSRequest(r *http.Request) error {
	if status := s.authStatus(r); status != 0 {
		return fmt.Errorf("webui ws auth rejected (%s)", http.StatusText(status))
	}
	return nil
}
