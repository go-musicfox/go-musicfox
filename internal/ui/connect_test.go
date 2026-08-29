package ui

import (
	"testing"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// TestPlayerShadowCompleteness is the S6 TC-3 shadow-integrity guard: with
// p.remote != nil and the embedded *core.Player nil (the connect-mode sentinel
// — NewNeteaseRemote never constructs a real core.Player, B9), every shadowed
// method must NOT fall through to the embedded player. A leaked shadow would
// nil-dereference the embedded field and panic. Each shadowed method is
// invoked over the live daemon subscription (real Call path) and must complete
// without a panic; the degraded local-playlist actions no-op via a nil
// netease shell (toasts dropped).
func TestPlayerShadowCompleteness(t *testing.T) {
	startRemoteDaemon(t)
	client := dialRemote(t)
	rp := NewRemotePlayer(nil, client)
	waitRemoteReady(t, rp)

	// netease/svc stay nil: toasts are no-ops and every method whose connect
	// branch touches svc returns before doing so.
	p := &Player{remote: rp}

	// Read side (playerRendererState + queries).
	_ = p.CurSong()
	_ = p.CurSongIndex()
	_ = p.PassedTime()
	_ = p.State()
	_ = p.Volume()
	_ = p.Mode()
	_ = p.Playlist()
	_ = p.User()
	_ = p.PlaylistUpdateAt()
	_ = p.PlayingMenuKey()
	_ = p.CompareWithCurPlaylist(nil)
	_ = p.CommandContext()

	// Control side (forwarded over the real subscription; the fresh empty
	// daemon turns every command into a safe no-op).
	p.Seek(0)
	p.Toggle()
	p.Pause()
	p.Resume()
	p.Stop()
	p.NextSong(true)
	p.PreviousSong(true)
	p.SwitchMode()
	p.SetMode(types.PmListLoop)
	p.SetMode(types.PmIntelligent)
	p.SetVolume(50)
	p.UpVolume()
	p.DownVolume()

	// Degraded local-playlist actions (toast no-ops with a nil shell).
	p.PlaySong(structs.Song{}, DurationNext)
	p.ReinitializePlaylist(0, nil)
	p.InitSongManager(0, nil)
	p.StartPlay()
	_, _ = p.RemoveSong(0)
	_, _ = p.NextPlaylistSong(true)
	p.MarkPlaylistUpdated()

	// Wrapper-own methods with connect guards.
	p.SetPlayingMenu("key", nil)
	p.MarkPlaylistModified()
	if page := p.Intelligence(false); page != nil {
		t.Fatalf("Intelligence connect = %v, want nil (degraded)", page)
	}
	_ = p.RenderTicker()
	p.RequestLogin(nil)
}

// TestNewNeteaseRemoteAssembly locks the connect-shell construction (S6 TC-3):
// no engine, no degraded lyric/spectrum renderers, search singleton built, the
// connect branch of Components, and the menuServices Player/User + ToLoginPage
// / ToSearchPage connect degradations.
func TestNewNeteaseRemoteAssembly(t *testing.T) {
	prevConfig := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	t.Cleanup(func() {
		configs.AppConfig = prevConfig
		connectMode = false
	})

	n := NewNeteaseRemote(model.NewApp(model.DefaultOptions()), nil)
	if n == nil {
		t.Fatal("NewNeteaseRemote returned nil")
	}
	if !n.ConnectMode() {
		t.Fatal("ConnectMode() = false, want true")
	}
	if n.engine != nil || n.ctx != nil {
		t.Fatal("connect shell built an engine/framework context (B9)")
	}
	if n.search == nil {
		t.Fatal("connect shell search singleton is nil")
	}
	if n.lyricRenderer != nil || n.spectrumRenderer != nil || n.spectrogramRenderer != nil {
		t.Fatal("connect shell built degraded renderers (lyric/spectrum, B5/B7)")
	}
	if n.songInfoRenderer == nil || n.progressRenderer == nil || n.coverRenderer == nil {
		t.Fatal("connect shell missing song info/progress/cover renderers")
	}

	// Components: song info + progress only (cover disabled by default config);
	// no lyric/spectrum renderers.
	comps := n.Components()
	if len(comps) != 2 {
		t.Fatalf("connect Components length = %d, want 2 (songInfo+progress)", len(comps))
	}
	for _, c := range comps {
		switch c.(type) {
		case *LyricRenderer, *SpectrumRenderer, *SpectrogramRenderer:
			t.Fatalf("connect Components contains degraded renderer %T", c)
		}
	}

	// menuServices connect branches: Player resolves the remote wrapper; User
	// resolves from the (empty) remote cache.
	svc := newMenuServices(n)
	if svc.Player() != n.player {
		t.Fatal("menuServices.Player() != shell wrapper in connect mode")
	}
	if u := svc.User(); u != nil {
		t.Fatalf("menuServices.User() = %+v, want nil before snapshot", u)
	}

	// Login gating degrades to a nil page (B8); search entry degrades too
	// (plugin menus absent).
	if page, cmd := n.ToLoginPage(nil); page != nil || cmd != nil {
		t.Fatalf("connect ToLoginPage = (%v, %v), want (nil, nil)", page, cmd)
	}
	if page, cmd := n.ToSearchPage(StSingleSong); page != nil || cmd != nil {
		t.Fatalf("connect ToSearchPage = (%v, %v), want (nil, nil)", page, cmd)
	}

	// The connect search-page fallback must be idempotent (no duplicate
	// provider panic on a second shell build).
	n2 := NewNeteaseRemote(model.NewApp(model.DefaultOptions()), nil)
	if n2 == nil || n2.search == nil {
		t.Fatal("second connect shell failed to build search singleton")
	}
}

// TestConnectMenuChainRelaxed proves the connect-mode after-anchor chain
// relaxation: plugin anchors are absent (the frontend scope that registers the
// 9 business plugins is skipped, B10), so a missing After anchor must re-anchor
// to the chain tail instead of panicking.
func TestConnectMenuChainRelaxed(t *testing.T) {
	connectMode = true
	t.Cleanup(func() { connectMode = false })

	entries := []mainMenuEntry{
		{key: "first", after: MainMenuStart, title: "第一项"},
		{key: "help", after: "last_fm", title: "帮助", builtin: true}, // last_fm is plugin-registered
		{key: "new_item", after: "help", title: "新项"},
	}
	got := orderMainMenuEntries(entries)
	want := []string{"first", "help", "new_item"}
	for i, key := range want {
		if got[i].key != key {
			t.Fatalf("chain[%d] = %q, want %q (full: %v)", i, got[i].key, key, chainKeysOf(got))
		}
	}
}
