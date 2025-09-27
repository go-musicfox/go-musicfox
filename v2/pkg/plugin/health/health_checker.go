// pkg/plugin/health_checker.go
package plugin

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"

	core "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
)

// DefaultHealthChecker 默认健康检查器实现
type DefaultHealthChecker struct {
	plugin   core.Plugin
	config   *ExtendedHealthCheckConfig
	logger   *slog.Logger
	
	// 状态管理
	currentStatus HealthStatus
	lastResult    *HealthCheckResult
	history       []*HealthCheckResult
	historyMutex  sync.RWMutex
	
	// 运行控制
	ctx       context.Context
	cancel    context.CancelFunc
	running   bool
	runMutex  sync.RWMutex
	
	// 回调和策略
	statusCallbacks    []func(HealthStatus, HealthStatus, *HealthCheckResult)
	metricsCollectors  map[string]MetricsCollector
	recoveryStrategies map[string]RecoveryStrategy
	healthStrategies   map[string]HealthCheckStrategy
	mutex              sync.RWMutex
	
	// 指标统计
	startTime     time.Time
	checkCount    int64
	errorCount    int64
	lastCheckTime time.Time
	statsMutex    sync.RWMutex
}

// NewHealthChecker 创建新的健康检查器
func NewDefaultHealthChecker(plugin core.Plugin, config *ExtendedHealthCheckConfig, logger *slog.Logger) *DefaultHealthChecker {
	
	hc := &DefaultHealthChecker{
		plugin:             plugin,
		config:             config,
		logger:             logger,
		currentStatus:      HealthStatusUnknown,
		history:            make([]*HealthCheckResult, 0),
		statusCallbacks:    make([]func(HealthStatus, HealthStatus, *HealthCheckResult), 0),
		metricsCollectors:  make(map[string]MetricsCollector),
		recoveryStrategies: make(map[string]RecoveryStrategy),
		healthStrategies:   make(map[string]HealthCheckStrategy),
		startTime:          time.Now(),
	}
	
	// 注册默认的健康检查策略
	hc.registerDefaultStrategies()
	
	// 注册默认的指标收集器
	hc.registerDefaultCollectors()
	
	// 注册默认的恢复策略
	hc.registerDefaultRecoveryStrategies()
	
	return hc
}

// Start 启动健康检查
func (hc *DefaultHealthChecker) Start(ctx context.Context) error {
	hc.runMutex.Lock()
	defer hc.runMutex.Unlock()
	
	if hc.running {
		return fmt.Errorf("health checker already running")
	}
	
	hc.ctx, hc.cancel = context.WithCancel(ctx)
	hc.running = true
	hc.startTime = time.Now()
	
	// 启动健康检查循环
	go hc.checkLoop()
	
	hc.logger.Info("Health checker started",
		"plugin", hc.plugin.GetInfo().Name,
		"interval", hc.config.CheckInterval)
	
	return nil
}

// Stop 停止健康检查
func (hc *DefaultHealthChecker) Stop() error {
	hc.runMutex.Lock()
	defer hc.runMutex.Unlock()
	
	if !hc.running {
		return nil
	}
	
	hc.cancel()
	hc.running = false
	
	hc.logger.Info("Health checker stopped",
		"plugin", hc.plugin.GetInfo().Name)
	
	return nil
}

// CheckHealth 执行单次健康检查（别名方法）
func (hc *DefaultHealthChecker) CheckHealth(ctx context.Context) (*HealthCheckResult, error) {
	return hc.Check(ctx)
}

// CheckHealthWithStrategy 使用指定策略执行健康检查
func (hc *DefaultHealthChecker) CheckHealthWithStrategy(ctx context.Context, strategy string) (*HealthCheckResult, error) {
	hc.mutex.RLock()
	strategyImpl, exists := hc.healthStrategies[strategy]
	hc.mutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("health strategy '%s' not found", strategy)
	}
	
	startTime := time.Now()
	
	// 创建检查上下文
	checkCtx, cancel := context.WithTimeout(ctx, hc.config.Timeout)
	defer cancel()
	
	// 使用指定策略执行检查
	result, err := strategyImpl.Check(checkCtx, hc.plugin)
	if err != nil {
		return &HealthCheckResult{
			Status:    HealthStatusCritical,
			Message:   fmt.Sprintf("Strategy '%s' failed: %v", strategy, err),
			CheckedAt: startTime,
			Duration:  time.Since(startTime),
		}, err
	}
	
	// 更新检查时间和持续时间
	result.CheckedAt = startTime
	result.Duration = time.Since(startTime)
	
	// 更新状态
	hc.updateStatus(result)
	
	// 添加到历史记录
	hc.addToHistory(result)
	
	return result, nil
}

