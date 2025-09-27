package plugin

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	// "reflect" // 暂时未使用
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"log/slog"
)

// ValidationLevel 验证级别
type ValidationLevel string

const (
	ValidationLevelBasic    ValidationLevel = "basic"    // 基础验证
	ValidationLevelStandard ValidationLevel = "standard" // 标准验证
	ValidationLevelStrict   ValidationLevel = "strict"   // 严格验证
	ValidationLevelCustom   ValidationLevel = "custom"   // 自定义验证
)

// ValidationSeverity 验证严重性
type ValidationSeverity string

const (
	ValidationSeverityInfo     ValidationSeverity = "info"
	ValidationSeverityWarning  ValidationSeverity = "warning"
	ValidationSeverityError    ValidationSeverity = "error"
	ValidationSeverityCritical ValidationSeverity = "critical"
)

// ValidationError 验证错误
type ValidationError struct {
	Field       string              `json:"field"`
	Value       interface{}         `json:"value,omitempty"`
	Message     string              `json:"message"`
	Severity    ValidationSeverity  `json:"severity"`
	Code        string              `json:"code"`
	Suggestion  string              `json:"suggestion,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// Error 实现error接口
func (ve *ValidationError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", ve.Severity, ve.Field, ve.Message)
}

// ValidationResult 验证结果
type ValidationResult struct {
	Valid      bool               `json:"valid"`
	Errors     []ValidationError  `json:"errors,omitempty"`
	Warnings   []ValidationError  `json:"warnings,omitempty"`
	Suggestions []ValidationError `json:"suggestions,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// HasErrors 检查是否有错误
func (vr *ValidationResult) HasErrors() bool {
	return len(vr.Errors) > 0
}

// HasWarnings 检查是否有警告
func (vr *ValidationResult) HasWarnings() bool {
	return len(vr.Warnings) > 0
}

// AddError 添加错误
func (vr *ValidationResult) AddError(field, message, code string, value interface{}) {
	vr.Errors = append(vr.Errors, ValidationError{
		Field:    field,
		Value:    value,
		Message:  message,
		Severity: ValidationSeverityError,
		Code:     code,
	})
	vr.Valid = false
}

// AddWarning 添加警告
func (vr *ValidationResult) AddWarning(field, message, code string, value interface{}) {
	vr.Warnings = append(vr.Warnings, ValidationError{
		Field:    field,
		Value:    value,
		Message:  message,
		Severity: ValidationSeverityWarning,
		Code:     code,
	})
}

// AddSuggestion 添加建议
func (vr *ValidationResult) AddSuggestion(field, message, suggestion string) {
	vr.Suggestions = append(vr.Suggestions, ValidationError{
		Field:      field,
		Message:    message,
		Severity:   ValidationSeverityInfo,
		Suggestion: suggestion,
	})
}

// Merge 合并验证结果
func (vr *ValidationResult) Merge(other *ValidationResult) {
	vr.Errors = append(vr.Errors, other.Errors...)
	vr.Warnings = append(vr.Warnings, other.Warnings...)
	vr.Suggestions = append(vr.Suggestions, other.Suggestions...)

	if other.HasErrors() {
		vr.Valid = false
	}

	if vr.Metadata == nil {
		vr.Metadata = make(map[string]interface{})
	}
	for k, v := range other.Metadata {
		vr.Metadata[k] = v
	}
}

// ValidationRule 验证规则接口
type ValidationRule interface {
	Validate(ctx context.Context, config *EnhancedPluginConfig) *ValidationResult
	GetName() string
	GetDescription() string
	GetSeverity() ValidationSeverity
}

// BaseValidationRule 基础验证规则
type BaseValidationRule struct {
	Name        string
	Description string
	Severity    ValidationSeverity
}

// GetName 获取规则名称
func (bvr *BaseValidationRule) GetName() string {
	return bvr.Name
}

// GetDescription 获取规则描述
func (bvr *BaseValidationRule) GetDescription() string {
	return bvr.Description
}

// GetSeverity 获取规则严重性
func (bvr *BaseValidationRule) GetSeverity() ValidationSeverity {
	return bvr.Severity
}

// RequiredFieldsRule 必填字段验证规则
type RequiredFieldsRule struct {
	BaseValidationRule
}

// NewRequiredFieldsRule 创建必填字段验证规则
func NewRequiredFieldsRule() *RequiredFieldsRule {
	return &RequiredFieldsRule{
		BaseValidationRule: BaseValidationRule{
			Name:        "required_fields",
			Description: "Validate required fields are present and not empty",
			Severity:    ValidationSeverityError,
		},
	}
}

// Validate 验证必填字段
func (rfr *RequiredFieldsRule) Validate(ctx context.Context, config *EnhancedPluginConfig) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// 验证基础字段
	if config.ID == "" {
		result.AddError("id", "Plugin ID is required", "REQUIRED_FIELD_MISSING", config.ID)
	}

	if config.Name == "" {
		result.AddError("name", "Plugin name is required", "REQUIRED_FIELD_MISSING", config.Name)
	}

	if config.Version == "" {
		result.AddError("version", "Plugin version is required", "REQUIRED_FIELD_MISSING", config.Version)
	}

	if config.PluginPath == "" {
		result.AddError("plugin_path", "Plugin path is required", "REQUIRED_FIELD_MISSING", config.PluginPath)
	}

	return result
}

