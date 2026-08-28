package webui

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/coder/websocket"

	"github.com/go-musicfox/go-musicfox/internal/core"
)

// Server serves the WebUI frontend over HTTP on a loopback listener. It
// mounts the static page, the token exchange endpoint and the authenticated
// /ws control channel; the token is generated up front and shared by the
// exchange endpoint (session cookie) and the WS/API auth layers. The data
// source is the Backend abstraction (a local engine or, in connect mode, a
// remote headless daemon), never a *core.Engine directly.
type Server struct {
	backend  Backend
	auth     bool // opts.Auth: true enables the token/cookie/Origin auth layers
	listener net.Listener
	token    string // crypto/rand 32-byte hex, generated in NewServer (session cookie + WS/API auth)
	mux      *http.ServeMux
	quit     chan struct{}
	quitOnce sync.Once
	httpSrv  *http.Server

	// conns tracks live WebSocket connections so Close can terminate them all
	// and future event broadcasts can reach connected clients.
	conns      map[int64]*wsConn
	connsMu    sync.Mutex
	nextConnID int64

	// broadcaster fans engine events out to every live connection; unsubscribe
	// removes the core event-bus listeners registered at construction (Close).
	broadcaster *broadcaster
	unsubscribe func()

	// readyErr reports listener readiness to Run: nil once the loopback
	// listener is bound, or the listen error when binding fails.
	readyErr chan error
}

// NewServer builds a WebUI server bound to the engine (the standalone
// convenience wrapper). The loopback listener is bound in Serve (not here),
// so the browser can be pointed at the actual ephemeral port.
func NewServer(engine *core.Engine) *Server {
	return NewServerWithOptions(&localBackend{engine: engine}, ServerOptions{Auth: true})
}

// NewServerWithBackend builds a WebUI server over an arbitrary Backend with
// the auth layer enabled (current standalone/connect behavior).
func NewServerWithBackend(backend Backend) *Server {
	return NewServerWithOptions(backend, ServerOptions{Auth: true})
}

// NewServerWithOptions builds a WebUI server over an arbitrary Backend with
// the given options. Auth=true registers the token/cookie/Origin auth layers
// (current behavior); Auth=false mounts the raw handlers so a GUI AssetServer
// scheme (no cookie exchange possible) can serve the page directly.
func NewServerWithOptions(backend Backend, opts ServerOptions) *Server {
	mux := http.NewServeMux()
	b := newBroadcaster()
	s := &Server{
		backend:     backend,
		auth:        opts.Auth,
		token:       randomToken(),
		mux:         mux,
		quit:        make(chan struct{}),
		httpSrv:     &http.Server{Handler: mux},
		conns:       make(map[int64]*wsConn),
		broadcaster: b,
		readyErr:    make(chan error, 1),
	}
	if backend != nil {
		// The WebUI frontend consumes playback/startup events through the
		// backend's event subscription: register the forwarding listeners at
		// construction so the startup phases (emitted by engine.Startup, which
		// Run calls after NewServer) reach the broadcaster. Close calls the
		// returned unsubscribe to clean up.
		s.unsubscribe = backend.SubscribeEvents(func(_ string, payload []byte) {
			s.broadcaster.broadcast(payload)
		})
	}
	// go1.22+ method-pattern routing: "GET /" matches every GET path (prefix
	// semantics, HEAD included); non-GET requests get 405. The token exchange,
	// /ws and the /api/* endpoints are explicit so they win over the "/"
	// prefix. /ws is gated by verifyWSRequest before the upgrade Accept (T4;
	// skipped when Auth=false); /api/* handlers are wrapped with
	// s.authMiddleware (T6/T7, skipped when Auth=false). The static root and
	// the token exchange stay unauthenticated.
	wrap := func(h http.HandlerFunc) http.HandlerFunc {
		if opts.Auth {
			return s.authMiddleware(h)
		}
		return h
	}
	mux.HandleFunc("GET /", s.handleStatic)
	mux.HandleFunc("GET /token", s.handleTokenExchange)
	mux.HandleFunc("GET /ws", s.handleWS)
	// T7 auxiliary endpoints.
	mux.HandleFunc("GET /api/status", wrap(s.handleStatus))
	mux.HandleFunc("GET /api/albumart", wrap(s.handleAlbumArt))
	mux.HandleFunc("GET /api/lyrics", wrap(s.handleLyrics))
	// T6 QR login endpoints.
	mux.HandleFunc("GET /api/login/qr/key", wrap(s.handleLoginQRKey))
	mux.HandleFunc("GET /api/login/qr/image", wrap(s.handleLoginQRImage))
	mux.HandleFunc("GET /api/login/qr/status", wrap(s.handleLoginQRStatus))
	// T6 Track-B command endpoints (GET list + POST exec by key).
	mux.HandleFunc("GET /api/commands", wrap(s.handleCommandsList))
	mux.HandleFunc("POST /api/commands/{key}", wrap(s.handleCommandExec))
	return s
}