// CollectMetrics 收集指定类型的指标
func (hc *DefaultHealthChecker) CollectMetrics(ctx context.Context, collectorName string) (map[string]interface{}, error) {
	hc.mutex.RLock()
	collector, exists := hc.metricsCollectors[collectorName]
	hc.mutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("metrics collector '%s' not found", collectorName)
	}
	
	return collector.Collect(ctx)
}

// Recover 执行恢复操作
func (hc *DefaultHealthChecker) Recover(ctx context.Context, strategyName string, result *HealthCheckResult) error {
	hc.mutex.RLock()
	strategy, exists := hc.recoveryStrategies[strategyName]
	hc.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("recovery strategy '%s' not found", strategyName)
	}

	// 检查是否可以恢复
	if !strategy.CanRecover(result) {
		return fmt.Errorf("strategy '%s' cannot recover from current state", strategyName)
	}

	// 执行恢复
	return strategy.Recover(ctx, hc.plugin, result)
}

// Check 执行单次健康检查
func (hc *DefaultHealthChecker) Check(ctx context.Context) (*HealthCheckResult, error) {
	startTime := time.Now()
	
	// 更新统计信息
	hc.statsMutex.Lock()
	hc.checkCount++
	hc.lastCheckTime = startTime
	hc.statsMutex.Unlock()
	
	// 创建检查上下文
	checkCtx, cancel := context.WithTimeout(ctx, hc.config.Timeout)
	defer cancel()
	
	// 收集基础指标
	metrics, err := hc.collectMetrics(checkCtx)
	if err != nil {
		hc.statsMutex.Lock()
		hc.errorCount++
		hc.statsMutex.Unlock()
		
		return &HealthCheckResult{
			Status:    HealthStatusCritical,
			Message:   fmt.Sprintf("Failed to collect metrics: %v", err),
			CheckedAt: startTime,
			Duration:  time.Since(startTime),
		}, err
	}
	
	// 执行健康检查策略
	status, message, details := hc.executeHealthStrategies(checkCtx, metrics)
	
	// 创建检查结果
	result := &HealthCheckResult{
		Status:    status,
		Message:   message,
		Metrics:   metrics,
		Details:   details,
		CheckedAt: startTime,
		Duration:  time.Since(startTime),
	}
	
	// 更新状态
	hc.updateStatus(result)
	
	// 添加到历史记录
	hc.addToHistory(result)
	
	return result, nil
}

// GetStatus 获取当前健康状态
func (hc *DefaultHealthChecker) GetStatus() HealthStatus {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()
	return hc.currentStatus
}

// GetMetrics 获取最新的健康指标
func (hc *DefaultHealthChecker) GetMetrics() *HealthMetrics {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()
	
	if hc.lastResult != nil {
		return hc.lastResult.Metrics
	}
	return nil
}

// GetHistory 获取健康检查历史
func (hc *DefaultHealthChecker) GetHistory(limit int) []*HealthCheckResult {
	hc.historyMutex.RLock()
	defer hc.historyMutex.RUnlock()
	
	if limit <= 0 || limit > len(hc.history) {
		limit = len(hc.history)
	}
	
	// 返回最近的记录
	start := len(hc.history) - limit
	result := make([]*HealthCheckResult, limit)
	copy(result, hc.history[start:])
	
	return result
}

// UpdateConfig 更新配置
func (hc *DefaultHealthChecker) UpdateConfig(config *ExtendedHealthCheckConfig) error {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	
	hc.config = config
	
	hc.logger.Info("Health checker config updated",
		"plugin", hc.plugin.GetInfo().Name,
		"interval", config.CheckInterval)
	
	return nil
}

// OnStatusChange 注册状态变化回调
func (hc *DefaultHealthChecker) OnStatusChange(callback func(HealthStatus, HealthStatus, *HealthCheckResult)) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	
	hc.statusCallbacks = append(hc.statusCallbacks, callback)
}

// RegisterMetricsCollector 注册指标收集器
func (hc *DefaultHealthChecker) RegisterMetricsCollector(name string, collector MetricsCollector) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	
	hc.metricsCollectors[name] = collector
	
	hc.logger.Debug("Metrics collector registered",
		"name", name,
		"plugin", hc.plugin.GetInfo().Name)
}

// RegisterRecoveryStrategy 注册恢复策略
func (hc *DefaultHealthChecker) RegisterRecoveryStrategy(name string, strategy RecoveryStrategy) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	
	hc.recoveryStrategies[name] = strategy
	
	hc.logger.Debug("Recovery strategy registered",
		"name", name,
		"plugin", hc.plugin.GetInfo().Name)
}

