package ui

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/pkg/constants"

	"github.com/skratchdot/open-golang/open"
)

type HelpMenu struct {
	baseMenu
	menus []model.MenuItem
}

func NewHelpMenu(base baseMenu) *HelpMenu {
	menu := &HelpMenu{
		baseMenu: base,
		menus: []model.MenuItem{
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
			{Title: "`", Subtitle: "将播放中歌曲加入歌单"},
			{Title: "Tab", Subtitle: "将选中歌曲加入歌单"},
			{Title: "~", Subtitle: "将播放中歌曲从歌单中删除"},
			{Title: "Shift+Tab", Subtitle: "将选中歌曲从歌单中删除"},
			{Title: "t", Subtitle: "标记播放中歌曲为不喜欢"},
			{Title: "T", Subtitle: "标记选中歌曲为不喜欢"},
			{Title: "d", Subtitle: "下载播放中音乐"},
			{Title: "D", Subtitle: "下载当前选中音乐"},
			{Title: "c/C", Subtitle: "当前播放列表"},
			{Title: "r/R", Subtitle: "重新渲染UI"},
			{Title: "/", Subtitle: "搜索当前列表"},
			{Title: "?", Subtitle: "帮助信息"},
			{Title: "a", Subtitle: "播放中歌曲的所属专辑"},
			{Title: "A", Subtitle: "选中歌曲的所属专辑"},
			{Title: "s", Subtitle: "播放中歌曲的所属歌手"},
			{Title: "S", Subtitle: "选中歌曲的所属歌手"},
			{Title: "o", Subtitle: "网页打开播放中歌曲"},
			{Title: "O", Subtitle: "网页打开选中歌曲/专辑..."},
			{Title: "e", Subtitle: "添加为下一曲播放"},
			{Title: "E", Subtitle: "添加到播放列表末尾"},
			{Title: "\\", Subtitle: "从播放列表删除选中歌曲"},
			{Title: "v/V", Subtitle: "快进5s/10s"},
			{Title: "x/X", Subtitle: "快退1s/5s"},
			{Title: ";/:", Subtitle: "收藏选中歌单"},
			{Title: "'/\"", Subtitle: "取消收藏选中歌单"},
			{Title: "u/U", Subtitle: "清除音乐缓存"},
		},
	}

	return menu
}

func (m *HelpMenu) GetMenuKey() string {
	return "help_menu"
}

func (m *HelpMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *HelpMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index == 0 {
		_ = open.Start(constants.AppGithubUrl)
	}
	return nil
}
