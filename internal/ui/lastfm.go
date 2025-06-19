package ui

import (
	"fmt"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/internal/lastfm"
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

func (m *Lastfm) MenuViews() []model.MenuItem {
	if !lastfm.IsAvailable() {
		return []model.MenuItem{
			{Title: "当前不可用，请设置 API key 及 secret"},
		}
	}

	if m.netease.lastfm.NeedAuth() {
		return []model.MenuItem{
			{Title: "去授权"},
		}
	}
	return []model.MenuItem{
		{Title: "查看用户信息"},
		{Title: "清除授权"},
	}
}

func (m *Lastfm) SubMenu(_ *model.App, index int) model.Menu {
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
		m.netease.lastfm.ClearUserInfo()
		return NewLastfmRes(m.baseMenu, "清除授权", nil, 2)
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
