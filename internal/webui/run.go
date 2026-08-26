package webui

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/skratchdot/open-golang/open"

	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

// Run starts the core engine with the WebUI frontend: it serves the static
// placeholder page on a loopback HTTP listener, opens the default browser at
// the page, and stays resident until SIGINT/SIGTERM or a server-triggered
// shutdown, then shuts the engine down cleanly.
func Run(ctx context.Context) error {
	// The engine owns its user slot and leaves desktop lyrics nil (nil-safe in
	// the player), same as the headless frontend.
	engine := core.NewEngine(core.EngineOptions{})

	// The WebUI observer is created with the server (NewServer also attaches it
	// to the engine player), so the server must be built before Startup can
	// broadcast the startup phases.
	server := NewServer(engine)

	if err := engine.Startup(ctx, server.observer); err != nil {
		// Defensive: Startup currently only returns non-nil on a truly fatal
		// failure (every degraded startup path logs and continues).
		slog.Error("webui startup failed", slogx.Error(err))
		_ = server.Close()
		_ = engine.Close()
		return err
	}
	serverCtx, cancelServer := context.WithCancel(ctx)
	defer cancelServer()
	go func() {
		if err := server.Serve(serverCtx); err != nil {
			// A listen/accept failure is fatal for the server: log it and
			// trigger shutdown so Run can exit.
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
		return engine.Close()
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

	server.Close()
	return engine.Close()
}
