package core

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// The beep speaker can only be initialized once per process, and the engine's
// ctrl goroutine processes playback signals asynchronously (reading configs and
// storage while doing so). The package therefore runs all Dispatcher tests
// against ONE shared engine with process-lifetime config/storage setup, set up
// here in TestMain so no test ever tears the state out from under the engine's
// background goroutines.
var (
	sharedEngineOnce sync.Once
	sharedEngineVal  *Engine

	testRoot string
)

func TestMain(m *testing.M) {
	var err error
	testRoot, err = os.MkdirTemp("", "musicfox-core-test-*")
	if err != nil {
		panic(err)
	}

	prevRoot := os.Getenv("MUSICFOX_ROOT")
	prevConfig := configs.AppConfig
	prevDB := storage.DBManager

	// The first app path access bootstraps the path manager from MUSICFOX_ROOT,
	// so any db/cache files land in the temp root, never in user dirs.
	_ = os.Setenv("MUSICFOX_ROOT", testRoot)
	configs.AppConfig = &configs.Config{Player: configs.PlayerConfig{Engine: types.BeepPlayer}}
	storage.DBManager = new(storage.LocalDBManager)

	code := m.Run()

	storage.DBManager = prevDB
	configs.AppConfig = prevConfig
	if prevRoot == "" {
		_ = os.Unsetenv("MUSICFOX_ROOT")
	} else {
		_ = os.Setenv("MUSICFOX_ROOT", prevRoot)
	}
	_ = os.RemoveAll(testRoot)
	os.Exit(code)
}

// newTestEngine returns the shared real core engine (no Startup). Its
// engine.Player() is usable immediately after construction.
func newTestEngine(t *testing.T) *Engine {
	t.Helper()
	sharedEngineOnce.Do(func() {
		sharedEngineVal = NewEngine(EngineOptions{})
	})
	return sharedEngineVal
}

// waitForMode polls until the player reaches the wanted play mode (control
// signals are processed asynchronously by the engine's ctrl goroutine).
func waitForMode(t *testing.T, engine *Engine, want types.Mode) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		if engine.Player().Mode() == want {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("player mode = %v, want %v", engine.Player().Mode(), want)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// numOf converts a JSON-decodable number (float64) or a native int to float64,
// since Dispatch results carry native Go types when called in-process and
// float64 when they have crossed the JSON wire.
func numOf(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	}
	return 0, false
}

func TestDispatcherUnknownCommand(t *testing.T) {
	d := NewDispatcher(newTestEngine(t))
	_, err := d.Dispatch(context.Background(), "nonsense", nil)
	if err == nil {
		t.Fatal("Dispatch(nonsense) expected an error")
	}
	if err.Error() != "未知命令: nonsense" {
		t.Fatalf("unknown-command error = %q, want %q", err.Error(), "未知命令: nonsense")
	}
}

func TestDispatcherStatusOnFreshEngine(t *testing.T) {
	d := NewDispatcher(newTestEngine(t))
	data, err := d.Dispatch(context.Background(), "status", nil)
	if err != nil {
		t.Fatalf("Dispatch(status) error = %v", err)
	}
	m, ok := data.(map[string]any)
	if !ok {
		t.Fatalf("status data = %T, want map[string]any", data)
	}
	if playlistLen, ok := numOf(m["playlistLen"]); !ok || int(playlistLen) != 0 {
		t.Fatalf("status playlistLen = %v, want 0", m["playlistLen"])
	}
	if state, ok := m["state"].(string); !ok || state == "" {
		t.Fatalf("status state = %v, want a non-empty string", m["state"])
	}
	if _, ok := m["song"].(map[string]any); !ok {
		t.Fatalf("status song = %T, want map[string]any", m["song"])
	}
}

func TestDispatcherVolumeGetSet(t *testing.T) {
	engine := newTestEngine(t)
	d := NewDispatcher(engine)

	// Get: a fresh engine reports the underlying player volume (no panic).
	get, err := d.Dispatch(context.Background(), "volume", nil)
	if err != nil {
		t.Fatalf("Dispatch(volume get) error = %v", err)
	}
	getMap, ok := get.(map[string]any)
	if !ok {
		t.Fatalf("volume get data = %T, want map[string]any", get)
	}
	if _, ok := numOf(getMap["volume"]); !ok {
		t.Fatalf("volume get missing 'volume' key: %v", getMap)
	}

	// Set then get: the round-trip value must stick.
	if _, err := d.Dispatch(context.Background(), "volume", map[string]any{"value": float64(30)}); err != nil {
		t.Fatalf("Dispatch(volume set) error = %v", err)
	}
	set, err := d.Dispatch(context.Background(), "volume", nil)
	if err != nil {
		t.Fatalf("Dispatch(volume get after set) error = %v", err)
	}
	setMap := set.(map[string]any)
	if v, ok := numOf(setMap["volume"]); !ok || int(v) != 30 {
		t.Fatalf("volume after set = %v, want 30", setMap["volume"])
	}
}

func TestDispatcherRepeatShuffleMapping(t *testing.T) {
	engine := newTestEngine(t)
	d := NewDispatcher(engine)

	// repeat: "off" maps to ordered mode, "one" to single-loop, "all" to
	// list-loop. Control signals are async, so poll until the mode settles.
	for _, tc := range []struct {
		mode string
		want types.Mode
	}{{"off", types.PmOrdered}, {"one", types.PmSingleLoop}, {"all", types.PmListLoop}} {
		if _, err := d.Dispatch(context.Background(), "repeat", map[string]any{"mode": tc.mode}); err != nil {
			t.Fatalf("Dispatch(repeat %s) error = %v", tc.mode, err)
		}
		waitForMode(t, engine, tc.want)
	}
	if _, err := d.Dispatch(context.Background(), "repeat", map[string]any{"mode": "bogus"}); err == nil {
		t.Fatal("Dispatch(repeat bogus) expected an error")
	}

	// shuffle: on → list-random, off → back to a non-random mode.
	if _, err := d.Dispatch(context.Background(), "shuffle", map[string]any{"on": true}); err != nil {
		t.Fatalf("Dispatch(shuffle on) error = %v", err)
	}
	waitForMode(t, engine, types.PmListRandom)
	if _, err := d.Dispatch(context.Background(), "shuffle", map[string]any{"on": false}); err != nil {
		t.Fatalf("Dispatch(shuffle off) error = %v", err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for engine.Player().Mode() == types.PmListRandom {
		if time.Now().After(deadline) {
			t.Fatal("shuffle off: mode still PmListRandom")
		}
		time.Sleep(5 * time.Millisecond)
	}
	if _, err := d.Dispatch(context.Background(), "shuffle", map[string]any{}); err == nil {
		t.Fatal("Dispatch(shuffle without on) expected an error")
	}
}

// TestDispatcherPlayEmptyQueryError verifies the "missing query" error path
// without any network call: an empty/missing query must fail before the search
// service is ever invoked.
func TestDispatcherPlayEmptyQueryError(t *testing.T) {
	d := NewDispatcher(newTestEngine(t))
	for _, args := range []map[string]any{nil, {}, {"query": ""}, {"query": "   "}} {
		if _, err := d.Dispatch(context.Background(), "play", args); err == nil {
			t.Fatalf("Dispatch(play with args %v) expected a missing-query error", args)
		}
	}
}
