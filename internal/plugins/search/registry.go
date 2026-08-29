// Package search implements the search cluster (搜索类型 / 搜索结果 / 搜索页)
// as the eighth real plugin. The two menus moved from internal/ui verbatim with
// their provider keys unchanged — search_type (搜索) and search_result (搜索
// 结果) — and search_type declares its own main-menu item (the built-in 搜索
// entry was removed from menu_main.go; plugin items are merged into the
// after-anchor chain at the original position, right after the album cluster's
// 专辑列表). The search page keeps its provider key "search" and is forwarded
// to ui.NewSearchPage: the SearchPage type and its wordsInput/result/searchType
// state stay in ui because the shell owns the page as a singleton, shared with
// the SearchResultMenu flow and operate.go. SearchType / St* constants,
// SearchResultOpts and the shared opts (AlbumDetailOpts / PlaylistDetailOpts /
// ArtistDetailOpts / UserPlaylistOpts / DjRadioDetailOpts) stay in ui — the
// search-result jump sites use them. Cross-menu jumps stay key-based
// (search_type -> "search_result"; search_result -> album_detail /
// playlist_detail / artist_detail / user_playlist / dj_radio_detail).
package search

import (
	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/framework"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// Plugin is the search business plugin (P5 cordis shape): its Start registers
// the search menu providers and the 搜索 main-menu entry — the registration
// window moves from package init() to the frontend scope Start.
type Plugin struct {
	framework.NoopPlugin
}

// Start registers the plugin's contributions inside a ui.WithPlugin scope so
// the attribution stamp records them under "search". Every key is identical to
// the one registered under in internal/ui before the extraction: search_type /
// search_result / search (page). search_type is a no-arg menu and declares its
// own main-menu item — the built-in 搜索 entry was removed from menu_main.go;
// the plugin item anchors after album_menu (专辑列表) to keep the original
// position. search_result is parameterized (SearchResultOpts carries the
// search type). The search page is registered by forwarding to ui.NewSearchPage;
// the shell keeps the built page as a singleton.
func (p *Plugin) Start(_ *framework.Context) error {
	ui.WithPlugin("search", "搜索", func() {
		ui.RegisterMenu("search_type", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
			return NewSearchTypeMenu(base), nil
		})
		ui.RegisterMenu("search_result", func(base ui.BaseMenu, opts ui.SearchResultOpts) (ui.Menu, error) {
			return NewSearchResultMenu(base, opts.SearchType), nil
		})
		ui.RegisterPage("search", func(opts ui.SearchPageOpts) (model.Page, error) {
			return ui.NewSearchPage(opts.Netease), nil
		})
		// 声明主菜单入口：NewMainMenu 经 After 锚点链归并复现插件化前的主菜单
		// 原始顺序（搜索跟在专辑列表（album 插件）后）。
		ui.RegisterMainMenuItemAfter("search_type", "搜索", "album_menu", nil)
	})
	return nil
}

// init is the compile-time registration entry (linked via the internal/plugins
// aggregator blank import, which cmd/musicfox.go pulls in) and only declares
// the plugin constructor — actual registrations happen in Start (frontend
// scope).
func init() {
	framework.RegisterPlugin("search", func() framework.Plugin { return &Plugin{} })
}
