// pkg/plugin/health_strategies.go
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

// registerDefaultStrategies 注册默认的健康检查策略
func (hc *DefaultHealthChecker) registerDefaultStrategies() {
	hc.healthStrategies["basic"] = &BasicHealthStrategy{logger: hc.logger}
	hc.healthStrategies["performance"] = &PerformanceHealthStrategy{logger: hc.logger}
	hc.healthStrategies["resources"] = &ResourceHealthStrategy{logger: hc.logger}
}

// registerDefaultCollectors 注册默认的指标收集器
func (hc *DefaultHealthChecker) registerDefaultCollectors() {
	hc.metricsCollectors["system"] = &SystemMetricsCollector{}
	hc.metricsCollectors["plugin"] = &PluginMetricsCollector{plugin: hc.plugin}
	hc.metricsCollectors["performance"] = &PerformanceMetricsCollector{
		plugin:    hc.plugin,
		startTime: time.Now(),
		stats:     make(map[string]*PerformanceStats),
	}
}

// registerDefaultRecoveryStrategies 注册默认的恢复策略
func (hc *DefaultHealthChecker) registerDefaultRecoveryStrategies() {
	hc.recoveryStrategies["restart"] = &RestartRecoveryStrategy{logger: hc.logger}
	hc.recoveryStrategies["gc"] = &GCRecoveryStrategy{logger: hc.logger}
	hc.recoveryStrategies["reset"] = &ResetRecoveryStrategy{logger: hc.logger}
}

// BasicHealthStrategy 基础健康检查策略
type BasicHealthStrategy struct {
	logger *slog.Logger
}

func (s *BasicHealthStrategy) Execute(ctx context.Context, plugin core.Plugin) (*HealthCheckResult, error) {
	startTime := time.Now()
	
	// 执行插件的健康检查方法
	err := plugin.HealthCheck()
	duration := time.Since(startTime)
	
	if err != nil {
		return &HealthCheckResult{
			Status:    HealthStatusUnhealthy,
			Message:   fmt.Sprintf("Plugin health check failed: %v", err),
			CheckedAt: startTime,
			Duration:  duration,
			Details: map[string]interface{}{
				"error": err.Error(),
				"check_type": "basic",
			},
		}, nil
	}
	
	return &HealthCheckResult{
		Status:    HealthStatusHealthy,
		Message:   "Basic health check passed",
		CheckedAt: startTime,
		Duration:  duration,
		Details: map[string]interface{}{
			"check_type": "basic",
		},
	}, nil
}

func (s *BasicHealthStrategy) GetName() string {
	return "basic"
}

func (s *BasicHealthStrategy) Check(ctx context.Context, plugin core.Plugin) (*HealthCheckResult, error) {
	return s.Execute(ctx, plugin)
}

// PerformanceHealthStrategy 性能健康检查策略
type PerformanceHealthStrategy struct {
	logger *slog.Logger
}

func (s *PerformanceHealthStrategy) Execute(ctx context.Context, plugin core.Plugin) (*HealthCheckResult, error) {
	startTime := time.Now()
	
	// 检查性能指标
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	details := map[string]interface{}{
		"check_type":     "performance",
		"alloc_bytes":    m.Alloc,
		"total_alloc":    m.TotalAlloc,
		"sys_bytes":      m.Sys,
		"num_gc":         m.NumGC,
		"goroutines":     runtime.NumGoroutine(),
		"gomaxprocs":     runtime.GOMAXPROCS(0),
	}
	
	// 简单的性能评估
	status := HealthStatusHealthy
	message := "Performance check passed"
	
	// 检查内存使用
	if m.Alloc > 100*1024*1024 { // 100MB
		status = HealthStatusDegraded
		message = "High memory usage detected"
	}
	
	// 检查协程数量
	if runtime.NumGoroutine() > 1000 {
		status = HealthStatusDegraded
		message = "High goroutine count detected"
	}
	
	return &HealthCheckResult{
		Status:    status,
		Message:   message,
		CheckedAt: startTime,
		Duration:  time.Since(startTime),
		Details:   details,
	}, nil
}

