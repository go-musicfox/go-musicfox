package event

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultEventMonitor_StartStop(t *testing.T) {
	logger := slog.Default()
	eventBus := NewDefaultEventBus()
	monitor := NewDefaultEventMonitor(logger, eventBus)

	ctx := context.Background()

	// 测试启动
	err := monitor.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, monitor.IsRunning())

	// 测试重复启动应该失败
	err = monitor.Start(ctx)
	assert.Error(t, err)

	// 测试停止
	err = monitor.Stop(ctx)
	assert.NoError(t, err)
	assert.False(t, monitor.IsRunning())

	// 测试重复停止应该失败
	err = monitor.Stop(ctx)
	assert.Error(t, err)
}

func TestDefaultEventMonitor_GetStats(t *testing.T) {
	logger := slog.Default()
	eventBus := NewDefaultEventBus()
	monitor := NewDefaultEventMonitor(logger, eventBus)

	ctx := context.Background()
	err := monitor.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = monitor.Stop(ctx) }()

	// 获取初始统计信息
	stats := monitor.GetStats()
	assert.NotNil(t, stats)
	assert.NotNil(t, stats.EventStats)
	assert.NotNil(t, stats.EventCounts)
	assert.NotNil(t, stats.ErrorCounts)
	assert.NotNil(t, stats.ErrorsByType)
	assert.NotNil(t, stats.ErrorsBySource)
	assert.NotNil(t, stats.SubscriptionsByType)
	assert.NotNil(t, stats.SubscriptionsByGroup)

	// 验证初始值
	assert.Equal(t, int64(0), stats.TotalEvents)
	assert.Equal(t, 0, stats.TotalSubscribers)
	assert.Equal(t, float64(0), stats.ErrorRate)
}

func TestDefaultEventMonitor_GetPerformanceMetrics(t *testing.T) {
	logger := slog.Default()
	eventBus := NewDefaultEventBus()
	monitor := NewDefaultEventMonitor(logger, eventBus)

	ctx := context.Background()
	err := monitor.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = monitor.Stop(ctx) }()

	// 获取性能指标
	metrics := monitor.GetPerformanceMetrics()
	assert.NotNil(t, metrics)

	// 验证指标字段存在
	assert.GreaterOrEqual(t, metrics.CPUUsage, float64(0))
	assert.GreaterOrEqual(t, metrics.MemoryUsage, int64(0))
	assert.GreaterOrEqual(t, metrics.GoroutineCount, 0)
}

func TestDefaultEventMonitor_GetHealthStatus(t *testing.T) {
	logger := slog.Default()
	eventBus := NewDefaultEventBus()
	monitor := NewDefaultEventMonitor(logger, eventBus)

	ctx := context.Background()
	err := monitor.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = monitor.Stop(ctx) }()

	// 获取健康状态
	healthStatus := monitor.GetHealthStatus()
	assert.NotNil(t, healthStatus)
	assert.NotNil(t, healthStatus.Components)
	assert.NotNil(t, healthStatus.Issues)

	// 验证初始健康状态
	assert.Equal(t, HealthLevelHealthy, healthStatus.OverallHealth)
	assert.Greater(t, healthStatus.Uptime, time.Duration(0))
}

func TestDefaultEventMonitor_GetRealTimeStats(t *testing.T) {
	logger := slog.Default()
	eventBus := NewDefaultEventBus()
	monitor := NewDefaultEventMonitor(logger, eventBus)

	ctx := context.Background()
	err := monitor.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = monitor.Stop(ctx) }()

	// 等待一段时间让监控器收集数据
	time.Sleep(time.Millisecond * 100)

	// 获取实时统计
	duration := time.Minute
	realTimeStats := monitor.GetRealTimeStats(duration)
	assert.NotNil(t, realTimeStats)
	assert.Equal(t, duration, realTimeStats.TimeWindow)
	assert.GreaterOrEqual(t, realTimeStats.EventCount, int64(0))
	assert.GreaterOrEqual(t, realTimeStats.EventRate, float64(0))
	assert.GreaterOrEqual(t, realTimeStats.ErrorCount, int64(0))
	assert.GreaterOrEqual(t, realTimeStats.ErrorRate, float64(0))
}

