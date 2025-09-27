package tui

import (
	"fmt"
	"time"

	ui "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/ui"
)

// NewMainViewComponent 创建主界面组件
func NewMainViewComponent() *ui.UIComponent {
	return &ui.UIComponent{
		ID:          "main-view",
		Name:        "Main View",
		Type:        ui.ComponentTypeCustom,
		Version:     "1.0.0",
		Description: "Main menu view component",
		Author:      "go-musicfox",
		Template:    "main-view-template",
		Props: map[string]interface{}{
			"title":     "go-musicfox",
			"menu_items": getMainMenuItems(),
		},
		Events:    []string{"menu_select", "menu_navigate"},
		Visible:   true,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewPlayerComponent 创建播放器组件
func NewPlayerComponent() *ui.UIComponent {
	return &ui.UIComponent{
		ID:          "player",
		Name:        "Player",
		Type:        ui.ComponentTypePlayer,
		Version:     "1.0.0",
		Description: "Music player component",
		Author:      "go-musicfox",
		Template:    "player-template",
		Props: map[string]interface{}{
			"show_lyrics":    true,
			"show_progress":  true,
			"show_volume":    true,
			"progress_style": "bar",
		},
		Events: []string{
			"play", "pause", "stop", "next", "previous",
			"volume_change", "seek", "mode_change",
		},
		Visible:   true,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewSearchComponent 创建搜索组件
func NewSearchComponent() *ui.UIComponent {
	return &ui.UIComponent{
		ID:          "search",
		Name:        "Search",
		Type:        ui.ComponentTypeCustom,
		Version:     "1.0.0",
		Description: "Music search component",
		Author:      "go-musicfox",
		Template:    "search-template",
		Props: map[string]interface{}{
			"search_types": []string{"song", "album", "artist", "playlist"},
			"placeholder":  "输入关键词搜索...",
			"max_results":  50,
		},
		Events: []string{
			"search", "result_select", "type_change",
		},
		Visible:   true,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewPlaylistComponent 创建播放列表组件
func NewPlaylistComponent() *ui.UIComponent {
	return &ui.UIComponent{
		ID:          "playlist",
		Name:        "Playlist",
		Type:        ui.ComponentTypePlaylist,
		Version:     "1.0.0",
		Description: "Music playlist component",
		Author:      "go-musicfox",
		Template:    "playlist-template",
		Props: map[string]interface{}{
			"show_index":    true,
			"show_duration": true,
			"show_artist":   true,
			"max_items":     100,
		},
		Events: []string{
			"song_select", "song_play", "song_remove",
			"playlist_clear", "playlist_shuffle",
		},
		Visible:   true,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewLyricsComponent 创建歌词组件
func NewLyricsComponent() *ui.UIComponent {
	return &ui.UIComponent{
		ID:          "lyrics",
		Name:        "Lyrics",
		Type:        ui.ComponentTypeLyrics,
		Version:     "1.0.0",
		Description: "Song lyrics component",
		Author:      "go-musicfox",
		Template:    "lyrics-template",
		Props: map[string]interface{}{
			"lines":         5,
			"auto_scroll":   true,
			"highlight":     true,
			"show_time":     false,
			"translation":   false,
		},
		Events: []string{
			"lyrics_scroll", "lyrics_toggle",
		},
		Visible:   true,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewStatusBarComponent 创建状态栏组件
func NewStatusBarComponent() *ui.UIComponent {
	return &ui.UIComponent{
		ID:          "status-bar",
		Name:        "Status Bar",
		Type:        ui.ComponentTypeCustom,
		Version:     "1.0.0",
		Description: "Application status bar component",
		Author:      "go-musicfox",
		Template:    "status-bar-template",
		Props: map[string]interface{}{
			"show_time":    true,
			"show_version": true,
			"show_user":    true,
			"show_status":  true,
		},
		Events:    []string{"status_update"},
		Visible:   true,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewProgressBarComponent 创建进度条组件
func NewProgressBarComponent() *ui.UIComponent {
	return &ui.UIComponent{
		ID:          "progress-bar",
		Name:        "Progress Bar",
		Type:        ui.ComponentTypeCustom,
		Version:     "1.0.0",
		Description: "Music progress bar component",
		Author:      "go-musicfox",
		Template:    "progress-bar-template",
		Props: map[string]interface{}{
			"width":      40,
			"style":      "bar",
			"show_time":  true,
			"interactive": true,
		},
		Events: []string{
			"progress_change", "seek",
		},
		Visible:   true,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewVolumeBarComponent 创建音量条组件
func NewVolumeBarComponent() *ui.UIComponent {
	return &ui.UIComponent{
		ID:          "volume-bar",
		Name:        "Volume Bar",
		Type:        ui.ComponentTypeCustom,
		Version:     "1.0.0",
		Description: "Volume control bar component",
		Author:      "go-musicfox",
		Template:    "volume-bar-template",
		Props: map[string]interface{}{
			"width":       10,
			"style":       "bar",
			"show_value":  true,
			"interactive": true,
			"mute_button": true,
		},
		Events: []string{
			"volume_change", "mute_toggle",
		},
		Visible:   true,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewMenuComponent 创建菜单组件
func NewMenuComponent(menuID, title string, items []string) *ui.UIComponent {
	return &ui.UIComponent{
		ID:          menuID,
		Name:        title,
		Type:        ui.ComponentTypeList,
		Version:     "1.0.0",
		Description: "Menu list component",
		Author:      "go-musicfox",
		Template:    "menu-template",
		Props: map[string]interface{}{
			"title":         title,
			"items":         items,
			"selected":      0,
			"show_index":    true,
			"show_cursor":   true,
			"cursor_style":  "arrow",
		},
		Events: []string{
			"item_select", "item_navigate",
		},
		Visible:   true,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// getMainMenuItems 获取主菜单项
func getMainMenuItems() []string {
	return []string{
		"每日推荐歌曲",
		"每日推荐歌单",
		"我的歌单",
		"我的收藏",
		"私人FM",
		"专辑列表",
		"搜索",
		"排行榜",
		"精选歌单",
		"热门歌手",
		"最近播放歌曲",
		"云盘",
		"主播电台",
		"LastFM",
		"帮助",
		"检查更新",
	}
}

// ComponentFactory 组件工厂
type ComponentFactory struct{}

// NewComponentFactory 创建组件工厂
func NewComponentFactory() *ComponentFactory {
	return &ComponentFactory{}
}

// CreateComponent 创建组件
func (cf *ComponentFactory) CreateComponent(componentType ui.ComponentType, config map[string]interface{}) *ui.UIComponent {
	switch componentType {
	case ui.ComponentTypePlayer:
		return NewPlayerComponent()
	case ui.ComponentTypePlaylist:
		return NewPlaylistComponent()
	case ui.ComponentTypeLyrics:
		return NewLyricsComponent()
	case ui.ComponentTypeList:
		if title, ok := config["title"].(string); ok {
			if items, ok := config["items"].([]string); ok {
				return NewMenuComponent("menu", title, items)
			}
		}
		return NewMenuComponent("menu", "Menu", []string{})
	case ui.ComponentTypeCustom:
		if id, ok := config["id"].(string); ok {
			switch id {
			case "main-view":
				return NewMainViewComponent()
			case "search":
				return NewSearchComponent()
			case "status-bar":
				return NewStatusBarComponent()
			case "progress-bar":
				return NewProgressBarComponent()
			case "volume-bar":
				return NewVolumeBarComponent()
			}
		}
		return NewMainViewComponent()
	default:
		return NewMainViewComponent()
	}
}

// GetAvailableComponents 获取可用组件列表
func (cf *ComponentFactory) GetAvailableComponents() []string {
	return []string{
		"main-view",
		"player",
		"search",
		"playlist",
		"lyrics",
		"status-bar",
		"progress-bar",
		"volume-bar",
		"menu",
	}
}

// ValidateComponent 验证组件配置
func (cf *ComponentFactory) ValidateComponent(component *ui.UIComponent) error {
	if component == nil {
		return fmt.Errorf("component cannot be nil")
	}

	if component.ID == "" {
		return fmt.Errorf("component ID cannot be empty")
	}

	if component.Name == "" {
		return fmt.Errorf("component name cannot be empty")
	}

	if component.Version == "" {
		return fmt.Errorf("component version cannot be empty")
	}

	return nil
}