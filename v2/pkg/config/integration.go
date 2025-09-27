package config

import (
	"context"
	"fmt"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/event"
)

// ConfigEventType 配置相关事件类型
const (
	// 配置变更事件
	EventConfigChanged        event.EventType = "config.changed"
	EventConfigReloaded       event.EventType = "config.reloaded"
	EventConfigVersionCreated event.EventType = "config.version.created"
	EventConfigRollback       event.EventType = "config.rollback"
	EventConfigTemplateApplied event.EventType = "config.template.applied"
	EventConfigEncrypted      event.EventType = "config.encrypted"
	EventConfigDecrypted      event.EventType = "config.decrypted"
	EventConfigImported       event.EventType = "config.imported"
	EventConfigExported       event.EventType = "config.exported"

	// 热更新事件
	EventConfigHotReloadEnabled  event.EventType = "config.hotreload.enabled"
	EventConfigHotReloadDisabled event.EventType = "config.hotreload.disabled"
	EventConfigFileChanged       event.EventType = "config.file.changed"
	EventConfigTemplateChanged   event.EventType = "config.template.changed"

	// 安全事件
	EventConfigAccessDenied     event.EventType = "config.access.denied"
	EventConfigKeyRotated       event.EventType = "config.key.rotated"
	EventConfigSecurityAudit    event.EventType = "config.security.audit"
	EventConfigVulnerabilityFound event.EventType = "config.vulnerability.found"

	// 事务事件
	EventConfigTransactionStarted   event.EventType = "config.transaction.started"
	EventConfigTransactionCommitted event.EventType = "config.transaction.committed"
	EventConfigTransactionRolledback event.EventType = "config.transaction.rolledback"
)

// ConfigEvent 配置事件
type ConfigEvent struct {
	*event.BaseEvent
	Key       string      `json:"key,omitempty"`
	OldValue  interface{} `json:"old_value,omitempty"`
	NewValue  interface{} `json:"new_value,omitempty"`
	User      string      `json:"user,omitempty"`
	Operation string      `json:"operation,omitempty"`
}

// ConfigVersionEvent 配置版本事件
type ConfigVersionEvent struct {
	*event.BaseEvent
	VersionID   string `json:"version_id"`
	Description string `json:"description"`
	Checksum    string `json:"checksum"`
}

