package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// EventStore 事件存储接口
type EventStore interface {
	// 存储事件
	Store(ctx context.Context, event Event) error
	StoreBatch(ctx context.Context, events []Event) error

	// 查询事件
	Get(ctx context.Context, eventID string) (Event, error)
	GetByType(ctx context.Context, eventType EventType, limit int, offset int) ([]Event, error)
	GetBySource(ctx context.Context, source string, limit int, offset int) ([]Event, error)
	GetByTimeRange(ctx context.Context, start, end time.Time, limit int, offset int) ([]Event, error)

	// 查询统计
	Count(ctx context.Context) (int64, error)
	CountByType(ctx context.Context, eventType EventType) (int64, error)
	CountBySource(ctx context.Context, source string) (int64, error)
	CountByTimeRange(ctx context.Context, start, end time.Time) (int64, error)

	// 删除事件
	Delete(ctx context.Context, eventID string) error
	DeleteByType(ctx context.Context, eventType EventType) error
	DeleteByTimeRange(ctx context.Context, start, end time.Time) error
	Clear(ctx context.Context) error

	// 生命周期管理
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	HealthCheck() error
}

// EventStoreConfig 事件存储配置
type EventStoreConfig struct {
	Type           string        `json:"type"`           // 存储类型：memory, file, database
	Path           string        `json:"path"`           // 存储路径
	MaxEvents      int64         `json:"max_events"`     // 最大事件数量
	RetentionTime  time.Duration `json:"retention_time"` // 事件保留时间
	FlushInterval  time.Duration `json:"flush_interval"` // 刷新间隔
	CompressionEnabled bool      `json:"compression_enabled"` // 是否启用压缩
}

// StoredEvent 存储的事件
type StoredEvent struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Metadata  map[string]string      `json:"metadata,omitempty"`
	StoredAt  time.Time              `json:"stored_at"`
}

// GetType 实现Event接口
func (e *StoredEvent) GetType() EventType {
	return e.Type
}

// GetData 实现Event接口
func (e *StoredEvent) GetData() interface{} {
	return e.Data
}

