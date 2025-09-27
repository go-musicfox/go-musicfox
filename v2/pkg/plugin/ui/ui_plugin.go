package ui

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	plugin "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
)

// BaseUIExtensionPlugin 基础UI扩展插件实现
type BaseUIExtensionPlugin struct {
	*plugin.BasePlugin

	// 核心组件
	renderer      UIRenderer
	inputHandler  InputHandler
	layoutManager LayoutManager
	themeManager  ThemeManager
	hotReloader   HotReloader

	// 状态管理
	currentLayout *Layout
	currentTheme  *Theme
	currentUIType UIType
	version       string

	// 组件管理
	components    map[string]*UIComponent
	componentsMux sync.RWMutex

	// 事件处理
	eventHandlers map[string]EventHandler
	eventMux      sync.RWMutex

	// 配置
	config        *UIPluginConfig
	logger        *slog.Logger
}

// UIPluginConfig UI插件配置
type UIPluginConfig struct {
	SupportedUITypes []UIType          `json:"supported_ui_types"`
	DefaultTheme     string            `json:"default_theme"`
	DefaultLayout    string            `json:"default_layout"`
	HotReloadEnabled bool              `json:"hot_reload_enabled"`
	MaxComponents    int               `json:"max_components"`
	Settings         map[string]string `json:"settings"`
}

// NewBaseUIExtensionPlugin 创建基础UI扩展插件
func NewBaseUIExtensionPlugin(info *plugin.PluginInfo, config *UIPluginConfig, logger *slog.Logger) *BaseUIExtensionPlugin {
	basePlugin := plugin.NewBasePlugin(info)

	return &BaseUIExtensionPlugin{
		BasePlugin:    basePlugin,
		config:        config,
		logger:        logger,
		version:       info.Version,
		components:    make(map[string]*UIComponent),
		eventHandlers: make(map[string]EventHandler),
		currentUIType: UITypeDesktop, // 默认桌面UI
	}
}

// Initialize 初始化插件
func (p *BaseUIExtensionPlugin) Initialize(ctx plugin.PluginContext) error {
	if err := p.BasePlugin.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize base plugin: %w", err)
	}

	// 初始化核心组件
	if err := p.initializeComponents(); err != nil {
		return fmt.Errorf("failed to initialize components: %w", err)
	}

	// 加载默认主题和布局
	if err := p.loadDefaults(); err != nil {
		return fmt.Errorf("failed to load defaults: %w", err)
	}

	p.logger.Info("UI extension plugin initialized", "name", p.GetInfo().Name)
	return nil
}

// initializeComponents 初始化核心组件
func (p *BaseUIExtensionPlugin) initializeComponents() error {
	// 初始化渲染器
	p.renderer = NewMultiUIRenderer(p.logger)

	// 初始化输入处理器
	p.inputHandler = NewMultiInputHandler(p.logger)

	// 初始化布局管理器
	p.layoutManager = NewDefaultLayoutManager(p.logger)

	// 初始化主题管理器
	p.themeManager = NewDefaultThemeManager(p.logger)

	// 初始化热重载器（如果启用）
	if p.config.HotReloadEnabled {
		p.hotReloader = NewDefaultHotReloader(p.logger)
	}

	return nil
}

// loadDefaults 加载默认配置
func (p *BaseUIExtensionPlugin) loadDefaults() error {
	// 加载默认主题
	if p.config.DefaultTheme != "" {
		theme := p.createDefaultTheme(p.config.DefaultTheme)
		if err := p.ApplyTheme(theme); err != nil {
			p.logger.Warn("Failed to apply default theme", "error", err)
		}
	}

	// 加载默认布局
	if p.config.DefaultLayout != "" {
		layout := p.createDefaultLayout(p.config.DefaultLayout)
		if err := p.SetLayout(layout); err != nil {
			p.logger.Warn("Failed to set default layout", "error", err)
		}
	}

	return nil
}

// Render 渲染UI
func (p *BaseUIExtensionPlugin) Render(ctx context.Context, state *AppState) error {
	if p.GetState() != plugin.PluginStateRunning {
		return fmt.Errorf("plugin is not running")
	}

	if p.renderer == nil {
		return fmt.Errorf("renderer not initialized")
	}

	// 渲染所有组件
	p.componentsMux.RLock()
	defer p.componentsMux.RUnlock()

	for _, component := range p.components {
		if component.Visible {
			if _, err := p.renderer.Render(ctx, component, state); err != nil {
				p.logger.Error("Failed to render component", "component", component.ID, "error", err)
			}
		}
	}

	return nil
}

// HandleInput 处理输入事件
func (p *BaseUIExtensionPlugin) HandleInput(ctx context.Context, input *InputEvent) error {
	if p.GetState() != plugin.PluginStateRunning {
		return fmt.Errorf("plugin is not running")
	}

	if p.inputHandler == nil {
		return fmt.Errorf("input handler not initialized")
	}

	return p.inputHandler.Handle(ctx, input)
}

// GetLayout 获取当前布局
func (p *BaseUIExtensionPlugin) GetLayout() *Layout {
	return p.currentLayout
}

