package recovery

import (
	"context"
	"log/slog"
	"time"

	pluginError "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/error"
)

// RecoveryIntegration 恢复策略集成器
type RecoveryIntegration struct {
	recoveryManager *RecoveryManager
	configManager   *ConfigManager
	autoRecovery    *AutoRecoveryManager
	logger          *slog.Logger
}

// NewRecoveryIntegration 创建新的恢复策略集成器
func NewRecoveryIntegration(logger *slog.Logger) *RecoveryIntegration {
	return &RecoveryIntegration{
		configManager: NewConfigManager(logger),
		logger:        logger,
	}
}

// Initialize 初始化恢复策略系统
func (ri *RecoveryIntegration) Initialize(ctx context.Context) error {
	// 1. 设置默认配置
	if err := ri.setupDefaultConfigs(); err != nil {
		return err
	}

	// 2. 创建恢复管理器
	rmConfig := DefaultRecoveryManagerConfig()
	ri.recoveryManager = NewRecoveryManager(rmConfig, ri.logger)

	// 3. 创建并注册策略
	if err := ri.createAndRegisterStrategies(); err != nil {
		return err
	}

	// 4. 创建自动恢复管理器
	autoConfig := DefaultAutoRecoveryConfig()
	ri.autoRecovery = NewAutoRecoveryManager(autoConfig, ri.logger)

	// 5. 启动管理器
	if err := ri.recoveryManager.Start(ctx); err != nil {
		return err
	}

	if err := ri.autoRecovery.Start(ctx); err != nil {
		return err
	}

	ri.logger.Info("Recovery integration initialized successfully")
	return nil
}

// setupDefaultConfigs 设置默认配置
func (ri *RecoveryIntegration) setupDefaultConfigs() error {
	// 熔断器配置
	cbConfig := DefaultCircuitBreakerConfig()
	cbConfig.FailureThreshold = 5
	cbConfig.RecoveryTimeout = 30 * time.Second
	ri.configManager.AddCircuitBreakerConfig("default-circuit-breaker", cbConfig)

	// 重试策略配置
	retryConfig := DefaultRetryConfig()
	retryConfig.MaxRetries = 3
	retryConfig.Policy = RetryPolicyExponential
	retryConfig.RetryableErrors = []string{"timeout", "connection", "temporary", "network"}
	ri.configManager.AddRetryConfig("default-retry", retryConfig)

	// 降级策略配置
	fallbackConfig := DefaultFallbackConfig()
	fallbackConfig.Type = FallbackTypeDefault
	fallbackConfig.DefaultValue = "Service temporarily unavailable"
	ri.configManager.AddFallbackConfig("default-fallback", fallbackConfig)

	// 策略配置
	policyConfig := &PolicyConfig{
		Name:        "default-policy",
		Description: "Default recovery policy",
		Enabled:     true,
		Priority:    1,
		Strategies:  []string{"default-circuit-breaker", "default-retry", "default-fallback"},
		Conditions:  []string{"error_rate > 0.1", "response_time > 5s"},
	}
	ri.configManager.AddPolicyConfig("default-policy", policyConfig)

	return nil
}

// createAndRegisterStrategies 创建并注册策略
func (ri *RecoveryIntegration) createAndRegisterStrategies() error {
	// 创建熔断器策略
	cbConfig, _ := ri.configManager.GetCircuitBreakerConfig("default-circuit-breaker")
	cb := NewCircuitBreaker("default-circuit-breaker", cbConfig, ri.logger)
	cbWrapper := &CircuitBreakerStrategyWrapper{cb: cb}
	ri.recoveryManager.RegisterStrategy(cbWrapper)

	// 创建重试策略
	retryConfig, _ := ri.configManager.GetRetryConfig("default-retry")
	rs := NewRetryStrategy("default-retry", retryConfig, ri.logger)
	retryWrapper := &RetryStrategyWrapper{rs: rs}
	ri.recoveryManager.RegisterStrategy(retryWrapper)

	// 创建降级策略
	fallbackConfig, _ := ri.configManager.GetFallbackConfig("default-fallback")
	fs := NewFallbackStrategy("default-fallback", fallbackConfig, ri.logger)
	fallbackWrapper := &FallbackStrategyWrapper{fs: fs}
	ri.recoveryManager.RegisterStrategy(fallbackWrapper)

	return nil
}

// HandlePluginError 处理插件错误，集成到现有错误处理系统
func (ri *RecoveryIntegration) HandlePluginError(ctx context.Context, pluginID string, err pluginError.PluginError, operation func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	// 根据错误类型选择恢复策略
	strategies := ri.selectStrategiesForError(err)

	// 执行恢复策略
	result, recoveryErr := ri.recoveryManager.ExecuteRecovery(ctx, pluginID, strategies, operation, nil)
	if recoveryErr == nil {
		return result, nil
	}

	// 如果恢复失败，记录到自动恢复系统
	ri.autoRecovery.RegisterPlugin(pluginID)

	return nil, recoveryErr
}

