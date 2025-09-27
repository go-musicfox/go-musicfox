package audio

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// PlayerManager 播放器管理器
type PlayerManager struct {
	factory       *PlayerFactory
	currentPlayer PlayerBackend
	currentConfig *BackendConfig
	defaultConfig *BackendConfig
	mutex         sync.RWMutex
	eventHandlers map[EventType][]EventHandler
	shutdownCh    chan struct{}
	running       bool
}

// NewPlayerManager 创建播放器管理器
func NewPlayerManager() *PlayerManager {
	manager := &PlayerManager{
		factory:       NewPlayerFactory(),
		eventHandlers: make(map[EventType][]EventHandler),
		shutdownCh:    make(chan struct{}),
		defaultConfig: &BackendConfig{
			Enabled:       true,
			Priority:      5,
			BufferSize:    4096,
			SampleRate:    44100,
			Channels:      2,
			DefaultVolume: 0.8,
			Settings:      make(map[string]interface{}),
		},
	}
	
	// 设置后端切换事件处理
	manager.factory.AddBackendSwitchHandler(manager.handleBackendSwitch)
	
	return manager
}

// Initialize 初始化播放器管理器
func (pm *PlayerManager) Initialize(ctx context.Context) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	if pm.running {
		return fmt.Errorf("player manager already running")
	}
	
	// 刷新后端可用性
	pm.factory.RefreshAvailability()
	
	// 选择最佳可用后端
	bestBackend, err := pm.factory.GetBestBackend()
	if err != nil {
		return fmt.Errorf("no available audio backends: %w", err)
	}
	
	// 创建默认播放器
	config := *pm.defaultConfig
	config.Name = bestBackend
	
	player, err := pm.factory.CreateBackend(bestBackend, &config)
	if err != nil {
		return fmt.Errorf("failed to create default player: %w", err)
	}
	
	pm.currentPlayer = player
	pm.currentConfig = &config
	pm.factory.SwitchBackend("", bestBackend)
	pm.running = true
	
	// 设置播放器事件处理
	pm.setupPlayerEventHandlers(player)
	
	return nil
}

// Shutdown 关闭播放器管理器
func (pm *PlayerManager) Shutdown(ctx context.Context) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	if !pm.running {
		return nil
	}
	
	pm.running = false
	close(pm.shutdownCh)
	
	// 清理当前播放器
	if pm.currentPlayer != nil {
		pm.currentPlayer.Stop()
		pm.currentPlayer.Cleanup()
		pm.currentPlayer = nil
	}
	
	pm.currentConfig = nil
	return nil
}

// GetCurrentPlayer 获取当前播放器
func (pm *PlayerManager) GetCurrentPlayer() PlayerBackend {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	return pm.currentPlayer
}

// GetCurrentBackendName 获取当前后端名称
func (pm *PlayerManager) GetCurrentBackendName() string {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	if pm.currentConfig != nil {
		return pm.currentConfig.Name
	}
	return ""
}

// SwitchBackend 切换播放器后端
func (pm *PlayerManager) SwitchBackend(backendName string, config *BackendConfig) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	if !pm.running {
		return fmt.Errorf("player manager not running")
	}
	
	// 检查后端是否存在且可用
	info, err := pm.factory.GetBackendInfo(backendName)
	if err != nil {
		return fmt.Errorf("backend not found: %w", err)
	}
	
	if !info.Available {
		return fmt.Errorf("backend '%s' is not available", backendName)
	}
	
	// 如果已经是当前后端，只更新配置
	if pm.currentConfig != nil && pm.currentConfig.Name == backendName {
		return pm.updateCurrentBackendConfig(config)
	}
	
	// 保存当前播放状态
	var currentState *PlaybackState
	var currentURL string
	var currentPosition time.Duration
	
	if pm.currentPlayer != nil {
		state := pm.currentPlayer.GetState()
		currentState = &state
		currentPosition, _ = pm.currentPlayer.GetPosition()
		
		// 停止当前播放器
		pm.currentPlayer.Stop()
		pm.currentPlayer.Cleanup()
	}
	
	// 创建新的播放器后端
	newConfig := config
	if newConfig == nil {
		newConfig = pm.defaultConfig
	}
	newConfig.Name = backendName
	
	newPlayer, err := pm.factory.CreateBackend(backendName, newConfig)
	if err != nil {
		return fmt.Errorf("failed to create new backend: %w", err)
	}
	
	// 设置新播放器
	oldBackend := ""
	if pm.currentConfig != nil {
		oldBackend = pm.currentConfig.Name
	}
	
	pm.currentPlayer = newPlayer
	pm.currentConfig = newConfig
	
	// 设置事件处理器
	pm.setupPlayerEventHandlers(newPlayer)
	
	// 通知工厂切换后端
	pm.factory.SwitchBackend(oldBackend, backendName)
	
	// 尝试恢复播放状态
	if currentState != nil && *currentState == StatePlaying && currentURL != "" {
		go pm.restorePlaybackState(currentURL, currentPosition)
	}
	
	return nil
}

