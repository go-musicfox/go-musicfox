package song

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// TestSimilarSongsMenuBuildsViaRegistryFactory proves the simi_songs provider
// is registered with its original key and builds through the real registry
// factory with the shared ui.SimiSongsOpts, keeping its dynamic GetMenuKey.
func TestSimilarSongsMenuBuildsViaRegistryFactory(t *testing.T) {
	menu, err := ui.BuildMenu("simi_songs", ui.BaseMenu{}, ui.SimiSongsOpts{Song: structs.Song{Id: 1}})
	if err != nil {
		t.Fatalf("BuildMenu(simi_songs) error = %v", err)
	}
	ss, ok := menu.(*SimilarSongsMenu)
	if !ok {
		t.Fatalf("BuildMenu(simi_songs) = %T, want *SimilarSongsMenu", menu)
	}
	if key := ss.GetMenuKey(); key != "simi_songs_1" {
		t.Fatalf("GetMenuKey() = %q, want %q", key, "simi_songs_1")
	}
}

// TestAddToUserPlaylistMenuBuildsViaRegistryFactory proves the
// add_to_user_playlist provider is registered with its original key and builds
// through the real registry factory with the shared ui.AddToUserPlaylistOpts
// (user id / song / add-or-del flag reach the menu).
func TestAddToUserPlaylistMenuBuildsViaRegistryFactory(t *testing.T) {
	menu, err := ui.BuildMenu("add_to_user_playlist", ui.BaseMenu{}, ui.AddToUserPlaylistOpts{UserID: 7, Song: structs.Song{Id: 9}, IsAdd: true})
	if err != nil {
		t.Fatalf("BuildMenu(add_to_user_playlist) error = %v", err)
	}
	au, ok := menu.(*AddToUserPlaylistMenu)
	if !ok {
		t.Fatalf("BuildMenu(add_to_user_playlist) = %T, want *AddToUserPlaylistMenu", menu)
	}
	if au.userID != 7 || au.song.Id != 9 || !au.action {
		t.Fatalf("opts not propagated: userID = %d, song.Id = %d, action = %v; want 7 / 9 / true", au.userID, au.song.Id, au.action)
	}
	if key := au.GetMenuKey(); key != "add_to_user_playlist_7" {
		t.Fatalf("GetMenuKey() = %q, want %q", key, "add_to_user_playlist_7")
	}
}
