package artist

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// TestAllMenusBuildViaRegistryFactory proves every moved provider key is
// registered with its original key (identical to the pre-extraction keys in
// registry_registrations.go) and builds through the real registry factory with
// a zero BaseMenu. Parameterized menus keep their dynamic GetMenuKey form.
func TestAllMenusBuildViaRegistryFactory(t *testing.T) {
	base := ui.BaseMenu{}

	noArgKeys := []struct{ key, want string }{
		{"hot_artists", "hot_artists"},
		{"artists_sub_list", "artists_sub_list"},
	}
	for _, tc := range noArgKeys {
		menu, err := ui.BuildMenu(tc.key, base, ui.NoArgMenuOpts{})
		if err != nil {
			t.Fatalf("BuildMenu(%s) error = %v", tc.key, err)
		}
		if key := menu.GetMenuKey(); key != tc.want {
			t.Fatalf("%s: GetMenuKey() = %q, want %q", tc.key, key, tc.want)
		}
	}

	detail, err := ui.BuildMenu("artist_detail", base, ui.ArtistDetailOpts{ArtistID: 456, Name: "artist"})
	if err != nil {
		t.Fatalf("BuildMenu(artist_detail) error = %v", err)
	}
	if key := detail.GetMenuKey(); key != "artist_detail_456" {
		t.Fatalf("artist_detail GetMenuKey() = %q, want artist_detail_456", key)
	}

	song, err := ui.BuildMenu("artist_song", base, ArtistSongOpts{ArtistID: 456})
	if err != nil {
		t.Fatalf("BuildMenu(artist_song) error = %v", err)
	}
	if key := song.GetMenuKey(); key != "artist_song_456" {
		t.Fatalf("artist_song GetMenuKey() = %q, want artist_song_456", key)
	}

	album, err := ui.BuildMenu("artist_album", base, ArtistAlbumOpts{ArtistID: 456})
	if err != nil {
		t.Fatalf("BuildMenu(artist_album) error = %v", err)
	}
	if key := album.GetMenuKey(); key != "artist_album_456" {
		t.Fatalf("artist_album GetMenuKey() = %q, want artist_album_456", key)
	}

	ofSong, err := ui.BuildMenu("artist_of_song", base, ui.ArtistsOfSongOpts{Song: structs.Song{Id: 789}})
	if err != nil {
		t.Fatalf("BuildMenu(artist_of_song) error = %v", err)
	}
	if key := ofSong.GetMenuKey(); key != "artist_of_song" {
		t.Fatalf("artist_of_song GetMenuKey() = %q, want artist_of_song", key)
	}
	if ofSong.(*ArtistsOfSongMenu).SongID() != 789 {
		t.Fatalf("artist_of_song SongID() = %d, want 789", ofSong.(*ArtistsOfSongMenu).SongID())
	}
}

// TestCrossMenuJumpsToArtistDetail proves the cluster's internal cross-menu
// jumps still resolve through the registry with the same keys: stubbing the
// network-fed artist list and calling SubMenu returns an *ArtistDetailMenu with
// the matching ID.
func TestCrossMenuJumpsToArtistDetail(t *testing.T) {
	base := ui.BaseMenu{}

	hot, err := ui.BuildMenu("hot_artists", base, ui.NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(hot_artists) error = %v", err)
	}
	hotMenu := hot.(*HotArtistsMenu)
	hotMenu.artists = []structs.Artist{{Id: 456, Name: "artist"}}
	sub := hotMenu.SubMenu(nil, 0)
	detail, ok := sub.(*ArtistDetailMenu)
	if !ok {
		t.Fatalf("hot_artists.SubMenu(0) = %T, want *ArtistDetailMenu", sub)
	}
	if detail.ArtistID() != 456 {
		t.Fatalf("ArtistID() = %d, want 456", detail.ArtistID())
	}

	subList, err := ui.BuildMenu("artists_sub_list", base, ui.NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(artists_sub_list) error = %v", err)
	}
	subListMenu := subList.(*ArtistsSubscribeListMenu)
	subListMenu.artists = []structs.Artist{{Id: 457, Name: "artist2"}}
	sub = subListMenu.SubMenu(nil, 0)
	detail, ok = sub.(*ArtistDetailMenu)
	if !ok {
		t.Fatalf("artists_sub_list.SubMenu(0) = %T, want *ArtistDetailMenu", sub)
	}
	if detail.ArtistID() != 457 {
		t.Fatalf("ArtistID() = %d, want 457", detail.ArtistID())
	}
}

// TestArtistDetailBuildsSubMenusThroughRegistry proves the artist_detail menu
// keeps building its parameterized sub-menus through the registry with the same
// keys (artist_detail -> artist_song / artist_album).
func TestArtistDetailBuildsSubMenusThroughRegistry(t *testing.T) {
	menu := NewArtistDetailMenu(ui.BaseMenu{}, 456, "artist")
	if views := menu.MenuViews(); len(views) != 2 {
		t.Fatalf("MenuViews() = %d items, want 2", len(views))
	}

	songSub := menu.SubMenu(nil, 0)
	songMenu, ok := songSub.(*ArtistSongMenu)
	if !ok {
		t.Fatalf("artist_detail.SubMenu(0) = %T, want *ArtistSongMenu", songSub)
	}
	if songMenu.artistID != 456 {
		t.Fatalf("artist_song artistID = %d, want 456", songMenu.artistID)
	}

	albumSub := menu.SubMenu(nil, 1)
	albumMenu, ok := albumSub.(*ArtistAlbumMenu)
	if !ok {
		t.Fatalf("artist_detail.SubMenu(1) = %T, want *ArtistAlbumMenu", albumSub)
	}
	if albumMenu.artistID != 456 {
		t.Fatalf("artist_album artistID = %d, want 456", albumMenu.artistID)
	}
}

// TestMainMenuItemRegistered proves the entry menu declares its own main-menu
// item (the built-in 热门歌手 entry was removed from menu_main.go, Phase 3.9.x).
func TestMainMenuItemRegistered(t *testing.T) {
	found := false
	for _, item := range ui.MainMenuPluginItems() {
		if item.Key == "hot_artists" && item.Title == "热门歌手" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("hot_artists main-menu item not registered")
	}
}