// updateCurrentBackendConfig 更新当前后端配置
func (pm *PlayerManager) updateCurrentBackendConfig(config *BackendConfig) error {
	if pm.currentPlayer == nil {
		return fmt.Errorf("no current player")
	}
	
	// 更新配置
	if config != nil {
		pm.currentConfig = config
		
		// 应用音量设置
		if config.DefaultVolume > 0 {
			pm.currentPlayer.SetVolume(config.DefaultVolume)
		}
	}
	
	return nil
}

// restorePlaybackState 恢复播放状态
func (pm *PlayerManager) restorePlaybackState(url string, position time.Duration) {
	if pm.currentPlayer == nil {
		return
	}
	
	// 播放音频
	if err := pm.currentPlayer.Play(url); err != nil {
		return
	}
	
	// 跳转到之前的位置
	if position > 0 {
		time.Sleep(100 * time.Millisecond) // 等待播放器准备就绪
		pm.currentPlayer.Seek(position)
	}
}

// GetAvailableBackends 获取可用后端列表
func (pm *PlayerManager) GetAvailableBackends() []string {
	return pm.factory.GetAvailableBackends()
}

// GetBackendInfo 获取后端信息
func (pm *PlayerManager) GetBackendInfo(name string) (*BackendInfo, error) {
	return pm.factory.GetBackendInfo(name)
}

// GetAllBackends 获取所有后端信息
func (pm *PlayerManager) GetAllBackends() map[string]*BackendInfo {
	return pm.factory.GetAllBackends()
}

// RefreshBackends 刷新后端可用性
func (pm *PlayerManager) RefreshBackends() {
	pm.factory.RefreshAvailability()
}

// AddEventHandler 添加事件处理器
func (pm *PlayerManager) AddEventHandler(eventType EventType, handler EventHandler) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	pm.eventHandlers[eventType] = append(pm.eventHandlers[eventType], handler)
	
	// 如果当前有播放器，也添加到播放器的事件处理器中
	if pm.currentPlayer != nil {
		pm.currentPlayer.AddEventHandler(eventType, handler)
	}
}

// RemoveEventHandler 移除事件处理器
func (pm *PlayerManager) RemoveEventHandler(eventType EventType, handler EventHandler) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	handlers := pm.eventHandlers[eventType]
	for i, h := range handlers {
		if &h == &handler {
			pm.eventHandlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
	
	// 从当前播放器中移除
	if pm.currentPlayer != nil {
		pm.currentPlayer.RemoveEventHandler(eventType, handler)
	}
}

// setupPlayerEventHandlers 设置播放器事件处理器
func (pm *PlayerManager) setupPlayerEventHandlers(player PlayerBackend) {
	// 为新播放器添加所有已注册的事件处理器
	for eventType, handlers := range pm.eventHandlers {
		for _, handler := range handlers {
			player.AddEventHandler(eventType, handler)
		}
	}
}

// handleBackendSwitch 处理后端切换事件
func (pm *PlayerManager) handleBackendSwitch(from, to string) {
	// 发送后端切换事件
	event := &Event{
		Type:      "backend_switched",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"from": from,
			"to":   to,
		},
		Source: "PlayerManager",
	}
	
	// 通知所有事件处理器
	pm.mutex.RLock()
	handlers := pm.eventHandlers["backend_switched"]
	pm.mutex.RUnlock()
	
	for _, handler := range handlers {
		go handler(event)
	}
}

