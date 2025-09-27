package audio

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"time"
)

// BackendCreator 播放器后端创建函数
type BackendCreator func(config *BackendConfig) (PlayerBackend, error)

// BackendInfo 播放器后端信息
type BackendInfo struct {
	Name         string                `json:"name"`
	Version      string                `json:"version"`
	Description  string                `json:"description"`
	Capabilities *BackendCapabilities  `json:"capabilities"`
	Platforms    []string              `json:"platforms"`
	Priority     int                   `json:"priority"`
	Available    bool                  `json:"available"`
	Creator      BackendCreator        `json:"-"`
}

// PlayerFactory 播放器工厂
type PlayerFactory struct {
	backends      map[string]*BackendInfo
	currentBackend string
	configWatcher *ConfigWatcher
	mutex         sync.RWMutex
	eventHandlers map[string][]func(string, string) // 后端切换事件处理器
}

// ConfigWatcher 配置监听器
type ConfigWatcher struct {
	callbacks []func(*BackendConfig)
	mutex     sync.RWMutex
}

// NewConfigWatcher 创建配置监听器
func NewConfigWatcher() *ConfigWatcher {
	return &ConfigWatcher{
		callbacks: make([]func(*BackendConfig), 0),
	}
}

// AddCallback 添加配置变化回调
func (cw *ConfigWatcher) AddCallback(callback func(*BackendConfig)) {
	cw.mutex.Lock()
	defer cw.mutex.Unlock()
	cw.callbacks = append(cw.callbacks, callback)
}

// NotifyConfigChange 通知配置变化
func (cw *ConfigWatcher) NotifyConfigChange(config *BackendConfig) {
	cw.mutex.RLock()
	defer cw.mutex.RUnlock()
	
	for _, callback := range cw.callbacks {
		go callback(config) // 异步通知
	}
}

// NewPlayerFactory 创建播放器工厂
func NewPlayerFactory() *PlayerFactory {
	factory := &PlayerFactory{
		backends:      make(map[string]*BackendInfo),
		configWatcher: NewConfigWatcher(),
		eventHandlers: make(map[string][]func(string, string)),
	}
	
	// 注册内置播放器后端
	factory.registerBuiltinBackends()
	
	// 设置配置热更新监听
	factory.setupConfigWatcher()
	
	return factory
}

