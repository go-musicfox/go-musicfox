package event

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// EventType 事件类型
type EventType string

// 预定义事件类型
const (
	// 特殊事件类型
	EventTypeAll EventType = "*" // 匹配所有事件类型

	// 插件相关事件
	EventPluginLoaded     EventType = "plugin.loaded"
	EventPluginStarted    EventType = "plugin.started"
	EventPluginStopped    EventType = "plugin.stopped"
	EventPluginUnloaded   EventType = "plugin.unloaded"
	EventPluginError      EventType = "plugin.error"
	EventPluginCrashed    EventType = "plugin.crashed"
	EventPluginRestarted  EventType = "plugin.restarted"
	EventPluginHealthCheck EventType = "plugin.health_check"

	// 播放器相关事件
	EventPlayerPlay           EventType = "player.play"
	EventPlayerPause          EventType = "player.pause"
	EventPlayerStop           EventType = "player.stop"
	EventPlayerNext           EventType = "player.next"
	EventPlayerPrevious       EventType = "player.previous"
	EventPlayerSeek           EventType = "player.seek"
	EventPlayerVolumeChanged  EventType = "player.volume_changed"
	EventPlayerStateChanged   EventType = "player.state_changed"
	EventPlayerSongChanged    EventType = "player.song_changed"
	EventPlayerPositionChanged EventType = "player.position_changed"
	EventPlayerModeChanged    EventType = "player.mode_changed"
	EventPlayerBuffering      EventType = "player.buffering"
	EventPlayerError          EventType = "player.error"

	// 播放列表相关事件
	EventPlaylistCreated  EventType = "playlist.created"
	EventPlaylistUpdated  EventType = "playlist.updated"
	EventPlaylistDeleted  EventType = "playlist.deleted"
	EventPlaylistCleared  EventType = "playlist.cleared"
	EventPlaylistSongAdded EventType = "playlist.song_added"
	EventPlaylistSongRemoved EventType = "playlist.song_removed"
	EventPlaylistSongMoved EventType = "playlist.song_moved"
	EventPlaylistLoaded   EventType = "playlist.loaded"

	// 音乐源相关事件
	EventMusicSourceConnected    EventType = "music_source.connected"
	EventMusicSourceDisconnected EventType = "music_source.disconnected"
	EventMusicSourceError        EventType = "music_source.error"
	EventMusicSourceSearchResult EventType = "music_source.search_result"
	EventMusicSourceCacheUpdated EventType = "music_source.cache_updated"

	// 用户相关事件
	EventUserLogin        EventType = "user.login"
	EventUserLogout       EventType = "user.logout"
	EventUserProfileUpdated EventType = "user.profile_updated"
	EventUserPreferencesChanged EventType = "user.preferences_changed"

	// 系统相关事件
	EventSystemStartup     EventType = "system.startup"
	EventSystemShutdown    EventType = "system.shutdown"
	EventSystemError       EventType = "system.error"
	EventSystemHealthCheck EventType = "system.health_check"
	EventConfigChanged     EventType = "config.changed"
	EventConfigReloaded    EventType = "config.reloaded"

	// 存储相关事件
	EventStorageRead      EventType = "storage.read"
	EventStorageWrite     EventType = "storage.write"
	EventStorageDelete    EventType = "storage.delete"
	EventStorageError     EventType = "storage.error"
	EventStorageCacheHit  EventType = "storage.cache_hit"
	EventStorageCacheMiss EventType = "storage.cache_miss"

	// UI相关事件
	EventUIRender     EventType = "ui.render"
	EventUIUpdate     EventType = "ui.update"
	EventUIKeyPress   EventType = "ui.key_press"
	EventUIMouseClick EventType = "ui.mouse_click"
	EventUIResize     EventType = "ui.resize"
	EventUIError      EventType = "ui.error"

	// 网络相关事件
	EventNetworkRequest  EventType = "network.request"
	EventNetworkResponse EventType = "network.response"
	EventNetworkTimeout  EventType = "network.timeout"
	EventNetworkError    EventType = "network.error"
)

// Event 事件接口
type Event interface {
	GetType() EventType
	GetData() interface{}
	GetTimestamp() time.Time
	GetSource() string
	GetID() string
}

// BaseEvent 基础事件实现
type BaseEvent struct {
	ID        string      `json:"id"`
	Type      EventType   `json:"type"`
	Data      interface{} `json:"data"`
	Source    string      `json:"source"`
	Timestamp time.Time   `json:"timestamp"`
}

func (e *BaseEvent) GetType() EventType {
	return e.Type
}

func (e *BaseEvent) GetData() interface{} {
	return e.Data
}

func (e *BaseEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e *BaseEvent) GetSource() string {
	return e.Source
}

func (e *BaseEvent) GetID() string {
	return e.ID
}

// EventHandler 事件处理器函数类型
type EventHandler func(ctx context.Context, event Event) error

// EventFilter 事件过滤器函数类型
type EventFilter func(event Event) bool

