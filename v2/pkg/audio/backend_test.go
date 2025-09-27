package audio

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockPlayerBackend 模拟播放器后端
type MockPlayerBackend struct {
	*BaseBackend
	isAvailable bool
	playURL     string
	playError   error
	pauseError  error
	resumeError error
	stopError   error
	seekError   error
}

// NewMockPlayerBackend 创建模拟播放器后端
func NewMockPlayerBackend(name string, available bool) *MockPlayerBackend {
	capabilities := &BackendCapabilities{
		SupportedFormats:   []string{"mp3", "wav"},
		SupportedPlatforms: []string{"test"},
		Features: map[string]bool{
			"seek":   true,
			"volume": true,
		},
		MaxVolume:        1.0,
		MinVolume:        0.0,
		SeekSupport:      true,
		StreamingSupport: true,
	}
	
	return &MockPlayerBackend{
		BaseBackend: NewBaseBackend(name, "1.0.0", capabilities),
		isAvailable: available,
	}
}

// Play 实现播放功能
func (m *MockPlayerBackend) Play(url string) error {
	if m.playError != nil {
		return m.playError
	}
	
	m.playURL = url
	m.SetState(StatePlaying)
	m.SetDuration(3 * time.Minute) // 模拟3分钟的音乐
	return nil
}

// Pause 实现暂停功能
func (m *MockPlayerBackend) Pause() error {
	if m.pauseError != nil {
		return m.pauseError
	}
	
	m.SetState(StatePaused)
	return nil
}

// Resume 实现恢复功能
func (m *MockPlayerBackend) Resume() error {
	if m.resumeError != nil {
		return m.resumeError
	}
	
	m.SetState(StatePlaying)
	return nil
}

// Stop 实现停止功能
func (m *MockPlayerBackend) Stop() error {
	if m.stopError != nil {
		return m.stopError
	}
	
	m.SetState(StateStopped)
	m.SetPosition(0)
	return nil
}

// Seek 实现跳转功能
func (m *MockPlayerBackend) Seek(position time.Duration) error {
	if m.seekError != nil {
		return m.seekError
	}
	
	m.SetPosition(position)
	return nil
}

// IsAvailable 检查是否可用
func (m *MockPlayerBackend) IsAvailable() bool {
	return m.isAvailable
}

// TestPlaybackState 测试播放状态枚举
func TestPlaybackState(t *testing.T) {
	tests := []struct {
		state    PlaybackState
		expected string
	}{
		{StateStopped, "stopped"},
		{StatePlaying, "playing"},
		{StatePaused, "paused"},
		{StateBuffering, "buffering"},
		{StateError, "error"},
		{PlaybackState(999), "unknown"},
	}
	
	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			assert.Equal(t, test.expected, test.state.String())
		})
	}
}

// TestBaseBackend 测试基础播放器后端
func TestBaseBackend(t *testing.T) {
	capabilities := &BackendCapabilities{
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
	}
	
	backend := NewBaseBackend("test", "1.0.0", capabilities)
	
	// 测试基本信息
	assert.Equal(t, "test", backend.GetName())
	assert.Equal(t, "1.0.0", backend.GetVersion())
	assert.Equal(t, capabilities, backend.GetCapabilities())
	
	// 测试初始状态
	assert.Equal(t, StateStopped, backend.GetState())
	assert.False(t, backend.IsPlaying())
	
	volume, err := backend.GetVolume()
	assert.NoError(t, err)
	assert.Equal(t, 0.8, volume)
	
	position, err := backend.GetPosition()
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), position)
	
	duration, err := backend.GetDuration()
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), duration)
}

// TestBaseBackendInitialize 测试基础后端初始化
func TestBaseBackendInitialize(t *testing.T) {
	backend := NewBaseBackend("test", "1.0.0", &BackendCapabilities{})
	
	config := &BackendConfig{
		Name:          "test",
		Enabled:       true,
		DefaultVolume: 0.6,
	}
	
	ctx := context.Background()
	err := backend.Initialize(ctx, config)
	assert.NoError(t, err)
	
	// 检查配置是否应用
	volume, _ := backend.GetVolume()
	assert.Equal(t, 0.6, volume)
	
	// 测试清理
	err = backend.Cleanup()
	assert.NoError(t, err)
}

