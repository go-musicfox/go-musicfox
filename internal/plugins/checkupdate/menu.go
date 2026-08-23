// Package checkupdate implements the "检查更新" (check for updates) menu as the
// first real external-style plugin: it embeds ui.BaseMenu (the exported menu
// base), registers through the compile-time registration boundary
// (init + ui.RegisterMenu, see registry.go), and is reached via ui.BuildMenu
// navigation. The check-and-notify logic previously lived in
// internal/ui/menu_check_update.go and is moved here verbatim (the only
// ui dependency, ui.BuildToastNotificationSpec, is the exported toast-spec
// helper). The shell's startup auto-check keeps its own inline copy because ui
// must not import plugins (see docs/plugin_development.md).
package checkupdate

import (
	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/skratchdot/open-golang/open"

	"github.com/go-musicfox/go-musicfox/internal/types"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
	"github.com/go-musicfox/go-musicfox/utils/notify"
	"github.com/go-musicfox/go-musicfox/utils/version"
)

// CheckUpdateMenu 检查更新菜单：嵌入 ui.BaseMenu 获得 ui.Menu 的默认实现，
// 仅覆写 GetMenuKey/MenuViews/SubMenu/Action 与可播放/可定位标记。它不作为
// 子菜单被渲染——主菜单选中「检查更新」时经注册表构建并调用其 Action，触发
// 版本检查并以 TUI 通知展示结果（与提取前行为一致：不跳转，仅弹通知）。
type CheckUpdateMenu struct {
	ui.BaseMenu
}

// GetMenuKey 返回菜单注册 key。
func (m *CheckUpdateMenu) GetMenuKey() string {
	return "check_update"
}

// IsPlayable 检查更新不是歌曲菜单，不可播放。
func (m *CheckUpdateMenu) IsPlayable() bool {
	return false
}

// IsLocatable 检查更新不参与播放自动定位。
func (m *CheckUpdateMenu) IsLocatable() bool {
	return false
}

// MenuViews 检查更新为动作菜单；返回一个静态项，不会作为子菜单被渲染。
func (m *CheckUpdateMenu) MenuViews() []model.MenuItem {
	return []model.MenuItem{{Title: "检查更新"}}
}

// SubMenu 检查更新不产生子菜单。
func (m *CheckUpdateMenu) SubMenu(_ *model.App, _ int) model.Menu {
	return nil
}

// Action 触发检查更新并留在当前页面（原 menu_main 行为：返回主页面 + 异步
// 检查命令，非 nil 的 page/cmd 会跳过子菜单导航）。
func (m *CheckUpdateMenu) Action(a *model.App, _ int) (model.Page, tea.Cmd) {
	return a.MustMain(), checkUpdateCmd()
}

func newVersionNotifyContent(newVersion string) notify.NotifyContent {
	return notify.NotifyContent{
		Title:       "发现新版本: " + newVersion,
		Text:        "去看看呗",
		Url:         types.AppGithubUrl + "/releases/tag/" + newVersion,
		ActionLabel: "前往 GitHub",
	}
}

func checkUpdateCmd() tea.Cmd {
	return func() tea.Msg {
		hasUpdate, latestVersion := version.CheckUpdate()
		return checkUpdateNotificationMsg(hasUpdate, latestVersion)
	}
}

func checkUpdateNotificationMsg(hasUpdate bool, latestVersion string) model.ShowNotificationMsg {
	var spec model.NotificationSpec
	switch {
	case hasUpdate:
		spec = ui.BuildToastNotificationSpec(newVersionNotifyContent(latestVersion), notify.ToastInfo, open.Start)
	case latestVersion != "":
		spec = model.NotificationSpec{
			Level:   model.NotificationSuccess,
			Title:   "检查更新",
			Message: types.AppVersion + " 已是最新版本",
		}
	default:
		spec = model.NotificationSpec{
			Level:   model.NotificationError,
			Title:   "检查更新失败",
			Message: "无法获取最新版本信息，请稍后重试",
		}
	}
	return model.ShowNotificationMsg{Spec: spec}
}
