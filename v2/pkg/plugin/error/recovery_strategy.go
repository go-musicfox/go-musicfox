package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RecoveryStrategy 错误恢复策略接口
type RecoveryStrategy interface {
	// CanRecover 判断是否可以恢复
	CanRecover(err error) bool
	// Recover 执行恢复操作
	Recover(ctx context.Context, pluginID string, err error) error
	// GetRecoveryTimeout 获取恢复超时时间
	GetRecoveryTimeout() time.Duration
	// GetRecoveryType 获取恢复类型
	GetRecoveryType() RecoveryType
	// GetPriority 获取策略优先级
	GetPriority() int
}

// BaseRecoveryStrategy 基础恢复策略
type BaseRecoveryStrategy struct {
	recoveryType    RecoveryType
	recoveryTimeout time.Duration
	priority        int
	logger          Logger
	metrics         MetricsCollector
	eventBus        EventBus
	pluginLoader    PluginLoader
}

// NewBaseRecoveryStrategy 创建基础恢复策略
func NewBaseRecoveryStrategy(recoveryType RecoveryType, timeout time.Duration, priority int, logger Logger, metrics MetricsCollector, eventBus EventBus, pluginLoader PluginLoader) *BaseRecoveryStrategy {
	return &BaseRecoveryStrategy{
		recoveryType:    recoveryType,
		recoveryTimeout: timeout,
		priority:        priority,
		logger:          logger,
		metrics:         metrics,
		eventBus:        eventBus,
		pluginLoader:    pluginLoader,
	}
}

// GetRecoveryTimeout 获取恢复超时时间
func (brs *BaseRecoveryStrategy) GetRecoveryTimeout() time.Duration {
	return brs.recoveryTimeout
}

// GetRecoveryType 获取恢复类型
func (brs *BaseRecoveryStrategy) GetRecoveryType() RecoveryType {
	return brs.recoveryType
}

// GetPriority 获取策略优先级
func (brs *BaseRecoveryStrategy) GetPriority() int {
	return brs.priority
}

// RestartRecoveryStrategy 重启恢复策略
type RestartRecoveryStrategy struct {
	*BaseRecoveryStrategy
	maxRestartAttempts int
	restartDelay       time.Duration
}

// NewRestartRecoveryStrategy 创建重启恢复策略
func NewRestartRecoveryStrategy(maxAttempts int, restartDelay time.Duration, logger Logger, metrics MetricsCollector, eventBus EventBus, pluginLoader PluginLoader) *RestartRecoveryStrategy {
	return &RestartRecoveryStrategy{
		BaseRecoveryStrategy: NewBaseRecoveryStrategy(RecoveryTypeRestart, 30*time.Second, 1, logger, metrics, eventBus, pluginLoader),
		maxRestartAttempts:   maxAttempts,
		restartDelay:         restartDelay,
	}
}

// CanRecover 判断是否可以恢复
func (rrs *RestartRecoveryStrategy) CanRecover(err error) bool {
	if pluginErr, ok := err.(*BasePluginError); ok {
		switch pluginErr.GetCode() {
		case ErrorCodePluginCrashed, ErrorCodePluginTimeout, ErrorCodePluginMemoryLimit:
			return true
		}
	}
	return false
}

