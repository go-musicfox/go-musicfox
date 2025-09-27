package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	koanfjson "github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	koanfyaml "github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"gopkg.in/yaml.v3"
)

// ConfigFormat 配置文件格式
type ConfigFormat string

const (
	ConfigFormatYAML ConfigFormat = "yaml"
	ConfigFormatJSON ConfigFormat = "json"
	ConfigFormatTOML ConfigFormat = "toml"
)

// PluginType 插件类型枚举
type PluginType string

const (
	PluginTypeGeneric        PluginType = "generic"         // 通用插件
	PluginTypeDynamicLibrary PluginType = "dynamic_library" // 动态链接库插件
	PluginTypeRPC           PluginType = "rpc"              // RPC插件
	PluginTypeWebAssembly   PluginType = "webassembly"     // WebAssembly插件
	PluginTypeHotReload     PluginType = "hot_reload"      // 热加载插件
	PluginTypeMusicSource   PluginType = "music_source"    // 音乐源插件
	PluginTypeAudioProcessor PluginType = "audio_processor" // 音频处理插件
	PluginTypeUI            PluginType = "ui"               // UI插件
	PluginTypeStorage       PluginType = "storage"          // 存储插件
	PluginTypeNetwork       PluginType = "network"          // 网络插件
)

// String 返回插件类型的字符串表示
func (pt PluginType) String() string {
	return string(pt)
}

// PluginPriority 插件优先级枚举
type PluginPriority int

const (
	PluginPriorityLow    PluginPriority = iota // 低优先级
	PluginPriorityNormal                       // 普通优先级
	PluginPriorityHigh                         // 高优先级
	PluginPriorityCritical                     // 关键优先级
)

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
	MinVersion     string                 `json:"min_version" yaml:"min_version"`         // 最小版本
	MaxVersion     string                 `json:"max_version" yaml:"max_version"`         // 最大版本
	PluginPath     string                 `json:"plugin_path" yaml:"plugin_path"`         // 插件路径
	LogPath        string                 `json:"log_path" yaml:"log_path"`               // 日志路径
	ResourceLimits *ResourceLimits        `json:"resource_limits" yaml:"resource_limits"` // 资源限制
	SecurityConfig *SecurityConfig        `json:"security_config" yaml:"security_config"` // 安全配置
	CustomConfig   map[string]interface{} `json:"custom_config" yaml:"custom_config"`     // 自定义配置
	CreatedAt      time.Time              `json:"created_at" yaml:"created_at"`           // 创建时间
	UpdatedAt      time.Time              `json:"updated_at" yaml:"updated_at"`           // 更新时间
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

// ConfigValidator 配置验证器接口
type ConfigValidator interface {
	Validate(ctx context.Context, config *EnhancedPluginConfig) error
}

// ConfigWatcher 配置监控器接口
type ConfigWatcher interface {
	// Start 开始监控
	Start(ctx context.Context) error
	// Stop 停止监控
	Stop()
	// Events 获取事件通道
	Events() <-chan *ConfigChangeEvent
}

