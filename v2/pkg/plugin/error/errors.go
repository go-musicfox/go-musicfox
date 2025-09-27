package plugin

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// PluginError 插件错误接口
type PluginError interface {
	error
	GetCode() ErrorCode
	GetType() ErrorType
	GetSeverity() ErrorSeverity
	GetContext() map[string]interface{}
	GetTimestamp() time.Time
	IsRetryable() bool
	GetRetryAfter() time.Duration
}

// BasePluginError 基础插件错误
type BasePluginError struct {
	Code       ErrorCode                `json:"code"`        // 错误代码
	Type       ErrorType                `json:"type"`        // 错误类型
	Message    string                   `json:"message"`     // 错误消息
	Severity   ErrorSeverity            `json:"severity"`    // 严重程度
	Context    map[string]interface{}   `json:"context"`     // 错误上下文
	Cause      error                    `json:"cause"`       // 原因错误
	Timestamp  time.Time                `json:"timestamp"`   // 时间戳
	Retryable  bool                     `json:"retryable"`   // 是否可重试
	RetryAfter time.Duration            `json:"retry_after"` // 重试间隔
	StackTrace string                   `json:"stack_trace"` // 堆栈跟踪
}

// Error 实现error接口
func (e *BasePluginError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %s)", e.Code.String(), e.Message, e.Cause.Error())
	}
	return fmt.Sprintf("%s: %s", e.Code.String(), e.Message)
}

// GetCode 获取错误代码
func (e *BasePluginError) GetCode() ErrorCode {
	return e.Code
}

// GetType 获取错误类型
func (e *BasePluginError) GetType() ErrorType {
	return e.Type
}

// GetSeverity 获取严重程度
func (e *BasePluginError) GetSeverity() ErrorSeverity {
	return e.Severity
}

// GetContext 获取错误上下文
func (e *BasePluginError) GetContext() map[string]interface{} {
	return e.Context
}

// GetTimestamp 获取时间戳
func (e *BasePluginError) GetTimestamp() time.Time {
	return e.Timestamp
}

// IsRetryable 是否可重试
func (e *BasePluginError) IsRetryable() bool {
	return e.Retryable
}

// GetRetryAfter 获取重试间隔
func (e *BasePluginError) GetRetryAfter() time.Duration {
	return e.RetryAfter
}

// ErrorCode 错误代码枚举
type ErrorCode int

const (
	// 通用错误
	ErrorCodeUnknown ErrorCode = iota
	ErrorCodeInternal
	ErrorCodeInvalidArgument
	ErrorCodeNotFound
	ErrorCodeAlreadyExists
	ErrorCodePermissionDenied
	ErrorCodeUnauthenticated
	ErrorCodeResourceExhausted
	ErrorCodeFailedPrecondition
	ErrorCodeAborted
	ErrorCodeOutOfRange
	ErrorCodeUnimplemented
	ErrorCodeUnavailable
	ErrorCodeDataLoss

	// 插件生命周期错误
	ErrorCodePluginNotFound
	ErrorCodePluginAlreadyLoaded
	ErrorCodePluginNotLoaded
	ErrorCodePluginInitFailed
	ErrorCodePluginStartFailed
	ErrorCodePluginStopFailed
	ErrorCodePluginDestroyFailed
	ErrorCodePluginConfigInvalid
	ErrorCodePluginDependencyMissing
	ErrorCodePluginVersionMismatch
	ErrorCodePluginIncompatible

	// 插件运行时错误
	ErrorCodePluginTimeout
	ErrorCodePluginCrashed
	ErrorCodePluginMemoryLimit
	ErrorCodePluginCPULimit
	ErrorCodePluginIOError
	ErrorCodePluginNetworkError
	ErrorCodePluginDatabaseError
	ErrorCodePluginFileSystemError
	ErrorCodePluginResourceLimit
	
	// 验证错误
	ErrorCodeValidation

	// 音频处理错误
	ErrorCodeAudioFormatUnsupported
	ErrorCodeAudioDecodeFailed
	ErrorCodeAudioEncodeFailed
	ErrorCodeAudioDeviceError
	ErrorCodeAudioBufferUnderrun
	ErrorCodeAudioBufferOverrun

	// 音乐源错误
	ErrorCodeMusicSourceUnavailable
	ErrorCodeMusicSourceAuthFailed
	ErrorCodeMusicSourceRateLimit
	ErrorCodeMusicSourceQuotaExceeded
	ErrorCodeMusicSourceContentNotFound
	ErrorCodeMusicSourceRegionBlocked
	ErrorCodeMusicSourceQualityUnavailable

	// 第三方服务错误
	ErrorCodeThirdPartyServiceDown
	ErrorCodeThirdPartyAuthExpired
	ErrorCodeThirdPartyRateLimit
	ErrorCodeThirdPartyQuotaExceeded
	ErrorCodeThirdPartyAPIChanged
	ErrorCodeThirdPartyNetworkError

	// UI扩展错误
	ErrorCodeUIComponentNotFound
	ErrorCodeUIThemeInvalid
	ErrorCodeUILayoutInvalid
	ErrorCodeUIRenderFailed
	ErrorCodeUIEventHandlerFailed
	ErrorCodeUIResourceNotFound
)

