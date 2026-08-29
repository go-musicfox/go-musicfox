package core

import (
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	neteaseutil "github.com/go-musicfox/netease-music/util"
	cookiejar "github.com/juju/persistent-cookiejar"

	"github.com/go-musicfox/go-musicfox/internal/composer"
	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/desktop_lyrics"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/lastfm"
	"github.com/go-musicfox/go-musicfox/internal/lyric"
	"github.com/go-musicfox/go-musicfox/internal/player"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/track"
)

// fakeDesktopLyrics is a minimal desktop_lyrics.Controller used only to make
// the service registration deterministic in unit tests.
type fakeDesktopLyrics struct{}

func (fakeDesktopLyrics) Show()                                                        {}
func (fakeDesktopLyrics) Hide()                                                        {}
func (fakeDesktopLyrics) IsVisible() bool                                              { return false }
func (fakeDesktopLyrics) Update(_, _ desktop_lyrics.LyricLine, _ int, _ int64, _ bool) {}
func (fakeDesktopLyrics) UpdateSpectrum(_ player.SpectrumFrame)                        {}
func (fakeDesktopLyrics) UpdateRawSamples(_ player.RawSampleFrame)                     {}
func (fakeDesktopLyrics) SetSpectrumAvailable(_ bool)                                  {}
func (fakeDesktopLyrics) Close()                                                       {}

// testServicePlugins builds the service constructor plugin set with
// config-free fake builders so unit tests can exercise the scope assembly
// without reading configs or starting a real audio player (the production
// builders read configs.AppConfig / storage and NewPlayer starts an engine).
// The plugin Start methods populate the engine service fields.
func testServicePlugins(e *Engine) []framework.Plugin {
	return []framework.Plugin{
		&lastfmPlugin{e: e, build: func() *lastfm.Client { return &lastfm.Client{} }},
		&trackManagerPlugin{e: e, build: func() *track.Manager { return &track.Manager{} }},
		&lyricServicePlugin{e: e, build: func(*track.Manager) *lyric.Service { return &lyric.Service{} }},
		&desktopLyricsPlugin{e: e, build: func() desktop_lyrics.Controller { return fakeDesktopLyrics{} }},
		newLoginServicePlugin(e),
		newUserServicePlugin(e),
		&shareSvcPlugin{e: e, build: func() *composer.ShareService { return &composer.ShareService{} }},
		newEventBusPlugin(e),
		&playerPlugin{e: e, build: func(PlayerOptions) *Player { return NewEmptyPlayer() }},
		newDispatcherPlugin(e),
	}
}

// startTestScope registers the test service plugins on a fresh scope and starts
// it against a fresh context. Callers must not dispose the returned scope (the
// test player is a NewEmptyPlayer whose Close would panic on the nil
// reporter/cancel); registration-order tests only need the started scope.
func startTestScope(e *Engine) (*framework.Scope, *framework.Context) {
	ctx := &framework.Context{}
	scope := framework.NewScope()
	for _, p := range testServicePlugins(e) {
		if err := scope.AddWithEnabled(p, true); err != nil {
			panic(err) // unreachable for a fresh scope
		}
	}
	if err := scope.Start(ctx); err != nil {
		panic(err) // unreachable for the static plugin set with correct ordering
	}
	return scope, ctx
}

// testEngine builds an Engine whose user slot is self-owned so the service
// plugins can capture it. It does not run NewEngine (which starts a real audio
// player and reads configs); the plugin Start methods populate the service
// fields.
func testEngine() *Engine {
	e := &Engine{}
	e.userSlot = &e.user
	return e
}

func TestServicePluginsRegisterExactSet(t *testing.T) {
	_, ctx := startTestScope(testEngine())

	got := ctx.Names()
	sort.Strings(got)
	want := append([]string(nil), registeredServiceNames...)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("registered services = %v, want %v", got, want)
	}
}

