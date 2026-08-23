package ui

import (
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"

	"github.com/go-musicfox/go-musicfox/internal/composer"
	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/desktop_lyrics"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/lastfm"
	"github.com/go-musicfox/go-musicfox/internal/lyric"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/track"
)

// Menu menu interface
type Menu interface {
	model.Menu

	// IsPlayable 当前菜单是否可播放？
	IsPlayable() bool

	// IsLocatable 当前菜单是否支持播放自动定位
	IsLocatable() bool
}

// DjMenu dj menu interface
type DjMenu interface {
	Menu
}

// DjRadioDetailSortable is implemented by the DJ radio detail menu (the
// concrete type now lives in the internal/plugins/dj plugin). The
// OpToggleSortOrder key handler toggles the program sort order through this
// interface instead of a concrete ui type — package ui must not import plugin
// packages.
type DjRadioDetailSortable interface {
	ToggleSortOrder()
	Reload() (bool, model.Page)
}

type SongsMenu interface {
	Menu
	Songs() []structs.Song
}

type PlaylistsMenu interface {
	Menu
	Playlists() []structs.Playlist
}

type AlbumsMenu interface {
	Menu
	Albums() []structs.Album
}

type ArtistsMenu interface {
	Menu
	Artists() []structs.Artist
}

// BaseMenu is the embeddable base for menus. It is exported so external plugin
// factories (outside package ui) can embed it and implement the ui.Menu
// interface; it carries the menuServices accessor through which plugins reach
// services and the thin-shell navigation surface (see the forwarding methods
// below and docs/plugin_development.md). baseMenu is an alias, so all existing
// internal code and registry registrations keep compiling untouched.
type BaseMenu struct {
	model.DefaultMenu
	svc *menuServices
}

// baseMenu is the unexported alias kept for internal call sites and the
// RegisterMenu/BuildMenu signatures (aliases are interchangeable, so factories
// written with BaseMenu are accepted as-is).
type baseMenu = BaseMenu

func newBaseMenu(netease *Netease) baseMenu {
	return baseMenu{
		svc: newMenuServices(netease),
	}
}

// NewBaseMenu builds a BaseMenu rooted at the given Netease shell. Exported for
// the app bootstrap (internal/commands), whose call sites predate the baseMenu
// signature and only hold the shell.
func NewBaseMenu(n *Netease) BaseMenu {
	return newBaseMenu(n)
}

// newBaseMenuFromSvc builds a baseMenu from an existing accessor, avoiding a
// round trip through the *Netease shell (Phase 3.3.3 menu constructions).
func newBaseMenuFromSvc(svc *menuServices) baseMenu {
	return baseMenu{svc: svc}
}

// --- Exported accessor forwarding (plugin boundary, Phase 3.6). External
// plugin menus embed BaseMenu and reach services/navigation through these
// methods; each forwards to the menuServices accessor and is nil-safe (a
// zero-value BaseMenu degrades to the zero value without a panic).

// Player resolves the player service.
func (e *BaseMenu) Player() *Player {
	if e.svc == nil {
		return nil
	}
	return e.svc.Player()
}

// User resolves the current user (nil until login).
func (e *BaseMenu) User() *structs.User {
	if e.svc == nil {
		return nil
	}
	return e.svc.User()
}

// TrackManager resolves the track manager service.
func (e *BaseMenu) TrackManager() *track.Manager {
	if e.svc == nil {
		return nil
	}
	return e.svc.TrackManager()
}

// LyricService resolves the lyric service.
func (e *BaseMenu) LyricService() *lyric.Service {
	if e.svc == nil {
		return nil
	}
	return e.svc.LyricService()
}

// DesktopLyrics resolves the desktop lyrics controller.
func (e *BaseMenu) DesktopLyrics() desktop_lyrics.Controller {
	if e.svc == nil {
		return nil
	}
	return e.svc.DesktopLyrics()
}

// CoverRenderer resolves the cover renderer.
func (e *BaseMenu) CoverRenderer() *CoverRenderer {
	if e.svc == nil {
		return nil
	}
	return e.svc.CoverRenderer()
}

// ShareSvc resolves the share service.
func (e *BaseMenu) ShareSvc() *composer.ShareService {
	if e.svc == nil {
		return nil
	}
	return e.svc.ShareSvc()
}

// Lastfm resolves the Last.fm client.
func (e *BaseMenu) Lastfm() *lastfm.Client {
	if e.svc == nil {
		return nil
	}
	return e.svc.Lastfm()
}

// Ctx returns the app-wide framework context backing the accessor.
func (e *BaseMenu) Ctx() *framework.Context {
	if e.svc == nil {
		return nil
	}
	return e.svc.Ctx()
}

// App returns the foxful app shell (nil when unset).
func (e *BaseMenu) App() *model.App {
	if e.svc == nil {
		return nil
	}
	return e.svc.App()
}

// MustMain returns the foxful main page (nil when the app has not started).
func (e *BaseMenu) MustMain() *model.Main {
	if e.svc == nil {
		return nil
	}
	return e.svc.MustMain()
}

// Rerender returns a tea.Cmd that forces a re-render on the app shell.
func (e *BaseMenu) Rerender() tea.Cmd {
	if e.svc == nil || e.svc.App() == nil {
		return nil
	}
	return e.svc.App().RerenderCmd(true)
}

// Search returns the shell-owned search page singleton (nil when unset).
func (e *BaseMenu) Search() *SearchPage {
	if e.svc == nil {
		return nil
	}
	return e.svc.Search()
}