// registerBuiltinBackends 注册内置播放器后端
func (f *PlayerFactory) registerBuiltinBackends() {
	// 注册 Beep 后端（跨平台）
	f.RegisterBackend(&BackendInfo{
		Name:        "beep",
		Version:     "1.0.0",
		Description: "Go native audio library with cross-platform support",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3", "wav", "flac", "ogg"},
			SupportedPlatforms: []string{"linux", "darwin", "windows"},
			Features: map[string]bool{
				"seek":      true,
				"streaming": true,
				"volume":    true,
			},
			MaxVolume:        1.0,
			MinVolume:        0.0,
			SeekSupport:      true,
			StreamingSupport: true,
		},
		Platforms: []string{"linux", "darwin", "windows"},
		Priority:  5,
		Creator:   createBeepBackend,
	})
	
	// 根据平台注册特定后端
	switch runtime.GOOS {
	case "darwin":
		// macOS 原生后端
		f.RegisterBackend(&BackendInfo{
			Name:        "osx",
			Version:     "1.0.0",
			Description: "macOS native AVAudioPlayer backend",
			Capabilities: &BackendCapabilities{
				SupportedFormats:   []string{"mp3", "wav", "m4a", "aac", "flac", "ogg"},
				SupportedPlatforms: []string{"darwin"},
				Features: map[string]bool{
					"seek":      true,
					"streaming": true,
					"volume":    true,
					"native":    true,
				},
				MaxVolume:        1.0,
				MinVolume:        0.0,
				SeekSupport:      true,
				StreamingSupport: true,
			},
			Platforms: []string{"darwin"},
			Priority:  10,
			Creator:   createOSXBackend,
		})
	case "windows":
		// Windows 后端
		f.RegisterBackend(&BackendInfo{
			Name:        "windows",
			Version:     "1.0.0",
			Description: "Windows Media Player API backend",
			Capabilities: &BackendCapabilities{
				SupportedFormats:   []string{"mp3", "wav", "wma", "m4a", "aac"},
				SupportedPlatforms: []string{"windows"},
				Features: map[string]bool{
					"seek":      true,
					"streaming": true,
					"volume":    true,
					"native":    true,
				},
				MaxVolume:        1.0,
				MinVolume:        0.0,
				SeekSupport:      true,
				StreamingSupport: true,
			},
			Platforms: []string{"windows"},
			Priority:  10,
			Creator:   createWindowsBackend,
		})
	}
	
	// MPD 后端（Linux/Unix）
	f.RegisterBackend(&BackendInfo{
		Name:        "mpd",
		Version:     "1.0.0",
		Description: "Music Player Daemon backend",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3", "wav", "flac", "ogg", "m4a", "aac"},
			SupportedPlatforms: []string{"linux", "unix"},
			Features: map[string]bool{
				"seek":      true,
				"streaming": true,
				"volume":    true,
				"remote":    true,
			},
			MaxVolume:        1.0,
			MinVolume:        0.0,
			SeekSupport:      true,
			StreamingSupport: true,
		},
		Platforms: []string{"linux", "unix"},
		Priority:  8,
		Creator:   createMPDBackend,
	})
	
	// MPV 后端（跨平台）
	f.RegisterBackend(&BackendInfo{
		Name:        "mpv",
		Version:     "1.0.0",
		Description: "MPV media player backend with extensive format support",
		Capabilities: &BackendCapabilities{
			SupportedFormats: []string{
				"mp3", "wav", "flac", "ogg", "m4a", "aac", "wma", "ape", "opus",
				"mp4", "mkv", "avi", "webm", // 视频格式也支持音频
			},
			SupportedPlatforms: []string{"linux", "darwin", "windows"},
			Features: map[string]bool{
				"seek":      true,
				"streaming": true,
				"volume":    true,
				"advanced":  true,
			},
			MaxVolume:        1.0,
			MinVolume:        0.0,
			SeekSupport:      true,
			StreamingSupport: true,
		},
		Platforms: []string{"linux", "darwin", "windows"},
		Priority:  7,
		Creator:   createMPVBackend,
	})
}

// RegisterBackend 注册播放器后端
func (f *PlayerFactory) RegisterBackend(info *BackendInfo) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	
	if info.Name == "" {
		return fmt.Errorf("backend name cannot be empty")
	}
	
	if info.Creator == nil {
		return fmt.Errorf("backend creator cannot be nil")
	}
	
	// 检查后端是否可用
	info.Available = f.checkBackendAvailability(info)
	
	f.backends[info.Name] = info
	return nil
}

// checkBackendAvailability 检查后端可用性
func (f *PlayerFactory) checkBackendAvailability(info *BackendInfo) bool {
	// 检查平台兼容性
	currentPlatform := runtime.GOOS
	platformSupported := false
	for _, platform := range info.Platforms {
		if platform == currentPlatform || platform == "test" {
			platformSupported = true
			break
		}
	}
	
	if !platformSupported {
		return false
	}
	
	// 尝试创建后端实例进行可用性检查
	backend, err := info.Creator(&BackendConfig{
		Name:    info.Name,
		Enabled: true,
	})
	
	if err != nil {
		return false
	}
	
	defer backend.Cleanup()
	return backend.IsAvailable()
}

// CreateBackend 创建播放器后端
func (f *PlayerFactory) CreateBackend(name string, config *BackendConfig) (PlayerBackend, error) {
	f.mutex.RLock()
	info, exists := f.backends[name]
	f.mutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("backend '%s' not found", name)
	}
	
	if !info.Available {
		return nil, fmt.Errorf("backend '%s' is not available on this system", name)
	}
	
	backend, err := info.Creator(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create backend '%s': %w", name, err)
	}
	
	// 初始化后端
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := backend.Initialize(ctx, config); err != nil {
		backend.Cleanup()
		return nil, fmt.Errorf("failed to initialize backend '%s': %w", name, err)
	}
	
	return backend, nil
}

// GetAvailableBackends 获取可用的播放器后端
func (f *PlayerFactory) GetAvailableBackends() []string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	
	var available []string
	for name, info := range f.backends {
		if info.Available {
			available = append(available, name)
		}
	}
	
	// 按优先级排序
	sort.Slice(available, func(i, j int) bool {
		return f.backends[available[i]].Priority > f.backends[available[j]].Priority
	})
	
	return available
}

