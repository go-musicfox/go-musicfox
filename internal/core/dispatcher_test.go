package core

import (
	"context"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	cookiejar "github.com/juju/persistent-cookiejar"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/player"
	"github.com/go-musicfox/go-musicfox/internal/playermiddleware"
	"github.com/go-musicfox/go-musicfox/internal/reporter"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/track"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/netease"
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

// recordingEngine records Play calls so tests can prove StartPlay reached the
// audio engine without relying on network resolution or real playback.
type recordingEngine struct {
	stubEngine
	played []player.URLMusic
}

func (s *recordingEngine) Play(m player.URLMusic) {
	s.played = append(s.played, m)
}

// fakeTrackProvider resolves every song to a fixed remote source so StartPlay
// completes without network access (the dispatcher play_list playback path).
type fakeTrackProvider struct{}

func (fakeTrackProvider) ResolvePlayableSource(_ context.Context, _ structs.Song) (track.PlayableSource, error) {
	return track.PlayableSource{
		Type: track.SourceRemote,
		Info: &netease.PlayableInfo{URL: "http://example.invalid/test.mp3", MusicType: "mp3"},
	}, nil
}

// newPlaybackTestEngine builds an engine whose player can complete the
// StartPlay path deterministically: a bare Player wired with the recording
// engine, a provider-backed track manager, a no-op reporter and an empty
// middleware chain — no real audio output, no network.
func newPlaybackTestEngine(t *testing.T) *Engine {
	t.Helper()
	e := testEngine()
	p := NewEmptyPlayer()
	p.Player = &recordingEngine{stubEngine: stubEngine{state: types.Stopped}}
	p.trackManager = track.NewManager(track.WithPlayableSourceProvider(fakeTrackProvider{}))
	p.reporter = reporter.NewService()
	p.middlewareChain = playermiddleware.NewChain()
	e.player = p
	return e
}

func TestDispatcherLoginQRKey(t *testing.T) {
	d := NewDispatcher(newTestEngine(t))

	prev := qrGetKey
	qrGetKey = func(_ http.CookieJar) (string, string, error) {
		return "unikey-test", "http://music.163.com/login?codekey=unikey-test", nil
	}
	defer func() { qrGetKey = prev }()

	data, err := d.Dispatch(context.Background(), "login_qr_key", nil)
	if err != nil {
		t.Fatalf("Dispatch(login_qr_key) error = %v", err)
	}
	m, ok := data.(map[string]any)
	if !ok {
		t.Fatalf("login_qr_key data = %T, want map[string]any", data)
	}
	if m["uniKey"] != "unikey-test" {
		t.Fatalf("uniKey = %v, want unikey-test", m["uniKey"])
	}
	if m["qrcodeUrl"] != "http://music.163.com/login?codekey=unikey-test" {
		t.Fatalf("qrcodeUrl = %v, want the reconstructed scan URL", m["qrcodeUrl"])
	}
}

func TestDispatcherLoginQRStatusPending(t *testing.T) {
	d := NewDispatcher(newTestEngine(t))

	prev := qrCheckStatus
	qrCheckStatus = func(_ string, _ http.CookieJar) (float64, []byte, error) {
		return 802, []byte(`{"code":802}`), nil
	}
	defer func() { qrCheckStatus = prev }()

	data, err := d.Dispatch(context.Background(), "login_qr_status", map[string]any{"key": "unikey-test"})
	if err != nil {
		t.Fatalf("Dispatch(login_qr_status) error = %v", err)
	}
	m, ok := data.(map[string]any)
	if !ok {
		t.Fatalf("login_qr_status data = %T, want map[string]any", data)
	}
	if code, ok := numOf(m["code"]); !ok || int(code) != 802 {
		t.Fatalf("code = %v, want 802", m["code"])
	}
	if _, ok := m["user"].(map[string]any); !ok {
		t.Fatalf("user = %T, want map[string]any (empty until 803)", m["user"])
	}

	// A missing key must fail before any network call.
	if _, err := d.Dispatch(context.Background(), "login_qr_status", nil); err == nil {
		t.Fatal("Dispatch(login_qr_status without key) expected an error")
	}
}

func TestDispatcherLoginQRStatus803CompletesLogin(t *testing.T) {
	engine := newTestEngine(t)
	d := NewDispatcher(engine)

	prevCheck := qrCheckStatus
	prevComplete := completeQRLogin
	prevUser := *engine.UserSlot()
	defer func() {
		qrCheckStatus = prevCheck
		completeQRLogin = prevComplete
		*engine.UserSlot() = prevUser
	}()

	qrCheckStatus = func(_ string, _ http.CookieJar) (float64, []byte, error) {
		return 803, []byte(`{"code":803}`), nil
	}
	called := false
	completeQRLogin = func(e *Engine, _ *cookiejar.Jar) error {
		called = true
		*(e.UserSlot()) = &structs.User{UserId: 424242, Nickname: "scanned-fox"}
		return nil
	}

	data, err := d.Dispatch(context.Background(), "login_qr_status", map[string]any{"key": "unikey-test"})
	if err != nil {
		t.Fatalf("Dispatch(login_qr_status) error = %v", err)
	}
	if !called {
		t.Fatal("CompleteQRLogin was not invoked on 803")
	}
	m, ok := data.(map[string]any)
	if !ok {
		t.Fatalf("login_qr_status data = %T, want map[string]any", data)
	}
	if code, ok := numOf(m["code"]); !ok || int(code) != 803 {
		t.Fatalf("code = %v, want 803", m["code"])
	}
	user, ok := m["user"].(map[string]any)
	if !ok {
		t.Fatalf("user = %T, want map[string]any", m["user"])
	}
	if id, ok := numOf(user["userId"]); !ok || int64(id) != 424242 {
		t.Fatalf("userId = %v, want 424242", user["userId"])
	}
	if nick, ok := user["nickname"].(string); !ok || nick != "scanned-fox" {
		t.Fatalf("nickname = %v, want scanned-fox", user["nickname"])
	}
}

func TestDispatcherPlayList(t *testing.T) {
	engine := newPlaybackTestEngine(t)
	d := NewDispatcher(engine)

	songs := []any{
		map[string]any{"id": float64(1), "name": "Song A", "artist": "Artist A", "album": "Album A"},
		map[string]any{"id": float64(2), "name": "Song B", "artist": "Artist B", "album": "Album B"},
	}

	// play=false rebuilds the queue at the given index without starting playback.
	data, err := d.Dispatch(context.Background(), "play_list", map[string]any{
		"songs": songs,
		"index": float64(1),
		"play":  false,
	})
	if err != nil {
		t.Fatalf("Dispatch(play_list play=false) error = %v", err)
	}
	m, ok := data.(map[string]any)
	if !ok {
		t.Fatalf("play_list data = %T, want map[string]any", data)
	}
	if started, ok := m["started"].(bool); !ok || started {
		t.Fatalf("started = %v, want false", m["started"])
	}
	idx, ok := numOf(m["index"])
	if !ok || int(idx) != 1 {
		t.Fatalf("index = %v, want 1", m["index"])
	}
	pl, ok := m["playlist"].([]map[string]any)
	if !ok || len(pl) != 2 {
		t.Fatalf("playlist = %T of len %d, want []map[string]any of len 2", m["playlist"], len(pl))
	}
	for k, want := range map[string]any{"id": int64(2), "name": "Song B", "artist": "Artist B", "album": "Album B"} {
		if pl[1][k] != want {
			t.Fatalf("playlist[1].%s = %v, want %v", k, pl[1][k], want)
		}
	}
	if len(engine.Player().Playlist()) != 2 || engine.Player().CurSongIndex() != 1 {
		t.Fatalf("player queue = len %d index %d, want len 2 index 1", len(engine.Player().Playlist()), engine.Player().CurSongIndex())
	}

	// play=true starts playback at the clamped index (99 → last).
	rec := engine.Player().Player.(*recordingEngine)
	data, err = d.Dispatch(context.Background(), "play_list", map[string]any{
		"songs": songs,
		"index": float64(99),
		"play":  true,
	})
	if err != nil {
		t.Fatalf("Dispatch(play_list play=true) error = %v", err)
	}
	m = data.(map[string]any)
	if started, ok := m["started"].(bool); !ok || !started {
		t.Fatalf("started = %v, want true", m["started"])
	}
	idx, ok = numOf(m["index"])
	if !ok || int(idx) != 1 {
		t.Fatalf("index = %v, want clamped 1", m["index"])
	}
	if len(rec.played) != 1 {
		t.Fatalf("engine Play calls = %d, want 1 (StartPlay reached the engine)", len(rec.played))
	}
	if rec.played[0].Song.Id != 2 {
		t.Fatalf("played song id = %d, want 2 (clamped index 1)", rec.played[0].Song.Id)
	}
}

func TestDispatcherPlayListErrors(t *testing.T) {
	d := NewDispatcher(newPlaybackTestEngine(t))

	for _, args := range []map[string]any{
		nil,
		{},
		{"songs": "not-an-array"},
		{"songs": []any{map[string]any{"name": "no-id"}}},
	} {
		if _, err := d.Dispatch(context.Background(), "play_list", args); err == nil {
			t.Fatalf("Dispatch(play_list with args %v) expected an error", args)
		}
	}

	if _, err := d.Dispatch(context.Background(), "play_list", map[string]any{
		"songs": []any{map[string]any{"id": float64(1)}},
		"index": "not-a-number",
	}); err == nil {
		t.Fatal("Dispatch(play_list with non-numeric index) expected an error")
	}

	// Guard rail: more than playListWireLimit songs is rejected.
	songs := make([]any, playListWireLimit+1)
	for i := range songs {
		songs[i] = map[string]any{"id": float64(i + 1)}
	}
	if _, err := d.Dispatch(context.Background(), "play_list", map[string]any{"songs": songs}); err == nil {
		t.Fatal("Dispatch(play_list with >playListWireLimit songs) expected an error")
	}
}
