package ui

import (
	"strings"
	"testing"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/structs"
)

// testBase is a zero-value baseMenu: provider construction is svc-free, so
// tests can build menus without a full Netease.
var testBase = baseMenu{}

func assertPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		t.Helper()
		if recover() == nil {
			t.Fatal("expected panic, got none")
		}
	}()
	fn()
}

// withDummyConfig makes notify.Notify safe in tests (it dereferences
// configs.AppConfig before consulting the toast/desktop switches).
func withDummyConfig(t *testing.T) {
	t.Helper()
	previous := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	t.Cleanup(func() { configs.AppConfig = previous })
}

// --- registry unit tests (3.2.3.a) ---

func TestRegisterAndBuildMenu(t *testing.T) {
	menu, err := BuildMenu("playlist_detail", testBase, PlaylistDetailOpts{PlaylistID: 42})
	if err != nil {
		t.Fatalf("BuildMenu(playlist_detail) error = %v", err)
	}
	pd, ok := menu.(*PlaylistDetailMenu)
	if !ok {
		t.Fatalf("BuildMenu(playlist_detail) = %T, want *PlaylistDetailMenu", menu)
	}
	if pd.playlistID != 42 {
		t.Fatalf("playlistId = %d, want 42", pd.playlistID)
	}
	if pd.GetMenuKey() != "playlist_detail_42" {
		t.Fatalf("GetMenuKey() = %q, want %q", pd.GetMenuKey(), "playlist_detail_42")
	}
}

func TestBuildMenuMissingKey(t *testing.T) {
	if _, err := BuildMenu("no_such_menu", testBase, NoArgMenuOpts{}); err == nil {
		t.Fatal("BuildMenu(missing key) error = nil, want error")
	}
}

func TestBuildMenuOptsTypeMismatch(t *testing.T) {
	// "ranks" is registered with NoArgMenuOpts; building it with a different
	// opts type must fail at the single runtime type assertion.
	if _, err := BuildMenu("ranks", testBase, PlaylistDetailOpts{PlaylistID: 1}); err == nil {
		t.Fatal("BuildMenu(ranks, PlaylistDetailOpts) error = nil, want opts type mismatch")
	}
}

func TestRegisterMenuDuplicateKey(t *testing.T) {
	// The key must be unique across the whole test binary; registering it
	// twice is a programmer error caught by the panic.
	const dupKey = "registry_test_dup"
	assertPanics(t, func() {
		RegisterMenu(dupKey, func(base baseMenu, _ NoArgMenuOpts) (Menu, error) { return nil, nil })
		RegisterMenu(dupKey, func(base baseMenu, _ NoArgMenuOpts) (Menu, error) { return nil, nil })
	})
}

func TestRegisterMenuEmptyKeyOrNilFactory(t *testing.T) {
	assertPanics(t, func() {
		RegisterMenu("", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) { return nil, nil })
	})
	assertPanics(t, func() {
		RegisterMenu[NoArgMenuOpts]("registry_test_nil_factory", nil)
	})
}

func TestMustBuildNoArg(t *testing.T) {
	ranks := mustBuildNoArg("ranks", testBase)
	if _, ok := ranks.(*RanksMenu); !ok {
		t.Fatalf("mustBuildNoArg(ranks) = %T, want *RanksMenu", ranks)
	}
	assertPanics(t, func() {
		mustBuildNoArg("no_such_menu", testBase)
	})
}

// --- plugin main-menu items (Phase 3.9) ---

func TestRegisterMainMenuItemValidation(t *testing.T) {
	// Empty key/title never register.
	assertPanics(t, func() {
		RegisterMainMenuItem("", "空 key")
	})
	assertPanics(t, func() {
		RegisterMainMenuItem("empty_title", "")
	})
	// Duplicate key panics; the first registration persists (compile-time
	// registration has no unregister), so pair it with a menu provider to keep
	// later NewMainMenu constructions buildable.
	RegisterMenu[NoArgMenuOpts]("registry_test_dup_main_item", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return &testCheckUpdateMenu{baseMenu: base}, nil
	})
	assertPanics(t, func() {
		RegisterMainMenuItem("registry_test_dup_main_item", "重复")
		RegisterMainMenuItem("registry_test_dup_main_item", "重复")
	})
}

func TestMainMenuPluginItemsSnapshot(t *testing.T) {
	// Snapshot must not alias the internal registry: mutating the returned
	// slice must not pollute later calls.
	before := len(MainMenuPluginItems())
	items := MainMenuPluginItems()
	items = append(items, MainMenuItem{Key: "registry_test_snapshot", Title: "快照"})
	for _, item := range items {
		if item.Key == "registry_test_snapshot" {
			if got := len(MainMenuPluginItems()); got != before {
				t.Fatalf("MainMenuPluginItems() mutated by appending to a snapshot: %d -> %d", before, got)
			}
			return
		}
	}
	t.Fatal("appended item not present in the local snapshot copy")
}

