package recovery

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// ConfigVersion 配置版本
type ConfigVersion string

const (
	// ConfigVersionV1 版本1
	ConfigVersionV1 ConfigVersion = "v1"
)

// RecoveryConfig 恢复策略配置
type RecoveryConfig struct {
	// Version 配置版本
	Version ConfigVersion `json:"version"`
	// Metadata 元数据
	Metadata ConfigMetadata `json:"metadata"`
	// Manager 管理器配置
	Manager *RecoveryManagerConfig `json:"manager"`
	// CircuitBreakers 熔断器配置
	CircuitBreakers map[string]*CircuitBreakerConfig `json:"circuit_breakers"`
	// RetryStrategies 重试策略配置
	RetryStrategies map[string]*RetryConfig `json:"retry_strategies"`
	// FallbackStrategies 降级策略配置
	FallbackStrategies map[string]*FallbackConfig `json:"fallback_strategies"`
	// AutoRecovery 自动恢复配置
	AutoRecovery *AutoRecoveryConfig `json:"auto_recovery"`
	// Policies 策略映射配置
	Policies map[string]*PolicyConfig `json:"policies"`
}

// ConfigMetadata 配置元数据
type ConfigMetadata struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Version     string            `json:"version"`
	Author      string            `json:"author"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Tags        []string          `json:"tags"`
	Labels      map[string]string `json:"labels"`
}

// PolicyConfig 策略配置
type PolicyConfig struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Enabled     bool     `json:"enabled"`
	Priority    int      `json:"priority"`
	Strategies  []string `json:"strategies"`
	Conditions  []string `json:"conditions"`
	PluginIDs   []string `json:"plugin_ids"`
	ErrorCodes  []string `json:"error_codes"`
}

// DefaultRecoveryConfig 返回默认恢复配置
func DefaultRecoveryConfig() *RecoveryConfig {
	return &RecoveryConfig{
		Version: ConfigVersionV1,
		Metadata: ConfigMetadata{
			Name:        "default-recovery-config",
			Description: "Default recovery configuration",
			Version:     "1.0.0",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Tags:        []string{"default", "recovery"},
			Labels:      make(map[string]string),
		},
		Manager:            DefaultRecoveryManagerConfig(),
		CircuitBreakers:    make(map[string]*CircuitBreakerConfig),
		RetryStrategies:    make(map[string]*RetryConfig),
		FallbackStrategies: make(map[string]*FallbackConfig),
		AutoRecovery:       DefaultAutoRecoveryConfig(),
		Policies:           make(map[string]*PolicyConfig),
	}
}

// ConfigValidator 配置验证器
type ConfigValidator struct {
	logger *slog.Logger
}

// NewConfigValidator 创建新的配置验证器
func NewConfigValidator(logger *slog.Logger) *ConfigValidator {
	return &ConfigValidator{
		logger: logger,
	}
}

// Validate 验证配置
func (cv *ConfigValidator) Validate(config *RecoveryConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// 验证版本
	if err := cv.validateVersion(config.Version); err != nil {
		return fmt.Errorf("invalid version: %w", err)
	}

	// 验证管理器配置
	if err := cv.validateManagerConfig(config.Manager); err != nil {
		return fmt.Errorf("invalid manager config: %w", err)
	}

	// 验证熔断器配置
	if err := cv.validateCircuitBreakerConfigs(config.CircuitBreakers); err != nil {
		return fmt.Errorf("invalid circuit breaker configs: %w", err)
	}

	// 验证重试策略配置
	if err := cv.validateRetryConfigs(config.RetryStrategies); err != nil {
		return fmt.Errorf("invalid retry configs: %w", err)
	}

	// 验证降级策略配置
	if err := cv.validateFallbackConfigs(config.FallbackStrategies); err != nil {
		return fmt.Errorf("invalid fallback configs: %w", err)
	}

	// 验证自动恢复配置
	if err := cv.validateAutoRecoveryConfig(config.AutoRecovery); err != nil {
		return fmt.Errorf("invalid auto recovery config: %w", err)
	}

	// 验证策略配置
	if err := cv.validatePolicyConfigs(config.Policies); err != nil {
		return fmt.Errorf("invalid policy configs: %w", err)
	}

	return nil
}

// validateVersion 验证版本
func (cv *ConfigValidator) validateVersion(version ConfigVersion) error {
	switch version {
	case ConfigVersionV1:
		return nil
	default:
		return fmt.Errorf("unsupported config version: %s", version)
	}
}

// validateManagerConfig 验证管理器配置
func (cv *ConfigValidator) validateManagerConfig(config *RecoveryManagerConfig) error {
	if config == nil {
		return fmt.Errorf("manager config cannot be nil")
	}

	if config.MaxConcurrentRecoveries <= 0 {
		return fmt.Errorf("max concurrent recoveries must be positive")
	}

	if config.RecoveryTimeout <= 0 {
		return fmt.Errorf("recovery timeout must be positive")
	}

	if config.HealthCheckInterval <= 0 {
		return fmt.Errorf("health check interval must be positive")
	}

	if config.MaxHistorySize <= 0 {
		return fmt.Errorf("max history size must be positive")
	}

	return nil
}

// validateCircuitBreakerConfigs 验证熔断器配置
func (cv *ConfigValidator) validateCircuitBreakerConfigs(configs map[string]*CircuitBreakerConfig) error {
	for name, config := range configs {
		if err := cv.validateCircuitBreakerConfig(name, config); err != nil {
			return err
		}
	}
	return nil
}

// validateCircuitBreakerConfig 验证单个熔断器配置
func (cv *ConfigValidator) validateCircuitBreakerConfig(name string, config *CircuitBreakerConfig) error {
	if config == nil {
		return fmt.Errorf("circuit breaker config '%s' cannot be nil", name)
	}

	if config.FailureThreshold <= 0 {
		return fmt.Errorf("circuit breaker '%s': failure threshold must be positive", name)
	}

	if config.SuccessThreshold <= 0 {
		return fmt.Errorf("circuit breaker '%s': success threshold must be positive", name)
	}

	if config.Timeout <= 0 {
		return fmt.Errorf("circuit breaker '%s': timeout must be positive", name)
	}

	if config.RecoveryTimeout <= 0 {
		return fmt.Errorf("circuit breaker '%s': recovery timeout must be positive", name)
	}

	if config.MaxRequests <= 0 {
		return fmt.Errorf("circuit breaker '%s': max requests must be positive", name)
	}

	return nil
}

// validateRetryConfigs 验证重试配置
func (cv *ConfigValidator) validateRetryConfigs(configs map[string]*RetryConfig) error {
	for name, config := range configs {
		if err := cv.validateRetryConfig(name, config); err != nil {
			return err
		}
	}
	return nil
}

// validateRetryConfig 验证单个重试配置
func (cv *ConfigValidator) validateRetryConfig(name string, config *RetryConfig) error {
	if config == nil {
		return fmt.Errorf("retry config '%s' cannot be nil", name)
	}

	if config.MaxRetries < 0 {
		return fmt.Errorf("retry config '%s': max retries cannot be negative", name)
	}

	if config.InitialDelay <= 0 {
		return fmt.Errorf("retry config '%s': initial delay must be positive", name)
	}

	if config.MaxDelay <= 0 {
		return fmt.Errorf("retry config '%s': max delay must be positive", name)
	}

	if config.BackoffFactor <= 0 {
		return fmt.Errorf("retry config '%s': backoff factor must be positive", name)
	}

	if config.JitterFactor < 0 || config.JitterFactor > 1 {
		return fmt.Errorf("retry config '%s': jitter factor must be between 0 and 1", name)
	}

	if config.Timeout <= 0 {
		return fmt.Errorf("retry config '%s': timeout must be positive", name)
	}

	return nil
}

// validateFallbackConfigs 验证降级配置
func (cv *ConfigValidator) validateFallbackConfigs(configs map[string]*FallbackConfig) error {
	for name, config := range configs {
		if err := cv.validateFallbackConfig(name, config); err != nil {
			return err
		}
	}
	return nil
}

// validateFallbackConfig 验证单个降级配置
func (cv *ConfigValidator) validateFallbackConfig(name string, config *FallbackConfig) error {
	if config == nil {
		return fmt.Errorf("fallback config '%s' cannot be nil", name)
	}

	if config.Timeout <= 0 {
		return fmt.Errorf("fallback config '%s': timeout must be positive", name)
	}

	if config.MaxConcurrency <= 0 {
		return fmt.Errorf("fallback config '%s': max concurrency must be positive", name)
	}

	if config.CacheExpiry <= 0 {
		return fmt.Errorf("fallback config '%s': cache expiry must be positive", name)
	}

	if config.Priority < 0 {
		return fmt.Errorf("fallback config '%s': priority cannot be negative", name)
	}

	return nil
}

// validateAutoRecoveryConfig 验证自动恢复配置
func (cv *ConfigValidator) validateAutoRecoveryConfig(config *AutoRecoveryConfig) error {
	if config == nil {
		return fmt.Errorf("auto recovery config cannot be nil")
	}

	if config.HealthCheckInterval <= 0 {
		return fmt.Errorf("health check interval must be positive")
	}

	if config.HealthCheckTimeout <= 0 {
		return fmt.Errorf("health check timeout must be positive")
	}

	if config.MaxRecoveryAttempts <= 0 {
		return fmt.Errorf("max recovery attempts must be positive")
	}

	if config.RecoveryDelay < 0 {
		return fmt.Errorf("recovery delay cannot be negative")
	}

	if config.FailureThreshold <= 0 {
		return fmt.Errorf("failure threshold must be positive")
	}

	return nil
}

// validatePolicyConfigs 验证策略配置
func (cv *ConfigValidator) validatePolicyConfigs(configs map[string]*PolicyConfig) error {
	for name, config := range configs {
		if err := cv.validatePolicyConfig(name, config); err != nil {
			return err
		}
	}
	return nil
}

// validatePolicyConfig 验证单个策略配置
func (cv *ConfigValidator) validatePolicyConfig(name string, config *PolicyConfig) error {
	if config == nil {
		return fmt.Errorf("policy config '%s' cannot be nil", name)
	}

	if config.Name == "" {
		return fmt.Errorf("policy config '%s': name cannot be empty", name)
	}

	if config.Priority < 0 {
		return fmt.Errorf("policy config '%s': priority cannot be negative", name)
	}

	if len(config.Strategies) == 0 {
		return fmt.Errorf("policy config '%s': strategies cannot be empty", name)
	}

	return nil
}

// ConfigManager 配置管理器
type ConfigManager struct {
	config    *RecoveryConfig
	validator *ConfigValidator
	logger    *slog.Logger
	mutex     sync.RWMutex

	// 配置变更回调
	onConfigChanged func(oldConfig, newConfig *RecoveryConfig)
}

// NewConfigManager 创建新的配置管理器
func NewConfigManager(logger *slog.Logger) *ConfigManager {
	return &ConfigManager{
		config:    DefaultRecoveryConfig(),
		validator: NewConfigValidator(logger),
		logger:    logger,
	}
}

// LoadConfig 加载配置
func (cm *ConfigManager) LoadConfig(data []byte) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	var newConfig RecoveryConfig
	if err := json.Unmarshal(data, &newConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 验证配置
	if err := cm.validator.Validate(&newConfig); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// 保存旧配置
	oldConfig := cm.config

	// 更新配置
	cm.config = &newConfig
	cm.config.Metadata.UpdatedAt = time.Now()

	// 触发配置变更回调
	if cm.onConfigChanged != nil {
		cm.onConfigChanged(oldConfig, cm.config)
	}

	if cm.logger != nil {
		cm.logger.Info("Recovery config loaded successfully",
			"name", newConfig.Metadata.Name,
			"version", newConfig.Metadata.Version)
	}

	return nil
}

// SaveConfig 保存配置
func (cm *ConfigManager) SaveConfig() ([]byte, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// 更新时间戳
	cm.config.Metadata.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	return data, nil
}

// GetConfig 获取配置
func (cm *ConfigManager) GetConfig() *RecoveryConfig {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// 返回配置的深拷贝
	data, _ := json.Marshal(cm.config)
	var configCopy RecoveryConfig
	json.Unmarshal(data, &configCopy)
	return &configCopy
}

// UpdateManagerConfig 更新管理器配置
func (cm *ConfigManager) UpdateManagerConfig(config *RecoveryManagerConfig) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if err := cm.validator.validateManagerConfig(config); err != nil {
		return err
	}

	oldConfig := cm.config
	cm.config.Manager = config
	cm.config.Metadata.UpdatedAt = time.Now()

	if cm.onConfigChanged != nil {
		cm.onConfigChanged(oldConfig, cm.config)
	}

	return nil
}

// AddCircuitBreakerConfig 添加熔断器配置
func (cm *ConfigManager) AddCircuitBreakerConfig(name string, config *CircuitBreakerConfig) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if err := cm.validator.validateCircuitBreakerConfig(name, config); err != nil {
		return err
	}

	oldConfig := cm.config
	cm.config.CircuitBreakers[name] = config
	cm.config.Metadata.UpdatedAt = time.Now()

	if cm.onConfigChanged != nil {
		cm.onConfigChanged(oldConfig, cm.config)
	}

	if cm.logger != nil {
		cm.logger.Info("Circuit breaker config added", "name", name)
	}

	return nil
}

// RemoveCircuitBreakerConfig 移除熔断器配置
func (cm *ConfigManager) RemoveCircuitBreakerConfig(name string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	oldConfig := cm.config
	delete(cm.config.CircuitBreakers, name)
	cm.config.Metadata.UpdatedAt = time.Now()

	if cm.onConfigChanged != nil {
		cm.onConfigChanged(oldConfig, cm.config)
	}

	if cm.logger != nil {
		cm.logger.Info("Circuit breaker config removed", "name", name)
	}
}

// AddRetryConfig 添加重试配置
func (cm *ConfigManager) AddRetryConfig(name string, config *RetryConfig) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if err := cm.validator.validateRetryConfig(name, config); err != nil {
		return err
	}

	oldConfig := cm.config
	cm.config.RetryStrategies[name] = config
	cm.config.Metadata.UpdatedAt = time.Now()

	if cm.onConfigChanged != nil {
		cm.onConfigChanged(oldConfig, cm.config)
	}

	if cm.logger != nil {
		cm.logger.Info("Retry config added", "name", name)
	}

	return nil
}

// RemoveRetryConfig 移除重试配置
func (cm *ConfigManager) RemoveRetryConfig(name string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	oldConfig := cm.config
	delete(cm.config.RetryStrategies, name)
	cm.config.Metadata.UpdatedAt = time.Now()

	if cm.onConfigChanged != nil {
		cm.onConfigChanged(oldConfig, cm.config)
	}

	if cm.logger != nil {
		cm.logger.Info("Retry config removed", "name", name)
	}
}

// AddFallbackConfig 添加降级配置
func (cm *ConfigManager) AddFallbackConfig(name string, config *FallbackConfig) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if err := cm.validator.validateFallbackConfig(name, config); err != nil {
		return err
	}

	oldConfig := cm.config
	cm.config.FallbackStrategies[name] = config
	cm.config.Metadata.UpdatedAt = time.Now()

	if cm.onConfigChanged != nil {
		cm.onConfigChanged(oldConfig, cm.config)
	}

	if cm.logger != nil {
		cm.logger.Info("Fallback config added", "name", name)
	}

	return nil
}

// RemoveFallbackConfig 移除降级配置
func (cm *ConfigManager) RemoveFallbackConfig(name string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	oldConfig := cm.config
	delete(cm.config.FallbackStrategies, name)
	cm.config.Metadata.UpdatedAt = time.Now()

	if cm.onConfigChanged != nil {
		cm.onConfigChanged(oldConfig, cm.config)
	}

	if cm.logger != nil {
		cm.logger.Info("Fallback config removed", "name", name)
	}
}

// UpdateAutoRecoveryConfig 更新自动恢复配置
func (cm *ConfigManager) UpdateAutoRecoveryConfig(config *AutoRecoveryConfig) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if err := cm.validator.validateAutoRecoveryConfig(config); err != nil {
		return err
	}

	oldConfig := cm.config
	cm.config.AutoRecovery = config
	cm.config.Metadata.UpdatedAt = time.Now()

	if cm.onConfigChanged != nil {
		cm.onConfigChanged(oldConfig, cm.config)
	}

	return nil
}

// AddPolicyConfig 添加策略配置
func (cm *ConfigManager) AddPolicyConfig(name string, config *PolicyConfig) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if err := cm.validator.validatePolicyConfig(name, config); err != nil {
		return err
	}

	oldConfig := cm.config
	cm.config.Policies[name] = config
	cm.config.Metadata.UpdatedAt = time.Now()

	if cm.onConfigChanged != nil {
		cm.onConfigChanged(oldConfig, cm.config)
	}

	if cm.logger != nil {
		cm.logger.Info("Policy config added", "name", name)
	}

	return nil
}

// RemovePolicyConfig 移除策略配置
func (cm *ConfigManager) RemovePolicyConfig(name string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	oldConfig := cm.config
	delete(cm.config.Policies, name)
	cm.config.Metadata.UpdatedAt = time.Now()

	if cm.onConfigChanged != nil {
		cm.onConfigChanged(oldConfig, cm.config)
	}

	if cm.logger != nil {
		cm.logger.Info("Policy config removed", "name", name)
	}
}

// SetOnConfigChanged 设置配置变更回调
func (cm *ConfigManager) SetOnConfigChanged(callback func(oldConfig, newConfig *RecoveryConfig)) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.onConfigChanged = callback
}

// GetCircuitBreakerConfig 获取熔断器配置
func (cm *ConfigManager) GetCircuitBreakerConfig(name string) (*CircuitBreakerConfig, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	config, exists := cm.config.CircuitBreakers[name]
	if !exists {
		return nil, false
	}

	// 返回配置的副本
	configCopy := *config
	return &configCopy, true
}

// GetRetryConfig 获取重试配置
func (cm *ConfigManager) GetRetryConfig(name string) (*RetryConfig, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	config, exists := cm.config.RetryStrategies[name]
	if !exists {
		return nil, false
	}

	// 返回配置的副本
	configCopy := *config
	return &configCopy, true
}

// GetFallbackConfig 获取降级配置
func (cm *ConfigManager) GetFallbackConfig(name string) (*FallbackConfig, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	config, exists := cm.config.FallbackStrategies[name]
	if !exists {
		return nil, false
	}

	// 返回配置的副本
	configCopy := *config
	return &configCopy, true
}

// GetPolicyConfig 获取策略配置
func (cm *ConfigManager) GetPolicyConfig(name string) (*PolicyConfig, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	config, exists := cm.config.Policies[name]
	if !exists {
		return nil, false
	}

	// 返回配置的副本
	configCopy := *config
	return &configCopy, true
}

// ListCircuitBreakerConfigs 列出所有熔断器配置
func (cm *ConfigManager) ListCircuitBreakerConfigs() map[string]*CircuitBreakerConfig {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	configs := make(map[string]*CircuitBreakerConfig)
	for name, config := range cm.config.CircuitBreakers {
		configCopy := *config
		configs[name] = &configCopy
	}

	return configs
}

// ListRetryConfigs 列出所有重试配置
func (cm *ConfigManager) ListRetryConfigs() map[string]*RetryConfig {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	configs := make(map[string]*RetryConfig)
	for name, config := range cm.config.RetryStrategies {
		configCopy := *config
		configs[name] = &configCopy
	}

	return configs
}

// ListFallbackConfigs 列出所有降级配置
func (cm *ConfigManager) ListFallbackConfigs() map[string]*FallbackConfig {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	configs := make(map[string]*FallbackConfig)
	for name, config := range cm.config.FallbackStrategies {
		configCopy := *config
		configs[name] = &configCopy
	}

	return configs
}

// ListPolicyConfigs 列出所有策略配置
func (cm *ConfigManager) ListPolicyConfigs() map[string]*PolicyConfig {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	configs := make(map[string]*PolicyConfig)
	for name, config := range cm.config.Policies {
		configCopy := *config
		configs[name] = &configCopy
	}

	return configs
}