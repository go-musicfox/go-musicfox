package audio

import (
	"context"
	"testing"
	"time"

	event "github.com/go-musicfox/go-musicfox/v2/pkg/event"
	model "github.com/go-musicfox/go-musicfox/v2/pkg/model"
	core "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEventBus 模拟事件总线
type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Subscribe(eventType event.EventType, handler event.EventHandler, options ...event.SubscribeOption) (*event.Subscription, error) {
	args := m.Called(eventType, handler, options)
	return args.Get(0).(*event.Subscription), args.Error(1)
}

func (m *MockEventBus) SubscribeWithFilter(eventType event.EventType, handler event.EventHandler, filter event.EventFilter, options ...event.SubscribeOption) (*event.Subscription, error) {
	args := m.Called(eventType, handler, filter, options)
	return args.Get(0).(*event.Subscription), args.Error(1)
}

func (m *MockEventBus) Unsubscribe(subscriptionID string) error {
	args := m.Called(subscriptionID)
	return args.Error(0)
}

func (m *MockEventBus) UnsubscribeAll(eventType event.EventType) error {
	args := m.Called(eventType)
	return args.Error(0)
}

func (m *MockEventBus) Publish(ctx context.Context, event event.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventBus) PublishAsync(ctx context.Context, event event.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventBus) PublishSync(ctx context.Context, event event.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventBus) PublishWithPriority(ctx context.Context, event event.Event, priority event.EventPriority) error {
	args := m.Called(ctx, event, priority)
	return args.Error(0)
}

func (m *MockEventBus) RegisterEventType(eventType event.EventType) error {
	args := m.Called(eventType)
	return args.Error(0)
}

func (m *MockEventBus) UnregisterEventType(eventType event.EventType) error {
	args := m.Called(eventType)
	return args.Error(0)
}

func (m *MockEventBus) GetRegisteredEventTypes() []event.EventType {
	args := m.Called()
	return args.Get(0).([]event.EventType)
}

func (m *MockEventBus) GetSubscriptionCount(eventType event.EventType) int {
	args := m.Called(eventType)
	return args.Int(0)
}

func (m *MockEventBus) GetTotalSubscriptions() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockEventBus) GetEventStats() *event.EventStats {
	args := m.Called()
	return args.Get(0).(*event.EventStats)
}

func (m *MockEventBus) GetStats() *event.EventStats {
	args := m.Called()
	return args.Get(0).(*event.EventStats)
}

func (m *MockEventBus) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockEventBus) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockEventBus) IsRunning() bool {
	args := m.Called()
	return args.Bool(0)
}

// MockPlayerBackend 模拟播放器后端
type MockPlayerBackend struct {
	mock.Mock
	name             string
	version          string
	supportedFormats []string
	playing          bool
	position         time.Duration
	duration         time.Duration
	volume           float64
}

func NewMockPlayerBackend() *MockPlayerBackend {
	return &MockPlayerBackend{
		name:             "Mock Player",
		version:          "1.0.0",
		supportedFormats: []string{"mp3", "wav"},
		volume:           0.8,
	}
}

func (m *MockPlayerBackend) Initialize(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockPlayerBackend) Cleanup() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPlayerBackend) Play(url string) error {
	args := m.Called(url)
	m.playing = true
	m.position = 0
	return args.Error(0)
}

func (m *MockPlayerBackend) Pause() error {
	args := m.Called()
	m.playing = false
	return args.Error(0)
}

func (m *MockPlayerBackend) Resume() error {
	args := m.Called()
	m.playing = true
	return args.Error(0)
}

func (m *MockPlayerBackend) Stop() error {
	args := m.Called()
	m.playing = false
	m.position = 0
	return args.Error(0)
}

func (m *MockPlayerBackend) Seek(position time.Duration) error {
	args := m.Called(position)
	m.position = position
	return args.Error(0)
}

func (m *MockPlayerBackend) GetPosition() (time.Duration, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return time.Duration(0), args.Error(1)
	}
	return args.Get(0).(time.Duration), args.Error(1)
}

