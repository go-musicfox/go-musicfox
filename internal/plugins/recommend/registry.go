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

	"github.com/go-musicfox/go-musicfox/internal/framework"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// enterMenuCallback mirrors ui.EnterMenuCallback (unexported there): the login
// callback re-enters the requesting menu once login succeeds.
func enterMenuCallback(main *model.Main) ui.LoginCallback {
	return func() model.Page {
		return main.EnterMenu(nil, nil)
	}
}

// Plugin is the recommend business plugin (P5 cordis shape): its Start
// registers the five recommend menu providers and their main-menu entries —
// the registration window moves from package init() to the frontend scope
// Start.
type Plugin struct {
	framework.NoopPlugin
}

// Start registers the plugin's contributions inside a ui.WithPlugin scope so
// the attribution stamp records them under "recommend". Every key is identical
// to the one the menu registered under in internal/ui before the extraction:
// daily_songs / daily_playlists / personal_fm / recent_songs / ranks. All five
// are no-arg menus and declare their own main-menu items — the built-in
// 每日推荐歌曲 / 每日推荐歌单 / 私人FM / 最近播放歌曲 / 排行榜 entries were
// removed from menu_main.go.
func (p *Plugin) Start(_ *framework.Context) error {
	ui.WithPlugin("recommend", "每日推荐", func() {
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
		// 声明主菜单入口：NewMainMenu 经 After 锚点链归并复现插件化前的主菜单
		// 原始顺序。每个入口声明其前驱项 key：每日推荐歌曲跟在链首（MainMenuStart）
		// 之后，每日推荐歌单跟在每日推荐歌曲后，私人FM 跟在我的收藏（playlist 插件）
		// 后，排行榜跟在搜索（search 插件）后，最近播放歌曲跟在热门歌手（artist 插件）后。
		ui.RegisterMainMenuItemAfter("daily_songs", "每日推荐歌曲", ui.MainMenuStart, nil)
		ui.RegisterMainMenuItemAfter("daily_playlists", "每日推荐歌单", "daily_songs", nil)
		ui.RegisterMainMenuItemAfter("personal_fm", "私人FM", "user_collect", nil)
		ui.RegisterMainMenuItemAfter("recent_songs", "最近播放歌曲", "hot_artists", nil)
		ui.RegisterMainMenuItemAfter("ranks", "排行榜", "search_type", nil)
	})
	return nil
}

// init is the compile-time registration entry (linked via the internal/plugins
// aggregator blank import, which cmd/musicfox.go pulls in) and only declares
// the plugin constructor — actual registrations happen in Start (frontend
// scope).
func init() {
	framework.RegisterPlugin("recommend", func() framework.Plugin { return &Plugin{} })
}
