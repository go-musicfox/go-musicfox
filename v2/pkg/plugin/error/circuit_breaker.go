package plugin

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// CircuitBreakerState 熔断器状态
type CircuitBreakerState int

const (
	CircuitBreakerStateClosed CircuitBreakerState = iota
	CircuitBreakerStateOpen
	CircuitBreakerStateHalfOpen
)

// String 返回熔断器状态的字符串表示
func (s CircuitBreakerState) String() string {
	switch s {
	case CircuitBreakerStateClosed:
		return "closed"
	case CircuitBreakerStateOpen:
		return "open"
	case CircuitBreakerStateHalfOpen:
		return "half_open"
	default:
		return "unknown"
	}
}

// CircuitBreaker 熔断器接口
type CircuitBreaker interface {
	// Execute 执行操作
	Execute(ctx context.Context, operation func() error) error
	// GetState 获取当前状态
	GetState() CircuitBreakerState
	// GetStats 获取统计信息
	GetStats() *CircuitBreakerStats
	// Reset 重置熔断器
	Reset()
	// ForceOpen 强制打开熔断器
	ForceOpen()
	// ForceClose 强制关闭熔断器
	ForceClose()
}

// CircuitBreakerStats 熔断器统计信息
type CircuitBreakerStats struct {
	State              CircuitBreakerState `json:"state"`               // 当前状态
	TotalRequests      int64               `json:"total_requests"`      // 总请求数
	SuccessfulRequests int64               `json:"successful_requests"` // 成功请求数
	FailedRequests     int64               `json:"failed_requests"`     // 失败请求数
	ConsecutiveFailures int64              `json:"consecutive_failures"` // 连续失败数
	LastFailureTime    *time.Time          `json:"last_failure_time"`   // 最后失败时间
	LastSuccessTime    *time.Time          `json:"last_success_time"`   // 最后成功时间
	StateChangedAt     time.Time           `json:"state_changed_at"`    // 状态改变时间
	FailureRate        float64             `json:"failure_rate"`        // 失败率
}

// SmartCircuitBreaker 智能熔断器实现
type SmartCircuitBreaker struct {
	config              *CircuitBreakerConfig
	state               CircuitBreakerState
	totalRequests       int64
	successfulRequests  int64
	failedRequests      int64
	consecutiveFailures int64
	lastFailureTime     *time.Time
	lastSuccessTime     *time.Time
	stateChangedAt      time.Time
	mutex               sync.RWMutex
	logger              Logger
	metrics             MetricsCollector
	eventBus            EventBus
}

// NewSmartCircuitBreaker 创建智能熔断器
func NewSmartCircuitBreaker(config *CircuitBreakerConfig, logger Logger, metrics MetricsCollector, eventBus EventBus) *SmartCircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}
	
	return &SmartCircuitBreaker{
		config:         config,
		state:          CircuitBreakerStateClosed,
		stateChangedAt: time.Now(),
		logger:         logger,
		metrics:        metrics,
		eventBus:       eventBus,
	}
}

// DefaultCircuitBreakerConfig 默认熔断器配置
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		FailureThreshold: 5,
		SuccessThreshold: 3,
		Timeout:          60 * time.Second,
		ResetTimeout:     30 * time.Second,
		MaxRequests:      10,
	}
}

// Execute 执行操作
func (cb *SmartCircuitBreaker) Execute(ctx context.Context, operation func() error) error {
	// 检查是否允许执行
	if !cb.allowRequest() {
		cb.recordMetrics("circuit_breaker_rejected")
		return NewPluginError(ErrorCodeUnavailable, "circuit breaker is open")
	}
	
	// 执行操作
	start := time.Now()
	err := operation()
	duration := time.Since(start)
	
	// 记录结果
	cb.recordResult(err, duration)
	
	return err
}

// allowRequest 检查是否允许请求
func (cb *SmartCircuitBreaker) allowRequest() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	now := time.Now()
	
	switch cb.state {
	case CircuitBreakerStateClosed:
		return true
		
	case CircuitBreakerStateOpen:
		// 检查是否可以转换到半开状态
		if now.Sub(cb.stateChangedAt) >= cb.config.ResetTimeout {
			cb.setState(CircuitBreakerStateHalfOpen, now)
			return true
		}
		return false
		
	case CircuitBreakerStateHalfOpen:
		// 半开状态下限制请求数量
		return cb.totalRequests-cb.getRequestsAtStateChange() < int64(cb.config.MaxRequests)
		
	default:
		return false
	}
}

