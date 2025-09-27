package recovery

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// StrategyType 策略类型
type StrategyType int

const (
	// StrategyTypeCircuitBreaker 熔断器策略
	StrategyTypeCircuitBreaker StrategyType = iota
	// StrategyTypeRetry 重试策略
	StrategyTypeRetry
	// StrategyTypeFallback 降级策略
	StrategyTypeFallback
	// StrategyTypeAutoRecovery 自动恢复策略
	StrategyTypeAutoRecovery
)

// String 返回策略类型的字符串表示
func (st StrategyType) String() string {
	switch st {
	case StrategyTypeCircuitBreaker:
		return "circuit_breaker"
	case StrategyTypeRetry:
		return "retry"
	case StrategyTypeFallback:
		return "fallback"
	case StrategyTypeAutoRecovery:
		return "auto_recovery"
	default:
		return "unknown"
	}
}

// RecoveryStrategy 恢复策略接口
type RecoveryStrategy interface {
	// GetName 获取策略名称
	GetName() string
	// GetType 获取策略类型
	GetType() StrategyType
	// Execute 执行策略
	Execute(ctx context.Context, operation func(ctx context.Context) (interface{}, error), args interface{}) (interface{}, error)
	// Reset 重置策略
	Reset()
	// IsHealthy 检查策略是否健康
	IsHealthy() bool
}

// StrategyConfig 策略配置
type StrategyConfig struct {
	Name     string        `json:"name"`
	Type     StrategyType  `json:"type"`
	Enabled  bool          `json:"enabled"`
	Priority int           `json:"priority"`
	Config   interface{}   `json:"config"`
}

// RecoveryManagerConfig 恢复管理器配置
type RecoveryManagerConfig struct {
	// Enabled 是否启用恢复管理器
	Enabled bool `json:"enabled"`
	// MaxConcurrentRecoveries 最大并发恢复数
	MaxConcurrentRecoveries int `json:"max_concurrent_recoveries"`
	// RecoveryTimeout 恢复超时时间
	RecoveryTimeout time.Duration `json:"recovery_timeout"`
	// HealthCheckInterval 健康检查间隔
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	// MetricsRetentionPeriod 指标保留期间
	MetricsRetentionPeriod time.Duration `json:"metrics_retention_period"`
	// EnableMetrics 是否启用指标收集
	EnableMetrics bool `json:"enable_metrics"`
	// EnableHistory 是否启用历史记录
	EnableHistory bool `json:"enable_history"`
	// MaxHistorySize 最大历史记录数
	MaxHistorySize int `json:"max_history_size"`
}

// DefaultRecoveryManagerConfig 返回默认恢复管理器配置
func DefaultRecoveryManagerConfig() *RecoveryManagerConfig {
	return &RecoveryManagerConfig{
		Enabled:                 true,
		MaxConcurrentRecoveries: 10,
		RecoveryTimeout:         5 * time.Minute,
		HealthCheckInterval:     30 * time.Second,
		MetricsRetentionPeriod:  24 * time.Hour,
		EnableMetrics:           true,
		EnableHistory:           true,
		MaxHistorySize:          1000,
	}
}

