package externalapi

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/gorilla/websocket"
)

const defaultPort = 14558

type Server struct {
	provider Provider
	options  Options
	http     *http.Server
	listener net.Listener
	hub      *hub
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func Start(provider Provider, options Options) (*Server, error) {
	if !options.Enable {
		return nil, nil
	}
	if options.Port == 0 {
		options.Port = defaultPort
	}
	host := "127.0.0.1"
	if options.AllowLAN {
		host = "0.0.0.0"
	}
	listener, err := net.Listen("tcp", net.JoinHostPort(host, strconv.Itoa(options.Port)))
	if err != nil {
		return nil, fmt.Errorf("listen external API: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	s := &Server{
		provider: provider,
		options:  options,
		listener: listener,
		hub:      newHub(),
		cancel:   cancel,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/info", s.handleInfo)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/now-playing", s.handleNowPlaying)
	mux.HandleFunc("/api/lyrics", s.handleLyrics)
	mux.HandleFunc("/api/play", s.controlHandler(s.provider.ExternalAPIPlay))
	mux.HandleFunc("/api/pause", s.controlHandler(s.provider.ExternalAPIPause))
	mux.HandleFunc("/api/stop", s.controlHandler(s.provider.ExternalAPIStop))
	mux.HandleFunc("/api/next", s.controlHandler(s.provider.ExternalAPINext))
	mux.HandleFunc("/api/prev", s.controlHandler(s.provider.ExternalAPIPrevious))
	mux.HandleFunc("/api/seek", s.handleSeek)
	mux.HandleFunc("/api/volume", s.handleVolume)
	mux.HandleFunc("/ws", s.handleWebSocket)

	s.http = &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	s.wg.Add(2)
	go func() {
		defer s.wg.Done()
		if err := s.http.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("external API server stopped", "error", err)
		}
	}()
	go func() {
		defer s.wg.Done()
		s.broadcastLoop(ctx)
	}()
	slog.Info("SPlayer-compatible external API started", "address", listener.Addr().String(), "allowLan", options.AllowLAN)
	return s, nil
}

func (s *Server) Close(ctx context.Context) error {
	s.cancel()
	s.hub.close()
	err := s.http.Shutdown(ctx)
	s.wg.Wait()
	return err
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	if !allowGet(w, r) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"name":      "go-musicfox",
		"version":   types.AppVersion,
		"wsClients": s.hub.count(),
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if !allowGet(w, r) {
		return
	}
	writeJSON(w, http.StatusOK, s.provider.ExternalAPISnapshot().Status)
}

func (s *Server) handleNowPlaying(w http.ResponseWriter, r *http.Request) {
	if !allowGet(w, r) {
		return
	}
	writeJSON(w, http.StatusOK, s.provider.ExternalAPISnapshot().NowPlaying)
}

func (s *Server) handleLyrics(w http.ResponseWriter, r *http.Request) {
	if !allowGet(w, r) {
		return
	}
	writeJSON(w, http.StatusOK, s.provider.ExternalAPISnapshot().Lyrics)
}

func (s *Server) controlHandler(action func()) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !allowControl(w, r) {
			return
		}
		action()
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

func (s *Server) handleSeek(w http.ResponseWriter, r *http.Request) {
	if !allowControl(w, r) {
		return
	}
	var request struct {
		PositionMs *float64 `json:"positionMs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || !validNonNegative(request.PositionMs) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "positionMs (number, >=0) required"})
		return
	}
	s.provider.ExternalAPISeek(int64(*request.PositionMs))
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleVolume(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if !allowGet(w, r) {
			return
		}
		writeJSON(w, http.StatusOK, map[string]float64{"volume": s.provider.ExternalAPISnapshot().Status.Volume})
		return
	}
	if !allowControl(w, r) {
		return
	}
	var request struct {
		Volume *float64 `json:"volume"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.Volume == nil ||
		math.IsNaN(*request.Volume) || math.IsInf(*request.Volume, 0) || *request.Volume < 0 || *request.Volume > 1 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "volume (number, 0..1) required"})
		return
	}
	s.provider.ExternalAPISetVolume(*request.Volume)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func validNonNegative(value *float64) bool {
	return value != nil && !math.IsNaN(*value) && !math.IsInf(*value, 0) && *value >= 0
}

