package audio

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewConfigManager 测试配置管理器创建
func TestNewConfigManager(t *testing.T) {
	// 创建临时配置文件
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "audio_config.json")
	
	cm, err := NewConfigManager(configPath)
	assert.NoError(t, err)
	assert.NotNil(t, cm)
	assert.Equal(t, configPath, cm.configPath)
	assert.NotNil(t, cm.watcher)
	assert.NotNil(t, cm.config)
	
	// 检查默认配置是否创建
	assert.FileExists(t, configPath)
	
	// 检查默认配置内容
	config := cm.GetConfig()
	assert.NotNil(t, config)
	assert.Equal(t, "beep", config.DefaultBackend)
	assert.True(t, config.HotReload)
	assert.NotNil(t, config.Backends)
	assert.NotNil(t, config.GlobalSettings)
	
	cm.StopWatching()
}

// TestConfigManager_LoadExistingConfig 测试加载已存在的配置文件
func TestConfigManager_LoadExistingConfig(t *testing.T) {
	// 创建临时配置文件
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "audio_config.json")
	
	// 创建测试配置
	testConfig := &AudioConfig{
		DefaultBackend: "mpv",
		Backends: map[string]*BackendConfig{
			"mpv": {
				Name:          "mpv",
				Enabled:       true,
				Priority:      8,
				DefaultVolume: 0.7,
			},
		},
		GlobalSettings: &GlobalAudioSettings{
			DefaultVolume: 0.7,
			BufferSize:    8192,
		},
		HotReload: false,
	}
	
	// 写入配置文件
	data, err := json.MarshalIndent(testConfig, "", "  ")
	require.NoError(t, err)
	
	err = ioutil.WriteFile(configPath, data, 0644)
	require.NoError(t, err)
	
	// 创建配置管理器
	cm, err := NewConfigManager(configPath)
	assert.NoError(t, err)
	
	// 检查配置是否正确加载
	config := cm.GetConfig()
	assert.Equal(t, "mpv", config.DefaultBackend)
	assert.False(t, config.HotReload)
	assert.Equal(t, 0.7, config.GlobalSettings.DefaultVolume)
	assert.Equal(t, 8192, config.GlobalSettings.BufferSize)
	
	cm.StopWatching()
}

// TestConfigManager_SaveConfig 测试保存配置
func TestConfigManager_SaveConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "audio_config.json")
	
	cm, err := NewConfigManager(configPath)
	require.NoError(t, err)
	
	// 修改配置
	config := cm.GetConfig()
	config.DefaultBackend = "mpv"
	config.GlobalSettings.DefaultVolume = 0.6
	
	err = cm.UpdateConfig(config)
	assert.NoError(t, err)
	
	// 重新加载配置验证保存是否成功
	cm2, err := NewConfigManager(configPath)
	assert.NoError(t, err)
	
	loadedConfig := cm2.GetConfig()
	assert.Equal(t, "mpv", loadedConfig.DefaultBackend)
	assert.Equal(t, 0.6, loadedConfig.GlobalSettings.DefaultVolume)
	
	cm.StopWatching()
	cm2.StopWatching()
}

// TestConfigManager_UpdateBackendConfig 测试更新后端配置
func TestConfigManager_UpdateBackendConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "audio_config.json")
	
	cm, err := NewConfigManager(configPath)
	require.NoError(t, err)
	
	// 更新后端配置
	newBackendConfig := &BackendConfig{
		Name:          "test_backend",
		Enabled:       true,
		Priority:      9,
		BufferSize:    8192,
		SampleRate:    48000,
		Channels:      2,
		DefaultVolume: 0.9,
		Settings: map[string]interface{}{
			"custom_setting": "value",
		},
	}
	
	err = cm.UpdateBackendConfig("test_backend", newBackendConfig)
	assert.NoError(t, err)
	
	// 验证更新结果
	backendConfig, err := cm.GetBackendConfig("test_backend")
	assert.NoError(t, err)
	assert.Equal(t, "test_backend", backendConfig.Name)
	assert.True(t, backendConfig.Enabled)
	assert.Equal(t, 9, backendConfig.Priority)
	assert.Equal(t, 8192, backendConfig.BufferSize)
	assert.Equal(t, 48000, backendConfig.SampleRate)
	assert.Equal(t, 2, backendConfig.Channels)
	assert.Equal(t, 0.9, backendConfig.DefaultVolume)
	assert.Equal(t, "value", backendConfig.Settings["custom_setting"])
	
	cm.StopWatching()
}

// TestConfigManager_SetDefaultBackend 测试设置默认后端
func TestConfigManager_SetDefaultBackend(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "audio_config.json")
	
	cm, err := NewConfigManager(configPath)
	require.NoError(t, err)
	
	// 设置默认后端
	err = cm.SetDefaultBackend("mpv")
	assert.NoError(t, err)
	
	// 验证设置结果
	defaultBackend := cm.GetDefaultBackend()
	assert.Equal(t, "mpv", defaultBackend)
	
	// 验证配置文件是否更新
	config := cm.GetConfig()
	assert.Equal(t, "mpv", config.DefaultBackend)
	
	cm.StopWatching()
}