// RecoveryEvent 恢复事件
type RecoveryEvent struct {
	ID          string                 `json:"id"`
	PluginID    string                 `json:"plugin_id"`
	StrategyName string                `json:"strategy_name"`
	EventType   string                 `json:"event_type"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        map[string]interface{} `json:"data"`
	Error       error                  `json:"error,omitempty"`
}

// RecoveryHistory 恢复历史记录
type RecoveryHistory struct {
	Events    []*RecoveryEvent `json:"events"`
	MaxSize   int              `json:"max_size"`
	mutex     sync.RWMutex
}

// NewRecoveryHistory 创建新的恢复历史记录
func NewRecoveryHistory(maxSize int) *RecoveryHistory {
	return &RecoveryHistory{
		Events:  make([]*RecoveryEvent, 0, maxSize),
		MaxSize: maxSize,
	}
}

// AddEvent 添加事件
func (rh *RecoveryHistory) AddEvent(event *RecoveryEvent) {
	rh.mutex.Lock()
	defer rh.mutex.Unlock()

	// 如果超过最大大小，移除最旧的事件
	if len(rh.Events) >= rh.MaxSize {
		rh.Events = rh.Events[1:]
	}

	rh.Events = append(rh.Events, event)
}

// GetEvents 获取事件列表
func (rh *RecoveryHistory) GetEvents() []*RecoveryEvent {
	rh.mutex.RLock()
	defer rh.mutex.RUnlock()

	// 返回事件的副本
	events := make([]*RecoveryEvent, len(rh.Events))
	copy(events, rh.Events)
	return events
}

// RecoveryManagerMetrics 恢复管理器指标
type RecoveryManagerMetrics struct {
	TotalStrategies       int           `json:"total_strategies"`
	ActiveStrategies      int           `json:"active_strategies"`
	TotalRecoveries       int64         `json:"total_recoveries"`
	SuccessfulRecoveries  int64         `json:"successful_recoveries"`
	FailedRecoveries      int64         `json:"failed_recoveries"`
	AverageRecoveryTime   time.Duration `json:"average_recovery_time"`
	ConcurrentRecoveries  int           `json:"concurrent_recoveries"`
	LastRecoveryTime      time.Time     `json:"last_recovery_time"`
	HealthyStrategies     int           `json:"healthy_strategies"`
	UnhealthyStrategies   int           `json:"unhealthy_strategies"`
}

// RecoveryManager 恢复策略管理器
type RecoveryManager struct {
	config   *RecoveryManagerConfig
	logger   *slog.Logger
	metrics  *RecoveryManagerMetrics
	history  *RecoveryHistory
	mutex    sync.RWMutex

	// 策略管理
	strategies       map[string]RecoveryStrategy
	strategiesByType map[StrategyType][]RecoveryStrategy
	strategiesMutex  sync.RWMutex

	// 并发控制
	concurrencyLimit chan struct{}

	// 控制通道
	stopChan chan struct{}
	done     chan struct{}

	// 回调函数
	onStrategyRegistered   func(strategy RecoveryStrategy)
	onStrategyUnregistered func(strategyName string)
	onRecoveryStarted      func(pluginID, strategyName string)
	onRecoveryCompleted    func(pluginID, strategyName string, success bool, duration time.Duration)
	onRecoveryFailed       func(pluginID, strategyName string, err error)
}

// NewRecoveryManager 创建新的恢复管理器
func NewRecoveryManager(config *RecoveryManagerConfig, logger *slog.Logger) *RecoveryManager {
	if config == nil {
		config = DefaultRecoveryManagerConfig()
	}

	rm := &RecoveryManager{
		config:           config,
		logger:           logger,
		metrics:          &RecoveryManagerMetrics{},
		strategies:       make(map[string]RecoveryStrategy),
		strategiesByType: make(map[StrategyType][]RecoveryStrategy),
		concurrencyLimit: make(chan struct{}, config.MaxConcurrentRecoveries),
		stopChan:         make(chan struct{}),
		done:             make(chan struct{}),
	}

	if config.EnableHistory {
		rm.history = NewRecoveryHistory(config.MaxHistorySize)
	}

	return rm
}

// Start 启动恢复管理器
func (rm *RecoveryManager) Start(ctx context.Context) error {
	if !rm.config.Enabled {
		if rm.logger != nil {
			rm.logger.Info("Recovery manager is disabled")
		}
		return nil
	}

	if rm.logger != nil {
		rm.logger.Info("Starting recovery manager",
			"max_concurrent_recoveries", rm.config.MaxConcurrentRecoveries,
			"health_check_interval", rm.config.HealthCheckInterval)
	}

	// 启动健康检查循环
	if rm.config.HealthCheckInterval > 0 {
		go rm.healthCheckLoop(ctx)
	}

	// 启动指标清理循环
	if rm.config.EnableMetrics && rm.config.MetricsRetentionPeriod > 0 {
		go rm.metricsCleanupLoop(ctx)
	}

	return nil
}

// Stop 停止恢复管理器
func (rm *RecoveryManager) Stop() error {
	close(rm.stopChan)
	<-rm.done

	if rm.logger != nil {
		rm.logger.Info("Recovery manager stopped")
	}

	return nil
}

// RegisterStrategy 注册恢复策略
func (rm *RecoveryManager) RegisterStrategy(strategy RecoveryStrategy) error {
	rm.strategiesMutex.Lock()
	defer rm.strategiesMutex.Unlock()

	name := strategy.GetName()
	if _, exists := rm.strategies[name]; exists {
		return fmt.Errorf("strategy with name '%s' already exists", name)
	}

	// 注册策略
	rm.strategies[name] = strategy

	// 按类型分组
	strategyType := strategy.GetType()
	rm.strategiesByType[strategyType] = append(rm.strategiesByType[strategyType], strategy)

	// 更新指标
	rm.updateStrategyMetrics()

	// 触发回调
	if rm.onStrategyRegistered != nil {
		rm.onStrategyRegistered(strategy)
	}

	// 记录历史
	if rm.history != nil {
		rm.history.AddEvent(&RecoveryEvent{
			ID:          fmt.Sprintf("reg_%d", time.Now().UnixNano()),
			StrategyName: name,
			EventType:   "strategy_registered",
			Timestamp:   time.Now(),
			Data: map[string]interface{}{
				"type": strategyType.String(),
			},
		})
	}

	if rm.logger != nil {
		rm.logger.Info("Recovery strategy registered",
			"name", name,
			"type", strategyType.String())
	}

	return nil
}

// UnregisterStrategy 取消注册恢复策略
func (rm *RecoveryManager) UnregisterStrategy(name string) error {
	rm.strategiesMutex.Lock()
	defer rm.strategiesMutex.Unlock()

	strategy, exists := rm.strategies[name]
	if !exists {
		return fmt.Errorf("strategy with name '%s' not found", name)
	}

	// 从策略映射中移除
	delete(rm.strategies, name)

	// 从类型分组中移除
	strategyType := strategy.GetType()
	strategies := rm.strategiesByType[strategyType]
	for i, s := range strategies {
		if s.GetName() == name {
			rm.strategiesByType[strategyType] = append(strategies[:i], strategies[i+1:]...)
			break
		}
	}

	// 更新指标
	rm.updateStrategyMetrics()

	// 触发回调
	if rm.onStrategyUnregistered != nil {
		rm.onStrategyUnregistered(name)
	}

	// 记录历史
	if rm.history != nil {
		rm.history.AddEvent(&RecoveryEvent{
			ID:          fmt.Sprintf("unreg_%d", time.Now().UnixNano()),
			StrategyName: name,
			EventType:   "strategy_unregistered",
			Timestamp:   time.Now(),
			Data: map[string]interface{}{
				"type": strategyType.String(),
			},
		})
	}

	if rm.logger != nil {
		rm.logger.Info("Recovery strategy unregistered", "name", name)
	}

	return nil
}

// ExecuteRecovery 执行恢复策略
func (rm *RecoveryManager) ExecuteRecovery(ctx context.Context, pluginID string, strategyNames []string, operation func(ctx context.Context) (interface{}, error), args interface{}) (interface{}, error) {
	if !rm.config.Enabled {
		return operation(ctx)
	}

	// 获取并发控制令牌
	select {
	case rm.concurrencyLimit <- struct{}{}:
		defer func() { <-rm.concurrencyLimit }()
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled while waiting for concurrency limit")
	}

	// 更新并发指标
	rm.mutex.Lock()
	rm.metrics.ConcurrentRecoveries++
	rm.mutex.Unlock()
	defer func() {
		rm.mutex.Lock()
		rm.metrics.ConcurrentRecoveries--
		rm.mutex.Unlock()
	}()

	// 按优先级排序策略
	strategies := rm.getStrategiesByNames(strategyNames)
	if len(strategies) == 0 {
		// 没有可用策略，直接执行操作
		return operation(ctx)
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, rm.config.RecoveryTimeout)
	defer cancel()

	// 按顺序尝试策略
	var lastErr error
	for _, strategy := range strategies {
		result, err := rm.executeWithStrategy(ctx, pluginID, strategy, operation, args)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("all recovery strategies failed, last error: %w", lastErr)
}

// executeWithStrategy 使用指定策略执行操作
func (rm *RecoveryManager) executeWithStrategy(ctx context.Context, pluginID string, strategy RecoveryStrategy, operation func(ctx context.Context) (interface{}, error), args interface{}) (interface{}, error) {
	startTime := time.Now()
	strategyName := strategy.GetName()

	// 触发恢复开始回调
	if rm.onRecoveryStarted != nil {
		rm.onRecoveryStarted(pluginID, strategyName)
	}

	if rm.logger != nil {
		rm.logger.Debug("Executing recovery strategy",
			"plugin_id", pluginID,
			"strategy", strategyName)
	}

	// 执行策略
	result, err := strategy.Execute(ctx, operation, args)
	duration := time.Since(startTime)

	// 更新指标
	rm.updateRecoveryMetrics(err == nil, duration)

	// 触发回调
	if rm.onRecoveryCompleted != nil {
		rm.onRecoveryCompleted(pluginID, strategyName, err == nil, duration)
	}

	if err != nil && rm.onRecoveryFailed != nil {
		rm.onRecoveryFailed(pluginID, strategyName, err)
	}

	// 记录历史
	if rm.history != nil {
		eventType := "recovery_success"
		if err != nil {
			eventType = "recovery_failed"
		}

		rm.history.AddEvent(&RecoveryEvent{
			ID:          fmt.Sprintf("rec_%d", time.Now().UnixNano()),
			PluginID:    pluginID,
			StrategyName: strategyName,
			EventType:   eventType,
			Timestamp:   time.Now(),
			Data: map[string]interface{}{
				"duration": duration.String(),
				"success":  err == nil,
			},
			Error: err,
		})
	}

	if err != nil {
		if rm.logger != nil {
			rm.logger.Error("Recovery strategy failed",
				"plugin_id", pluginID,
				"strategy", strategyName,
				"duration", duration,
				"error", err.Error())
		}
		return nil, err
	}

	if rm.logger != nil {
		rm.logger.Info("Recovery strategy succeeded",
			"plugin_id", pluginID,
			"strategy", strategyName,
			"duration", duration)
	}

	return result, nil
}

// getStrategiesByNames 根据名称获取策略列表
func (rm *RecoveryManager) getStrategiesByNames(names []string) []RecoveryStrategy {
	rm.strategiesMutex.RLock()
	defer rm.strategiesMutex.RUnlock()

	var strategies []RecoveryStrategy
	for _, name := range names {
		if strategy, exists := rm.strategies[name]; exists {
			strategies = append(strategies, strategy)
		}
	}

	return strategies
}

// GetStrategiesByType 根据类型获取策略列表
func (rm *RecoveryManager) GetStrategiesByType(strategyType StrategyType) []RecoveryStrategy {
	rm.strategiesMutex.RLock()
	defer rm.strategiesMutex.RUnlock()

	strategies := rm.strategiesByType[strategyType]
	result := make([]RecoveryStrategy, len(strategies))
	copy(result, strategies)

	return result
}

// GetStrategy 获取指定名称的策略
func (rm *RecoveryManager) GetStrategy(name string) (RecoveryStrategy, bool) {
	rm.strategiesMutex.RLock()
	defer rm.strategiesMutex.RUnlock()

	strategy, exists := rm.strategies[name]
	return strategy, exists
}

// GetAllStrategies 获取所有策略
func (rm *RecoveryManager) GetAllStrategies() map[string]RecoveryStrategy {
	rm.strategiesMutex.RLock()
	defer rm.strategiesMutex.RUnlock()

	strategies := make(map[string]RecoveryStrategy)
	for name, strategy := range rm.strategies {
		strategies[name] = strategy
	}

	return strategies
}

// healthCheckLoop 健康检查循环
func (rm *RecoveryManager) healthCheckLoop(ctx context.Context) {
	defer close(rm.done)

	ticker := time.NewTicker(rm.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-rm.stopChan:
			return
		case <-ticker.C:
			rm.performHealthChecks()
		}
	}
}

// performHealthChecks 执行健康检查
func (rm *RecoveryManager) performHealthChecks() {
	rm.strategiesMutex.RLock()
	strategies := make([]RecoveryStrategy, 0, len(rm.strategies))
	for _, strategy := range rm.strategies {
		strategies = append(strategies, strategy)
	}
	rm.strategiesMutex.RUnlock()

	healthyCount := 0
	unhealthyCount := 0

	for _, strategy := range strategies {
		if strategy.IsHealthy() {
			healthyCount++
		} else {
			unhealthyCount++
			if rm.logger != nil {
				rm.logger.Warn("Unhealthy recovery strategy detected",
					"strategy", strategy.GetName(),
					"type", strategy.GetType().String())
			}
		}
	}

	// 更新健康状态指标
	rm.mutex.Lock()
	rm.metrics.HealthyStrategies = healthyCount
	rm.metrics.UnhealthyStrategies = unhealthyCount
	rm.mutex.Unlock()
}

// metricsCleanupLoop 指标清理循环
func (rm *RecoveryManager) metricsCleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(rm.config.MetricsRetentionPeriod / 10) // 每10分之一保留期间清理一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-rm.stopChan:
			return
		case <-ticker.C:
			rm.cleanupOldMetrics()
		}
	}
}

// cleanupOldMetrics 清理旧指标
func (rm *RecoveryManager) cleanupOldMetrics() {
	if rm.history == nil {
		return
	}

	cutoff := time.Now().Add(-rm.config.MetricsRetentionPeriod)
	events := rm.history.GetEvents()

	// 过滤掉过期的事件
	var validEvents []*RecoveryEvent
	for _, event := range events {
		if event.Timestamp.After(cutoff) {
			validEvents = append(validEvents, event)
		}
	}

	// 更新历史记录
	rm.history.mutex.Lock()
	rm.history.Events = validEvents
	rm.history.mutex.Unlock()
}

// updateStrategyMetrics 更新策略指标
func (rm *RecoveryManager) updateStrategyMetrics() {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	rm.metrics.TotalStrategies = len(rm.strategies)

	activeCount := 0
	for _, strategy := range rm.strategies {
		if strategy.IsHealthy() {
			activeCount++
		}
	}
	rm.metrics.ActiveStrategies = activeCount
}

// updateRecoveryMetrics 更新恢复指标
func (rm *RecoveryManager) updateRecoveryMetrics(success bool, duration time.Duration) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	rm.metrics.TotalRecoveries++
	rm.metrics.LastRecoveryTime = time.Now()

	if success {
		rm.metrics.SuccessfulRecoveries++
	} else {
		rm.metrics.FailedRecoveries++
	}

	// 更新平均恢复时间
	if rm.metrics.SuccessfulRecoveries > 0 {
		totalTime := time.Duration(rm.metrics.AverageRecoveryTime.Nanoseconds()*int64(rm.metrics.SuccessfulRecoveries-1)) + duration
		rm.metrics.AverageRecoveryTime = totalTime / time.Duration(rm.metrics.SuccessfulRecoveries)
	}
}

// GetMetrics 获取指标
func (rm *RecoveryManager) GetMetrics() *RecoveryManagerMetrics {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	// 返回指标的副本
	metrics := *rm.metrics
	return &metrics
}

// GetHistory 获取历史记录
func (rm *RecoveryManager) GetHistory() []*RecoveryEvent {
	if rm.history == nil {
		return nil
	}
	return rm.history.GetEvents()
}

// ResetAllStrategies 重置所有策略
func (rm *RecoveryManager) ResetAllStrategies() {
	rm.strategiesMutex.RLock()
	strategies := make([]RecoveryStrategy, 0, len(rm.strategies))
	for _, strategy := range rm.strategies {
		strategies = append(strategies, strategy)
	}
	rm.strategiesMutex.RUnlock()

	for _, strategy := range strategies {
		strategy.Reset()
	}

	if rm.logger != nil {
		rm.logger.Info("All recovery strategies reset")
	}
}

// SetOnStrategyRegistered 设置策略注册回调
func (rm *RecoveryManager) SetOnStrategyRegistered(callback func(strategy RecoveryStrategy)) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	rm.onStrategyRegistered = callback
}

// SetOnStrategyUnregistered 设置策略取消注册回调
func (rm *RecoveryManager) SetOnStrategyUnregistered(callback func(strategyName string)) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	rm.onStrategyUnregistered = callback
}

// SetOnRecoveryStarted 设置恢复开始回调
func (rm *RecoveryManager) SetOnRecoveryStarted(callback func(pluginID, strategyName string)) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	rm.onRecoveryStarted = callback
}

// SetOnRecoveryCompleted 设置恢复完成回调
func (rm *RecoveryManager) SetOnRecoveryCompleted(callback func(pluginID, strategyName string, success bool, duration time.Duration)) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	rm.onRecoveryCompleted = callback
}

// SetOnRecoveryFailed 设置恢复失败回调
func (rm *RecoveryManager) SetOnRecoveryFailed(callback func(pluginID, strategyName string, err error)) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	rm.onRecoveryFailed = callback
}

// GetSuccessRate 获取成功率
func (rm *RecoveryManager) GetSuccessRate() float64 {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	if rm.metrics.TotalRecoveries == 0 {
		return 0.0
	}

	return float64(rm.metrics.SuccessfulRecoveries) / float64(rm.metrics.TotalRecoveries)
}

// GetHealthyStrategyRate 获取健康策略比率
func (rm *RecoveryManager) GetHealthyStrategyRate() float64 {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	if rm.metrics.TotalStrategies == 0 {
		return 0.0
	}

	return float64(rm.metrics.HealthyStrategies) / float64(rm.metrics.TotalStrategies)
}