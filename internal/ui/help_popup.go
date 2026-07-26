package ui

import (
	"slices"
	"sort"
	"strings"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/keybindings"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// buildHelpMarkdown 依据当前生效的键位绑定生成帮助文档（Markdown 表格）。
// 内建操作与自定义操作分两组展示。
func buildHelpMarkdown() string {
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
	sort.Slice(builtinOps, func(i, j int) bool { return builtinOps[i] > builtinOps[j] })
	slices.Sort(customOps)

	var sb strings.Builder
	sb.WriteString("# go-musicfox 快捷键\n\n")
	sb.WriteString("[进来给个 star ⭐️ 呗~](")
	sb.WriteString(types.AppGithubUrl)
	sb.WriteString(")\n\n")

	writeGroup := func(title string, ops []keybindings.OperateType) {
		rows := make([]string, 0, len(ops))
		for _, op := range ops {
			keysRaw, ok := boundOperations[op]
			if !ok {
				continue
			}
			validKeys := make([]string, 0, len(keysRaw))
			seenKeys := make(map[string]struct{})
			for _, k := range keysRaw {
				if k == "" {
					continue
				}
				formatted := keybindings.FormatKeyForDisplay(k)
				if _, seen := seenKeys[formatted]; !seen {
					seenKeys[formatted] = struct{}{}
					validKeys = append(validKeys, formatted)
				}
			}
			if len(validKeys) == 0 {
				continue
			}
			rows = append(rows, "| "+strings.Join(validKeys, " / ")+" | "+op.Desc()+" |")
		}
		if len(rows) == 0 {
			return
		}
		sb.WriteString("## ")
		sb.WriteString(title)
		sb.WriteString("\n\n| 按键 | 操作 |\n| --- | --- |\n")
		sb.WriteString(strings.Join(rows, "\n"))
		sb.WriteString("\n\n")
	}

	writeGroup("内建操作", builtinOps)
	writeGroup("自定义操作", customOps)

	return sb.String()
}

// showHelpPopup 以 Markdown 弹窗展示帮助（快捷键说明），替代旧的整页帮助菜单。
// 弹窗按终端尺寸限制为可滚动视口，支持鼠标滚动、Esc/q 关闭。
func showHelpPopup(app *model.App) {
	popupWidth := app.WindowWidth() * 70 / 100
	if popupWidth < 40 {
		popupWidth = 40
	}
	if popupWidth > 120 {
		popupWidth = 120
	}

	maxHeight := app.WindowHeight() * 80 / 100
	if maxHeight < 10 {
		maxHeight = 10
	}

	popup, err := model.NewMarkdownPopup(model.MarkdownPopupSpec{
		Title:              " 帮助 ",
		MarkdownContent:    buildHelpMarkdown(),
		MarkdownDarkStyle:  configs.GetCurrentMarkdownTheme(true),
		MarkdownLightStyle: configs.GetCurrentMarkdownTheme(false),
		MarkdownEmoji:      true,
		MaxWidth:           popupWidth,
		MaxHeight:          maxHeight,
		Anchor:             model.AnchorCenter,
		DisableResize:      false,
		CloseKeys:          []string{"esc", "q"},
		Actions: []model.PopupAction{
			{ID: "ok", Label: "OK"},
		},
	})
	if err != nil {
		return
	}
	app.ShowPopup(popup)
}