// checkLoop 健康检查循环
func (hc *DefaultHealthChecker) checkLoop() {
	ticker := time.NewTicker(hc.config.CheckInterval)
	defer ticker.Stop()
	
	// 立即执行一次检查
	if _, err := hc.Check(hc.ctx); err != nil {
		hc.logger.Error("Initial health check failed",
			"plugin", hc.plugin.GetInfo().Name,
			"error", err)
	}
	
	for {
		select {
		case <-hc.ctx.Done():
			return
		case <-ticker.C:
			if _, err := hc.Check(hc.ctx); err != nil {
				hc.logger.Error("Health check failed",
					"plugin", hc.plugin.GetInfo().Name,
					"error", err)
			}
		}
	}
}

// collectMetrics 收集健康指标
func (hc *DefaultHealthChecker) collectMetrics(ctx context.Context) (*HealthMetrics, error) {
	metrics := &HealthMetrics{
		Timestamp: time.Now(),
		Uptime:    time.Since(hc.startTime),
	}
	
	// 收集基础系统指标
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	metrics.MemoryUsage = int64(m.Alloc)
	metrics.Goroutines = runtime.NumGoroutine()
	
	// 收集自定义指标
	hc.mutex.RLock()
	collectors := make(map[string]MetricsCollector)
	for name, collector := range hc.metricsCollectors {
		collectors[name] = collector
	}
	hc.mutex.RUnlock()
	
	for name, collector := range collectors {
		if customMetrics, err := collector.Collect(ctx); err == nil {
			// 将自定义指标合并到主指标中
			if cpuUsage, ok := customMetrics["cpu_usage"].(float64); ok {
				metrics.CPUUsage = cpuUsage
			}
			if responseTime, ok := customMetrics["response_time"].(time.Duration); ok {
				metrics.ResponseTime = responseTime
			}
			if throughput, ok := customMetrics["throughput"].(float64); ok {
				metrics.Throughput = throughput
			}
			if errorRate, ok := customMetrics["error_rate"].(float64); ok {
				metrics.ErrorRate = errorRate
			}
		} else {
			hc.logger.Warn("Failed to collect custom metrics",
				"collector", name,
				"error", err)
		}
	}
	
	// 计算成功率
	hc.statsMutex.RLock()
	if hc.checkCount > 0 {
		metrics.SuccessRate = float64(hc.checkCount-hc.errorCount) / float64(hc.checkCount)
	} else {
		metrics.SuccessRate = 1.0
	}
	hc.statsMutex.RUnlock()
	
	return metrics, nil
}

// executeHealthStrategies 执行健康检查策略
func (hc *DefaultHealthChecker) executeHealthStrategies(ctx context.Context, metrics *HealthMetrics) (HealthStatus, string, map[string]interface{}) {
	status := HealthStatusHealthy
	messages := make([]string, 0)
	details := make(map[string]interface{})
	
	// 检查阈值
	thresholds := hc.config.Thresholds
	
	// CPU检查
	if metrics.CPUUsage >= thresholds.CPUCritical {
		status = HealthStatusCritical
		messages = append(messages, fmt.Sprintf("CPU usage critical: %.2f%%", metrics.CPUUsage))
		details["cpu_status"] = "critical"
	} else if metrics.CPUUsage >= thresholds.CPUWarning {
		if status < HealthStatusDegraded {
			status = HealthStatusDegraded
		}
		messages = append(messages, fmt.Sprintf("CPU usage high: %.2f%%", metrics.CPUUsage))
		details["cpu_status"] = "warning"
	} else {
		details["cpu_status"] = "healthy"
	}
	
	// 内存检查
	if metrics.MemoryUsage >= thresholds.MemoryCritical {
		status = HealthStatusCritical
		messages = append(messages, fmt.Sprintf("Memory usage critical: %d bytes", metrics.MemoryUsage))
		details["memory_status"] = "critical"
	} else if metrics.MemoryUsage >= thresholds.MemoryWarning {
		if status < HealthStatusDegraded {
			status = HealthStatusDegraded
		}
		messages = append(messages, fmt.Sprintf("Memory usage high: %d bytes", metrics.MemoryUsage))
		details["memory_status"] = "warning"
	} else {
		details["memory_status"] = "healthy"
	}
	
	// 响应时间检查
	if metrics.ResponseTime >= thresholds.ResponseTimeCritical {
		status = HealthStatusCritical
		messages = append(messages, fmt.Sprintf("Response time critical: %v", metrics.ResponseTime))
		details["response_time_status"] = "critical"
	} else if metrics.ResponseTime >= thresholds.ResponseTimeWarning {
		if status < HealthStatusDegraded {
			status = HealthStatusDegraded
		}
		messages = append(messages, fmt.Sprintf("Response time high: %v", metrics.ResponseTime))
		details["response_time_status"] = "warning"
	} else {
		details["response_time_status"] = "healthy"
	}
	
	// 错误率检查
	if metrics.ErrorRate >= thresholds.ErrorRateCritical {
		status = HealthStatusCritical
		messages = append(messages, fmt.Sprintf("Error rate critical: %.2f%%", metrics.ErrorRate*100))
		details["error_rate_status"] = "critical"
	} else if metrics.ErrorRate >= thresholds.ErrorRateWarning {
		if status < HealthStatusDegraded {
			status = HealthStatusDegraded
		}
		messages = append(messages, fmt.Sprintf("Error rate high: %.2f%%", metrics.ErrorRate*100))
		details["error_rate_status"] = "warning"
	} else {
		details["error_rate_status"] = "healthy"
	}
	
	// 协程数检查
	if metrics.Goroutines >= thresholds.GoroutineCritical {
		status = HealthStatusCritical
		messages = append(messages, fmt.Sprintf("Goroutine count critical: %d", metrics.Goroutines))
		details["goroutine_status"] = "critical"
	} else if metrics.Goroutines >= thresholds.GoroutineWarning {
		if status < HealthStatusDegraded {
			status = HealthStatusDegraded
		}
		messages = append(messages, fmt.Sprintf("Goroutine count high: %d", metrics.Goroutines))
		details["goroutine_status"] = "warning"
	} else {
		details["goroutine_status"] = "healthy"
	}
	
	// 执行插件自身的健康检查
	if err := hc.plugin.HealthCheck(); err != nil {
		status = HealthStatusUnhealthy
		messages = append(messages, fmt.Sprintf("Plugin health check failed: %v", err))
		details["plugin_health"] = "failed"
	} else {
		details["plugin_health"] = "healthy"
	}
	
	// 构建消息
	var message string
	if len(messages) == 0 {
		message = "All health checks passed"
	} else {
		message = fmt.Sprintf("Health issues detected: %v", messages)
	}
	
	return status, message, details
}