// Subscription 订阅信息
type Subscription struct {
	ID       string       `json:"id"`
	Type     EventType    `json:"type"`
	Handler  EventHandler `json:"-"`
	Filter   EventFilter  `json:"-"`
	Priority int          `json:"priority"`
	Async    bool         `json:"async"`
	CreatedAt time.Time   `json:"created_at"`
}

// EventPriority 事件优先级
type EventPriority int

const (
	PriorityLow    EventPriority = 0
	PriorityNormal EventPriority = 100
	PriorityHigh   EventPriority = 200
	PriorityCritical EventPriority = 300
)

// PriorityEvent 带优先级的事件
type PriorityEvent struct {
	*BaseEvent
	Priority EventPriority `json:"priority"`
}

// PluginEvent 插件事件
type PluginEvent struct {
	*BaseEvent
	PluginName string `json:"plugin_name"`
	PluginType string `json:"plugin_type"`
	Path       string `json:"path"`
	Error      string `json:"error,omitempty"`
}

// PlayerEvent 播放器事件
type PlayerEvent struct {
	*BaseEvent
	SongID   string        `json:"song_id,omitempty"`
	Position time.Duration `json:"position,omitempty"`
	Volume   float64       `json:"volume,omitempty"`
}

// SystemEvent 系统事件
type SystemEvent struct {
	*BaseEvent
	Message string `json:"message"`
	Level   string `json:"level"`
}

// SubscriberInfo 订阅者信息
type SubscriberInfo struct {
	ID          string        `json:"id"`
	EventType   EventType     `json:"event_type"`
	Handler     EventHandler  `json:"-"`
	Filter      EventFilter   `json:"-"`
	Priority    EventPriority `json:"priority"`
	Async       bool          `json:"async"`
	Active      bool          `json:"active"`
	Group       string        `json:"group"`
	CreatedAt   time.Time     `json:"created_at"`
	LastEventAt *time.Time    `json:"last_event_at,omitempty"`
	EventCount  int64         `json:"event_count"`
}

// EventTypeRegistry 事件类型注册表
type EventTypeRegistry struct {
	registeredTypes map[EventType]EventTypeInfo
	mu              sync.RWMutex
}

