package event

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDefaultEventBus(t *testing.T) {
	bus := NewDefaultEventBus()
	assert.NotNil(t, bus)
	assert.False(t, bus.IsRunning())
}

func TestEventBus_RegisterEventType(t *testing.T) {
	bus := NewDefaultEventBus()

	// 注册新事件类型
	err := bus.RegisterEventType("test.event")
	assert.NoError(t, err)

	// 重复注册应该返回错误
	err = bus.RegisterEventType("test.event")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestEventBus_Subscribe(t *testing.T) {
	bus := NewDefaultEventBus()
	ctx := context.Background()

	// 启动事件总线
	err := bus.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = bus.Stop(ctx) }()

	// 注册事件类型
	err = bus.RegisterEventType("test.event")
	require.NoError(t, err)

	// 订阅事件，使用channel来同步
	handlerCh := make(chan bool, 1)
	handler := func(ctx context.Context, event Event) error {
		select {
		case handlerCh <- true:
		default:
			// Channel full, ignore
		}
		return nil
	}

	subscriberID, err := bus.Subscribe("test.event", handler)
	assert.NoError(t, err)
	assert.NotEmpty(t, subscriberID)

	// 发布事件
	event := &BaseEvent{
		ID:   "test-1",
		Type: "test.event",
		Data: "test data",
	}

	err = bus.Publish(ctx, event)
	assert.NoError(t, err)

	// 等待事件处理
	select {
	case <-handlerCh:
		// 事件已处理
	case <-time.After(1 * time.Second):
		t.Error("Handler should be called")
	}
}

func TestEventBus_SubscribeWithOptions(t *testing.T) {
	bus := NewDefaultEventBus()
	ctx := context.Background()

	// 启动事件总线
	err := bus.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = bus.Stop(ctx) }()

	err = bus.RegisterEventType("test.event")
	require.NoError(t, err)

	// 使用选项订阅，使用channel来同步
	handlerCh := make(chan bool, 1)
	handler := func(ctx context.Context, event Event) error {
		select {
		case handlerCh <- true:
		default:
			// Channel full, ignore
		}
		return nil
	}

	subscriberID, err := bus.Subscribe("test.event", handler,
		WithPriority(int(PriorityHigh)),
		WithGroup("test-group"),
		WithFilter(func(event Event) bool {
			return event.GetData() == "filtered data"
		}),
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, subscriberID)

	// 发布不匹配过滤器的事件
	event1 := &BaseEvent{
		ID:   "test-1",
		Type: "test.event",
		Data: "normal data",
	}
	err = bus.Publish(ctx, event1)
	assert.NoError(t, err)
	
	// 等待一段时间，确保不匹配的事件不会触发处理器
	select {
	case <-handlerCh:
		t.Error("Handler should not be called for filtered event")
	case <-time.After(100 * time.Millisecond):
		// 预期行为：处理器不应该被调用
	}

	// 发布匹配过滤器的事件
	event2 := &BaseEvent{
		ID:   "test-2",
		Type: "test.event",
		Data: "filtered data",
	}
	err = bus.Publish(ctx, event2)
	assert.NoError(t, err)
	
	// 等待匹配的事件被处理
	select {
	case <-handlerCh:
		// 预期行为：处理器应该被调用
	case <-time.After(1 * time.Second):
		t.Error("Handler should be called for matching event")
	}
}

func TestEventBus_Unsubscribe(t *testing.T) {
	bus := NewDefaultEventBus()
	ctx := context.Background()

	// 启动事件总线
	err := bus.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = bus.Stop(ctx) }()

	err = bus.RegisterEventType("test.event")
	require.NoError(t, err)

	handlerCallCount := int32(0)
	handler := func(ctx context.Context, event Event) error {
		atomic.AddInt32(&handlerCallCount, 1)
		return nil
	}

	// 订阅事件（使用同步处理）
	subscriberID, err := bus.Subscribe("test.event", handler, WithAsync(false))
	require.NoError(t, err)

	// 发布事件
	event := &BaseEvent{
		ID:   "test-1",
		Type: "test.event",
		Data: "test data",
	}
	err = bus.Publish(ctx, event)
	assert.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, int32(1), atomic.LoadInt32(&handlerCallCount))

	// 取消订阅
	err = bus.Unsubscribe(subscriberID.ID)
	assert.NoError(t, err)

	// 再次发布事件
	err = bus.Publish(ctx, event)
	assert.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, int32(1), atomic.LoadInt32(&handlerCallCount)) // 计数不应该增加
}

