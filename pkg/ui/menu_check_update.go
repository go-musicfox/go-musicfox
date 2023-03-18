package ui

import (
	"github.com/go-musicfox/go-musicfox/pkg/constants"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/skratchdot/open-golang/open"
)

type CheckUpdateMenu struct {
	DefaultMenu
	hasUpdate bool
}

func NewCheckUpdateMenu() *CheckUpdateMenu {
	return &CheckUpdateMenu{}
}

func (m *CheckUpdateMenu) GetMenuKey() string {
	return "check_update"
}

func (m *CheckUpdateMenu) MenuViews() []MenuItem {
	if m.hasUpdate {
		return []MenuItem{
			{Title: "检查到新版本，回车查看~", Subtitle: "ENTER"},
		}
	}

	return []MenuItem{
		{Title: "已是最新版本"},
	}
}

func (m *CheckUpdateMenu) SubMenu(_ *NeteaseModel, _ int) Menu {
	if m.hasUpdate {
		_ = open.Start(constants.AppGithubUrl)
	}
	return nil
}

func (m *CheckUpdateMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {
		m.hasUpdate = utils.CheckUpdate()
		return true
	}
}
