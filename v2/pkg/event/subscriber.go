package event

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// SubscriberManager 订阅者管理器
type SubscriberManager struct {
	mu          sync.RWMutex
	subscribers map[string]*SubscriberInfo
	groupMap    map[string][]string // group -> subscriber IDs
	stats       *SubscriberStats
}

// SubscriberStats 订阅者统计信息
type SubscriberStats struct {
	mu                sync.RWMutex
	TotalSubscribers  int64
	ActiveSubscribers int64
	EventsSent        int64
	ErrorCount        int64
}

// NewSubscriberManager 创建订阅者管理器
func NewSubscriberManager() *SubscriberManager {
	return &SubscriberManager{
		subscribers: make(map[string]*SubscriberInfo),
		groupMap:    make(map[string][]string),
		stats:       &SubscriberStats{},
	}
}

// AddSubscriber 添加订阅者
func (sm *SubscriberManager) AddSubscriber(eventType EventType, handler EventHandler, options ...SubscribeOption) string {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	subscriber := &SubscriberInfo{
		ID:        uuid.New().String(),
		EventType: eventType,
		Handler:   handler,
		Priority:  PriorityNormal,
		Async:     true,
		Active:    true,
		CreatedAt: time.Now(),
	}

	subscribeOpts := &SubscribeOptions{}
	for _, option := range options {
		option(subscribeOpts)
	}

	// 应用选项到订阅者信息
	if subscribeOpts.Priority != 0 {
		subscriber.Priority = EventPriority(subscribeOpts.Priority)
	}
	if subscribeOpts.Group != "" {
		subscriber.Group = subscribeOpts.Group
	}
	if subscribeOpts.Filter != nil {
		subscriber.Filter = subscribeOpts.Filter
	}
	subscriber.Async = subscribeOpts.Async

	sm.subscribers[subscriber.ID] = subscriber

	// 添加到组映射
	if subscriber.Group != "" {
		sm.groupMap[subscriber.Group] = append(sm.groupMap[subscriber.Group], subscriber.ID)
	}

	// 更新统计
	sm.stats.mu.Lock()
	sm.stats.TotalSubscribers++
	sm.stats.ActiveSubscribers++
	sm.stats.mu.Unlock()

	return subscriber.ID
}

// RemoveSubscriber 移除订阅者
func (sm *SubscriberManager) RemoveSubscriber(subscriberID string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	subscriber, exists := sm.subscribers[subscriberID]
	if !exists {
		return false
	}

	// 从组映射中移除
	if subscriber.Group != "" {
		if groupIDs, exists := sm.groupMap[subscriber.Group]; exists {
			for i, id := range groupIDs {
				if id == subscriberID {
					sm.groupMap[subscriber.Group] = append(groupIDs[:i], groupIDs[i+1:]...)
					break
				}
			}
			// 如果组为空，删除组
			if len(sm.groupMap[subscriber.Group]) == 0 {
				delete(sm.groupMap, subscriber.Group)
			}
		}
	}

	delete(sm.subscribers, subscriberID)

	// 更新统计
	sm.stats.mu.Lock()
	if subscriber.Active {
		sm.stats.ActiveSubscribers--
	}
	sm.stats.mu.Unlock()

	return true
}

// GetSubscriber 获取订阅者信息
func (sm *SubscriberManager) GetSubscriber(subscriberID string) (*SubscriberInfo, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	subscriber, exists := sm.subscribers[subscriberID]
	if !exists {
		return nil, false
	}

	// 返回副本以避免并发修改
	copy := *subscriber
	return &copy, true
}

// GetSubscribersByEventType 根据事件类型获取订阅者
func (sm *SubscriberManager) GetSubscribersByEventType(eventType EventType) []*SubscriberInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var subscribers []*SubscriberInfo
	for _, subscriber := range sm.subscribers {
		if subscriber.Active && (subscriber.EventType == eventType || subscriber.EventType == EventTypeAll) {
			// 返回副本以避免并发修改
			copy := *subscriber
			subscribers = append(subscribers, &copy)
		}
	}

	return subscribers
}

// GetSubscribersByGroup 根据组获取订阅者
func (sm *SubscriberManager) GetSubscribersByGroup(group string) []*SubscriberInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	subscriberIDs, exists := sm.groupMap[group]
	if !exists {
		return nil
	}

	var subscribers []*SubscriberInfo
	for _, id := range subscriberIDs {
		if subscriber, exists := sm.subscribers[id]; exists && subscriber.Active {
			// 返回副本以避免并发修改
			copy := *subscriber
			subscribers = append(subscribers, &copy)
		}
	}

	return subscribers
}