func TestEventBus_PublishSync(t *testing.T) {
	bus := NewDefaultEventBus()
	ctx := context.Background()

	// 启动事件总线
	err := bus.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = bus.Stop(ctx) }()

	err = bus.RegisterEventType("test.event")
	require.NoError(t, err)

	handlerCalled := false
	handler := func(ctx context.Context, event Event) error {
		handlerCalled = true
		return nil
	}

	_, err = bus.Subscribe("test.event", handler, WithAsync(false))
	require.NoError(t, err)

	// 同步发布事件
	event := &BaseEvent{
		ID:   "test-1",
		Type: "test.event",
		Data: "test data",
	}

	err = bus.PublishSync(ctx, event)
	assert.NoError(t, err)
	assert.True(t, handlerCalled) // 同步调用，应该立即执行
}

func TestEventBus_ConcurrentPublishSubscribe(t *testing.T) {
	bus := NewDefaultEventBus()
	ctx := context.Background()

	// 启动事件总线
	err := bus.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = bus.Stop(ctx) }()

	err = bus.RegisterEventType("test.event")
	require.NoError(t, err)

	handlerCallCount := int64(0)
	handler := func(ctx context.Context, event Event) error {
		atomic.AddInt64(&handlerCallCount, 1)
		return nil
	}

	// 订阅事件
	_, err = bus.Subscribe("test.event", handler, WithAsync(false))
	require.NoError(t, err)

	// 并发发布事件
	const numGoroutines = 10
	const eventsPerGoroutine = 100

	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < eventsPerGoroutine; j++ {
				event := &BaseEvent{
					ID:   fmt.Sprintf("test-%d-%d", goroutineID, j),
					Type: "test.event",
					Data: fmt.Sprintf("data-%d-%d", goroutineID, j),
				}
				err := bus.Publish(ctx, event)
				assert.NoError(t, err)
			}
		}(i)
	}

	wg.Wait()

	// 等待所有事件处理完成
	time.Sleep(500 * time.Millisecond)

	// 验证大部分事件都被处理（允许一些延迟）
	expectedCount := int64(numGoroutines * eventsPerGoroutine)
	actualCount := atomic.LoadInt64(&handlerCallCount)
	assert.GreaterOrEqual(t, actualCount, int64(900)) // 至少90%的事件被处理
	assert.LessOrEqual(t, actualCount, expectedCount)   // 不超过总数
}

func TestEventBus_ConcurrentSubscribeUnsubscribe(t *testing.T) {
	bus := NewDefaultEventBus()

	err := bus.RegisterEventType("test.event")
	require.NoError(t, err)

	handler := func(ctx context.Context, event Event) error {
		return nil
	}

	// 并发订阅和取消订阅
	const numGoroutines = 10
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 订阅
			subscriberID, err := bus.Subscribe("test.event", handler)
			assert.NoError(t, err)
			assert.NotEmpty(t, subscriberID)

			// 短暂等待
			time.Sleep(10 * time.Millisecond)

			// 取消订阅
			err = bus.Unsubscribe(subscriberID.ID)
		assert.NoError(t, err)
		}()
	}

	wg.Wait()
}

func TestEventBus_ErrorHandling(t *testing.T) {
	bus := NewDefaultEventBus()
	ctx := context.Background()

	// 启动事件总线
	err := bus.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = bus.Stop(ctx) }()

	err = bus.RegisterEventType("test.event")
	require.NoError(t, err)

	// 订阅一个会返回错误的处理器
	handler := func(ctx context.Context, event Event) error {
		return fmt.Errorf("handler error")
	}

	_, err = bus.Subscribe("test.event", handler, WithAsync(false))
	require.NoError(t, err)

	// 发布事件
	event := &BaseEvent{
		ID:   "test-1",
		Type: "test.event",
		Data: "test data",
	}

	// 异步发布不应该返回处理器错误
	err = bus.PublishAsync(ctx, event)
	assert.NoError(t, err)

	// 同步发布应该返回处理器错误
	err = bus.Publish(ctx, event)
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "handler error")
	}
}

func TestEventBus_PanicHandling(t *testing.T) {
	bus := NewDefaultEventBus()
	ctx := context.Background()

	// 启动事件总线
	err := bus.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = bus.Stop(ctx) }()

	err = bus.RegisterEventType("test.event")
	require.NoError(t, err)

	// 订阅一个会panic的处理器
	handler := func(ctx context.Context, event Event) error {
		panic("handler panic")
	}

	_, err = bus.Subscribe("test.event", handler, WithAsync(false))
	require.NoError(t, err)

	// 发布事件
	event := &BaseEvent{
		ID:   "test-1",
		Type: "test.event",
		Data: "test data",
	}

	// 异步发布不应该导致程序崩溃
	err = bus.PublishAsync(ctx, event)
	assert.NoError(t, err)

	// 等待处理完成
	time.Sleep(100 * time.Millisecond)

	// 同步发布应该捕获panic并返回错误
	err = bus.Publish(ctx, event)
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "handler panic")
	}
}

