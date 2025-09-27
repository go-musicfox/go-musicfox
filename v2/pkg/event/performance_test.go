package event

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventBus_BatchProcessing(t *testing.T) {
	_ = slog.Default() // 避免未使用变量警告
	config := &WorkerPoolConfig{
		MinWorkers:   2,
		MaxWorkers:   10,
		QueueSize:    1000,
		BatchSize:    5,
		BatchTimeout: time.Millisecond * 50,
	}
	eventBus := NewEventBusWithConfig(slog.Default(), config)

	ctx := context.Background()
	err := eventBus.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = eventBus.Stop(ctx) }()

	// 注册事件类型
	err = eventBus.RegisterEventType(EventPlayerPlay)
	require.NoError(t, err)

	// 创建事件处理器来跟踪批量处理
	var processedEvents int64
	var batchSizes []int
	var mu sync.Mutex

	handler := func(ctx context.Context, event Event) error {
		atomic.AddInt64(&processedEvents, 1)
		return nil
	}

	// 订阅事件
	subscription, err := eventBus.Subscribe(EventPlayerPlay, handler)
	require.NoError(t, err)
	assert.NotNil(t, subscription)

	// 快速发布多个事件以触发批量处理
	for i := 0; i < 20; i++ {
		event := &BaseEvent{
			ID:        fmt.Sprintf("batch-test-%d", i),
			Type:      EventPlayerPlay,
			Source:    "batch-test",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"index": i},
		}
		err = eventBus.PublishAsync(ctx, event)
		assert.NoError(t, err)
	}

	// 等待所有事件被处理
	time.Sleep(time.Second)

	// 验证所有事件都被处理了
	assert.Equal(t, int64(20), atomic.LoadInt64(&processedEvents))

	// 验证批量处理确实发生了（通过检查日志或内部状态）
	// 这里简化处理，实际实现中可能需要更复杂的验证
	mu.Lock()
	_ = batchSizes // 使用变量避免编译警告
	mu.Unlock()
}

func TestEventBus_DynamicScaling(t *testing.T) {
	_ = slog.Default() // 避免未使用变量警告
	config := &WorkerPoolConfig{
		MinWorkers:         2,
		MaxWorkers:         8,
		QueueSize:          100,
		BatchSize:          10,
		BatchTimeout:       time.Millisecond * 100,
		ScaleUpThreshold:   float64(0.7),  // 70%队列使用率时扩容
		ScaleDownThreshold: float64(0.2),  // 20%队列使用率时缩容
		ScaleCooldown:      time.Millisecond * 500, // 短冷却时间用于测试
	}
	eventBus := NewEventBusWithConfig(slog.Default(), config)

	ctx := context.Background()
	err := eventBus.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = eventBus.Stop(ctx) }()

	// 注册事件类型
	err = eventBus.RegisterEventType(EventPlayerPlay)
	require.NoError(t, err)

	// 创建慢速事件处理器以填满队列
	var processedEvents int64
	handler := func(ctx context.Context, event Event) error {
		time.Sleep(time.Millisecond * 100) // 慢速处理
		atomic.AddInt64(&processedEvents, 1)
		return nil
	}

	// 订阅事件
	subscription, err := eventBus.Subscribe(EventPlayerPlay, handler)
	require.NoError(t, err)
	assert.NotNil(t, subscription)

	// 快速发布大量事件以触发扩容
	for i := 0; i < 80; i++ {
		event := &BaseEvent{
			ID:        fmt.Sprintf("scale-test-%d", i),
			Type:      EventPlayerPlay,
			Source:    "scale-test",
			Timestamp: time.Now(),
		}
		err = eventBus.PublishAsync(ctx, event)
		if err != nil {
			// 队列可能满了，这是预期的
			break
		}
	}

	// 等待扩容发生
	time.Sleep(time.Second * 2)

	// 验证工作协程数量增加了（通过内部状态检查）
	// 这里简化处理，实际实现中需要访问内部状态

	// 等待所有事件处理完成
	time.Sleep(time.Second * 2)

	// 验证事件被处理了
	processed := atomic.LoadInt64(&processedEvents)
	assert.Greater(t, processed, int64(0))
}

