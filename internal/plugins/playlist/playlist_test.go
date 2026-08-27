package playlist

import (
	"fmt"
	"testing"

	"github.com/anhoder/foxful-cli/model"

	// user_collect 的构造器经 ui.MustBuildNoArg 构建 album_sub_list /
	// artists_sub_list 子菜单，其 provider 由 album / artist 插件注册；本测试
	// 二进制经空导入链接这两个插件（init() 只声明构造器），TestMain 在测试
	// 运行前启动它们（P5：注册窗口从 init() 后移到插件 Start）。
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/album"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/artist"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// TestAllMenusBuildViaRegistryFactory proves every moved provider key is
// registered with its original key (identical to the pre-extraction keys in
// registry_registrations.go) and builds through the real registry factory with
// a zero BaseMenu. user_playlist is parameterized (UserPlaylistOpts carries
// the user ID); the other three are no-arg menus.
func TestAllMenusBuildViaRegistryFactory(t *testing.T) {
	base := ui.BaseMenu{}

	// user_playlist builds with a user ID (the parameterized contract).
	up, err := ui.BuildMenu("user_playlist", base, ui.UserPlaylistOpts{UserID: 987})
	if err != nil {
		t.Fatalf("BuildMenu(user_playlist) error = %v", err)
	}
	upMenu, ok := up.(*UserPlaylistMenu)
	if !ok {
		t.Fatalf("BuildMenu(user_playlist) = %T, want *UserPlaylistMenu", up)
	}
	if upMenu.userID != 987 {
		t.Fatalf("userID = %d, want 987 (opts must reach the menu)", upMenu.userID)
	}
	if key := upMenu.GetMenuKey(); key != "user_playlist_987" {
		t.Fatalf("GetMenuKey() = %q, want %q", key, "user_playlist_987")
	}

	for _, tc := range []struct{ key, want string }{
		{"user_collect", "user_collect"},
		{"high_quality_playlists", "high_quality_playlists"},
		{"could", "could"},
	} {
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
// *PlaylistDetailMenu (now provided by this plugin) with the matching dynamic
// key.
func TestCrossMenuJumpsToPlaylistDetail(t *testing.T) {
	base := ui.BaseMenu{}

	check := func(t *testing.T, name string, sub model.Menu, wantID int64) {
		t.Helper()
		pd, ok := sub.(*PlaylistDetailMenu)
		if !ok {
			t.Fatalf("%s.SubMenu(0) = %T, want *PlaylistDetailMenu", name, sub)
		}
		wantKey := fmt.Sprintf("playlist_detail_%d", wantID)
		if key := pd.GetMenuKey(); key != wantKey {
			t.Fatalf("%s GetMenuKey() = %q, want %q", name, key, wantKey)
		}
	}

	// user_playlist -> playlist_detail (SubMenu edge; stub the network-fed
	// playlists).
	up, err := ui.BuildMenu("user_playlist", base, ui.UserPlaylistOpts{UserID: 1})
	if err != nil {
		t.Fatalf("BuildMenu(user_playlist) error = %v", err)
	}
	upMenu := up.(*UserPlaylistMenu)
	upMenu.playlists = []structs.Playlist{{Id: 123}}
	check(t, "user_playlist", upMenu.SubMenu(nil, 0), 123)

	// high_quality_playlists -> playlist_detail (SubMenu edge; stub the
	// network-fed high quality playlists).
	hq, err := ui.BuildMenu("high_quality_playlists", base, ui.NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(high_quality_playlists) error = %v", err)
	}
	hqMenu := hq.(*HighQualityPlaylistsMenu)
	hqMenu.playlists = []structs.Playlist{{Id: 456}}
	check(t, "high_quality_playlists", hqMenu.SubMenu(nil, 0), 456)
}

// TestUserCollectionBuildsCrossPluginSubMenus proves user_collect builds its
// album_sub_list / artists_sub_list sub-menus through the registry keys that
// the album / artist plugins register (both plugins are blank-imported in this
// test binary).
func TestUserCollectionBuildsCrossPluginSubMenus(t *testing.T) {
	base := ui.BaseMenu{}
	uc, err := ui.BuildMenu("user_collect", base, ui.NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(user_collect) error = %v", err)
	}
	ucMenu := uc.(*UserCollectionMenu)
	if views := ucMenu.MenuViews(); len(views) != 2 {
		t.Fatalf("user_collect MenuViews() = %v, want 2 static items", views)
	}
	for i, wantKey := range []string{"album_sub_list", "artists_sub_list"} {
		sub := ucMenu.SubMenu(nil, i)
		if sub == nil {
			t.Fatalf("user_collect.SubMenu(%d) = nil, want %s", i, wantKey)
		}
		if key := sub.GetMenuKey(); key != wantKey {
			t.Fatalf("user_collect.SubMenu(%d) key = %q, want %q", i, key, wantKey)
		}
	}
}

// TestPlaylistDetailBuildsViaRegistryFactory proves playlist_detail is now
// provided by this plugin (its registration moved out of ui together with the
// menu, Phase 3.9.x): it builds through the real registry factory with the
// shared ui.PlaylistDetailOpts and keeps its dynamic GetMenuKey.
func TestPlaylistDetailBuildsViaRegistryFactory(t *testing.T) {
	menu, err := ui.BuildMenu("playlist_detail", ui.BaseMenu{}, ui.PlaylistDetailOpts{PlaylistID: 42})
	if err != nil {
		t.Fatalf("BuildMenu(playlist_detail) error = %v", err)
	}
	pd, ok := menu.(*PlaylistDetailMenu)
	if !ok {
		t.Fatalf("BuildMenu(playlist_detail) = %T, want *PlaylistDetailMenu", menu)
	}
	if key := pd.GetMenuKey(); key != "playlist_detail_42" {
		t.Fatalf("GetMenuKey() = %q, want %q", key, "playlist_detail_42")
	}
}

// TestMainMenuItemsRegistered proves all four entry menus declare their own
// main-menu items — the built-in 我的歌单 / 我的收藏 / 精选歌单 / 云盘 entries
// were removed from menu_main.go (Phase 3.9.x). user_playlist is the
// parameterized entry: it carries a Build that constructs the menu with
// UserID = ui.CurUser (当前用户歌单), exactly like the removed built-in entry.
func TestMainMenuItemsRegistered(t *testing.T) {
	want := map[string]string{
		"user_playlist":          "我的歌单",
		"user_collect":           "我的收藏",
		"high_quality_playlists": "精选歌单",
		"could":                  "云盘",
	}
	found := map[string]string{}
	var userPlaylistBuilder func(base ui.BaseMenu) ui.Menu
	for _, item := range ui.MainMenuPluginItems() {
		if title, ok := want[item.Key]; ok && title == item.Title {
			found[item.Key] = item.Title
		}
		if item.Key == "user_playlist" {
			userPlaylistBuilder = item.Build
		}
	}
	for key, title := range want {
		if found[key] != title {
			t.Fatalf("main-menu item %q (%q) not registered", key, title)
		}
	}

	// The user_playlist entry must use the parameterized builder form and build
	// the menu with UserID = ui.CurUser (the behavior of the removed built-in
	// entry).
	if userPlaylistBuilder == nil {
		t.Fatal("user_playlist main-menu item has no Build (want RegisterMainMenuItemWith builder)")
	}
	menu := userPlaylistBuilder(ui.BaseMenu{})
	up, ok := menu.(*UserPlaylistMenu)
	if !ok {
		t.Fatalf("user_playlist builder = %T, want *UserPlaylistMenu", menu)
	}
	if up.userID != ui.CurUser {
		t.Fatalf("user_playlist builder userID = %d, want ui.CurUser (%d)", up.userID, ui.CurUser)
	}
}
