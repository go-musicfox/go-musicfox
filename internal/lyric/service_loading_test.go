package lyric

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/structs"
)

type lyricFetchFunc func(context.Context, structs.Song) (structs.LRCData, error)

func (f lyricFetchFunc) GetLyric(ctx context.Context, song structs.Song) (structs.LRCData, error) {
	return f(ctx, song)
}

func awaitLoad(t *testing.T, done <-chan error) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-time.After(2 * time.Second):
		t.Fatal("lyric operation blocked")
		return nil
	}
}

func TestStateRemainsResponsiveWhileLoading(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	defer close(release)
	s := NewService(lyricFetchFunc(func(context.Context, structs.Song) (structs.LRCData, error) {
		close(entered)
		<-release
		return structs.LRCData{Original: "[00:00.00]loaded"}, nil
	}), false, 0, false)
	done := make(chan error, 1)
	go func() { done <- s.SetSong(context.Background(), structs.Song{Id: 1}) }()
	<-entered
	responsive := make(chan error, 1)
	go func() {
		s.State()
		s.UpdatePosition(time.Second)
		s.SetOffset(time.Second)
		s.EnableTranslation(true)
		responsive <- nil
	}()
	if err := awaitLoad(t, responsive); err != nil {
		t.Fatal(err)
	}
}

func TestLateLyricsCannotReplaceNewerRequest(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer unblock()
	calls := 0
	s := NewService(lyricFetchFunc(func(context.Context, structs.Song) (structs.LRCData, error) {
		calls++
		if calls == 1 {
			close(entered)
			<-release
			return structs.LRCData{Original: "[00:00.00]old"}, nil
		}
		return structs.LRCData{Original: "[00:00.00]new"}, nil
	}), false, 0, false)
	oldDone := make(chan error, 1)
	go func() { oldDone <- s.SetSong(context.Background(), structs.Song{Id: 1}) }()
	<-entered
	newDone := make(chan error, 1)
	// Repeat the same song ID: identity alone must not authorize an old result.
	go func() { newDone <- s.SetSong(context.Background(), structs.Song{Id: 1}) }()
	if err := awaitLoad(t, newDone); err != nil {
		t.Fatal(err)
	}
	if got := s.State().Fragments[0].Content; got != "new" {
		t.Fatalf("lyric = %q", got)
	}
	// The old request must not publish even if its fetcher ignores cancellation.
	unblock()
	if err := awaitLoad(t, oldDone); !errors.Is(err, context.Canceled) {
		t.Fatalf("old request error = %v", err)
	}
	if got := s.State().Fragments[0].Content; got != "new" {
		t.Fatalf("late result replaced lyrics: %q", got)
	}
}

func TestStopCancelsPendingLyrics(t *testing.T) {
	entered := make(chan struct{})
	s := NewService(lyricFetchFunc(func(ctx context.Context, _ structs.Song) (structs.LRCData, error) {
		close(entered)
		<-ctx.Done()
		return structs.LRCData{Original: "[00:00.00]stale"}, nil
	}), false, 0, false)
	done := make(chan error, 1)
	go func() { done <- s.SetSong(context.Background(), structs.Song{Id: 1}) }()
	<-entered
	stopped := make(chan error, 1)
	go func() { s.Stop(); stopped <- nil }()
	if err := awaitLoad(t, stopped); err != nil {
		t.Fatal(err)
	}
	if err := awaitLoad(t, done); !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v", err)
	}
	if state := s.State(); state.IsRunning || len(state.Fragments) != 0 {
		t.Fatalf("stopped state = %+v", state)
	}
}

func TestLoadSongResetsSynchronouslyAndUsesCurrentOptions(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer unblock()
	s := NewService(lyricFetchFunc(func(_ context.Context, song structs.Song) (structs.LRCData, error) {
		if song.Id == 2 {
			close(entered)
			<-release
		}
		return structs.LRCData{Original: "[00:00.00]original", Translated: "[00:00.00]translated"}, nil
	}), false, 0, false)
	if err := s.SetSong(context.Background(), structs.Song{Id: 1}); err != nil {
		t.Fatal(err)
	}
	s.LoadSong(context.Background(), structs.Song{Id: 2})
	if state := s.State(); state.IsRunning || len(state.Fragments) != 0 {
		t.Fatalf("old lyrics visible after starting load: %+v", state)
	}
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("asynchronous load did not start")
	}
	s.EnableTranslation(true)
	s.SetOffset(250 * time.Millisecond)
	unblock()
	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		state := s.State()
		if state.IsRunning {
			if state.TranslatedFragments[0] != "translated" || state.OffsetMs != 250 || !state.ShowTranslation {
				t.Fatalf("load discarded current options: %+v", state)
			}
			return
		}
		select {
		case <-deadline.C:
			t.Fatal("asynchronous load did not publish")
		case <-ticker.C:
		}
	}
}

func TestCanceledSongDoesNotResetCurrentLyrics(t *testing.T) {
	s := NewService(lyricFetchFunc(func(context.Context, structs.Song) (structs.LRCData, error) {
		return structs.LRCData{Original: "[00:00.00]current"}, nil
	}), false, 0, false)
	if err := s.SetSong(context.Background(), structs.Song{Id: 1}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.SetSong(ctx, structs.Song{Id: 2}); !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v", err)
	}
	s.LoadSong(ctx, structs.Song{Id: 2})
	if state := s.State(); !state.IsRunning || len(state.Fragments) != 1 || state.Fragments[0].Content != "current" {
		t.Fatalf("canceled load reset current lyrics: %+v", state)
	}
}
