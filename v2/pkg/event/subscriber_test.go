package event

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSubscriberManager(t *testing.T) {
	sm := NewSubscriberManager()
	assert.NotNil(t, sm)
	assert.NotNil(t, sm.subscribers)
	assert.NotNil(t, sm.groupMap)
	assert.NotNil(t, sm.stats)
}

func TestSubscriberManager_AddSubscriber(t *testing.T) {
	sm := NewSubscriberManager()

	handler := func(ctx context.Context, event Event) error {
		return nil
	}

	// 添加基本订阅者
	subscriberID := sm.AddSubscriber("test.event", handler)
	assert.NotEmpty(t, subscriberID)

	// 验证订阅者被添加
	subscriber, exists := sm.GetSubscriber(subscriberID)
	assert.True(t, exists)
	assert.Equal(t, "test.event", string(subscriber.EventType))
	assert.True(t, subscriber.Active)
	assert.Equal(t, PriorityNormal, subscriber.Priority)

	// 验证统计信息
	stats := sm.GetStats()
	assert.Equal(t, int64(1), stats.TotalSubscribers)
	assert.Equal(t, int64(1), stats.ActiveSubscribers)
}

func TestSubscriberManager_RemoveSubscriber(t *testing.T) {
	sm := NewSubscriberManager()

	handler := func(ctx context.Context, event Event) error {
		return nil
	}

	// 添加订阅者
	subscriberID := sm.AddSubscriber("test.event", handler, WithGroup("test-group"))

	// 验证订阅者存在
	_, exists := sm.GetSubscriber(subscriberID)
	assert.True(t, exists)

	// 移除订阅者
	removed := sm.RemoveSubscriber(subscriberID)
	assert.True(t, removed)

	// 验证订阅者不存在
	_, exists = sm.GetSubscriber(subscriberID)
	assert.False(t, exists)

	// 移除不存在的订阅者
	removed = sm.RemoveSubscriber("non-existent")
	assert.False(t, removed)
}

func TestSubscriberManager_ExecuteHandler(t *testing.T) {
	sm := NewSubscriberManager()
	ctx := context.Background()

	// 测试正常处理器
	handlerCalled := false
	handler := func(ctx context.Context, event Event) error {
		handlerCalled = true
		return nil
	}

	subscriberID := sm.AddSubscriber("test.event", handler)
	subscriber, _ := sm.GetSubscriber(subscriberID)

	event := &BaseEvent{
		ID:   "test-1",
		Type: "test.event",
		Data: "test data",
	}

	err := sm.ExecuteHandler(ctx, subscriber, event)
	assert.NoError(t, err)
	assert.True(t, handlerCalled)
}

func TestSubscriberManager_ConcurrentOperations(t *testing.T) {
	sm := NewSubscriberManager()

	handler := func(ctx context.Context, event Event) error {
		return nil
	}

	// 并发添加和移除订阅者
	const numGoroutines = 5
	const operationsPerGoroutine = 10

	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				// 添加订阅者
				subscriberID := sm.AddSubscriber(EventType(fmt.Sprintf("event.%d.%d", goroutineID, j)), handler)

				// 获取订阅者
				_, exists := sm.GetSubscriber(subscriberID)
				assert.True(t, exists)

				// 移除订阅者
				removed := sm.RemoveSubscriber(subscriberID)
				assert.True(t, removed)
			}
		}(i)
	}

	wg.Wait()

	// 验证最终状态
	assert.Equal(t, 0, sm.GetActiveSubscriberCount())
}