package migration

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// reportModel 是用于展示迁移报告的 bubbletea 模型。
type reportModel struct {
	failedItems  []migrationResult
	successItems []migrationResult
	skippedItems []migrationResult
	displayItems []migrationResult
	hasErrors    bool
}

// newReportModel 创建并初始化报告模型，它会将传入的结果进行分组。
func newReportModel(results []migrationResult) reportModel {
	m := reportModel{}

	for _, res := range results {
		if res.Item.displayOnly {
			m.displayItems = append(m.displayItems, res)
			continue
		}

		switch res.Status {
		case StatusFailure:
			m.failedItems = append(m.failedItems, res)
		case StatusSuccess:
			m.successItems = append(m.successItems, res)
		case StatusSkipped:
			m.skippedItems = append(m.skippedItems, res)
		}
	}

	m.hasErrors = len(m.failedItems) > 0
	return m
}

func (m reportModel) Init() tea.Cmd {
	return nil
}

func (m reportModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	}
	return m, nil
}
func (m reportModel) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true)
	sectionTitleStyle := lipgloss.NewStyle().Bold(true).Underline(true)
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))  // Green
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))     // Red
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12"))     // Light Cyan/Blue
	skippedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // Grey

	var blocks []string

	if m.hasErrors {
		blocks = append(blocks, titleStyle.Render("go-musicfox 迁移失败"))
	} else {
		blocks = append(blocks, titleStyle.Render("go-musicfox 迁移报告"))
	}

	if len(m.failedItems) > 0 {
		blocks = append(blocks, "")
		blocks = append(blocks, sectionTitleStyle.Render("迁移失败 (需要您手动干预)"))
		for _, res := range m.failedItems {
			errorText := fmt.Sprintf("  ✖ %s: 失败\n    原因: %v\n    路径: %s -> %s", res.Description, res.Error, res.Item.oldPathFn(), res.Item.newPathFn())
			blocks = append(blocks, errorStyle.Render(errorText))
		}
	}

	if len(m.successItems) > 0 {
		blocks = append(blocks, "") 
		blocks = append(blocks, sectionTitleStyle.Render("迁移成功"))
		for _, res := range m.successItems {
			successText := fmt.Sprintf("  ✔ %s\n    变更: %s -> %s", res.Description, res.Item.oldPathFn(), res.Item.newPathFn())
			blocks = append(blocks, successStyle.Render(successText))
		}
	}

	if len(m.skippedItems) > 0 {
		blocks = append(blocks, "")
		blocks = append(blocks, sectionTitleStyle.Render("无需变更 (已是最新)"))
		for _, res := range m.skippedItems {
			skippedText := fmt.Sprintf("  - %s\n    请留意: %s -> %s", res.Description, res.Item.oldPathFn(), res.Item.newPathFn())
			blocks = append(blocks, skippedStyle.Render(skippedText))
		}
	}

	if len(m.displayItems) > 0 {
		blocks = append(blocks, "")
		blocks = append(blocks, sectionTitleStyle.Render("手动操作提示"))
		for _, res := range m.displayItems {
			infoText := fmt.Sprintf("  • %s\n    参考路径: %s -> %s", res.Description, res.Item.oldPathFn(), res.Item.newPathFn())
			blocks = append(blocks, infoStyle.Render(infoText))
		}
	}

	blocks = append(blocks, "")
	if m.hasErrors {
		blocks = append(blocks, "请处理错误后重试。按任意键退出。")
	} else {
		blocks = append(blocks, "迁移完成！请重新启动程序以加载新配置。\n按任意键退出。")
	}

	finalContent := lipgloss.JoinVertical(lipgloss.Left, blocks...)

	return lipgloss.NewStyle().Padding(1, 2).Render(finalContent)
}
