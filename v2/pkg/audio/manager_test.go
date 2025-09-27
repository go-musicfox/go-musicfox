package audio

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewPlayerManager 测试播放器管理器创建
func TestNewPlayerManager(t *testing.T) {
	manager := NewPlayerManager()
	
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.factory)
	assert.NotNil(t, manager.eventHandlers)
	assert.NotNil(t, manager.shutdownCh)
	assert.NotNil(t, manager.defaultConfig)
	assert.False(t, manager.running)
	
	// 检查默认配置
	assert.True(t, manager.defaultConfig.Enabled)
	assert.Equal(t, 5, manager.defaultConfig.Priority)
	assert.Equal(t, 4096, manager.defaultConfig.BufferSize)
	assert.Equal(t, 44100, manager.defaultConfig.SampleRate)
	assert.Equal(t, 2, manager.defaultConfig.Channels)
	assert.Equal(t, 0.8, manager.defaultConfig.DefaultVolume)
}

// TestPlayerManager_Initialize 测试播放器管理器初始化
func TestPlayerManager_Initialize(t *testing.T) {
	manager := NewPlayerManager()
	
	// 注册测试后端
	info := &BackendInfo{
		Name:        "test",
		Version:     "1.0.0",
		Description: "Test backend",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3"},
			SupportedPlatforms: []string{"test"},
		},
		Platforms: []string{"test"},
		Priority:  10, // 高优先级确保被选为最佳后端
		Creator: func(config *BackendConfig) (PlayerBackend, error) {
			return NewMockPlayerBackend("test", true), nil
		},
	}
	
	err := manager.factory.RegisterBackend(info)
	require.NoError(t, err)
	
	// 初始化管理器
	ctx := context.Background()
	err = manager.Initialize(ctx)
	assert.NoError(t, err)
	
	// 检查初始化状态
	assert.True(t, manager.running)
	assert.NotNil(t, manager.currentPlayer)
	assert.NotNil(t, manager.currentConfig)
	assert.Equal(t, "test", manager.GetCurrentBackendName())
	
	// 清理
	manager.Shutdown(ctx)
}

// TestPlayerManager_InitializeNoBackends 测试无可用后端时的初始化
func TestPlayerManager_InitializeNoBackends(t *testing.T) {
	// 创建一个空的工厂（没有可用后端）
	manager := &PlayerManager{
		factory:       &PlayerFactory{backends: make(map[string]*BackendInfo)},
		eventHandlers: make(map[EventType][]EventHandler),
		shutdownCh:    make(chan struct{}),
		defaultConfig: &BackendConfig{},
	}
	
	ctx := context.Background()
	err := manager.Initialize(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no available audio backends")
	assert.False(t, manager.running)
}

// TestPlayerManager_Shutdown 测试播放器管理器关闭
func TestPlayerManager_Shutdown(t *testing.T) {
	manager := NewPlayerManager()
	
	// 注册测试后端
	info := &BackendInfo{
		Name:        "test",
		Version:     "1.0.0",
		Description: "Test backend",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3"},
			SupportedPlatforms: []string{"test"},
		},
		Platforms: []string{"test"},
		Priority:  10,
		Creator: func(config *BackendConfig) (PlayerBackend, error) {
			return NewMockPlayerBackend("test", true), nil
		},
	}
	
	err := manager.factory.RegisterBackend(info)
	require.NoError(t, err)
	
	// 初始化管理器
	ctx := context.Background()
	err = manager.Initialize(ctx)
	require.NoError(t, err)
	
	// 检查初始化状态
	assert.True(t, manager.running)
	assert.NotNil(t, manager.currentPlayer)
	
	// 关闭管理器
	err = manager.Shutdown(ctx)
	assert.NoError(t, err)
	
	// 检查关闭状态
	assert.False(t, manager.running)
	assert.Nil(t, manager.currentPlayer)
	assert.Nil(t, manager.currentConfig)
	
	// 重复关闭应该不报错
	err = manager.Shutdown(ctx)
	assert.NoError(t, err)
}

