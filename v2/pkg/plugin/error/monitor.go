package plugin

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// MetricsCollector 指标收集器接口
type MetricsCollector interface {
	// IncrementCounter 增加计数器
	IncrementCounter(name string, labels map[string]string)
	// SetGauge 设置仪表盘值
	SetGauge(name string, value float64, labels map[string]string)
	// RecordHistogram 记录直方图
	RecordHistogram(name string, value float64, labels map[string]string)
	// RecordTimer 记录计时器
	RecordTimer(name string, duration time.Duration, labels map[string]string)
	// GetMetrics 获取指标
	GetMetrics() map[string]interface{}
	// Reset 重置指标
	Reset()
}

// ErrorMonitor 错误监控器接口
type ErrorMonitor interface {
	// RecordError 记录错误
	RecordError(ctx context.Context, err PluginError, pluginID string)
	// GetErrorStats 获取错误统计
	GetErrorStats(pluginID string) *ErrorStats
	// GetAllErrorStats 获取所有错误统计
	GetAllErrorStats() map[string]*ErrorStats
	// GetErrorRate 获取错误率
	GetErrorRate(pluginID string, window time.Duration) float64
	// GetMTTR 获取平均恢复时间
	GetMTTR(pluginID string) time.Duration
	// GetMTBF 获取平均故障间隔时间
	GetMTBF(pluginID string) time.Duration
	// SetAlertThreshold 设置告警阈值
	SetAlertThreshold(pluginID string, threshold AlertThreshold)
	// CheckAlerts 检查告警
	CheckAlerts() []Alert
	// RegisterAlertHandler 注册告警处理器
	RegisterAlertHandler(handler AlertHandler)
	// Start 启动监控
	Start(ctx context.Context) error
	// Stop 停止监控
	Stop() error
}

// AlertThreshold 告警阈值
type AlertThreshold struct {
	ErrorRate     float64       `json:"error_rate"`     // 错误率阈值
	ErrorCount    int64         `json:"error_count"`    // 错误数量阈值
	MTTR          time.Duration `json:"mttr"`           // MTTR阈值
	MTBF          time.Duration `json:"mtbf"`           // MTBF阈值
	TimeWindow    time.Duration `json:"time_window"`    // 时间窗口
	SeverityLevel ErrorSeverity `json:"severity_level"` // 严重程度阈值
	Enabled       bool          `json:"enabled"`        // 是否启用
}

// Alert 告警
type Alert struct {
	ID          string                 `json:"id"`          // 告警ID
	Type        AlertType              `json:"type"`        // 告警类型
	PluginID    string                 `json:"plugin_id"`   // 插件ID
	Message     string                 `json:"message"`     // 告警消息
	Severity    AlertSeverity          `json:"severity"`    // 告警严重程度
	Timestamp   time.Time              `json:"timestamp"`   // 时间戳
	Metadata    map[string]interface{} `json:"metadata"`    // 元数据
	Resolved    bool                   `json:"resolved"`    // 是否已解决
	ResolvedAt  *time.Time             `json:"resolved_at"` // 解决时间
}

// AlertType 告警类型
type AlertType int

const (
	AlertTypeErrorRate AlertType = iota
	AlertTypeErrorCount
	AlertTypeMTTR
	AlertTypeMTBF
	AlertTypePluginDown
	AlertTypeResourceExhausted
	AlertTypePerformanceDegraded
)

// String 返回告警类型的字符串表示
func (a AlertType) String() string {
	switch a {
	case AlertTypeErrorRate:
		return "error_rate"
	case AlertTypeErrorCount:
		return "error_count"
	case AlertTypeMTTR:
		return "mttr"
	case AlertTypeMTBF:
		return "mtbf"
	case AlertTypePluginDown:
		return "plugin_down"
	case AlertTypeResourceExhausted:
		return "resource_exhausted"
	case AlertTypePerformanceDegraded:
		return "performance_degraded"
	default:
		return "unknown"
	}
}

// AlertSeverity 告警严重程度
type AlertSeverity int

const (
	AlertSeverityLow AlertSeverity = iota
	AlertSeverityMedium
	AlertSeverityHigh
	AlertSeverityCritical
)

