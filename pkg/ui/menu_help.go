package ui

import (
	"github.com/skratchdot/open-golang/open"
	"go-musicfox/pkg/constants"
)

type HelpMenu struct {
	DefaultMenu
	menus []MenuItem
}

func NewHelpMenu() *HelpMenu {
	menu := new(HelpMenu)
	menu.menus = []MenuItem{
		{Title: "进来给个star⭐️呗~"},
		{Title: "SPACE", Subtitle: "播放/暂停"},
		{Title: "h/H/LEFT", Subtitle: "左"},
		{Title: "l/L/RIGHT", Subtitle: "右"},
		{Title: "k/K/UP", Subtitle: "上"},
		{Title: "j/J/DOWN", Subtitle: "下"},
		{Title: "g", Subtitle: "上移到顶部"},
		{Title: "G", Subtitle: "下移到底部"},
		{Title: "[", Subtitle: "上一首"},
		{Title: "]", Subtitle: "下一首"},
		{Title: "-", Subtitle: "减小音量"},
		{Title: "=", Subtitle: "加大音量"},
		{Title: "n/N/ENTER", Subtitle: "进入"},
		{Title: "b/B/ESC", Subtitle: "返回"},
		{Title: "q/Q", Subtitle: "退出"},
		{Title: "w/W", Subtitle: "注销并退出"},
		{Title: "p", Subtitle: "切换播放模式"},
		{Title: "P", Subtitle: "心动模式"},
		{Title: ",", Subtitle: "喜欢播放中歌曲"},
		{Title: "<", Subtitle: "喜欢选中歌曲"},
		{Title: ".", Subtitle: "取消喜欢播放中歌曲"},
		{Title: ">", Subtitle: "取消喜欢选中歌曲"},
		{Title: "t", Subtitle: "标记播放中歌曲为不喜欢"},
		{Title: "T", Subtitle: "标记选中歌曲为不喜欢"},
		{Title: "d", Subtitle: "下载播放中音乐"},
		{Title: "D", Subtitle: "下载当前选中音乐"},
		{Title: "c/C", Subtitle: "当前播放列表"},
		{Title: "/", Subtitle: "搜索当前列表"},
		{Title: "r/R", Subtitle: "重新渲染UI"},
	}

	return menu
}

func (m *HelpMenu) GetMenuKey() string {
	return "help_menu"
}

func (m *HelpMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *HelpMenu) SubMenu(_ *NeteaseModel, index int) IMenu {
	if index == 0 {
		_ = open.Start(constants.AppGithubUrl)
	}
	return nil
}