// ResourceLimits 资源限制配置
type ResourceLimits struct {
	MaxMemoryMB    int `json:"max_memory_mb" yaml:"max_memory_mb"`
	MaxCPUPercent  int `json:"max_cpu_percent" yaml:"max_cpu_percent"`
	MaxDiskIOKBps  int `json:"max_disk_io_kbps" yaml:"max_disk_io_kbps"`
	MaxNetIOKBps   int `json:"max_net_io_kbps" yaml:"max_net_io_kbps"`
	MaxGoroutines  int `json:"max_goroutines" yaml:"max_goroutines"`
	MaxFileHandles int `json:"max_file_handles" yaml:"max_file_handles"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	EnableSeccomp     bool     `json:"enable_seccomp" yaml:"enable_seccomp"`
	EnableNamespace   bool     `json:"enable_namespace" yaml:"enable_namespace"`
	AllowedPaths      []string `json:"allowed_paths" yaml:"allowed_paths"`
	BlockedPaths      []string `json:"blocked_paths" yaml:"blocked_paths"`
	AllowedHosts      []string `json:"allowed_hosts" yaml:"allowed_hosts"`
	BlockedHosts      []string `json:"blocked_hosts" yaml:"blocked_hosts"`
	AllowedPorts      []int    `json:"allowed_ports" yaml:"allowed_ports"`
	BlockedPorts      []int    `json:"blocked_ports" yaml:"blocked_ports"`
	AllowedSyscalls   []string `json:"allowed_syscalls" yaml:"allowed_syscalls"`
	BlockedSyscalls   []string `json:"blocked_syscalls" yaml:"blocked_syscalls"`
	MaxConnections    int      `json:"max_connections" yaml:"max_connections"`
	ConnectionTimeout int      `json:"connection_timeout" yaml:"connection_timeout"`
}

// PluginConfigManager 插件配置管理器
type PluginConfigManager struct {
	logger       *slog.Logger
	configDir    string
	templateDir  string
	configs      map[string]*EnhancedPluginConfig
	templates    map[string]*ConfigTemplate
	validators   map[PluginType]ConfigValidator
	watchers     map[string]ConfigWatcher
	mu           sync.RWMutex
	defaultLimits *ResourceLimits
	defaultSecurity *SecurityConfig
}

// EnhancedPluginConfig 增强的插件配置
type EnhancedPluginConfig struct {
	*BasePluginConfig
	ConfigPath   string                 `json:"config_path" yaml:"config_path"`     // 配置文件路径
	ConfigFormat ConfigFormat          `json:"config_format" yaml:"config_format"` // 配置文件格式
	AutoReload   bool                   `json:"auto_reload" yaml:"auto_reload"`     // 是否自动重载
	BackupCount  int                    `json:"backup_count" yaml:"backup_count"`   // 备份数量
	Validation   *ValidationConfig      `json:"validation" yaml:"validation"`       // 验证配置
	Metadata     map[string]interface{} `json:"metadata" yaml:"metadata"`           // 元数据
	CustomConfig map[string]interface{} `json:"custom_config" yaml:"custom_config"` // 自定义配置
	Checksum     string                 `json:"checksum" yaml:"checksum"`           // 配置校验和
	LastLoaded   time.Time              `json:"last_loaded" yaml:"last_loaded"`     // 最后加载时间
}

// Validate 验证配置
func (c *EnhancedPluginConfig) Validate() error {
	if c.BasePluginConfig == nil {
		return fmt.Errorf("base plugin config is required")
	}
	return c.BasePluginConfig.Validate()
}

// ValidationConfig 验证配置
type ValidationConfig struct {
	Enabled        bool     `json:"enabled" yaml:"enabled"`               // 是否启用验证
	Strict         bool     `json:"strict" yaml:"strict"`                 // 严格模式
	RequiredFields []string `json:"required_fields" yaml:"required_fields"` // 必需字段
	CustomRules    []string `json:"custom_rules" yaml:"custom_rules"`     // 自定义规则
}

// ConfigChangeEvent 配置变更事件
type ConfigChangeEvent struct {
	PluginID   string                 `json:"plugin_id"`
	ChangeType ConfigChangeType       `json:"change_type"`
	OldConfig  *EnhancedPluginConfig  `json:"old_config,omitempty"`
	NewConfig  *EnhancedPluginConfig  `json:"new_config,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ConfigChangeType 配置变更类型
type ConfigChangeType string

const (
	ConfigChangeTypeCreate ConfigChangeType = "create"
	ConfigChangeTypeUpdate ConfigChangeType = "update"
	ConfigChangeTypeDelete ConfigChangeType = "delete"
	ConfigChangeTypeReload ConfigChangeType = "reload"
)

// NewPluginConfigManager 创建插件配置管理器
func NewPluginConfigManager(logger *slog.Logger) *PluginConfigManager {
	return &PluginConfigManager{
		logger:          logger,
		configDir:       "./configs",
		templateDir:     "./templates",
		configs:         make(map[string]*EnhancedPluginConfig),
		templates:       make(map[string]*ConfigTemplate),
		validators:      make(map[PluginType]ConfigValidator),
		watchers:        make(map[string]ConfigWatcher),
		defaultLimits:   &ResourceLimits{MaxMemoryMB: 100, MaxCPUPercent: 50},
		defaultSecurity: &SecurityConfig{EnableSeccomp: false, EnableNamespace: false},
	}
}

// Initialize 初始化配置管理器
func (cm *PluginConfigManager) Initialize(ctx context.Context) error {
	// 创建配置目录
	if err := os.MkdirAll(cm.configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 创建模板目录
	if err := os.MkdirAll(cm.templateDir, 0755); err != nil {
		return fmt.Errorf("failed to create template directory: %w", err)
	}

	// 加载现有配置
	if err := cm.loadAllConfigs(ctx); err != nil {
		return fmt.Errorf("failed to load configs: %w", err)
	}

	// 加载配置模板
	if err := cm.loadAllTemplates(ctx); err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	cm.logger.Info("Plugin config manager initialized",
		"config_dir", cm.configDir,
		"template_dir", cm.templateDir,
		"configs_loaded", len(cm.configs),
		"templates_loaded", len(cm.templates))

	return nil
}

// LoadConfig 加载插件配置
func (cm *PluginConfigManager) LoadConfig(ctx context.Context, pluginID string) (*EnhancedPluginConfig, error) {
	cm.mu.RLock()
	if config, exists := cm.configs[pluginID]; exists {
		cm.mu.RUnlock()
		return config, nil
	}
	cm.mu.RUnlock()

	// 尝试从文件加载
	config, err := cm.loadConfigFromFile(ctx, pluginID)
	if err != nil {
		return nil, fmt.Errorf("failed to load config for plugin %s: %w", pluginID, err)
	}

	cm.mu.Lock()
	cm.configs[pluginID] = config
	cm.mu.Unlock()

	return config, nil
}

// SaveConfig 保存插件配置
func (cm *PluginConfigManager) SaveConfig(ctx context.Context, config *EnhancedPluginConfig) error {
	if err := cm.validateConfig(ctx, config); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// 更新时间戳和校验和
	config.UpdatedAt = time.Now()
	config.Checksum = cm.calculateChecksum(config)

	// 保存到文件
	if err := cm.saveConfigToFile(ctx, config); err != nil {
		return fmt.Errorf("failed to save config to file: %w", err)
	}

	cm.mu.Lock()
	cm.configs[config.ID] = config
	cm.mu.Unlock()

	cm.logger.Info("Plugin config saved", "plugin_id", config.ID)
	return nil
}

// DeleteConfig 删除插件配置
func (cm *PluginConfigManager) DeleteConfig(ctx context.Context, pluginID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	config, exists := cm.configs[pluginID]
	if !exists {
		return fmt.Errorf("config not found for plugin %s", pluginID)
	}

	// 删除配置文件
	if config.ConfigPath != "" {
		if err := os.Remove(config.ConfigPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete config file: %w", err)
		}
	}

	// 停止监控
	if watcher, exists := cm.watchers[pluginID]; exists {
		watcher.Stop()
		delete(cm.watchers, pluginID)
	}

	delete(cm.configs, pluginID)

	cm.logger.Info("Plugin config deleted", "plugin_id", pluginID)
	return nil
}

// ListConfigs 列出所有配置
func (cm *PluginConfigManager) ListConfigs(ctx context.Context) ([]*EnhancedPluginConfig, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	configs := make([]*EnhancedPluginConfig, 0, len(cm.configs))
	for _, config := range cm.configs {
		configs = append(configs, config)
	}

	return configs, nil
}

// GetTemplate 获取配置模板
func (cm *PluginConfigManager) GetTemplate(ctx context.Context, pluginType PluginType) (*ConfigTemplate, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	template, exists := cm.templates[string(pluginType)]
	if !exists {
		return nil, fmt.Errorf("template not found for plugin type %s", pluginType)
	}

	return template, nil
}

// CreateDefaultConfig 创建默认配置
func (cm *PluginConfigManager) CreateDefaultConfig(pluginID, name, version string) *EnhancedPluginConfig {
	return &EnhancedPluginConfig{
		BasePluginConfig: NewBasePluginConfig(pluginID, name, version, PluginTypeGeneric),
		ConfigFormat:     ConfigFormatYAML,
		AutoReload:       true,
		BackupCount:      5,
		Validation:       &ValidationConfig{Enabled: true},
		Metadata:         make(map[string]interface{}),
		CustomConfig:     make(map[string]interface{}),
		LastLoaded:       time.Now(),
	}
}

// CreateFromTemplate 从模板创建配置
func (cm *PluginConfigManager) CreateFromTemplate(ctx context.Context, pluginID string, pluginType PluginType, overrides map[string]interface{}) (*EnhancedPluginConfig, error) {
	template, err := cm.GetTemplate(ctx, pluginType)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// 创建基础配置
	config := &EnhancedPluginConfig{
		BasePluginConfig: NewBasePluginConfig(pluginID, pluginID, "1.0.0", pluginType),
		ConfigFormat:     ConfigFormatYAML,
		AutoReload:       true,
		BackupCount:      5,
		Validation:       &ValidationConfig{Enabled: true},
		Metadata:         make(map[string]interface{}),
		CustomConfig:     make(map[string]interface{}),
		LastLoaded:       time.Now(),
	}

	// 应用模板配置
	for key, value := range template.Config {
		config.CustomConfig[key] = value
	}

	// 应用覆盖配置
	for key, value := range overrides {
		config.CustomConfig[key] = value
	}

	// 设置默认资源限制和安全配置
	if config.ResourceLimits == nil {
		config.ResourceLimits = cm.defaultLimits
	}
	if config.SecurityConfig == nil {
		config.SecurityConfig = cm.defaultSecurity
	}

	return config, nil
}

// RegisterValidator 注册配置验证器
func (cm *PluginConfigManager) RegisterValidator(pluginType PluginType, validator ConfigValidator) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.validators[pluginType] = validator
}

// WatchConfig 监控配置变更
func (cm *PluginConfigManager) WatchConfig(ctx context.Context, pluginID string) (<-chan *ConfigChangeEvent, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	config, exists := cm.configs[pluginID]
	if !exists {
		return nil, fmt.Errorf("config not found for plugin %s", pluginID)
	}

	if !config.AutoReload {
		return nil, fmt.Errorf("auto reload is disabled for plugin %s", pluginID)
	}

	watcher, exists := cm.watchers[pluginID]
	if exists {
		return watcher.Events(), nil
	}

	watcher = NewConfigWatcher(cm.logger, config.ConfigPath, pluginID)
	if err := watcher.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start config watcher: %w", err)
	}

	cm.watchers[pluginID] = watcher
	return watcher.Events(), nil
}

