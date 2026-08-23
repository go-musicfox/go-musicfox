package ui

import (
	"reflect"
	"sort"
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/composer"
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

// testNetease builds a Netease whose fields are all non-nil so registerServices
// registers every startup service with the right concrete type.
func testNetease() *Netease {
	return &Netease{
		player:        &Player{},
		lyricService:  &lyric.Service{},
		trackManager:  &track.Manager{},
		desktopLyrics: fakeDesktopLyrics{},
		coverRenderer: &CoverRenderer{},
		shareSvc:      &composer.ShareService{},
		lastfm:        &lastfm.Client{},
	}
}

func TestRegisterServicesRegistersExactSet(t *testing.T) {
	ctx := &framework.Context{}
	if err := registerServices(ctx, testNetease()); err != nil {
		t.Fatalf("registerServices() error = %v", err)
	}

	got := ctx.Names()
	sort.Strings(got)
	want := append([]string(nil), registeredServiceNames...)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("registered services = %v, want %v", got, want)
	}
}

func TestRegisterServicesRegistersRightConcreteTypes(t *testing.T) {
	ctx := &framework.Context{}
	n := testNetease()
	if err := registerServices(ctx, n); err != nil {
		t.Fatalf("registerServices() error = %v", err)
	}

	if svc, ok := framework.ServiceOf[*Player](ctx, ServicePlayer); !ok || svc != n.player {
		t.Errorf("ServiceOf(player) = %v, %v; want %T, true", svc, ok, n.player)
	}
	if svc, ok := framework.ServiceOf[*lyric.Service](ctx, ServiceLyricService); !ok || svc != n.lyricService {
		t.Errorf("ServiceOf(lyricService) = %v, %v; want %T, true", svc, ok, n.lyricService)
	}
	if svc, ok := framework.ServiceOf[*track.Manager](ctx, ServiceTrackManager); !ok || svc != n.trackManager {
		t.Errorf("ServiceOf(trackManager) = %v, %v; want %T, true", svc, ok, n.trackManager)
	}
	if _, ok := framework.ServiceOf[desktop_lyrics.Controller](ctx, ServiceDesktopLyrics); !ok {
		t.Error("ServiceOf(desktopLyrics) not resolvable as desktop_lyrics.Controller")
	}
	if svc, ok := framework.ServiceOf[*CoverRenderer](ctx, ServiceCoverRenderer); !ok || svc != n.coverRenderer {
		t.Errorf("ServiceOf(coverRenderer) = %v, %v; want %T, true", svc, ok, n.coverRenderer)
	}
	if svc, ok := framework.ServiceOf[*UserService](ctx, ServiceUserService); !ok || svc == nil {
		t.Errorf("ServiceOf(userService) = %v, %v; want *UserService, true", svc, ok)
	}
	if svc, ok := framework.ServiceOf[*LoginService](ctx, ServiceLoginService); !ok || svc == nil {
		t.Errorf("ServiceOf(loginService) = %v, %v; want *LoginService, true", svc, ok)
	}
	if svc, ok := framework.ServiceOf[*composer.ShareService](ctx, ServiceShareSvc); !ok || svc != n.shareSvc {
		t.Errorf("ServiceOf(shareSvc) = %v, %v; want %T, true", svc, ok, n.shareSvc)
	}
	if svc, ok := framework.ServiceOf[*lastfm.Client](ctx, ServiceLastfm); !ok || svc != n.lastfm {
		t.Errorf("ServiceOf(lastfm) = %v, %v; want %T, true", svc, ok, n.lastfm)
	}
}

func TestUserServiceReflectsLiveUserSlot(t *testing.T) {
	ctx := &framework.Context{}
	n := testNetease()
	if err := registerServices(ctx, n); err != nil {
		t.Fatalf("registerServices() error = %v", err)
	}

	svc, ok := framework.ServiceOf[*UserService](ctx, ServiceUserService)
	if !ok {
		t.Fatal("userService not resolvable")
	}
	if svc.User != &n.user {
		t.Fatal("userService.User does not point at the Netease user slot")
	}
	if *svc.User != nil {
		t.Fatal("initial user should be nil")
	}

	n.user = &structs.User{Nickname: "tester"}
	if got := *svc.User; got != n.user {
		t.Fatalf("live user = %v, want %v", got, n.user)
	}
}

func TestLoginServiceReflectsLiveCookieJar(t *testing.T) {
	ctx := &framework.Context{}
	n := testNetease()
	if err := registerServices(ctx, n); err != nil {
		t.Fatalf("registerServices() error = %v", err)
	}

	svc, ok := framework.ServiceOf[*LoginService](ctx, ServiceLoginService)
	if !ok {
		t.Fatal("loginService not resolvable")
	}
	if svc.CookieJar != &appCookieJar {
		t.Fatal("loginService.CookieJar does not point at the appCookieJar slot")
	}
}