// Recover 执行恢复操作
func (rrs *RestartRecoveryStrategy) Recover(ctx context.Context, pluginID string, err error) error {
	if rrs.logger != nil {
		rrs.logger.Info("Attempting to restart plugin", map[string]interface{}{
			"plugin_id": pluginID,
			"error": err,
		})
	}
	
	// 记录恢复尝试指标
	if rrs.metrics != nil {
		rrs.metrics.IncrementCounter("recovery_attempt_total", map[string]string{
			"plugin_id":     pluginID,
			"recovery_type": rrs.recoveryType.String(),
		})
	}
	
	// 发送恢复开始事件
	if rrs.eventBus != nil {
		rrs.eventBus.Publish("recovery_started", map[string]interface{}{
			"plugin_id":     pluginID,
			"recovery_type": rrs.recoveryType.String(),
			"error":         err.Error(),
			"timestamp":     time.Now(),
		})
	}
	
	for attempt := 1; attempt <= rrs.maxRestartAttempts; attempt++ {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		// 等待重启延迟
		if attempt > 1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(rrs.restartDelay):
			}
		}
		
		// 尝试重启插件
		if rrs.pluginLoader != nil {
			if err := rrs.pluginLoader.RestartPlugin(pluginID); err != nil {
				if rrs.logger != nil {
					rrs.logger.Warn("Plugin restart failed", map[string]interface{}{
						"plugin_id": pluginID,
						"attempt": attempt,
						"error": err,
					})
				}
				continue
			}
			
			// 重启成功
			if rrs.logger != nil {
				rrs.logger.Info("Plugin restart successful", map[string]interface{}{
					"plugin_id": pluginID,
					"attempt": attempt,
				})
			}
			
			// 记录成功指标
			if rrs.metrics != nil {
				rrs.metrics.IncrementCounter("recovery_success_total", map[string]string{
					"plugin_id":     pluginID,
					"recovery_type": rrs.recoveryType.String(),
				})
			}
			
			// 发送恢复成功事件
			if rrs.eventBus != nil {
				rrs.eventBus.Publish("recovery_succeeded", map[string]interface{}{
					"plugin_id":     pluginID,
					"recovery_type": rrs.recoveryType.String(),
					"attempts":      attempt,
					"timestamp":     time.Now(),
				})
			}
			
			return nil
		}
	}
	
	// 所有重启尝试都失败了
	recoveryErr := fmt.Errorf("failed to restart plugin %s after %d attempts", pluginID, rrs.maxRestartAttempts)
	
	// 记录失败指标
	if rrs.metrics != nil {
		rrs.metrics.IncrementCounter("recovery_failure_total", map[string]string{
			"plugin_id":     pluginID,
			"recovery_type": rrs.recoveryType.String(),
		})
	}
	
	// 发送恢复失败事件
	if rrs.eventBus != nil {
		rrs.eventBus.Publish("recovery_failed", map[string]interface{}{
			"plugin_id":     pluginID,
			"recovery_type": rrs.recoveryType.String(),
			"attempts":      rrs.maxRestartAttempts,
			"error":         recoveryErr.Error(),
			"timestamp":     time.Now(),
		})
	}
	
	return recoveryErr
}

// ReloadRecoveryStrategy 重载恢复策略
type ReloadRecoveryStrategy struct {
	*BaseRecoveryStrategy
	maxReloadAttempts int
	reloadDelay       time.Duration
}

// NewReloadRecoveryStrategy 创建重载恢复策略
func NewReloadRecoveryStrategy(maxAttempts int, reloadDelay time.Duration, logger Logger, metrics MetricsCollector, eventBus EventBus, pluginLoader PluginLoader) *ReloadRecoveryStrategy {
	return &ReloadRecoveryStrategy{
		BaseRecoveryStrategy: NewBaseRecoveryStrategy(RecoveryTypeRestart, 20*time.Second, 2, logger, metrics, eventBus, pluginLoader),
		maxReloadAttempts:    maxAttempts,
		reloadDelay:          reloadDelay,
	}
}

// CanRecover 判断是否可以恢复
func (rrs *ReloadRecoveryStrategy) CanRecover(err error) bool {
	if pluginErr, ok := err.(*BasePluginError); ok {
		switch pluginErr.GetCode() {
		case ErrorCodePluginInitFailed, ErrorCodePluginConfigInvalid, ErrorCodePluginDependencyMissing:
			return true
		}
	}
	return false
}

