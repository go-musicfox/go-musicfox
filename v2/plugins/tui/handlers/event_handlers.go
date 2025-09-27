package handlers

import (
	"context"
	"fmt"
	"time"

	ui "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/ui"
)

// EventHandler 事件处理器接口
type EventHandler interface {
	Handle(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error
	CanHandle(event *ui.UIEvent) bool
	GetPriority() int
}

// BaseEventHandler 基础事件处理器
type BaseEventHandler struct {
	name     string
	priority int
}

// NewBaseEventHandler 创建基础事件处理器
func NewBaseEventHandler(name string, priority int) *BaseEventHandler {
	return &BaseEventHandler{
		name:     name,
		priority: priority,
	}
}

// GetPriority 获取优先级
func (h *BaseEventHandler) GetPriority() int {
	return h.priority
}

// PlaybackEventHandler 播放控制事件处理器
type PlaybackEventHandler struct {
	*BaseEventHandler
}

// NewPlaybackEventHandler 创建播放控制事件处理器
func NewPlaybackEventHandler() *PlaybackEventHandler {
	return &PlaybackEventHandler{
		BaseEventHandler: NewBaseEventHandler("playback", 100),
	}
}

// CanHandle 检查是否可以处理事件
func (h *PlaybackEventHandler) CanHandle(event *ui.UIEvent) bool {
	playbackEvents := []string{
		"play", "pause", "stop", "next", "previous",
		"shuffle", "repeat", "seek", "volume_change",
	}

	for _, eventType := range playbackEvents {
		if event.Type == eventType {
			return true
		}
	}
	return false
}

// Handle 处理播放控制事件
func (h *PlaybackEventHandler) Handle(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	if state.Player == nil {
		return fmt.Errorf("player state is nil")
	}

	switch event.Type {
	case "play":
		return h.handlePlay(ctx, event, state)
	case "pause":
		return h.handlePause(ctx, event, state)
	case "stop":
		return h.handleStop(ctx, event, state)
	case "next":
		return h.handleNext(ctx, event, state)
	case "previous":
		return h.handlePrevious(ctx, event, state)
	case "shuffle":
		return h.handleShuffle(ctx, event, state)
	case "repeat":
		return h.handleRepeat(ctx, event, state)
	case "seek":
		return h.handleSeek(ctx, event, state)
	case "volume_change":
		return h.handleVolumeChange(ctx, event, state)
	default:
		return fmt.Errorf("unsupported playback event: %s", event.Type)
	}
}

func (h *PlaybackEventHandler) handlePlay(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	state.Player.Status = ui.PlayStatusPlaying
	return nil
}

func (h *PlaybackEventHandler) handlePause(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	state.Player.Status = ui.PlayStatusPaused
	return nil
}

func (h *PlaybackEventHandler) handleStop(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	state.Player.Status = ui.PlayStatusStopped
	state.Player.Position = 0
	return nil
}

func (h *PlaybackEventHandler) handleNext(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	// 检查是否有播放列表信息在Config中
	if playlistSongs, ok := state.Config["playlist_songs"]; ok && playlistSongs != "" {
		if currentIndexStr, ok := state.Config["current_index"]; ok {
			// 简单的字符串数字递增
			switch currentIndexStr {
			case "0":
				state.Config["current_index"] = "1"
			case "1":
				state.Config["current_index"] = "2"
			case "2":
				state.Config["current_index"] = "0" // 循环回到开始
			}
		}
		return nil
	}
	
	if state.Player == nil || len(state.Player.Queue) == 0 {
		return fmt.Errorf("no songs in queue")
	}

	// 简单实现：切换到下一首
	// 实际实现需要更复杂的队列管理逻辑
	return nil
}

func (h *PlaybackEventHandler) handlePrevious(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	if state.Player == nil || len(state.Player.History) == 0 {
		return fmt.Errorf("no songs in history")
	}

	// 简单实现：切换到上一首
	// 实际实现需要更复杂的历史管理逻辑
	return nil
}

func (h *PlaybackEventHandler) handleShuffle(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	// 切换播放模式到随机或顺序
	if state.Player.PlayMode == ui.PlayModeShuffle {
		state.Player.PlayMode = ui.PlayModeSequential
	} else {
		state.Player.PlayMode = ui.PlayModeShuffle
	}
	return nil
}

func (h *PlaybackEventHandler) handleRepeat(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	switch state.Player.PlayMode {
	case ui.PlayModeSequential:
		state.Player.PlayMode = ui.PlayModeRepeatOne
	case ui.PlayModeRepeatOne:
		state.Player.PlayMode = ui.PlayModeRepeatAll
	case ui.PlayModeRepeatAll:
		state.Player.PlayMode = ui.PlayModeSequential
	case ui.PlayModeShuffle:
		// 随机模式下切换重复模式
		state.Player.PlayMode = ui.PlayModeRepeatAll
	}
	return nil
}

func (h *PlaybackEventHandler) handleSeek(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	if position, ok := event.Data["position"].(float64); ok {
		// 转换为time.Duration
		state.Player.Position = time.Duration(position * float64(time.Second))
	}
	return nil
}

func (h *PlaybackEventHandler) handleVolumeChange(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	if volume, ok := event.Data["volume"].(float64); ok {
		if volume >= 0 && volume <= 1.0 {
			state.Player.Volume = volume
			state.Player.IsMuted = volume == 0
		}
	}
	return nil
}

// NavigationEventHandler 导航事件处理器
type NavigationEventHandler struct {
	*BaseEventHandler
}

// NewNavigationEventHandler 创建导航事件处理器
func NewNavigationEventHandler() *NavigationEventHandler {
	return &NavigationEventHandler{
		BaseEventHandler: NewBaseEventHandler("navigation", 90),
	}
}

// CanHandle 检查是否可以处理事件
func (h *NavigationEventHandler) CanHandle(event *ui.UIEvent) bool {
	navigationEvents := []string{
		"navigate_up", "navigate_down", "navigate_left", "navigate_right",
		"page_up", "page_down", "home", "end", "select", "back",
		"view_change",
	}

	for _, eventType := range navigationEvents {
		if event.Type == eventType {
			return true
		}
	}
	return false
}

// Handle 处理导航事件
func (h *NavigationEventHandler) Handle(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	switch event.Type {
	case "navigate_up":
		return h.handleNavigateUp(ctx, event, state)
	case "navigate_down":
		return h.handleNavigateDown(ctx, event, state)
	case "navigate_left":
		return h.handleNavigateLeft(ctx, event, state)
	case "navigate_right":
		return h.handleNavigateRight(ctx, event, state)
	case "page_up":
		return h.handlePageUp(ctx, event, state)
	case "page_down":
		return h.handlePageDown(ctx, event, state)
	case "home":
		return h.handleHome(ctx, event, state)
	case "end":
		return h.handleEnd(ctx, event, state)
	case "select":
		return h.handleSelect(ctx, event, state)
	case "back":
		return h.handleBack(ctx, event, state)
	case "view_change":
		return h.handleViewChange(ctx, event, state)
	default:
		return fmt.Errorf("unsupported navigation event: %s", event.Type)
	}
}

func (h *NavigationEventHandler) handleNavigateUp(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	// 导航向上逻辑
	if selectedIndexStr, ok := state.Config["selected_index"]; ok {
		if selectedIndexStr != "0" {
			// 简单的字符串数字递减
			switch selectedIndexStr {
			case "1":
				state.Config["selected_index"] = "0"
			case "2":
				state.Config["selected_index"] = "1"
			case "3":
				state.Config["selected_index"] = "2"
			case "4":
				state.Config["selected_index"] = "3"
			case "5":
				state.Config["selected_index"] = "4"
			case "10":
				state.Config["selected_index"] = "9"
			}
		}
	}
	return nil
}

func (h *NavigationEventHandler) handleNavigateDown(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	// 导航向下逻辑
	if selectedIndexStr, ok := state.Config["selected_index"]; ok {
		// 简单的字符串数字递增
		switch selectedIndexStr {
		case "0":
			state.Config["selected_index"] = "1"
		case "1":
			state.Config["selected_index"] = "2"
		case "2":
			state.Config["selected_index"] = "3"
		case "3":
			state.Config["selected_index"] = "4"
		case "4":
			state.Config["selected_index"] = "5"
		case "9":
			state.Config["selected_index"] = "10"
		}
	}
	return nil
}

func (h *NavigationEventHandler) handleNavigateLeft(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	// 处理左导航，可能是切换标签页或返回上级菜单
	return nil
}

func (h *NavigationEventHandler) handleNavigateRight(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	// 处理右导航，可能是进入子菜单或切换标签页
	return nil
}

func (h *NavigationEventHandler) handlePageUp(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	// 页面向上逻辑
	// 实际实现需要在具体的视图中处理
	return nil
}

func (h *NavigationEventHandler) handlePageDown(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	// 页面向下逻辑
	// 实际实现需要在具体的视图中处理
	return nil
}

func (h *NavigationEventHandler) handleHome(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	// 跳转到开头逻辑
	state.Config["selected_index"] = "0"
	return nil
}

func (h *NavigationEventHandler) handleEnd(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	// 跳转到结尾逻辑
	// 实际实现需要在具体的视图中处理
	return nil
}

func (h *NavigationEventHandler) handleSelect(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	// 处理选择事件，具体行为取决于当前视图
	switch state.CurrentView {
	case "main":
		// 主菜单选择
		if menuItem, ok := event.Data["menu_item"].(string); ok {
			state.CurrentView = menuItem
		}
	case "playlist":
		// 播放列表选择
		if songIndex, ok := event.Data["song_index"].(int); ok {
			if state.Player != nil && state.Player.Queue != nil && songIndex < len(state.Player.Queue) {
				// 设置当前歌曲
				state.Player.CurrentSong = state.Player.Queue[songIndex]
				state.Player.Status = ui.PlayStatusPlaying
			}
		}
	}
	return nil
}

func (h *NavigationEventHandler) handleBack(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	// 返回上一个视图
	// 简单实现：返回主菜单
	state.CurrentView = "main"
	return nil
}

func (h *NavigationEventHandler) handleViewChange(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	if newView, ok := event.Data["view"].(string); ok {
		// 切换视图
		state.CurrentView = newView
	}
	return nil
}

// SearchEventHandler 搜索事件处理器
type SearchEventHandler struct {
	*BaseEventHandler
}

// NewSearchEventHandler 创建搜索事件处理器
func NewSearchEventHandler() *SearchEventHandler {
	return &SearchEventHandler{
		BaseEventHandler: NewBaseEventHandler("search", 80),
	}
}

// CanHandle 检查是否可以处理事件
func (h *SearchEventHandler) CanHandle(event *ui.UIEvent) bool {
	searchEvents := []string{
		"search", "search_type_change", "search_clear",
		"search_history", "search_suggestion",
	}

	for _, eventType := range searchEvents {
		if event.Type == eventType {
			return true
		}
	}
	return false
}

// Handle 处理搜索事件
func (h *SearchEventHandler) Handle(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	switch event.Type {
	case "search":
		return h.handleSearch(ctx, event, state)
	case "search_type_change":
		return h.handleSearchTypeChange(ctx, event, state)
	case "search_clear":
		return h.handleSearchClear(ctx, event, state)
	case "search_history":
		return h.handleSearchHistory(ctx, event, state)
	case "search_suggestion":
		return h.handleSearchSuggestion(ctx, event, state)
	default:
		return fmt.Errorf("unsupported search event: %s", event.Type)
	}
}

func (h *SearchEventHandler) handleSearch(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	if query, ok := event.Data["query"].(string); ok {
		// 更新搜索查询
		state.Config["search_query"] = query
		state.Config["is_searching"] = "true"
		// 添加到搜索历史
		state.Config["search_history"] = query
	}
	return nil
}

func (h *SearchEventHandler) handleSearchTypeChange(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	// 搜索类型切换逻辑
	// 实际实现需要在具体的搜索视图中处理
	return nil
}

func (h *SearchEventHandler) handleSearchClear(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	// 清除搜索逻辑
	// 实际实现需要在具体的搜索视图中处理
	return nil
}

func (h *SearchEventHandler) handleSearchHistory(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	// 搜索历史逻辑
	// 实际实现需要在具体的搜索视图中处理
	return nil
}

func (h *SearchEventHandler) handleSearchSuggestion(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	// 搜索建议逻辑
	// 实际实现需要在具体的搜索视图中处理
	return nil
}

// ThemeEventHandler 主题事件处理器
type ThemeEventHandler struct {
	*BaseEventHandler
}

// NewThemeEventHandler 创建主题事件处理器
func NewThemeEventHandler() *ThemeEventHandler {
	return &ThemeEventHandler{
		BaseEventHandler: NewBaseEventHandler("theme", 70),
	}
}

// CanHandle 检查是否可以处理事件
func (h *ThemeEventHandler) CanHandle(event *ui.UIEvent) bool {
	themeEvents := []string{
		"theme_change", "theme_toggle", "theme_reset",
	}

	for _, eventType := range themeEvents {
		if event.Type == eventType {
			return true
		}
	}
	return false
}

// Handle 处理主题事件
func (h *ThemeEventHandler) Handle(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	switch event.Type {
	case "theme_change":
		return h.handleThemeChange(ctx, event, state)
	case "theme_toggle":
		return h.handleThemeToggle(ctx, event, state)
	case "theme_reset":
		return h.handleThemeReset(ctx, event, state)
	default:
		return fmt.Errorf("unsupported theme event: %s", event.Type)
	}
}

func (h *ThemeEventHandler) handleThemeChange(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	// 主题切换逻辑
	if theme, ok := event.Data["theme"].(string); ok {
		state.Config["theme"] = theme
	}
	return nil
}

func (h *ThemeEventHandler) handleThemeToggle(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	// 主题切换逻辑
	if currentTheme, ok := state.Config["theme"]; ok {
		if currentTheme == "default" {
			state.Config["theme"] = "dark"
		} else {
			state.Config["theme"] = "light"
		}
	} else {
		state.Config["theme"] = "dark"
	}
	return nil
}

func (h *ThemeEventHandler) handleThemeReset(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	// 主题重置逻辑
	// 实际实现需要在具体的UI管理器中处理
	return nil
}

// EventDispatcher 事件分发器
type EventDispatcher struct {
	handlers []EventHandler
}

// NewEventDispatcher 创建事件分发器
func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		handlers: []EventHandler{},
	}
}

