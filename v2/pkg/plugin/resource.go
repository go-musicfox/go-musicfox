package plugin

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"log/slog"
)

// ResourceType 资源类型
type ResourceType string

const (
	ResourceTypeMemory    ResourceType = "memory"
	ResourceTypeCPU       ResourceType = "cpu"
	ResourceTypeDiskIO    ResourceType = "disk_io"
	ResourceTypeNetworkIO ResourceType = "network_io"
	ResourceTypeGoroutine ResourceType = "goroutine"
	ResourceTypeFileDesc  ResourceType = "file_descriptor"
)

// ResourceUnit 资源单位
type ResourceUnit string

const (
	ResourceUnitBytes       ResourceUnit = "bytes"
	ResourceUnitKB          ResourceUnit = "kb"
	ResourceUnitMB          ResourceUnit = "mb"
	ResourceUnitGB          ResourceUnit = "gb"
	ResourceUnitPercent     ResourceUnit = "percent"
	ResourceUnitCount       ResourceUnit = "count"
	ResourceUnitBytesPerSec ResourceUnit = "bytes_per_sec"
	ResourceUnitOpsPerSec   ResourceUnit = "ops_per_sec"
)

// ResourceLimit 单个资源限制
type ResourceLimit struct {
	Type        ResourceType `json:"type" yaml:"type"`
	SoftLimit   int64        `json:"soft_limit" yaml:"soft_limit"`     // 软限制
	HardLimit   int64        `json:"hard_limit" yaml:"hard_limit"`     // 硬限制
	Unit        ResourceUnit `json:"unit" yaml:"unit"`
	Enabled     bool         `json:"enabled" yaml:"enabled"`
	Description string       `json:"description,omitempty" yaml:"description,omitempty"`
}

// Validate 验证资源限制配置
func (rl *ResourceLimit) Validate() error {
	if rl.Type == "" {
		return fmt.Errorf("resource type cannot be empty")
	}

	if rl.Enabled {
		if rl.SoftLimit <= 0 {
			return fmt.Errorf("soft limit must be positive when enabled")
		}
		if rl.HardLimit <= 0 {
			return fmt.Errorf("hard limit must be positive when enabled")
		}
		if rl.SoftLimit > rl.HardLimit {
			return fmt.Errorf("soft limit cannot exceed hard limit")
		}
	}

	return nil
}

// IsExceeded 检查是否超过限制
func (rl *ResourceLimit) IsExceeded(current int64) (bool, bool) {
	if !rl.Enabled {
		return false, false
	}

	softExceeded := current > rl.SoftLimit
	hardExceeded := current > rl.HardLimit

	return softExceeded, hardExceeded
}

// GetUtilization 获取资源利用率（百分比）
func (rl *ResourceLimit) GetUtilization(current int64) float64 {
	if !rl.Enabled || rl.HardLimit <= 0 {
		return 0
	}

	utilization := float64(current) / float64(rl.HardLimit) * 100
	if utilization > 100 {
		return 100
	}
	return utilization
}

