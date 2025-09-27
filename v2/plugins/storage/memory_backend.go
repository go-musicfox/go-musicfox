package storage

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// MemoryBackend 内存存储后端
type MemoryBackend struct {
	data      map[string]interface{} // 数据存储
	expires   map[string]time.Time   // 过期时间
	mu        sync.RWMutex           // 读写锁
	stats     BackendStats           // 统计信息
	statsMu   sync.RWMutex           // 统计信息锁
	closed    bool                   // 是否已关闭
	cleanupCh chan struct{}          // 清理通道
}

// NewMemoryBackend 创建内存存储后端
func NewMemoryBackend() *MemoryBackend {
	mb := &MemoryBackend{
		data:      make(map[string]interface{}),
		expires:   make(map[string]time.Time),
		stats:     BackendStats{},
		cleanupCh: make(chan struct{}),
	}
	
	// 启动过期清理协程
	go mb.cleanupExpired()
	
	return mb
}

// Get 获取值
func (mb *MemoryBackend) Get(key string) (interface{}, error) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	if mb.closed {
		return nil, fmt.Errorf("backend is closed")
	}

	// 更新统计信息
	mb.updateStats(func(stats *BackendStats) {
		stats.ReadCount++
	})

	// 检查是否过期
	if expireTime, exists := mb.expires[key]; exists {
		if time.Now().After(expireTime) {
			// 已过期，删除数据
			delete(mb.data, key)
			delete(mb.expires, key)
			return nil, fmt.Errorf("key not found: %s", key)
		}
	}

	value, exists := mb.data[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	return value, nil
}

// Set 设置值
func (mb *MemoryBackend) Set(key string, value interface{}, ttl time.Duration) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if mb.closed {
		return fmt.Errorf("backend is closed")
	}

	// 更新统计信息
	mb.updateStats(func(stats *BackendStats) {
		stats.WriteCount++
	})

	mb.data[key] = value

	// 设置过期时间
	if ttl > 0 {
		mb.expires[key] = time.Now().Add(ttl)
	} else {
		// 删除过期时间（永不过期）
		delete(mb.expires, key)
	}

	return nil
}

// Delete 删除值
func (mb *MemoryBackend) Delete(key string) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if mb.closed {
		return fmt.Errorf("backend is closed")
	}

	// 更新统计信息
	mb.updateStats(func(stats *BackendStats) {
		stats.DeleteCount++
	})

	delete(mb.data, key)
	delete(mb.expires, key)

	return nil
}

// Exists 检查键是否存在
func (mb *MemoryBackend) Exists(key string) (bool, error) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	if mb.closed {
		return false, fmt.Errorf("backend is closed")
	}

	// 检查是否过期
	if expireTime, exists := mb.expires[key]; exists {
		if time.Now().After(expireTime) {
			// 已过期，删除数据
			delete(mb.data, key)
			delete(mb.expires, key)
			return false, nil
		}
	}

	_, exists := mb.data[key]
	return exists, nil
}

// GetBatch 批量获取值
func (mb *MemoryBackend) GetBatch(keys []string) (map[string]interface{}, error) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	if mb.closed {
		return nil, fmt.Errorf("backend is closed")
	}

	// 更新统计信息
	mb.updateStats(func(stats *BackendStats) {
		stats.ReadCount += int64(len(keys))
	})

	result := make(map[string]interface{})
	now := time.Now()

	for _, key := range keys {
		// 检查是否过期
		if expireTime, exists := mb.expires[key]; exists {
			if now.After(expireTime) {
				// 已过期，删除数据
				delete(mb.data, key)
				delete(mb.expires, key)
				continue
			}
		}

		if value, exists := mb.data[key]; exists {
			result[key] = value
		}
	}

	return result, nil
}

// SetBatch 批量设置值
func (mb *MemoryBackend) SetBatch(items map[string]interface{}, ttl time.Duration) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if mb.closed {
		return fmt.Errorf("backend is closed")
	}

	// 更新统计信息
	mb.updateStats(func(stats *BackendStats) {
		stats.WriteCount += int64(len(items))
	})

	expireTime := time.Time{}
	if ttl > 0 {
		expireTime = time.Now().Add(ttl)
	}

	for key, value := range items {
		mb.data[key] = value
		if ttl > 0 {
			mb.expires[key] = expireTime
		} else {
			delete(mb.expires, key)
		}
	}

	return nil
}

