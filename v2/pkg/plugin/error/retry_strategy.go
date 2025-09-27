package plugin

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// RetryStrategy 重试策略接口
type RetryStrategy interface {
	// ShouldRetry 判断是否应该重试
	ShouldRetry(attempt int, err error) bool
	// GetDelay 获取重试延迟
	GetDelay(attempt int) time.Duration
	// GetMaxAttempts 获取最大重试次数
	GetMaxAttempts() int
	// Reset 重置策略状态
	Reset()
}

// BackoffType 退避类型
type BackoffType int

const (
	BackoffTypeFixed BackoffType = iota
	BackoffTypeLinear
	BackoffTypeExponential
	BackoffTypeCustom
)

// String 返回退避类型的字符串表示
func (b BackoffType) String() string {
	switch b {
	case BackoffTypeFixed:
		return "fixed"
	case BackoffTypeLinear:
		return "linear"
	case BackoffTypeExponential:
		return "exponential"
	case BackoffTypeCustom:
		return "custom"
	default:
		return "unknown"
	}
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxAttempts   int                    `json:"max_attempts"`   // 最大重试次数
	BaseDelay     time.Duration          `json:"base_delay"`     // 基础延迟
	MaxDelay      time.Duration          `json:"max_delay"`      // 最大延迟
	BackoffType   BackoffType            `json:"backoff_type"`   // 退避类型
	BackoffFactor float64                `json:"backoff_factor"` // 退避因子
	Jitter        bool                   `json:"jitter"`         // 是否添加抖动
	JitterFactor  float64                `json:"jitter_factor"`  // 抖动因子
	RetryableErrors []ErrorCode          `json:"retryable_errors"` // 可重试的错误代码
	NonRetryableErrors []ErrorCode       `json:"non_retryable_errors"` // 不可重试的错误代码
	CustomDelayFunc func(int) time.Duration `json:"-"` // 自定义延迟函数
}

// DefaultRetryConfig 默认重试配置
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:   3,
		BaseDelay:     100 * time.Millisecond,
		MaxDelay:      30 * time.Second,
		BackoffType:   BackoffTypeExponential,
		BackoffFactor: 2.0,
		Jitter:        true,
		JitterFactor:  0.1,
		RetryableErrors: []ErrorCode{
			ErrorCodeUnavailable,
			ErrorCodeResourceExhausted,
			ErrorCodePluginTimeout,
			ErrorCodePluginNetworkError,
			ErrorCodeMusicSourceRateLimit,
			ErrorCodeThirdPartyServiceDown,
			ErrorCodeThirdPartyRateLimit,
		},
		NonRetryableErrors: []ErrorCode{
			ErrorCodeInvalidArgument,
			ErrorCodePermissionDenied,
			ErrorCodeUnauthenticated,
			ErrorCodeNotFound,
			ErrorCodePluginConfigInvalid,
		},
	}
}

// SmartRetryStrategy 智能重试策略实现
type SmartRetryStrategy struct {
	config    *RetryConfig
	attempts  int
	lastError error
	mutex     sync.RWMutex
	random    *rand.Rand
}

