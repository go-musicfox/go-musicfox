// Package lastfm implements the Last.fm cluster as the second real
// external-style plugin. It demonstrates the plugin ecosystem beyond the
// checkupdate menu: service access via the exported accessor (svc.Lastfm()),
// page plugins (lastfm_auth / lastfm_custom_api registered through
// ui.RegisterPage) and a plugin main-menu item (ui.RegisterMainMenuItem).
// The menu and pages embed/use exported ui boundary types (ui.BaseMenu /
// ui.MenuServices) — no unexported ui symbols. Everything moved verbatim from
// internal/ui/lastfm*.go with only the access path changed (m.svc.X →
// m.X() via BaseMenu forwarding, page opts carrying ui.MenuServices).
package lastfm

import (
	"fmt"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/types"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
	"github.com/go-musicfox/go-musicfox/utils/notify"
)

// Lastfm 是 Last.fm 功能菜单：管理授权、前往主页、启停 Scrobble、清空队列。
// 嵌入 ui.BaseMenu（导出基座），服务经 BaseMenu 转发方法解析（m.Lastfm()），
// 确认弹窗经 ui.ShowConfirmPopup（导出的弹窗助手）。
type Lastfm struct {
	ui.BaseMenu
}

// NewLastfm builds the Last.fm menu from the exported base.
func NewLastfm(base ui.BaseMenu) *Lastfm {
	return &Lastfm{BaseMenu: base}
}

func (m *Lastfm) GetMenuKey() string {
	return "last_fm"
}

func (m *Lastfm) MenuViews() []model.MenuItem {
	getControlTitle := func() string {
		if m.Lastfm().Tracker.Status() {
			return "关闭功能"
		}
		return "启用功能"
	}

	return []model.MenuItem{
		{Title: "管理授权"},
		{Title: "前往主页"},
		{Title: getControlTitle()},
		{Title: "清空队列", Subtitle: fmt.Sprintf("[共 %d 条]", m.Lastfm().Tracker.Count())},
	}
}

func (m *Lastfm) SubMenu(app *model.App, index int) model.Menu {
	switch index {
	case 0:
		return NewLastfmProfile(m.BaseMenu)
	case 1:
		m.Lastfm().OpenUserHomePage()
	case 2:
		turningOn := !m.Lastfm().Tracker.Status()
		title := "关闭 Last.fm 功能"
		content := "确定关闭 Last.fm Scrobble 功能吗？"
		if turningOn {
			title = "启用 Last.fm 功能"
			content = "确定启用 Last.fm Scrobble 功能吗？"
		}
		ui.ShowConfirmPopup(app, title, content, func() {
			m.Lastfm().Tracker.Toggle()
			m.MustMain().RefreshMenuList()
		})
		return nil
	case 3:
		count := m.Lastfm().Tracker.Count()
		ui.ShowConfirmPopup(app, "清空 Last.fm 队列", fmt.Sprintf("确定清空 Last.fm Scrobble 队列吗？（共 %d 条）", count), func() {
			m.Lastfm().Tracker.Clear()
			notify.Notify(notify.NotifyContent{
				Title:   "清除 last.fm Scrobble 队列成功",
				Text:    "Last.fm Scrobble 队列已清除",
				GroupId: types.GroupID,
			})
			m.MustMain().RefreshMenuList()
		})
		return nil
	}
	return nil
}

func (m *Lastfm) FormatMenuItem(item *model.MenuItem) {
	item.Subtitle = "[未授权]"
	if !m.Lastfm().NeedAuth() {
		if username := m.Lastfm().UserName(); username != "" {
			item.Subtitle = fmt.Sprintf("[%s]", username)
		}
	}
}