// DetailedResourceLimits 详细资源限制集合
type DetailedResourceLimits struct {
	Memory       *ResourceLimit            `json:"memory,omitempty" yaml:"memory,omitempty"`
	CPU          *ResourceLimit            `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	DiskIO       *ResourceLimit            `json:"disk_io,omitempty" yaml:"disk_io,omitempty"`
	NetworkIO    *ResourceLimit            `json:"network_io,omitempty" yaml:"network_io,omitempty"`
	Goroutines   *ResourceLimit            `json:"goroutines,omitempty" yaml:"goroutines,omitempty"`
	FileDesc     *ResourceLimit            `json:"file_descriptors,omitempty" yaml:"file_descriptors,omitempty"`
	CustomLimits map[string]*ResourceLimit `json:"custom_limits,omitempty" yaml:"custom_limits,omitempty"`
	Enabled      bool                      `json:"enabled" yaml:"enabled"`
}

// NewDetailedResourceLimits 创建详细资源限制
func NewDetailedResourceLimits() *DetailedResourceLimits {
	return &DetailedResourceLimits{
		Memory: &ResourceLimit{
			Type:        ResourceTypeMemory,
			SoftLimit:   100 * 1024 * 1024, // 100MB
			HardLimit:   200 * 1024 * 1024, // 200MB
			Unit:        ResourceUnitBytes,
			Enabled:     true,
			Description: "Memory usage limit",
		},
		CPU: &ResourceLimit{
			Type:        ResourceTypeCPU,
			SoftLimit:   50, // 50%
			HardLimit:   80, // 80%
			Unit:        ResourceUnitPercent,
			Enabled:     true,
			Description: "CPU usage limit",
		},
		Goroutines: &ResourceLimit{
			Type:        ResourceTypeGoroutine,
			SoftLimit:   100,
			HardLimit:   200,
			Unit:        ResourceUnitCount,
			Enabled:     true,
			Description: "Goroutine count limit",
		},
		FileDesc: &ResourceLimit{
			Type:        ResourceTypeFileDesc,
			SoftLimit:   50,
			HardLimit:   100,
			Unit:        ResourceUnitCount,
			Enabled:     true,
			Description: "File descriptor limit",
		},
		CustomLimits: make(map[string]*ResourceLimit),
		Enabled:      true,
	}
}

// Validate 验证资源限制配置
func (rl *DetailedResourceLimits) Validate() error {
	if !rl.Enabled {
		return nil
	}

	// 验证各个资源限制
	limits := []*ResourceLimit{rl.Memory, rl.CPU, rl.DiskIO, rl.NetworkIO, rl.Goroutines, rl.FileDesc}
	for _, limit := range limits {
		if limit != nil {
			if err := limit.Validate(); err != nil {
				return fmt.Errorf("invalid %s limit: %w", limit.Type, err)
			}
		}
	}

	// 验证自定义限制
	for name, limit := range rl.CustomLimits {
		if limit != nil {
			if err := limit.Validate(); err != nil {
				return fmt.Errorf("invalid custom limit %s: %w", name, err)
			}
		}
	}

	return nil
}

// GetLimit 获取指定类型的资源限制
func (rl *DetailedResourceLimits) GetLimit(resourceType ResourceType) *ResourceLimit {
	if !rl.Enabled {
		return nil
	}

	switch resourceType {
	case ResourceTypeMemory:
		return rl.Memory
	case ResourceTypeCPU:
		return rl.CPU
	case ResourceTypeDiskIO:
		return rl.DiskIO
	case ResourceTypeNetworkIO:
		return rl.NetworkIO
	case ResourceTypeGoroutine:
		return rl.Goroutines
	case ResourceTypeFileDesc:
		return rl.FileDesc
	default:
		return nil
	}
}

// GetCustomLimit 获取自定义资源限制
func (rl *DetailedResourceLimits) GetCustomLimit(name string) *ResourceLimit {
	if !rl.Enabled {
		return nil
	}
	return rl.CustomLimits[name]
}

// SetCustomLimit 设置自定义资源限制
func (rl *DetailedResourceLimits) SetCustomLimit(name string, limit *ResourceLimit) error {
	if limit != nil {
		if err := limit.Validate(); err != nil {
			return fmt.Errorf("invalid custom limit: %w", err)
		}
	}

	if rl.CustomLimits == nil {
		rl.CustomLimits = make(map[string]*ResourceLimit)
	}

	rl.CustomLimits[name] = limit
	return nil
}

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	PluginID    string                 `json:"plugin_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Memory      int64                  `json:"memory"`      // 字节
	CPU         float64                `json:"cpu"`         // 百分比
	DiskIO      int64                  `json:"disk_io"`     // 字节/秒
	NetworkIO   int64                  `json:"network_io"`  // 字节/秒
	Goroutines  int                    `json:"goroutines"`  // 数量
	FileDesc    int                    `json:"file_desc"`   // 数量
	CustomUsage map[string]interface{} `json:"custom_usage,omitempty"`
}

