package ui

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/player"
	"github.com/go-musicfox/go-musicfox/internal/playlist"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

type gaplessCancelTestPlayer struct {
	player.Player
	music      player.URLMusic
	preloadNum int
}

func (p *gaplessCancelTestPlayer) CurMusic() player.URLMusic { return p.music }
func (p *gaplessCancelTestPlayer) Preload(music player.URLMusic) {
	p.preloadNum++
	p.music = music
}
func (p *gaplessCancelTestPlayer) CancelPreload()                                         {}
func (p *gaplessCancelTestPlayer) GaplessTransitionChan() <-chan player.GaplessTransition { return nil }

func TestCancelledGaplessRequestCannotPreload(t *testing.T) {
	gapless := &gaplessCancelTestPlayer{music: player.URLMusic{Song: structs.Song{Id: 1}}}
	p := &Player{Player: gapless, gaplessGeneration: 1}
	generation := p.gaplessGeneration
	p.cancelGaplessPreload()
	if p.finishGaplessPreload(gapless, player.URLMusic{Song: structs.Song{Id: 2}}, 1, generation) {
		t.Fatal("cancelled preload unexpectedly succeeded")
	}
	if gapless.preloadNum != 0 {
		t.Fatalf("cancelled preload called Preload %d times", gapless.preloadNum)
	}
}

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

func TestGaplessDisabledUntilInsertedPlaybackFinishes(t *testing.T) {
	gapless := &gaplessCancelTestPlayer{music: player.URLMusic{Song: structs.Song{Id: 1}}}
	p := &Player{Player: gapless, playlistManager: playlist.NewPlaylistManager()}
	if err := p.playlistManager.Initialize(0, []structs.Song{{Id: 1}, {Id: 2}, {Id: 3}}); err != nil {
		t.Fatal(err)
	}
	p.addSongsToNext([]structs.Song{{Id: 9}})
	if _, _, _, ok := p.beginGaplessPreload(); ok {
		t.Fatal("gapless preload started with an inserted song queued")
	}
	if song, err := p.playlistManager.NextSong(false); err != nil || song.Id != 9 {
		t.Fatalf("inserted NextSong() = %d, %v; want 9", song.Id, err)
	}
	gapless.music = player.URLMusic{Song: structs.Song{Id: 9}}
	if _, _, _, ok := p.beginGaplessPreload(); ok {
		t.Fatal("gapless preload started while inserted song was playing")
	}
	if song, err := p.playlistManager.NextSong(false); err != nil || song.Id != 2 {
		t.Fatalf("normal NextSong() = %d, %v; want 2", song.Id, err)
	}
	gapless.music = player.URLMusic{Song: structs.Song{Id: 2}}
	next, fromID, _, ok := p.beginGaplessPreload()
	if !ok || next.Id != 3 || fromID != 2 {
		t.Fatalf("gapless did not resume after inserted playback: next=%d from=%d ok=%v", next.Id, fromID, ok)
	}
}

func TestAddingInsertedSongInvalidatesInFlightGaplessResult(t *testing.T) {
	gapless := &gaplessCancelTestPlayer{music: player.URLMusic{Song: structs.Song{Id: 1}}}
	p := &Player{Player: gapless, playlistManager: playlist.NewPlaylistManager()}
	if err := p.playlistManager.Initialize(0, []structs.Song{{Id: 1}, {Id: 2}}); err != nil {
		t.Fatal(err)
	}
	_, fromID, generation, ok := p.beginGaplessPreload()
	if !ok {
		t.Fatal("expected normal-song gapless preload reservation")
	}
	p.addSongsToNext([]structs.Song{{Id: 1}})
	if p.finishGaplessPreload(gapless, player.URLMusic{Song: structs.Song{Id: 2}}, fromID, generation) {
		t.Fatal("in-flight normal-song preload survived inserted-song enqueue")
	}
	if gapless.preloadNum != 0 {
		t.Fatalf("stale request called Preload %d times", gapless.preloadNum)
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
