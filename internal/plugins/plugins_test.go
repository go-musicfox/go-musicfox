package plugins

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/structs"
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

// TestBlankImportRegistersRecommendPlugin proves the sixth real plugin
// (internal/plugins/recommend — the whole recommend cluster) links through the
// same aggregator: its menu providers are registered with their original keys
// and every entry menu declares its own main-menu item.
func TestBlankImportRegistersRecommendPlugin(t *testing.T) {
	for _, key := range []string{"daily_songs", "daily_playlists", "personal_fm", "recent_songs", "ranks"} {
		if !(ui.MenuRegistry{}).Registered(key) {
			t.Fatalf("recommend menu provider %q not registered via aggregator blank import", key)
		}
	}
	for _, tc := range []struct{ key, title string }{
		{"daily_songs", "每日推荐歌曲"},
		{"daily_playlists", "每日推荐歌单"},
		{"personal_fm", "私人FM"},
		{"recent_songs", "最近播放歌曲"},
		{"ranks", "排行榜"},
	} {
		menu, err := ui.BuildMenu(tc.key, ui.BaseMenu{}, ui.NoArgMenuOpts{})
		if err != nil {
			t.Fatalf("BuildMenu(%s) error = %v", tc.key, err)
		}
		if key := menu.GetMenuKey(); key != tc.key {
			t.Fatalf("GetMenuKey() = %q, want %q", key, tc.key)
		}
		found := false
		for _, item := range ui.MainMenuPluginItems() {
			if item.Key == tc.key && item.Title == tc.title {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("%s main-menu item (%s) not registered via aggregator blank import", tc.key, tc.title)
		}
	}
}

// TestBlankImportRegistersPlaylistPlugin proves the seventh real plugin
// (internal/plugins/playlist — the whole playlist & cloud cluster) links
// through the same aggregator: its menu providers are registered with their
// original keys and every entry menu declares its own main-menu item.
// user_playlist is the parameterized entry — it must carry a Build (the
// RegisterMainMenuItemWith form) that constructs the menu with ui.CurUser.
func TestBlankImportRegistersPlaylistPlugin(t *testing.T) {
	for _, key := range []string{"user_playlist", "user_collect", "high_quality_playlists", "could"} {
		if !(ui.MenuRegistry{}).Registered(key) {
			t.Fatalf("playlist menu provider %q not registered via aggregator blank import", key)
		}
	}

	// user_playlist is parameterized: it builds with UserPlaylistOpts.
	up, err := ui.BuildMenu("user_playlist", ui.BaseMenu{}, ui.UserPlaylistOpts{UserID: 0})
	if err != nil {
		t.Fatalf("BuildMenu(user_playlist) error = %v", err)
	}
	if key := up.GetMenuKey(); key != "user_playlist_0" {
		t.Fatalf("GetMenuKey() = %q, want %q", key, "user_playlist_0")
	}

	for _, tc := range []struct{ key, title string }{
		{"user_collect", "我的收藏"},
		{"high_quality_playlists", "精选歌单"},
		{"could", "云盘"},
	} {
		menu, err := ui.BuildMenu(tc.key, ui.BaseMenu{}, ui.NoArgMenuOpts{})
		if err != nil {
			t.Fatalf("BuildMenu(%s) error = %v", tc.key, err)
		}
		if key := menu.GetMenuKey(); key != tc.key {
			t.Fatalf("GetMenuKey() = %q, want %q", key, tc.key)
		}
	}

	wantItems := map[string]string{
		"user_playlist":          "我的歌单",
		"user_collect":           "我的收藏",
		"high_quality_playlists": "精选歌单",
		"could":                  "云盘",
	}
	foundItems := map[string]string{}
	var userPlaylistBuilder func(base ui.BaseMenu) ui.Menu
	for _, item := range ui.MainMenuPluginItems() {
		if title, ok := wantItems[item.Key]; ok && title == item.Title {
			foundItems[item.Key] = item.Title
		}
		if item.Key == "user_playlist" {
			userPlaylistBuilder = item.Build
		}
	}
	for key, title := range wantItems {
		if foundItems[key] != title {
			t.Fatalf("%s main-menu item (%s) not registered via aggregator blank import", key, title)
		}
	}

	// user_playlist must be the parameterized main-menu entry: it carries a
	// Build (not built via mustBuildNoArg — it is a parameterized provider)
	// that constructs the menu with ui.CurUser (0).
	if userPlaylistBuilder == nil {
		t.Fatal("user_playlist main-menu item has no Build (want RegisterMainMenuItemWith builder)")
	}
	menu := userPlaylistBuilder(ui.BaseMenu{})
	if key := menu.GetMenuKey(); key != "user_playlist_0" {
		t.Fatalf("user_playlist builder GetMenuKey() = %q, want %q (built with ui.CurUser)", key, "user_playlist_0")
	}
}

// TestBlankImportRegistersSearchPlugin proves the eighth real plugin
// (internal/plugins/search — the whole search cluster) links through the same
// aggregator: its menu providers (search_type / search_result) and the search
// page registration are visible, and it declares the 搜索 main-menu item
// (after album_menu, the pre-extraction position).
func TestBlankImportRegistersSearchPlugin(t *testing.T) {
	for _, key := range []string{"search_type", "search_result"} {
		if !(ui.MenuRegistry{}).Registered(key) {
			t.Fatalf("search menu provider %q not registered via aggregator blank import", key)
		}
	}
	if !(ui.PageRegistry{}).Registered("search") {
		t.Fatal("search page provider not registered via aggregator blank import")
	}

	// search_type is the no-arg main-menu entry; search_result is
	// parameterized (SearchResultOpts carries the search type).
	st, err := ui.BuildMenu("search_type", ui.BaseMenu{}, ui.NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(search_type) error = %v", err)
	}
	if key := st.GetMenuKey(); key != "search_type" {
		t.Fatalf("GetMenuKey() = %q, want search_type", key)
	}

	found := false
	for _, item := range ui.MainMenuPluginItems() {
		if item.Key == "search_type" && item.Title == "搜索" && item.After == "album_menu" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("search_type main-menu item (搜索, after album_menu) not registered via aggregator blank import")
	}
}

// TestBlankImportRegistersSongPlugin proves the ninth real plugin
// (internal/plugins/song — the single-song cluster) links through the same
// aggregator: both parameterized jump-target providers (simi_songs /
// add_to_user_playlist) are registered with their original keys.
func TestBlankImportRegistersSongPlugin(t *testing.T) {
	for _, key := range []string{"simi_songs", "add_to_user_playlist"} {
		if !(ui.MenuRegistry{}).Registered(key) {
			t.Fatalf("song menu provider %q not registered via aggregator blank import", key)
		}
	}

	ss, err := ui.BuildMenu("simi_songs", ui.BaseMenu{}, ui.SimiSongsOpts{Song: structs.Song{Id: 1}})
	if err != nil {
		t.Fatalf("BuildMenu(simi_songs) error = %v", err)
	}
	if key := ss.GetMenuKey(); key != "simi_songs_1" {
		t.Fatalf("GetMenuKey() = %q, want simi_songs_1", key)
	}

	am, err := ui.BuildMenu("add_to_user_playlist", ui.BaseMenu{}, ui.AddToUserPlaylistOpts{
		UserID: 7, Song: structs.Song{Id: 9}, IsAdd: true,
	})
	if err != nil {
		t.Fatalf("BuildMenu(add_to_user_playlist) error = %v", err)
	}
	if key := am.GetMenuKey(); key != "add_to_user_playlist_7" {
		t.Fatalf("GetMenuKey() = %q, want add_to_user_playlist_7", key)
	}
}

// TestMainMenuPreservesOriginalOrder proves the after-anchor chain merge in
// NewMainMenu: plugin entries (搜索 after album_menu included — it moved into
// the search plugin with the same anchor) and the built-in 帮助 entry chain by
// their After anchors and reproduce the original pre-extraction main-menu
// sequence exactly (16 items). This is the integration view of the same
// mechanism the ui package tests at the unit level — here the full plugin set
// is linked via the aggregator. Titles() is side-effect free, so a zero-base
// menu (no live services) is enough to read the display order.
func TestMainMenuPreservesOriginalOrder(t *testing.T) {
	got := ui.NewMainMenu(ui.NewBaseMenu(nil)).Titles()
	want := []string{
		"每日推荐歌曲", "每日推荐歌单", "我的歌单", "我的收藏", "私人FM",
		"专辑列表", "搜索", "排行榜", "精选歌单", "热门歌手", "最近播放歌曲",
		"云盘", "主播电台", "LastFM", "帮助", "检查更新",
	}
	if len(got) != len(want) {
		t.Fatalf("main menu has %d items, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("main menu item %d = %q, want %q (full sequence: %v)", i, got[i], want[i], got)
		}
	}
}

// TestMainMenuHidesDisabledPluginItems proves the plugin configurable
// enable/disable consumption point at the aggregator level: with the search
// plugin disabled in [plugins], its 搜索 main-menu item disappears while the
// after-anchor chain still builds without panicking. The disabled item's key
// (search_type) is a live After anchor of the recommend plugin's 排行榜 — this
// proves anchor integrity is computed over all registered entries (including
// disabled ones) and the chain skips the hidden entry at display.
func TestMainMenuHidesDisabledPluginItems(t *testing.T) {
	previous := configs.AppConfig
	configs.AppConfig = &configs.Config{Plugins: configs.PluginsConfig{Disabled: []string{"search"}}}
	t.Cleanup(func() { configs.AppConfig = previous })

	got := ui.NewMainMenu(ui.NewBaseMenu(nil)).Titles()
	want := []string{
		"每日推荐歌曲", "每日推荐歌单", "我的歌单", "我的收藏", "私人FM",
		"专辑列表", "排行榜", "精选歌单", "热门歌手", "最近播放歌曲",
		"云盘", "主播电台", "LastFM", "帮助", "检查更新",
	}
	if len(got) != len(want) {
		t.Fatalf("main menu with search disabled has %d items, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("main menu item %d = %q, want %q (full sequence: %v)", i, got[i], want[i], got)
		}
	}
}
