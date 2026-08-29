package ui

import (
	"context"
	"encoding/json"
	"io"
	"runtime"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/charmbracelet/x/ansi"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/headless"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// startRemoteDaemonWithEvents is startRemoteDaemon plus the headless
// DaemonPlugin, so engine event-bus frames stream to subscription connections
// — the production daemon shape the TUI connect shell subscribes to
// (emitter → DaemonPlugin → broadcastEvent → socket → SubscribeClient). It
// shares the process-lifetime engine/config fixtures (setupRemoteEnv) and the
// socket path (headless.ListenAddr) with the TC-2 daemon tests, which run
// sequentially.
func startRemoteDaemonWithEvents(t *testing.T) *headless.Server {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("daemon integration tests use a unix socket; the Windows TCP path is covered by the manual smoke checklist")
	}
	setupRemoteEnv()

	server := headless.NewServerWithAddr(remoteEngine, "unix", headless.ListenAddr())
	scope := framework.NewScope()
	if err := scope.Add(headless.NewDaemonPlugin(server)); err != nil {
		t.Fatalf("scope.Add(daemonPlugin) error = %v", err)
	}
	if err := scope.Start(remoteEngine.Ctx()); err != nil {
		t.Fatalf("scope.Start error = %v", err)
	}
	t.Cleanup(func() { _ = scope.Dispose() })
	t.Cleanup(server.Close)
	waitRemoteDaemon(t)
	return server
}

// newConnectShell assembles the TUI-connect shell the way RunConnect does —
// NewNeteaseRemote plus the main menu and the connect renderer set — without
// running the full TUI main loop: the foxful App is started with a cancelled
// context (the command_view_flow pattern) and driven by hand through
// app.Update, so the event consumer's Rerender poke is a safe no-op on the
// terminated program while the render path stays observable via app.View.
func newConnectShell(t *testing.T, client *headless.SubscribeClient) (*model.App, *Netease) {
	t.Helper()
	prevConfig := configs.AppConfig
	prevLocale := model.DefaultCatalog().Locale()
	configs.AppConfig = &configs.Config{Player: configs.PlayerConfig{Engine: types.BeepPlayer}}
	t.Cleanup(func() {
		configs.AppConfig = prevConfig
		model.SetLocale(prevLocale)
		connectMode = false
	})
	// Mirror the RunConnect assembly order: global assignments and the zh
	// catalog must be set before NewNeteaseRemote builds the shell (the search
	// page renders catalog-localized strings).
	model.Submit = types.SubmitText
	model.SearchPlaceholder = types.SearchPlaceholder
	model.SearchResult = types.SearchResult
	SetupI18n(configs.AppConfig.Main.Locale)

	opts := model.DefaultOptions()
	opts.EnableStartup = false
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	opts.TeaOptions = []tea.ProgramOption{
		tea.WithContext(ctx),
		tea.WithInput(nil),
		tea.WithOutput(io.Discard),
	}

	app := model.NewApp(opts)
	n := NewNeteaseRemote(app, client)
	n.With(
		model.WithMainMenu(NewMainMenu(NewBaseMenu(n)), &model.MenuItem{Title: "网易云音乐"}),
		func(options *model.Options) {
			options.Components = append(options.Components, n.Components()...)
		},
	)
	_ = app.Run()
	if app.Main() == nil {
		t.Fatal("main page was not initialized")
	}
	_, _ = app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	return app, n
}

