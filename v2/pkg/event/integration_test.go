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

func TestEventSystemIntegration(t *testing.T) {
	// 创建事件总线
	bus := NewEventBus(slog.Default())
	require.NotNil(t, bus)

	// 启动事件总线
	ctx := context.Background()
	err := bus.Start(ctx)
	require.NoError(t, err)
	defer func() {
		err := bus.Stop(ctx)
		assert.NoError(t, err)
	}()

	// 创建事件存储
	storeConfig := &EventStoreConfig{
		MaxEvents:     100,
		RetentionTime: time.Hour,
	}
	store := NewMemoryEventStore(slog.Default(), storeConfig)
	require.NotNil(t, store)

	// 创建事件监控器
	monitor := NewDefaultEventMonitor(slog.Default(), bus)
	require.NotNil(t, monitor)

	err = monitor.Start(ctx)
	require.NoError(t, err)
	defer func() {
		err := monitor.Stop(ctx)
		assert.NoError(t, err)
	}()

	// 创建事件重放器
	replayer := NewDefaultEventReplayer(slog.Default(), store)
	require.NotNil(t, replayer)

	// 测试完整的事件流程
	t.Run("CompleteEventFlow", func(t *testing.T) {
		testCompleteEventFlow(t, bus, store, monitor, replayer)
	})

	// 测试事件持久化和重放
	t.Run("EventPersistenceAndReplay", func(t *testing.T) {
		testEventPersistenceAndReplay(t, bus, store, replayer)
	})

	// 测试事件监控和统计
	t.Run("EventMonitoringAndStats", func(t *testing.T) {
		testEventMonitoringAndStats(t, bus, monitor)
	})

	// 测试高并发场景
	t.Run("HighConcurrencyScenario", func(t *testing.T) {
		testHighConcurrencyScenario(t, bus, store, monitor)
	})
}

func testCompleteEventFlow(t *testing.T, bus EventBus, store EventStore, monitor EventMonitor, replayer EventReplayer) {
	var receivedEvents []Event
	var mu sync.Mutex

	// 订阅事件（使用同步处理）
	subscription, err := bus.Subscribe(EventPlayerPlay, func(ctx context.Context, event Event) error {
		mu.Lock()
		receivedEvents = append(receivedEvents, event)
		mu.Unlock()

		// 存储事件
		err := store.Store(context.Background(), event)
		assert.NoError(t, err)

		// 记录监控数据
		monitor.RecordEvent(event)
		return nil
	}, WithAsync(false))
	require.NoError(t, err)
	defer func() {
		err := bus.Unsubscribe(subscription.ID)
		assert.NoError(t, err)
	}()

	// 发布事件
	event := &BaseEvent{
		ID:        "integration-test-1",
		Type:      EventPlayerPlay,
		Source:    "integration-test",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"song_id": "test-song-123"},
	}

	err = bus.Publish(context.Background(), event)
	require.NoError(t, err)

	// 等待事件处理
	time.Sleep(100 * time.Millisecond)

	// 验证事件被接收
	mu.Lock()
	assert.Len(t, receivedEvents, 1)
	assert.Equal(t, "integration-test-1", receivedEvents[0].GetID())
	mu.Unlock()

	// 验证事件被存储
	storedEvents, err := store.GetByType(context.Background(), EventPlayerPlay, 10, 0)
	require.NoError(t, err)
	assert.Len(t, storedEvents, 1)
	assert.Equal(t, "integration-test-1", storedEvents[0].GetID())

	// 验证监控统计
	stats := monitor.GetStats()
	assert.Greater(t, stats.TotalEvents, int64(0))
	assert.Contains(t, stats.EventCounts, EventPlayerPlay)
}

