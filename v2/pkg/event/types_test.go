package event

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventTypeRegistry_RegisterEventType(t *testing.T) {
	registry := NewEventTypeRegistry()

	// 测试注册新事件类型
	customEventType := EventType("custom.test.event")
	info := EventTypeInfo{
		Category:    "custom",
		Description: "Custom test event",
		Schema:      "{\"type\": \"object\"}",
	}

	err := registry.RegisterEventType(customEventType, info)
	assert.NoError(t, err)

	// 验证事件类型已注册
	assert.True(t, registry.IsRegistered(customEventType))

	// 测试重复注册应该失败
	err = registry.RegisterEventType(customEventType, info)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestEventTypeRegistry_UnregisterEventType(t *testing.T) {
	registry := NewEventTypeRegistry()

	// 注册一个事件类型
	customEventType := EventType("custom.test.event")
	info := EventTypeInfo{
		Category:    "custom",
		Description: "Custom test event",
	}

	err := registry.RegisterEventType(customEventType, info)
	require.NoError(t, err)

	// 验证已注册
	assert.True(t, registry.IsRegistered(customEventType))

	// 注销事件类型
	err = registry.UnregisterEventType(customEventType)
	assert.NoError(t, err)

	// 验证已注销
	assert.False(t, registry.IsRegistered(customEventType))

	// 测试注销不存在的事件类型应该失败
	err = registry.UnregisterEventType(EventType("non.existent.event"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not registered")
}

func TestEventTypeRegistry_GetEventTypeInfo(t *testing.T) {
	registry := NewEventTypeRegistry()

	// 测试获取预定义事件类型信息
	info, exists := registry.GetEventTypeInfo(EventPlayerPlay)
	assert.True(t, exists)
	assert.Equal(t, EventPlayerPlay, info.Type)
	assert.Equal(t, "player", info.Category)
	assert.Equal(t, "Player started playing", info.Description)
	assert.False(t, info.RegisteredAt.IsZero())

	// 测试获取不存在的事件类型信息
	info, exists = registry.GetEventTypeInfo(EventType("non.existent.event"))
	assert.False(t, exists)
	assert.Equal(t, EventType(""), info.Type)
}

func TestEventTypeRegistry_GetAllEventTypes(t *testing.T) {
	registry := NewEventTypeRegistry()

	// 获取所有事件类型
	allTypes := registry.GetAllEventTypes()
	assert.Greater(t, len(allTypes), 0)

	// 验证包含预定义事件类型
	found := false
	for _, info := range allTypes {
		if info.Type == EventPlayerPlay {
			found = true
			break
		}
	}
	assert.True(t, found, "Should contain predefined event types")

	// 注册自定义事件类型
	customEventType := EventType("custom.test.event")
	customInfo := EventTypeInfo{
		Category:    "custom",
		Description: "Custom test event",
	}
	err := registry.RegisterEventType(customEventType, customInfo)
	require.NoError(t, err)

	// 验证自定义事件类型也被包含
	allTypesAfter := registry.GetAllEventTypes()
	assert.Equal(t, len(allTypes)+1, len(allTypesAfter))

	found = false
	for _, info := range allTypesAfter {
		if info.Type == customEventType {
			found = true
			break
		}
	}
	assert.True(t, found, "Should contain custom event type")
}

func TestEventTypeRegistry_GetEventTypesByCategory(t *testing.T) {
	registry := NewEventTypeRegistry()

	// 获取播放器类别的事件类型
	playerTypes := registry.GetEventTypesByCategory("player")
	assert.Greater(t, len(playerTypes), 0)

	// 验证所有返回的事件类型都属于播放器类别
	for _, info := range playerTypes {
		assert.Equal(t, "player", info.Category)
	}

	// 验证包含预期的播放器事件类型
	foundPlay := false
	foundPause := false
	for _, info := range playerTypes {
		if info.Type == EventPlayerPlay {
			foundPlay = true
		}
		if info.Type == EventPlayerPause {
			foundPause = true
		}
	}
	assert.True(t, foundPlay, "Should contain EventPlayerPlay")
	assert.True(t, foundPause, "Should contain EventPlayerPause")

	// 测试不存在的类别
	nonExistentTypes := registry.GetEventTypesByCategory("non_existent")
	assert.Empty(t, nonExistentTypes)
}

func TestDefaultEventValidator_Validate(t *testing.T) {
	registry := NewEventTypeRegistry()
	validator := NewDefaultEventValidator(registry)

	// 测试有效事件
	validEvent := &BaseEvent{
		ID:        "test-event-1",
		Type:      EventPlayerPlay,
		Source:    "test-source",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"test": "data"},
	}

	err := validator.Validate(validEvent)
	assert.NoError(t, err)

	// 测试nil事件
	err = validator.Validate(nil)
	assert.Error(t, err)
	assert.IsType(t, &ValidationError{}, err)
	validationErr := err.(*ValidationError)
	assert.Equal(t, "event", validationErr.Field)
	assert.Contains(t, validationErr.Message, "cannot be nil")
}

func TestDefaultEventValidator_ValidateEventType(t *testing.T) {
	registry := NewEventTypeRegistry()
	validator := NewDefaultEventValidator(registry)

	// 测试空事件类型
	emptyTypeEvent := &BaseEvent{
		ID:        "test-event-1",
		Type:      "",
		Source:    "test-source",
		Timestamp: time.Now(),
	}

	err := validator.Validate(emptyTypeEvent)
	assert.Error(t, err)
	assert.IsType(t, &ValidationError{}, err)
	validationErr := err.(*ValidationError)
	assert.Equal(t, "type", validationErr.Field)
	assert.Contains(t, validationErr.Message, "cannot be empty")

	// 测试未注册的事件类型
	unregisteredEvent := &BaseEvent{
		ID:        "test-event-1",
		Type:      "unregistered.event.type",
		Source:    "test-source",
		Timestamp: time.Now(),
	}

	err = validator.Validate(unregisteredEvent)
	assert.Error(t, err)
	assert.IsType(t, &ValidationError{}, err)
	validationErr = err.(*ValidationError)
	assert.Equal(t, "type", validationErr.Field)
	assert.Contains(t, validationErr.Message, "not registered")
	assert.Equal(t, EventType("unregistered.event.type"), validationErr.Value)

	// 测试通配符事件类型（应该被允许）
	wildcardEvent := &BaseEvent{
		ID:        "test-event-1",
		Type:      EventTypeAll,
		Source:    "test-source",
		Timestamp: time.Now(),
	}

	err = validator.Validate(wildcardEvent)
	assert.NoError(t, err)
}

func TestDefaultEventValidator_ValidateEventID(t *testing.T) {
	registry := NewEventTypeRegistry()
	validator := NewDefaultEventValidator(registry)

	// 测试空事件ID
	emptyIDEvent := &BaseEvent{
		ID:        "",
		Type:      EventPlayerPlay,
		Source:    "test-source",
		Timestamp: time.Now(),
	}

	err := validator.Validate(emptyIDEvent)
	assert.Error(t, err)
	assert.IsType(t, &ValidationError{}, err)
	validationErr := err.(*ValidationError)
	assert.Equal(t, "id", validationErr.Field)
	assert.Contains(t, validationErr.Message, "cannot be empty")
}

func TestDefaultEventValidator_ValidateEventSource(t *testing.T) {
	registry := NewEventTypeRegistry()
	validator := NewDefaultEventValidator(registry)

	// 测试空事件源
	emptySourceEvent := &BaseEvent{
		ID:        "test-event-1",
		Type:      EventPlayerPlay,
		Source:    "",
		Timestamp: time.Now(),
	}

	err := validator.Validate(emptySourceEvent)
	assert.Error(t, err)
	assert.IsType(t, &ValidationError{}, err)
	validationErr := err.(*ValidationError)
	assert.Equal(t, "source", validationErr.Field)
	assert.Contains(t, validationErr.Message, "cannot be empty")
}

func TestDefaultEventValidator_ValidateEventTimestamp(t *testing.T) {
	registry := NewEventTypeRegistry()
	validator := NewDefaultEventValidator(registry)

	// 测试零时间戳
	zeroTimestampEvent := &BaseEvent{
		ID:        "test-event-1",
		Type:      EventPlayerPlay,
		Source:    "test-source",
		Timestamp: time.Time{}, // 零值时间
	}

	err := validator.Validate(zeroTimestampEvent)
	assert.Error(t, err)
	assert.IsType(t, &ValidationError{}, err)
	validationErr := err.(*ValidationError)
	assert.Equal(t, "timestamp", validationErr.Field)
	assert.Contains(t, validationErr.Message, "cannot be zero")
}

func TestValidationError_Error(t *testing.T) {
	validationErr := &ValidationError{
		Field:   "test_field",
		Message: "test message",
		Value:   "test_value",
	}

	errorMsg := validationErr.Error()
	assert.Contains(t, errorMsg, "test_field")
	assert.Contains(t, errorMsg, "test message")
}

func TestIsValidEventType(t *testing.T) {
	tests := []struct {
		eventType EventType
		expected  bool
	}{
		{"", false},                    // 空事件类型
		{"valid.event.type", true},     // 有效事件类型
		{EventTypeAll, true},           // 通配符事件类型
		{EventPlayerPlay, true},        // 预定义事件类型
		{"a", true},                    // 单字符事件类型
	}

	for _, test := range tests {
		result := IsValidEventType(test.eventType)
		assert.Equal(t, test.expected, result, "EventType: %s", test.eventType)
	}
}

func TestGetEventCategory(t *testing.T) {
	tests := []struct {
		eventType EventType
		expected  string
	}{
		{"player.play", "player"},
		{"system.startup", "system"},
		{"plugin.loaded", "plugin"},
		{"single", "single"},
		{"", "unknown"},
		{"no.category.here.just.dots", "no"},
	}

	for _, test := range tests {
		result := GetEventCategory(test.eventType)
		assert.Equal(t, test.expected, result, "EventType: %s", test.eventType)
	}
}

func TestBaseEvent_Interface(t *testing.T) {
	now := time.Now()
	event := &BaseEvent{
		ID:        "test-event-1",
		Type:      EventPlayerPlay,
		Source:    "test-source",
		Timestamp: now,
		Data:      map[string]interface{}{"key": "value"},
	}

	// 测试Event接口方法
	assert.Equal(t, "test-event-1", event.GetID())
	assert.Equal(t, EventPlayerPlay, event.GetType())
	assert.Equal(t, "test-source", event.GetSource())
	assert.Equal(t, now, event.GetTimestamp())
	assert.Equal(t, map[string]interface{}{"key": "value"}, event.GetData())
}

func TestPriorityEvent(t *testing.T) {
	now := time.Now()
	baseEvent := &BaseEvent{
		ID:        "priority-test-1",
		Type:      EventPlayerPlay,
		Source:    "test-source",
		Timestamp: now,
		Data:      map[string]interface{}{"priority": "high"},
	}

	priorityEvent := &PriorityEvent{
		BaseEvent: baseEvent,
		Priority:  PriorityHigh,
	}

	// 测试继承的Event接口方法
	assert.Equal(t, "priority-test-1", priorityEvent.GetID())
	assert.Equal(t, EventPlayerPlay, priorityEvent.GetType())
	assert.Equal(t, "test-source", priorityEvent.GetSource())
	assert.Equal(t, now, priorityEvent.GetTimestamp())
	assert.Equal(t, map[string]interface{}{"priority": "high"}, priorityEvent.GetData())

	// 测试优先级
	assert.Equal(t, PriorityHigh, priorityEvent.Priority)
}

func TestPluginEvent(t *testing.T) {
	now := time.Now()
	baseEvent := &BaseEvent{
		ID:        "plugin-test-1",
		Type:      EventPluginLoaded,
		Source:    "plugin-manager",
		Timestamp: now,
		Data:      map[string]interface{}{"plugin_id": "test-plugin"},
	}

	pluginEvent := &PluginEvent{
		BaseEvent:  baseEvent,
		PluginName: "test-plugin",
		PluginType: "audio",
		Path:       "/plugins/test-plugin.so",
	}

	// 测试继承的Event接口方法
	assert.Equal(t, "plugin-test-1", pluginEvent.GetID())
	assert.Equal(t, EventPluginLoaded, pluginEvent.GetType())
	assert.Equal(t, "plugin-manager", pluginEvent.GetSource())
	assert.Equal(t, now, pluginEvent.GetTimestamp())

	// 测试插件特定字段
	assert.Equal(t, "test-plugin", pluginEvent.PluginName)
	assert.Equal(t, "audio", pluginEvent.PluginType)
	assert.Equal(t, "/plugins/test-plugin.so", pluginEvent.Path)
	assert.Empty(t, pluginEvent.Error)
}

func TestPlayerEvent(t *testing.T) {
	now := time.Now()
	baseEvent := &BaseEvent{
		ID:        "player-test-1",
		Type:      EventPlayerPlay,
		Source:    "audio-player",
		Timestamp: now,
		Data:      map[string]interface{}{"song_id": "123"},
	}

	playerEvent := &PlayerEvent{
		BaseEvent: baseEvent,
		SongID:    "123",
		Position:  time.Second * 30,
		Volume:    0.8,
	}

	// 测试继承的Event接口方法
	assert.Equal(t, "player-test-1", playerEvent.GetID())
	assert.Equal(t, EventPlayerPlay, playerEvent.GetType())
	assert.Equal(t, "audio-player", playerEvent.GetSource())
	assert.Equal(t, now, playerEvent.GetTimestamp())

	// 测试播放器特定字段
	assert.Equal(t, "123", playerEvent.SongID)
	assert.Equal(t, time.Second*30, playerEvent.Position)
	assert.Equal(t, 0.8, playerEvent.Volume)
}

func TestSystemEvent(t *testing.T) {
	now := time.Now()
	baseEvent := &BaseEvent{
		ID:        "system-test-1",
		Type:      EventSystemStartup,
		Source:    "system",
		Timestamp: now,
		Data:      map[string]interface{}{"version": "2.0.0"},
	}

	systemEvent := &SystemEvent{
		BaseEvent: baseEvent,
		Message:   "System started successfully",
		Level:     "info",
	}

	// 测试继承的Event接口方法
	assert.Equal(t, "system-test-1", systemEvent.GetID())
	assert.Equal(t, EventSystemStartup, systemEvent.GetType())
	assert.Equal(t, "system", systemEvent.GetSource())
	assert.Equal(t, now, systemEvent.GetTimestamp())

	// 测试系统特定字段
	assert.Equal(t, "System started successfully", systemEvent.Message)
	assert.Equal(t, "info", systemEvent.Level)
}

func TestSubscriberInfo(t *testing.T) {
	now := time.Now()
	lastEventTime := now.Add(-time.Minute)

	subscriberInfo := SubscriberInfo{
		ID:          "subscriber-1",
		EventType:   EventPlayerPlay,
		Priority:    PriorityNormal,
		Async:       true,
		Active:      true,
		Group:       "audio-group",
		CreatedAt:   now,
		LastEventAt: &lastEventTime,
		EventCount:  42,
	}

	assert.Equal(t, "subscriber-1", subscriberInfo.ID)
	assert.Equal(t, EventPlayerPlay, subscriberInfo.EventType)
	assert.Equal(t, PriorityNormal, subscriberInfo.Priority)
	assert.True(t, subscriberInfo.Async)
	assert.True(t, subscriberInfo.Active)
	assert.Equal(t, "audio-group", subscriberInfo.Group)
	assert.Equal(t, now, subscriberInfo.CreatedAt)
	assert.Equal(t, lastEventTime, *subscriberInfo.LastEventAt)
	assert.Equal(t, int64(42), subscriberInfo.EventCount)
}

func TestPredefinedEventTypes(t *testing.T) {
	registry := NewEventTypeRegistry()

	// 测试一些预定义事件类型是否已注册
	predefinedTypes := []EventType{
		EventPluginLoaded,
		EventPluginStarted,
		EventPlayerPlay,
		EventPlayerPause,
		EventSystemStartup,
		EventSystemShutdown,
		EventConfigChanged,
	}

	for _, eventType := range predefinedTypes {
		assert.True(t, registry.IsRegistered(eventType), "EventType %s should be registered", eventType)
		
		info, exists := registry.GetEventTypeInfo(eventType)
		assert.True(t, exists, "EventType %s should have info", eventType)
		assert.Equal(t, eventType, info.Type)
		assert.NotEmpty(t, info.Category, "EventType %s should have category", eventType)
		assert.NotEmpty(t, info.Description, "EventType %s should have description", eventType)
		assert.False(t, info.RegisteredAt.IsZero(), "EventType %s should have registration time", eventType)
	}
}