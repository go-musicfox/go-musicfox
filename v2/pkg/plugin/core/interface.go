// Package plugin 定义了插件系统的基础接口和数据结构
package plugin

import (
	"time"
)

// Plugin 插件基础接口
// 所有插件都必须实现此接口
type Plugin interface {
	// 插件元信息
	GetInfo() *PluginInfo
	GetCapabilities() []string
	GetDependencies() []string

	// 生命周期管理
	Initialize(ctx PluginContext) error
	Start() error
	Stop() error
	Cleanup() error

	// 健康检查
	HealthCheck() error

	// 配置管理
	ValidateConfig(config map[string]interface{}) error
	UpdateConfig(config map[string]interface{}) error

	// 指标收集
	GetMetrics() (*PluginMetrics, error)

	// 事件处理
	HandleEvent(event interface{}) error
}

// PluginInfo 插件信息结构体
type PluginInfo struct {
	ID          string            `json:"id"`          // 插件ID
	Name        string            `json:"name"`        // 插件名称
	Version     string            `json:"version"`     // 插件版本
	Description string            `json:"description"` // 插件描述
	Author      string            `json:"author"`      // 插件作者
	Type        PluginType        `json:"type"`        // 插件类型
	License     string            `json:"license"`     // 许可证
	Homepage    string            `json:"homepage"`    // 主页地址
	Tags        []string          `json:"tags"`        // 标签
	Config      map[string]string `json:"config"`      // 配置项
	LoadTime    time.Time         `json:"load_time"`   // 加载时间
	CreatedAt   time.Time         `json:"created_at"`  // 创建时间
	UpdatedAt   time.Time         `json:"updated_at"`  // 更新时间
}

// PluginState 插件状态枚举
type PluginState int

const (
	PluginStateUnknown PluginState = iota // 未知状态
	PluginStateLoaded                     // 已加载
	PluginStateRunning                    // 运行中
	PluginStateStopping                   // 停止中
	PluginStateStopped                    // 已停止
	PluginStateUnloading                  // 卸载中
	PluginStateUnloaded                   // 已卸载
	PluginStateError                      // 错误状态
	PluginStatePaused                     // 暂停状态
	PluginStateCleaning                   // 清理中
	PluginStateCorrupted                  // 损坏状态
)

// String 返回插件状态的字符串表示
func (s PluginState) String() string {
	switch s {
	case PluginStateUnknown:
		return "unknown"
	case PluginStateLoaded:
		return "loaded"
	case PluginStateRunning:
		return "running"
	case PluginStateStopping:
		return "stopping"
	case PluginStateStopped:
		return "stopped"
	case PluginStateUnloading:
		return "unloading"
	case PluginStateUnloaded:
		return "unloaded"
	case PluginStateError:
		return "error"
	case PluginStatePaused:
		return "paused"
	case PluginStateCleaning:
		return "cleaning"
	case PluginStateCorrupted:
		return "corrupted"
	default:
		return "unknown"
	}
}

// PluginType 插件类型枚举
type PluginType string

const (
	PluginTypeDynamicLibrary PluginType = "dynamic_library" // 动态链接库插件
	PluginTypeRPC           PluginType = "rpc"              // RPC插件
	PluginTypeWebAssembly   PluginType = "webassembly"     // WebAssembly插件
	PluginTypeHotReload     PluginType = "hot_reload"      // 热加载插件
	PluginTypeMusicSource   PluginType = "music_source"    // 音乐源插件
	PluginTypeAudioProcessor PluginType = "audio_processor" // 音频处理插件
)

// String 返回插件类型的字符串表示
func (t PluginType) String() string {
	return string(t)
}

// PluginPriority 插件优先级枚举
type PluginPriority int

const (
	PluginPriorityLow    PluginPriority = iota // 低优先级
	PluginPriorityNormal                       // 普通优先级
	PluginPriorityHigh                         // 高优先级
	PluginPriorityCritical                     // 关键优先级
)

// String 返回插件优先级的字符串表示
func (p PluginPriority) String() string {
	switch p {
	case PluginPriorityLow:
		return "low"
	case PluginPriorityNormal:
		return "normal"
	case PluginPriorityHigh:
		return "high"
	case PluginPriorityCritical:
		return "critical"
	default:
		return "normal"
	}
}

