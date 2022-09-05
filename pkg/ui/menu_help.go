package ui

import (
	"go-musicfox/pkg/constants"
	"go-musicfox/utils"
)

type HelpMenu struct {
    menus []MenuItem
}

func NewHelpMenu() *HelpMenu {
    menu := new(HelpMenu)
    menu.menus = []MenuItem{
        {Title: "Enter进来给个star吧, 谢谢~v~"},
        {Title: "r/R", Subtitle: "重新渲染UI"},
        {Title: "h/H/LEFT", Subtitle: "左"},
        {Title: "l/L/RIGHT", Subtitle: "右"},
        {Title: "j/J/DOWN", Subtitle: "下"},
        {Title: "q/Q", Subtitle: "退出"},
        {Title: "SPACE", Subtitle: "播放/暂停"},
        {Title: "[", Subtitle: "上一首"},
        {Title: "]", Subtitle: "下一首"},
        {Title: "-", Subtitle: "减小音量"},
        {Title: "=", Subtitle: "加大音量"},
        {Title: "n/N/ENTER", Subtitle: "进入"},
        {Title: "b/B/ESC", Subtitle: "返回"},
        {Title: "w/W", Subtitle: "注销并退出"},
        {Title: "p", Subtitle: "切换播放模式"},
        {Title: "P", Subtitle: "心动模式"},
        {Title: ",", Subtitle: "喜欢播放中歌曲"},
        {Title: "<", Subtitle: "喜欢选中歌曲"},
        {Title: ".", Subtitle: "取消喜欢播放中歌曲"},
        {Title: ">", Subtitle: "取消喜欢选中歌曲"},
        {Title: "/", Subtitle: "标记播放中歌曲为不喜欢"},
        {Title: "?", Subtitle: "标记选中歌曲为不喜欢"},
    }

    return menu
}

func (m *HelpMenu) MenuData() interface{} {
    return nil
}

func (m *HelpMenu) IsPlayable() bool {
    return false
}

func (m *HelpMenu) ResetPlaylistWhenPlay() bool {
    return false
}

func (m *HelpMenu) GetMenuKey() string {
    return "help_menu"
}

func (m *HelpMenu) MenuViews() []MenuItem {
    return m.menus
}

func (m *HelpMenu) SubMenu(_ *NeteaseModel, index int) IMenu {
    if index == 0 {
        _ = utils.OpenUrl(constants.AppGithubUrl)
    }
    return nil
}

func (m *HelpMenu) BeforePrePageHook() Hook {
    // Nothing to do
    return nil
}

func (m *HelpMenu) BeforeNextPageHook() Hook {
    // Nothing to do
    return nil
}

func (m *HelpMenu) BeforeEnterMenuHook() Hook {
    // Nothing to do
    return nil
}

func (m *HelpMenu) BeforeBackMenuHook() Hook {
    // Nothing to do
    return nil
}

func (m *HelpMenu) BottomOutHook() Hook {
    // Nothing to do
    return nil
}

func (m *HelpMenu) TopOutHook() Hook {
    // Nothing to do
    return nil
}
