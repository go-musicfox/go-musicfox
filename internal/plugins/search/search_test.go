package search

import (
	"testing"

	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// TestSearchTypeMenuBuildsViaRegistryFactory proves the search_type provider is
// registered with its original key and builds through the real registry factory
// with a zero BaseMenu, keeping its 7 static search-type items.
func TestSearchTypeMenuBuildsViaRegistryFactory(t *testing.T) {
	menu, err := ui.BuildMenu("search_type", ui.BaseMenu{}, ui.NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(search_type) error = %v", err)
	}
	st, ok := menu.(*SearchTypeMenu)
	if !ok {
		t.Fatalf("BuildMenu(search_type) = %T, want *SearchTypeMenu", menu)
	}
	if key := st.GetMenuKey(); key != "search_type" {
		t.Fatalf("GetMenuKey() = %q, want search_type", key)
	}
	if views := st.MenuViews(); len(views) != 7 {
		t.Fatalf("MenuViews() = %v, want 7 static items", views)
	}
}

// TestSearchTypeSubMenuJumpsToSearchResult proves search_type -> search_result
// still resolves through the registry with the same key as before: SubMenu
// returns a *SearchResultMenu built via the search_result provider.
func TestSearchTypeSubMenuJumpsToSearchResult(t *testing.T) {
	menu, err := ui.BuildMenu("search_type", ui.BaseMenu{}, ui.NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(search_type) error = %v", err)
	}
	st := menu.(*SearchTypeMenu)
	sub := st.SubMenu(nil, 0)
	sr, ok := sub.(*SearchResultMenu)
	if !ok {
		t.Fatalf("SearchTypeMenu.SubMenu(0) = %T, want *SearchResultMenu", sub)
	}
	if !sr.IsSearchable() {
		t.Fatal("SearchResultMenu.IsSearchable() = false, want true")
	}
}

// TestSearchResultMenuBuildsViaRegistryFactory proves the search_result provider
// is registered with its original key and builds through the real registry
// factory with the shared ui.SearchResultOpts.
func TestSearchResultMenuBuildsViaRegistryFactory(t *testing.T) {
	menu, err := ui.BuildMenu("search_result", ui.BaseMenu{}, ui.SearchResultOpts{SearchType: ui.StSingleSong})
	if err != nil {
		t.Fatalf("BuildMenu(search_result) error = %v", err)
	}
	if _, ok := menu.(*SearchResultMenu); !ok {
		t.Fatalf("BuildMenu(search_result) = %T, want *SearchResultMenu", menu)
	}
}

// TestSearchPageRegisteredViaForwarding proves the search page provider is
// forwarded to ui.NewSearchPage (the page type and its singleton state stay in
// ui): BuildPage builds a *ui.SearchPage through the real registry factory.
func TestSearchPageRegisteredViaForwarding(t *testing.T) {
	page, err := ui.BuildPage("search", ui.SearchPageOpts{Netease: nil})
	if err != nil {
		t.Fatalf("BuildPage(search) error = %v", err)
	}
	if _, ok := page.(*ui.SearchPage); !ok {
		t.Fatalf("BuildPage(search) = %T, want *ui.SearchPage", page)
	}
}

// TestSearchTypeMainMenuItemRegistered proves the plugin declares its own
// main-menu entry (the built-in 搜索 entry was removed from menu_main.go,
// Phase 3.9.x) anchored after album_menu.
func TestSearchTypeMainMenuItemRegistered(t *testing.T) {
	found := false
	for _, item := range ui.MainMenuPluginItems() {
		if item.Key == "search_type" && item.Title == "搜索" {
			found = true
			if item.After != "album_menu" {
				t.Fatalf("search_type main-menu item After = %q, want album_menu", item.After)
			}
			break
		}
	}
	if !found {
		t.Fatal("search_type main-menu item (搜索) not registered")
	}
}
