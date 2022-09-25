package ui

type SearchTypeMenu struct {
	DefaultMenu
	menus []MenuItem
}

func NewSearchTypeMenu() *SearchTypeMenu {
	typeMenu := new(SearchTypeMenu)
	typeMenu.menus = []MenuItem{
		{Title: "按单曲"},
		{Title: "按专辑"},
		{Title: "按歌手"},
		{Title: "按歌单"},
		{Title: "按用户"},
		{Title: "按歌词"},
		{Title: "按电台"},
	}

	return typeMenu
}

func (m *SearchTypeMenu) GetMenuKey() string {
	return "search_type"
}

func (m *SearchTypeMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *SearchTypeMenu) SubMenu(_ *NeteaseModel, index int) IMenu {

	typeArr := []SearchType{
		StSingleSong,
		StAlbum,
		StSinger,
		StPlaylist,
		StUser,
		StLyric,
		StRadio,
	}

	if index >= len(typeArr) {
		return nil
	}

	return NewSearchResultMenu(typeArr[index])
}