// loadAllConfigs 加载所有配置文件
func (cm *PluginConfigManager) loadAllConfigs(ctx context.Context) error {
	return filepath.WalkDir(cm.configDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// 检查文件扩展名
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" && ext != ".json" && ext != ".toml" {
			return nil
		}

		// 提取插件ID
		pluginID := strings.TrimSuffix(d.Name(), ext)

		// 加载配置
		config, err := cm.loadConfigFromPath(ctx, path, pluginID)
		if err != nil {
			cm.logger.Warn("Failed to load config", "path", path, "error", err)
			return nil // 继续处理其他文件
		}

		cm.configs[pluginID] = config
		return nil
	})
}

// loadAllTemplates 加载所有配置模板
func (cm *PluginConfigManager) loadAllTemplates(ctx context.Context) error {
	return filepath.WalkDir(cm.templateDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// 检查文件扩展名
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" && ext != ".json" && ext != ".toml" {
			return nil
		}

		// 加载模板
		template, err := cm.loadTemplateFromPath(ctx, path)
		if err != nil {
			cm.logger.Warn("Failed to load template", "path", path, "error", err)
			return nil // 继续处理其他文件
		}

		cm.templates[template.Name] = template
		return nil
	})
}

// loadConfigFromFile 从文件加载配置
func (cm *PluginConfigManager) loadConfigFromFile(ctx context.Context, pluginID string) (*EnhancedPluginConfig, error) {
	// 尝试不同的文件格式
	formats := []string{".yaml", ".yml", ".json", ".toml"}
	for _, ext := range formats {
		path := filepath.Join(cm.configDir, pluginID+ext)
		if _, err := os.Stat(path); err == nil {
			return cm.loadConfigFromPath(ctx, path, pluginID)
		}
	}

	return nil, fmt.Errorf("config file not found for plugin %s", pluginID)
}