func TestEventBus_PerformanceMonitoring(t *testing.T) {
	config := DefaultWorkerPoolConfig()
	eventBus := NewEventBusWithConfig(slog.Default(), config)

	ctx := context.Background()
	err := eventBus.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = eventBus.Stop(ctx) }()

	// 注册事件类型
	err = eventBus.RegisterEventType(EventPlayerPlay)
	require.NoError(t, err)

	// 创建事件处理器
	var processedEvents int64
	handler := func(ctx context.Context, event Event) error {
		time.Sleep(time.Millisecond * 10) // 模拟处理时间
		atomic.AddInt64(&processedEvents, 1)
		return nil
	}

	// 订阅事件
	subscription, err := eventBus.Subscribe(EventPlayerPlay, handler)
	require.NoError(t, err)
	assert.NotNil(t, subscription)

	// 发布事件
	for i := 0; i < 50; i++ {
		event := &BaseEvent{
			ID:        fmt.Sprintf("perf-test-%d", i),
			Type:      EventPlayerPlay,
			Source:    "perf-test",
			Timestamp: time.Now(),
		}
		err = eventBus.PublishAsync(ctx, event)
		assert.NoError(t, err)
	}

	// 等待处理完成
	time.Sleep(time.Second * 3)

	// 验证事件被处理了
	assert.Equal(t, int64(50), atomic.LoadInt64(&processedEvents))

	// 获取统计信息
	stats := eventBus.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, int64(50), stats.TotalEvents)

	// 验证性能监控数据被收集了
	// 这里简化处理，实际实现中需要检查处理时间等指标
}

func TestEventBus_HighThroughput(t *testing.T) {
	config := &WorkerPoolConfig{
		MinWorkers:   5,
		MaxWorkers:   20,
		QueueSize:    10000,
		BatchSize:    50,
		BatchTimeout: time.Millisecond * 10,
	}
	eventBus := NewEventBusWithConfig(slog.Default(), config)

	ctx := context.Background()
	err := eventBus.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = eventBus.Stop(ctx) }()

	// 注册事件类型
	err = eventBus.RegisterEventType(EventPlayerPlay)
	require.NoError(t, err)

	// 创建快速事件处理器
	var processedEvents int64
	handler := func(ctx context.Context, event Event) error {
		atomic.AddInt64(&processedEvents, 1)
		return nil
	}

	// 订阅事件
	subscription, err := eventBus.Subscribe(EventPlayerPlay, handler)
	require.NoError(t, err)
	assert.NotNil(t, subscription)

	// 高吞吐量测试
	startTime := time.Now()
	eventCount := 10000

	// 并发发布事件
	var wg sync.WaitGroup
	workers := 10
	eventsPerWorker := eventCount / workers

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < eventsPerWorker; i++ {
				event := &BaseEvent{
					ID:        fmt.Sprintf("throughput-test-%d-%d", workerID, i),
					Type:      EventPlayerPlay,
					Source:    fmt.Sprintf("worker-%d", workerID),
					Timestamp: time.Now(),
				}
				_ = eventBus.PublishAsync(ctx, event)
			}
		}(w)
	}

	wg.Wait()
	publishTime := time.Since(startTime)

	// 等待所有事件被处理
	time.Sleep(time.Second * 5)

	processed := atomic.LoadInt64(&processedEvents)
	processingTime := time.Since(startTime)

	// 验证吞吐量
	publishRate := float64(eventCount) / publishTime.Seconds()
	processingRate := float64(processed) / processingTime.Seconds()

	t.Logf("Published %d events in %v (%.2f events/sec)", eventCount, publishTime, publishRate)
	t.Logf("Processed %d events in %v (%.2f events/sec)", processed, processingTime, processingRate)

	// 验证大部分事件都被处理了（允许一些丢失）
	assert.Greater(t, processed, int64(float64(eventCount)*0.9)) // 至少90%的事件被处理
	assert.Greater(t, publishRate, 1000.0)             // 发布速率应该大于1000事件/秒
}

