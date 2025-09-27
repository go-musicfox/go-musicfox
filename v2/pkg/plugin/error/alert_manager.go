package plugin

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// AlertManager 告警管理器接口
type AlertManager interface {
	// RegisterAlert 注册告警规则
	RegisterAlert(rule *AlertRule) error
	// UnregisterAlert 注销告警规则
	UnregisterAlert(ruleID string) error
	// TriggerAlert 触发告警
	TriggerAlert(ruleID string, data map[string]interface{}) error
	// GetActiveAlerts 获取活跃告警
	GetActiveAlerts() []*Alert
	// GetAlert 获取指定告警
	GetAlert(alertID string) (*Alert, error)
	// AcknowledgeAlert 确认告警
	AcknowledgeAlert(alertID string, acknowledgedBy string) error
	// ResolveAlert 解决告警
	ResolveAlert(alertID string, resolvedBy string) error
	// RegisterHandler 注册告警处理器
	RegisterHandler(handler AlertHandler) error
	// UnregisterHandler 注销告警处理器
	UnregisterHandler(handlerName string) error
	// Start 启动告警管理器
	Start(ctx context.Context) error
	// Stop 停止告警管理器
	Stop() error
	// GetStats 获取告警统计
	GetStats() *AlertManagerStats
}

// AlertRule 告警规则
type AlertRule struct {
	ID          string            `json:"id"`          // 规则ID
	Name        string            `json:"name"`        // 规则名称
	Description string            `json:"description"` // 规则描述
	Condition   string            `json:"condition"`   // 告警条件
	Threshold   float64           `json:"threshold"`   // 阈值
	Duration    time.Duration     `json:"duration"`    // 持续时间
	Severity    AlertSeverity     `json:"severity"`    // 严重程度
	Enabled     bool              `json:"enabled"`     // 是否启用
	Actions     []AlertAction     `json:"actions"`     // 告警动作
	Labels      map[string]string `json:"labels"`      // 标签
	Annotations map[string]string `json:"annotations"` // 注释
	CreatedAt   time.Time         `json:"created_at"`  // 创建时间
	UpdatedAt   time.Time         `json:"updated_at"`  // 更新时间
}

// AlertAction 告警动作
type AlertAction struct {
	Type       AlertActionType   `json:"type"`       // 动作类型
	Target     string            `json:"target"`     // 目标
	Parameters map[string]string `json:"parameters"` // 参数
	Enabled    bool              `json:"enabled"`    // 是否启用
}

// AlertActionType 告警动作类型
type AlertActionType int

const (
	AlertActionTypeLog AlertActionType = iota
	AlertActionTypeEmail
	AlertActionTypeWebhook
	AlertActionTypeSMS
	AlertActionTypeSlack
	AlertActionTypeCustom
)

// String 返回告警动作类型的字符串表示
func (a AlertActionType) String() string {
	switch a {
	case AlertActionTypeLog:
		return "log"
	case AlertActionTypeEmail:
		return "email"
	case AlertActionTypeWebhook:
		return "webhook"
	case AlertActionTypeSMS:
		return "sms"
	case AlertActionTypeSlack:
		return "slack"
	case AlertActionTypeCustom:
		return "custom"
	default:
		return "unknown"
	}
}

// AlertManagerStats 告警管理器统计
type AlertManagerStats struct {
	TotalRules      int                        `json:"total_rules"`      // 总规则数
	ActiveRules     int                        `json:"active_rules"`     // 活跃规则数
	TotalAlerts     int                        `json:"total_alerts"`     // 总告警数
	ActiveAlerts    int                        `json:"active_alerts"`    // 活跃告警数
	ResolvedAlerts  int                        `json:"resolved_alerts"`  // 已解决告警数
	AlertsBySeverity map[AlertSeverity]int     `json:"alerts_by_severity"` // 按严重程度分组的告警数
	AlertsByType    map[AlertType]int          `json:"alerts_by_type"`    // 按类型分组的告警数
	LastAlert       *Alert                     `json:"last_alert"`       // 最后一个告警
	UpdatedAt       time.Time                  `json:"updated_at"`       // 更新时间
}

// SmartAlertManager 智能告警管理器实现
type SmartAlertManager struct {
	rules        map[string]*AlertRule
	alerts       map[string]*Alert
	handlers     map[string]AlertHandler
	mutex        sync.RWMutex
	logger       Logger
	metrics      MetricsCollector
	eventBus     EventBus
	running      bool
	stopChan     chan struct{}
	checkInterval time.Duration
}

// NewSmartAlertManager 创建智能告警管理器
func NewSmartAlertManager(logger Logger, metrics MetricsCollector, eventBus EventBus) *SmartAlertManager {
	return &SmartAlertManager{
		rules:         make(map[string]*AlertRule),
		alerts:        make(map[string]*Alert),
		handlers:      make(map[string]AlertHandler),
		logger:        logger,
		metrics:       metrics,
		eventBus:      eventBus,
		stopChan:      make(chan struct{}),
		checkInterval: 30 * time.Second,
	}
}

