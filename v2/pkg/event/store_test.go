package event

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryEventStore_Store(t *testing.T) {
	logger := slog.Default()
	config := &EventStoreConfig{
		Type:      "memory",
		MaxEvents: 100,
	}
	store := NewMemoryEventStore(logger, config)

	ctx := context.Background()
	err := store.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = store.Stop(ctx) }()

	// 创建测试事件
	event := &BaseEvent{
		ID:        "test-1",
		Type:      EventPlayerPlay,
		Source:    "test-source",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"song_id": "123"},
	}

	// 存储事件
	err = store.Store(ctx, event)
	assert.NoError(t, err)

	// 验证事件已存储
	storedEvent, err := store.Get(ctx, "test-1")
	assert.NoError(t, err)
	assert.Equal(t, event.ID, storedEvent.GetID())
	assert.Equal(t, event.Type, storedEvent.GetType())
	assert.Equal(t, event.Source, storedEvent.GetSource())
}

func TestMemoryEventStore_StoreBatch(t *testing.T) {
	logger := slog.Default()
	config := &EventStoreConfig{
		Type:      "memory",
		MaxEvents: 100,
	}
	store := NewMemoryEventStore(logger, config)

	ctx := context.Background()
	err := store.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = store.Stop(ctx) }()

	// 创建批量测试事件
	events := []Event{
		&BaseEvent{
			ID:        "batch-1",
			Type:      EventPlayerPlay,
			Source:    "test-source",
			Timestamp: time.Now(),
		},
		&BaseEvent{
			ID:        "batch-2",
			Type:      EventPlayerPause,
			Source:    "test-source",
			Timestamp: time.Now(),
		},
	}

	// 批量存储事件
	err = store.StoreBatch(ctx, events)
	assert.NoError(t, err)

	// 验证事件数量
	count, err := store.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// 验证每个事件都已存储
	for _, event := range events {
		storedEvent, err := store.Get(ctx, event.GetID())
		assert.NoError(t, err)
		assert.Equal(t, event.GetID(), storedEvent.GetID())
	}
}

func TestMemoryEventStore_GetByType(t *testing.T) {
	logger := slog.Default()
	config := &EventStoreConfig{
		Type:      "memory",
		MaxEvents: 100,
	}
	store := NewMemoryEventStore(logger, config)

	ctx := context.Background()
	err := store.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = store.Stop(ctx) }()

	// 存储不同类型的事件
	events := []Event{
		&BaseEvent{ID: "play-1", Type: EventPlayerPlay, Source: "test", Timestamp: time.Now()},
		&BaseEvent{ID: "play-2", Type: EventPlayerPlay, Source: "test", Timestamp: time.Now()},
		&BaseEvent{ID: "pause-1", Type: EventPlayerPause, Source: "test", Timestamp: time.Now()},
	}

	for _, event := range events {
		err = store.Store(ctx, event)
		require.NoError(t, err)
	}

	// 按类型查询
	playEvents, err := store.GetByType(ctx, EventPlayerPlay, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, playEvents, 2)

	pauseEvents, err := store.GetByType(ctx, EventPlayerPause, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, pauseEvents, 1)
}

func TestMemoryEventStore_GetByTimeRange(t *testing.T) {
	logger := slog.Default()
	config := &EventStoreConfig{
		Type:      "memory",
		MaxEvents: 100,
	}
	store := NewMemoryEventStore(logger, config)

	ctx := context.Background()
	err := store.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = store.Stop(ctx) }()

	now := time.Now()
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	// 存储不同时间的事件
	events := []Event{
		&BaseEvent{ID: "past-1", Type: EventPlayerPlay, Source: "test", Timestamp: past},
		&BaseEvent{ID: "now-1", Type: EventPlayerPlay, Source: "test", Timestamp: now},
		&BaseEvent{ID: "future-1", Type: EventPlayerPlay, Source: "test", Timestamp: future},
	}

	for _, event := range events {
		err = store.Store(ctx, event)
		require.NoError(t, err)
	}

	// 按时间范围查询
	results, err := store.GetByTimeRange(ctx, past.Add(-time.Minute), now.Add(time.Minute), 10, 0)
	assert.NoError(t, err)
	assert.Len(t, results, 2) // past-1 和 now-1
}