func (m *MockPlayerBackend) GetDuration() (time.Duration, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return time.Duration(0), args.Error(1)
	}
	return args.Get(0).(time.Duration), args.Error(1)
}

func (m *MockPlayerBackend) IsPlaying() bool {
	return m.playing
}

func (m *MockPlayerBackend) SetVolume(volume float64) error {
	args := m.Called(volume)
	m.volume = volume
	return args.Error(0)
}

func (m *MockPlayerBackend) GetVolume() (float64, error) {
	args := m.Called()
	return m.volume, args.Error(0)
}

func (m *MockPlayerBackend) GetName() string {
	return m.name
}

func (m *MockPlayerBackend) GetVersion() string {
	return m.version
}

func (m *MockPlayerBackend) GetSupportedFormats() []string {
	return m.supportedFormats
}

func (m *MockPlayerBackend) IsAvailable() bool {
	args := m.Called()
	return args.Bool(0)
}

// TestAudioPlugin_NewAudioPlugin 测试音频插件创建
func TestAudioPlugin_NewAudioPlugin(t *testing.T) {
	mockEventBus := &MockEventBus{}
	plugin := NewAudioPlugin(mockEventBus)

	assert.NotNil(t, plugin)
	assert.Equal(t, "Audio Processor Plugin", plugin.GetInfo().Name)
	assert.Equal(t, "1.0.0", plugin.GetInfo().Version)
	assert.NotNil(t, plugin.playerFactory)
	assert.NotNil(t, plugin.state)
	assert.Equal(t, 80, plugin.volume)
}

// MockPluginContext 模拟插件上下文
type MockPluginContext struct {
	config map[string]interface{}
}

func (m *MockPluginContext) GetContext() context.Context {
	return context.Background()
}

func (m *MockPluginContext) GetContainer() core.ServiceRegistry {
	return nil
}

func (m *MockPluginContext) GetEventBus() core.EventBus {
	return nil
}

func (m *MockPluginContext) GetServiceRegistry() core.ServiceRegistry {
	return nil
}

func (m *MockPluginContext) GetLogger() core.Logger {
	return nil
}

func (m *MockPluginContext) GetPluginConfig() core.PluginConfig {
	return &MockPluginConfig{config: m.config}
}

func (m *MockPluginContext) UpdateConfig(config core.PluginConfig) error {
	return nil
}

func (m *MockPluginContext) GetDataDir() string {
	return "/tmp"
}

func (m *MockPluginContext) GetTempDir() string {
	return "/tmp"
}

func (m *MockPluginContext) SendMessage(topic string, data interface{}) error {
	return nil
}

func (m *MockPluginContext) Subscribe(topic string, handler core.EventHandler) error {
	return nil
}

func (m *MockPluginContext) Unsubscribe(topic string, handler core.EventHandler) error {
	return nil
}

func (m *MockPluginContext) BroadcastMessage(message interface{}) error {
	return nil
}

func (m *MockPluginContext) GetResourceMonitor() *core.ResourceMonitor {
	return nil
}

func (m *MockPluginContext) GetSecurityManager() *core.SecurityManager {
	return nil
}

func (m *MockPluginContext) GetIsolationGroup() *core.IsolationGroup {
	return nil
}

func (m *MockPluginContext) Shutdown() error {
	return nil
}

// MockPluginConfig 模拟插件配置
type MockPluginConfig struct {
	config map[string]interface{}
}

func (m *MockPluginConfig) GetID() string {
	return "test-plugin"
}

func (m *MockPluginConfig) GetName() string {
	return "Test Plugin"
}

func (m *MockPluginConfig) GetVersion() string {
	return "1.0.0"
}

func (m *MockPluginConfig) GetEnabled() bool {
	return true
}

func (m *MockPluginConfig) GetPriority() core.PluginPriority {
	return core.PluginPriorityNormal
}

func (m *MockPluginConfig) GetDependencies() []string {
	return nil
}