// String 返回错误代码的字符串表示
func (e ErrorCode) String() string {
	switch e {
	// 通用错误
	case ErrorCodeUnknown:
		return "UNKNOWN"
	case ErrorCodeInternal:
		return "INTERNAL"
	case ErrorCodeInvalidArgument:
		return "INVALID_ARGUMENT"
	case ErrorCodeNotFound:
		return "NOT_FOUND"
	case ErrorCodeAlreadyExists:
		return "ALREADY_EXISTS"
	case ErrorCodePermissionDenied:
		return "PERMISSION_DENIED"
	case ErrorCodeUnauthenticated:
		return "UNAUTHENTICATED"
	case ErrorCodeResourceExhausted:
		return "RESOURCE_EXHAUSTED"
	case ErrorCodeFailedPrecondition:
		return "FAILED_PRECONDITION"
	case ErrorCodeAborted:
		return "ABORTED"
	case ErrorCodeOutOfRange:
		return "OUT_OF_RANGE"
	case ErrorCodeUnimplemented:
		return "UNIMPLEMENTED"
	case ErrorCodeUnavailable:
		return "UNAVAILABLE"
	case ErrorCodeDataLoss:
		return "DATA_LOSS"

	// 插件生命周期错误
	case ErrorCodePluginNotFound:
		return "PLUGIN_NOT_FOUND"
	case ErrorCodePluginAlreadyLoaded:
		return "PLUGIN_ALREADY_LOADED"
	case ErrorCodePluginNotLoaded:
		return "PLUGIN_NOT_LOADED"
	case ErrorCodePluginInitFailed:
		return "PLUGIN_INIT_FAILED"
	case ErrorCodePluginStartFailed:
		return "PLUGIN_START_FAILED"
	case ErrorCodePluginStopFailed:
		return "PLUGIN_STOP_FAILED"
	case ErrorCodePluginDestroyFailed:
		return "PLUGIN_DESTROY_FAILED"
	case ErrorCodePluginConfigInvalid:
		return "PLUGIN_CONFIG_INVALID"
	case ErrorCodePluginDependencyMissing:
		return "PLUGIN_DEPENDENCY_MISSING"
	case ErrorCodePluginVersionMismatch:
		return "PLUGIN_VERSION_MISMATCH"
	case ErrorCodePluginIncompatible:
		return "PLUGIN_INCOMPATIBLE"

	// 插件运行时错误
	case ErrorCodePluginTimeout:
		return "PLUGIN_TIMEOUT"
	case ErrorCodePluginCrashed:
		return "PLUGIN_CRASHED"
	case ErrorCodePluginMemoryLimit:
		return "PLUGIN_MEMORY_LIMIT"
	case ErrorCodePluginCPULimit:
		return "PLUGIN_CPU_LIMIT"
	case ErrorCodePluginIOError:
		return "PLUGIN_IO_ERROR"
	case ErrorCodePluginNetworkError:
		return "PLUGIN_NETWORK_ERROR"
	case ErrorCodePluginDatabaseError:
		return "PLUGIN_DATABASE_ERROR"
	case ErrorCodePluginFileSystemError:
		return "PLUGIN_FILESYSTEM_ERROR"
	case ErrorCodePluginResourceLimit:
		return "PLUGIN_RESOURCE_LIMIT"

	// 验证错误
	case ErrorCodeValidation:
		return "VALIDATION"

	// 音频处理错误
	case ErrorCodeAudioFormatUnsupported:
		return "AUDIO_FORMAT_UNSUPPORTED"
	case ErrorCodeAudioDecodeFailed:
		return "AUDIO_DECODE_FAILED"
	case ErrorCodeAudioEncodeFailed:
		return "AUDIO_ENCODE_FAILED"
	case ErrorCodeAudioDeviceError:
		return "AUDIO_DEVICE_ERROR"
	case ErrorCodeAudioBufferUnderrun:
		return "AUDIO_BUFFER_UNDERRUN"
	case ErrorCodeAudioBufferOverrun:
		return "AUDIO_BUFFER_OVERRUN"

	// 音乐源错误
	case ErrorCodeMusicSourceUnavailable:
		return "MUSIC_SOURCE_UNAVAILABLE"
	case ErrorCodeMusicSourceAuthFailed:
		return "MUSIC_SOURCE_AUTH_FAILED"
	case ErrorCodeMusicSourceRateLimit:
		return "MUSIC_SOURCE_RATE_LIMIT"
	case ErrorCodeMusicSourceQuotaExceeded:
		return "MUSIC_SOURCE_QUOTA_EXCEEDED"
	case ErrorCodeMusicSourceContentNotFound:
		return "MUSIC_SOURCE_CONTENT_NOT_FOUND"
	case ErrorCodeMusicSourceRegionBlocked:
		return "MUSIC_SOURCE_REGION_BLOCKED"
	case ErrorCodeMusicSourceQualityUnavailable:
		return "MUSIC_SOURCE_QUALITY_UNAVAILABLE"

	// 第三方服务错误
	case ErrorCodeThirdPartyServiceDown:
		return "THIRD_PARTY_SERVICE_DOWN"
	case ErrorCodeThirdPartyAuthExpired:
		return "THIRD_PARTY_AUTH_EXPIRED"
	case ErrorCodeThirdPartyRateLimit:
		return "THIRD_PARTY_RATE_LIMIT"
	case ErrorCodeThirdPartyQuotaExceeded:
		return "THIRD_PARTY_QUOTA_EXCEEDED"
	case ErrorCodeThirdPartyAPIChanged:
		return "THIRD_PARTY_API_CHANGED"
	case ErrorCodeThirdPartyNetworkError:
		return "THIRD_PARTY_NETWORK_ERROR"

	// UI扩展错误
	case ErrorCodeUIComponentNotFound:
		return "UI_COMPONENT_NOT_FOUND"
	case ErrorCodeUIThemeInvalid:
		return "UI_THEME_INVALID"
	case ErrorCodeUILayoutInvalid:
		return "UI_LAYOUT_INVALID"
	case ErrorCodeUIRenderFailed:
		return "UI_RENDER_FAILED"
	case ErrorCodeUIEventHandlerFailed:
		return "UI_EVENT_HANDLER_FAILED"
	case ErrorCodeUIResourceNotFound:
		return "UI_RESOURCE_NOT_FOUND"

	default:
		return "UNKNOWN"
	}
}

