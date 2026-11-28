package ui

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/lyric"
	playerpkg "github.com/go-musicfox/go-musicfox/internal/player"
	"github.com/go-musicfox/go-musicfox/internal/playlist"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

func TestPeekGaplessSongDeterministicModes(t *testing.T) {
	p := &Player{playlistManager: playlist.NewPlaylistManager()}
	songs := []structs.Song{{Id: 1}, {Id: 2}, {Id: 3}}
	if err := p.playlistManager.Initialize(1, songs); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		mode types.Mode
		want int64
		ok   bool
	}{
		{types.PmOrdered, 3, true},
		{types.PmListLoop, 3, true},
		{types.PmSingleLoop, 2, true},
		{types.PmListRandom, 0, false},
		{types.PmInfRandom, 0, false},
		{types.PmIntelligent, 0, false},
	}
	for _, test := range tests {
		if err := p.playlistManager.SetPlayMode(test.mode); err != nil {
			t.Fatal(err)
		}
		got, ok := p.peekGaplessSong()
		if ok != test.ok || got.Id != test.want {
			t.Errorf("mode %v: got id=%d ok=%v, want id=%d ok=%v", test.mode, got.Id, ok, test.want, test.ok)
		}
	}
}

func TestPeekGaplessSongWrapsListLoop(t *testing.T) {
	p := &Player{playlistManager: playlist.NewPlaylistManager()}
	if err := p.playlistManager.Initialize(1, []structs.Song{{Id: 1}, {Id: 2}}); err != nil {
		t.Fatal(err)
	}
	if err := p.playlistManager.SetPlayMode(types.PmListLoop); err != nil {
		t.Fatal(err)
	}
	got, ok := p.peekGaplessSong()
	if !ok || got.Id != 1 {
		t.Fatalf("got id=%d ok=%v, want id=1 ok=true", got.Id, ok)
	}
}

type staticLyricFetcher struct{}

func (staticLyricFetcher) GetLyric(_ context.Context, _ structs.Song) (structs.LRCData, error) {
	return structs.LRCData{Original: "[00:00.00]lyrics-for-song-1"}, nil
}

type playingInfoPlayer struct{}

func (*playingInfoPlayer) Play(playerpkg.URLMusic)        {}
func (*playingInfoPlayer) CurMusic() playerpkg.URLMusic   { return playerpkg.URLMusic{} }
func (*playingInfoPlayer) Pause()                         {}
func (*playingInfoPlayer) Resume()                        {}
func (*playingInfoPlayer) Stop()                          {}
func (*playingInfoPlayer) Toggle()                        {}
func (*playingInfoPlayer) Seek(time.Duration)             {}
func (*playingInfoPlayer) PassedTime() time.Duration      { return 0 }
func (*playingInfoPlayer) PlayedTime() time.Duration      { return 0 }
func (*playingInfoPlayer) TimeChan() <-chan time.Duration { return nil }
func (*playingInfoPlayer) State() types.State             { return types.Playing }
func (*playingInfoPlayer) StateChan() <-chan types.State  { return nil }
func (*playingInfoPlayer) Volume() int                    { return 50 }
func (*playingInfoPlayer) SetVolume(int)                  {}
func (*playingInfoPlayer) UpVolume()                      {}
func (*playingInfoPlayer) DownVolume()                    {}
func (*playingInfoPlayer) Close()                         {}

func TestPlayingInfoOnlyIncludesLyricsForCurrentSong(t *testing.T) {
	lyricService := lyric.NewService(staticLyricFetcher{}, false, 0, false)
	loadID := lyricService.BeginSong(1)
	loaded, err := lyricService.SetSong(context.Background(), structs.Song{Id: 1}, loadID)
	if err != nil || !loaded {
		t.Fatalf("SetSong() = loaded %v, error %v; want loaded true", loaded, err)
	}

	p := &Player{
		playlistManager: playlist.NewPlaylistManager(),
		lyricService:    lyricService,
		Player:          &playingInfoPlayer{},
	}
	if err := p.playlistManager.Initialize(0, []structs.Song{{Id: 1}}); err != nil {
		t.Fatal(err)
	}
	if lrc := p.PlayingInfo().LRCText; !strings.Contains(lrc, "lyrics-for-song-1") {
		t.Fatalf("matching lyrics = %q, want song 1 lyrics", lrc)
	}

	if err := p.playlistManager.Initialize(0, []structs.Song{{Id: 2}}); err != nil {
		t.Fatal(err)
	}
	if lrc := p.PlayingInfo().LRCText; lrc != "" {
		t.Fatalf("mismatched lyrics = %q, want empty", lrc)
	}
}
