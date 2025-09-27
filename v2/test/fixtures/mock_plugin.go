package fixtures

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/event"
	"github.com/go-musicfox/go-musicfox/v2/pkg/kernel"
)

// MockPlugin 模拟插件实现
type MockPlugin struct {
	info         *kernel.PluginInfo
	capabilities []string
	dependencies []string
	config       map[string]interface{}
	ctx          kernel.PluginContext
	logger       *slog.Logger
	running      bool
	mutex        sync.RWMutex

	// 测试用的状态跟踪
	initializeCalled bool
	startCalled      bool
	stopCalled       bool
	cleanupCalled    bool
	healthCheckCount int
	configUpdates    int
}

// NewMockPlugin 创建模拟插件
func NewMockPlugin(name, version string) *MockPlugin {
	return &MockPlugin{
		info: &kernel.PluginInfo{
			Name:        name,
			Version:     version,
			Description: fmt.Sprintf("Mock plugin %s for testing", name),
			Author:      "Test Suite",
			License:     "MIT",
			Homepage:    "https://example.com",
			Tags:        []string{"test", "mock"},
			Config:      make(map[string]string),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		capabilities: []string{"basic", "configurable"},
		dependencies: []string{},
		config:       make(map[string]interface{}),
	}
}

// NewMockPluginWithDependencies 创建带依赖的模拟插件
func NewMockPluginWithDependencies(name, version string, dependencies []string) *MockPlugin {
	plugin := NewMockPlugin(name, version)
	plugin.dependencies = dependencies
	return plugin
}

// GetInfo 获取插件信息
func (mp *MockPlugin) GetInfo() *kernel.PluginInfo {
	return mp.info
}

// GetCapabilities 获取插件能力
func (mp *MockPlugin) GetCapabilities() []string {
	return mp.capabilities
}

// GetDependencies 获取插件依赖
func (mp *MockPlugin) GetDependencies() []string {
	return mp.dependencies
}

// Initialize 初始化插件
func (mp *MockPlugin) Initialize(ctx kernel.PluginContext) error {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	mp.ctx = ctx
	mp.logger = ctx.GetLogger()
	mp.initializeCalled = true

	mp.logger.Info("Mock plugin initialized", "name", mp.info.Name)
	return nil
}

// Start 启动插件
func (mp *MockPlugin) Start() error {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	if !mp.initializeCalled {
		return fmt.Errorf("plugin not initialized")
	}

	mp.running = true
	mp.startCalled = true

	if mp.logger != nil {
		mp.logger.Info("Mock plugin started", "name", mp.info.Name)
	}
	return nil
}

// Stop 停止插件
func (mp *MockPlugin) Stop() error {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	mp.running = false
	mp.stopCalled = true

	if mp.logger != nil {
		mp.logger.Info("Mock plugin stopped", "name", mp.info.Name)
	}
	return nil
}

// Cleanup 清理插件
func (mp *MockPlugin) Cleanup() error {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	mp.cleanupCalled = true

	if mp.logger != nil {
		mp.logger.Info("Mock plugin cleaned up", "name", mp.info.Name)
	}
	return nil
}

// HealthCheck 健康检查
func (mp *MockPlugin) HealthCheck() error {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	mp.healthCheckCount++

	if !mp.running {
		return fmt.Errorf("plugin not running")
	}

	return nil
}

// ValidateConfig 验证配置
func (mp *MockPlugin) ValidateConfig(config map[string]interface{}) error {
	// 简单的配置验证逻辑
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// 检查必需的配置项
	if requiredKey, exists := config["required_key"]; exists {
		if requiredKey == "" {
			return fmt.Errorf("required_key cannot be empty")
		}
	}

	return nil
}

// UpdateConfig 更新配置
func (mp *MockPlugin) UpdateConfig(config map[string]interface{}) error {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	if err := mp.ValidateConfig(config); err != nil {
		return err
	}

	mp.config = config
	mp.configUpdates++

	if mp.logger != nil {
		mp.logger.Info("Mock plugin config updated", "name", mp.info.Name, "updates", mp.configUpdates)
	}

	return nil
}

// 测试辅助方法

// IsInitializeCalled 检查是否调用了Initialize
func (mp *MockPlugin) IsInitializeCalled() bool {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	return mp.initializeCalled
}

// IsStartCalled 检查是否调用了Start
func (mp *MockPlugin) IsStartCalled() bool {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	return mp.startCalled
}

// IsStopCalled 检查是否调用了Stop
func (mp *MockPlugin) IsStopCalled() bool {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	return mp.stopCalled
}

// IsCleanupCalled 检查是否调用了Cleanup
func (mp *MockPlugin) IsCleanupCalled() bool {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	return mp.cleanupCalled
}

// GetHealthCheckCount 获取健康检查调用次数
func (mp *MockPlugin) GetHealthCheckCount() int {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	return mp.healthCheckCount
}

// GetConfigUpdates 获取配置更新次数
func (mp *MockPlugin) GetConfigUpdates() int {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	return mp.configUpdates
}

// IsRunning 检查插件是否运行中
func (mp *MockPlugin) IsRunning() bool {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	return mp.running
}

// MockPluginContext 模拟插件上下文
type MockPluginContext struct {
	logger          *slog.Logger
	eventBus        event.EventBus
	serviceRegistry kernel.ServiceRegistry
	securityManager kernel.SecurityManager
	config          map[string]interface{}
}

// NewMockPluginContext 创建模拟插件上下文
func NewMockPluginContext() *MockPluginContext {
	return &MockPluginContext{
		logger: slog.Default(),
		config: make(map[string]interface{}),
	}
}

// GetLogger 获取日志器
func (mpc *MockPluginContext) GetLogger() *slog.Logger {
	return mpc.logger
}

// GetEventBus 获取事件总线
func (mpc *MockPluginContext) GetEventBus() event.EventBus {
	return mpc.eventBus
}

// GetServiceRegistry 获取服务注册表
func (mpc *MockPluginContext) GetServiceRegistry() kernel.ServiceRegistry {
	return mpc.serviceRegistry
}

// GetSecurityManager 获取安全管理器
func (mpc *MockPluginContext) GetSecurityManager() kernel.SecurityManager {
	return mpc.securityManager
}

// GetConfig 获取配置
func (mpc *MockPluginContext) GetConfig() map[string]interface{} {
	return mpc.config
}

// SetEventBus 设置事件总线
func (mpc *MockPluginContext) SetEventBus(eventBus event.EventBus) {
	mpc.eventBus = eventBus
}

// SetServiceRegistry 设置服务注册表
func (mpc *MockPluginContext) SetServiceRegistry(serviceRegistry kernel.ServiceRegistry) {
	mpc.serviceRegistry = serviceRegistry
}

// SetSecurityManager 设置安全管理器
func (mpc *MockPluginContext) SetSecurityManager(securityManager kernel.SecurityManager) {
	mpc.securityManager = securityManager
}

// SetConfig 设置配置
func (mpc *MockPluginContext) SetConfig(config map[string]interface{}) {
	mpc.config = config
}

// MockAudioProcessorPlugin 模拟音频处理插件
type MockAudioProcessorPlugin struct {
	*MockPlugin
	processCount int
	volumeAdjustments int
	effectsApplied int
}

// NewMockAudioProcessorPlugin 创建模拟音频处理插件
func NewMockAudioProcessorPlugin() *MockAudioProcessorPlugin {
	mockPlugin := NewMockPlugin("mock-audio-processor", "1.0.0")
	mockPlugin.capabilities = append(mockPlugin.capabilities, "audio_processing", "volume_control", "effects")

	return &MockAudioProcessorPlugin{
		MockPlugin: mockPlugin,
	}
}

// ProcessAudio 处理音频
func (mapp *MockAudioProcessorPlugin) ProcessAudio(input []byte, sampleRate int, channels int) ([]byte, error) {
	mapp.mutex.Lock()
	defer mapp.mutex.Unlock()

	mapp.processCount++

	if mapp.logger != nil {
		mapp.logger.Debug("Processing audio", "size", len(input), "sampleRate", sampleRate, "channels", channels)
	}

	// 模拟音频处理，直接返回输入
	return input, nil
}

// AdjustVolume 调节音量
func (mapp *MockAudioProcessorPlugin) AdjustVolume(input []byte, volume float64) ([]byte, error) {
	mapp.mutex.Lock()
	defer mapp.mutex.Unlock()

	mapp.volumeAdjustments++

	if volume < 0 || volume > 1 {
		return nil, fmt.Errorf("volume must be between 0 and 1")
	}

	if mapp.logger != nil {
		mapp.logger.Debug("Adjusting volume", "volume", volume)
	}

	// 模拟音量调节，直接返回输入
	return input, nil
}

// ApplyEffect 应用音效
func (mapp *MockAudioProcessorPlugin) ApplyEffect(input []byte, effect string) ([]byte, error) {
	mapp.mutex.Lock()
	defer mapp.mutex.Unlock()

	mapp.effectsApplied++

	if mapp.logger != nil {
		mapp.logger.Debug("Applying effect", "effect", effect)
	}

	// 模拟音效应用，直接返回输入
	return input, nil
}

// GetProcessCount 获取处理次数
func (mapp *MockAudioProcessorPlugin) GetProcessCount() int {
	mapp.mutex.RLock()
	defer mapp.mutex.RUnlock()
	return mapp.processCount
}

// GetVolumeAdjustments 获取音量调节次数
func (mapp *MockAudioProcessorPlugin) GetVolumeAdjustments() int {
	mapp.mutex.RLock()
	defer mapp.mutex.RUnlock()
	return mapp.volumeAdjustments
}

// GetEffectsApplied 获取音效应用次数
func (mapp *MockAudioProcessorPlugin) GetEffectsApplied() int {
	mapp.mutex.RLock()
	defer mapp.mutex.RUnlock()
	return mapp.effectsApplied
}

// MockMusicSourcePlugin 模拟音乐源插件
type MockMusicSourcePlugin struct {
	*MockPlugin
	searchCount int
	playlistRequests int
	songRequests int
}

// NewMockMusicSourcePlugin 创建模拟音乐源插件
func NewMockMusicSourcePlugin() *MockMusicSourcePlugin {
	mockPlugin := NewMockPlugin("mock-music-source", "1.0.0")
	mockPlugin.capabilities = append(mockPlugin.capabilities, "search", "playlist", "song_info")

	return &MockMusicSourcePlugin{
		MockPlugin: mockPlugin,
	}
}

// Search 搜索音乐
func (mms *MockMusicSourcePlugin) Search(ctx context.Context, query string) ([]map[string]interface{}, error) {
	mms.mutex.Lock()
	defer mms.mutex.Unlock()

	mms.searchCount++

	if mms.logger != nil {
		mms.logger.Debug("Searching music", "query", query)
	}

	// 返回模拟搜索结果
	return []map[string]interface{}{
		{
			"id":     "1",
			"title":  "Mock Song 1",
			"artist": "Mock Artist",
			"album":  "Mock Album",
		},
		{
			"id":     "2",
			"title":  "Mock Song 2",
			"artist": "Mock Artist 2",
			"album":  "Mock Album 2",
		},
	}, nil
}

// GetPlaylist 获取播放列表
func (mms *MockMusicSourcePlugin) GetPlaylist(ctx context.Context, id string) (map[string]interface{}, error) {
	mms.mutex.Lock()
	defer mms.mutex.Unlock()

	mms.playlistRequests++

	if mms.logger != nil {
		mms.logger.Debug("Getting playlist", "id", id)
	}

	// 返回模拟播放列表
	return map[string]interface{}{
		"id":          id,
		"name":        "Mock Playlist",
		"description": "A mock playlist for testing",
		"songs": []map[string]interface{}{
			{
				"id":     "1",
				"title":  "Mock Song 1",
				"artist": "Mock Artist",
			},
		},
	}, nil
}

// GetSong 获取歌曲信息
func (mms *MockMusicSourcePlugin) GetSong(ctx context.Context, id string) (map[string]interface{}, error) {
	mms.mutex.Lock()
	defer mms.mutex.Unlock()

	mms.songRequests++

	if mms.logger != nil {
		mms.logger.Debug("Getting song", "id", id)
	}

	// 返回模拟歌曲信息
	return map[string]interface{}{
		"id":       id,
		"title":    "Mock Song",
		"artist":   "Mock Artist",
		"album":    "Mock Album",
		"duration": 180,
		"url":      "https://example.com/mock-song.mp3",
	}, nil
}

// GetSearchCount 获取搜索次数
func (mms *MockMusicSourcePlugin) GetSearchCount() int {
	mms.mutex.RLock()
	defer mms.mutex.RUnlock()
	return mms.searchCount
}

// GetPlaylistRequests 获取播放列表请求次数
func (mms *MockMusicSourcePlugin) GetPlaylistRequests() int {
	mms.mutex.RLock()
	defer mms.mutex.RUnlock()
	return mms.playlistRequests
}

// GetSongRequests 获取歌曲请求次数
func (mms *MockMusicSourcePlugin) GetSongRequests() int {
	mms.mutex.RLock()
	defer mms.mutex.RUnlock()
	return mms.songRequests
}