package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime"
	"sort"
	"sync"
	"time"
)

// EventMonitor 事件监控器接口
type EventMonitor interface {
	// 监控控制
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool

	// 统计信息
	GetStats() *EnhancedEventStats
	GetPerformanceMetrics() *PerformanceMetrics
	GetHealthStatus() *MonitorHealthStatus

	// 实时监控
	GetRealTimeStats(duration time.Duration) *RealTimeStats
	GetEventRateStats(duration time.Duration) *EventRateStats
	GetLatencyStats() *LatencyStats

	// 告警和通知
	SetThresholds(thresholds *MonitorThresholds)
	GetActiveAlerts() []*MonitorAlert
	SubscribeToAlerts(handler AlertHandler) string
	UnsubscribeFromAlerts(subscriptionID string)

	// 数据导出
	ExportStats(format ExportFormat) ([]byte, error)
	ExportMetrics(format ExportFormat, duration time.Duration) ([]byte, error)

	// 事件记录
	RecordEvent(event Event)
}

// EnhancedEventStats 增强的事件统计信息
type EnhancedEventStats struct {
	*EventStats

	// 性能指标
	AverageProcessingTime time.Duration            `json:"average_processing_time"`
	MaxProcessingTime     time.Duration            `json:"max_processing_time"`
	MinProcessingTime     time.Duration            `json:"min_processing_time"`
	Throughput           float64                   `json:"throughput"` // 事件/秒
	LatencyP50           time.Duration            `json:"latency_p50"`
	LatencyP95           time.Duration            `json:"latency_p95"`
	LatencyP99           time.Duration            `json:"latency_p99"`

	// 错误统计
	ErrorRate            float64                   `json:"error_rate"`
	ErrorsByType         map[EventType]int64       `json:"errors_by_type"`
	ErrorsBySource       map[string]int64          `json:"errors_by_source"`

	// 队列统计
	QueueSize            int                       `json:"queue_size"`
	QueueCapacity        int                       `json:"queue_capacity"`
	QueueUtilization     float64                   `json:"queue_utilization"`
	MaxQueueSize         int                       `json:"max_queue_size"`

	// 工作池统计
	ActiveWorkers        int                       `json:"active_workers"`
	MaxWorkers           int                       `json:"max_workers"`
	WorkerUtilization    float64                   `json:"worker_utilization"`
	IdleWorkers          int                       `json:"idle_workers"`

	// 时间窗口统计
	EventsPerMinute      float64                   `json:"events_per_minute"`
	EventsPerHour        float64                   `json:"events_per_hour"`
	PeakEventsPerMinute  float64                   `json:"peak_events_per_minute"`

	// 订阅者统计
	ActiveSubscriptions  int                       `json:"active_subscriptions"`
	SubscriptionsByType  map[EventType]int         `json:"subscriptions_by_type"`
	SubscriptionsByGroup map[string]int            `json:"subscriptions_by_group"`
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	CPUUsage        float64           `json:"cpu_usage"`
	MemoryUsage     int64             `json:"memory_usage"`
	GoroutineCount  int               `json:"goroutine_count"`
	GCPauseTime     time.Duration     `json:"gc_pause_time"`
	AllocatedMemory int64             `json:"allocated_memory"`
	HeapSize        int64             `json:"heap_size"`
	CollectedAt     time.Time         `json:"collected_at"`
}

// MonitorHealthStatus 监控健康状态
type MonitorHealthStatus struct {
	OverallHealth   HealthLevel       `json:"overall_health"`
	Components      map[string]HealthLevel `json:"components"`
	Issues          []HealthIssue     `json:"issues"`
	LastCheckTime   time.Time         `json:"last_check_time"`
	Uptime          time.Duration     `json:"uptime"`
}

// HealthLevel 健康级别
type HealthLevel int

const (
	HealthLevelHealthy HealthLevel = iota
	HealthLevelWarning
	HealthLevelCritical
	HealthLevelUnknown
)

