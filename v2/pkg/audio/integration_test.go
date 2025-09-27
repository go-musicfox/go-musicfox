package audio

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// AudioIntegrationTestSuite 音频系统集成测试套件
type AudioIntegrationTestSuite struct {
	suite.Suite
	manager       *PlayerManager
	configManager *ConfigManager
	tempDir       string
	configPath    string
}

// SetupSuite 设置测试套件
func (suite *AudioIntegrationTestSuite) SetupSuite() {
	suite.tempDir = suite.T().TempDir()
	suite.configPath = filepath.Join(suite.tempDir, "audio_config.json")
	
	// 创建配置管理器
	cm, err := NewConfigManager(suite.configPath)
	suite.Require().NoError(err)
	suite.configManager = cm
}

// SetupTest 设置每个测试
func (suite *AudioIntegrationTestSuite) SetupTest() {
	// 为每个测试创建新的播放器管理器
	suite.manager = NewPlayerManager()
	
	// 注册测试后端
	suite.registerTestBackends()
}

// TearDownTest 清理每个测试
func (suite *AudioIntegrationTestSuite) TearDownTest() {
	if suite.manager != nil {
		ctx := context.Background()
		suite.manager.Shutdown(ctx)
		suite.manager = nil
	}
}

// TearDownSuite 清理测试套件
func (suite *AudioIntegrationTestSuite) TearDownSuite() {
	if suite.configManager != nil {
		suite.configManager.StopWatching()
	}
}

// registerTestBackends 注册测试后端
func (suite *AudioIntegrationTestSuite) registerTestBackends() {
	// 注册高优先级测试后端
	highPriorityInfo := &BackendInfo{
		Name:        "high_priority",
		Version:     "1.0.0",
		Description: "High priority test backend",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3", "wav", "flac"},
			SupportedPlatforms: []string{"test"},
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
		Platforms: []string{"test"},
		Priority:  10,
		Creator: func(config *BackendConfig) (PlayerBackend, error) {
			return NewMockPlayerBackend("high_priority", true), nil
		},
	}
	
	// 注册低优先级测试后端
	lowPriorityInfo := &BackendInfo{
		Name:        "low_priority",
		Version:     "1.0.0",
		Description: "Low priority test backend",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3", "wav"},
			SupportedPlatforms: []string{"test"},
			Features: map[string]bool{
				"seek":      true,
				"streaming": false,
				"volume":    true,
			},
			MaxVolume:        1.0,
			MinVolume:        0.0,
			SeekSupport:      true,
			StreamingSupport: false,
		},
		Platforms: []string{"test"},
		Priority:  5,
		Creator: func(config *BackendConfig) (PlayerBackend, error) {
			return NewMockPlayerBackend("low_priority", true), nil
		},
	}
	
	// 注册不可用后端
	unavailableInfo := &BackendInfo{
		Name:        "unavailable",
		Version:     "1.0.0",
		Description: "Unavailable test backend",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3"},
			SupportedPlatforms: []string{"test"},
		},
		Platforms: []string{"test"},
		Priority:  8,
		Creator: func(config *BackendConfig) (PlayerBackend, error) {
			return NewMockPlayerBackend("unavailable", false), nil
		},
	}
	
	err := suite.manager.factory.RegisterBackend(highPriorityInfo)
	suite.Require().NoError(err)
	
	err = suite.manager.factory.RegisterBackend(lowPriorityInfo)
	suite.Require().NoError(err)
	
	err = suite.manager.factory.RegisterBackend(unavailableInfo)
	suite.Require().NoError(err)
}

