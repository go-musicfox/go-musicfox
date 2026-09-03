package kitty

import (
	"context"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	// tmuxClientCacheTTL is how long a successful list-clients count is cached.
	// Cover IsEnabled/View may poll every frame; a short TTL avoids forking
	// tmux on every paint while still noticing a second attach quickly.
	tmuxClientCacheTTL = time.Second
	// tmuxClientFailCacheTTL negatively caches list-clients failures so a
	// broken tmux binary does not get re-exec'd every frame. Fail-open:
	// failures are treated as "not multiple" so covers stay available.
	tmuxClientFailCacheTTL = time.Second
)

var (
	tmuxClientMu     sync.Mutex
	tmuxClientCount  int
	tmuxClientCached bool
	tmuxClientAt     time.Time
	tmuxClientFailAt time.Time
)

// TmuxHasMultipleClients reports whether more than one client is attached to
// the current tmux server. Used to suppress Kitty graphics passthrough when
// DCS would fan out to multiple terminals (e.g. Ghostty + Kitty), which has
// caused CPU/GPU hangs. Fail-open: returns false on query error so a broken
// tmux binary does not permanently kill covers.
func TmuxHasMultipleClients() bool {
	return cachedTmuxClientCount() > 1
}

// cachedTmuxClientCount returns the attached client count, using a short
// success/failure TTL cache. On query failure it returns 0 (fail-open).
func cachedTmuxClientCount() int {
	tmuxClientMu.Lock()
	defer tmuxClientMu.Unlock()

	now := time.Now()
	if tmuxClientCached && now.Sub(tmuxClientAt) < tmuxClientCacheTTL {
		return tmuxClientCount
	}
	if !tmuxClientFailAt.IsZero() && now.Sub(tmuxClientFailAt) < tmuxClientFailCacheTTL {
		return 0
	}

	if os.Getenv("TMUX") == "" {
		tmuxClientFailAt = now
		tmuxClientCached = false
		return 0
	}

	// Include termname for potential debug logging; count uses non-empty lines.
	output, err := tmuxExecFn(context.Background(), "list-clients", "-F", "#{client_tty}\t#{client_termname}")
	if err != nil {
		tmuxClientFailAt = now
		tmuxClientCached = false
		return 0
	}
	count := countTmuxClients(string(output))
	tmuxClientCount = count
	tmuxClientCached = true
	tmuxClientAt = now
	tmuxClientFailAt = time.Time{}
	return count
}

// countTmuxClients counts non-empty lines from `tmux list-clients -F ...`
// output. Empty or whitespace-only output yields 0.
func countTmuxClients(output string) int {
	n := 0
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		n++
	}
	return n
}

// InvalidateTmuxClientCache resets the client-count caches (success and
// failure), forcing the next TmuxHasMultipleClients call to re-query tmux.
func InvalidateTmuxClientCache() {
	tmuxClientMu.Lock()
	defer tmuxClientMu.Unlock()
	tmuxClientCached = false
	tmuxClientFailAt = time.Time{}
}