func (h HealthLevel) String() string {
	switch h {
	case HealthLevelHealthy:
		return "healthy"
	case HealthLevelWarning:
		return "warning"
	case HealthLevelCritical:
		return "critical"
	case HealthLevelUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

// HealthIssue 健康问题
type HealthIssue struct {
	Component   string      `json:"component"`
	Level       HealthLevel `json:"level"`
	Message     string      `json:"message"`
	DetectedAt  time.Time   `json:"detected_at"`
	ResolvedAt  *time.Time  `json:"resolved_at,omitempty"`
}

// RealTimeStats 实时统计
type RealTimeStats struct {
	TimeWindow      time.Duration     `json:"time_window"`
	EventCount      int64             `json:"event_count"`
	EventRate       float64           `json:"event_rate"`
	ErrorCount      int64             `json:"error_count"`
	ErrorRate       float64           `json:"error_rate"`
	AverageLatency  time.Duration     `json:"average_latency"`
	Throughput      float64           `json:"throughput"`
	CollectedAt     time.Time         `json:"collected_at"`
}

// EventRateStats 事件速率统计
type EventRateStats struct {
	TimeWindow      time.Duration     `json:"time_window"`
	CurrentRate     float64           `json:"current_rate"`
	AverageRate     float64           `json:"average_rate"`
	PeakRate        float64           `json:"peak_rate"`
	MinRate         float64           `json:"min_rate"`
	RateHistory     []RateDataPoint   `json:"rate_history"`
	CollectedAt     time.Time         `json:"collected_at"`
}

// RateDataPoint 速率数据点
type RateDataPoint struct {
	Timestamp   time.Time `json:"timestamp"`
	Rate        float64   `json:"rate"`
	EventCount  int64     `json:"event_count"`
}

// LatencyStats 延迟统计
type LatencyStats struct {
	AverageLatency  time.Duration     `json:"average_latency"`
	MinLatency      time.Duration     `json:"min_latency"`
	MaxLatency      time.Duration     `json:"max_latency"`
	P50Latency      time.Duration     `json:"p50_latency"`
	P95Latency      time.Duration     `json:"p95_latency"`
	P99Latency      time.Duration     `json:"p99_latency"`
	LatencyHistory  []LatencyDataPoint `json:"latency_history"`
	CollectedAt     time.Time         `json:"collected_at"`
}

// LatencyDataPoint 延迟数据点
type LatencyDataPoint struct {
	Timestamp time.Time     `json:"timestamp"`
	Latency   time.Duration `json:"latency"`
	EventType EventType     `json:"event_type"`
}

// MonitorThresholds 监控阈值
type MonitorThresholds struct {
	MaxEventRate        float64       `json:"max_event_rate"`
	MaxErrorRate        float64       `json:"max_error_rate"`
	MaxLatency          time.Duration `json:"max_latency"`
	MaxQueueUtilization float64       `json:"max_queue_utilization"`
	MaxWorkerUtilization float64      `json:"max_worker_utilization"`
	MaxMemoryUsage      int64         `json:"max_memory_usage"`
	MinThroughput       float64       `json:"min_throughput"`
}

// MonitorAlert 监控告警
type MonitorAlert struct {
	ID          string        `json:"id"`
	Level       AlertLevel    `json:"level"`
	Type        AlertType     `json:"type"`
	Message     string        `json:"message"`
	Metric      string        `json:"metric"`
	Threshold   interface{}   `json:"threshold"`
	CurrentValue interface{}  `json:"current_value"`
	TriggeredAt time.Time     `json:"triggered_at"`
	ResolvedAt  *time.Time    `json:"resolved_at,omitempty"`
	Count       int           `json:"count"`
}

// AlertLevel 告警级别
type AlertLevel int

const (
	AlertLevelInfo AlertLevel = iota
	AlertLevelWarning
	AlertLevelCritical
)

func (a AlertLevel) String() string {
	switch a {
	case AlertLevelInfo:
		return "info"
	case AlertLevelWarning:
		return "warning"
	case AlertLevelCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// AlertType 告警类型
type AlertType int

const (
	AlertTypeEventRate AlertType = iota
	AlertTypeErrorRate
	AlertTypeLatency
	AlertTypeQueueUtilization
	AlertTypeWorkerUtilization
	AlertTypeMemoryUsage
	AlertTypeThroughput
)

func (a AlertType) String() string {
	switch a {
	case AlertTypeEventRate:
		return "event_rate"
	case AlertTypeErrorRate:
		return "error_rate"
	case AlertTypeLatency:
		return "latency"
	case AlertTypeQueueUtilization:
		return "queue_utilization"
	case AlertTypeWorkerUtilization:
		return "worker_utilization"
	case AlertTypeMemoryUsage:
		return "memory_usage"
	case AlertTypeThroughput:
		return "throughput"
	default:
		return "unknown"
	}
}

// AlertHandler 告警处理器
type AlertHandler func(alert *MonitorAlert)

// ExportFormat 导出格式
type ExportFormat int

const (
	ExportFormatJSON ExportFormat = iota
	ExportFormatCSV
	ExportFormatPrometheus
)

// DefaultEventMonitor 默认事件监控器实现
type DefaultEventMonitor struct {
	logger    *slog.Logger
	eventBus  EventBus

	// 监控状态
	running   bool
	startTime time.Time
	mutex     sync.RWMutex

	// 统计数据
	stats           *EnhancedEventStats
	performanceMetrics *PerformanceMetrics
	healthStatus    *MonitorHealthStatus
	latencyHistory  []LatencyDataPoint
	rateHistory     []RateDataPoint

	// 阈值和告警
	thresholds      *MonitorThresholds
	activeAlerts    map[string]*MonitorAlert
	alertHandlers   map[string]AlertHandler
	alertMutex      sync.RWMutex

	// 控制信号
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// 数据收集
	metricsInterval time.Duration
	historySize     int
}

// NewDefaultEventMonitor 创建默认事件监控器
func NewDefaultEventMonitor(logger *slog.Logger, eventBus EventBus) *DefaultEventMonitor {
	return &DefaultEventMonitor{
		logger:   logger,
		eventBus: eventBus,
		stats: &EnhancedEventStats{
			EventStats: &EventStats{
				EventCounts: make(map[EventType]int64),
				ErrorCounts: make(map[EventType]int64),
			},
			ErrorsByType:         make(map[EventType]int64),
			ErrorsBySource:       make(map[string]int64),
			SubscriptionsByType:  make(map[EventType]int),
			SubscriptionsByGroup: make(map[string]int),
		},
		performanceMetrics: &PerformanceMetrics{},
		healthStatus: &MonitorHealthStatus{
			OverallHealth: HealthLevelHealthy,
			Components:    make(map[string]HealthLevel),
			Issues:        make([]HealthIssue, 0),
		},
		activeAlerts:    make(map[string]*MonitorAlert),
		alertHandlers:   make(map[string]AlertHandler),
		latencyHistory:  make([]LatencyDataPoint, 0),
		rateHistory:     make([]RateDataPoint, 0),
		metricsInterval: time.Second * 10,
		historySize:     1000,
		thresholds: &MonitorThresholds{
			MaxEventRate:         1000.0,
			MaxErrorRate:         0.05, // 5%
			MaxLatency:           time.Second,
			MaxQueueUtilization:  0.8, // 80%
			MaxWorkerUtilization: 0.9, // 90%
			MaxMemoryUsage:       1024 * 1024 * 1024, // 1GB
			MinThroughput:        10.0,
		},
	}
}

// Start 启动监控器
func (m *DefaultEventMonitor) Start(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.running {
		return fmt.Errorf("event monitor is already running")
	}

	m.ctx, m.cancel = context.WithCancel(ctx)
	m.startTime = time.Now()
	m.running = true

	// 启动监控协程
	m.wg.Add(3)
	go m.metricsCollector()
	go m.healthChecker()
	go m.alertManager()

	m.logger.Info("Event monitor started")
	return nil
}

// Stop 停止监控器
func (m *DefaultEventMonitor) Stop(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.running {
		return fmt.Errorf("event monitor is not running")
	}

	m.cancel()
	m.wg.Wait()
	m.running = false

	m.logger.Info("Event monitor stopped")
	return nil
}

// IsRunning 检查监控器是否运行中
func (m *DefaultEventMonitor) IsRunning() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.running
}

// GetStats 获取增强统计信息
func (m *DefaultEventMonitor) GetStats() *EnhancedEventStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 复制统计信息
	stats := *m.stats
	stats.EventStats = &EventStats{
		TotalEvents:      m.stats.EventStats.TotalEvents,
		TotalSubscribers: m.stats.EventStats.TotalSubscribers,
		EventCounts:      make(map[EventType]int64),
		ErrorCounts:      make(map[EventType]int64),
		LastEventTime:    m.stats.EventStats.LastEventTime,
	}

	for k, v := range m.stats.EventStats.EventCounts {
		stats.EventStats.EventCounts[k] = v
	}
	for k, v := range m.stats.EventStats.ErrorCounts {
		stats.EventStats.ErrorCounts[k] = v
	}

	return &stats
}

// GetPerformanceMetrics 获取性能指标
func (m *DefaultEventMonitor) GetPerformanceMetrics() *PerformanceMetrics {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	metrics := *m.performanceMetrics
	return &metrics
}

// GetHealthStatus 获取健康状态
func (m *DefaultEventMonitor) GetHealthStatus() *MonitorHealthStatus {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	status := *m.healthStatus
	status.Components = make(map[string]HealthLevel)
	for k, v := range m.healthStatus.Components {
		status.Components[k] = v
	}

	status.Issues = make([]HealthIssue, len(m.healthStatus.Issues))
	copy(status.Issues, m.healthStatus.Issues)

	// 更新运行时间
	if m.running && !m.startTime.IsZero() {
		status.Uptime = time.Since(m.startTime)
	}

	return &status
}

// GetRealTimeStats 获取实时统计
func (m *DefaultEventMonitor) GetRealTimeStats(duration time.Duration) *RealTimeStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	now := time.Now()
	cutoff := now.Add(-duration)

	var eventCount, errorCount int64
	var totalLatency time.Duration
	var latencyCount int

	// 从历史数据中计算实时统计
	for _, point := range m.rateHistory {
		if point.Timestamp.After(cutoff) {
			eventCount += point.EventCount
		}
	}

	for _, point := range m.latencyHistory {
		if point.Timestamp.After(cutoff) {
			totalLatency += point.Latency
			latencyCount++
		}
	}

	var averageLatency time.Duration
	if latencyCount > 0 {
		averageLatency = totalLatency / time.Duration(latencyCount)
	}

	eventRate := float64(eventCount) / duration.Seconds()
	errorRate := float64(errorCount) / float64(eventCount)
	if eventCount == 0 {
		errorRate = 0
	}

	return &RealTimeStats{
		TimeWindow:     duration,
		EventCount:     eventCount,
		EventRate:      eventRate,
		ErrorCount:     errorCount,
		ErrorRate:      errorRate,
		AverageLatency: averageLatency,
		Throughput:     eventRate,
		CollectedAt:    now,
	}
}

// GetEventRateStats 获取事件速率统计
func (m *DefaultEventMonitor) GetEventRateStats(duration time.Duration) *EventRateStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	now := time.Now()
	cutoff := now.Add(-duration)

	var rates []float64
	var filteredHistory []RateDataPoint

	for _, point := range m.rateHistory {
		if point.Timestamp.After(cutoff) {
			rates = append(rates, point.Rate)
			filteredHistory = append(filteredHistory, point)
		}
	}

	if len(rates) == 0 {
		return &EventRateStats{
			TimeWindow:  duration,
			CurrentRate: 0,
			AverageRate: 0,
			PeakRate:    0,
			MinRate:     0,
			RateHistory: make([]RateDataPoint, 0),
			CollectedAt: now,
		}
	}

	// 计算统计值
	sort.Float64s(rates)
	currentRate := rates[len(rates)-1]
	peakRate := rates[len(rates)-1]
	minRate := rates[0]

	var totalRate float64
	for _, rate := range rates {
		totalRate += rate
	}
	averageRate := totalRate / float64(len(rates))

	return &EventRateStats{
		TimeWindow:  duration,
		CurrentRate: currentRate,
		AverageRate: averageRate,
		PeakRate:    peakRate,
		MinRate:     minRate,
		RateHistory: filteredHistory,
		CollectedAt: now,
	}
}

