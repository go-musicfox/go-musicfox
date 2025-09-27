package playlist

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/event"
	"github.com/go-musicfox/go-musicfox/v2/pkg/model"
	plugin "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
)

// MockEventBus 模拟事件总线
type MockEventBus struct {
	publishedEvents []event.Event
	subscriptions   map[event.EventType][]event.EventHandler
}

func NewMockEventBus() *MockEventBus {
	return &MockEventBus{
		publishedEvents: make([]event.Event, 0),
		subscriptions:   make(map[event.EventType][]event.EventHandler),
	}
}

func (m *MockEventBus) Publish(ctx context.Context, e event.Event) error {
	m.publishedEvents = append(m.publishedEvents, e)
	return nil
}

func (m *MockEventBus) Subscribe(eventType event.EventType, handler event.EventHandler, options ...event.SubscribeOption) (*event.Subscription, error) {
	if m.subscriptions[eventType] == nil {
		m.subscriptions[eventType] = make([]event.EventHandler, 0)
	}
	m.subscriptions[eventType] = append(m.subscriptions[eventType], handler)
	return &event.Subscription{ID: "mock-subscription-id"}, nil
}

func (m *MockEventBus) Unsubscribe(subscriptionID string) error {
	return nil
}

func (m *MockEventBus) GetEventStats() *event.EventStats {
	return &event.EventStats{}
}

func (m *MockEventBus) GetRegisteredEventTypes() []event.EventType {
	return []event.EventType{}
}

func (m *MockEventBus) GetStats() *event.EventStats {
	return &event.EventStats{}
}

func (m *MockEventBus) GetSubscriptionCount(eventType event.EventType) int {
	return 0
}

func (m *MockEventBus) GetTotalSubscriptions() int {
	return 0
}

func (m *MockEventBus) SubscribeWithFilter(eventType event.EventType, handler event.EventHandler, filter event.EventFilter, options ...event.SubscribeOption) (*event.Subscription, error) {
	return &event.Subscription{}, nil
}

func (m *MockEventBus) UnsubscribeAll(eventType event.EventType) error {
	return nil
}

func (m *MockEventBus) PublishAsync(ctx context.Context, e event.Event) error {
	return m.Publish(ctx, e)
}

func (m *MockEventBus) PublishSync(ctx context.Context, e event.Event) error {
	return m.Publish(ctx, e)
}

func (m *MockEventBus) PublishWithPriority(ctx context.Context, e event.Event, priority event.EventPriority) error {
	return m.Publish(ctx, e)
}

func (m *MockEventBus) RegisterEventType(eventType event.EventType) error {
	return nil
}

func (m *MockEventBus) UnregisterEventType(eventType event.EventType) error {
	return nil
}

func (m *MockEventBus) Start(ctx context.Context) error {
	return nil
}

func (m *MockEventBus) Stop(ctx context.Context) error {
	return nil
}

func (m *MockEventBus) IsRunning() bool {
	return true
}

func (m *MockEventBus) GetPublishedEvents() []event.Event {
	return m.publishedEvents
}

func (m *MockEventBus) ClearEvents() {
	m.publishedEvents = make([]event.Event, 0)
}

// MockPluginContext 模拟插件上下文
type MockPluginContext struct {
	services map[string]interface{}
}

func NewMockPluginContext(eventBus event.EventBus) *MockPluginContext {
	return &MockPluginContext{
		services: map[string]interface{}{
			"eventBus": eventBus, // 修正服务名称
		},
	}
}

func (m *MockPluginContext) GetService(name string) interface{} {
	return m.services[name]
}

func (m *MockPluginContext) GetServiceRegistry() plugin.ServiceRegistry {
	return &MockServiceRegistry{services: m.services}
}

func (m *MockPluginContext) GetContext() context.Context {
	return context.Background()
}

func (m *MockPluginContext) GetConfig() map[string]interface{} {
	return make(map[string]interface{})
}

func (m *MockPluginContext) GetLogger() plugin.Logger {
	return nil
}

func (m *MockPluginContext) GetContainer() plugin.ServiceRegistry {
	return nil
}

func (m *MockPluginContext) GetEventBus() plugin.EventBus {
	return nil
}

func (m *MockPluginContext) GetPluginConfig() plugin.PluginConfig {
	return &plugin.BasePluginConfig{
		ID:       "test",
		Name:     "test",
		Enabled:  true,
		Version:  "1.0.0",
		Priority: plugin.PluginPriorityNormal,
	}
}

