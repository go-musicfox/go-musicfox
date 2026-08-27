package headless

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/frontend"
	"github.com/go-musicfox/go-musicfox/utils/app"
)

// ipcReadTimeout bounds how long a connection may take to send a single
// request. A stuck/half-open client must never hold a server goroutine forever.
// It only applies to the one-shot legacy path and the first (opting-in) request
// of a subscription session; once a connection subscribes it is long-lived and
// the read deadline is cleared.
const ipcReadTimeout = 5 * time.Second

// daemonWriteTimeout bounds a single frame write on a subscription connection
// (responses and event broadcasts alike). A stuck client must never pin its
// write goroutine forever; the frame is dropped for that connection alone.
const daemonWriteTimeout = 5 * time.Second

// snapshotTimeout bounds the snapshot build: it dispatches a "status" command
// through the shared dispatcher mutex and reads the playlist.
const snapshotTimeout = 5 * time.Second

// ErrQuit is the transport-layer shutdown sentinel: it marks the "quit"
// command as a graceful shutdown request rather than a failure. The transport
// layer decides the quit behavior (the core Dispatcher never shuts anything
// down itself).
var ErrQuit = errors.New("quit")

// ListenAddr returns the control-channel socket path used by the headless
// daemon on non-Windows platforms. On Windows the channel is a TCP listener on
// an ephemeral port persisted to a port file, so the empty string is returned
// here and serverAddr/DialAddr take over.
func ListenAddr() string {
	if runtime.GOOS == "windows" {
		return ""
	}
	return filepath.Join(app.DataDir(), "musicfox.sock")
}

// DialAddr returns the (network, address) pair used to connect to a running
// headless daemon: a unix socket on non-Windows, and the persisted ephemeral
// TCP port on Windows. It returns empty strings when no daemon is running.
func DialAddr() (network, addr string) {
	if runtime.GOOS == "windows" {
		portFile := filepath.Join(app.DataDir(), "musicfox.port")
		data, err := os.ReadFile(portFile)
		if err != nil {
			return "", ""
		}
		return "tcp", "127.0.0.1:" + strings.TrimSpace(string(data))
	}
	return "unix", ListenAddr()
}

// serverListenAddr returns the default listen (network, address) pair: the
// unix socket on non-Windows, an ephemeral TCP loopback port on Windows.
func serverListenAddr() (network, addr string) {
	if runtime.GOOS == "windows" {
		return "tcp", "127.0.0.1:0"
	}
	return "unix", ListenAddr()
}

// daemonConn is one long-lived control connection. writeMu serializes response
// writes and event broadcasts on the same connection so frames can never
// interleave; Write also owns the transport framing (a trailing newline so the
// streaming client can delimit JSON frames).
type daemonConn struct {
	conn    net.Conn
	writeMu sync.Mutex
	id      int64
}

// Write implements frontend.EventSinkConn. It appends a newline terminator
// (the headless wire protocol is newline-delimited JSON) and bounds the write
// by daemonWriteTimeout.
func (c *daemonConn) Write(payload []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	_ = c.conn.SetWriteDeadline(time.Now().Add(daemonWriteTimeout))
	payload = append(payload, '\n')
	_, err := c.conn.Write(payload)
	return err
}

// Server serves the headless control channel. It listens on a unix socket
// (non-Windows) or an ephemeral TCP loopback port (Windows). Each connection
// is dual-capability (P7, docs/plugin_ecosystem.md §3.4):
//
//   - Legacy one-shot path: the first request is not "subscribe", so the
//     connection is treated exactly as before — one JSON Request in, one JSON
//     Response out, then close. musicfox ctrl (internal/headless/client.go)
//     opens a fresh connection per Call and is unaffected.
//   - Subscription path: the first request is "subscribe", so the connection
//     becomes a long-lived session: it receives a snapshot frame, then a stream
//     of event frames for its subscribed event names (filtered by the
//     per-connection subscription set) interleaved with request/response
//     command handling.
type Server struct {
	engine     *core.Engine
	network    string
	addr       string
	listenerMu sync.Mutex // guards listener between Serve (listen) and Close
	listener   net.Listener
	dispatcher *core.Dispatcher
	quit       chan struct{}
	quitOnce   sync.Once

	// sink fans event frames out to every subscribed connection (the shared
	// frontend.EventSink, same mechanism as the WebUI broadcaster).
	sink *frontend.EventSink

	// subs is the daemon's subscription state: conn id → set of subscribed
	// wire event names. It is data-plane state and stays out of the
	// Scope/Context (the scope manages lifecycle, subscriptions manage the
	// data plane — orthogonal, docs/plugin_ecosystem.md §3.4).
	subs   map[int64]map[string]bool
	subsMu sync.Mutex

	// conns tracks live subscription connections so Close can terminate them all.
	conns      map[int64]*daemonConn
	connsMu    sync.Mutex
	nextConnID int64
}