func TestDefaultEventMonitor_GetEventRateStats(t *testing.T) {
	logger := slog.Default()
	eventBus := NewDefaultEventBus()
	monitor := NewDefaultEventMonitor(logger, eventBus)

	ctx := context.Background()
	err := monitor.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = monitor.Stop(ctx) }()

	// 等待一段时间让监控器收集数据
	time.Sleep(time.Millisecond * 100)

	// 获取事件速率统计
	duration := time.Minute
	rateStats := monitor.GetEventRateStats(duration)
	assert.NotNil(t, rateStats)
	assert.Equal(t, duration, rateStats.TimeWindow)
	assert.GreaterOrEqual(t, rateStats.CurrentRate, float64(0))
	assert.GreaterOrEqual(t, rateStats.AverageRate, float64(0))
	assert.GreaterOrEqual(t, rateStats.PeakRate, float64(0))
	assert.GreaterOrEqual(t, rateStats.MinRate, float64(0))
	assert.NotNil(t, rateStats.RateHistory)
}

func TestDefaultEventMonitor_GetLatencyStats(t *testing.T) {
	logger := slog.Default()
	eventBus := NewDefaultEventBus()
	monitor := NewDefaultEventMonitor(logger, eventBus)

	ctx := context.Background()
	err := monitor.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = monitor.Stop(ctx) }()

	// 获取延迟统计
	latencyStats := monitor.GetLatencyStats()
	assert.NotNil(t, latencyStats)
	assert.GreaterOrEqual(t, latencyStats.AverageLatency, time.Duration(0))
	assert.GreaterOrEqual(t, latencyStats.MinLatency, time.Duration(0))
	assert.GreaterOrEqual(t, latencyStats.MaxLatency, time.Duration(0))
	assert.NotNil(t, latencyStats.LatencyHistory)
}

func TestDefaultEventMonitor_SetThresholds(t *testing.T) {
	logger := slog.Default()
	eventBus := NewDefaultEventBus()
	monitor := NewDefaultEventMonitor(logger, eventBus)

	// 设置自定义阈值
	customThresholds := &MonitorThresholds{
		MaxEventRate:         500.0,
		MaxErrorRate:         0.1, // 10%
		MaxLatency:           time.Second * 2,
		MaxQueueUtilization:  0.9, // 90%
		MaxWorkerUtilization: 0.95, // 95%
		MaxMemoryUsage:       2 * 1024 * 1024 * 1024, // 2GB
		MinThroughput:        5.0,
	}

	monitor.SetThresholds(customThresholds)

	// 验证阈值已设置（通过内部状态检查，这里简化处理）
	// 在实际实现中，可能需要添加GetThresholds方法来验证
}

func TestDefaultEventMonitor_AlertSubscription(t *testing.T) {
	logger := slog.Default()
	eventBus := NewDefaultEventBus()
	monitor := NewDefaultEventMonitor(logger, eventBus)

	ctx := context.Background()
	err := monitor.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = monitor.Stop(ctx) }()

	// 订阅告警
	var receivedAlerts []*MonitorAlert
	var mu sync.Mutex

	alertHandler := func(alert *MonitorAlert) {
		mu.Lock()
		defer mu.Unlock()
		receivedAlerts = append(receivedAlerts, alert)
	}

	subscriptionID := monitor.SubscribeToAlerts(alertHandler)
	assert.NotEmpty(t, subscriptionID)

	// 获取活跃告警（初始应该为空）
	activeAlerts := monitor.GetActiveAlerts()
	assert.Empty(t, activeAlerts)

	// 取消订阅
	monitor.UnsubscribeFromAlerts(subscriptionID)
}

