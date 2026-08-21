package playermiddleware

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/player"
	"github.com/go-musicfox/go-musicfox/internal/structs"
)

func testURLMusic(url string) *player.URLMusic {
	return &player.URLMusic{URL: url, Song: structs.Song{}, Type: player.Mp3}
}

func TestEmptyChainIsNoop(t *testing.T) {
	c := NewChain()
	if c.Len() != 0 {
		t.Fatalf("expected 0 middleware, got %d", c.Len())
	}
	if err := c.Execute(context.Background(), testURLMusic("https://example.com/a.mp3")); err != nil {
		t.Fatalf("empty chain should not error, got %v", err)
	}
}

func TestChainRunsInOrder(t *testing.T) {
	var order []string
	c := NewChain()
	c.Use(
		func(_ context.Context, m *player.URLMusic, next MiddlewareNext) error {
			order = append(order, "first-in")
			err := next(context.Background(), m)
			order = append(order, "first-out")
			return err
		},
		func(_ context.Context, m *player.URLMusic, next MiddlewareNext) error {
			order = append(order, "second-in")
			err := next(context.Background(), m)
			order = append(order, "second-out")
			return err
		},
	)
	if err := c.Execute(context.Background(), testURLMusic("https://example.com/a.mp3")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "first-in,second-in,second-out,first-out"
	if got := strings.Join(order, ","); got != want {
		t.Fatalf("order mismatch: got %q want %q", got, want)
	}
}

func TestChainMiddlewareCanRewriteURL(t *testing.T) {
	c := NewChain()
	c.Use(func(_ context.Context, m *player.URLMusic, next MiddlewareNext) error {
		m.URL = "https://rewritten.example.com/x.mp3"
		return next(context.Background(), m)
	})
	m := testURLMusic("https://original.example.com/x.mp3")
	if err := c.Execute(context.Background(), m); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.URL != "https://rewritten.example.com/x.mp3" {
		t.Fatalf("URL was not rewritten, got %q", m.URL)
	}
}

func TestChainErrorAbortsAndPropagates(t *testing.T) {
	boom := errors.New("boom")
	var laterCalled bool
	c := NewChain()
	c.Use(func(_ context.Context, m *player.URLMusic, next MiddlewareNext) error {
		return boom
	}, func(_ context.Context, m *player.URLMusic, next MiddlewareNext) error {
		laterCalled = true
		return next(context.Background(), m)
	})
	err := c.Execute(context.Background(), testURLMusic("https://example.com/a.mp3"))
	if !errors.Is(err, boom) {
		t.Fatalf("expected boom, got %v", err)
	}
	if laterCalled {
		t.Fatal("later middleware should not run after an error")
	}
}

func TestChainContextPassthrough(t *testing.T) {
	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "value")
	var seen any
	c := NewChain()
	c.Use(func(c context.Context, m *player.URLMusic, next MiddlewareNext) error {
		seen = c.Value(ctxKey{})
		return next(c, m)
	})
	if err := c.Execute(ctx, testURLMusic("https://example.com/a.mp3")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if seen != "value" {
		t.Fatalf("context value not propagated, got %v", seen)
	}
}
