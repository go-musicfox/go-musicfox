package storage

import (
	"fmt"
	"time"

	core "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
)

// StoragePlugin 存储插件接口
// 基于设计文档中的StoragePlugin接口定义
type StoragePlugin interface {
	core.Plugin

	// 基础存储操作
	Get(key string) (interface{}, error)
	Set(key string, value interface{}, ttl time.Duration) error
	Delete(key string) error
	Exists(key string) (bool, error)

	// 批量操作
	GetBatch(keys []string) (map[string]interface{}, error)
	SetBatch(items map[string]interface{}, ttl time.Duration) error
	DeleteBatch(keys []string) error

	// 查询操作
	Find(pattern string, limit int) (map[string]interface{}, error)
	Count(pattern string) (int64, error)
	Keys(pattern string) ([]string, error)

	// 事务支持
	BeginTransaction() (Transaction, error)

	// 缓存管理
	ClearCache() error
	GetCacheStats() CacheStats
}

// Transaction 事务接口
// 提供ACID事务特性
type Transaction interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}) error
	Delete(key string) error
	Commit() error
	Rollback() error
	GetID() string
	IsActive() bool
}

// CacheStats 缓存统计信息
type CacheStats struct {
	Hits   int64 `json:"hits"`   // 命中次数
	Misses int64 `json:"misses"` // 未命中次数
	Size   int64 `json:"size"`   // 缓存大小
}

// HitRate 计算缓存命中率
func (cs *CacheStats) HitRate() float64 {
	total := cs.Hits + cs.Misses
	if total == 0 {
		return 0.0
	}
	return float64(cs.Hits) / float64(total)
}

// StorageBackend 存储后端接口
type StorageBackend interface {
	// 基础操作
	Get(key string) (interface{}, error)
	Set(key string, value interface{}, ttl time.Duration) error
	Delete(key string) error
	Exists(key string) (bool, error)

	// 批量操作
	GetBatch(keys []string) (map[string]interface{}, error)
	SetBatch(items map[string]interface{}, ttl time.Duration) error
	DeleteBatch(keys []string) error

	// 查询操作
	Find(pattern string, limit int) (map[string]interface{}, error)
	Count(pattern string) (int64, error)
	Keys(pattern string) ([]string, error)

	// 生命周期
	Initialize() error
	Close() error

	// 统计信息
	GetStats() BackendStats
}

// BackendStats 后端统计信息
type BackendStats struct {
	KeyCount     int64 `json:"key_count"`     // 键数量
	ReadCount    int64 `json:"read_count"`    // 读取次数
	WriteCount   int64 `json:"write_count"`   // 写入次数
	DeleteCount  int64 `json:"delete_count"`  // 删除次数
	MemoryUsage  int64 `json:"memory_usage"`  // 内存使用量
}

// StorageConfig 存储配置
type StorageConfig struct {
	Backend         string                `json:"backend"`          // 存储后端类型：memory, file, database, local
	CacheEnabled    bool                  `json:"cache_enabled"`    // 是否启用缓存
	CacheMaxSize    int64                 `json:"cache_max_size"`   // 缓存最大大小
	CacheTTL        time.Duration         `json:"cache_ttl"`        // 缓存TTL
	FileConfig      *FileBackendConfig    `json:"file_config"`      // 文件后端配置
	LocalConfig     *LocalStorageConfig   `json:"local_config"`     // 本地存储后端配置
}

// DefaultStorageConfig 默认存储配置
func DefaultStorageConfig() *StorageConfig {
	return &StorageConfig{
		Backend:      "memory",
		CacheEnabled: true,
		CacheMaxSize: 1000,
		CacheTTL:     5 * time.Minute,
		FileConfig:   DefaultFileBackendConfig(),
		LocalConfig:  DefaultLocalStorageConfig(),
	}
}

// Validate 验证存储配置
func (sc *StorageConfig) Validate() error {
	if sc.Backend == "" {
		return fmt.Errorf("backend cannot be empty")
	}

	if sc.CacheMaxSize < 0 {
		return fmt.Errorf("cache max size cannot be negative")
	}

	return nil
}