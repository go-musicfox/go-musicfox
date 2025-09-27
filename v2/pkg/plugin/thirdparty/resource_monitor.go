// Package thirdparty 实现资源监控功能
package thirdparty

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// ResourceMonitor 资源监控器
type ResourceMonitor struct {
	limits      *ResourceLimits
	usage       *ResourceUsage
	violations  []ResourceViolation
	mu          sync.RWMutex
	running     int32
	ctx         context.Context
	cancel      context.CancelFunc
	updateInterval time.Duration
	alertThreshold float64 // 告警阈值（0-1）
	callbacks   []ResourceCallback
}

// ResourceViolation 资源违规记录
type ResourceViolation struct {
	Type        ResourceViolationType `json:"type"`        // 违规类型
	Description string                `json:"description"` // 描述
	Timestamp   time.Time             `json:"timestamp"`   // 时间戳
	CurrentValue interface{}          `json:"current_value"` // 当前值
	LimitValue   interface{}          `json:"limit_value"`   // 限制值
	Severity     Severity             `json:"severity"`      // 严重程度
	Action       ViolationAction      `json:"action"`       // 采取的行动
}

// ResourceViolationType 资源违规类型枚举
type ResourceViolationType int

const (
	ResourceViolationMemory ResourceViolationType = iota // 内存违规
	ResourceViolationCPU                                 // CPU违规
	ResourceViolationDiskIO                              // 磁盘IO违规
	ResourceViolationNetworkIO                           // 网络IO违规
	ResourceViolationGoroutines                          // 协程数违规
	ResourceViolationFileHandles                         // 文件句柄违规
	ResourceViolationTimeout                             // 超时违规
)

// String 返回资源违规类型的字符串表示
func (r ResourceViolationType) String() string {
	switch r {
	case ResourceViolationMemory:
		return "memory"
	case ResourceViolationCPU:
		return "cpu"
	case ResourceViolationDiskIO:
		return "disk_io"
	case ResourceViolationNetworkIO:
		return "network_io"
	case ResourceViolationGoroutines:
		return "goroutines"
	case ResourceViolationFileHandles:
		return "file_handles"
	case ResourceViolationTimeout:
		return "timeout"
	default:
		return "unknown"
	}
}

// ViolationAction 违规行动枚举
type ViolationAction int

const (
	ViolationActionNone ViolationAction = iota // 无行动
	ViolationActionWarn                         // 警告
	ViolationActionThrottle                     // 限流
	ViolationActionTerminate                    // 终止
)

// String 返回违规行动的字符串表示
func (v ViolationAction) String() string {
	switch v {
	case ViolationActionNone:
		return "none"
	case ViolationActionWarn:
		return "warn"
	case ViolationActionThrottle:
		return "throttle"
	case ViolationActionTerminate:
		return "terminate"
	default:
		return "unknown"
	}
}

// ResourceCallback 资源回调函数类型
type ResourceCallback func(violation ResourceViolation)

// NewResourceMonitor 创建新的资源监控器
func NewResourceMonitor(limits *ResourceLimits) (*ResourceMonitor, error) {
	if limits == nil {
		return nil, fmt.Errorf("resource limits cannot be nil")
	}

	rm := &ResourceMonitor{
		limits:         limits,
		usage:          &ResourceUsage{LastUpdated: time.Now()},
		violations:     make([]ResourceViolation, 0),
		updateInterval: 1 * time.Second,
		alertThreshold: 0.8, // 80%阈值告警
		callbacks:      make([]ResourceCallback, 0),
	}

	return rm, nil
}

// Start 启动资源监控
func (rm *ResourceMonitor) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&rm.running, 0, 1) {
		return fmt.Errorf("resource monitor already running")
	}

	rm.ctx, rm.cancel = context.WithCancel(ctx)

	// 启动监控goroutine
	go rm.monitorLoop()

	return nil
}

// Stop 停止资源监控
func (rm *ResourceMonitor) Stop() {
	if !atomic.CompareAndSwapInt32(&rm.running, 1, 0) {
		return
	}

	if rm.cancel != nil {
		rm.cancel()
	}
}

// GetUsage 获取当前资源使用情况
func (rm *ResourceMonitor) GetUsage() *ResourceUsage {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// 更新资源使用情况
	rm.updateUsage()

	// 返回副本
	usage := *rm.usage
	return &usage
}

// UpdateLimits 更新资源限制
func (rm *ResourceMonitor) UpdateLimits(limits *ResourceLimits) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if limits == nil {
		return fmt.Errorf("resource limits cannot be nil")
	}

	rm.limits = limits
	return nil
}

// GetViolations 获取资源违规记录
func (rm *ResourceMonitor) GetViolations() []ResourceViolation {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return append([]ResourceViolation{}, rm.violations...)
}

// ClearViolations 清除违规记录
func (rm *ResourceMonitor) ClearViolations() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.violations = make([]ResourceViolation, 0)
}