// FormatValidationRule 格式验证规则
type FormatValidationRule struct {
	BaseValidationRule
}

// NewFormatValidationRule 创建格式验证规则
func NewFormatValidationRule() *FormatValidationRule {
	return &FormatValidationRule{
		BaseValidationRule: BaseValidationRule{
			Name:        "format_validation",
			Description: "Validate field formats and patterns",
			Severity:    ValidationSeverityError,
		},
	}
}

// Validate 验证字段格式
func (fvr *FormatValidationRule) Validate(ctx context.Context, config *EnhancedPluginConfig) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// 验证ID格式
	if config.ID != "" {
		if !fvr.isValidID(config.ID) {
			result.AddError("id", "Plugin ID must contain only alphanumeric characters, hyphens, and underscores", "INVALID_FORMAT", config.ID)
		}
	}

	// 验证版本格式
	if config.Version != "" {
		if !fvr.isValidVersion(config.Version) {
			result.AddError("version", "Plugin version must follow semantic versioning (e.g., 1.0.0)", "INVALID_FORMAT", config.Version)
		}
	}

	// 验证路径格式
	if config.PluginPath != "" {
		if !filepath.IsAbs(config.PluginPath) {
			result.AddError("plugin_path", "Plugin path must be absolute", "INVALID_FORMAT", config.PluginPath)
		}
	}

	// 验证配置路径格式
	if config.ConfigPath != "" {
		if !filepath.IsAbs(config.ConfigPath) {
			result.AddError("config_path", "Config path must be absolute", "INVALID_FORMAT", config.ConfigPath)
		}
	}

	return result
}

// isValidID 检查ID格式是否有效
func (fvr *FormatValidationRule) isValidID(id string) bool {
	pattern := `^[a-zA-Z0-9_-]+$`
	matched, _ := regexp.MatchString(pattern, id)
	return matched && len(id) >= 2 && len(id) <= 64
}

// isValidVersion 检查版本格式是否有效
func (fvr *FormatValidationRule) isValidVersion(version string) bool {
	// 简单的语义版本验证
	pattern := `^\d+\.\d+\.\d+(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$`
	matched, _ := regexp.MatchString(pattern, version)
	return matched
}

// PathValidationRule 路径验证规则
type PathValidationRule struct {
	BaseValidationRule
}

// NewPathValidationRule 创建路径验证规则
func NewPathValidationRule() *PathValidationRule {
	return &PathValidationRule{
		BaseValidationRule: BaseValidationRule{
			Name:        "path_validation",
			Description: "Validate file and directory paths",
			Severity:    ValidationSeverityError,
		},
	}
}

