package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ErrorPropagator 错误传播器接口
type ErrorPropagator interface {
	// PropagateError 传播错误
	PropagateError(ctx context.Context, err PluginError, source string, targets []string) error
	// RegisterPropagationHandler 注册传播处理器
	RegisterPropagationHandler(handler PropagationHandler) error
	// UnregisterPropagationHandler 注销传播处理器
	UnregisterPropagationHandler(handlerID string) error
	// SetPropagationRules 设置传播规则
	SetPropagationRules(rules []PropagationRule) error
}

// PropagationHandler 传播处理器接口
type PropagationHandler interface {
	// GetID 获取处理器ID
	GetID() string
	// CanHandle 是否可以处理该错误
	CanHandle(err PluginError) bool
	// Handle 处理错误传播
	Handle(ctx context.Context, err PluginError, source string, targets []string) error
	// GetPriority 获取优先级
	GetPriority() int
}

// PropagationRule 传播规则
type PropagationRule struct {
	ID          string                 `json:"id"`          // 规则ID
	Name        string                 `json:"name"`        // 规则名称
	Condition   PropagationCondition   `json:"condition"`   // 传播条件
	Action      PropagationAction      `json:"action"`      // 传播动作
	Targets     []string               `json:"targets"`     // 目标插件
	Enabled     bool                   `json:"enabled"`     // 是否启用
	Priority    int                    `json:"priority"`    // 优先级
	Timeout     time.Duration          `json:"timeout"`     // 超时时间
	RetryCount  int                    `json:"retry_count"` // 重试次数
	RetryDelay  time.Duration          `json:"retry_delay"` // 重试延迟
	Metadata    map[string]interface{} `json:"metadata"`    // 元数据
}

// PropagationCondition 传播条件
type PropagationCondition struct {
	ErrorCodes     []ErrorCode     `json:"error_codes"`     // 错误代码
	ErrorTypes     []ErrorType     `json:"error_types"`     // 错误类型
	SeverityLevels []ErrorSeverity `json:"severity_levels"` // 严重程度
	SourcePlugins  []string        `json:"source_plugins"`  // 源插件
	TimeWindow     time.Duration   `json:"time_window"`     // 时间窗口
	Frequency      int             `json:"frequency"`       // 频率阈值
}

// PropagationAction 传播动作
type PropagationAction struct {
	Type       PropagationActionType  `json:"type"`       // 动作类型
	Parameters map[string]interface{} `json:"parameters"` // 参数
	Async      bool                   `json:"async"`      // 是否异步
	Batch      bool                   `json:"batch"`      // 是否批处理
	Transform  bool                   `json:"transform"`  // 是否转换错误
}

// PropagationActionType 传播动作类型
type PropagationActionType int

const (
	PropagationActionNotify PropagationActionType = iota
	PropagationActionRestart
	PropagationActionStop
	PropagationActionIsolate
	PropagationActionDegrade
	PropagationActionLog
	PropagationActionMetrics
	PropagationActionAlert
)

// String 返回传播动作类型的字符串表示
func (p PropagationActionType) String() string {
	switch p {
	case PropagationActionNotify:
		return "notify"
	case PropagationActionRestart:
		return "restart"
	case PropagationActionStop:
		return "stop"
	case PropagationActionIsolate:
		return "isolate"
	case PropagationActionDegrade:
		return "degrade"
	case PropagationActionLog:
		return "log"
	case PropagationActionMetrics:
		return "metrics"
	case PropagationActionAlert:
		return "alert"
	default:
		return "unknown"
	}
}

// DefaultErrorPropagator 默认错误传播器
type DefaultErrorPropagator struct {
	handlers []PropagationHandler
	rules    []PropagationRule
	mutex    sync.RWMutex
	logger   Logger
	metrics  MetricsCollector
}

