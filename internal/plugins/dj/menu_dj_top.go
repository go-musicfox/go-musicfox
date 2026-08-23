package dj

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

// DjHotType 热门电台类型（热门/新晋）。
type DjHotType string

const (
	DjHot    DjHotType = "hot"
	DjNotHot DjHotType = "not_hot"
)

// DjHotMenu 展示热门/新晋电台列表；与提取前行为一致（仅嵌入基座换为
// ui.BaseMenu，子菜单跳转经 ui.BuildMenuOrToast 走注册表）。
type DjHotMenu struct {
	ui.BaseMenu
	menus   []model.MenuItem
	radios  []structs.DjRadio
	hotType DjHotType
}

func NewDjHotMenu(base ui.BaseMenu, hotType DjHotType) *DjHotMenu {
	return &DjHotMenu{
		BaseMenu: base,
		hotType:  hotType,
	}
}

func (m *DjHotMenu) IsSearchable() bool {
	return true
}

func (m *DjHotMenu) GetMenuKey() string {
	return "dj_hot"
}

func (m *DjHotMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *DjHotMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.radios) {
		return nil
	}

	return ui.BuildMenuOrToast("dj_radio_detail", m.BaseMenu, ui.DjRadioDetailOpts{DjRadioID: m.radios[index].Id})
}

func (m *DjHotMenu) ItemToShare(index int) any {
	if index >= 0 && index < len(m.radios) {
		return m.radios[index]
	}
	return nil
}

func (m *DjHotMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		// 不重复请求
		if len(m.menus) > 0 && len(m.radios) > 0 {
			return true, nil
		}

		djTopService := service.DjToplistService{
			Type: string(m.hotType),
		}
		code, response := djTopService.DjToplist()
		codeType := _struct.CheckCode(code)
		if codeType != _struct.Success {
			return false, nil
		}

		m.radios = _struct.GetDjRadiosOfTopDj(response)
		m.menus = menux.GetViewFromDjRadios(m.radios)

		return true, nil
	}
}