// NewServer builds a server on the default listen address.
func NewServer(engine *core.Engine) *Server {
	network, addr := serverListenAddr()
	return NewServerWithAddr(engine, network, addr)
}

// NewServerWithAddr builds a server on an explicit (network, address) pair.
// Both Run and the integration tests share this constructor so tests can point
// the server at a temp socket path.
func NewServerWithAddr(engine *core.Engine, network, addr string) *Server {
	return &Server{
		engine:     engine,
		network:    network,
		addr:       addr,
		dispatcher: core.NewDispatcher(engine),
		quit:       make(chan struct{}),
		sink:       frontend.NewEventSink(),
		subs:       make(map[int64]map[string]bool),
		conns:      make(map[int64]*daemonConn),
	}
}

// ShutdownCh is closed when the server shuts down (e.g. after a "quit"
// command), so headless.Run can break out of its signal wait.
func (s *Server) ShutdownCh() <-chan struct{} {
	return s.quit
}

// Serve listens (if not already listening) and accepts control connections
// until Close is called or ctx is cancelled.
func (s *Server) Serve(ctx context.Context) error {
	s.listenerMu.Lock()
	if s.listener == nil {
		if err := s.listen(); err != nil {
			s.listenerMu.Unlock()
			return err
		}
	}
	s.listenerMu.Unlock()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return nil // normal shutdown
			case <-ctx.Done():
				return nil
			default:
			}
			return err
		}
		go s.handleConn(ctx, conn)
	}
}

// listen binds the configured address. On unix it probes a pre-existing socket
// file: if nothing is listening (a previous daemon crashed), the stale socket
// is removed before binding; if a live daemon is listening, listening fails.
// On Windows (tcp) the chosen ephemeral port is persisted to a port file.
// It must be called with listenerMu held (Serve does so), so the listener
// assignment stays synchronized with Close.
func (s *Server) listen() error {
	if s.network == "unix" {
		if _, err := os.Stat(s.addr); err == nil {
			probe, derr := net.DialTimeout("unix", s.addr, time.Second)
			if derr != nil {
				// Stale socket from a crashed daemon: remove and rebind.
				_ = os.Remove(s.addr)
			} else {
				_ = probe.Close()
				return fmt.Errorf("headless daemon already running at %s", s.addr)
			}
		}
	}

	ln, err := net.Listen(s.network, s.addr)
	if err != nil {
		return err
	}
	if s.network == "unix" {
		// Only the owning user may connect to the control channel.
		if err := os.Chmod(s.addr, 0600); err != nil {
			_ = ln.Close()
			return err
		}
	}
	s.listener = ln

	if s.network == "tcp" {
		if tcpAddr, ok := ln.Addr().(*net.TCPAddr); ok {
			portFile := filepath.Join(app.DataDir(), "musicfox.port")
			_ = os.WriteFile(portFile, []byte(strconv.Itoa(tcpAddr.Port)), 0600)
		}
	}
	return nil
}

// handleConn reads the first Request and dispatches it. A connection whose
// FIRST request is "subscribe" opts into the long-lived subscription session
// (snapshot + event stream + request/response); anything else keeps the legacy
// one-shot behavior. The connection is always closed on exit; panics are
// recovered so a broken client can never kill the accept loop.
func (s *Server) handleConn(ctx context.Context, conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("headless IPC connection panic", slog.Any("panic", r))
		}
	}()
	defer func() { _ = conn.Close() }()

	_ = conn.SetReadDeadline(time.Now().Add(ipcReadTimeout))
	var req core.Request
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		slog.Debug("headless IPC: read request failed", slog.Any("err", err))
		return
	}

	if req.Cmd == "subscribe" {
		s.handleSubscription(ctx, conn, req)
		return
	}
	s.handleOneShot(ctx, conn, req)
}

