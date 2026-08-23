package ui

import (
	"fmt"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/notify"
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

	// Phase 3.3.1 parameterized menu contracts. Menus whose GetMenuKey() is
	// parameterized (e.g. album_top_<area>) register under the static key
	// prefix; the runtime menu key keeps its dynamic form.
	AlbumTopOpts         struct{ Area string }
	AlbumNewOpts         struct{ Area string }
	ArtistAlbumOpts      struct{ ArtistID int64 }
	ArtistSongOpts       struct{ ArtistID int64 }
	ArtistsOfSongOpts    struct{ Song structs.Song }
	SimiSongsOpts        struct{ Song structs.Song }
	DjCategoryDetailOpts struct{ CategoryID int64 }
	DjHotOpts            struct{ HotType DjHotType }

	// Phase 3.3.3 menu contracts: the last hardcoded constructions (event
	// handler current-playlist, operate action menu, main menu Last.fm entry).
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

// --- Page registry ---

// Parameter-object types (the discoverable "contract") of registered pages.
type (
	LoginPageOpts      struct{ Netease *Netease }
	LastfmAuthPageOpts struct{ Netease *Netease }
	// Phase 3.3.2 page contracts. SearchPageOpts builds the shell-owned search
	// singleton (its wordsInput/result/searchType state is shared with the
	// SearchResultMenu flow); LastfmCustomApiPageOpts is the Last.fm profile
	// "设置 API account" entry.
	SearchPageOpts          struct{ Netease *Netease }
	LastfmCustomApiPageOpts struct{ Netease *Netease }
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

// toastRegistryError surfaces a provider build failure through the toast
// mechanism (registerToastHook infra), mirroring existing error UX.
func toastRegistryError(title string, err error) {
	notify.Notify(notify.NotifyContent{
		Title: title,
		Text:  err.Error(),
		Level: notify.ToastError,
	})
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
// sync with the init() registrations in registry_registrations.go.
var expectedMenuKeys = []string{
	// Phase 3.2 base set + demo migrations.
	"playlist_detail",
	"artist_detail",
	"search_result",
	"add_to_user_playlist",
	"ranks",
	"album_detail",
	"user_playlist",
	"dj_radio_detail",
	// Phase 3.3.1 batch 1: DJ / radio cluster.
	"dj_category_detail",
	"dj_category",
	"dj_program_rank",
	"dj_program_hour_rank",
	"dj_hot",
	"dj_sub",
	"dj_recommend",
	"dj_today_recommend",
	"radio_dj_type",
	// Phase 3.3.1 batch 2: album cluster.
	"album_new_area",
	"album_top_area",
	"album_new_hot",
	"album_menu",
	"album_top",
	"album_new",
	"album_sub_list",
	"high_quality_playlists",
	// Phase 3.3.1 batch 3: artist cluster.
	"artist_of_song",
	"artist_album",
	"artist_song",
	"artists_sub_list",
	"hot_artists",
	"simi_songs",
	"user_collect",
	// Phase 3.3.1 batch 4: main menu cluster + misc.
	"personal_fm",
	"could",
	"recent_songs",
	"daily_playlists",
	"daily_songs",
	"search_type",
	"local_search",
	"main_menu",
	// Phase 3.3.3: last hardcoded constructions.
	"cur_playlist",
	"action_menu",
	"last_fm",
}

// expectedPageKeys is the canonical page provider key set.
var expectedPageKeys = []string{
	"login",
	"lastfm_auth",
	"search",
	"lastfm_custom_api",
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
