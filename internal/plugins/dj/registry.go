// Package dj implements the DJ / radio cluster (主播电台) as the third real
// plugin. It is the largest visible cluster extraction: all ten menus moved
// from internal/ui verbatim with their provider keys unchanged — the eight
// sub-menus plus the two parameterized details, and the radio_dj_type entry
// menu which now declares the plugin main-menu item. It demonstrates bulk
// menu extraction with cross-menu jumps: the cluster navigates internally via
// ui.BuildMenuOrToast / ui.MustBuildNoArg / ui.MustBuild with the same keys as
// before, and the ui side (menu_search_result) keeps jumping into
// "dj_radio_detail" unchanged.
package dj

import (
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// DjCategoryDetailOpts is the parameter contract of the "dj_category_detail"
// menu provider (the category radio list). Moved from ui together with the
// cluster — its only consumers are inside this package.
type DjCategoryDetailOpts struct {
	CategoryID int64
}

// DjHotOpts is the parameter contract of the "dj_hot" menu provider
// (hot / not-hot radio lists). Moved from ui together with the cluster.
type DjHotOpts struct {
	HotType DjHotType
}

// init is the compile-time registration entry (linked via the internal/plugins
// aggregator blank import, which cmd/musicfox.go pulls in). Every key is
// identical to the one the menu registered under in internal/ui before the
// extraction: dj_radio_detail / dj_category_detail / dj_category /
// dj_program_rank / dj_program_hour_rank / dj_hot / dj_sub / dj_recommend /
// dj_today_recommend / radio_dj_type. Note "dj_radio_detail" keeps its shared
// opts type in ui (DjRadioDetailOpts — ui's search-result menu also jumps into
// it). The radio_dj_type entry menu declares the main-menu item 主播电台: the
// built-in entry was removed from menu_main.go (plugin items are appended after
// all built-ins).
func init() {
	ui.WithPlugin("dj", "主播电台", func() {
		ui.RegisterMenu("dj_radio_detail", func(base ui.BaseMenu, opts ui.DjRadioDetailOpts) (ui.Menu, error) {
			return NewDjRadioDetailMenu(base, opts.DjRadioID), nil
		})
		ui.RegisterMenu("dj_category_detail", func(base ui.BaseMenu, opts DjCategoryDetailOpts) (ui.Menu, error) {
			return NewDjCategoryDetailMenu(base, opts.CategoryID), nil
		})
		ui.RegisterMenu("dj_category", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
			return NewDjCategoryMenu(base), nil
		})
		ui.RegisterMenu("dj_program_rank", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
			return NewDjProgramRankMenu(base), nil
		})
		ui.RegisterMenu("dj_program_hour_rank", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
			return NewDjProgramHoursRankMenu(base), nil
		})
		ui.RegisterMenu("dj_hot", func(base ui.BaseMenu, opts DjHotOpts) (ui.Menu, error) {
			return NewDjHotMenu(base, opts.HotType), nil
		})
		ui.RegisterMenu("dj_sub", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
			return NewDjSubListMenu(base), nil
		})
		ui.RegisterMenu("dj_recommend", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
			return NewDjRecommendMenu(base), nil
		})
		ui.RegisterMenu("dj_today_recommend", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
			return NewDjTodayRecommendMenu(base), nil
		})
		ui.RegisterMenu("radio_dj_type", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
			return NewRadioDjTypeMenu(base), nil
		})
		// 声明主菜单入口：NewMainMenu 经 After 锚点链归并复现插件化前的主菜单
		// 原始顺序（主播电台跟在云盘（playlist 插件）后、LastFM（lastfm 插件）前）。
		ui.RegisterMainMenuItemAfter("radio_dj_type", "主播电台", "could", nil)
	})
}
