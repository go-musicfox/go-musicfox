package recommend

import (
	"fmt"
	"testing"

	"github.com/anhoder/foxful-cli/model"

	// playlist_detail is now provided by the playlist plugin; the SubMenu jump
	// assertions below type-assert its concrete menu.
	playlistplugin "github.com/go-musicfox/go-musicfox/internal/plugins/playlist"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// TestAllMenusBuildViaRegistryFactory proves every moved provider key is
// registered with its original key (identical to the pre-extraction keys in
// registry_registrations.go) and builds through the real registry factory with
// a zero BaseMenu. All five recommend menus are no-arg: they build with the
// plain NoArgMenuOpts form and keep their static GetMenuKey.
func TestAllMenusBuildViaRegistryFactory(t *testing.T) {
	base := ui.BaseMenu{}

	noArgKeys := []struct{ key, want string }{
		{"daily_songs", "daily_songs"},
		{"daily_playlists", "daily_playlists"},
		{"personal_fm", "personal_fm"},
		{"recent_songs", "recent_songs"},
		{"ranks", "ranks"},
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
}

// TestCrossMenuJumpsToPlaylistDetail proves the cluster's internal cross-menu
// jumps still resolve through the registry with the same keys as before:
// stubbing the network-fed data and calling SubMenu returns a
// *playlistplugin.PlaylistDetailMenu with the matching dynamic key
// (playlist_detail is plugin-supplied now).
func TestCrossMenuJumpsToPlaylistDetail(t *testing.T) {
	base := ui.BaseMenu{}

	check := func(t *testing.T, name string, sub model.Menu, wantID int64) {
		t.Helper()
		pd, ok := sub.(*playlistplugin.PlaylistDetailMenu)
		if !ok {
			t.Fatalf("%s.SubMenu(0) = %T, want *playlistplugin.PlaylistDetailMenu", name, sub)
		}
		wantKey := fmt.Sprintf("playlist_detail_%d", wantID)
		if key := pd.GetMenuKey(); key != wantKey {
			t.Fatalf("%s GetMenuKey() = %q, want %q", name, key, wantKey)
		}
	}

	// ranks -> playlist_detail (SubMenu edge; stub the network-fed ranks).
	ranks, err := ui.BuildMenu("ranks", base, ui.NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(ranks) error = %v", err)
	}
	ranksMenu := ranks.(*RanksMenu)
	ranksMenu.ranks = []structs.Rank{{Id: 123}}
	check(t, "ranks", ranksMenu.SubMenu(nil, 0), 123)

	// daily_playlists -> playlist_detail (SubMenu edge; stub the network-fed
	// daily recommend playlists).
	daily, err := ui.BuildMenu("daily_playlists", base, ui.NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(daily_playlists) error = %v", err)
	}
	dailyMenu := daily.(*DailyRecommendPlaylistsMenu)
	dailyMenu.playlists = []structs.Playlist{{Id: 456}}
	check(t, "daily_playlists", dailyMenu.SubMenu(nil, 0), 456)
}

// TestMainMenuItemsRegistered proves all five entry menus declare their own
// main-menu items (the built-in 每日推荐歌曲 / 每日推荐歌单 / 私人FM /
// 最近播放歌曲 / 排行榜 entries were removed from menu_main.go, Phase 3.9.x).
func TestMainMenuItemsRegistered(t *testing.T) {
	want := map[string]string{
		"daily_songs":     "每日推荐歌曲",
		"daily_playlists": "每日推荐歌单",
		"personal_fm":     "私人FM",
		"recent_songs":    "最近播放歌曲",
		"ranks":           "排行榜",
	}
	found := map[string]string{}
	for _, item := range ui.MainMenuPluginItems() {
		if title, ok := want[item.Key]; ok && title == item.Title {
			found[item.Key] = item.Title
		}
	}
	for key, title := range want {
		if found[key] != title {
			t.Fatalf("main-menu item %q (%q) not registered", key, title)
		}
	}
}