func TestServicePluginsRegisterRightConcreteTypes(t *testing.T) {
	e := testEngine()
	_, ctx := startTestScope(e)

	if svc, ok := framework.ServiceOf[*Player](ctx, ServicePlayer); !ok || svc != e.player {
		t.Errorf("ServiceOf(player) = %v, %v; want %T, true", svc, ok, e.player)
	}
	if svc, ok := framework.ServiceOf[*lyric.Service](ctx, ServiceLyricService); !ok || svc != e.lyricService {
		t.Errorf("ServiceOf(lyricService) = %v, %v; want %T, true", svc, ok, e.lyricService)
	}
	if svc, ok := framework.ServiceOf[*track.Manager](ctx, ServiceTrackManager); !ok || svc != e.trackManager {
		t.Errorf("ServiceOf(trackManager) = %v, %v; want %T, true", svc, ok, e.trackManager)
	}
	if _, ok := framework.ServiceOf[desktop_lyrics.Controller](ctx, ServiceDesktopLyrics); !ok {
		t.Error("ServiceOf(desktopLyrics) not resolvable as desktop_lyrics.Controller")
	}
	if svc, ok := framework.ServiceOf[*UserService](ctx, ServiceUserService); !ok || svc == nil {
		t.Errorf("ServiceOf(userService) = %v, %v; want *UserService, true", svc, ok)
	}
	if svc, ok := framework.ServiceOf[*LoginService](ctx, ServiceLoginService); !ok || svc == nil {
		t.Errorf("ServiceOf(loginService) = %v, %v; want *LoginService, true", svc, ok)
	}
	if svc, ok := framework.ServiceOf[*composer.ShareService](ctx, ServiceShareSvc); !ok || svc != e.shareSvc {
		t.Errorf("ServiceOf(shareSvc) = %v, %v; want %T, true", svc, ok, e.shareSvc)
	}
	if svc, ok := framework.ServiceOf[*lastfm.Client](ctx, ServiceLastfm); !ok || svc != e.lastfm {
		t.Errorf("ServiceOf(lastfm) = %v, %v; want %T, true", svc, ok, e.lastfm)
	}
}

func TestUserServiceReflectsLiveUserSlot(t *testing.T) {
	e := testEngine()
	_, ctx := startTestScope(e)

	svc, ok := framework.ServiceOf[*UserService](ctx, ServiceUserService)
	if !ok {
		t.Fatal("userService not resolvable")
	}
	if svc.User != e.UserSlot() {
		t.Fatal("userService.User does not point at the engine user slot")
	}
	if *svc.User != nil {
		t.Fatal("initial user should be nil")
	}

	e.user = &structs.User{Nickname: "tester"}
	if got := *svc.User; got != e.user {
		t.Fatalf("live user = %v, want %v", got, e.user)
	}
}

func TestLoginServiceReflectsLiveCookieJar(t *testing.T) {
	e := testEngine()
	_, ctx := startTestScope(e)

	svc, ok := framework.ServiceOf[*LoginService](ctx, ServiceLoginService)
	if !ok {
		t.Fatal("loginService not resolvable")
	}
	if svc.CookieJar != &appCookieJar {
		t.Fatal("loginService.CookieJar does not point at the appCookieJar slot")
	}
}

// TestLoginServiceInitJarCreatesAndSyncsJar verifies InitJar is invocable
// standalone (as the startup sequence requires: jar 先于 userService 回调):
// it creates a persistent jar, assigns the appCookieJar slot and syncs the
// netease-music global cookie jar.
func TestLoginServiceInitJarCreatesAndSyncsJar(t *testing.T) {
	var jarSlot *cookiejar.Jar
	var userSlot *structs.User
	svc := &LoginService{CookieJar: &jarSlot, User: &userSlot}

	cookiePath := filepath.Join(t.TempDir(), "cookie")
	jar, err := svc.InitJar(cookiePath)
	if err != nil {
		t.Fatalf("InitJar() error = %v", err)
	}
	if jar == nil {
		t.Fatal("InitJar() returned nil jar")
	}
	if jarSlot != jar {
		t.Fatal("appCookieJar slot was not assigned")
	}
	if got := neteaseutil.GetGlobalCookieJar(); got != http.CookieJar(jar) {
		t.Fatalf("global cookie jar not synced: got %v, want %v", got, jar)
	}
}

// TestLoginServiceInitJarBacksUpCorruptFile verifies the corrupt-file
// backup/reset logic: a corrupt cookie file is renamed to a .bak.<ts> backup
// and a fresh jar is created at the original path.
func TestLoginServiceInitJarBacksUpCorruptFile(t *testing.T) {
	var jarSlot *cookiejar.Jar
	var userSlot *structs.User
	svc := &LoginService{CookieJar: &jarSlot, User: &userSlot}

	cookiePath := filepath.Join(t.TempDir(), "cookie")
	if err := os.WriteFile(cookiePath, []byte("this is not a valid cookie jar"), 0600); err != nil {
		t.Fatalf("write corrupt cookie file: %v", err)
	}

	jar, err := svc.InitJar(cookiePath)
	if err != nil {
		t.Fatalf("InitJar() on corrupt file error = %v", err)
	}
	if jar == nil {
		t.Fatal("InitJar() returned nil jar after corrupt reset")
	}
	if jarSlot != jar {
		t.Fatal("appCookieJar slot was not assigned after corrupt reset")
	}

	backups, err := filepath.Glob(cookiePath + ".bak.*")
	if err != nil || len(backups) != 1 {
		t.Fatalf("expected exactly one .bak backup, got %v (err %v)", backups, err)
	}
}