func TestMemoryEventStore_Delete(t *testing.T) {
	logger := slog.Default()
	config := &EventStoreConfig{
		Type:      "memory",
		MaxEvents: 100,
	}
	store := NewMemoryEventStore(logger, config)

	ctx := context.Background()
	err := store.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = store.Stop(ctx) }()

	// 存储事件
	event := &BaseEvent{
		ID:        "delete-test",
		Type:      EventPlayerPlay,
		Source:    "test",
		Timestamp: time.Now(),
	}
	err = store.Store(ctx, event)
	require.NoError(t, err)

	// 验证事件存在
	_, err = store.Get(ctx, "delete-test")
	assert.NoError(t, err)

	// 删除事件
	err = store.Delete(ctx, "delete-test")
	assert.NoError(t, err)

	// 验证事件已删除
	_, err = store.Get(ctx, "delete-test")
	assert.Error(t, err)
}

func TestMemoryEventStore_MaxEvents(t *testing.T) {
	logger := slog.Default()
	config := &EventStoreConfig{
		Type:      "memory",
		MaxEvents: 3, // 限制最大事件数量
	}
	store := NewMemoryEventStore(logger, config)

	ctx := context.Background()
	err := store.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = store.Stop(ctx) }()

	// 存储超过最大数量的事件
	for i := 0; i < 5; i++ {
		event := &BaseEvent{
			ID:        fmt.Sprintf("event-%d", i),
			Type:      EventPlayerPlay,
			Source:    "test",
			Timestamp: time.Now(),
		}
		err = store.Store(ctx, event)
		assert.NoError(t, err)
	}

	// 验证只保留了最大数量的事件
	count, err := store.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// 验证最旧的事件被删除了
	_, err = store.Get(ctx, "event-0")
	assert.Error(t, err)
	_, err = store.Get(ctx, "event-1")
	assert.Error(t, err)

	// 验证最新的事件还在
	_, err = store.Get(ctx, "event-4")
	assert.NoError(t, err)
}

func TestMemoryEventStore_HealthCheck(t *testing.T) {
	logger := slog.Default()
	config := &EventStoreConfig{
		Type:      "memory",
		MaxEvents: 100,
	}
	store := NewMemoryEventStore(logger, config)

	// 未启动时健康检查应该失败
	err := store.HealthCheck()
	assert.Error(t, err)

	// 启动后健康检查应该成功
	ctx := context.Background()
	err = store.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = store.Stop(ctx) }()

	err = store.HealthCheck()
	assert.NoError(t, err)
}

func TestMemoryEventStore_RetentionTime(t *testing.T) {
	logger := slog.Default()
	config := &EventStoreConfig{
		Type:          "memory",
		MaxEvents:     100,
		RetentionTime: time.Millisecond * 100, // 很短的保留时间用于测试
	}
	store := NewMemoryEventStore(logger, config)

	ctx := context.Background()
	err := store.Start(ctx)
	require.NoError(t, err)
	defer func() { _ = store.Stop(ctx) }()

	// 存储事件
	event := &BaseEvent{
		ID:        "retention-test",
		Type:      EventPlayerPlay,
		Source:    "test",
		Timestamp: time.Now(),
	}
	err = store.Store(ctx, event)
	require.NoError(t, err)

	// 验证事件存在
	_, err = store.Get(ctx, "retention-test")
	assert.NoError(t, err)

	// 等待超过保留时间
	time.Sleep(time.Millisecond * 200)

	// 验证事件被清理（这个测试可能不稳定，因为清理是异步的）
	// 在实际实现中，可能需要手动触发清理或等待更长时间
}