// ResourceAlert 资源告警
type ResourceAlert struct {
	PluginID     string       `json:"plugin_id"`
	ResourceType ResourceType `json:"resource_type"`
	AlertLevel   AlertLevel   `json:"alert_level"`
	CurrentUsage int64        `json:"current_usage"`
	Limit        int64        `json:"limit"`
	Utilization  float64      `json:"utilization"`
	Timestamp    time.Time    `json:"timestamp"`
	Message      string       `json:"message"`
}

// AlertLevel 告警级别
type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelCritical AlertLevel = "critical"
)

// ResourceMonitor 资源监控器
type ResourceMonitor struct {
	logger       *slog.Logger
	pluginID     string
	limits       *DetailedResourceLimits
	usage        *ResourceUsage
	alerts       chan *ResourceAlert
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.RWMutex
	isRunning    bool
	interval     time.Duration
	alertHistory []ResourceAlert
	maxHistory   int

	// 统计计数器
	memoryUsage    int64
	cpuUsage       int64
	diskIOBytes    int64
	networkIOBytes int64
	goroutineCount int32
	fileDescCount  int32
}

// ResourceMonitorOptions 资源监控器选项
type ResourceMonitorOptions struct {
	Interval   time.Duration // 监控间隔
	MaxHistory int           // 最大告警历史记录数
}

// DefaultResourceMonitorOptions 默认资源监控器选项
func DefaultResourceMonitorOptions() *ResourceMonitorOptions {
	return &ResourceMonitorOptions{
		Interval:   5 * time.Second,
		MaxHistory: 100,
	}
}

// NewResourceMonitor 创建资源监控器
func NewResourceMonitor(logger *slog.Logger, pluginID string, limits *DetailedResourceLimits) *ResourceMonitor {
	return NewResourceMonitorWithOptions(logger, pluginID, limits, DefaultResourceMonitorOptions())
}

// NewResourceMonitorWithOptions 使用选项创建资源监控器
func NewResourceMonitorWithOptions(logger *slog.Logger, pluginID string, limits *DetailedResourceLimits, options *ResourceMonitorOptions) *ResourceMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	return &ResourceMonitor{
		logger:       logger,
		pluginID:     pluginID,
		limits:       limits,
		alerts:       make(chan *ResourceAlert, 100),
		ctx:          ctx,
		cancel:       cancel,
		interval:     options.Interval,
		maxHistory:   options.MaxHistory,
		alertHistory: make([]ResourceAlert, 0, options.MaxHistory),
		usage: &ResourceUsage{
			PluginID:    pluginID,
			CustomUsage: make(map[string]interface{}),
		},
	}
}

// Start 启动资源监控
func (rm *ResourceMonitor) Start(ctx context.Context) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.isRunning {
		return fmt.Errorf("resource monitor is already running")
	}

	if rm.limits == nil {
		return fmt.Errorf("resource limits not configured")
	}

	rm.isRunning = true

	// 启动监控协程
	go rm.monitorLoop()

	rm.logger.Info("Resource monitor started", "plugin_id", rm.pluginID, "interval", rm.interval)
	return nil
}

// Stop 停止资源监控
func (rm *ResourceMonitor) Stop() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if !rm.isRunning {
		return
	}

	rm.isRunning = false
	rm.cancel()
	close(rm.alerts)

	rm.logger.Info("Resource monitor stopped", "plugin_id", rm.pluginID)
}

// IsRunning 检查监控器是否运行中
func (rm *ResourceMonitor) IsRunning() bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.isRunning
}

// GetUsage 获取当前资源使用情况
func (rm *ResourceMonitor) GetUsage() *ResourceUsage {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// 创建副本以避免并发访问问题
	usage := *rm.usage
	usage.CustomUsage = make(map[string]interface{})
	for k, v := range rm.usage.CustomUsage {
		usage.CustomUsage[k] = v
	}

	return &usage
}

