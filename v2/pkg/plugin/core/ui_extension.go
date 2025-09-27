package plugin

import (
	"context"
	"time"
)

// UIExtensionPlugin UI扩展插件接口
type UIExtensionPlugin interface {
	Plugin

	// 组件管理
	RegisterComponent(ctx context.Context, component *UIComponent) error
	UnregisterComponent(ctx context.Context, componentID string) error
	GetComponent(ctx context.Context, componentID string) (*UIComponent, error)
	ListComponents(ctx context.Context) ([]*UIComponent, error)

	// 主题管理
	RegisterTheme(ctx context.Context, theme *Theme) error
	UnregisterTheme(ctx context.Context, themeID string) error
	GetTheme(ctx context.Context, themeID string) (*Theme, error)
	ListThemes(ctx context.Context) ([]*Theme, error)
	ApplyTheme(ctx context.Context, themeID string) error

	// 布局管理
	RegisterLayout(ctx context.Context, layout *Layout) error
	UnregisterLayout(ctx context.Context, layoutID string) error
	GetLayout(ctx context.Context, layoutID string) (*Layout, error)
	ListLayouts(ctx context.Context) ([]*Layout, error)
	ApplyLayout(ctx context.Context, layoutID string) error

	// 快捷键管理
	RegisterShortcut(ctx context.Context, shortcut *Shortcut) error
	UnregisterShortcut(ctx context.Context, shortcutID string) error
	GetShortcut(ctx context.Context, shortcutID string) (*Shortcut, error)
	ListShortcuts(ctx context.Context) ([]*Shortcut, error)

	// 菜单管理
	RegisterMenu(ctx context.Context, menu *Menu) error
	UnregisterMenu(ctx context.Context, menuID string) error
	GetMenu(ctx context.Context, menuID string) (*Menu, error)
	ListMenus(ctx context.Context) ([]*Menu, error)

	// 工具栏管理
	RegisterToolbar(ctx context.Context, toolbar *Toolbar) error
	UnregisterToolbar(ctx context.Context, toolbarID string) error
	GetToolbar(ctx context.Context, toolbarID string) (*Toolbar, error)
	ListToolbars(ctx context.Context) ([]*Toolbar, error)

	// 窗口管理
	CreateWindow(ctx context.Context, window *WindowConfig) (*Window, error)
	CloseWindow(ctx context.Context, windowID string) error
	GetWindow(ctx context.Context, windowID string) (*Window, error)
	ListWindows(ctx context.Context) ([]*Window, error)
	FocusWindow(ctx context.Context, windowID string) error

	// 对话框管理
	ShowDialog(ctx context.Context, dialog *Dialog) (*DialogResult, error)
	CloseDialog(ctx context.Context, dialogID string) error

	// 通知管理
	ShowNotification(ctx context.Context, notification *UINotification) error
	HideNotification(ctx context.Context, notificationID string) error

	// 状态栏管理
	UpdateStatusBar(ctx context.Context, status *StatusBarUpdate) error
	GetStatusBar(ctx context.Context) (*StatusBar, error)

	// 获取支持的UI类型
	GetSupportedUITypes() []UIType

	// 获取UI配置
	GetUIConfig() *UIConfig
}

