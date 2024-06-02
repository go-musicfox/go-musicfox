package ui

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/skratchdot/open-golang/open"

	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/version"
)

type CheckUpdateMenu struct {
	baseMenu
	hasUpdate bool
}

func NewCheckUpdateMenu(base baseMenu) *CheckUpdateMenu {
	return &CheckUpdateMenu{
		baseMenu: base,
	}
}

func (m *CheckUpdateMenu) GetMenuKey() string {
	return "check_update"
}

func (m *CheckUpdateMenu) MenuViews() []model.MenuItem {
	if m.hasUpdate {
		return []model.MenuItem{
			{Title: "检查到新版本，回车查看~", Subtitle: "ENTER"},
		}
	}

	return []model.MenuItem{
		{Title: "已是最新版本"},
	}
}

func (m *CheckUpdateMenu) SubMenu(_ *model.App, _ int) model.Menu {
	if m.hasUpdate {
		_ = open.Start(types.AppGithubUrl)
	}
	return nil
}

func (m *CheckUpdateMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		m.hasUpdate, _ = version.CheckUpdate()
		return true, nil
	}
}
