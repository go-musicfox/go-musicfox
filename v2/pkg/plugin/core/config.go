package plugin

import (
	"context"
	"fmt"
	"time"
)

// ErrorCode 错误代码类型
type ErrorCode int

const (
	ErrorCodePluginConfigInvalid ErrorCode = iota + 1000
	ErrorCodePluginNotFound
	ErrorCodePluginLoadFailed
	ErrorCodePluginStartFailed
	ErrorCodePermissionDenied
)

// PluginError 插件错误
type PluginError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Cause   error     `json:"cause,omitempty"`
}

// Error 实现error接口
func (e *PluginError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("plugin error [%d]: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("plugin error [%d]: %s", e.Code, e.Message)
}

// NewPluginError 创建新的插件错误
func NewPluginError(code ErrorCode, message string) *PluginError {
	return &PluginError{
		Code:    code,
		Message: message,
	}
}

// PluginConfig 插件配置接口
type PluginConfig interface {
	GetID() string
	GetName() string
	GetVersion() string
	GetEnabled() bool
	GetPriority() PluginPriority
	GetDependencies() []string
	GetResourceLimits() *ResourceLimits
	GetSecurityConfig() *SecurityConfig
	GetCustomConfig() map[string]interface{}
	Validate() error
}

// BasePluginConfig 基础插件配置
type BasePluginConfig struct {
	ID             string                 `json:"id" yaml:"id"`                           // 插件ID
	Name           string                 `json:"name" yaml:"name"`                       // 插件名称
	Version        string                 `json:"version" yaml:"version"`                 // 插件版本
	Description    string                 `json:"description" yaml:"description"`         // 插件描述
	Author         string                 `json:"author" yaml:"author"`                   // 插件作者
	Homepage       string                 `json:"homepage" yaml:"homepage"`               // 插件主页
	License        string                 `json:"license" yaml:"license"`                 // 许可证
	Enabled        bool                   `json:"enabled" yaml:"enabled"`                 // 是否启用
	Priority       PluginPriority         `json:"priority" yaml:"priority"`               // 优先级
	Type           PluginType             `json:"type" yaml:"type"`                       // 插件类型
	Tags           []string               `json:"tags" yaml:"tags"`                       // 标签
	Dependencies   []string               `json:"dependencies" yaml:"dependencies"`       // 依赖项
	Conflicts      []string               `json:"conflicts" yaml:"conflicts"`             // 冲突项
	MinKernelVersion string               `json:"min_kernel_version" yaml:"min_kernel_version"` // 最小内核版本
	MaxKernelVersion string               `json:"max_kernel_version" yaml:"max_kernel_version"` // 最大内核版本
	ResourceLimits *ResourceLimits        `json:"resource_limits" yaml:"resource_limits"` // 资源限制
	SecurityConfig *SecurityConfig        `json:"security_config" yaml:"security_config"` // 安全配置
	CustomConfig   map[string]interface{} `json:"custom_config" yaml:"custom_config"`     // 自定义配置
	CreatedAt      time.Time              `json:"created_at" yaml:"created_at"`           // 创建时间
	UpdatedAt      time.Time              `json:"updated_at" yaml:"updated_at"`           // 更新时间
}

// GetID 获取插件ID
func (c *BasePluginConfig) GetID() string {
	return c.ID
}

// GetName 获取插件名称
func (c *BasePluginConfig) GetName() string {
	return c.Name
}

// GetVersion 获取插件版本
func (c *BasePluginConfig) GetVersion() string {
	return c.Version
}

// GetEnabled 获取是否启用
func (c *BasePluginConfig) GetEnabled() bool {
	return c.Enabled
}

// GetPriority 获取优先级
func (c *BasePluginConfig) GetPriority() PluginPriority {
	return c.Priority
}

// GetDependencies 获取依赖项
func (c *BasePluginConfig) GetDependencies() []string {
	return c.Dependencies
}

// GetResourceLimits 获取资源限制
func (c *BasePluginConfig) GetResourceLimits() *ResourceLimits {
	return c.ResourceLimits
}

// GetSecurityConfig 获取安全配置
func (c *BasePluginConfig) GetSecurityConfig() *SecurityConfig {
	return c.SecurityConfig
}

// GetCustomConfig 获取自定义配置
func (c *BasePluginConfig) GetCustomConfig() map[string]interface{} {
	return c.CustomConfig
}

// Validate 验证配置
func (c *BasePluginConfig) Validate() error {
	if c.ID == "" {
		return NewPluginError(ErrorCodePluginConfigInvalid, "plugin ID is required")
	}
	if c.Name == "" {
		return NewPluginError(ErrorCodePluginConfigInvalid, "plugin name is required")
	}
	if c.Version == "" {
		return NewPluginError(ErrorCodePluginConfigInvalid, "plugin version is required")
	}
	if c.Type == "" {
		return NewPluginError(ErrorCodePluginConfigInvalid, "plugin type is required")
	}
	
	// 验证资源限制
	if c.ResourceLimits != nil {
		if err := c.ResourceLimits.Validate(); err != nil {
			return err
		}
	}
	
	// 验证安全配置
	if c.SecurityConfig != nil {
		if err := c.SecurityConfig.Validate(); err != nil {
			return err
		}
	}
	
	return nil
}

// ResourceLimits 资源限制配置
type ResourceLimits struct {
	MaxMemoryMB    int64         `json:"max_memory_mb" yaml:"max_memory_mb"`       // 最大内存(MB)
	MaxCPUPercent  float64       `json:"max_cpu_percent" yaml:"max_cpu_percent"`   // 最大CPU使用率(%)
	MaxGoroutines  int           `json:"max_goroutines" yaml:"max_goroutines"`     // 最大协程数
	MaxFileHandles int           `json:"max_file_handles" yaml:"max_file_handles"` // 最大文件句柄数
	MaxNetworkConn int           `json:"max_network_conn" yaml:"max_network_conn"` // 最大网络连接数
	MaxFileDescriptors int       `json:"max_file_descriptors" yaml:"max_file_descriptors"` // 最大文件描述符数
	MaxNetworkConnections int    `json:"max_network_connections" yaml:"max_network_connections"` // 最大网络连接数
	MaxDiskUsageMB int64         `json:"max_disk_usage_mb" yaml:"max_disk_usage_mb"` // 最大磁盘使用(MB)
	ExecutionTimeout time.Duration `json:"execution_timeout" yaml:"execution_timeout"` // 执行超时
	IdleTimeout    time.Duration `json:"idle_timeout" yaml:"idle_timeout"`         // 空闲超时
	StartupTimeout time.Duration `json:"startup_timeout" yaml:"startup_timeout"`   // 启动超时
	ShutdownTimeout time.Duration `json:"shutdown_timeout" yaml:"shutdown_timeout"` // 关闭超时
	Enabled        bool          `json:"enabled" yaml:"enabled"`                   // 是否启用限制
	EnforceMode    EnforceMode   `json:"enforce_mode" yaml:"enforce_mode"`         // 强制模式
}

// Validate 验证资源限制配置
func (r *ResourceLimits) Validate() error {
	if r.MaxMemoryMB < 0 {
		return NewPluginError(ErrorCodePluginConfigInvalid, "max memory must be non-negative")
	}
	if r.MaxCPUPercent < 0 || r.MaxCPUPercent > 100 {
		return NewPluginError(ErrorCodePluginConfigInvalid, "max CPU percent must be between 0 and 100")
	}
	if r.MaxGoroutines < 0 {
		return NewPluginError(ErrorCodePluginConfigInvalid, "max goroutines must be non-negative")
	}
	if r.MaxFileHandles < 0 {
		return NewPluginError(ErrorCodePluginConfigInvalid, "max file handles must be non-negative")
	}
	if r.MaxNetworkConn < 0 {
		return NewPluginError(ErrorCodePluginConfigInvalid, "max network connections must be non-negative")
	}
	if r.MaxFileDescriptors < 0 {
		return NewPluginError(ErrorCodePluginConfigInvalid, "max file descriptors must be non-negative")
	}
	if r.MaxNetworkConnections < 0 {
		return NewPluginError(ErrorCodePluginConfigInvalid, "max network connections must be non-negative")
	}
	if r.MaxDiskUsageMB < 0 {
		return NewPluginError(ErrorCodePluginConfigInvalid, "max disk usage must be non-negative")
	}
	return nil
}

// EnforceMode 强制模式枚举
type EnforceMode int

const (
	EnforceModeWarn EnforceMode = iota // 警告模式
	EnforceModeLimit                   // 限制模式
	EnforceModeKill                    // 终止模式
)

// String 返回强制模式的字符串表示
func (e EnforceMode) String() string {
	switch e {
	case EnforceModeWarn:
		return "warn"
	case EnforceModeLimit:
		return "limit"
	case EnforceModeKill:
		return "kill"
	default:
		return "unknown"
	}
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	SandboxEnabled    bool                   `json:"sandbox_enabled" yaml:"sandbox_enabled"`       // 是否启用沙箱
	Permissions       []Permission           `json:"permissions" yaml:"permissions"`               // 权限列表
	AllowedHosts      []string               `json:"allowed_hosts" yaml:"allowed_hosts"`           // 允许的主机
	AllowedPorts      []int                  `json:"allowed_ports" yaml:"allowed_ports"`           // 允许的端口
	AllowedPaths      []string               `json:"allowed_paths" yaml:"allowed_paths"`           // 允许的路径
	BlockedPaths      []string               `json:"blocked_paths" yaml:"blocked_paths"`           // 禁止的路径
	AllowedCommands   []string               `json:"allowed_commands" yaml:"allowed_commands"`     // 允许的命令
	BlockedCommands   []string               `json:"blocked_commands" yaml:"blocked_commands"`     // 禁止的命令
	EncryptionEnabled bool                   `json:"encryption_enabled" yaml:"encryption_enabled"` // 是否启用加密
	SignatureRequired bool                   `json:"signature_required" yaml:"signature_required"` // 是否需要签名
	TrustedSources    []string               `json:"trusted_sources" yaml:"trusted_sources"`       // 可信来源
	SecurityLevel     SecurityLevel          `json:"security_level" yaml:"security_level"`         // 安全级别
	AuditEnabled      bool                   `json:"audit_enabled" yaml:"audit_enabled"`           // 是否启用审计
	CustomRules       map[string]interface{} `json:"custom_rules" yaml:"custom_rules"`             // 自定义规则
}

// Validate 验证安全配置
func (s *SecurityConfig) Validate() error {
	// 验证权限
	for _, perm := range s.Permissions {
		if perm == PermissionUnknown {
			return NewPluginError(ErrorCodePluginConfigInvalid, "invalid permission")
		}
	}
	
	// 验证端口范围
	for _, port := range s.AllowedPorts {
		if port < 1 || port > 65535 {
			return NewPluginError(ErrorCodePluginConfigInvalid, "invalid port number")
		}
	}
	
	return nil
}

// Permission 权限枚举
type Permission int

const (
	PermissionUnknown Permission = iota
	PermissionFileRead
	PermissionFileWrite
	PermissionFileExecute
	PermissionNetworkAccess
	PermissionSystemCall
	PermissionProcessControl
	PermissionEnvironmentAccess
	PermissionConfigAccess
	PermissionDatabaseAccess
	PermissionAudioAccess
	PermissionUIAccess
	PermissionPluginManagement
	PermissionKernelAccess
	PermissionEventAccess
	PermissionServiceAccess
	PermissionAll
)

// String 返回权限的字符串表示
func (p Permission) String() string {
	switch p {
	case PermissionFileRead:
		return "file_read"
	case PermissionFileWrite:
		return "file_write"
	case PermissionFileExecute:
		return "file_execute"
	case PermissionNetworkAccess:
		return "network_access"
	case PermissionSystemCall:
		return "system_call"
	case PermissionProcessControl:
		return "process_control"
	case PermissionEnvironmentAccess:
		return "environment_access"
	case PermissionConfigAccess:
		return "config_access"
	case PermissionDatabaseAccess:
		return "database_access"
	case PermissionAudioAccess:
		return "audio_access"
	case PermissionUIAccess:
		return "ui_access"
	case PermissionPluginManagement:
		return "plugin_management"
	case PermissionKernelAccess:
		return "kernel_access"
	case PermissionEventAccess:
		return "event_access"
	case PermissionServiceAccess:
		return "service_access"
	case PermissionAll:
		return "all"
	default:
		return "unknown"
	}
}

// SecurityLevel 安全级别枚举
type SecurityLevel int

const (
	SecurityLevelNone SecurityLevel = iota
	SecurityLevelLow
	SecurityLevelMedium
	SecurityLevelHigh
	SecurityLevelStrict
)

// String 返回安全级别的字符串表示
func (s SecurityLevel) String() string {
	switch s {
	case SecurityLevelNone:
		return "none"
	case SecurityLevelLow:
		return "low"
	case SecurityLevelMedium:
		return "medium"
	case SecurityLevelHigh:
		return "high"
	case SecurityLevelStrict:
		return "strict"
	default:
		return "unknown"
	}
}

// ConfigManager 配置管理器接口
type ConfigManager interface {
	LoadConfig(ctx context.Context, pluginID string) (PluginConfig, error)
	SaveConfig(ctx context.Context, config PluginConfig) error
	DeleteConfig(ctx context.Context, pluginID string) error
	ListConfigs(ctx context.Context) ([]PluginConfig, error)
	ValidateConfig(ctx context.Context, config PluginConfig) error
	GetDefaultConfig(ctx context.Context, pluginType PluginType) (PluginConfig, error)
	MergeConfig(ctx context.Context, base, override PluginConfig) (PluginConfig, error)
	WatchConfig(ctx context.Context, pluginID string) (<-chan PluginConfig, error)
}

// ConfigValidator 配置验证器接口
type ConfigValidator interface {
	Validate(ctx context.Context, config PluginConfig) error
	GetSchema(ctx context.Context, pluginType PluginType) (interface{}, error)
	ValidateSchema(ctx context.Context, config PluginConfig, schema interface{}) error
}

// ConfigTemplate 配置模板
type ConfigTemplate struct {
	Name        string                 `json:"name" yaml:"name"`               // 模板名称
	Description string                 `json:"description" yaml:"description"` // 模板描述
	Type        PluginType             `json:"type" yaml:"type"`               // 插件类型
	Version     string                 `json:"version" yaml:"version"`         // 模板版本
	Config      map[string]interface{} `json:"config" yaml:"config"`           // 配置内容
	Schema      interface{}            `json:"schema" yaml:"schema"`           // 配置模式
	CreatedAt   time.Time              `json:"created_at" yaml:"created_at"`   // 创建时间
	UpdatedAt   time.Time              `json:"updated_at" yaml:"updated_at"`   // 更新时间
}

// NewBasePluginConfig 创建基础插件配置
func NewBasePluginConfig(id, name, version string, pluginType PluginType) *BasePluginConfig {
	return &BasePluginConfig{
		ID:           id,
		Name:         name,
		Version:      version,
		Type:         pluginType,
		Enabled:      true,
		Priority:     PluginPriorityNormal,
		CustomConfig: make(map[string]interface{}),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// NewResourceLimits 创建资源限制配置
func NewResourceLimits() *ResourceLimits {
	return &ResourceLimits{
		MaxMemoryMB:     512,  // 默认512MB
		MaxCPUPercent:   50.0, // 默认50%
		MaxGoroutines:   100,  // 默认100个协程
		MaxFileHandles:  50,   // 默认50个文件句柄
		MaxNetworkConn:  10,   // 默认10个网络连接
		MaxFileDescriptors: 50, // 默认50个文件描述符
		MaxNetworkConnections: 10, // 默认10个网络连接
		MaxDiskUsageMB:  100,  // 默认100MB磁盘
		ExecutionTimeout: 30 * time.Second, // 默认30秒执行超时
		IdleTimeout:     5 * time.Minute,   // 默认5分钟空闲超时
		StartupTimeout:  10 * time.Second,  // 默认10秒启动超时
		ShutdownTimeout: 5 * time.Second,   // 默认5秒关闭超时
		Enabled:        true,
		EnforceMode:    EnforceModeWarn,
	}
}

// NewSecurityConfig 创建安全配置
func NewSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		SandboxEnabled:    true,
		Permissions:       []Permission{PermissionFileRead, PermissionNetworkAccess},
		AllowedHosts:      []string{},
		AllowedPorts:      []int{80, 443},
		AllowedPaths:      []string{},
		BlockedPaths:      []string{"/etc", "/sys", "/proc"},
		AllowedCommands:   []string{},
		BlockedCommands:   []string{"rm", "dd", "mkfs"},
		EncryptionEnabled: false,
		SignatureRequired: false,
		TrustedSources:    []string{},
		SecurityLevel:     SecurityLevelMedium,
		AuditEnabled:      true,
		CustomRules:       make(map[string]interface{}),
	}
}