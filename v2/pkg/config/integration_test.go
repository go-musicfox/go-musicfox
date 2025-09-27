package config

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockEventBus 模拟事件总线
type MockEventBus struct {
	registeredTypes map[event.EventType]bool
	subscriptions   map[event.EventType][]*event.Subscription
	publishedEvents []event.Event
}

func NewMockEventBus() *MockEventBus {
	return &MockEventBus{
		registeredTypes: make(map[event.EventType]bool),
		subscriptions:   make(map[event.EventType][]*event.Subscription),
		publishedEvents: make([]event.Event, 0),
	}
}

func (m *MockEventBus) Subscribe(eventType event.EventType, handler event.EventHandler, options ...event.SubscribeOption) (*event.Subscription, error) {
	sub := &event.Subscription{
		ID:        generateID(),
		Type:      eventType,
		Handler:   handler,
		CreatedAt: time.Now(),
	}

	if m.subscriptions[eventType] == nil {
		m.subscriptions[eventType] = make([]*event.Subscription, 0)
	}
	m.subscriptions[eventType] = append(m.subscriptions[eventType], sub)

	return sub, nil
}

func (m *MockEventBus) SubscribeWithFilter(eventType event.EventType, handler event.EventHandler, filter event.EventFilter, options ...event.SubscribeOption) (*event.Subscription, error) {
	return m.Subscribe(eventType, handler, options...)
}

func (m *MockEventBus) Unsubscribe(subscriptionID string) error {
	for eventType, subs := range m.subscriptions {
		for i, sub := range subs {
			if sub.ID == subscriptionID {
				m.subscriptions[eventType] = append(subs[:i], subs[i+1:]...)
				return nil
			}
		}
	}
	return nil
}

func (m *MockEventBus) UnsubscribeAll(eventType event.EventType) error {
	delete(m.subscriptions, eventType)
	return nil
}

func (m *MockEventBus) Publish(ctx context.Context, event event.Event) error {
	m.publishedEvents = append(m.publishedEvents, event)
	return nil
}

func (m *MockEventBus) PublishAsync(ctx context.Context, event event.Event) error {
	return m.Publish(ctx, event)
}

func (m *MockEventBus) PublishSync(ctx context.Context, event event.Event) error {
	return m.Publish(ctx, event)
}

func (m *MockEventBus) PublishWithPriority(ctx context.Context, event event.Event, priority event.EventPriority) error {
	return m.Publish(ctx, event)
}

func (m *MockEventBus) RegisterEventType(eventType event.EventType) error {
	m.registeredTypes[eventType] = true
	return nil
}

func (m *MockEventBus) UnregisterEventType(eventType event.EventType) error {
	delete(m.registeredTypes, eventType)
	return nil
}

func (m *MockEventBus) GetRegisteredEventTypes() []event.EventType {
	types := make([]event.EventType, 0, len(m.registeredTypes))
	for eventType := range m.registeredTypes {
		types = append(types, eventType)
	}
	return types
}

func (m *MockEventBus) GetSubscriptionCount(eventType event.EventType) int {
	return len(m.subscriptions[eventType])
}

func (m *MockEventBus) GetTotalSubscriptions() int {
	total := 0
	for _, subs := range m.subscriptions {
		total += len(subs)
	}
	return total
}

func (m *MockEventBus) GetEventStats() *event.EventStats {
	return &event.EventStats{
		TotalEvents:      int64(len(m.publishedEvents)),
		TotalSubscribers: m.GetTotalSubscriptions(),
		EventCounts:      make(map[event.EventType]int64),
		ErrorCounts:      make(map[event.EventType]int64),
		LastEventTime:    time.Now(),
	}
}