// NewSmartRetryStrategy 创建智能重试策略
func NewSmartRetryStrategy(config *RetryConfig) *SmartRetryStrategy {
	if config == nil {
		config = DefaultRetryConfig()
	}
	return &SmartRetryStrategy{
		config: config,
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// ShouldRetry 判断是否应该重试
func (s *SmartRetryStrategy) ShouldRetry(attempt int, err error) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.attempts = attempt
	s.lastError = err
	
	// 检查是否超过最大重试次数
	if attempt >= s.config.MaxAttempts {
		return false
	}
	
	// 检查错误是否可重试
	if pluginErr, ok := err.(*BasePluginError); ok {
		// 检查是否在不可重试列表中
		for _, code := range s.config.NonRetryableErrors {
			if pluginErr.GetCode() == code {
				return false
			}
		}
		
		// 检查是否在可重试列表中
		for _, code := range s.config.RetryableErrors {
			if pluginErr.GetCode() == code {
				return true
			}
		}
		
		// 使用插件错误的重试配置
		return pluginErr.IsRetryable()
	}
	
	// 对于非插件错误，检查是否是临时性错误
	return IsTemporary(err)
}

// GetDelay 获取重试延迟
func (s *SmartRetryStrategy) GetDelay(attempt int) time.Duration {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	var delay time.Duration
	
	// 如果有自定义延迟函数，使用自定义函数
	if s.config.CustomDelayFunc != nil {
		delay = s.config.CustomDelayFunc(attempt)
	} else {
		// 根据退避类型计算延迟
		switch s.config.BackoffType {
		case BackoffTypeFixed:
			delay = s.config.BaseDelay
		case BackoffTypeLinear:
			delay = time.Duration(float64(s.config.BaseDelay) * float64(attempt+1))
		case BackoffTypeExponential:
			delay = time.Duration(float64(s.config.BaseDelay) * math.Pow(s.config.BackoffFactor, float64(attempt)))
		default:
			delay = s.config.BaseDelay
		}
	}
	
	// 限制最大延迟
	if delay > s.config.MaxDelay {
		delay = s.config.MaxDelay
	}
	
	// 添加抖动
	if s.config.Jitter {
		jitterRange := float64(delay) * s.config.JitterFactor
		jitter := time.Duration(s.random.Float64()*jitterRange - jitterRange/2)
		delay += jitter
		if delay < 0 {
			delay = s.config.BaseDelay
		}
	}
	
	return delay
}

// GetMaxAttempts 获取最大重试次数
func (s *SmartRetryStrategy) GetMaxAttempts() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.config.MaxAttempts
}

// Reset 重置策略状态
func (s *SmartRetryStrategy) Reset() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.attempts = 0
	s.lastError = nil
}

// GetAttempts 获取当前重试次数
func (s *SmartRetryStrategy) GetAttempts() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.attempts
}

// GetLastError 获取最后一个错误
func (s *SmartRetryStrategy) GetLastError() error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.lastError
}

// RetryExecutor 重试执行器
type RetryExecutor struct {
	strategy RetryStrategy
	logger   Logger
	metrics  MetricsCollector
}

// NewRetryExecutor 创建重试执行器
func NewRetryExecutor(strategy RetryStrategy, logger Logger, metrics MetricsCollector) *RetryExecutor {
	return &RetryExecutor{
		strategy: strategy,
		logger:   logger,
		metrics:  metrics,
	}
}

// Execute 执行带重试的操作
func (r *RetryExecutor) Execute(ctx context.Context, operation func() error) error {
	r.strategy.Reset()
	
	for attempt := 0; attempt < r.strategy.GetMaxAttempts(); attempt++ {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		// 执行操作
		err := operation()
		if err == nil {
			// 操作成功，记录指标
			if r.metrics != nil {
				r.metrics.IncrementCounter("retry_success_total", map[string]string{
					"attempts": fmt.Sprintf("%d", attempt+1),
				})
			}
			return nil
		}
		
		// 记录错误指标
		if r.metrics != nil {
			r.metrics.IncrementCounter("retry_attempt_total", map[string]string{
				"attempt": fmt.Sprintf("%d", attempt+1),
				"error_type": getErrorType(err),
			})
		}
		
		// 判断是否应该重试
		if !r.strategy.ShouldRetry(attempt, err) {
			if r.logger != nil {
				r.logger.Error("Operation failed, not retrying", map[string]interface{}{
				"error": err,
				"attempt": attempt+1,
			})
			}
			return err
		}
		
		// 计算延迟时间
		delay := r.strategy.GetDelay(attempt)
		
		if r.logger != nil {
			r.logger.Warn("Operation failed, retrying", map[string]interface{}{
				"error": err,
				"attempt": attempt+1,
				"delay": delay,
			})
		}
		
		// 等待重试
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// 继续重试
		}
	}
	
	// 所有重试都失败了
	lastErr := r.strategy.(*SmartRetryStrategy).GetLastError()
	if r.logger != nil {
		r.logger.Error("Operation failed after all retries", map[string]interface{}{
			"error": lastErr,
			"attempts": r.strategy.GetMaxAttempts(),
		})
	}
	
	if r.metrics != nil {
		r.metrics.IncrementCounter("retry_exhausted_total", map[string]string{
			"error_type": getErrorType(lastErr),
		})
	}
	
	return fmt.Errorf("operation failed after %d attempts: %w", r.strategy.GetMaxAttempts(), lastErr)
}

// getErrorType 获取错误类型字符串
func getErrorType(err error) string {
	if pluginErr, ok := err.(*BasePluginError); ok {
		return pluginErr.GetType().String()
	}
	return "unknown"
}