package ui

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils"
)

type DjProgramHoursRankMenu struct {
	baseMenu
	menus []model.MenuItem
	songs []structs.Song
}

func NewDjProgramHoursRankMenu(base baseMenu) *DjProgramHoursRankMenu {
	return &DjProgramHoursRankMenu{
		baseMenu: base,
	}
}

func (m *DjProgramHoursRankMenu) IsSearchable() bool {
	return true
}

func (m *DjProgramHoursRankMenu) IsPlayable() bool {
	return true
}

func (m *DjProgramHoursRankMenu) GetMenuKey() string {
	return "dj_program_hour_rank"
}

func (m *DjProgramHoursRankMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *DjProgramHoursRankMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		djProgramService := service.DjProgramToplistHoursService{
			Limit: "100",
		}
		code, response := djProgramService.DjProgramToplistHours()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false, nil
		}
		m.songs = utils.GetSongsOfDjHoursRank(response)
		m.menus = utils.GetViewFromSongs(m.songs)

		return true, nil
	}
}

func (m *DjProgramHoursRankMenu) Songs() []structs.Song {
	return m.songs
}
