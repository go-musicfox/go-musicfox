package ui

import (
	"github.com/skratchdot/open-golang/open"
	"go-musicfox/pkg/storage"
	"go-musicfox/utils"
)

type LastfmAuth struct {
	DefaultMenu
	model *NeteaseModel
	token string
	url   string
	err   error
}

func NewLastfmAuth(model *NeteaseModel) *LastfmAuth {
	return &LastfmAuth{model: model}
}

func (m *LastfmAuth) GetMenuKey() string {
	return "last_fm_auth"
}

func (m *LastfmAuth) MenuViews() []MenuItem {
	return []MenuItem{
		{Title: "已点击，继续授权"},
	}
}

func (m *LastfmAuth) BeforeBackMenuHook() Hook {
	return func(model *NeteaseModel) bool {
		m.token, m.url, m.err = "", "", nil
		return true
	}
}

func (m *LastfmAuth) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {
		m.token, m.url, m.err = model.lastfm.GetAuthUrlWithToken()
		if m.url != "" {
			_ = open.Start(m.url)
		}
		utils.Logger().Println("[lastfm] auth url: " + m.url)
		return true
	}
}

func (m *LastfmAuth) SubMenu(model *NeteaseModel, _ int) IMenu {
	var err error

	loading := NewLoading(model)
	loading.start()

	if m.model.lastfmUser == nil {
		m.model.lastfmUser = &storage.LastfmUser{}
	}
	m.model.lastfmUser.SessionKey, err = model.lastfm.GetSession(m.token)
	if err != nil {
		loading.complete()
		return NewLastfmRes("授权", err, 1)
	}
	user, err := model.lastfm.GetUserInfo(map[string]interface{}{})
	loading.complete()
	if err != nil {
		return NewLastfmRes("授权", err, 1)
	}
	m.model.lastfmUser.Id = user.Id
	m.model.lastfmUser.Name = user.Name
	m.model.lastfmUser.RealName = user.RealName
	m.model.lastfmUser.Url = user.Url
	m.model.lastfmUser.Store()
	return NewLastfmRes("授权", nil, 3)
}

func (m *LastfmAuth) FormatMenuItem(item *MenuItem) {
	if m.err != nil {
		item.Subtitle = "[错误: " + m.err.Error() + "]"
		return
	}
	if m.url != "" {
		item.Subtitle = "打开链接进行授权"
		return
	}
	item.Subtitle = ""
}
