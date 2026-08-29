package search

import (
	"github.com/anhoder/foxful-cli/model"

	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

type SearchTypeMenu struct {
	ui.BaseMenu
	menus []model.MenuItem
}

func NewSearchTypeMenu(base ui.BaseMenu) *SearchTypeMenu {
	typeMenu := &SearchTypeMenu{
		BaseMenu: base,
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
	typeArr := []ui.SearchType{
		ui.StSingleSong,
		ui.StAlbum,
		ui.StSinger,
		ui.StPlaylist,
		ui.StUser,
		ui.StLyric,
		ui.StRadio,
	}

	if index >= len(typeArr) {
		return nil
	}

	searchResultMenu, err := ui.BuildMenu("search_result", m.BaseMenu, ui.SearchResultOpts{SearchType: typeArr[index]})
	if err != nil {
		return nil
	}
	return searchResultMenu
}
