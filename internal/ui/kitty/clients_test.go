package kitty

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestCountTmuxClients(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   int
	}{
		{"empty", "", 0},
		{"whitespace", "  \n\t\n", 0},
		{"one", "/dev/ttys001\txterm-ghostty\n", 1},
		{"two", "/dev/ttys001\txterm-ghostty\n/dev/ttys002\txterm-kitty\n", 2},
		{"trailing newline", "/dev/ttys001\n/dev/ttys002\n", 2},
		{"crlf", "/dev/ttys001\r\n/dev/ttys002\r\n", 2},
		{"blank line between", "/dev/ttys001\n\n/dev/ttys002\n", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := countTmuxClients(tt.output); got != tt.want {
				t.Fatalf("countTmuxClients(%q) = %d, want %d", tt.output, got, tt.want)
			}
		})
	}
}

func stubTmuxListClients(t *testing.T, output string, err error) {
	t.Helper()
	orig := tmuxExecFn
	t.Cleanup(func() {
		tmuxExecFn = orig
		InvalidateTmuxClientCache()
	})
	InvalidateTmuxClientCache()
	tmuxExecFn = func(_ context.Context, args ...string) ([]byte, error) {
		joined := strings.Join(args, " ")
		if !strings.HasPrefix(joined, "list-clients") {
			t.Fatalf("unexpected tmux args: %v", args)
		}
		if err != nil {
			return nil, err
		}
		return []byte(output), nil
	}
}

func TestTmuxHasMultipleClients(t *testing.T) {
	t.Setenv("TMUX", "/tmp/go-musicfox-kitty-test-socket,123,0")

	t.Run("zero clients", func(t *testing.T) {
		stubTmuxListClients(t, "", nil)
		if TmuxHasMultipleClients() {
			t.Fatal("expected false for zero clients")
		}
	})

	t.Run("one client", func(t *testing.T) {
		stubTmuxListClients(t, "/dev/ttys001\txterm-ghostty\n", nil)
		if TmuxHasMultipleClients() {
			t.Fatal("expected false for a single client")
		}
	})

	t.Run("two clients", func(t *testing.T) {
		stubTmuxListClients(t, "/dev/ttys001\txterm-ghostty\n/dev/ttys002\txterm-kitty\n", nil)
		if !TmuxHasMultipleClients() {
			t.Fatal("expected true for two clients")
		}
	})

	t.Run("exec error fail-open", func(t *testing.T) {
		stubTmuxListClients(t, "", errors.New("tmux missing"))
		if TmuxHasMultipleClients() {
			t.Fatal("expected fail-open false on exec error")
		}
	})

	t.Run("empty env fail-open", func(t *testing.T) {
		t.Setenv("TMUX", "")
		InvalidateTmuxClientCache()
		t.Cleanup(InvalidateTmuxClientCache)
		calls := 0
		orig := tmuxExecFn
		t.Cleanup(func() { tmuxExecFn = orig })
		tmuxExecFn = func(context.Context, ...string) ([]byte, error) {
			calls++
			return nil, errors.New("should not run")
		}
		if TmuxHasMultipleClients() {
			t.Fatal("expected false when TMUX is unset")
		}
		if calls != 0 {
			t.Fatalf("expected no tmux exec when TMUX unset, got %d calls", calls)
		}
	})
}

func TestTmuxHasMultipleClientsCache(t *testing.T) {
	t.Setenv("TMUX", "/tmp/go-musicfox-kitty-test-socket,123,0")
	InvalidateTmuxClientCache()
	t.Cleanup(InvalidateTmuxClientCache)

	calls := 0
	orig := tmuxExecFn
	t.Cleanup(func() { tmuxExecFn = orig })
	tmuxExecFn = func(_ context.Context, args ...string) ([]byte, error) {
		calls++
		return []byte("/dev/ttys001\n/dev/ttys002\n"), nil
	}

	if !TmuxHasMultipleClients() {
		t.Fatal("expected true on first query")
	}
	if !TmuxHasMultipleClients() {
		t.Fatal("expected cached true on second query")
	}
	if calls != 1 {
		t.Fatalf("expected one exec within TTL, got %d", calls)
	}

	InvalidateTmuxClientCache()
	if !TmuxHasMultipleClients() {
		t.Fatal("expected true after invalidate")
	}
	if calls != 2 {
		t.Fatalf("expected re-query after invalidate, got %d calls", calls)
	}
}

func TestTmuxHasMultipleClientsNegativeCache(t *testing.T) {
	t.Setenv("TMUX", "/tmp/go-musicfox-kitty-test-socket,123,0")
	InvalidateTmuxClientCache()
	t.Cleanup(InvalidateTmuxClientCache)

	calls := 0
	orig := tmuxExecFn
	t.Cleanup(func() { tmuxExecFn = orig })
	tmuxExecFn = func(context.Context, ...string) ([]byte, error) {
		calls++
		return nil, errors.New("boom")
	}

	if TmuxHasMultipleClients() {
		t.Fatal("expected fail-open false")
	}
	if TmuxHasMultipleClients() {
		t.Fatal("expected negative cache to keep fail-open false")
	}
	if calls != 1 {
		t.Fatalf("expected one exec within fail TTL, got %d", calls)
	}

	// Seed an expired failure so the next call re-queries.
	tmuxClientMu.Lock()
	tmuxClientFailAt = time.Now().Add(-2 * tmuxClientFailCacheTTL)
	tmuxClientMu.Unlock()

	if TmuxHasMultipleClients() {
		t.Fatal("expected fail-open false after TTL expiry")
	}
	if calls != 2 {
		t.Fatalf("expected re-query after fail TTL, got %d calls", calls)
	}
}

func TestInvalidateTmuxClientCache(t *testing.T) {
	tmuxClientMu.Lock()
	tmuxClientCached = true
	tmuxClientAt = time.Now()
	tmuxClientCount = 2
	tmuxClientFailAt = time.Now()
	tmuxClientMu.Unlock()
	t.Cleanup(InvalidateTmuxClientCache)

	InvalidateTmuxClientCache()

	tmuxClientMu.Lock()
	defer tmuxClientMu.Unlock()
	if tmuxClientCached {
		t.Error("expected success cache cleared")
	}
	if !tmuxClientFailAt.IsZero() {
		t.Error("expected failure cache cleared")
	}
}