// TestPlayerManager_SwitchBackend 测试后端切换
func TestPlayerManager_SwitchBackend(t *testing.T) {
	manager := NewPlayerManager()
	
	// 注册两个测试后端
	info1 := &BackendInfo{
		Name:        "backend1",
		Version:     "1.0.0",
		Description: "Backend 1",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3"},
			SupportedPlatforms: []string{"test"},
		},
		Platforms: []string{"test"},
		Priority:  10,
		Creator: func(config *BackendConfig) (PlayerBackend, error) {
			return NewMockPlayerBackend("backend1", true), nil
		},
	}
	
	info2 := &BackendInfo{
		Name:        "backend2",
		Version:     "1.0.0",
		Description: "Backend 2",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3"},
			SupportedPlatforms: []string{"test"},
		},
		Platforms: []string{"test"},
		Priority:  5,
		Creator: func(config *BackendConfig) (PlayerBackend, error) {
			return NewMockPlayerBackend("backend2", true), nil
		},
	}
	
	err := manager.factory.RegisterBackend(info1)
	require.NoError(t, err)
	
	err = manager.factory.RegisterBackend(info2)
	require.NoError(t, err)
	
	// 初始化管理器
	ctx := context.Background()
	err = manager.Initialize(ctx)
	require.NoError(t, err)
	
	// 检查初始后端（应该是高优先级的 backend1）
	assert.Equal(t, "backend1", manager.GetCurrentBackendName())
	
	// 切换到 backend2
	config := &BackendConfig{
		Name:    "backend2",
		Enabled: true,
	}
	
	err = manager.SwitchBackend("backend2", config)
	assert.NoError(t, err)
	
	// 检查切换结果
	assert.Equal(t, "backend2", manager.GetCurrentBackendName())
	assert.Equal(t, "backend2", manager.currentPlayer.GetName())
	
	// 清理
	manager.Shutdown(ctx)
}

// TestPlayerManager_SwitchBackendWithPlayback 测试播放过程中的后端切换
func TestPlayerManager_SwitchBackendWithPlayback(t *testing.T) {
	manager := NewPlayerManager()
	
	// 注册两个测试后端
	info1 := &BackendInfo{
		Name:        "backend1",
		Version:     "1.0.0",
		Description: "Backend 1",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3"},
			SupportedPlatforms: []string{"test"},
		},
		Platforms: []string{"test"},
		Priority:  10,
		Creator: func(config *BackendConfig) (PlayerBackend, error) {
			return NewMockPlayerBackend("backend1", true), nil
		},
	}
	
	info2 := &BackendInfo{
		Name:        "backend2",
		Version:     "1.0.0",
		Description: "Backend 2",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3"},
			SupportedPlatforms: []string{"test"},
		},
		Platforms: []string{"test"},
		Priority:  5,
		Creator: func(config *BackendConfig) (PlayerBackend, error) {
			return NewMockPlayerBackend("backend2", true), nil
		},
	}
	
	err := manager.factory.RegisterBackend(info1)
	require.NoError(t, err)
	
	err = manager.factory.RegisterBackend(info2)
	require.NoError(t, err)
	
	// 初始化管理器
	ctx := context.Background()
	err = manager.Initialize(ctx)
	require.NoError(t, err)
	
	// 开始播放
	testURL := "http://example.com/test.mp3"
	err = manager.Play(testURL)
	assert.NoError(t, err)
	assert.True(t, manager.IsPlaying())
	
	// 跳转到某个位置
	seekPos := 30 * time.Second
	err = manager.Seek(seekPos)
	assert.NoError(t, err)
	
	// 切换后端
	config := &BackendConfig{
		Name:    "backend2",
		Enabled: true,
	}
	
	err = manager.SwitchBackend("backend2", config)
	assert.NoError(t, err)
	
	// 检查切换结果
	assert.Equal(t, "backend2", manager.GetCurrentBackendName())
	
	// 等待播放状态恢复（异步操作）
	time.Sleep(200 * time.Millisecond)
	
	// 清理
	manager.Shutdown(ctx)
}

