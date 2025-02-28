package ui

import (
	"fmt"

	"github.com/anhoder/foxful-cli/model"
	"github.com/skratchdot/open-golang/open"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/storage"
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

var getScrobbleSubtitle = func() string {
	if !configs.ConfigRegistry.Lastfm.Scrobble {
		return "[未启用]"
	}
	return "[已启用]"
}

func (m *Lastfm) MenuViews() []model.MenuItem {
	if configs.ConfigRegistry.Lastfm.Key != "" {
		if m.netease.lastfmUser == nil || m.netease.lastfmUser.SessionKey == "" {
			return []model.MenuItem{
				{Title: "去授权"},
			}
		}

		return []model.MenuItem{
			{Title: "查看用户信息"},
			{Title: "上报", Subtitle: getScrobbleSubtitle()},
			{Title: "清除授权"},
		}
	} else {
		return []model.MenuItem{
			{Title: "缺少 last.fm API key，请前往配置文件进行配置"},
		}
	}
}

func (m *Lastfm) SubMenu(app *model.App, index int) model.Menu {
	if configs.ConfigRegistry.Lastfm.Key != "" {
		if m.netease.lastfmUser == nil || m.netease.lastfmUser.SessionKey == "" {
			return m.auth
		}
		switch index {
		case 0:
			_ = open.Start(m.netease.lastfmUser.Url)
		case 1:
			configs.ConfigRegistry.Lastfm.Scrobble = !configs.ConfigRegistry.Lastfm.Scrobble
			app.MustMain().MenuList()[index].Subtitle = getScrobbleSubtitle()
		case 2:
			m.netease.lastfmUser = &storage.LastfmUser{}
			m.netease.lastfmUser.Clear()
			return NewLastfmRes(m.baseMenu, "清除授权", nil, 2)
		}
		return nil
	}
	app.MustMain().BackMenu()
	return nil
}

func (m *Lastfm) FormatMenuItem(item *model.MenuItem) {
	if m.netease.lastfmUser == nil || m.netease.lastfmUser.SessionKey == "" {
		item.Subtitle = "[未授权]"
	} else {
		item.Subtitle = fmt.Sprintf("[%s]", m.netease.lastfmUser.Name)
	}
}