// RegisterAlert 注册告警规则
func (sam *SmartAlertManager) RegisterAlert(rule *AlertRule) error {
	sam.mutex.Lock()
	defer sam.mutex.Unlock()
	
	if rule.ID == "" {
		return NewPluginError(ErrorCodeInvalidArgument, "alert rule ID cannot be empty")
	}
	
	if _, exists := sam.rules[rule.ID]; exists {
		return NewPluginError(ErrorCodeAlreadyExists, fmt.Sprintf("alert rule %s already exists", rule.ID))
	}
	
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	sam.rules[rule.ID] = rule
	
	if sam.logger != nil {
		sam.logger.Info("Alert rule registered", map[string]interface{}{
			"rule_id": rule.ID,
			"name": rule.Name,
		})
	}
	
	// 记录指标
	if sam.metrics != nil {
		sam.metrics.IncrementCounter("alert_rules_registered_total", map[string]string{
			"severity": rule.Severity.String(),
		})
	}
	
	// 发送事件
	if sam.eventBus != nil {
		sam.eventBus.Publish("alert_rule_registered", map[string]interface{}{
			"rule_id":   rule.ID,
			"rule_name": rule.Name,
			"severity":  rule.Severity.String(),
			"timestamp": time.Now(),
		})
	}
	
	return nil
}

// UnregisterAlert 注销告警规则
func (sam *SmartAlertManager) UnregisterAlert(ruleID string) error {
	sam.mutex.Lock()
	defer sam.mutex.Unlock()
	
	if _, exists := sam.rules[ruleID]; !exists {
		return NewPluginError(ErrorCodeNotFound, fmt.Sprintf("alert rule %s not found", ruleID))
	}
	
	delete(sam.rules, ruleID)
	
	if sam.logger != nil {
		sam.logger.Info("Alert rule unregistered", map[string]interface{}{
			"rule_id": ruleID,
		})
	}
	
	return nil
}

// TriggerAlert 触发告警
func (sam *SmartAlertManager) TriggerAlert(ruleID string, data map[string]interface{}) error {
	sam.mutex.Lock()
	rule, exists := sam.rules[ruleID]
	sam.mutex.Unlock()
	
	if !exists {
		return NewPluginError(ErrorCodeNotFound, fmt.Sprintf("alert rule %s not found", ruleID))
	}
	
	if !rule.Enabled {
		return nil // 规则未启用，忽略
	}
	
	// 创建告警
	alert := &Alert{
		ID:        fmt.Sprintf("%s-%d", ruleID, time.Now().UnixNano()),
		Type:      sam.getAlertTypeFromRule(rule),
		PluginID:  sam.getPluginIDFromData(data),
		Message:   sam.buildAlertMessage(rule, data),
		Severity:  rule.Severity,
		Timestamp: time.Now(),
		Metadata:  data,
		Resolved:  false,
	}
	
	// 存储告警
	sam.mutex.Lock()
	sam.alerts[alert.ID] = alert
	sam.mutex.Unlock()
	
	if sam.logger != nil {
		sam.logger.Warn("Alert triggered", map[string]interface{}{
			"alert_id": alert.ID,
			"rule_id": ruleID,
			"message": alert.Message,
		})
	}
	
	// 记录指标
	if sam.metrics != nil {
		sam.metrics.IncrementCounter("alerts_triggered_total", map[string]string{
			"severity": alert.Severity.String(),
			"type":     alert.Type.String(),
		})
	}
	
	// 发送事件
	if sam.eventBus != nil {
		sam.eventBus.Publish("alert_triggered", map[string]interface{}{
			"alert_id":  alert.ID,
			"rule_id":   ruleID,
			"severity":  alert.Severity.String(),
			"message":   alert.Message,
			"timestamp": alert.Timestamp,
		})
	}
	
	// 执行告警动作
	go sam.executeAlertActions(alert, rule)
	
	return nil
}

// GetActiveAlerts 获取活跃告警
func (sam *SmartAlertManager) GetActiveAlerts() []*Alert {
	sam.mutex.RLock()
	defer sam.mutex.RUnlock()
	
	var activeAlerts []*Alert
	for _, alert := range sam.alerts {
		if !alert.Resolved {
			activeAlerts = append(activeAlerts, alert)
		}
	}
	
	// 按时间戳排序
	sort.Slice(activeAlerts, func(i, j int) bool {
		return activeAlerts[i].Timestamp.After(activeAlerts[j].Timestamp)
	})
	
	return activeAlerts
}

