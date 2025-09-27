package config

import (
	"time"
)

// TUIConfig TUI插件配置
type TUIConfig struct {
	// 基础配置
	Enabled     bool   `json:"enabled" yaml:"enabled"`
	Theme       string `json:"theme" yaml:"theme"`
	Language    string `json:"language" yaml:"language"`
	DefaultView string `json:"default_view" yaml:"default_view"`
	
	// 界面配置
	DualColumn      bool `json:"dual_column" yaml:"dual_column"`
	DynamicRowCount bool `json:"dynamic_row_count" yaml:"dynamic_row_count"`
	MaxMenuStartRow int  `json:"max_menu_start_row" yaml:"max_menu_start_row"`
	DisplayTitle    bool `json:"display_title" yaml:"display_title"`
	
	// 歌词配置
	ShowLyrics   bool `json:"show_lyrics" yaml:"show_lyrics"`
	LyricLines   int  `json:"lyric_lines" yaml:"lyric_lines"`
	LyricOffset  int  `json:"lyric_offset" yaml:"lyric_offset"`
	
	// 颜色配置
	PrimaryColor   string `json:"primary_color" yaml:"primary_color"`
	SecondaryColor string `json:"secondary_color" yaml:"secondary_color"`
	AccentColor    string `json:"accent_color" yaml:"accent_color"`
	
	// 性能配置
	RefreshInterval time.Duration `json:"refresh_interval" yaml:"refresh_interval"`
	MaxComponents   int           `json:"max_components" yaml:"max_components"`
	
	// 键盘快捷键配置
	KeyBindings map[string]string `json:"key_bindings" yaml:"key_bindings"`
}

// NewTUIConfig 创建默认TUI配置
func NewTUIConfig() *TUIConfig {
	return &TUIConfig{
		// 基础配置
		Enabled:     true,
		Theme:       "default",
		Language:    "zh-CN",
		DefaultView: "main",
		
		// 界面配置
		DualColumn:      true,
		DynamicRowCount: true,
		MaxMenuStartRow: 10,
		DisplayTitle:    true,
		
		// 歌词配置
		ShowLyrics:  true,
		LyricLines:  5,
		LyricOffset: 0,
		
		// 颜色配置
		PrimaryColor:   "#1DB954", // Spotify绿色
		SecondaryColor: "#191414", // Spotify黑色
		AccentColor:    "#1ED760", // Spotify亮绿色
		
		// 性能配置
		RefreshInterval: 100 * time.Millisecond,
		MaxComponents:   50,
		
		// 键盘快捷键配置
		KeyBindings: getDefaultKeyBindings(),
	}
}

// getDefaultKeyBindings 获取默认键绑定
func getDefaultKeyBindings() map[string]string {
	return map[string]string{
		"q":      "quit",
		"Q":      "quit",
		"ctrl+c": "quit",
		"space":  "toggle_play",
		"n":      "next_song",
		"p":      "prev_song",
		"s":      "search",
		"l":      "toggle_lyrics",
		"m":      "toggle_play_mode",
		"up":     "menu_up",
		"down":   "menu_down",
		"enter":  "menu_select",
		"esc":    "menu_back",
		"left":   "seek_backward",
		"right":  "seek_forward",
		"+":      "volume_up",
		"=":      "volume_up",
		"-":      "volume_down",
		"0":      "toggle_mute",
		"j":      "menu_down",
		"k":      "menu_up",
		"h":      "menu_back",
		"L":      "menu_select",
		"/":      "search",
		"?":      "help",
		"r":      "refresh",
		"f":      "favorite",
		"d":      "download",
		"c":      "copy",
		"v":      "volume",
		"1":      "play_mode_sequence",
		"2":      "play_mode_loop",
		"3":      "play_mode_random",
		"4":      "play_mode_single",
	}
}

// Validate 验证配置
func (c *TUIConfig) Validate() error {
	// 验证主题
	if c.Theme == "" {
		c.Theme = "default"
	}
	
	// 验证语言
	validLanguages := []string{"zh-CN", "en-US", "ja-JP"}
	validLang := false
	for _, lang := range validLanguages {
		if c.Language == lang {
			validLang = true
			break
		}
	}
	if !validLang {
		c.Language = "zh-CN"
	}
	
	// 验证歌词行数
	if c.LyricLines < 3 {
		c.LyricLines = 3
	} else if c.LyricLines > 7 {
		c.LyricLines = 7
	}
	
	// 验证刷新间隔
	if c.RefreshInterval < 50*time.Millisecond {
		c.RefreshInterval = 50 * time.Millisecond
	} else if c.RefreshInterval > 1*time.Second {
		c.RefreshInterval = 1 * time.Second
	}
	
	// 验证最大组件数
	if c.MaxComponents < 10 {
		c.MaxComponents = 10
	} else if c.MaxComponents > 100 {
		c.MaxComponents = 100
	}
	
	return nil
}

// Clone 克隆配置
func (c *TUIConfig) Clone() *TUIConfig {
	newConfig := *c
	
	// 深拷贝键绑定
	newConfig.KeyBindings = make(map[string]string)
	for k, v := range c.KeyBindings {
		newConfig.KeyBindings[k] = v
	}
	
	return &newConfig
}

// Merge 合并配置
func (c *TUIConfig) Merge(other *TUIConfig) {
	if other == nil {
		return
	}
	
	if other.Theme != "" {
		c.Theme = other.Theme
	}
	if other.Language != "" {
		c.Language = other.Language
	}
	if other.DefaultView != "" {
		c.DefaultView = other.DefaultView
	}
	if other.PrimaryColor != "" {
		c.PrimaryColor = other.PrimaryColor
	}
	if other.SecondaryColor != "" {
		c.SecondaryColor = other.SecondaryColor
	}
	if other.AccentColor != "" {
		c.AccentColor = other.AccentColor
	}
	
	// 合并键绑定
	if other.KeyBindings != nil {
		if c.KeyBindings == nil {
			c.KeyBindings = make(map[string]string)
		}
		for k, v := range other.KeyBindings {
			c.KeyBindings[k] = v
		}
	}
}

// GetKeyBinding 获取键绑定
func (c *TUIConfig) GetKeyBinding(key string) (string, bool) {
	if c.KeyBindings == nil {
		return "", false
	}
	action, exists := c.KeyBindings[key]
	return action, exists
}

// SetKeyBinding 设置键绑定
func (c *TUIConfig) SetKeyBinding(key, action string) {
	if c.KeyBindings == nil {
		c.KeyBindings = make(map[string]string)
	}
	c.KeyBindings[key] = action
}

// RemoveKeyBinding 移除键绑定
func (c *TUIConfig) RemoveKeyBinding(key string) {
	if c.KeyBindings != nil {
		delete(c.KeyBindings, key)
	}
}