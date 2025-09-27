package plugin

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

// ErrorWrapper 错误包装器接口
type ErrorWrapper interface {
	// WrapError 包装错误
	WrapError(err error, code ErrorCode, message string) PluginError
	// WrapWithContext 带上下文包装错误
	WrapWithContext(ctx context.Context, err error, code ErrorCode, message string) PluginError
	// WrapWithPlugin 带插件信息包装错误
	WrapWithPlugin(err error, code ErrorCode, message string, pluginID string) PluginError
	// UnwrapError 解包错误
	UnwrapError(err error) error
	// IsPluginError 判断是否为插件错误
	IsPluginError(err error) bool
	// GetErrorChain 获取错误链
	GetErrorChain(err error) []error
	// GetRootCause 获取根因错误
	GetRootCause(err error) error
	// AddErrorContext 添加错误上下文
	AddErrorContext(err error, context map[string]interface{}) PluginError
}

// DefaultErrorWrapper 默认错误包装器
type DefaultErrorWrapper struct {
	includeStackTrace bool
	maxStackDepth     int
}

// NewErrorWrapper 创建新的错误包装器
func NewErrorWrapper(includeStackTrace bool, maxStackDepth int) ErrorWrapper {
	return &DefaultErrorWrapper{
		includeStackTrace: includeStackTrace,
		maxStackDepth:     maxStackDepth,
	}
}

// WrapError 包装错误
func (w *DefaultErrorWrapper) WrapError(err error, code ErrorCode, message string) PluginError {
	pluginErr := NewPluginError(code, message)
	pluginErr.Cause = err
	
	if w.includeStackTrace {
		pluginErr.StackTrace = w.captureStackTrace()
	}
	
	return pluginErr
}

// WrapWithContext 带上下文包装错误
func (w *DefaultErrorWrapper) WrapWithContext(ctx context.Context, err error, code ErrorCode, message string) PluginError {
	pluginErr := w.WrapError(err, code, message).(*BasePluginError)
	
	// 从上下文中提取相关信息
	if pluginID, ok := ctx.Value("plugin_id").(string); ok {
		pluginErr.WithContext("plugin_id", pluginID)
	}
	
	if requestID, ok := ctx.Value("request_id").(string); ok {
		pluginErr.WithContext("request_id", requestID)
	}
	
	if userID, ok := ctx.Value("user_id").(string); ok {
		pluginErr.WithContext("user_id", userID)
	}
	
	return pluginErr
}

// WrapWithPlugin 带插件信息包装错误
func (w *DefaultErrorWrapper) WrapWithPlugin(err error, code ErrorCode, message string, pluginID string) PluginError {
	pluginErr := w.WrapError(err, code, message).(*BasePluginError)
	pluginErr.WithContext("plugin_id", pluginID)
	return pluginErr
}

// UnwrapError 解包错误
func (w *DefaultErrorWrapper) UnwrapError(err error) error {
	if pluginErr, ok := err.(PluginError); ok {
		if baseErr, ok := pluginErr.(*BasePluginError); ok {
			return baseErr.Cause
		}
	}
	return err
}

// IsPluginError 判断是否为插件错误
func (w *DefaultErrorWrapper) IsPluginError(err error) bool {
	_, ok := err.(PluginError)
	return ok
}

// captureStackTrace 捕获堆栈跟踪
func (w *DefaultErrorWrapper) captureStackTrace() string {
	var stackTrace strings.Builder
	
	for i := 2; i < w.maxStackDepth+2; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			stackTrace.WriteString(fmt.Sprintf("%s:%d %s\n", file, line, fn.Name()))
		} else {
			stackTrace.WriteString(fmt.Sprintf("%s:%d\n", file, line))
		}
	}
	
	return stackTrace.String()
}

// ErrorChain 错误链
type ErrorChain struct {
	errors []PluginError
	mutex  sync.RWMutex
}

// NewErrorChain 创建新的错误链
func NewErrorChain() *ErrorChain {
	return &ErrorChain{
		errors: make([]PluginError, 0),
	}
}

// AddError 添加错误到链中
func (ec *ErrorChain) AddError(err PluginError) {
	ec.mutex.Lock()
	defer ec.mutex.Unlock()
	
	ec.errors = append(ec.errors, err)
}

// GetErrors 获取所有错误
func (ec *ErrorChain) GetErrors() []PluginError {
	ec.mutex.RLock()
	defer ec.mutex.RUnlock()
	
	errors := make([]PluginError, len(ec.errors))
	copy(errors, ec.errors)
	return errors
}

// GetLastError 获取最后一个错误
func (ec *ErrorChain) GetLastError() PluginError {
	ec.mutex.RLock()
	defer ec.mutex.RUnlock()
	
	if len(ec.errors) == 0 {
		return nil
	}
	
	return ec.errors[len(ec.errors)-1]
}

// HasErrors 是否有错误
func (ec *ErrorChain) HasErrors() bool {
	ec.mutex.RLock()
	defer ec.mutex.RUnlock()
	
	return len(ec.errors) > 0
}

// Clear 清空错误链
func (ec *ErrorChain) Clear() {
	ec.mutex.Lock()
	defer ec.mutex.Unlock()
	
	ec.errors = ec.errors[:0]
}

// Error 实现error接口
func (ec *ErrorChain) Error() string {
	ec.mutex.RLock()
	defer ec.mutex.RUnlock()
	
	if len(ec.errors) == 0 {
		return "no errors"
	}
	
	var messages []string
	for _, err := range ec.errors {
		messages = append(messages, err.Error())
	}
	
	return fmt.Sprintf("error chain: %s", strings.Join(messages, " -> "))
}