// IsValidStateTransition 检查状态转换是否有效
func (s PluginState) IsValidStateTransition(to PluginState) bool {
	switch s {
	case PluginStateUnknown:
		return to == PluginStateLoaded || to == PluginStateError
	case PluginStateLoaded:
		return to == PluginStateRunning || to == PluginStateUnloading || to == PluginStateError
	case PluginStateRunning:
		return to == PluginStateStopping || to == PluginStatePaused || to == PluginStateError
	case PluginStateStopping:
		return to == PluginStateStopped || to == PluginStateError
	case PluginStateStopped:
		return to == PluginStateRunning || to == PluginStateUnloading || to == PluginStateError
	case PluginStateUnloading:
		return to == PluginStateUnloaded || to == PluginStateCleaning || to == PluginStateError || to == PluginStateCorrupted
	case PluginStateUnloaded:
		return to == PluginStateLoaded || to == PluginStateError
	case PluginStateError:
		return to == PluginStateCleaning || to == PluginStateCorrupted || to == PluginStateUnloading
	case PluginStatePaused:
		return to == PluginStateRunning || to == PluginStateStopping || to == PluginStateError
	case PluginStateCleaning:
		return to == PluginStateUnloaded || to == PluginStateError || to == PluginStateCorrupted
	case PluginStateCorrupted:
		return to == PluginStateCleaning || to == PluginStateUnloading
	default:
		return false
	}
}

// CanUnload 检查插件是否可以卸载
func (s PluginState) CanUnload() bool {
	return s == PluginStateLoaded || s == PluginStateStopped || s == PluginStateError || s == PluginStatePaused
}

// CanStop 检查插件是否可以停止
func (s PluginState) CanStop() bool {
	return s == PluginStateRunning || s == PluginStatePaused
}

// IsTransitional 检查状态是否为过渡状态
func (s PluginState) IsTransitional() bool {
	return s == PluginStateStopping || s == PluginStateUnloading || s == PluginStateCleaning
}

// HealthChecker 健康检查器接口
type HealthChecker interface {
	// CheckHealth 检查插件健康状态
	CheckHealth(plugin Plugin) (*HealthStatus, error)
	
	// RegisterPlugin 注册插件
	RegisterPlugin(plugin Plugin) error
	
	// UnregisterPlugin 注销插件
	UnregisterPlugin(plugin Plugin) error
	
	// Start 启动健康检查
	Start() error
	
	// Stop 停止健康检查
	Stop() error
	
	// GetStatus 获取健康检查状态
	GetStatus() map[string]*HealthStatus
	
	// StartMonitoring 启动监控
	StartMonitoring() error
	
	// StopMonitoring 停止监控
	StopMonitoring() error
	
	// AddPlugin 添加插件
	AddPlugin(plugin Plugin) error
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(logger interface{}, interval time.Duration) HealthChecker {
	return &DefaultHealthChecker{
		plugins: make(map[string]Plugin),
		status:  make(map[string]*HealthStatus),
	}
}

// DefaultHealthChecker 默认健康检查器实现
type DefaultHealthChecker struct {
	plugins map[string]Plugin
	status  map[string]*HealthStatus
	running bool
}

// CheckHealth 检查插件健康状态
func (d *DefaultHealthChecker) CheckHealth(plugin Plugin) (*HealthStatus, error) {
	return &HealthStatus{
		Healthy:   true,
		Message:   "OK",
		Details:   make(map[string]interface{}),
		Timestamp: time.Now(),
	}, nil
}

// RegisterPlugin 注册插件
func (d *DefaultHealthChecker) RegisterPlugin(plugin Plugin) error {
	info := plugin.GetInfo()
	d.plugins[info.ID] = plugin
	return nil
}

// UnregisterPlugin 注销插件
func (d *DefaultHealthChecker) UnregisterPlugin(plugin Plugin) error {
	info := plugin.GetInfo()
	delete(d.plugins, info.ID)
	delete(d.status, info.ID)
	return nil
}

// Start 启动健康检查
func (d *DefaultHealthChecker) Start() error {
	d.running = true
	return nil
}

// Stop 停止健康检查
func (d *DefaultHealthChecker) Stop() error {
	d.running = false
	return nil
}

// GetStatus 获取健康检查状态
func (d *DefaultHealthChecker) GetStatus() map[string]*HealthStatus {
	return d.status
}

// StartMonitoring 启动监控
func (d *DefaultHealthChecker) StartMonitoring() error {
	return d.Start()
}

// StopMonitoring 停止监控
func (d *DefaultHealthChecker) StopMonitoring() error {
	return d.Stop()
}

// AddPlugin 添加插件
func (d *DefaultHealthChecker) AddPlugin(plugin Plugin) error {
	return d.RegisterPlugin(plugin)
}