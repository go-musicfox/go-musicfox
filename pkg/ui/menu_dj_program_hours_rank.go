package ui

import (
	"github.com/anhoder/netease-music/service"
	"go-musicfox/pkg/structs"
	"go-musicfox/utils"
)

type DjProgramHoursRankMenu struct {
	DefaultMenu
	menus []MenuItem
	songs []structs.Song
}

func NewDjProgramHoursRankMenu() *DjProgramHoursRankMenu {
	return &DjProgramHoursRankMenu{}
}

func (m *DjProgramHoursRankMenu) MenuData() interface{} {
	return m.songs
}

func (m *DjProgramHoursRankMenu) IsPlayable() bool {
	return true
}

func (m *DjProgramHoursRankMenu) GetMenuKey() string {
	return "dj_program_hour_rank"
}

func (m *DjProgramHoursRankMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *DjProgramHoursRankMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {

		djProgramService := service.DjProgramToplistHoursService{
			Limit: "100",
		}
		code, response := djProgramService.DjProgramToplistHours()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false
		}
		m.songs = utils.GetSongsOfDjHoursRank(response)
		m.menus = GetViewFromSongs(m.songs)

		return true
	}
}
