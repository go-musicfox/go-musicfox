package audio

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// PlayerBackend 播放器后端接口
type PlayerBackend interface {
	// 生命周期管理
	Initialize(ctx context.Context) error
	Cleanup() error

	// 播放控制
	Play(url string) error
	Pause() error
	Resume() error
	Stop() error
	Seek(position time.Duration) error

	// 状态查询
	GetPosition() (time.Duration, error)
	GetDuration() (time.Duration, error)
	IsPlaying() bool

	// 音量控制
	SetVolume(volume float64) error
	GetVolume() (float64, error)

	// 后端信息
	GetName() string
	GetVersion() string
	GetSupportedFormats() []string
	IsAvailable() bool
}

// PlayerFactory 播放器工厂
type PlayerFactory struct {
	backends map[string]PlayerBackendCreator
	mutex    sync.RWMutex
}

// PlayerBackendCreator 播放器后端创建函数
type PlayerBackendCreator func(config map[string]interface{}) (PlayerBackend, error)

// NewPlayerFactory 创建播放器工厂
func NewPlayerFactory() *PlayerFactory {
	factory := &PlayerFactory{
		backends: make(map[string]PlayerBackendCreator),
	}

	// 注册内置播放器后端
	factory.registerBuiltinBackends()

	return factory
}

// registerBuiltinBackends 注册内置播放器后端
func (f *PlayerFactory) registerBuiltinBackends() {
	// 注册 Beep 后端（跨平台）
	f.RegisterBackend("beep", func(config map[string]interface{}) (PlayerBackend, error) {
		return NewBeepPlayer(config), nil
	})

	// 根据平台注册特定后端
	switch runtime.GOOS {
	case "darwin":
		// macOS 原生后端
		f.RegisterBackend("osx", func(config map[string]interface{}) (PlayerBackend, error) {
			return newOSXPlayer(config)
		})
	case "windows":
		// Windows 后端
		f.RegisterBackend("windows", func(config map[string]interface{}) (PlayerBackend, error) {
			return newWindowsPlayer(config)
		})
	}

	// MPD 后端
	f.RegisterBackend("mpd", func(config map[string]interface{}) (PlayerBackend, error) {
		return NewMPDPlayer(config), nil
	})

	// MPV 后端（跨平台）
	f.RegisterBackend("mpv", func(config map[string]interface{}) (PlayerBackend, error) {
		return NewMPVPlayer(config), nil
	})
}

// RegisterBackend 注册播放器后端
func (f *PlayerFactory) RegisterBackend(name string, creator PlayerBackendCreator) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if name == "" {
		return fmt.Errorf("backend name cannot be empty")
	}

	if creator == nil {
		return fmt.Errorf("backend creator cannot be nil")
	}

	f.backends[name] = creator
	return nil
}

// CreatePlayer 创建播放器后端
func (f *PlayerFactory) CreatePlayer(name string, config map[string]interface{}) (PlayerBackend, error) {
	f.mutex.RLock()
	creator, exists := f.backends[name]
	f.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("player backend '%s' not found", name)
	}

	player, err := creator(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create player backend '%s': %w", name, err)
	}

	// 检查后端是否可用
	if !player.IsAvailable() {
		return nil, fmt.Errorf("player backend '%s' is not available on this system", name)
	}

	return player, nil
}

// GetAvailableBackends 获取可用的播放器后端
func (f *PlayerFactory) GetAvailableBackends() []string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	var available []string
	for name, creator := range f.backends {
		// 创建临时实例检查可用性
		if player, err := creator(nil); err == nil && player.IsAvailable() {
			available = append(available, name)
		}
	}

	return available
}

// GetBackendInfo 获取后端信息
func (f *PlayerFactory) GetBackendInfo(name string) (map[string]interface{}, error) {
	f.mutex.RLock()
	creator, exists := f.backends[name]
	f.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("backend '%s' not found", name)
	}

	player, err := creator(nil)
	if err != nil {
		return nil, err
	}

	info := map[string]interface{}{
		"name":              player.GetName(),
		"version":           player.GetVersion(),
		"supported_formats": player.GetSupportedFormats(),
		"available":         player.IsAvailable(),
	}

	return info, nil
}

// BasePlayer 基础播放器实现
type BasePlayer struct {
	name             string
	version          string
	supportedFormats []string
	volume           float64
	position         time.Duration
	duration         time.Duration
	playing          bool
	mutex            sync.RWMutex
}

// NewBasePlayer 创建基础播放器
func NewBasePlayer() *BasePlayer {
	return &BasePlayer{
		name:             "Base Player",
		version:          "1.0.0",
		supportedFormats: []string{},
		volume:           0.8, // 默认音量 80%
	}
}

// NewBasePlayerWithInfo 创建带信息的基础播放器
func NewBasePlayerWithInfo(name, version string, formats []string) *BasePlayer {
	return &BasePlayer{
		name:             name,
		version:          version,
		supportedFormats: formats,
		volume:           0.8, // 默认音量 80%
	}
}

// GetName 获取播放器名称
func (p *BasePlayer) GetName() string {
	return p.name
}

// GetVersion 获取播放器版本
func (p *BasePlayer) GetVersion() string {
	return p.version
}

// GetSupportedFormats 获取支持的格式
func (p *BasePlayer) GetSupportedFormats() []string {
	return p.supportedFormats
}

// SetVolume 设置音量
func (p *BasePlayer) SetVolume(volume float64) error {
	if volume < 0 || volume > 1 {
		return fmt.Errorf("volume must be between 0 and 1")
	}

	p.mutex.Lock()
	p.volume = volume
	p.mutex.Unlock()

	return nil
}

// GetVolume 获取音量
func (p *BasePlayer) GetVolume() (float64, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.volume, nil
}

// GetPosition 获取播放位置
func (p *BasePlayer) GetPosition() (time.Duration, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.position, nil
}

// GetDuration 获取音乐时长
func (p *BasePlayer) GetDuration() (time.Duration, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.duration, nil
}

// IsPlaying 检查是否正在播放
func (p *BasePlayer) IsPlaying() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.playing
}

// setPlaying 设置播放状态（内部使用）
func (p *BasePlayer) setPlaying(playing bool) {
	p.mutex.Lock()
	p.playing = playing
	p.mutex.Unlock()
}

// setPosition 设置播放位置（内部使用）
func (p *BasePlayer) setPosition(position time.Duration) {
	p.mutex.Lock()
	p.position = position
	p.mutex.Unlock()
}

// setDuration 设置音乐时长（内部使用）
func (p *BasePlayer) setDuration(duration time.Duration) {
	p.mutex.Lock()
	p.duration = duration
	p.mutex.Unlock()
}