// TestConfigManager_ConfigChangeCallback 测试配置变化回调
func TestConfigManager_ConfigChangeCallback(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "audio_config.json")
	
	cm, err := NewConfigManager(configPath)
	require.NoError(t, err)
	
	// 添加配置变化回调
	callbackCalled := false
	var oldConfig, newConfig *AudioConfig
	
	callback := func(old, new *AudioConfig) error {
		callbackCalled = true
		oldConfig = old
		newConfig = new
		return nil
	}
	
	cm.AddConfigChangeCallback(callback)
	
	// 更新配置
	config := cm.GetConfig()
	originalBackend := config.DefaultBackend
	config.DefaultBackend = "mpv"
	
	err = cm.UpdateConfig(config)
	assert.NoError(t, err)
	
	// 等待异步回调
	time.Sleep(20 * time.Millisecond)
	
	// 验证回调是否被调用
	assert.True(t, callbackCalled)
	assert.NotNil(t, oldConfig)
	assert.NotNil(t, newConfig)
	assert.Equal(t, originalBackend, oldConfig.DefaultBackend)
	assert.Equal(t, "mpv", newConfig.DefaultBackend)
	
	cm.StopWatching()
}

// TestConfigManager_HotReload 测试配置热重载
func TestConfigManager_HotReload(t *testing.T) {
	// 跳过在某些CI环境中可能不稳定的文件监听测试
	if os.Getenv("CI") != "" {
		t.Skip("Skipping file watcher test in CI environment")
	}
	
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "audio_config.json")
	
	cm, err := NewConfigManager(configPath)
	require.NoError(t, err)
	
	// 启动配置监听
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err = cm.StartWatching(ctx)
	assert.NoError(t, err)
	
	// 添加配置变化回调
	callbackCalled := false
	callback := func(old, new *AudioConfig) error {
		callbackCalled = true
		return nil
	}
	
	cm.AddConfigChangeCallback(callback)
	
	// 修改配置文件
	config := cm.GetConfig()
	config.DefaultBackend = "mpv"
	config.GlobalSettings.DefaultVolume = 0.5
	
	data, err := json.MarshalIndent(config, "", "  ")
	require.NoError(t, err)
	
	// 写入文件触发热重载
	err = ioutil.WriteFile(configPath, data, 0644)
	assert.NoError(t, err)
	
	// 等待文件监听和配置重载
	time.Sleep(500 * time.Millisecond)
	
	// 验证配置是否更新
	updatedConfig := cm.GetConfig()
	assert.Equal(t, "mpv", updatedConfig.DefaultBackend)
	assert.Equal(t, 0.5, updatedConfig.GlobalSettings.DefaultVolume)
	
	// 验证回调是否被调用
	assert.True(t, callbackCalled)
	
	cm.StopWatching()
}

// TestConfigManager_ValidateConfig 测试配置验证
func TestConfigManager_ValidateConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "audio_config.json")
	
	cm, err := NewConfigManager(configPath)
	require.NoError(t, err)
	
	// 测试有效配置
	validConfig := &AudioConfig{
		DefaultBackend: "beep",
		Backends: map[string]*BackendConfig{
			"beep": {
				Name:          "beep",
				Enabled:       true,
				DefaultVolume: 0.8,
			},
		},
		GlobalSettings: &GlobalAudioSettings{
			DefaultVolume: 0.8,
			BufferSize:    4096,
			SampleRate:    44100,
			Channels:      2,
		},
		HotReload: true,
	}
	
	err = cm.ValidateConfig(validConfig)
	assert.NoError(t, err)
	
	// 测试空配置
	err = cm.ValidateConfig(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config cannot be nil")
	
	// 测试空默认后端
	invalidConfig1 := &AudioConfig{
		DefaultBackend: "",
	}
	
	err = cm.ValidateConfig(invalidConfig1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "default backend cannot be empty")
	
	// 测试默认后端不在后端列表中
	invalidConfig2 := &AudioConfig{
		DefaultBackend: "nonexistent",
		Backends: map[string]*BackendConfig{
			"beep": {
				Name: "beep",
			},
		},
	}
	
	err = cm.ValidateConfig(invalidConfig2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "default backend 'nonexistent' not found in backends list")
	
	// 测试无效音量
	invalidConfig3 := &AudioConfig{
		DefaultBackend: "beep",
		GlobalSettings: &GlobalAudioSettings{
			DefaultVolume: 1.5, // 无效音量
			BufferSize:    4096,
			SampleRate:    44100,
			Channels:      2,
		},
	}
	
	err = cm.ValidateConfig(invalidConfig3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "default volume must be between 0 and 1")
	
	// 测试无效缓冲区大小
	invalidConfig4 := &AudioConfig{
		DefaultBackend: "beep",
		GlobalSettings: &GlobalAudioSettings{
			DefaultVolume: 0.8,
			BufferSize:    -1, // 无效缓冲区大小
			SampleRate:    44100,
			Channels:      2,
		},
	}
	
	err = cm.ValidateConfig(invalidConfig4)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "buffer size must be positive")
	
	// 测试后端名称不匹配
	invalidConfig5 := &AudioConfig{
		DefaultBackend: "beep",
		Backends: map[string]*BackendConfig{
			"beep": {
				Name: "different_name", // 名称不匹配
			},
		},
	}
	
	err = cm.ValidateConfig(invalidConfig5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backend name mismatch")
	
	cm.StopWatching()
}