// ToLoginPage forwards to the thin-shell login navigation.
func (e *BaseMenu) ToLoginPage(callback func() model.Page) (model.Page, tea.Cmd) {
	if e.svc == nil {
		return nil, nil
	}
	return e.svc.ToLoginPage(callback)
}

// ToSearchPage forwards to the thin-shell search navigation.
func (e *BaseMenu) ToSearchPage(searchType SearchType) (model.Page, tea.Cmd) {
	if e.svc == nil {
		return nil, nil
	}
	return e.svc.ToSearchPage(searchType)
}

// Services returns the underlying menuServices accessor (exported alias).
// Plugins pass it into page-opts fields and constructors that require the
// accessor type — the unexported svc field itself is unreachable outside ui.
func (e *BaseMenu) Services() MenuServices {
	return e.svc
}

// Netease returns the Netease shell. Escape hatch for legacy helper calls that
// still take *Netease; new external plugin code should prefer the accessor
// methods above.
func (e *BaseMenu) Netease() *Netease {
	if e.svc == nil {
		return nil
	}
	return e.svc.Netease()
}

func (e *BaseMenu) HelpHints() []model.HelpHint {
	return nil
}

func (e *BaseMenu) Action(a *model.App, index int) (model.Page, tea.Cmd) {
	menu := a.MustMain().CurMenu()
	songsMenu, ok := menu.(SongsMenu)
	if !ok || !songsMenu.IsPlayable() {
		return nil, nil
	}

	selectedIndex := menu.RealDataIndex(index)
	if selectedIndex < 0 || selectedIndex >= len(songsMenu.Songs()) {
		return nil, nil
	}

	playOrToggle(e.svc, selectedIndex)
	return nil, a.RerenderCmd(true)
}

func (e *BaseMenu) ContextMenuItems(a *model.App, index int) []model.ContextMenuItem {
	main := a.MustMain()
	menu := main.CurMenu()
	if menu.GetMenuKey() == actionMenuKey {
		return nil
	}

	var items []model.ContextMenuItem

	if index >= 0 && index < len(menu.MenuViews()) {
		selActions := actionItemsForMenu(e.svc, menu.GetMenuKey(), false, index)
		if len(selActions) > 0 {
			title := selectedContextTitle(menu, index)
			items = append(items, buildGroupItems("sel", title, selActions, false)...)
		}
	}

	if song, ok := getTargetSong(e.svc, false); ok {
		playActions := actionItemsForMenu(e.svc, menu.GetMenuKey(), true, -1)
		if len(playActions) > 0 {
			title := iconMusicNote + "当前播放：" + songTitleBrief(song.Name)
			items = append(items, buildGroupItems("play", title, playActions, len(items) > 0)...)
		}
	}

	return appendContextMenuGlobalItems(items, len(e.svc.Player().Playlist()) > 0)
}

func (e *BaseMenu) ContextMenuAction(a *model.App, index int, item model.ContextMenuItem) (model.Page, tea.Cmd) {
	main := a.MustMain()
	menu := main.CurMenu()
	if menu.GetMenuKey() == actionMenuKey {
		return nil, nil
	}
	if strings.HasPrefix(item.ID, "generic:") {
		return handleGenericContextAction(e.svc, a, item.ID)
	}

	if rest, ok := strings.CutPrefix(item.ID, "sel:"); ok {
		i, err := strconv.Atoi(rest)
		if err != nil {
			return nil, nil
		}
		actions := actionItemsForMenu(e.svc, menu.GetMenuKey(), false, index)
		return runContextAction(actions, i, a)
	}
	if rest, ok := strings.CutPrefix(item.ID, "play:"); ok {
		i, err := strconv.Atoi(rest)
		if err != nil {
			return nil, nil
		}
		actions := actionItemsForMenu(e.svc, menu.GetMenuKey(), true, -1)
		return runContextAction(actions, i, a)
	}
	return nil, nil
}

func runContextAction(actions []ActionItem, i int, a *model.App) (model.Page, tea.Cmd) {
	if i < 0 || i >= len(actions) {
		return nil, nil
	}
	action := actions[i]
	if action.page != nil {
		return action.page(), nil
	}
	if action.action != nil {
		action.action()
		return nil, a.RerenderCmd(true)
	}
	return nil, nil
}

func handleGenericContextAction(svc *menuServices, a *model.App, id string) (model.Page, tea.Cmd) {
	main := a.MustMain()
	switch id {
	case "generic:refresh":
		main.RefreshMenuWithLoading()
		return nil, nil
	case "generic:toggle":
		svc.Player().Toggle()
		return nil, a.RerenderCmd(true)
	case "generic:prev":
		svc.Player().PreviousSong(true)
		return nil, a.RerenderCmd(true)
	case "generic:next":
		svc.Player().NextSong(true)
		return nil, a.RerenderCmd(true)
	case "generic:switchTheme":
		registry := configs.CurrentThemeRegistry()
		newSS := registry.NextStyleSet(style.HasDarkBackground())
		if newSS != nil {
			syncActiveThemePair(a, registry)
			style.SetStyleSet(*newSS)
			a.SetStyleSet(*newSS)
			themeName := registry.CurrentName(style.HasDarkBackground())
			svc.SaveActiveTheme(themeName)
			svc.NotifyThemeSwitch(a, "切换主题", themeName)
		}
		return nil, a.RerenderCmd(true)
	}
	return nil, nil
}

func (e *BaseMenu) IsPlayable() bool {
	return false
}

func (e *BaseMenu) IsLocatable() bool {
	return true
}