func TestNewMainMenuAppendsPluginItems(t *testing.T) {
	// A test-double plugin menu + main-menu item must be appended after the
	// built-ins (the registry is a package global shared with other tests, so
	// only the presence/position-relative-to-builtins is asserted).
	const builtinMenuCount = 14
	RegisterMenu("registry_test_plugin_menu", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return &testCheckUpdateMenu{baseMenu: base}, nil
	})
	RegisterMainMenuItem("registry_test_plugin_menu", "插件菜单项")

	menu := NewMainMenu(testBase)
	if len(menu.menus) != len(menu.menuList) {
		t.Fatalf("menus=%d menuList=%d, want equal lengths", len(menu.menus), len(menu.menuList))
	}
	pluginIndex := -1
	for i, item := range menu.menus {
		if item.Title == "插件菜单项" {
			pluginIndex = i
			break
		}
	}
	if pluginIndex < 0 {
		t.Fatal("main menu does not contain the plugin item")
	}
	if pluginIndex < builtinMenuCount {
		t.Fatalf("plugin item index = %d, want appended after built-ins (%d+)", pluginIndex, builtinMenuCount)
	}
	if menu.menus[mainMenuHelpIndex].Title != "帮助" {
		t.Fatalf("help entry shifted: menus[%d] = %q", mainMenuHelpIndex, menu.menus[mainMenuHelpIndex].Title)
	}
	if submenu := menu.SubMenu(nil, pluginIndex); submenu == nil {
		t.Fatal("plugin item SubMenu is nil, want the built plugin menu")
	} else if _, ok := submenu.(*testCheckUpdateMenu); !ok {
		t.Fatalf("plugin item SubMenu = %T, want *testCheckUpdateMenu", submenu)
	}

	// A main-menu item whose key has no menu provider is a startup integrity
	// failure (mustBuildNoArg would panic; the explicit assert in NewMainMenu
	// must fire with a clear message). Registered last so no later test in this
	// binary builds a main menu.
	RegisterMainMenuItem("registry_test_missing_menu", "缺失注册")
	assertPanics(t, func() {
		NewMainMenu(testBase)
	})
}

// --- plugin startup hooks (Phase 3.9) ---

func TestRegisterStartupHookRejectsNil(t *testing.T) {
	assertPanics(t, func() {
		RegisterStartupHook(nil)
	})
}

func TestRunStartupHooksPanicIsolation(t *testing.T) {
	var order []string
	RegisterStartupHook(func() { order = append(order, "first") })
	RegisterStartupHook(func() { panic("boom") })
	RegisterStartupHook(func() { order = append(order, "third") })

	runStartupHooks()

	// Registration order preserved, panicking hook skipped without aborting.
	if len(order) != 2 || order[0] != "first" || order[1] != "third" {
		t.Fatalf("startup hooks ran in order %v, want [first third]", order)
	}
}

func TestBuildMenuOrToastMissingKeyReturnsNil(t *testing.T) {
	withDummyConfig(t)
	if menu := buildMenuOrToast("no_such_menu", testBase, NoArgMenuOpts{}); menu != nil {
		t.Fatalf("buildMenuOrToast(missing key) = %v, want nil", menu)
	}
}

func TestBuildPageMissingKey(t *testing.T) {
	if _, err := BuildPage("no_such_page", LoginPageOpts{}); err == nil {
		t.Fatal("BuildPage(missing key) error = nil, want error")
	}
}

func TestBuildPageOrToastMissingKeyReturnsNil(t *testing.T) {
	withDummyConfig(t)
	if page := buildPageOrToast("no_such_page", LoginPageOpts{}); page != nil {
		t.Fatalf("buildPageOrToast(missing key) = %v, want nil", page)
	}
}

// --- page build + navigation smoke (3.3.2) ---

func TestRegisterAndBuildSearchPage(t *testing.T) {
	page, err := BuildPage("search", SearchPageOpts{Netease: nil})
	if err != nil {
		t.Fatalf("BuildPage(search) error = %v", err)
	}
	if _, ok := page.(*SearchPage); !ok {
		t.Fatalf("BuildPage(search) = %T, want *SearchPage", page)
	}
}