// UIComponent UI组件
type UIComponent struct {
	ID          string            `json:"id"`           // 组件ID
	Name        string            `json:"name"`         // 组件名称
	Type        ComponentType     `json:"type"`         // 组件类型
	Version     string            `json:"version"`      // 版本
	Description string            `json:"description"`  // 描述
	Author      string            `json:"author"`       // 作者
	Icon        string            `json:"icon"`         // 图标
	Template    string            `json:"template"`     // 模板
	Styles      string            `json:"styles"`       // 样式
	Script      string            `json:"script"`       // 脚本
	Props       map[string]interface{} `json:"props"`   // 属性
	Events      []string          `json:"events"`       // 事件
	Dependencies []string         `json:"dependencies"` // 依赖
	Position    *Position         `json:"position"`     // 位置
	Size        *Size             `json:"size"`         // 大小
	Visible     bool              `json:"visible"`      // 是否可见
	Enabled     bool              `json:"enabled"`      // 是否启用
	CreatedAt   time.Time         `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time         `json:"updated_at"`   // 更新时间
}

// ComponentType 组件类型枚举
type ComponentType int

const (
	ComponentTypeButton ComponentType = iota
	ComponentTypeInput
	ComponentTypeLabel
	ComponentTypeImage
	ComponentTypeList
	ComponentTypeTable
	ComponentTypeTree
	ComponentTypeChart
	ComponentTypePlayer
	ComponentTypePlaylist
	ComponentTypeLyrics
	ComponentTypeVisualizer
	ComponentTypeEqualizer
	ComponentTypeCustom
)

// String 返回组件类型的字符串表示
func (c ComponentType) String() string {
	switch c {
	case ComponentTypeButton:
		return "button"
	case ComponentTypeInput:
		return "input"
	case ComponentTypeLabel:
		return "label"
	case ComponentTypeImage:
		return "image"
	case ComponentTypeList:
		return "list"
	case ComponentTypeTable:
		return "table"
	case ComponentTypeTree:
		return "tree"
	case ComponentTypeChart:
		return "chart"
	case ComponentTypePlayer:
		return "player"
	case ComponentTypePlaylist:
		return "playlist"
	case ComponentTypeLyrics:
		return "lyrics"
	case ComponentTypeVisualizer:
		return "visualizer"
	case ComponentTypeEqualizer:
		return "equalizer"
	case ComponentTypeCustom:
		return "custom"
	default:
		return "unknown"
	}
}

// Position 位置信息
type Position struct {
	X int `json:"x"` // X坐标
	Y int `json:"y"` // Y坐标
	Z int `json:"z"` // Z坐标（层级）
}

// Size 大小信息
type Size struct {
	Width  int `json:"width"`  // 宽度
	Height int `json:"height"` // 高度
}

// Theme 主题
type Theme struct {
	ID          string            `json:"id"`          // 主题ID
	Name        string            `json:"name"`        // 主题名称
	Version     string            `json:"version"`     // 版本
	Description string            `json:"description"` // 描述
	Author      string            `json:"author"`      // 作者
	Preview     string            `json:"preview"`     // 预览图
	Colors      *ColorScheme      `json:"colors"`      // 颜色方案
	Fonts       *FontScheme       `json:"fonts"`       // 字体方案
	Spacing     *SpacingScheme    `json:"spacing"`     // 间距方案
	Borders     *BorderScheme     `json:"borders"`     // 边框方案
	Shadows     *ShadowScheme     `json:"shadows"`     // 阴影方案
	Animations  *AnimationScheme  `json:"animations"`  // 动画方案
	CustomCSS   string            `json:"custom_css"`  // 自定义CSS
	Variables   map[string]string `json:"variables"`   // 变量
	IsDark      bool              `json:"is_dark"`     // 是否暗色主题
	IsDefault   bool              `json:"is_default"`  // 是否默认主题
	CreatedAt   time.Time         `json:"created_at"`  // 创建时间
	UpdatedAt   time.Time         `json:"updated_at"`  // 更新时间
}

// ColorScheme 颜色方案
type ColorScheme struct {
	Primary     string `json:"primary"`     // 主色
	Secondary   string `json:"secondary"`   // 次色
	Accent      string `json:"accent"`      // 强调色
	Background  string `json:"background"`  // 背景色
	Surface     string `json:"surface"`     // 表面色
	Text        string `json:"text"`        // 文本色
	TextSecondary string `json:"text_secondary"` // 次要文本色
	Border      string `json:"border"`      // 边框色
	Error       string `json:"error"`       // 错误色
	Warning     string `json:"warning"`     // 警告色
	Success     string `json:"success"`     // 成功色
	Info        string `json:"info"`        // 信息色
}

// FontScheme 字体方案
type FontScheme struct {
	Primary   *FontConfig `json:"primary"`   // 主字体
	Secondary *FontConfig `json:"secondary"` // 次字体
	Monospace *FontConfig `json:"monospace"` // 等宽字体
	Display   *FontConfig `json:"display"`   // 显示字体
}

// FontConfig 字体配置
type FontConfig struct {
	Family string `json:"family"` // 字体族
	Size   int    `json:"size"`   // 字体大小
	Weight string `json:"weight"` // 字体粗细
	Style  string `json:"style"`  // 字体样式
}

// SpacingScheme 间距方案
type SpacingScheme struct {
	XS string `json:"xs"` // 极小间距
	SM string `json:"sm"` // 小间距
	MD string `json:"md"` // 中间距
	LG string `json:"lg"` // 大间距
	XL string `json:"xl"` // 极大间距
}

// BorderScheme 边框方案
type BorderScheme struct {
	Width  string `json:"width"`  // 边框宽度
	Style  string `json:"style"`  // 边框样式
	Radius string `json:"radius"` // 边框圆角
}

// ShadowScheme 阴影方案
type ShadowScheme struct {
	Small  string `json:"small"`  // 小阴影
	Medium string `json:"medium"` // 中阴影
	Large  string `json:"large"`  // 大阴影
}

// AnimationScheme 动画方案
type AnimationScheme struct {
	Duration string `json:"duration"` // 动画时长
	Easing   string `json:"easing"`   // 缓动函数
	Delay    string `json:"delay"`    // 延迟时间
}

// Layout 布局
type Layout struct {
	ID          string            `json:"id"`          // 布局ID
	Name        string            `json:"name"`        // 布局名称
	Description string            `json:"description"` // 描述
	Type        LayoutType        `json:"type"`        // 布局类型
	Template    string            `json:"template"`    // 布局模板
	Areas       []*LayoutArea     `json:"areas"`       // 布局区域
	Grid        *GridConfig       `json:"grid"`        // 网格配置
	Flex        *FlexConfig       `json:"flex"`        // 弹性配置
	Responsive  *ResponsiveConfig `json:"responsive"`  // 响应式配置
	IsDefault   bool              `json:"is_default"`  // 是否默认布局
	CreatedAt   time.Time         `json:"created_at"`  // 创建时间
	UpdatedAt   time.Time         `json:"updated_at"`  // 更新时间
}

// LayoutType 布局类型枚举
type LayoutType int

const (
	LayoutTypeGrid LayoutType = iota
	LayoutTypeFlex
	LayoutTypeAbsolute
	LayoutTypeFloat
	LayoutTypeTable
	LayoutTypeCustom
)

// LayoutArea 布局区域
type LayoutArea struct {
	ID         string    `json:"id"`         // 区域ID
	Name       string    `json:"name"`       // 区域名称
	Type       string    `json:"type"`       // 区域类型
	Position   *Position `json:"position"`   // 位置
	Size       *Size     `json:"size"`       // 大小
	Components []string  `json:"components"` // 组件列表
	Visible    bool      `json:"visible"`    // 是否可见
	Resizable  bool      `json:"resizable"`  // 是否可调整大小
	Movable    bool      `json:"movable"`    // 是否可移动
}

// GridConfig 网格配置
type GridConfig struct {
	Rows    int    `json:"rows"`    // 行数
	Columns int    `json:"columns"` // 列数
	Gap     string `json:"gap"`     // 间隙
}

// FlexConfig 弹性配置
type FlexConfig struct {
	Direction string `json:"direction"` // 方向
	Wrap      string `json:"wrap"`      // 换行
	Justify   string `json:"justify"`   // 主轴对齐
	Align     string `json:"align"`     // 交叉轴对齐
	Gap       string `json:"gap"`       // 间隙
}

// ResponsiveConfig 响应式配置
type ResponsiveConfig struct {
	Breakpoints map[string]int    `json:"breakpoints"` // 断点
	Layouts     map[string]string `json:"layouts"`     // 不同断点的布局
}

// Shortcut 快捷键
type Shortcut struct {
	ID          string    `json:"id"`          // 快捷键ID
	Name        string    `json:"name"`        // 快捷键名称
	Description string    `json:"description"` // 描述
	Key         string    `json:"key"`         // 按键组合
	Action      string    `json:"action"`      // 动作
	Context     string    `json:"context"`     // 上下文
	Enabled     bool      `json:"enabled"`     // 是否启用
	Global      bool      `json:"global"`      // 是否全局
	CreatedAt   time.Time `json:"created_at"`  // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`  // 更新时间
}