// GetBackendInfo 获取后端信息
func (f *PlayerFactory) GetBackendInfo(name string) (*BackendInfo, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	
	info, exists := f.backends[name]
	if !exists {
		return nil, fmt.Errorf("backend '%s' not found", name)
	}
	
	// 返回副本以避免外部修改
	infoCopy := *info
	infoCopy.Creator = nil // 不暴露创建函数
	return &infoCopy, nil
}

// GetAllBackends 获取所有后端信息
func (f *PlayerFactory) GetAllBackends() map[string]*BackendInfo {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	
	result := make(map[string]*BackendInfo)
	for name, info := range f.backends {
		infoCopy := *info
		infoCopy.Creator = nil // 不暴露创建函数
		result[name] = &infoCopy
	}
	
	return result
}

// GetBestBackend 获取最佳可用后端
func (f *PlayerFactory) GetBestBackend() (string, error) {
	available := f.GetAvailableBackends()
	if len(available) == 0 {
		return "", fmt.Errorf("no available backends")
	}
	
	return available[0], nil // 已按优先级排序
}

// SwitchBackend 切换播放器后端
func (f *PlayerFactory) SwitchBackend(from, to string) error {
	f.mutex.Lock()
	oldBackend := f.currentBackend
	f.currentBackend = to
	f.mutex.Unlock()
	
	// 触发后端切换事件
	f.emitBackendSwitchEvent(oldBackend, to)
	
	return nil
}

// GetCurrentBackend 获取当前后端
func (f *PlayerFactory) GetCurrentBackend() string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.currentBackend
}

// AddBackendSwitchHandler 添加后端切换事件处理器
func (f *PlayerFactory) AddBackendSwitchHandler(handler func(string, string)) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	
	f.eventHandlers["switch"] = append(f.eventHandlers["switch"], handler)
}

// emitBackendSwitchEvent 发送后端切换事件
func (f *PlayerFactory) emitBackendSwitchEvent(from, to string) {
	f.mutex.RLock()
	handlers := f.eventHandlers["switch"]
	f.mutex.RUnlock()
	
	for _, handler := range handlers {
		go handler(from, to) // 异步处理
	}
}

// setupConfigWatcher 设置配置监听器
func (f *PlayerFactory) setupConfigWatcher() {
	f.configWatcher.AddCallback(func(config *BackendConfig) {
		// 处理配置热更新
		f.handleConfigUpdate(config)
	})
}

// handleConfigUpdate 处理配置更新
func (f *PlayerFactory) handleConfigUpdate(config *BackendConfig) {
	// 如果配置指定了新的后端，尝试切换
	if config.Name != "" && config.Name != f.GetCurrentBackend() {
		if err := f.SwitchBackend(f.GetCurrentBackend(), config.Name); err == nil {
			// 切换成功，可以记录日志或发送通知
		}
	}
}

// RefreshAvailability 刷新所有后端的可用性
func (f *PlayerFactory) RefreshAvailability() {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	
	for _, info := range f.backends {
		info.Available = f.checkBackendAvailability(info)
	}
}

// 后端创建函数（需要在具体实现中定义）
func createBeepBackend(config *BackendConfig) (PlayerBackend, error) {
	// 这里应该返回实际的 Beep 后端实现
	return nil, fmt.Errorf("beep backend not implemented")
}

func createOSXBackend(config *BackendConfig) (PlayerBackend, error) {
	// 这里应该返回实际的 OSX 后端实现
	return nil, fmt.Errorf("osx backend not implemented")
}

func createWindowsBackend(config *BackendConfig) (PlayerBackend, error) {
	// 这里应该返回实际的 Windows 后端实现
	return nil, fmt.Errorf("windows backend not implemented")
}

func createMPDBackend(config *BackendConfig) (PlayerBackend, error) {
	// 这里应该返回实际的 MPD 后端实现
	return nil, fmt.Errorf("mpd backend not implemented")
}

func createMPVBackend(config *BackendConfig) (PlayerBackend, error) {
	// 这里应该返回实际的 MPV 后端实现
	return nil, fmt.Errorf("mpv backend not implemented")
}