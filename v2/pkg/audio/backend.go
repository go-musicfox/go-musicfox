package audio

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// PlaybackState 播放状态枚举
type PlaybackState int

const (
	StateStopped PlaybackState = iota
	StatePlaying
	StatePaused
	StateBuffering
	StateError
)

// String 返回播放状态的字符串表示
func (s PlaybackState) String() string {
	switch s {
	case StateStopped:
		return "stopped"
	case StatePlaying:
		return "playing"
	case StatePaused:
		return "paused"
	case StateBuffering:
		return "buffering"
	case StateError:
		return "error"
	default:
		return "unknown"
	}
}

// BackendCapabilities 播放器后端能力描述
type BackendCapabilities struct {
	SupportedFormats []string          `json:"supported_formats"` // 支持的音频格式
	SupportedPlatforms []string        `json:"supported_platforms"` // 支持的平台
	Features         map[string]bool   `json:"features"`          // 支持的功能特性
	MaxVolume        float64           `json:"max_volume"`        // 最大音量
	MinVolume        float64           `json:"min_volume"`        // 最小音量
	SeekSupport      bool              `json:"seek_support"`      // 是否支持跳转
	StreamingSupport bool              `json:"streaming_support"` // 是否支持流媒体
	Metadata         map[string]string `json:"metadata"`          // 额外元数据
}

// BackendConfig 播放器后端配置
type BackendConfig struct {
	Name         string                 `json:"name"`          // 后端名称
	Enabled      bool                   `json:"enabled"`       // 是否启用
	Priority     int                    `json:"priority"`      // 优先级
	Settings     map[string]interface{} `json:"settings"`      // 后端特定设置
	BufferSize   int                    `json:"buffer_size"`   // 缓冲区大小
	SampleRate   int                    `json:"sample_rate"`   // 采样率
	Channels     int                    `json:"channels"`      // 声道数
	DefaultVolume float64               `json:"default_volume"` // 默认音量
}

// EventType 事件类型
type EventType string

const (
	EventStateChanged    EventType = "state_changed"
	EventPositionChanged EventType = "position_changed"
	EventVolumeChanged   EventType = "volume_changed"
	EventTrackChanged    EventType = "track_changed"
	EventError           EventType = "error"
	EventBuffering       EventType = "buffering"
)

// Event 播放器事件
type Event struct {
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Source    string                 `json:"source"`
}

// EventHandler 事件处理器
type EventHandler func(event *Event)

// PlayerBackend 播放器后端接口
type PlayerBackend interface {
	// 生命周期管理
	Initialize(ctx context.Context, config *BackendConfig) error
	Cleanup() error

	// 播放控制
	Play(url string) error
	Pause() error
	Resume() error
	Stop() error
	Seek(position time.Duration) error

	// 状态查询
	GetState() PlaybackState
	GetPosition() (time.Duration, error)
	GetDuration() (time.Duration, error)
	IsPlaying() bool

	// 音量控制
	SetVolume(volume float64) error
	GetVolume() (float64, error)

	// 后端信息
	GetName() string
	GetVersion() string
	GetCapabilities() *BackendCapabilities
	IsAvailable() bool

	// 事件系统
	AddEventHandler(eventType EventType, handler EventHandler) error
	RemoveEventHandler(eventType EventType, handler EventHandler) error
	EmitEvent(event *Event)

	// 健康检查
	HealthCheck() error
}

// BaseBackend 基础播放器后端实现
type BaseBackend struct {
	name         string
	version      string
	capabilities *BackendCapabilities
	config       *BackendConfig
	state        PlaybackState
	volume       float64
	position     time.Duration
	duration     time.Duration
	eventHandlers map[EventType][]EventHandler
	mutex        sync.RWMutex
}

// NewBaseBackend 创建基础播放器后端
func NewBaseBackend(name, version string, capabilities *BackendCapabilities) *BaseBackend {
	return &BaseBackend{
		name:          name,
		version:       version,
		capabilities:  capabilities,
		state:         StateStopped,
		volume:        0.8, // 默认音量 80%
		eventHandlers: make(map[EventType][]EventHandler),
	}
}

// Initialize 初始化后端
func (b *BaseBackend) Initialize(ctx context.Context, config *BackendConfig) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	
	b.config = config
	if config.DefaultVolume > 0 {
		b.volume = config.DefaultVolume
	}
	
	return nil
}

// Cleanup 清理资源
func (b *BaseBackend) Cleanup() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	
	b.eventHandlers = make(map[EventType][]EventHandler)
	b.state = StateStopped
	
	return nil
}

// GetName 获取后端名称
func (b *BaseBackend) GetName() string {
	return b.name
}

