package ui

import (
	"github.com/anhoder/foxful-cli/model"
)

type ConfirmItem struct {
	title     model.MenuItem
	action    func()
	backLevel int
}

type ConfirmMenu struct {
	baseMenu
	item []ConfirmItem
}

// NewConfirmMenu 确认操作
func NewConfirmMenu(base baseMenu, item []ConfirmItem) *ConfirmMenu {
	return &ConfirmMenu{
		baseMenu: base,
		item:     item,
	}
}

func (m *ConfirmMenu) GetMenuKey() string {
	return "confirm_menu"
}

func (m *ConfirmMenu) MenuViews() []model.MenuItem {
	menuItems := make([]model.MenuItem, 0, len(m.item))
	for _, item := range m.item {
		menuItems = append(menuItems, item.title)
	}
	return menuItems
}

func (m *ConfirmMenu) SubMenu(app *model.App, index int) model.Menu {
	m.item[index].action()
	for range m.item[index].backLevel {
		app.MustMain().BackMenu()
	}
	app.MustMain().RefreshMenuList()
	return nil
}