// GetLatencyStats 获取延迟统计
func (m *DefaultEventMonitor) GetLatencyStats() *LatencyStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if len(m.latencyHistory) == 0 {
		return &LatencyStats{
			AverageLatency: 0,
			MinLatency:     0,
			MaxLatency:     0,
			P50Latency:     0,
			P95Latency:     0,
			P99Latency:     0,
			LatencyHistory: make([]LatencyDataPoint, 0),
			CollectedAt:    time.Now(),
		}
	}

	// 提取延迟值并排序
	latencies := make([]time.Duration, len(m.latencyHistory))
	for i, point := range m.latencyHistory {
		latencies[i] = point.Latency
	}
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	// 计算统计值
	minLatency := latencies[0]
	maxLatency := latencies[len(latencies)-1]

	var totalLatency time.Duration
	for _, latency := range latencies {
		totalLatency += latency
	}
	averageLatency := totalLatency / time.Duration(len(latencies))

	// 计算百分位数
	p50Index := len(latencies) * 50 / 100
	p95Index := len(latencies) * 95 / 100
	p99Index := len(latencies) * 99 / 100

	p50Latency := latencies[p50Index]
	p95Latency := latencies[p95Index]
	p99Latency := latencies[p99Index]

	return &LatencyStats{
		AverageLatency: averageLatency,
		MinLatency:     minLatency,
		MaxLatency:     maxLatency,
		P50Latency:     p50Latency,
		P95Latency:     p95Latency,
		P99Latency:     p99Latency,
		LatencyHistory: m.latencyHistory,
		CollectedAt:    time.Now(),
	}
}

