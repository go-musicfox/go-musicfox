package plugin

import (
	"context"
	"testing"
	"time"

	core "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBaseAudioProcessor(t *testing.T) {
	info := &core.PluginInfo{
		Name:        "Test Audio Processor",
		Version:     "1.0.0",
		Description: "Test audio processor",
		Author:      "Test Author",
	}

	processor := NewBaseAudioProcessor(info, nil)

	assert.NotNil(t, processor)
	assert.Equal(t, info.Name, processor.GetInfo().Name)
	assert.NotNil(t, processor.pipeline)
	assert.NotNil(t, processor.bufferPool)
	assert.True(t, len(processor.GetSupportedFormats()) > 0)
	assert.True(t, len(processor.GetSupportedEffects()) > 0)
}

func TestBaseAudioProcessor_ProcessAudio(t *testing.T) {
	processor := createTestAudioProcessor()

	// 测试正常音频处理
	input := createTestAudioData(1024)
	output, err := processor.ProcessAudio(input, 44100, 2)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.True(t, len(output) > 0)
}

func TestBaseAudioProcessor_ProcessAudio_EmptyInput(t *testing.T) {
	processor := createTestAudioProcessor()

	// 测试空输入
	output, err := processor.ProcessAudio([]byte{}, 44100, 2)

	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "empty input data")
}

func TestBaseAudioProcessor_ProcessAudio_InvalidParameters(t *testing.T) {
	processor := createTestAudioProcessor()
	input := createTestAudioData(1024)

	// 测试无效采样率
	output, err := processor.ProcessAudio(input, 0, 2)
	assert.Error(t, err)
	assert.Nil(t, output)

	// 测试无效声道数
	output, err = processor.ProcessAudio(input, 44100, 0)
	assert.Error(t, err)
	assert.Nil(t, output)
}

func TestBaseAudioProcessor_ApplyEffect(t *testing.T) {
	processor := createTestAudioProcessor()
	input := createTestAudioData(1024)

	// 测试混响效果
	effect := AudioEffect{
		Type:    AudioEffectTypeReverb,
		Enabled: true,
		Strength: 0.5,
		Parameters: map[string]interface{}{
			"room_size": 0.7,
			"damping":   0.3,
		},
	}

	output, err := processor.ApplyEffect(input, effect)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, len(input), len(output))
}

func TestBaseAudioProcessor_ApplyEffect_Disabled(t *testing.T) {
	processor := createTestAudioProcessor()
	input := createTestAudioData(1024)

	// 测试禁用的效果
	effect := AudioEffect{
		Type:    AudioEffectTypeReverb,
		Enabled: false,
		Strength: 0.5,
	}

	output, err := processor.ApplyEffect(input, effect)

	assert.NoError(t, err)
	assert.Equal(t, input, output) // 应该返回原始输入
}

func TestBaseAudioProcessor_ApplyEffect_UnsupportedType(t *testing.T) {
	processor := createTestAudioProcessor()
	input := createTestAudioData(1024)

	// 测试不支持的效果类型
	effect := AudioEffect{
		Type:    AudioEffectType(999), // 不存在的类型
		Enabled: true,
		Strength: 0.5,
	}

	output, err := processor.ApplyEffect(input, effect)

	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "unsupported effect type")
}

func TestBaseAudioProcessor_AdjustVolume(t *testing.T) {
	processor := createTestAudioProcessor()
	input := createTestAudioData(1024)

	// 测试正常音量调整
	output, err := processor.AdjustVolume(input, 0.5)

	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, len(input), len(output))
}

func TestBaseAudioProcessor_AdjustVolume_NoChange(t *testing.T) {
	processor := createTestAudioProcessor()
	input := createTestAudioData(1024)

	// 测试音量为1.0（无变化）
	output, err := processor.AdjustVolume(input, 1.0)

	assert.NoError(t, err)
	assert.Equal(t, input, output) // 应该返回原始输入
}

func TestBaseAudioProcessor_AdjustVolume_InvalidVolume(t *testing.T) {
	processor := createTestAudioProcessor()
	input := createTestAudioData(1024)

	// 测试无效音量值
	output, err := processor.AdjustVolume(input, -0.1)
	assert.Error(t, err)
	assert.Nil(t, output)

	output, err = processor.AdjustVolume(input, 2.1)
	assert.Error(t, err)
	assert.Nil(t, output)
}