// Validate 验证路径
func (pvr *PathValidationRule) Validate(ctx context.Context, config *EnhancedPluginConfig) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// 验证插件路径
	if config.PluginPath != "" {
		if _, err := os.Stat(config.PluginPath); os.IsNotExist(err) {
			result.AddError("plugin_path", "Plugin file does not exist", "PATH_NOT_FOUND", config.PluginPath)
		} else if err != nil {
			result.AddError("plugin_path", fmt.Sprintf("Cannot access plugin file: %v", err), "PATH_ACCESS_ERROR", config.PluginPath)
		}
	}

	// 验证配置路径
	if config.ConfigPath != "" {
		configDir := filepath.Dir(config.ConfigPath)
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			result.AddWarning("config_path", "Config directory does not exist, will be created", "PATH_WILL_CREATE", configDir)
		}
	}

	// 验证日志路径
	if config.LogPath != "" {
		logDir := filepath.Dir(config.LogPath)
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			result.AddWarning("log_path", "Log directory does not exist, will be created", "PATH_WILL_CREATE", logDir)
		}
	}

	return result
}

// DependencyValidationRule 依赖验证规则
type DependencyValidationRule struct {
	BaseValidationRule
	availablePlugins map[string]*EnhancedPluginConfig
	mu               sync.RWMutex
}

// NewDependencyValidationRule 创建依赖验证规则
func NewDependencyValidationRule() *DependencyValidationRule {
	return &DependencyValidationRule{
		BaseValidationRule: BaseValidationRule{
			Name:        "dependency_validation",
			Description: "Validate plugin dependencies",
			Severity:    ValidationSeverityError,
		},
		availablePlugins: make(map[string]*EnhancedPluginConfig),
	}
}

// SetAvailablePlugins 设置可用插件列表
func (dvr *DependencyValidationRule) SetAvailablePlugins(plugins map[string]*EnhancedPluginConfig) {
	dvr.mu.Lock()
	defer dvr.mu.Unlock()
	dvr.availablePlugins = plugins
}

// Validate 验证依赖
func (dvr *DependencyValidationRule) Validate(ctx context.Context, config *EnhancedPluginConfig) *ValidationResult {
	result := &ValidationResult{Valid: true}

	dvr.mu.RLock()
	defer dvr.mu.RUnlock()

	// 检查循环依赖
	if dvr.hasCyclicDependency(config.ID, config.Dependencies, make(map[string]bool)) {
		result.AddError("dependencies", "Cyclic dependency detected", "CYCLIC_DEPENDENCY", config.Dependencies)
	}

	// 检查依赖是否存在
	for _, dep := range config.Dependencies {
		if _, exists := dvr.availablePlugins[dep]; !exists {
			result.AddError("dependencies", fmt.Sprintf("Dependency '%s' not found", dep), "DEPENDENCY_NOT_FOUND", dep)
		}
	}

	// 检查版本兼容性
	for _, dep := range config.Dependencies {
		if depConfig, exists := dvr.availablePlugins[dep]; exists {
			if !dvr.isVersionCompatible(config.MinVersion, depConfig.Version) {
				result.AddWarning("dependencies", fmt.Sprintf("Dependency '%s' version may not be compatible", dep), "VERSION_COMPATIBILITY", dep)
			}
		}
	}

	return result
}

// hasCyclicDependency 检查循环依赖
func (dvr *DependencyValidationRule) hasCyclicDependency(pluginID string, dependencies []string, visited map[string]bool) bool {
	if visited[pluginID] {
		return true
	}

	visited[pluginID] = true

	for _, dep := range dependencies {
		if depConfig, exists := dvr.availablePlugins[dep]; exists {
			if dvr.hasCyclicDependency(dep, depConfig.Dependencies, visited) {
				return true
			}
		}
	}

	delete(visited, pluginID)
	return false
}

// isVersionCompatible 检查版本兼容性
func (dvr *DependencyValidationRule) isVersionCompatible(minVersion, actualVersion string) bool {
	// 简单的版本比较实现
	if minVersion == "" {
		return true
	}

	minParts := strings.Split(minVersion, ".")
	actualParts := strings.Split(actualVersion, ".")

	for i := 0; i < len(minParts) && i < len(actualParts); i++ {
		minNum, err1 := strconv.Atoi(minParts[i])
		actualNum, err2 := strconv.Atoi(actualParts[i])

		if err1 != nil || err2 != nil {
			return true // 无法比较，假设兼容
		}

		if actualNum < minNum {
			return false
		} else if actualNum > minNum {
			return true
		}
	}

	return true
}

