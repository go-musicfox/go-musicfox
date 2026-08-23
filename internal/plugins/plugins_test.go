package plugins

import (
	"testing"

	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// TestBlankImportRegistersCheckUpdateMenu proves the aggregator's blank-import
// chain (cmd/musicfox.go → internal/plugins → internal/plugins/checkupdate)
// links the plugin and triggers its init() registration: the "check_update"
// provider is visible on the registry, BuildMenu resolves it through the
// exported base, and the plugin-declared main-menu entry is appended.
func TestBlankImportRegistersCheckUpdateMenu(t *testing.T) {
	if !(ui.MenuRegistry{}).Registered("check_update") {
		t.Fatal("check_update menu provider not registered via aggregator blank import")
	}
	menu, err := ui.BuildMenu("check_update", ui.BaseMenu{}, ui.NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(check_update) error = %v", err)
	}
	if key := menu.GetMenuKey(); key != "check_update" {
		t.Fatalf("GetMenuKey() = %q, want check_update", key)
	}

	// The plugin also declares its main-menu entry (Phase 3.9).
	found := false
	for _, item := range ui.MainMenuPluginItems() {
		if item.Key == "check_update" && item.Title == "检查更新" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("check_update main-menu item not registered via aggregator blank import")
	}
}

// TestBlankImportRegistersAlbumPlugin proves the fourth real plugin
// (internal/plugins/album — the whole album cluster) links through the same
// aggregator: its menu providers are registered with their original keys and
// its entry menu declares the 专辑列表 main-menu item.
func TestBlankImportRegistersAlbumPlugin(t *testing.T) {
	for _, key := range []string{"album_menu", "album_new_area", "album_top_area", "album_new_hot", "album_new", "album_top", "album_sub_list", "album_detail"} {
		if !(ui.MenuRegistry{}).Registered(key) {
			t.Fatalf("album menu provider %q not registered via aggregator blank import", key)
		}
	}
	entry, err := ui.BuildMenu("album_menu", ui.BaseMenu{}, ui.NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(album_menu) error = %v", err)
	}
	if key := entry.GetMenuKey(); key != "album_menu" {
		t.Fatalf("GetMenuKey() = %q, want album_menu", key)
	}
	found := false
	for _, item := range ui.MainMenuPluginItems() {
		if item.Key == "album_menu" && item.Title == "专辑列表" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("album_menu main-menu item not registered via aggregator blank import")
	}
}

// TestBlankImportRegistersDjPlugin proves the third real plugin
// (internal/plugins/dj — the whole DJ/radio cluster) links through the same
// aggregator: its menu providers are registered with their original keys and
// its entry menu declares the 主播电台 main-menu item.
func TestBlankImportRegistersDjPlugin(t *testing.T) {
	for _, key := range []string{"dj_radio_detail", "dj_category_detail", "dj_category", "dj_program_rank", "dj_program_hour_rank", "dj_hot", "dj_sub", "dj_recommend", "dj_today_recommend", "radio_dj_type"} {
		if !(ui.MenuRegistry{}).Registered(key) {
			t.Fatalf("dj menu provider %q not registered via aggregator blank import", key)
		}
	}
	entry, err := ui.BuildMenu("radio_dj_type", ui.BaseMenu{}, ui.NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(radio_dj_type) error = %v", err)
	}
	if key := entry.GetMenuKey(); key != "radio_dj_type" {
		t.Fatalf("GetMenuKey() = %q, want radio_dj_type", key)
	}
	found := false
	for _, item := range ui.MainMenuPluginItems() {
		if item.Key == "radio_dj_type" && item.Title == "主播电台" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("radio_dj_type main-menu item not registered via aggregator blank import")
	}
}

// TestBlankImportRegistersArtistPlugin proves the fifth real plugin
// (internal/plugins/artist — the whole artist cluster) links through the same
// aggregator: its menu providers are registered with their original keys and
// its entry menu declares the 热门歌手 main-menu item.
func TestBlankImportRegistersArtistPlugin(t *testing.T) {
	for _, key := range []string{"hot_artists", "artist_detail", "artist_song", "artist_album", "artist_of_song", "artists_sub_list"} {
		if !(ui.MenuRegistry{}).Registered(key) {
			t.Fatalf("artist menu provider %q not registered via aggregator blank import", key)
		}
	}
	entry, err := ui.BuildMenu("hot_artists", ui.BaseMenu{}, ui.NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(hot_artists) error = %v", err)
	}
	if key := entry.GetMenuKey(); key != "hot_artists" {
		t.Fatalf("GetMenuKey() = %q, want hot_artists", key)
	}
	found := false
	for _, item := range ui.MainMenuPluginItems() {
		if item.Key == "hot_artists" && item.Title == "热门歌手" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("hot_artists main-menu item not registered via aggregator blank import")
	}
}

// TestBlankImportRegistersLastfmPlugin proves the second real plugin
// (internal/plugins/lastfm) links through the same aggregator: its menu and
// page providers are registered and its main-menu item is appended after the
// built-ins.
func TestBlankImportRegistersLastfmPlugin(t *testing.T) {
	if !(ui.MenuRegistry{}).Registered("last_fm") {
		t.Fatal("last_fm menu provider not registered via aggregator blank import")
	}
	if !(ui.PageRegistry{}).Registered("lastfm_auth") {
		t.Fatal("lastfm_auth page provider not registered via aggregator blank import")
	}
	if !(ui.PageRegistry{}).Registered("lastfm_custom_api") {
		t.Fatal("lastfm_custom_api page provider not registered via aggregator blank import")
	}
	menu, err := ui.BuildMenu("last_fm", ui.BaseMenu{}, ui.NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(last_fm) error = %v", err)
	}
	if key := menu.GetMenuKey(); key != "last_fm" {
		t.Fatalf("GetMenuKey() = %q, want last_fm", key)
	}

	found := false
	for _, item := range ui.MainMenuPluginItems() {
		if item.Key == "last_fm" && item.Title == "LastFM" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("last_fm main-menu item not registered via aggregator blank import")
	}
}
