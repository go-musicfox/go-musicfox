package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	core "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
)

// 类型别名
type Cache = core.Cache

// MemoryCache 内存缓存实现
type MemoryCache struct {
	data      map[string]*CacheItem
	mu        sync.RWMutex
	maxSize   int
	defaultTTL time.Duration
	stats     *CacheStats
}

// CacheItem 缓存项
type CacheItem struct {
	Value     interface{} `json:"value"`
	ExpiresAt time.Time   `json:"expires_at"`
	CreatedAt time.Time   `json:"created_at"`
	AccessCount int64     `json:"access_count"`
	LastAccess  time.Time `json:"last_access"`
	Size       int64     `json:"size"`
}

// CacheStats 缓存统计
type CacheStats struct {
	Hits        int64 `json:"hits"`
	Misses      int64 `json:"misses"`
	Sets        int64 `json:"sets"`
	Deletes     int64 `json:"deletes"`
	Evictions   int64 `json:"evictions"`
	TotalSize   int64 `json:"total_size"`
	ItemCount   int64 `json:"item_count"`
	mu          sync.RWMutex
}

// NewMemoryCache 创建内存缓存
func NewMemoryCache(maxSize int, defaultTTL time.Duration) *MemoryCache {
	cache := &MemoryCache{
		data:       make(map[string]*CacheItem),
		maxSize:    maxSize,
		defaultTTL: defaultTTL,
		stats:      &CacheStats{},
	}

	// 启动清理goroutine
	go cache.startCleanup()

	return cache
}

// Get 获取缓存值
func (c *MemoryCache) Get(ctx context.Context, key string) (interface{}, error) {
	c.mu.RLock()
	item, exists := c.data[key]
	c.mu.RUnlock()

	if !exists {
		c.stats.recordMiss()
		return nil, fmt.Errorf("key not found")
	}

	// 检查是否过期
	if time.Now().After(item.ExpiresAt) {
		c.mu.Lock()
		delete(c.data, key)
		c.mu.Unlock()
		c.stats.recordMiss()
		c.stats.recordEviction()
		return nil, fmt.Errorf("key expired")
	}

	// 更新访问统计
	c.mu.Lock()
	item.AccessCount++
	item.LastAccess = time.Now()
	c.mu.Unlock()

	c.stats.recordHit()
	return item.Value, nil
}

// Set 设置缓存值
func (c *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.defaultTTL
	}

	// 计算值的大小
	size := c.calculateSize(value)

	item := &CacheItem{
		Value:       value,
		ExpiresAt:   time.Now().Add(ttl),
		CreatedAt:   time.Now(),
		AccessCount: 0,
		LastAccess:  time.Now(),
		Size:        size,
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查是否需要驱逐
	if len(c.data) >= c.maxSize {
		c.evictLRU()
	}

	c.data[key] = item
	c.stats.recordSet()
	c.stats.addSize(size)

	return nil
}

// Delete 删除缓存值
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.data[key]
	if !exists {
		return fmt.Errorf("key not found")
	}

	delete(c.data, key)
	c.stats.recordDelete()
	c.stats.removeSize(item.Size)

	return nil
}

// Exists 检查键是否存在
func (c *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	item, exists := c.data[key]
	c.mu.RUnlock()

	if !exists {
		return false, nil
	}

	// 检查是否过期
	if time.Now().After(item.ExpiresAt) {
		c.mu.Lock()
		delete(c.data, key)
		c.mu.Unlock()
		return false, nil
	}

	return true, nil
}

// Clear 清空缓存
func (c *MemoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*CacheItem)
	c.stats.reset()

	return nil
}

