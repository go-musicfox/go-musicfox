package ui

import (
	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/notify"
)

type LastfmRevokeMenu struct {
	baseMenu
	backLevel int
}

// 移除授权二次确认
func NewLastfmRevokeMenu(base baseMenu, backLevel int) *LastfmRevokeMenu {
	return &LastfmRevokeMenu{
		baseMenu:  base,
		backLevel: backLevel,
	}
}

func (m *LastfmRevokeMenu) GetMenuKey() string {
	return "last_fm_revoke"
}

func (m *LastfmRevokeMenu) MenuViews() []model.MenuItem {
	return []model.MenuItem{
		{Title: "确定"},
	}
}

func (m *LastfmRevokeMenu) SubMenu(app *model.App, _ int) model.Menu {
	m.netease.lastfmUser = &storage.LastfmUser{}
	m.netease.lastfmUser.Clear()

	notify.Notify(notify.NotifyContent{
		Title:   "清除授权成功",
		Text:    "Last.fm 授权已清除",
		GroupId: types.GroupID,
	})

	for range m.backLevel {
		app.MustMain().BackMenu()
	}
	return nil
}