// ErrorType 错误类型枚举
type ErrorType int

const (
	ErrorTypeSystem ErrorType = iota
	ErrorTypePlugin
	ErrorTypeNetwork
	ErrorTypeIO
	ErrorTypeValidation
	ErrorTypeAuthentication
	ErrorTypeAuthorization
	ErrorTypeConfiguration
	ErrorTypeResource
	ErrorTypeTimeout
	ErrorTypeRateLimit
	ErrorTypeCompatibility
	ErrorTypeData
	ErrorTypeUI
	ErrorTypeAudio
	ErrorTypeMusicSource
	ErrorTypeThirdParty
)

// String 返回错误类型的字符串表示
func (e ErrorType) String() string {
	switch e {
	case ErrorTypeSystem:
		return "system"
	case ErrorTypePlugin:
		return "plugin"
	case ErrorTypeNetwork:
		return "network"
	case ErrorTypeIO:
		return "io"
	case ErrorTypeValidation:
		return "validation"
	case ErrorTypeAuthentication:
		return "authentication"
	case ErrorTypeAuthorization:
		return "authorization"
	case ErrorTypeConfiguration:
		return "configuration"
	case ErrorTypeResource:
		return "resource"
	case ErrorTypeTimeout:
		return "timeout"
	case ErrorTypeRateLimit:
		return "rate_limit"
	case ErrorTypeCompatibility:
		return "compatibility"
	case ErrorTypeData:
		return "data"
	case ErrorTypeUI:
		return "ui"
	case ErrorTypeAudio:
		return "audio"
	case ErrorTypeMusicSource:
		return "music_source"
	case ErrorTypeThirdParty:
		return "third_party"
	default:
		return "unknown"
	}
}

// ErrorSeverity 错误严重程度枚举
type ErrorSeverity int

const (
	ErrorSeverityTrace ErrorSeverity = iota
	ErrorSeverityDebug
	ErrorSeverityInfo
	ErrorSeverityWarning
	ErrorSeverityError
	ErrorSeverityFatal
	ErrorSeverityCritical
)