// String 返回告警严重程度的字符串表示
func (a AlertSeverity) String() string {
	switch a {
	case AlertSeverityLow:
		return "low"
	case AlertSeverityMedium:
		return "medium"
	case AlertSeverityHigh:
		return "high"
	case AlertSeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// AlertHandler 告警处理器接口
type AlertHandler interface {
	// HandleAlert 处理告警
	HandleAlert(ctx context.Context, alert Alert) error
	// GetName 获取处理器名称
	GetName() string
}

// DefaultMetricsCollector 默认指标收集器
type DefaultMetricsCollector struct {
	counters   map[string]*Counter
	gauges     map[string]*Gauge
	histograms map[string]*Histogram
	timers     map[string]*Timer
	mutex      sync.RWMutex
}

// Counter 计数器
type Counter struct {
	value  int64
	labels map[string]string
}

// Gauge 仪表盘
type Gauge struct {
	value  float64
	labels map[string]string
	mutex  sync.RWMutex
}

// Histogram 直方图
type Histogram struct {
	values []float64
	labels map[string]string
	mutex  sync.RWMutex
}

// Timer 计时器
type Timer struct {
	durations []time.Duration
	labels    map[string]string
	mutex     sync.RWMutex
}

// NewMetricsCollector 创建新的指标收集器
func NewMetricsCollector() MetricsCollector {
	return &DefaultMetricsCollector{
		counters:   make(map[string]*Counter),
		gauges:     make(map[string]*Gauge),
		histograms: make(map[string]*Histogram),
		timers:     make(map[string]*Timer),
	}
}

// IncrementCounter 增加计数器
func (mc *DefaultMetricsCollector) IncrementCounter(name string, labels map[string]string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	key := mc.buildKey(name, labels)
	if counter, exists := mc.counters[key]; exists {
		atomic.AddInt64(&counter.value, 1)
	} else {
		mc.counters[key] = &Counter{
			value:  1,
			labels: labels,
		}
	}
}

// SetGauge 设置仪表盘值
func (mc *DefaultMetricsCollector) SetGauge(name string, value float64, labels map[string]string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	key := mc.buildKey(name, labels)
	if gauge, exists := mc.gauges[key]; exists {
		gauge.mutex.Lock()
		gauge.value = value
		gauge.mutex.Unlock()
	} else {
		mc.gauges[key] = &Gauge{
			value:  value,
			labels: labels,
		}
	}
}

// RecordHistogram 记录直方图
func (mc *DefaultMetricsCollector) RecordHistogram(name string, value float64, labels map[string]string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	key := mc.buildKey(name, labels)
	if histogram, exists := mc.histograms[key]; exists {
		histogram.mutex.Lock()
		histogram.values = append(histogram.values, value)
		histogram.mutex.Unlock()
	} else {
		mc.histograms[key] = &Histogram{
			values: []float64{value},
			labels: labels,
		}
	}
}

// RecordTimer 记录计时器
func (mc *DefaultMetricsCollector) RecordTimer(name string, duration time.Duration, labels map[string]string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	key := mc.buildKey(name, labels)
	if timer, exists := mc.timers[key]; exists {
		timer.mutex.Lock()
		timer.durations = append(timer.durations, duration)
		timer.mutex.Unlock()
	} else {
		mc.timers[key] = &Timer{
			durations: []time.Duration{duration},
			labels:    labels,
		}
	}
}

// GetMetrics 获取指标
func (mc *DefaultMetricsCollector) GetMetrics() map[string]interface{} {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	
	metrics := make(map[string]interface{})
	
	// 收集计数器指标
	counters := make(map[string]interface{})
	for key, counter := range mc.counters {
		counters[key] = map[string]interface{}{
			"value": atomic.LoadInt64(&counter.value),
			"labels": counter.labels,
		}
	}
	metrics["counters"] = counters
	
	// 收集仪表盘指标
	gauges := make(map[string]interface{})
	for key, gauge := range mc.gauges {
		gauge.mutex.RLock()
		gauges[key] = map[string]interface{}{
			"value": gauge.value,
			"labels": gauge.labels,
		}
		gauge.mutex.RUnlock()
	}
	metrics["gauges"] = gauges
	
	// 收集直方图指标
	histograms := make(map[string]interface{})
	for key, histogram := range mc.histograms {
		histogram.mutex.RLock()
		values := make([]float64, len(histogram.values))
		copy(values, histogram.values)
		histograms[key] = map[string]interface{}{
			"values": values,
			"labels": histogram.labels,
		}
		histogram.mutex.RUnlock()
	}
	metrics["histograms"] = histograms
	
	// 收集计时器指标
	timers := make(map[string]interface{})
	for key, timer := range mc.timers {
		timer.mutex.RLock()
		durations := make([]time.Duration, len(timer.durations))
		copy(durations, timer.durations)
		timers[key] = map[string]interface{}{
			"durations": durations,
			"labels": timer.labels,
		}
		timer.mutex.RUnlock()
	}
	metrics["timers"] = timers
	
	return metrics
}

// Reset 重置指标
func (mc *DefaultMetricsCollector) Reset() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	mc.counters = make(map[string]*Counter)
	mc.gauges = make(map[string]*Gauge)
	mc.histograms = make(map[string]*Histogram)
	mc.timers = make(map[string]*Timer)
}

// buildKey 构建指标键
func (mc *DefaultMetricsCollector) buildKey(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}
	
	key := name
	for k, v := range labels {
		key += fmt.Sprintf("_%s_%s", k, v)
	}
	return key
}



