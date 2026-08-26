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
	"github.com/go-musicfox/go-musicfox/utils/app"
)

// ipcReadTimeout bounds how long a connection may take to send a single
// request. A stuck/half-open client must never hold a server goroutine forever.
const ipcReadTimeout = 5 * time.Second

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

// Server serves the headless control channel. It listens on a unix socket
// (non-Windows) or an ephemeral TCP loopback port (Windows), accepts one JSON
// Request per connection and replies with one JSON Response.
type Server struct {
	engine     *core.Engine
	network    string
	addr       string
	listener   net.Listener
	dispatcher *core.Dispatcher
	quit       chan struct{}
	quitOnce   sync.Once
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
	if s.listener == nil {
		if err := s.listen(); err != nil {
			return err
		}
	}

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

// handleConn reads one Request, dispatches it and writes one Response. The
// connection is always closed on exit; panics are recovered so a broken client
// can never kill the accept loop.
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

	if req.Cmd == "quit" {
		// Transport-layer shutdown semantics: the transport decides whether
		// "quit" shuts it down. The command is answered with {ok:true} first,
		// then the server shuts down (listener + socket/port file) so
		// headless.Run can exit.
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

// Close shuts the server down: it closes the listener (unblocking Accept) and
// removes the socket/port file. It is idempotent.
func (s *Server) Close() {
	s.quitOnce.Do(func() {
		close(s.quit)
		if s.listener != nil {
			_ = s.listener.Close()
		}
		if s.network == "unix" {
			_ = os.Remove(s.addr)
		} else {
			_ = os.Remove(filepath.Join(app.DataDir(), "musicfox.port"))
		}
	})
}
