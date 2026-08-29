package lastfm

import (
	"github.com/anhoder/foxful-cli/model"

	lastfm "github.com/go-musicfox/go-musicfox/internal/lastfm"
	"github.com/go-musicfox/go-musicfox/internal/types"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
	"github.com/go-musicfox/go-musicfox/utils/notify"
)

// LastfmProfile 是 Last.fm 子菜单：设置 API account 与授权管理。页面跳转经
// ui.BuildPageOrToast 走页面注册表（lastfm_custom_api / lastfm_auth），
// opts 携带 ui.MenuServices（经 m.Services() 取自身访问器）。
type LastfmProfile struct {
	ui.BaseMenu
}

func NewLastfmProfile(base ui.BaseMenu) *LastfmProfile {
	return &LastfmProfile{
		BaseMenu: base,
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
		if m.Lastfm().NeedAuth() {
			return "去授权"
		}
		return "取消授权"
	}
	return []model.MenuItem{{Title: "设置 API account", Subtitle: "[已设置]"}, {Title: getAuthTitle()}}
}

func (m *LastfmProfile) SubMenu(app *model.App, index int) model.Menu {
	switch index {
	case 0:
		page := ui.BuildPageOrToast("lastfm_custom_api", LastfmCustomAPIPageOpts{Svc: m.Services()})
		if page == nil {
			return nil
		}
		page.(*LastfmCustomAPIPage).AfterAction = func() {
			app.MustMain().RefreshMenuList()
		}
		return ui.NewMenuToPage(m.BaseMenu, page, m.CoverRenderer().ClearDisplayed)
	case 1:
		if m.Lastfm().NeedAuth() {
			page := ui.BuildPageOrToast("lastfm_auth", LastfmAuthPageOpts{Svc: m.Services()})
			if page == nil {
				return nil
			}
			page.(*LastfmAuthPage).AfterAction = func() {
				app.MustMain().RefreshMenuList()
			}
			return ui.NewMenuToPage(m.BaseMenu, page, m.CoverRenderer().ClearDisplayed)
		}
		ui.ShowConfirmPopup(app, "清除 Last.fm 授权", "确定清除 Last.fm 授权信息吗？", func() {
			m.Lastfm().ClearUserInfo()
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