// Menu 菜单
type Menu struct {
	ID        string      `json:"id"`        // 菜单ID
	Name      string      `json:"name"`      // 菜单名称
	Type      MenuType    `json:"type"`      // 菜单类型
	Items     []*MenuItem `json:"items"`     // 菜单项
	Position  string      `json:"position"`  // 位置
	Visible   bool        `json:"visible"`   // 是否可见
	Enabled   bool        `json:"enabled"`   // 是否启用
	CreatedAt time.Time   `json:"created_at"` // 创建时间
	UpdatedAt time.Time   `json:"updated_at"` // 更新时间
}

// MenuType 菜单类型枚举
type MenuType int

const (
	MenuTypeMain MenuType = iota
	MenuTypeContext
	MenuTypePopup
	MenuTypeDropdown
	MenuTypeTab
)

// MenuItem 菜单项
type MenuItem struct {
	ID        string      `json:"id"`        // 菜单项ID
	Label     string      `json:"label"`     // 标签
	Icon      string      `json:"icon"`      // 图标
	Action    string      `json:"action"`    // 动作
	Shortcut  string      `json:"shortcut"`  // 快捷键
	Submenu   []*MenuItem `json:"submenu"`   // 子菜单
	Separator bool        `json:"separator"` // 是否分隔符
	Visible   bool        `json:"visible"`   // 是否可见
	Enabled   bool        `json:"enabled"`   // 是否启用
	Checked   bool        `json:"checked"`   // 是否选中
}