// TestPlayerManager_PlaybackControls 测试播放控制
func TestPlayerManager_PlaybackControls(t *testing.T) {
	manager := NewPlayerManager()
	
	// 注册测试后端
	info := &BackendInfo{
		Name:        "test",
		Version:     "1.0.0",
		Description: "Test backend",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3"},
			SupportedPlatforms: []string{"test"},
		},
		Platforms: []string{"test"},
		Priority:  10,
		Creator: func(config *BackendConfig) (PlayerBackend, error) {
			return NewMockPlayerBackend("test", true), nil
		},
	}
	
	err := manager.factory.RegisterBackend(info)
	require.NoError(t, err)
	
	// 初始化管理器
	ctx := context.Background()
	err = manager.Initialize(ctx)
	require.NoError(t, err)
	
	// 测试播放
	testURL := "http://example.com/test.mp3"
	err = manager.Play(testURL)
	assert.NoError(t, err)
	assert.Equal(t, StatePlaying, manager.GetState())
	assert.True(t, manager.IsPlaying())
	
	// 测试暂停
	err = manager.Pause()
	assert.NoError(t, err)
	assert.Equal(t, StatePaused, manager.GetState())
	assert.False(t, manager.IsPlaying())
	
	// 测试恢复
	err = manager.Resume()
	assert.NoError(t, err)
	assert.Equal(t, StatePlaying, manager.GetState())
	assert.True(t, manager.IsPlaying())
	
	// 测试跳转
	seekPos := 60 * time.Second
	err = manager.Seek(seekPos)
	assert.NoError(t, err)
	
	position, err := manager.GetPosition()
	assert.NoError(t, err)
	assert.Equal(t, seekPos, position)
	
	// 测试停止
	err = manager.Stop()
	assert.NoError(t, err)
	assert.Equal(t, StateStopped, manager.GetState())
	assert.False(t, manager.IsPlaying())
	
	// 清理
	manager.Shutdown(ctx)
}

// TestPlayerManager_VolumeControl 测试音量控制
func TestPlayerManager_VolumeControl(t *testing.T) {
	manager := NewPlayerManager()
	
	// 注册测试后端
	info := &BackendInfo{
		Name:        "test",
		Version:     "1.0.0",
		Description: "Test backend",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3"},
			SupportedPlatforms: []string{"test"},
		},
		Platforms: []string{"test"},
		Priority:  10,
		Creator: func(config *BackendConfig) (PlayerBackend, error) {
			return NewMockPlayerBackend("test", true), nil
		},
	}
	
	err := manager.factory.RegisterBackend(info)
	require.NoError(t, err)
	
	// 初始化管理器
	ctx := context.Background()
	err = manager.Initialize(ctx)
	require.NoError(t, err)
	
	// 测试设置音量
	err = manager.SetVolume(0.6)
	assert.NoError(t, err)
	
	volume, err := manager.GetVolume()
	assert.NoError(t, err)
	assert.Equal(t, 0.6, volume)
	
	// 测试设置无效音量
	err = manager.SetVolume(-0.1)
	assert.Error(t, err)
	
	err = manager.SetVolume(1.1)
	assert.Error(t, err)
	
	// 音量应该保持不变
	volume, _ = manager.GetVolume()
	assert.Equal(t, 0.6, volume)
	
	// 清理
	manager.Shutdown(ctx)
}

// TestPlayerManager_EventHandlers 测试事件处理器
func TestPlayerManager_EventHandlers(t *testing.T) {
	manager := NewPlayerManager()
	
	// 注册测试后端
	info := &BackendInfo{
		Name:        "test",
		Version:     "1.0.0",
		Description: "Test backend",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3"},
			SupportedPlatforms: []string{"test"},
		},
		Platforms: []string{"test"},
		Priority:  10,
		Creator: func(config *BackendConfig) (PlayerBackend, error) {
			return NewMockPlayerBackend("test", true), nil
		},
	}
	
	err := manager.factory.RegisterBackend(info)
	require.NoError(t, err)
	
	// 添加事件处理器
	stateChangeReceived := false
	volumeChangeReceived := false
	
	stateHandler := func(event *Event) {
		stateChangeReceived = true
	}
	
	volumeHandler := func(event *Event) {
		volumeChangeReceived = true
	}
	
	manager.AddEventHandler(EventStateChanged, stateHandler)
	manager.AddEventHandler(EventVolumeChanged, volumeHandler)
	
	// 初始化管理器
	ctx := context.Background()
	err = manager.Initialize(ctx)
	require.NoError(t, err)
	
	// 触发状态变化事件
	err = manager.Play("http://example.com/test.mp3")
	assert.NoError(t, err)
	
	// 触发音量变化事件
	err = manager.SetVolume(0.5)
	assert.NoError(t, err)
	
	// 等待事件处理
	time.Sleep(20 * time.Millisecond)
	
	// 检查事件是否被处理
	assert.True(t, stateChangeReceived)
	assert.True(t, volumeChangeReceived)
	
	// 清理
	manager.Shutdown(ctx)
}

