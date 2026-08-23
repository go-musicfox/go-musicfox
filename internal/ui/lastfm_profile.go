package ui

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/internal/lastfm"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/notify"
)

type LastfmProfile struct {
	baseMenu
}

func NewLastfmProfile(base baseMenu) *LastfmProfile {
	return &LastfmProfile{
		baseMenu: base,
	}
}

func (m *LastfmProfile) GetMenuKey() string {
	return "lastfm_profile"
}

func (m *LastfmProfile) MenuViews() (menu []model.MenuItem) {
	if !lastfm.IsAvailable() {
		return []model.MenuItem{{Title: "设置 API account", Subtitle: "[待设置]"}}
	}

	getAuthTitle := func() string {
		if m.svc.Lastfm().NeedAuth() {
			return "去授权"
		}
		return "取消授权"
	}
	return []model.MenuItem{{Title: "设置 API account", Subtitle: "[已设置]"}, {Title: getAuthTitle()}}
}

func (m *LastfmProfile) SubMenu(app *model.App, index int) model.Menu {
	switch index {
	case 0:
		page := NewLastfmCustomApiPage(m.svc.Netease())
		page.AfterAction = func() {
			app.MustMain().RefreshMenuList()
		}
		return NewMenuToPage(m.baseMenu, page, m.svc.CoverRenderer().ClearDisplayed)
	case 1:
		if m.svc.Lastfm().NeedAuth() {
			page := NewLastfmAuthPage(m.svc.Netease())
			page.AfterAction = func() {
				app.MustMain().RefreshMenuList()
			}
			return NewMenuToPage(m.baseMenu, page, m.svc.CoverRenderer().ClearDisplayed)
		}
		showConfirmPopup(app, "清除 Last.fm 授权", "确定清除 Last.fm 授权信息吗？", func() {
			m.svc.Lastfm().ClearUserInfo()
			notify.Notify(notify.NotifyContent{
				Title:   "清除授权成功",
				Text:    "Last.fm 授权已清除",
				GroupId: types.GroupID,
			})
			// Original ConfirmMenu pushed a level and used backLevel:2 to land
			// on the Lastfm menu; the popup pushes nothing, so a single BackMenu
			// (out of LastfmProfile) reaches the same landing menu.
			app.MustMain().BackMenu()
			app.MustMain().RefreshMenuList()
		})
		return nil
	default:
		return nil
	}
}
