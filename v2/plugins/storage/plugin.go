package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	core "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
)

// MockPluginContext 测试用的插件上下文实现
type MockPluginContext struct{}

func (m *MockPluginContext) GetContext() context.Context {
	return context.Background()
}

func (m *MockPluginContext) GetContainer() core.ServiceRegistry {
	return nil
}

func (m *MockPluginContext) GetEventBus() core.EventBus {
	return nil
}

func (m *MockPluginContext) GetServiceRegistry() core.ServiceRegistry {
	return nil
}

func (m *MockPluginContext) GetLogger() core.Logger {
	return nil
}

func (m *MockPluginContext) GetPluginConfig() core.PluginConfig {
	return nil
}

func (m *MockPluginContext) UpdateConfig(config core.PluginConfig) error {
	return nil
}

func (m *MockPluginContext) GetDataDir() string {
	return "/tmp"
}

func (m *MockPluginContext) GetTempDir() string {
	return "/tmp"
}

func (m *MockPluginContext) SendMessage(topic string, data interface{}) error {
	return nil
}

func (m *MockPluginContext) Subscribe(topic string, handler core.EventHandler) error {
	return nil
}

func (m *MockPluginContext) Unsubscribe(topic string, handler core.EventHandler) error {
	return nil
}

func (m *MockPluginContext) BroadcastMessage(message interface{}) error {
	return nil
}

func (m *MockPluginContext) GetResourceMonitor() *core.ResourceMonitor {
	return nil
}

func (m *MockPluginContext) GetSecurityManager() *core.SecurityManager {
	return nil
}

func (m *MockPluginContext) GetIsolationGroup() *core.IsolationGroup {
	return nil
}

func (m *MockPluginContext) Shutdown() error {
	return nil
}

// Plugin 存储插件实现
type Plugin struct {
	*core.BasePlugin
	backend     StorageBackend    // 存储后端
	cache       map[string]interface{} // 缓存
	cacheStats  CacheStats        // 缓存统计
	transactions map[string]Transaction // 活跃事务
	mu          sync.RWMutex      // 读写锁
	cacheMu     sync.RWMutex      // 缓存锁
	txMu        sync.RWMutex      // 事务锁
	config      *StorageConfig    // 配置
	closed      bool              // 是否已关闭

	// 控制通道
	cacheCleanupCh chan struct{}
	txCleanupCh    chan struct{}
}

// NewPlugin 创建存储插件
func NewPlugin(config *StorageConfig) (*Plugin, error) {
	if config == nil {
		config = DefaultStorageConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 创建基础插件
	basePlugin := core.NewBasePlugin(&core.PluginInfo{
		ID:          "storage",
		Name:        "storage",
		Version:     "1.0.0",
		Description: "Storage plugin for go-musicfox",
		Author:      "go-musicfox",
		Type:        core.PluginTypeDynamicLibrary,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})

	// 创建存储后端
	var backend StorageBackend

	switch config.Backend {
	case "memory":
		backend = NewMemoryBackend()
	case "file":
		backend = NewFileBackend(config.FileConfig)
	case "local":
		backend = NewLocalStorageBackend(config.LocalConfig)
	default:
		return nil, fmt.Errorf("unsupported backend: %s", config.Backend)
	}

	return &Plugin{
		BasePlugin:     basePlugin,
		backend:        backend,
		cache:          make(map[string]interface{}),
		cacheStats:     CacheStats{},
		transactions:   make(map[string]Transaction),
		config:         config,
		cacheCleanupCh: make(chan struct{}),
		txCleanupCh:    make(chan struct{}),
	}, nil
}

// Initialize 初始化插件
func (p *Plugin) Initialize(ctx core.PluginContext) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("plugin is closed")
	}

	// 初始化存储后端
	if err := p.backend.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize backend: %w", err)
	}

	// 调用基础插件初始化
	return p.BasePlugin.Initialize(ctx)
}

// Start 启动插件
func (p *Plugin) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("plugin is closed")
	}

	// 启动缓存清理协程
	if p.config.CacheEnabled {
		go p.cacheCleanupWorker()
	}

	// 启动事务清理协程
	go p.transactionCleanupWorker()

	// 调用基础插件启动
	return p.BasePlugin.Start()
}

