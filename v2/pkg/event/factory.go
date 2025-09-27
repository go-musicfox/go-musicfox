package event

import (
	"time"

	"github.com/google/uuid"
)

// EventFactory 事件工厂
type EventFactory struct{}

// NewEventFactory 创建事件工厂
func NewEventFactory() *EventFactory {
	return &EventFactory{}
}

// CreateEvent 创建基础事件
func (f *EventFactory) CreateEvent(eventType EventType, data interface{}, source string) Event {
	return &BaseEvent{
		ID:        uuid.New().String(),
		Type:      eventType,
		Data:      data,
		Source:    source,
		Timestamp: time.Now(),
	}
}

// CreatePriorityEvent 创建带优先级的事件
func (f *EventFactory) CreatePriorityEvent(eventType EventType, data interface{}, source string, priority EventPriority) Event {
	baseEvent := &BaseEvent{
		ID:        uuid.New().String(),
		Type:      eventType,
		Data:      data,
		Source:    source,
		Timestamp: time.Now(),
	}

	return &PriorityEvent{
		BaseEvent: baseEvent,
		Priority:  priority,
	}
}

// CreatePluginEvent 创建插件事件
func (f *EventFactory) CreatePluginEvent(eventType EventType, pluginName, pluginType, path string, err error) Event {
	baseEvent := &BaseEvent{
		ID:        uuid.New().String(),
		Type:      eventType,
		Data:      nil,
		Source:    "plugin_manager",
		Timestamp: time.Now(),
	}

	pluginEvent := &PluginEvent{
		BaseEvent:  baseEvent,
		PluginName: pluginName,
		PluginType: pluginType,
		Path:       path,
	}

	if err != nil {
		pluginEvent.Error = err.Error()
	}

	return pluginEvent
}

// CreatePlayerEvent 创建播放器事件
func (f *EventFactory) CreatePlayerEvent(eventType EventType, songID string, position time.Duration, volume float64) Event {
	baseEvent := &BaseEvent{
		ID:        uuid.New().String(),
		Type:      eventType,
		Data:      nil,
		Source:    "player",
		Timestamp: time.Now(),
	}

	return &PlayerEvent{
		BaseEvent: baseEvent,
		SongID:    songID,
		Position:  position,
		Volume:    volume,
	}
}

// CreateSystemEvent 创建系统事件
func (f *EventFactory) CreateSystemEvent(eventType EventType, message, level string) Event {
	baseEvent := &BaseEvent{
		ID:        uuid.New().String(),
		Type:      eventType,
		Data:      nil,
		Source:    "system",
		Timestamp: time.Now(),
	}

	return &SystemEvent{
		BaseEvent: baseEvent,
		Message:   message,
		Level:     level,
	}
}

// 便捷函数

// NewPluginLoadedEvent 创建插件加载事件
func NewPluginLoadedEvent(pluginName, pluginType, path string) Event {
	factory := NewEventFactory()
	return factory.CreatePluginEvent(EventPluginLoaded, pluginName, pluginType, path, nil)
}

// NewPluginStartedEvent 创建插件启动事件
func NewPluginStartedEvent(pluginName, pluginType string) Event {
	factory := NewEventFactory()
	return factory.CreatePluginEvent(EventPluginStarted, pluginName, pluginType, "", nil)
}

// NewPluginStoppedEvent 创建插件停止事件
func NewPluginStoppedEvent(pluginName, pluginType string) Event {
	factory := NewEventFactory()
	return factory.CreatePluginEvent(EventPluginStopped, pluginName, pluginType, "", nil)
}

// NewPluginUnloadedEvent 创建插件卸载事件
func NewPluginUnloadedEvent(pluginName, pluginType string) Event {
	factory := NewEventFactory()
	return factory.CreatePluginEvent(EventPluginUnloaded, pluginName, pluginType, "", nil)
}

// NewPluginErrorEvent 创建插件错误事件
func NewPluginErrorEvent(pluginName, pluginType string, err error) Event {
	factory := NewEventFactory()
	return factory.CreatePluginEvent(EventPluginError, pluginName, pluginType, "", err)
}

// NewPlayerPlayEvent 创建播放事件
func NewPlayerPlayEvent(songID string) Event {
	factory := NewEventFactory()
	return factory.CreatePlayerEvent(EventPlayerPlay, songID, 0, 0)
}

// NewPlayerPauseEvent 创建暂停事件
func NewPlayerPauseEvent(songID string, position time.Duration) Event {
	factory := NewEventFactory()
	return factory.CreatePlayerEvent(EventPlayerPause, songID, position, 0)
}

// NewPlayerStopEvent 创建停止事件
func NewPlayerStopEvent() Event {
	factory := NewEventFactory()
	return factory.CreatePlayerEvent(EventPlayerStop, "", 0, 0)
}

// NewPlayerVolumeChangedEvent 创建音量变化事件
func NewPlayerVolumeChangedEvent(volume float64) Event {
	factory := NewEventFactory()
	return factory.CreatePlayerEvent(EventPlayerVolumeChanged, "", 0, volume)
}

// NewSystemStartupEvent 创建系统启动事件
func NewSystemStartupEvent() Event {
	factory := NewEventFactory()
	return factory.CreateSystemEvent(EventSystemStartup, "System started successfully", "info")
}

// NewSystemShutdownEvent 创建系统关闭事件
func NewSystemShutdownEvent() Event {
	factory := NewEventFactory()
	return factory.CreateSystemEvent(EventSystemShutdown, "System is shutting down", "info")
}

// NewSystemErrorEvent 创建系统错误事件
func NewSystemErrorEvent(message string) Event {
	factory := NewEventFactory()
	return factory.CreateSystemEvent(EventSystemError, message, "error")
}

// NewConfigChangedEvent 创建配置变更事件
func NewConfigChangedEvent(configKey string) Event {
	factory := NewEventFactory()
	return factory.CreateEvent(EventConfigChanged, map[string]string{"key": configKey}, "config_manager")
}