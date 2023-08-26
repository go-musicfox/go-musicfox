package ui

import "github.com/anhoder/foxful-cli/model"

type LastfmRes struct {
	baseMenu
	err            error
	opName         string
	backLevel      int
	originTitle    string
	originSubTitle string
}

func NewLastfmRes(base baseMenu, opName string, err error, backLevel int) *LastfmRes {
	return &LastfmRes{
		baseMenu:  base,
		opName:    opName,
		err:       err,
		backLevel: backLevel,
	}
}

func (m *LastfmRes) GetMenuKey() string {
	return "last_fm_res"
}

func (m *LastfmRes) MenuViews() []model.MenuItem {
	return []model.MenuItem{
		{Title: "返回"},
	}
}

func (m *LastfmRes) SubMenu(app *model.App, _ int) model.Menu {
	level := m.backLevel // 避免后续被更新
	for i := 0; i < level; i++ {
		app.MustMain().BackMenu()
	}
	return nil
}

func (m *LastfmRes) BeforeBackMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		m.opName, m.err, m.backLevel = "", nil, 0
		return true, nil
	}
}

func (m *LastfmRes) FormatMenuItem(item *model.MenuItem) {
	if m.opName == "" {
		item.Title = m.originTitle
		item.Subtitle = m.originSubTitle
		return
	}
	m.originTitle = item.Title
	m.originSubTitle = item.Subtitle
	if m.err != nil {
		item.Title = m.opName + "失败"
		item.Subtitle = "[错误: " + m.err.Error() + "]"
		return
	}
	item.Title = m.opName + "成功"
}