// ResourceValidationRule 资源验证规则
type ResourceValidationRule struct {
	BaseValidationRule
}

// NewResourceValidationRule 创建资源验证规则
func NewResourceValidationRule() *ResourceValidationRule {
	return &ResourceValidationRule{
		BaseValidationRule: BaseValidationRule{
			Name:        "resource_validation",
			Description: "Validate resource limits and configurations",
			Severity:    ValidationSeverityWarning,
		},
	}
}

// Validate 验证资源配置
func (rvr *ResourceValidationRule) Validate(ctx context.Context, config *EnhancedPluginConfig) *ValidationResult {
	result := &ValidationResult{Valid: true}

	if config.ResourceLimits == nil {
		result.AddSuggestion("resource_limits", "No resource limits configured", "Consider setting resource limits for better system stability")
		return result
	}

	// TODO: 验证资源限制配置
	// 由于ResourceLimits结构已简化，需要重新设计验证逻辑
	// if err := config.ResourceLimits.Validate(); err != nil {
	//     result.AddError("resource_limits", fmt.Sprintf("Invalid resource limits: %v", err), "INVALID_RESOURCE_LIMITS", config.ResourceLimits)
	// }

	// 检查内存限制是否合理
	if config.ResourceLimits.MaxMemoryMB > 1024 { // 1GB
		result.AddWarning("resource_limits.memory", "Memory limit is very high (>1GB)", "HIGH_MEMORY_LIMIT", config.ResourceLimits.MaxMemoryMB)
	}
	if config.ResourceLimits.MaxMemoryMB < 1 { // 1MB
		result.AddWarning("resource_limits.memory", "Memory limit is very low (<1MB)", "LOW_MEMORY_LIMIT", config.ResourceLimits.MaxMemoryMB)
	}

	// TODO: 检查CPU限制是否合理
	// 由于ResourceLimits结构已简化，需要重新设计CPU限制检查逻辑
	// if config.ResourceLimits.CPU != nil && config.ResourceLimits.CPU.Enabled {
	//     if config.ResourceLimits.CPU.HardLimit > 100 {
	//         result.AddError("resource_limits.cpu", "CPU limit cannot exceed 100%", "INVALID_CPU_LIMIT", config.ResourceLimits.CPU.HardLimit)
	//     }
	// }
	if config.ResourceLimits.MaxCPUPercent > 100 {
		result.AddError("resource_limits.cpu", "CPU limit cannot exceed 100%", "INVALID_CPU_LIMIT", config.ResourceLimits.MaxCPUPercent)
	}

	return result
}

// SecurityValidationRule 安全验证规则
type SecurityValidationRule struct {
	BaseValidationRule
}

// NewSecurityValidationRule 创建安全验证规则
func NewSecurityValidationRule() *SecurityValidationRule {
	return &SecurityValidationRule{
		BaseValidationRule: BaseValidationRule{
			Name:        "security_validation",
			Description: "Validate security configurations",
			Severity:    ValidationSeverityWarning,
		},
	}
}

