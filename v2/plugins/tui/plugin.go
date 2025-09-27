package tui

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	plugin "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
	"github.com/go-musicfox/go-musicfox/v2/plugins/tui/config"
)

// User 用户信息
type User struct {
	ID       string
	Nickname string
	Avatar   string
}

// TUIPlugin TUI插件实现，基于bubbletea
type TUIPlugin struct {
	*plugin.BasePlugin
	
	// bubbletea相关
	program *tea.Program
	model   *TUIModel
	
	// foxful-cli相关
	main   interface{} // 主界面模型
	app    interface{} // 应用实例
	player interface{} // 播放器实例
	
	// 组件
	inputHandler *InputHandler
	viewManager  *ViewManager
	
	// 用户信息
	user *User
	
	// 日志记录器
	logger *slog.Logger
	
	// 状态
	isRunning  bool
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// NewTUIPlugin 创建TUI插件实例
func NewTUIPlugin() plugin.Plugin {
	info := &plugin.PluginInfo{
		ID:          "tui",
		Name:        "TUI Plugin",
		Version:     "1.0.0",
		Description: "Terminal User Interface Plugin for go-musicfox",
		Author:      "go-musicfox",
		Type:        "ui",
	}
	
	return &TUIPlugin{
		BasePlugin: plugin.NewBasePlugin(info),
		logger: slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
	}
}

// Initialize 初始化插件
func (p *TUIPlugin) Initialize(ctx plugin.PluginContext) error {
	p.logger.Info("Initializing TUI plugin")
	
	// 调用基类初始化
	if err := p.BasePlugin.Initialize(ctx); err != nil {
		return err
	}
	
	// 创建默认配置
	cfg := config.NewTUIConfig()
	
	// 创建组件
	p.inputHandler = NewInputHandler(p, cfg)
	p.viewManager = NewViewManager(p, cfg)
	p.player = NewPlayer(p)
	
	// 创建TUI模型
	p.model = NewTUIModel(p)
	
	p.logger.Info("TUI plugin initialized successfully")
	return nil
}

// Start 启动插件
func (p *TUIPlugin) Start() error {
	if p.isRunning {
		return nil
	}
	
	p.logger.Info("Starting TUI plugin")
	
	// 调用基类启动
	if err := p.BasePlugin.Start(); err != nil {
		return err
	}
	
	// 创建取消上下文
	var cancel context.CancelFunc
	p.ctx, cancel = context.WithCancel(context.Background())
	p.cancelFunc = cancel
	
	// 创建bubbletea程序
	p.program = tea.NewProgram(p.model, tea.WithAltScreen())
	
	// 在goroutine中启动TUI
	go func() {
		defer func() {
			if r := recover(); r != nil {
				p.logger.Error("TUI application panic", "error", r)
			}
			p.isRunning = false
		}()
		
		p.logger.Info("Starting bubbletea TUI program")
		if _, err := p.program.Run(); err != nil {
			p.logger.Error("Failed to run TUI program", "error", err)
		}
		p.logger.Info("TUI program finished")
	}()
	
	p.isRunning = true
	p.logger.Info("TUI plugin started successfully")
	return nil
}

// Stop 停止插件
func (p *TUIPlugin) Stop() error {
	if !p.isRunning {
		return nil
	}
	
	p.logger.Info("Stopping TUI plugin")
	
	// 调用基类停止
	if err := p.BasePlugin.Stop(); err != nil {
		return err
	}
	
	if p.cancelFunc != nil {
		p.cancelFunc()
	}
	
	// 退出bubbletea程序
	if p.program != nil {
		p.program.Quit()
	}
	
	p.isRunning = false
	p.logger.Info("TUI plugin stopped successfully")
	return nil
}

// IsRunning 检查插件是否正在运行
func (p *TUIPlugin) IsRunning() bool {
	return p.isRunning
}

// GetLogger 获取日志记录器（内部使用）
func (p *TUIPlugin) GetLogger() *slog.Logger {
	return p.logger
}

// SetUser 设置用户信息
func (p *TUIPlugin) SetUser(user *User) {
	p.user = user
	p.logger.Info("User information updated", "nickname", user.Nickname)
}

// GetUser 获取用户信息
func (p *TUIPlugin) GetUser() *User {
	return p.user
}

// ClearUser 清除用户信息
func (p *TUIPlugin) ClearUser() {
	p.user = nil
	p.logger.Info("User information cleared")
}

// TUIModel bubbletea模型
type TUIModel struct {
	plugin *TUIPlugin
	width  int
	height int
	currentView string
	statusMsg string
}

// NewTUIModel 创建TUI模型
func NewTUIModel(plugin *TUIPlugin) *TUIModel {
	return &TUIModel{
		plugin: plugin,
		currentView: "main",
		statusMsg: "Welcome to MusicFox MVP!",
	}
}

// Init 初始化模型
func (m *TUIModel) Init() tea.Cmd {
	return nil
}

// Update 更新模型
func (m *TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "h":
			m.statusMsg = "Help: q=quit, h=help, p=play, s=stop"
		case "p":
			m.statusMsg = "Playing music..."
		case "s":
			m.statusMsg = "Music stopped."
		}
	}
	return m, nil
}

// View 渲染视图
func (m *TUIModel) View() string {
	header := "🎵 MusicFox MVP - Terminal Music Player\n\n"
	
	main := fmt.Sprintf("Current View: %s\n", m.currentView)
	main += fmt.Sprintf("Status: %s\n\n", m.statusMsg)
	
	controls := "Controls:\n"
	controls += "  h - Show help\n"
	controls += "  p - Play music\n"
	controls += "  s - Stop music\n"
	controls += "  q - Quit\n\n"
	
	footer := "Press 'q' or Ctrl+C to quit"
	
	return header + main + controls + footer
}