// TestConfigManager_GetBackendConfig 测试获取后端配置
func TestConfigManager_GetBackendConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "audio_config.json")
	
	cm, err := NewConfigManager(configPath)
	require.NoError(t, err)
	
	// 获取存在的后端配置
	beepConfig, err := cm.GetBackendConfig("beep")
	assert.NoError(t, err)
	assert.NotNil(t, beepConfig)
	assert.Equal(t, "beep", beepConfig.Name)
	
	// 获取不存在的后端配置
	_, err = cm.GetBackendConfig("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backend 'nonexistent' not found in config")
	
	cm.StopWatching()
}

// TestGlobalAudioSettings 测试全局音频设置
func TestGlobalAudioSettings(t *testing.T) {
	settings := &GlobalAudioSettings{
		DefaultVolume:     0.8,
		BufferSize:        4096,
		SampleRate:        44100,
		Channels:          2,
		AutoSwitchBackend: true,
		RetryAttempts:     3,
		RetryDelay:        "1s",
	}
	
	assert.Equal(t, 0.8, settings.DefaultVolume)
	assert.Equal(t, 4096, settings.BufferSize)
	assert.Equal(t, 44100, settings.SampleRate)
	assert.Equal(t, 2, settings.Channels)
	assert.True(t, settings.AutoSwitchBackend)
	assert.Equal(t, 3, settings.RetryAttempts)
	assert.Equal(t, "1s", settings.RetryDelay)
}

// TestAudioConfig 测试音频配置结构
func TestAudioConfig(t *testing.T) {
	config := &AudioConfig{
		DefaultBackend: "beep",
		Backends: map[string]*BackendConfig{
			"beep": {
				Name:          "beep",
				Enabled:       true,
				Priority:      5,
				BufferSize:    4096,
				SampleRate:    44100,
				Channels:      2,
				DefaultVolume: 0.8,
				Settings:      make(map[string]interface{}),
			},
		},
		GlobalSettings: &GlobalAudioSettings{
			DefaultVolume:     0.8,
			BufferSize:        4096,
			SampleRate:        44100,
			Channels:          2,
			AutoSwitchBackend: true,
			RetryAttempts:     3,
			RetryDelay:        "1s",
		},
		HotReload:  true,
		ConfigPath: "/path/to/config.json",
	}
	
	assert.Equal(t, "beep", config.DefaultBackend)
	assert.NotNil(t, config.Backends)
	assert.Contains(t, config.Backends, "beep")
	assert.NotNil(t, config.GlobalSettings)
	assert.True(t, config.HotReload)
	assert.Equal(t, "/path/to/config.json", config.ConfigPath)
	
	// 测试后端配置
	beepConfig := config.Backends["beep"]
	assert.Equal(t, "beep", beepConfig.Name)
	assert.True(t, beepConfig.Enabled)
	assert.Equal(t, 5, beepConfig.Priority)
	assert.Equal(t, 4096, beepConfig.BufferSize)
	assert.Equal(t, 44100, beepConfig.SampleRate)
	assert.Equal(t, 2, beepConfig.Channels)
	assert.Equal(t, 0.8, beepConfig.DefaultVolume)
	assert.NotNil(t, beepConfig.Settings)
}

// TestConfigManager_StopWatching 测试停止监听
func TestConfigManager_StopWatching(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "audio_config.json")
	
	cm, err := NewConfigManager(configPath)
	require.NoError(t, err)
	
	// 启动监听
	ctx := context.Background()
	err = cm.StartWatching(ctx)
	assert.NoError(t, err)
	assert.True(t, cm.running)
	
	// 停止监听
	err = cm.StopWatching()
	assert.NoError(t, err)
	assert.False(t, cm.running)
	
	// 重复停止应该不报错
	err = cm.StopWatching()
	assert.NoError(t, err)
}

// BenchmarkConfigManager_LoadConfig 基准测试配置加载性能
func BenchmarkConfigManager_LoadConfig(b *testing.B) {
	tempDir := b.TempDir()
	configPath := filepath.Join(tempDir, "audio_config.json")
	
	// 创建配置管理器
	cm, err := NewConfigManager(configPath)
	if err != nil {
		b.Fatal(err)
	}
	defer cm.StopWatching()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.LoadConfig()
	}
}

// BenchmarkConfigManager_SaveConfig 基准测试配置保存性能
func BenchmarkConfigManager_SaveConfig(b *testing.B) {
	tempDir := b.TempDir()
	configPath := filepath.Join(tempDir, "audio_config.json")
	
	// 创建配置管理器
	cm, err := NewConfigManager(configPath)
	if err != nil {
		b.Fatal(err)
	}
	defer cm.StopWatching()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.SaveConfig()
	}
}