// GetAlerts 获取告警通道
func (rm *ResourceMonitor) GetAlerts() <-chan *ResourceAlert {
	return rm.alerts
}

// GetAlertHistory 获取告警历史
func (rm *ResourceMonitor) GetAlertHistory() []ResourceAlert {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// 创建副本
	history := make([]ResourceAlert, len(rm.alertHistory))
	copy(history, rm.alertHistory)
	return history
}

// UpdateLimits 更新资源限制
func (rm *ResourceMonitor) UpdateLimits(limits *DetailedResourceLimits) error {
	if limits != nil {
		if err := limits.Validate(); err != nil {
			return fmt.Errorf("invalid resource limits: %w", err)
		}
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.limits = limits
	rm.logger.Info("Resource limits updated", "plugin_id", rm.pluginID)
	return nil
}

// RecordMemoryUsage 记录内存使用
func (rm *ResourceMonitor) RecordMemoryUsage(bytes int64) {
	atomic.StoreInt64(&rm.memoryUsage, bytes)
}

// RecordCPUUsage 记录CPU使用
func (rm *ResourceMonitor) RecordCPUUsage(percent float64) {
	atomic.StoreInt64(&rm.cpuUsage, int64(percent*100)) // 存储为百分比*100
}

// RecordDiskIO 记录磁盘IO
func (rm *ResourceMonitor) RecordDiskIO(bytes int64) {
	atomic.AddInt64(&rm.diskIOBytes, bytes)
}

// RecordNetworkIO 记录网络IO
func (rm *ResourceMonitor) RecordNetworkIO(bytes int64) {
	atomic.AddInt64(&rm.networkIOBytes, bytes)
}

// RecordGoroutineCount 记录协程数量
func (rm *ResourceMonitor) RecordGoroutineCount(count int) {
	atomic.StoreInt32(&rm.goroutineCount, int32(count))
}

// RecordFileDescCount 记录文件描述符数量
func (rm *ResourceMonitor) RecordFileDescCount(count int) {
	atomic.StoreInt32(&rm.fileDescCount, int32(count))
}

// RecordCustomUsage 记录自定义资源使用
func (rm *ResourceMonitor) RecordCustomUsage(name string, value interface{}) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.usage.CustomUsage[name] = value
}

// monitorLoop 监控循环
func (rm *ResourceMonitor) monitorLoop() {
	ticker := time.NewTicker(rm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return

		case <-ticker.C:
			rm.collectAndCheckUsage()
		}
	}
}

// collectAndCheckUsage 收集并检查资源使用情况
func (rm *ResourceMonitor) collectAndCheckUsage() {
	// 收集系统资源使用情况
	rm.collectSystemUsage()

	// 更新使用情况记录
	rm.updateUsageRecord()

	// 检查资源限制
	rm.checkResourceLimits()
}

// collectSystemUsage 收集系统资源使用情况
func (rm *ResourceMonitor) collectSystemUsage() {
	// 收集内存使用情况
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	rm.RecordMemoryUsage(int64(m.Alloc))

	// 收集协程数量
	rm.RecordGoroutineCount(runtime.NumGoroutine())

	// 注意：CPU使用率和文件描述符数量需要更复杂的实现
	// 这里提供基础框架，实际实现可能需要使用系统调用或第三方库
}

// updateUsageRecord 更新使用情况记录
func (rm *ResourceMonitor) updateUsageRecord() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.usage.Timestamp = time.Now()
	rm.usage.Memory = atomic.LoadInt64(&rm.memoryUsage)
	rm.usage.CPU = float64(atomic.LoadInt64(&rm.cpuUsage)) / 100
	rm.usage.DiskIO = atomic.SwapInt64(&rm.diskIOBytes, 0) // 重置计数器
	rm.usage.NetworkIO = atomic.SwapInt64(&rm.networkIOBytes, 0)
	rm.usage.Goroutines = int(atomic.LoadInt32(&rm.goroutineCount))
	rm.usage.FileDesc = int(atomic.LoadInt32(&rm.fileDescCount))
}