// SetThresholds 设置监控阈值
func (m *DefaultEventMonitor) SetThresholds(thresholds *MonitorThresholds) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.thresholds = thresholds
}

// GetActiveAlerts 获取活跃告警
func (m *DefaultEventMonitor) GetActiveAlerts() []*MonitorAlert {
	m.alertMutex.RLock()
	defer m.alertMutex.RUnlock()

	alerts := make([]*MonitorAlert, 0, len(m.activeAlerts))
	for _, alert := range m.activeAlerts {
		if alert.ResolvedAt == nil {
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// SubscribeToAlerts 订阅告警
func (m *DefaultEventMonitor) SubscribeToAlerts(handler AlertHandler) string {
	m.alertMutex.Lock()
	defer m.alertMutex.Unlock()

	subscriptionID := fmt.Sprintf("alert_sub_%d", time.Now().UnixNano())
	m.alertHandlers[subscriptionID] = handler
	return subscriptionID
}

// UnsubscribeFromAlerts 取消告警订阅
func (m *DefaultEventMonitor) UnsubscribeFromAlerts(subscriptionID string) {
	m.alertMutex.Lock()
	defer m.alertMutex.Unlock()
	delete(m.alertHandlers, subscriptionID)
}

// ExportStats 导出统计信息
func (m *DefaultEventMonitor) ExportStats(format ExportFormat) ([]byte, error) {
	stats := m.GetStats()

	switch format {
	case ExportFormatJSON:
		return m.exportJSON(stats)
	case ExportFormatCSV:
		return m.exportCSV(stats)
	case ExportFormatPrometheus:
		return m.exportPrometheus(stats)
	default:
		return nil, fmt.Errorf("unsupported export format: %d", format)
	}
}

// ExportMetrics 导出性能指标
func (m *DefaultEventMonitor) ExportMetrics(format ExportFormat, duration time.Duration) ([]byte, error) {
	metrics := m.GetPerformanceMetrics()
	realTimeStats := m.GetRealTimeStats(duration)

	data := map[string]interface{}{
		"performance_metrics": metrics,
		"real_time_stats":     realTimeStats,
		"exported_at":         time.Now(),
	}

	switch format {
	case ExportFormatJSON:
		return m.exportJSON(data)
	default:
		return nil, fmt.Errorf("unsupported export format for metrics: %d", format)
	}
}

// RecordEvent 记录事件
func (m *DefaultEventMonitor) RecordEvent(event Event) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 更新事件计数
	m.stats.TotalEvents++
	m.stats.EventCounts[event.GetType()]++
	m.stats.LastEventTime = event.GetTimestamp()

	// 记录延迟数据点
	latency := time.Since(event.GetTimestamp())
	latencyPoint := LatencyDataPoint{
		Timestamp: time.Now(),
		Latency:   latency,
		EventType: event.GetType(),
	}
	m.latencyHistory = append(m.latencyHistory, latencyPoint)
	if len(m.latencyHistory) > m.historySize {
		m.latencyHistory = m.latencyHistory[1:]
	}

	// 更新延迟统计
	if m.stats.MinProcessingTime == 0 || latency < m.stats.MinProcessingTime {
		m.stats.MinProcessingTime = latency
	}
	if latency > m.stats.MaxProcessingTime {
		m.stats.MaxProcessingTime = latency
	}
}

// 内部方法

// metricsCollector 指标收集器
func (m *DefaultEventMonitor) metricsCollector() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.metricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.collectMetrics()
		}
	}
}