// ShutdownCh is closed when the server shuts down, so Run can break out of
// its signal wait.
func (s *Server) ShutdownCh() <-chan struct{} {
	return s.quit
}

// Addr returns the bound loopback address ("127.0.0.1:<port>"). It is valid
// after Serve has bound the listener (or equivalently after waitReady).
func (s *Server) Addr() string {
	return s.listener.Addr().String()
}

// Token returns the server's auth token, embedded in the browser URL by Run so
// the first visit can exchange it for the session cookie.
func (s *Server) Token() string {
	return s.token
}

// waitReady blocks until the loopback listener is bound and returns the bind
// error (if any). It exists because Serve runs in its own goroutine while Run
// needs the real address to open the browser.
func (s *Server) waitReady() error {
	return <-s.readyErr
}

// Serve binds the loopback listener (if not already bound) and serves HTTP
// until Close is called or ctx is cancelled. It reports readiness/failure on
// readyErr before blocking in http.Server.Serve.
func (s *Server) Serve(ctx context.Context) error {
	if s.listener == nil {
		if err := s.listen(); err != nil {
			s.readyErr <- err
			return err
		}
	}
	s.readyErr <- nil

	err := s.httpSrv.Serve(s.listener)
	if errors.Is(err, http.ErrServerClosed) || errors.Is(err, net.ErrClosed) {
		return nil
	}
	select {
	case <-s.quit:
		return nil // normal shutdown
	case <-ctx.Done():
		return nil
	default:
	}
	return err
}

// listen binds the TCP loopback listener on an ephemeral port.
func (s *Server) listen() error {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	s.listener = ln
	return nil
}

// addConn registers a live WebSocket connection and returns its registry id.
// It returns 0 when the server is already shutting down (the caller then
// closes the connection itself), so a connection racing Close can never leak.
func (s *Server) addConn(c *wsConn) int64 {
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

// handleStatic serves the embedded placeholder page via http.FileServer.
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	http.FileServer(http.FS(staticRoot)).ServeHTTP(w, r)
}

// Close shuts the server down gracefully. It is idempotent.
func (s *Server) Close() error {
	var closeErr error
	s.quitOnce.Do(func() {
		close(s.quit)
		// Unsubscribe the core event-bus listeners so no further engine event
		// targets this server's broadcaster after teardown.
		if s.unsubscribe != nil {
			s.unsubscribe()
		}
		// Terminate every live WebSocket connection first so their read loops
		// unblock and the connection goroutines wind down (they remove
		// themselves from conns as they exit).
		s.connsMu.Lock()
		for id, conn := range s.conns {
			delete(s.conns, id)
			_ = conn.c.Close(websocket.StatusGoingAway, "server shutting down")
		}
		s.connsMu.Unlock()
		// Drop every connection from the broadcaster so no further event
		// broadcast targets a terminated socket.
		s.broadcaster.reset()
		if s.httpSrv != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := s.httpSrv.Shutdown(ctx); err != nil {
				closeErr = err
				// Force-close so Serve returns and Run can exit promptly.
				_ = s.httpSrv.Close()
			}
		}
		if s.listener != nil {
			// Shutdown already closes the listener; this is a belt-and-braces
			// no-op returning net.ErrClosed on success.
			_ = s.listener.Close()
		}
	})
	return closeErr
}

// randomToken returns a 32-byte crypto/rand token hex-encoded (64 chars). It
// guards the future API/WS routes; T2 only generates it.
func randomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand never fails on supported platforms; fall back to a
		// pid+timestamp seed so the server still starts.
		return fmt.Sprintf("%d-%d", os.Getpid(), time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
