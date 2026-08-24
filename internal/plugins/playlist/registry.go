// Package playlist implements the playlist & cloud cluster (我的歌单 / 我的收藏 /
// 精选歌单 / 云盘 / 歌单详情) as the seventh real plugin. The five menus moved
// from internal/ui verbatim with their provider keys unchanged — user_playlist
// (我的歌单), user_collect (我的收藏), high_quality_playlists (精选歌单), could
// (云盘) and playlist_detail (歌单详情) — and each declares its own main-menu
// item except playlist_detail, which is a jump target, not a top-level entry
// (the four built-in entries were removed from menu_main.go; plugin items are
// appended after all built-ins). user_playlist is the parameterized main-menu
// entry demo (Phase 3.9.9): it is built via RegisterMainMenuItemWith with
// ui.UserPlaylistOpts{UserID: ui.CurUser}, exactly like the built-in entry
// used to construct it. Cross-menu jumps into ui stay key-based (user_playlist
// / high_quality_playlists / search_result -> "playlist_detail", now provided
// by this plugin), and user_collect builds its album_sub_list / artists_sub_list
// sub-menus through the same registry keys the album / artist plugins register
// (their keys are plugin-supplied now). ui.PlaylistDetailOpts stays in ui —
// operate.go / player.go / the search_result jump site keep using it.
package playlist

import (
	"github.com/anhoder/foxful-cli/model"

	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// enterMenuCallback mirrors ui.EnterMenuCallback (unexported there): the login
// callback re-enters the requesting menu once login succeeds.
func enterMenuCallback(main *model.Main) ui.LoginCallback {
	return func() model.Page {
		return main.EnterMenu(nil, nil)
	}
}

// init is the compile-time registration entry (linked via the internal/plugins
// aggregator blank import, which cmd/musicfox.go pulls in). Every key is
// identical to the one the menu registered under in internal/ui before the
// extraction: user_playlist / user_collect / high_quality_playlists / could /
// playlist_detail. The user_playlist menu is parameterized (UserPlaylistOpts
// carries the user ID); its main-menu entry uses the RegisterMainMenuItemWith
// builder form so it is constructed with UserID = ui.CurUser (当前用户歌单)
// exactly as the built-in main-menu entry did. playlist_detail is
// parameterized too (PlaylistDetailOpts carries the playlist ID) and is a pure
// jump target, not a main-menu entry. The other three are no-arg menus and
// declare their own main-menu items — the built-in 我的歌单 / 我的收藏 / 精选歌单 /
// 云盘 entries were removed from menu_main.go (plugin items are appended after
// all built-ins).
func init() {
	ui.RegisterMenu("user_playlist", func(base ui.BaseMenu, opts ui.UserPlaylistOpts) (ui.Menu, error) {
		return NewUserPlaylistMenu(base, opts.UserID), nil
	})
	ui.RegisterMenu("user_collect", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
		return NewUserCollectionMenu(base), nil
	})
	ui.RegisterMenu("high_quality_playlists", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
		return NewHighQualityPlaylistsMenu(base), nil
	})
	ui.RegisterMenu("could", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
		return NewCloudMenu(base), nil
	})
	ui.RegisterMenu("playlist_detail", func(base ui.BaseMenu, opts ui.PlaylistDetailOpts) (ui.Menu, error) {
		return NewPlaylistDetailMenu(base, opts.PlaylistID), nil
	})
	// 声明主菜单入口：NewMainMenu 经 After 锚点链归并复现插件化前的主菜单
	// 原始顺序。每个入口声明其前驱项 key：我的歌单跟在每日推荐歌单（recommend
	// 插件）后，我的收藏跟在我的歌单后，精选歌单跟在排行榜（recommend 插件）后，
	// 云盘跟在最近播放歌曲（recommend 插件）后。user_playlist 经参数化 builder
	// 构造（UserID = ui.CurUser，与内置入口行为一致）。
	ui.RegisterMainMenuItemAfter("user_playlist", "我的歌单", "daily_playlists", func(base ui.BaseMenu) ui.Menu {
		return ui.MustBuild("user_playlist", base, ui.UserPlaylistOpts{UserID: ui.CurUser})
	})
	ui.RegisterMainMenuItemAfter("user_collect", "我的收藏", "user_playlist", nil)
	ui.RegisterMainMenuItemAfter("high_quality_playlists", "精选歌单", "ranks", nil)
	ui.RegisterMainMenuItemAfter("could", "云盘", "recent_songs", nil)
}
