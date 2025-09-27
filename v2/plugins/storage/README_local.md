# 本地存储插件 (Local Storage Plugin)

本地存储插件为 go-musicfox v2 提供基于 SQLite 数据库的持久化存储解决方案，支持数据迁移、备份恢复、事务处理等高级功能。

## 功能特性

### 🗄️ 核心存储功能
- **SQLite 数据库**: 基于 SQLite 的可靠数据存储
- **ACID 事务**: 完整的事务支持，确保数据一致性
- **批量操作**: 高效的批量读写操作
- **查询功能**: 支持模式匹配查询和统计
- **TTL 支持**: 自动过期数据清理
- **并发安全**: 线程安全的并发访问

### 🔄 数据迁移
- **版本管理**: 自动数据库版本管理
- **向前兼容**: 支持数据库结构升级
- **向后兼容**: 支持数据库结构降级
- **迁移历史**: 完整的迁移记录追踪
- **自动迁移**: 启动时自动执行待处理迁移

### 💾 备份恢复
- **多种格式**: 支持 JSON、SQL、CSV 格式备份
- **全量备份**: 完整数据备份
- **增量备份**: 基于时间戳的增量备份
- **压缩支持**: GZIP 压缩减少存储空间
- **加密支持**: AES 加密保护敏感数据
- **自动备份**: 定时自动备份功能
- **备份管理**: 自动清理过期备份

### ⚡ 性能优化
- **连接池**: 数据库连接池管理
- **WAL 模式**: 写前日志提升并发性能
- **索引优化**: 自动创建性能索引
- **缓存集成**: 与插件缓存系统集成
- **统计信息**: 详细的性能统计

### 🔒 安全特性
- **数据加密**: 支持数据库文件加密
- **访问控制**: 基于角色的访问控制
- **SQL 注入防护**: 参数化查询防护
- **备份加密**: 备份文件加密保护

## 快速开始

### 1. 基础配置

```go
package main

import (
    "time"
    storage "github.com/go-musicfox/go-musicfox/v2/plugins/storage"
)

func main() {
    // 创建本地存储配置
    config := &storage.StorageConfig{
        Backend:      "local",
        CacheEnabled: true,
        CacheMaxSize: 1000,
        CacheTTL:     5 * time.Minute,
        LocalConfig: &storage.LocalStorageConfig{
            SQLiteConfig: &storage.SQLiteBackendConfig{
                DatabasePath:    "./data/musicfox.db",
                MaxConnections:  10,
                ConnectionTTL:   30 * time.Minute,
                WALMode:         true,
                SyncMode:        "NORMAL",
                CacheSize:       8192, // 8MB
                BusyTimeout:     30 * time.Second,
                JournalMode:     "WAL",
                AutoVacuum:      true,
                ForeignKeys:     true,
            },
            BackupDir:       "./data/backups",
            AutoMigrate:     true,
            AutoBackup:      false,
            BackupInterval:  24 * time.Hour,
            MaxBackups:      7,
            CompressionType: "gzip",
            EncryptionKey:   "your-encryption-key",
        },
    }

    // 创建并初始化插件
    plugin, err := storage.NewPlugin(config)
    if err != nil {
        panic(err)
    }

    ctx := &storage.MockPluginContext{}
    if err := plugin.Initialize(ctx); err != nil {
        panic(err)
    }

    if err := plugin.Start(); err != nil {
        panic(err)
    }

    defer func() {
        plugin.Stop()
        plugin.Cleanup()
    }()

    // 现在可以使用插件了
}
```

### 2. 基础存储操作

```go
// 存储数据
userConfig := map[string]interface{}{
    "theme":    "dark",
    "language": "zh-CN",
    "volume":   0.8,
}

err := plugin.Set("user:config", userConfig, 0)
if err != nil {
    log.Fatal(err)
}

// 读取数据
retrieved, err := plugin.Get("user:config")
if err != nil {
    log.Fatal(err)
}

// 检查存在性
exists, err := plugin.Exists("user:config")
if err != nil {
    log.Fatal(err)
}

// 删除数据
err = plugin.Delete("user:config")
if err != nil {
    log.Fatal(err)
}
```

### 3. 批量操作

```go
// 批量存储
playlists := map[string]interface{}{
    "playlist:1": map[string]interface{}{"name": "我的收藏", "count": 25},
    "playlist:2": map[string]interface{}{"name": "最近播放", "count": 10},
    "playlist:3": map[string]interface{}{"name": "每日推荐", "count": 30},
}

err := plugin.SetBatch(playlists, 0)
if err != nil {
    log.Fatal(err)
}

// 批量读取
keys := []string{"playlist:1", "playlist:2", "playlist:3"}
retrieved, err := plugin.GetBatch(keys)
if err != nil {
    log.Fatal(err)
}

// 批量删除
deleteKeys := []string{"playlist:3"}
err = plugin.DeleteBatch(deleteKeys)
if err != nil {
    log.Fatal(err)
}
```