// TestManagerInitialization 测试管理器初始化
func (suite *AudioIntegrationTestSuite) TestManagerInitialization() {
	ctx := context.Background()
	
	// 初始化管理器
	err := suite.manager.Initialize(ctx)
	suite.Assert().NoError(err)
	
	// 检查初始化状态
	suite.Assert().True(suite.manager.running)
	suite.Assert().NotNil(suite.manager.currentPlayer)
	suite.Assert().Equal("high_priority", suite.manager.GetCurrentBackendName())
	
	// 检查可用后端
	available := suite.manager.GetAvailableBackends()
	suite.Assert().Contains(available, "high_priority")
	suite.Assert().Contains(available, "low_priority")
	suite.Assert().NotContains(available, "unavailable")
	
	// 检查后端按优先级排序
	suite.Assert().Equal("high_priority", available[0])
}

// TestBackendSwitching 测试后端切换
func (suite *AudioIntegrationTestSuite) TestBackendSwitching() {
	ctx := context.Background()
	err := suite.manager.Initialize(ctx)
	suite.Require().NoError(err)
	
	// 检查初始后端
	suite.Assert().Equal("high_priority", suite.manager.GetCurrentBackendName())
	
	// 切换到低优先级后端
	config := &BackendConfig{
		Name:          "low_priority",
		Enabled:       true,
		DefaultVolume: 0.7,
	}
	
	err = suite.manager.SwitchBackend("low_priority", config)
	suite.Assert().NoError(err)
	suite.Assert().Equal("low_priority", suite.manager.GetCurrentBackendName())
	
	// 检查音量是否应用
	volume, err := suite.manager.GetVolume()
	suite.Assert().NoError(err)
	suite.Assert().Equal(0.7, volume)
	
	// 尝试切换到不可用后端
	unavailableConfig := &BackendConfig{
		Name:    "unavailable",
		Enabled: true,
	}
	
	err = suite.manager.SwitchBackend("unavailable", unavailableConfig)
	suite.Assert().Error(err)
	suite.Assert().Contains(err.Error(), "is not available")
	
	// 当前后端应该保持不变
	suite.Assert().Equal("low_priority", suite.manager.GetCurrentBackendName())
}

// TestPlaybackFlow 测试完整播放流程
func (suite *AudioIntegrationTestSuite) TestPlaybackFlow() {
	ctx := context.Background()
	err := suite.manager.Initialize(ctx)
	suite.Require().NoError(err)
	
	// 测试播放
	testURL := "http://example.com/test.mp3"
	err = suite.manager.Play(testURL)
	suite.Assert().NoError(err)
	suite.Assert().Equal(StatePlaying, suite.manager.GetState())
	suite.Assert().True(suite.manager.IsPlaying())
	
	// 测试音量控制
	err = suite.manager.SetVolume(0.6)
	suite.Assert().NoError(err)
	
	volume, err := suite.manager.GetVolume()
	suite.Assert().NoError(err)
	suite.Assert().Equal(0.6, volume)
	
	// 测试跳转
	seekPos := 30 * time.Second
	err = suite.manager.Seek(seekPos)
	suite.Assert().NoError(err)
	
	position, err := suite.manager.GetPosition()
	suite.Assert().NoError(err)
	suite.Assert().Equal(seekPos, position)
	
	// 测试暂停
	err = suite.manager.Pause()
	suite.Assert().NoError(err)
	suite.Assert().Equal(StatePaused, suite.manager.GetState())
	suite.Assert().False(suite.manager.IsPlaying())
	
	// 测试恢复
	err = suite.manager.Resume()
	suite.Assert().NoError(err)
	suite.Assert().Equal(StatePlaying, suite.manager.GetState())
	suite.Assert().True(suite.manager.IsPlaying())
	
	// 测试停止
	err = suite.manager.Stop()
	suite.Assert().NoError(err)
	suite.Assert().Equal(StateStopped, suite.manager.GetState())
	suite.Assert().False(suite.manager.IsPlaying())
}