// Stop 停止插件
func (p *Plugin) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	// 停止worker协程（安全关闭通道）
	select {
	case <-p.cacheCleanupCh:
		// 通道已关闭
	default:
		close(p.cacheCleanupCh)
	}

	select {
	case <-p.txCleanupCh:
		// 通道已关闭
	default:
		close(p.txCleanupCh)
	}

	// 回滚所有活跃事务
	p.txMu.Lock()
	for id, tx := range p.transactions {
		if memTx, ok := tx.(*MemoryTransaction); ok {
			if memTx.IsActive() {
				if err := tx.Rollback(); err != nil {
					// 记录错误但继续清理
					fmt.Printf("Failed to rollback transaction %s: %v\n", id, err)
				}
			}
		}
	}
	p.transactions = make(map[string]Transaction)
	p.txMu.Unlock()

	// 调用基础插件停止
	return p.BasePlugin.Stop()
}

// Cleanup 清理插件
func (p *Plugin) Cleanup() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	// 关闭存储后端
	if err := p.backend.Close(); err != nil {
		return fmt.Errorf("failed to close backend: %w", err)
	}

	// 清空缓存
	p.cacheMu.Lock()
	p.cache = nil
	p.cacheMu.Unlock()

	// 调用基础插件清理
	return p.BasePlugin.Cleanup()
}

// GetBackend 获取存储后端
func (p *Plugin) GetBackend() StorageBackend {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.backend
}

// Get 获取值
func (p *Plugin) Get(key string) (interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return nil, fmt.Errorf("plugin is closed")
	}

	// 先从缓存获取
	if p.config.CacheEnabled {
		if value, found := p.getFromCache(key); found {
			return value, nil
		}
	}

	// 从后端获取
	value, err := p.backend.Get(key)
	if err != nil {
		return nil, err
	}

	// 添加到缓存
	if p.config.CacheEnabled {
		p.setToCache(key, value)
	}

	return value, nil
}

// Set 设置值
func (p *Plugin) Set(key string, value interface{}, ttl time.Duration) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return fmt.Errorf("plugin is closed")
	}

	// 设置到后端
	if err := p.backend.Set(key, value, ttl); err != nil {
		return err
	}

	// 更新缓存
	if p.config.CacheEnabled {
		p.setToCache(key, value)
	}

	return nil
}

// Delete 删除值
func (p *Plugin) Delete(key string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return fmt.Errorf("plugin is closed")
	}

	// 从后端删除
	if err := p.backend.Delete(key); err != nil {
		return err
	}

	// 从缓存删除
	if p.config.CacheEnabled {
		p.deleteFromCache(key)
	}

	return nil
}

// Exists 检查键是否存在
func (p *Plugin) Exists(key string) (bool, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return false, fmt.Errorf("plugin is closed")
	}

	// 先检查缓存
	if p.config.CacheEnabled {
		if _, found := p.getFromCache(key); found {
			return true, nil
		}
	}

	// 检查后端
	return p.backend.Exists(key)
}

// GetBatch 批量获取值
func (p *Plugin) GetBatch(keys []string) (map[string]interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return nil, fmt.Errorf("plugin is closed")
	}

	result := make(map[string]interface{})
	missingKeys := make([]string, 0)

	// 先从缓存获取
	if p.config.CacheEnabled {
		for _, key := range keys {
			if value, found := p.getFromCache(key); found {
				result[key] = value
			} else {
				missingKeys = append(missingKeys, key)
			}
		}
	} else {
		missingKeys = keys
	}

	// 从后端获取缺失的键
	if len(missingKeys) > 0 {
		backendResult, err := p.backend.GetBatch(missingKeys)
		if err != nil {
			return nil, err
		}

		// 合并结果并更新缓存
		for key, value := range backendResult {
			result[key] = value
			if p.config.CacheEnabled {
				p.setToCache(key, value)
			}
		}
	}

	return result, nil
}

// SetBatch 批量设置值
func (p *Plugin) SetBatch(items map[string]interface{}, ttl time.Duration) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return fmt.Errorf("plugin is closed")
	}

	// 设置到后端
	if err := p.backend.SetBatch(items, ttl); err != nil {
		return err
	}

	// 更新缓存
	if p.config.CacheEnabled {
		for key, value := range items {
			p.setToCache(key, value)
		}
	}

	return nil
}

// DeleteBatch 批量删除值
func (p *Plugin) DeleteBatch(keys []string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return fmt.Errorf("plugin is closed")
	}

	// 从后端删除
	if err := p.backend.DeleteBatch(keys); err != nil {
		return err
	}

	// 从缓存删除
	if p.config.CacheEnabled {
		for _, key := range keys {
			p.deleteFromCache(key)
		}
	}

	return nil
}

// Find 查找匹配的键值对
func (p *Plugin) Find(pattern string, limit int) (map[string]interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return nil, fmt.Errorf("plugin is closed")
	}

	return p.backend.Find(pattern, limit)
}

