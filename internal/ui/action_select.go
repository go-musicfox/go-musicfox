package ui

import (
	"log/slog"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/internal/composer"
	"github.com/go-musicfox/go-musicfox/internal/structs"
)

const actionMenuKey = "action_menu"

// Nerd Font 图标常量（统一使用 Nerd Font v3 Material Design Icons）
const (
	iconAlbum          = "󰀥 " // 专辑
	iconArtist         = "󰠃 " // 歌手/麦克风
	iconHeartFilled    = "󰋑 " // 收藏（实心）
	iconHeartOutline   = "󰋕 " // 取消收藏（空心）
	iconDownload       = "󰇚 " // 下载
	iconDownloadDoc    = "󰈙 " // 下载文档（歌词）
	iconThumbUp        = "󰋍 " // 喜欢
	iconThumbDown      = "󰋓 " // 移除喜欢
	iconThumbBan       = "󰋙 " // 不喜欢
	iconPlaylistAdd    = "󰐕 " // 添加至歌单
	iconPlaylistRemove = "󰝚 " // 从歌单移除
	iconDelete         = "󰆴 " // 删除
	iconSearch         = "󰍉 " // 搜索
	iconShuffle        = "󰒟 " // 相似
	iconShare          = "󰒖 " // 分享
	iconWeb            = "󰖟 " // 网页打开
	iconRefresh        = "󰑐 " // 刷新
	iconPlayPause      = "󰏦 " // 播放/暂停
	iconSkipPrevious   = "󰒮 " // 上一首
	iconSkipNext       = "󰒭 " // 下一首
	iconCrosshairs     = "󰋇 " // 当前选中（目标，通用回退）
	iconMusicNote      = "󰝚 " // 当前播放（音符）
	iconTune           = "󰘳 " // 播放控制（调节）
	iconSong           = "󰎈 " // 歌曲（音符）
	iconPlaylist       = "󰲹 " // 歌单（播放列表）
)

// itemIndent 为分组标题（Header）下的操作项前导缩进，
// 与 Header 文案形成父子层级。放在图标之前（缩进 + icon + 文案）。
const itemIndent = "  "

// TODO: 自适应添加
type ActionItem struct {
	title  model.MenuItem
	action func()
	// menu   model.Menu
	page  func() model.Page
	group string
}

type ActionMenu struct {
	baseMenu
	from        string // 发起 action 的页面
	playing     bool   // 是否针对当前播放
	items       []ActionItem
	playingSong structs.Song // 当前播放
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
	isSelected := !m.playing
	m.buildActionItems()
	if isSelected && len(m.items) == 0 {
		slog.Debug("无针对选择项的操作，改为操作当前播放")
		m.playing = true
		m.buildActionItems()
	}
	return nil
}

func (m *ActionMenu) FormatMenuItem(item *model.MenuItem) {
	if m.playing {
		item.Title = "操作当前播放"
		if m.playingSong.Id != 0 {
			item.Subtitle = m.playingSong.Name
		} else {
			item.Subtitle = "当前无播放"
		}
	}
}

func (m *ActionMenu) buildActionItems() {
	if m.playing {
		var ok bool
		if m.playingSong, ok = getTargetSong(m.netease, false); !ok {
			slog.Debug("无法获取到当前播放歌曲")
			m.items = nil
			return
		}
	}

	m.items = actionItemsForMenu(m.netease, m.from, m.playing, m.netease.MustMain().SelectedIndex())
}

func actionItemsForMenu(n *Netease, from string, playing bool, selectedIndex int) []ActionItem {
	isSelected := !playing
	var actions []ActionItem
	menu := n.MustMain().CurMenu()

	if isSelected && isAlbumsProvider(menu) {
		actions = append(actions, buildAlbumActions(n, selectedIndex)...)
	} else if playing || isSongsProvider(menu) {
		actions = append(actions, buildSongActions(n, isSelected, selectedIndex)...)
	}

	if canCollectPlaylist(menu) {
		actions = append(actions, buildPlaylistActions(n)...)
	}

	if isSelected && from == CurPlaylistKey {
		actions = append(actions, ActionItem{
			title: model.MenuItem{Title: iconDelete + "从播放列表移除"},
			page:  func() model.Page { return delSongFromPlaylist(n) },
			group: "playlist",
		})
	}

	if playing || canShare(menu) {
		actions = append(actions, ActionItem{
			title:  model.MenuItem{Title: iconShare + "分享"},
			action: func() { shareItem(n, isSelected, selectedIndex) },
			group:  "share",
		})
	}

	if playing || canOpenInWeb(menu) {
		actions = append(actions, ActionItem{
			title:  model.MenuItem{Title: iconWeb + "在网页打开"},
			action: func() { openInWeb(n, isSelected, selectedIndex) },
			group:  "share",
		})
	}

	return actions
}

