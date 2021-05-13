package ui

import (
    "github.com/anhoder/netease-music/service"
    "go-musicfox/ds"
    "go-musicfox/utils"
)

type DjProgramHoursRankMenu struct {
    menus     []MenuItem
    songs     []ds.Song
}

func NewDjProgramHoursRankMenu() *DjProgramHoursRankMenu {
    return &DjProgramHoursRankMenu{}
}

func (m *DjProgramHoursRankMenu) MenuData() interface{} {
    return m.songs
}

func (m *DjProgramHoursRankMenu) BeforeBackMenuHook() Hook {
    return nil
}

func (m *DjProgramHoursRankMenu) IsPlayable() bool {
    return true
}

func (m *DjProgramHoursRankMenu) ResetPlaylistWhenPlay() bool {
    return false
}

func (m *DjProgramHoursRankMenu) GetMenuKey() string {
    return "dj_program_hour_rank"
}

func (m *DjProgramHoursRankMenu) MenuViews() []MenuItem {
    return m.menus
}

func (m *DjProgramHoursRankMenu) SubMenu(_ *NeteaseModel, _ int) IMenu {
    return nil
}

func (m *DjProgramHoursRankMenu) BeforePrePageHook() Hook {
    // Nothing to do
    return nil
}

func (m *DjProgramHoursRankMenu) BeforeNextPageHook() Hook {
    // Nothing to do
    return nil
}

func (m *DjProgramHoursRankMenu) BeforeEnterMenuHook() Hook {
    return func(model *NeteaseModel) bool {

        djProgramService := service.DjProgramToplistHoursService{
            Limit:  "100",
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

func (m *DjProgramHoursRankMenu) BottomOutHook() Hook {
    // Nothing to do
    return nil
}

func (m *DjProgramHoursRankMenu) TopOutHook() Hook {
    // Nothing to do
    return nil
}