// DefaultErrorMonitor 默认错误监控器
type DefaultErrorMonitor struct {
	errorStats      map[string]*ErrorStats
	alertThresholds map[string]AlertThreshold
	alertHandlers   []AlertHandler
	metrics         MetricsCollector
	logger          Logger
	mutex           sync.RWMutex
	running         bool
	stopChan        chan struct{}
	wg              sync.WaitGroup
}

// NewErrorMonitor 创建新的错误监控器
func NewErrorMonitor(metrics MetricsCollector, logger Logger) ErrorMonitor {
	return &DefaultErrorMonitor{
		errorStats:      make(map[string]*ErrorStats),
		alertThresholds: make(map[string]AlertThreshold),
		alertHandlers:   make([]AlertHandler, 0),
		metrics:         metrics,
		logger:          logger,
		stopChan:        make(chan struct{}),
	}
}

// RecordError 记录错误
func (em *DefaultErrorMonitor) RecordError(ctx context.Context, err PluginError, pluginID string) {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	
	// 获取或创建错误统计
	stats, exists := em.errorStats[pluginID]
	if !exists {
		stats = &ErrorStats{
			TotalErrors:  0,
			ErrorsByType: make(map[ErrorType]int),
			ErrorRate:    0.0,
			LastError:    nil,
			FirstError:   nil,
		}
		em.errorStats[pluginID] = stats
	}
	
	// 更新统计
	stats.TotalErrors++
	stats.ErrorsByType[err.GetType()]++
	
	// 创建错误记录
	errorRecord := &ErrorRecord{
		Timestamp: time.Now(),
		Error:     err,
		ErrorType: err.GetType(),
	}
	
	stats.LastError = errorRecord
	if stats.FirstError == nil {
		stats.FirstError = errorRecord
	}
	
	// 更新指标
	if em.metrics != nil {
		em.metrics.IncrementCounter("plugin_errors_total", map[string]string{
			"plugin_id": pluginID,
			"error_code": err.GetCode().String(),
			"error_type": err.GetType().String(),
			"severity": err.GetSeverity().String(),
		})
	}
	
	// 记录日志
	if em.logger != nil {
		em.logger.LogError(ctx, err, pluginID)
	}
}

// GetErrorStats 获取错误统计
func (em *DefaultErrorMonitor) GetErrorStats(pluginID string) *ErrorStats {
	em.mutex.RLock()
	defer em.mutex.RUnlock()
	
	if stats, exists := em.errorStats[pluginID]; exists {
		// 深拷贝
		statsCopy := &ErrorStats{
			TotalErrors:  stats.TotalErrors,
			ErrorsByType: make(map[ErrorType]int),
			ErrorRate:    stats.ErrorRate,
		}
		
		// 拷贝ErrorsByType map
		for k, v := range stats.ErrorsByType {
			statsCopy.ErrorsByType[k] = v
		}
		
		// 深拷贝LastError
		if stats.LastError != nil {
			statsCopy.LastError = &ErrorRecord{
				Timestamp: stats.LastError.Timestamp,
				Error:     stats.LastError.Error,
				ErrorType: stats.LastError.ErrorType,
				Context:   make(map[string]interface{}),
			}
			for k, v := range stats.LastError.Context {
				statsCopy.LastError.Context[k] = v
			}
		}
		
		// 深拷贝FirstError
		if stats.FirstError != nil {
			statsCopy.FirstError = &ErrorRecord{
				Timestamp: stats.FirstError.Timestamp,
				Error:     stats.FirstError.Error,
				ErrorType: stats.FirstError.ErrorType,
				Context:   make(map[string]interface{}),
			}
			for k, v := range stats.FirstError.Context {
				statsCopy.FirstError.Context[k] = v
			}
		}
		
		return statsCopy
	}
	
	return nil
}

