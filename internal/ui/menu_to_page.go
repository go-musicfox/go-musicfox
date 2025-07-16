package ui

import (
	"github.com/anhoder/foxful-cli/model"
)

type MenuToPage struct {
	baseMenu
	page model.Page
}

func NewMenuToPage(base baseMenu, page model.Page) *MenuToPage {
	return &MenuToPage{baseMenu: base, page: page}
}

func (m *MenuToPage) GetMenuKey() string {
	return "menu_to_page"
}

func (m *MenuToPage) MenuViews() []model.MenuItem {
	return []model.MenuItem{}
}

func (m *MenuToPage) SubMenu(app *model.App, index int) model.Menu {
	return nil
}

func (m *MenuToPage) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		return false, m.page
	}
}