// SetSubscriberActive 设置订阅者活跃状态
func (sm *SubscriberManager) SetSubscriberActive(subscriberID string, active bool) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	subscriber, exists := sm.subscribers[subscriberID]
	if !exists {
		return false
	}

	oldActive := subscriber.Active
	subscriber.Active = active

	// 更新统计
	sm.stats.mu.Lock()
	if oldActive && !active {
		sm.stats.ActiveSubscribers--
	} else if !oldActive && active {
		sm.stats.ActiveSubscribers++
	}
	sm.stats.mu.Unlock()

	return true
}

// RemoveGroup 移除整个组的订阅者
func (sm *SubscriberManager) RemoveGroup(group string) int {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	subscriberIDs, exists := sm.groupMap[group]
	if !exists {
		return 0
	}

	removedCount := 0
	for _, id := range subscriberIDs {
		if subscriber, exists := sm.subscribers[id]; exists {
			delete(sm.subscribers, id)
			// 更新统计
			sm.stats.mu.Lock()
			if subscriber.Active {
				sm.stats.ActiveSubscribers--
			}
			sm.stats.mu.Unlock()
			removedCount++
		}
	}

	delete(sm.groupMap, group)
	return removedCount
}

// GetStats 获取统计信息
func (sm *SubscriberManager) GetStats() SubscriberStats {
	sm.stats.mu.RLock()
	defer sm.stats.mu.RUnlock()

	// 返回不包含锁的副本以避免copylocks
	return SubscriberStats{
		TotalSubscribers:  sm.stats.TotalSubscribers,
		ActiveSubscribers: sm.stats.ActiveSubscribers,
		EventsSent:        sm.stats.EventsSent,
		ErrorCount:        sm.stats.ErrorCount,
	}
}

// IncrementEventsSent 增加发送事件计数
func (sm *SubscriberManager) IncrementEventsSent() {
	sm.stats.mu.Lock()
	sm.stats.EventsSent++
	sm.stats.mu.Unlock()
}

// IncrementErrorCount 增加错误计数
func (sm *SubscriberManager) IncrementErrorCount() {
	sm.stats.mu.Lock()
	sm.stats.ErrorCount++
	sm.stats.mu.Unlock()
}

// GetAllSubscribers 获取所有订阅者
func (sm *SubscriberManager) GetAllSubscribers() map[string]*SubscriberInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make(map[string]*SubscriberInfo)
	for id, subscriber := range sm.subscribers {
		// 返回副本以避免并发修改
		copy := *subscriber
		result[id] = &copy
	}

	return result
}

// GetActiveSubscriberCount 获取活跃订阅者数量
func (sm *SubscriberManager) GetActiveSubscriberCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	count := 0
	for _, subscriber := range sm.subscribers {
		if subscriber.Active {
			count++
		}
	}

	return count
}

// CleanupInactiveSubscribers 清理非活跃订阅者
func (sm *SubscriberManager) CleanupInactiveSubscribers(maxInactiveTime time.Duration) int {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	removedCount := 0

	for id, subscriber := range sm.subscribers {
		if !subscriber.Active && now.Sub(subscriber.CreatedAt) > maxInactiveTime {
			// 从组映射中移除
			if subscriber.Group != "" {
				if groupIDs, exists := sm.groupMap[subscriber.Group]; exists {
					for i, gid := range groupIDs {
						if gid == id {
							sm.groupMap[subscriber.Group] = append(groupIDs[:i], groupIDs[i+1:]...)
							break
						}
					}
					if len(sm.groupMap[subscriber.Group]) == 0 {
						delete(sm.groupMap, subscriber.Group)
					}
				}
			}

			delete(sm.subscribers, id)
			removedCount++
		}
	}

	return removedCount
}

// ExecuteHandler 安全执行事件处理器
func (sm *SubscriberManager) ExecuteHandler(ctx context.Context, subscriber *SubscriberInfo, event Event) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("handler panic: %v", r)
			sm.IncrementErrorCount()
		}
	}()

	// 检查过滤器
	if subscriber.Filter != nil && !subscriber.Filter(event) {
		return nil
	}

	// 执行处理器
	err = subscriber.Handler(ctx, event)
	if err != nil {
		sm.IncrementErrorCount()
	} else {
		sm.IncrementEventsSent()
	}

	return err
}