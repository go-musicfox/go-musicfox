package plugins

import (
	"testing"

	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// TestBlankImportRegistersCheckUpdateMenu proves the aggregator's blank-import
// chain (cmd/musicfox.go → internal/plugins → internal/plugins/checkupdate)
// links the plugin and triggers its init() registration: the "check_update"
// provider is visible on the registry, BuildMenu resolves it through the
// exported base, and the plugin-declared main-menu entry is appended.
func TestBlankImportRegistersCheckUpdateMenu(t *testing.T) {
	if !(ui.MenuRegistry{}).Registered("check_update") {
		t.Fatal("check_update menu provider not registered via aggregator blank import")
	}
	menu, err := ui.BuildMenu("check_update", ui.BaseMenu{}, ui.NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(check_update) error = %v", err)
	}
	if key := menu.GetMenuKey(); key != "check_update" {
		t.Fatalf("GetMenuKey() = %q, want check_update", key)
	}

	// The plugin also declares its main-menu entry (Phase 3.9).
	found := false
	for _, item := range ui.MainMenuPluginItems() {
		if item.Key == "check_update" && item.Title == "检查更新" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("check_update main-menu item not registered via aggregator blank import")
	}
}