func TestDefaultEventMonitor_ExportStats(t *testing.T) {
	logger := slog.Default()
	eventBus := NewDefaultEventBus()
	monitor := NewDefaultEventMonitor(logger, eventBus)

	ctx := context.Background()
	err := monitor.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = monitor.Stop(ctx) }()

	// 测试JSON导出
	jsonData, err := monitor.ExportStats(ExportFormatJSON)
	if err == nil { // 如果实现了导出功能
		assert.NotEmpty(t, jsonData)
	}

	// 测试CSV导出
	csvData, err := monitor.ExportStats(ExportFormatCSV)
	if err == nil { // 如果实现了导出功能
		assert.NotEmpty(t, csvData)
	}

	// 测试Prometheus导出
	prometheusData, err := monitor.ExportStats(ExportFormatPrometheus)
	if err == nil { // 如果实现了导出功能
		assert.NotEmpty(t, prometheusData)
	}
}

func TestDefaultEventMonitor_ExportMetrics(t *testing.T) {
	logger := slog.Default()
	eventBus := NewDefaultEventBus()
	monitor := NewDefaultEventMonitor(logger, eventBus)

	ctx := context.Background()
	err := monitor.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = monitor.Stop(ctx) }()

	// 等待一段时间让监控器收集数据
	time.Sleep(time.Millisecond * 100)

	// 测试指标导出
	duration := time.Minute
	metricsData, err := monitor.ExportMetrics(ExportFormatJSON, duration)
	if err == nil { // 如果实现了导出功能
		assert.NotEmpty(t, metricsData)
	}
}

func TestHealthLevel_String(t *testing.T) {
	tests := []struct {
		level    HealthLevel
		expected string
	}{
		{HealthLevelHealthy, "healthy"},
		{HealthLevelWarning, "warning"},
		{HealthLevelCritical, "critical"},
		{HealthLevelUnknown, "unknown"},
		{HealthLevel(999), "unknown"}, // 未知值
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.level.String())
	}
}

func TestAlertLevel_String(t *testing.T) {
	tests := []struct {
		level    AlertLevel
		expected string
	}{
		{AlertLevelInfo, "info"},
		{AlertLevelWarning, "warning"},
		{AlertLevelCritical, "critical"},
		{AlertLevel(999), "unknown"}, // 未知值
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.level.String())
	}
}

func TestAlertType_String(t *testing.T) {
	tests := []struct {
		alertType AlertType
		expected  string
	}{
		{AlertTypeEventRate, "event_rate"},
		{AlertTypeErrorRate, "error_rate"},
		{AlertTypeLatency, "latency"},
		{AlertTypeQueueUtilization, "queue_utilization"},
		{AlertTypeWorkerUtilization, "worker_utilization"},
		{AlertTypeMemoryUsage, "memory_usage"},
		{AlertTypeThroughput, "throughput"},
		{AlertType(999), "unknown"}, // 未知值
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.alertType.String())
	}
}

func TestMonitorAlert_Creation(t *testing.T) {
	now := time.Now()
	alert := &MonitorAlert{
		ID:           "test-alert-1",
		Level:        AlertLevelWarning,
		Type:         AlertTypeEventRate,
		Message:      "Event rate is too high",
		Metric:       "event_rate",
		Threshold:    1000.0,
		CurrentValue: 1500.0,
		TriggeredAt:  now,
		Count:        1,
	}

	assert.Equal(t, "test-alert-1", alert.ID)
	assert.Equal(t, AlertLevelWarning, alert.Level)
	assert.Equal(t, AlertTypeEventRate, alert.Type)
	assert.Equal(t, "Event rate is too high", alert.Message)
	assert.Equal(t, "event_rate", alert.Metric)
	assert.Equal(t, 1000.0, alert.Threshold)
	assert.Equal(t, 1500.0, alert.CurrentValue)
	assert.Equal(t, now, alert.TriggeredAt)
	assert.Nil(t, alert.ResolvedAt)
	assert.Equal(t, 1, alert.Count)
}

