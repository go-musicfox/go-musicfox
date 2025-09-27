package audio

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewPlayerFactory 测试播放器工厂创建
func TestNewPlayerFactory(t *testing.T) {
	factory := NewPlayerFactory()
	
	assert.NotNil(t, factory)
	assert.NotNil(t, factory.backends)
	assert.NotNil(t, factory.configWatcher)
	assert.NotNil(t, factory.eventHandlers)
	
	// 检查是否注册了内置后端
	backends := factory.GetAllBackends()
	assert.Contains(t, backends, "beep")
	assert.Contains(t, backends, "mpd")
	assert.Contains(t, backends, "mpv")
	
	// 检查平台特定后端
	switch runtime.GOOS {
	case "darwin":
		assert.Contains(t, backends, "osx")
	case "windows":
		assert.Contains(t, backends, "windows")
	}
}

// TestPlayerFactory_RegisterBackend 测试后端注册
func TestPlayerFactory_RegisterBackend(t *testing.T) {
	factory := NewPlayerFactory()
	
	// 测试注册有效后端
	info := &BackendInfo{
		Name:        "test",
		Version:     "1.0.0",
		Description: "Test backend",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3"},
			SupportedPlatforms: []string{runtime.GOOS},
		},
		Platforms: []string{runtime.GOOS},
		Priority:  5,
		Creator: func(config *BackendConfig) (PlayerBackend, error) {
			return NewMockPlayerBackend("test", true), nil
		},
	}
	
	err := factory.RegisterBackend(info)
	assert.NoError(t, err)
	
	// 检查后端是否已注册
	registeredInfo, err := factory.GetBackendInfo("test")
	assert.NoError(t, err)
	assert.Equal(t, "test", registeredInfo.Name)
	assert.Equal(t, "1.0.0", registeredInfo.Version)
	assert.Equal(t, "Test backend", registeredInfo.Description)
	
	// 测试注册无效后端（空名称）
	invalidInfo := &BackendInfo{
		Name:    "",
		Creator: func(config *BackendConfig) (PlayerBackend, error) { return nil, nil },
	}
	
	err = factory.RegisterBackend(invalidInfo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backend name cannot be empty")
	
	// 测试注册无效后端（空创建函数）
	invalidInfo2 := &BackendInfo{
		Name:    "invalid",
		Creator: nil,
	}
	
	err = factory.RegisterBackend(invalidInfo2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backend creator cannot be nil")
}

// TestPlayerFactory_CreateBackend 测试后端创建
func TestPlayerFactory_CreateBackend(t *testing.T) {
	factory := NewPlayerFactory()
	
	// 注册测试后端
	info := &BackendInfo{
		Name:        "test",
		Version:     "1.0.0",
		Description: "Test backend",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3"},
			SupportedPlatforms: []string{runtime.GOOS},
		},
		Platforms: []string{runtime.GOOS},
		Priority:  5,
		Creator: func(config *BackendConfig) (PlayerBackend, error) {
			return NewMockPlayerBackend("test", true), nil
		},
	}
	
	err := factory.RegisterBackend(info)
	require.NoError(t, err)
	
	// 测试创建有效后端
	config := &BackendConfig{
		Name:    "test",
		Enabled: true,
	}
	
	backend, err := factory.CreateBackend("test", config)
	assert.NoError(t, err)
	assert.NotNil(t, backend)
	assert.Equal(t, "test", backend.GetName())
	
	// 清理
	backend.Cleanup()
	
	// 测试创建不存在的后端
	_, err = factory.CreateBackend("nonexistent", config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backend 'nonexistent' not found")
}

// TestPlayerFactory_GetAvailableBackends 测试获取可用后端
func TestPlayerFactory_GetAvailableBackends(t *testing.T) {
	factory := NewPlayerFactory()
	
	// 注册可用后端
	availableInfo := &BackendInfo{
		Name:        "available",
		Version:     "1.0.0",
		Description: "Available backend",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3"},
			SupportedPlatforms: []string{runtime.GOOS},
		},
		Platforms: []string{runtime.GOOS},
		Priority:  10,
		Creator: func(config *BackendConfig) (PlayerBackend, error) {
			return NewMockPlayerBackend("available", true), nil
		},
	}
	
	// 注册不可用后端
	unavailableInfo := &BackendInfo{
		Name:        "unavailable",
		Version:     "1.0.0",
		Description: "Unavailable backend",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3"},
			SupportedPlatforms: []string{runtime.GOOS},
		},
		Platforms: []string{runtime.GOOS},
		Priority:  5,
		Creator: func(config *BackendConfig) (PlayerBackend, error) {
			return NewMockPlayerBackend("unavailable", false), nil
		},
	}
	
	err := factory.RegisterBackend(availableInfo)
	require.NoError(t, err)
	
	err = factory.RegisterBackend(unavailableInfo)
	require.NoError(t, err)
	
	// 获取可用后端
	available := factory.GetAvailableBackends()
	
	// 检查可用后端包含我们注册的可用后端
	assert.Contains(t, available, "available")
	
	// 检查不可用后端不在列表中
	assert.NotContains(t, available, "unavailable")
	
	// 检查后端按优先级排序（高优先级在前）
	if len(available) >= 2 {
		// 找到我们注册的后端在列表中的位置
		availableIndex := -1
		for i, name := range available {
			if name == "available" {
				availableIndex = i
				break
			}
		}
		
		// 检查高优先级后端在前面
		if availableIndex > 0 {
			prevBackendInfo, _ := factory.GetBackendInfo(available[availableIndex-1])
			currentBackendInfo, _ := factory.GetBackendInfo(available[availableIndex])
			assert.GreaterOrEqual(t, prevBackendInfo.Priority, currentBackendInfo.Priority)
		}
	}
}

