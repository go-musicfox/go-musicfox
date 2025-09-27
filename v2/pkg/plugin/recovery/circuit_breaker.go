// Package recovery 提供插件错误恢复策略实现
package recovery

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// CircuitBreakerState 熔断器状态
type CircuitBreakerState int

const (
	// StateClosed 关闭状态 - 正常工作
	StateClosed CircuitBreakerState = iota
	// StateOpen 开启状态 - 熔断开启，拒绝请求
	StateOpen
	// StateHalfOpen 半开状态 - 允许少量请求测试服务是否恢复
	StateHalfOpen
)

// String 返回状态的字符串表示
func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	// FailureThreshold 失败阈值，达到此值时开启熔断
	FailureThreshold int `json:"failure_threshold"`
	// SuccessThreshold 成功阈值，半开状态下连续成功此次数后关闭熔断
	SuccessThreshold int `json:"success_threshold"`
	// Timeout 请求超时时间
	Timeout time.Duration `json:"timeout"`
	// RecoveryTimeout 恢复超时时间，开启状态下等待此时间后转为半开状态
	RecoveryTimeout time.Duration `json:"recovery_timeout"`
	// MaxRequests 半开状态下允许的最大请求数
	MaxRequests int `json:"max_requests"`
	// MonitoringWindow 监控窗口时间
	MonitoringWindow time.Duration `json:"monitoring_window"`
}

// DefaultCircuitBreakerConfig 返回默认熔断器配置
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		FailureThreshold: 5,
		SuccessThreshold: 3,
		Timeout:          30 * time.Second,
		RecoveryTimeout:  60 * time.Second,
		MaxRequests:      10,
		MonitoringWindow: 5 * time.Minute,
	}
}

// CircuitBreakerMetrics 熔断器指标
type CircuitBreakerMetrics struct {
	TotalRequests   int64     `json:"total_requests"`
	SuccessRequests int64     `json:"success_requests"`
	FailureRequests int64     `json:"failure_requests"`
	RejectedRequests int64    `json:"rejected_requests"`
	LastFailureTime time.Time `json:"last_failure_time"`
	LastSuccessTime time.Time `json:"last_success_time"`
	StateChanges    int64     `json:"state_changes"`
}

// CircuitBreaker 熔断器实现
type CircuitBreaker struct {
	name     string
	config   *CircuitBreakerConfig
	logger   *slog.Logger
	mutex    sync.RWMutex

	// 状态管理
	state            CircuitBreakerState
	lastStateChange  time.Time

	// 计数器
	failureCount     int
	successCount     int
	requestCount     int

	// 时间记录
	lastFailureTime  time.Time
	lastSuccessTime  time.Time

	// 指标
	metrics          *CircuitBreakerMetrics

	// 回调函数
	onStateChange    func(name string, from, to CircuitBreakerState)
	onRequestRejected func(name string)
}

// NewCircuitBreaker 创建新的熔断器
func NewCircuitBreaker(name string, config *CircuitBreakerConfig, logger *slog.Logger) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}

	return &CircuitBreaker{
		name:            name,
		config:          config,
		logger:          logger,
		state:           StateClosed,
		lastStateChange: time.Now(),
		metrics: &CircuitBreakerMetrics{
			LastFailureTime: time.Time{},
			LastSuccessTime: time.Now(),
		},
	}
}

// Execute 执行操作，如果熔断器开启则拒绝执行
func (cb *CircuitBreaker) Execute(ctx context.Context, operation func(ctx context.Context) error) error {
	if !cb.canExecute() {
		cb.recordRejection()
		return fmt.Errorf("circuit breaker '%s' is open, request rejected", cb.name)
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, cb.config.Timeout)
	defer cancel()

	// 执行操作
	err := operation(ctx)

	if err != nil {
		cb.recordFailure()
		return fmt.Errorf("operation failed in circuit breaker '%s': %w", cb.name, err)
	}

	cb.recordSuccess()
	return nil
}

// canExecute 检查是否可以执行操作
func (cb *CircuitBreaker) canExecute() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// 检查是否可以转为半开状态
		if time.Since(cb.lastStateChange) >= cb.config.RecoveryTimeout {
			cb.mutex.RUnlock()
			cb.transitionToHalfOpen()
			cb.mutex.RLock()
			return true
		}
		return false
	case StateHalfOpen:
		// 半开状态下限制请求数量
		return cb.requestCount < cb.config.MaxRequests
	default:
		return false
	}
}

// recordSuccess 记录成功
func (cb *CircuitBreaker) recordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.metrics.TotalRequests++
	cb.metrics.SuccessRequests++
	cb.metrics.LastSuccessTime = time.Now()
	cb.lastSuccessTime = time.Now()

	switch cb.state {
	case StateClosed:
		// 关闭状态下重置失败计数
		cb.failureCount = 0
	case StateHalfOpen:
		// 半开状态下增加成功计数
		cb.successCount++
		if cb.successCount >= cb.config.SuccessThreshold {
			cb.transitionToClosed()
		}
	}

	if cb.logger != nil {
		cb.logger.Debug("Circuit breaker recorded success",
			"name", cb.name,
			"state", cb.state.String(),
			"success_count", cb.successCount,
			"failure_count", cb.failureCount)
	}
}

