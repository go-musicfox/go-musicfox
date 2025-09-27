// Package thirdparty 实现第三方插件接口，支持WebAssembly插件的安全执行
package thirdparty

import (
	"context"
	"time"

	core "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
)

// ThirdPartyPlugin 第三方插件接口（WebAssembly实现）
type ThirdPartyPlugin interface {
	core.Plugin

	// WASM特定方法
	GetWASMModule() []byte
	GetExportedFunctions() []string

	// 沙箱执行
	ExecuteFunction(functionName string, args []interface{}) (interface{}, error)

	// 资源限制
	GetResourceLimits() *ResourceLimits
	SetResourceLimits(limits *ResourceLimits) error

	// 沙箱管理
	GetSandboxConfig() *SandboxConfig
	SetSandboxConfig(config *SandboxConfig) error

	// 资源监控
	GetResourceUsage() *ResourceUsage
	StartResourceMonitoring(ctx context.Context) error
	StopResourceMonitoring() error
}

// ResourceLimits 资源限制配置
type ResourceLimits struct {
	MaxMemory     int64         `json:"max_memory"`      // 最大内存使用量（字节）
	MaxCPU        float64       `json:"max_cpu"`         // 最大CPU使用率（0-1）
	MaxDiskIO     int64         `json:"max_disk_io"`     // 最大磁盘IO（字节/秒）
	MaxNetworkIO  int64         `json:"max_network_io"`  // 最大网络IO（字节/秒）
	Timeout       time.Duration `json:"timeout"`         // 执行超时时间
	MaxGoroutines int           `json:"max_goroutines"`  // 最大协程数
	MaxFileSize   int64         `json:"max_file_size"`   // 最大文件大小
	MaxOpenFiles  int           `json:"max_open_files"`  // 最大打开文件数
}

// SandboxConfig 沙箱配置
type SandboxConfig struct {
	Enabled         bool     `json:"enabled"`           // 是否启用沙箱
	AllowedPaths    []string `json:"allowed_paths"`     // 允许访问的路径
	AllowedNetworks []string `json:"allowed_networks"`  // 允许访问的网络
	AllowedSyscalls []string `json:"allowed_syscalls"`  // 允许的系统调用
	TrustedSources  []string `json:"trusted_sources"`   // 可信的插件源
	IsolationLevel  IsolationLevel `json:"isolation_level"` // 隔离级别
	NetworkAccess   bool     `json:"network_access"`    // 是否允许网络访问
	FileSystemAccess bool    `json:"filesystem_access"` // 是否允许文件系统访问
}

// IsolationLevel 隔离级别枚举
type IsolationLevel int

const (
	IsolationLevelNone IsolationLevel = iota // 无隔离
	IsolationLevelBasic                      // 基础隔离
	IsolationLevelStrict                     // 严格隔离
	IsolationLevelComplete                   // 完全隔离
)

// String 返回隔离级别的字符串表示
func (l IsolationLevel) String() string {
	switch l {
	case IsolationLevelNone:
		return "none"
	case IsolationLevelBasic:
		return "basic"
	case IsolationLevelStrict:
		return "strict"
	case IsolationLevelComplete:
		return "complete"
	default:
		return "unknown"
	}
}

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	MemoryUsage   int64         `json:"memory_usage"`   // 当前内存使用量
	CPUUsage      float64       `json:"cpu_usage"`      // 当前CPU使用率
	DiskIOUsage   int64         `json:"disk_io_usage"`  // 当前磁盘IO使用量
	NetworkIOUsage int64        `json:"network_io_usage"` // 当前网络IO使用量
	GoroutineCount int          `json:"goroutine_count"` // 当前协程数量
	OpenFileCount int          `json:"open_file_count"` // 当前打开文件数
	Uptime        time.Duration `json:"uptime"`         // 运行时间
	LastUpdated   time.Time     `json:"last_updated"`   // 最后更新时间
}

// WASMFunction WASM函数信息
type WASMFunction struct {
	Name       string                 `json:"name"`        // 函数名称
	Params     []WASMType            `json:"params"`      // 参数类型
	Results    []WASMType            `json:"results"`     // 返回值类型
	Exported   bool                  `json:"exported"`    // 是否导出
	Metadata   map[string]interface{} `json:"metadata"`   // 元数据
}

// WASMType WASM类型枚举
type WASMType int

const (
	WASMTypeI32 WASMType = iota // 32位整数
	WASMTypeI64                 // 64位整数
	WASMTypeF32                 // 32位浮点数
	WASMTypeF64                 // 64位浮点数
	WASMTypeV128                // 128位向量
	WASMTypeFuncRef             // 函数引用
	WASMTypeExternRef           // 外部引用
)

// String 返回WASM类型的字符串表示
func (t WASMType) String() string {
	switch t {
	case WASMTypeI32:
		return "i32"
	case WASMTypeI64:
		return "i64"
	case WASMTypeF32:
		return "f32"
	case WASMTypeF64:
		return "f64"
	case WASMTypeV128:
		return "v128"
	case WASMTypeFuncRef:
		return "funcref"
	case WASMTypeExternRef:
		return "externref"
	default:
		return "unknown"
	}
}

// ExecutionContext 执行上下文
type ExecutionContext struct {
	Timeout     time.Duration          `json:"timeout"`      // 执行超时
	MemoryLimit int64                 `json:"memory_limit"` // 内存限制
	GasLimit    int64                 `json:"gas_limit"`    // Gas限制
	Metadata    map[string]interface{} `json:"metadata"`    // 元数据
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	Value       interface{}   `json:"value"`        // 返回值
	Error       string        `json:"error"`        // 错误信息
	GasUsed     int64         `json:"gas_used"`     // 使用的Gas
	MemoryUsed  int64         `json:"memory_used"`  // 使用的内存
	Duration    time.Duration `json:"duration"`     // 执行时间
	Success     bool          `json:"success"`      // 是否成功
}

// SecurityPolicy 安全策略
type SecurityPolicy struct {
	AllowUnsafeOperations bool     `json:"allow_unsafe_operations"` // 是否允许不安全操作
	TrustedDomains        []string `json:"trusted_domains"`        // 可信域名
	BlockedDomains        []string `json:"blocked_domains"`        // 阻止的域名
	MaxRequestSize        int64    `json:"max_request_size"`       // 最大请求大小
	RateLimitRPS          int      `json:"rate_limit_rps"`         // 速率限制（请求/秒）
}