// updateStatus 更新健康状态
func (hc *DefaultHealthChecker) updateStatus(result *HealthCheckResult) {
	hc.mutex.Lock()
	oldStatus := hc.currentStatus
	newStatus := result.Status
	hc.currentStatus = newStatus
	hc.lastResult = result
	callbacks := make([]func(HealthStatus, HealthStatus, *HealthCheckResult), len(hc.statusCallbacks))
	copy(callbacks, hc.statusCallbacks)
	hc.mutex.Unlock()
	
	// 如果状态发生变化，触发回调和恢复策略
	if oldStatus != newStatus {
		hc.logger.Info("Plugin health status changed",
			"plugin", hc.plugin.GetInfo().Name,
			"old_status", oldStatus.String(),
			"new_status", newStatus.String(),
			"message", result.Message)
		
		// 触发状态变化回调
		for _, callback := range callbacks {
			go func(cb func(HealthStatus, HealthStatus, *HealthCheckResult)) {
				defer func() {
					if r := recover(); r != nil {
						hc.logger.Error("Health status callback panicked",
							"plugin", hc.plugin.GetInfo().Name,
							"panic", r)
					}
				}()
				cb(oldStatus, newStatus, result)
			}(callback)
		}
		
		// 如果状态变为不健康且启用了自动恢复，尝试恢复
		if newStatus != HealthStatusHealthy && hc.config.RecoveryEnabled {
			go hc.attemptRecovery(result)
		}
	}
}

// addToHistory 添加到历史记录
func (hc *DefaultHealthChecker) addToHistory(result *HealthCheckResult) {
	hc.historyMutex.Lock()
	defer hc.historyMutex.Unlock()
	
	hc.history = append(hc.history, result)
	
	// 限制历史记录数量
	maxHistory := 100
	if len(hc.history) > maxHistory {
		hc.history = hc.history[len(hc.history)-maxHistory:]
	}
}

// attemptRecovery 尝试恢复
func (hc *DefaultHealthChecker) attemptRecovery(result *HealthCheckResult) {
	hc.mutex.RLock()
	strategies := make(map[string]RecoveryStrategy)
	for name, strategy := range hc.recoveryStrategies {
		strategies[name] = strategy
	}
	hc.mutex.RUnlock()
	
	for name, strategy := range strategies {
		if strategy.CanRecover(result) {
			hc.logger.Info("Attempting recovery",
				"plugin", hc.plugin.GetInfo().Name,
				"strategy", name)
			
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := strategy.Recover(ctx, hc.plugin, result); err != nil {
				hc.logger.Error("Recovery failed",
					"plugin", hc.plugin.GetInfo().Name,
					"strategy", name,
					"error", err)
			} else {
				hc.logger.Info("Recovery successful",
					"plugin", hc.plugin.GetInfo().Name,
					"strategy", name)
				
				// 恢复成功后立即进行一次健康检查
				go func() {
					time.Sleep(5 * time.Second)
					hc.Check(context.Background())
				}()
			}
			cancel()
			return
		}
	}
	
	hc.logger.Warn("No suitable recovery strategy found",
		"plugin", hc.plugin.GetInfo().Name,
		"status", result.Status.String())
}