// GetAlert 获取指定告警
func (sam *SmartAlertManager) GetAlert(alertID string) (*Alert, error) {
	sam.mutex.RLock()
	defer sam.mutex.RUnlock()
	
	alert, exists := sam.alerts[alertID]
	if !exists {
		return nil, NewPluginError(ErrorCodeNotFound, fmt.Sprintf("alert %s not found", alertID))
	}
	
	return alert, nil
}

// AcknowledgeAlert 确认告警
func (sam *SmartAlertManager) AcknowledgeAlert(alertID string, acknowledgedBy string) error {
	sam.mutex.Lock()
	defer sam.mutex.Unlock()
	
	alert, exists := sam.alerts[alertID]
	if !exists {
		return NewPluginError(ErrorCodeNotFound, fmt.Sprintf("alert %s not found", alertID))
	}
	
	if alert.Resolved {
		return NewPluginError(ErrorCodeInvalidArgument, "cannot acknowledge resolved alert")
	}
	
	// 这里可以添加确认逻辑，比如设置确认标志
	if alert.Metadata == nil {
		alert.Metadata = make(map[string]interface{})
	}
	alert.Metadata["acknowledged_by"] = acknowledgedBy
	alert.Metadata["acknowledged_at"] = time.Now()
	
	if sam.logger != nil {
		sam.logger.Info("Alert acknowledged", map[string]interface{}{
			"alert_id": alertID,
			"acknowledged_by": acknowledgedBy,
		})
	}
	
	return nil
}

// ResolveAlert 解决告警
func (sam *SmartAlertManager) ResolveAlert(alertID string, resolvedBy string) error {
	sam.mutex.Lock()
	defer sam.mutex.Unlock()
	
	alert, exists := sam.alerts[alertID]
	if !exists {
		return NewPluginError(ErrorCodeNotFound, fmt.Sprintf("alert %s not found", alertID))
	}
	
	if alert.Resolved {
		return NewPluginError(ErrorCodeInvalidArgument, "alert already resolved")
	}
	
	now := time.Now()
	alert.Resolved = true
	alert.ResolvedAt = &now
	
	if alert.Metadata == nil {
		alert.Metadata = make(map[string]interface{})
	}
	alert.Metadata["resolved_by"] = resolvedBy
	
	if sam.logger != nil {
		sam.logger.Info("Alert resolved", map[string]interface{}{
			"alert_id": alertID,
			"resolved_by": resolvedBy,
		})
	}
	
	// 记录指标
	if sam.metrics != nil {
		sam.metrics.IncrementCounter("alerts_resolved_total", map[string]string{
			"severity": alert.Severity.String(),
		})
	}
	
	// 发送事件
	if sam.eventBus != nil {
		sam.eventBus.Publish("alert_resolved", map[string]interface{}{
			"alert_id":    alertID,
			"resolved_by": resolvedBy,
			"timestamp":   now,
		})
	}
	
	return nil
}

// RegisterHandler 注册告警处理器
func (sam *SmartAlertManager) RegisterHandler(handler AlertHandler) error {
	sam.mutex.Lock()
	defer sam.mutex.Unlock()
	
	name := handler.GetName()
	if _, exists := sam.handlers[name]; exists {
		return NewPluginError(ErrorCodeAlreadyExists, fmt.Sprintf("alert handler %s already exists", name))
	}
	
	sam.handlers[name] = handler
	
	if sam.logger != nil {
		sam.logger.Info("Alert handler registered", map[string]interface{}{
			"handler_name": name,
		})
	}
	
	return nil
}

// UnregisterHandler 注销告警处理器
func (sam *SmartAlertManager) UnregisterHandler(handlerName string) error {
	sam.mutex.Lock()
	defer sam.mutex.Unlock()
	
	if _, exists := sam.handlers[handlerName]; !exists {
		return NewPluginError(ErrorCodeNotFound, fmt.Sprintf("alert handler %s not found", handlerName))
	}
	
	delete(sam.handlers, handlerName)
	
	if sam.logger != nil {
		sam.logger.Info("Alert handler unregistered", map[string]interface{}{
			"handler_name": handlerName,
		})
	}
	
	return nil
}

// Start 启动告警管理器
func (sam *SmartAlertManager) Start(ctx context.Context) error {
	sam.mutex.Lock()
	if sam.running {
		sam.mutex.Unlock()
		return NewPluginError(ErrorCodeAlreadyExists, "alert manager already running")
	}
	sam.running = true
	sam.mutex.Unlock()
	
	if sam.logger != nil {
		sam.logger.Info("Alert manager started", map[string]interface{}{})
	}
	
	// 启动定期检查协程
	go sam.periodicCheck(ctx)
	
	return nil
}

