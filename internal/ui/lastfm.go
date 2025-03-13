package ui

import (
	"fmt"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/internal/lastfm"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/notify"
)

type Lastfm struct {
	baseMenu
	auth *LastfmAuth
}

func NewLastfm(base baseMenu) *Lastfm {
	return &Lastfm{
		baseMenu: base,
		auth:     NewLastfmAuth(base),
	}
}

func (m *Lastfm) GetMenuKey() string {
	return "last_fm"
}

var getScrobbleSubtitle = func(m *lastfm.Tracker) string {
	if m.Status() {
		return "[已启用]"
	}
	return "[未启用]"
}

var getScrobbleCountSubtitle = func(m *lastfm.Tracker) string {
	return fmt.Sprintf("[共 %d 条]", m.Count())
}

func (m *Lastfm) MenuViews() []model.MenuItem {
	if !lastfm.IsAvailable() {
		return []model.MenuItem{
			{Title: "当前不可用，请设置 API key 及 secret"},
			{Title: "本地记录", Subtitle: getScrobbleSubtitle(m.netease.lastfm.Tracker)},
			{Title: "清空队列", Subtitle: getScrobbleCountSubtitle(m.netease.lastfm.Tracker)},
		}
	}
	if m.netease.lastfm.NeedAuth() {
		return []model.MenuItem{
			{Title: "去授权"},
			{Title: "本地记录", Subtitle: getScrobbleSubtitle(m.netease.lastfm.Tracker)},
			{Title: "清空队列", Subtitle: getScrobbleCountSubtitle(m.netease.lastfm.Tracker)},
		}
	}
	return []model.MenuItem{
		{Title: "查看用户信息"},
		{Title: "切换功能状态", Subtitle: getScrobbleSubtitle(m.netease.lastfm.Tracker)},
		{Title: "清空队列", Subtitle: getScrobbleCountSubtitle(m.netease.lastfm.Tracker)},
		{Title: "清除授权"},
	}
}

func (m *Lastfm) SubMenu(app *model.App, index int) model.Menu {
	switch index {
	case 0:
		if !lastfm.IsAvailable() {
			return nil
		}
		if m.netease.lastfm.NeedAuth() {
			return m.auth
		}
		m.netease.lastfm.OpenUserHomePage()
	case 1:
		action := func() {
			m.netease.lastfm.Tracker.Toggle()
		}
		menu := NewConfirmMenu(m.baseMenu, []ConfirmItem{
			{title: model.MenuItem{Title: "确定"}, action: action, backLevel: 1},
		})
		app.MustMain().EnterMenu(menu, &m.MenuViews()[index])
	case 2:
		action := func() {
			m.netease.lastfm.Tracker.Clear()

			notify.Notify(notify.NotifyContent{
				Title:   "清除 last.fm Scrobble 队列成功",
				Text:    "Last.fm Scrobble 队列已清除",
				GroupId: types.GroupID,
			})
		}

		return NewConfirmMenu(m.baseMenu, []ConfirmItem{
			{title: model.MenuItem{Title: "确定"}, action: action, backLevel: 1},
		})
	case 3:
		action := func() {
			m.netease.lastfm.ClearUserInfo()

			notify.Notify(notify.NotifyContent{
				Title:   "清除授权成功",
				Text:    "Last.fm 授权已清除",
				GroupId: types.GroupID,
			})
		}

		return NewConfirmMenu(m.baseMenu, []ConfirmItem{
			{title: model.MenuItem{Title: "确定"}, action: action, backLevel: 2},
		})
	}
	return nil
}

func (m *Lastfm) FormatMenuItem(item *model.MenuItem) {
	item.Subtitle = "[未授权]"
	if !m.netease.lastfm.NeedAuth() {
		if username := m.netease.lastfm.UserName(); username != "" {
			item.Subtitle = fmt.Sprintf("[%s]", username)
		}
	}
}