// AddCallback 添加资源回调
func (rm *ResourceMonitor) AddCallback(callback ResourceCallback) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.callbacks = append(rm.callbacks, callback)
}

// SetUpdateInterval 设置更新间隔
func (rm *ResourceMonitor) SetUpdateInterval(interval time.Duration) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.updateInterval = interval
}

// SetAlertThreshold 设置告警阈值
func (rm *ResourceMonitor) SetAlertThreshold(threshold float64) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if threshold >= 0 && threshold <= 1 {
		rm.alertThreshold = threshold
	}
}

// monitorLoop 监控循环
func (rm *ResourceMonitor) monitorLoop() {
	ticker := time.NewTicker(rm.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.checkResources()
		}
	}
}

// checkResources 检查资源使用情况
func (rm *ResourceMonitor) checkResources() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// 更新资源使用情况
	rm.updateUsage()

	// 检查各项资源是否超限
	rm.checkMemoryUsage()
	rm.checkCPUUsage()
	rm.checkGoroutineCount()
	rm.checkFileHandles()
}

// updateUsage 更新资源使用情况
func (rm *ResourceMonitor) updateUsage() {
	now := time.Now()

	// 更新内存使用情况
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	rm.usage.MemoryUsage = int64(memStats.Alloc)

	// 更新协程数量
	rm.usage.GoroutineCount = runtime.NumGoroutine()

	// 更新运行时间
	if rm.usage.LastUpdated.IsZero() {
		rm.usage.Uptime = 0
	} else {
		rm.usage.Uptime = now.Sub(rm.usage.LastUpdated)
	}

	// 更新CPU使用率（简化实现）
	rm.usage.CPUUsage = rm.calculateCPUUsage()

	// 更新文件句柄数（简化实现）
	rm.usage.OpenFileCount = rm.getOpenFileCount()

	rm.usage.LastUpdated = now
}

// checkMemoryUsage 检查内存使用情况
func (rm *ResourceMonitor) checkMemoryUsage() {
	if rm.limits.MaxMemory <= 0 {
		return
	}

	usageRatio := float64(rm.usage.MemoryUsage) / float64(rm.limits.MaxMemory)

	if usageRatio >= 1.0 {
		// 超过限制
		violation := ResourceViolation{
			Type:         ResourceViolationMemory,
			Description:  "Memory usage exceeded limit",
			Timestamp:    time.Now(),
			CurrentValue: rm.usage.MemoryUsage,
			LimitValue:   rm.limits.MaxMemory,
			Severity:     SeverityCritical,
			Action:       ViolationActionTerminate,
		}
		rm.recordViolation(violation)
	} else if usageRatio >= rm.alertThreshold {
		// 接近限制
		violation := ResourceViolation{
			Type:         ResourceViolationMemory,
			Description:  "Memory usage approaching limit",
			Timestamp:    time.Now(),
			CurrentValue: rm.usage.MemoryUsage,
			LimitValue:   rm.limits.MaxMemory,
			Severity:     SeverityHigh,
			Action:       ViolationActionWarn,
		}
		rm.recordViolation(violation)
	}
}

// checkCPUUsage 检查CPU使用情况
func (rm *ResourceMonitor) checkCPUUsage() {
	if rm.limits.MaxCPU <= 0 {
		return
	}

	usageRatio := rm.usage.CPUUsage / rm.limits.MaxCPU

	if usageRatio >= 1.0 {
		// 超过限制
		violation := ResourceViolation{
			Type:         ResourceViolationCPU,
			Description:  "CPU usage exceeded limit",
			Timestamp:    time.Now(),
			CurrentValue: rm.usage.CPUUsage,
			LimitValue:   rm.limits.MaxCPU,
			Severity:     SeverityHigh,
			Action:       ViolationActionThrottle,
		}
		rm.recordViolation(violation)
	} else if usageRatio >= rm.alertThreshold {
		// 接近限制
		violation := ResourceViolation{
			Type:         ResourceViolationCPU,
			Description:  "CPU usage approaching limit",
			Timestamp:    time.Now(),
			CurrentValue: rm.usage.CPUUsage,
			LimitValue:   rm.limits.MaxCPU,
			Severity:     SeverityMedium,
			Action:       ViolationActionWarn,
		}
		rm.recordViolation(violation)
	}
}

