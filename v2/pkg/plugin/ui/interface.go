package ui

import (
	"context"
	"time"

	plugin "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
)

// UIExtensionPlugin UI扩展插件接口
type UIExtensionPlugin interface {
	plugin.Plugin

	// UI渲染
	Render(ctx context.Context, state *AppState) error
	HandleInput(ctx context.Context, input *InputEvent) error

	// 布局管理
	GetLayout() *Layout
	SetLayout(layout *Layout) error
	GetSupportedLayouts() []*Layout

	// 主题支持
	GetSupportedThemes() []*Theme
	ApplyTheme(theme *Theme) error
	GetCurrentTheme() *Theme

	// 热更新
	CanHotReload() bool
	HotReload(newVersion []byte) error
	GetVersion() string

	// UI类型支持
	GetSupportedUITypes() []UIType
	SetUIType(uiType UIType) error
	GetCurrentUIType() UIType

	// 组件管理
	RegisterComponent(ctx context.Context, component *UIComponent) error
	UnregisterComponent(ctx context.Context, componentID string) error
	GetComponent(ctx context.Context, componentID string) (*UIComponent, error)
	ListComponents(ctx context.Context) ([]*UIComponent, error)

	// 事件处理
	RegisterEventHandler(eventType string, handler EventHandler) error
	UnregisterEventHandler(eventType string) error
}

// AppState 应用状态
type AppState struct {
	Player      *PlayerState      `json:"player"`
	CurrentView string            `json:"current_view"`
	User        *User             `json:"user"`
	Config      map[string]string `json:"config"`
	Plugins     []*plugin.PluginInfo `json:"plugins"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// PlayerState 播放器状态
type PlayerState struct {
	Status      PlayStatus    `json:"status"`
	CurrentSong *Song         `json:"current_song"`
	Position    time.Duration `json:"position"`
	Duration    time.Duration `json:"duration"`
	Volume      float64       `json:"volume"`
	IsMuted     bool          `json:"is_muted"`
	PlayMode    PlayMode      `json:"play_mode"`
	Queue       []*Song       `json:"queue"`
	History     []*Song       `json:"history"`
}

// PlayStatus 播放状态枚举
type PlayStatus int

const (
	PlayStatusStopped PlayStatus = iota
	PlayStatusPlaying
	PlayStatusPaused
	PlayStatusBuffering
	PlayStatusError
)

// PlayMode 播放模式枚举
type PlayMode int

const (
	PlayModeSequential PlayMode = iota
	PlayModeRepeatOne
	PlayModeRepeatAll
	PlayModeShuffle
)

// Song 歌曲信息
type Song struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Artist      string            `json:"artist"`
	Album       string            `json:"album"`
	Duration    time.Duration     `json:"duration"`
	Source      string            `json:"source"`
	URL         string            `json:"url"`
	CoverURL    string            `json:"cover_url"`
	Quality     Quality           `json:"quality"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Quality 音质枚举
type Quality int

const (
	QualityLow Quality = iota
	QualityMedium
	QualityHigh
	QualityLossless
)

// User 用户信息
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Email    string `json:"email"`
}

// InputEvent 输入事件
type InputEvent struct {
	Type      InputType         `json:"type"`
	Key       string            `json:"key"`
	Modifiers []string          `json:"modifiers"`
	Position  *Position         `json:"position"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time         `json:"timestamp"`
}

// InputType 输入类型枚举
type InputType int

const (
	InputTypeKeyboard InputType = iota
	InputTypeMouse
	InputTypeTouch
	InputTypeGamepad
	InputTypeVoice
)

// Position 位置信息
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

// Layout 布局
type Layout struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        LayoutType        `json:"type"`
	Template    string            `json:"template"`
	Areas       []*LayoutArea     `json:"areas"`
	Grid        *GridConfig       `json:"grid"`
	Flex        *FlexConfig       `json:"flex"`
	Responsive  *ResponsiveConfig `json:"responsive"`
	IsDefault   bool              `json:"is_default"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
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
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Position   *Position `json:"position"`
	Size       *Size     `json:"size"`
	Components []string  `json:"components"`
	Visible    bool      `json:"visible"`
	Resizable  bool      `json:"resizable"`
	Movable    bool      `json:"movable"`
}

// Size 大小信息
type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// GridConfig 网格配置
type GridConfig struct {
	Rows    int    `json:"rows"`
	Columns int    `json:"columns"`
	Gap     string `json:"gap"`
}

// FlexConfig 弹性配置
type FlexConfig struct {
	Direction string `json:"direction"`
	Wrap      string `json:"wrap"`
	Justify   string `json:"justify"`
	Align     string `json:"align"`
	Gap       string `json:"gap"`
}