// loadConfigFromPath 从指定路径加载配置
func (cm *PluginConfigManager) loadConfigFromPath(ctx context.Context, path, pluginID string) (*EnhancedPluginConfig, error) {
	k := koanf.New(".")

	// 确定配置格式
	format := cm.detectConfigFormat(path)
	var parser koanf.Parser
	switch format {
	case ConfigFormatYAML:
		parser = koanfyaml.Parser()
	case ConfigFormatJSON:
		parser = koanfjson.Parser()
	case ConfigFormatTOML:
		parser = toml.Parser()
	default:
		return nil, fmt.Errorf("unsupported config format: %s", format)
	}

	// 加载配置文件
	if err := k.Load(file.Provider(path), parser); err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	// 解析配置
	config := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{},
		ConfigPath:       path,
		ConfigFormat:     format,
		LastLoaded:       time.Now(),
	}

	if err := k.Unmarshal("", config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 设置插件ID（如果未设置）
	if config.ID == "" {
		config.ID = pluginID
	}

	// 计算校验和
	config.Checksum = cm.calculateChecksum(config)

	return config, nil
}

// loadTemplateFromPath 从指定路径加载模板
func (cm *PluginConfigManager) loadTemplateFromPath(ctx context.Context, path string) (*ConfigTemplate, error) {
	k := koanf.New(".")

	// 确定配置格式
	format := cm.detectConfigFormat(path)
	var parser koanf.Parser
	switch format {
	case ConfigFormatYAML:
		parser = koanfyaml.Parser()
	case ConfigFormatJSON:
		parser = koanfjson.Parser()
	case ConfigFormatTOML:
		parser = toml.Parser()
	default:
		return nil, fmt.Errorf("unsupported config format: %s", format)
	}

	// 加载模板文件
	if err := k.Load(file.Provider(path), parser); err != nil {
		return nil, fmt.Errorf("failed to load template file: %w", err)
	}

	// 解析模板
	template := &ConfigTemplate{}
	if err := k.Unmarshal("", template); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template: %w", err)
	}

	return template, nil
}

