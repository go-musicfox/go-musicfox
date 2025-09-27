# Storage Plugin

存储插件为go-musicfox v2提供统一的数据存储接口，支持多种存储后端和事务操作。

## 功能特性

- **多存储后端支持**：内存存储、文件存储、数据库存储
- **事务支持**：提供ACID事务特性
- **批量操作**：支持批量读写操作以提高性能
- **查询功能**：支持模式匹配查询和统计
- **缓存管理**：内置缓存统计和清理功能
- **并发安全**：所有操作都是线程安全的

## 接口说明

### StoragePlugin 接口

```go
type StoragePlugin interface {
    Plugin
    
    // 基础存储操作
    Get(ctx context.Context, key string) (interface{}, error)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) bool
    
    // 批量操作
    GetBatch(ctx context.Context, keys []string) (map[string]interface{}, error)
    SetBatch(ctx context.Context, data map[string]interface{}) error
    DeleteBatch(ctx context.Context, keys []string) error
    
    // 查询操作
    Find(ctx context.Context, pattern string) (map[string]interface{}, error)
    Count(ctx context.Context, pattern string) (int, error)
    Keys(ctx context.Context, pattern string) ([]string, error)
    
    // 事务支持
    BeginTransaction(ctx context.Context) (Transaction, error)
    
    // 缓存管理
    ClearCache(ctx context.Context) error
    GetCacheStats(ctx context.Context) (*CacheStats, error)
}
```

### Transaction 接口

```go
type Transaction interface {
    Get(key string) (interface{}, error)
    Set(key string, value interface{}) error
    Delete(key string) error
    Commit() error
    Rollback() error
}
```

## 存储后端

### 内存存储
- 适用于临时数据和测试环境
- 高性能，但数据不持久化
- 支持TTL过期机制

### 文件存储
- 适用于小到中等规模的数据持久化
- 基于JSON格式存储
- 支持原子写入和备份

### 数据库存储（预留）
- 适用于大规模数据存储
- 支持SQLite、MySQL、PostgreSQL等
- 完整的ACID事务支持

## 使用示例

```go
// 基础操作
err := storage.Set(ctx, "user:123", userData, time.Hour)
value, err := storage.Get(ctx, "user:123")
exists := storage.Exists(ctx, "user:123")
err = storage.Delete(ctx, "user:123")

// 批量操作
data := map[string]interface{}{
    "key1": "value1",
    "key2": "value2",
}
err = storage.SetBatch(ctx, data)
results, err := storage.GetBatch(ctx, []string{"key1", "key2"})

// 事务操作
tx, err := storage.BeginTransaction(ctx)
if err != nil {
    return err
}
defer tx.Rollback()

err = tx.Set("key1", "value1")
if err != nil {
    return err
}

err = tx.Set("key2", "value2")
if err != nil {
    return err
}

return tx.Commit()
```

## 配置选项

```yaml
storage:
  backend: "memory"  # memory, file, database
  file_path: "/tmp/musicfox/storage.json"
  cache_size: 1000
  ttl_cleanup_interval: "5m"
```

## 错误处理

插件使用统一的错误码系统：

- `ErrStorageReadFailed` (5001): 读取操作失败
- `ErrStorageWriteFailed` (5002): 写入操作失败
- `ErrStorageCorrupted` (5003): 存储数据损坏
- `ErrStorageNotFound` (5004): 数据不存在

## 性能特性

- 支持并发读写操作
- 内置连接池和缓存机制
- 批量操作优化
- 异步写入支持
- 内存使用优化