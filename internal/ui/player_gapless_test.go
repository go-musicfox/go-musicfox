package ui

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/playlist"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

func TestPeekGaplessSongDeterministicModes(t *testing.T) {
	p := &Player{playlistManager: playlist.NewPlaylistManager()}
	songs := []structs.Song{{Id: 1}, {Id: 2}, {Id: 3}}
	p.ReinitializePlaylist(1, songs)

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
	p.ReinitializePlaylist(1, []structs.Song{{Id: 1}, {Id: 2}})
	if err := p.playlistManager.SetPlayMode(types.PmListLoop); err != nil {
		t.Fatal(err)
	}
	got, ok := p.peekGaplessSong()
	if !ok || got.Id != 1 {
		t.Fatalf("got id=%d ok=%v, want id=1 ok=true", got.Id, ok)
	}
}