// GetTimestamp 实现Event接口
func (e *StoredEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

// GetSource 实现Event接口
func (e *StoredEvent) GetSource() string {
	return e.Source
}

// GetID 实现Event接口
func (e *StoredEvent) GetID() string {
	return e.ID
}

// MemoryEventStore 内存事件存储实现
type MemoryEventStore struct {
	logger *slog.Logger
	config *EventStoreConfig

	// 存储
	events    map[string]*StoredEvent
	typeIndex map[EventType][]string
	sourceIndex map[string][]string
	timeIndex []string // 按时间排序的事件ID列表
	mutex     sync.RWMutex

	// 生命周期
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// NewMemoryEventStore 创建内存事件存储
func NewMemoryEventStore(logger *slog.Logger, config *EventStoreConfig) *MemoryEventStore {
	ctx, cancel := context.WithCancel(context.Background())

	return &MemoryEventStore{
		logger:      logger,
		config:      config,
		events:      make(map[string]*StoredEvent),
		typeIndex:   make(map[EventType][]string),
		sourceIndex: make(map[string][]string),
		timeIndex:   make([]string, 0),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start 启动存储
func (s *MemoryEventStore) Start(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		return fmt.Errorf("memory event store is already running")
	}

	// 启动清理协程
	if s.config.RetentionTime > 0 {
		s.wg.Add(1)
		go s.cleanupWorker()
	}

	s.running = true
	s.logger.Info("Memory event store started")
	return nil
}

// Stop 停止存储
func (s *MemoryEventStore) Stop(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return fmt.Errorf("memory event store is not running")
	}

	s.cancel()
	s.wg.Wait()

	s.running = false
	s.logger.Info("Memory event store stopped")
	return nil
}

// HealthCheck 健康检查
func (s *MemoryEventStore) HealthCheck() error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if !s.running {
		return fmt.Errorf("memory event store is not running")
	}

	return nil
}

// Store 存储单个事件
func (s *MemoryEventStore) Store(ctx context.Context, event Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 先清理过期事件
	if s.config.RetentionTime > 0 {
		s.cleanupExpiredEventsNoLock()
	}

	// 检查是否超过最大事件数量
	if s.config.MaxEvents > 0 && int64(len(s.events)) >= s.config.MaxEvents {
		// 删除最旧的事件
		s.removeOldestEvent()
	}

	// 创建存储事件
	storedEvent := &StoredEvent{
		ID:        event.GetID(),
		Type:      event.GetType(),
		Source:    event.GetSource(),
		Timestamp: event.GetTimestamp(),
		StoredAt:  time.Now(),
	}

	// 序列化数据
	if event.GetData() != nil {
		if dataMap, ok := event.GetData().(map[string]interface{}); ok {
			storedEvent.Data = dataMap
		} else {
			// 尝试JSON序列化
			dataBytes, err := json.Marshal(event.GetData())
			if err != nil {
				return fmt.Errorf("failed to serialize event data: %w", err)
			}
			storedEvent.Data = map[string]interface{}{
				"_serialized": string(dataBytes),
			}
		}
	}

	// 存储事件
	s.events[storedEvent.ID] = storedEvent

	// 更新索引
	s.updateIndexes(storedEvent)

	s.logger.Debug("Event stored", "event_id", storedEvent.ID, "type", storedEvent.Type)
	return nil
}

// StoreBatch 批量存储事件
func (s *MemoryEventStore) StoreBatch(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 先清理过期事件
	if s.config.RetentionTime > 0 {
		s.cleanupExpiredEventsNoLock()
	}

	for _, event := range events {
		if event == nil {
			continue
		}

		// 检查是否超过最大事件数量
		if s.config.MaxEvents > 0 && int64(len(s.events)) >= s.config.MaxEvents {
			s.removeOldestEvent()
		}

		// 创建存储事件
		storedEvent := &StoredEvent{
			ID:        event.GetID(),
			Type:      event.GetType(),
			Source:    event.GetSource(),
			Timestamp: event.GetTimestamp(),
			StoredAt:  time.Now(),
		}

		// 序列化数据
		if event.GetData() != nil {
			if dataMap, ok := event.GetData().(map[string]interface{}); ok {
				storedEvent.Data = dataMap
			} else {
				dataBytes, err := json.Marshal(event.GetData())
				if err != nil {
					s.logger.Warn("Failed to serialize event data", "event_id", event.GetID(), "error", err)
					continue
				}
				storedEvent.Data = map[string]interface{}{
					"_serialized": string(dataBytes),
				}
			}
		}

		// 存储事件
		s.events[storedEvent.ID] = storedEvent

		// 更新索引
		s.updateIndexes(storedEvent)
	}

	s.logger.Debug("Events stored in batch", "count", len(events))
	return nil
}

// Get 获取单个事件
func (s *MemoryEventStore) Get(ctx context.Context, eventID string) (Event, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	event, exists := s.events[eventID]
	if !exists {
		return nil, fmt.Errorf("event not found: %s", eventID)
	}

	return event, nil
}

// GetByType 按类型获取事件
func (s *MemoryEventStore) GetByType(ctx context.Context, eventType EventType, limit int, offset int) ([]Event, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	eventIDs, exists := s.typeIndex[eventType]
	if !exists {
		return []Event{}, nil
	}

	return s.getEventsByIDs(eventIDs, limit, offset), nil
}

// GetBySource 按源获取事件
func (s *MemoryEventStore) GetBySource(ctx context.Context, source string, limit int, offset int) ([]Event, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	eventIDs, exists := s.sourceIndex[source]
	if !exists {
		return []Event{}, nil
	}

	return s.getEventsByIDs(eventIDs, limit, offset), nil
}

// GetByTimeRange 按时间范围获取事件
func (s *MemoryEventStore) GetByTimeRange(ctx context.Context, start, end time.Time, limit int, offset int) ([]Event, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var matchingIDs []string
	for _, eventID := range s.timeIndex {
		event := s.events[eventID]
		if event.Timestamp.After(start) && event.Timestamp.Before(end) {
			matchingIDs = append(matchingIDs, eventID)
		}
	}

	return s.getEventsByIDs(matchingIDs, limit, offset), nil
}

// Count 获取事件总数
func (s *MemoryEventStore) Count(ctx context.Context) (int64, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return int64(len(s.events)), nil
}

// CountByType 按类型统计事件数量
func (s *MemoryEventStore) CountByType(ctx context.Context, eventType EventType) (int64, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	eventIDs, exists := s.typeIndex[eventType]
	if !exists {
		return 0, nil
	}

	return int64(len(eventIDs)), nil
}

// CountBySource 按源统计事件数量
func (s *MemoryEventStore) CountBySource(ctx context.Context, source string) (int64, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	eventIDs, exists := s.sourceIndex[source]
	if !exists {
		return 0, nil
	}

	return int64(len(eventIDs)), nil
}

// CountByTimeRange 按时间范围统计事件数量
func (s *MemoryEventStore) CountByTimeRange(ctx context.Context, start, end time.Time) (int64, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var count int64
	for _, eventID := range s.timeIndex {
		event := s.events[eventID]
		if event.Timestamp.After(start) && event.Timestamp.Before(end) {
			count++
		}
	}

	return count, nil
}

// Delete 删除单个事件
func (s *MemoryEventStore) Delete(ctx context.Context, eventID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	event, exists := s.events[eventID]
	if !exists {
		return fmt.Errorf("event not found: %s", eventID)
	}

	// 从索引中移除
	s.removeFromIndexes(event)

	// 删除事件
	delete(s.events, eventID)

	s.logger.Debug("Event deleted", "event_id", eventID)
	return nil
}

// DeleteByType 按类型删除事件
func (s *MemoryEventStore) DeleteByType(ctx context.Context, eventType EventType) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	eventIDs, exists := s.typeIndex[eventType]
	if !exists {
		return nil
	}

	for _, eventID := range eventIDs {
		event := s.events[eventID]
		s.removeFromIndexes(event)
		delete(s.events, eventID)
	}

	s.logger.Debug("Events deleted by type", "type", eventType, "count", len(eventIDs))
	return nil
}