// Keys 获取匹配模式的键
func (c *MemoryCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0)
	for key := range c.data {
		// 简单的模式匹配，支持*通配符
		if c.matchPattern(key, pattern) {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// TTL 获取键的剩余生存时间
func (c *MemoryCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	c.mu.RLock()
	item, exists := c.data[key]
	c.mu.RUnlock()

	if !exists {
		return 0, fmt.Errorf("key not found")
	}

	ttl := time.Until(item.ExpiresAt)
	if ttl < 0 {
		return 0, nil
	}

	return ttl, nil
}

// evictLRU 驱逐最近最少使用的项
func (c *MemoryCache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, item := range c.data {
		if first || item.LastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.LastAccess
			first = false
		}
	}

	if oldestKey != "" {
		item := c.data[oldestKey]
		delete(c.data, oldestKey)
		c.stats.recordEviction()
		c.stats.removeSize(item.Size)
	}
}

// startCleanup 启动清理过期项的goroutine
func (c *MemoryCache) startCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup 清理过期项
func (c *MemoryCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.data {
		if now.After(item.ExpiresAt) {
			delete(c.data, key)
			c.stats.recordEviction()
			c.stats.removeSize(item.Size)
		}
	}
}

// calculateSize 计算值的大小
func (c *MemoryCache) calculateSize(value interface{}) int64 {
	// 简单的大小计算，实际应用中可能需要更精确的计算
	data, err := json.Marshal(value)
	if err != nil {
		return 0
	}
	return int64(len(data))
}

// matchPattern 简单的模式匹配
func (c *MemoryCache) matchPattern(key, pattern string) bool {
	if pattern == "*" {
		return true
	}
	// 这里可以实现更复杂的模式匹配逻辑
	return key == pattern
}

// GetStats 获取缓存统计信息
func (c *MemoryCache) GetStats() *CacheStats {
	c.stats.mu.RLock()
	defer c.stats.mu.RUnlock()

	c.mu.RLock()
	itemCount := int64(len(c.data))
	c.mu.RUnlock()

	return &CacheStats{
		Hits:      c.stats.Hits,
		Misses:    c.stats.Misses,
		Sets:      c.stats.Sets,
		Deletes:   c.stats.Deletes,
		Evictions: c.stats.Evictions,
		TotalSize: c.stats.TotalSize,
		ItemCount: itemCount,
	}
}

// CacheStats 方法
func (s *CacheStats) recordHit() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Hits++
}

func (s *CacheStats) recordMiss() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Misses++
}

func (s *CacheStats) recordSet() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Sets++
	s.ItemCount++
}

func (s *CacheStats) recordDelete() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Deletes++
	s.ItemCount--
}

func (s *CacheStats) recordEviction() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Evictions++
	s.ItemCount--
}

func (s *CacheStats) addSize(size int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalSize += size
}

func (s *CacheStats) removeSize(size int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalSize -= size
}

func (s *CacheStats) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Hits = 0
	s.Misses = 0
	s.Sets = 0
	s.Deletes = 0
	s.Evictions = 0
	s.TotalSize = 0
	s.ItemCount = 0
}

// GetHitRate 获取命中率
func (s *CacheStats) GetHitRate() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := s.Hits + s.Misses
	if total == 0 {
		return 0
	}
	return float64(s.Hits) / float64(total)
}

// MusicCache 音乐缓存实现
type MusicCache struct {
	core.Cache
}

// NewMusicCache 创建音乐缓存
func NewMusicCache(cache core.Cache) *MusicCache {
	return &MusicCache{
		Cache: cache,
	}
}

// GetCacheInstance 获取缓存实例
func GetCacheInstance() core.Cache {
	return NewMemoryCache(1000, 30*time.Minute)
}

// MultiLevelCache 多级缓存
type MultiLevelCache struct {
	l1Cache Cache // 一级缓存（内存）
	l2Cache Cache // 二级缓存（可能是Redis等）
	mu      sync.RWMutex
}

// NewMultiLevelCache 创建多级缓存
func NewMultiLevelCache(l1Cache, l2Cache Cache) *MultiLevelCache {
	return &MultiLevelCache{
		l1Cache: l1Cache,
		l2Cache: l2Cache,
	}
}