// checkResourceLimits 检查资源限制
func (rm *ResourceMonitor) checkResourceLimits() {
	if rm.limits == nil || !rm.limits.Enabled {
		return
	}

	usage := rm.GetUsage()

	// 检查各种资源限制
	rm.checkLimit(ResourceTypeMemory, usage.Memory, rm.limits.Memory)
	rm.checkLimit(ResourceTypeCPU, int64(usage.CPU), rm.limits.CPU)
	rm.checkLimit(ResourceTypeDiskIO, usage.DiskIO, rm.limits.DiskIO)
	rm.checkLimit(ResourceTypeNetworkIO, usage.NetworkIO, rm.limits.NetworkIO)
	rm.checkLimit(ResourceTypeGoroutine, int64(usage.Goroutines), rm.limits.Goroutines)
	rm.checkLimit(ResourceTypeFileDesc, int64(usage.FileDesc), rm.limits.FileDesc)

	// 检查自定义限制
	for name, limit := range rm.limits.CustomLimits {
		if customValue, exists := usage.CustomUsage[name]; exists {
			if intValue, ok := customValue.(int64); ok {
				rm.checkCustomLimit(name, intValue, limit)
			}
		}
	}
}

// checkLimit 检查单个资源限制
func (rm *ResourceMonitor) checkLimit(resourceType ResourceType, current int64, limit *ResourceLimit) {
	if limit == nil || !limit.Enabled {
		return
	}

	softExceeded, hardExceeded := limit.IsExceeded(current)

	if hardExceeded {
		rm.sendAlert(resourceType, AlertLevelCritical, current, limit.HardLimit, limit.GetUtilization(current))
	} else if softExceeded {
		rm.sendAlert(resourceType, AlertLevelWarning, current, limit.SoftLimit, limit.GetUtilization(current))
	}
}

// checkCustomLimit 检查自定义资源限制
func (rm *ResourceMonitor) checkCustomLimit(name string, current int64, limit *ResourceLimit) {
	if limit == nil || !limit.Enabled {
		return
	}

	softExceeded, hardExceeded := limit.IsExceeded(current)

	if hardExceeded {
		rm.sendCustomAlert(name, AlertLevelCritical, current, limit.HardLimit, limit.GetUtilization(current))
	} else if softExceeded {
		rm.sendCustomAlert(name, AlertLevelWarning, current, limit.SoftLimit, limit.GetUtilization(current))
	}
}

// sendAlert 发送告警
func (rm *ResourceMonitor) sendAlert(resourceType ResourceType, level AlertLevel, current, limit int64, utilization float64) {
	alert := &ResourceAlert{
		PluginID:     rm.pluginID,
		ResourceType: resourceType,
		AlertLevel:   level,
		CurrentUsage: current,
		Limit:        limit,
		Utilization:  utilization,
		Timestamp:    time.Now(),
		Message:      fmt.Sprintf("Plugin %s %s usage %s: current=%d, limit=%d, utilization=%.2f%%", rm.pluginID, resourceType, level, current, limit, utilization),
	}

	rm.addToHistory(*alert)

	select {
	case rm.alerts <- alert:
		rm.logger.Warn("Resource alert",
			"plugin_id", rm.pluginID,
			"resource_type", resourceType,
			"level", level,
			"current", current,
			"limit", limit,
			"utilization", utilization)
	default:
		rm.logger.Error("Alert channel full, dropping alert", "plugin_id", rm.pluginID)
	}
}

