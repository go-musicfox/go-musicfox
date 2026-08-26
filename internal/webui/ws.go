package webui

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/go-musicfox/go-musicfox/internal/core"
)

// wsConn is one accepted WebSocket connection. writeMu serializes response
// writes and future event broadcasts on the same connection so frames can
// never interleave.
type wsConn struct {
	c       *websocket.Conn
	writeMu sync.Mutex
	id      int64
}

// handleWS is the "GET /ws" handler: it applies the same cookie + host +
// origin validation as the API routes before upgrading, then runs the read
// loop that dispatches core.Requests to the shared Dispatcher. "quit" is
// intercepted here (transport-layer shutdown semantics) and never reaches the
// Dispatcher, matching the headless control channel. Only Dispatcher commands
// are exposed — there is no exec or other out-of-band surface.
func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	defer func() {
		// A panicking connection goroutine must never kill the HTTP server:
		// recover and log (mirrors internal/headless/server.go).
		if rec := recover(); rec != nil {
			slog.Error("webui ws: connection panic", slog.Any("panic", rec))
		}
	}()

	if err := s.verifyWSRequest(r); err != nil {
		// Reject before upgrading; the failed handshake surfaces as a plain
		// 401 response to the client.
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// Defense in depth: Accept re-checks the Origin even though
		// verifyWSRequest already ran.
		OriginPatterns: []string{"http://127.0.0.1:*", "http://localhost:*"},
	})
	if err != nil {
		slog.Debug("webui ws: accept failed", slog.Any("err", err))
		return
	}

	conn := &wsConn{c: c}
	id := s.addConn(conn)
	if id == 0 {
		// The server is already shutting down: refuse the connection.
		_ = c.Close(websocket.StatusGoingAway, "server shutting down")
		return
	}
	// Register in the broadcaster BEFORE sending the snapshot: a connection
	// that joins mid-playback must not miss the current state. Duplicates are
	// idempotent for the frontend.
	s.broadcaster.add(conn)
	defer s.broadcaster.remove(id)
	defer s.removeConn(id)
	defer func() { _ = c.Close(websocket.StatusNormalClosure, "") }()

	// Snapshot first, then serve requests: the client always sees the current
	// state before any command responses.
	s.writeSnapshot(conn)
	s.serveWS(conn)
}

// snapshotTimeout bounds the snapshot build: it dispatches a "status" command
// through the shared dispatcher mutex and reads the playlist.
const snapshotTimeout = 5 * time.Second

// writeSnapshot builds and sends the initial state frame to a fresh connection.
func (s *Server) writeSnapshot(conn *wsConn) {
	ctx, cancel := context.WithTimeout(context.Background(), snapshotTimeout)
	defer cancel()
	payload, err := s.buildSnapshot(ctx)
	if err != nil {
		slog.Debug("webui ws: snapshot build failed", slog.Any("err", err))
		return
	}
	conn.writeMu.Lock()
	defer conn.writeMu.Unlock()
	if err := conn.c.Write(ctx, websocket.MessageText, payload); err != nil {
		slog.Debug("webui ws: snapshot write failed", slog.Any("err", err))
	}
}

// buildSnapshot merges the Dispatcher "status" result with the current
// playlist (trimmed to id/name/artist/album) into the snapshot frame:
// {"type":"snapshot","data":{...status..., "playlist":[...]}}.
func (s *Server) buildSnapshot(ctx context.Context) ([]byte, error) {
	if s.engine == nil {
		return nil, errors.New("no engine")
	}
	status, err := s.dispatcher.Dispatch(ctx, "status", nil)
	if err != nil {
		return nil, err
	}
	data := map[string]any{}
	if m, ok := status.(map[string]any); ok {
		for k, v := range m {
			data[k] = v
		}
	}
	playlist := make([]map[string]any, 0)
	for _, song := range s.engine.Player().Playlist() {
		playlist = append(playlist, map[string]any{
			"id":     song.Id,
			"name":   song.Name,
			"artist": song.ArtistName(),
			"album":  song.Album.Name,
		})
	}
	data["playlist"] = playlist
	return json.Marshal(map[string]any{"type": "snapshot", "data": data})
}

// serveWS runs the read/dispatch/write loop for one connection until the
// client disconnects, the server shuts down, or a panic is recovered.
func (s *Server) serveWS(conn *wsConn) {
	ctx := context.Background()
	for {
		var req core.Request
		if err := wsjson.Read(ctx, conn.c, &req); err != nil {
			slog.Debug("webui ws: read failed", slog.Any("err", err))
			return
		}

		if req.Cmd == "quit" {
			// Transport-layer shutdown semantics: answer {ok:true} first, then
			// shut the server down. Close closes ShutdownCh, so Run breaks out
			// of its signal wait and cleans up the engine.
			if err := s.writeResponse(conn, ctx, core.Response{V: core.ProtocolVersion, ID: req.ID, Ok: true}); err != nil {
				slog.Debug("webui ws: quit response write failed", slog.Any("err", err))
			}
			s.Close()
			return
		}

		data, err := s.dispatcher.Dispatch(ctx, req.Cmd, req.Args)
		resp := core.Response{V: core.ProtocolVersion, ID: req.ID, Ok: err == nil, Data: data}
		if err != nil {
			resp.Error = err.Error()
		}
		if err := s.writeResponse(conn, ctx, resp); err != nil {
			slog.Debug("webui ws: response write failed", slog.Any("err", err))
			return
		}
	}
}

// writeResponse sends a Response under the connection's write lock so
// concurrent writers (responses + future event broadcasts) never interleave.
func (s *Server) writeResponse(conn *wsConn, ctx context.Context, resp core.Response) error {
	conn.writeMu.Lock()
	defer conn.writeMu.Unlock()
	return wsjson.Write(ctx, conn.c, &resp)
}
