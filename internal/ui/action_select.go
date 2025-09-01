package ui

import (
	"log/slog"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/internal/composer"
)

const actionMenuKey = "action_menu"

// TODO: 自适应添加
type ActionItem struct {
	title  model.MenuItem
	action func()
	// menu   model.Menu
	page func() model.Page
}

type ActionMenu struct {
	baseMenu
	from    string // 发起 action 的页面
	playing bool   // 是否针对当前播放
	items    []ActionItem
}

// NewActionMenu 新建操作页
func NewActionMenu(base baseMenu, from string, curPlaying bool) *ActionMenu {
	return &ActionMenu{
		baseMenu: base,
		from:     from,
		playing:  curPlaying,
	}
}

func (m *ActionMenu) GetMenuKey() string {
	return actionMenuKey
}

func (m *ActionMenu) MenuViews() []model.MenuItem {
	menuItems := make([]model.MenuItem, 0, len(m.items))
	for _, item := range m.items {
		menuItems = append(menuItems, item.title)
	}
	return menuItems
}

func (m *ActionMenu) SubMenu(app *model.App, index int) model.Menu {
	// FIXME: 快速返回后执行操作，以达到直接调用的（Loading）显示效果。（异步？）
	app.MustMain().BackMenu()
	app.MustMain().RefreshMenuList()
	if m.items[index].action != nil {
		m.items[index].action()
		return nil
	}
	if m.items[index].page != nil {
		return NewMenuToPage(m.baseMenu, m.items[index].page())
	}

	return nil
}

func (m *ActionMenu) BeforeEnterMenuHook() model.Hook {
	m.buildActionItems()
	if !m.playing && len(m.items) == 0 {
		slog.Debug("无针对选择项的操作，改为操作当前播放")
		m.playing = true
		m.buildActionItems()
	}
	return nil
}

func (m *ActionMenu) FormatMenuItem(item *model.MenuItem) {
	if m.playing {
		item.Title = "操作当前播放"
		if song, ok := getTargetSong(m.netease, false); ok {
			item.Subtitle = song.Name
		}
	}
}

func (m *ActionMenu) buildActionItems() {
	isSelected := !m.playing
	var actions []ActionItem
	main := m.netease.MustMain()
	menu := main.CurMenu()

	if m.playing || isSongsProvider(menu) {
		actions = append(actions, buildSongActions(m.netease, isSelected)...)
	}

	if canCollectPlaylist(menu) {
		actions = append(actions, buildPlaylistActions(m.netease)...)
	}

	if isSelected && m.from == CurPlaylistKey {
		actions = append(actions, ActionItem{
			title: model.MenuItem{Title: "从播放列表移除"},
			page:  func() model.Page { return delSongFromPlaylist(m.netease) },
		})
	}

	if m.playing || canShare(menu) {
		actions = append(actions, ActionItem{
			title:  model.MenuItem{Title: "分享"},
			action: func() { shareItem(m.netease, isSelected) },
		})
	}

	if m.playing || canOpenInWeb(menu) {
		actions = append(actions, ActionItem{
			title:  model.MenuItem{Title: "在网页打开"},
			action: func() { openInWeb(m.netease, isSelected) },
		})
	}

	m.items = actions
}

func buildPlaylistActions(n *Netease) []ActionItem {
	items := []ActionItem{
		{
			title: model.MenuItem{Title: "收藏"},
			page:  func() model.Page { return collectSelectedPlaylist(n, true) },
		}, {
			title: model.MenuItem{Title: "取消收藏"},
			page:  func() model.Page { return collectSelectedPlaylist(n, false) },
		},
	}
	return items
}

func buildSongActions(n *Netease, isSelected bool) []ActionItem {
	items := []ActionItem{
		{
			title:  model.MenuItem{Title: "所属专辑"},
			action: func() { goToAlbumOfSong(n, isSelected) },
		},
		{
			title:  model.MenuItem{Title: "所属歌手"},
			action: func() { goToArtistOfSong(n, isSelected) },
		},
		{
			title:  model.MenuItem{Title: "下载"},
			action: func() { downloadSong(n, isSelected) },
		},
		{
			title:  model.MenuItem{Title: "下载歌词"},
			action: func() { downloadSongLrc(n, isSelected) },
		},
		{
			title: model.MenuItem{Title: "添加到喜欢"},
			page:  func() model.Page { return likeSong(n, true, isSelected) },
		},
		{
			title: model.MenuItem{Title: "从喜欢移除"},
			page:  func() model.Page { return likeSong(n, false, isSelected) },
		},
		{
			title: model.MenuItem{Title: "添加至歌单"},
			page:  func() model.Page { return openAddSongToUserPlaylistMenu(n, isSelected, true) },
		},
		{
			title: model.MenuItem{Title: "从歌单移除"},
			page:  func() model.Page { return openAddSongToUserPlaylistMenu(n, isSelected, false) },
		},
		{
			title: model.MenuItem{Title: "标记为不喜欢"},
			page:  func() model.Page { return trashSong(n, isSelected) },
		},
		{
			title:  model.MenuItem{Title: "相似的歌曲"},
			action: func() { findSimilarSongs(n, isSelected) },
		},
	}
	return items
}

func canShare(menu model.Menu) bool {
	if _, ok := menu.(composer.Sharer); ok {
		return true
	}
	switch menu.(type) {
	case SongsMenu, AlbumsMenu, ArtistsMenu, PlaylistsMenu:
		return true
	default:
		return false
	}
}

func canOpenInWeb(menu model.Menu) bool {
	// 判断逻辑一致
	return canShare(menu)
}

func canCollectPlaylist(menu model.Menu) bool {
	_, ok := menu.(PlaylistsMenu)
	return ok
}

func isSongsProvider(menu model.Menu) bool {
	_, ok := menu.(SongsMenu)
	return ok
}
