package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
)

const (
	mainMenuHelpIndex        = 14
	mainMenuCheckUpdateIndex = 15
)

type MainMenu struct {
	baseMenu
	menus    []model.MenuItem
	menuList []Menu
}

func NewMainMenu(netease *Netease) *MainMenu {
	base := newBaseMenu(netease)
	mainMenu := &MainMenu{
		baseMenu: base,
		menus: []model.MenuItem{
			{Title: "每日推荐歌曲"},
			{Title: "每日推荐歌单"},
			{Title: "我的歌单"},
			{Title: "我的收藏"},
			{Title: "私人FM"},
			{Title: "专辑列表"},
			{Title: "搜索"},
			{Title: "排行榜"},
			{Title: "精选歌单"},
			{Title: "热门歌手"},
			{Title: "最近播放歌曲"},
			{Title: "云盘"},
			{Title: "主播电台"},
			{Title: "LastFM"},
			{Title: "帮助"},
			{Title: "检查更新"},
		},
		menuList: []Menu{
			NewDailyRecommendSongsMenu(base),
			NewDailyRecommendPlaylistMenu(base),
			NewUserPlaylistMenu(base, CurUser),
			NewUserCollectionMenu(base),
			NewPersonalFmMenu(base),
			NewAlbumListMenu(base),
			NewSearchTypeMenu(base),
			mustBuildMenuB("ranks", base, NoArgMenuOpts{}),
			NewHighQualityPlaylistsMenu(base),
			NewHotArtistsMenu(base),
			NewRecentSongsMenu(base),
			NewCloudMenu(base),
			NewRadioDjTypeMenu(base),
			NewLastfm(base),
			nil, // 帮助由 Action 直接打开 Markdown 弹窗，不再进入子菜单。
			nil, // 检查更新由 Action 异步执行，并直接显示 TUI 通知。
		},
	}
	return mainMenu
}

func (m *MainMenu) FormatMenuItem(item *model.MenuItem) {
	subtitle := "[未登录]"
	if m.svc.User() != nil {
		subtitle = "[" + m.svc.User().Nickname + "]"
	}
	item.Subtitle = subtitle
}

func (m *MainMenu) GetMenuKey() string {
	return "main_menu"
}

func (m *MainMenu) MenuViews() []model.MenuItem {
	for i, menu := range m.menuList {
		if menu != nil {
			menu.FormatMenuItem(&m.menus[i])
		}
	}
	return m.menus
}

func (m *MainMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index < 0 || index >= len(m.menuList) {
		return nil
	}
	return m.menuList[index]
}

func (m *MainMenu) Action(app *model.App, index int) (model.Page, tea.Cmd) {
	switch index {
	case mainMenuHelpIndex:
		showHelpPopup(app)
		return app.MustMain(), nil
	case mainMenuCheckUpdateIndex:
		return app.MustMain(), checkUpdateCmd()
	default:
		return m.baseMenu.Action(app, index)
	}
}
