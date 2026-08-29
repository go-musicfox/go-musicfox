package core

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/playlist"
	"github.com/go-musicfox/go-musicfox/internal/structs"
)

func TestPlayerPlaylistStateAPI(t *testing.T) {
	p := &Player{playlistManager: playlist.NewPlaylistManager()}
	songs := []structs.Song{{Id: 1}, {Id: 2}, {Id: 3}}

	p.ReinitializePlaylist(1, songs)
	if len(p.Playlist()) != 3 || p.CurSongIndex() != 1 {
		t.Fatalf("ReinitializePlaylist: got playlist=%v index=%d", p.Playlist(), p.CurSongIndex())
	}

	if !p.PlaylistUpdateAt().IsZero() {
		t.Fatalf("PlaylistUpdateAt should be zero before MarkPlaylistUpdated")
	}
	p.MarkPlaylistUpdated()
	if p.PlaylistUpdateAt().IsZero() {
		t.Fatalf("PlaylistUpdateAt should be set after MarkPlaylistUpdated")
	}

	p.SetPlayingMenu("cur_playlist")
	if p.PlayingMenuKey() != "cur_playlist" {
		t.Fatalf("SetPlayingMenu: got key=%q", p.PlayingMenuKey())
	}

	p.MarkPlaylistModified()
	if p.PlayingMenuKey() != "cur_playlistmodified" {
		t.Fatalf("MarkPlaylistModified: got key=%q", p.PlayingMenuKey())
	}

	// RemoveSong returns the next song to play and removes the target by index.
	removed, err := p.RemoveSong(0)
	if err != nil || removed.Id != 2 {
		t.Fatalf("RemoveSong: err=%v removed=%v", err, removed)
	}
	if len(p.Playlist()) != 2 || p.Playlist()[0].Id != 2 || p.Playlist()[1].Id != 3 {
		t.Fatalf("RemoveSong: playlist=%v", p.Playlist())
	}
}