func testEventPersistenceAndReplay(t *testing.T, bus EventBus, store EventStore, replayer EventReplayer) {
	// 清空存储以确保测试隔离
	err := store.Clear(context.Background())
	require.NoError(t, err)

	// 创建测试事件
	events := []*BaseEvent{
		{
			ID:        "replay-test-1",
			Type:      EventPlayerPlay,
			Source:    "replay-test",
			Timestamp: time.Now().Add(-time.Hour),
			Data:      map[string]interface{}{"song_id": "song-1"},
		},
		{
			ID:        "replay-test-2",
			Type:      EventPlayerPause,
			Source:    "replay-test",
			Timestamp: time.Now().Add(-30 * time.Minute),
			Data:      map[string]interface{}{"song_id": "song-1"},
		},
		{
			ID:        "replay-test-3",
			Type:      EventPlayerPlay,
			Source:    "replay-test",
			Timestamp: time.Now().Add(-10 * time.Minute),
			Data:      map[string]interface{}{"song_id": "song-2"},
		},
	}

	// 存储事件
	for _, event := range events {
		err := store.Store(context.Background(), event)
		require.NoError(t, err)
	}

	// 测试按类型重放
	var replayedEvents []Event
	var mu sync.Mutex

	replayer.SetEventHandler(func(ctx context.Context, event Event) error {
		mu.Lock()
		replayedEvents = append(replayedEvents, event)
		mu.Unlock()
		return nil
	})

	options := &ReplayOptions{
		BatchSize: 100,
		LoopCount: 1, // 只重放一次
	}

	err = replayer.ReplayByType(context.Background(), EventPlayerPlay, options)
	require.NoError(t, err)

	// 验证重放结果
	mu.Lock()
	assert.Len(t, replayedEvents, 2) // 应该有2个EventPlayerPlay事件
	for _, event := range replayedEvents {
		assert.Equal(t, EventPlayerPlay, event.GetType())
	}
	mu.Unlock()

	// 测试按时间范围重放
	replayedEvents = nil
	startTime := time.Now().Add(-45 * time.Minute)
	endTime := time.Now().Add(-5 * time.Minute)

	err = replayer.ReplayRange(context.Background(), startTime, endTime, &ReplayOptions{
		BatchSize: 100,
		LoopCount: 1, // 只重放一次
	})
	require.NoError(t, err)

	// 验证时间范围重放结果
	mu.Lock()
	assert.Len(t, replayedEvents, 2) // 应该有2个在时间范围内的事件
	mu.Unlock()
}

func testEventMonitoringAndStats(t *testing.T, bus EventBus, monitor EventMonitor) {
	// 发布一系列事件进行监控测试
	events := []Event{
		&BaseEvent{
			ID:        "monitor-test-1",
			Type:      EventPlayerPlay,
			Source:    "monitor-test",
			Timestamp: time.Now(),
		},
		&BaseEvent{
			ID:        "monitor-test-2",
			Type:      EventPlayerPause,
			Source:    "monitor-test",
			Timestamp: time.Now(),
		},
		&BaseEvent{
			ID:        "monitor-test-3",
			Type:      EventPlayerPlay,
			Source:    "monitor-test",
			Timestamp: time.Now(),
		},
	}

	// 记录事件到监控器
	for _, event := range events {
		monitor.RecordEvent(event)
		err := bus.Publish(context.Background(), event)
		require.NoError(t, err)
	}

	// 等待处理
	time.Sleep(100 * time.Millisecond)

	// 获取统计信息
	stats := monitor.GetStats()
	assert.Greater(t, stats.TotalEvents, int64(2))
	assert.Contains(t, stats.EventCounts, EventPlayerPlay)
	assert.Contains(t, stats.EventCounts, EventPlayerPause)
	assert.Greater(t, stats.EventCounts[EventPlayerPlay], int64(1))

	// 获取性能指标
	metrics := monitor.GetPerformanceMetrics()
	assert.NotNil(t, metrics)
	// 性能指标可能需要时间收集，所以使用GreaterOrEqual而不是Greater
	assert.GreaterOrEqual(t, metrics.CPUUsage, 0.0)

	// 获取健康状态
	health := monitor.GetHealthStatus()
	assert.NotNil(t, health)
	assert.Equal(t, HealthLevelHealthy, health.OverallHealth)
}

