// Package loader 定义插件加载器共享的类型和接口
package loader

import (
	"context"
	"time"
)

// 常量定义
const (
	EnforceModeLimit = "limit"
	EnforceModeWarn  = "warn"
	EnforceModeKill  = "kill"
)

// LoaderType 插件加载器类型枚举
type LoaderType string

const (
	LoaderTypeDynamic LoaderType = "dynamic"    // 动态库加载器
	LoaderTypeRPC     LoaderType = "rpc"        // RPC加载器
	LoaderTypeWASM    LoaderType = "wasm"       // WASM加载器
	LoaderTypeHotReload LoaderType = "hotreload" // 热重载加载器
)

// String 返回加载器类型的字符串表示
func (lt LoaderType) String() string {
	return string(lt)
}

// PluginType 插件类型枚举
type PluginType string

const (
	PluginTypeDynamicLibrary PluginType = "dynamic_library" // 动态链接库插件
	PluginTypeRPC           PluginType = "rpc"              // RPC插件
	PluginTypeWebAssembly   PluginType = "webassembly"     // WebAssembly插件
	PluginTypeHotReload     PluginType = "hot_reload"      // 热加载插件
)

// PluginState 插件状态类型
type PluginState int

const (
	PluginStateUnknown PluginState = iota
	PluginStateLoaded
	PluginStateInitialized
	PluginStateRunning
	PluginStateStopping
	PluginStateStopped
	PluginStateUnloading
	PluginStateError
	PluginStateUnloaded
)

// String 返回插件状态的字符串表示
func (s PluginState) String() string {
	switch s {
	case PluginStateUnknown:
		return "unknown"
	case PluginStateLoaded:
		return "loaded"
	case PluginStateInitialized:
		return "initialized"
	case PluginStateRunning:
		return "running"
	case PluginStateStopping:
		return "stopping"
	case PluginStateStopped:
		return "stopped"
	case PluginStateUnloading:
		return "unloading"
	case PluginStateError:
		return "error"
	case PluginStateUnloaded:
		return "unloaded"
	default:
		return "unknown"
	}
}

// Plugin 插件接口
type Plugin interface {
	// GetInfo 获取插件信息
	GetInfo() *PluginInfo
	
	// GetCapabilities 获取插件能力
	GetCapabilities() []string
	
	// GetDependencies 获取插件依赖
	GetDependencies() []string
	
	// Initialize 初始化插件
	Initialize(ctx PluginContext) error
	
	// Start 启动插件
	Start() error
	
	// Stop 停止插件
	Stop() error
	
	// Cleanup 清理插件资源
	Cleanup() error
	
	// GetMetrics 获取插件指标
	GetMetrics() (*PluginMetrics, error)
}

// PluginInfo 插件信息
type PluginInfo struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	Type        string                 `json:"type"`
	Path        string                 `json:"path"`
	Config      map[string]interface{} `json:"config"`
	Dependencies []string              `json:"dependencies"`
	Metadata    map[string]interface{} `json:"metadata"`
	LoadTime    time.Time              `json:"load_time"`
}

// PluginContext 插件上下文接口
type PluginContext interface {
	// GetContext 获取上下文
	GetContext() context.Context
	
	// GetEventBus 获取事件总线
	GetEventBus() EventBus
	
	// GetServiceRegistry 获取服务注册表
	GetServiceRegistry() ServiceRegistry
	
	// GetLogger 获取日志器
	GetLogger() Logger
	
	// GetConfig 获取配置
	GetConfig() map[string]interface{}
}

// EventBus 事件总线接口
type EventBus interface {
	// Publish 发布事件
	Publish(event string, data interface{}) error
	
	// Subscribe 订阅事件
	Subscribe(event string, handler func(interface{})) error
	
	// Unsubscribe 取消订阅事件
	Unsubscribe(event string, handler func(interface{})) error
}

// ServiceRegistry 服务注册表接口
type ServiceRegistry interface {
	// RegisterService 注册服务
	RegisterService(name string, service interface{}) error
	
	// GetService 获取服务
	GetService(name string) (interface{}, error)
	
	// UnregisterService 注销服务
	UnregisterService(name string) error
	
	// ListServices 列出所有服务
	ListServices() []string
}

// Logger 日志接口
type Logger interface {
	// Debug 调试日志
	Debug(msg string, args ...interface{})
	
	// Info 信息日志
	Info(msg string, args ...interface{})
	
	// Warn 警告日志
	Warn(msg string, args ...interface{})
	
	// Error 错误日志
	Error(msg string, args ...interface{})
}

// PluginMetrics 插件指标
type PluginMetrics struct {
	PluginID      string                 `json:"plugin_id"`
	Uptime        time.Duration          `json:"uptime"`
	MemoryUsage   int64                  `json:"memory_usage"`
	CPUUsage      float64                `json:"cpu_usage"`
	RequestCount  int64                  `json:"request_count"`
	ErrorCount    int64                  `json:"error_count"`
	SuccessRate   float64                `json:"success_rate"`
	CustomMetrics map[string]interface{} `json:"custom_metrics"`
	Timestamp     time.Time              `json:"timestamp"`
}

// ResourceLimits 资源限制
type ResourceLimits struct {
	MaxMemoryMB      int           `json:"max_memory_mb"`
	MaxCPUPercent    float64       `json:"max_cpu_percent"`
	MaxGoroutines    int           `json:"max_goroutines"`
	MaxFileHandles   int           `json:"max_file_handles"`
	MaxNetworkConn   int           `json:"max_network_conn"`
	ExecutionTimeout time.Duration `json:"execution_timeout"`
	IdleTimeout      time.Duration `json:"idle_timeout"`
	StartupTimeout   time.Duration `json:"startup_timeout"`
	ShutdownTimeout  time.Duration `json:"shutdown_timeout"`
	Enabled          bool          `json:"enabled"`
	EnforceMode      string        `json:"enforce_mode"`
}

// PluginLoader 插件加载器接口
type PluginLoader interface {
	// LoadPlugin 从指定路径加载插件
	LoadPlugin(ctx context.Context, pluginPath string) (Plugin, error)
	
	// UnloadPlugin 卸载指定的插件
	UnloadPlugin(ctx context.Context, pluginID string) error
	
	// GetLoadedPlugins 获取已加载的插件列表
	GetLoadedPlugins() map[string]Plugin
	
	// GetPluginInfo 获取插件信息
	GetPluginInfo(pluginID string) (*PluginInfo, error)
	
	// ValidatePlugin 验证插件是否有效
	ValidatePlugin(pluginPath string) error
	
	// GetLoaderType 获取加载器类型
	GetLoaderType() PluginType
	
	// GetLoaderInfo 获取加载器信息
	GetLoaderInfo() map[string]interface{}
	
	// Shutdown 关闭加载器
	Shutdown(ctx context.Context) error
	
	// ReloadPlugin 重新加载插件
	ReloadPlugin(ctx context.Context, pluginID string) error
	
	// IsPluginLoaded 检查插件是否已加载
	IsPluginLoaded(pluginID string) bool
	
	// Cleanup 清理资源
	Cleanup() error
}