// Validate 验证安全配置
func (svr *SecurityValidationRule) Validate(ctx context.Context, config *EnhancedPluginConfig) *ValidationResult {
	result := &ValidationResult{Valid: true}

	if config.SecurityConfig == nil {
		result.AddWarning("security_config", "No security configuration found", "NO_SECURITY_CONFIG", nil)
		return result
	}

	// 检查危险配置
	dangerousCount := 0

	// 检查是否允许访问敏感路径
	for _, path := range config.SecurityConfig.AllowedPaths {
		if strings.HasPrefix(path, "/etc") || strings.HasPrefix(path, "/root") || strings.HasPrefix(path, "/sys") {
			result.AddWarning("security_config.allowed_paths", fmt.Sprintf("Dangerous path access allowed: %s", path), "DANGEROUS_PERMISSION", path)
			dangerousCount++
		}
	}

	// 检查是否允许访问危险端口
	for _, port := range config.SecurityConfig.AllowedPorts {
		// 排除常用的HTTP/HTTPS端口
		if port == 22 || port == 23 || port == 3389 || (port < 1024 && port != 80 && port != 443) {
			result.AddWarning("security_config.allowed_ports", fmt.Sprintf("Dangerous port access allowed: %d", port), "DANGEROUS_PERMISSION", port)
			dangerousCount++
		}
	}

	// 检查是否允许危险系统调用
	for _, syscall := range config.SecurityConfig.AllowedSyscalls {
		if syscall == "execve" || syscall == "fork" || syscall == "clone" {
			result.AddWarning("security_config.allowed_syscalls", fmt.Sprintf("Dangerous syscall allowed: %s", syscall), "DANGEROUS_PERMISSION", syscall)
			dangerousCount++
		}
	}

	// 检查是否禁用了重要的安全特性
	if !config.SecurityConfig.EnableSeccomp {
		result.AddWarning("security_config.enable_seccomp", "Seccomp is disabled, which may allow dangerous system calls", "DANGEROUS_PERMISSION", false)
		dangerousCount++
	}

	if !config.SecurityConfig.EnableNamespace {
		result.AddWarning("security_config.enable_namespace", "Namespace isolation is disabled, which may allow privilege escalation", "DANGEROUS_PERMISSION", false)
		dangerousCount++
	}

	// 检查连接数限制是否过高
	if config.SecurityConfig.MaxConnections > 100 {
		result.AddWarning("security_config.max_connections", fmt.Sprintf("High connection limit may cause resource exhaustion: %d", config.SecurityConfig.MaxConnections), "DANGEROUS_PERMISSION", config.SecurityConfig.MaxConnections)
		dangerousCount++
	}

	// 检查连接超时是否过长
	if config.SecurityConfig.ConnectionTimeout > 300 {
		result.AddWarning("security_config.connection_timeout", fmt.Sprintf("Long connection timeout may cause resource leaks: %d seconds", config.SecurityConfig.ConnectionTimeout), "DANGEROUS_PERMISSION", config.SecurityConfig.ConnectionTimeout)
		dangerousCount++
	}

	return result
}

// NetworkValidationRule 网络验证规则
type NetworkValidationRule struct {
	BaseValidationRule
}

// NewNetworkValidationRule 创建网络验证规则
func NewNetworkValidationRule() *NetworkValidationRule {
	return &NetworkValidationRule{
		BaseValidationRule: BaseValidationRule{
			Name:        "network_validation",
			Description: "Validate network configurations",
			Severity:    ValidationSeverityError,
		},
	}
}

// Validate 验证网络配置
func (nvr *NetworkValidationRule) Validate(ctx context.Context, config *EnhancedPluginConfig) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// 从自定义配置中提取网络相关配置
	networkConfigs := nvr.extractNetworkConfigs(config.CustomConfig)

	for field, value := range networkConfigs {
		switch {
		case strings.Contains(strings.ToLower(field), "url"):
			if urlStr, ok := value.(string); ok {
				if !nvr.isValidURL(urlStr) {
					result.AddError(field, "Invalid URL format", "INVALID_URL", urlStr)
				}
			}

		case strings.Contains(strings.ToLower(field), "port"):
			var port int
			switch v := value.(type) {
			case int:
				port = v
			case float64:
				port = int(v)
			case string:
				if p, err := strconv.Atoi(v); err == nil {
					port = p
				}
			}

			if port <= 0 || port > 65535 {
				result.AddError(field, "Port number must be between 1 and 65535", "INVALID_PORT", port)
			}

		case strings.Contains(strings.ToLower(field), "host") || strings.Contains(strings.ToLower(field), "address"):
			if hostStr, ok := value.(string); ok {
				if !nvr.isValidHost(hostStr) {
					result.AddError(field, "Invalid host format", "INVALID_HOST", hostStr)
				}
			}
		}
	}

	return result
}

// extractNetworkConfigs 提取网络相关配置
func (nvr *NetworkValidationRule) extractNetworkConfigs(config map[string]interface{}) map[string]interface{} {
	networkConfigs := make(map[string]interface{})
	nvr.extractNetworkConfigsRecursive("", config, networkConfigs)
	return networkConfigs
}