// saveConfigToFile 保存配置到文件
func (cm *PluginConfigManager) saveConfigToFile(ctx context.Context, config *EnhancedPluginConfig) error {
	// 确定保存路径
	if config.ConfigPath == "" {
		ext := ".yaml"
		switch config.ConfigFormat {
		case ConfigFormatJSON:
			ext = ".json"
		case ConfigFormatTOML:
			ext = ".toml"
		}
		config.ConfigPath = filepath.Join(cm.configDir, config.ID+ext)
	}

	// 创建备份
	if err := cm.createBackup(config.ConfigPath); err != nil {
		cm.logger.Warn("Failed to create backup", "path", config.ConfigPath, "error", err)
	}

	// 序列化配置
	var data []byte
	var err error
	switch config.ConfigFormat {
	case ConfigFormatYAML:
		data, err = yaml.Marshal(config)
	case ConfigFormatJSON:
		data, err = json.MarshalIndent(config, "", "  ")
	case ConfigFormatTOML:
		// TOML序列化需要使用第三方库
		return fmt.Errorf("TOML serialization not implemented")
	default:
		return fmt.Errorf("unsupported config format: %s", config.ConfigFormat)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(config.ConfigPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// validateConfig 验证配置
func (cm *PluginConfigManager) validateConfig(ctx context.Context, config *EnhancedPluginConfig) error {
	// 基础验证
	if err := config.Validate(); err != nil {
		return err
	}

	// 使用注册的验证器
	if validator, exists := cm.validators[config.Type]; exists {
		if err := validator.Validate(ctx, config); err != nil {
			return err
		}
	}

	// 验证配置
	if config.Validation != nil && config.Validation.Enabled {
		if err := cm.validateCustomRules(ctx, config); err != nil {
			return err
		}
	}

	return nil
}

// validateCustomRules 验证自定义规则
func (cm *PluginConfigManager) validateCustomRules(ctx context.Context, config *EnhancedPluginConfig) error {
	// 检查必需字段
	for _, field := range config.Validation.RequiredFields {
		if _, exists := config.CustomConfig[field]; !exists {
			return fmt.Errorf("required field missing: %s", field)
		}
	}

	// TODO: 实现更多自定义验证规则

	return nil
}

// detectConfigFormat 检测配置文件格式
func (cm *PluginConfigManager) detectConfigFormat(path string) ConfigFormat {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return ConfigFormatYAML
	case ".json":
		return ConfigFormatJSON
	case ".toml":
		return ConfigFormatTOML
	default:
		return ConfigFormatYAML // 默认格式
	}
}

// calculateChecksum 计算配置校验和
func (cm *PluginConfigManager) calculateChecksum(config *EnhancedPluginConfig) string {
	// 简单的校验和实现，实际应用中可以使用更复杂的算法
	data, _ := json.Marshal(config.CustomConfig)
	return fmt.Sprintf("%x", len(data))
}

// createBackup 创建配置备份
func (cm *PluginConfigManager) createBackup(configPath string) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil // 文件不存在，无需备份
	}

	backupPath := configPath + ".backup." + time.Now().Format("20060102150405")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	return os.WriteFile(backupPath, data, 0644)
}