// String 返回错误严重程度的字符串表示
func (e ErrorSeverity) String() string {
	switch e {
	case ErrorSeverityTrace:
		return "trace"
	case ErrorSeverityDebug:
		return "debug"
	case ErrorSeverityInfo:
		return "info"
	case ErrorSeverityWarning:
		return "warning"
	case ErrorSeverityError:
		return "error"
	case ErrorSeverityFatal:
		return "fatal"
	case ErrorSeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}



// ErrorHandlerFunc 错误处理器函数类型
type ErrorHandlerFunc func(ctx context.Context, err PluginError) error

// ErrorRecoveryStrategy 错误恢复策略
type ErrorRecoveryStrategy struct {
	Type              RecoveryType           `json:"type"`               // 恢复类型
	MaxRetries        int                    `json:"max_retries"`        // 最大重试次数
	RetryDelay        time.Duration          `json:"retry_delay"`        // 重试延迟
	BackoffFactor     float64                `json:"backoff_factor"`     // 退避因子
	MaxDelay          time.Duration          `json:"max_delay"`          // 最大延迟
	Jitter            bool                   `json:"jitter"`             // 是否添加抖动
	Fallback          string                 `json:"fallback"`           // 降级策略
	CircuitBreaker    *CircuitBreakerConfig  `json:"circuit_breaker"`    // 熔断器配置
	DegradationConfig *DegradationConfig     `json:"degradation_config"` // 降级配置
}

// RecoveryType 恢复类型枚举
type RecoveryType int

const (
	RecoveryTypeNone RecoveryType = iota
	RecoveryTypeRetry
	RecoveryTypeFallback
	RecoveryTypeRestart
	RecoveryTypeCircuitBreaker
	RecoveryTypeGracefulDegradation
)

// String 返回恢复类型的字符串表示
func (r RecoveryType) String() string {
	switch r {
	case RecoveryTypeNone:
		return "none"
	case RecoveryTypeRetry:
		return "retry"
	case RecoveryTypeFallback:
		return "fallback"
	case RecoveryTypeRestart:
		return "restart"
	case RecoveryTypeCircuitBreaker:
		return "circuit_breaker"
	case RecoveryTypeGracefulDegradation:
		return "graceful_degradation"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	FailureThreshold int           `json:"failure_threshold"` // 失败阈值
	SuccessThreshold int           `json:"success_threshold"` // 成功阈值
	Timeout          time.Duration `json:"timeout"`           // 超时时间
	ResetTimeout     time.Duration `json:"reset_timeout"`     // 重置超时
	MaxRequests      int           `json:"max_requests"`      // 最大请求数
}

// ErrorMetrics 错误指标
type ErrorMetrics struct {
	TotalErrors     int64                    `json:"total_errors"`     // 总错误数
	ErrorsByCode    map[ErrorCode]int64      `json:"errors_by_code"`   // 按错误代码统计
	ErrorsByType    map[ErrorType]int64      `json:"errors_by_type"`   // 按错误类型统计
	ErrorsBySeverity map[ErrorSeverity]int64 `json:"errors_by_severity"` // 按严重程度统计
	ErrorRate       float64                  `json:"error_rate"`       // 错误率
	MTTR            time.Duration            `json:"mttr"`             // 平均恢复时间
	MTBF            time.Duration            `json:"mtbf"`             // 平均故障间隔时间
	LastError       *BasePluginError         `json:"last_error"`       // 最后一个错误
	UpdatedAt       time.Time                `json:"updated_at"`       // 更新时间
}

// NewPluginError 创建新的插件错误
func NewPluginError(code ErrorCode, message string) *BasePluginError {
	return &BasePluginError{
		Code:      code,
		Type:      getErrorTypeByCode(code),
		Message:   message,
		Severity:  getErrorSeverityByCode(code),
		Context:   make(map[string]interface{}),
		Timestamp: time.Now(),
		Retryable: isRetryableByCode(code),
	}
}

// NewPluginErrorWithCause 创建带原因的插件错误
func NewPluginErrorWithCause(code ErrorCode, message string, cause error) *BasePluginError {
	err := NewPluginError(code, message)
	err.Cause = cause
	err.StackTrace = captureStackTrace()
	return err
}

// WithContext 添加错误上下文
func (e *BasePluginError) WithContext(key string, value interface{}) *BasePluginError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithRetryConfig 设置重试配置
func (e *BasePluginError) WithRetryConfig(retryable bool, retryAfter time.Duration) *BasePluginError {
	e.Retryable = retryable
	e.RetryAfter = retryAfter
	return e
}

// WithSeverity 设置错误严重程度
func (e *BasePluginError) WithSeverity(severity ErrorSeverity) *BasePluginError {
	e.Severity = severity
	return e
}

// WithStackTrace 设置堆栈跟踪
func (e *BasePluginError) WithStackTrace(stack string) *BasePluginError {
	e.StackTrace = stack
	return e
}

// GetStackTrace 获取堆栈跟踪
func (e *BasePluginError) GetStackTrace() string {
	return e.StackTrace
}

// GetCause 获取原因错误
func (e *BasePluginError) GetCause() error {
	return e.Cause
}

// Unwrap 实现errors.Unwrap接口
func (e *BasePluginError) Unwrap() error {
	return e.Cause
}

// Is 实现errors.Is接口
func (e *BasePluginError) Is(target error) bool {
	if t, ok := target.(*BasePluginError); ok {
		return e.Code == t.Code
	}
	return false
}

// As 实现errors.As接口
func (e *BasePluginError) As(target interface{}) bool {
	if t, ok := target.(**BasePluginError); ok {
		*t = e
		return true
	}
	return false
}

// captureStackTrace 捕获堆栈跟踪
func captureStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// NewMusicFoxError 创建MusicFox错误（兼容设计文档）
func NewMusicFoxError(code ErrorCode, message, details, source string) *BasePluginError {
	err := NewPluginError(code, message)
	err.WithContext("details", details)
	err.WithContext("source", source)
	err.StackTrace = captureStackTrace()
	return err
}

// WrapError 包装标准错误为插件错误
func WrapError(err error, code ErrorCode, message string) *BasePluginError {
	if err == nil {
		return nil
	}
	pluginErr := NewPluginErrorWithCause(code, message, err)
	return pluginErr
}

// IsTemporary 判断错误是否是临时性的
func IsTemporary(err error) bool {
	if pluginErr, ok := err.(*BasePluginError); ok {
		return pluginErr.IsRetryable()
	}
	if temp, ok := err.(interface{ Temporary() bool }); ok {
		return temp.Temporary()
	}
	return false
}

// IsTimeout 判断错误是否是超时错误
func IsTimeout(err error) bool {
	if pluginErr, ok := err.(*BasePluginError); ok {
		// 检查错误代码是否为超时相关
		return pluginErr.GetCode() == ErrorCodePluginTimeout
	}
	if timeout, ok := err.(interface{ Timeout() bool }); ok {
		return timeout.Timeout()
	}
	return false
}

// getErrorTypeByCode 根据错误代码获取错误类型
func getErrorTypeByCode(code ErrorCode) ErrorType {
	switch {
	case code >= ErrorCodePluginNotFound && code <= ErrorCodePluginIncompatible:
		return ErrorTypePlugin
	case code >= ErrorCodePluginTimeout && code <= ErrorCodePluginFileSystemError:
		return ErrorTypeSystem
	case code >= ErrorCodeAudioFormatUnsupported && code <= ErrorCodeAudioBufferOverrun:
		return ErrorTypeAudio
	case code >= ErrorCodeMusicSourceUnavailable && code <= ErrorCodeMusicSourceQualityUnavailable:
		return ErrorTypeMusicSource
	case code >= ErrorCodeThirdPartyServiceDown && code <= ErrorCodeThirdPartyNetworkError:
		return ErrorTypeThirdParty
	case code >= ErrorCodeUIComponentNotFound && code <= ErrorCodeUIResourceNotFound:
		return ErrorTypeUI
	default:
		return ErrorTypeSystem
	}
}

// getErrorSeverityByCode 根据错误代码获取错误严重程度
func getErrorSeverityByCode(code ErrorCode) ErrorSeverity {
	switch code {
	case ErrorCodePluginCrashed, ErrorCodeDataLoss:
		return ErrorSeverityCritical
	case ErrorCodeInternal, ErrorCodePluginInitFailed, ErrorCodePluginStartFailed:
		return ErrorSeverityFatal
	case ErrorCodePluginTimeout, ErrorCodeUnavailable, ErrorCodeResourceExhausted:
		return ErrorSeverityError
	case ErrorCodePluginConfigInvalid, ErrorCodeInvalidArgument:
		return ErrorSeverityWarning
	default:
		return ErrorSeverityError
	}
}

// isRetryableByCode 根据错误代码判断是否可重试
func isRetryableByCode(code ErrorCode) bool {
	switch code {
	case ErrorCodeUnavailable, ErrorCodeResourceExhausted, ErrorCodePluginTimeout,
		ErrorCodePluginNetworkError, ErrorCodeMusicSourceRateLimit,
		ErrorCodeThirdPartyServiceDown, ErrorCodeThirdPartyRateLimit:
		return true
	default:
		return false
	}
}