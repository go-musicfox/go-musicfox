package ui

import (
	"github.com/anhoder/netease-music/service"
	"go-musicfox/pkg/structs"
	"go-musicfox/utils"
)

type DjProgramRankMenu struct {
    menus     []MenuItem
    songs     []structs.Song
}

func NewDjProgramRankMenu() *DjProgramRankMenu {
    return &DjProgramRankMenu{}
}

func (m *DjProgramRankMenu) MenuData() interface{} {
    return m.songs
}

func (m *DjProgramRankMenu) BeforeBackMenuHook() Hook {
    return nil
}

func (m *DjProgramRankMenu) IsPlayable() bool {
    return true
}

func (m *DjProgramRankMenu) ResetPlaylistWhenPlay() bool {
    return false
}

func (m *DjProgramRankMenu) GetMenuKey() string {
    return "dj_program_rank"
}

func (m *DjProgramRankMenu) MenuViews() []MenuItem {
    return m.menus
}

func (m *DjProgramRankMenu) SubMenu(_ *NeteaseModel, _ int) IMenu {
    return nil
}

func (m *DjProgramRankMenu) BeforePrePageHook() Hook {
    // Nothing to do
    return nil
}

func (m *DjProgramRankMenu) BeforeNextPageHook() Hook {
    // Nothing to do
    return nil
}

func (m *DjProgramRankMenu) BeforeEnterMenuHook() Hook {
    return func(model *NeteaseModel) bool {

        djProgramService := service.DjProgramToplistService{
            Limit:  "100",
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

func (m *DjProgramRankMenu) BottomOutHook() Hook {
    // Nothing to do
    return nil
}

func (m *DjProgramRankMenu) TopOutHook() Hook {
    // Nothing to do
    return nil
}