// recordResult 记录操作结果
func (cb *SmartCircuitBreaker) recordResult(err error, duration time.Duration) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	now := time.Now()
	atomic.AddInt64(&cb.totalRequests, 1)
	
	if err == nil {
		// 操作成功
		atomic.AddInt64(&cb.successfulRequests, 1)
		atomic.StoreInt64(&cb.consecutiveFailures, 0)
		cb.lastSuccessTime = &now
		
		// 记录成功指标
		cb.recordMetrics("circuit_breaker_success")
		if cb.metrics != nil {
			cb.metrics.RecordTimer("circuit_breaker_duration", duration, map[string]string{
				"result": "success",
				"state":  cb.state.String(),
			})
		}
		
		// 检查是否可以关闭熔断器
		if cb.state == CircuitBreakerStateHalfOpen {
			successCount := cb.successfulRequests - cb.getSuccessfulRequestsAtStateChange()
			if successCount >= int64(cb.config.SuccessThreshold) {
				cb.setState(CircuitBreakerStateClosed, now)
			}
		}
	} else {
		// 操作失败
		atomic.AddInt64(&cb.failedRequests, 1)
		atomic.AddInt64(&cb.consecutiveFailures, 1)
		cb.lastFailureTime = &now
		
		// 记录失败指标
		cb.recordMetrics("circuit_breaker_failure")
		if cb.metrics != nil {
			cb.metrics.RecordTimer("circuit_breaker_duration", duration, map[string]string{
				"result": "failure",
				"state":  cb.state.String(),
			})
		}
		
		// 检查是否需要打开熔断器
		if cb.state == CircuitBreakerStateClosed || cb.state == CircuitBreakerStateHalfOpen {
			if cb.consecutiveFailures >= int64(cb.config.FailureThreshold) {
				cb.setState(CircuitBreakerStateOpen, now)
			}
		}
	}
}

// setState 设置熔断器状态
func (cb *SmartCircuitBreaker) setState(newState CircuitBreakerState, now time.Time) {
	oldState := cb.state
	cb.state = newState
	cb.stateChangedAt = now
	
	// 记录状态变化
	if cb.logger != nil {
		cb.logger.Info("Circuit breaker state changed", map[string]interface{}{
			"old_state": oldState.String(),
			"new_state": newState.String(),
			"consecutive_failures": cb.consecutiveFailures,
		})
	}
	
	// 发送状态变化事件
	if cb.eventBus != nil {
		cb.eventBus.Publish("circuit_breaker_state_changed", map[string]interface{}{
			"old_state":            oldState.String(),
			"new_state":            newState.String(),
			"consecutive_failures": cb.consecutiveFailures,
			"timestamp":            now,
		})
	}
	
	// 记录状态变化指标
	if cb.metrics != nil {
		cb.metrics.IncrementCounter("circuit_breaker_state_changes", map[string]string{
			"from_state": oldState.String(),
			"to_state":   newState.String(),
		})
	}
}

// GetState 获取当前状态
func (cb *SmartCircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// GetStats 获取统计信息
func (cb *SmartCircuitBreaker) GetStats() *CircuitBreakerStats {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	
	totalReq := atomic.LoadInt64(&cb.totalRequests)
	successReq := atomic.LoadInt64(&cb.successfulRequests)
	failedReq := atomic.LoadInt64(&cb.failedRequests)
	
	var failureRate float64
	if totalReq > 0 {
		failureRate = float64(failedReq) / float64(totalReq)
	}
	
	return &CircuitBreakerStats{
		State:               cb.state,
		TotalRequests:       totalReq,
		SuccessfulRequests:  successReq,
		FailedRequests:      failedReq,
		ConsecutiveFailures: atomic.LoadInt64(&cb.consecutiveFailures),
		LastFailureTime:     cb.lastFailureTime,
		LastSuccessTime:     cb.lastSuccessTime,
		StateChangedAt:      cb.stateChangedAt,
		FailureRate:         failureRate,
	}
}

// Reset 重置熔断器
func (cb *SmartCircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	atomic.StoreInt64(&cb.totalRequests, 0)
	atomic.StoreInt64(&cb.successfulRequests, 0)
	atomic.StoreInt64(&cb.failedRequests, 0)
	atomic.StoreInt64(&cb.consecutiveFailures, 0)
	cb.lastFailureTime = nil
	cb.lastSuccessTime = nil
	cb.setState(CircuitBreakerStateClosed, time.Now())
	
	if cb.logger != nil {
		cb.logger.Info("Circuit breaker reset", map[string]interface{}{})
	}
}

// ForceOpen 强制打开熔断器
func (cb *SmartCircuitBreaker) ForceOpen() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	cb.setState(CircuitBreakerStateOpen, time.Now())
	
	if cb.logger != nil {
		cb.logger.Warn("Circuit breaker forced open", map[string]interface{}{})
	}
}

