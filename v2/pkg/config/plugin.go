package config

import (
	"time"
)

// PluginConfig 插件配置
type PluginConfig struct {
	// 基础配置
	Name      string `koanf:"name" json:"name" validate:"required"`
	Type      string `koanf:"type" json:"type" validate:"required,oneof=dynamic_library rpc webassembly hot_reload"`
	Path      string `koanf:"path" json:"path" validate:"required"`
	Enabled   bool   `koanf:"enabled" json:"enabled"`
	AutoStart bool   `koanf:"auto_start" json:"auto_start"`
	Priority  int    `koanf:"priority" json:"priority" validate:"min=0,max=100"`

	// 特定配置
	Config map[string]string `koanf:"config" json:"config"`

	// 资源限制
	Resources ResourceLimits `koanf:"resources" json:"resources"`

	// 安全配置
	Security PluginSecurityConfig `koanf:"security" json:"security"`
}

// ResourceLimits 资源限制
type ResourceLimits struct {
	MaxMemory     int64         `koanf:"max_memory" json:"max_memory" validate:"min=0"`           // 最大内存使用量（字节）
	MaxCPU        float64       `koanf:"max_cpu" json:"max_cpu" validate:"min=0,max=1"`           // 最大CPU使用率（0-1）
	MaxDiskIO     int64         `koanf:"max_disk_io" json:"max_disk_io" validate:"min=0"`         // 最大磁盘IO（字节/秒）
	MaxNetworkIO  int64         `koanf:"max_network_io" json:"max_network_io" validate:"min=0"`   // 最大网络IO（字节/秒）
	Timeout       time.Duration `koanf:"timeout" json:"timeout"`                                 // 执行超时时间
	MaxGoroutines int           `koanf:"max_goroutines" json:"max_goroutines" validate:"min=1"`   // 最大协程数
}

// PluginSecurityConfig 插件安全配置
type PluginSecurityConfig struct {
	Sandbox         bool     `koanf:"sandbox" json:"sandbox"`                     // 是否启用沙箱
	AllowedPaths    []string `koanf:"allowed_paths" json:"allowed_paths"`         // 允许访问的路径
	AllowedNetworks []string `koanf:"allowed_networks" json:"allowed_networks"`   // 允许访问的网络
	AllowedSyscalls []string `koanf:"allowed_syscalls" json:"allowed_syscalls"`   // 允许的系统调用
	TrustedSources  []string `koanf:"trusted_sources" json:"trusted_sources"`     // 可信的插件源
}

// Validate 验证插件配置
func (pc *PluginConfig) Validate() error {
	if pc.Name == "" {
		return ErrInvalidPluginName
	}
	if pc.Type == "" {
		return ErrInvalidPluginType
	}
	if pc.Path == "" {
		return ErrInvalidPluginPath
	}
	if pc.Priority < 0 || pc.Priority > 100 {
		return ErrInvalidPluginPriority
	}
	return pc.Resources.Validate()
}

// Validate 验证资源限制配置
func (rl *ResourceLimits) Validate() error {
	if rl.MaxMemory < 0 {
		return ErrInvalidMaxMemory
	}
	if rl.MaxCPU < 0 || rl.MaxCPU > 1 {
		return ErrInvalidMaxCPU
	}
	if rl.MaxDiskIO < 0 {
		return ErrInvalidMaxDiskIO
	}
	if rl.MaxNetworkIO < 0 {
		return ErrInvalidMaxNetworkIO
	}
	if rl.MaxGoroutines < 1 {
		return ErrInvalidMaxGoroutines
	}
	return nil
}

// Validate 验证安全配置
func (sc *PluginSecurityConfig) Validate() error {
	// 基本验证，可根据需要扩展
	return nil
}

// SetDefaults 设置默认值
func (pc *PluginConfig) SetDefaults() {
	if pc.Priority == 0 {
		pc.Priority = 50 // 默认优先级
	}
	if pc.Config == nil {
		pc.Config = make(map[string]string)
	}
	pc.Resources.SetDefaults()
	pc.Security.SetDefaults()
}

// SetDefaults 设置资源限制默认值
func (rl *ResourceLimits) SetDefaults() {
	if rl.MaxMemory == 0 {
		rl.MaxMemory = 100 * 1024 * 1024 // 100MB
	}
	if rl.MaxCPU == 0 {
		rl.MaxCPU = 0.5 // 50%
	}
	if rl.MaxDiskIO == 0 {
		rl.MaxDiskIO = 10 * 1024 * 1024 // 10MB/s
	}
	if rl.MaxNetworkIO == 0 {
		rl.MaxNetworkIO = 10 * 1024 * 1024 // 10MB/s
	}
	if rl.Timeout == 0 {
		rl.Timeout = 30 * time.Second
	}
	if rl.MaxGoroutines == 0 {
		rl.MaxGoroutines = 100
	}
}

// SetDefaults 设置安全配置默认值
func (sc *PluginSecurityConfig) SetDefaults() {
	if sc.AllowedPaths == nil {
		sc.AllowedPaths = []string{}
	}
	if sc.AllowedNetworks == nil {
		sc.AllowedNetworks = []string{}
	}
	if sc.AllowedSyscalls == nil {
		sc.AllowedSyscalls = []string{}
	}
	if sc.TrustedSources == nil {
		sc.TrustedSources = []string{}
	}
}