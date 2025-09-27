package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileBackendConfig 文件后端配置
type FileBackendConfig struct {
	DataDir     string        `json:"data_dir"`     // 数据目录
	SyncMode    bool          `json:"sync_mode"`    // 同步模式
	Compression bool          `json:"compression"`  // 是否压缩
	BackupCount int           `json:"backup_count"` // 备份数量
	FlushInterval time.Duration `json:"flush_interval"` // 刷新间隔
}

// DefaultFileBackendConfig 默认文件后端配置
func DefaultFileBackendConfig() *FileBackendConfig {
	return &FileBackendConfig{
		DataDir:       "./data/storage",
		SyncMode:      false,
		Compression:   false,
		BackupCount:   3,
		FlushInterval: 5 * time.Second,
	}
}

// FileEntry 文件条目
type FileEntry struct {
	Key       string      `json:"key"`        // 键
	Value     interface{} `json:"value"`      // 值
	ExpireAt  *time.Time  `json:"expire_at"`  // 过期时间
	CreatedAt time.Time   `json:"created_at"` // 创建时间
	UpdatedAt time.Time   `json:"updated_at"` // 更新时间
}

// IsExpired 检查是否过期
func (fe *FileEntry) IsExpired() bool {
	if fe.ExpireAt == nil {
		return false
	}
	return time.Now().After(*fe.ExpireAt)
}

// FileBackend 文件存储后端
type FileBackend struct {
	config    *FileBackendConfig // 配置
	data      map[string]*FileEntry // 内存缓存
	mu        sync.RWMutex       // 读写锁
	stats     BackendStats       // 统计信息
	statsMu   sync.RWMutex       // 统计信息锁
	closed    bool               // 是否已关闭
	flushCh   chan struct{}      // 刷新通道
	cleanupCh chan struct{}      // 清理通道
	dirty     bool               // 是否有未保存的更改
}

// NewFileBackend 创建文件存储后端
func NewFileBackend(config *FileBackendConfig) *FileBackend {
	if config == nil {
		config = DefaultFileBackendConfig()
	}

	fb := &FileBackend{
		config:    config,
		data:      make(map[string]*FileEntry),
		stats:     BackendStats{},
		flushCh:   make(chan struct{}),
		cleanupCh: make(chan struct{}),
	}

	return fb
}

