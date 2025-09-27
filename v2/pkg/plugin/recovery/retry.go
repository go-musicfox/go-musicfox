package recovery

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"sync"
	"time"
)

// RetryPolicy 重试策略类型
type RetryPolicy int

const (
	// RetryPolicyFixed 固定间隔重试
	RetryPolicyFixed RetryPolicy = iota
	// RetryPolicyLinear 线性增长重试
	RetryPolicyLinear
	// RetryPolicyExponential 指数退避重试
	RetryPolicyExponential
	// RetryPolicyCustom 自定义重试策略
	RetryPolicyCustom
)

// String 返回重试策略的字符串表示
func (rp RetryPolicy) String() string {
	switch rp {
	case RetryPolicyFixed:
		return "fixed"
	case RetryPolicyLinear:
		return "linear"
	case RetryPolicyExponential:
		return "exponential"
	case RetryPolicyCustom:
		return "custom"
	default:
		return "unknown"
	}
}

// RetryConfig 重试配置
type RetryConfig struct {
	// MaxRetries 最大重试次数
	MaxRetries int `json:"max_retries"`
	// InitialDelay 初始延迟时间
	InitialDelay time.Duration `json:"initial_delay"`
	// MaxDelay 最大延迟时间
	MaxDelay time.Duration `json:"max_delay"`
	// BackoffFactor 退避因子（指数退避时使用）
	BackoffFactor float64 `json:"backoff_factor"`
	// Jitter 是否添加抖动
	Jitter bool `json:"jitter"`
	// JitterFactor 抖动因子 (0.0-1.0)
	JitterFactor float64 `json:"jitter_factor"`
	// Policy 重试策略
	Policy RetryPolicy `json:"policy"`
	// Timeout 单次操作超时时间
	Timeout time.Duration `json:"timeout"`
	// RetryableErrors 可重试的错误类型
	RetryableErrors []string `json:"retryable_errors"`
	// CustomDelayFunc 自定义延迟计算函数
	CustomDelayFunc func(attempt int) time.Duration `json:"-"`
}

// DefaultRetryConfig 返回默认重试配置
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:      3,
		InitialDelay:    100 * time.Millisecond,
		MaxDelay:        30 * time.Second,
		BackoffFactor:   2.0,
		Jitter:          true,
		JitterFactor:    0.1,
		Policy:          RetryPolicyExponential,
		Timeout:         30 * time.Second,
		RetryableErrors: []string{"timeout", "connection", "network", "temporary", "retryable"},
	}
}

// RetryMetrics 重试指标
type RetryMetrics struct {
	TotalAttempts    int64         `json:"total_attempts"`
	SuccessAttempts  int64         `json:"success_attempts"`
	FailedAttempts   int64         `json:"failed_attempts"`
	TotalRetries     int64         `json:"total_retries"`
	AverageDelay     time.Duration `json:"average_delay"`
	MaxDelay         time.Duration `json:"max_delay"`
	LastAttemptTime  time.Time     `json:"last_attempt_time"`
	LastSuccessTime  time.Time     `json:"last_success_time"`
}

// RetryStrategy 重试策略实现
type RetryStrategy struct {
	name    string
	config  *RetryConfig
	logger  *slog.Logger
	metrics *RetryMetrics
	mutex   sync.RWMutex
	random  *rand.Rand

	// 回调函数
	onRetryAttempt func(attempt int, delay time.Duration, err error)
	onMaxRetriesExceeded func(attempts int, lastErr error)
}

// NewRetryStrategy 创建新的重试策略
func NewRetryStrategy(name string, config *RetryConfig, logger *slog.Logger) *RetryStrategy {
	if config == nil {
		config = DefaultRetryConfig()
	}

	return &RetryStrategy{
		name:   name,
		config: config,
		logger: logger,
		metrics: &RetryMetrics{
			LastSuccessTime: time.Now(),
		},
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Execute 执行带重试的操作
func (rs *RetryStrategy) Execute(ctx context.Context, operation func(ctx context.Context) error) error {
	var lastErr error

	for attempt := 0; attempt <= rs.config.MaxRetries; attempt++ {
		// 检查主上下文是否已取消
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled before retry attempt in strategy '%s': %w", rs.name, ctx.Err())
		default:
		}

		// 创建带超时的上下文
		opCtx, cancel := context.WithTimeout(ctx, rs.config.Timeout)

		// 执行操作
		err := operation(opCtx)
		cancel()

		// 记录尝试
		rs.recordAttempt(attempt, err)

		if err == nil {
			// 成功，记录并返回
			rs.recordSuccess(attempt)
			return nil
		}

		lastErr = err

		// 检查是否可重试
		if !rs.isRetryable(err) {
			rs.recordFailure(attempt, err, "non-retryable error")
			return fmt.Errorf("non-retryable error in retry strategy '%s': %w", rs.name, err)
		}

		// 如果已达到最大重试次数，退出
		if attempt >= rs.config.MaxRetries {
			rs.recordFailure(attempt, err, "max retries exceeded")
			break
		}

		// 计算延迟时间
		delay := rs.calculateDelay(attempt)

		// 触发重试回调
		if rs.onRetryAttempt != nil {
			rs.onRetryAttempt(attempt+1, delay, err)
		}

		if rs.logger != nil {
			rs.logger.Warn("Retry attempt failed, retrying",
				"strategy", rs.name,
				"attempt", attempt+1,
				"max_retries", rs.config.MaxRetries,
				"delay", delay,
				"error", err.Error())
		}

		// 等待延迟时间
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("context cancelled during retry delay in strategy '%s': %w", rs.name, ctx.Err())
		case <-timer.C:
			// 继续下一次重试
		}
	}

	// 触发最大重试次数超出回调
	if rs.onMaxRetriesExceeded != nil {
		rs.onMaxRetriesExceeded(rs.config.MaxRetries+1, lastErr)
	}

	return fmt.Errorf("max retries (%d) exceeded in strategy '%s', last error: %w",
		rs.config.MaxRetries, rs.name, lastErr)
}

