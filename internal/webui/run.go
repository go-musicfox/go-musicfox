package webui

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/skratchdot/open-golang/open"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/frontend"
	"github.com/go-musicfox/go-musicfox/internal/wasm"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

// RunWithOptions runs the WebUI frontend, dispatching on the launch mode:
// standalone builds its own engine (current behavior); connect dials the local
// headless daemon and serves the same surface against the daemon (connect.go).
func RunWithOptions(ctx context.Context, opts frontend.LaunchOptions) error {
	if opts.Mode == frontend.ModeConnect {
		return connectRun(ctx)
	}
	return runStandalone(ctx, opts)
}

// runStandalone starts the core engine with the WebUI frontend: it serves the
// static placeholder page on a loopback HTTP listener, opens the default
// browser at the page, and stays resident until SIGINT/SIGTERM or a
// server-triggered shutdown, then shuts the engine down cleanly.
func runStandalone(ctx context.Context, _ frontend.LaunchOptions) error {
	// The engine owns its user slot and leaves desktop lyrics nil (nil-safe in
	// the player), same as the headless frontend.
	engine := core.NewEngine(core.EngineOptions{})

	// The event-bus subscription is registered when the server is built
	// (NewServer wires the core emitter to the broadcaster), so the server must
	// be built before Startup can broadcast the startup phases. Startup takes a
	// no-op observer — the WebUI consumes events through the emitter, not the
	// observer seam.
	server := NewServer(engine)

	// Load WASM plugin commands into the frontend registry through the P6 scope
	// pipeline: a dedicated wasm scope owns the app-wide manager
	// (wasm.ManagerPlugin) and one wasmPlugin adapter per loaded plugin
	// directory (wasm.LoadIntoScope). The scope lifecycle replaces the former
	// LoadAndRegister + defer mgr.Close — Stop unregisters the commands and
	// Dispose closes the instances + manager on exit. A missing dir is not an
	// error; per-plugin load or registration failures are collected and only
	// logged, so a bad plugin never blocks startup. Headless does not load WASM
	// plugins — its control protocol has no command consumer.
	wasmScope := framework.NewScope()
	if err := wasmScope.Add(&wasm.ManagerPlugin{}); err != nil {
		slog.Error("webui: wasm manager plugin registration failed", slogx.Error(err))
		_ = server.Close()
		_ = engine.Close()
		return err
	}
	wasmDir := configs.AppConfig.Plugins.WasmDir
	if wasmDir == "" {
		wasmDir = filepath.Join(app.ConfigDir(), "wasm-plugins")
	}
	if _, errs := wasm.LoadIntoScope(ctx, engine.Ctx(), wasmScope, wasmDir, webuiWasmSink{}); len(errs) != 0 {
		for _, err := range errs {
			slog.Warn("webui: wasm plugin load issue", slogx.Error(err))
		}
	}
	if err := wasmScope.Start(engine.Ctx()); err != nil {
		slog.Error("webui: wasm scope start failed", slogx.Error(err))
		_ = wasmScope.Dispose()
		_ = server.Close()
		_ = engine.Close()
		return err
	}
	defer wasmScope.Dispose()

	if err := engine.Startup(ctx, webuiNoopObserver{}); err != nil {
		// Defensive: Startup currently only returns non-nil on a truly fatal
		// failure (every degraded startup path logs and continues).
		slog.Error("webui startup failed", slogx.Error(err))
		_ = server.Close()
		_ = engine.Close()
		return err
	}

	if err := runServer(ctx, server); err != nil {
		_ = engine.Close()
		return err
	}
	return engine.Close()
}

// runServer serves an already-built WebUI server: it starts Serve, waits for
// the loopback listener, points the default browser at the token exchange URL
// and blocks until SIGINT/SIGTERM, a server-triggered shutdown (WS "quit") or
// ctx cancellation, then closes the server. It is shared by the standalone and
// connect modes.
func runServer(ctx context.Context, server *Server) error {
	serverCtx, cancelServer := context.WithCancel(ctx)
	defer cancelServer()
	go func() {
		if err := server.Serve(serverCtx); err != nil {
			// A listen/accept failure is fatal for the server: log it and
			// trigger shutdown so the wait below breaks.
			slog.Error("webui server failed", slogx.Error(err))
			server.Close()
		}
	}()

	// Wait for the loopback listener to be bound, then point the default
	// browser at the placeholder page (best effort: a missing browser must not
	// break the frontend).
	if err := server.waitReady(); err != nil {
		slog.Error("webui listen failed", slogx.Error(err))
		server.Close()
		return err
	}

	// Point the browser at the token exchange endpoint first: it validates the
	// URL token, sets the session cookie, and redirects to "/" so the page and
	// its /ws + /api requests carry the authenticated cookie.
	url := "http://" + server.Addr() + "/token?token=" + server.Token()
	if err := open.Start(url); err != nil {
		slog.Warn("webui: failed to open browser, open the URL manually", slogx.Error(err))
	}
	slog.Info("webui listening", slog.String("url", url))
	fmt.Println("musicfox WebUI is running at " + url)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case <-sigCh:
	case <-server.ShutdownCh():
	case <-ctx.Done():
	}

	return server.Close()
}