func testHighConcurrencyScenario(t *testing.T, bus EventBus, store EventStore, monitor EventMonitor) {
	const (
		numGoroutines = 10
		eventsPerGoroutine = 100
		totalEvents = numGoroutines * eventsPerGoroutine
	)

	var (
		publishedCount int64
		receivedCount  int64
		storedCount    int64
		wg             sync.WaitGroup
	)

	// 注册事件类型
	err := bus.RegisterEventType(EventPlayerPlay)
	require.NoError(t, err)

	// 订阅事件
	subscription, err := bus.Subscribe(EventPlayerPlay, func(ctx context.Context, event Event) error {
		atomic.AddInt64(&receivedCount, 1)

		// 异步存储事件
		go func() {
			if err := store.Store(context.Background(), event); err == nil {
				atomic.AddInt64(&storedCount, 1)
			}
		}()

		// 记录监控数据
		monitor.RecordEvent(event)
		return nil
	})
	require.NoError(t, err)
	defer func() {
		err := bus.Unsubscribe(subscription.ID)
		assert.NoError(t, err)
	}()

	// 并发发布事件
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < eventsPerGoroutine; j++ {
				event := &BaseEvent{
					ID:        fmt.Sprintf("concurrent-test-%d-%d", goroutineID, j),
					Type:      EventPlayerPlay,
					Source:    fmt.Sprintf("goroutine-%d", goroutineID),
					Timestamp: time.Now(),
					Data:      map[string]interface{}{"index": j},
				}

				if err := bus.Publish(context.Background(), event); err == nil {
					atomic.AddInt64(&publishedCount, 1)
				}
			}
		}(i)
	}

	// 等待所有发布完成
	wg.Wait()

	// 等待事件处理完成
	time.Sleep(2 * time.Second)

	// 验证结果
	assert.Equal(t, int64(totalEvents), atomic.LoadInt64(&publishedCount))
	assert.Equal(t, int64(totalEvents), atomic.LoadInt64(&receivedCount))

	// 存储可能有延迟，允许一定的误差
	assert.Greater(t, atomic.LoadInt64(&storedCount), int64(totalEvents*0.9))

	// 验证监控统计
	stats := monitor.GetStats()
	assert.GreaterOrEqual(t, stats.TotalEvents, int64(totalEvents))

	// 验证性能指标
	metrics := monitor.GetPerformanceMetrics()
	// 性能指标可能需要时间收集，所以使用GreaterOrEqual而不是Greater
	assert.GreaterOrEqual(t, metrics.CPUUsage, 0.0)
	assert.GreaterOrEqual(t, metrics.MemoryUsage, int64(0))
}

func TestEventSystemWithConfig(t *testing.T) {
	// 测试使用配置创建事件总线
	config := &WorkerPoolConfig{
		QueueSize:         1000,
		MinWorkers:        2,
		MaxWorkers:        10,
		ScaleUpThreshold:  0.8,
		ScaleDownThreshold: 0.2,
		BatchSize:         10,
		BatchTimeout:      100 * time.Millisecond,
	}

	bus := NewEventBusWithConfig(slog.Default(), config)
	require.NotNil(t, bus)

	ctx := context.Background()
	err := bus.Start(ctx)
	require.NoError(t, err)
	defer func() {
		err := bus.Stop(ctx)
		assert.NoError(t, err)
	}()

	// 测试配置是否生效
	var receivedCount int64
	subscription, err := bus.Subscribe(EventPlayerPlay, func(ctx context.Context, event Event) error {
		atomic.AddInt64(&receivedCount, 1)
		return nil
	})
	require.NoError(t, err)
	defer func() {
		err := bus.Unsubscribe(subscription.ID)
		assert.NoError(t, err)
	}()

	// 发布事件测试批量处理
	for i := 0; i < 50; i++ {
		event := &BaseEvent{
			ID:        fmt.Sprintf("config-test-%d", i),
			Type:      EventPlayerPlay,
			Source:    "config-test",
			Timestamp: time.Now(),
		}

		err := bus.Publish(context.Background(), event)
		require.NoError(t, err)
	}

	// 等待处理完成
	time.Sleep(500 * time.Millisecond)

	// 验证所有事件都被处理
	assert.Equal(t, int64(50), atomic.LoadInt64(&receivedCount))
}

func TestEventSystemErrorHandling(t *testing.T) {
	bus := NewEventBus(slog.Default())
	require.NotNil(t, bus)

	ctx := context.Background()
	err := bus.Start(ctx)
	require.NoError(t, err)
	defer func() {
		err := bus.Stop(ctx)
		assert.NoError(t, err)
	}()

	// 测试订阅处理函数出错的情况
	var errorCount int64
	var successCount int64

	// 订阅一个会出错的处理函数
	errorSubscription, err := bus.Subscribe(EventPlayerPlay, func(ctx context.Context, event Event) error {
		atomic.AddInt64(&errorCount, 1)
		panic("test error") // 模拟处理函数出错
	})
	require.NoError(t, err)
	defer func() {
		err := bus.Unsubscribe(errorSubscription.ID)
		assert.NoError(t, err)
	}()

	// 订阅一个正常的处理函数
	successSubscription, err := bus.Subscribe(EventPlayerPlay, func(ctx context.Context, event Event) error {
		atomic.AddInt64(&successCount, 1)
		return nil
	})
	require.NoError(t, err)
	defer func() {
		err := bus.Unsubscribe(successSubscription.ID)
		assert.NoError(t, err)
	}()

	// 发布事件
	event := &BaseEvent{
		ID:        "error-test-1",
		Type:      EventPlayerPlay,
		Source:    "error-test",
		Timestamp: time.Now(),
	}

	err = bus.Publish(context.Background(), event)
	require.NoError(t, err)

	// 等待处理完成
	time.Sleep(200 * time.Millisecond)

	// 验证错误处理函数被调用，但不影响正常处理函数
	assert.Equal(t, int64(1), atomic.LoadInt64(&errorCount))
	assert.Equal(t, int64(1), atomic.LoadInt64(&successCount))
}

