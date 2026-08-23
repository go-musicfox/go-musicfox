package ui

import (
	"fmt"
	"log/slog"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/notify"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

// Production menu/page provider registry — the adjudicated Candidate B shape
// (generic typed registration, Phase 3.0 verdict). Providers are keyed
// factories whose parameter contract lives in an explicit opts type; the ONLY
// runtime type assertion in the whole mechanism is hidden inside
// BuildMenu/BuildPage, never at call sites or in providers.
//
// Phase 3.0 prototypes (registry_proto_a.go / registry_proto_b.go) are kept as
// evidence until the 3.3.4 cleanup; production code does not wire through them.

// --- Menu registry ---

// Parameter-object types (the discoverable "contract") of registered menus.
type (
	PlaylistDetailOpts struct{ PlaylistID int64 }
	ArtistDetailOpts   struct {
		ArtistID int64
		Name     string
	}
	SearchResultOpts      struct{ SearchType SearchType }
	AddToUserPlaylistOpts struct {
		UserID int64
		Song   structs.Song
		IsAdd  bool
	}
	AlbumDetailOpts   struct{ AlbumID int64 }
	UserPlaylistOpts  struct{ UserID int64 }
	DjRadioDetailOpts struct{ DjRadioID int64 }

	// ArtistsOfSongOpts is the parameter contract of the "artist_of_song" menu
	// provider (the artists-of-song menu, multi-artist jump). It stays shared
	// with ui — operate.go's goToArtistOfSong carries the song payload. The
	// following Phase 3.3.1 parameterized menu contracts register under the
	// static key prefix; the runtime menu key keeps its dynamic form (e.g.
	// artist_album_<id>). AlbumTopOpts / AlbumNewOpts moved into the
	// internal/plugins/album plugin with the album cluster (Phase 3.9.x), and
	// ArtistAlbumOpts / ArtistSongOpts moved into internal/plugins/artist with
	// the artist cluster (Phase 3.9.x). ArtistDetailOpts stays shared —
	// search_result / operate jump into that menu.
	ArtistsOfSongOpts struct{ Song structs.Song }
	SimiSongsOpts     struct{ Song structs.Song }

	// CurPlaylistOpts and the following Phase 3.3.3 menu contracts: the last
	// hardcoded constructions (event handler current-playlist, operate action
	// menu, main menu Last.fm entry).
	CurPlaylistOpts struct{ Songs []structs.Song }
	ActionMenuOpts  struct {
		From       string
		CurPlaying bool
	}

	// NoArgMenuOpts is the shared placeholder opts type for no-arg menus.
	NoArgMenuOpts struct{}
)

// menuFactory[T] is a typed menu provider stored behind `any` in the registry.
type menuFactory[T any] struct {
	Key   string
	Build func(base baseMenu, opts T) (Menu, error)
}

// menuRegistry is the production menu provider registry: key -> menuFactory[T].
var menuRegistry = map[string]any{}

// RegisterMenu registers a typed menu provider under key (panics on programmer
// error: empty key, nil factory or duplicate key).
func RegisterMenu[T any](key string, f func(base baseMenu, opts T) (Menu, error)) {
	if key == "" || f == nil {
		panic("RegisterMenu: empty key or nil factory")
	}
	if _, dup := menuRegistry[key]; dup {
		panic("RegisterMenu: duplicate key " + key)
	}
	menuRegistry[key] = menuFactory[T]{Key: key, Build: f}
}

// BuildMenu resolves the typed factory for key and invokes it with opts.
// T is normally inferred from the opts argument; a mismatch with the
// registered T is surfaced here as a runtime error (the ONLY type assertion).
func BuildMenu[T any](key string, base baseMenu, opts T) (Menu, error) {
	factory, ok := menuRegistry[key].(menuFactory[T])
	if !ok {
		return nil, fmt.Errorf("menu %q not registered or opts type mismatch (want %T)", key, opts)
	}
	return factory.Build(base, opts)
}

// mustBuild is the typed equivalent of mustBuildNoArg for parameterized menus
// whose construction sites are static code (menuList initializers): a build
// error with the in-code registry is a programmer error, so panic.
func mustBuild[T any](key string, base baseMenu, opts T) Menu {
	menu, err := BuildMenu(key, base, opts)
	if err != nil {
		panic(fmt.Sprintf("mustBuild(%q): %v", key, err))
	}
	return menu
}

// mustBuildNoArg is the compact helper for no-arg menus (uses NoArgMenuOpts{}).
// Registration errors with the static in-code registry are programmer errors,
// so panic instead of threading an error through a bootstrap constructor.
func mustBuildNoArg(key string, base baseMenu) Menu {
	return mustBuild(key, base, NoArgMenuOpts{})
}

// buildMenuOrToast resolves a menu through the registry for jump sites that
// have no error channel: on error it toasts the failure (registerToastHook
// infra) and returns nil — no panic, matching the existing menu-error behavior.
func buildMenuOrToast[T any](key string, base baseMenu, opts T) Menu {
	menu, err := BuildMenu(key, base, opts)
	if err != nil {
		toastRegistryError("菜单加载失败", err)
		return nil
	}
	return menu
}

// MustBuild is the exported form of mustBuild (Phase 3.9 plugin boundary):
// plugin menu-list initializers build parameterized menus exactly like the
// internal main menu does — a build error against the static in-code registry
// is a programmer error, so panic.
func MustBuild[T any](key string, base baseMenu, opts T) Menu {
	return mustBuild(key, base, opts)
}

// MustBuildNoArg is the exported form of mustBuildNoArg for plugin menu-list
// initializers of no-arg menus (e.g. the DJ/radio cluster entry menu).
func MustBuildNoArg(key string, base baseMenu) Menu {
	return mustBuildNoArg(key, base)
}

// BuildMenuOrToast is the exported form of buildMenuOrToast (Phase 3.9 plugin
// boundary): plugin SubMenu jump sites with no error channel toast the build
// failure and return nil, matching the internal menu-error behavior.
func BuildMenuOrToast[T any](key string, base baseMenu, opts T) Menu {
	return buildMenuOrToast(key, base, opts)
}

// --- Page registry ---

// Parameter-object types (the discoverable "contract") of registered pages.
type (
	LoginPageOpts struct{ Netease *Netease }
	// SearchPageOpts is the Phase 3.3.2 page contract. It builds the
	// shell-owned search singleton (its wordsInput/result/searchType state is
	// shared with the SearchResultMenu flow). The Last.fm page opts
	// (lastfm_auth / lastfm_custom_api) moved into the internal/plugins/lastfm
	// plugin with the pages themselves (Phase 3.9).
	SearchPageOpts struct{ Netease *Netease }
)

// pageFactory[T] is a typed page provider stored behind `any` in the registry.
type pageFactory[T any] struct {
	Key   string
	Build func(opts T) (model.Page, error)
}

// pageRegistry is the production page provider registry: key -> pageFactory[T].
var pageRegistry = map[string]any{}

// RegisterPage registers a typed page provider under key (panics on programmer
// error: empty key, nil factory or duplicate key).
func RegisterPage[T any](key string, f func(opts T) (model.Page, error)) {
	if key == "" || f == nil {
		panic("RegisterPage: empty key or nil factory")
	}
	if _, dup := pageRegistry[key]; dup {
		panic("RegisterPage: duplicate key " + key)
	}
	pageRegistry[key] = pageFactory[T]{Key: key, Build: f}
}

// BuildPage resolves the typed page factory for key and invokes it with opts.
func BuildPage[T any](key string, opts T) (model.Page, error) {
	factory, ok := pageRegistry[key].(pageFactory[T])
	if !ok {
		return nil, fmt.Errorf("page %q not registered or opts type mismatch (want %T)", key, opts)
	}
	return factory.Build(opts)
}

// buildPageOrToast is the page analog of buildMenuOrToast.
func buildPageOrToast[T any](key string, opts T) model.Page {
	page, err := BuildPage(key, opts)
	if err != nil {
		toastRegistryError("页面加载失败", err)
		return nil
	}
	return page
}

// BuildPageOrToast is the exported form of buildPageOrToast (Phase 3.9 plugin
// boundary): plugin page-opening sites outside package ui use it to preserve
// the toast-on-failure + nil-return behavior.
func BuildPageOrToast[T any](key string, opts T) model.Page {
	return buildPageOrToast(key, opts)
}

// toastRegistryError surfaces a provider build failure through the toast
// mechanism (registerToastHook infra), mirroring existing error UX.
func toastRegistryError(title string, err error) {
	notify.Notify(notify.NotifyContent{
		Title: title,
		Text:  err.Error(),
		Level: notify.ToastError,
	})
}

// --- Plugin main-menu items (Phase 3.9) ---

// MainMenuItem is a plugin-declared entry appended to the main menu after the
// built-in items. The Key MUST be registered in the menu registry: NewMainMenu
// asserts it explicitly so an unregistered key surfaces as a startup panic
// (programmer error / startup integrity signal). Build optionally constructs
// the entry menu with the plugin's own options (e.g. a parameterized provider);
// when nil the main menu builds the entry via mustBuildNoArg, so the key must
// then be a no-arg menu provider.
type MainMenuItem struct {
	Key   string
	Title string
	// Build optionally constructs the entry menu from the menu base. nil means
	// the main menu builds the entry via mustBuildNoArg(Key, base).
	Build func(base BaseMenu) Menu
}

// mainMenuPluginItems holds the plugin-declared main-menu items in
// registration order (compile-time registration via init()).
var mainMenuPluginItems []MainMenuItem

// RegisterMainMenuItem appends a plugin main-menu entry with a nil builder
// (the entry menu is built via mustBuildNoArg, so the key must be a registered
// no-arg menu provider). Panics on empty key/title or a duplicate key
// (programmer error).
func RegisterMainMenuItem(key, title string) {
	RegisterMainMenuItemWith(key, title, nil)
}

// RegisterMainMenuItemWith appends a plugin main-menu entry with an optional
// builder. When build is nil the main menu builds the entry via mustBuildNoArg
// (the key must be a registered no-arg menu provider); when non-nil the builder
// constructs the entry menu with its own options, letting a parameterized
// provider serve as a main-menu entry. Panics on empty key/title or a duplicate
// key (programmer error).
func RegisterMainMenuItemWith(key, title string, build func(base BaseMenu) Menu) {
	if key == "" || title == "" {
		panic("RegisterMainMenuItem: empty key or title")
	}
	for _, item := range mainMenuPluginItems {
		if item.Key == key {
			panic("RegisterMainMenuItem: duplicate key " + key)
		}
	}
	mainMenuPluginItems = append(mainMenuPluginItems, MainMenuItem{Key: key, Title: title, Build: build})
}

// MainMenuPluginItems returns a snapshot of the registered plugin main-menu
// items in registration order (NewMainMenu reads it once at construction).
func MainMenuPluginItems() []MainMenuItem {
	items := make([]MainMenuItem, len(mainMenuPluginItems))
	copy(items, mainMenuPluginItems)
	return items
}

// --- Plugin startup hooks (Phase 3.9) ---

// startupHooks are the plugin-registered startup tasks, invoked by the shell's
// InitHook after user/login init (services registered, toast hook wired) at
// the position where the old shell-level startup auto-check ran. Registration
// order is preserved.
var startupHooks []func()

// RegisterStartupHook registers a startup task. The shell calls hooks with the
// app running (services registered); each hook runs with panic isolation (a
// panicking hook is logged and does not crash startup). Panics on a nil hook
// (programmer error).
func RegisterStartupHook(fn func()) {
	if fn == nil {
		panic("RegisterStartupHook: nil hook")
	}
	startupHooks = append(startupHooks, fn)
}

// runStartupHooks invokes the registered startup hooks in order, each wrapped
// in a recover so a panicking hook is logged and skipped without crashing
// startup (framework-hardening house style).
func runStartupHooks() {
	for _, hook := range startupHooks {
		func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("startup hook panicked", slogx.Error(r))
				}
			}()
			hook()
		}()
	}
}