func (m *MockPluginContext) UpdateConfig(config plugin.PluginConfig) error {
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

func (m *MockPluginContext) Subscribe(topic string, handler plugin.EventHandler) error {
	return nil
}

func (m *MockPluginContext) Unsubscribe(topic string, handler plugin.EventHandler) error {
	return nil
}

func (m *MockPluginContext) BroadcastMessage(message interface{}) error {
	return nil
}

func (m *MockPluginContext) GetResourceMonitor() *plugin.ResourceMonitor {
	return nil
}

func (m *MockPluginContext) GetSecurityManager() *plugin.SecurityManager {
	return nil
}

func (m *MockPluginContext) GetIsolationGroup() *plugin.IsolationGroup {
	return nil
}

func (m *MockPluginContext) Shutdown() error {
	return nil
}

// MockServiceRegistry 模拟服务注册表
type MockServiceRegistry struct {
	services map[string]interface{}
}

func (m *MockServiceRegistry) RegisterService(name string, service interface{}) error {
	m.services[name] = service
	return nil
}

func (m *MockServiceRegistry) GetService(name string) (interface{}, error) {
	if service, ok := m.services[name]; ok {
		return service, nil
	}
	return nil, fmt.Errorf("service not found: %s", name)
}

func (m *MockServiceRegistry) UnregisterService(name string) error {
	delete(m.services, name)
	return nil
}

func (m *MockServiceRegistry) ListServices() []string {
	keys := make([]string, 0, len(m.services))
	for k := range m.services {
		keys = append(keys, k)
	}
	return keys
}

func (m *MockServiceRegistry) HasService(name string) bool {
	_, ok := m.services[name]
	return ok
}

// 测试辅助函数
func createTestSong(id, title, artist string) *model.Song {
	return &model.Song{
		ID:        id,
		Title:     title,
		Artist:    artist,
		Album:     "Test Album",
		Source:    "test",
		URL:       "http://example.com/" + id + ".mp3",
		Duration:  3 * time.Minute,
		Quality:   model.QualityHigh,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func setupTestPlugin() (*PlaylistPluginImpl, *MockEventBus) {
	mockEventBus := NewMockEventBus()
	plugin := NewPlaylistPlugin()
	ctx := NewMockPluginContext(mockEventBus)
	
	err := plugin.Initialize(ctx)
	if err != nil {
		panic("Failed to initialize plugin: " + err.Error())
	}
	
	return plugin, mockEventBus
}

// TestNewPlaylistPlugin 测试插件创建
func TestNewPlaylistPlugin(t *testing.T) {
	plugin := NewPlaylistPlugin()
	
	if plugin == nil {
		t.Fatal("Plugin should not be nil")
	}
	
	info := plugin.GetInfo()
	if info.ID != "playlist" {
		t.Errorf("Expected plugin ID 'playlist', got '%s'", info.ID)
	}
	
	if info.Name != "Playlist Plugin" {
		t.Errorf("Expected plugin name 'Playlist Plugin', got '%s'", info.Name)
	}
	
	capabilities := plugin.GetCapabilities()
	expectedCapabilities := []string{
		"playlist_management",
		"queue_management",
		"history_management",
		"play_mode_management",
		"event_integration",
	}
	
	if len(capabilities) != len(expectedCapabilities) {
		t.Errorf("Expected %d capabilities, got %d", len(expectedCapabilities), len(capabilities))
	}
}

// TestPluginLifecycle 测试插件生命周期
func TestPluginLifecycle(t *testing.T) {
	plugin, mockEventBus := setupTestPlugin()
	
	// 测试启动
	err := plugin.Start()
	if err != nil {
		t.Errorf("Failed to start plugin: %v", err)
	}
	
	// 检查启动事件
	events := mockEventBus.GetPublishedEvents()
	if len(events) == 0 {
		t.Error("Expected start event to be published")
	}
	
	mockEventBus.ClearEvents()
	
	// 测试停止
	err = plugin.Stop()
	if err != nil {
		t.Errorf("Failed to stop plugin: %v", err)
	}
	
	// 检查停止事件
	events = mockEventBus.GetPublishedEvents()
	if len(events) == 0 {
		t.Error("Expected stop event to be published")
	}
	
	// 测试清理
	err = plugin.Cleanup()
	if err != nil {
		t.Errorf("Failed to cleanup plugin: %v", err)
	}
}

// TestHealthCheck 测试健康检查
func TestHealthCheck(t *testing.T) {
	plugin, _ := setupTestPlugin()
	
	err := plugin.HealthCheck()
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}
	
	// 测试损坏状态
	plugin.playlists = nil
	err = plugin.HealthCheck()
	if err == nil {
		t.Error("Expected health check to fail with nil playlists")
	}
}

// TestConfigValidationAndUpdate 测试配置验证和更新
func TestConfigValidationAndUpdate(t *testing.T) {
	plugin, _ := setupTestPlugin()
	
	// 测试有效配置
	validConfig := map[string]interface{}{
		"max_history": 50,
	}
	
	err := plugin.ValidateConfig(validConfig)
	if err != nil {
		t.Errorf("Valid config should pass validation: %v", err)
	}
	
	err = plugin.UpdateConfig(validConfig)
	if err != nil {
		t.Errorf("Failed to update config: %v", err)
	}
	
	if plugin.maxHistory != 50 {
		t.Errorf("Expected max_history to be 50, got %d", plugin.maxHistory)
	}
	
	// 测试无效配置
	invalidConfig := map[string]interface{}{
		"max_history": -1,
	}
	
	err = plugin.ValidateConfig(invalidConfig)
	if err == nil {
		t.Error("Invalid config should fail validation")
	}
}

// TestGetMetrics 测试指标获取
func TestGetMetrics(t *testing.T) {
	plugin, _ := setupTestPlugin()
	
	// 添加一些测试数据
	ctx := context.Background()
	playlist, _ := plugin.CreatePlaylist(ctx, "Test Playlist", "Test Description")
	song := createTestSong("song1", "Test Song", "Test Artist")
	plugin.AddSong(ctx, playlist.ID, song)
	plugin.AddToQueue(ctx, song)
	plugin.AddToHistory(ctx, song)
	
	metrics, err := plugin.GetMetrics()
	if err != nil {
		t.Errorf("Failed to get metrics: %v", err)
	}
	
	if metrics.CustomMetrics["playlists_count"] != 1 {
		t.Errorf("Expected playlists_count to be 1, got %v", metrics.CustomMetrics["playlists_count"])
	}
	
	if metrics.CustomMetrics["queue_length"] != 1 {
		t.Errorf("Expected queue_length to be 1, got %v", metrics.CustomMetrics["queue_length"])
	}
	
	if metrics.CustomMetrics["history_length"] != 1 {
		t.Errorf("Expected history_length to be 1, got %v", metrics.CustomMetrics["history_length"])
	}
}

// TestHandleEvent 测试事件处理
func TestHandleEvent(t *testing.T) {
	plugin, _ := setupTestPlugin()
	
	// 创建测试歌曲和队列
	ctx := context.Background()
	song := createTestSong("song1", "Test Song", "Test Artist")
	plugin.AddToQueue(ctx, song)
	
	// 创建播放器事件
	playerEvent := &event.BaseEvent{
		ID:        "test-event",
		Type:      event.EventPlayerSongChanged,
		Source:    "test-player",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"song_id": "song1",
		},
	}
	
	err := plugin.HandleEvent(playerEvent)
	if err != nil {
		t.Errorf("Failed to handle event: %v", err)
	}
	
	// 检查歌曲是否被添加到历史记录
	history, err := plugin.GetHistory(ctx, 10)
	if err != nil {
		t.Errorf("Failed to get history: %v", err)
	}
	
	if len(history) != 1 {
		t.Errorf("Expected 1 song in history, got %d", len(history))
	}
	
	if history[0].ID != "song1" {
		t.Errorf("Expected song ID 'song1' in history, got '%s'", history[0].ID)
	}
}

// BenchmarkCreatePlaylist 播放列表创建性能测试
func BenchmarkCreatePlaylist(b *testing.B) {
	plugin, _ := setupTestPlugin()
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := "Playlist " + string(rune(i))
		_, err := plugin.CreatePlaylist(ctx, name, "Benchmark playlist")
		if err != nil {
			b.Errorf("Failed to create playlist: %v", err)
		}
	}
}

// BenchmarkAddToQueuePlugin 队列添加性能测试
func BenchmarkAddToQueuePlugin(b *testing.B) {
	plugin, _ := setupTestPlugin()
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		song := createTestSong(string(rune(i)), "Song "+string(rune(i)), "Artist")
		err := plugin.AddToQueue(ctx, song)
		if err != nil {
			b.Errorf("Failed to add to queue: %v", err)
		}
	}
}

// BenchmarkGetNextSong 获取下一首歌曲性能测试
func BenchmarkGetNextSong(b *testing.B) {
	plugin, _ := setupTestPlugin()
	ctx := context.Background()
	
	// 创建测试队列
	songs := make([]*model.Song, 1000)
	for i := 0; i < 1000; i++ {
		songs[i] = createTestSong(string(rune(i)), "Song "+string(rune(i)), "Artist")
	}
	plugin.SetCurrentQueue(ctx, songs)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		currentSong := songs[i%len(songs)]
		_, err := plugin.GetNextSong(ctx, currentSong)
		if err != nil {
			b.Errorf("Failed to get next song: %v", err)
		}
	}
}