// ErrorAggregator 错误聚合器
type ErrorAggregator struct {
	errorsByPlugin map[string]*ErrorChain
	mutex          sync.RWMutex
}

// NewErrorAggregator 创建新的错误聚合器
func NewErrorAggregator() *ErrorAggregator {
	return &ErrorAggregator{
		errorsByPlugin: make(map[string]*ErrorChain),
	}
}

// AddError 添加错误
func (ea *ErrorAggregator) AddError(pluginID string, err PluginError) {
	ea.mutex.Lock()
	defer ea.mutex.Unlock()
	
	if _, exists := ea.errorsByPlugin[pluginID]; !exists {
		ea.errorsByPlugin[pluginID] = NewErrorChain()
	}
	
	ea.errorsByPlugin[pluginID].AddError(err)
}

// GetErrorsForPlugin 获取插件的错误
func (ea *ErrorAggregator) GetErrorsForPlugin(pluginID string) []PluginError {
	ea.mutex.RLock()
	defer ea.mutex.RUnlock()
	
	if chain, exists := ea.errorsByPlugin[pluginID]; exists {
		return chain.GetErrors()
	}
	
	return nil
}

// GetAllErrors 获取所有错误
func (ea *ErrorAggregator) GetAllErrors() map[string][]PluginError {
	ea.mutex.RLock()
	defer ea.mutex.RUnlock()
	
	allErrors := make(map[string][]PluginError)
	for pluginID, chain := range ea.errorsByPlugin {
		allErrors[pluginID] = chain.GetErrors()
	}
	
	return allErrors
}

// ClearErrorsForPlugin 清空插件的错误
func (ea *ErrorAggregator) ClearErrorsForPlugin(pluginID string) {
	ea.mutex.Lock()
	defer ea.mutex.Unlock()
	
	if chain, exists := ea.errorsByPlugin[pluginID]; exists {
		chain.Clear()
	}
}

// ClearAllErrors 清空所有错误
func (ea *ErrorAggregator) ClearAllErrors() {
	ea.mutex.Lock()
	defer ea.mutex.Unlock()
	
	for _, chain := range ea.errorsByPlugin {
		chain.Clear()
	}
}

// GetErrorChain 获取错误链
func (w *DefaultErrorWrapper) GetErrorChain(err error) []error {
	var chain []error
	current := err
	
	for current != nil && len(chain) < w.maxStackDepth {
		chain = append(chain, current)
		
		// 尝试解包错误
		if pluginErr, ok := current.(PluginError); ok {
			if baseErr, ok := pluginErr.(*BasePluginError); ok {
				current = baseErr.Cause
			} else {
				break
			}
		} else if unwrapper, ok := current.(interface{ Unwrap() error }); ok {
			current = unwrapper.Unwrap()
		} else {
			break
		}
	}
	
	return chain
}

// GetRootCause 获取根因错误
func (w *DefaultErrorWrapper) GetRootCause(err error) error {
	chain := w.GetErrorChain(err)
	if len(chain) == 0 {
		return err
	}
	return chain[len(chain)-1]
}

// AddErrorContext 添加错误上下文
func (w *DefaultErrorWrapper) AddErrorContext(err error, context map[string]interface{}) PluginError {
	if pluginErr, ok := err.(PluginError); ok {
		if baseErr, ok := pluginErr.(*BasePluginError); ok {
			for key, value := range context {
				baseErr.WithContext(key, value)
			}
			return baseErr
		}
	}
	
	// 如果不是插件错误，包装成插件错误并添加上下文
	wrappedErr := w.WrapError(err, ErrorCodeInternal, err.Error()).(*BasePluginError)
	for key, value := range context {
		wrappedErr.WithContext(key, value)
	}
	return wrappedErr
}

// GetErrorStats 获取错误统计
func (ea *ErrorAggregator) GetErrorStats(pluginID string) *ErrorStats {
	ea.mutex.RLock()
	defer ea.mutex.RUnlock()
	
	chain, exists := ea.errorsByPlugin[pluginID]
	if !exists || !chain.HasErrors() {
		return nil
	}
	
	errors := chain.GetErrors()
	stats := &ErrorStats{
		TotalErrors:  len(errors),
		ErrorsByType: make(map[ErrorType]int),
		ErrorRate:    0.0,
		LastError:    nil,
		FirstError:   nil,
	}
	
	for _, err := range errors {
		stats.ErrorsByType[err.GetType()]++
	}
	
	// 设置最后一个错误
	if len(errors) > 0 {
		lastErr := errors[len(errors)-1]
		stats.LastError = &ErrorRecord{
			Timestamp: time.Now(),
			Error:     lastErr,
			ErrorType: lastErr.GetType(),
		}
		stats.FirstError = &ErrorRecord{
			Timestamp: time.Now(),
			Error:     errors[0],
			ErrorType: errors[0].GetType(),
		}
	}
	
	return stats
}

// GetRootCause 获取错误链的根因
func (ec *ErrorChain) GetRootCause() PluginError {
	ec.mutex.RLock()
	defer ec.mutex.RUnlock()
	
	if len(ec.errors) == 0 {
		return nil
	}
	
	return ec.errors[0]
}

// GetLatestError 获取最新错误
func (ec *ErrorChain) GetLatestError() PluginError {
	ec.mutex.RLock()
	defer ec.mutex.RUnlock()
	
	if len(ec.errors) == 0 {
		return nil
	}
	
	return ec.errors[len(ec.errors)-1]
}

// Length 获取错误链长度
func (ec *ErrorChain) Length() int {
	ec.mutex.RLock()
	defer ec.mutex.RUnlock()
	
	return len(ec.errors)
}