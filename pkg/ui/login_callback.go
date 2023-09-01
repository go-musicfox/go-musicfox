package ui

import "github.com/anhoder/foxful-cli/model"

type LoginCallback func() model.Page

func EnterMenuCallback(m *model.Main) LoginCallback {
	return func() model.Page {
		return m.EnterMenu(nil, nil)
	}
}
