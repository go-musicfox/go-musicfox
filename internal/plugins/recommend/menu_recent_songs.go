package recommend

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

type RecentSongsMenu struct {
	ui.BaseMenu
	menus []model.MenuItem
	songs []structs.Song
}

func NewRecentSongsMenu(base ui.BaseMenu) *RecentSongsMenu {
	return &RecentSongsMenu{
		BaseMenu: base,
	}
}

func (m *RecentSongsMenu) IsSearchable() bool {
	return true
}

func (m *RecentSongsMenu) IsPlayable() bool {
	return true
}

func (m *RecentSongsMenu) GetMenuKey() string {
	return "recent_songs"
}

func (m *RecentSongsMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *RecentSongsMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		if _struct.CheckUserInfo(m.User()) == _struct.NeedLogin {
			page, _ := m.ToLoginPage(enterMenuCallback(main))
			return false, page
		}

		recentSongService := service.RecordRecentSongsService{}
		code, response, _ := recentSongService.RecordRecentSongs()
		codeType := _struct.CheckCode(code)
		if codeType == _struct.NeedLogin {
			page, _ := m.ToLoginPage(enterMenuCallback(main))
			return false, page
		} else if codeType != _struct.Success {
			return false, nil
		}
		m.songs = _struct.GetRecentSongs(response)
		m.menus = menux.GetViewFromSongs(m.songs)

		return true, nil
	}
}

func (m *RecentSongsMenu) Songs() []structs.Song {
	return m.songs
}
