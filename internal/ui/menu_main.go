package ui

import (
	"github.com/anhoder/foxful-cli/model"
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
			NewPersonalFmMenu(base),
			NewAlbumListMenu(base),
			NewSearchTypeMenu(base),
			NewRanksMenu(base),
			NewHighQualityPlaylistsMenu(base),
			NewHotArtistsMenu(base),
			NewRecentSongsMenu(base),
			NewCloudMenu(base),
			NewRadioDjTypeMenu(base),
			NewLastfm(base),
			NewHelpMenu(base),
			NewCheckUpdateMenu(base),
		},
	}
	return mainMenu
}

func (m *MainMenu) FormatMenuItem(item *model.MenuItem) {
	var subtitle = "[未登录]"
	if m.netease.user != nil {
		subtitle = "[" + m.netease.user.Nickname + "]"
	}
	item.Subtitle = subtitle
}

func (m *MainMenu) GetMenuKey() string {
	return "main_menu"
}

func (m *MainMenu) MenuViews() []model.MenuItem {
	for i, menu := range m.menuList {
		menu.FormatMenuItem(&m.menus[i])
	}
	return m.menus
}

func (m *MainMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.menuList) {
		return nil
	}

	return m.menuList[index]
}
