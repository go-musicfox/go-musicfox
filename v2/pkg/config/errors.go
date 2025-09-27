package config

import "errors"

// 配置相关错误定义
var (
	// 插件配置错误
	ErrInvalidPluginName     = errors.New("invalid plugin name")
	ErrInvalidPluginType     = errors.New("invalid plugin type")
	ErrInvalidPluginPath     = errors.New("invalid plugin path")
	ErrInvalidPluginPriority = errors.New("invalid plugin priority, must be between 0 and 100")

	// 资源限制错误
	ErrInvalidMaxMemory     = errors.New("invalid max memory, must be >= 0")
	ErrInvalidMaxCPU        = errors.New("invalid max CPU, must be between 0 and 1")
	ErrInvalidMaxDiskIO     = errors.New("invalid max disk IO, must be >= 0")
	ErrInvalidMaxNetworkIO  = errors.New("invalid max network IO, must be >= 0")
	ErrInvalidMaxGoroutines = errors.New("invalid max goroutines, must be >= 1")

	// 配置管理错误
	ErrConfigNotFound    = errors.New("configuration not found")
	ErrConfigLoadFailed  = errors.New("failed to load configuration")
	ErrConfigParseFailed = errors.New("failed to parse configuration")
	ErrConfigMergeFailed = errors.New("failed to merge configuration")
	ErrConfigSaveFailed  = errors.New("failed to save configuration")

	// 配置源错误
	ErrUnsupportedConfigSource = errors.New("unsupported configuration source")
	ErrConfigSourceNotFound    = errors.New("configuration source not found")
	ErrConfigSourceInvalid     = errors.New("invalid configuration source")
)