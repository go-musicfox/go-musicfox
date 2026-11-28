package track

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/netease"
)

type lyricFetcherStub struct {
	regular      structs.LRCData
	cloud        structs.LRCData
	cloudErr     error
	userID       int64
	songID       int64
	regularCalls int
	cloudCalls   int
}

func (f lyricFetcherStub) FetchPlayableInfo(context.Context, int64) (*netease.PlayableInfo, error) {
	return nil, nil
}

func (f lyricFetcherStub) FetchStream(context.Context, PlayableSource) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}

func (f *lyricFetcherStub) FetchLyric(context.Context, int64) (structs.LRCData, error) {
	f.regularCalls++
	return f.regular, nil
}

func (f *lyricFetcherStub) FetchCloudLyric(_ context.Context, userID, songID int64) (structs.LRCData, error) {
	f.cloudCalls++
	f.userID = userID
	f.songID = songID
	return f.cloud, f.cloudErr
}

func TestGetLyricUsesEmbeddedLyricsForUnmatchedCloudSong(t *testing.T) {
	fetcher := &lyricFetcherStub{
		regular: structs.LRCData{Original: "[00:00.00]regular"},
		cloud:   structs.LRCData{Original: "[00:00.00]embedded"},
	}
	manager := &Manager{fetcher: fetcher}
	manager.SetCloudUserID(54321)
	manager.SetCloudUserID(12345)

	got, err := manager.GetLyric(context.Background(), structs.Song{
		Id:        42,
		UnMatched: true,
	})
	if err != nil {
		t.Fatalf("get lyric: %v", err)
	}
	if got.Original != "[00:00.00]embedded" {
		t.Fatalf("original lyric = %q, want embedded cloud lyric", got.Original)
	}
	if fetcher.userID != 12345 || fetcher.songID != 42 {
		t.Fatalf("cloud lyric IDs = (%d, %d), want (12345, 42)", fetcher.userID, fetcher.songID)
	}
	if fetcher.cloudCalls != 1 || fetcher.regularCalls != 0 {
		t.Fatalf("fetch calls = (cloud %d, regular %d), want (1, 0)", fetcher.cloudCalls, fetcher.regularCalls)
	}
}

func TestGetLyricUsesRegularLyricsWithoutCloudUser(t *testing.T) {
	manager := &Manager{fetcher: &lyricFetcherStub{
		regular: structs.LRCData{Original: "[00:00.00]regular"},
		cloud:   structs.LRCData{Original: "[00:00.00]embedded"},
	}}

	got, err := manager.GetLyric(context.Background(), structs.Song{
		Id:        42,
		UnMatched: true,
	})
	if err != nil {
		t.Fatalf("get lyric: %v", err)
	}
	if got.Original != "[00:00.00]regular" {
		t.Fatalf("original lyric = %q, want regular lyric without current cloud user", got.Original)
	}
}

func TestGetLyricFallsBackWhenEmbeddedCloudLyricsAreUnavailable(t *testing.T) {
	fetcher := &lyricFetcherStub{
		regular:  structs.LRCData{Original: "[00:00.00]regular"},
		cloudErr: errors.New("no embedded lyric"),
	}
	manager := &Manager{fetcher: fetcher}
	manager.SetCloudUserID(12345)

	got, err := manager.GetLyric(context.Background(), structs.Song{
		Id:        42,
		UnMatched: true,
	})
	if err != nil {
		t.Fatalf("get lyric: %v", err)
	}
	if got.Original != "[00:00.00]regular" {
		t.Fatalf("original lyric = %q, want regular lyric fallback", got.Original)
	}
	if fetcher.cloudCalls != 1 || fetcher.regularCalls != 1 {
		t.Fatalf("fetch calls = (cloud %d, regular %d), want (1, 1)", fetcher.cloudCalls, fetcher.regularCalls)
	}
}

func TestGetLyricUsesCloudLyricsForLegacyUnmatchedSnapshot(t *testing.T) {
	fetcher := &lyricFetcherStub{
		regular: structs.LRCData{Original: "[00:00.00]regular"},
		cloud:   structs.LRCData{Original: "[00:00.00]embedded"},
	}
	manager := &Manager{fetcher: fetcher}
	manager.SetCloudUserID(12345)

	got, err := manager.GetLyric(context.Background(), structs.Song{
		Id:      42,
		Album:   structs.Album{Id: 0},
		Artists: []structs.Artist{{Id: 0, Name: "artist"}},
	})
	if err != nil {
		t.Fatalf("get lyric: %v", err)
	}
	if got.Original != "[00:00.00]embedded" {
		t.Fatalf("original lyric = %q, want embedded cloud lyric", got.Original)
	}
	if fetcher.cloudCalls != 1 || fetcher.regularCalls != 0 {
		t.Fatalf("fetch calls = (cloud %d, regular %d), want (1, 0)", fetcher.cloudCalls, fetcher.regularCalls)
	}
}

func TestGetLyricKeepsRegularLyricsForMatchedCloudSong(t *testing.T) {
	fetcher := &lyricFetcherStub{
		regular: structs.LRCData{Original: "[00:00.00]regular"},
		cloud:   structs.LRCData{Original: "[00:00.00]embedded"},
	}
	manager := &Manager{fetcher: fetcher}
	manager.SetCloudUserID(12345)

	got, err := manager.GetLyric(context.Background(), structs.Song{
		Id: 186001,
	})
	if err != nil {
		t.Fatalf("get lyric: %v", err)
	}
	if got.Original != "[00:00.00]regular" {
		t.Fatalf("original lyric = %q, want regular lyric", got.Original)
	}
	if fetcher.cloudCalls != 0 || fetcher.regularCalls != 1 {
		t.Fatalf("fetch calls = (cloud %d, regular %d), want (0, 1)", fetcher.cloudCalls, fetcher.regularCalls)
	}
}

// blockingLyricFetcher simulates an SDK request that cannot be interrupted.
type blockingLyricFetcher struct {
	*lyricFetcherStub
	entered  chan context.Context
	release  chan struct{}
	finished chan struct{}
}

func (f *blockingLyricFetcher) FetchLyric(ctx context.Context, _ int64) (structs.LRCData, error) {
	f.entered <- ctx
	<-f.release
	close(f.finished)
	return structs.LRCData{Original: "[00:00.00]shared"}, nil
}

func TestGetLyricCancellationStopsWaiting(t *testing.T) {
	f := &blockingLyricFetcher{
		lyricFetcherStub: &lyricFetcherStub{},
		entered:          make(chan context.Context, 1),
		release:          make(chan struct{}),
		finished:         make(chan struct{}),
	}
	defer close(f.release)
	m := &Manager{fetcher: f}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := m.GetLyric(ctx, structs.Song{Id: 42}); done <- err }()
	var sharedCtx context.Context
	select {
	case sharedCtx = <-f.entered:
	case <-time.After(2 * time.Second):
		t.Fatal("request did not start")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("cancellation did not stop waiting")
	}
	if sharedCtx.Err() != nil {
		t.Fatalf("caller canceled shared request: %v", sharedCtx.Err())
	}
	select {
	case <-f.finished:
		t.Fatal("request should still be blocked")
	default:
	}
}

func TestGetLyricAlreadyCanceledDoesNotFetch(t *testing.T) {
	f := &lyricFetcherStub{}
	m := &Manager{fetcher: f}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := m.GetLyric(ctx, structs.Song{Id: 42})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v", err)
	}
	if f.regularCalls != 0 || f.cloudCalls != 0 {
		t.Fatal("canceled request fetched lyrics")
	}
}
