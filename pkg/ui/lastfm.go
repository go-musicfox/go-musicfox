package ui

import (
	"fmt"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/pkg/storage"

	"github.com/skratchdot/open-golang/open"
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
	if m.netease.lastfmUser == nil || m.netease.lastfmUser.SessionKey == "" {
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
	if m.netease.lastfmUser == nil || m.netease.lastfmUser.SessionKey == "" {
		return m.auth
	}
	switch index {
	case 0:
		_ = open.Start(m.netease.lastfmUser.Url)
	case 1:
		m.netease.lastfmUser = &storage.LastfmUser{}
		m.netease.lastfmUser.Clear()
		return NewLastfmRes(m.baseMenu, "清除授权", nil, 2)
	}
	return nil
}

func (m *Lastfm) FormatMenuItem(item *model.MenuItem) {
	if m.netease.lastfmUser == nil || m.netease.lastfmUser.SessionKey == "" {
		item.Subtitle = "[未授权]"
	} else {
		item.Subtitle = fmt.Sprintf("[%s]", m.netease.lastfmUser.Name)
	}
}