// waitFor polls cond until it returns true or the deadline expires.
func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		if cond() {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("condition not met within timeout")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// readWireEvent reads client.Events() until a frame for the given event name
// arrives (the snapshot frame and unrelated events are skipped) or the timeout
// expires.
func readWireEvent(t *testing.T, client *headless.SubscribeClient, want string, timeout time.Duration) {
	t.Helper()
	deadline := time.After(timeout)
	for {
		select {
		case raw, ok := <-client.Events():
			if !ok {
				t.Fatalf("spy Events channel closed before %q arrived", want)
			}
			var f struct {
				Type  string `json:"type"`
				Event string `json:"event"`
			}
			if err := json.Unmarshal(raw, &f); err != nil {
				continue
			}
			if f.Type == "event" && f.Event == want {
				return
			}
		case <-deadline:
			t.Fatalf("spy did not receive %q within %v", want, timeout)
		}
	}
}

// connectAssertViewContains drives the app message loop with the poke
// App.Rerender sends (the active page's Msg) — the S6 "model.App message
// loop" pattern — and asserts the rendered view contains want: the observable
// redraw effect of the event-consumer goroutine's Rerender call on the shell.
func connectAssertViewContains(t *testing.T, app *model.App, want string) {
	t.Helper()
	if page := app.CurPage(); page != nil {
		_, _ = app.Update(page.Msg())
	}
	view := ansi.Strip(app.View().Content)
	if !strings.Contains(view, want) {
		t.Fatalf("shell view does not contain %q after redraw:\n%s", want, view)
	}
}

// TestConnectIntegrationSnapshotMapping locks the snapshot mapping of the full
// chain against a real daemon: a seeded daemon player state (song/volume/mode/
// playlist/user) reaches the shell cache over the subscription, agrees with the
// daemon's own status report, and the shell view renders the mapped song.
func TestConnectIntegrationSnapshotMapping(t *testing.T) {
	startRemoteDaemonWithEvents(t)

	// Seed the daemon player BEFORE subscribing so the snapshot carries a
	// meaningful, assertable state. Seeding is test-side (the shared engine),
	// not a production change.
	remoteEngine.Player().ReinitializePlaylist(0, []structs.Song{
		{Id: 1, Name: "第一首", Artists: []structs.Artist{{Name: "歌手A"}}, Album: structs.Album{Name: "专辑X"}},
		{Id: 2, Name: "第二首", Artists: []structs.Artist{{Name: "歌手B"}}, Album: structs.Album{Name: "专辑Y"}},
	})
	remoteEngine.Player().SetVolume(66)
	remoteEngine.Player().SetMode(types.PmListLoop)
	*remoteEngine.UserSlot() = &structs.User{Nickname: "tester"}
	t.Cleanup(func() {
		*remoteEngine.UserSlot() = nil
		remoteEngine.Player().ReinitializePlaylist(0, nil)
	})

	client := dialRemote(t)
	app, n := newConnectShell(t, client)
	rp := n.Player().remote
	waitRemoteReady(t, rp)

	// The shell read surface agrees with the daemon's own status report on
	// every mapped field (the snapshot → cache mapping).
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	resp, err := client.Call(ctx, "status", nil)
	if err != nil || resp == nil || !resp.Ok {
		t.Fatalf("daemon status call = %v err = %v, want ok", resp, err)
	}
	data, _ := resp.Data.(map[string]any)

	song := n.Player().CurSong()
	if song.Id != 1 || song.Name != "第一首" || len(song.Artists) != 1 || song.Artists[0].Name != "歌手A" || song.Album.Name != "专辑X" {
		t.Fatalf("CurSong() = %+v, want 第一首/歌手A/专辑X", song)
	}
	if got := n.Player().CurSongIndex(); got != 0 {
		t.Fatalf("CurSongIndex() = %d, want 0", got)
	}
	if got := n.Player().State(); got != stateFromWire(strOf(data["state"])) {
		t.Fatalf("State() = %v, want daemon status %v", got, data["state"])
	}
	if got := n.Player().Volume(); got != 66 {
		t.Fatalf("Volume() = %d, want 66", got)
	}
	if v, ok := data["volume"].(float64); !ok || int(v) != n.Player().Volume() {
		t.Fatalf("daemon status volume = %v, shell Volume() = %d, want equal", data["volume"], n.Player().Volume())
	}
	if got := n.Player().Mode(); got != types.PmListLoop {
		t.Fatalf("Mode() = %v, want PmListLoop", got)
	}
	pl := n.Player().Playlist()
	if len(pl) != 2 || pl[0].Id != 1 || pl[1].Id != 2 {
		t.Fatalf("Playlist() = %+v, want [1 2]", pl)
	}
	if u := n.Player().User(); u == nil || u.Nickname != "tester" {
		t.Fatalf("User() = %+v, want nickname tester", u)
	}

	// The renderer pipeline consumes the mapped state: the shell view renders
	// the current song (the redraw source).
	connectAssertViewContains(t, app, "第一首")
}

// TestConnectIntegrationEventRedrawAndControl locks the event + control plane
// of the full chain: the TUI shell's next action forwards Call("next") to the
// daemon (whose player emits player.playlist_exhausted on an empty queue — the
// spy subscription observes the wire delivery), a daemon song_changed event
// arrives at the shell cache and drives a redraw of the rendered view, and a
// SetVolume forward takes effect on the daemon.
func TestConnectIntegrationEventRedrawAndControl(t *testing.T) {
	startRemoteDaemonWithEvents(t)

	client := dialRemote(t)
	app, n := newConnectShell(t, client)
	rp := n.Player().remote
	waitRemoteReady(t, rp)

	// ctrl next: the shell's NextSong forwards Call("next") to the daemon. With
	// an empty queue the daemon player emits player.playlist_exhausted; a spy
	// subscription observes that wire frame — the daemon-side effect of the
	// control command the shell itself consumes silently.
	spy, err := headless.DialSubscribe([]string{core.EvPlaylistEnd})
	if err != nil {
		t.Fatalf("spy DialSubscribe error = %v", err)
	}
	t.Cleanup(func() { _ = spy.Close() })
	n.Player().NextSong(true)
	readWireEvent(t, spy, core.EvPlaylistEnd, 3*time.Second)

	// Event arrival + redraw: emit the song_changed frame the daemon produces
	// when a next lands on a loaded song (core emits the same shape) and assert
	// it reaches the shell cache AND the renderer pipeline — the consumer
	// goroutine's Rerender makes the view show the event's duration.
	emitter, ok := framework.ServiceOf[*framework.EventEmitter](remoteEngine.Ctx(), core.ServiceEventBus)
	if !ok {
		t.Fatal("eventBus not resolved from engine ctx")
	}
	if err := emitter.Emit(remoteEngine.Ctx(), core.EvSongChanged, map[string]any{
		"id":              int64(42),
		"name":            "新歌",
		"artist":          "歌手C",
		"album":           "专辑Z",
		"picUrl":          "https://p1.music.126.net/cover.jpg",
		"durationSeconds": 180.0,
	}); err != nil {
		t.Fatalf("emit song_changed: %v", err)
	}
	waitFor(t, 3*time.Second, func() bool { return n.Player().CurSong().Id == 42 })
	if song := n.Player().CurSong(); song.Name != "新歌" || song.Duration != 180*time.Second {
		t.Fatalf("CurSong() = %+v after song_changed, want 新歌/180s", song)
	}
	// The progress renderer shows the event's duration: the empty snapshot had
	// no song (zero duration → no progress bar), so "03:00" proves the redrawn
	// view reflects the arriving event.
	connectAssertViewContains(t, app, "03:00")

	// Control forwarding: the wrapper's SetVolume reaches the daemon through
	// Call("volume", {value:33}) and takes effect there.
	n.Player().SetVolume(33)
	waitFor(t, 3*time.Second, func() bool {
		resp, err := client.Call(context.Background(), "volume", nil)
		if err != nil || resp == nil || !resp.Ok {
			return false
		}
		data, _ := resp.Data.(map[string]any)
		v, _ := data["volume"].(float64)
		return v == 33
	})
}

// TestConnectIntegrationMenuChain locks the S6-R1 menu outcome through the
// real integration path: a daemon-connected NewNeteaseRemote plus the main
// menu the shell actually navigates with (the app's CurMenu) yields the full
// browsing chain — the test binary's 14 after-anchor test-double items in
// their production order plus the built-in 帮助 at the tail. The 8 mounted
// business plugins (lastfm excluded) keep the local browsing tree complete in
// the remote shell (roadmap §8.4 browsing rows).
func TestConnectIntegrationMenuChain(t *testing.T) {
	startRemoteDaemonWithEvents(t)
	client := dialRemote(t)
	app, n := newConnectShell(t, client)
	rp := n.Player().remote
	waitRemoteReady(t, rp)

	menu, ok := app.MustMain().CurMenu().(*MainMenu)
	if !ok {
		t.Fatalf("current menu = %T, want *MainMenu", app.MustMain().CurMenu())
	}
	titles := menu.Titles()
	// The test binary's 14 test-double items keep the production browsing
	// chain; 帮助 follows LastFM (last_fm is a direct test-double item here —
	// in production the lastfm plugin is excluded from the connect scope and
	// the built-in 帮助 re-anchors to the tail the same way).
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

// TestConnectIntegrationSearchNavigation locks the S6-R1 search-flow outcome
// through the real integration path: with a daemon-connected shell whose
// search plugin is mounted, ToSearchPage runs the standalone flow — it returns
// the shell search singleton with the type set (search is local, no connect
// degradation toast), and the returned tick cmd flows through the app message
// loop without touching the search API (the tick is a keep-page no-op; the API
// only runs on submit).
func TestConnectIntegrationSearchNavigation(t *testing.T) {
	startRemoteDaemonWithEvents(t)
	client := dialRemote(t)
	app, n := newConnectShell(t, client)
	rp := n.Player().remote
	waitRemoteReady(t, rp)

	page, cmd := n.ToSearchPage(StSingleSong)
	if page == nil || cmd == nil {
		t.Fatalf("connect ToSearchPage = (%v, %v), want (search page, tick cmd)", page, cmd)
	}
	sp, ok := page.(*SearchPage)
	if !ok {
		t.Fatalf("connect ToSearchPage page = %T, want *SearchPage", page)
	}
	if sp != n.search {
		t.Fatal("connect ToSearchPage did not return the shell search singleton")
	}
	if got := sp.SearchType(); got != StSingleSong {
		t.Fatalf("searchType = %v, want StSingleSong", got)
	}

	// Drive the navigation through the app message loop WITHOUT running the
	// search API: the returned cmd yields the tickSearchMsg the loop delivers
	// to the page; the tick keeps the page (no submit → no API), and rendering
	// the page is shell-local.
	msg := cmd()
	if _, ok := msg.(tickSearchMsg); !ok {
		t.Fatalf("ToSearchPage cmd produced %T, want tickSearchMsg", msg)
	}
	_, _ = app.Update(msg)
	if _, pageCmd := sp.Update(msg, app); pageCmd != nil {
		t.Fatalf("search page Update(tickSearchMsg) returned %v, want nil (no search API)", pageCmd)
	}
	view := ansi.Strip(sp.View(app))
	if !strings.Contains(view, "输入关键词") {
		t.Fatalf("search page view does not render the local search input:\n%s", view)
	}
}

// TestConnectIntegrationDisconnect locks the disconnect degradation (D-TC-4):
// closing the daemon server terminates the subscription connection, the Events
// channel closes and consumeEvents marks the shell not ready. Reads keep
// returning the stale cache and a control forward fails gracefully (no panic),
// while the shell still renders.
func TestConnectIntegrationDisconnect(t *testing.T) {
	server := startRemoteDaemonWithEvents(t)
	client := dialRemote(t)
	app, n := newConnectShell(t, client)
	rp := n.Player().remote
	waitRemoteReady(t, rp)

	// server.Close terminates the subscription connection → Events closes →
	// consumeEvents sets ready=false.
	server.Close()
	waitFor(t, 3*time.Second, func() bool { return !rp.Ready() })

	// Degraded path: reads still return the stale cache, a control forward
	// fails fast (logged, the notification is dropped on the terminated
	// program), and the shell keeps rendering (static, not crashed).
	if song := n.Player().CurSong(); song.Id != 0 {
		t.Fatalf("CurSong() = %+v after disconnect, want cached zero song", song)
	}
	n.Player().NextSong(true)
	if rp.Ready() {
		t.Fatal("Ready() = true after disconnect + failed control forward")
	}
	if view := ansi.Strip(app.View().Content); view == "" {
		t.Fatal("shell view empty after disconnect (should keep rendering the stale state)")
	}
}
