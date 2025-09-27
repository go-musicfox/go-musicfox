// Package thirdparty 资源监控测试
package thirdparty

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewResourceMonitor 测试创建资源监控器
func TestNewResourceMonitor(t *testing.T) {
	limits := &ResourceLimits{
		MaxMemory:     64 * 1024 * 1024, // 64MB
		MaxCPU:        0.5,              // 50%
		MaxDiskIO:     10 * 1024 * 1024, // 10MB/s
		MaxNetworkIO:  5 * 1024 * 1024,  // 5MB/s
		Timeout:       30 * time.Second,
		MaxGoroutines: 10,
		MaxFileSize:   10 * 1024 * 1024, // 10MB
		MaxOpenFiles:  20,
	}

	monitor, err := NewResourceMonitor(limits)
	assert.NoError(t, err)
	assert.NotNil(t, monitor)
	assert.Equal(t, limits, monitor.limits)
	assert.False(t, monitor.IsRunning())

	// 测试nil限制
	_, err = NewResourceMonitor(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

// TestResourceMonitorStartStop 测试资源监控器启动和停止
func TestResourceMonitorStartStop(t *testing.T) {
	limits := &ResourceLimits{
		MaxMemory:     64 * 1024 * 1024,
		MaxCPU:        0.8,
		MaxGoroutines: 20,
	}

	monitor, err := NewResourceMonitor(limits)
	require.NoError(t, err)

	// 测试启动
	ctx := context.Background()
	err = monitor.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, monitor.IsRunning())

	// 测试重复启动
	err = monitor.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// 等待一小段时间让监控器运行
	time.Sleep(100 * time.Millisecond)

	// 测试停止
	monitor.Stop()
	assert.False(t, monitor.IsRunning())

	// 测试重复停止（应该不会出错）
	monitor.Stop()
	assert.False(t, monitor.IsRunning())
}

// TestResourceUsage 测试资源使用情况获取
func TestResourceUsage(t *testing.T) {
	limits := &ResourceLimits{
		MaxMemory:     128 * 1024 * 1024,
		MaxCPU:        0.8,
		MaxGoroutines: 50,
	}

	monitor, err := NewResourceMonitor(limits)
	require.NoError(t, err)

	// 获取初始使用情况
	usage := monitor.GetUsage()
	assert.NotNil(t, usage)
	assert.NotZero(t, usage.LastUpdated)
	assert.GreaterOrEqual(t, usage.MemoryUsage, int64(0))
	assert.GreaterOrEqual(t, usage.GoroutineCount, 1) // 至少有当前测试的goroutine

	// 启动监控器
	ctx := context.Background()
	err = monitor.Start(ctx)
	require.NoError(t, err)

	// 等待监控器更新数据
	time.Sleep(1100 * time.Millisecond) // 稍微超过更新间隔

	// 再次获取使用情况
	newUsage := monitor.GetUsage()
	assert.NotNil(t, newUsage)
	assert.True(t, newUsage.LastUpdated.After(usage.LastUpdated))

	monitor.Stop()
}

// TestUpdateLimits 测试更新资源限制
func TestUpdateLimits(t *testing.T) {
	initialLimits := &ResourceLimits{
		MaxMemory:     64 * 1024 * 1024,
		MaxCPU:        0.5,
		MaxGoroutines: 10,
	}

	monitor, err := NewResourceMonitor(initialLimits)
	require.NoError(t, err)

	// 测试更新限制
	newLimits := &ResourceLimits{
		MaxMemory:     128 * 1024 * 1024,
		MaxCPU:        0.8,
		MaxGoroutines: 20,
	}

	err = monitor.UpdateLimits(newLimits)
	assert.NoError(t, err)
	assert.Equal(t, newLimits, monitor.limits)

	// 测试更新为nil
	err = monitor.UpdateLimits(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

// TestResourceViolations 测试资源违规检测
func TestResourceViolations(t *testing.T) {
	// 设置很低的限制以便触发违规
	limits := &ResourceLimits{
		MaxMemory:     1024, // 1KB - 很低的内存限制
		MaxCPU:        0.01, // 1% - 很低的CPU限制
		MaxGoroutines: 1,    // 很低的goroutine限制
		MaxOpenFiles:  1,    // 很低的文件句柄限制
	}

	monitor, err := NewResourceMonitor(limits)
	require.NoError(t, err)

	// 设置较短的更新间隔以便快速检测
	monitor.SetUpdateInterval(100 * time.Millisecond)

	// 启动监控器
	ctx := context.Background()
	err = monitor.Start(ctx)
	require.NoError(t, err)

	// 等待监控器检测到违规
	time.Sleep(500 * time.Millisecond)

	// 检查违规记录
	violations := monitor.GetViolations()
	assert.NotEmpty(t, violations, "应该检测到资源违规")

	// 验证违规记录的内容
	for _, violation := range violations {
		assert.NotZero(t, violation.Timestamp)
		assert.NotEmpty(t, violation.Description)
		assert.NotNil(t, violation.CurrentValue)
		assert.NotNil(t, violation.LimitValue)
	}

	// 测试清除违规记录
	monitor.ClearViolations()
	violations = monitor.GetViolations()
	assert.Empty(t, violations)

	monitor.Stop()
}

// TestResourceCallbacks 测试资源回调
func TestResourceCallbacks(t *testing.T) {
	limits := &ResourceLimits{
		MaxMemory:     1024, // 很低的限制以触发违规
		MaxGoroutines: 1,
	}

	monitor, err := NewResourceMonitor(limits)
	require.NoError(t, err)

	// 使用channel来同步回调
	callbackCh := make(chan ResourceViolation, 1)

	monitor.AddCallback(func(violation ResourceViolation) {
		select {
		case callbackCh <- violation:
		default:
			// 如果channel已满，忽略
		}
	})

	// 设置较短的更新间隔
	monitor.SetUpdateInterval(100 * time.Millisecond)

	// 启动监控器
	ctx := context.Background()
	err = monitor.Start(ctx)
	require.NoError(t, err)

	// 等待回调被调用
	select {
	case receivedViolation := <-callbackCh:
		// 验证回调被调用
		assert.NotZero(t, receivedViolation.Timestamp)
	case <-time.After(1 * time.Second):
		t.Error("回调函数应该被调用")
	}

	monitor.Stop()
}

// TestResourceMonitorConfiguration 测试资源监控器配置
func TestResourceMonitorConfiguration(t *testing.T) {
	limits := &ResourceLimits{
		MaxMemory:     64 * 1024 * 1024,
		MaxGoroutines: 20,
	}

	monitor, err := NewResourceMonitor(limits)
	require.NoError(t, err)

	// 测试设置更新间隔
	newInterval := 2 * time.Second
	monitor.SetUpdateInterval(newInterval)
	assert.Equal(t, newInterval, monitor.updateInterval)

	// 测试设置告警阈值
	newThreshold := 0.9
	monitor.SetAlertThreshold(newThreshold)
	assert.Equal(t, newThreshold, monitor.alertThreshold)

	// 测试无效阈值
	monitor.SetAlertThreshold(-0.1) // 无效值
	assert.Equal(t, newThreshold, monitor.alertThreshold) // 应该保持不变

	monitor.SetAlertThreshold(1.5) // 无效值
	assert.Equal(t, newThreshold, monitor.alertThreshold) // 应该保持不变

	// 测试有效阈值
	monitor.SetAlertThreshold(0.7)
	assert.Equal(t, 0.7, monitor.alertThreshold)
}

// TestResourceMonitorStats 测试资源监控器统计信息
func TestResourceMonitorStats(t *testing.T) {
	limits := &ResourceLimits{
		MaxMemory:     64 * 1024 * 1024,
		MaxCPU:        0.8,
		MaxGoroutines: 20,
		MaxOpenFiles:  50,
		Timeout:       30 * time.Second,
	}

	monitor, err := NewResourceMonitor(limits)
	require.NoError(t, err)

	// 添加一个回调
	monitor.AddCallback(func(violation ResourceViolation) {})

	// 获取统计信息
	stats := monitor.GetStats()
	assert.NotNil(t, stats)

	// 验证基本统计信息
	assert.Equal(t, false, stats["running"])
	assert.Equal(t, "1s", stats["update_interval"])
	assert.Equal(t, 0.8, stats["alert_threshold"])
	assert.Equal(t, 0, stats["violation_count"])
	assert.Equal(t, 1, stats["callback_count"])

	// 验证当前使用情况
	currentUsage, ok := stats["current_usage"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, currentUsage, "memory_usage")
	assert.Contains(t, currentUsage, "cpu_usage")
	assert.Contains(t, currentUsage, "goroutine_count")
	assert.Contains(t, currentUsage, "open_file_count")
	assert.Contains(t, currentUsage, "uptime")

	// 验证资源限制
	limitsStats, ok := stats["limits"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, limits.MaxMemory, limitsStats["max_memory"])
	assert.Equal(t, limits.MaxCPU, limitsStats["max_cpu"])
	assert.Equal(t, limits.MaxGoroutines, limitsStats["max_goroutines"])
	assert.Equal(t, limits.MaxOpenFiles, limitsStats["max_open_files"])
	assert.Equal(t, limits.Timeout.String(), limitsStats["timeout"])

	// 验证违规统计
	violationStats, ok := stats["violation_stats"].(map[string]int)
	assert.True(t, ok)
	assert.NotNil(t, violationStats)
}

// TestResourceViolationTypes 测试资源违规类型
func TestResourceViolationTypes(t *testing.T) {
	testCases := []struct {
		violationType ResourceViolationType
		expected      string
	}{
		{ResourceViolationMemory, "memory"},
		{ResourceViolationCPU, "cpu"},
		{ResourceViolationDiskIO, "disk_io"},
		{ResourceViolationNetworkIO, "network_io"},
		{ResourceViolationGoroutines, "goroutines"},
		{ResourceViolationFileHandles, "file_handles"},
		{ResourceViolationTimeout, "timeout"},
		{ResourceViolationType(999), "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.violationType.String())
		})
	}
}

// TestViolationActions 测试违规行动
func TestViolationActions(t *testing.T) {
	testCases := []struct {
		action   ViolationAction
		expected string
	}{
		{ViolationActionNone, "none"},
		{ViolationActionWarn, "warn"},
		{ViolationActionThrottle, "throttle"},
		{ViolationActionTerminate, "terminate"},
		{ViolationAction(999), "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.action.String())
		})
	}
}

// TestResourceMonitorConcurrency 测试资源监控器并发安全
func TestResourceMonitorConcurrency(t *testing.T) {
	limits := &ResourceLimits{
		MaxMemory:     64 * 1024 * 1024,
		MaxGoroutines: 50,
	}

	monitor, err := NewResourceMonitor(limits)
	require.NoError(t, err)

	// 启动监控器
	ctx := context.Background()
	err = monitor.Start(ctx)
	require.NoError(t, err)

	// 并发访问监控器
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// 并发获取使用情况
			usage := monitor.GetUsage()
			assert.NotNil(t, usage)

			// 并发获取违规记录
			violations := monitor.GetViolations()
			assert.NotNil(t, violations)

			// 并发获取统计信息
			stats := monitor.GetStats()
			assert.NotNil(t, stats)
		}()
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}

	monitor.Stop()
}

// TestResourceMonitorMemoryLeak 测试资源监控器内存泄漏防护
func TestResourceMonitorMemoryLeak(t *testing.T) {
	limits := &ResourceLimits{
		MaxMemory:     1024, // 很低的限制以触发大量违规
		MaxGoroutines: 1,
	}

	monitor, err := NewResourceMonitor(limits)
	require.NoError(t, err)

	// 设置很短的更新间隔以产生大量违规记录
	monitor.SetUpdateInterval(10 * time.Millisecond)

	// 启动监控器
	ctx := context.Background()
	err = monitor.Start(ctx)
	require.NoError(t, err)

	// 等待产生大量违规记录
	time.Sleep(2 * time.Second)

	// 检查违规记录数量是否被限制
	violations := monitor.GetViolations()
	assert.LessOrEqual(t, len(violations), 1000, "违规记录数量应该被限制在1000以内")

	monitor.Stop()
}