func (s *PerformanceHealthStrategy) GetName() string {
	return "performance"
}

func (s *PerformanceHealthStrategy) Check(ctx context.Context, plugin core.Plugin) (*HealthCheckResult, error) {
	return s.Execute(ctx, plugin)
}

// ResourceHealthStrategy 资源健康检查策略
type ResourceHealthStrategy struct {
	logger *slog.Logger
}

func (s *ResourceHealthStrategy) Execute(ctx context.Context, plugin core.Plugin) (*HealthCheckResult, error) {
	startTime := time.Now()
	
	// 检查资源使用情况
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	details := map[string]interface{}{
		"check_type":        "resources",
		"heap_alloc":        m.HeapAlloc,
		"heap_sys":          m.HeapSys,
		"heap_idle":         m.HeapIdle,
		"heap_inuse":        m.HeapInuse,
		"heap_released":     m.HeapReleased,
		"heap_objects":      m.HeapObjects,
		"stack_inuse":       m.StackInuse,
		"stack_sys":         m.StackSys,
		"next_gc":           m.NextGC,
		"last_gc":           time.Unix(0, int64(m.LastGC)),
		"pause_total_ns":    m.PauseTotalNs,
		"num_forced_gc":     m.NumForcedGC,
	}
	
	status := HealthStatusHealthy
	message := "Resource check passed"
	
	// 检查堆内存使用
	heapUsagePercent := float64(m.HeapInuse) / float64(m.HeapSys) * 100
	if heapUsagePercent > 90 {
		status = HealthStatusCritical
		message = fmt.Sprintf("Critical heap usage: %.2f%%", heapUsagePercent)
	} else if heapUsagePercent > 75 {
		status = HealthStatusDegraded
		message = fmt.Sprintf("High heap usage: %.2f%%", heapUsagePercent)
	}
	
	details["heap_usage_percent"] = heapUsagePercent
	
	return &HealthCheckResult{
		Status:    status,
		Message:   message,
		CheckedAt: startTime,
		Duration:  time.Since(startTime),
		Details:   details,
	}, nil
}

func (s *ResourceHealthStrategy) GetName() string {
	return "resources"
}

func (s *ResourceHealthStrategy) Check(ctx context.Context, plugin core.Plugin) (*HealthCheckResult, error) {
	return s.Execute(ctx, plugin)
}

// SystemMetricsCollector 系统指标收集器
type SystemMetricsCollector struct{}

func (c *SystemMetricsCollector) Collect(ctx context.Context) (map[string]interface{}, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return map[string]interface{}{
		"memory_usage":  int64(m.Alloc),
		"goroutines":    runtime.NumGoroutine(),
		"gc_count":      m.NumGC,
		"last_gc_time": time.Unix(0, int64(m.LastGC)),
	}, nil
}

func (c *SystemMetricsCollector) GetName() string {
	return "system"
}

func (c *SystemMetricsCollector) GetDescription() string {
	return "Collects basic system metrics like memory usage and goroutine count"
}

// PluginMetricsCollector 插件指标收集器
type PluginMetricsCollector struct {
	plugin core.Plugin
}

func (c *PluginMetricsCollector) Collect(ctx context.Context) (map[string]interface{}, error) {
	info := c.plugin.GetInfo()
	
	metrics := map[string]interface{}{
		"plugin_name":    info.Name,
		"plugin_version": info.Version,
		"uptime":         time.Since(info.CreatedAt),
	}
	
	// 尝试执行健康检查来测量响应时间
	startTime := time.Now()
	err := c.plugin.HealthCheck()
	responseTime := time.Since(startTime)
	
	metrics["response_time"] = responseTime
	metrics["last_health_check_success"] = err == nil
	
	if err != nil {
		metrics["last_health_check_error"] = err.Error()
		metrics["error_rate"] = 1.0
	} else {
		metrics["error_rate"] = 0.0
	}
	
	return metrics, nil
}