// Stop 停止告警管理器
func (sam *SmartAlertManager) Stop() error {
	sam.mutex.Lock()
	if !sam.running {
		sam.mutex.Unlock()
		return NewPluginError(ErrorCodeInvalidArgument, "alert manager not running")
	}
	sam.running = false
	sam.mutex.Unlock()
	
	close(sam.stopChan)
	
	if sam.logger != nil {
		sam.logger.Info("Alert manager stopped", map[string]interface{}{})
	}
	
	return nil
}

// GetStats 获取告警统计
func (sam *SmartAlertManager) GetStats() *AlertManagerStats {
	sam.mutex.RLock()
	defer sam.mutex.RUnlock()
	
	stats := &AlertManagerStats{
		TotalRules:       len(sam.rules),
		TotalAlerts:      len(sam.alerts),
		AlertsBySeverity: make(map[AlertSeverity]int),
		AlertsByType:     make(map[AlertType]int),
		UpdatedAt:        time.Now(),
	}
	
	// 统计活跃规则
	for _, rule := range sam.rules {
		if rule.Enabled {
			stats.ActiveRules++
		}
	}
	
	// 统计告警
	var lastAlert *Alert
	for _, alert := range sam.alerts {
		if alert.Resolved {
			stats.ResolvedAlerts++
		} else {
			stats.ActiveAlerts++
		}
		
		stats.AlertsBySeverity[alert.Severity]++
		stats.AlertsByType[alert.Type]++
		
		if lastAlert == nil || alert.Timestamp.After(lastAlert.Timestamp) {
			lastAlert = alert
		}
	}
	
	stats.LastAlert = lastAlert
	return stats
}

// executeAlertActions 执行告警动作
func (sam *SmartAlertManager) executeAlertActions(alert *Alert, rule *AlertRule) {
	for _, action := range rule.Actions {
		if !action.Enabled {
			continue
		}
		
		switch action.Type {
		case AlertActionTypeLog:
			if sam.logger != nil {
				sam.logger.Error("ALERT", map[string]interface{}{
					"alert_id": alert.ID,
					"message": alert.Message,
					"severity": alert.Severity.String(),
				})
			}
			
		case AlertActionTypeWebhook:
			// 这里可以实现webhook调用
			if sam.logger != nil {
				sam.logger.Info("Webhook alert action", map[string]interface{}{
					"alert_id": alert.ID,
					"target": action.Target,
				})
			}
			
		case AlertActionTypeCustom:
			// 调用自定义处理器
			if handler, exists := sam.handlers[action.Target]; exists {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				
				if err := handler.HandleAlert(ctx, *alert); err != nil && sam.logger != nil {
					sam.logger.Error("Alert handler failed", map[string]interface{}{
						"handler": action.Target,
						"error": err,
					})
				}
			}
		}
	}
}

// periodicCheck 定期检查
func (sam *SmartAlertManager) periodicCheck(ctx context.Context) {
	ticker := time.NewTicker(sam.checkInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-sam.stopChan:
			return
		case <-ticker.C:
			sam.cleanupOldAlerts()
		}
	}
}

// cleanupOldAlerts 清理旧告警
func (sam *SmartAlertManager) cleanupOldAlerts() {
	sam.mutex.Lock()
	defer sam.mutex.Unlock()
	
	cutoff := time.Now().Add(-24 * time.Hour) // 保留24小时内的告警
	
	for alertID, alert := range sam.alerts {
		if alert.Resolved && alert.ResolvedAt != nil && alert.ResolvedAt.Before(cutoff) {
			delete(sam.alerts, alertID)
		}
	}
}

// getAlertTypeFromRule 从规则获取告警类型
func (sam *SmartAlertManager) getAlertTypeFromRule(rule *AlertRule) AlertType {
	// 根据规则条件判断告警类型
	switch rule.Condition {
	case "error_rate":
		return AlertTypeErrorRate
	case "error_count":
		return AlertTypeErrorCount
	case "mttr":
		return AlertTypeMTTR
	case "mtbf":
		return AlertTypeMTBF
	default:
		return AlertTypeErrorRate
	}
}

// getPluginIDFromData 从数据获取插件ID
func (sam *SmartAlertManager) getPluginIDFromData(data map[string]interface{}) string {
	if pluginID, ok := data["plugin_id"].(string); ok {
		return pluginID
	}
	return "unknown"
}

// buildAlertMessage 构建告警消息
func (sam *SmartAlertManager) buildAlertMessage(rule *AlertRule, data map[string]interface{}) string {
	message := fmt.Sprintf("Alert: %s", rule.Name)
	if rule.Description != "" {
		message += fmt.Sprintf(" - %s", rule.Description)
	}
	
	if pluginID := sam.getPluginIDFromData(data); pluginID != "unknown" {
		message += fmt.Sprintf(" (Plugin: %s)", pluginID)
	}
	
	return message
}