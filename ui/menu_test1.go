package ui

type TestMenu struct {}

func (m *TestMenu) BeforeBackMenuHook() Hook {
	return nil
}

func (m *TestMenu) IsPlayable() bool {
	return false
}

func (m *TestMenu) ResetPlaylistWhenEnter() bool {
	return false
}

func (m *TestMenu) GetMenuKey() string {
	return "main_menu"
}

func (m *TestMenu) MenuViews() []string {
	return []string{
		"3213",
		"测试",
		"测试",
		"432432",
		"测试",
		"dasdsa",
		"dsada",
		"测试",
		"3432432",
		"4wfddw",
		"主播电台",
		"帮助",
	}
}

func (m *TestMenu) SubMenu(index int) IMenu {
	return nil
}

func (m *TestMenu) ExtraView() string {
	return ""
}

func (m *TestMenu) BeforePrePageHook() Hook {
	// Nothing to do
	return nil
}

func (m *TestMenu) BeforeNextPageHook() Hook {
	// Nothing to do
	return nil
}

func (m *TestMenu) BeforeEnterMenuHook() Hook {
	// Nothing to do
	return nil
}

func (m *TestMenu) BottomOutHook() Hook {
	// Nothing to do
	return nil
}

func (m *TestMenu) TopOutHook() Hook {
	// Nothing to do
	return nil
}