// EventTypeInfo 事件类型信息
type EventTypeInfo struct {
	Type        EventType `json:"type"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Schema      string    `json:"schema,omitempty"`
	Deprecated  bool      `json:"deprecated"`
	RegisteredAt time.Time `json:"registered_at"`
}

// EventValidator 事件验证器接口
type EventValidator interface {
	Validate(event Event) error
}

// DefaultEventValidator 默认事件验证器
type DefaultEventValidator struct {
	registry *EventTypeRegistry
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// 预定义事件类型信息
var predefinedEventTypes = map[EventType]EventTypeInfo{
	// 插件相关事件
	EventPluginLoaded:     {Type: EventPluginLoaded, Category: "plugin", Description: "Plugin loaded successfully"},
	EventPluginStarted:    {Type: EventPluginStarted, Category: "plugin", Description: "Plugin started successfully"},
	EventPluginStopped:    {Type: EventPluginStopped, Category: "plugin", Description: "Plugin stopped"},
	EventPluginUnloaded:   {Type: EventPluginUnloaded, Category: "plugin", Description: "Plugin unloaded"},
	EventPluginError:      {Type: EventPluginError, Category: "plugin", Description: "Plugin error occurred"},
	EventPluginCrashed:    {Type: EventPluginCrashed, Category: "plugin", Description: "Plugin crashed"},
	EventPluginRestarted:  {Type: EventPluginRestarted, Category: "plugin", Description: "Plugin restarted"},
	EventPluginHealthCheck: {Type: EventPluginHealthCheck, Category: "plugin", Description: "Plugin health check"},

	// 播放器相关事件
	EventPlayerPlay:           {Type: EventPlayerPlay, Category: "player", Description: "Player started playing"},
	EventPlayerPause:          {Type: EventPlayerPause, Category: "player", Description: "Player paused"},
	EventPlayerStop:           {Type: EventPlayerStop, Category: "player", Description: "Player stopped"},
	EventPlayerNext:           {Type: EventPlayerNext, Category: "player", Description: "Player moved to next track"},
	EventPlayerPrevious:       {Type: EventPlayerPrevious, Category: "player", Description: "Player moved to previous track"},
	EventPlayerSeek:           {Type: EventPlayerSeek, Category: "player", Description: "Player seeked to position"},
	EventPlayerVolumeChanged:  {Type: EventPlayerVolumeChanged, Category: "player", Description: "Player volume changed"},
	EventPlayerStateChanged:   {Type: EventPlayerStateChanged, Category: "player", Description: "Player state changed"},
	EventPlayerSongChanged:    {Type: EventPlayerSongChanged, Category: "player", Description: "Current song changed"},
	EventPlayerPositionChanged: {Type: EventPlayerPositionChanged, Category: "player", Description: "Player position changed"},
	EventPlayerModeChanged:    {Type: EventPlayerModeChanged, Category: "player", Description: "Player mode changed"},
	EventPlayerBuffering:      {Type: EventPlayerBuffering, Category: "player", Description: "Player is buffering"},
	EventPlayerError:          {Type: EventPlayerError, Category: "player", Description: "Player error occurred"},

	// 系统相关事件
	EventSystemStartup:     {Type: EventSystemStartup, Category: "system", Description: "System startup"},
	EventSystemShutdown:    {Type: EventSystemShutdown, Category: "system", Description: "System shutdown"},
	EventSystemError:       {Type: EventSystemError, Category: "system", Description: "System error occurred"},
	EventSystemHealthCheck: {Type: EventSystemHealthCheck, Category: "system", Description: "System health check"},
	EventConfigChanged:     {Type: EventConfigChanged, Category: "config", Description: "Configuration changed"},
	EventConfigReloaded:    {Type: EventConfigReloaded, Category: "config", Description: "Configuration reloaded"},
}

// NewEventTypeRegistry 创建事件类型注册表
func NewEventTypeRegistry() *EventTypeRegistry {
	registry := &EventTypeRegistry{
		registeredTypes: make(map[EventType]EventTypeInfo),
	}
	
	// 注册预定义事件类型
	for eventType, info := range predefinedEventTypes {
		info.RegisteredAt = time.Now()
		registry.registeredTypes[eventType] = info
	}
	
	return registry
}

// RegisterEventType 注册事件类型
func (r *EventTypeRegistry) RegisterEventType(eventType EventType, info EventTypeInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.registeredTypes[eventType]; exists {
		return fmt.Errorf("event type %s is already registered", eventType)
	}
	
	info.Type = eventType
	info.RegisteredAt = time.Now()
	r.registeredTypes[eventType] = info
	return nil
}

// UnregisterEventType 注销事件类型
func (r *EventTypeRegistry) UnregisterEventType(eventType EventType) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.registeredTypes[eventType]; !exists {
		return fmt.Errorf("event type %s is not registered", eventType)
	}
	
	delete(r.registeredTypes, eventType)
	return nil
}

// IsRegistered 检查事件类型是否已注册
func (r *EventTypeRegistry) IsRegistered(eventType EventType) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	_, exists := r.registeredTypes[eventType]
	return exists
}

// GetEventTypeInfo 获取事件类型信息
func (r *EventTypeRegistry) GetEventTypeInfo(eventType EventType) (EventTypeInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	info, exists := r.registeredTypes[eventType]
	return info, exists
}

// GetAllEventTypes 获取所有已注册的事件类型
func (r *EventTypeRegistry) GetAllEventTypes() []EventTypeInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	types := make([]EventTypeInfo, 0, len(r.registeredTypes))
	for _, info := range r.registeredTypes {
		types = append(types, info)
	}
	return types
}

// GetEventTypesByCategory 按类别获取事件类型
func (r *EventTypeRegistry) GetEventTypesByCategory(category string) []EventTypeInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var types []EventTypeInfo
	for _, info := range r.registeredTypes {
		if info.Category == category {
			types = append(types, info)
		}
	}
	return types
}

// NewDefaultEventValidator 创建默认事件验证器
func NewDefaultEventValidator(registry *EventTypeRegistry) *DefaultEventValidator {
	return &DefaultEventValidator{
		registry: registry,
	}
}

// Validate 验证事件
func (v *DefaultEventValidator) Validate(event Event) error {
	if event == nil {
		return &ValidationError{
			Field:   "event",
			Message: "event cannot be nil",
		}
	}
	
	// 验证事件类型
	eventType := event.GetType()
	if eventType == "" {
		return &ValidationError{
			Field:   "type",
			Message: "event type cannot be empty",
		}
	}
	
	// 检查事件类型是否已注册（除了通配符类型）
	if eventType != EventTypeAll && !v.registry.IsRegistered(eventType) {
		return &ValidationError{
			Field:   "type",
			Message: "event type is not registered",
			Value:   eventType,
		}
	}
	
	// 验证事件ID
	if event.GetID() == "" {
		return &ValidationError{
			Field:   "id",
			Message: "event ID cannot be empty",
		}
	}
	
	// 验证事件源
	if event.GetSource() == "" {
		return &ValidationError{
			Field:   "source",
			Message: "event source cannot be empty",
		}
	}
	
	// 验证时间戳
	if event.GetTimestamp().IsZero() {
		return &ValidationError{
			Field:   "timestamp",
			Message: "event timestamp cannot be zero",
		}
	}
	
	return nil
}

// IsValidEventType 检查事件类型是否有效
func IsValidEventType(eventType EventType) bool {
	return eventType != "" && (eventType == EventTypeAll || len(string(eventType)) > 0)
}

// GetEventCategory 从事件类型获取类别
func GetEventCategory(eventType EventType) string {
	if eventType == "" {
		return "unknown"
	}
	parts := strings.Split(string(eventType), ".")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return "unknown"
}