// TestPlayerFactory_GetBestBackend 测试获取最佳后端
func TestPlayerFactory_GetBestBackend(t *testing.T) {
	factory := NewPlayerFactory()
	
	// 注册高优先级后端
	highPriorityInfo := &BackendInfo{
		Name:        "high_priority",
		Version:     "1.0.0",
		Description: "High priority backend",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3"},
			SupportedPlatforms: []string{runtime.GOOS},
		},
		Platforms: []string{runtime.GOOS},
		Priority:  10,
		Creator: func(config *BackendConfig) (PlayerBackend, error) {
			return NewMockPlayerBackend("high_priority", true), nil
		},
	}
	
	// 注册低优先级后端
	lowPriorityInfo := &BackendInfo{
		Name:        "low_priority",
		Version:     "1.0.0",
		Description: "Low priority backend",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3"},
			SupportedPlatforms: []string{runtime.GOOS},
		},
		Platforms: []string{runtime.GOOS},
		Priority:  5,
		Creator: func(config *BackendConfig) (PlayerBackend, error) {
			return NewMockPlayerBackend("low_priority", true), nil
		},
	}
	
	err := factory.RegisterBackend(highPriorityInfo)
	require.NoError(t, err)
	
	err = factory.RegisterBackend(lowPriorityInfo)
	require.NoError(t, err)
	
	// 获取最佳后端
	bestBackend, err := factory.GetBestBackend()
	assert.NoError(t, err)
	assert.Equal(t, "high_priority", bestBackend)
}

// TestPlayerFactory_SwitchBackend 测试后端切换
func TestPlayerFactory_SwitchBackend(t *testing.T) {
	factory := NewPlayerFactory()
	
	// 先设置一个初始后端
	err := factory.SwitchBackend("", "old")
	assert.NoError(t, err)
	
	// 测试后端切换事件
	eventReceived := false
	var fromBackend, toBackend string
	
	handler := func(from, to string) {
		eventReceived = true
		fromBackend = from
		toBackend = to
	}
	
	factory.AddBackendSwitchHandler(handler)
	
	// 执行后端切换
	err = factory.SwitchBackend("old", "new")
	assert.NoError(t, err)
	
	// 等待事件处理
	time.Sleep(10 * time.Millisecond)
	
	// 检查事件是否触发
	assert.True(t, eventReceived)
	assert.Equal(t, "old", fromBackend)
	assert.Equal(t, "new", toBackend)
	
	// 检查当前后端
	current := factory.GetCurrentBackend()
	assert.Equal(t, "new", current)
}

// TestPlayerFactory_RefreshAvailability 测试刷新可用性
func TestPlayerFactory_RefreshAvailability(t *testing.T) {
	factory := NewPlayerFactory()
	
	// 创建一个可以动态改变可用性的模拟后端
	mockBackend := NewMockPlayerBackend("dynamic", true)
	
	dynamicInfo := &BackendInfo{
		Name:        "dynamic",
		Version:     "1.0.0",
		Description: "Dynamic availability backend",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3"},
			SupportedPlatforms: []string{runtime.GOOS},
		},
		Platforms: []string{runtime.GOOS},
		Priority:  5,
		Creator: func(config *BackendConfig) (PlayerBackend, error) {
			return mockBackend, nil
		},
	}
	
	err := factory.RegisterBackend(dynamicInfo)
	require.NoError(t, err)
	
	// 初始状态应该可用
	available := factory.GetAvailableBackends()
	assert.Contains(t, available, "dynamic")
	
	// 改变后端可用性
	mockBackend.isAvailable = false
	
	// 刷新可用性
	factory.RefreshAvailability()
	
	// 现在应该不可用
	available = factory.GetAvailableBackends()
	assert.NotContains(t, available, "dynamic")
	
	// 恢复可用性
	mockBackend.isAvailable = true
	factory.RefreshAvailability()
	
	// 应该再次可用
	available = factory.GetAvailableBackends()
	assert.Contains(t, available, "dynamic")
}

