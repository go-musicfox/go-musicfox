package playermiddleware

import (
	"context"
	"errors"
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/player"
)

// withUNMConfig applies a fresh UNM config for the test and restores the
// previous global config afterwards.
func withUNMConfig(t *testing.T, proxyURL string, skipInvalid bool) {
	t.Helper()
	previous := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	configs.AppConfig.UNM.ProxyURL = proxyURL
	configs.AppConfig.UNM.SkipInvalidTracks = skipInvalid
	t.Cleanup(func() { configs.AppConfig = previous })
}

func TestUNMProxyURLDoesNotAffectBannedInterception(t *testing.T) {
	withUNMConfig(t, "http://127.0.0.1:8080", true)
	c := NewChain().Use(NewUNMMiddleware())
	// A configured UNM proxy only disables the SDK engine (util/config.go);
	// the SkipInvalidTracks banned-path interception still applies, exactly
	// like today's player.go behavior.
	m := testURLMusic("https://music.163.com/resource/n2/73/84/3759149332.mp3")
	err := c.Execute(context.Background(), m)
	if !errors.Is(err, ErrBlockedTrack) {
		t.Fatalf("expected ErrBlockedTrack despite proxy config, got %v", err)
	}
}

func TestUNMBannedPathInterception(t *testing.T) {
	withUNMConfig(t, "", true)
	c := NewChain().Use(NewUNMMiddleware())
	// bannedLinkFeatures in utils/netease/filter.go blocks URLs whose path
	// ends with /resource/n2/73/84/3759149332.mp3.
	m := testURLMusic("https://music.163.com/resource/n2/73/84/3759149332.mp3")
	err := c.Execute(context.Background(), m)
	if !errors.Is(err, ErrBlockedTrack) {
		t.Fatalf("expected ErrBlockedTrack for banned path, got %v", err)
	}
}

func TestUNMBypassWhenSkipInvalidTracksOff(t *testing.T) {
	withUNMConfig(t, "", false)
	c := NewChain().Use(NewUNMMiddleware())
	m := testURLMusic("https://music.163.com/resource/n2/73/84/3759149332.mp3")
	if err := c.Execute(context.Background(), m); err != nil {
		t.Fatalf("expected pass-through when SkipInvalidTracks is off, got %v", err)
	}
}

func TestUNMNormalURLPasses(t *testing.T) {
	withUNMConfig(t, "", true)
	c := NewChain().Use(NewUNMMiddleware())
	m := testURLMusic("https://m801.music.126.net/abc/def.mp3")
	if err := c.Execute(context.Background(), m); err != nil {
		t.Fatalf("expected pass-through for a normal URL, got %v", err)
	}
	if m.URL != "https://m801.music.126.net/abc/def.mp3" {
		t.Fatalf("URL must stay unchanged, got %q", m.URL)
	}
}

func TestUNMNextCalledOnSuccess(t *testing.T) {
	withUNMConfig(t, "", false)
	var nextCalled bool
	c := NewChain().Use(NewUNMMiddleware(), func(_ context.Context, m *player.URLMusic, next MiddlewareNext) error {
		nextCalled = true
		return next(context.Background(), m)
	})
	if err := c.Execute(context.Background(), testURLMusic("https://m801.music.126.net/abc/def.mp3")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !nextCalled {
		t.Fatal("next middleware was not called")
	}
}

func TestUNMReceivesRewrittenURLFromUpstream(t *testing.T) {
	withUNMConfig(t, "", false)
	// A rewriting middleware upstream hands a new URLMusic to next; the UNM
	// middleware must operate on that propagated value, not a stale one.
	c := NewChain().Use(
		func(_ context.Context, _ *player.URLMusic, next MiddlewareNext) error {
			return next(context.Background(), testURLMusic("https://m801.music.126.net/abc/def.mp3"))
		},
		NewUNMMiddleware(),
		func(_ context.Context, m *player.URLMusic, next MiddlewareNext) error {
			if m.URL != "https://m801.music.126.net/abc/def.mp3" {
				t.Fatalf("UNM passed a wrong URLMusic downstream, got %q", m.URL)
			}
			return next(context.Background(), m)
		},
	)
	if err := c.Execute(context.Background(), testURLMusic("https://original.example.com/x.mp3")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