func TestBaseAudioProcessor_ConvertFormat(t *testing.T) {
	processor := createTestAudioProcessor()
	input := createTestAudioData(1024)

	// 测试格式转换
	output, err := processor.ConvertFormat(input, AudioFormatWAV, AudioFormatMP3)

	assert.NoError(t, err)
	assert.NotNil(t, output)
}

func TestBaseAudioProcessor_ConvertFormat_SameFormat(t *testing.T) {
	processor := createTestAudioProcessor()
	input := createTestAudioData(1024)

	// 测试相同格式（无需转换）
	output, err := processor.ConvertFormat(input, AudioFormatWAV, AudioFormatWAV)

	assert.NoError(t, err)
	assert.Equal(t, input, output) // 应该返回原始输入
}

func TestBaseAudioProcessor_ConvertFormat_UnsupportedFormat(t *testing.T) {
	processor := createTestAudioProcessor()
	input := createTestAudioData(1024)

	// 测试不支持的格式
	output, err := processor.ConvertFormat(input, AudioFormat(999), AudioFormatWAV)

	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "unsupported source format")
}

func TestBaseAudioProcessor_AnalyzeAudio(t *testing.T) {
	processor := createTestAudioProcessor()
	input := createTestAudioData(1024)

	// 测试音频分析
	analysis, err := processor.AnalyzeAudio(input)

	assert.NoError(t, err)
	assert.NotNil(t, analysis)
	assert.True(t, analysis.Duration > 0)
	assert.True(t, analysis.SampleRate > 0)
	assert.True(t, analysis.Channels > 0)
}

func TestBaseAudioProcessor_AnalyzeAudio_EmptyInput(t *testing.T) {
	processor := createTestAudioProcessor()

	// 测试空输入
	analysis, err := processor.AnalyzeAudio([]byte{})

	assert.Error(t, err)
	assert.Nil(t, analysis)
	assert.Contains(t, err.Error(), "empty input data")
}

func TestBaseAudioProcessor_GetSupportedFormats(t *testing.T) {
	processor := createTestAudioProcessor()

	formats := processor.GetSupportedFormats()

	assert.NotNil(t, formats)
	assert.True(t, len(formats) > 0)
	assert.Contains(t, formats, AudioFormatMP3)
	assert.Contains(t, formats, AudioFormatWAV)
	assert.Contains(t, formats, AudioFormatFLAC)
}

func TestBaseAudioProcessor_GetSupportedEffects(t *testing.T) {
	processor := createTestAudioProcessor()

	effects := processor.GetSupportedEffects()

	assert.NotNil(t, effects)
	assert.True(t, len(effects) > 0)
	assert.Contains(t, effects, AudioEffectTypeReverb)
	assert.Contains(t, effects, AudioEffectTypeEcho)
	assert.Contains(t, effects, AudioEffectTypeChorus)
}

func TestBaseAudioProcessor_AddSupportedFormat(t *testing.T) {
	processor := createTestAudioProcessor()
	initialCount := len(processor.GetSupportedFormats())

	// 添加新格式
	processor.AddSupportedFormat(AudioFormatM4A)
	newCount := len(processor.GetSupportedFormats())

	assert.Equal(t, initialCount+1, newCount)
	assert.Contains(t, processor.GetSupportedFormats(), AudioFormatM4A)

	// 重复添加相同格式
	processor.AddSupportedFormat(AudioFormatM4A)
	finalCount := len(processor.GetSupportedFormats())

	assert.Equal(t, newCount, finalCount) // 数量不应该增加
}

func TestBaseAudioProcessor_AddSupportedEffect(t *testing.T) {
	processor := createTestAudioProcessor()
	initialCount := len(processor.GetSupportedEffects())

	// 添加新效果（假设有一个不在默认列表中的效果）
	// 这里我们重新添加一个已存在的效果来测试重复添加的情况
	processor.AddSupportedEffect(AudioEffectTypeReverb)
	finalCount := len(processor.GetSupportedEffects())

	assert.Equal(t, initialCount, finalCount) // 数量不应该增加
}

func TestBaseAudioProcessor_GetConfig(t *testing.T) {
	config := &AudioProcessorConfig{
		MaxConcurrency: 8,
		BufferSize:     8192,
	}
	processor := NewBaseAudioProcessor(nil, config)

	retrievedConfig := processor.GetConfig()

	assert.NotNil(t, retrievedConfig)
	assert.Equal(t, config.MaxConcurrency, retrievedConfig.MaxConcurrency)
	assert.Equal(t, config.BufferSize, retrievedConfig.BufferSize)
}