// DeleteByTimeRange 按时间范围删除事件
func (s *MemoryEventStore) DeleteByTimeRange(ctx context.Context, start, end time.Time) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var toDelete []string
	for _, eventID := range s.timeIndex {
		event := s.events[eventID]
		if event.Timestamp.After(start) && event.Timestamp.Before(end) {
			toDelete = append(toDelete, eventID)
		}
	}

	for _, eventID := range toDelete {
		event := s.events[eventID]
		s.removeFromIndexes(event)
		delete(s.events, eventID)
	}

	s.logger.Debug("Events deleted by time range", "count", len(toDelete))
	return nil
}

// Clear 清空所有事件
func (s *MemoryEventStore) Clear(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	count := len(s.events)
	s.events = make(map[string]*StoredEvent)
	s.typeIndex = make(map[EventType][]string)
	s.sourceIndex = make(map[string][]string)
	s.timeIndex = make([]string, 0)

	s.logger.Debug("All events cleared", "count", count)
	return nil
}

// 辅助方法

// updateIndexes 更新索引
func (s *MemoryEventStore) updateIndexes(event *StoredEvent) {
	// 更新类型索引
	s.typeIndex[event.Type] = append(s.typeIndex[event.Type], event.ID)

	// 更新源索引
	s.sourceIndex[event.Source] = append(s.sourceIndex[event.Source], event.ID)

	// 更新时间索引（保持排序）
	s.insertIntoTimeIndex(event.ID, event.Timestamp)
}