// ResponsiveConfig 响应式配置
type ResponsiveConfig struct {
	Breakpoints map[string]int    `json:"breakpoints"`
	Layouts     map[string]string `json:"layouts"`
}

// Theme 主题
type Theme struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Preview     string            `json:"preview"`
	Colors      *ColorScheme      `json:"colors"`
	Fonts       *FontScheme       `json:"fonts"`
	Spacing     *SpacingScheme    `json:"spacing"`
	Borders     *BorderScheme     `json:"borders"`
	Shadows     *ShadowScheme     `json:"shadows"`
	Animations  *AnimationScheme  `json:"animations"`
	CustomCSS   string            `json:"custom_css"`
	Variables   map[string]string `json:"variables"`
	IsDark      bool              `json:"is_dark"`
	IsDefault   bool              `json:"is_default"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// ColorScheme 颜色方案
type ColorScheme struct {
	Primary       string `json:"primary"`
	Secondary     string `json:"secondary"`
	Accent        string `json:"accent"`
	Background    string `json:"background"`
	Surface       string `json:"surface"`
	Text          string `json:"text"`
	TextSecondary string `json:"text_secondary"`
	Border        string `json:"border"`
	Error         string `json:"error"`
	Warning       string `json:"warning"`
	Success       string `json:"success"`
	Info          string `json:"info"`
}

// FontScheme 字体方案
type FontScheme struct {
	Primary   *FontConfig `json:"primary"`
	Secondary *FontConfig `json:"secondary"`
	Monospace *FontConfig `json:"monospace"`
	Display   *FontConfig `json:"display"`
}

// FontConfig 字体配置
type FontConfig struct {
	Family string `json:"family"`
	Size   int    `json:"size"`
	Weight string `json:"weight"`
	Style  string `json:"style"`
}

// SpacingScheme 间距方案
type SpacingScheme struct {
	XS string `json:"xs"`
	SM string `json:"sm"`
	MD string `json:"md"`
	LG string `json:"lg"`
	XL string `json:"xl"`
}

// BorderScheme 边框方案
type BorderScheme struct {
	Width  string `json:"width"`
	Style  string `json:"style"`
	Radius string `json:"radius"`
}

// ShadowScheme 阴影方案
type ShadowScheme struct {
	Small  string `json:"small"`
	Medium string `json:"medium"`
	Large  string `json:"large"`
}

// AnimationScheme 动画方案
type AnimationScheme struct {
	Duration string `json:"duration"`
	Easing   string `json:"easing"`
	Delay    string `json:"delay"`
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

// UIComponent UI组件
type UIComponent struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         ComponentType          `json:"type"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Author       string                 `json:"author"`
	Icon         string                 `json:"icon"`
	Template     string                 `json:"template"`
	Styles       string                 `json:"styles"`
	Script       string                 `json:"script"`
	Props        map[string]interface{} `json:"props"`
	Events       []string               `json:"events"`
	Dependencies []string               `json:"dependencies"`
	Position     *Position              `json:"position"`
	Size         *Size                  `json:"size"`
	Visible      bool                   `json:"visible"`
	Enabled      bool                   `json:"enabled"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
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

// UIEvent UI事件
type UIEvent struct {
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Target    string                 `json:"target"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Cancelled bool                   `json:"cancelled"`
}

// EventHandler 事件处理器
type EventHandler interface {
	Handle(ctx context.Context, event *UIEvent) error
	GetType() string
	GetPriority() int
}

// UIRenderer UI渲染器接口
type UIRenderer interface {
	Render(ctx context.Context, component *UIComponent, state *AppState) ([]byte, error)
	SupportsType(uiType UIType) bool
	GetSupportedTypes() []UIType
}

// InputHandler 输入处理器接口
type InputHandler interface {
	Handle(ctx context.Context, event *InputEvent) error
	SupportsType(inputType InputType) bool
	GetSupportedTypes() []InputType
}

// LayoutManager 布局管理器接口
type LayoutManager interface {
	ApplyLayout(ctx context.Context, layout *Layout) error
	GetCurrentLayout() *Layout
	ValidateLayout(layout *Layout) error
	GetSupportedTypes() []LayoutType
}

// ThemeManager 主题管理器接口
type ThemeManager interface {
	ApplyTheme(ctx context.Context, theme *Theme) error
	GetCurrentTheme() *Theme
	ValidateTheme(theme *Theme) error
	GenerateCSS(theme *Theme) (string, error)
}

// HotReloader 热重载器接口
type HotReloader interface {
	Reload(ctx context.Context, data []byte) error
	CanReload() bool
	GetVersion() string
	ValidateVersion(version string) error
}