// SetLayout 设置布局
func (p *BaseUIExtensionPlugin) SetLayout(layout *Layout) error {
	if layout == nil {
		return fmt.Errorf("layout cannot be nil")
	}

	if p.layoutManager != nil {
		if err := p.layoutManager.ValidateLayout(layout); err != nil {
			return fmt.Errorf("invalid layout: %w", err)
		}

		if err := p.layoutManager.ApplyLayout(context.Background(), layout); err != nil {
			return fmt.Errorf("failed to apply layout: %w", err)
		}
	}

	p.currentLayout = layout
	p.logger.Info("Layout applied", "layout", layout.Name)
	return nil
}

// GetSupportedLayouts 获取支持的布局
func (p *BaseUIExtensionPlugin) GetSupportedLayouts() []*Layout {
	// 返回预定义的布局
	return []*Layout{
		p.createDefaultLayout("grid"),
		p.createDefaultLayout("flex"),
		p.createDefaultLayout("absolute"),
	}
}

// GetSupportedThemes 获取支持的主题
func (p *BaseUIExtensionPlugin) GetSupportedThemes() []*Theme {
	// 返回预定义的主题
	return []*Theme{
		p.createDefaultTheme("light"),
		p.createDefaultTheme("dark"),
		p.createDefaultTheme("auto"),
	}
}

// ApplyTheme 应用主题
func (p *BaseUIExtensionPlugin) ApplyTheme(theme *Theme) error {
	if theme == nil {
		return fmt.Errorf("theme cannot be nil")
	}

	if p.themeManager != nil {
		if err := p.themeManager.ValidateTheme(theme); err != nil {
			return fmt.Errorf("invalid theme: %w", err)
		}

		if err := p.themeManager.ApplyTheme(context.Background(), theme); err != nil {
			return fmt.Errorf("failed to apply theme: %w", err)
		}
	}

	p.currentTheme = theme
	p.logger.Info("Theme applied", "theme", theme.Name)
	return nil
}

// GetCurrentTheme 获取当前主题
func (p *BaseUIExtensionPlugin) GetCurrentTheme() *Theme {
	return p.currentTheme
}

// CanHotReload 是否支持热重载
func (p *BaseUIExtensionPlugin) CanHotReload() bool {
	return p.config.HotReloadEnabled && p.hotReloader != nil
}

// HotReload 热重载
func (p *BaseUIExtensionPlugin) HotReload(newVersion []byte) error {
	if !p.CanHotReload() {
		return fmt.Errorf("hot reload not supported")
	}

	return p.hotReloader.Reload(context.Background(), newVersion)
}

// GetVersion 获取版本
func (p *BaseUIExtensionPlugin) GetVersion() string {
	return p.version
}

// GetSupportedUITypes 获取支持的UI类型
func (p *BaseUIExtensionPlugin) GetSupportedUITypes() []UIType {
	return p.config.SupportedUITypes
}

// SetUIType 设置UI类型
func (p *BaseUIExtensionPlugin) SetUIType(uiType UIType) error {
	// 检查是否支持该UI类型
	supported := false
	for _, supportedType := range p.config.SupportedUITypes {
		if supportedType == uiType {
			supported = true
			break
		}
	}

	if !supported {
		return fmt.Errorf("UI type %s not supported", uiType.String())
	}

	p.currentUIType = uiType
	p.logger.Info("UI type changed", "type", uiType.String())
	return nil
}

// GetCurrentUIType 获取当前UI类型
func (p *BaseUIExtensionPlugin) GetCurrentUIType() UIType {
	return p.currentUIType
}

// RegisterComponent 注册组件
func (p *BaseUIExtensionPlugin) RegisterComponent(ctx context.Context, component *UIComponent) error {
	if component == nil {
		return fmt.Errorf("component cannot be nil")
	}

	if component.ID == "" {
		return fmt.Errorf("component ID cannot be empty")
	}

	p.componentsMux.Lock()
	defer p.componentsMux.Unlock()

	// 检查组件数量限制
	if len(p.components) >= p.config.MaxComponents {
		return fmt.Errorf("maximum number of components reached: %d", p.config.MaxComponents)
	}

	// 检查组件是否已存在
	if _, exists := p.components[component.ID]; exists {
		return fmt.Errorf("component with ID %s already exists", component.ID)
	}

	// 设置时间戳
	component.CreatedAt = time.Now()
	component.UpdatedAt = time.Now()

	p.components[component.ID] = component
	p.logger.Info("Component registered", "id", component.ID, "name", component.Name)
	return nil
}

// UnregisterComponent 注销组件
func (p *BaseUIExtensionPlugin) UnregisterComponent(ctx context.Context, componentID string) error {
	if componentID == "" {
		return fmt.Errorf("component ID cannot be empty")
	}

	p.componentsMux.Lock()
	defer p.componentsMux.Unlock()

	if _, exists := p.components[componentID]; !exists {
		return fmt.Errorf("component with ID %s not found", componentID)
	}

	delete(p.components, componentID)
	p.logger.Info("Component unregistered", "id", componentID)
	return nil
}