// GetAllErrorStats 获取所有错误统计
func (em *DefaultErrorMonitor) GetAllErrorStats() map[string]*ErrorStats {
	em.mutex.RLock()
	defer em.mutex.RUnlock()
	
	allStats := make(map[string]*ErrorStats)
	for pluginID, stats := range em.errorStats {
		// 深拷贝
		statsCopy := &ErrorStats{
			TotalErrors:  stats.TotalErrors,
			ErrorsByType: make(map[ErrorType]int),
			ErrorRate:    stats.ErrorRate,
		}
		
		// 拷贝ErrorsByType map
		for k, v := range stats.ErrorsByType {
			statsCopy.ErrorsByType[k] = v
		}
		
		// 深拷贝LastError
		if stats.LastError != nil {
			statsCopy.LastError = &ErrorRecord{
				Timestamp: stats.LastError.Timestamp,
				Error:     stats.LastError.Error,
				ErrorType: stats.LastError.ErrorType,
				Context:   make(map[string]interface{}),
			}
			for k, v := range stats.LastError.Context {
				statsCopy.LastError.Context[k] = v
			}
		}
		
		// 深拷贝FirstError
		if stats.FirstError != nil {
			statsCopy.FirstError = &ErrorRecord{
				Timestamp: stats.FirstError.Timestamp,
				Error:     stats.FirstError.Error,
				ErrorType: stats.FirstError.ErrorType,
				Context:   make(map[string]interface{}),
			}
			for k, v := range stats.FirstError.Context {
				statsCopy.FirstError.Context[k] = v
			}
		}
		
		allStats[pluginID] = statsCopy
	}
	
	return allStats
}

// GetErrorRate 获取错误率
func (em *DefaultErrorMonitor) GetErrorRate(pluginID string, window time.Duration) float64 {
	// 简化实现，实际应该基于时间窗口计算
	stats := em.GetErrorStats(pluginID)
	if stats == nil {
		return 0.0
	}
	
	return stats.ErrorRate
}

// GetMTTR 获取平均恢复时间
func (em *DefaultErrorMonitor) GetMTTR(pluginID string) time.Duration {
	// 简化实现，实际应该基于历史数据计算
	return time.Minute // 默认返回1分钟
}

// GetMTBF 获取平均故障间隔时间
func (em *DefaultErrorMonitor) GetMTBF(pluginID string) time.Duration {
	// 简化实现，实际应该基于历史数据计算
	return time.Hour // 默认返回1小时
}

// SetAlertThreshold 设置告警阈值
func (em *DefaultErrorMonitor) SetAlertThreshold(pluginID string, threshold AlertThreshold) {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	
	em.alertThresholds[pluginID] = threshold
}

// CheckAlerts 检查告警
func (em *DefaultErrorMonitor) CheckAlerts() []Alert {
	em.mutex.RLock()
	defer em.mutex.RUnlock()
	
	var alerts []Alert
	
	for pluginID, threshold := range em.alertThresholds {
		if !threshold.Enabled {
			continue
		}
		
		stats := em.errorStats[pluginID]
		if stats == nil {
			continue
		}
		
		// 检查错误率
		if threshold.ErrorRate > 0 && stats.ErrorRate > threshold.ErrorRate {
			alerts = append(alerts, Alert{
				ID:        fmt.Sprintf("%s_error_rate_%d", pluginID, time.Now().Unix()),
				Type:      AlertTypeErrorRate,
				PluginID:  pluginID,
				Message:   fmt.Sprintf("Error rate %.2f%% exceeds threshold %.2f%%", stats.ErrorRate*100, threshold.ErrorRate*100),
				Severity:  AlertSeverityHigh,
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"current_rate": stats.ErrorRate,
					"threshold":    threshold.ErrorRate,
				},
			})
		}
		
		// 检查错误数量
		if threshold.ErrorCount > 0 && int64(stats.TotalErrors) > threshold.ErrorCount {
			alerts = append(alerts, Alert{
				ID:        fmt.Sprintf("%s_error_count_%d", pluginID, time.Now().Unix()),
				Type:      AlertTypeErrorCount,
				PluginID:  pluginID,
				Message:   fmt.Sprintf("Error count %d exceeds threshold %d", stats.TotalErrors, threshold.ErrorCount),
				Severity:  AlertSeverityHigh,
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"current_count": stats.TotalErrors,
					"threshold":     threshold.ErrorCount,
				},
			})
		}
		
		// 检查严重程度级别
		if stats.LastError != nil && stats.LastError.Error != nil {
			if pluginErr, ok := stats.LastError.Error.(PluginError); ok {
				if pluginErr.GetSeverity() >= threshold.SeverityLevel {
					alerts = append(alerts, Alert{
						ID:        fmt.Sprintf("%s_severity_%d", pluginID, time.Now().Unix()),
						Type:      AlertTypeErrorCount, // 使用通用类型，因为AlertTypeSeverityLevel未定义
						PluginID:  pluginID,
						Message:   fmt.Sprintf("Error severity %s meets or exceeds threshold %s", pluginErr.GetSeverity().String(), threshold.SeverityLevel.String()),
						Severity:  em.mapErrorSeverityToAlertSeverity(pluginErr.GetSeverity()),
						Timestamp: time.Now(),
						Metadata: map[string]interface{}{
							"current_severity": pluginErr.GetSeverity().String(),
							"threshold":        threshold.SeverityLevel.String(),
						},
					})
				}
			}
		}
	}
	
	return alerts
}

