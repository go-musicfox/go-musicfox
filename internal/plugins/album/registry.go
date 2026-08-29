// Package album implements the album cluster (专辑列表) as the fourth real
// plugin. It is the second bulk cluster extraction after the DJ/radio cluster:
// all eight album menus moved from internal/ui verbatim with their provider
// keys unchanged — album_menu (the 专辑列表 entry menu, which now declares the
// plugin main-menu item), the two area choosers (album_new_area / album_top_area),
// the three album lists (album_new / album_top / album_new_hot), the subscribed
// album list (album_sub_list) and the shared album_detail menu. The cluster
// navigates internally via ui.BuildMenuOrToast / ui.MustBuildNoArg with the
// same keys as before, and the ui side (search_result / artist_album / operate
// goToAlbumOfSong) keeps jumping into "album_detail" unchanged.
package album

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

// AlbumTopOpts is the parameter contract of the "album_top" menu provider
// (top albums by area). Moved from ui together with the cluster — its only
// consumers are inside this package.
type AlbumTopOpts struct {
	Area string
}

// AlbumNewOpts is the parameter contract of the "album_new" menu provider
// (new albums by area). Moved from ui together with the cluster.
type AlbumNewOpts struct {
	Area string
}

// Plugin is the album business plugin (P5 cordis shape): its Start registers
// the eight album menu providers and the 专辑列表 main-menu entry — the
// registration window moves from package init() to the frontend scope Start.
type Plugin struct {
	framework.NoopPlugin
}

// Start registers the plugin's contributions inside a ui.WithPlugin scope so
// the attribution stamp records them under "album". Every key is identical to
// the one the menu registered under in internal/ui before the extraction:
// album_menu / album_new_area / album_top_area / album_new_hot /
// album_new / album_top / album_sub_list / album_detail. Note "album_detail"
// keeps its shared opts type in ui (ui.AlbumDetailOpts — ui's search-result
// menu, artist-album menu and operate.go goToAlbumOfSong also jump into it).
// The album_menu entry menu declares the main-menu item 专辑列表: the built-in
// entry was removed from menu_main.go.
func (p *Plugin) Start(_ *framework.Context) error {
	ui.WithPlugin("album", "专辑列表", func() {
		ui.RegisterMenu("album_menu", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
			return NewAlbumListMenu(base), nil
		})
		ui.RegisterMenu("album_new_area", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
			return NewAlbumNewAreaMenu(base), nil
		})
		ui.RegisterMenu("album_top_area", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
			return NewAlbumTopAreaMenu(base), nil
		})
		ui.RegisterMenu("album_new_hot", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
			return NewAlbumNewestMenu(base), nil
		})
		ui.RegisterMenu("album_new", func(base ui.BaseMenu, opts AlbumNewOpts) (ui.Menu, error) {
			return NewAlbumNewMenu(base, opts.Area), nil
		})
		ui.RegisterMenu("album_top", func(base ui.BaseMenu, opts AlbumTopOpts) (ui.Menu, error) {
			return NewAlbumTopMenu(base, opts.Area), nil
		})
		ui.RegisterMenu("album_sub_list", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
			return NewAlbumSubscribeListMenu(base), nil
		})
		ui.RegisterMenu("album_detail", func(base ui.BaseMenu, opts ui.AlbumDetailOpts) (ui.Menu, error) {
			return NewAlbumDetailMenu(base, opts.AlbumID), nil
		})
		// 声明主菜单入口：NewMainMenu 经 After 锚点链归并复现插件化前的主菜单
		// 原始顺序（专辑列表跟在私人FM（recommend 插件）后、搜索（search 插件）前）。
		ui.RegisterMainMenuItemAfter("album_menu", "专辑列表", "personal_fm", nil)
	})
	return nil
}

// init is the compile-time registration entry (linked via the internal/plugins
// aggregator blank import, which cmd/musicfox.go pulls in) and only declares
// the plugin constructor — actual registrations happen in Start (frontend
// scope).
func init() {
	framework.RegisterPlugin("album", func() framework.Plugin { return &Plugin{} })
}