// calculateDelay 计算延迟时间
func (rs *RetryStrategy) calculateDelay(attempt int) time.Duration {
	var delay time.Duration

	switch rs.config.Policy {
	case RetryPolicyFixed:
		delay = rs.config.InitialDelay
	case RetryPolicyLinear:
		delay = rs.config.InitialDelay * time.Duration(attempt+1)
	case RetryPolicyExponential:
		delay = time.Duration(float64(rs.config.InitialDelay) * math.Pow(rs.config.BackoffFactor, float64(attempt)))
	case RetryPolicyCustom:
		if rs.config.CustomDelayFunc != nil {
			delay = rs.config.CustomDelayFunc(attempt)
		} else {
			delay = rs.config.InitialDelay
		}
	default:
		delay = rs.config.InitialDelay
	}

	// 应用最大延迟限制
	if delay > rs.config.MaxDelay {
		delay = rs.config.MaxDelay
	}

	// 添加抖动
	if rs.config.Jitter {
		delay = rs.addJitter(delay)
	}

	return delay
}

// addJitter 添加抖动
func (rs *RetryStrategy) addJitter(delay time.Duration) time.Duration {
	if rs.config.JitterFactor <= 0 {
		return delay
	}

	// 计算抖动范围
	jitterRange := float64(delay) * rs.config.JitterFactor
	jitter := rs.random.Float64()*jitterRange*2 - jitterRange

	newDelay := time.Duration(float64(delay) + jitter)
	if newDelay < 0 {
		newDelay = delay / 2
	}

	return newDelay
}

// isRetryable 检查错误是否可重试
func (rs *RetryStrategy) isRetryable(err error) bool {
	if err == nil {
		return false
	}

	errorStr := err.Error()
	for _, retryableError := range rs.config.RetryableErrors {
		if contains(errorStr, retryableError) {
			return true
		}
	}

	return false
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// recordAttempt 记录尝试
func (rs *RetryStrategy) recordAttempt(attempt int, err error) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	rs.metrics.TotalAttempts++
	rs.metrics.LastAttemptTime = time.Now()

	if attempt > 0 {
		rs.metrics.TotalRetries++
	}

	if err != nil {
		rs.metrics.FailedAttempts++
	}
}

// recordSuccess 记录成功
func (rs *RetryStrategy) recordSuccess(attempt int) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	rs.metrics.SuccessAttempts++
	rs.metrics.LastSuccessTime = time.Now()

	if rs.logger != nil {
		rs.logger.Info("Retry strategy succeeded",
			"strategy", rs.name,
			"attempts", attempt+1,
			"total_retries", rs.metrics.TotalRetries)
	}
}

// recordFailure 记录失败
func (rs *RetryStrategy) recordFailure(attempt int, err error, reason string) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	if rs.logger != nil {
		rs.logger.Error("Retry strategy failed",
			"strategy", rs.name,
			"attempts", attempt+1,
			"reason", reason,
			"error", err.Error())
	}
}

// GetMetrics 获取指标
func (rs *RetryStrategy) GetMetrics() *RetryMetrics {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	// 返回指标的副本
	metrics := *rs.metrics
	return &metrics
}

// Reset 重置重试策略
func (rs *RetryStrategy) Reset() {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	rs.metrics = &RetryMetrics{
		LastSuccessTime: time.Now(),
	}

	if rs.logger != nil {
		rs.logger.Info("Retry strategy reset", "strategy", rs.name)
	}
}

// SetOnRetryAttempt 设置重试尝试回调
func (rs *RetryStrategy) SetOnRetryAttempt(callback func(attempt int, delay time.Duration, err error)) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	rs.onRetryAttempt = callback
}

// SetOnMaxRetriesExceeded 设置最大重试次数超出回调
func (rs *RetryStrategy) SetOnMaxRetriesExceeded(callback func(attempts int, lastErr error)) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	rs.onMaxRetriesExceeded = callback
}

// GetName 获取策略名称
func (rs *RetryStrategy) GetName() string {
	return rs.name
}

// GetConfig 获取配置
func (rs *RetryStrategy) GetConfig() *RetryConfig {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	// 返回配置的副本
	config := *rs.config
	return &config
}

// UpdateConfig 更新配置
func (rs *RetryStrategy) UpdateConfig(config *RetryConfig) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	rs.config = config

	if rs.logger != nil {
		rs.logger.Info("Retry strategy config updated",
			"strategy", rs.name,
			"max_retries", config.MaxRetries,
			"policy", config.Policy.String())
	}
}

// GetSuccessRate 获取成功率
func (rs *RetryStrategy) GetSuccessRate() float64 {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	if rs.metrics.TotalAttempts == 0 {
		return 0.0
	}

	return float64(rs.metrics.SuccessAttempts) / float64(rs.metrics.TotalAttempts)
}

// GetAverageRetries 获取平均重试次数
func (rs *RetryStrategy) GetAverageRetries() float64 {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	if rs.metrics.SuccessAttempts == 0 {
		return 0.0
	}

	return float64(rs.metrics.TotalRetries) / float64(rs.metrics.SuccessAttempts)
}