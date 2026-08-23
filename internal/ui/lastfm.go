package ui

import (
	"fmt"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/notify"
)

type Lastfm struct {
	baseMenu
}

func NewLastfm(base baseMenu) *Lastfm {
	return &Lastfm{baseMenu: base}
}

func (m *Lastfm) GetMenuKey() string {
	return "last_fm"
}

func (m *Lastfm) MenuViews() []model.MenuItem {
	getControlTitle := func() string {
		if m.svc.Lastfm().Tracker.Status() {
			return "关闭功能"
		}
		return "启用功能"
	}

	return []model.MenuItem{
		{Title: "管理授权"},
		{Title: "前往主页"},
		{Title: getControlTitle()},
		{Title: "清空队列", Subtitle: fmt.Sprintf("[共 %d 条]", m.svc.Lastfm().Tracker.Count())},
	}
}

func (m *Lastfm) SubMenu(app *model.App, index int) model.Menu {
	switch index {
	case 0:
		return NewLastfmProfile(m.baseMenu)
	case 1:
		m.svc.Lastfm().OpenUserHomePage()
	case 2:
		turningOn := !m.svc.Lastfm().Tracker.Status()
		title := "关闭 Last.fm 功能"
		content := "确定关闭 Last.fm Scrobble 功能吗？"
		if turningOn {
			title = "启用 Last.fm 功能"
			content = "确定启用 Last.fm Scrobble 功能吗？"
		}
		showConfirmPopup(app, title, content, func() {
			m.svc.Lastfm().Tracker.Toggle()
			m.svc.Netease().MustMain().RefreshMenuList()
		})
		return nil
	case 3:
		count := m.svc.Lastfm().Tracker.Count()
		showConfirmPopup(app, "清空 Last.fm 队列", fmt.Sprintf("确定清空 Last.fm Scrobble 队列吗？（共 %d 条）", count), func() {
			m.svc.Lastfm().Tracker.Clear()
			notify.Notify(notify.NotifyContent{
				Title:   "清除 last.fm Scrobble 队列成功",
				Text:    "Last.fm Scrobble 队列已清除",
				GroupId: types.GroupID,
			})
			m.svc.Netease().MustMain().RefreshMenuList()
		})
		return nil
	}
	return nil
}

func (m *Lastfm) FormatMenuItem(item *model.MenuItem) {
	item.Subtitle = "[未授权]"
	if !m.svc.Lastfm().NeedAuth() {
		if username := m.svc.Lastfm().UserName(); username != "" {
			item.Subtitle = fmt.Sprintf("[%s]", username)
		}
	}
}
