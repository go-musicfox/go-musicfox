package lyric

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/structs"
)

func TestStateIncludesLyricOffset(t *testing.T) {
	service := NewService(nil, false, 250*time.Millisecond, false)
	if got, want := service.State().OffsetMs, int64(250); got != want {
		t.Errorf("initial offset = %dms, want %dms", got, want)
	}

	service.SetOffset(-100 * time.Millisecond)
	if got, want := service.State().OffsetMs, int64(-100); got != want {
		t.Errorf("updated offset = %dms, want %dms", got, want)
	}
}

type lyricFetchResult struct {
	loaded bool
	err    error
}

type controlledLyricFetcher struct {
	started  chan int64
	releases map[int64]chan struct{}
}

func (f *controlledLyricFetcher) GetLyric(ctx context.Context, song structs.Song) (structs.LRCData, error) {
	select {
	case f.started <- song.Id:
	case <-ctx.Done():
		return structs.LRCData{}, ctx.Err()
	}
	select {
	case <-f.releases[song.Id]:
		return structs.LRCData{Original: fmt.Sprintf("[00:00.00]song-%d", song.Id)}, nil
	case <-ctx.Done():
		return structs.LRCData{}, ctx.Err()
	}
}

func TestSetSongDoesNotBlockStateWhileFetching(t *testing.T) {
	release := make(chan struct{})
	fetcher := &controlledLyricFetcher{
		started:  make(chan int64, 1),
		releases: map[int64]chan struct{}{1: release},
	}
	service := NewService(fetcher, false, 0, false)
	loadID := service.BeginSong(1)
	result := make(chan lyricFetchResult, 1)
	go func() {
		loaded, err := service.SetSong(context.Background(), structs.Song{Id: 1}, loadID)
		result <- lyricFetchResult{loaded: loaded, err: err}
	}()

	awaitStartedSong(t, fetcher.started, 1)
	stateRead := make(chan struct{})
	go func() {
		_ = service.State()
		close(stateRead)
	}()
	select {
	case <-stateRead:
	case <-time.After(time.Second):
		close(release)
		t.Fatal("State blocked while lyrics were being fetched")
	}

	close(release)
	got := <-result
	if got.err != nil || !got.loaded {
		t.Fatalf("SetSong() = loaded %v, error %v; want loaded true", got.loaded, got.err)
	}
}

func TestSetSongDiscardsOutdatedLoad(t *testing.T) {
	releaseFirst := make(chan struct{})
	releaseSecond := make(chan struct{})
	fetcher := &controlledLyricFetcher{
		started: make(chan int64, 2),
		releases: map[int64]chan struct{}{
			1: releaseFirst,
			2: releaseSecond,
		},
	}
	service := NewService(fetcher, false, 0, false)

	firstLoadID := service.BeginSong(1)
	firstResult := make(chan lyricFetchResult, 1)
	go func() {
		loaded, err := service.SetSong(context.Background(), structs.Song{Id: 1}, firstLoadID)
		firstResult <- lyricFetchResult{loaded: loaded, err: err}
	}()
	awaitStartedSong(t, fetcher.started, 1)

	secondLoadID := service.BeginSong(2)
	if state := service.State(); state.CurrentSongID != 0 || state.IsRunning {
		t.Fatalf("state after BeginSong() = song %d, running %v; want cleared", state.CurrentSongID, state.IsRunning)
	}
	secondResult := make(chan lyricFetchResult, 1)
	go func() {
		loaded, err := service.SetSong(context.Background(), structs.Song{Id: 2}, secondLoadID)
		secondResult <- lyricFetchResult{loaded: loaded, err: err}
	}()
	awaitStartedSong(t, fetcher.started, 2)

	close(releaseSecond)
	if got := <-secondResult; got.err != nil || !got.loaded {
		t.Fatalf("latest SetSong() = loaded %v, error %v; want loaded true", got.loaded, got.err)
	}
	close(releaseFirst)
	if got := <-firstResult; got.err != nil || got.loaded {
		t.Fatalf("outdated SetSong() = loaded %v, error %v; want loaded false", got.loaded, got.err)
	}

	state := service.State()
	if state.CurrentSongID != 2 {
		t.Fatalf("current song ID = %d, want 2", state.CurrentSongID)
	}
	if lrc := state.FormatAsLRC(); !strings.Contains(lrc, "song-2") || strings.Contains(lrc, "song-1") {
		t.Fatalf("current lyrics = %q, want only song-2", lrc)
	}
}

func awaitStartedSong(t *testing.T, started <-chan int64, want int64) {
	t.Helper()
	select {
	case got := <-started:
		if got != want {
			t.Fatalf("started song ID = %d, want %d", got, want)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for song %d fetch", want)
	}
}