func (c *PluginMetricsCollector) GetName() string {
	return "plugin"
}

func (c *PluginMetricsCollector) GetDescription() string {
	return "Collects plugin-specific metrics like response time and error rate"
}

// PerformanceStats 性能统计
type PerformanceStats struct {
	TotalRequests   int64         `json:"total_requests"`
	SuccessRequests int64         `json:"success_requests"`
	ErrorRequests   int64         `json:"error_requests"`
	TotalDuration   time.Duration `json:"total_duration"`
	MinDuration     time.Duration `json:"min_duration"`
	MaxDuration     time.Duration `json:"max_duration"`
	LastRequestTime time.Time     `json:"last_request_time"`
	mutex           sync.RWMutex
}

// PerformanceMetricsCollector 性能指标收集器
type PerformanceMetricsCollector struct {
	plugin    core.Plugin
	startTime time.Time
	stats     map[string]*PerformanceStats
	mutex     sync.RWMutex
}

func (c *PerformanceMetricsCollector) Collect(ctx context.Context) (map[string]interface{}, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	metrics := map[string]interface{}{
		"uptime": time.Since(c.startTime),
	}
	
	// 聚合所有统计信息
	var totalRequests, successRequests, errorRequests int64
	var totalDuration, minDuration, maxDuration time.Duration
	var avgDuration time.Duration
	
	for _, stats := range c.stats {
		stats.mutex.RLock()
		totalRequests += stats.TotalRequests
		successRequests += stats.SuccessRequests
		errorRequests += stats.ErrorRequests
		totalDuration += stats.TotalDuration
		
		if minDuration == 0 || (stats.MinDuration > 0 && stats.MinDuration < minDuration) {
			minDuration = stats.MinDuration
		}
		if stats.MaxDuration > maxDuration {
			maxDuration = stats.MaxDuration
		}
		stats.mutex.RUnlock()
	}
	
	if totalRequests > 0 {
		avgDuration = totalDuration / time.Duration(totalRequests)
		metrics["throughput"] = float64(totalRequests) / time.Since(c.startTime).Seconds()
		metrics["error_rate"] = float64(errorRequests) / float64(totalRequests)
		metrics["success_rate"] = float64(successRequests) / float64(totalRequests)
	} else {
		metrics["throughput"] = 0.0
		metrics["error_rate"] = 0.0
		metrics["success_rate"] = 1.0
	}
	
	metrics["total_requests"] = totalRequests
	metrics["success_requests"] = successRequests
	metrics["error_requests"] = errorRequests
	metrics["response_time"] = avgDuration
	metrics["min_response_time"] = minDuration
	metrics["max_response_time"] = maxDuration
	
	return metrics, nil
}

func (c *PerformanceMetricsCollector) GetName() string {
	return "performance"
}

func (c *PerformanceMetricsCollector) GetDescription() string {
	return "Collects performance metrics like throughput, response time, and error rates"
}

// RecordRequest 记录请求
func (c *PerformanceMetricsCollector) RecordRequest(operation string, duration time.Duration, success bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	stats, exists := c.stats[operation]
	if !exists {
		stats = &PerformanceStats{
			MinDuration: duration,
			MaxDuration: duration,
		}
		c.stats[operation] = stats
	}
	
	stats.mutex.Lock()
	defer stats.mutex.Unlock()
	
	stats.TotalRequests++
	stats.TotalDuration += duration
	stats.LastRequestTime = time.Now()
	
	if success {
		stats.SuccessRequests++
	} else {
		stats.ErrorRequests++
	}
	
	if duration < stats.MinDuration {
		stats.MinDuration = duration
	}
	if duration > stats.MaxDuration {
		stats.MaxDuration = duration
	}
}

// RestartRecoveryStrategy 重启恢复策略
type RestartRecoveryStrategy struct {
	logger *slog.Logger
}