// buildGroupItems 构建一个带标题行的操作分组。
// prefix 用于 ID 前缀（"sel"/"play"），headerTitle 为标题文案，
// leadSeparator 为 true 时在标题前插入分隔线（用于非首组）。
func buildGroupItems(prefix, headerTitle string, actions []ActionItem, leadSeparator bool) []model.ContextMenuItem {
	items := make([]model.ContextMenuItem, 0, len(actions)+2)
	if leadSeparator {
		items = append(items, model.ContextMenuItem{Separator: true})
	}
	items = append(items, model.ContextMenuItem{Header: true, Label: headerTitle})
	for i, action := range actions {
		items = append(items, model.ContextMenuItem{
			ID:    prefix + ":" + strconv.Itoa(i),
			Label: itemIndent + action.title.OriginString(),
		})
	}
	return items
}

// buildGenericGroupItems 构建“播放控制”分组。
func buildGenericGroupItems(leadSeparator bool) []model.ContextMenuItem {
	var items []model.ContextMenuItem
	if leadSeparator {
		items = append(items, model.ContextMenuItem{Separator: true})
	}
	items = append(items, model.ContextMenuItem{Header: true, Label: iconTune + "播放控制"})
	items = append(items,
		model.ContextMenuItem{ID: "generic:toggle", Label: itemIndent + iconPlayPause + "播放/暂停"},
		model.ContextMenuItem{ID: "generic:prev", Label: itemIndent + iconSkipPrevious + "上一首"},
		model.ContextMenuItem{ID: "generic:next", Label: itemIndent + iconSkipNext + "下一首"},
	)
	return items
}

// songTitleBrief 截断过长歌名（超过 20 个 rune 加省略号）。
func songTitleBrief(name string) string {
	r := []rune(name)
	if len(r) > 20 {
		return "「" + string(r[:20]) + "…」"
	}
	return "「" + name + "」"
}

// selectedContextTitle 返回选中项的标题文案，并按当前菜单类型选择区分图标。
func selectedContextTitle(menu model.Menu, index int) string {
	icon := selectedTypeIcon(menu)
	views := menu.MenuViews()
	realIdx := menu.RealDataIndex(index)
	if realIdx >= 0 && realIdx < len(views) {
		return icon + "当前选中：" + songTitleBrief(views[realIdx].Title)
	}
	return icon + "当前选中"
}

// selectedTypeIcon 根据菜单实现的类型接口返回对应的 Nerd Font 图标，
// 用于在右键菜单「当前选中」标题中区分歌曲/歌单/专辑/歌手。
// 无法识别的类型回退到通用的目标图标。
func selectedTypeIcon(menu model.Menu) string {
	switch {
	case isSongsProvider(menu):
		return iconSong
	case isPlaylistsProvider(menu):
		return iconPlaylist
	case isAlbumsProvider(menu):
		return iconAlbum
	case isArtistsProvider(menu):
		return iconArtist
	default:
		return iconCrosshairs
	}
}

func genericContextMenuItems() []model.ContextMenuItem {
	return []model.ContextMenuItem{
		{ID: "generic:refresh", Label: iconRefresh + "刷新当前列表"},
		{ID: "generic:switchTheme", Label: iconTune + "切换主题"},
	}
}

func appendContextMenuGlobalItems(items []model.ContextMenuItem, hasPlaylist bool) []model.ContextMenuItem {
	if hasPlaylist {
		items = append(items, buildGenericGroupItems(len(items) > 0)...)
	}
	// 「刷新当前列表」作为独立菜单项追加到末尾（无分组、无缩进）。
	if len(items) > 0 {
		items = append(items, model.ContextMenuItem{Separator: true})
	}
	return append(items, genericContextMenuItems()...)
}

func buildPlaylistActions(n *Netease) []ActionItem {
	items := []ActionItem{
		{
			title: model.MenuItem{Title: iconHeartFilled + "收藏"},
			page:  func() model.Page { return collectSelectedPlaylist(n, true) },
			group: "subscribe",
		}, {
			title: model.MenuItem{Title: iconHeartOutline + "取消收藏"},
			page:  func() model.Page { return collectSelectedPlaylist(n, false) },
			group: "subscribe",
		},
	}
	return items
}