// TestBaseBackendVolumeControl 测试音量控制
func TestBaseBackendVolumeControl(t *testing.T) {
	backend := NewBaseBackend("test", "1.0.0", &BackendCapabilities{})
	
	// 测试设置有效音量
	err := backend.SetVolume(0.5)
	assert.NoError(t, err)
	
	volume, err := backend.GetVolume()
	assert.NoError(t, err)
	assert.Equal(t, 0.5, volume)
	
	// 测试设置无效音量
	err = backend.SetVolume(-0.1)
	assert.Error(t, err)
	
	err = backend.SetVolume(1.1)
	assert.Error(t, err)
	
	// 音量应该保持不变
	volume, _ = backend.GetVolume()
	assert.Equal(t, 0.5, volume)
}

// TestBaseBackendStateManagement 测试状态管理
func TestBaseBackendStateManagement(t *testing.T) {
	backend := NewBaseBackend("test", "1.0.0", &BackendCapabilities{})
	
	// 测试状态变化事件
	eventReceived := false
	var receivedEvent *Event
	
	handler := func(event *Event) {
		eventReceived = true
		receivedEvent = event
	}
	
	err := backend.AddEventHandler(EventStateChanged, handler)
	assert.NoError(t, err)
	
	// 改变状态
	backend.SetState(StatePlaying)
	
	// 等待事件处理
	time.Sleep(10 * time.Millisecond)
	
	assert.True(t, eventReceived)
	assert.NotNil(t, receivedEvent)
	assert.Equal(t, EventStateChanged, receivedEvent.Type)
	assert.Equal(t, "test", receivedEvent.Source)
	assert.Equal(t, "stopped", receivedEvent.Data["old_state"])
	assert.Equal(t, "playing", receivedEvent.Data["new_state"])
}

// TestBaseBackendPositionManagement 测试位置管理
func TestBaseBackendPositionManagement(t *testing.T) {
	backend := NewBaseBackend("test", "1.0.0", &BackendCapabilities{})
	
	// 测试位置变化事件
	eventReceived := false
	var receivedEvent *Event
	
	handler := func(event *Event) {
		eventReceived = true
		receivedEvent = event
	}
	
	err := backend.AddEventHandler(EventPositionChanged, handler)
	assert.NoError(t, err)
	
	// 设置位置
	position := 30 * time.Second
	backend.SetPosition(position)
	
	// 等待事件处理
	time.Sleep(10 * time.Millisecond)
	
	// 检查位置
	currentPosition, err := backend.GetPosition()
	assert.NoError(t, err)
	assert.Equal(t, position, currentPosition)
	
	// 检查事件
	assert.True(t, eventReceived)
	assert.NotNil(t, receivedEvent)
	assert.Equal(t, EventPositionChanged, receivedEvent.Type)
	assert.Equal(t, position.Seconds(), receivedEvent.Data["position"])
}

// TestBaseBackendEventHandlers 测试事件处理器管理
func TestBaseBackendEventHandlers(t *testing.T) {
	backend := NewBaseBackend("test", "1.0.0", &BackendCapabilities{})
	
	// 测试添加空处理器
	err := backend.AddEventHandler(EventStateChanged, nil)
	assert.Error(t, err)
	
	// 测试添加有效处理器
	handler1Called := false
	handler2Called := false
	
	handler1 := func(event *Event) {
		handler1Called = true
	}
	
	handler2 := func(event *Event) {
		handler2Called = true
	}
	
	err = backend.AddEventHandler(EventStateChanged, handler1)
	assert.NoError(t, err)
	
	err = backend.AddEventHandler(EventStateChanged, handler2)
	assert.NoError(t, err)
	
	// 触发事件
	backend.SetState(StatePlaying)
	
	// 等待事件处理
	time.Sleep(10 * time.Millisecond)
	
	// 检查两个处理器都被调用
	assert.True(t, handler1Called)
	assert.True(t, handler2Called)
}

// TestMockPlayerBackend 测试模拟播放器后端
func TestMockPlayerBackend(t *testing.T) {
	mock := NewMockPlayerBackend("mock", true)
	
	// 测试基本信息
	assert.Equal(t, "mock", mock.GetName())
	assert.Equal(t, "1.0.0", mock.GetVersion())
	assert.True(t, mock.IsAvailable())
	
	// 测试播放功能
	url := "http://example.com/test.mp3"
	err := mock.Play(url)
	assert.NoError(t, err)
	assert.Equal(t, url, mock.playURL)
	assert.Equal(t, StatePlaying, mock.GetState())
	assert.True(t, mock.IsPlaying())
	
	// 测试暂停功能
	err = mock.Pause()
	assert.NoError(t, err)
	assert.Equal(t, StatePaused, mock.GetState())
	assert.False(t, mock.IsPlaying())
	
	// 测试恢复功能
	err = mock.Resume()
	assert.NoError(t, err)
	assert.Equal(t, StatePlaying, mock.GetState())
	assert.True(t, mock.IsPlaying())
	
	// 测试跳转功能
	seekPos := 60 * time.Second
	err = mock.Seek(seekPos)
	assert.NoError(t, err)
	
	position, err := mock.GetPosition()
	assert.NoError(t, err)
	assert.Equal(t, seekPos, position)
	
	// 测试停止功能
	err = mock.Stop()
	assert.NoError(t, err)
	assert.Equal(t, StateStopped, mock.GetState())
	assert.False(t, mock.IsPlaying())
	
	position, _ = mock.GetPosition()
	assert.Equal(t, time.Duration(0), position)
}