// RegisterAlertHandler 注册告警处理器
func (em *DefaultErrorMonitor) RegisterAlertHandler(handler AlertHandler) {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	
	em.alertHandlers = append(em.alertHandlers, handler)
}

// Start 启动监控
func (em *DefaultErrorMonitor) Start(ctx context.Context) error {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	
	if em.running {
		return fmt.Errorf("error monitor is already running")
	}
	
	em.running = true
	// 创建新的stopChan，以防之前被关闭
	em.stopChan = make(chan struct{})
	em.wg.Add(1)
	
	go em.monitorLoop(ctx)
	
	if em.logger != nil {
		em.logger.Info("Error monitor started", nil)
	}
	
	return nil
}

// Stop 停止监控
func (em *DefaultErrorMonitor) Stop() error {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	
	if !em.running {
		return fmt.Errorf("error monitor is not running")
	}
	
	em.running = false
	
	// 安全关闭stopChan
	select {
	case <-em.stopChan:
		// 已经关闭
	default:
		close(em.stopChan)
	}
	
	em.wg.Wait()
	
	if em.logger != nil {
		em.logger.Info("Error monitor stopped", nil)
	}
	
	return nil
}

// monitorLoop 监控循环
func (em *DefaultErrorMonitor) monitorLoop(ctx context.Context) {
	defer em.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-em.stopChan:
			return
		case <-ticker.C:
			// 检查告警
			alerts := em.CheckAlerts()
			for _, alert := range alerts {
				// 处理告警
				for _, handler := range em.alertHandlers {
					if err := handler.HandleAlert(ctx, alert); err != nil && em.logger != nil {
						em.logger.Error("Failed to handle alert", map[string]interface{}{
							"alert_id": alert.ID,
							"handler":  handler.GetName(),
							"error":    err.Error(),
						})
					}
				}
			}
		}
	}
}

// LogAlertHandler 日志告警处理器
type LogAlertHandler struct {
	name   string
	logger Logger
}

// NewLogAlertHandler 创建日志告警处理器
func NewLogAlertHandler(name string, logger Logger) AlertHandler {
	return &LogAlertHandler{
		name:   name,
		logger: logger,
	}
}

// HandleAlert 处理告警
func (h *LogAlertHandler) HandleAlert(ctx context.Context, alert Alert) error {
	h.logger.Error("Alert triggered", map[string]interface{}{
		"alert_id":   alert.ID,
		"alert_type": alert.Type.String(),
		"plugin_id":  alert.PluginID,
		"message":    alert.Message,
		"severity":   alert.Severity.String(),
		"timestamp":  alert.Timestamp.Format(time.RFC3339),
		"metadata":   alert.Metadata,
	})
	
	return nil
}

// GetName 获取处理器名称
func (h *LogAlertHandler) GetName() string {
	return h.name
}

// mapErrorSeverityToAlertSeverity 映射错误严重程度到告警严重程度
func (em *DefaultErrorMonitor) mapErrorSeverityToAlertSeverity(errorSeverity ErrorSeverity) AlertSeverity {
	switch errorSeverity {
	case ErrorSeverityTrace, ErrorSeverityDebug, ErrorSeverityInfo:
		return AlertSeverityLow
	case ErrorSeverityWarning:
		return AlertSeverityMedium
	case ErrorSeverityError:
		return AlertSeverityHigh
	case ErrorSeverityFatal, ErrorSeverityCritical:
		return AlertSeverityCritical
	default:
		return AlertSeverityMedium
	}
}