// GetVersion 获取后端版本
func (b *BaseBackend) GetVersion() string {
	return b.version
}

// GetCapabilities 获取后端能力
func (b *BaseBackend) GetCapabilities() *BackendCapabilities {
	return b.capabilities
}

// GetState 获取播放状态
func (b *BaseBackend) GetState() PlaybackState {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.state
}

// SetState 设置播放状态（内部使用）
func (b *BaseBackend) SetState(state PlaybackState) {
	b.mutex.Lock()
	oldState := b.state
	b.state = state
	b.mutex.Unlock()
	
	if oldState != state {
		b.EmitEvent(&Event{
			Type:      EventStateChanged,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"old_state": oldState.String(),
				"new_state": state.String(),
			},
			Source: b.name,
		})
	}
}

// GetVolume 获取音量
func (b *BaseBackend) GetVolume() (float64, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.volume, nil
}

// SetVolume 设置音量
func (b *BaseBackend) SetVolume(volume float64) error {
	if volume < 0 || volume > 1 {
		return fmt.Errorf("volume must be between 0 and 1")
	}
	
	b.mutex.Lock()
	oldVolume := b.volume
	b.volume = volume
	b.mutex.Unlock()
	
	if oldVolume != volume {
		b.EmitEvent(&Event{
			Type:      EventVolumeChanged,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"old_volume": oldVolume,
				"new_volume": volume,
			},
			Source: b.name,
		})
	}
	
	return nil
}

// GetPosition 获取播放位置
func (b *BaseBackend) GetPosition() (time.Duration, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.position, nil
}

// SetPosition 设置播放位置（内部使用）
func (b *BaseBackend) SetPosition(position time.Duration) {
	b.mutex.Lock()
	b.position = position
	b.mutex.Unlock()
	
	b.EmitEvent(&Event{
		Type:      EventPositionChanged,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"position": position.Seconds(),
		},
		Source: b.name,
	})
}

// GetDuration 获取音乐时长
func (b *BaseBackend) GetDuration() (time.Duration, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.duration, nil
}

// SetDuration 设置音乐时长（内部使用）
func (b *BaseBackend) SetDuration(duration time.Duration) {
	b.mutex.Lock()
	b.duration = duration
	b.mutex.Unlock()
}

// IsPlaying 检查是否正在播放
func (b *BaseBackend) IsPlaying() bool {
	return b.GetState() == StatePlaying
}

// AddEventHandler 添加事件处理器
func (b *BaseBackend) AddEventHandler(eventType EventType, handler EventHandler) error {
	if handler == nil {
		return fmt.Errorf("event handler cannot be nil")
	}
	
	b.mutex.Lock()
	defer b.mutex.Unlock()
	
	b.eventHandlers[eventType] = append(b.eventHandlers[eventType], handler)
	return nil
}

// RemoveEventHandler 移除事件处理器
func (b *BaseBackend) RemoveEventHandler(eventType EventType, handler EventHandler) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	
	handlers := b.eventHandlers[eventType]
	for i, h := range handlers {
		// 由于函数比较困难，这里简单移除第一个匹配的处理器
		if &h == &handler {
			b.eventHandlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			return nil
		}
	}
	
	return fmt.Errorf("event handler not found")
}

// EmitEvent 发送事件
func (b *BaseBackend) EmitEvent(event *Event) {
	b.mutex.RLock()
	handlers := b.eventHandlers[event.Type]
	b.mutex.RUnlock()
	
	for _, handler := range handlers {
		go handler(event) // 异步处理事件
	}
}

// HealthCheck 健康检查
func (b *BaseBackend) HealthCheck() error {
	// 基础实现，子类可以重写
	return nil
}

// 默认实现的播放控制方法（子类需要重写）

// Play 播放音频（需要子类实现）
func (b *BaseBackend) Play(url string) error {
	return fmt.Errorf("Play method not implemented")
}

// Pause 暂停播放（需要子类实现）
func (b *BaseBackend) Pause() error {
	return fmt.Errorf("Pause method not implemented")
}

// Resume 恢复播放（需要子类实现）
func (b *BaseBackend) Resume() error {
	return fmt.Errorf("Resume method not implemented")
}

// Stop 停止播放（需要子类实现）
func (b *BaseBackend) Stop() error {
	return fmt.Errorf("Stop method not implemented")
}

// Seek 跳转到指定位置（需要子类实现）
func (b *BaseBackend) Seek(position time.Duration) error {
	return fmt.Errorf("Seek method not implemented")
}

// IsAvailable 检查后端是否可用（需要子类实现）
func (b *BaseBackend) IsAvailable() bool {
	return false
}