package recovery

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// RecoveryAction 恢复动作类型
type RecoveryAction int

const (
	// RecoveryActionRestart 重启插件
	RecoveryActionRestart RecoveryAction = iota
	// RecoveryActionReload 重新加载插件
	RecoveryActionReload
	// RecoveryActionFailover 故障转移
	RecoveryActionFailover
	// RecoveryActionReset 重置插件状态
	RecoveryActionReset
	// RecoveryActionCustom 自定义恢复动作
	RecoveryActionCustom
)

// String 返回恢复动作的字符串表示
func (ra RecoveryAction) String() string {
	switch ra {
	case RecoveryActionRestart:
		return "restart"
	case RecoveryActionReload:
		return "reload"
	case RecoveryActionFailover:
		return "failover"
	case RecoveryActionReset:
		return "reset"
	case RecoveryActionCustom:
		return "custom"
	default:
		return "unknown"
	}
}

// HealthCheckStatus 健康检查状态
type HealthCheckStatus int

const (
	// HealthStatusHealthy 健康
	HealthStatusHealthy HealthCheckStatus = iota
	// HealthStatusUnhealthy 不健康
	HealthStatusUnhealthy
	// HealthStatusDegraded 降级
	HealthStatusDegraded
	// HealthStatusUnknown 未知
	HealthStatusUnknown
)

