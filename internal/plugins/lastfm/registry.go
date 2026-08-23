package lastfm

import (
	"github.com/anhoder/foxful-cli/model"

	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// LastfmAuthPageOpts is the parameter contract of the "lastfm_auth" page
// provider. It carries the ui.MenuServices accessor through which the page
// resolves the shell (MustMain/App) and services (Lastfm). The opts type lives
// in the plugin with the page — the registration moved here with it (Phase
// 3.9), and the field is exported so the accessor can be passed from any
// package.
type LastfmAuthPageOpts struct {
	Svc ui.MenuServices
}

// LastfmCustomAPIPageOpts is the parameter contract of the
// "lastfm_custom_api" page provider (the Last.fm profile "设置 API account"
// entry).
type LastfmCustomAPIPageOpts struct {
	Svc ui.MenuServices
}

// init is the compile-time registration entry (linked via the internal/plugins
// aggregator blank import, which cmd/musicfox.go pulls in). The Last.fm menu
// keeps its original key "last_fm"; the two pages keep "lastfm_auth" /
// "lastfm_custom_api"; the plugin declares its own main-menu entry, so the
// built-in main menu entry for LastFM was removed (plugin items are appended
// after all built-ins). The QR page is not registered — it is reached only from
// the auth page (authByQRCode) inside this plugin.
func init() {
	ui.RegisterMenu("last_fm", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
		return NewLastfm(base), nil
	})
	ui.RegisterPage("lastfm_auth", func(opts LastfmAuthPageOpts) (model.Page, error) {
		return NewLastfmAuthPage(opts.Svc), nil
	})
	ui.RegisterPage("lastfm_custom_api", func(opts LastfmCustomAPIPageOpts) (model.Page, error) {
		return NewLastfmCustomAPIPage(opts.Svc), nil
	})
	// 声明主菜单入口：NewMainMenu 经 After 锚点链归并复现插件化前的主菜单
	// 原始顺序（LastFM 跟在主播电台（dj 插件）后、帮助（内置）前）。
	ui.RegisterMainMenuItemAfter("last_fm", "LastFM", "radio_dj_type", nil)
}
