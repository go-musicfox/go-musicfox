package ui

import "github.com/anhoder/foxful-cli/model"

type SearchTypeMenu struct {
	baseMenu
	menus []model.MenuItem
}

func NewSearchTypeMenu(base baseMenu) *SearchTypeMenu {
	typeMenu := &SearchTypeMenu{
		baseMenu: base,
		menus: []model.MenuItem{
			{Title: "按单曲"},
			{Title: "按专辑"},
			{Title: "按歌手"},
			{Title: "按歌单"},
			{Title: "按用户"},
			{Title: "按歌词"},
			{Title: "按电台"},
		},
	}

	return typeMenu
}

func (m *SearchTypeMenu) GetMenuKey() string {
	return "search_type"
}

func (m *SearchTypeMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *SearchTypeMenu) SubMenu(_ *model.App, index int) model.Menu {
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

	return NewSearchResultMenu(m.baseMenu, typeArr[index])
}