// Get 从多级缓存获取值
func (m *MultiLevelCache) Get(ctx context.Context, key string) (interface{}, error) {
	// 先从一级缓存获取
	value, err := m.l1Cache.Get(ctx, key)
	if err == nil {
		return value, nil
	}

	// 从二级缓存获取
	if m.l2Cache != nil {
		value, err = m.l2Cache.Get(ctx, key)
		if err == nil {
			// 回写到一级缓存
			m.l1Cache.Set(ctx, key, value, 10*time.Minute)
			return value, nil
		}
	}

	return nil, fmt.Errorf("key not found in any cache level")
}

// Set 设置多级缓存值
func (m *MultiLevelCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// 设置到一级缓存
	err := m.l1Cache.Set(ctx, key, value, ttl)
	if err != nil {
		return err
	}

	// 设置到二级缓存
	if m.l2Cache != nil {
		return m.l2Cache.Set(ctx, key, value, ttl)
	}

	return nil
}

// Delete 从多级缓存删除值
func (m *MultiLevelCache) Delete(ctx context.Context, key string) error {
	// 从一级缓存删除
	m.l1Cache.Delete(ctx, key)

	// 从二级缓存删除
	if m.l2Cache != nil {
		m.l2Cache.Delete(ctx, key)
	}

	return nil
}

// Exists 检查键是否存在于多级缓存
func (m *MultiLevelCache) Exists(ctx context.Context, key string) (bool, error) {
	// 检查一级缓存
	exists, err := m.l1Cache.Exists(ctx, key)
	if err == nil && exists {
		return true, nil
	}

	// 检查二级缓存
	if m.l2Cache != nil {
		return m.l2Cache.Exists(ctx, key)
	}

	return false, nil
}

// Clear 清空多级缓存
func (m *MultiLevelCache) Clear(ctx context.Context) error {
	m.l1Cache.Clear(ctx)
	if m.l2Cache != nil {
		m.l2Cache.Clear(ctx)
	}
	return nil
}

// Keys 获取多级缓存的键
func (m *MultiLevelCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	// 合并两级缓存的键
	keys1, _ := m.l1Cache.Keys(ctx, pattern)
	keysMap := make(map[string]bool)

	for _, key := range keys1 {
		keysMap[key] = true
	}

	if m.l2Cache != nil {
		keys2, _ := m.l2Cache.Keys(ctx, pattern)
		for _, key := range keys2 {
			keysMap[key] = true
		}
	}

	result := make([]string, 0, len(keysMap))
	for key := range keysMap {
		result = append(result, key)
	}

	return result, nil
}

// TTL 获取多级缓存键的TTL
func (m *MultiLevelCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	// 优先返回一级缓存的TTL
	ttl, err := m.l1Cache.TTL(ctx, key)
	if err == nil {
		return ttl, nil
	}

	// 返回二级缓存的TTL
	if m.l2Cache != nil {
		return m.l2Cache.TTL(ctx, key)
	}

	return 0, fmt.Errorf("key not found")
}

// CacheManager 缓存管理器
type CacheManager struct {
	caches map[string]Cache
	mu     sync.RWMutex
}

// NewCacheManager 创建缓存管理器
func NewCacheManager() *CacheManager {
	return &CacheManager{
		caches: make(map[string]Cache),
	}
}

// RegisterCache 注册缓存
func (cm *CacheManager) RegisterCache(name string, cache Cache) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.caches[name] = cache
}

// GetCache 获取缓存
func (cm *CacheManager) GetCache(name string) (Cache, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	cache, exists := cm.caches[name]
	if !exists {
		return nil, fmt.Errorf("cache %s not found", name)
	}

	return cache, nil
}

// GetDefaultCache 获取默认缓存
func (cm *CacheManager) GetDefaultCache() Cache {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// 返回第一个注册的缓存作为默认缓存
	for _, cache := range cm.caches {
		return cache
	}

	// 如果没有注册的缓存，创建一个默认的内存缓存
	return NewMemoryCache(1000, 30*time.Minute)
}

// ClearAllCaches 清空所有缓存
func (cm *CacheManager) ClearAllCaches(ctx context.Context) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, cache := range cm.caches {
		cache.Clear(ctx)
	}
}