// extractNetworkConfigsRecursive 递归提取网络配置
func (nvr *NetworkValidationRule) extractNetworkConfigsRecursive(prefix string, obj interface{}, result map[string]interface{}) {
	switch v := obj.(type) {
	case map[string]interface{}:
		for key, value := range v {
			fullKey := key
			if prefix != "" {
				fullKey = prefix + "." + key
			}

			lowerKey := strings.ToLower(key)
			if strings.Contains(lowerKey, "url") || strings.Contains(lowerKey, "port") ||
				strings.Contains(lowerKey, "host") || strings.Contains(lowerKey, "address") {
				result[fullKey] = value
			} else {
				nvr.extractNetworkConfigsRecursive(fullKey, value, result)
			}
		}
	case []interface{}:
		for i, item := range v {
			fullKey := fmt.Sprintf("%s[%d]", prefix, i)
			nvr.extractNetworkConfigsRecursive(fullKey, item, result)
		}
	}
}

// isValidURL 检查URL是否有效
func (nvr *NetworkValidationRule) isValidURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// isValidHost 检查主机地址是否有效
func (nvr *NetworkValidationRule) isValidHost(host string) bool {
	// 检查是否为IP地址
	if net.ParseIP(host) != nil {
		return true
	}

	// 检查是否为有效的域名
	if len(host) == 0 || len(host) > 253 {
		return false
	}

	// 简单的域名格式检查
	pattern := `^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`
	matched, _ := regexp.MatchString(pattern, host)
	return matched
}

// CustomValidationRule 自定义验证规则
type CustomValidationRule struct {
	BaseValidationRule
	ValidateFunc func(ctx context.Context, config *EnhancedPluginConfig) *ValidationResult
}

// NewCustomValidationRule 创建自定义验证规则
func NewCustomValidationRule(name, description string, severity ValidationSeverity, validateFunc func(ctx context.Context, config *EnhancedPluginConfig) *ValidationResult) *CustomValidationRule {
	return &CustomValidationRule{
		BaseValidationRule: BaseValidationRule{
			Name:        name,
			Description: description,
			Severity:    severity,
		},
		ValidateFunc: validateFunc,
	}
}

// Validate 执行自定义验证
func (cvr *CustomValidationRule) Validate(ctx context.Context, config *EnhancedPluginConfig) *ValidationResult {
	if cvr.ValidateFunc != nil {
		return cvr.ValidateFunc(ctx, config)
	}
	return &ValidationResult{Valid: true}
}

// DetailedConfigValidator 详细配置验证器
type DetailedConfigValidator struct {
	logger *slog.Logger
	rules  []ValidationRule
	level  ValidationLevel
	mu     sync.RWMutex
}

// ValidateSimple 实现ConfigValidator接口的简单验证方法
func (cv *DetailedConfigValidator) ValidateSimple(ctx context.Context, config *EnhancedPluginConfig) error {
	result := cv.Validate(ctx, config)
	if !result.Valid {
		if len(result.Errors) > 0 {
			return fmt.Errorf("validation failed: %s", result.Errors[0].Message)
		}
		return fmt.Errorf("validation failed")
	}
	return nil
}

// ConfigValidatorAdapter 配置验证器适配器
type ConfigValidatorAdapter struct {
	detailedValidator *DetailedConfigValidator
}

// NewConfigValidatorAdapter 创建配置验证器适配器
func NewConfigValidatorAdapter(detailedValidator *DetailedConfigValidator) *ConfigValidatorAdapter {
	return &ConfigValidatorAdapter{
		detailedValidator: detailedValidator,
	}
}

// Validate 实现ConfigValidator接口
func (cva *ConfigValidatorAdapter) Validate(ctx context.Context, config *EnhancedPluginConfig) error {
	return cva.detailedValidator.ValidateSimple(ctx, config)
}

// ConfigValidatorOptions 配置验证器选项
type ConfigValidatorOptions struct {
	Level        ValidationLevel
	CustomRules  []ValidationRule
	SkipBuiltins bool
}

// DefaultConfigValidatorOptions 默认配置验证器选项
func DefaultConfigValidatorOptions() *ConfigValidatorOptions {
	return &ConfigValidatorOptions{
		Level:        ValidationLevelStandard,
		CustomRules:  nil,
		SkipBuiltins: false,
	}
}