// TestPageNavigationSmoke exercises the thin-shell navigation methods through
// the page provider registry: ToLoginPage builds a fresh login page with the
// AfterLogin callback wired; ToSearchPage returns the shell-owned search
// singleton (its shared wordsInput/result/searchType state is read back by the
// SearchResultMenu flow).
func TestPageNavigationSmoke(t *testing.T) {
	withDummyConfig(t)
	n := testNetease()
	n.search = &SearchPage{}

	// ToLoginPage: fresh page through the "login" provider, callback wired.
	callback := func() model.Page { return nil }
	page, cmd := n.ToLoginPage(callback)
	login, ok := page.(*LoginPage)
	if !ok {
		t.Fatalf("ToLoginPage() = %T, want *LoginPage", page)
	}
	if login.AfterLogin == nil {
		t.Fatal("ToLoginPage() did not wire AfterLogin")
	}
	if cmd == nil {
		t.Fatal("ToLoginPage() cmd is nil")
	}

	// ToSearchPage: returns the shell-owned singleton with searchType set.
	spage, scmd := n.ToSearchPage(StSingleSong)
	search, ok := spage.(*SearchPage)
	if !ok {
		t.Fatalf("ToSearchPage() = %T, want *SearchPage", spage)
	}
	if search != n.search {
		t.Fatal("ToSearchPage() did not return the shell-owned search singleton")
	}
	if search.searchType != StSingleSong {
		t.Fatalf("searchType = %d, want %d", search.searchType, StSingleSong)
	}
	if scmd == nil {
		t.Fatal("ToSearchPage() cmd is nil")
	}
}

// --- menu -> menu navigation smoke (3.2.3.b) ---

// TestMenuNavigationSmoke builds navigation chains through the production
// registry (Ranks -> PlaylistDetail; SearchType -> SearchResult -> the demo
// album/artist/user/dj menus) and asserts each constructed menu has the
// expected concrete type, key and non-nil views. Network-fed menu data is
// stubbed directly; the edges exercised are the real SubMenu implementations.
func TestMenuNavigationSmoke(t *testing.T) {
	// Ranks -> PlaylistDetail (SubMenu edge; stub the network-fed ranks).
	ranks, err := BuildMenu("ranks", testBase, NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(ranks) error = %v", err)
	}
	ranksMenu := ranks.(*RanksMenu)
	ranksMenu.ranks = []structs.Rank{{Id: 123}}
	sub := ranksMenu.SubMenu(nil, 0)
	pd, ok := sub.(*PlaylistDetailMenu)
	if !ok {
		t.Fatalf("ranks.SubMenu(0) = %T, want *PlaylistDetailMenu", sub)
	}
	if pd.playlistID != 123 {
		t.Fatalf("playlistId = %d, want 123", pd.playlistID)
	}
	if pd.GetMenuKey() != "playlist_detail_123" {
		t.Fatalf("GetMenuKey() = %q, want %q", pd.GetMenuKey(), "playlist_detail_123")
	}

	// SearchType -> SearchResult -> demo sub-menus (all provider-built).
	st := NewSearchTypeMenu(testBase)
	if st.MenuViews() == nil || len(st.MenuViews()) != 7 {
		t.Fatalf("search type MenuViews() = %v, want 7 static items", st.MenuViews())
	}
	srSub := st.SubMenu(nil, 0)
	sr, ok := srSub.(*SearchResultMenu)
	if !ok {
		t.Fatalf("searchType.SubMenu(0) = %T, want *SearchResultMenu", srSub)
	}
	if !sr.IsSearchable() {
		t.Fatal("search result menu not marked searchable")
	}

	checkSub := func(name string, result any, wantType any) {
		t.Helper()
		sr.result = result
		got := sr.SubMenu(nil, 0)
		switch wantType.(type) {
		case *AlbumDetailMenu:
			if album, ok := got.(*AlbumDetailMenu); !ok || album.albumID != result.([]structs.Album)[0].Id {
				t.Fatalf("%s.SubMenu(0) = %T, want *AlbumDetailMenu(id=%d)", name, got, result.([]structs.Album)[0].Id)
			}
		case *ArtistDetailMenu:
			if artist, ok := got.(*ArtistDetailMenu); !ok || artist.artistId != result.([]structs.Artist)[0].Id {
				t.Fatalf("%s.SubMenu(0) = %T, want *ArtistDetailMenu(id=%d)", name, got, result.([]structs.Artist)[0].Id)
			}
		case *UserPlaylistMenu:
			if user, ok := got.(*UserPlaylistMenu); !ok || user.userID != result.([]structs.User)[0].UserId {
				t.Fatalf("%s.SubMenu(0) = %T, want *UserPlaylistMenu(userId=%d)", name, got, result.([]structs.User)[0].UserId)
			}
		case *DjRadioDetailMenu:
			if dj, ok := got.(*DjRadioDetailMenu); !ok || dj.djRadioId != result.([]structs.DjRadio)[0].Id {
				t.Fatalf("%s.SubMenu(0) = %T, want *DjRadioDetailMenu(id=%d)", name, got, result.([]structs.DjRadio)[0].Id)
			}
		}
	}

	checkSub("album", []structs.Album{{Id: 456}}, (*AlbumDetailMenu)(nil))
	checkSub("artist", []structs.Artist{{Id: 457, Name: "artist"}}, (*ArtistDetailMenu)(nil))
	checkSub("user", []structs.User{{UserId: 458}}, (*UserPlaylistMenu)(nil))
	checkSub("dj", []structs.DjRadio{{Id: 459}}, (*DjRadioDetailMenu)(nil))

	// add_to_user_playlist builds through the registry with its song payload.
	addMenu, err := BuildMenu("add_to_user_playlist", testBase, AddToUserPlaylistOpts{
		UserID: 7,
		Song:   structs.Song{Id: 9},
		IsAdd:  true,
	})
	if err != nil {
		t.Fatalf("BuildMenu(add_to_user_playlist) error = %v", err)
	}
	if _, ok := addMenu.(*AddToUserPlaylistMenu); !ok {
		t.Fatalf("BuildMenu(add_to_user_playlist) = %T, want *AddToUserPlaylistMenu", addMenu)
	}
}