// Initialize 初始化后端
func (fb *FileBackend) Initialize() error {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	if fb.closed {
		return fmt.Errorf("backend is closed")
	}

	// 创建数据目录
	if err := os.MkdirAll(fb.config.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// 加载现有数据
	if err := fb.loadData(); err != nil {
		return fmt.Errorf("failed to load data: %w", err)
	}

	// 启动后台任务
	go fb.flushWorker()
	go fb.cleanupWorker()

	return nil
}

// Get 获取值
func (fb *FileBackend) Get(key string) (interface{}, error) {
	fb.mu.RLock()
	defer fb.mu.RUnlock()

	if fb.closed {
		return nil, fmt.Errorf("backend is closed")
	}

	// 更新统计信息
	fb.updateStats(func(stats *BackendStats) {
		stats.ReadCount++
	})

	entry, exists := fb.data[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	// 检查是否过期
	if entry.IsExpired() {
		delete(fb.data, key)
		fb.dirty = true
		return nil, fmt.Errorf("key not found: %s", key)
	}

	return entry.Value, nil
}

// Set 设置值
func (fb *FileBackend) Set(key string, value interface{}, ttl time.Duration) error {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	if fb.closed {
		return fmt.Errorf("backend is closed")
	}

	// 更新统计信息
	fb.updateStats(func(stats *BackendStats) {
		stats.WriteCount++
	})

	now := time.Now()
	entry := &FileEntry{
		Key:       key,
		Value:     value,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 设置过期时间
	if ttl > 0 {
		expireAt := now.Add(ttl)
		entry.ExpireAt = &expireAt
	}

	// 如果是更新现有键，保留创建时间
	if existing, exists := fb.data[key]; exists {
		entry.CreatedAt = existing.CreatedAt
	}

	fb.data[key] = entry
	fb.dirty = true

	// 同步模式下立即刷新
	if fb.config.SyncMode {
		return fb.flushToDisk()
	}

	return nil
}

// Delete 删除值
func (fb *FileBackend) Delete(key string) error {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	if fb.closed {
		return fmt.Errorf("backend is closed")
	}

	// 更新统计信息
	fb.updateStats(func(stats *BackendStats) {
		stats.DeleteCount++
	})

	delete(fb.data, key)
	fb.dirty = true

	// 同步模式下立即刷新
	if fb.config.SyncMode {
		return fb.flushToDisk()
	}

	return nil
}

// Exists 检查键是否存在
func (fb *FileBackend) Exists(key string) (bool, error) {
	fb.mu.RLock()
	defer fb.mu.RUnlock()

	if fb.closed {
		return false, fmt.Errorf("backend is closed")
	}

	entry, exists := fb.data[key]
	if !exists {
		return false, nil
	}

	// 检查是否过期
	if entry.IsExpired() {
		delete(fb.data, key)
		fb.dirty = true
		return false, nil
	}

	return true, nil
}

// GetBatch 批量获取值
func (fb *FileBackend) GetBatch(keys []string) (map[string]interface{}, error) {
	fb.mu.RLock()
	defer fb.mu.RUnlock()

	if fb.closed {
		return nil, fmt.Errorf("backend is closed")
	}

	// 更新统计信息
	fb.updateStats(func(stats *BackendStats) {
		stats.ReadCount += int64(len(keys))
	})

	result := make(map[string]interface{})

	for _, key := range keys {
		entry, exists := fb.data[key]
		if !exists {
			continue
		}

		// 检查是否过期
		if entry.IsExpired() {
			delete(fb.data, key)
			fb.dirty = true
			continue
		}

		result[key] = entry.Value
	}

	return result, nil
}

// SetBatch 批量设置值
func (fb *FileBackend) SetBatch(items map[string]interface{}, ttl time.Duration) error {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	if fb.closed {
		return fmt.Errorf("backend is closed")
	}

	// 更新统计信息
	fb.updateStats(func(stats *BackendStats) {
		stats.WriteCount += int64(len(items))
	})

	now := time.Now()
	var expireAt *time.Time
	if ttl > 0 {
		expire := now.Add(ttl)
		expireAt = &expire
	}

	for key, value := range items {
		entry := &FileEntry{
			Key:       key,
			Value:     value,
			ExpireAt:  expireAt,
			CreatedAt: now,
			UpdatedAt: now,
		}

		// 如果是更新现有键，保留创建时间
		if existing, exists := fb.data[key]; exists {
			entry.CreatedAt = existing.CreatedAt
		}

		fb.data[key] = entry
	}

	fb.dirty = true

	// 同步模式下立即刷新
	if fb.config.SyncMode {
		return fb.flushToDisk()
	}

	return nil
}

// DeleteBatch 批量删除值
func (fb *FileBackend) DeleteBatch(keys []string) error {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	if fb.closed {
		return fmt.Errorf("backend is closed")
	}

	// 更新统计信息
	fb.updateStats(func(stats *BackendStats) {
		stats.DeleteCount += int64(len(keys))
	})

	for _, key := range keys {
		delete(fb.data, key)
	}

	fb.dirty = true

	// 同步模式下立即刷新
	if fb.config.SyncMode {
		return fb.flushToDisk()
	}

	return nil
}

// Find 查找匹配的键值对
func (fb *FileBackend) Find(pattern string, limit int) (map[string]interface{}, error) {
	fb.mu.RLock()
	defer fb.mu.RUnlock()

	if fb.closed {
		return nil, fmt.Errorf("backend is closed")
	}

	result := make(map[string]interface{})
	count := 0

	for key, entry := range fb.data {
		// 检查是否过期
		if entry.IsExpired() {
			delete(fb.data, key)
			fb.dirty = true
			continue
		}

		if matchPattern(key, pattern) {
			result[key] = entry.Value
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}

	return result, nil
}

// Count 统计匹配的键数量
func (fb *FileBackend) Count(pattern string) (int64, error) {
	fb.mu.RLock()
	defer fb.mu.RUnlock()

	if fb.closed {
		return 0, fmt.Errorf("backend is closed")
	}

	count := int64(0)

	for key, entry := range fb.data {
		// 检查是否过期
		if entry.IsExpired() {
			delete(fb.data, key)
			fb.dirty = true
			continue
		}

		if matchPattern(key, pattern) {
			count++
		}
	}

	return count, nil
}

// Keys 获取所有匹配的键
func (fb *FileBackend) Keys(pattern string) ([]string, error) {
	fb.mu.RLock()
	defer fb.mu.RUnlock()

	if fb.closed {
		return nil, fmt.Errorf("backend is closed")
	}

	var keys []string

	for key, entry := range fb.data {
		// 检查是否过期
		if entry.IsExpired() {
			delete(fb.data, key)
			fb.dirty = true
			continue
		}

		if matchPattern(key, pattern) {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// Close 关闭后端
func (fb *FileBackend) Close() error {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	if fb.closed {
		return nil
	}

	fb.closed = true
	close(fb.flushCh)
	close(fb.cleanupCh)

	// 最后一次刷新数据
	if fb.dirty {
		if err := fb.flushToDisk(); err != nil {
			return fmt.Errorf("failed to flush data on close: %w", err)
		}
	}

	// 清空内存数据
	fb.data = nil

	return nil
}

// GetStats 获取统计信息
func (fb *FileBackend) GetStats() BackendStats {
	fb.statsMu.RLock()
	defer fb.statsMu.RUnlock()

	fb.mu.RLock()
	defer fb.mu.RUnlock()

	stats := fb.stats
	stats.KeyCount = int64(len(fb.data))
	stats.MemoryUsage = fb.estimateMemoryUsage()

	return stats
}

// updateStats 更新统计信息
func (fb *FileBackend) updateStats(fn func(*BackendStats)) {
	fb.statsMu.Lock()
	defer fb.statsMu.Unlock()
	fn(&fb.stats)
}

// estimateMemoryUsage 估算内存使用量
func (fb *FileBackend) estimateMemoryUsage() int64 {
	// 简单估算：每个条目大约占用200字节
	return int64(len(fb.data)) * 200
}

// loadData 从磁盘加载数据
func (fb *FileBackend) loadData() error {
	dataFile := filepath.Join(fb.config.DataDir, "data.json")

	// 检查文件是否存在
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		return nil // 文件不存在，跳过加载
	}

	// 读取文件
	data, err := os.ReadFile(dataFile)
	if err != nil {
		return fmt.Errorf("failed to read data file: %w", err)
	}

	// 解析JSON
	var entries map[string]*FileEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	// 过滤过期数据
	for key, entry := range entries {
		if !entry.IsExpired() {
			fb.data[key] = entry
		}
	}

	return nil
}

// flushToDisk 刷新数据到磁盘
func (fb *FileBackend) flushToDisk() error {
	if !fb.dirty {
		return nil
	}

	dataFile := filepath.Join(fb.config.DataDir, "data.json")
	tempFile := dataFile + ".tmp"

	// 序列化数据
	data, err := json.MarshalIndent(fb.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// 写入临时文件
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// 原子性重命名
	if err := os.Rename(tempFile, dataFile); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	fb.dirty = false
	return nil
}

// flushWorker 刷新工作协程
func (fb *FileBackend) flushWorker() {
	ticker := time.NewTicker(fb.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fb.mu.Lock()
			if fb.dirty && !fb.closed {
				fb.flushToDisk()
			}
			fb.mu.Unlock()
		case <-fb.flushCh:
			return
		}
	}
}

// cleanupWorker 清理工作协程
func (fb *FileBackend) cleanupWorker() {
	ticker := time.NewTicker(1 * time.Minute) // 每分钟清理一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fb.mu.Lock()
			if !fb.closed {
				for key, entry := range fb.data {
					if entry.IsExpired() {
						delete(fb.data, key)
						fb.dirty = true
					}
				}
			}
			fb.mu.Unlock()
		case <-fb.cleanupCh:
			return
		}
	}
}