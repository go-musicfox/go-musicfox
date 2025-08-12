package ui

import (
	"github.com/anhoder/foxful-cli/model"
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
	item    []ActionItem
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
	menuItems := make([]model.MenuItem, 0, len(m.item))
	for _, item := range m.item {
		menuItems = append(menuItems, item.title)
	}
	return menuItems
}

func (m *ActionMenu) SubMenu(app *model.App, index int) model.Menu {
	// FIXME: 快速返回后执行操作，以达到直接调用的（Loading）显示效果。（异步？）
	app.MustMain().BackMenu()
	app.MustMain().RefreshMenuList()
	if m.item[index].action != nil {
		m.item[index].action()
		return nil
	}
	if m.item[index].page != nil {
		return NewMenuToPage(m.baseMenu, m.item[index].page())
	}

	return nil
}

func (m *ActionMenu) BeforeEnterMenuHook() model.Hook {
	// slog.Info(fmt.Sprintf("from: %v, isplaying: %v", m.from, m.playing))

	var (
		main = m.netease.MustMain()
		menu = main.CurMenu()
	)

	switch menu.(type) {
	case SongsMenu:
		if m.playing {
			m.curSongAction()
		} else {
			m.selectSongAction()
		}
	case PlaylistsMenu:
		m.playlistAction()
	default:
		m.curSongAction()
	}
	return nil
}

func (m *ActionMenu) selectSongAction() {
	m.item = []ActionItem{
		{
			title:  model.MenuItem{Title: "所属专辑"},
			action: func() { albumOfSelectedSong(m.netease) },
		}, {
			title:  model.MenuItem{Title: "所属歌手"},
			action: func() { artistOfSelectedSong(m.netease) },
		}, {
			title:  model.MenuItem{Title: "下载"},
			action: func() { downloadSelectedSong(m.netease) },
		}, {
			title: model.MenuItem{Title: "添加到喜欢"},
			page:  func() model.Page { return likeSelectedSong(m.netease, true) },
		}, {
			title: model.MenuItem{Title: "从喜欢移除"},
			page:  func() model.Page { return likeSelectedSong(m.netease, false) },
		}, {
			title: model.MenuItem{Title: "添加至歌单"},
			page:  func() model.Page { return openAddSongToUserPlaylistMenu(m.netease, true, true) },
		}, {
			title: model.MenuItem{Title: "从歌单移除"},
			page:  func() model.Page { return openAddSongToUserPlaylistMenu(m.netease, true, false) },
		}, {
			title:  model.MenuItem{Title: "在网页打开"},
			action: func() { openSelectedItemInWeb(m.netease) },
		}, {
			title: model.MenuItem{Title: "标记为不喜欢"},
			page:  func() model.Page { return trashSelectedSong(m.netease) },
		}, {
			title: model.MenuItem{Title: "相似的歌曲"},
			action:  func() { simiSongsOfSelectedSong(m.netease) },
		},
	}
	if m.from == CurPlaylistKey { // 仅在当前播放界面生效
		m.item = append(m.item, ActionItem{
			title: model.MenuItem{Title: "从播放列表移除"},
			page:  func() model.Page { return delSongFromPlaylist(m.netease) },
		})
	}
}

func (m *ActionMenu) curSongAction() {
	m.item = []ActionItem{
		{
			title:  model.MenuItem{Title: "所属专辑"},
			action: func() { albumOfPlayingSong(m.netease) },
		}, {
			title:  model.MenuItem{Title: "所属歌手"},
			action: func() { artistOfPlayingSong(m.netease) },
		}, {
			title:  model.MenuItem{Title: "下载音乐"},
			action: func() { downloadPlayingSong(m.netease) },
		}, {
			title:  model.MenuItem{Title: "下载歌词"},
			action: func() { downloadPlayingSongLrc(m.netease) },
		}, {
			title: model.MenuItem{Title: "添加到喜欢"},
			page:  func() model.Page { return likePlayingSong(m.netease, true) },
		}, {
			title: model.MenuItem{Title: "从喜欢移除"},
			page:  func() model.Page { return likePlayingSong(m.netease, false) },
		}, {
			title: model.MenuItem{Title: "添加至歌单"},
			page:  func() model.Page { return openAddSongToUserPlaylistMenu(m.netease, false, true) },
		}, {
			title: model.MenuItem{Title: "从歌单移除"},
			page:  func() model.Page { return openAddSongToUserPlaylistMenu(m.netease, false, false) },
		}, {
			title:  model.MenuItem{Title: "在网页打开"},
			action: func() { openPlayingSongInWeb(m.netease) },
		}, {
			title: model.MenuItem{Title: "标记为不喜欢"},
			page:  func() model.Page { return trashPlayingSong(m.netease) },
		}, {
			title: model.MenuItem{Title: "相似的歌曲"},
			action:  func() { simiSongsOfPlayingSong(m.netease) },
		},
	}
}

func (m *ActionMenu) playlistAction() {
	m.item = []ActionItem{
		{
			title: model.MenuItem{Title: "收藏"},
			page:  func() model.Page { return collectSelectedPlaylist(m.netease, true) },
		}, {
			title: model.MenuItem{Title: "取消收藏"},
			page:  func() model.Page { return collectSelectedPlaylist(m.netease, false) },
		},
	}
}