// GetComponent 获取组件
func (p *BaseUIExtensionPlugin) GetComponent(ctx context.Context, componentID string) (*UIComponent, error) {
	if componentID == "" {
		return nil, fmt.Errorf("component ID cannot be empty")
	}

	p.componentsMux.RLock()
	defer p.componentsMux.RUnlock()

	component, exists := p.components[componentID]
	if !exists {
		return nil, fmt.Errorf("component with ID %s not found", componentID)
	}

	return component, nil
}

// ListComponents 列出所有组件
func (p *BaseUIExtensionPlugin) ListComponents(ctx context.Context) ([]*UIComponent, error) {
	p.componentsMux.RLock()
	defer p.componentsMux.RUnlock()

	components := make([]*UIComponent, 0, len(p.components))
	for _, component := range p.components {
		components = append(components, component)
	}

	return components, nil
}

// HandleEvent 处理事件 (实现plugin.Plugin接口)
func (p *BaseUIExtensionPlugin) HandleEvent(event interface{}) error {
	// 尝试将事件转换为UIEvent
	uiEvent, ok := event.(*UIEvent)
	if !ok {
		p.logger.Debug("Event is not a UIEvent, ignoring", "event", event)
		return nil
	}

	p.eventMux.RLock()
	handler, exists := p.eventHandlers[uiEvent.Type]
	p.eventMux.RUnlock()

	if !exists {
		p.logger.Debug("No handler found for event type", "type", uiEvent.Type)
		return nil
	}

	return handler.Handle(context.Background(), uiEvent)
}

// HandleUIEvent 处理UI事件 (UI特定的方法)
func (p *BaseUIExtensionPlugin) HandleUIEvent(ctx context.Context, event *UIEvent) error {
	return p.HandleEvent(event)
}

// RegisterEventHandler 注册事件处理器
func (p *BaseUIExtensionPlugin) RegisterEventHandler(eventType string, handler EventHandler) error {
	if eventType == "" {
		return fmt.Errorf("event type cannot be empty")
	}

	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	p.eventMux.Lock()
	defer p.eventMux.Unlock()

	p.eventHandlers[eventType] = handler
	p.logger.Info("Event handler registered", "type", eventType)
	return nil
}

// UnregisterEventHandler 注销事件处理器
func (p *BaseUIExtensionPlugin) UnregisterEventHandler(eventType string) error {
	if eventType == "" {
		return fmt.Errorf("event type cannot be empty")
	}

	p.eventMux.Lock()
	defer p.eventMux.Unlock()

	delete(p.eventHandlers, eventType)
	p.logger.Info("Event handler unregistered", "type", eventType)
	return nil
}

// createDefaultTheme 创建默认主题
func (p *BaseUIExtensionPlugin) createDefaultTheme(name string) *Theme {
	switch name {
	case "dark":
		return &Theme{
			ID:          "dark",
			Name:        "Dark Theme",
			Version:     "1.0.0",
			Description: "Default dark theme",
			Author:      "go-musicfox",
			Colors: &ColorScheme{
				Primary:       "#1976d2",
				Secondary:     "#424242",
				Accent:        "#ff4081",
				Background:    "#121212",
				Surface:       "#1e1e1e",
				Text:          "#ffffff",
				TextSecondary: "#b3b3b3",
				Border:        "#333333",
				Error:         "#f44336",
				Warning:       "#ff9800",
				Success:       "#4caf50",
				Info:          "#2196f3",
			},
			IsDark:    true,
			IsDefault: false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	case "light":
		return &Theme{
			ID:          "light",
			Name:        "Light Theme",
			Version:     "1.0.0",
			Description: "Default light theme",
			Author:      "go-musicfox",
			Colors: &ColorScheme{
				Primary:       "#1976d2",
				Secondary:     "#757575",
				Accent:        "#ff4081",
				Background:    "#ffffff",
				Surface:       "#f5f5f5",
				Text:          "#212121",
				TextSecondary: "#757575",
				Border:        "#e0e0e0",
				Error:         "#f44336",
				Warning:       "#ff9800",
				Success:       "#4caf50",
				Info:          "#2196f3",
			},
			IsDark:    false,
			IsDefault: true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	default:
		return p.createDefaultTheme("light")
	}
}

// createDefaultLayout 创建默认布局
func (p *BaseUIExtensionPlugin) createDefaultLayout(name string) *Layout {
	switch name {
	case "grid":
		return &Layout{
			ID:          "grid",
			Name:        "Grid Layout",
			Description: "Default grid layout",
			Type:        LayoutTypeGrid,
			Grid: &GridConfig{
				Rows:    3,
				Columns: 3,
				Gap:     "10px",
			},
			IsDefault: true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	case "flex":
		return &Layout{
			ID:          "flex",
			Name:        "Flex Layout",
			Description: "Default flex layout",
			Type:        LayoutTypeFlex,
			Flex: &FlexConfig{
				Direction: "column",
				Wrap:      "nowrap",
				Justify:   "flex-start",
				Align:     "stretch",
				Gap:       "10px",
			},
			IsDefault: false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	default:
		return p.createDefaultLayout("grid")
	}
}