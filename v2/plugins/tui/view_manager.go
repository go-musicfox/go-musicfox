package tui

import (
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/go-musicfox/go-musicfox/v2/plugins/tui/config"
)

// ViewManager 视图管理器
type ViewManager struct {
	plugin *TUIPlugin
	config *config.TUIConfig
	
	// 布局相关
	width  int
	height int
	isDualColumn bool
	
	// 显示状态
	showPlayer bool
	showStatusBar bool
	showTitle bool
}

// NewViewManager 创建视图管理器
func NewViewManager(plugin *TUIPlugin, cfg *config.TUIConfig) *ViewManager {
	return &ViewManager{
		plugin: plugin,
		config: cfg,
		showPlayer: true,
		showStatusBar: true,
		showTitle: true,
		isDualColumn: false,
	}
}

// SetSize 设置视图大小
func (vm *ViewManager) SetSize(width, height int) {
	vm.width = width
	vm.height = height
	
	// 根据宽度决定是否使用双列布局
	minWidth := 120 // 默认最小宽度
	vm.isDualColumn = width >= minWidth
}

// Render 渲染完整界面
func (vm *ViewManager) Render() string {
	var builder strings.Builder
	
	// 计算各部分高度
	heights := vm.calculateHeights()
	
	// 渲染标题栏
	if vm.showTitle {
		builder.WriteString(vm.renderTitle())
		builder.WriteString("\n")
	}
	
	// 渲染主内容区域
	if vm.isDualColumn {
		builder.WriteString(vm.renderDualColumn(heights.main))
	} else {
		builder.WriteString(vm.renderSingleColumn(heights.main))
	}
	
	// 渲染播放器
	if vm.showPlayer && vm.plugin.player != nil {
		builder.WriteString("\n")
		builder.WriteString(vm.renderPlayerSection(heights.player))
	}
	
	// 渲染状态栏
	if vm.showStatusBar {
		builder.WriteString("\n")
		builder.WriteString(vm.renderStatusBar())
	}
	
	return builder.String()
}

// Heights 各部分高度
type Heights struct {
	title  int
	main   int
	player int
	status int
}

// calculateHeights 计算各部分高度
func (vm *ViewManager) calculateHeights() Heights {
	heights := Heights{}
	
	// 标题栏高度
	if vm.showTitle {
		heights.title = 1
	}
	
	// 状态栏高度
	if vm.showStatusBar {
		heights.status = 1
	}
	
	// 播放器高度
	if vm.showPlayer {
		heights.player = 6 // 默认播放器高度
		if heights.player < 3 {
			heights.player = 3 // 最小高度
		}
	}
	
	// 主内容区域高度
	heights.main = vm.height - heights.title - heights.player - heights.status
	if heights.main < 5 {
		heights.main = 5 // 最小高度
	}
	
	return heights
}

// renderTitle 渲染标题栏
func (vm *ViewManager) renderTitle() string {
	title := "go-musicfox v2"
	if vm.plugin.user != nil {
		title += " - " + vm.plugin.user.Nickname
	}
	
	// 居中显示标题
	padding := (vm.width - len(title)) / 2
	if padding < 0 {
		padding = 0
	}
	
	return strings.Repeat(" ", padding) + title
}

// renderSingleColumn 渲染单列布局
func (vm *ViewManager) renderSingleColumn(height int) string {
	if vm.plugin.main == nil {
		return "正在初始化..."
	}
	
	// TODO: 添加适当的类型断言来调用View方法
	return "主界面内容"
}

// renderDualColumn 渲染双列布局
func (vm *ViewManager) renderDualColumn(height int) string {
	// 双列布局：左侧菜单，右侧内容
	leftWidth := vm.width / 3
	rightWidth := vm.width - leftWidth - 1 // 减1是为了分隔符
	
	if vm.plugin.main == nil {
		return "正在初始化..."
	}
	
	// TODO: 添加适当的类型断言来调用View方法
	mainContent := "主界面内容"
	lines := strings.Split(mainContent, "\n")
	
	var builder strings.Builder
	for i := 0; i < height && i < len(lines); i++ {
		line := lines[i]
		
		// 左侧内容（截断或填充）
		if len(line) > leftWidth {
			builder.WriteString(line[:leftWidth])
		} else {
			builder.WriteString(line + strings.Repeat(" ", leftWidth-len(line)))
		}
		
		// 分隔符
		builder.WriteString("|")
		
		// 右侧内容（暂时为空，可以显示详细信息）
		builder.WriteString(strings.Repeat(" ", rightWidth))
		
		if i < height-1 {
			builder.WriteString("\n")
		}
	}
	
	return builder.String()
}

// renderPlayerSection 渲染播放器区域
func (vm *ViewManager) renderPlayerSection(height int) string {
	if vm.plugin.player == nil {
		return strings.Repeat("-", vm.width)
	}
	
	// 添加分隔线
	var builder strings.Builder
	builder.WriteString(strings.Repeat("-", vm.width))
	builder.WriteString("\n")
	
	// TODO: 添加适当的类型断言来调用Render方法
	playerContent := "播放器内容"
	builder.WriteString(playerContent)
	
	return builder.String()
}

// renderStatusBar 渲染状态栏
func (vm *ViewManager) renderStatusBar() string {
	var parts []string
	
	// 添加快捷键提示
	parts = append(parts, "q:退出")
	parts = append(parts, "?:帮助")
	parts = append(parts, "/:搜索")
	
	// TODO: 添加适当的类型断言来访问播放器属性
	if vm.plugin.player != nil {
		parts = append(parts, "空格:播放/暂停")
		parts = append(parts, "[]:上/下一首")
		parts = append(parts, "+/-:音量")
	}
	
	// 右侧显示时间或其他信息
	rightInfo := ""
	if vm.plugin.player != nil {
		// TODO: 添加适当的类型断言来调用播放器方法
		rightInfo = "00:00/00:00"
	}
	
	// 组合状态栏
	leftPart := strings.Join(parts, " | ")
	availableWidth := vm.width - len(leftPart) - len(rightInfo)
	if availableWidth < 0 {
		availableWidth = 0
	}
	
	return leftPart + 
		strings.Repeat(" ", availableWidth) + 
		rightInfo
}

// TogglePlayer 切换播放器显示
func (vm *ViewManager) TogglePlayer() {
	vm.showPlayer = !vm.showPlayer
}

// ToggleStatusBar 切换状态栏显示
func (vm *ViewManager) ToggleStatusBar() {
	vm.showStatusBar = !vm.showStatusBar
}

// ToggleTitle 切换标题栏显示
func (vm *ViewManager) ToggleTitle() {
	vm.showTitle = !vm.showTitle
}

// ToggleDualColumn 切换双列布局
func (vm *ViewManager) ToggleDualColumn() {
	vm.isDualColumn = !vm.isDualColumn
}

// HandleWindowSizeMsg 处理窗口大小变化消息
func (vm *ViewManager) HandleWindowSizeMsg(msg tea.WindowSizeMsg) {
	vm.SetSize(msg.Width, msg.Height)
	
	// 通知主界面窗口大小变化
	if vm.plugin.main != nil {
		// TODO: 添加适当的类型断言来调用Update方法
	}
}

// GetMainContentHeight 获取主内容区域高度
func (vm *ViewManager) GetMainContentHeight() int {
	heights := vm.calculateHeights()
	return heights.main
}

// GetMainContentWidth 获取主内容区域宽度
func (vm *ViewManager) GetMainContentWidth() int {
	if vm.isDualColumn {
		return vm.width / 3 // 左侧菜单宽度
	}
	return vm.width
}