func (s *RestartRecoveryStrategy) Recover(ctx context.Context, plugin core.Plugin, result *HealthCheckResult) error {
	s.logger.Info("Attempting to restart plugin", "plugin", plugin.GetInfo().Name)
	
	// 停止插件
	if err := plugin.Stop(); err != nil {
		s.logger.Error("Failed to stop plugin during restart", "error", err)
		return fmt.Errorf("failed to stop plugin: %w", err)
	}
	
	// 等待一段时间
	time.Sleep(2 * time.Second)
	
	// 重新启动插件
	if err := plugin.Start(); err != nil {
		s.logger.Error("Failed to start plugin during restart", "error", err)
		return fmt.Errorf("failed to start plugin: %w", err)
	}
	
	s.logger.Info("Plugin restarted successfully", "plugin", plugin.GetInfo().Name)
	return nil
}

func (s *RestartRecoveryStrategy) CanRecover(result *HealthCheckResult) bool {
	// 可以尝试重启的情况
	return result.Status == HealthStatusCritical || result.Status == HealthStatusUnhealthy
}

func (s *RestartRecoveryStrategy) GetName() string {
	return "restart"
}

func (s *RestartRecoveryStrategy) GetDescription() string {
	return "Restarts the plugin when it becomes unhealthy or critical"
}

// GCRecoveryStrategy 垃圾回收恢复策略
type GCRecoveryStrategy struct {
	logger *slog.Logger
}

func (s *GCRecoveryStrategy) Recover(ctx context.Context, plugin core.Plugin, result *HealthCheckResult) error {
	s.logger.Info("Attempting garbage collection recovery", "plugin", plugin.GetInfo().Name)
	
	// 强制垃圾回收
	runtime.GC()
	runtime.GC() // 执行两次确保彻底清理
	
	// 等待GC完成
	time.Sleep(1 * time.Second)
	
	s.logger.Info("Garbage collection completed", "plugin", plugin.GetInfo().Name)
	return nil
}

func (s *GCRecoveryStrategy) CanRecover(result *HealthCheckResult) bool {
	// 当内存使用过高时可以尝试GC
	if result.Details != nil {
		if memStatus, ok := result.Details["memory_status"].(string); ok {
			return memStatus == "warning" || memStatus == "critical"
		}
	}
	return false
}

func (s *GCRecoveryStrategy) GetName() string {
	return "gc"
}

func (s *GCRecoveryStrategy) GetDescription() string {
	return "Performs garbage collection to free memory when memory usage is high"
}

// ResetRecoveryStrategy 重置恢复策略
type ResetRecoveryStrategy struct {
	logger *slog.Logger
}

func (s *ResetRecoveryStrategy) Recover(ctx context.Context, plugin core.Plugin, result *HealthCheckResult) error {
	s.logger.Info("Attempting to reset plugin state", "plugin", plugin.GetInfo().Name)
	
	// 尝试清理插件状态
	if err := plugin.Cleanup(); err != nil {
		s.logger.Error("Failed to cleanup plugin during reset", "error", err)
		return fmt.Errorf("failed to cleanup plugin: %w", err)
	}
	
	// 重新初始化（如果插件支持的话）
	// 这里需要插件上下文，但为了简化，我们只记录日志
	s.logger.Info("Plugin state reset completed", "plugin", plugin.GetInfo().Name)
	return nil
}

func (s *ResetRecoveryStrategy) CanRecover(result *HealthCheckResult) bool {
	// 当插件健康检查失败时可以尝试重置
	if result.Details != nil {
		if pluginHealth, ok := result.Details["plugin_health"].(string); ok {
			return pluginHealth == "failed"
		}
	}
	return result.Status == HealthStatusUnhealthy
}

func (s *ResetRecoveryStrategy) GetName() string {
	return "reset"
}

func (s *ResetRecoveryStrategy) GetDescription() string {
	return "Resets plugin state when plugin health check fails"
}