func buildSongActions(n *Netease, isSelected bool, selectedIndex int) []ActionItem {
	items := []ActionItem{
		{
			title:  model.MenuItem{Title: iconAlbum + "所属专辑"},
			action: func() { goToAlbumOfSong(n, isSelected) },
			group:  "nav",
		},
		{
			title:  model.MenuItem{Title: iconArtist + "所属歌手"},
			action: func() { goToArtistOfSong(n, isSelected) },
			group:  "nav",
		},
		{
			title: model.MenuItem{Title: iconHeartFilled + "收藏歌手"},
			page:  func() model.Page { return subscribeArtist(n, true, isSelected) },
			group: "subscribe",
		},
		{
			title: model.MenuItem{Title: iconHeartOutline + "取消收藏歌手"},
			page:  func() model.Page { return subscribeArtist(n, false, isSelected) },
			group: "subscribe",
		},
		{
			title:  model.MenuItem{Title: iconDownload + "下载"},
			action: func() { downloadSong(n, isSelected) },
			group:  "download",
		},
		{
			title:  model.MenuItem{Title: iconDownloadDoc + "下载歌词"},
			action: func() { downloadSongLrc(n, isSelected) },
			group:  "download",
		},
		{
			title: model.MenuItem{Title: iconThumbUp + "添加到喜欢"},
			page:  func() model.Page { return likeSong(n, true, isSelected) },
			group: "like",
		},
		{
			title: model.MenuItem{Title: iconThumbDown + "从喜欢移除"},
			page:  func() model.Page { return likeSong(n, false, isSelected) },
			group: "like",
		},
		{
			title:  model.MenuItem{Title: iconThumbBan + "标记为不喜欢"},
			action: func() { confirmTrashSong(n, isSelected) },
			group:  "like",
		},
		{
			title: model.MenuItem{Title: iconPlaylistAdd + "添加至歌单"},
			page:  func() model.Page { return openAddSongToUserPlaylistMenu(n, isSelected, true) },
			group: "playlist",
		},
		{
			title: model.MenuItem{Title: iconPlaylistRemove + "从歌单移除"},
			page:  func() model.Page { return openAddSongToUserPlaylistMenu(n, isSelected, false) },
			group: "playlist",
		},
		{
			title:  model.MenuItem{Title: iconShuffle + "相似的歌曲"},
			action: func() { findSimilarSongs(n, isSelected) },
			group:  "discover",
		},
		{
			title:  model.MenuItem{Title: iconSearch + "搜索歌名"},
			action: func() { searchSong(n, isSelected) },
			group:  "discover",
		},
	}

	albumAction := buildAlbumSubscriptionAction(n, isSelected, selectedIndex)
	items = append(items[:2], append([]ActionItem{albumAction}, items[2:]...)...)
	return items
}

func buildAlbumActions(n *Netease, selectedIndex int) []ActionItem {
	albumAction := buildAlbumSubscriptionAction(n, true, selectedIndex)
	return []ActionItem{albumAction}
}

func buildAlbumSubscriptionAction(n *Netease, isSelected bool, selectedIndex int) ActionItem {
	action := ActionItem{
		title: model.MenuItem{Title: iconHeartFilled + "收藏专辑"},
		page:  func() model.Page { return subscribeAlbum(n, true, isSelected) },
		group: "subscribe",
	}
	// Unknown state must never turn an explicit subscribe action into an unsubscribe.
	if album, ok := targetAlbum(n, isSelected, selectedIndex); ok {
		if subscribed, known := n.albumSubscriptions.get(album.Id); known && subscribed {
			action.title.Title = iconHeartOutline + "取消收藏专辑"
			action.page = func() model.Page { return subscribeAlbum(n, false, isSelected) }
		}
	}
	return action
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
	return isPlaylistsProvider(menu)
}

func isSongsProvider(menu model.Menu) bool {
	songs, ok := menu.(SongsMenu)
	return ok && len(songs.Songs()) > 0
}

func isAlbumsProvider(menu model.Menu) bool {
	albums, ok := menu.(AlbumsMenu)
	return ok && len(albums.Albums()) > 0
}

func isPlaylistsProvider(menu model.Menu) bool {
	playlists, ok := menu.(PlaylistsMenu)
	return ok && len(playlists.Playlists()) > 0
}

func isArtistsProvider(menu model.Menu) bool {
	artists, ok := menu.(ArtistsMenu)
	return ok && len(artists.Artists()) > 0
}