// TestEventSystem 测试事件系统
func (suite *AudioIntegrationTestSuite) TestEventSystem() {
	ctx := context.Background()
	err := suite.manager.Initialize(ctx)
	suite.Require().NoError(err)
	
	// 添加事件处理器
	stateEvents := make([]*Event, 0)
	volumeEvents := make([]*Event, 0)
	backendSwitchEvents := make([]*Event, 0)
	
	stateHandler := func(event *Event) {
		stateEvents = append(stateEvents, event)
	}
	
	volumeHandler := func(event *Event) {
		volumeEvents = append(volumeEvents, event)
	}
	
	backendSwitchHandler := func(event *Event) {
		backendSwitchEvents = append(backendSwitchEvents, event)
	}
	
	suite.manager.AddEventHandler(EventStateChanged, stateHandler)
	suite.manager.AddEventHandler(EventVolumeChanged, volumeHandler)
	suite.manager.AddEventHandler("backend_switched", backendSwitchHandler)
	
	// 触发状态变化事件
	err = suite.manager.Play("http://example.com/test.mp3")
	suite.Assert().NoError(err)
	
	err = suite.manager.Pause()
	suite.Assert().NoError(err)
	
	// 触发音量变化事件
	err = suite.manager.SetVolume(0.5)
	suite.Assert().NoError(err)
	
	// 触发后端切换事件
	config := &BackendConfig{
		Name:    "low_priority",
		Enabled: true,
	}
	
	err = suite.manager.SwitchBackend("low_priority", config)
	suite.Assert().NoError(err)
	
	// 等待事件处理
	time.Sleep(100 * time.Millisecond)
	
	// 检查事件是否被触发
	suite.Assert().Greater(len(stateEvents), 0)
	suite.Assert().Greater(len(volumeEvents), 0)
	suite.Assert().Greater(len(backendSwitchEvents), 0)
	
	// 检查状态变化事件
	found := false
	for _, event := range stateEvents {
		if event.Type == EventStateChanged {
			if newState, ok := event.Data["new_state"].(string); ok && newState == "playing" {
				found = true
				break
			}
		}
	}
	suite.Assert().True(found, "Should receive state change event to playing")
	
	// 检查音量变化事件
	found = false
	for _, event := range volumeEvents {
		if event.Type == EventVolumeChanged {
			if newVolume, ok := event.Data["new_volume"].(float64); ok && newVolume == 0.5 {
				found = true
				break
			}
		}
	}
	suite.Assert().True(found, "Should receive volume change event")
	
	// 检查后端切换事件
	found = false
	for _, event := range backendSwitchEvents {
		if event.Type == "backend_switched" {
			if toBackend, ok := event.Data["to"].(string); ok && toBackend == "low_priority" {
				found = true
				break
			}
		}
	}
	suite.Assert().True(found, "Should receive backend switch event")
}

// TestConfigIntegration 测试配置集成
func (suite *AudioIntegrationTestSuite) TestConfigIntegration() {
	// 添加新的后端配置
	newBackendConfig := &BackendConfig{
		Name:          "test_backend",
		Enabled:       true,
		Priority:      7,
		DefaultVolume: 0.7,
		Settings: map[string]interface{}{
			"custom_setting": "test_value",
		},
	}
	
	err := suite.configManager.UpdateBackendConfig("test_backend", newBackendConfig)
	suite.Assert().NoError(err)
	
	// 更新配置
	config := suite.configManager.GetConfig()
	originalBackend := config.DefaultBackend
	config.GlobalSettings.DefaultVolume = 0.6
	
	err = suite.configManager.UpdateConfig(config)
	suite.Assert().NoError(err)
	
	// 验证配置更新
	updatedConfig := suite.configManager.GetConfig()
	suite.Assert().Equal(originalBackend, updatedConfig.DefaultBackend) // 保持原始默认后端
	suite.Assert().Equal(0.6, updatedConfig.GlobalSettings.DefaultVolume)
	
	// 验证后端配置
	backendConfig, err := suite.configManager.GetBackendConfig("test_backend")
	suite.Assert().NoError(err)
	suite.Assert().Equal("test_backend", backendConfig.Name)
	suite.Assert().True(backendConfig.Enabled)
	suite.Assert().Equal(7, backendConfig.Priority)
	suite.Assert().Equal(0.7, backendConfig.DefaultVolume)
	suite.Assert().Equal("test_value", backendConfig.Settings["custom_setting"])
	
	// 测试配置验证
	err = suite.configManager.ValidateConfig(updatedConfig)
	suite.Assert().NoError(err)
}

