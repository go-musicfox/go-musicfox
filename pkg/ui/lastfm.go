package ui

import (
	"fmt"

	"go-musicfox/pkg/storage"

	"github.com/skratchdot/open-golang/open"
)

type Lastfm struct {
	DefaultMenu
	model *NeteaseModel
	auth  *LastfmAuth
}

func NewLastfm(model *NeteaseModel) *Lastfm {
	return &Lastfm{
		model: model,
		auth:  NewLastfmAuth(model),
	}
}

func (m *Lastfm) GetMenuKey() string {
	return "last_fm"
}

func (m *Lastfm) MenuViews() []MenuItem {
	if m.model.lastfmUser == nil || m.model.lastfmUser.SessionKey == "" {
		return []MenuItem{
			{Title: "去授权"},
		}
	}
	return []MenuItem{
		{Title: "查看用户信息"},
		{Title: "清除授权"},
	}
}

func (m *Lastfm) SubMenu(_ *NeteaseModel, index int) IMenu {
	if m.model.lastfmUser == nil || m.model.lastfmUser.SessionKey == "" {
		return m.auth
	}
	switch index {
	case 0:
		_ = open.Start(m.model.lastfmUser.Url)
	case 1:
		m.model.lastfmUser = &storage.LastfmUser{}
		m.model.lastfmUser.Clear()
		return NewLastfmRes("清楚授权", nil, 2)
	}
	return nil
}

func (m *Lastfm) FormatMenuItem(item *MenuItem) {
	if m.model.lastfmUser == nil || m.model.lastfmUser.SessionKey == "" {
		item.Subtitle = "[未授权]"
	} else {
		item.Subtitle = fmt.Sprintf("[%s]", m.model.lastfmUser.Name)
	}
}