// Play 播放音频
func (pm *PlayerManager) Play(url string) error {
	pm.mutex.RLock()
	player := pm.currentPlayer
	pm.mutex.RUnlock()
	
	if player == nil {
		return fmt.Errorf("no player available")
	}
	
	return player.Play(url)
}

// Pause 暂停播放
func (pm *PlayerManager) Pause() error {
	pm.mutex.RLock()
	player := pm.currentPlayer
	pm.mutex.RUnlock()
	
	if player == nil {
		return fmt.Errorf("no player available")
	}
	
	return player.Pause()
}

// Resume 恢复播放
func (pm *PlayerManager) Resume() error {
	pm.mutex.RLock()
	player := pm.currentPlayer
	pm.mutex.RUnlock()
	
	if player == nil {
		return fmt.Errorf("no player available")
	}
	
	return player.Resume()
}

// Stop 停止播放
func (pm *PlayerManager) Stop() error {
	pm.mutex.RLock()
	player := pm.currentPlayer
	pm.mutex.RUnlock()
	
	if player == nil {
		return fmt.Errorf("no player available")
	}
	
	return player.Stop()
}

// Seek 跳转到指定位置
func (pm *PlayerManager) Seek(position time.Duration) error {
	pm.mutex.RLock()
	player := pm.currentPlayer
	pm.mutex.RUnlock()
	
	if player == nil {
		return fmt.Errorf("no player available")
	}
	
	return player.Seek(position)
}

// SetVolume 设置音量
func (pm *PlayerManager) SetVolume(volume float64) error {
	pm.mutex.RLock()
	player := pm.currentPlayer
	pm.mutex.RUnlock()
	
	if player == nil {
		return fmt.Errorf("no player available")
	}
	
	return player.SetVolume(volume)
}

// GetVolume 获取音量
func (pm *PlayerManager) GetVolume() (float64, error) {
	pm.mutex.RLock()
	player := pm.currentPlayer
	pm.mutex.RUnlock()
	
	if player == nil {
		return 0, fmt.Errorf("no player available")
	}
	
	return player.GetVolume()
}

// GetState 获取播放状态
func (pm *PlayerManager) GetState() PlaybackState {
	pm.mutex.RLock()
	player := pm.currentPlayer
	pm.mutex.RUnlock()
	
	if player == nil {
		return StateStopped
	}
	
	return player.GetState()
}

// GetPosition 获取播放位置
func (pm *PlayerManager) GetPosition() (time.Duration, error) {
	pm.mutex.RLock()
	player := pm.currentPlayer
	pm.mutex.RUnlock()
	
	if player == nil {
		return 0, fmt.Errorf("no player available")
	}
	
	return player.GetPosition()
}

// GetDuration 获取音乐时长
func (pm *PlayerManager) GetDuration() (time.Duration, error) {
	pm.mutex.RLock()
	player := pm.currentPlayer
	pm.mutex.RUnlock()
	
	if player == nil {
		return 0, fmt.Errorf("no player available")
	}
	
	return player.GetDuration()
}

// IsPlaying 检查是否正在播放
func (pm *PlayerManager) IsPlaying() bool {
	pm.mutex.RLock()
	player := pm.currentPlayer
	pm.mutex.RUnlock()
	
	if player == nil {
		return false
	}
	
	return player.IsPlaying()
}

// HealthCheck 健康检查
func (pm *PlayerManager) HealthCheck() error {
	pm.mutex.RLock()
	player := pm.currentPlayer
	pm.mutex.RUnlock()
	
	if player == nil {
		return fmt.Errorf("no player available")
	}
	
	return player.HealthCheck()
}