// seqRecorder records startup events in call order so tests can assert the
// startup order constraints (3.1.3).
type seqRecorder struct {
	events []string
}

func (r *seqRecorder) record(ev string) {
	r.events = append(r.events, ev)
}

func (r *seqRecorder) indexOf(ev string) int {
	for i, e := range r.events {
		if e == ev {
			return i
		}
	}
	return -1
}

// TestJarInitPrecedesUserCallback asserts the cross-service order constraint:
// loginService.InitJar (jar 初始化) MUST run before the userService callback
// (LoginCallback persists the jar via appCookieJar.Save()). The sequence is
// InitJar → LoadFromStorage → LoginWithCookie, and the recorded order must put
// "jar-init" before "user-callback".
func TestJarInitPrecedesUserCallback(t *testing.T) {
	previousConfig := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	t.Cleanup(func() { configs.AppConfig = previousConfig })

	rec := &seqRecorder{}
	var jarSlot *cookiejar.Jar
	var userSlot *structs.User

	loginSvc := &LoginService{CookieJar: &jarSlot, User: &userSlot}
	userSvc := &UserService{
		User:           &userSlot,
		Jar:            &jarSlot,
		Login:          func() error { rec.record("user-callback"); return nil },
		loadStoredUser: func() (*structs.User, bool) { return nil, false },
		refreshJar:     func() (*cookiejar.Jar, error) { return jarSlot, nil },
	}

	// A cookie in the env drives LoginWithCookie through the cookie-login path,
	// reaching the Login callback after a (stubbed) successful token refresh.
	t.Setenv("MUSICFOX_COOKIE", "MUSIC_U=test-cookie")

	jar, err := loginSvc.InitJar(filepath.Join(t.TempDir(), "cookie"))
	if err != nil {
		t.Fatalf("InitJar() error = %v", err)
	}
	rec.record("jar-init")

	userSvc.LoadFromStorage(jar)
	userSvc.LoginWithCookie(jar, filepath.Join(t.TempDir(), "cookie"))

	jarIdx := rec.indexOf("jar-init")
	cbIdx := rec.indexOf("user-callback")
	if jarIdx == -1 {
		t.Fatalf("jar-init was not recorded; events=%v", rec.events)
	}
	if cbIdx == -1 {
		t.Fatalf("user-callback was not recorded (LoginWithCookie must reach the callback); events=%v", rec.events)
	}
	if jarIdx >= cbIdx {
		t.Fatalf("order constraint violated: jar-init(%d) must precede user-callback(%d); events=%v", jarIdx, cbIdx, rec.events)
	}
}

// TestUserServiceTouristMode verifies the guest path: without a stored user and
// without a cookie, LoadFromStorage + LoginWithCookie leave the user slot nil,
// never invoke the Login callback and never panic.
func TestUserServiceTouristMode(t *testing.T) {
	previousConfig := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	t.Cleanup(func() { configs.AppConfig = previousConfig })

	rec := &seqRecorder{}
	var jarSlot *cookiejar.Jar
	var userSlot *structs.User

	loginSvc := &LoginService{CookieJar: &jarSlot, User: &userSlot}
	userSvc := &UserService{
		User:           &userSlot,
		Jar:            &jarSlot,
		Login:          func() error { rec.record("user-callback"); return nil },
		loadStoredUser: func() (*structs.User, bool) { return nil, false },
		refreshJar:     func() (*cookiejar.Jar, error) { return jarSlot, nil },
	}

	t.Setenv("MUSICFOX_COOKIE", "") // no cookie → tourist mode

	jar, err := loginSvc.InitJar(filepath.Join(t.TempDir(), "cookie"))
	if err != nil {
		t.Fatalf("InitJar() error = %v", err)
	}

	userSvc.LoadFromStorage(jar)
	userSvc.LoginWithCookie(jar, filepath.Join(t.TempDir(), "cookie"))

	if userSlot != nil {
		t.Fatalf("tourist mode: user should stay nil, got %v", userSlot)
	}
	if cbIdx := rec.indexOf("user-callback"); cbIdx != -1 {
		t.Fatalf("tourist mode: Login callback must not fire, got event at %d; events=%v", cbIdx, rec.events)
	}
}