// Recover 执行恢复操作
func (rrs *ReloadRecoveryStrategy) Recover(ctx context.Context, pluginID string, err error) error {
	if rrs.logger != nil {
		rrs.logger.Info("Attempting to reload plugin", map[string]interface{}{
			"plugin_id": pluginID,
			"error": err,
		})
	}
	
	for attempt := 1; attempt <= rrs.maxReloadAttempts; attempt++ {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		// 等待重载延迟
		if attempt > 1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(rrs.reloadDelay):
			}
		}
		
		// 尝试重载插件
		if rrs.pluginLoader != nil {
			if err := rrs.pluginLoader.ReloadPlugin(pluginID); err != nil {
				if rrs.logger != nil {
					rrs.logger.Warn("Plugin reload failed", map[string]interface{}{
						"plugin_id": pluginID,
						"attempt": attempt,
						"error": err,
					})
				}
				continue
			}
			
			// 重载成功
			if rrs.logger != nil {
				rrs.logger.Info("Plugin reload successful", map[string]interface{}{
					"plugin_name": pluginID,
					"attempt": attempt,
					"error": err,
				})
			}
			return nil
		}
	}
	
	return fmt.Errorf("failed to reload plugin %s after %d attempts", pluginID, rrs.maxReloadAttempts)
}

// FallbackRecoveryStrategy 降级恢复策略
type FallbackRecoveryStrategy struct {
	*BaseRecoveryStrategy
	fallbackPlugins map[string]string // 插件ID -> 降级插件ID
	degradationConfig *DegradationConfig
}

// NewFallbackRecoveryStrategy 创建降级恢复策略
func NewFallbackRecoveryStrategy(fallbackPlugins map[string]string, degradationConfig *DegradationConfig, logger Logger, metrics MetricsCollector, eventBus EventBus, pluginLoader PluginLoader) *FallbackRecoveryStrategy {
	return &FallbackRecoveryStrategy{
		BaseRecoveryStrategy: NewBaseRecoveryStrategy(RecoveryTypeFallback, 10*time.Second, 3, logger, metrics, eventBus, pluginLoader),
		fallbackPlugins:      fallbackPlugins,
		degradationConfig:    degradationConfig,
	}
}

// CanRecover 判断是否可以恢复
func (frs *FallbackRecoveryStrategy) CanRecover(err error) bool {
	// 降级策略可以处理大部分错误
	return true
}

// Recover 执行恢复操作
func (frs *FallbackRecoveryStrategy) Recover(ctx context.Context, pluginID string, err error) error {
	if frs.logger != nil {
		frs.logger.Info("Attempting fallback recovery", map[string]interface{}{
			"plugin_id": pluginID,
			"error": err,
		})
	}
	
	// 检查是否有降级插件
	if fallbackID, exists := frs.fallbackPlugins[pluginID]; exists {
		if frs.logger != nil {
			frs.logger.Info("Switching to fallback plugin", map[string]interface{}{
				"original_plugin": pluginID,
				"fallback_plugin": fallbackID,
			})
		}
		
		// 这里应该实现切换到降级插件的逻辑
		// 由于需要与插件管理器集成，这里只是记录日志
		
		// 发送降级事件
		if frs.eventBus != nil {
			frs.eventBus.Publish("plugin_fallback", map[string]interface{}{
				"original_plugin": pluginID,
				"fallback_plugin": fallbackID,
				"error":           err.Error(),
				"timestamp":       time.Now(),
			})
		}
		
		return nil
	}
	
	// 如果没有降级插件，启用降级模式
	if frs.degradationConfig != nil {
		if frs.logger != nil {
			frs.logger.Info("Enabling degradation mode", map[string]interface{}{
				"plugin_id": pluginID,
			})
		}
		
		// 发送降级模式事件
		if frs.eventBus != nil {
			frs.eventBus.Publish("degradation_mode_enabled", map[string]interface{}{
				"plugin_id":          pluginID,
				"disabled_features":  frs.degradationConfig.DisableFeatures,
				"reduce_quality":     frs.degradationConfig.ReduceQuality,
				"fallback_mode":      frs.degradationConfig.FallbackMode,
				"timestamp":          time.Now(),
			})
		}
		
		return nil
	}
	
	return fmt.Errorf("no fallback strategy available for plugin %s", pluginID)
}

