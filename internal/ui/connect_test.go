package ui

import (
	"testing"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// connectTestPlugin is the S6-R1 test-double business plugin: the ui test
// binary cannot link the real plugins (ui must not import internal/plugins),
// so connect_test registers framework constructors for the 9 business plugin
// IDs that mirror the real shape — Start ignores ctx and registers a unique
// menu key per id inside a WithPlugin scope (the search double additionally
// registers the "search" page, like the real plugin). No main-menu items are
// registered, so the after-anchor chain tests stay stable.
//
// The registrations are dup-guarded: the test binary constructs several shells
// in one process (the integration tests), and the package-level registries are
// process-global — re-running Start must not panic on duplicate keys (the
// production invariant is one shell per process, which is why the real plugins
// need no guard).
type connectTestPlugin struct {
	framework.NoopPlugin
	id string
}

func (p *connectTestPlugin) Start(*framework.Context) error {
	WithPlugin(p.id, p.id, func() {
		if _, dup := menuRegistry["r1_"+p.id+"_menu"]; !dup {
			RegisterMenu("r1_"+p.id+"_menu", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
				return &testCheckUpdateMenu{baseMenu: base}, nil
			})
		}
		if p.id == "search" {
			if _, dup := pageRegistry["search"]; !dup {
				RegisterPage("search", func(opts SearchPageOpts) (model.Page, error) {
					return NewSearchPage(opts.Netease), nil
				})
			}
		}
	})
	return nil
}

// init registers framework plugin constructors for all 9 business plugin ids
// (including lastfm) so the S6-R1 mount path can be exercised in this binary.
// lastfm is registered too: the connect scope must exclude it by id even when
// its constructor exists.
func init() {
	for _, id := range businessPluginIDs {
		framework.RegisterPlugin(id, func(id string) func() framework.Plugin {
			return func() framework.Plugin { return &connectTestPlugin{id: id} }
		}(id))
	}
}

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