// Count 统计匹配的键数量
func (p *Plugin) Count(pattern string) (int64, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return 0, fmt.Errorf("plugin is closed")
	}

	return p.backend.Count(pattern)
}

// Keys 获取所有匹配的键
func (p *Plugin) Keys(pattern string) ([]string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return nil, fmt.Errorf("plugin is closed")
	}

	return p.backend.Keys(pattern)
}

// PluginTransaction 插件事务实现，支持缓存清理
type PluginTransaction struct {
	*MemoryTransaction
	plugin *Plugin // 插件实例，用于缓存清理
}

// NewPluginTransaction 创建插件事务
func NewPluginTransaction(backend StorageBackend, plugin *Plugin) *PluginTransaction {
	return &PluginTransaction{
		MemoryTransaction: NewMemoryTransaction(backend),
		plugin:           plugin,
	}
}

// Commit 提交事务并清理缓存
func (pt *PluginTransaction) Commit() error {
	// 先提交事务
	if err := pt.MemoryTransaction.Commit(); err != nil {
		return err
	}

	// 清理相关的缓存项
	if pt.plugin.config.CacheEnabled {
		for _, entry := range pt.MemoryTransaction.operations {
			pt.plugin.deleteFromCache(entry.Key)
		}
	}

	return nil
}

// BeginTransaction 开始事务
func (p *Plugin) BeginTransaction() (Transaction, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return nil, fmt.Errorf("plugin is closed")
	}

	tx := NewPluginTransaction(p.backend, p)

	p.txMu.Lock()
	p.transactions[tx.GetID()] = tx
	p.txMu.Unlock()

	return tx, nil
}

// ClearCache 清空缓存
func (p *Plugin) ClearCache() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return fmt.Errorf("plugin is closed")
	}

	if !p.config.CacheEnabled {
		return fmt.Errorf("cache is not enabled")
	}

	p.cacheMu.Lock()
	p.cache = make(map[string]interface{})
	p.cacheStats = CacheStats{}
	p.cacheMu.Unlock()

	return nil
}

// GetCacheStats 获取缓存统计
func (p *Plugin) GetCacheStats() CacheStats {
	p.cacheMu.RLock()
	defer p.cacheMu.RUnlock()

	stats := p.cacheStats
	stats.Size = int64(len(p.cache))
	return stats
}

// getFromCache 从缓存获取值
func (p *Plugin) getFromCache(key string) (interface{}, bool) {
	p.cacheMu.Lock()
	defer p.cacheMu.Unlock()

	value, found := p.cache[key]
	if found {
		p.cacheStats.Hits++
	} else {
		p.cacheStats.Misses++
	}

	return value, found
}

// setToCache 设置值到缓存
func (p *Plugin) setToCache(key string, value interface{}) {
	p.cacheMu.Lock()
	defer p.cacheMu.Unlock()

	// 检查缓存大小限制
	if p.config.CacheMaxSize > 0 && int64(len(p.cache)) >= p.config.CacheMaxSize {
		// 简单的LRU：删除第一个元素
		for k := range p.cache {
			delete(p.cache, k)
			break
		}
	}

	p.cache[key] = value
}

// deleteFromCache 从缓存删除值
func (p *Plugin) deleteFromCache(key string) {
	p.cacheMu.Lock()
	defer p.cacheMu.Unlock()

	delete(p.cache, key)
}

// cacheCleanupWorker 缓存清理工作协程
func (p *Plugin) cacheCleanupWorker() {
	ticker := time.NewTicker(p.config.CacheTTL / 2) // 每半个TTL清理一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 检查插件是否已关闭
			p.mu.RLock()
			if p.closed {
				p.mu.RUnlock()
				return
			}
			p.mu.RUnlock()

			// 简单的缓存清理：清空所有缓存
			// 在实际应用中，应该实现更智能的TTL管理
			if p.config.CacheTTL > 0 {
				p.cacheMu.Lock()
				p.cache = make(map[string]interface{})
				p.cacheMu.Unlock()
			}
		case <-p.cacheCleanupCh:
			return
		}
	}
}

// transactionCleanupWorker 事务清理工作协程
func (p *Plugin) transactionCleanupWorker() {
	ticker := time.NewTicker(1 * time.Minute) // 每分钟清理一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 检查插件是否已关闭
			p.mu.RLock()
			if p.closed {
				p.mu.RUnlock()
				return
			}
			p.mu.RUnlock()

			p.txMu.Lock()
			for id, tx := range p.transactions {
				if memTx, ok := tx.(*MemoryTransaction); ok {
					if !memTx.IsActive() {
						delete(p.transactions, id)
					}
				}
			}
			p.txMu.Unlock()
		case <-p.txCleanupCh:
			return
		}
	}
}