// RecoveryManager 恢复管理器
type RecoveryManager struct {
	strategies []RecoveryStrategy
	mutex      sync.RWMutex
	logger     Logger
	metrics    MetricsCollector
	eventBus   EventBus
}

// NewRecoveryManager 创建恢复管理器
func NewRecoveryManager(logger Logger, metrics MetricsCollector, eventBus EventBus) *RecoveryManager {
	return &RecoveryManager{
		strategies: make([]RecoveryStrategy, 0),
		logger:     logger,
		metrics:    metrics,
		eventBus:   eventBus,
	}
}

// AddStrategy 添加恢复策略
func (rm *RecoveryManager) AddStrategy(strategy RecoveryStrategy) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	// 按优先级插入策略
	inserted := false
	for i, existing := range rm.strategies {
		if strategy.GetPriority() < existing.GetPriority() {
			rm.strategies = append(rm.strategies[:i], append([]RecoveryStrategy{strategy}, rm.strategies[i:]...)...)
			inserted = true
			break
		}
	}
	
	if !inserted {
		rm.strategies = append(rm.strategies, strategy)
	}
}

// RemoveStrategy 移除恢复策略
func (rm *RecoveryManager) RemoveStrategy(recoveryType RecoveryType) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	for i, strategy := range rm.strategies {
		if strategy.GetRecoveryType() == recoveryType {
			rm.strategies = append(rm.strategies[:i], rm.strategies[i+1:]...)
			break
		}
	}
}

// Recover 执行恢复操作
func (rm *RecoveryManager) Recover(ctx context.Context, pluginID string, err error) error {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	if rm.logger != nil {
		rm.logger.Info("Starting recovery process", map[string]interface{}{
			"plugin_id": pluginID,
			"error": err,
		})
	}
	
	// 尝试每个恢复策略
	for _, strategy := range rm.strategies {
		if !strategy.CanRecover(err) {
			continue
		}
		
		if rm.logger != nil {
			rm.logger.Info("Trying recovery strategy", map[string]interface{}{
				"plugin_id": pluginID,
				"strategy": strategy.GetRecoveryType().String(),
			})
		}
		
		// 创建带超时的上下文
		recoveryCtx, cancel := context.WithTimeout(ctx, strategy.GetRecoveryTimeout())
		defer cancel()
		
		// 尝试恢复
		if recoveryErr := strategy.Recover(recoveryCtx, pluginID, err); recoveryErr == nil {
			if rm.logger != nil {
				rm.logger.Info("Recovery successful", map[string]interface{}{
					"plugin_id": pluginID,
					"strategy": strategy.GetRecoveryType().String(),
				})
			}
			return nil
		} else {
			if rm.logger != nil {
				rm.logger.Warn("Recovery strategy failed", map[string]interface{}{
					"plugin_id": pluginID,
					"strategy": strategy.GetRecoveryType().String(),
					"error": recoveryErr,
				})
			}
		}
	}
	
	// 所有恢复策略都失败了
	recoveryErr := fmt.Errorf("all recovery strategies failed for plugin %s", pluginID)
	
	if rm.logger != nil {
		rm.logger.Error("Recovery process failed", map[string]interface{}{
			"plugin_id": pluginID,
			"error": recoveryErr,
		})
	}
	
	return recoveryErr
}

// GetStrategies 获取所有恢复策略
func (rm *RecoveryManager) GetStrategies() []RecoveryStrategy {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	strategies := make([]RecoveryStrategy, len(rm.strategies))
	copy(strategies, rm.strategies)
	return strategies
}