// ConfigSecurityEvent 配置安全事件
type ConfigSecurityEvent struct {
	*event.BaseEvent
	SecurityLevel string   `json:"security_level"`
	Vulnerabilities []string `json:"vulnerabilities,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
	SecurityScore   int      `json:"security_score,omitempty"`
}

// ConfigTransactionEvent 配置事务事件
type ConfigTransactionEvent struct {
	*event.BaseEvent
	TransactionID string            `json:"transaction_id"`
	Changes       map[string]interface{} `json:"changes,omitempty"`
	Deletions     []string          `json:"deletions,omitempty"`
}

// EventIntegration 事件系统集成
type EventIntegration struct {
	eventBus      event.EventBus
	configManager *AdvancedManager
	subscriptions []*event.Subscription
	enabled       bool
}

// NewEventIntegration 创建事件系统集成
func NewEventIntegration(eventBus event.EventBus, configManager *AdvancedManager) *EventIntegration {
	return &EventIntegration{
		eventBus:      eventBus,
		configManager: configManager,
		subscriptions: make([]*event.Subscription, 0),
		enabled:       false,
	}
}

// Enable 启用事件集成
func (ei *EventIntegration) Enable(ctx context.Context) error {
	if ei.enabled {
		return fmt.Errorf("event integration is already enabled")
	}

	// 注册配置相关事件类型
	if err := ei.registerEventTypes(); err != nil {
		return fmt.Errorf("failed to register event types: %w", err)
	}

	// 设置配置变更回调
	if err := ei.setupConfigChangeCallbacks(); err != nil {
		return fmt.Errorf("failed to setup config change callbacks: %w", err)
	}

	// 订阅系统事件
	if err := ei.subscribeToSystemEvents(ctx); err != nil {
		return fmt.Errorf("failed to subscribe to system events: %w", err)
	}

	ei.enabled = true
	return nil
}

// Disable 禁用事件集成
func (ei *EventIntegration) Disable() error {
	if !ei.enabled {
		return nil
	}

	// 取消所有订阅
	for _, subscription := range ei.subscriptions {
		if err := ei.eventBus.Unsubscribe(subscription.ID); err != nil {
			fmt.Printf("Warning: failed to unsubscribe %s: %v\n", subscription.ID, err)
		}
	}

	ei.subscriptions = ei.subscriptions[:0]
	ei.enabled = false
	return nil
}

// registerEventTypes 注册事件类型
func (ei *EventIntegration) registerEventTypes() error {
	eventTypes := []event.EventType{
		EventConfigChanged,
		EventConfigReloaded,
		EventConfigVersionCreated,
		EventConfigRollback,
		EventConfigTemplateApplied,
		EventConfigEncrypted,
		EventConfigDecrypted,
		EventConfigImported,
		EventConfigExported,
		EventConfigHotReloadEnabled,
		EventConfigHotReloadDisabled,
		EventConfigFileChanged,
		EventConfigTemplateChanged,
		EventConfigAccessDenied,
		EventConfigKeyRotated,
		EventConfigSecurityAudit,
		EventConfigVulnerabilityFound,
		EventConfigTransactionStarted,
		EventConfigTransactionCommitted,
		EventConfigTransactionRolledback,
	}

	for _, eventType := range eventTypes {
		if err := ei.eventBus.RegisterEventType(eventType); err != nil {
			return fmt.Errorf("failed to register event type %s: %w", eventType, err)
		}
	}

	return nil
}

// setupConfigChangeCallbacks 设置配置变更回调
func (ei *EventIntegration) setupConfigChangeCallbacks() error {
	// 注册配置变更回调
	return ei.configManager.OnConfigChanged(func(change *ConfigChange) error {
		return ei.publishConfigChangeEvent(change)
	})
}

// publishConfigChangeEvent 发布配置变更事件
func (ei *EventIntegration) publishConfigChangeEvent(change *ConfigChange) error {
	ctx := context.Background()

	// 创建配置事件
	configEvent := &ConfigEvent{
		BaseEvent: &event.BaseEvent{
			ID:        change.ID,
			Type:      EventConfigChanged,
			Source:    "config_manager",
			Timestamp: change.Timestamp,
			Data:      change,
		},
		Key:       change.Key,
		OldValue:  change.OldValue,
		NewValue:  change.NewValue,
		User:      change.User,
		Operation: change.Operation,
	}

	// 根据操作类型发布不同的事件
	switch change.Operation {
	case "rollback":
		configEvent.Type = EventConfigRollback
	case "template_apply":
		configEvent.Type = EventConfigTemplateApplied
	case "encrypt":
		configEvent.Type = EventConfigEncrypted
	case "decrypt":
		configEvent.Type = EventConfigDecrypted
	case "import":
		configEvent.Type = EventConfigImported
	case "export":
		configEvent.Type = EventConfigExported
	case "hot_reload":
		configEvent.Type = EventConfigReloaded
	case "transaction_commit":
		configEvent.Type = EventConfigTransactionCommitted
	case "transaction_rollback":
		configEvent.Type = EventConfigTransactionRolledback
	}

	return ei.eventBus.PublishAsync(ctx, configEvent)
}

// subscribeToSystemEvents 订阅系统事件
func (ei *EventIntegration) subscribeToSystemEvents(ctx context.Context) error {
	// 订阅插件事件，处理插件配置变更
	pluginSub, err := ei.eventBus.Subscribe(
		event.EventPluginLoaded,
		ei.handlePluginEvent,
		event.WithAsync(true),
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to plugin events: %w", err)
	}
	ei.subscriptions = append(ei.subscriptions, pluginSub)

	// 订阅系统启动事件
	systemSub, err := ei.eventBus.Subscribe(
		event.EventSystemStartup,
		ei.handleSystemEvent,
		event.WithAsync(true),
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to system events: %w", err)
	}
	ei.subscriptions = append(ei.subscriptions, systemSub)

	return nil
}

// handlePluginEvent 处理插件事件
func (ei *EventIntegration) handlePluginEvent(ctx context.Context, evt event.Event) error {
	pluginEvent, ok := evt.(*event.PluginEvent)
	if !ok {
		return fmt.Errorf("invalid plugin event type")
	}

	// 当插件加载时，检查是否需要应用插件特定的配置
	if pluginEvent.GetType() == event.EventPluginLoaded {
		return ei.applyPluginConfig(pluginEvent.PluginName)
	}

	return nil
}

// handleSystemEvent 处理系统事件
func (ei *EventIntegration) handleSystemEvent(ctx context.Context, evt event.Event) error {
	systemEvent, ok := evt.(*event.SystemEvent)
	if !ok {
		return fmt.Errorf("invalid system event type")
	}

	// 当系统启动时，执行配置初始化
	if systemEvent.GetType() == event.EventSystemStartup {
		return ei.initializeConfigOnStartup()
	}

	return nil
}

// applyPluginConfig 应用插件配置
func (ei *EventIntegration) applyPluginConfig(pluginName string) error {
	// 检查是否有插件特定的配置模板
	templateKey := fmt.Sprintf("plugin_%s_template", pluginName)
	if ei.configManager.k.Exists(templateKey) {
		templateName := ei.configManager.k.String(templateKey)
		if templateName != "" {
			// 应用模板
			variables := map[string]interface{}{
				"plugin_name": pluginName,
				"timestamp":   time.Now().Format(time.RFC3339),
			}
			return ei.configManager.ApplyTemplate(templateName, variables)
		}
	}

	return nil
}

// initializeConfigOnStartup 系统启动时初始化配置
func (ei *EventIntegration) initializeConfigOnStartup() error {
	// 创建启动快照
	if _, err := ei.configManager.CreateSnapshot("System startup snapshot"); err != nil {
		fmt.Printf("Warning: failed to create startup snapshot: %v\n", err)
	}

	// 执行安全审计
	audit := ei.configManager.CreateSecurityAudit()
	if len(audit.Vulnerabilities) > 0 {
		// 发布安全审计事件
		securityEvent := &ConfigSecurityEvent{
			BaseEvent: &event.BaseEvent{
				ID:        generateID(),
				Type:      EventConfigSecurityAudit,
				Source:    "config_manager",
				Timestamp: time.Now(),
				Data:      audit,
			},
			SecurityLevel:   "warning",
			Vulnerabilities: audit.Vulnerabilities,
			Recommendations: audit.Recommendations,
			SecurityScore:   audit.SecurityScore,
		}

		ctx := context.Background()
		if err := ei.eventBus.PublishAsync(ctx, securityEvent); err != nil {
			fmt.Printf("Warning: failed to publish security audit event: %v\n", err)
		}
	}

	return nil
}

// PublishVersionEvent 发布版本事件
func (ei *EventIntegration) PublishVersionEvent(version *ConfigVersion) error {
	if !ei.enabled {
		return nil
	}

	versionEvent := &ConfigVersionEvent{
		BaseEvent: &event.BaseEvent{
			ID:        generateID(),
			Type:      EventConfigVersionCreated,
			Source:    "config_manager",
			Timestamp: time.Now(),
			Data:      version,
		},
		VersionID:   version.ID,
		Description: version.Description,
		Checksum:    version.Checksum,
	}

	ctx := context.Background()
	return ei.eventBus.PublishAsync(ctx, versionEvent)
}

// PublishTransactionEvent 发布事务事件
func (ei *EventIntegration) PublishTransactionEvent(eventType event.EventType, transactionID string, changes map[string]interface{}, deletions []string) error {
	if !ei.enabled {
		return nil
	}

	transactionEvent := &ConfigTransactionEvent{
		BaseEvent: &event.BaseEvent{
			ID:        generateID(),
			Type:      eventType,
			Source:    "config_manager",
			Timestamp: time.Now(),
		},
		TransactionID: transactionID,
		Changes:       changes,
		Deletions:     deletions,
	}

	ctx := context.Background()
	return ei.eventBus.PublishAsync(ctx, transactionEvent)
}

// PublishHotReloadEvent 发布热更新事件
func (ei *EventIntegration) PublishHotReloadEvent(eventType event.EventType, path string, operation string) error {
	if !ei.enabled {
		return nil
	}

	hotReloadEvent := &ConfigEvent{
		BaseEvent: &event.BaseEvent{
			ID:        generateID(),
			Type:      eventType,
			Source:    "hot_reload_manager",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"path":      path,
				"operation": operation,
			},
		},
		Operation: operation,
	}

	ctx := context.Background()
	return ei.eventBus.PublishAsync(ctx, hotReloadEvent)
}

// PublishSecurityEvent 发布安全事件
func (ei *EventIntegration) PublishSecurityEvent(eventType event.EventType, level string, message string, data interface{}) error {
	if !ei.enabled {
		return nil
	}

	securityEvent := &ConfigSecurityEvent{
		BaseEvent: &event.BaseEvent{
			ID:        generateID(),
			Type:      eventType,
			Source:    "config_security",
			Timestamp: time.Now(),
			Data:      data,
		},
		SecurityLevel: level,
	}

	ctx := context.Background()
	return ei.eventBus.PublishAsync(ctx, securityEvent)
}

// GetEventStats 获取配置相关事件统计
func (ei *EventIntegration) GetEventStats() map[event.EventType]int64 {
	if !ei.enabled {
		return make(map[event.EventType]int64)
	}

	stats := ei.eventBus.GetEventStats()
	configStats := make(map[event.EventType]int64)

	// 过滤配置相关事件
	for eventType, count := range stats.EventCounts {
		if ei.isConfigEvent(eventType) {
			configStats[eventType] = count
		}
	}

	return configStats
}

// isConfigEvent 检查是否是配置相关事件
func (ei *EventIntegration) isConfigEvent(eventType event.EventType) bool {
	configEventPrefix := "config."
	return len(eventType) > len(configEventPrefix) && string(eventType)[:len(configEventPrefix)] == configEventPrefix
}

// IsEnabled 检查事件集成是否启用
func (ei *EventIntegration) IsEnabled() bool {
	return ei.enabled
}