// TestNewNeteaseRemoteAssembly locks the connect-shell construction (S6 TC-3 +
// R1): no engine, no degraded lyric/spectrum renderers, the connect frontend
// scope mounting the 8 engine-independent business plugins (lastfm excluded),
// the search singleton built through the search plugin's provider, the connect
// branch of Components, and the menuServices Player/User + ToLoginPage connect
// degradations. ToSearchPage now returns the search page (search is local).
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
	if n.frontendScope == nil {
		t.Fatal("connect shell has no frontend scope (R1)")
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

	// The connect scope mounted the 8 engine-independent business plugins
	// (test-doubles) and excluded lastfm (skipIDs) even though its constructor
	// is registered. Asserted on the scope's own plugin set (order-independent)
	// plus the Start effect (the per-plugin menu keys landed in menuRegistry).
	mounted := map[string]bool{}
	for _, p := range n.frontendScope.Plugins() {
		if idp, ok := p.(*identifiedPlugin); ok {
			mounted[idp.PluginID()] = true
		}
	}
	for _, id := range []string{"checkupdate", "search", "dj", "album", "artist", "recommend", "playlist", "song"} {
		if !mounted[id] {
			t.Fatalf("connect scope did not mount plugin %q (mounted set: %v)", id, mounted)
		}
		if _, ok := menuRegistry["r1_"+id+"_menu"]; !ok {
			t.Fatalf("mounted plugin %q did not register its menu (r1_%s_menu missing)", id, id)
		}
	}
	if mounted["lastfm"] {
		t.Fatal("lastfm was mounted by the connect scope (must be excluded: its Deps needs engine services)")
	}
	if _, ok := menuRegistry["r1_lastfm_menu"]; ok {
		t.Fatal("r1_lastfm_menu registered — the connect scope must not mount lastfm")
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

	// Login gating degrades to a nil page (B8).
	if page, cmd := n.ToLoginPage(nil); page != nil || cmd != nil {
		t.Fatalf("connect ToLoginPage = (%v, %v), want (nil, nil)", page, cmd)
	}

	// Search is local (S6-R1): the search plugin is mounted, so ToSearchPage
	// returns the shell search singleton with the type set (standalone flow).
	page, cmd := n.ToSearchPage(StSingleSong)
	if page == nil || cmd == nil {
		t.Fatalf("connect ToSearchPage = (%v, %v), want (search page, tick cmd)", page, cmd)
	}
	if sp, ok := page.(*SearchPage); !ok || sp.searchType != StSingleSong {
		t.Fatalf("connect ToSearchPage page = %T with searchType %v, want *SearchPage with StSingleSong", page, sp.searchType)
	}

	// A second shell construction must not panic: the test-double plugins'
	// Start dup-guards their registrations (the test process builds several
	// shells; production builds one shell per process).
	n2 := NewNeteaseRemote(model.NewApp(model.DefaultOptions()), nil)
	if n2 == nil || n2.search == nil {
		t.Fatal("second connect shell failed to build search singleton")
	}
}

// TestConnectSearchFallbackWithDisabledSearch proves the registerConnectProviders
// fallback guard (S6-R1): when [plugins] disabled contains "search", the search
// plugin is not mounted, so the shell re-provides the "search" page and the
// singleton still builds.
func TestConnectSearchFallbackWithDisabledSearch(t *testing.T) {
	prevConfig := configs.AppConfig
	configs.AppConfig = &configs.Config{Plugins: configs.PluginsConfig{Disabled: []string{"search"}}}
	t.Cleanup(func() {
		configs.AppConfig = prevConfig
		connectMode = false
	})

	n := NewNeteaseRemote(model.NewApp(model.DefaultOptions()), nil)
	if n == nil {
		t.Fatal("NewNeteaseRemote returned nil with search disabled")
	}
	// AddWithEnabled(false) keeps the plugin in the scope slice but never
	// starts it (P5 "disabled = nonexistent"): assert the enabled flag, not
	// absence from Plugins().
	for _, p := range n.frontendScope.Plugins() {
		if idp, ok := p.(*identifiedPlugin); ok && idp.PluginID() == "search" {
			if idp.IsEnabled() {
				t.Fatal("search plugin enabled despite [plugins] disabled (Start must be skipped)")
			}
		}
	}
	if n.search == nil {
		t.Fatal("connect shell with search disabled failed to build the search singleton via the fallback provider")
	}
}

// TestConnectMainMenuFullChain locks the S6-R1 outcome: with the business
// plugins mounted in the connect scope, NewMainMenu builds the full menu tree
// (the test binary's 14 test-double items + built-in 帮助 — the equivalent of
// the production 8-plugin chain minus LastFM) instead of the TC-3 skeleton's
// single 帮助 item. This is the "8 plugins mounted, lastfm missing" complete
// chain case: the connect scope excludes lastfm (its Deps needs engine
// services), yet the chain still builds completely — the test binary registers
// the last_fm main-menu item directly, so the built-in 帮助 anchors onto it
// and lands at the tail exactly as the production relaxed re-anchoring would.
func TestConnectMainMenuFullChain(t *testing.T) {
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
	menu := NewMainMenu(NewBaseMenu(n))
	titles := menu.Titles()

	// The 14 test-double items keep the production browsing chain order
	// (roadmap §8.4 browsing rows); 帮助 sits at the chain tail after LastFM.
	wantHead := []string{
		"每日推荐歌曲", "每日推荐歌单", "我的歌单", "我的收藏", "私人FM",
		"专辑列表", "搜索", "排行榜", "精选歌单", "热门歌手",
		"最近播放歌曲", "云盘", "主播电台", "LastFM",
	}
	if len(titles) < len(wantHead)+1 {
		t.Fatalf("connect main menu has %d items, want >= %d: %v", len(titles), len(wantHead)+1, titles)
	}
	for i, want := range wantHead {
		if titles[i] != want {
			t.Fatalf("menu[%d] = %q, want %q (full: %v)", i, titles[i], want, titles)
		}
	}
	if titles[len(wantHead)] != "帮助" {
		t.Fatalf("menu[%d] = %q, want 帮助 (chain tail)", len(wantHead), titles[len(wantHead)])
	}
}

// TestConnectMenuChainRelaxed proves the connect-mode after-anchor chain
// relaxation mechanics on a focused synthetic entry set: a missing After anchor
// re-anchors the dependent entry (and its followers) to the chain tail instead
// of panicking. The real "8 plugins mounted, lastfm excluded" complete-chain
// case — NewNeteaseRemote + NewMainMenu after the connect scope mount — is
// covered by TestConnectMainMenuFullChain (and, over a live daemon, by
// TestConnectIntegrationMenuChain).
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