// recordFailure 记录失败
func (cb *CircuitBreaker) recordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.metrics.TotalRequests++
	cb.metrics.FailureRequests++
	cb.metrics.LastFailureTime = time.Now()
	cb.lastFailureTime = time.Now()
	cb.failureCount++

	switch cb.state {
	case StateClosed:
		// 关闭状态下检查是否需要开启熔断
		if cb.failureCount >= cb.config.FailureThreshold {
			cb.transitionToOpen()
		}
	case StateHalfOpen:
		// 半开状态下立即开启熔断
		cb.transitionToOpen()
	}

	if cb.logger != nil {
		cb.logger.Warn("Circuit breaker recorded failure",
			"name", cb.name,
			"state", cb.state.String(),
			"failure_count", cb.failureCount,
			"threshold", cb.config.FailureThreshold)
	}
}

// recordRejection 记录拒绝
func (cb *CircuitBreaker) recordRejection() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.metrics.RejectedRequests++

	if cb.onRequestRejected != nil {
		cb.onRequestRejected(cb.name)
	}

	if cb.logger != nil {
		cb.logger.Debug("Circuit breaker rejected request",
			"name", cb.name,
			"state", cb.state.String(),
			"rejected_count", cb.metrics.RejectedRequests)
	}
}

// transitionToClosed 转换到关闭状态
func (cb *CircuitBreaker) transitionToClosed() {
	oldState := cb.state
	cb.state = StateClosed
	cb.lastStateChange = time.Now()
	cb.failureCount = 0
	cb.successCount = 0
	cb.requestCount = 0
	cb.metrics.StateChanges++

	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, oldState, StateClosed)
	}

	if cb.logger != nil {
		cb.logger.Info("Circuit breaker transitioned to closed",
			"name", cb.name,
			"from_state", oldState.String())
	}
}

// transitionToOpen 转换到开启状态
func (cb *CircuitBreaker) transitionToOpen() {
	oldState := cb.state
	cb.state = StateOpen
	cb.lastStateChange = time.Now()
	cb.successCount = 0
	cb.requestCount = 0
	cb.metrics.StateChanges++

	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, oldState, StateOpen)
	}

	if cb.logger != nil {
		cb.logger.Error("Circuit breaker transitioned to open",
			"name", cb.name,
			"from_state", oldState.String(),
			"failure_count", cb.failureCount,
			"threshold", cb.config.FailureThreshold)
	}
}

// transitionToHalfOpen 转换到半开状态
func (cb *CircuitBreaker) transitionToHalfOpen() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	oldState := cb.state
	cb.state = StateHalfOpen
	cb.lastStateChange = time.Now()
	cb.failureCount = 0
	cb.successCount = 0
	cb.requestCount = 0
	cb.metrics.StateChanges++

	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, oldState, StateHalfOpen)
	}

	if cb.logger != nil {
		cb.logger.Info("Circuit breaker transitioned to half-open",
			"name", cb.name,
			"from_state", oldState.String())
	}
}

// GetState 获取当前状态
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// GetMetrics 获取指标
func (cb *CircuitBreaker) GetMetrics() *CircuitBreakerMetrics {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	// 返回指标的副本
	metrics := *cb.metrics
	return &metrics
}

// Reset 重置熔断器
func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	oldState := cb.state
	cb.state = StateClosed
	cb.lastStateChange = time.Now()
	cb.failureCount = 0
	cb.successCount = 0
	cb.requestCount = 0
	cb.lastFailureTime = time.Time{}
	cb.lastSuccessTime = time.Now()

	// 重置指标
	cb.metrics = &CircuitBreakerMetrics{
		LastSuccessTime: time.Now(),
	}

	if cb.logger != nil {
		cb.logger.Info("Circuit breaker reset",
			"name", cb.name,
			"from_state", oldState.String())
	}
}

// SetOnStateChange 设置状态变化回调
func (cb *CircuitBreaker) SetOnStateChange(callback func(name string, from, to CircuitBreakerState)) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.onStateChange = callback
}

// SetOnRequestRejected 设置请求拒绝回调
func (cb *CircuitBreaker) SetOnRequestRejected(callback func(name string)) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.onRequestRejected = callback
}

// IsHealthy 检查熔断器是否健康
func (cb *CircuitBreaker) IsHealthy() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state == StateClosed
}

// GetFailureRate 获取失败率
func (cb *CircuitBreaker) GetFailureRate() float64 {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	if cb.metrics.TotalRequests == 0 {
		return 0.0
	}

	return float64(cb.metrics.FailureRequests) / float64(cb.metrics.TotalRequests)
}

// GetName 获取熔断器名称
func (cb *CircuitBreaker) GetName() string {
	return cb.name
}

// GetConfig 获取配置
func (cb *CircuitBreaker) GetConfig() *CircuitBreakerConfig {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	// 返回配置的副本
	config := *cb.config
	return &config
}