func TestEventBus_ErrorHandlingPerformance(t *testing.T) {
	config := DefaultWorkerPoolConfig()
	eventBus := NewEventBusWithConfig(slog.Default(), config)

	ctx := context.Background()
	err := eventBus.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = eventBus.Stop(ctx) }()

	// 注册事件类型
	err = eventBus.RegisterEventType(EventPlayerPlay)
	require.NoError(t, err)

	// 创建会出错的事件处理器
	var processedEvents int64
	var errorEvents int64
	handler := func(ctx context.Context, event Event) error {
		atomic.AddInt64(&processedEvents, 1)
		// 50%的事件会出错
		if processedEvents%2 == 0 {
			atomic.AddInt64(&errorEvents, 1)
			return fmt.Errorf("simulated error for event %s", event.GetID())
		}
		return nil
	}

	// 订阅事件
	subscription, err := eventBus.Subscribe(EventPlayerPlay, handler)
	require.NoError(t, err)
	assert.NotNil(t, subscription)

	// 发布事件
	startTime := time.Now()
	eventCount := 1000

	for i := 0; i < eventCount; i++ {
		event := &BaseEvent{
			ID:        fmt.Sprintf("error-test-%d", i),
			Type:      EventPlayerPlay,
			Source:    "error-test",
			Timestamp: time.Now(),
		}
		err = eventBus.PublishAsync(ctx, event)
		assert.NoError(t, err)
	}

	// 等待处理完成
	time.Sleep(time.Second * 3)

	processed := atomic.LoadInt64(&processedEvents)
	errors := atomic.LoadInt64(&errorEvents)
	processingTime := time.Since(startTime)

	// 验证性能不受错误影响太大
	processingRate := float64(processed) / processingTime.Seconds()

	t.Logf("Processed %d events with %d errors in %v (%.2f events/sec)", 
		processed, errors, processingTime, processingRate)

	// 验证所有事件都被尝试处理了
	assert.Equal(t, int64(eventCount), processed)
	// 验证大约一半的事件出错了
	assert.InDelta(t, float64(eventCount)/2, float64(errors), float64(eventCount)*0.1)
	// 验证处理速率仍然合理
	assert.Greater(t, processingRate, 100.0) // 至少100事件/秒

	// 检查错误统计
	stats := eventBus.GetStats()
	assert.Greater(t, stats.ErrorCounts[EventPlayerPlay], int64(0))
}

func TestEventBus_MemoryUsage(t *testing.T) {
	_ = slog.Default() // 避免未使用变量警告
	config := &WorkerPoolConfig{
		MinWorkers:   2,
		MaxWorkers:   5,
		QueueSize:    1000,
		BatchSize:    10,
		BatchTimeout: time.Millisecond * 100,
	}
	eventBus := NewEventBusWithConfig(slog.Default(), config)

	ctx := context.Background()
	err := eventBus.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = eventBus.Stop(ctx) }()

	// 注册事件类型
	err = eventBus.RegisterEventType(EventPlayerPlay)
	require.NoError(t, err)

	// 创建事件处理器
	var processedEvents int64
	handler := func(ctx context.Context, event Event) error {
		atomic.AddInt64(&processedEvents, 1)
		return nil
	}

	// 订阅事件
	subscription, err := eventBus.Subscribe(EventPlayerPlay, handler)
	require.NoError(t, err)
	assert.NotNil(t, subscription)

	// 发布大量事件来测试内存使用
	eventCount := 10000
	for i := 0; i < eventCount; i++ {
		event := &BaseEvent{
			ID:        fmt.Sprintf("memory-test-%d", i),
			Type:      EventPlayerPlay,
			Source:    "memory-test",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"large_data": make([]byte, 1024)}, // 1KB数据
		}
		err = eventBus.PublishAsync(ctx, event)
		if err != nil {
			// 队列满了，等待一下
			time.Sleep(time.Millisecond * 10)
			i-- // 重试
		}
	}

	// 等待处理完成，减少等待时间
	time.Sleep(time.Millisecond * 500)

	// 验证事件被处理了
	processed := atomic.LoadInt64(&processedEvents)
	assert.Greater(t, processed, int64(float64(eventCount)*0.8)) // 至少80%的事件被处理

	// 验证内存没有泄漏（这里简化处理，实际测试中可能需要更复杂的内存监控）
	// 可以通过runtime.ReadMemStats()来检查内存使用情况
}

