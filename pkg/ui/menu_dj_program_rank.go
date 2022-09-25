package ui

import (
	"github.com/anhoder/netease-music/service"
	"go-musicfox/pkg/structs"
	"go-musicfox/utils"
)

type DjProgramRankMenu struct {
	DefaultMenu
	menus []MenuItem
	songs []structs.Song
}

func NewDjProgramRankMenu() *DjProgramRankMenu {
	return &DjProgramRankMenu{}
}

func (m *DjProgramRankMenu) MenuData() interface{} {
	return m.songs
}

func (m *DjProgramRankMenu) IsPlayable() bool {
	return true
}

func (m *DjProgramRankMenu) GetMenuKey() string {
	return "dj_program_rank"
}

func (m *DjProgramRankMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *DjProgramRankMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {

		djProgramService := service.DjProgramToplistService{
			Limit: "100",
		}
		code, response := djProgramService.DjProgramToplist()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false
		}
		m.songs = utils.GetSongsOfDjRank(response)
		m.menus = GetViewFromSongs(m.songs)

		return true
	}
}
