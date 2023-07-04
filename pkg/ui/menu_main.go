package ui

type MainMenu struct {
	DefaultMenu
	menus    []MenuItem
	menuList []Menu
	model    *NeteaseModel
}

func NewMainMenu(m *NeteaseModel) *MainMenu {

	mainMenu := new(MainMenu)
	mainMenu.menus = []MenuItem{
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
		{Title: "清除缓存"},
	}
	mainMenu.menuList = []Menu{
		NewDailyRecommendSongsMenu(),
		NewDailyRecommendPlaylistMenu(),
		NewUserPlaylistMenu(CurUser),
		NewPersonalFmMenu(),
		NewAlbumListMenu(),
		NewSearchTypeMenu(),
		NewRanksMenu(),
		NewHighQualityPlaylistsMenu(),
		NewHotArtistsMenu(),
		NewRecentSongsMenu(),
		NewCloudMenu(),
		NewRadioDjTypeMenu(),
		NewLastfm(m),
		NewHelpMenu(),
		NewCheckUpdateMenu(),
		NewClearCacheMenu(),
	}

	mainMenu.model = m

	return mainMenu
}

func (m *MainMenu) FormatMenuItem(item *MenuItem) {
	var subtitle = "[未登录]"
	if m.model.user != nil {
		subtitle = "[" + m.model.user.Nickname + "]"
	}
	item.Subtitle = subtitle
}

func (m *MainMenu) GetMenuKey() string {
	return "main_menu"
}

func (m *MainMenu) MenuViews() []MenuItem {
	for i, menu := range m.menuList {
		menu.FormatMenuItem(&m.menus[i])
	}
	return m.menus
}

func (m *MainMenu) SubMenu(_ *NeteaseModel, index int) Menu {
	if index >= len(m.menuList) {
		return nil
	}

	return m.menuList[index]
}
