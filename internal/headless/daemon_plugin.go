// P7 headless daemon plugin (docs/plugin_ecosystem.md §3.4): the control
// server becomes a cordis plugin hosted by a dedicated frontend scope. The
// plugin owns the server lifecycle and the core event-bus subscription that
// streams playback/startup events to subscribed control connections.
//
// Concurrency contract (docs/plugin_ecosystem.md §四): emitter listeners run
// synchronously on the emitting player goroutine, so they are strictly
// enqueue-only — each listener writes a daemonEvent into a bounded channel and
// returns, never touching a socket. A dedicated drain goroutine consumes the
// queue, builds the frame and hands it to the server, which filters it by the
// per-connection subscription set and broadcasts through the shared
// frontend.EventSink (non-blocking per-connection goroutine writes).
package headless

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/frontend"
)

// daemonEventQueueSize bounds the queue between the emitter listeners and the
// drain goroutine. When full, the newest frame is dropped — the slow-subscriber
// philosophy from the WebUI (a slow consumer only loses its own frames, never
// the emitter callback).
const daemonEventQueueSize = 256

// daemonPositionMinInterval is the daemon-side position throttle. The engine
// already throttles position ticks; this second layer keeps a stable event
// rate regardless of the UI frame rate (same value as the WebUI).
const daemonPositionMinInterval = 250 * time.Millisecond

// daemonWireEvents is the full core event-bus wire set the daemon forwards.
// Subscribers opt in per event name via "subscribe"; the wire name is echoed
// in the event frame's "event" field so a client can correlate 1:1.
var daemonWireEvents = []string{
	core.EvSongChanged,
	core.EvStateChanged,
	core.EvPosition,
	core.EvPlaylistEnd,
	core.EvRerender,
	core.EvLogin,
	core.EvStartupPhase,
}

// daemonEvent is one queued event-bus frame awaiting broadcast.
type daemonEvent struct {
	name    string
	payload any
}

// positionThrottle drops position events closer than daemonPositionMinInterval
// to the previous emit. Not safe for concurrent use — the event-bus listener
// runs on the emitting player goroutine (a single goroutine per loop).
type positionThrottle struct {
	lastAt time.Time
}

// shouldEmit reports whether now is far enough from the last emit, updating
// the state on a hit.
func (t *positionThrottle) shouldEmit(now time.Time) bool {
	if now.Sub(t.lastAt) >= daemonPositionMinInterval {
		t.lastAt = now
		return true
	}
	return false
}

// DaemonPlugin is the headless daemon as a cordis plugin. Deps resolves
// ServiceDispatcher and ServiceEventBus from the engine context; Start listens
// on the control channel and subscribes the event bus; Stop closes the server;
// Dispose unsubscribes the emitter and stops the drain goroutine.
type DaemonPlugin struct {
	framework.PluginBase
	server *Server

	// Deps-resolved services.
	dispatcher *core.Dispatcher
	emitter    *framework.EventEmitter

	eventCh      chan daemonEvent
	stopped      chan struct{}
	stopOnce     sync.Once
	unsub        func()
	serverCancel context.CancelFunc
}

// NewDaemonPlugin builds the daemon plugin around an already-created control
// Server. The server's default dispatcher is replaced by ServiceDispatcher in
// Deps/Start (the framework-resolvable mount point).
func NewDaemonPlugin(server *Server) *DaemonPlugin {
	return &DaemonPlugin{
		server:  server,
		stopped: make(chan struct{}),
	}
}

// Deps resolves the daemon's service dependencies from the engine context.
func (p *DaemonPlugin) Deps(ctx *framework.Context) error {
	d, ok := framework.ServiceOf[*core.Dispatcher](ctx, core.ServiceDispatcher)
	if !ok {
		return errors.New("daemonPlugin: ServiceDispatcher not resolved")
	}
	e, ok := framework.ServiceOf[*framework.EventEmitter](ctx, core.ServiceEventBus)
	if !ok {
		return errors.New("daemonPlugin: ServiceEventBus not resolved")
	}
	p.dispatcher = d
	p.emitter = e
	return nil
}

// Start wires the daemon: it points the server transport at the resolved
// ServiceDispatcher, starts the accept loop, registers the event-bus listeners
// and launches the drain goroutine.
func (p *DaemonPlugin) Start(*framework.Context) error {
	p.server.dispatcher = p.dispatcher

	serverCtx, cancel := context.WithCancel(context.Background())
	p.serverCancel = cancel
	go func() {
		if err := p.server.Serve(serverCtx); err != nil {
			// A listen/accept failure is fatal for the resident daemon: log it
			// and close so headless.Run's ShutdownCh wait breaks.
			slog.Error("headless daemon: control server failed", slog.Any("err", err))
			p.server.Close()
		}
	}()

	p.eventCh = make(chan daemonEvent, daemonEventQueueSize)
	p.unsub = p.subscribeEvents()
	go p.drain()
	return nil
}

// subscribeEvents registers the enqueue-only listeners that forward the core
// event bus to the drain queue. The returned func unsubscribes the whole set
// on teardown (per-wire coarse EventEmitter.Unregister — the daemon owns its
// subscription for its lifetime).
func (p *DaemonPlugin) subscribeEvents() func() {
	posThrottle := new(positionThrottle)
	for _, name := range daemonWireEvents {
		name := name
		p.emitter.Listener(name, func(_ *framework.Context, payload any) error {
			if name == core.EvPosition && !posThrottle.shouldEmit(time.Now()) {
				return nil
			}
			p.enqueue(name, payload)
			return nil
		})
	}
	return func() {
		for _, name := range daemonWireEvents {
			p.emitter.Unregister(name)
		}
	}
}

// enqueue delivers one event to the drain queue. It is the only thing an
// emitter listener may do: it never blocks (bounded buffer, drop on full) and
// never touches a socket.
func (p *DaemonPlugin) enqueue(name string, payload any) {
	select {
	case p.eventCh <- daemonEvent{name: name, payload: payload}:
	default:
		slog.Debug("headless daemon: event queue full, dropping frame", slog.String("event", name))
	}
}

// drain consumes the event queue and delivers filtered frames through the
// server's broadcaster. It exits when the plugin is disposed (stopped closed).
func (p *DaemonPlugin) drain() {
	for {
		select {
		case <-p.stopped:
			return
		case ev := <-p.eventCh:
			frame := frontend.EventFrame(ev.name, ev.payload)
			if frame == nil {
				continue
			}
			p.server.broadcastEvent(ev.name, frame)
		}
	}
}

// Stop closes the control server (listener + subscription connections + sink).
func (p *DaemonPlugin) Stop() error {
	if p.serverCancel != nil {
		p.serverCancel()
		p.serverCancel = nil
	}
	p.server.Close()
	return nil
}

// Dispose unsubscribes the event-bus listeners and stops the drain goroutine.
// Closing the subscription connections is covered by Stop (server.Close); this
// is belt-and-braces for a Dispose-without-Stop path.
func (p *DaemonPlugin) Dispose() error {
	if p.unsub != nil {
		p.unsub()
		p.unsub = nil
	}
	p.stopOnce.Do(func() { close(p.stopped) })
	p.server.Close()
	return nil
}