// NewErrorPropagator 创建新的错误传播器
func NewErrorPropagator(logger Logger, metrics MetricsCollector) ErrorPropagator {
	return &DefaultErrorPropagator{
		handlers: make([]PropagationHandler, 0),
		rules:    make([]PropagationRule, 0),
		logger:   logger,
		metrics:  metrics,
	}
}

// PropagateError 传播错误
func (p *DefaultErrorPropagator) PropagateError(ctx context.Context, err PluginError, source string, targets []string) error {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	// 记录传播开始
	p.logger.Info("Starting error propagation", map[string]interface{}{
		"error_code": err.GetCode(),
		"source":     source,
		"targets":    targets,
	})
	
	// 应用传播规则
	applicableRules := p.getApplicableRules(err, source)
	if len(applicableRules) == 0 {
		p.logger.Debug("No applicable propagation rules found")
		return nil
	}
	
	// 执行传播
	for _, rule := range applicableRules {
		if err := p.executeRule(ctx, rule, err, source, targets); err != nil {
			p.logger.Error("Failed to execute propagation rule", map[string]interface{}{
				"rule_id": rule.ID,
				"error":   err.Error(),
			})
			continue
		}
	}
	
	// 使用处理器传播
	for _, handler := range p.handlers {
		if handler.CanHandle(err) {
			if handlerErr := handler.Handle(ctx, err, source, targets); handlerErr != nil {
				p.logger.Error("Propagation handler failed", map[string]interface{}{
					"handler_id": handler.GetID(),
					"error":      handlerErr.Error(),
				})
			}
		}
	}
	
	// 更新指标
	if p.metrics != nil {
		p.metrics.IncrementCounter("error_propagation_total", map[string]string{
			"source": source,
			"type":   err.GetType().String(),
		})
	}
	
	return nil
}

// RegisterPropagationHandler 注册传播处理器
func (p *DefaultErrorPropagator) RegisterPropagationHandler(handler PropagationHandler) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	// 检查是否已存在
	for _, h := range p.handlers {
		if h.GetID() == handler.GetID() {
			return fmt.Errorf("handler with ID %s already exists", handler.GetID())
		}
	}
	
	// 按优先级插入
	inserted := false
	for i, h := range p.handlers {
		if handler.GetPriority() > h.GetPriority() {
			p.handlers = append(p.handlers[:i], append([]PropagationHandler{handler}, p.handlers[i:]...)...)
			inserted = true
			break
		}
	}
	
	if !inserted {
		p.handlers = append(p.handlers, handler)
	}
	
	p.logger.Info("Registered propagation handler", map[string]interface{}{
		"handler_id": handler.GetID(),
		"priority":   handler.GetPriority(),
	})
	
	return nil
}

// UnregisterPropagationHandler 注销传播处理器
func (p *DefaultErrorPropagator) UnregisterPropagationHandler(handlerID string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	for i, handler := range p.handlers {
		if handler.GetID() == handlerID {
			p.handlers = append(p.handlers[:i], p.handlers[i+1:]...)
			p.logger.Info("Unregistered propagation handler", map[string]interface{}{
				"handler_id": handlerID,
			})
			return nil
		}
	}
	
	return fmt.Errorf("handler with ID %s not found", handlerID)
}

// SetPropagationRules 设置传播规则
func (p *DefaultErrorPropagator) SetPropagationRules(rules []PropagationRule) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	p.rules = make([]PropagationRule, len(rules))
	copy(p.rules, rules)
	
	p.logger.Info("Updated propagation rules", map[string]interface{}{
		"rule_count": len(rules),
	})
	
	return nil
}

// getApplicableRules 获取适用的规则
func (p *DefaultErrorPropagator) getApplicableRules(err PluginError, source string) []PropagationRule {
	var applicableRules []PropagationRule
	
	for _, rule := range p.rules {
		if !rule.Enabled {
			continue
		}
		
		if p.matchesCondition(rule.Condition, err, source) {
			applicableRules = append(applicableRules, rule)
		}
	}
	
	return applicableRules
}

