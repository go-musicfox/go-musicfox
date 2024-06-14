package ui

import (
	"fmt"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

type DjRadioDetailMenu struct {
	baseMenu
	menus     []model.MenuItem
	songs     []structs.Song
	djRadioId int64
	limit     int
	offset    int
	total     int
}

func NewDjRadioDetailMenu(base baseMenu, djRadioId int64) *DjRadioDetailMenu {
	return &DjRadioDetailMenu{
		baseMenu:  base,
		djRadioId: djRadioId,
		limit:     50,
	}
}

func (m *DjRadioDetailMenu) IsSearchable() bool {
	return true
}

func (m *DjRadioDetailMenu) IsPlayable() bool {
	return true
}

func (m *DjRadioDetailMenu) GetMenuKey() string {
	return fmt.Sprintf("dj_radio_detail_%d", m.djRadioId)
}

func (m *DjRadioDetailMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *DjRadioDetailMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		djProgramService := service.DjProgramService{
			RID:    strconv.FormatInt(m.djRadioId, 10),
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := djProgramService.DjProgram()
		codeType := _struct.CheckCode(code)
		if codeType != _struct.Success {
			return false, nil
		}
		m.songs = _struct.GetSongsOfDjRadio(response)
		if total, err := jsonparser.GetInt(response, "count"); err == nil {
			m.total = int(total)
		}
		m.menus = menux.GetViewFromSongs(m.songs)

		return true, nil
	}
}

func (m *DjRadioDetailMenu) BottomOutHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		if len(m.songs) >= m.total {
			return true, nil
		}
		offset := m.offset + m.limit
		djProgramService := service.DjProgramService{
			RID:    strconv.FormatInt(m.djRadioId, 10),
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(offset),
		}
		code, response := djProgramService.DjProgram()
		codeType := _struct.CheckCode(code)
		if codeType != _struct.Success {
			return false, nil
		}
		songs := _struct.GetSongsOfDjRadio(response)
		m.songs = append(m.songs, songs...)
		m.menus = append(m.menus, menux.GetViewFromSongs(songs)...)
		m.offset = offset

		return true, nil
	}
}

func (m *DjRadioDetailMenu) Songs() []structs.Song {
	return m.songs
}

func (m *DjRadioDetailMenu) DjRadioId() int64 {
	return m.djRadioId
}
