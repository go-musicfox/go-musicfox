// pkg/plugin/health.go
package plugin

import (
	"context"
	"time"

	core "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
)

// HealthStatus 健康状态枚举
type HealthStatus int

const (
	HealthStatusUnknown HealthStatus = iota
	HealthStatusHealthy
	HealthStatusUnhealthy
	HealthStatusDegraded
	HealthStatusCritical
)

func (s HealthStatus) String() string {
	switch s {
	case HealthStatusHealthy:
		return "healthy"
	case HealthStatusUnhealthy:
		return "unhealthy"
	case HealthStatusDegraded:
		return "degraded"
	case HealthStatusCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// HealthMetrics 健康指标
type HealthMetrics struct {
	// 基础指标
	CPUUsage    float64   `json:"cpu_usage"`     // CPU使用率 (0-100)
	MemoryUsage int64         `json:"memory_usage"` // 内存使用量 (字节)
	Goroutines  int           `json:"goroutines"`   // 协程数量
	Uptime      time.Duration `json:"uptime"`   // 运行时间
	   
	// 性能指标
	ResponseTime    time.Duration `json:"response_time"`    // 平均响应时间
	Throughput   float64       `json:"throughput"`    // 吞吐量 (请求/秒)
	ErrorRate    float64       `json:"error_rate"`    // 错误率 (0-1)
	SuccessRate  float64       `json:"success_rate"`  // 成功率 (0-1)
	
// 资源指标
	DiskUsage       int64 `json:"disk_usage"`       // 磁盘使用量 (字节)
	NetworkInBytes  int64 `json:"network_in_bytes"`  // 网络入流量 (字节)
	NetworkOutBytes int64 `json:"network_out_bytes"`  // 网络出流量 (字节)
	
// 时间戳
	Timestamp time.Time `json:"timestamp"`
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	Status      HealthStatus    `json:"status"`
	Message     string          `json:"message"`
	Metrics   *HealthMetrics         `json:"metrics"`
	Details   map[string]interface{} `json:"details,omitempty"`
	CheckedAt time.Time              `json:"checked_at"`
	Duration  time.Duration          `json:"duration"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	// 检查间隔
	Interval time.Duration `json:"interval"`

	// 超时设置
	Timeout time.Duration `json:"timeout"`

	// 阈值配置
	Thresholds HealthThresholds `json:"thresholds"`

	// 重试配置
	MaxRetries int           `json:"max_retries"`
	RetryDelay time.Duration `json:"retry_delay"`

	// 启用的检查项
	EnabledChecks []string `json:"enabled_checks"`

	// 自动恢复
	AutoRecover bool `json:"auto_recover"`
}

// HealthThresholds 健康阈值配置
type HealthThresholds struct {
	// CPU阈值
	CPUWarning  float64 `json:"cpu_warning"`  // CPU警告阈值
	CPUCritical float64 `json:"cpu_critical"` // CPU严重阈值

	// 内存阈值
	MemoryWarning  int64 `json:"memory_warning"`  // 内存警告阈值 (字节)
	MemoryCritical int64 `json:"memory_critical"` // 内存严重阈值 (字节)

	// 响应时间阈值
	ResponseTimeWarning  time.Duration `json:"response_time_warning"`  // 响应时间警告阈值
	ResponseTimeCritical time.Duration `json:"response_time_critical"` // 响应时间严重阈值

	// 错误率阈值
	ErrorRateWarning  float64 `json:"error_rate_warning"`  // 错误率警告阈值
	ErrorRateCritical float64 `json:"error_rate_critical"` // 错误率严重阈值

	// 协程数阈值
	GoroutineWarning  int `json:"goroutine_warning"`  // 协程数警告阈值
	GoroutineCritical int `json:"goroutine_critical"` // 协程数严重阈值
}

// HealthChecker 健康检查器接口
type HealthChecker interface {
	// 启动健康检查
	Start(ctx context.Context) error

	// 停止健康检查
	Stop() error

	// 执行单次健康检查
	Check(ctx context.Context) (*HealthCheckResult, error)

	// 获取当前健康状态
	GetStatus() HealthStatus

	// 获取最新的健康指标
	GetMetrics() *HealthMetrics

	// 获取健康检查历史
	GetHistory(limit int) []*HealthCheckResult

	// 更新配置
	UpdateConfig(config *HealthCheckConfig) error

	// 注册健康检查回调
	OnStatusChange(callback func(oldStatus, newStatus HealthStatus, result *HealthCheckResult))

	// 注册指标收集器
	RegisterMetricsCollector(name string, collector MetricsCollector)

	// 注册恢复策略
	RegisterRecoveryStrategy(name string, strategy RecoveryStrategy)

	// 添加插件到健康检查
	AddPlugin(plugin interface{})

	// 开始监控指定插件
	StartMonitoring()

	// 停止监控指定插件
	StopMonitoring()
}

// MetricsCollector 指标收集器接口
type MetricsCollector interface {
	// 收集指标
	Collect(ctx context.Context) (map[string]interface{}, error)

	// 获取收集器名称
	GetName() string

	// 获取收集器描述
	GetDescription() string
}

// RecoveryStrategy 恢复策略接口
type RecoveryStrategy interface {
	// 执行恢复操作
	Recover(ctx context.Context, plugin core.Plugin, result *HealthCheckResult) error

	// 检查是否可以恢复
	CanRecover(result *HealthCheckResult) bool

	// 获取策略名称
	GetName() string

	// 获取策略描述
	GetDescription() string
}

// HealthEvent 健康事件
type HealthEvent struct {
	PluginName string                 `json:"plugin_name"`
	OldStatus  HealthStatus           `json:"old_status"`
	NewStatus  HealthStatus           `json:"new_status"`
	Result     *HealthCheckResult     `json:"result"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// HealthCheckStrategy 健康检查策略
type HealthCheckStrategy interface {
	// Execute 执行健康检查
	Execute(ctx context.Context, plugin core.Plugin) (*HealthCheckResult, error)

	// 检查健康状态（Execute的别名方法）
	Check(ctx context.Context, plugin core.Plugin) (*HealthCheckResult, error)

	// 获取策略名称
	GetName() string
}

// 默认健康阈值
var DefaultHealthThresholds = HealthThresholds{
	CPUWarning:           70.0,              // 70% CPU使用率警告
	CPUCritical:          90.0,              // 90% CPU使用率严重
	MemoryWarning:        100 * 1024 * 1024, // 100MB内存警告
	MemoryCritical:       500 * 1024 * 1024, // 500MB内存严重
	ResponseTimeWarning:  time.Second,       // 1秒响应时间警告
	ResponseTimeCritical: 5 * time.Second,   // 5秒响应时间严重
	ErrorRateWarning:     0.05,              // 5%错误率警告
	ErrorRateCritical:    0.20,              // 20%错误率严重
	GoroutineWarning:     100,               // 100个协程警告
	GoroutineCritical:    500,               // 500个协程严重
}            

// 默认健康检查配置
var DefaultHealthCheckConfig = HealthCheckConfig{
	Interval:      30 * time.Second,
	Timeout:       10 * time.Second,
	Thresholds:    DefaultHealthThresholds,
	MaxRetries:    3,
	RetryDelay:    5 * time.Second,
	EnabledChecks: []string{"basic", "performance", "resources"},
	AutoRecover:   true,
}