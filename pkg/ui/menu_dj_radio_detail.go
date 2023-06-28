package ui

import (
	"fmt"
	"strconv"

	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"
)

type DjRadioDetailMenu struct {
	DefaultMenu
	menus     []MenuItem
	songs     []structs.Song
	djRadioId int64
	limit     int
	offset    int
	total     int
}

func NewDjRadioDetailMenu(djRadioId int64) *DjRadioDetailMenu {
	return &DjRadioDetailMenu{
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

func (m *DjRadioDetailMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *DjRadioDetailMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {
		djProgramService := service.DjProgramService{
			RID:    strconv.FormatInt(m.djRadioId, 10),
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := djProgramService.DjProgram()
		utils.Logger().Println(string(response))
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false
		}
		m.songs = utils.GetSongsOfDjRadio(response)
		if total, err := jsonparser.GetInt(response, "count"); err == nil {
			m.total = int(total)
		}
		m.menus = GetViewFromSongs(m.songs)

		return true
	}
}

func (m *DjRadioDetailMenu) BottomOutHook() Hook {
	return func(model *NeteaseModel) bool {
		if len(m.songs) >= m.total {
			return true
		}
		offset := m.offset + m.limit
		djProgramService := service.DjProgramService{
			RID:    strconv.FormatInt(m.djRadioId, 10),
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(offset),
		}
		code, response := djProgramService.DjProgram()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false
		}
		songs := utils.GetSongsOfDjRadio(response)
		m.songs = append(m.songs, songs...)
		m.menus = append(m.menus, GetViewFromSongs(songs)...)
		m.offset = offset

		return true
	}
}

func (m *DjRadioDetailMenu) Songs() []structs.Song {
	return m.songs
}