func (s *Server) broadcastLoop(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	var lastTrackID string
	var lastLyricHash [32]byte
	var lastState PlayerState
	var lastLPCTrackID string
	var lastStatusAt time.Time
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			snapshot := s.provider.ExternalAPISnapshot()
			trackID := ""
			if snapshot.NowPlaying.Track != nil {
				trackID = snapshot.NowPlaying.Track.ID
			}
			if trackID != lastTrackID {
				lastTrackID = trackID
				s.hub.broadcast("track", map[string]any{"track": snapshot.NowPlaying.Track})
			}
			lyricHash := sha256.Sum256(mustJSON(snapshot.Lyrics.Lyric))
			if snapshot.Lyrics.TrackID != lastLPCTrackID || lyricHash != lastLyricHash {
				lastLPCTrackID = snapshot.Lyrics.TrackID
				lastLyricHash = lyricHash
				s.hub.broadcast("lyric", map[string]any{"source": snapshot.Lyrics.Source, "lyric": snapshot.Lyrics.Lyric})
			}
			if snapshot.NowPlaying.State != lastState || time.Since(lastStatusAt) >= time.Second {
				lastState = snapshot.NowPlaying.State
				lastStatusAt = time.Now()
				s.hub.broadcast("status", snapshot.Status)
			}
		}
	}
}

func mustJSON(value any) []byte {
	data, _ := json.Marshal(value)
	return data
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if !s.options.EnableWS {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "WebSocket disabled"})
		return
	}
	if !allowedOrigin(r.Header.Get("Origin")) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "origin not allowed"})
		return
	}
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		return allowedOrigin(r.Header.Get("Origin"))
	}}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := s.hub.add(conn)
	defer s.hub.remove(client)

	snapshot := s.provider.ExternalAPISnapshot()
	_ = client.writeJSON(map[string]any{"kind": "hello", "clients": s.hub.count()})
	if snapshot.NowPlaying.Track != nil {
		_ = client.writeJSON(map[string]any{"kind": "event", "type": "track", "data": map[string]any{"track": snapshot.NowPlaying.Track}})
	}
	_ = client.writeJSON(map[string]any{"kind": "event", "type": "lyric", "data": map[string]any{"source": snapshot.Lyrics.Source, "lyric": snapshot.Lyrics.Lyric}})
	_ = client.writeJSON(map[string]any{"kind": "event", "type": "status", "data": snapshot.Status})

	for {
		var request struct {
			Op         string   `json:"op"`
			PositionMs *float64 `json:"positionMs"`
			Volume     *float64 `json:"volume"`
		}
		if err := conn.ReadJSON(&request); err != nil {
			return
		}
		if err := s.dispatchWebSocketCommand(request.Op, request.PositionMs, request.Volume); err != nil {
			_ = client.writeJSON(map[string]any{"kind": "error", "op": request.Op, "error": err.Error()})
			continue
		}
		_ = client.writeJSON(map[string]any{"kind": "ack", "op": request.Op})
	}
}

func (s *Server) dispatchWebSocketCommand(op string, positionMs, volume *float64) error {
	switch op {
	case "play":
		s.provider.ExternalAPIPlay()
	case "pause":
		s.provider.ExternalAPIPause()
	case "stop":
		s.provider.ExternalAPIStop()
	case "next":
		s.provider.ExternalAPINext()
	case "prev":
		s.provider.ExternalAPIPrevious()
	case "seek":
		if !validNonNegative(positionMs) {
			return errors.New("positionMs (number, >=0) required")
		}
		s.provider.ExternalAPISeek(int64(*positionMs))
	case "setVolume":
		if volume == nil || math.IsNaN(*volume) || math.IsInf(*volume, 0) || *volume < 0 || *volume > 1 {
			return errors.New("volume (number, 0..1) required")
		}
		s.provider.ExternalAPISetVolume(*volume)
	default:
		return fmt.Errorf("unknown op %q", op)
	}
	return nil
}

func allowGet(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return false
	}
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return false
	}
	return true
}

func allowControl(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return false
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return false
	}
	if !allowedOrigin(r.Header.Get("Origin")) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "origin not allowed"})
		return false
	}
	return true
}

func allowedOrigin(origin string) bool {
	if origin == "" || origin == "null" {
		return true
	}
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if strings.EqualFold(parsed.Scheme, "file") {
		return true
	}
	host := strings.Trim(parsed.Hostname(), "[]")
	return host == "127.0.0.1" || host == "::1" || host == "localhost"
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