// RegisterHandler 注册事件处理器
func (d *EventDispatcher) RegisterHandler(handler EventHandler) {
	d.handlers = append(d.handlers, handler)
	
	// 按优先级排序（高优先级在前）
	for i := len(d.handlers) - 1; i > 0; i-- {
		if d.handlers[i].GetPriority() > d.handlers[i-1].GetPriority() {
			d.handlers[i], d.handlers[i-1] = d.handlers[i-1], d.handlers[i]
		} else {
			break
		}
	}
}

// DispatchEvent 分发事件
func (d *EventDispatcher) DispatchEvent(ctx context.Context, event *ui.UIEvent, state *ui.AppState) error {
	for _, handler := range d.handlers {
		if handler.CanHandle(event) {
			if err := handler.Handle(ctx, event, state); err != nil {
				return fmt.Errorf("handler %T failed: %w", handler, err)
			}
			return nil // 只有第一个匹配的处理器处理事件
		}
	}
	return fmt.Errorf("no handler found for event type: %s", event.Type)
}

// GetHandlers 获取所有处理器
func (d *EventDispatcher) GetHandlers() []EventHandler {
	return d.handlers
}

// RemoveHandler 移除事件处理器
func (d *EventDispatcher) RemoveHandler(handler EventHandler) {
	for i, h := range d.handlers {
		if h == handler {
			d.handlers = append(d.handlers[:i], d.handlers[i+1:]...)
			break
		}
	}
}

// CreateDefaultEventDispatcher 创建默认事件分发器
func CreateDefaultEventDispatcher() *EventDispatcher {
	dispatcher := NewEventDispatcher()
	
	// 注册默认事件处理器
	dispatcher.RegisterHandler(NewPlaybackEventHandler())
	dispatcher.RegisterHandler(NewNavigationEventHandler())
	dispatcher.RegisterHandler(NewSearchEventHandler())
	dispatcher.RegisterHandler(NewThemeEventHandler())
	
	return dispatcher
}