// matchesCondition 检查是否匹配条件
func (p *DefaultErrorPropagator) matchesCondition(condition PropagationCondition, err PluginError, source string) bool {
	// 检查错误代码
	if len(condition.ErrorCodes) > 0 {
		matched := false
		for _, code := range condition.ErrorCodes {
			if err.GetCode() == code {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	
	// 检查错误类型
	if len(condition.ErrorTypes) > 0 {
		matched := false
		for _, errType := range condition.ErrorTypes {
			if err.GetType() == errType {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	
	// 检查严重程度
	if len(condition.SeverityLevels) > 0 {
		matched := false
		for _, severity := range condition.SeverityLevels {
			if err.GetSeverity() == severity {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	
	// 检查源插件
	if len(condition.SourcePlugins) > 0 {
		matched := false
		for _, plugin := range condition.SourcePlugins {
			if source == plugin {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	
	return true
}

// executeRule 执行规则
func (p *DefaultErrorPropagator) executeRule(ctx context.Context, rule PropagationRule, err PluginError, source string, targets []string) error {
	// 创建带超时的上下文
	ruleCtx := ctx
	if rule.Timeout > 0 {
		var cancel context.CancelFunc
		ruleCtx, cancel = context.WithTimeout(ctx, rule.Timeout)
		defer cancel()
	}
	
	// 确定目标
	ruleTargets := targets
	if len(rule.Targets) > 0 {
		ruleTargets = rule.Targets
	}
	
	// 执行动作
	return p.executeAction(ruleCtx, rule.Action, err, source, ruleTargets)
}

// executeAction 执行动作
func (p *DefaultErrorPropagator) executeAction(ctx context.Context, action PropagationAction, err PluginError, source string, targets []string) error {
	switch action.Type {
	case PropagationActionLog:
		p.logger.Error("Propagated error", map[string]interface{}{
			"error_code": err.GetCode(),
			"source":     source,
			"targets":    targets,
			"message":    err.Error(),
		})
	
	case PropagationActionMetrics:
		if p.metrics != nil {
			p.metrics.IncrementCounter("error_propagated", map[string]string{
				"source": source,
				"type":   err.GetType().String(),
				"code":   err.GetCode().String(),
			})
		}
	
	case PropagationActionNotify:
		// 通知目标插件
		for _, target := range targets {
			p.logger.Info("Notifying target plugin", map[string]interface{}{
				"target": target,
				"error":  err.Error(),
			})
		}
	
	default:
		p.logger.Warn("Unknown propagation action type", map[string]interface{}{
			"action_type": action.Type,
		})
	}
	
	return nil
}

// NotificationPropagationHandler 通知传播处理器
type NotificationPropagationHandler struct {
	id       string
	priority int
	logger   Logger
}

// NewNotificationPropagationHandler 创建通知传播处理器
func NewNotificationPropagationHandler(id string, priority int, logger Logger) PropagationHandler {
	return &NotificationPropagationHandler{
		id:       id,
		priority: priority,
		logger:   logger,
	}
}

// GetID 获取处理器ID
func (h *NotificationPropagationHandler) GetID() string {
	return h.id
}

// CanHandle 是否可以处理该错误
func (h *NotificationPropagationHandler) CanHandle(err PluginError) bool {
	// 处理所有错误
	return true
}

// Handle 处理错误传播
func (h *NotificationPropagationHandler) Handle(ctx context.Context, err PluginError, source string, targets []string) error {
	h.logger.Info("Handling error propagation via notification", map[string]interface{}{
		"error_code": err.GetCode(),
		"source":     source,
		"targets":    targets,
	})
	
	// 这里可以实现实际的通知逻辑
	// 例如发送事件、调用回调等
	
	return nil
}

// GetPriority 获取优先级
func (h *NotificationPropagationHandler) GetPriority() int {
	return h.priority
}