// String 返回健康状态的字符串表示
func (hs HealthCheckStatus) String() string {
	switch hs {
	case HealthStatusHealthy:
		return "healthy"
	case HealthStatusUnhealthy:
		return "unhealthy"
	case HealthStatusDegraded:
		return "degraded"
	case HealthStatusUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

// AutoRecoveryConfig 自动恢复配置
type AutoRecoveryConfig struct {
	// Enabled 是否启用自动恢复
	Enabled bool `json:"enabled"`
	// HealthCheckInterval 健康检查间隔
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	// HealthCheckTimeout 健康检查超时时间
	HealthCheckTimeout time.Duration `json:"health_check_timeout"`
	// MaxRecoveryAttempts 最大恢复尝试次数
	MaxRecoveryAttempts int `json:"max_recovery_attempts"`
	// RecoveryDelay 恢复延迟时间
	RecoveryDelay time.Duration `json:"recovery_delay"`
	// FailureThreshold 失败阈值
	FailureThreshold int `json:"failure_threshold"`
	// RecoveryActions 恢复动作列表（按优先级排序）
	RecoveryActions []RecoveryAction `json:"recovery_actions"`
	// FailoverTargets 故障转移目标
	FailoverTargets []string `json:"failover_targets"`
	// CustomRecoveryFunc 自定义恢复函数
	CustomRecoveryFunc func(ctx context.Context, pluginID string) error `json:"-"`
}

// DefaultAutoRecoveryConfig 返回默认自动恢复配置
func DefaultAutoRecoveryConfig() *AutoRecoveryConfig {
	return &AutoRecoveryConfig{
		Enabled:             true,
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  10 * time.Second,
		MaxRecoveryAttempts: 3,
		RecoveryDelay:       5 * time.Second,
		FailureThreshold:    3,
		RecoveryActions: []RecoveryAction{
			RecoveryActionReset,
			RecoveryActionRestart,
			RecoveryActionReload,
			RecoveryActionFailover,
		},
		FailoverTargets: []string{},
	}
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	PluginID    string            `json:"plugin_id"`
	Status      HealthCheckStatus `json:"status"`
	Message     string            `json:"message"`
	Details     map[string]interface{} `json:"details"`
	CheckTime   time.Time         `json:"check_time"`
	Latency     time.Duration     `json:"latency"`
	Error       error             `json:"error,omitempty"`
}

// RecoveryAttempt 恢复尝试记录
type RecoveryAttempt struct {
	PluginID     string         `json:"plugin_id"`
	Action       RecoveryAction `json:"action"`
	AttemptTime  time.Time      `json:"attempt_time"`
	Success      bool           `json:"success"`
	Error        error          `json:"error,omitempty"`
	Duration     time.Duration  `json:"duration"`
}

// AutoRecoveryMetrics 自动恢复指标
type AutoRecoveryMetrics struct {
	TotalHealthChecks    int64     `json:"total_health_checks"`
	FailedHealthChecks   int64     `json:"failed_health_checks"`
	TotalRecoveryAttempts int64    `json:"total_recovery_attempts"`
	SuccessfulRecoveries int64     `json:"successful_recoveries"`
	FailedRecoveries     int64     `json:"failed_recoveries"`
	AverageRecoveryTime  time.Duration `json:"average_recovery_time"`
	LastHealthCheckTime  time.Time `json:"last_health_check_time"`
	LastRecoveryTime     time.Time `json:"last_recovery_time"`
}

// PluginHealthState 插件健康状态
type PluginHealthState struct {
	PluginID         string               `json:"plugin_id"`
	CurrentStatus    HealthCheckStatus    `json:"current_status"`
	LastHealthCheck  *HealthCheckResult   `json:"last_health_check"`
	FailureCount     int                  `json:"failure_count"`
	RecoveryAttempts int                  `json:"recovery_attempts"`
	LastRecovery     *RecoveryAttempt     `json:"last_recovery"`
	IsRecovering     bool                 `json:"is_recovering"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
}

// AutoRecoveryManager 自动恢复管理器
type AutoRecoveryManager struct {
	config       *AutoRecoveryConfig
	logger       *slog.Logger
	metrics      *AutoRecoveryMetrics
	mutex        sync.RWMutex

	// 插件健康状态
	pluginStates map[string]*PluginHealthState
	statesMutex  sync.RWMutex

	// 健康检查函数
	healthCheckFunc func(ctx context.Context, pluginID string) (*HealthCheckResult, error)

	// 恢复动作函数
	recoveryFuncs map[RecoveryAction]func(ctx context.Context, pluginID string) error

	// 控制通道
	stopChan chan struct{}
	done     chan struct{}
	started  bool

	// 回调函数
	onHealthCheckFailed func(pluginID string, result *HealthCheckResult)
	onRecoveryStarted   func(pluginID string, action RecoveryAction)
	onRecoveryCompleted func(pluginID string, attempt *RecoveryAttempt)
}

// NewAutoRecoveryManager 创建新的自动恢复管理器
func NewAutoRecoveryManager(config *AutoRecoveryConfig, logger *slog.Logger) *AutoRecoveryManager {
	if config == nil {
		config = DefaultAutoRecoveryConfig()
	}

	arm := &AutoRecoveryManager{
		config:       config,
		logger:       logger,
		metrics:      &AutoRecoveryMetrics{},
		pluginStates: make(map[string]*PluginHealthState),
		recoveryFuncs: make(map[RecoveryAction]func(ctx context.Context, pluginID string) error),
		stopChan:     make(chan struct{}),
		done:         make(chan struct{}),
	}

	// 注册默认恢复函数
	arm.registerDefaultRecoveryFuncs()

	return arm
}

// registerDefaultRecoveryFuncs 注册默认恢复函数
func (arm *AutoRecoveryManager) registerDefaultRecoveryFuncs() {
	// 重置恢复函数
	arm.recoveryFuncs[RecoveryActionReset] = func(ctx context.Context, pluginID string) error {
		// 这里应该调用插件管理器的重置方法
		if arm.logger != nil {
			arm.logger.Info("Executing reset recovery action", "plugin_id", pluginID)
		}
		return nil // 实际实现需要调用插件管理器
	}

	// 重启恢复函数
	arm.recoveryFuncs[RecoveryActionRestart] = func(ctx context.Context, pluginID string) error {
		// 这里应该调用插件管理器的重启方法
		if arm.logger != nil {
			arm.logger.Info("Executing restart recovery action", "plugin_id", pluginID)
		}
		return nil // 实际实现需要调用插件管理器
	}

	// 重新加载恢复函数
	arm.recoveryFuncs[RecoveryActionReload] = func(ctx context.Context, pluginID string) error {
		// 这里应该调用插件管理器的重新加载方法
		if arm.logger != nil {
			arm.logger.Info("Executing reload recovery action", "plugin_id", pluginID)
		}
		return nil // 实际实现需要调用插件管理器
	}

	// 故障转移恢复函数
	arm.recoveryFuncs[RecoveryActionFailover] = func(ctx context.Context, pluginID string) error {
		// 这里应该实现故障转移逻辑
		if arm.logger != nil {
			arm.logger.Info("Executing failover recovery action", "plugin_id", pluginID)
		}
		return nil // 实际实现需要故障转移逻辑
	}
}

// Start 启动自动恢复管理器
func (arm *AutoRecoveryManager) Start(ctx context.Context) error {
	arm.mutex.Lock()
	defer arm.mutex.Unlock()

	if arm.started {
		if arm.logger != nil {
			arm.logger.Warn("Auto recovery manager is already started")
		}
		return nil
	}

	if !arm.config.Enabled {
		if arm.logger != nil {
			arm.logger.Info("Auto recovery is disabled")
		}
		return nil
	}

	if arm.logger != nil {
		arm.logger.Info("Starting auto recovery manager",
			"health_check_interval", arm.config.HealthCheckInterval,
			"max_recovery_attempts", arm.config.MaxRecoveryAttempts)
	}

	arm.started = true

	// 启动健康检查循环
	go arm.healthCheckLoop(ctx)

	return nil
}

// Stop 停止自动恢复管理器
func (arm *AutoRecoveryManager) Stop() error {
	arm.mutex.Lock()
	defer arm.mutex.Unlock()

	if !arm.started {
		return nil
	}

	select {
	case <-arm.stopChan:
		// Already stopped
		return nil
	default:
		close(arm.stopChan)
	}

	// 等待健康检查循环结束，但设置超时避免永久阻塞
	select {
	case <-arm.done:
		// 正常结束
	case <-time.After(5 * time.Second):
		// 超时，强制结束
		if arm.logger != nil {
			arm.logger.Warn("Auto recovery manager stop timeout")
		}
	}

	arm.started = false

	if arm.logger != nil {
		arm.logger.Info("Auto recovery manager stopped")
	}

	return nil
}

// healthCheckLoop 健康检查循环
func (arm *AutoRecoveryManager) healthCheckLoop(ctx context.Context) {
	defer func() {
		select {
		case <-arm.done:
			// channel already closed
		default:
			close(arm.done)
		}
	}()

	ticker := time.NewTicker(arm.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-arm.stopChan:
			return
		case <-ticker.C:
			arm.performHealthChecks(ctx)
		}
	}
}

// performHealthChecks 执行健康检查
func (arm *AutoRecoveryManager) performHealthChecks(ctx context.Context) {
	arm.statesMutex.RLock()
	pluginIDs := make([]string, 0, len(arm.pluginStates))
	for pluginID := range arm.pluginStates {
		pluginIDs = append(pluginIDs, pluginID)
	}
	arm.statesMutex.RUnlock()

	for _, pluginID := range pluginIDs {
		arm.checkPluginHealth(ctx, pluginID)
	}
}

// checkPluginHealth 检查插件健康状态
func (arm *AutoRecoveryManager) checkPluginHealth(ctx context.Context, pluginID string) {
	if arm.healthCheckFunc == nil {
		return
	}

	// 创建带超时的上下文
	checkCtx, cancel := context.WithTimeout(ctx, arm.config.HealthCheckTimeout)
	defer cancel()

	// 执行健康检查
	result, err := arm.healthCheckFunc(checkCtx, pluginID)
	if err != nil {
		result = &HealthCheckResult{
			PluginID:  pluginID,
			Status:    HealthStatusUnknown,
			Message:   "Health check failed",
			CheckTime: time.Now(),
			Error:     err,
		}
	}

	// 更新指标
	arm.updateHealthCheckMetrics(result)

	// 更新插件状态
	arm.updatePluginState(pluginID, result)

	// 检查是否需要恢复
	if result.Status == HealthStatusUnhealthy {
		arm.triggerRecovery(ctx, pluginID)
	}
}

// updateHealthCheckMetrics 更新健康检查指标
func (arm *AutoRecoveryManager) updateHealthCheckMetrics(result *HealthCheckResult) {
	arm.mutex.Lock()
	defer arm.mutex.Unlock()

	arm.metrics.TotalHealthChecks++
	arm.metrics.LastHealthCheckTime = time.Now()

	if result.Status == HealthStatusUnhealthy || result.Error != nil {
		arm.metrics.FailedHealthChecks++
	}
}

// updatePluginState 更新插件状态
func (arm *AutoRecoveryManager) updatePluginState(pluginID string, result *HealthCheckResult) {
	arm.statesMutex.Lock()
	defer arm.statesMutex.Unlock()

	state, exists := arm.pluginStates[pluginID]
	if !exists {
		state = &PluginHealthState{
			PluginID:  pluginID,
			CreatedAt: time.Now(),
		}
		arm.pluginStates[pluginID] = state
	}

	state.LastHealthCheck = result
	state.CurrentStatus = result.Status
	state.UpdatedAt = time.Now()

	// 更新失败计数
	if result.Status == HealthStatusUnhealthy {
		state.FailureCount++
	} else if result.Status == HealthStatusHealthy {
		state.FailureCount = 0 // 重置失败计数
	}
}

// triggerRecovery 触发恢复
func (arm *AutoRecoveryManager) triggerRecovery(ctx context.Context, pluginID string) {
	arm.statesMutex.RLock()
	state, exists := arm.pluginStates[pluginID]
	arm.statesMutex.RUnlock()

	if !exists {
		return
	}

	// 检查是否已经在恢复中
	if state.IsRecovering {
		return
	}

	// 检查失败阈值
	if state.FailureCount < arm.config.FailureThreshold {
		return
	}

	// 检查最大恢复尝试次数
	if state.RecoveryAttempts >= arm.config.MaxRecoveryAttempts {
		if arm.logger != nil {
			arm.logger.Error("Max recovery attempts exceeded",
				"plugin_id", pluginID,
				"attempts", state.RecoveryAttempts,
				"max_attempts", arm.config.MaxRecoveryAttempts)
		}
		return
	}

	// 标记为恢复中
	arm.statesMutex.Lock()
	state.IsRecovering = true
	arm.statesMutex.Unlock()

	// 异步执行恢复
	go arm.executeRecovery(ctx, pluginID)
}

// executeRecovery 执行恢复
func (arm *AutoRecoveryManager) executeRecovery(ctx context.Context, pluginID string) {
	defer func() {
		// 恢复完成后重置恢复状态
		arm.statesMutex.Lock()
		if state, exists := arm.pluginStates[pluginID]; exists {
			state.IsRecovering = false
		}
		arm.statesMutex.Unlock()
	}()

	// 等待恢复延迟
	select {
	case <-ctx.Done():
		return
	case <-time.After(arm.config.RecoveryDelay):
	}

	// 按优先级尝试恢复动作
	for _, action := range arm.config.RecoveryActions {
		if arm.attemptRecovery(ctx, pluginID, action) {
			return // 恢复成功
		}
	}

	if arm.logger != nil {
		arm.logger.Error("All recovery actions failed", "plugin_id", pluginID)
	}
}

// attemptRecovery 尝试恢复
func (arm *AutoRecoveryManager) attemptRecovery(ctx context.Context, pluginID string, action RecoveryAction) bool {
	startTime := time.Now()

	// 触发恢复开始回调
	if arm.onRecoveryStarted != nil {
		arm.onRecoveryStarted(pluginID, action)
	}

	if arm.logger != nil {
		arm.logger.Info("Attempting recovery",
			"plugin_id", pluginID,
			"action", action.String())
	}

	// 获取恢复函数
	var recoveryFunc func(ctx context.Context, pluginID string) error
	if action == RecoveryActionCustom && arm.config.CustomRecoveryFunc != nil {
		recoveryFunc = arm.config.CustomRecoveryFunc
	} else {
		recoveryFunc = arm.recoveryFuncs[action]
	}

	if recoveryFunc == nil {
		if arm.logger != nil {
			arm.logger.Warn("No recovery function for action",
				"plugin_id", pluginID,
				"action", action.String())
		}
		return false
	}

	// 执行恢复
	err := recoveryFunc(ctx, pluginID)
	duration := time.Since(startTime)

	// 创建恢复尝试记录
	attempt := &RecoveryAttempt{
		PluginID:    pluginID,
		Action:      action,
		AttemptTime: startTime,
		Success:     err == nil,
		Error:       err,
		Duration:    duration,
	}

	// 更新状态和指标
	arm.updateRecoveryMetrics(attempt)
	arm.updatePluginRecoveryState(pluginID, attempt)

	// 触发恢复完成回调
	if arm.onRecoveryCompleted != nil {
		arm.onRecoveryCompleted(pluginID, attempt)
	}

	if err != nil {
		if arm.logger != nil {
			arm.logger.Error("Recovery attempt failed",
				"plugin_id", pluginID,
				"action", action.String(),
				"error", err.Error(),
				"duration", duration)
		}
		return false
	}

	if arm.logger != nil {
		arm.logger.Info("Recovery attempt succeeded",
			"plugin_id", pluginID,
			"action", action.String(),
			"duration", duration)
	}

	return true
}

// updateRecoveryMetrics 更新恢复指标
func (arm *AutoRecoveryManager) updateRecoveryMetrics(attempt *RecoveryAttempt) {
	arm.mutex.Lock()
	defer arm.mutex.Unlock()

	arm.metrics.TotalRecoveryAttempts++
	arm.metrics.LastRecoveryTime = time.Now()

	if attempt.Success {
		arm.metrics.SuccessfulRecoveries++
	} else {
		arm.metrics.FailedRecoveries++
	}

	// 更新平均恢复时间
	if arm.metrics.SuccessfulRecoveries > 0 {
		totalTime := time.Duration(arm.metrics.AverageRecoveryTime.Nanoseconds()*int64(arm.metrics.SuccessfulRecoveries-1)) + attempt.Duration
		arm.metrics.AverageRecoveryTime = totalTime / time.Duration(arm.metrics.SuccessfulRecoveries)
	}
}

// updatePluginRecoveryState 更新插件恢复状态
func (arm *AutoRecoveryManager) updatePluginRecoveryState(pluginID string, attempt *RecoveryAttempt) {
	arm.statesMutex.Lock()
	defer arm.statesMutex.Unlock()

	state, exists := arm.pluginStates[pluginID]
	if !exists {
		return
	}

	state.LastRecovery = attempt
	state.RecoveryAttempts++
	state.UpdatedAt = time.Now()

	if attempt.Success {
		// 恢复成功，重置失败计数和恢复尝试次数
		state.FailureCount = 0
		state.RecoveryAttempts = 0
	}
}

// RegisterPlugin 注册插件进行监控
func (arm *AutoRecoveryManager) RegisterPlugin(pluginID string) {
	arm.statesMutex.Lock()
	defer arm.statesMutex.Unlock()

	if _, exists := arm.pluginStates[pluginID]; !exists {
		arm.pluginStates[pluginID] = &PluginHealthState{
			PluginID:      pluginID,
			CurrentStatus: HealthStatusUnknown,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if arm.logger != nil {
			arm.logger.Info("Plugin registered for auto recovery", "plugin_id", pluginID)
		}
	}
}

// UnregisterPlugin 取消注册插件
func (arm *AutoRecoveryManager) UnregisterPlugin(pluginID string) {
	arm.statesMutex.Lock()
	defer arm.statesMutex.Unlock()

	delete(arm.pluginStates, pluginID)

	if arm.logger != nil {
		arm.logger.Info("Plugin unregistered from auto recovery", "plugin_id", pluginID)
	}
}

// SetHealthCheckFunc 设置健康检查函数
func (arm *AutoRecoveryManager) SetHealthCheckFunc(fn func(ctx context.Context, pluginID string) (*HealthCheckResult, error)) {
	arm.mutex.Lock()
	defer arm.mutex.Unlock()
	arm.healthCheckFunc = fn
}

// RegisterRecoveryFunc 注册恢复函数
func (arm *AutoRecoveryManager) RegisterRecoveryFunc(action RecoveryAction, fn func(ctx context.Context, pluginID string) error) {
	arm.mutex.Lock()
	defer arm.mutex.Unlock()
	arm.recoveryFuncs[action] = fn
}

// GetPluginState 获取插件状态
func (arm *AutoRecoveryManager) GetPluginState(pluginID string) (*PluginHealthState, bool) {
	arm.statesMutex.RLock()
	defer arm.statesMutex.RUnlock()

	state, exists := arm.pluginStates[pluginID]
	if !exists {
		return nil, false
	}

	// 返回状态的副本
	stateCopy := *state
	return &stateCopy, true
}

// GetAllPluginStates 获取所有插件状态
func (arm *AutoRecoveryManager) GetAllPluginStates() map[string]*PluginHealthState {
	arm.statesMutex.RLock()
	defer arm.statesMutex.RUnlock()

	states := make(map[string]*PluginHealthState)
	for pluginID, state := range arm.pluginStates {
		// 返回状态的副本
		stateCopy := *state
		states[pluginID] = &stateCopy
	}

	return states
}

// GetMetrics 获取指标
func (arm *AutoRecoveryManager) GetMetrics() *AutoRecoveryMetrics {
	arm.mutex.RLock()
	defer arm.mutex.RUnlock()

	// 返回指标的副本
	metrics := *arm.metrics
	return &metrics
}

// SetOnHealthCheckFailed 设置健康检查失败回调
func (arm *AutoRecoveryManager) SetOnHealthCheckFailed(callback func(pluginID string, result *HealthCheckResult)) {
	arm.mutex.Lock()
	defer arm.mutex.Unlock()
	arm.onHealthCheckFailed = callback
}

// SetOnRecoveryStarted 设置恢复开始回调
func (arm *AutoRecoveryManager) SetOnRecoveryStarted(callback func(pluginID string, action RecoveryAction)) {
	arm.mutex.Lock()
	defer arm.mutex.Unlock()
	arm.onRecoveryStarted = callback
}

// SetOnRecoveryCompleted 设置恢复完成回调
func (arm *AutoRecoveryManager) SetOnRecoveryCompleted(callback func(pluginID string, attempt *RecoveryAttempt)) {
	arm.mutex.Lock()
	defer arm.mutex.Unlock()
	arm.onRecoveryCompleted = callback
}