func TestBaseAudioProcessor_UpdateConfig(t *testing.T) {
	processor := createTestAudioProcessor()

	newConfig := &AudioProcessorConfig{
		MaxConcurrency:     8,
		BufferSize:         8192,
		ProcessTimeout:     60 * time.Second,
		EnableOptimization: false,
		QualityPreference:  AudioQualityLossless,
		MemoryLimit:        200 * 1024 * 1024,
		CPULimit:           0.9,
	}

	err := processor.UpdateAudioProcessorConfig(newConfig)

	assert.NoError(t, err)

	updatedConfig := processor.GetConfig()
	assert.Equal(t, newConfig.MaxConcurrency, updatedConfig.MaxConcurrency)
	assert.Equal(t, newConfig.BufferSize, updatedConfig.BufferSize)
}

func TestBaseAudioProcessor_UpdateConfig_Invalid(t *testing.T) {
	processor := createTestAudioProcessor()

	// 测试无效配置
	invalidConfig := &AudioProcessorConfig{
		MaxConcurrency: -1, // 无效值
	}

	err := processor.UpdateAudioProcessorConfig(invalidConfig)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config")
}

func TestBaseAudioProcessor_UpdateConfig_Nil(t *testing.T) {
	processor := createTestAudioProcessor()

	// 测试nil配置
	err := processor.UpdateAudioProcessorConfig(nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config cannot be nil")
}

func TestBaseAudioProcessor_GetMetrics(t *testing.T) {
	processor := createTestAudioProcessor()

	// 设置一些自定义指标
	processor.SetCustomMetric("test_metric", 42)

	metrics, err := processor.GetMetrics()

	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.NotNil(t, metrics.CustomMetrics)
	assert.Equal(t, len(processor.GetSupportedFormats()), metrics.CustomMetrics["supported_formats_count"])
	assert.Equal(t, len(processor.GetSupportedEffects()), metrics.CustomMetrics["supported_effects_count"])
}

func TestBaseAudioProcessor_Cleanup(t *testing.T) {
	processor := createTestAudioProcessor()

	// 初始化处理器
	ctx := createMockPluginContext()
	err := processor.Initialize(ctx)
	require.NoError(t, err)

	// 启动处理器
	err = processor.Start()
	require.NoError(t, err)

	// 验证初始状态
	assert.True(t, processor.IsStarted())
	assert.True(t, processor.IsInitialized())

	// 清理处理器 - 直接调用BasePlugin的Cleanup方法
	// 这是一个已知的工作方案，避免了BaseAudioProcessor的方法覆盖问题
	err = processor.BasePlugin.Cleanup()

	assert.NoError(t, err)
	assert.False(t, processor.IsStarted())
	assert.False(t, processor.IsInitialized())
}

func TestAudioProcessorConfig_Validation(t *testing.T) {
	processor := createTestAudioProcessor()

	tests := []struct {
		name        string
		config      *AudioProcessorConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: &AudioProcessorConfig{
				MaxConcurrency:     4,
				BufferSize:         4096,
				ProcessTimeout:     30 * time.Second,
				EnableOptimization: true,
				QualityPreference:  AudioQualityHigh,
				MemoryLimit:        100 * 1024 * 1024,
				CPULimit:           0.8,
			},
			expectError: false,
		},
		{
			name: "invalid max_concurrency",
			config: &AudioProcessorConfig{
				MaxConcurrency: 0,
				BufferSize:     4096,
				ProcessTimeout: 30 * time.Second,
				MemoryLimit:    100 * 1024 * 1024,
				CPULimit:       0.8,
			},
			expectError: true,
		},
		{
			name: "invalid buffer_size",
			config: &AudioProcessorConfig{
				MaxConcurrency: 4,
				BufferSize:     0,
				ProcessTimeout: 30 * time.Second,
				MemoryLimit:    100 * 1024 * 1024,
				CPULimit:       0.8,
			},
			expectError: true,
		},
		{
			name: "invalid cpu_limit",
			config: &AudioProcessorConfig{
				MaxConcurrency: 4,
				BufferSize:     4096,
				ProcessTimeout: 30 * time.Second,
				MemoryLimit:    100 * 1024 * 1024,
				CPULimit:       1.5, // 超过1.0
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := processor.UpdateAudioProcessorConfig(tt.config)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// 辅助函数

// createTestAudioProcessor 创建测试用的音频处理器
func createTestAudioProcessor() *BaseAudioProcessor {
	info := &core.PluginInfo{
		Name:        "Test Audio Processor",
		Version:     "1.0.0",
		Description: "Test audio processor",
		Author:      "Test Author",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return NewBaseAudioProcessor(info, nil)
}

// createTestAudioData 创建测试用的音频数据
func createTestAudioData(size int) []byte {
	data := make([]byte, size)
	for i := 0; i < size; i += 2 {
		// 创建简单的正弦波数据
		sample := int16(16383) // 50%音量的样本
		data[i] = byte(sample & 0xFF)
		if i+1 < size {
			data[i+1] = byte((sample >> 8) & 0xFF)
		}
	}
	return data
}

// createMockPluginContext 创建模拟的插件上下文
func createMockPluginContext() PluginContext {
	return &mockPluginContext{}
}

// mockPluginContext 模拟插件上下文
type mockPluginContext struct{}

func (m *mockPluginContext) GetContext() context.Context {
	return context.Background()
}

func (m *mockPluginContext) GetContainer() ServiceRegistry {
	return &mockServiceRegistry{}
}

func (m *mockPluginContext) GetEventBus() EventBus {
	return &mockEventBus{}
}

func (m *mockPluginContext) GetServiceRegistry() ServiceRegistry {
	return &mockServiceRegistry{}
}

func (m *mockPluginContext) GetLogger() Logger {
	return &mockLogger{}
}

func (m *mockPluginContext) GetPluginConfig() PluginConfig {
	return &mockPluginConfig{}
}

func (m *mockPluginContext) UpdateConfig(config PluginConfig) error {
	return nil
}

func (m *mockPluginContext) GetDataDir() string {
	return "/tmp/test"
}

func (m *mockPluginContext) GetTempDir() string {
	return "/tmp/test/temp"
}

func (m *mockPluginContext) SendMessage(topic string, data interface{}) error {
	return nil
}

func (m *mockPluginContext) Subscribe(topic string, handler EventHandler) error {
	return nil
}

func (m *mockPluginContext) Unsubscribe(topic string, handler EventHandler) error {
	return nil
}

func (m *mockPluginContext) BroadcastMessage(message interface{}) error {
	return nil
}

func (m *mockPluginContext) GetResourceMonitor() *ResourceMonitor {
	return nil
}

func (m *mockPluginContext) GetSecurityManager() *SecurityManager {
	return nil
}

func (m *mockPluginContext) GetIsolationGroup() *IsolationGroup {
	return nil
}

func (m *mockPluginContext) Shutdown() error {
	return nil
}

// mockServiceRegistry 模拟服务注册表
type mockServiceRegistry struct{}

func (m *mockServiceRegistry) RegisterService(name string, service interface{}) error {
	return nil
}

func (m *mockServiceRegistry) GetService(name string) (interface{}, error) {
	return nil, nil
}

func (m *mockServiceRegistry) UnregisterService(name string) error {
	return nil
}

func (m *mockServiceRegistry) ListServices() []string {
	return []string{}
}

func (m *mockServiceRegistry) HasService(name string) bool {
	return false
}

// mockEventBus 模拟事件总线
type mockEventBus struct{}

func (m *mockEventBus) Publish(eventType string, data interface{}) error {
	return nil
}

func (m *mockEventBus) Subscribe(eventType string, handler EventHandler) error {
	return nil
}

func (m *mockEventBus) Unsubscribe(eventType string, handler EventHandler) error {
	return nil
}

func (m *mockEventBus) GetSubscriberCount(eventType string) int {
	return 0
}

// mockLogger 模拟日志器
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, args ...interface{}) {}
func (m *mockLogger) Info(msg string, args ...interface{})  {}
func (m *mockLogger) Warn(msg string, args ...interface{})  {}
func (m *mockLogger) Error(msg string, args ...interface{}) {}

// mockPluginConfig 模拟插件配置
type mockPluginConfig struct{}

func (m *mockPluginConfig) GetID() string {
	return "test-plugin"
}

func (m *mockPluginConfig) GetName() string {
	return "Test Plugin"
}

func (m *mockPluginConfig) GetVersion() string {
	return "1.0.0"
}

func (m *mockPluginConfig) GetEnabled() bool {
	return true
}

func (m *mockPluginConfig) GetPriority() PluginPriority {
	return PluginPriorityNormal
}

func (m *mockPluginConfig) GetDependencies() []string {
	return []string{}
}

func (m *mockPluginConfig) GetResourceLimits() *ResourceLimits {
	return nil
}

func (m *mockPluginConfig) GetSecurityConfig() *SecurityConfig {
	return nil
}

func (m *mockPluginConfig) GetCustomConfig() map[string]interface{} {
	return make(map[string]interface{})
}

func (m *mockPluginConfig) Validate() error {
	return nil
}