// healthChecker 健康检查器
func (m *DefaultEventMonitor) healthChecker() {
	defer m.wg.Done()

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkHealth()
		}
	}
}

// alertManager 告警管理器
func (m *DefaultEventMonitor) alertManager() {
	defer m.wg.Done()

	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkAlerts()
		}
	}
}

// collectMetrics 收集指标
func (m *DefaultEventMonitor) collectMetrics() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()

	// 更新基础统计
	if m.eventBus != nil {
		busStats := m.eventBus.GetStats()
		m.stats.EventStats = busStats
		m.stats.TotalSubscribers = m.eventBus.GetTotalSubscriptions()
	}

	// 计算速率
	if len(m.rateHistory) > 0 {
		lastPoint := m.rateHistory[len(m.rateHistory)-1]
		timeDiff := now.Sub(lastPoint.Timestamp).Seconds()
		eventDiff := m.stats.TotalEvents - lastPoint.EventCount
		if timeDiff > 0 {
			currentRate := float64(eventDiff) / timeDiff
			m.addRateDataPoint(RateDataPoint{
				Timestamp:  now,
				Rate:       currentRate,
				EventCount: m.stats.TotalEvents,
			})
		}
	} else {
		m.addRateDataPoint(RateDataPoint{
			Timestamp:  now,
			Rate:       0,
			EventCount: m.stats.TotalEvents,
		})
	}

	// 更新性能指标
	m.updatePerformanceMetrics()
}