// Toolbar 工具栏
type Toolbar struct {
	ID        string         `json:"id"`        // 工具栏ID
	Name      string         `json:"name"`      // 工具栏名称
	Position  ToolbarPosition `json:"position"`  // 位置
	Buttons   []*ToolbarButton `json:"buttons"`   // 按钮
	Visible   bool           `json:"visible"`   // 是否可见
	Enabled   bool           `json:"enabled"`   // 是否启用
	Movable   bool           `json:"movable"`   // 是否可移动
	CreatedAt time.Time      `json:"created_at"` // 创建时间
	UpdatedAt time.Time      `json:"updated_at"` // 更新时间
}

// ToolbarPosition 工具栏位置枚举
type ToolbarPosition int

const (
	ToolbarPositionTop ToolbarPosition = iota
	ToolbarPositionBottom
	ToolbarPositionLeft
	ToolbarPositionRight
	ToolbarPositionFloating
)

// ToolbarButton 工具栏按钮
type ToolbarButton struct {
	ID       string `json:"id"`       // 按钮ID
	Label    string `json:"label"`    // 标签
	Icon     string `json:"icon"`     // 图标
	Action   string `json:"action"`   // 动作
	Tooltip  string `json:"tooltip"`  // 提示
	Visible  bool   `json:"visible"`  // 是否可见
	Enabled  bool   `json:"enabled"`  // 是否启用
	Pressed  bool   `json:"pressed"`  // 是否按下
	Toggle   bool   `json:"toggle"`   // 是否切换按钮
}