func (m *MockPluginConfig) GetResourceLimits() *core.ResourceLimits {
	return nil
}

func (m *MockPluginConfig) GetSecurityConfig() *core.SecurityConfig {
	return nil
}

func (m *MockPluginConfig) GetCustomConfig() map[string]interface{} {
	return m.config
}

func (m *MockPluginConfig) Validate() error {
	return nil
}

// TestAudioPlugin_Initialize 测试插件初始化
func TestAudioPlugin_Initialize(t *testing.T) {
	mockEventBus := &MockEventBus{}
	plugin := NewAudioPlugin(mockEventBus)

	mockCtx := &MockPluginContext{
		config: map[string]interface{}{
			"backend": "beep",
			"volume":  90,
		},
	}

	err := plugin.Initialize(mockCtx)
	assert.NoError(t, err)
	assert.Equal(t, "beep", plugin.config.DefaultBackend)
	assert.Equal(t, 90, plugin.volume)
}

// TestAudioPlugin_Play 测试播放功能
func TestAudioPlugin_Play(t *testing.T) {
	mockEventBus := &MockEventBus{}
	mockEventBus.On("PublishAsync", mock.Anything, mock.Anything).Return(nil)

	plugin := NewAudioPlugin(mockEventBus)

	// 创建模拟播放器后端
	mockPlayer := NewMockPlayerBackend()
	mockPlayer.On("Play", mock.Anything).Return(nil)
	mockPlayer.On("GetPosition").Return(time.Duration(0), nil)
	mockPlayer.On("GetDuration").Return(time.Duration(0), nil)

	plugin.currentPlayer = mockPlayer

	song := &model.Song{
		ID:     "test-song-1",
		Title:  "Test Song",
		Artist: "Test Artist",
		URL:    "http://example.com/test.mp3",
	}

	err := plugin.Play(song)
	assert.NoError(t, err)
	assert.Equal(t, model.PlayStatusPlaying, plugin.state.Status)
	assert.Equal(t, song, plugin.state.CurrentSong)

	state := plugin.GetPlayState()
	assert.Equal(t, model.PlayStatusPlaying, state.Status)
	assert.Equal(t, song, state.CurrentSong)

	mockPlayer.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

// TestAudioPlugin_GetPlayState 测试获取播放状态
func TestAudioPlugin_GetPlayState(t *testing.T) {
	mockEventBus := &MockEventBus{}
	plugin := NewAudioPlugin(mockEventBus)

	song := &model.Song{
		ID:    "test-song",
		Title: "Test Song",
	}

	plugin.state.CurrentSong = song
	plugin.state.Status = model.PlayStatusPlaying
	plugin.volume = 80

	state := plugin.GetPlayState()
	assert.Equal(t, model.PlayStatusPlaying, state.Status)
	assert.Equal(t, song, state.CurrentSong)
}

// TestAudioPlugin_Pause 测试暂停功能
func TestAudioPlugin_Pause(t *testing.T) {
	mockEventBus := &MockEventBus{}
	plugin := NewAudioPlugin(mockEventBus)

	mockPlayer := NewMockPlayerBackend()
	mockPlayer.On("Pause").Return(nil)
	plugin.currentPlayer = mockPlayer

	err := plugin.Pause()
	assert.NoError(t, err)
	assert.Equal(t, model.PlayStatusPaused, plugin.state.Status)

	mockPlayer.AssertExpectations(t)
}

// TestAudioPlugin_Resume 测试恢复播放功能
func TestAudioPlugin_Resume(t *testing.T) {
	mockEventBus := &MockEventBus{}
	plugin := NewAudioPlugin(mockEventBus)

	mockPlayer := NewMockPlayerBackend()
	mockPlayer.On("Resume").Return(nil)
	plugin.currentPlayer = mockPlayer

	err := plugin.Resume()
	assert.NoError(t, err)
	assert.Equal(t, model.PlayStatusPlaying, plugin.state.Status)

	mockPlayer.AssertExpectations(t)
}

// TestAudioPlugin_Stop 测试停止功能
func TestAudioPlugin_Stop(t *testing.T) {
	mockEventBus := &MockEventBus{}
	plugin := NewAudioPlugin(mockEventBus)

	mockPlayer := NewMockPlayerBackend()
	mockPlayer.On("Stop").Return(nil)
	plugin.currentPlayer = mockPlayer

	err := plugin.StopPlayback()
	assert.NoError(t, err)
	assert.Equal(t, model.PlayStatusStopped, plugin.state.Status)
	assert.Equal(t, time.Duration(0), plugin.state.Position)

	mockPlayer.AssertExpectations(t)
}

// TestAudioPlugin_SetVolume 测试音量设置
func TestAudioPlugin_SetVolume(t *testing.T) {
	mockEventBus := &MockEventBus{}
	plugin := NewAudioPlugin(mockEventBus)

	mockPlayer := NewMockPlayerBackend()
	mockPlayer.On("SetVolume", 0.5).Return(nil)
	plugin.currentPlayer = mockPlayer

	err := plugin.SetVolume(50)
	assert.NoError(t, err)
	assert.Equal(t, 50, plugin.volume)

	mockPlayer.AssertExpectations(t)
}

// TestAudioPlugin_SetVolumeInvalidRange 测试无效音量范围
func TestAudioPlugin_SetVolumeInvalidRange(t *testing.T) {
	mockEventBus := &MockEventBus{}
	plugin := NewAudioPlugin(mockEventBus)

	// 测试负数音量
	err := plugin.SetVolume(-10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "volume must be between 0 and 100")

	// 测试超过100的音量
	err = plugin.SetVolume(150)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "volume must be between 0 and 100")
}