// Shutdown 关闭配置管理器
func (cm *PluginConfigManager) Shutdown(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 停止所有监控器
	for _, watcher := range cm.watchers {
		watcher.Stop()
	}

	cm.logger.Info("Plugin config manager shutdown")
	return nil
}

// DefaultResourceLimits 返回默认资源限制
func DefaultResourceLimits() *ResourceLimits {
	return &ResourceLimits{
		MaxMemoryMB:    100,
		MaxCPUPercent:  50,
		MaxDiskIOKBps:  1000,
		MaxNetIOKBps:   1000,
		MaxGoroutines:  100,
		MaxFileHandles: 50,
	}
}

// DefaultSecurityConfig 返回默认安全配置
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		EnableSeccomp:    true,
		EnableNamespace:  false,
		AllowedPaths:     []string{"./data", "./temp"},
		BlockedPaths:     []string{"/etc", "/sys", "/proc"},
		AllowedHosts:     []string{"localhost", "127.0.0.1"},
		BlockedHosts:     []string{},
		AllowedPorts:     []int{80, 443, 8080},
		BlockedPorts:     []int{22, 23, 25},
		AllowedSyscalls:  []string{"read", "write", "open", "close"},
		BlockedSyscalls:  []string{"execve", "fork", "clone"},
		MaxConnections:   10,
		ConnectionTimeout: 30,
	}
}

// serializeConfig 序列化配置
func (cm *PluginConfigManager) serializeConfig(config *EnhancedPluginConfig, format ConfigFormat) ([]byte, error) {
	switch format {
	case ConfigFormatYAML:
		return yaml.Marshal(config)
	case ConfigFormatJSON:
		return json.MarshalIndent(config, "", "  ")
	case ConfigFormatTOML:
		return nil, fmt.Errorf("TOML serialization not implemented")
	default:
		return nil, fmt.Errorf("unsupported config format: %s", format)
	}
}

// deserializeConfig 反序列化配置
func (cm *PluginConfigManager) deserializeConfig(data []byte, format ConfigFormat) (*EnhancedPluginConfig, error) {
	config := &EnhancedPluginConfig{}
	switch format {
	case ConfigFormatYAML:
		err := yaml.Unmarshal(data, config)
		return config, err
	case ConfigFormatJSON:
		err := json.Unmarshal(data, config)
		return config, err
	case ConfigFormatTOML:
		return nil, fmt.Errorf("TOML deserialization not implemented")
	default:
		return nil, fmt.Errorf("unsupported config format: %s", format)
	}
}