// handleOneShot is the legacy path: dispatch one request, write one response
// and close. "quit" is answered first, then shuts the server down
// (transport-layer shutdown semantics).
func (s *Server) handleOneShot(ctx context.Context, conn net.Conn, req core.Request) {
	if req.Cmd == "quit" {
		if err := json.NewEncoder(conn).Encode(core.Response{V: core.ProtocolVersion, ID: req.ID, Ok: true}); err != nil {
			slog.Debug("headless IPC: write response failed", slog.Any("err", err))
		}
		s.Close()
		return
	}

	data, err := s.dispatcher.Dispatch(ctx, req.Cmd, req.Args)

	resp := core.Response{V: core.ProtocolVersion, ID: req.ID, Ok: err == nil, Data: data}
	if err != nil {
		resp.Error = err.Error()
	}
	if err := json.NewEncoder(conn).Encode(resp); err != nil {
		slog.Debug("headless IPC: write response failed", slog.Any("err", err))
	}
}

// handleSubscription runs the long-lived session for a connection that opted
// in with a "subscribe" first request: register the connection, send the
// snapshot frame first, ack the subscribe, then serve requests (subscribe /
// unsubscribe / dispatcher commands) while event frames stream in. "quit"
// keeps transport-layer shutdown semantics and terminates the whole server.
func (s *Server) handleSubscription(ctx context.Context, conn net.Conn, req core.Request) {
	dc := &daemonConn{conn: conn}
	id := s.addConn(dc)
	if id == 0 {
		return // server is shutting down; handleConn's deferred Close cleans up
	}
	defer s.removeConn(id)

	events := parseSubEvents(req.Args)
	if len(events) == 0 {
		_ = s.writeResponse(dc, core.Response{
			V:     core.ProtocolVersion,
			ID:    req.ID,
			Ok:    false,
			Error: "subscribe 需要非空 events 数组 (usage: subscribe [\"player.song_changed\", ...])",
		})
		return
	}

	// Snapshot first, and BEFORE the connection is registered in the event
	// sink: the first frame a subscriber receives is always the snapshot (an
	// event racing the registration could otherwise beat the snapshot write on
	// the connection lock and break the snapshot-first contract). Events fired
	// during the build are reflected in the snapshot; those fired after
	// registration are delivered in order.
	if payload, err := s.buildSnapshot(ctx); err != nil {
		slog.Debug("headless IPC: snapshot build failed", slog.Any("err", err))
	} else if err := dc.Write(payload); err != nil {
		slog.Debug("headless IPC: snapshot write failed", slog.Any("err", err))
		return
	}

	s.sink.Add(id, dc)
	defer s.sink.Remove(id)
	s.setSubscription(id, events)
	defer s.clearSubscription(id)

	if err := s.writeResponse(dc, core.Response{
		V:    core.ProtocolVersion,
		ID:   req.ID,
		Ok:   true,
		Data: map[string]any{"events": events},
	}); err != nil {
		slog.Debug("headless IPC: subscribe ack write failed", slog.Any("err", err))
		return
	}

	// Long-lived session: clear the one-shot read deadline (an idle subscriber
	// stays connected; Close terminates it on shutdown).
	_ = conn.SetReadDeadline(time.Time{})

	for {
		var r core.Request
		if err := json.NewDecoder(conn).Decode(&r); err != nil {
			return // client disconnected / connection closed by shutdown
		}
		switch r.Cmd {
		case "unsubscribe":
			// Specific events when given, the whole subscription otherwise.
			if evs := parseSubEvents(r.Args); len(evs) > 0 {
				s.unsetSubscription(id, evs)
			} else {
				s.clearSubscription(id)
			}
			if err := s.writeResponse(dc, core.Response{V: core.ProtocolVersion, ID: r.ID, Ok: true}); err != nil {
				slog.Debug("headless IPC: unsubscribe ack write failed", slog.Any("err", err))
				return
			}
		case "quit":
			// Transport-layer shutdown semantics: answer {ok:true} first, then
			// shut the server down.
			if err := s.writeResponse(dc, core.Response{V: core.ProtocolVersion, ID: r.ID, Ok: true}); err != nil {
				slog.Debug("headless IPC: quit response write failed", slog.Any("err", err))
			}
			s.Close()
			return
		default:
			data, err := s.dispatcher.Dispatch(ctx, r.Cmd, r.Args)
			resp := core.Response{V: core.ProtocolVersion, ID: r.ID, Ok: err == nil, Data: data}
			if err != nil {
				resp.Error = err.Error()
			}
			if err := s.writeResponse(dc, resp); err != nil {
				slog.Debug("headless IPC: response write failed", slog.Any("err", err))
				return
			}
		}
	}
}

