package ui

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

type DjProgramRankMenu struct {
	baseMenu
	menus []model.MenuItem
	songs []structs.Song
}

func NewDjProgramRankMenu(base baseMenu) *DjProgramRankMenu {
	return &DjProgramRankMenu{
		baseMenu: base,
	}
}

func (m *DjProgramRankMenu) IsSearchable() bool {
	return true
}

func (m *DjProgramRankMenu) IsPlayable() bool {
	return true
}

func (m *DjProgramRankMenu) GetMenuKey() string {
	return "dj_program_rank"
}

func (m *DjProgramRankMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *DjProgramRankMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {

		djProgramService := service.DjProgramToplistService{
			Limit: "100",
		}
		code, response := djProgramService.DjProgramToplist()
		codeType := _struct.CheckCode(code)
		if codeType != _struct.Success {
			return false, nil
		}
		m.songs = _struct.GetSongsOfDjRank(response)
		m.menus = menux.GetViewFromSongs(m.songs)

		return true, nil
	}
}

func (m *DjProgramRankMenu) Songs() []structs.Song {
	return m.songs
}