// --- last hardcoded menu constructions (3.3.3) ---

func TestBuildLastHardcodedMenus(t *testing.T) {
	// cur_playlist carries the current playlist snapshot.
	cp, err := BuildMenu("cur_playlist", testBase, CurPlaylistOpts{Songs: []structs.Song{{Id: 1}, {Id: 2}}})
	if err != nil {
		t.Fatalf("BuildMenu(cur_playlist) error = %v", err)
	}
	cpMenu, ok := cp.(*CurPlaylist)
	if !ok {
		t.Fatalf("BuildMenu(cur_playlist) = %T, want *CurPlaylist", cp)
	}
	if len(cpMenu.Songs()) != 2 {
		t.Fatalf("cur_playlist Songs() = %d, want 2", len(cpMenu.Songs()))
	}

	// action_menu carries the originating menu key + playing flag.
	am, err := BuildMenu("action_menu", testBase, ActionMenuOpts{From: "playlist_detail", CurPlaying: true})
	if err != nil {
		t.Fatalf("BuildMenu(action_menu) error = %v", err)
	}
	amMenu, ok := am.(*ActionMenu)
	if !ok {
		t.Fatalf("BuildMenu(action_menu) = %T, want *ActionMenu", am)
	}
	if amMenu.from != "playlist_detail" || !amMenu.playing {
		t.Fatalf("action_menu from=%q playing=%v, want playlist_detail/true", amMenu.from, amMenu.playing)
	}

	// last_fm moved into the internal/plugins/lastfm plugin (Phase 3.9): its
	// build is asserted by the plugin test package.
}

// Login callback ordering (jar-init before user-callback) is covered by
// TestJarInitPrecedesUserCallback in user_service_test.go (3.2.3.c); the
// registry exercises the same LoginPage provider used by that flow via
// TestRegisterAndBuildPage, so no duplication here.

// --- bootstrap completeness assertion (3.2.3.d) ---

func TestAssertMenuRegistryComplete(t *testing.T) {
	// The canonical set is registered by init(); the assertion must pass.
	AssertMenuRegistryComplete(expectedMenuKeys...)

	// A missing key must fail loudly, listing the absent provider.
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("AssertMenuRegistryComplete(missing key) did not panic")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "missing providers") || !strings.Contains(msg, "no_such_key") {
			t.Fatalf("panic message = %v, want it to list missing providers", r)
		}
	}()
	AssertMenuRegistryComplete("no_such_key")
}

func TestAssertPageRegistryComplete(t *testing.T) {
	AssertPageRegistryComplete(expectedPageKeys...)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("AssertPageRegistryComplete(missing key) did not panic")
		}
	}()
	AssertPageRegistryComplete("no_such_page")
}

// --- framework service resolution ---

func TestRegistryServicesResolvable(t *testing.T) {
	ctx := &framework.Context{}
	if err := registerServices(ctx, testNetease()); err != nil {
		t.Fatalf("registerServices() error = %v", err)
	}

	if svc, ok := framework.ServiceOf[MenuRegistry](ctx, ServiceMenuRegistry); !ok || !svc.Registered("playlist_detail") {
		t.Errorf("ServiceOf(menuRegistry) not resolvable or missing playlist_detail")
	}
	if svc, ok := framework.ServiceOf[PageRegistry](ctx, ServicePageRegistry); !ok || !svc.Registered("login") {
		t.Errorf("ServiceOf(pageRegistry) not resolvable or missing login")
	}
}
