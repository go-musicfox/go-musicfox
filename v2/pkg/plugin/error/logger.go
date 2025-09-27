package plugin

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"sync"
	"time"
)

// Logger 日志记录器接口
type Logger interface {
	// Trace 跟踪级别日志
	Trace(message string, fields map[string]interface{})
	// Debug 调试级别日志
	Debug(message string, fields ...map[string]interface{})
	// Info 信息级别日志
	Info(message string, fields map[string]interface{})
	// Warn 警告级别日志
	Warn(message string, fields map[string]interface{})
	// Error 错误级别日志
	Error(message string, fields map[string]interface{})
	// Fatal 致命级别日志
	Fatal(message string, fields map[string]interface{})
	// LogError 记录插件错误
	LogError(ctx context.Context, err PluginError, pluginID string)
	// LogErrorWithContext 带上下文记录错误
	LogErrorWithContext(ctx context.Context, err PluginError, pluginID string, additionalFields map[string]interface{})
	// SetLevel 设置日志级别
	SetLevel(level LogLevel)
	// GetLevel 获取日志级别
	GetLevel() LogLevel
}

// LogLevel 日志级别
type LogLevel int

const (
	LogLevelTrace LogLevel = iota
	LogLevelDebug
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// String 返回日志级别的字符串表示
func (l LogLevel) String() string {
	switch l {
	case LogLevelTrace:
		return "TRACE"
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ErrorLogger 错误日志记录器
type ErrorLogger struct {
	logger    *slog.Logger
	level     LogLevel
	mutex     sync.RWMutex
	formatters map[string]ErrorFormatter
	filters   []ErrorFilter
	enrichers []ErrorEnricher
}

// ErrorFormatter 错误格式化器接口
type ErrorFormatter interface {
	// Format 格式化错误
	Format(err PluginError, pluginID string, additionalFields map[string]interface{}) map[string]interface{}
	// GetName 获取格式化器名称
	GetName() string
}

// ErrorFilter 错误过滤器接口
type ErrorFilter interface {
	// ShouldLog 是否应该记录日志
	ShouldLog(err PluginError, pluginID string) bool
	// GetName 获取过滤器名称
	GetName() string
}

// ErrorEnricher 错误丰富器接口
type ErrorEnricher interface {
	// Enrich 丰富错误信息
	Enrich(ctx context.Context, err PluginError, pluginID string) map[string]interface{}
	// GetName 获取丰富器名称
	GetName() string
}

// NewErrorLogger 创建新的错误日志记录器
func NewErrorLogger(logger *slog.Logger, level LogLevel) Logger {
	return &ErrorLogger{
		logger:     logger,
		level:      level,
		formatters: make(map[string]ErrorFormatter),
		filters:    make([]ErrorFilter, 0),
		enrichers:  make([]ErrorEnricher, 0),
	}
}

// Trace 跟踪级别日志
func (el *ErrorLogger) Trace(message string, fields map[string]interface{}) {
	if el.shouldLog(LogLevelTrace) {
		el.logWithFields(slog.LevelDebug-1, message, fields)
	}
}

// Debug 调试级别日志
func (el *ErrorLogger) Debug(message string, fields ...map[string]interface{}) {
	if el.shouldLog(LogLevelDebug) {
		var allFields map[string]interface{}
		if len(fields) > 0 {
			allFields = fields[0]
		}
		el.logWithFields(slog.LevelDebug, message, allFields)
	}
}

// Info 信息级别日志
func (el *ErrorLogger) Info(message string, fields map[string]interface{}) {
	if el.shouldLog(LogLevelInfo) {
		el.logWithFields(slog.LevelInfo, message, fields)
	}
}

// Warn 警告级别日志
func (el *ErrorLogger) Warn(message string, fields map[string]interface{}) {
	if el.shouldLog(LogLevelWarn) {
		el.logWithFields(slog.LevelWarn, message, fields)
	}
}

// Error 错误级别日志
func (el *ErrorLogger) Error(message string, fields map[string]interface{}) {
	if el.shouldLog(LogLevelError) {
		el.logWithFields(slog.LevelError, message, fields)
	}
}

// Fatal 致命级别日志
func (el *ErrorLogger) Fatal(message string, fields map[string]interface{}) {
	if el.shouldLog(LogLevelFatal) {
		el.logWithFields(slog.LevelError+1, message, fields)
	}
}

// LogError 记录插件错误
func (el *ErrorLogger) LogError(ctx context.Context, err PluginError, pluginID string) {
	el.LogErrorWithContext(ctx, err, pluginID, nil)
}

// LogErrorWithContext 带上下文记录错误
func (el *ErrorLogger) LogErrorWithContext(ctx context.Context, err PluginError, pluginID string, additionalFields map[string]interface{}) {
	// 应用过滤器
	for _, filter := range el.filters {
		if !filter.ShouldLog(err, pluginID) {
			return
		}
	}
	
	// 确定日志级别
	logLevel := el.mapSeverityToLogLevel(err.GetSeverity())
	if !el.shouldLog(logLevel) {
		return
	}
	
	// 构建基础字段
	fields := map[string]interface{}{
		"plugin_id":    pluginID,
		"error_code":   err.GetCode().String(),
		"error_type":   err.GetType().String(),
		"error_severity": err.GetSeverity().String(),
		"timestamp":    err.GetTimestamp().Format(time.RFC3339),
		"retryable":    err.IsRetryable(),
	}
	
	// 添加错误上下文
	if errContext := err.GetContext(); errContext != nil {
		fields["error_context"] = errContext
	}
	
	// 添加重试信息
	if err.GetRetryAfter() > 0 {
		fields["retry_after"] = err.GetRetryAfter().String()
	}
	
	// 添加堆栈跟踪
	if baseErr, ok := err.(*BasePluginError); ok && baseErr.StackTrace != "" {
		fields["stack_trace"] = baseErr.StackTrace
	}
	
	// 添加原因错误
	if baseErr, ok := err.(*BasePluginError); ok && baseErr.Cause != nil {
		fields["cause"] = baseErr.Cause.Error()
	}
	
	// 应用丰富器
	for _, enricher := range el.enrichers {
		enrichedFields := enricher.Enrich(ctx, err, pluginID)
		for k, v := range enrichedFields {
			fields[k] = v
		}
	}
	
	// 添加额外字段
	if additionalFields != nil {
		for k, v := range additionalFields {
			fields[k] = v
		}
	}
	
	// 应用格式化器
	for _, formatter := range el.formatters {
		formattedFields := formatter.Format(err, pluginID, additionalFields)
		for k, v := range formattedFields {
			fields[k] = v
		}
	}
	
	// 记录日志
	slogLevel := el.mapLogLevelToSlogLevel(logLevel)
	el.logWithFields(slogLevel, err.Error(), fields)
}

// SetLevel 设置日志级别
func (el *ErrorLogger) SetLevel(level LogLevel) {
	el.mutex.Lock()
	defer el.mutex.Unlock()
	el.level = level
}

// GetLevel 获取日志级别
func (el *ErrorLogger) GetLevel() LogLevel {
	el.mutex.RLock()
	defer el.mutex.RUnlock()
	return el.level
}

// AddFormatter 添加格式化器
func (el *ErrorLogger) AddFormatter(formatter ErrorFormatter) {
	el.mutex.Lock()
	defer el.mutex.Unlock()
	el.formatters[formatter.GetName()] = formatter
}

// RemoveFormatter 移除格式化器
func (el *ErrorLogger) RemoveFormatter(name string) {
	el.mutex.Lock()
	defer el.mutex.Unlock()
	delete(el.formatters, name)
}

// AddFilter 添加过滤器
func (el *ErrorLogger) AddFilter(filter ErrorFilter) {
	el.mutex.Lock()
	defer el.mutex.Unlock()
	el.filters = append(el.filters, filter)
}

// RemoveFilter 移除过滤器
func (el *ErrorLogger) RemoveFilter(name string) {
	el.mutex.Lock()
	defer el.mutex.Unlock()
	
	for i, filter := range el.filters {
		if filter.GetName() == name {
			el.filters = append(el.filters[:i], el.filters[i+1:]...)
			break
		}
	}
}

// AddEnricher 添加丰富器
func (el *ErrorLogger) AddEnricher(enricher ErrorEnricher) {
	el.mutex.Lock()
	defer el.mutex.Unlock()
	el.enrichers = append(el.enrichers, enricher)
}

// RemoveEnricher 移除丰富器
func (el *ErrorLogger) RemoveEnricher(name string) {
	el.mutex.Lock()
	defer el.mutex.Unlock()
	
	for i, enricher := range el.enrichers {
		if enricher.GetName() == name {
			el.enrichers = append(el.enrichers[:i], el.enrichers[i+1:]...)
			break
		}
	}
}

// shouldLog 是否应该记录日志
func (el *ErrorLogger) shouldLog(level LogLevel) bool {
	el.mutex.RLock()
	defer el.mutex.RUnlock()
	return level >= el.level
}

// logWithFields 带字段记录日志
func (el *ErrorLogger) logWithFields(level slog.Level, message string, fields map[string]interface{}) {
	if fields == nil || len(fields) == 0 {
		el.logger.Log(context.Background(), level, message)
		return
	}
	
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	
	el.logger.Log(context.Background(), level, message, args...)
}

// mapSeverityToLogLevel 映射严重程度到日志级别
func (el *ErrorLogger) mapSeverityToLogLevel(severity ErrorSeverity) LogLevel {
	switch severity {
	case ErrorSeverityTrace:
		return LogLevelTrace
	case ErrorSeverityDebug:
		return LogLevelDebug
	case ErrorSeverityInfo:
		return LogLevelInfo
	case ErrorSeverityWarning:
		return LogLevelWarn
	case ErrorSeverityError:
		return LogLevelError
	case ErrorSeverityFatal, ErrorSeverityCritical:
		return LogLevelFatal
	default:
		return LogLevelError
	}
}

// mapLogLevelToSlogLevel 映射日志级别到slog级别
func (el *ErrorLogger) mapLogLevelToSlogLevel(level LogLevel) slog.Level {
	switch level {
	case LogLevelTrace:
		return slog.LevelDebug - 1
	case LogLevelDebug:
		return slog.LevelDebug
	case LogLevelInfo:
		return slog.LevelInfo
	case LogLevelWarn:
		return slog.LevelWarn
	case LogLevelError:
		return slog.LevelError
	case LogLevelFatal:
		return slog.LevelError + 1
	default:
		return slog.LevelError
	}
}

// JSONErrorFormatter JSON错误格式化器
type JSONErrorFormatter struct {
	name string
}

// NewJSONErrorFormatter 创建JSON错误格式化器
func NewJSONErrorFormatter() ErrorFormatter {
	return &JSONErrorFormatter{
		name: "json",
	}
}

// Format 格式化错误
func (f *JSONErrorFormatter) Format(err PluginError, pluginID string, additionalFields map[string]interface{}) map[string]interface{} {
	errorData := map[string]interface{}{
		"code":      err.GetCode().String(),
		"type":      err.GetType().String(),
		"severity":  err.GetSeverity().String(),
		"message":   err.Error(),
		"timestamp": err.GetTimestamp().Format(time.RFC3339),
		"retryable": err.IsRetryable(),
		"context":   err.GetContext(),
	}
	
	if err.GetRetryAfter() > 0 {
		errorData["retry_after"] = err.GetRetryAfter().String()
	}
	
	errorJSON, _ := json.Marshal(errorData)
	
	return map[string]interface{}{
		"error_json": string(errorJSON),
	}
}

// GetName 获取格式化器名称
func (f *JSONErrorFormatter) GetName() string {
	return f.name
}

// SeverityErrorFilter 严重程度错误过滤器
type SeverityErrorFilter struct {
	name           string
	minSeverity    ErrorSeverity
	allowedCodes   map[ErrorCode]bool
	blockedCodes   map[ErrorCode]bool
	allowedPlugins map[string]bool
	blockedPlugins map[string]bool
}

// NewSeverityErrorFilter 创建严重程度错误过滤器
func NewSeverityErrorFilter(name string, minSeverity ErrorSeverity) ErrorFilter {
	return &SeverityErrorFilter{
		name:           name,
		minSeverity:    minSeverity,
		allowedCodes:   make(map[ErrorCode]bool),
		blockedCodes:   make(map[ErrorCode]bool),
		allowedPlugins: make(map[string]bool),
		blockedPlugins: make(map[string]bool),
	}
}

// ShouldLog 是否应该记录日志
func (f *SeverityErrorFilter) ShouldLog(err PluginError, pluginID string) bool {
	// 检查严重程度
	if err.GetSeverity() < f.minSeverity {
		return false
	}
	
	// 检查被阻止的错误代码
	if f.blockedCodes[err.GetCode()] {
		return false
	}
	
	// 检查允许的错误代码
	if len(f.allowedCodes) > 0 && !f.allowedCodes[err.GetCode()] {
		return false
	}
	
	// 检查被阻止的插件
	if f.blockedPlugins[pluginID] {
		return false
	}
	
	// 检查允许的插件
	if len(f.allowedPlugins) > 0 && !f.allowedPlugins[pluginID] {
		return false
	}
	
	return true
}

// GetName 获取过滤器名称
func (f *SeverityErrorFilter) GetName() string {
	return f.name
}

// ContextErrorEnricher 上下文错误丰富器
type ContextErrorEnricher struct {
	name string
}

// NewContextErrorEnricher 创建上下文错误丰富器
func NewContextErrorEnricher() ErrorEnricher {
	return &ContextErrorEnricher{
		name: "context",
	}
}

// Enrich 丰富错误信息
func (e *ContextErrorEnricher) Enrich(ctx context.Context, err PluginError, pluginID string) map[string]interface{} {
	fields := make(map[string]interface{})
	
	// 添加请求ID
	if requestID, ok := ctx.Value("request_id").(string); ok {
		fields["request_id"] = requestID
	}
	
	// 添加用户ID
	if userID, ok := ctx.Value("user_id").(string); ok {
		fields["user_id"] = userID
	}
	
	// 添加会话ID
	if sessionID, ok := ctx.Value("session_id").(string); ok {
		fields["session_id"] = sessionID
	}
	
	// 添加跟踪ID
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		fields["trace_id"] = traceID
	}
	
	// 添加系统信息
	fields["hostname"], _ = os.Hostname()
	fields["pid"] = os.Getpid()
	
	return fields
}

// GetName 获取丰富器名称
func (e *ContextErrorEnricher) GetName() string {
	return e.name
}