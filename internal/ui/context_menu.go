package ui

import (
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
)

// ContextMenuContext 描述右键菜单展开时的上下文：当前菜单、选中行索引
// （-1 表示无选中）、是否针对当前播放（playing 组）。
type ContextMenuContext struct {
	Menu          Menu
	SelectedIndex int
	Playing       bool
}

// ContextMenuContrib 是插件贡献的一个右键菜单操作项：Show 决定其在当前
// 上下文是否显示，Action 在用户点击时执行（返回页面与命令；导航自行处理）。
type ContextMenuContrib struct {
	Title  string // 菜单项文案（不含图标）
	Show   func(ctx ContextMenuContext) bool
	Action func(svc MenuServices, ctx ContextMenuContext) (model.Page, tea.Cmd)
}

// contextMenuContribs 保存插件注册的右键菜单贡献项，按注册序（编译期 init()）。
var contextMenuContribs []ContextMenuContrib

// RegisterContextMenuContrib 注册插件右键菜单项。全部注册项归入一个
// "插件"分组（Header: "插件"），追加在 generic 全局组之后，按注册序显示。
// 空 Title 或 nil Action 为编程错误，直接 panic 拒绝注册。
func RegisterContextMenuContrib(contrib ContextMenuContrib) {
	if contrib.Title == "" {
		panic("RegisterContextMenuContrib: empty title")
	}
	if contrib.Action == nil {
		panic("RegisterContextMenuContrib: nil Action")
	}
	contextMenuContribs = append(contextMenuContribs, contrib)
}

// ContextMenuContribs 返回已注册贡献项的快照副本（按注册序），供右键菜单
// 构造与测试读取；返回切片不与内部注册表共享底层数组。
func ContextMenuContribs() []ContextMenuContrib {
	contribs := make([]ContextMenuContrib, len(contextMenuContribs))
	copy(contribs, contextMenuContribs)
	return contribs
}

// contextMenuPluginHeader 是插件贡献项统一归入的分组标题。
const contextMenuPluginHeader = "插件"

// buildPluginContextMenuItems 将已注册的插件贡献项转换为右键菜单项：满足
// Show(ctx) 的项生成 ID 为 "plugin:<注册序号>" 的项（Label 带 itemIndent
// 缩进，与 buildGroupItems 的操作项一致），归入一个 "插件" Header 分组。
// 无满足项时返回 nil。前置分隔线由调用方按 leadSeparator 语义处理。
func buildPluginContextMenuItems(ctx ContextMenuContext) []model.ContextMenuItem {
	var items []model.ContextMenuItem
	for i, contrib := range ContextMenuContribs() {
		if contrib.Show != nil && !contrib.Show(ctx) {
			continue
		}
		if items == nil {
			items = append(items, model.ContextMenuItem{Header: true, Label: contextMenuPluginHeader})
		}
		items = append(items, model.ContextMenuItem{
			ID:    "plugin:" + strconv.Itoa(i),
			Label: itemIndent + contrib.Title,
		})
	}
	return items
}

// handlePluginContextAction 分发 "plugin:<注册序号>" ID 到对应贡献项的
// Action，返回其 (model.Page, tea.Cmd)。序号非数字、越界或 Action 缺失时
// 返回 nil, nil。
func handlePluginContextAction(svc MenuServices, ctx ContextMenuContext, id string) (model.Page, tea.Cmd) {
	rest, ok := strings.CutPrefix(id, "plugin:")
	if !ok {
		return nil, nil
	}
	i, err := strconv.Atoi(rest)
	if err != nil {
		return nil, nil
	}
	contribs := ContextMenuContribs()
	if i < 0 || i >= len(contribs) {
		return nil, nil
	}
	contrib := contribs[i]
	if contrib.Action == nil {
		return nil, nil
	}
	return contrib.Action(svc, ctx)
}