// writeResponse marshals resp and writes it as a newline-delimited frame under
// the connection's write lock (so it can never interleave with event frames).
func (s *Server) writeResponse(dc *daemonConn, resp core.Response) error {
	payload, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return dc.Write(payload)
}

// parseSubEvents extracts the args.events string array (JSON-decoded as
// []any). A missing or empty array yields nil.
func parseSubEvents(args map[string]any) []string {
	raw, _ := args["events"].([]any)
	if len(raw) == 0 {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}

// addConn registers a live subscription connection and returns its id. It
// returns 0 when the server is already shutting down (the caller then leaves
// cleanup to handleConn's deferred Close), so a connection racing Close can
// never leak.
func (s *Server) addConn(c *daemonConn) int64 {
	s.connsMu.Lock()
	defer s.connsMu.Unlock()
	select {
	case <-s.quit:
		return 0
	default:
	}
	s.nextConnID++
	c.id = s.nextConnID
	s.conns[c.id] = c
	return c.id
}

// removeConn drops a connection from the registry. It is idempotent: Close
// clears the map wholesale, so a connection goroutine winding down after a
// server shutdown removes an already-missing entry.
func (s *Server) removeConn(id int64) {
	s.connsMu.Lock()
	defer s.connsMu.Unlock()
	delete(s.conns, id)
}

// setSubscription records the events a connection subscribes to.
func (s *Server) setSubscription(id int64, events []string) {
	s.subsMu.Lock()
	defer s.subsMu.Unlock()
	m := s.subs[id]
	if m == nil {
		m = make(map[string]bool)
		s.subs[id] = m
	}
	for _, e := range events {
		m[e] = true
	}
}

// unsetSubscription removes specific events from a connection's subscription.
func (s *Server) unsetSubscription(id int64, events []string) {
	s.subsMu.Lock()
	defer s.subsMu.Unlock()
	m := s.subs[id]
	if m == nil {
		return
	}
	for _, e := range events {
		delete(m, e)
	}
	if len(m) == 0 {
		delete(s.subs, id)
	}
}

// clearSubscription removes a connection's whole subscription (idempotent).
func (s *Server) clearSubscription(id int64) {
	s.subsMu.Lock()
	defer s.subsMu.Unlock()
	delete(s.subs, id)
}

// broadcastEvent delivers an event frame to every connection subscribed to the
// event name. The subscription set is snapshotted under subsMu first so the
// sink's keep predicate never touches subsMu (no nested locking).
func (s *Server) broadcastEvent(name string, payload []byte) {
	s.subsMu.Lock()
	subs := make(map[int64]bool, len(s.subs))
	for id, events := range s.subs {
		if events[name] {
			subs[id] = true
		}
	}
	s.subsMu.Unlock()
	s.sink.BroadcastFiltered(payload, func(id int64) bool { return subs[id] })
}

// buildSnapshot builds the initial state frame for a fresh subscription
// connection: the Dispatcher "status" result merged with the playlist
// (trimmed to id/name/artist/album):
// {"type":"snapshot","data":{...status..., "playlist":[...]}}.
func (s *Server) buildSnapshot(ctx context.Context) ([]byte, error) {
	if s.engine == nil {
		return nil, errors.New("no engine")
	}
	sctx, cancel := context.WithTimeout(ctx, snapshotTimeout)
	defer cancel()
	status, err := s.dispatcher.Dispatch(sctx, "status", nil)
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

// Close shuts the server down: it closes the listener (unblocking Accept),
// terminates every live subscription connection (unblocking their read loops),
// drops all subscriptions and event-sink registrations, and removes the
// socket/port file. It is idempotent.
func (s *Server) Close() {
	s.quitOnce.Do(func() {
		close(s.quit)
		s.listenerMu.Lock()
		if s.listener != nil {
			_ = s.listener.Close()
		}
		s.listenerMu.Unlock()
		// Terminate live subscription connections first so their read loops
		// unblock and the connection goroutines wind down.
		s.connsMu.Lock()
		for id, c := range s.conns {
			delete(s.conns, id)
			_ = c.conn.Close()
		}
		s.connsMu.Unlock()
		s.subsMu.Lock()
		s.subs = make(map[int64]map[string]bool)
		s.subsMu.Unlock()
		// Drop every connection from the sink so no further event broadcast
		// targets a terminated socket.
		s.sink.Reset()
		if s.network == "unix" {
			_ = os.Remove(s.addr)
		} else {
			_ = os.Remove(filepath.Join(app.DataDir(), "musicfox.port"))
		}
	})
}