// TestAudioPlugin_SwitchBackend 测试播放器后端切换
func TestAudioPlugin_SwitchBackend(t *testing.T) {
	mockEventBus := &MockEventBus{}
	plugin := NewAudioPlugin(mockEventBus)

	// 模拟当前播放器
	currentPlayer := NewMockPlayerBackend()
	currentPlayer.On("Stop").Return(nil)
	currentPlayer.On("Cleanup").Return(nil)
	plugin.currentPlayer = currentPlayer

	// 注册新的播放器后端
	newPlayer := NewMockPlayerBackend()
	newPlayer.name = "New Mock Player"
	newPlayer.On("Initialize", mock.Anything).Return(nil)
	newPlayer.On("IsAvailable").Return(true)

	plugin.playerFactory.RegisterBackend("new_mock", func(config map[string]interface{}) (PlayerBackend, error) {
		return newPlayer, nil
	})

	err := plugin.SwitchBackend("new_mock")
	assert.NoError(t, err)
	assert.Equal(t, "new_mock", plugin.config.DefaultBackend)
	assert.Equal(t, newPlayer, plugin.currentPlayer)

	currentPlayer.AssertExpectations(t)
	newPlayer.AssertExpectations(t)
}

// TestAudioPlugin_GetAvailableBackends 测试获取可用后端
func TestAudioPlugin_GetAvailableBackends(t *testing.T) {
	mockEventBus := &MockEventBus{}
	plugin := NewAudioPlugin(mockEventBus)

	backends := plugin.GetAvailableBackends()
	assert.NotEmpty(t, backends)
	assert.Contains(t, backends, "beep") // beep应该总是可用的
}

// TestAudioPlugin_NoPlayerBackend 测试没有播放器后端的情况
func TestAudioPlugin_NoPlayerBackend(t *testing.T) {
	mockEventBus := &MockEventBus{}
	plugin := NewAudioPlugin(mockEventBus)
	plugin.currentPlayer = nil

	song := &model.Song{
		ID:  "test-song",
		URL: "http://example.com/test.mp3",
	}

	err := plugin.Play(song)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no player backend available")

	err = plugin.Pause()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no player backend available")

	err = plugin.Resume()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no player backend available")

	err = plugin.StopPlayback()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no player backend available")
}
