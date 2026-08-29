package album

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// TestAllMenusBuildViaRegistryFactory proves every moved provider key is
// registered with its original key (identical to the pre-extraction keys in
// registry_registrations.go) and builds through the real registry factory with
// a zero BaseMenu. Parameterized menus keep their dynamic GetMenuKey form.
func TestAllMenusBuildViaRegistryFactory(t *testing.T) {
	base := ui.BaseMenu{}

	noArgKeys := []struct{ key, want string }{
		{"album_menu", "album_menu"},
		{"album_new_area", "album_new_area"},
		{"album_top_area", "album_top_area"},
		{"album_new_hot", "album_new_hot"},
		{"album_sub_list", "album_sub_list"},
	}
	for _, tc := range noArgKeys {
		menu, err := ui.BuildMenu(tc.key, base, ui.NoArgMenuOpts{})
		if err != nil {
			t.Fatalf("BuildMenu(%s) error = %v", tc.key, err)
		}
		if key := menu.GetMenuKey(); key != tc.want {
			t.Fatalf("%s: GetMenuKey() = %q, want %q", tc.key, key, tc.want)
		}
	}

	detail, err := ui.BuildMenu("album_detail", base, ui.AlbumDetailOpts{AlbumID: 456})
	if err != nil {
		t.Fatalf("BuildMenu(album_detail) error = %v", err)
	}
	if key := detail.GetMenuKey(); key != "album_detail_456" {
		t.Fatalf("album_detail GetMenuKey() = %q, want album_detail_456", key)
	}

	top, err := ui.BuildMenu("album_top", base, AlbumTopOpts{Area: "ALL"})
	if err != nil {
		t.Fatalf("BuildMenu(album_top) error = %v", err)
	}
	if key := top.GetMenuKey(); key != "album_top_ALL" {
		t.Fatalf("album_top GetMenuKey() = %q, want album_top_ALL", key)
	}

	newMenu, err := ui.BuildMenu("album_new", base, AlbumNewOpts{Area: "ZH"})
	if err != nil {
		t.Fatalf("BuildMenu(album_new) error = %v", err)
	}
	if key := newMenu.GetMenuKey(); key != "album_new_ZH" {
		t.Fatalf("album_new GetMenuKey() = %q, want album_new_ZH", key)
	}
}

// TestCrossMenuJumpsToAlbumDetail proves the cluster's internal cross-menu
// jumps still resolve through the registry with the same keys: stubbing the
// network-fed albums list and calling SubMenu returns an *AlbumDetailMenu with
// the matching ID.
func TestCrossMenuJumpsToAlbumDetail(t *testing.T) {
	top, err := ui.BuildMenu("album_top", ui.BaseMenu{}, AlbumTopOpts{Area: "ALL"})
	if err != nil {
		t.Fatalf("BuildMenu(album_top) error = %v", err)
	}
	topMenu := top.(*AlbumTopMenu)
	topMenu.albums = []structs.Album{{Id: 456}}
	sub := topMenu.SubMenu(nil, 0)
	detail, ok := sub.(*AlbumDetailMenu)
	if !ok {
		t.Fatalf("album_top.SubMenu(0) = %T, want *AlbumDetailMenu", sub)
	}
	if detail.albumID != 456 {
		t.Fatalf("albumID = %d, want 456", detail.albumID)
	}

	newest, err := ui.BuildMenu("album_new_hot", ui.BaseMenu{}, ui.NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(album_new_hot) error = %v", err)
	}
	newestMenu := newest.(*AlbumNewestMenu)
	newestMenu.albums = []structs.Album{{Id: 457}}
	sub = newestMenu.SubMenu(nil, 0)
	detail, ok = sub.(*AlbumDetailMenu)
	if !ok {
		t.Fatalf("album_new_hot.SubMenu(0) = %T, want *AlbumDetailMenu", sub)
	}
	if detail.albumID != 457 {
		t.Fatalf("albumID = %d, want 457", detail.albumID)
	}
}

// TestAreaMenusBuildSubMenusThroughRegistry proves the area choosers keep
// building their parameterized sub-menus through the registry with the same
// keys (album_new_area -> album_new, album_top_area -> album_top).
func TestAreaMenusBuildSubMenusThroughRegistry(t *testing.T) {
	newArea := NewAlbumNewAreaMenu(ui.BaseMenu{})
	newSub := newArea.SubMenu(nil, 1)
	newMenu, ok := newSub.(*AlbumNewMenu)
	if !ok {
		t.Fatalf("album_new_area.SubMenu(1) = %T, want *AlbumNewMenu", newSub)
	}
	if newMenu.area != "ZH" {
		t.Fatalf("album_new area = %q, want ZH", newMenu.area)
	}

	topArea := NewAlbumTopAreaMenu(ui.BaseMenu{})
	topSub := topArea.SubMenu(nil, 2)
	topMenu, ok := topSub.(*AlbumTopMenu)
	if !ok {
		t.Fatalf("album_top_area.SubMenu(2) = %T, want *AlbumTopMenu", topSub)
	}
	if topMenu.area != "EA" {
		t.Fatalf("album_top area = %q, want EA", topMenu.area)
	}
}

// TestAlbumListEntryMenuListsAllSubMenus proves the 专辑列表 entry menu keeps
// its 3 static entries and builds every sub-menu through the registry (the
// mustBuild* panic path on a broken registration surfaces here as a test
// failure instead of a startup panic).
func TestAlbumListEntryMenuListsAllSubMenus(t *testing.T) {
	menu := NewAlbumListMenu(ui.BaseMenu{})
	if views := menu.MenuViews(); len(views) != 3 {
		t.Fatalf("MenuViews() = %d items, want 3", len(views))
	}
	for i := range menu.menuList {
		if menu.SubMenu(nil, i) == nil {
			t.Fatalf("SubMenu(%d) = nil, want built menu", i)
		}
	}
}

// TestMainMenuItemRegistered proves the entry menu declares its own main-menu
// item (the built-in 专辑列表 entry was removed from menu_main.go, Phase 3.9.x).
func TestMainMenuItemRegistered(t *testing.T) {
	found := false
	for _, item := range ui.MainMenuPluginItems() {
		if item.Key == "album_menu" && item.Title == "专辑列表" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("album_menu main-menu item not registered")
	}
}