// ForceClose 强制关闭熔断器
func (cb *SmartCircuitBreaker) ForceClose() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	atomic.StoreInt64(&cb.consecutiveFailures, 0)
	cb.setState(CircuitBreakerStateClosed, time.Now())
	
	if cb.logger != nil {
		cb.logger.Info("Circuit breaker forced closed", map[string]interface{}{})
	}
}

// recordMetrics 记录指标
func (cb *SmartCircuitBreaker) recordMetrics(metricName string) {
	if cb.metrics != nil {
		cb.metrics.IncrementCounter(metricName, map[string]string{
			"state": cb.state.String(),
		})
	}
}

// getRequestsAtStateChange 获取状态改变时的请求数
func (cb *SmartCircuitBreaker) getRequestsAtStateChange() int64 {
	// 这里可以实现更复杂的逻辑来跟踪状态改变时的请求数
	// 为简化实现，这里返回当前总请求数
	return atomic.LoadInt64(&cb.totalRequests)
}

// getSuccessfulRequestsAtStateChange 获取状态改变时的成功请求数
func (cb *SmartCircuitBreaker) getSuccessfulRequestsAtStateChange() int64 {
	// 这里可以实现更复杂的逻辑来跟踪状态改变时的成功请求数
	// 为简化实现，这里返回当前成功请求数
	return atomic.LoadInt64(&cb.successfulRequests)
}

// CircuitBreakerManager 熔断器管理器
type CircuitBreakerManager struct {
	circuitBreakers map[string]CircuitBreaker
	mutex           sync.RWMutex
	logger          Logger
	metrics         MetricsCollector
	eventBus        EventBus
}

// NewCircuitBreakerManager 创建熔断器管理器
func NewCircuitBreakerManager(logger Logger, metrics MetricsCollector, eventBus EventBus) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		circuitBreakers: make(map[string]CircuitBreaker),
		logger:          logger,
		metrics:         metrics,
		eventBus:        eventBus,
	}
}

// GetOrCreateCircuitBreaker 获取或创建熔断器
func (cbm *CircuitBreakerManager) GetOrCreateCircuitBreaker(name string, config *CircuitBreakerConfig) CircuitBreaker {
	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()
	
	if cb, exists := cbm.circuitBreakers[name]; exists {
		return cb
	}
	
	cb := NewSmartCircuitBreaker(config, cbm.logger, cbm.metrics, cbm.eventBus)
	cbm.circuitBreakers[name] = cb
	return cb
}

// GetCircuitBreaker 获取熔断器
func (cbm *CircuitBreakerManager) GetCircuitBreaker(name string) (CircuitBreaker, bool) {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()
	
	cb, exists := cbm.circuitBreakers[name]
	return cb, exists
}

// RemoveCircuitBreaker 移除熔断器
func (cbm *CircuitBreakerManager) RemoveCircuitBreaker(name string) {
	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()
	
	delete(cbm.circuitBreakers, name)
}

// GetAllStats 获取所有熔断器统计信息
func (cbm *CircuitBreakerManager) GetAllStats() map[string]*CircuitBreakerStats {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()
	
	stats := make(map[string]*CircuitBreakerStats)
	for name, cb := range cbm.circuitBreakers {
		stats[name] = cb.GetStats()
	}
	return stats
}