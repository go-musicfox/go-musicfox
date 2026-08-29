package headless

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// sockCounter and shortSockPath produce short unique unix socket paths:
// macOS's sockaddr_un.sun_path is only 104 bytes, so t.TempDir() paths (which
// nest under /var/folders/...) are far too long for net.Listen/Dial.
var sockCounter int

func shortSockPath(t *testing.T) string {
	t.Helper()
	sockCounter++
	return filepath.Join(os.TempDir(), fmt.Sprintf("mfox-test-%d-%d.sock", os.Getpid(), sockCounter))
}

// waitForSocket polls until the server is accepting connections on sock.
func waitForSocket(t *testing.T, sock string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		conn, err := net.Dial("unix", sock)
		if err == nil {
			_ = conn.Close()
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("server did not start listening on %s: %v", sock, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// testCtrlClient builds a CtrlClient pointed at the test server socket.
func testCtrlClient(t *testing.T, sock string) *CtrlClient {
	t.Helper()
	return &CtrlClient{network: "unix", addr: sock}
}

// TestServerClientIntegration exercises the full JSON-lines control channel
// over a real unix socket: status ok, unknown-command failure, and a quit that
// shuts the server down.
func TestServerClientIntegration(t *testing.T) {
	engine := newTestEngine(t)
	sock := shortSockPath(t)

	server := NewServerWithAddr(engine, "unix", sock)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	serveDone := make(chan error, 1)
	go func() { serveDone <- server.Serve(ctx) }()
	t.Cleanup(server.Close)
	waitForSocket(t, sock)

	client := testCtrlClient(t, sock)

	// status → ok:true with data.
	resp, err := client.Call(context.Background(), "status", nil)
	if err != nil {
		t.Fatalf("Call(status) error = %v", err)
	}
	if !resp.Ok {
		t.Fatalf("Call(status) ok = false, error = %q", resp.Error)
	}
	if resp.Data == nil {
		t.Fatal("Call(status) returned nil data")
	}
	if _, ok := resp.Data.(map[string]any); !ok {
		t.Fatalf("Call(status) data = %T, want map[string]any", resp.Data)
	}

	// nonsense → ok:false with an error message.
	resp, err = client.Call(context.Background(), "nonsense", nil)
	if err != nil {
		t.Fatalf("Call(nonsense) error = %v", err)
	}
	if resp.Ok {
		t.Fatal("Call(nonsense) ok = true, want false")
	}
	if resp.Error == "" {
		t.Fatal("Call(nonsense) missing error message")
	}

	// quit → ok:true, then the server shuts down.
	resp, err = client.Call(context.Background(), "quit", nil)
	if err != nil {
		t.Fatalf("Call(quit) error = %v", err)
	}
	if !resp.Ok {
		t.Fatalf("Call(quit) ok = false, error = %q", resp.Error)
	}

	select {
	case <-server.ShutdownCh():
	case <-time.After(3 * time.Second):
		t.Fatal("server did not shut down after quit command")
	}
	if err := <-serveDone; err != nil {
		t.Fatalf("Serve returned error = %v", err)
	}
}

// TestServerStaleSocketRemoved verifies the stale-socket probe: a leftover
// socket file with no live listener is removed so Serve can bind it.
func TestServerStaleSocketRemoved(t *testing.T) {
	engine := newTestEngine(t)
	sock := shortSockPath(t)
	if err := os.WriteFile(sock, []byte("stale"), 0600); err != nil {
		t.Fatalf("create stale socket file: %v", err)
	}

	server := NewServerWithAddr(engine, "unix", sock)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = server.Serve(ctx) }()
	t.Cleanup(server.Close)
	waitForSocket(t, sock)
}