// DeleteBatch 批量删除值
func (mb *MemoryBackend) DeleteBatch(keys []string) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if mb.closed {
		return fmt.Errorf("backend is closed")
	}

	// 更新统计信息
	mb.updateStats(func(stats *BackendStats) {
		stats.DeleteCount += int64(len(keys))
	})

	for _, key := range keys {
		delete(mb.data, key)
		delete(mb.expires, key)
	}

	return nil
}

// Find 查找匹配的键值对
func (mb *MemoryBackend) Find(pattern string, limit int) (map[string]interface{}, error) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	if mb.closed {
		return nil, fmt.Errorf("backend is closed")
	}

	result := make(map[string]interface{})
	count := 0
	now := time.Now()

	for key, value := range mb.data {
		// 检查是否过期
		if expireTime, exists := mb.expires[key]; exists {
			if now.After(expireTime) {
				// 已过期，跳过
				continue
			}
		}

		// 简单的通配符匹配
		if matchPattern(key, pattern) {
			result[key] = value
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}

	return result, nil
}

// Count 统计匹配的键数量
func (mb *MemoryBackend) Count(pattern string) (int64, error) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	if mb.closed {
		return 0, fmt.Errorf("backend is closed")
	}

	count := int64(0)
	now := time.Now()

	for key := range mb.data {
		// 检查是否过期
		if expireTime, exists := mb.expires[key]; exists {
			if now.After(expireTime) {
				// 已过期，跳过
				continue
			}
		}

		if matchPattern(key, pattern) {
			count++
		}
	}

	return count, nil
}

// Keys 获取所有匹配的键
func (mb *MemoryBackend) Keys(pattern string) ([]string, error) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	if mb.closed {
		return nil, fmt.Errorf("backend is closed")
	}

	var keys []string
	now := time.Now()

	for key := range mb.data {
		// 检查是否过期
		if expireTime, exists := mb.expires[key]; exists {
			if now.After(expireTime) {
				// 已过期，跳过
				continue
			}
		}

		if matchPattern(key, pattern) {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// Initialize 初始化后端
func (mb *MemoryBackend) Initialize() error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if mb.closed {
		return fmt.Errorf("backend is closed")
	}

	// 内存后端无需特殊初始化
	return nil
}

// Close 关闭后端
func (mb *MemoryBackend) Close() error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if mb.closed {
		return nil
	}

	mb.closed = true
	close(mb.cleanupCh)

	// 清空数据
	mb.data = nil
	mb.expires = nil

	return nil
}

// GetStats 获取统计信息
func (mb *MemoryBackend) GetStats() BackendStats {
	mb.statsMu.RLock()
	defer mb.statsMu.RUnlock()

	mb.mu.RLock()
	defer mb.mu.RUnlock()

	stats := mb.stats
	stats.KeyCount = int64(len(mb.data))
	stats.MemoryUsage = mb.estimateMemoryUsage()

	return stats
}

// updateStats 更新统计信息
func (mb *MemoryBackend) updateStats(fn func(*BackendStats)) {
	mb.statsMu.Lock()
	defer mb.statsMu.Unlock()
	fn(&mb.stats)
}

// estimateMemoryUsage 估算内存使用量
func (mb *MemoryBackend) estimateMemoryUsage() int64 {
	// 简单估算：每个键值对大约占用100字节
	return int64(len(mb.data)) * 100
}

// cleanupExpired 清理过期数据
func (mb *MemoryBackend) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute) // 每分钟清理一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mb.mu.Lock()
			now := time.Now()
			for key, expireTime := range mb.expires {
				if now.After(expireTime) {
					delete(mb.data, key)
					delete(mb.expires, key)
				}
			}
			mb.mu.Unlock()
		case <-mb.cleanupCh:
			return
		}
	}
}

// matchPattern 简单的通配符匹配
func matchPattern(text, pattern string) bool {
	if pattern == "*" {
		return true
	}

	if strings.Contains(pattern, "*") {
		// 简单的前缀/后缀匹配
		if strings.HasPrefix(pattern, "*") {
			suffix := pattern[1:]
			return strings.HasSuffix(text, suffix)
		}
		if strings.HasSuffix(pattern, "*") {
			prefix := pattern[:len(pattern)-1]
			return strings.HasPrefix(text, prefix)
		}
		// 中间包含通配符的情况
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			return strings.HasPrefix(text, parts[0]) && strings.HasSuffix(text, parts[1])
		}
	}

	return text == pattern
}