// checkHealth 检查健康状态
func (m *DefaultEventMonitor) checkHealth() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.healthStatus.LastCheckTime = time.Now()
	m.healthStatus.Uptime = time.Since(m.startTime)

	// 检查各组件健康状态
	m.healthStatus.Components["event_bus"] = m.checkEventBusHealth()
	m.healthStatus.Components["memory"] = m.checkMemoryHealth()
	m.healthStatus.Components["performance"] = m.checkPerformanceHealth()

	// 计算整体健康状态
	m.healthStatus.OverallHealth = m.calculateOverallHealth()
}

// checkAlerts 检查告警
func (m *DefaultEventMonitor) checkAlerts() {
	stats := m.GetStats()
	realTimeStats := m.GetRealTimeStats(time.Minute * 5)

	// 检查各种阈值
	m.checkEventRateAlert(realTimeStats.EventRate)
	m.checkErrorRateAlert(realTimeStats.ErrorRate)
	m.checkLatencyAlert(realTimeStats.AverageLatency)
	m.checkQueueUtilizationAlert(stats.QueueUtilization)
	m.checkWorkerUtilizationAlert(stats.WorkerUtilization)
	m.checkThroughputAlert(realTimeStats.Throughput)
}

// 辅助方法实现省略...
// 这里只展示核心结构，实际实现会包含所有辅助方法