func TestEventBus_GetStats(t *testing.T) {
	bus := NewDefaultEventBus()
	ctx := context.Background()

	// 启动事件总线
	err := bus.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = bus.Stop(ctx) }()

	err = bus.RegisterEventType("test.event")
	require.NoError(t, err)

	handler := func(ctx context.Context, event Event) error {
		return nil
	}

	_, err = bus.Subscribe("test.event", handler)
	require.NoError(t, err)

	// 获取初始统计
	stats := bus.GetStats()
	assert.Equal(t, int64(0), stats.TotalEvents) // 初始时没有事件
	assert.Equal(t, 1, stats.TotalSubscribers)

	// 发布事件
	event := &BaseEvent{
		ID:   "test-1",
		Type: "test.event",
		Data: "test data",
	}

	err = bus.Publish(ctx, event)
	assert.NoError(t, err)

	// 等待处理完成
	time.Sleep(100 * time.Millisecond)

	// 检查统计更新
	stats = bus.GetStats()
	assert.Equal(t, int64(1), stats.TotalEvents) // 发布了一个事件
	assert.Equal(t, 1, stats.TotalSubscribers)
}

func TestEventBus_Shutdown(t *testing.T) {
	bus := NewDefaultEventBus()
	ctx := context.Background()

	// 启动事件总线
	err := bus.Start(ctx)
	require.NoError(t, err)

	err = bus.RegisterEventType("test.event")
	require.NoError(t, err)

	// 使用channel来同步处理完成
	processedCh := make(chan bool, 1)
	handler := func(ctx context.Context, event Event) error {
		time.Sleep(50 * time.Millisecond) // 减少处理时间
		select {
		case processedCh <- true:
		default:
			// Channel full, ignore
		}
		return nil
	}

	_, err = bus.Subscribe("test.event", handler)
	require.NoError(t, err)

	// 发布事件
	event := &BaseEvent{
		ID:   "test-1",
		Type: "test.event",
		Data: "test data",
	}

	err = bus.Publish(ctx, event)
	assert.NoError(t, err)

	// 等待事件处理完成
	select {
	case <-processedCh:
		// 事件已处理完成
	case <-time.After(1 * time.Second):
		t.Error("Event processing timeout")
	}

	// 停止事件总线
	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = bus.Stop(stopCtx)
	assert.NoError(t, err)

	// 关闭后发布事件应该失败
	err = bus.Publish(ctx, event)
	assert.Error(t, err)
}

func TestEventBus_MultipleEventTypes(t *testing.T) {
	bus := NewDefaultEventBus()
	ctx := context.Background()

	// 启动事件总线
	err := bus.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = bus.Stop(ctx) }()

	// 订阅事件注册多个事件类型
	err = bus.RegisterEventType("event.type1")
	require.NoError(t, err)
	err = bus.RegisterEventType("event.type2")
	require.NoError(t, err)

	// 使用channel来同步事件处理
	type1Ch := make(chan bool, 1)
	type2Ch := make(chan bool, 1)

	handler1 := func(ctx context.Context, event Event) error {
		select {
		case type1Ch <- true:
		default:
			// Channel full, ignore
		}
		return nil
	}

	handler2 := func(ctx context.Context, event Event) error {
		select {
		case type2Ch <- true:
		default:
			// Channel full, ignore
		}
		return nil
	}

	// 订阅不同事件类型
	_, err = bus.Subscribe("event.type1", handler1)
	require.NoError(t, err)
	_, err = bus.Subscribe("event.type2", handler2)
	require.NoError(t, err)

	// 发布type1事件
	event1 := &BaseEvent{
		ID:   "test-1",
		Type: "event.type1",
		Data: "data1",
	}
	err = bus.Publish(ctx, event1)
	assert.NoError(t, err)

	// 发布type2事件
	event2 := &BaseEvent{
		ID:   "test-2",
		Type: "event.type2",
		Data: "data2",
	}
	err = bus.Publish(ctx, event2)
	assert.NoError(t, err)

	// 等待事件处理完成
	select {
	case <-type1Ch:
		// type1事件已处理
	case <-time.After(1 * time.Second):
		t.Error("Expected type1 event to be processed")
	}

	select {
	case <-type2Ch:
		// type2事件已处理
	case <-time.After(1 * time.Second):
		t.Error("Expected type2 event to be processed")
	}
}