// WindowConfig 窗口配置
type WindowConfig struct {
	Title     string    `json:"title"`     // 标题
	Type      WindowType `json:"type"`      // 窗口类型
	Position  *Position `json:"position"`  // 位置
	Size      *Size     `json:"size"`      // 大小
	MinSize   *Size     `json:"min_size"`  // 最小大小
	MaxSize   *Size     `json:"max_size"`  // 最大大小
	Resizable bool      `json:"resizable"` // 是否可调整大小
	Movable   bool      `json:"movable"`   // 是否可移动
	Closable  bool      `json:"closable"`  // 是否可关闭
	Modal     bool      `json:"modal"`     // 是否模态
	AlwaysOnTop bool    `json:"always_on_top"` // 是否置顶
	Content   string    `json:"content"`   // 内容
}

// WindowType 窗口类型枚举
type WindowType int

const (
	WindowTypeMain WindowType = iota
	WindowTypeDialog
	WindowTypePopup
	WindowTypeTooltip
	WindowTypeNotification
	WindowTypeFloating
)

// Window 窗口
type Window struct {
	ID        string        `json:"id"`        // 窗口ID
	Config    *WindowConfig `json:"config"`    // 配置
	State     WindowState   `json:"state"`     // 状态
	Visible   bool          `json:"visible"`   // 是否可见
	Focused   bool          `json:"focused"`   // 是否聚焦
	CreatedAt time.Time     `json:"created_at"` // 创建时间
	UpdatedAt time.Time     `json:"updated_at"` // 更新时间
}

// WindowState 窗口状态枚举
type WindowState int

const (
	WindowStateNormal WindowState = iota
	WindowStateMinimized
	WindowStateMaximized
	WindowStateFullscreen
	WindowStateHidden
)

// Dialog 对话框
type Dialog struct {
	ID       string      `json:"id"`       // 对话框ID
	Type     DialogType  `json:"type"`     // 对话框类型
	Title    string      `json:"title"`    // 标题
	Message  string      `json:"message"`  // 消息
	Icon     string      `json:"icon"`     // 图标
	Buttons  []*DialogButton `json:"buttons"`  // 按钮
	Inputs   []*DialogInput  `json:"inputs"`   // 输入框
	Modal    bool        `json:"modal"`    // 是否模态
	Closable bool        `json:"closable"` // 是否可关闭
	Timeout  *time.Duration `json:"timeout"`  // 超时时间
}

// DialogType 对话框类型枚举
type DialogType int

const (
	DialogTypeInfo DialogType = iota
	DialogTypeWarning
	DialogTypeError
	DialogTypeConfirm
	DialogTypePrompt
	DialogTypeCustom
)

// DialogButton 对话框按钮
type DialogButton struct {
	ID      string `json:"id"`      // 按钮ID
	Label   string `json:"label"`   // 标签
	Type    string `json:"type"`    // 类型
	Default bool   `json:"default"` // 是否默认
	Cancel  bool   `json:"cancel"`  // 是否取消
}

// DialogInput 对话框输入框
type DialogInput struct {
	ID          string `json:"id"`          // 输入框ID
	Label       string `json:"label"`       // 标签
	Type        string `json:"type"`        // 类型
	Placeholder string `json:"placeholder"` // 占位符
	Value       string `json:"value"`       // 值
	Required    bool   `json:"required"`    // 是否必填
}

// DialogResult 对话框结果
type DialogResult struct {
	ButtonID string            `json:"button_id"` // 按钮ID
	Inputs   map[string]string `json:"inputs"`    // 输入值
	Cancelled bool             `json:"cancelled"` // 是否取消
}

// UINotification UI通知
type UINotification struct {
	ID       string               `json:"id"`       // 通知ID
	Type     UINotificationType   `json:"type"`     // 通知类型
	Title    string               `json:"title"`    // 标题
	Message  string               `json:"message"`  // 消息
	Icon     string               `json:"icon"`     // 图标
	Position NotificationPosition `json:"position"` // 位置
	Duration *time.Duration       `json:"duration"` // 持续时间
	Actions  []*NotificationAction `json:"actions"`  // 动作
	Closable bool                 `json:"closable"` // 是否可关闭
}