// MergeConfigs 合并配置
func (cm *PluginConfigManager) MergeConfigs(base *EnhancedPluginConfig, override map[string]interface{}) *EnhancedPluginConfig {
	merged := &EnhancedPluginConfig{
		BasePluginConfig: base.BasePluginConfig,
		ConfigPath:       base.ConfigPath,
		ConfigFormat:     base.ConfigFormat,
		AutoReload:       base.AutoReload,
		CustomConfig:     make(map[string]interface{}),
		Validation:       base.Validation,
		LastLoaded:       base.LastLoaded,
		Checksum:         base.Checksum,
	}

	// 复制基础配置
	for k, v := range base.CustomConfig {
		merged.CustomConfig[k] = v
	}

	// 应用覆盖配置
	for k, v := range override {
		merged.CustomConfig[k] = v
	}

	return merged
}

// ProcessEnvironmentVariables 处理环境变量
func (cm *PluginConfigManager) ProcessEnvironmentVariables(config *EnhancedPluginConfig) *EnhancedPluginConfig {
	processed := &EnhancedPluginConfig{
		BasePluginConfig: config.BasePluginConfig,
		ConfigPath:       config.ConfigPath,
		ConfigFormat:     config.ConfigFormat,
		AutoReload:       config.AutoReload,
		CustomConfig:     make(map[string]interface{}),
		Validation:       config.Validation,
		LastLoaded:       config.LastLoaded,
		Checksum:         config.Checksum,
	}

	// 处理自定义配置中的环境变量
	for k, v := range config.CustomConfig {
		if str, ok := v.(string); ok {
			processed.CustomConfig[k] = cm.expandEnvironmentVariable(str)
		} else {
			processed.CustomConfig[k] = v
		}
	}

	return processed
}

// expandEnvironmentVariable 展开环境变量
func (cm *PluginConfigManager) expandEnvironmentVariable(value string) string {
	// 简单的环境变量替换实现
	// 支持 ${VAR} 和 ${VAR:default} 格式
	if len(value) < 3 || !strings.HasPrefix(value, "${") || !strings.HasSuffix(value, "}") {
		return value
	}

	varExpr := value[2 : len(value)-1] // 去掉 ${ 和 }
	parts := strings.SplitN(varExpr, ":", 2)
	varName := parts[0]

	envValue := os.Getenv(varName)
	if envValue != "" {
		return envValue
	}

	// 如果环境变量不存在且有默认值，返回默认值
	if len(parts) > 1 {
		return parts[1]
	}

	// 否则返回原值
	return value
}

// NewBasePluginConfig 创建基础插件配置
func NewBasePluginConfig(id, name, version string, pluginType PluginType) *BasePluginConfig {
	return &BasePluginConfig{
		ID:              id,
		Name:            name,
		Version:         version,
		Type:            pluginType,
		Enabled:         true,
		Priority:        PluginPriorityNormal,
		Tags:            []string{},
		Dependencies:    []string{},
		Conflicts:       []string{},
		CustomConfig:    make(map[string]interface{}),
		ResourceLimits:  DefaultResourceLimits(),
		SecurityConfig:  DefaultSecurityConfig(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// Validate 验证基础插件配置
func (c *BasePluginConfig) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("plugin ID is required")
	}
	if c.Name == "" {
		return fmt.Errorf("plugin name is required")
	}
	if c.Version == "" {
		return fmt.Errorf("plugin version is required")
	}
	// 简单的版本格式验证
	if !regexp.MustCompile(`^\d+\.\d+\.\d+`).MatchString(c.Version) {
		return fmt.Errorf("invalid version format: %s (expected format: x.y.z)", c.Version)
	}
	return nil
}

// TODO: NewConfigWatcher函数已在config_watcher.go中定义，此处删除重复声明
// 简单配置监控器实现已移至config_watcher.go文件