### 4. 查询操作

```go
// 查找匹配模式的数据
playlists, err := plugin.Find("playlist:%", 10) // 限制返回10条
if err != nil {
    log.Fatal(err)
}

// 统计匹配数量
count, err := plugin.Count("playlist:%")
if err != nil {
    log.Fatal(err)
}

// 获取匹配的键
keys, err := plugin.Keys("playlist:%")
if err != nil {
    log.Fatal(err)
}
```

### 5. 事务操作

```go
// 开始事务
tx, err := plugin.BeginTransaction()
if err != nil {
    log.Fatal(err)
}

// 在事务中执行操作
err = tx.Set("key1", "value1")
if err != nil {
    tx.Rollback()
    log.Fatal(err)
}

err = tx.Set("key2", "value2")
if err != nil {
    tx.Rollback()
    log.Fatal(err)
}

// 提交事务
err = tx.Commit()
if err != nil {
    log.Fatal(err)
}
```

## 高级功能

### 数据迁移

```go
// 获取本地存储后端
localBackend := plugin.GetBackend().(*storage.LocalStorageBackend)

// 获取迁移状态
status, err := localBackend.GetMigrationStatus()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("当前版本: %d, 最新版本: %d\n", status.CurrentVersion, status.LatestVersion)

// 手动迁移到指定版本
err = localBackend.MigrateToVersion(3)
if err != nil {
    log.Fatal(err)
}

// 获取迁移历史
mgr := localBackend.GetMigrationManager()
history, err := mgr.GetMigrationHistory()
if err != nil {
    log.Fatal(err)
}
```

### 备份和恢复

```go
// 创建备份
backupOptions := &storage.BackupOptions{
    Name:        "daily_backup",
    Description: "每日自动备份",
    Type:        storage.BackupTypeFull,
    Format:      storage.BackupFormatJSON,
    Compress:    true,
    Encrypt:     true,
    Password:    "backup-password",
}

backupInfo, err := localBackend.CreateBackup(backupOptions)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("备份已创建: %s (ID: %d)\n", backupInfo.Name, backupInfo.ID)

// 列出所有备份
backups, err := localBackend.ListBackups()
if err != nil {
    log.Fatal(err)
}

// 恢复备份
err = localBackend.RestoreBackup(backupInfo.ID, "backup-password")
if err != nil {
    log.Fatal(err)
}

// 删除备份
err = localBackend.DeleteBackup(backupInfo.ID)
if err != nil {
    log.Fatal(err)
}
```

### 数据库管理

```go
// 获取数据库信息
dbInfo, err := localBackend.GetDatabaseInfo()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("数据库大小: %d 字节\n", dbInfo.Size)
fmt.Printf("表数量: %d\n", dbInfo.TableCount)

// 数据库优化
err = localBackend.Analyze() // 更新统计信息
if err != nil {
    log.Fatal(err)
}

err = localBackend.Compact() // 压缩数据库
if err != nil {
    log.Fatal(err)
}

// 获取性能统计
stats := localBackend.GetStats()
fmt.Printf("读取次数: %d, 写入次数: %d\n", stats.ReadCount, stats.WriteCount)
```

## 配置选项

### SQLite 配置

```go
type SQLiteBackendConfig struct {
    DatabasePath    string        // 数据库文件路径
    MaxConnections  int           // 最大连接数
    ConnectionTTL   time.Duration // 连接TTL
    WALMode         bool          // 是否启用WAL模式
    SyncMode        string        // 同步模式：OFF, NORMAL, FULL
    CacheSize       int           // 缓存大小(KB)
    BusyTimeout     time.Duration // 忙等超时
    JournalMode     string        // 日志模式：DELETE, WAL, MEMORY
    AutoVacuum      bool          // 自动清理
    ForeignKeys     bool          // 外键约束
}
```

### 本地存储配置

```go
type LocalStorageConfig struct {
    SQLiteConfig    *SQLiteBackendConfig // SQLite配置
    BackupDir       string               // 备份目录
    AutoMigrate     bool                 // 自动迁移
    AutoBackup      bool                 // 自动备份
    BackupInterval  time.Duration        // 备份间隔
    MaxBackups      int                  // 最大备份数量
    CompressionType string               // 压缩类型
    EncryptionKey   string               // 加密密钥
}
```

## 性能调优

### 1. 连接池优化

```go
config.LocalConfig.SQLiteConfig.MaxConnections = 20  // 增加连接数
config.LocalConfig.SQLiteConfig.ConnectionTTL = 1 * time.Hour // 延长连接生命周期
```

### 2. 缓存优化

```go
config.LocalConfig.SQLiteConfig.CacheSize = 16384 // 16MB缓存
config.CacheEnabled = true
config.CacheMaxSize = 5000
config.CacheTTL = 10 * time.Minute
```

