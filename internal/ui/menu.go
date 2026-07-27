package ui

import (
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/structs"
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

type baseMenu struct {
	model.DefaultMenu
	netease *Netease
}

func newBaseMenu(netease *Netease) baseMenu {
	return baseMenu{
		netease: netease,
	}
}

func (e *baseMenu) HelpHints() []model.HelpHint {
	return nil
}

func (e *baseMenu) Action(a *model.App, index int) (model.Page, tea.Cmd) {
	menu := a.MustMain().CurMenu()
	songsMenu, ok := menu.(SongsMenu)
	if !ok || !songsMenu.IsPlayable() {
		return nil, nil
	}

	selectedIndex := menu.RealDataIndex(index)
	if selectedIndex < 0 || selectedIndex >= len(songsMenu.Songs()) {
		return nil, nil
	}

	playOrToggle(e.netease, selectedIndex)
	return nil, a.RerenderCmd(true)
}

func (e *baseMenu) ContextMenuItems(a *model.App, index int) []model.ContextMenuItem {
	main := a.MustMain()
	menu := main.CurMenu()
	if menu.GetMenuKey() == actionMenuKey {
		return nil
	}

	var items []model.ContextMenuItem

	if index >= 0 && index < len(menu.MenuViews()) {
		selActions := actionItemsForMenu(e.netease, menu.GetMenuKey(), false, index)
		if len(selActions) > 0 {
			title := selectedContextTitle(menu, index)
			items = append(items, buildGroupItems("sel", title, selActions, false)...)
		}
	}

	if song, ok := getTargetSong(e.netease, false); ok {
		playActions := actionItemsForMenu(e.netease, menu.GetMenuKey(), true, -1)
		if len(playActions) > 0 {
			title := iconMusicNote + "当前播放：" + songTitleBrief(song.Name)
			items = append(items, buildGroupItems("play", title, playActions, len(items) > 0)...)
		}
	}

	return appendContextMenuGlobalItems(items, len(e.netease.Player().Playlist()) > 0)
}

func (e *baseMenu) ContextMenuAction(a *model.App, index int, item model.ContextMenuItem) (model.Page, tea.Cmd) {
	main := a.MustMain()
	menu := main.CurMenu()
	if menu.GetMenuKey() == actionMenuKey {
		return nil, nil
	}
	if strings.HasPrefix(item.ID, "generic:") {
		return handleGenericContextAction(e.netease, a, item.ID)
	}

	if rest, ok := strings.CutPrefix(item.ID, "sel:"); ok {
		i, err := strconv.Atoi(rest)
		if err != nil {
			return nil, nil
		}
		actions := actionItemsForMenu(e.netease, menu.GetMenuKey(), false, index)
		return runContextAction(actions, i, a)
	}
	if rest, ok := strings.CutPrefix(item.ID, "play:"); ok {
		i, err := strconv.Atoi(rest)
		if err != nil {
			return nil, nil
		}
		actions := actionItemsForMenu(e.netease, menu.GetMenuKey(), true, -1)
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

func handleGenericContextAction(n *Netease, a *model.App, id string) (model.Page, tea.Cmd) {
	main := a.MustMain()
	switch id {
	case "generic:refresh":
		main.RefreshMenuWithLoading()
		return nil, nil
	case "generic:toggle":
		n.player.Toggle()
		return nil, a.RerenderCmd(true)
	case "generic:prev":
		n.player.PreviousSong(true)
		return nil, a.RerenderCmd(true)
	case "generic:next":
		n.player.NextSong(true)
		return nil, a.RerenderCmd(true)
	case "generic:switchTheme":
		registry := configs.CurrentThemeRegistry()
		newSS := registry.NextStyleSet(style.HasDarkBackground())
		if newSS != nil {
			style.SetStyleSet(*newSS)
			a.SetStyleSet(*newSS)
			n.notifyThemeSwitch(a, "切换主题", registry.CurrentName(style.HasDarkBackground()))
		}
		return nil, a.RerenderCmd(true)
	}
	return nil, nil
}

func (e *baseMenu) IsPlayable() bool {
	return false
}

func (e *baseMenu) IsLocatable() bool {
	return true
}