// UINotificationType UI通知类型枚举
type UINotificationType int

const (
	UINotificationTypeInfo UINotificationType = iota
	UINotificationTypeSuccess
	UINotificationTypeWarning
	UINotificationTypeError
	UINotificationTypeProgress
)

// NotificationPosition 通知位置枚举
type NotificationPosition int

const (
	NotificationPositionTopLeft NotificationPosition = iota
	NotificationPositionTopCenter
	NotificationPositionTopRight
	NotificationPositionBottomLeft
	NotificationPositionBottomCenter
	NotificationPositionBottomRight
	NotificationPositionCenter
)

// NotificationAction 通知动作
type NotificationAction struct {
	ID    string `json:"id"`    // 动作ID
	Label string `json:"label"` // 标签
	Action string `json:"action"` // 动作
}

// StatusBarUpdate 状态栏更新
type StatusBarUpdate struct {
	Text     string `json:"text"`     // 文本
	Icon     string `json:"icon"`     // 图标
	Progress *int   `json:"progress"` // 进度（0-100）
	Color    string `json:"color"`    // 颜色
	Blink    bool   `json:"blink"`    // 是否闪烁
}

// StatusBar 状态栏
type StatusBar struct {
	Text     string    `json:"text"`       // 文本
	Icon     string    `json:"icon"`       // 图标
	Progress int       `json:"progress"`   // 进度
	Color    string    `json:"color"`      // 颜色
	Blink    bool      `json:"blink"`      // 是否闪烁
	Visible  bool      `json:"visible"`    // 是否可见
	UpdatedAt time.Time `json:"updated_at"` // 更新时间
}

// UIType UI类型枚举
type UIType int

const (
	UITypeDesktop UIType = iota
	UITypeWeb
	UITypeMobile
	UITypeTerminal
	UITypeEmbedded
)

// String 返回UI类型的字符串表示
func (u UIType) String() string {
	switch u {
	case UITypeDesktop:
		return "desktop"
	case UITypeWeb:
		return "web"
	case UITypeMobile:
		return "mobile"
	case UITypeTerminal:
		return "terminal"
	case UITypeEmbedded:
		return "embedded"
	default:
		return "unknown"
	}
}

// UIConfig UI配置
type UIConfig struct {
	Types       []UIType          `json:"types"`        // 支持的UI类型
	Themes      []string          `json:"themes"`       // 支持的主题
	Layouts     []string          `json:"layouts"`      // 支持的布局
	Components  []string          `json:"components"`   // 支持的组件
	Features    []UIFeature       `json:"features"`     // 支持的功能
	Settings    map[string]interface{} `json:"settings"` // 设置
	Constraints *UIConstraints    `json:"constraints"`  // 约束
}

// UIFeature UI功能枚举
type UIFeature int

const (
	UIFeatureThemes UIFeature = iota
	UIFeatureLayouts
	UIFeatureShortcuts
	UIFeatureMenus
	UIFeatureToolbars
	UIFeatureWindows
	UIFeatureDialogs
	UIFeatureNotifications
	UIFeatureStatusBar
	UIFeatureCustomComponents
	UIFeatureResponsive
	UIFeatureAccessibility
)

// UIConstraints UI约束
type UIConstraints struct {
	MaxWindows     int `json:"max_windows"`     // 最大窗口数
	MaxComponents  int `json:"max_components"`  // 最大组件数
	MaxThemes      int `json:"max_themes"`      // 最大主题数
	MaxLayouts     int `json:"max_layouts"`     // 最大布局数
	MaxShortcuts   int `json:"max_shortcuts"`   // 最大快捷键数
	MaxMenus       int `json:"max_menus"`       // 最大菜单数
	MaxToolbars    int `json:"max_toolbars"`    // 最大工具栏数
	MaxNotifications int `json:"max_notifications"` // 最大通知数
}