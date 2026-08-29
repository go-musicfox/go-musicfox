package headless

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

// Run starts the core engine in headless mode.
//
//   - When once is non-empty it runs a single control command (first
//     space-separated token = cmd, the rest = query used as args.query for
//     "play", empty args otherwise), prints the result as a compact JSON object
//     on stdout and exits. No control server is started in once mode.
//   - Otherwise it stays resident: the daemon is a cordis plugin (DaemonPlugin)
//     hosted by a dedicated frontend scope — its Deps resolves
//     ServiceDispatcher + ServiceEventBus from the engine context, its Start
//     serves the control channel (legacy one-shot commands + the P7
//     subscribe/unsubscribe event stream) and subscribes the event bus. Run
//     blocks until SIGINT/SIGTERM or a server-triggered shutdown (a "quit"
//     control command), then stops the daemon scope and shuts the engine down
//     cleanly.
func Run(once string) error {
	// No user slot and no desktop-lyrics controller: the engine owns its user
	// slot and leaves desktop lyrics nil (nil-safe in the player).
	engine := core.NewEngine(core.EngineOptions{})

	if err := engine.Startup(context.Background(), &HeadlessObserver{}); err != nil {
		// Defensive: Startup currently only returns non-nil on a truly fatal
		// failure (every degraded startup path logs and continues).
		slog.Error("headless startup failed", slogx.Error(err))
		_ = engine.Close()
		return err
	}

	if once != "" {
		return runOnce(engine, once)
	}

	// Stay resident until SIGINT/SIGTERM or a "quit" control command, then
	// stop the daemon scope and shut the engine down cleanly (scrobble report
	// + cookie jar save happen inside engine.Close).
	server := NewServer(engine)
	scope := framework.NewScope()
	if err := scope.Add(NewDaemonPlugin(server)); err != nil {
		// Unreachable for a fresh scope; close so Run exits.
		server.Close()
		_ = engine.Close()
		return err
	}
	if err := scope.Start(engine.Ctx()); err != nil {
		// Unreachable for the static plugin set (the engine always registers
		// ServiceDispatcher + ServiceEventBus); close so Run exits.
		slog.Error("headless daemon scope start failed", slogx.Error(err))
		server.Close()
		_ = engine.Close()
		return err
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case <-sigCh:
	case <-server.ShutdownCh():
	}

	server.Close()
	var errs []error
	errs = append(errs, scope.Dispose())
	errs = append(errs, engine.Close())
	return errors.Join(errs...)
}

// parseOnce splits a "--once" string into its command and query parts: the
// first space-separated token is the command, the remainder (trimmed) is the
// query. Both parts are empty for an empty input.
func parseOnce(s string) (cmd, query string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", ""
	}
	if idx := strings.Index(s, " "); idx >= 0 {
		return s[:idx], strings.TrimSpace(s[idx+1:])
	}
	return s, ""
}

// runOnce executes a single control command against the engine, prints the
// result as a compact JSON object and shuts the engine down. It returns nil on
// success and a non-nil error on failure (callers surface it as a CLI error).
func runOnce(engine *core.Engine, once string) error {
	cmd, query := parseOnce(once)
	var args map[string]any
	if cmd == "play" && query != "" {
		args = map[string]any{"query": query}
	}

	if cmd == "quit" {
		// Transport-layer shutdown semantics: the transport decides the quit
		// behavior. In --once mode "quit" mirrors the historical dispatcher
		// ErrQuit result: print {ok:false,"error":"quit"} and return an error.
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"ok": false, "error": "quit"})
		_ = engine.Close()
		return errors.New("quit")
	}

	data, err := core.NewDispatcher(engine).Dispatch(context.Background(), cmd, args)
	if err != nil {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"ok": false, "error": err.Error()})
		_ = engine.Close()
		return err
	}

	if err := json.NewEncoder(os.Stdout).Encode(map[string]any{"ok": true, "data": data}); err != nil {
		_ = engine.Close()
		return errors.Join(err, engine.Close())
	}
	return engine.Close()
}