// TestPlayerManager_NoPlayerOperations 测试无播放器时的操作
func TestPlayerManager_NoPlayerOperations(t *testing.T) {
	manager := NewPlayerManager()
	
	// 在未初始化时测试操作
	err := manager.Play("test.mp3")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no player available")
	
	err = manager.Pause()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no player available")
	
	err = manager.Resume()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no player available")
	
	err = manager.Stop()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no player available")
	
	err = manager.Seek(time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no player available")
	
	err = manager.SetVolume(0.5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no player available")
	
	_, err = manager.GetVolume()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no player available")
	
	_, err = manager.GetPosition()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no player available")
	
	_, err = manager.GetDuration()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no player available")
	
	err = manager.HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no player available")
	
	// 状态查询应该返回默认值
	assert.Equal(t, StateStopped, manager.GetState())
	assert.False(t, manager.IsPlaying())
	assert.Equal(t, "", manager.GetCurrentBackendName())
}

// TestPlayerManager_HealthCheck 测试健康检查
func TestPlayerManager_HealthCheck(t *testing.T) {
	manager := NewPlayerManager()
	
	// 注册测试后端
	info := &BackendInfo{
		Name:        "test",
		Version:     "1.0.0",
		Description: "Test backend",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3"},
			SupportedPlatforms: []string{"test"},
		},
		Platforms: []string{"test"},
		Priority:  10,
		Creator: func(config *BackendConfig) (PlayerBackend, error) {
			return NewMockPlayerBackend("test", true), nil
		},
	}
	
	err := manager.factory.RegisterBackend(info)
	require.NoError(t, err)
	
	// 初始化管理器
	ctx := context.Background()
	err = manager.Initialize(ctx)
	require.NoError(t, err)
	
	// 测试健康检查
	err = manager.HealthCheck()
	assert.NoError(t, err)
	
	// 清理
	manager.Shutdown(ctx)
}

// BenchmarkPlayerManager_PlaybackControls 基准测试播放控制性能
func BenchmarkPlayerManager_PlaybackControls(b *testing.B) {
	manager := NewPlayerManager()
	
	// 注册测试后端
	info := &BackendInfo{
		Name:        "benchmark",
		Version:     "1.0.0",
		Description: "Benchmark backend",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3"},
			SupportedPlatforms: []string{"test"},
		},
		Platforms: []string{"test"},
		Priority:  10,
		Creator: func(config *BackendConfig) (PlayerBackend, error) {
			return NewMockPlayerBackend("benchmark", true), nil
		},
	}
	
	manager.factory.RegisterBackend(info)
	
	// 初始化管理器
	ctx := context.Background()
	manager.Initialize(ctx)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.Play("test.mp3")
		manager.Pause()
		manager.Resume()
		manager.Stop()
	}
	
	manager.Shutdown(ctx)
}

// BenchmarkPlayerManager_VolumeControl 基准测试音量控制性能
func BenchmarkPlayerManager_VolumeControl(b *testing.B) {
	manager := NewPlayerManager()
	
	// 注册测试后端
	info := &BackendInfo{
		Name:        "benchmark",
		Version:     "1.0.0",
		Description: "Benchmark backend",
		Capabilities: &BackendCapabilities{
			SupportedFormats:   []string{"mp3"},
			SupportedPlatforms: []string{"test"},
		},
		Platforms: []string{"test"},
		Priority:  10,
		Creator: func(config *BackendConfig) (PlayerBackend, error) {
			return NewMockPlayerBackend("benchmark", true), nil
		},
	}
	
	manager.factory.RegisterBackend(info)
	
	// 初始化管理器
	ctx := context.Background()
	manager.Initialize(ctx)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		volume := float64(i%100) / 100.0
		manager.SetVolume(volume)
		manager.GetVolume()
	}
	
	manager.Shutdown(ctx)
}