// selectStrategiesForError 根据错误类型选择恢复策略
func (ri *RecoveryIntegration) selectStrategiesForError(err pluginError.PluginError) []string {
	switch err.GetCode() {
	case pluginError.ErrorCodePluginTimeout:
		return []string{"default-circuit-breaker", "default-retry"}
	case pluginError.ErrorCodePluginCrashed:
		return []string{"default-fallback"}
	case pluginError.ErrorCodePluginInitFailed:
		return []string{"default-retry", "default-fallback"}
	case pluginError.ErrorCodePluginNotFound:
		return []string{"default-retry"}
	default:
		return []string{"default-circuit-breaker", "default-retry", "default-fallback"}
	}
}

// GetRecoveryManager 获取恢复管理器
func (ri *RecoveryIntegration) GetRecoveryManager() *RecoveryManager {
	return ri.recoveryManager
}

// GetConfigManager 获取配置管理器
func (ri *RecoveryIntegration) GetConfigManager() *ConfigManager {
	return ri.configManager
}

// GetAutoRecoveryManager 获取自动恢复管理器
func (ri *RecoveryIntegration) GetAutoRecoveryManager() *AutoRecoveryManager {
	return ri.autoRecovery
}

// Shutdown 关闭恢复策略系统
func (ri *RecoveryIntegration) Shutdown() error {
	if ri.autoRecovery != nil {
		ri.autoRecovery.Stop()
	}

	if ri.recoveryManager != nil {
		ri.recoveryManager.Stop()
	}

	ri.logger.Info("Recovery integration shutdown completed")
	return nil
}

// 策略包装器实现

// CircuitBreakerStrategyWrapper 熔断器策略包装器
type CircuitBreakerStrategyWrapper struct {
	cb *CircuitBreaker
}

func (cbw *CircuitBreakerStrategyWrapper) GetName() string {
	return cbw.cb.GetName()
}

func (cbw *CircuitBreakerStrategyWrapper) GetType() StrategyType {
	return StrategyTypeCircuitBreaker
}

func (cbw *CircuitBreakerStrategyWrapper) Execute(ctx context.Context, operation func(ctx context.Context) (interface{}, error), args interface{}) (interface{}, error) {
	var result interface{}
	err := cbw.cb.Execute(ctx, func(ctx context.Context) error {
		var opErr error
		result, opErr = operation(ctx)
		return opErr
	})
	return result, err
}

func (cbw *CircuitBreakerStrategyWrapper) Reset() {
	cbw.cb.Reset()
}

func (cbw *CircuitBreakerStrategyWrapper) IsHealthy() bool {
	return cbw.cb.IsHealthy()
}

// RetryStrategyWrapper 重试策略包装器
type RetryStrategyWrapper struct {
	rs *RetryStrategy
}

func (rsw *RetryStrategyWrapper) GetName() string {
	return rsw.rs.GetName()
}

func (rsw *RetryStrategyWrapper) GetType() StrategyType {
	return StrategyTypeRetry
}

func (rsw *RetryStrategyWrapper) Execute(ctx context.Context, operation func(ctx context.Context) (interface{}, error), args interface{}) (interface{}, error) {
	var result interface{}
	err := rsw.rs.Execute(ctx, func(ctx context.Context) error {
		var opErr error
		result, opErr = operation(ctx)
		return opErr
	})
	return result, err
}

func (rsw *RetryStrategyWrapper) Reset() {
	rsw.rs.Reset()
}

func (rsw *RetryStrategyWrapper) IsHealthy() bool {
	return rsw.rs.GetSuccessRate() > 0.5
}

// FallbackStrategyWrapper 降级策略包装器
type FallbackStrategyWrapper struct {
	fs *FallbackStrategy
}

func (fsw *FallbackStrategyWrapper) GetName() string {
	return fsw.fs.GetName()
}

func (fsw *FallbackStrategyWrapper) GetType() StrategyType {
	return StrategyTypeFallback
}

func (fsw *FallbackStrategyWrapper) Execute(ctx context.Context, operation func(ctx context.Context) (interface{}, error), args interface{}) (interface{}, error) {
	return fsw.fs.Execute(ctx, operation, args)
}

func (fsw *FallbackStrategyWrapper) Reset() {
	fsw.fs.Reset()
}

func (fsw *FallbackStrategyWrapper) IsHealthy() bool {
	return fsw.fs.GetSuccessRate() > 0.8
}