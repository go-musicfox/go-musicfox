package dj

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
		{"radio_dj_type", "radio_dj_type"},
		{"dj_sub", "dj_sub"},
		{"dj_recommend", "dj_recommend"},
		{"dj_today_recommend", "dj_today_recommend"},
		{"dj_category", "dj_category"},
		{"dj_program_rank", "dj_program_rank"},
		{"dj_program_hour_rank", "dj_program_hour_rank"},
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

	radio, err := ui.BuildMenu("dj_radio_detail", base, ui.DjRadioDetailOpts{DjRadioID: 459})
	if err != nil {
		t.Fatalf("BuildMenu(dj_radio_detail) error = %v", err)
	}
	if key := radio.GetMenuKey(); key != "dj_radio_detail_459" {
		t.Fatalf("dj_radio_detail GetMenuKey() = %q, want dj_radio_detail_459", key)
	}

	cate, err := ui.BuildMenu("dj_category_detail", base, DjCategoryDetailOpts{CategoryID: 7})
	if err != nil {
		t.Fatalf("BuildMenu(dj_category_detail) error = %v", err)
	}
	if key := cate.GetMenuKey(); key != "dj_category_detail_7" {
		t.Fatalf("dj_category_detail GetMenuKey() = %q, want dj_category_detail_7", key)
	}

	for _, hotType := range []DjHotType{DjHot, DjNotHot} {
		hot, err := ui.BuildMenu("dj_hot", base, DjHotOpts{HotType: hotType})
		if err != nil {
			t.Fatalf("BuildMenu(dj_hot) error = %v", err)
		}
		if key := hot.GetMenuKey(); key != "dj_hot" {
			t.Fatalf("dj_hot GetMenuKey() = %q, want dj_hot", key)
		}
	}
}

// TestCrossMenuJumpsToRadioDetail proves the cluster's internal cross-menu
// jumps still resolve through the registry with the same keys: stubbing the
// network-fed radios list and calling SubMenu returns a *DjRadioDetailMenu
// with the matching ID (the behavior previously asserted from ui's
// TestMenuNavigationSmoke before the extraction).
func TestCrossMenuJumpsToRadioDetail(t *testing.T) {
	rec, err := ui.BuildMenu("dj_recommend", ui.BaseMenu{}, ui.NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(dj_recommend) error = %v", err)
	}
	recMenu := rec.(*DjRecommendMenu)
	recMenu.radios = []structs.DjRadio{{Id: 459}}
	sub := recMenu.SubMenu(nil, 0)
	dj, ok := sub.(*DjRadioDetailMenu)
	if !ok {
		t.Fatalf("dj_recommend.SubMenu(0) = %T, want *DjRadioDetailMenu", sub)
	}
	if dj.djRadioID != 459 {
		t.Fatalf("djRadioID = %d, want 459", dj.djRadioID)
	}
}

// TestRadioDjTypeEntryMenuListsAllSubMenus proves the 主播电台 entry menu keeps
// its 8 static entries and builds every sub-menu through the registry (the
// mustBuild* panic path on a broken registration surfaces here as a test
// failure instead of a startup panic).
func TestRadioDjTypeEntryMenuListsAllSubMenus(t *testing.T) {
	menu := NewRadioDjTypeMenu(ui.BaseMenu{})
	if views := menu.MenuViews(); len(views) != 8 {
		t.Fatalf("MenuViews() = %d items, want 8", len(views))
	}
	for i := range menu.menuList {
		if menu.SubMenu(nil, i) == nil {
			t.Fatalf("SubMenu(%d) = nil, want built menu", i)
		}
	}
}

// TestMainMenuItemRegistered proves the entry menu declares its own main-menu
// item (the built-in 主播电台 entry was removed from menu_main.go, Phase 3.9.x).
func TestMainMenuItemRegistered(t *testing.T) {
	found := false
	for _, item := range ui.MainMenuPluginItems() {
		if item.Key == "radio_dj_type" && item.Title == "主播电台" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("radio_dj_type main-menu item not registered")
	}
}
