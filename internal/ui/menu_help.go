package ui

import (
	"slices"
	"sort"
	"strings"

	"github.com/anhoder/foxful-cli/model"
	"github.com/skratchdot/open-golang/open"

	"github.com/go-musicfox/go-musicfox/internal/keybindings"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

type HelpMenu struct {
	baseMenu
	menus []model.MenuItem
}

func NewHelpMenu(base baseMenu) *HelpMenu {
	menu := &HelpMenu{
		baseMenu: base,
		menus:    []model.MenuItem{},
	}

	// 准备最终的菜单项列表，先添加固定项
	menuItems := []model.MenuItem{
		{Title: "进来给个star⭐️呗~"},
	}

	boundOperations := keybindings.UserOperateToKeys()

	var builtinOps []keybindings.OperateType
	var customOps []keybindings.OperateType
	for op := range boundOperations {
		if op < 0 {
			builtinOps = append(builtinOps, op)
		} else {
			customOps = append(customOps, op)
		}
	}

	sort.Slice(builtinOps, func(i, j int) bool {
		return builtinOps[i] > builtinOps[j]
	})
	slices.Sort(customOps)

	// 生成菜单项
	processOp := func(op keybindings.OperateType) {
		_, ok := boundOperations[op]
		if !ok {
			return
		}

		keysRaw := boundOperations[op]

		validKeys := make([]string, 0, len(keysRaw))
		seenKeys := make(map[string]struct{}) // 用于快速检查重复

		for _, k := range keysRaw {
			if k == "" {
				continue
			}
			formattedKey := keybindings.FormatKeyForDisplay(k)
			if _, seen := seenKeys[formattedKey]; !seen {
				seenKeys[formattedKey] = struct{}{}
				validKeys = append(validKeys, formattedKey)
			}
		}

		// 如果没有有效的快捷键与此操作关联，则不显示此帮助项
		if len(validKeys) == 0 {
			return
		}

		keyStr := strings.Join(validKeys, "/")

		menuItems = append(menuItems, model.MenuItem{
			Title:    keyStr,
			Subtitle: op.Desc(),
		})
	}

	for _, op := range builtinOps {
		processOp(op)
	}
	for _, op := range customOps {
		processOp(op)
	}

	menu.menus = menuItems
	return menu
}

func (m *HelpMenu) GetMenuKey() string {
	return "help_menu"
}

func (m *HelpMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *HelpMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index == 0 {
		_ = open.Start(types.AppGithubUrl)
	}
	return nil
}