// TestBackendCapabilities 测试后端能力检测
func (suite *AudioIntegrationTestSuite) TestBackendCapabilities() {
	ctx := context.Background()
	err := suite.manager.Initialize(ctx)
	suite.Require().NoError(err)
	
	// 获取高优先级后端信息
	highPriorityInfo, err := suite.manager.GetBackendInfo("high_priority")
	suite.Assert().NoError(err)
	suite.Assert().NotNil(highPriorityInfo)
	
	// 检查能力
	caps := highPriorityInfo.Capabilities
	suite.Assert().NotNil(caps)
	suite.Assert().Contains(caps.SupportedFormats, "mp3")
	suite.Assert().Contains(caps.SupportedFormats, "wav")
	suite.Assert().Contains(caps.SupportedFormats, "flac")
	suite.Assert().True(caps.Features["seek"])
	suite.Assert().True(caps.Features["streaming"])
	suite.Assert().True(caps.Features["volume"])
	suite.Assert().True(caps.SeekSupport)
	suite.Assert().True(caps.StreamingSupport)
	
	// 获取低优先级后端信息
	lowPriorityInfo, err := suite.manager.GetBackendInfo("low_priority")
	suite.Assert().NoError(err)
	suite.Assert().NotNil(lowPriorityInfo)
	
	// 检查能力差异
	lowCaps := lowPriorityInfo.Capabilities
	suite.Assert().NotContains(lowCaps.SupportedFormats, "flac")
	suite.Assert().False(lowCaps.Features["streaming"])
	suite.Assert().False(lowCaps.StreamingSupport)
	
	// 获取所有后端信息
	allBackends := suite.manager.GetAllBackends()
	suite.Assert().Contains(allBackends, "high_priority")
	suite.Assert().Contains(allBackends, "low_priority")
	suite.Assert().Contains(allBackends, "unavailable")
	
	// 检查不可用后端
	unavailableInfo := allBackends["unavailable"]
	suite.Assert().False(unavailableInfo.Available)
}

// TestHealthCheck 测试健康检查
func (suite *AudioIntegrationTestSuite) TestHealthCheck() {
	ctx := context.Background()
	err := suite.manager.Initialize(ctx)
	suite.Require().NoError(err)
	
	// 测试管理器健康检查
	err = suite.manager.HealthCheck()
	suite.Assert().NoError(err)
	
	// 测试当前播放器健康检查
	currentPlayer := suite.manager.GetCurrentPlayer()
	suite.Assert().NotNil(currentPlayer)
	
	err = currentPlayer.HealthCheck()
	suite.Assert().NoError(err)
}

// TestConcurrentOperations 测试并发操作
func (suite *AudioIntegrationTestSuite) TestConcurrentOperations() {
	ctx := context.Background()
	err := suite.manager.Initialize(ctx)
	suite.Require().NoError(err)
	
	// 并发播放控制
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(index int) {
			defer func() { done <- true }()
			
			// 随机操作
			switch index % 4 {
			case 0:
				suite.manager.Play("http://example.com/test.mp3")
			case 1:
				suite.manager.SetVolume(float64(index%10) / 10.0)
			case 2:
				suite.manager.Seek(time.Duration(index) * time.Second)
			case 3:
				suite.manager.GetState()
			}
		}(i)
	}
	
	// 等待所有操作完成
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// 操作完成
		case <-time.After(5 * time.Second):
			suite.T().Fatal("Concurrent operations timeout")
		}
	}
	
	// 系统应该仍然正常工作
	err = suite.manager.HealthCheck()
	suite.Assert().NoError(err)
}