// TestMockPlayerBackendErrors 测试模拟播放器后端错误处理
func TestMockPlayerBackendErrors(t *testing.T) {
	mock := NewMockPlayerBackend("mock", true)
	
	// 设置错误
	playError := assert.AnError
	mock.playError = playError
	
	// 测试播放错误
	err := mock.Play("test.mp3")
	assert.Error(t, err)
	assert.Equal(t, playError, err)
	
	// 清除播放错误，设置暂停错误
	mock.playError = nil
	mock.pauseError = assert.AnError
	
	// 播放应该成功
	err = mock.Play("test.mp3")
	assert.NoError(t, err)
	
	// 暂停应该失败
	err = mock.Pause()
	assert.Error(t, err)
}

// TestBackendCapabilities 测试后端能力描述
func TestBackendCapabilities(t *testing.T) {
	capabilities := &BackendCapabilities{
		SupportedFormats:   []string{"mp3", "wav", "flac"},
		SupportedPlatforms: []string{"linux", "darwin", "windows"},
		Features: map[string]bool{
			"seek":      true,
			"streaming": true,
			"volume":    true,
			"equalizer": false,
		},
		MaxVolume:        1.0,
		MinVolume:        0.0,
		SeekSupport:      true,
		StreamingSupport: true,
		Metadata: map[string]string{
			"author":  "test",
			"license": "MIT",
		},
	}
	
	backend := NewBaseBackend("test", "1.0.0", capabilities)
	resultCaps := backend.GetCapabilities()
	
	assert.Equal(t, capabilities, resultCaps)
	assert.Equal(t, []string{"mp3", "wav", "flac"}, resultCaps.SupportedFormats)
	assert.Equal(t, []string{"linux", "darwin", "windows"}, resultCaps.SupportedPlatforms)
	assert.True(t, resultCaps.Features["seek"])
	assert.True(t, resultCaps.Features["streaming"])
	assert.True(t, resultCaps.Features["volume"])
	assert.False(t, resultCaps.Features["equalizer"])
	assert.Equal(t, 1.0, resultCaps.MaxVolume)
	assert.Equal(t, 0.0, resultCaps.MinVolume)
	assert.True(t, resultCaps.SeekSupport)
	assert.True(t, resultCaps.StreamingSupport)
	assert.Equal(t, "test", resultCaps.Metadata["author"])
	assert.Equal(t, "MIT", resultCaps.Metadata["license"])
}

// TestBackendConfig 测试后端配置
func TestBackendConfig(t *testing.T) {
	config := &BackendConfig{
		Name:          "test",
		Enabled:       true,
		Priority:      5,
		BufferSize:    4096,
		SampleRate:    44100,
		Channels:      2,
		DefaultVolume: 0.8,
		Settings: map[string]interface{}{
			"custom_setting": "value",
			"numeric_setting": 42,
		},
	}
	
	assert.Equal(t, "test", config.Name)
	assert.True(t, config.Enabled)
	assert.Equal(t, 5, config.Priority)
	assert.Equal(t, 4096, config.BufferSize)
	assert.Equal(t, 44100, config.SampleRate)
	assert.Equal(t, 2, config.Channels)
	assert.Equal(t, 0.8, config.DefaultVolume)
	assert.Equal(t, "value", config.Settings["custom_setting"])
	assert.Equal(t, 42, config.Settings["numeric_setting"])
}

// BenchmarkBaseBackendStateChange 基准测试状态变化性能
func BenchmarkBaseBackendStateChange(b *testing.B) {
	backend := NewBaseBackend("test", "1.0.0", &BackendCapabilities{})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			backend.SetState(StatePlaying)
		} else {
			backend.SetState(StateStopped)
		}
	}
}

// BenchmarkBaseBackendVolumeChange 基准测试音量变化性能
func BenchmarkBaseBackendVolumeChange(b *testing.B) {
	backend := NewBaseBackend("test", "1.0.0", &BackendCapabilities{})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		volume := float64(i%100) / 100.0
		backend.SetVolume(volume)
	}
}