// NewConfigValidator 创建配置验证器
func NewConfigValidator(logger *slog.Logger, options *ConfigValidatorOptions) *DetailedConfigValidator {
	if options == nil {
		options = DefaultConfigValidatorOptions()
	}

	validator := &DetailedConfigValidator{
		logger: logger,
		rules:  make([]ValidationRule, 0),
		level:  options.Level,
	}

	// 添加内置规则
	if !options.SkipBuiltins {
		validator.addBuiltinRules()
	}

	// 添加自定义规则
	for _, rule := range options.CustomRules {
		validator.AddRule(rule)
	}

	return validator
}

// addBuiltinRules 添加内置验证规则
func (cv *DetailedConfigValidator) addBuiltinRules() {
	// 基础规则（所有级别都包含）
	cv.rules = append(cv.rules, NewRequiredFieldsRule())
	cv.rules = append(cv.rules, NewFormatValidationRule())

	// 标准级别及以上
	if cv.level == ValidationLevelStandard || cv.level == ValidationLevelStrict {
		cv.rules = append(cv.rules, NewPathValidationRule())
		cv.rules = append(cv.rules, NewResourceValidationRule())
		cv.rules = append(cv.rules, NewSecurityValidationRule())
	}

	// 严格级别
	if cv.level == ValidationLevelStrict {
		cv.rules = append(cv.rules, NewDependencyValidationRule())
		cv.rules = append(cv.rules, NewNetworkValidationRule())
	}
}

// AddRule 添加验证规则
func (cv *DetailedConfigValidator) AddRule(rule ValidationRule) {
	cv.mu.Lock()
	defer cv.mu.Unlock()
	cv.rules = append(cv.rules, rule)
}

// RemoveRule 移除验证规则
func (cv *DetailedConfigValidator) RemoveRule(name string) {
	cv.mu.Lock()
	defer cv.mu.Unlock()

	newRules := make([]ValidationRule, 0, len(cv.rules))
	for _, rule := range cv.rules {
		if rule.GetName() != name {
			newRules = append(newRules, rule)
		}
	}
	cv.rules = newRules
}

// GetRules 获取所有规则
func (cv *DetailedConfigValidator) GetRules() []ValidationRule {
	cv.mu.RLock()
	defer cv.mu.RUnlock()

	rules := make([]ValidationRule, len(cv.rules))
	copy(rules, cv.rules)
	return rules
}

// SetLevel 设置验证级别
func (cv *DetailedConfigValidator) SetLevel(level ValidationLevel) {
	cv.mu.Lock()
	defer cv.mu.Unlock()

	cv.level = level
	// 重新构建规则列表
	cv.rules = make([]ValidationRule, 0)
	cv.addBuiltinRules()
}

// Validate 验证配置
func (cv *DetailedConfigValidator) Validate(ctx context.Context, config *EnhancedPluginConfig) *ValidationResult {
	cv.mu.RLock()
	defer cv.mu.RUnlock()

	result := &ValidationResult{
		Valid:    true,
		Metadata: make(map[string]interface{}),
	}

	// 记录验证开始时间
	startTime := time.Now()

	// 执行所有验证规则
	for _, rule := range cv.rules {
		ruleResult := rule.Validate(ctx, config)
		result.Merge(ruleResult)

		cv.logger.Debug("Validation rule executed",
			"rule", rule.GetName(),
			"plugin_id", config.ID,
			"errors", len(ruleResult.Errors),
			"warnings", len(ruleResult.Warnings))
	}

	// 记录验证统计信息
	result.Metadata["validation_duration"] = time.Since(startTime)
	result.Metadata["validation_level"] = cv.level
	result.Metadata["rules_executed"] = len(cv.rules)
	result.Metadata["plugin_id"] = config.ID

	cv.logger.Info("Config validation completed",
		"plugin_id", config.ID,
		"valid", result.Valid,
		"errors", len(result.Errors),
		"warnings", len(result.Warnings),
		"duration", time.Since(startTime))

	return result
}