// sendCustomAlert 发送自定义资源告警
func (rm *ResourceMonitor) sendCustomAlert(name string, level AlertLevel, current, limit int64, utilization float64) {
	alert := &ResourceAlert{
		PluginID:     rm.pluginID,
		ResourceType: ResourceType(name),
		AlertLevel:   level,
		CurrentUsage: current,
		Limit:        limit,
		Utilization:  utilization,
		Timestamp:    time.Now(),
		Message:      fmt.Sprintf("Plugin %s custom resource %s usage %s: current=%d, limit=%d, utilization=%.2f%%", rm.pluginID, name, level, current, limit, utilization),
	}

	rm.addToHistory(*alert)

	select {
	case rm.alerts <- alert:
		rm.logger.Warn("Custom resource alert",
			"plugin_id", rm.pluginID,
			"resource_name", name,
			"level", level,
			"current", current,
			"limit", limit,
			"utilization", utilization)
	default:
		rm.logger.Error("Alert channel full, dropping custom alert", "plugin_id", rm.pluginID)
	}
}

// addToHistory 添加到告警历史
func (rm *ResourceMonitor) addToHistory(alert ResourceAlert) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// 如果历史记录已满，移除最旧的记录
	if len(rm.alertHistory) >= rm.maxHistory {
		rm.alertHistory = rm.alertHistory[1:]
	}

	rm.alertHistory = append(rm.alertHistory, alert)
}

// ResourceManager 资源管理器
type ResourceManager struct {
	logger   *slog.Logger
	monitors map[string]*ResourceMonitor
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewResourceManager 创建资源管理器
func NewResourceManager(logger *slog.Logger) *ResourceManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &ResourceManager{
		logger:   logger,
		monitors: make(map[string]*ResourceMonitor),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start 启动资源管理器
func (rm *ResourceManager) Start(ctx context.Context) error {
	rm.logger.Info("Resource manager started")
	return nil
}

// Stop 停止资源管理器
func (rm *ResourceManager) Stop() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.cancel()

	// 停止所有监控器
	for _, monitor := range rm.monitors {
		monitor.Stop()
	}

	rm.logger.Info("Resource manager stopped")
}

// AddMonitor 添加资源监控器
func (rm *ResourceManager) AddMonitor(pluginID string, limits *DetailedResourceLimits) (*ResourceMonitor, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.monitors[pluginID]; exists {
		return nil, fmt.Errorf("monitor for plugin %s already exists", pluginID)
	}

	monitor := NewResourceMonitor(rm.logger, pluginID, limits)
	if err := monitor.Start(rm.ctx); err != nil {
		return nil, fmt.Errorf("failed to start monitor: %w", err)
	}

	rm.monitors[pluginID] = monitor
	return monitor, nil
}

// RemoveMonitor 移除资源监控器
func (rm *ResourceManager) RemoveMonitor(pluginID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if monitor, exists := rm.monitors[pluginID]; exists {
		monitor.Stop()
		delete(rm.monitors, pluginID)
	}
}

// GetMonitor 获取资源监控器
func (rm *ResourceManager) GetMonitor(pluginID string) *ResourceMonitor {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return rm.monitors[pluginID]
}

// GetAllUsage 获取所有插件的资源使用情况
func (rm *ResourceManager) GetAllUsage() map[string]*ResourceUsage {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	usage := make(map[string]*ResourceUsage)
	for pluginID, monitor := range rm.monitors {
		usage[pluginID] = monitor.GetUsage()
	}

	return usage
}

// GetSystemResourceUsage 获取系统总体资源使用情况
func (rm *ResourceManager) GetSystemResourceUsage() *ResourceUsage {
	allUsage := rm.GetAllUsage()

	systemUsage := &ResourceUsage{
		PluginID:    "system",
		Timestamp:   time.Now(),
		CustomUsage: make(map[string]interface{}),
	}

	// 聚合所有插件的资源使用情况
	for _, usage := range allUsage {
		systemUsage.Memory += usage.Memory
		systemUsage.CPU += usage.CPU
		systemUsage.DiskIO += usage.DiskIO
		systemUsage.NetworkIO += usage.NetworkIO
		systemUsage.Goroutines += usage.Goroutines
		systemUsage.FileDesc += usage.FileDesc
	}

	return systemUsage
}