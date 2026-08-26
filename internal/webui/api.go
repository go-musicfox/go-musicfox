package webui

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	apputils "github.com/go-musicfox/go-musicfox/utils/app"
)

// handleStatus answers the same "status" snapshot as the WS control channel,
// wrapped as {"ok":true,"data":{...}} (same shape as the WS snapshot frame).
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	data, err := s.dispatcher.Dispatch(r.Context(), "status", nil)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "data": data})
}

// albumArtClient fetches album art behind the Referer header so the music.163
// CDN does not reject the backend proxy request as hotlinking.
var albumArtClient = &http.Client{Timeout: 10 * time.Second}

// handleAlbumArt proxies the current song's cover image. The PicUrl is
// validated against a strict allowlist (http/https + .music.163.com suffix) to
// prevent SSRF through a tampered song record. No caching (YAGNI).
func (s *Server) handleAlbumArt(w http.ResponseWriter, r *http.Request) {
	if s.engine == nil {
		http.NotFound(w, r)
		return
	}
	info := s.engine.Player().PlayingInfo()
	if info.PicUrl == "" {
		http.NotFound(w, r)
		return
	}
	u, err := url.Parse(info.PicUrl)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		http.NotFound(w, r)
		return
	}
	if host := u.Hostname(); host != "music.163.com" && !strings.HasSuffix(host, ".music.163.com") {
		http.NotFound(w, r)
		return
	}

	picURL := info.PicUrl
	if size := r.URL.Query().Get("size"); size != "" {
		if n, err := strconv.Atoi(size); err == nil && n > 0 {
			picURL = apputils.AddResizeParamForPicUrl(picURL, int64(n))
		}
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, picURL, nil)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	req.Header.Set("Referer", "https://music.163.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; musicfox-webui)")

	resp, err := albumArtClient.Do(req)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, resp.Body)
}

// lyricFragment is the wire form of one lyric line.
type lyricFragment struct {
	StartTimeMs int64  `json:"startTimeMs"`
	Content     string `json:"content"`
}

// lyricsResponse is the wire shape of GET /api/lyrics.
type lyricsResponse struct {
	Fragments           []lyricFragment  `json:"fragments"`
	TranslatedFragments map[int64]string `json:"translatedFragments"`
	CurrentIndex        int              `json:"currentIndex"`
	OffsetMs            int64            `json:"offsetMs"`
}

// handleLyrics answers the current lyric service state. No song / no lyrics
// yields an empty structure (not 500).
func (s *Server) handleLyrics(w http.ResponseWriter, r *http.Request) {
	if s.engine == nil {
		http.NotFound(w, r)
		return
	}
	st := s.engine.LyricService().State()

	fragments := make([]lyricFragment, 0, len(st.Fragments))
	for _, f := range st.Fragments {
		fragments = append(fragments, lyricFragment{StartTimeMs: f.StartTimeMs, Content: f.Content})
	}
	trans := st.TranslatedFragments
	if trans == nil {
		trans = map[int64]string{}
	}

	writeJSON(w, http.StatusOK, lyricsResponse{
		Fragments:           fragments,
		TranslatedFragments: trans,
		CurrentIndex:        st.CurrentIndex,
		OffsetMs:            st.OffsetMs,
	})
}