// ValidateBatch 批量验证配置
func (cv *DetailedConfigValidator) ValidateBatch(ctx context.Context, configs map[string]*EnhancedPluginConfig) map[string]*ValidationResult {
	results := make(map[string]*ValidationResult)

	// 设置依赖验证规则的可用插件
	for _, rule := range cv.rules {
		if depRule, ok := rule.(*DependencyValidationRule); ok {
			depRule.SetAvailablePlugins(configs)
			break
		}
	}

	// 并发验证
	var wg sync.WaitGroup
	var mu sync.Mutex

	for pluginID, config := range configs {
		wg.Add(1)
		go func(id string, cfg *EnhancedPluginConfig) {
			defer wg.Done()

			result := cv.Validate(ctx, cfg)

			mu.Lock()
			results[id] = result
			mu.Unlock()
		}(pluginID, config)
	}

	wg.Wait()

	cv.logger.Info("Batch validation completed", "configs", len(configs))
	return results
}

// ValidateField 验证单个字段
func (cv *DetailedConfigValidator) ValidateField(ctx context.Context, config *EnhancedPluginConfig, fieldPath string, value interface{}) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// 根据字段路径执行相应的验证
	switch fieldPath {
	case "id":
		if str, ok := value.(string); ok {
			if str == "" {
				result.AddError(fieldPath, "ID cannot be empty", "REQUIRED_FIELD_MISSING", value)
			} else {
				pattern := `^[a-zA-Z0-9_-]+$`
				if matched, _ := regexp.MatchString(pattern, str); !matched {
					result.AddError(fieldPath, "ID contains invalid characters", "INVALID_FORMAT", value)
				}
			}
		}

	case "version":
		if str, ok := value.(string); ok {
			pattern := `^\d+\.\d+\.\d+(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$`
			if matched, _ := regexp.MatchString(pattern, str); !matched {
				result.AddError(fieldPath, "Invalid version format", "INVALID_FORMAT", value)
			}
		}

	case "plugin_path":
		if str, ok := value.(string); ok {
			if !filepath.IsAbs(str) {
				result.AddError(fieldPath, "Path must be absolute", "INVALID_FORMAT", value)
			}
		}

	default:
		// 对于未知字段，执行通用验证
		if value == nil {
			result.AddWarning(fieldPath, "Field is null", "NULL_VALUE", value)
		}
	}

	return result
}

// GetValidationSchema 获取验证模式
func (cv *DetailedConfigValidator) GetValidationSchema() map[string]interface{} {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":    "string",
				"pattern": "^[a-zA-Z0-9_-]+$",
				"minLength": 2,
				"maxLength": 64,
			},
			"name": map[string]interface{}{
				"type":      "string",
				"minLength": 1,
				"maxLength": 100,
			},
			"version": map[string]interface{}{
				"type":    "string",
				"pattern": "^\\d+\\.\\d+\\.\\d+(-[a-zA-Z0-9.-]+)?(\\+[a-zA-Z0-9.-]+)?$",
			},
			"plugin_path": map[string]interface{}{
				"type": "string",
			},
			"enabled": map[string]interface{}{
				"type": "boolean",
			},
		},
		"required": []string{"id", "name", "version", "plugin_path"},
	}

	return schema
}

// ValidationReport 验证报告
type ValidationReport struct {
	Summary    ValidationSummary            `json:"summary"`
	Results    map[string]*ValidationResult `json:"results"`
	Timestamp  time.Time                    `json:"timestamp"`
	Duration   time.Duration                `json:"duration"`
	Level      ValidationLevel              `json:"level"`
}

// ValidationSummary 验证摘要
type ValidationSummary struct {
	Total       int `json:"total"`
	Valid       int `json:"valid"`
	Invalid     int `json:"invalid"`
	TotalErrors int `json:"total_errors"`
	TotalWarnings int `json:"total_warnings"`
}

// GenerateReport 生成验证报告
func (cv *DetailedConfigValidator) GenerateReport(ctx context.Context, configs map[string]*EnhancedPluginConfig) *ValidationReport {
	startTime := time.Now()

	results := cv.ValidateBatch(ctx, configs)

	summary := ValidationSummary{
		Total: len(configs),
	}

	for _, result := range results {
		if result.Valid {
			summary.Valid++
		} else {
			summary.Invalid++
		}
		summary.TotalErrors += len(result.Errors)
		summary.TotalWarnings += len(result.Warnings)
	}

	return &ValidationReport{
		Summary:   summary,
		Results:   results,
		Timestamp: startTime,
		Duration:  time.Since(startTime),
		Level:     cv.level,
	}
}