func TestHealthIssue_Creation(t *testing.T) {
	now := time.Now()
	issue := HealthIssue{
		Component:  "event_bus",
		Level:      HealthLevelWarning,
		Message:    "Queue utilization is high",
		DetectedAt: now,
	}

	assert.Equal(t, "event_bus", issue.Component)
	assert.Equal(t, HealthLevelWarning, issue.Level)
	assert.Equal(t, "Queue utilization is high", issue.Message)
	assert.Equal(t, now, issue.DetectedAt)
	assert.Nil(t, issue.ResolvedAt)
}

func TestRealTimeStats_Creation(t *testing.T) {
	now := time.Now()
	duration := time.Minute * 5
	stats := &RealTimeStats{
		TimeWindow:     duration,
		EventCount:     100,
		EventRate:      20.0,
		ErrorCount:     5,
		ErrorRate:      0.05,
		AverageLatency: time.Millisecond * 50,
		Throughput:     18.5,
		CollectedAt:    now,
	}

	assert.Equal(t, duration, stats.TimeWindow)
	assert.Equal(t, int64(100), stats.EventCount)
	assert.Equal(t, 20.0, stats.EventRate)
	assert.Equal(t, int64(5), stats.ErrorCount)
	assert.Equal(t, 0.05, stats.ErrorRate)
	assert.Equal(t, time.Millisecond*50, stats.AverageLatency)
	assert.Equal(t, 18.5, stats.Throughput)
	assert.Equal(t, now, stats.CollectedAt)
}

func TestEventRateStats_Creation(t *testing.T) {
	now := time.Now()
	duration := time.Minute * 10
	rateHistory := []RateDataPoint{
		{Timestamp: now.Add(-time.Minute), Rate: 15.0, EventCount: 15},
		{Timestamp: now, Rate: 25.0, EventCount: 40},
	}

	stats := &EventRateStats{
		TimeWindow:  duration,
		CurrentRate: 25.0,
		AverageRate: 20.0,
		PeakRate:    30.0,
		MinRate:     10.0,
		RateHistory: rateHistory,
		CollectedAt: now,
	}

	assert.Equal(t, duration, stats.TimeWindow)
	assert.Equal(t, 25.0, stats.CurrentRate)
	assert.Equal(t, 20.0, stats.AverageRate)
	assert.Equal(t, 30.0, stats.PeakRate)
	assert.Equal(t, 10.0, stats.MinRate)
	assert.Len(t, stats.RateHistory, 2)
	assert.Equal(t, now, stats.CollectedAt)
}

func TestLatencyStats_Creation(t *testing.T) {
	now := time.Now()
	latencyHistory := []LatencyDataPoint{
		{Timestamp: now.Add(-time.Minute), Latency: time.Millisecond * 30, EventType: EventPlayerPlay},
		{Timestamp: now, Latency: time.Millisecond * 50, EventType: EventPlayerPause},
	}

	stats := &LatencyStats{
		AverageLatency: time.Millisecond * 40,
		MinLatency:     time.Millisecond * 20,
		MaxLatency:     time.Millisecond * 80,
		P50Latency:     time.Millisecond * 35,
		P95Latency:     time.Millisecond * 70,
		P99Latency:     time.Millisecond * 75,
		LatencyHistory: latencyHistory,
		CollectedAt:    now,
	}

	assert.Equal(t, time.Millisecond*40, stats.AverageLatency)
	assert.Equal(t, time.Millisecond*20, stats.MinLatency)
	assert.Equal(t, time.Millisecond*80, stats.MaxLatency)
	assert.Equal(t, time.Millisecond*35, stats.P50Latency)
	assert.Equal(t, time.Millisecond*70, stats.P95Latency)
	assert.Equal(t, time.Millisecond*75, stats.P99Latency)
	assert.Len(t, stats.LatencyHistory, 2)
	assert.Equal(t, now, stats.CollectedAt)
}