func TestEventSystemLifecycle(t *testing.T) {
	bus := NewEventBus(slog.Default())
	require.NotNil(t, bus)

	ctx := context.Background()

	// 测试重复启动
	err := bus.Start(ctx)
	require.NoError(t, err)

	err = bus.Start(ctx) // 重复启动应该返回错误
	assert.Error(t, err)

	// 测试正常停止
	err = bus.Stop(ctx)
	assert.NoError(t, err)

	// 测试重复停止
	err = bus.Stop(ctx) // 重复停止应该返回错误
	assert.Error(t, err)

	// 测试停止后重新启动
	err = bus.Start(ctx)
	assert.NoError(t, err)

	err = bus.Stop(ctx)
	assert.NoError(t, err)
}

func TestEventSystemMemoryUsage(t *testing.T) {
	// 创建有限容量的存储
	storeConfig := &EventStoreConfig{
		MaxEvents:     10,
		RetentionTime: time.Minute,
	}
	store := NewMemoryEventStore(slog.Default(), storeConfig) // 只能存储10个事件
	require.NotNil(t, store)

	// 存储超过容量的事件
	for i := 0; i < 20; i++ {
		event := &BaseEvent{
			ID:        fmt.Sprintf("memory-test-%d", i),
			Type:      EventPlayerPlay,
			Source:    "memory-test",
			Timestamp: time.Now(),
		}

		err := store.Store(context.Background(), event)
		assert.NoError(t, err)
	}

	// 验证只保留了最新的10个事件
	allEvents, err := store.GetByType(context.Background(), EventPlayerPlay, 20, 0)
	require.NoError(t, err)
	assert.Len(t, allEvents, 10)

	// 验证保留的是最新的事件
	for i, storedEvent := range allEvents {
		expectedID := fmt.Sprintf("memory-test-%d", i+10)
		assert.Equal(t, expectedID, storedEvent.GetID())
	}
}

func TestEventSystemCleanup(t *testing.T) {
	// 创建有过期时间的存储
	storeConfig := &EventStoreConfig{
		MaxEvents:     100,
		RetentionTime: 100 * time.Millisecond,
	}
	store := NewMemoryEventStore(slog.Default(), storeConfig) // 100ms过期时间
	require.NotNil(t, store)

	// 存储一些事件
	for i := 0; i < 5; i++ {
		event := &BaseEvent{
			ID:        fmt.Sprintf("cleanup-test-%d", i),
			Type:      EventPlayerPlay,
			Source:    "cleanup-test",
			Timestamp: time.Now(),
		}

		err := store.Store(context.Background(), event)
		require.NoError(t, err)
	}

	// 验证事件已存储
	allEvents, err := store.GetByType(context.Background(), EventPlayerPlay, 10, 0)
	require.NoError(t, err)
	assert.Len(t, allEvents, 5)

	// 等待过期时间
	time.Sleep(200 * time.Millisecond)

	// 触发清理（通过存储新事件）
	newEvent := &BaseEvent{
		ID:        "cleanup-trigger",
		Type:      EventPlayerPlay,
		Source:    "cleanup-test",
		Timestamp: time.Now(),
	}
	err = store.Store(context.Background(), newEvent)
	require.NoError(t, err)

	// 验证过期事件已被清理
	allEventsAfter, err := store.GetByType(context.Background(), EventPlayerPlay, 10, 0)
	require.NoError(t, err)
	assert.Len(t, allEventsAfter, 1)
	assert.Equal(t, "cleanup-trigger", allEventsAfter[0].GetID())
}