// TestPlayerFactory_GetAllBackends 测试获取所有后端
func TestPlayerFactory_GetAllBackends(t *testing.T) {
	factory := NewPlayerFactory()
	
	allBackends := factory.GetAllBackends()
	
	// 检查内置后端
	assert.Contains(t, allBackends, "beep")
	assert.Contains(t, allBackends, "mpd")
	assert.Contains(t, allBackends, "mpv")
	
	// 检查后端信息
	beepInfo := allBackends["beep"]
	assert.NotNil(t, beepInfo)
	assert.Equal(t, "beep", beepInfo.Name)
	assert.Equal(t, "1.0.0", beepInfo.Version)
	assert.NotNil(t, beepInfo.Capabilities)
	assert.Nil(t, beepInfo.Creator) // Creator 应该被隐藏
	
	// 检查能力信息
	assert.Contains(t, beepInfo.Capabilities.SupportedFormats, "mp3")
	assert.Contains(t, beepInfo.Capabilities.SupportedPlatforms, runtime.GOOS)
	assert.True(t, beepInfo.Capabilities.Features["seek"])
	assert.True(t, beepInfo.Capabilities.Features["streaming"])
	assert.True(t, beepInfo.Capabilities.Features["volume"])
}

// TestConfigWatcher 测试配置监听器
func TestConfigWatcher(t *testing.T) {
	watcher := NewConfigWatcher()
	
	// 测试添加回调
	callbackCalled := false
	var receivedConfig *BackendConfig
	
	callback := func(config *BackendConfig) {
		callbackCalled = true
		receivedConfig = config
	}
	
	watcher.AddCallback(callback)
	
	// 测试通知配置变化
	testConfig := &BackendConfig{
		Name:    "test",
		Enabled: true,
	}
	
	watcher.NotifyConfigChange(testConfig)
	
	// 等待异步处理
	time.Sleep(10 * time.Millisecond)
	
	// 检查回调是否被调用
	assert.True(t, callbackCalled)
	assert.Equal(t, testConfig, receivedConfig)
}

// TestBackendInfo 测试后端信息结构
func TestBackendInfo(t *testing.T) {
	info := &BackendInfo{
		Name:        "test",
		Version:     "1.0.0",
		Description: "Test backend",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3", "wav"},
			SupportedPlatforms: []string{"linux", "darwin"},
			Features: map[string]bool{
				"seek":   true,
				"volume": true,
			},
			MaxVolume:        1.0,
			MinVolume:        0.0,
			SeekSupport:      true,
			StreamingSupport: true,
		},
		Platforms: []string{"linux", "darwin"},
		Priority:  5,
		Available: true,
		Creator: func(config *BackendConfig) (PlayerBackend, error) {
			return nil, nil
		},
	}
	
	assert.Equal(t, "test", info.Name)
	assert.Equal(t, "1.0.0", info.Version)
	assert.Equal(t, "Test backend", info.Description)
	assert.NotNil(t, info.Capabilities)
	assert.Equal(t, []string{"linux", "darwin"}, info.Platforms)
	assert.Equal(t, 5, info.Priority)
	assert.True(t, info.Available)
	assert.NotNil(t, info.Creator)
}

// BenchmarkPlayerFactory_CreateBackend 基准测试后端创建性能
func BenchmarkPlayerFactory_CreateBackend(b *testing.B) {
	factory := NewPlayerFactory()
	
	// 注册测试后端
	info := &BackendInfo{
		Name:        "benchmark",
		Version:     "1.0.0",
		Description: "Benchmark backend",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3"},
			SupportedPlatforms: []string{runtime.GOOS},
		},
		Platforms: []string{runtime.GOOS},
		Priority:  5,
		Creator: func(config *BackendConfig) (PlayerBackend, error) {
			return NewMockPlayerBackend("benchmark", true), nil
		},
	}
	
	factory.RegisterBackend(info)
	
	config := &BackendConfig{
		Name:    "benchmark",
		Enabled: true,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		backend, err := factory.CreateBackend("benchmark", config)
		if err != nil {
			b.Fatal(err)
		}
		backend.Cleanup()
	}
}

// BenchmarkPlayerFactory_GetAvailableBackends 基准测试获取可用后端性能
func BenchmarkPlayerFactory_GetAvailableBackends(b *testing.B) {
	factory := NewPlayerFactory()
	
	// 注册多个测试后端
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("backend_%d", i)
		info := &BackendInfo{
			Name:        name,
			Version:     "1.0.0",
			Description: fmt.Sprintf("Backend %d", i),
			Capabilities: &BackendCapabilities{
				SupportedFormats:   []string{"mp3"},
				SupportedPlatforms: []string{runtime.GOOS},
			},
			Platforms: []string{runtime.GOOS},
			Priority:  i,
			Creator: func(config *BackendConfig) (PlayerBackend, error) {
				return NewMockPlayerBackend(name, true), nil
			},
		}
		factory.RegisterBackend(info)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		factory.GetAvailableBackends()
	}
}