package ui

import (
	"time"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils"
)

type DjTodayRecommendMenu struct {
	baseMenu
	menus     []model.MenuItem
	radios    []structs.DjRadio
	fetchTime time.Time
}

func NewDjTodayRecommendMenu(base baseMenu) *DjTodayRecommendMenu {
	return &DjTodayRecommendMenu{
		baseMenu: base,
	}
}

func (m *DjTodayRecommendMenu) IsSearchable() bool {
	return true
}

func (m *DjTodayRecommendMenu) GetMenuKey() string {
	return "dj_today_recommend"
}

func (m *DjTodayRecommendMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *DjTodayRecommendMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.radios) {
		return nil
	}

	return NewDjRadioDetailMenu(m.baseMenu, m.radios[index].Id)
}

func (m *DjTodayRecommendMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		// 不重复请求
		now := time.Now()
		if len(m.menus) > 0 && len(m.radios) > 0 && utils.IsSameDate(m.fetchTime, now) {
			return true, nil
		}

		djTodayService := service.DjTodayPerferedService{}
		code, response := djTodayService.DjTodayPerfered()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false, nil
		}

		m.radios = utils.GetDjRadiosOfToday(response)
		m.menus = utils.GetViewFromDjRadios(m.radios)
		m.fetchTime = now

		return true, nil
	}
}

func (m *DjTodayRecommendMenu) BottomOutHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		djTodayService := service.DjTodayPerferedService{}
		code, response := djTodayService.DjTodayPerfered()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false, nil
		}

		radios := utils.GetDjRadiosOfToday(response)
		menus := utils.GetViewFromDjRadios(radios)

		m.radios = append(m.radios, radios...)
		m.menus = append(m.menus, menus...)

		return true, nil
	}
}
