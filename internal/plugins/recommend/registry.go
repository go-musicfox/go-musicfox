// Package recommend implements the recommend cluster (每日推荐 / 私人FM /
// 最近播放 / 排行榜) as the sixth real plugin. All five menus moved from
// internal/ui verbatim with their provider keys unchanged — daily_songs (每日
// 推荐歌曲), daily_playlists (每日推荐歌单), personal_fm (私人FM), recent_songs
// (最近播放歌曲) and ranks (排行榜) — and each declares its own main-menu item
// (the five built-in entries were removed from menu_main.go; plugin items are
// appended after all built-ins). The cluster is login-gated via the BaseMenu
// ToLoginPage forwarding and reaches the player service (personal_fm's
// BottomOutHook) through the same accessor. Cross-menu jumps into ui stay
// key-based (ranks / daily_playlists -> "playlist_detail", which remains in ui).
package recommend

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
// extraction: daily_songs / daily_playlists / personal_fm / recent_songs /
// ranks. All five are no-arg menus and declare their own main-menu items —
// the built-in 每日推荐歌曲 / 每日推荐歌单 / 私人FM / 最近播放歌曲 / 排行榜
// entries were removed from menu_main.go (plugin items are appended after all
// built-ins).
func init() {
	ui.RegisterMenu("daily_songs", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
		return NewDailyRecommendSongsMenu(base), nil
	})
	ui.RegisterMenu("daily_playlists", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
		return NewDailyRecommendPlaylistMenu(base), nil
	})
	ui.RegisterMenu("personal_fm", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
		return NewPersonalFmMenu(base), nil
	})
	ui.RegisterMenu("recent_songs", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
		return NewRecentSongsMenu(base), nil
	})
	ui.RegisterMenu("ranks", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
		return NewRanksMenu(base), nil
	})
	// 声明主菜单入口：NewMainMenu 在全部内置项之后追加这些项（原先为内置
	// 索引 0/1/4/6/8，现为插件主菜单项，排在全部内置项之后）。
	ui.RegisterMainMenuItem("daily_songs", "每日推荐歌曲")
	ui.RegisterMainMenuItem("daily_playlists", "每日推荐歌单")
	ui.RegisterMainMenuItem("personal_fm", "私人FM")
	ui.RegisterMainMenuItem("recent_songs", "最近播放歌曲")
	ui.RegisterMainMenuItem("ranks", "排行榜")
}
