package ui

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/core"
)

// TestPlayerWrapperMenuAPI covers the TUI wrapper's playing-menu bookkeeping:
// the 2-arg SetPlayingMenu stores the shell menu reference, PlayingMenu reads
// it back, and the wrapper MarkPlaylistModified clears it while the core still
// appends "modified" to the key.
func TestPlayerWrapperMenuAPI(t *testing.T) {
	p := &Player{Player: core.NewEmptyPlayer()}
	cp := &CurPlaylist{}
	p.SetPlayingMenu(cp.GetMenuKey(), cp)
	if p.PlayingMenuKey() != CurPlaylistKey || p.PlayingMenu() != cp {
		t.Fatalf("SetPlayingMenu: got key=%q menu=%v", p.PlayingMenuKey(), p.PlayingMenu())
	}

	p.MarkPlaylistModified()
	if p.PlayingMenu() != nil {
		t.Fatalf("MarkPlaylistModified should clear the playing menu")
	}
	if p.PlayingMenuKey() != CurPlaylistKey+"modified" {
		t.Fatalf("MarkPlaylistModified: got key=%q", p.PlayingMenuKey())
	}
}