### 3. WAL 模式

```go
config.LocalConfig.SQLiteConfig.WALMode = true
config.LocalConfig.SQLiteConfig.JournalMode = "WAL"
config.LocalConfig.SQLiteConfig.SyncMode = "NORMAL"
```

### 4. 批量操作

```go
// 使用批量操作而不是单个操作
items := make(map[string]interface{})
for i := 0; i < 1000; i++ {
    items[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
}
plugin.SetBatch(items, 0) // 比1000次单独Set操作快得多
```

## 最佳实践

### 1. 错误处理

```go
// 总是检查错误
if err := plugin.Set("key", "value", 0); err != nil {
    log.Printf("Failed to set key: %v", err)
    // 处理错误
}

// 使用事务确保数据一致性
tx, err := plugin.BeginTransaction()
if err != nil {
    return err
}
defer func() {
    if err != nil {
        tx.Rollback()
    }
}()

// 执行操作...

return tx.Commit()
```

### 2. 资源管理

```go
// 总是正确关闭插件
defer func() {
    if err := plugin.Stop(); err != nil {
        log.Printf("Failed to stop plugin: %v", err)
    }
    if err := plugin.Cleanup(); err != nil {
        log.Printf("Failed to cleanup plugin: %v", err)
    }
}()
```

### 3. 备份策略

```go
// 启用自动备份
config.LocalConfig.AutoBackup = true
config.LocalConfig.BackupInterval = 6 * time.Hour // 每6小时备份
config.LocalConfig.MaxBackups = 14 // 保留14个备份

// 定期创建手动备份
go func() {
    ticker := time.NewTicker(24 * time.Hour)
    defer ticker.Stop()
    
    for range ticker.C {
        backupOptions := &storage.BackupOptions{
            Name:        fmt.Sprintf("manual_%s", time.Now().Format("20060102")),
            Description: "手动每日备份",
            Type:        storage.BackupTypeFull,
            Format:      storage.BackupFormatJSON,
            Compress:    true,
            Encrypt:     true,
            Password:    os.Getenv("BACKUP_PASSWORD"),
        }
        
        if _, err := localBackend.CreateBackup(backupOptions); err != nil {
            log.Printf("Failed to create manual backup: %v", err)
        }
    }
}()
```

### 4. 监控和诊断

```go
// 定期检查性能统计
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        stats := plugin.GetCacheStats()
        if stats.HitRate() < 0.8 { // 命中率低于80%
            log.Printf("Low cache hit rate: %.2f%%", stats.HitRate()*100)
        }
        
        if localBackend, ok := plugin.GetBackend().(*storage.LocalStorageBackend); ok {
            backendStats := localBackend.GetStats()
            log.Printf("Backend stats - Keys: %d, Reads: %d, Writes: %d",
                backendStats.KeyCount, backendStats.ReadCount, backendStats.WriteCount)
        }
    }
}()
```

## 故障排除

### 常见问题

1. **数据库锁定错误**
   ```
   解决方案：增加 BusyTimeout 或启用 WAL 模式
   config.LocalConfig.SQLiteConfig.BusyTimeout = 60 * time.Second
   config.LocalConfig.SQLiteConfig.WALMode = true
   ```

2. **性能问题**
   ```
   解决方案：优化缓存和连接池设置
   config.LocalConfig.SQLiteConfig.CacheSize = 32768 // 32MB
   config.LocalConfig.SQLiteConfig.MaxConnections = 50
   ```

3. **磁盘空间不足**
   ```
   解决方案：启用自动清理和压缩
   config.LocalConfig.SQLiteConfig.AutoVacuum = true
   定期调用 localBackend.Compact()
   ```

4. **备份失败**
   ```
   解决方案：检查备份目录权限和磁盘空间
   确保 BackupDir 目录存在且可写
   检查加密密钥是否正确
   ```

### 调试模式

```go
// 启用详细日志
config.LocalConfig.SQLiteConfig.BusyTimeout = 1 * time.Second // 快速失败以便调试

// 监控数据库操作
if localBackend, ok := plugin.GetBackend().(*storage.LocalStorageBackend); ok {
    go func() {
        ticker := time.NewTicker(10 * time.Second)
        defer ticker.Stop()
        
        for range ticker.C {
            dbInfo, err := localBackend.GetDatabaseInfo()
            if err != nil {
                log.Printf("Debug: Failed to get DB info: %v", err)
                continue
            }
            
            log.Printf("Debug: DB size: %d bytes, Tables: %d, Schema version: %d",
                dbInfo.Size, dbInfo.TableCount, dbInfo.SchemaVersion)
        }
    }()
}
```

## 示例项目

完整的使用示例请参考 `examples/local_storage_example.go` 文件，其中包含了所有功能的详细演示。

## 许可证

本项目采用 MIT 许可证，详情请参阅 LICENSE 文件。