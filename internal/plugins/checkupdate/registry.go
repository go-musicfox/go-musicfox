package checkupdate

import (
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// init is the compile-time registration entry: it runs when the plugin package
// is linked into the binary (via the internal/plugins aggregator blank import,
// which cmd/musicfox.go pulls in). The factory's base parameter is written as
// ui.BaseMenu — the registry signature uses its alias baseMenu, and the two are
// interchangeable (BaseMenu is exported exactly for external factories).
func init() {
	ui.RegisterMenu("check_update", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
		return &CheckUpdateMenu{BaseMenu: base}, nil
	})
}