// TestErrorHandling 测试错误处理
func (suite *AudioIntegrationTestSuite) TestErrorHandling() {
	ctx := context.Background()
	err := suite.manager.Initialize(ctx)
	suite.Require().NoError(err)
	
	// 测试切换到不存在的后端
	err = suite.manager.SwitchBackend("nonexistent", &BackendConfig{Name: "nonexistent"})
	suite.Assert().Error(err)
	suite.Assert().Contains(err.Error(), "backend not found")
	
	// 测试无效音量设置
	err = suite.manager.SetVolume(-0.1)
	suite.Assert().Error(err)
	
	err = suite.manager.SetVolume(1.1)
	suite.Assert().Error(err)
	
	// 系统应该仍然正常工作
	err = suite.manager.HealthCheck()
	suite.Assert().NoError(err)
	
	// 当前状态应该保持一致
	backendName := suite.manager.GetCurrentBackendName()
	suite.Assert().NotEmpty(backendName)
}

// TestIntegration 运行集成测试套件
func TestIntegration(t *testing.T) {
	suite.Run(t, new(AudioIntegrationTestSuite))
}

// TestEndToEndScenario 端到端场景测试
func TestEndToEndScenario(t *testing.T) {
	// 创建临时目录和配置
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "audio_config.json")
	
	// 创建配置管理器
	configManager, err := NewConfigManager(configPath)
	require.NoError(t, err)
	defer configManager.StopWatching()
	
	// 创建播放器管理器
	manager := NewPlayerManager()
	
	// 注册测试后端
	testInfo := &BackendInfo{
		Name:        "e2e_test",
		Version:     "1.0.0",
		Description: "End-to-end test backend",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3", "wav"},
			SupportedPlatforms: []string{"test"},
			Features: map[string]bool{
				"seek":      true,
				"streaming": true,
				"volume":    true,
			},
			SeekSupport:      true,
			StreamingSupport: true,
		},
		Platforms: []string{"test"},
		Priority:  10,
		Creator: func(config *BackendConfig) (PlayerBackend, error) {
			return NewMockPlayerBackend("e2e_test", true), nil
		},
	}
	
	err = manager.factory.RegisterBackend(testInfo)
	require.NoError(t, err)
	
	// 场景1：系统初始化
	ctx := context.Background()
	err = manager.Initialize(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "e2e_test", manager.GetCurrentBackendName())
	
	// 场景2：播放音乐
	testURL := "http://example.com/song.mp3"
	err = manager.Play(testURL)
	assert.NoError(t, err)
	assert.True(t, manager.IsPlaying())
	
	// 场景3：调整音量
	err = manager.SetVolume(0.7)
	assert.NoError(t, err)
	
	volume, err := manager.GetVolume()
	assert.NoError(t, err)
	assert.Equal(t, 0.7, volume)
	
	// 场景4：跳转播放位置
	seekPos := 45 * time.Second
	err = manager.Seek(seekPos)
	assert.NoError(t, err)
	
	position, err := manager.GetPosition()
	assert.NoError(t, err)
	assert.Equal(t, seekPos, position)
	
	// 场景5：暂停和恢复
	err = manager.Pause()
	assert.NoError(t, err)
	assert.False(t, manager.IsPlaying())
	
	err = manager.Resume()
	assert.NoError(t, err)
	assert.True(t, manager.IsPlaying())
	
	// 场景6：配置管理
	config := configManager.GetConfig()
	config.GlobalSettings.DefaultVolume = 0.8
	
	err = configManager.UpdateConfig(config)
	assert.NoError(t, err)
	
	updatedConfig := configManager.GetConfig()
	assert.Equal(t, 0.8, updatedConfig.GlobalSettings.DefaultVolume)
	
	// 场景7：健康检查
	err = manager.HealthCheck()
	assert.NoError(t, err)
	
	// 场景8：系统关闭
	err = manager.Shutdown(ctx)
	assert.NoError(t, err)
	assert.False(t, manager.running)
	assert.Nil(t, manager.currentPlayer)
}