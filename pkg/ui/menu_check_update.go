package ui

import (
    "go-musicfox/constants"
    "go-musicfox/utils"
)

type CheckUpdateMenu struct {
    hasUpdate bool
}

func NewCheckUpdateMenu() *CheckUpdateMenu {
    return &CheckUpdateMenu{}
}

func (m *CheckUpdateMenu) MenuData() interface{} {
    return nil
}

func (m *CheckUpdateMenu) BeforeBackMenuHook() Hook {
    return nil
}

func (m *CheckUpdateMenu) IsPlayable() bool {
    return false
}

func (m *CheckUpdateMenu) ResetPlaylistWhenPlay() bool {
    return false
}

func (m *CheckUpdateMenu) GetMenuKey() string {
    return "check_update"
}

func (m *CheckUpdateMenu) MenuViews() []MenuItem {
    if m.hasUpdate {
        return []MenuItem{
            {Title: "检查到新版本，回车查看~", Subtitle: "ENTER"},
        }
    }

    return []MenuItem{
        {Title: "已是最新版本"},
    }
}

func (m *CheckUpdateMenu) SubMenu(_ *NeteaseModel, _ int) IMenu {
    if m.hasUpdate {
        _ = utils.OpenUrl(constants.AppGithubUrl)
    }
    return nil
}

func (m *CheckUpdateMenu) BeforePrePageHook() Hook {
    // Nothing to do
    return nil
}

func (m *CheckUpdateMenu) BeforeNextPageHook() Hook {
    // Nothing to do
    return nil
}

func (m *CheckUpdateMenu) BeforeEnterMenuHook() Hook {
    return func(model *NeteaseModel) bool {
        m.hasUpdate = utils.CheckUpdate()
        return true
    }
}

func (m *CheckUpdateMenu) BottomOutHook() Hook {
    // Nothing to do
    return nil
}

func (m *CheckUpdateMenu) TopOutHook() Hook {
    // Nothing to do
    return nil
}