func (m *DefaultEventMonitor) exportJSON(data interface{}) ([]byte, error) {
	// JSON导出实现
	return json.Marshal(data)
}

func (m *DefaultEventMonitor) exportCSV(data interface{}) ([]byte, error) {
	// CSV导出实现 - 简单实现
	return []byte("CSV export not fully implemented"), nil
}

func (m *DefaultEventMonitor) exportPrometheus(data interface{}) ([]byte, error) {
	// Prometheus导出实现 - 简单实现
	return []byte("# Prometheus export not fully implemented"), nil
}

func (m *DefaultEventMonitor) addRateDataPoint(point RateDataPoint) {
	m.rateHistory = append(m.rateHistory, point)
	if len(m.rateHistory) > m.historySize {
		m.rateHistory = m.rateHistory[1:]
	}
}

func (m *DefaultEventMonitor) updatePerformanceMetrics() {
	// 获取内存统计
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	m.performanceMetrics.MemoryUsage = int64(memStats.Alloc)
	m.performanceMetrics.AllocatedMemory = int64(memStats.TotalAlloc)
	m.performanceMetrics.HeapSize = int64(memStats.HeapAlloc)
	m.performanceMetrics.GoroutineCount = runtime.NumGoroutine()
	m.performanceMetrics.GCPauseTime = time.Duration(memStats.PauseNs[(memStats.NumGC+255)%256])
	m.performanceMetrics.CollectedAt = time.Now()
	
	// 简单的CPU使用率模拟（实际应用中应该使用更精确的方法）
	m.performanceMetrics.CPUUsage = float64(runtime.NumGoroutine()) / 100.0
	if m.performanceMetrics.CPUUsage > 1.0 {
		m.performanceMetrics.CPUUsage = 1.0
	}
}

func (m *DefaultEventMonitor) checkEventBusHealth() HealthLevel {
	return HealthLevelHealthy
}

func (m *DefaultEventMonitor) checkMemoryHealth() HealthLevel {
	return HealthLevelHealthy
}

func (m *DefaultEventMonitor) checkPerformanceHealth() HealthLevel {
	return HealthLevelHealthy
}

func (m *DefaultEventMonitor) calculateOverallHealth() HealthLevel {
	return HealthLevelHealthy
}

func (m *DefaultEventMonitor) checkEventRateAlert(rate float64) {
	// 事件速率告警检查
}

func (m *DefaultEventMonitor) checkErrorRateAlert(rate float64) {
	// 错误率告警检查
}

func (m *DefaultEventMonitor) checkLatencyAlert(latency time.Duration) {
	// 延迟告警检查
}

func (m *DefaultEventMonitor) checkQueueUtilizationAlert(utilization float64) {
	// 队列利用率告警检查
}

func (m *DefaultEventMonitor) checkWorkerUtilizationAlert(utilization float64) {
	// 工作池利用率告警检查
}

func (m *DefaultEventMonitor) checkThroughputAlert(throughput float64) {
	// 吞吐量告警检查
}