package ui

type LastfmRes struct {
	DefaultMenu
	err            error
	opName         string
	backLevel      int
	originTitle    string
	originSubTitle string
}

func NewLastfmRes(opName string, err error, backLevel int) *LastfmRes {
	return &LastfmRes{
		opName:    opName,
		err:       err,
		backLevel: backLevel,
	}
}

func (m *LastfmRes) GetMenuKey() string {
	return "last_fm_res"
}

func (m *LastfmRes) MenuViews() []MenuItem {
	return []MenuItem{
		{Title: "返回"},
	}
}

func (m *LastfmRes) SubMenu(model *NeteaseModel, _ int) Menu {
	level := m.backLevel // 避免后续被更新
	for i := 0; i < level; i++ {
		backMenu(model)
	}
	return nil
}

func (m *LastfmRes) BeforeBackMenuHook() Hook {
	return func(model *NeteaseModel) bool {
		m.opName, m.err, m.backLevel = "", nil, 0
		return true
	}
}

func (m *LastfmRes) FormatMenuItem(item *MenuItem) {
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
