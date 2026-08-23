package dj

import (
	"fmt"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

// DjRadioDetailMenu 展示单个电台的节目列表（支持排序切换与滚动加载）。菜单
// 本体已移入本插件包，ui 侧 OpToggleSortOrder 键处理经 ui.DjRadioDetailSortable
// 接口访问（ToggleSortOrder/Reload），不再依赖具体类型。
type DjRadioDetailMenu struct {
	ui.BaseMenu
	menus     []model.MenuItem
	songs     []structs.Song
	djRadioID int64
	limit     int
	offset    int
	total     int
	sortOrder string
}

func NewDjRadioDetailMenu(base ui.BaseMenu, djRadioID int64) *DjRadioDetailMenu {
	return &DjRadioDetailMenu{
		BaseMenu:  base,
		djRadioID: djRadioID,
		limit:     300,
		sortOrder: "true",
	}
}

func (m *DjRadioDetailMenu) ToggleSortOrder() {
	if m.sortOrder == "true" {
		m.sortOrder = "false"
	} else {
		m.sortOrder = "true"
	}
}

func (m *DjRadioDetailMenu) IsSearchable() bool {
	return true
}

func (m *DjRadioDetailMenu) IsPlayable() bool {
	return true
}

func (m *DjRadioDetailMenu) GetMenuKey() string {
	return fmt.Sprintf("dj_radio_detail_%d", m.djRadioID)
}

func (m *DjRadioDetailMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *DjRadioDetailMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		djProgramService := service.DjProgramService{
			RID:    strconv.FormatInt(m.djRadioID, 10),
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
			Asc:    m.sortOrder,
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
			RID:    strconv.FormatInt(m.djRadioID, 10),
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(offset),
			Asc:    m.sortOrder,
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

func (m *DjRadioDetailMenu) DjRadioID() int64 {
	return m.djRadioID
}

func (m *DjRadioDetailMenu) Reload() (bool, model.Page) {
	m.offset = 0
	djProgramService := service.DjProgramService{
		RID:    strconv.FormatInt(m.djRadioID, 10),
		Limit:  strconv.Itoa(m.limit),
		Offset: "0",
		Asc:    m.sortOrder,
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
