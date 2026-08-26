package webui

import (
	"encoding/json"
	"net/http"

	neteaseutil "github.com/go-musicfox/netease-music/util"
	cookiejar "github.com/juju/persistent-cookiejar"
	"github.com/skip2/go-qrcode"

	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/core/qrlogin"
)

// qrGetKey and qrCheckStatus are package-level overridable so tests can stub
// the netease network calls (mirrors the UserService overridable pattern in
// internal/core/services.go).
var qrGetKey = qrlogin.GetKey
var qrCheckStatus = qrlogin.CheckStatus

// completeQRLogin runs the post-scan login completion (app jar replacement +
// LoginCallback, both network-bound). Package-level overridable so tests can
// stub it without hitting the netease account API.
var completeQRLogin = func(e *core.Engine, jar *cookiejar.Jar) error {
	return e.CompleteQRLogin(jar)
}

// qrHTTPJar returns the app cookie jar when available (the persistent jar the
// netease login cookies land in), falling back to the global jar otherwise.
func qrHTTPJar() http.CookieJar {
	if jar := core.AppCookieJar(); jar != nil {
		return jar
	}
	return neteaseutil.GetGlobalCookieJar()
}

// handleLoginQRKey answers the QR login unikey and its qrcode URL.
func (s *Server) handleLoginQRKey(w http.ResponseWriter, r *http.Request) {
	uniKey, qrcodeUrl, err := qrGetKey(qrHTTPJar())
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"uniKey": uniKey, "qrcodeUrl": qrcodeUrl})
}

// handleLoginQRImage renders the QR PNG for a unikey. The qrcode URL is
// reconstructed the same way qrlogin.GetKey builds it (the chainId is a
// per-request tracking token and need not match the key response).
func (s *Server) handleLoginQRImage(w http.ResponseWriter, r *http.Request) {
	uniKey := r.URL.Query().Get("key")
	if uniKey == "" {
		writeJSONError(w, http.StatusBadRequest, "missing key")
		return
	}
	chainID := neteaseutil.GenerateChainID(qrHTTPJar())
	qrcodeUrl := "http://music.163.com/login?codekey=" + uniKey + "&chainId=" + chainID
	png, err := qrcode.Encode(qrcodeUrl, qrcode.Medium, 256)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "二维码生成失败")
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(png)
}

// handleLoginQRStatus polls the scan status. Code semantics: 800 = expired /
// invalid, 801 = awaiting scan, 802 = scanned, awaiting confirm, 803 =
// confirmed — the login is completed synchronously here (CompleteQRLogin is
// network-bound and blocks this connection for at most a few seconds; the
// frontend is waiting for 803 anyway).
func (s *Server) handleLoginQRStatus(w http.ResponseWriter, r *http.Request) {
	uniKey := r.URL.Query().Get("key")
	if uniKey == "" {
		writeJSONError(w, http.StatusBadRequest, "missing key")
		return
	}

	code, _, err := qrCheckStatus(uniKey, qrHTTPJar())
	if err != nil {
		// Surface as error JSON so the frontend shows 「获取二维码失败」 instead
		// of polling an invalid key forever.
		writeJSONError(w, http.StatusBadGateway, "获取二维码失败")
		return
	}
	if code != 803 {
		writeJSON(w, http.StatusOK, map[string]any{"code": code})
		return
	}

	if s.engine == nil {
		writeJSONError(w, http.StatusInternalServerError, "engine unavailable")
		return
	}
	var appJar *cookiejar.Jar
	if j := core.AppCookieJar(); j != nil {
		appJar = j
	}
	if err := completeQRLogin(s.engine, appJar); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	user := s.engine.User()
	userData := map[string]any{}
	if user != nil {
		userData = map[string]any{
			"userId":    user.UserId,
			"nickname":  user.Nickname,
			"avatarUrl": user.AvatarUrl,
		}
	}
	s.broadcaster.broadcast(eventFrame("login", map[string]any{"user": userData}))
	writeJSON(w, http.StatusOK, map[string]any{"code": 803, "message": "登录成功"})
}

// writeJSON writes data as a JSON response with the given status.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// writeJSONError writes an error JSON body: {"ok":false,"error":...}.
func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"ok": false, "error": msg})
}