// --- Framework service handles ---

// MenuRegistry is the framework service handle for the menu provider registry.
// Register/Build stay package-level generic functions (Go forbids generic
// methods); this handle makes the registry resolvable as a framework service
// and offers non-generic introspection for completeness assertions and tests.
type MenuRegistry struct{}

// Registered reports whether a provider exists under key.
func (MenuRegistry) Registered(key string) bool {
	_, ok := menuRegistry[key]
	return ok
}

// Keys returns the registered menu provider keys.
func (MenuRegistry) Keys() []string {
	keys := make([]string, 0, len(menuRegistry))
	for k := range menuRegistry {
		keys = append(keys, k)
	}
	return keys
}

// PageRegistry is the framework service handle for the page provider registry.
type PageRegistry struct{}

// Registered reports whether a provider exists under key.
func (PageRegistry) Registered(key string) bool {
	_, ok := pageRegistry[key]
	return ok
}

// Keys returns the registered page provider keys.
func (PageRegistry) Keys() []string {
	keys := make([]string, 0, len(pageRegistry))
	for k := range pageRegistry {
		keys = append(keys, k)
	}
	return keys
}

// --- Startup completeness assertions ---

// expectedMenuKeys is the canonical menu provider key set that must be
// registered at startup (Phase 3.2 bootstrap completeness assertion). Keep in
// sync with the init() registrations in registry_registrations.go. Keys moved
// into plugins (check_update / last_fm / the DJ radio cluster / the album
// cluster / the artist cluster / the recommend cluster) are plugin-supplied
// and intentionally absent — the assertion only locks the built-in set.
var expectedMenuKeys = []string{
	// Phase 3.2 base set + demo migrations.
	"playlist_detail",
	"search_result",
	"add_to_user_playlist",
	"user_playlist",
	"high_quality_playlists",
	"simi_songs",
	"user_collect",
	// Phase 3.3.1 batch 4: main menu cluster + misc.
	"could",
	"search_type",
	"local_search",
	"main_menu",
	// Phase 3.3.3: last hardcoded constructions.
	"cur_playlist",
	"action_menu",
}

// expectedPageKeys is the canonical page provider key set.
var expectedPageKeys = []string{
	"login",
	"search",
}

// AssertMenuRegistryComplete panics listing the missing keys when any expected
// key has no menu provider. Called from the bootstrap (NewNetease) so startup
// fails loudly on an incomplete provider set.
func AssertMenuRegistryComplete(expectedKeys ...string) {
	if missing := missingRegistryKeys(menuRegistry, expectedKeys); len(missing) > 0 {
		panic(fmt.Sprintf("menu registry incomplete, missing providers: %v", missing))
	}
}

// AssertPageRegistryComplete is the page analog of AssertMenuRegistryComplete.
func AssertPageRegistryComplete(expectedKeys ...string) {
	if missing := missingRegistryKeys(pageRegistry, expectedKeys); len(missing) > 0 {
		panic(fmt.Sprintf("page registry incomplete, missing providers: %v", missing))
	}
}

func missingRegistryKeys(registry map[string]any, expected []string) []string {
	var missing []string
	for _, key := range expected {
		if _, ok := registry[key]; !ok {
			missing = append(missing, key)
		}
	}
	return missing
}