func (m *MockEventBus) GetStats() *event.EventStats {
	return m.GetEventStats()
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

// GetPublishedEvents 获取已发布的事件（测试辅助方法）
func (m *MockEventBus) GetPublishedEvents() []event.Event {
	return m.publishedEvents
}

// ClearPublishedEvents 清空已发布的事件（测试辅助方法）
func (m *MockEventBus) ClearPublishedEvents() {
	m.publishedEvents = m.publishedEvents[:0]
}

// TestEventIntegration_Enable 测试事件集成启用
func TestEventIntegration_Enable(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewAdvancedManager(tempDir, "")
	mockEventBus := NewMockEventBus()

	integration := NewEventIntegration(mockEventBus, manager)
	assert.False(t, integration.IsEnabled())

	// 启用事件集成
	ctx := context.Background()
	err := integration.Enable(ctx)
	require.NoError(t, err)
	assert.True(t, integration.IsEnabled())

	// 验证事件类型已注册
	registeredTypes := mockEventBus.GetRegisteredEventTypes()
	assert.Contains(t, registeredTypes, EventConfigChanged)
	assert.Contains(t, registeredTypes, EventConfigReloaded)
	assert.Contains(t, registeredTypes, EventConfigVersionCreated)

	// 验证订阅已创建
	assert.True(t, len(integration.subscriptions) > 0)

	// 禁用事件集成
	err = integration.Disable()
	require.NoError(t, err)
	assert.False(t, integration.IsEnabled())
}

// TestEventIntegration_ConfigChangeEvents 测试配置变更事件
func TestEventIntegration_ConfigChangeEvents(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewAdvancedManager(tempDir, "")
	mockEventBus := NewMockEventBus()

	integration := NewEventIntegration(mockEventBus, manager)
	ctx := context.Background()
	err := integration.Enable(ctx)
	require.NoError(t, err)

	// 清空已发布的事件
	mockEventBus.ClearPublishedEvents()

	// 手动触发配置变更事件（模拟实际的配置变更）
	change := &ConfigChange{
		ID:        "test-change-1",
		Timestamp: time.Now(),
		Operation: "set",
		Key:       "test.key",
		OldValue:  nil,
		NewValue:  "test_value",
		User:      "test",
		Source:    "test",
	}

	// 直接调用配置变更回调来模拟配置变更
	if integration.configManager != nil {
		// 通过recordChange来触发回调
		manager.recordChange(change.Operation, change.Key, change.OldValue, change.NewValue, change.User, change.Source)
		// 手动触发回调
		for _, callback := range manager.callbacks {
			callback(change)
		}
	}

	// 等待事件处理
	time.Sleep(10 * time.Millisecond)

	// 验证配置变更事件已发布
	publishedEvents := mockEventBus.GetPublishedEvents()
	assert.True(t, len(publishedEvents) > 0)

	// 查找配置变更事件
	var configEvent *ConfigEvent
	for _, evt := range publishedEvents {
		if evt.GetType() == EventConfigChanged {
			configEvent = evt.(*ConfigEvent)
			break
		}
	}

	assert.NotNil(t, configEvent)
	if configEvent != nil {
		assert.Equal(t, "test.key", configEvent.Key)
		assert.Equal(t, "test_value", configEvent.NewValue)
	}
}

// TestEventIntegration_VersionEvents 测试版本事件
func TestEventIntegration_VersionEvents(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewAdvancedManager(tempDir, "")
	mockEventBus := NewMockEventBus()

	integration := NewEventIntegration(mockEventBus, manager)
	ctx := context.Background()
	err := integration.Enable(ctx)
	require.NoError(t, err)

	// 设置一些配置
	manager.k.Set("version.test", "value")

	// 创建版本快照
	version, err := manager.CreateSnapshot("Test version")
	require.NoError(t, err)

	// 发布版本事件
	err = integration.PublishVersionEvent(version)
	require.NoError(t, err)

	// 验证版本事件已发布
	publishedEvents := mockEventBus.GetPublishedEvents()
	var versionEvent *ConfigVersionEvent
	for _, evt := range publishedEvents {
		if evt.GetType() == EventConfigVersionCreated {
			versionEvent = evt.(*ConfigVersionEvent)
			break
		}
	}

	assert.NotNil(t, versionEvent)
	assert.Equal(t, version.ID, versionEvent.VersionID)
	assert.Equal(t, version.Description, versionEvent.Description)
	assert.Equal(t, version.Checksum, versionEvent.Checksum)
}

// TestEventIntegration_TransactionEvents 测试事务事件
func TestEventIntegration_TransactionEvents(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewAdvancedManager(tempDir, "")
	mockEventBus := NewMockEventBus()

	integration := NewEventIntegration(mockEventBus, manager)
	ctx := context.Background()
	err := integration.Enable(ctx)
	require.NoError(t, err)

	// 发布事务开始事件
	changes := map[string]interface{}{
		"tx.key1": "value1",
		"tx.key2": "value2",
	}
	deletions := []string{"tx.key3"}

	err = integration.PublishTransactionEvent(EventConfigTransactionStarted, "tx123", changes, deletions)
	require.NoError(t, err)

	// 验证事务事件已发布
	publishedEvents := mockEventBus.GetPublishedEvents()
	var transactionEvent *ConfigTransactionEvent
	for _, evt := range publishedEvents {
		if evt.GetType() == EventConfigTransactionStarted {
			transactionEvent = evt.(*ConfigTransactionEvent)
			break
		}
	}

	assert.NotNil(t, transactionEvent)
	assert.Equal(t, "tx123", transactionEvent.TransactionID)
	assert.Equal(t, changes, transactionEvent.Changes)
	assert.Equal(t, deletions, transactionEvent.Deletions)
}

// TestEventIntegration_SecurityEvents 测试安全事件
func TestEventIntegration_SecurityEvents(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewAdvancedManager(tempDir, "")
	mockEventBus := NewMockEventBus()

	integration := NewEventIntegration(mockEventBus, manager)
	ctx := context.Background()
	err := integration.Enable(ctx)
	require.NoError(t, err)

	// 发布安全事件
	auditData := map[string]interface{}{
		"vulnerabilities": []string{"Unencrypted sensitive data"},
		"score":           75,
	}

	err = integration.PublishSecurityEvent(EventConfigSecurityAudit, "warning", "Security audit completed", auditData)
	require.NoError(t, err)

	// 验证安全事件已发布
	publishedEvents := mockEventBus.GetPublishedEvents()
	var securityEvent *ConfigSecurityEvent
	for _, evt := range publishedEvents {
		if evt.GetType() == EventConfigSecurityAudit {
			securityEvent = evt.(*ConfigSecurityEvent)
			break
		}
	}

	assert.NotNil(t, securityEvent)
	assert.Equal(t, "warning", securityEvent.SecurityLevel)
	assert.Equal(t, auditData, securityEvent.GetData())
}

// TestEventIntegration_HotReloadEvents 测试热更新事件
func TestEventIntegration_HotReloadEvents(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewAdvancedManager(tempDir, "")
	mockEventBus := NewMockEventBus()

	integration := NewEventIntegration(mockEventBus, manager)
	ctx := context.Background()
	err := integration.Enable(ctx)
	require.NoError(t, err)

	// 发布热更新事件
	err = integration.PublishHotReloadEvent(EventConfigFileChanged, "/path/to/config.json", "write")
	require.NoError(t, err)

	// 验证热更新事件已发布
	publishedEvents := mockEventBus.GetPublishedEvents()
	var hotReloadEvent *ConfigEvent
	for _, evt := range publishedEvents {
		if evt.GetType() == EventConfigFileChanged {
			hotReloadEvent = evt.(*ConfigEvent)
			break
		}
	}

	assert.NotNil(t, hotReloadEvent)
	assert.Equal(t, "write", hotReloadEvent.Operation)
	assert.Equal(t, "hot_reload_manager", hotReloadEvent.GetSource())

	// 验证事件数据
	eventData, ok := hotReloadEvent.GetData().(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "/path/to/config.json", eventData["path"])
	assert.Equal(t, "write", eventData["operation"])
}

// TestEventIntegration_PluginConfigApplication 测试插件配置应用
func TestEventIntegration_PluginConfigApplication(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewAdvancedManager(tempDir, "")
	mockEventBus := NewMockEventBus()

	integration := NewEventIntegration(mockEventBus, manager)
	ctx := context.Background()
	err := integration.Enable(ctx)
	require.NoError(t, err)

	// 设置插件模板配置
	manager.k.Set("plugin_test_plugin_template", "test_template")

	// 创建模板文件
	templateFile := filepath.Join(tempDir, "test_template.json")
	templateData := map[string]interface{}{
		"plugin": map[string]interface{}{
			"name":    "${plugin_name}",
			"enabled": true,
			"config": map[string]interface{}{
				"timestamp": "${timestamp}",
			},
		},
	}
	templateJSON, err := json.MarshalIndent(templateData, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(templateFile, templateJSON, 0644))

	// 加载模板
	err = manager.LoadTemplate(templateFile)
	require.NoError(t, err)

	// 模拟插件加载事件
	pluginEvent := &event.PluginEvent{
		BaseEvent: &event.BaseEvent{
			ID:        generateID(),
			Type:      event.EventPluginLoaded,
			Source:    "plugin_manager",
			Timestamp: time.Now(),
		},
		PluginName: "test_plugin",
		PluginType: "audio",
	}

	// 处理插件事件
	err = integration.handlePluginEvent(ctx, pluginEvent)
	require.NoError(t, err)

	// 验证插件配置已应用
	assert.Equal(t, "test_plugin", manager.GetString("plugin.name"))
	assert.True(t, manager.GetBool("plugin.enabled"))
	assert.NotEmpty(t, manager.GetString("plugin.config.timestamp"))
}

// TestEventIntegration_SystemStartupHandling 测试系统启动处理
func TestEventIntegration_SystemStartupHandling(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewAdvancedManager(tempDir, "")
	mockEventBus := NewMockEventBus()

	integration := NewEventIntegration(mockEventBus, manager)
	ctx := context.Background()
	err := integration.Enable(ctx)
	require.NoError(t, err)

	// 设置一些敏感配置（触发安全审计）
	manager.k.Set("database.password", "secret123")
	manager.k.Set("api.key", "apikey456")

	// 清空已发布的事件
	mockEventBus.ClearPublishedEvents()

	// 模拟系统启动事件
	systemEvent := &event.SystemEvent{
		BaseEvent: &event.BaseEvent{
			ID:        generateID(),
			Type:      event.EventSystemStartup,
			Source:    "system",
			Timestamp: time.Now(),
		},
		Message: "System startup completed",
		Level:   "info",
	}

	// 处理系统事件
	err = integration.handleSystemEvent(ctx, systemEvent)
	require.NoError(t, err)

	// 验证启动快照已创建
	history, err := manager.GetVersionHistory()
	require.NoError(t, err)
	assert.True(t, len(history) > 0)

	// 查找启动快照
	var startupSnapshot *ConfigVersion
	for _, version := range history {
		if version.Description == "System startup snapshot" {
			startupSnapshot = version
			break
		}
	}
	assert.NotNil(t, startupSnapshot)

	// 验证安全审计事件已发布
	publishedEvents := mockEventBus.GetPublishedEvents()
	var securityAuditEvent *ConfigSecurityEvent
	for _, evt := range publishedEvents {
		if evt.GetType() == EventConfigSecurityAudit {
			securityAuditEvent = evt.(*ConfigSecurityEvent)
			break
		}
	}

	assert.NotNil(t, securityAuditEvent)
	assert.Equal(t, "warning", securityAuditEvent.SecurityLevel)
	assert.True(t, len(securityAuditEvent.Vulnerabilities) > 0)
}

// TestEventIntegration_EventStats 测试事件统计
func TestEventIntegration_EventStats(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewAdvancedManager(tempDir, "")
	mockEventBus := NewMockEventBus()

	integration := NewEventIntegration(mockEventBus, manager)
	ctx := context.Background()
	err := integration.Enable(ctx)
	require.NoError(t, err)

	// 发布一些配置相关事件
	err = integration.PublishHotReloadEvent(EventConfigFileChanged, "/path/to/config.json", "write")
	require.NoError(t, err)

	version, err := manager.CreateSnapshot("Test version")
	require.NoError(t, err)
	err = integration.PublishVersionEvent(version)
	require.NoError(t, err)

	err = integration.PublishSecurityEvent(EventConfigSecurityAudit, "info", "Audit completed", nil)
	require.NoError(t, err)

	// 模拟非配置事件
	mockEventBus.Publish(ctx, &event.BaseEvent{
		Type:   event.EventPlayerPlay,
		Source: "player",
	})

	// 获取配置相关事件统计
	configStats := integration.GetEventStats()

	// 验证只包含配置相关事件
	for eventType := range configStats {
		assert.True(t, integration.isConfigEvent(eventType), "Event type %s should be config-related", eventType)
	}

	// 验证非配置事件不包含在内
	_, exists := configStats[event.EventPlayerPlay]
	assert.False(t, exists)
}