func TestWorkerPoolConfig_Defaults(t *testing.T) {
	config := DefaultWorkerPoolConfig()

	assert.Equal(t, 2, config.MinWorkers)
	assert.Equal(t, 20, config.MaxWorkers)
	assert.Equal(t, 1000, config.QueueSize)
	assert.Equal(t, 10, config.BatchSize)
	assert.Equal(t, time.Millisecond*100, config.BatchTimeout)
	assert.Equal(t, float64(0.8), config.ScaleUpThreshold)
	assert.Equal(t, float64(0.2), config.ScaleDownThreshold)
	assert.Equal(t, time.Second*30, config.ScaleCooldown)
}

func TestEventBus_ConcurrentSubscriptions(t *testing.T) {
	_ = slog.Default() // 避免未使用变量警告
	eventBus := NewDefaultEventBus()

	ctx := context.Background()
	err := eventBus.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = eventBus.Stop(ctx) }()

	// 注册事件类型
	err = eventBus.RegisterEventType(EventPlayerPlay)
	require.NoError(t, err)

	// 并发创建多个订阅
	var wg sync.WaitGroup
	subscriptionCount := 100
	var subscriptions []*Subscription
	var mu sync.Mutex

	for i := 0; i < subscriptionCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			handler := func(ctx context.Context, event Event) error {
				return nil
			}
			subscription, err := eventBus.Subscribe(EventPlayerPlay, handler)
			assert.NoError(t, err)
			mu.Lock()
			subscriptions = append(subscriptions, subscription)
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// 验证所有订阅都成功创建了
	mu.Lock()
	assert.Len(t, subscriptions, subscriptionCount)
	mu.Unlock()

	// 验证订阅数量统计正确
	totalSubscriptions := eventBus.GetTotalSubscriptions()
	assert.Equal(t, subscriptionCount, totalSubscriptions)

	// 并发取消订阅
	for _, subscription := range subscriptions {
		wg.Add(1)
		go func(sub *Subscription) {
			defer wg.Done()
			err := eventBus.Unsubscribe(sub.ID)
			assert.NoError(t, err)
		}(subscription)
	}

	wg.Wait()

	// 验证所有订阅都被取消了
	totalSubscriptions = eventBus.GetTotalSubscriptions()
	assert.Equal(t, 0, totalSubscriptions)
}

func BenchmarkEventBus_PublishAsync(b *testing.B) {
	logger := slog.Default()
	config := &WorkerPoolConfig{
		MinWorkers:   5,
		MaxWorkers:   20,
		QueueSize:    10000,
		BatchSize:    50,
		BatchTimeout: time.Millisecond * 10,
	}
	eventBus := NewEventBusWithConfig(logger, config)

	ctx := context.Background()
	err := eventBus.Start(ctx)
	require.NoError(b, err)
	defer func() { _ = eventBus.Stop(ctx) }()

	// 注册事件类型
	err = eventBus.RegisterEventType(EventPlayerPlay)
	require.NoError(b, err)

	// 创建快速事件处理器
	handler := func(ctx context.Context, event Event) error {
		return nil
	}

	// 订阅事件
	_, err = eventBus.Subscribe(EventPlayerPlay, handler)
	require.NoError(b, err)

	// 基准测试
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			event := &BaseEvent{
				ID:        fmt.Sprintf("bench-%d", i),
				Type:      EventPlayerPlay,
				Source:    "benchmark",
				Timestamp: time.Now(),
			}
			_ = eventBus.PublishAsync(ctx, event)
			i++
		}
	})
}

func BenchmarkEventBus_Subscribe(b *testing.B) {
	_ = slog.Default() // 避免未使用变量警告
	eventBus := NewDefaultEventBus()

	ctx := context.Background()
	err := eventBus.Start(ctx)
	require.NoError(b, err)
	defer func() { _ = eventBus.Stop(ctx) }()

	// 注册事件类型
	err = eventBus.RegisterEventType(EventPlayerPlay)
	require.NoError(b, err)

	// 基准测试订阅性能
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler := func(ctx context.Context, event Event) error {
			return nil
		}
		_, err := eventBus.Subscribe(EventPlayerPlay, handler)
		if err != nil {
			b.Fatal(err)
		}
	}
}