// checkGoroutineCount 检查协程数量
func (rm *ResourceMonitor) checkGoroutineCount() {
	if rm.limits.MaxGoroutines <= 0 {
		return
	}

	usageRatio := float64(rm.usage.GoroutineCount) / float64(rm.limits.MaxGoroutines)

	if usageRatio >= 1.0 {
		// 超过限制
		violation := ResourceViolation{
			Type:         ResourceViolationGoroutines,
			Description:  "Goroutine count exceeded limit",
			Timestamp:    time.Now(),
			CurrentValue: rm.usage.GoroutineCount,
			LimitValue:   rm.limits.MaxGoroutines,
			Severity:     SeverityHigh,
			Action:       ViolationActionThrottle,
		}
		rm.recordViolation(violation)
	} else if usageRatio >= rm.alertThreshold {
		// 接近限制
		violation := ResourceViolation{
			Type:         ResourceViolationGoroutines,
			Description:  "Goroutine count approaching limit",
			Timestamp:    time.Now(),
			CurrentValue: rm.usage.GoroutineCount,
			LimitValue:   rm.limits.MaxGoroutines,
			Severity:     SeverityMedium,
			Action:       ViolationActionWarn,
		}
		rm.recordViolation(violation)
	}
}

// checkFileHandles 检查文件句柄数量
func (rm *ResourceMonitor) checkFileHandles() {
	if rm.limits.MaxOpenFiles <= 0 {
		return
	}

	usageRatio := float64(rm.usage.OpenFileCount) / float64(rm.limits.MaxOpenFiles)

	if usageRatio >= 1.0 {
		// 超过限制
		violation := ResourceViolation{
			Type:         ResourceViolationFileHandles,
			Description:  "Open file count exceeded limit",
			Timestamp:    time.Now(),
			CurrentValue: rm.usage.OpenFileCount,
			LimitValue:   rm.limits.MaxOpenFiles,
			Severity:     SeverityHigh,
			Action:       ViolationActionThrottle,
		}
		rm.recordViolation(violation)
	} else if usageRatio >= rm.alertThreshold {
		// 接近限制
		violation := ResourceViolation{
			Type:         ResourceViolationFileHandles,
			Description:  "Open file count approaching limit",
			Timestamp:    time.Now(),
			CurrentValue: rm.usage.OpenFileCount,
			LimitValue:   rm.limits.MaxOpenFiles,
			Severity:     SeverityMedium,
			Action:       ViolationActionWarn,
		}
		rm.recordViolation(violation)
	}
}

// recordViolation 记录资源违规
func (rm *ResourceMonitor) recordViolation(violation ResourceViolation) {
	// 限制违规记录数量，避免内存泄漏
	if len(rm.violations) >= 1000 {
		// 移除最旧的记录
		rm.violations = rm.violations[1:]
	}
	rm.violations = append(rm.violations, violation)

	// 复制回调函数列表，避免并发访问问题
	callbacks := make([]ResourceCallback, len(rm.callbacks))
	copy(callbacks, rm.callbacks)

	// 调用回调函数
	for _, callback := range callbacks {
		go func(cb ResourceCallback, v ResourceViolation) {
			defer func() {
				if r := recover(); r != nil {
					// 忽略回调函数中的panic
				}
			}()
			cb(v)
		}(callback, violation)
	}
}

// calculateCPUUsage 计算CPU使用率（简化实现）
func (rm *ResourceMonitor) calculateCPUUsage() float64 {
	// 这里应该实现真正的CPU使用率计算
	// 简化实现，返回一个模拟值
	return 0.1 // 10%
}

// getOpenFileCount 获取打开的文件数量（简化实现）
func (rm *ResourceMonitor) getOpenFileCount() int {
	// 这里应该实现真正的文件句柄计数
	// 简化实现，返回一个模拟值
	return 10
}

// GetStats 获取监控统计信息
func (rm *ResourceMonitor) GetStats() map[string]interface{} {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["running"] = atomic.LoadInt32(&rm.running) == 1
	stats["update_interval"] = rm.updateInterval.String()
	stats["alert_threshold"] = rm.alertThreshold
	stats["violation_count"] = len(rm.violations)
	stats["callback_count"] = len(rm.callbacks)

	// 当前使用情况
	stats["current_usage"] = map[string]interface{}{
		"memory_usage":    rm.usage.MemoryUsage,
		"cpu_usage":       rm.usage.CPUUsage,
		"goroutine_count": rm.usage.GoroutineCount,
		"open_file_count": rm.usage.OpenFileCount,
		"uptime":          rm.usage.Uptime.String(),
	}

	// 资源限制
	stats["limits"] = map[string]interface{}{
		"max_memory":      rm.limits.MaxMemory,
		"max_cpu":         rm.limits.MaxCPU,
		"max_goroutines":  rm.limits.MaxGoroutines,
		"max_open_files":  rm.limits.MaxOpenFiles,
		"timeout":         rm.limits.Timeout.String(),
	}

	// 违规统计
	violationStats := make(map[string]int)
	for _, violation := range rm.violations {
		violationStats[violation.Type.String()]++
	}
	stats["violation_stats"] = violationStats

	return stats
}

// IsRunning 检查监控器是否正在运行
func (rm *ResourceMonitor) IsRunning() bool {
	return atomic.LoadInt32(&rm.running) == 1
}