// removeFromIndexes 从索引中移除
func (s *MemoryEventStore) removeFromIndexes(event *StoredEvent) {
	// 从类型索引中移除
	if typeIDs, exists := s.typeIndex[event.Type]; exists {
		s.typeIndex[event.Type] = s.removeFromSlice(typeIDs, event.ID)
		if len(s.typeIndex[event.Type]) == 0 {
			delete(s.typeIndex, event.Type)
		}
	}

	// 从源索引中移除
	if sourceIDs, exists := s.sourceIndex[event.Source]; exists {
		s.sourceIndex[event.Source] = s.removeFromSlice(sourceIDs, event.ID)
		if len(s.sourceIndex[event.Source]) == 0 {
			delete(s.sourceIndex, event.Source)
		}
	}

	// 从时间索引中移除
	s.timeIndex = s.removeFromSlice(s.timeIndex, event.ID)
}

// insertIntoTimeIndex 插入到时间索引（保持排序）
func (s *MemoryEventStore) insertIntoTimeIndex(eventID string, timestamp time.Time) {
	// 找到正确的插入位置以保持时间排序
	insertPos := len(s.timeIndex)
	for i, existingID := range s.timeIndex {
		if existingEvent, exists := s.events[existingID]; exists {
			if timestamp.Before(existingEvent.Timestamp) {
				insertPos = i
				break
			}
		}
	}
	
	// 插入到正确位置
	if insertPos == len(s.timeIndex) {
		s.timeIndex = append(s.timeIndex, eventID)
	} else {
		s.timeIndex = append(s.timeIndex[:insertPos+1], s.timeIndex[insertPos:]...)
		s.timeIndex[insertPos] = eventID
	}
}

// removeFromSlice 从切片中移除元素
func (s *MemoryEventStore) removeFromSlice(slice []string, item string) []string {
	for i, v := range slice {
		if v == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// getEventsByIDs 根据ID列表获取事件
func (s *MemoryEventStore) getEventsByIDs(eventIDs []string, limit int, offset int) []Event {
	if offset >= len(eventIDs) {
		return []Event{}
	}

	end := offset + limit
	if limit <= 0 || end > len(eventIDs) {
		end = len(eventIDs)
	}

	result := make([]Event, 0, end-offset)
	for i := offset; i < end; i++ {
		if event, exists := s.events[eventIDs[i]]; exists {
			result = append(result, event)
		}
	}

	return result
}

// removeOldestEvent 移除最旧的事件
func (s *MemoryEventStore) removeOldestEvent() {
	if len(s.timeIndex) == 0 {
		return
	}

	oldestID := s.timeIndex[0]
	if event, exists := s.events[oldestID]; exists {
		s.removeFromIndexes(event)
		delete(s.events, oldestID)
	}
}

// cleanupWorker 清理工作协程
func (s *MemoryEventStore) cleanupWorker() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.RetentionTime / 10) // 每1/10保留时间检查一次
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.cleanupExpiredEvents()
		}
	}
}

// cleanupExpiredEvents 清理过期事件
func (s *MemoryEventStore) cleanupExpiredEvents() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.cleanupExpiredEventsNoLock()
}

// cleanupExpiredEventsNoLock 清理过期事件（不加锁版本）
func (s *MemoryEventStore) cleanupExpiredEventsNoLock() {
	now := time.Now()
	expiredBefore := now.Add(-s.config.RetentionTime)

	var toDelete []string
	for _, eventID := range s.timeIndex {
		event := s.events[eventID]
		if event.StoredAt.Before(expiredBefore) {
			toDelete = append(toDelete, eventID)
		} else {
			break // 由于时间索引是排序的，后面的事件都不会过期
		}
	}

	for _, eventID := range toDelete {
		event := s.events[eventID]
		s.removeFromIndexes(event)
		delete(s.events, eventID)
	}

	if len(toDelete) > 0 {
		s.logger.Debug("Expired events cleaned up", "count", len(toDelete))
	}
}