package ui

import "time"

type MainMenu struct {}

func (m *MainMenu) IsPlayable() bool {
    return false
}

func (m *MainMenu) ResetPlaylistWhenEnter() bool {
    return false
}

func (m *MainMenu) GetMenuKey() string {
    return "main_menu"
}

func (m *MainMenu) GetSubMenuViews() []string {
    return []string{
        "测试1",
        "测试2",
        "测试3",
        "测试1",
        "测试2",
        "测试3",
        "测试1",
        "测试2",
        "测试3",
        "测试1",
        "测试2",
        "测试3",
        "测试1",
        "测试2",
        "测试3",
    }
}

func (m *MainMenu) SubMenu(index int) IMenu {
    return nil
}

func (m *MainMenu) ExtraView() string {
    return ""
}

func (m *MainMenu) BeforePrePageHook(model *NeteaseModel) {
    time.Sleep(time.Second)
}

func (m *MainMenu) BeforeNextPageHook(model *NeteaseModel) {
    time.Sleep(time.Second)
}

func (m *MainMenu) BeforeEnterMenuHook(model *NeteaseModel) []string {
    return nil
}

func (m *MainMenu) BottomOutHook(model *NeteaseModel) {
    time.Sleep(time.Second)
}

func (m *MainMenu) TopOutHook(model *NeteaseModel) {
    time.Sleep(time.Second)
}

