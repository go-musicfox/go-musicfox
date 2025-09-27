package main

import (
	"fmt"
	"log"
	"time"

	storage "github.com/go-musicfox/go-musicfox/v2/plugins/storage"
)

// LocalStorageExample 本地存储插件使用示例
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
			AutoBackup:      false, // 可以启用自动备份
			BackupInterval:  24 * time.Hour,
			MaxBackups:      7,
			CompressionType: "gzip",
			EncryptionKey:   "your-encryption-key-here", // 生产环境中应该从环境变量获取
		},
	}

	// 创建存储插件
	plugin, err := storage.NewPlugin(config)
	if err != nil {
		log.Fatalf("Failed to create storage plugin: %v", err)
	}

	// 初始化插件
	ctx := &storage.MockPluginContext{} // 在实际使用中应该是真实的PluginContext
	if err := plugin.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize plugin: %v", err)
	}

	// 启动插件
	if err := plugin.Start(); err != nil {
		log.Fatalf("Failed to start plugin: %v", err)
	}

	defer func() {
		plugin.Stop()
		plugin.Cleanup()
	}()

	fmt.Println("=== 本地存储插件示例 ===")

	// 1. 基础存储操作
	fmt.Println("\n1. 基础存储操作")
	demoBasicOperations(plugin)

	// 2. 批量操作
	fmt.Println("\n2. 批量操作")
	demoBatchOperations(plugin)

	// 3. 查询操作
	fmt.Println("\n3. 查询操作")
	demoQueryOperations(plugin)

	// 4. 事务操作
	fmt.Println("\n4. 事务操作")
	demoTransactionOperations(plugin)

	// 5. 数据迁移
	fmt.Println("\n5. 数据迁移")
	demoMigrationOperations(plugin)

	// 6. 备份和恢复
	fmt.Println("\n6. 备份和恢复")
	demoBackupOperations(plugin)

	// 7. 高级功能
	fmt.Println("\n7. 高级功能")
	demoAdvancedFeatures(plugin)

	fmt.Println("\n=== 示例完成 ===")
}

// demoBasicOperations 演示基础存储操作
func demoBasicOperations(plugin *storage.Plugin) {
	// 存储用户配置
	userConfig := map[string]interface{}{
		"theme":    "dark",
		"language": "zh-CN",
		"volume":   0.8,
		"autoplay": true,
	}

	err := plugin.Set("user:config", userConfig, 0)
	if err != nil {
		log.Printf("Failed to set user config: %v", err)
		return
	}
	fmt.Println("✓ 用户配置已保存")

	// 读取用户配置
	retrieved, err := plugin.Get("user:config")
	if err != nil {
		log.Printf("Failed to get user config: %v", err)
		return
	}
	fmt.Printf("✓ 用户配置已读取: %+v\n", retrieved)

	// 检查键是否存在
	exists, err := plugin.Exists("user:config")
	if err != nil {
		log.Printf("Failed to check existence: %v", err)
		return
	}
	fmt.Printf("✓ 键存在性检查: %t\n", exists)

	// 存储带TTL的临时数据
	err = plugin.Set("temp:session", "session-123", 5*time.Minute)
	if err != nil {
		log.Printf("Failed to set temp session: %v", err)
		return
	}
	fmt.Println("✓ 临时会话数据已保存（5分钟TTL）")
}

// demoBatchOperations 演示批量操作
func demoBatchOperations(plugin *storage.Plugin) {
	// 批量存储播放列表
	playlists := map[string]interface{}{
		"playlist:1": map[string]interface{}{
			"name":  "我的收藏",
			"count": 25,
			"songs": []string{"song1", "song2", "song3"},
		},
		"playlist:2": map[string]interface{}{
			"name":  "最近播放",
			"count": 10,
			"songs": []string{"song4", "song5"},
		},
		"playlist:3": map[string]interface{}{
			"name":  "每日推荐",
			"count": 30,
			"songs": []string{"song6", "song7", "song8"},
		},
	}

	err := plugin.SetBatch(playlists, 0)
	if err != nil {
		log.Printf("Failed to set batch playlists: %v", err)
		return
	}
	fmt.Printf("✓ 批量保存了 %d 个播放列表\n", len(playlists))

	// 批量读取
	keys := []string{"playlist:1", "playlist:2", "playlist:3"}
	retrieved, err := plugin.GetBatch(keys)
	if err != nil {
		log.Printf("Failed to get batch playlists: %v", err)
		return
	}
	fmt.Printf("✓ 批量读取了 %d 个播放列表\n", len(retrieved))

	// 批量删除部分数据
	deleteKeys := []string{"playlist:3"}
	err = plugin.DeleteBatch(deleteKeys)
	if err != nil {
		log.Printf("Failed to delete batch: %v", err)
		return
	}
	fmt.Printf("✓ 批量删除了 %d 个播放列表\n", len(deleteKeys))
}

// demoQueryOperations 演示查询操作
func demoQueryOperations(plugin *storage.Plugin) {
	// 查找所有播放列表
	playlists, err := plugin.Find("playlist:%", 0)
	if err != nil {
		log.Printf("Failed to find playlists: %v", err)
		return
	}
	fmt.Printf("✓ 找到 %d 个播放列表\n", len(playlists))

	// 统计播放列表数量
	count, err := plugin.Count("playlist:%")
	if err != nil {
		log.Printf("Failed to count playlists: %v", err)
		return
	}
	fmt.Printf("✓ 播放列表总数: %d\n", count)

	// 获取所有播放列表的键
	keys, err := plugin.Keys("playlist:%")
	if err != nil {
		log.Printf("Failed to get playlist keys: %v", err)
		return
	}
	fmt.Printf("✓ 播放列表键: %v\n", keys)
}

// demoTransactionOperations 演示事务操作
func demoTransactionOperations(plugin *storage.Plugin) {
	// 开始事务
	tx, err := plugin.BeginTransaction()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return
	}

	fmt.Printf("✓ 开始事务: %s\n", tx.GetID())

	// 在事务中执行操作
	err = tx.Set("tx:test1", "value1")
	if err != nil {
		log.Printf("Failed to set in transaction: %v", err)
		tx.Rollback()
		return
	}

	err = tx.Set("tx:test2", "value2")
	if err != nil {
		log.Printf("Failed to set in transaction: %v", err)
		tx.Rollback()
		return
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return
	}

	fmt.Println("✓ 事务已提交")

	// 验证事务结果
	value1, err := plugin.Get("tx:test1")
	if err != nil {
		log.Printf("Failed to get tx:test1: %v", err)
		return
	}
	fmt.Printf("✓ 事务结果验证: tx:test1 = %v\n", value1)
}

// demoMigrationOperations 演示数据迁移
func demoMigrationOperations(plugin *storage.Plugin) {
	// 获取本地存储后端（如果是本地存储插件）
	if localBackend, ok := plugin.GetBackend().(*storage.LocalStorageBackend); ok {
		// 获取迁移状态
		status, err := localBackend.GetMigrationStatus()
		if err != nil {
			log.Printf("Failed to get migration status: %v", err)
			return
		}

		fmt.Printf("✓ 当前数据库版本: %d\n", status.CurrentVersion)
		fmt.Printf("✓ 最新可用版本: %d\n", status.LatestVersion)
		fmt.Printf("✓ 待执行迁移数: %d\n", status.PendingCount)
		fmt.Printf("✓ 需要迁移: %t\n", status.NeedsMigration)

		// 获取迁移历史
		mgr := localBackend.GetMigrationManager()
		history, err := mgr.GetMigrationHistory()
		if err != nil {
			log.Printf("Failed to get migration history: %v", err)
			return
		}

		fmt.Printf("✓ 迁移历史 (%d 条记录):\n", len(history))
		for _, record := range history {
			fmt.Printf("  - 版本 %d: %s (应用于 %s)\n",
				record.Version, record.Description, record.AppliedAt.Format("2006-01-02 15:04:05"))
		}
	} else {
		fmt.Println("✓ 当前后端不支持数据迁移功能")
	}
}

// demoBackupOperations 演示备份和恢复
func demoBackupOperations(plugin *storage.Plugin) {
	// 获取本地存储后端（如果是本地存储插件）
	if localBackend, ok := plugin.GetBackend().(*storage.LocalStorageBackend); ok {
		// 创建备份
		backupOptions := &storage.BackupOptions{
			Name:        "demo_backup",
			Description: "演示备份",
			Type:        storage.BackupTypeFull,
			Format:      storage.BackupFormatJSON,
			Compress:    true,
			Encrypt:     false, // 演示中不加密
		}

		backupInfo, err := localBackend.CreateBackup(backupOptions)
		if err != nil {
			log.Printf("Failed to create backup: %v", err)
			return
		}

		fmt.Printf("✓ 备份已创建: %s\n", backupInfo.Name)
		fmt.Printf("  - ID: %d\n", backupInfo.ID)
		fmt.Printf("  - 文件大小: %d 字节\n", backupInfo.FileSize)
		fmt.Printf("  - 条目数量: %d\n", backupInfo.EntryCount)
		fmt.Printf("  - 校验和: %s\n", backupInfo.Checksum)

		// 列出所有备份
		backups, err := localBackend.ListBackups()
		if err != nil {
			log.Printf("Failed to list backups: %v", err)
			return
		}

		fmt.Printf("✓ 备份列表 (%d 个备份):\n", len(backups))
		for _, backup := range backups {
			fmt.Printf("  - %s (ID: %d, 大小: %d 字节, 创建于: %s)\n",
				backup.Name, backup.ID, backup.FileSize,
				backup.CreatedAt.Format("2006-01-02 15:04:05"))
		}
	} else {
		fmt.Println("✓ 当前后端不支持备份功能")
	}
}

// demoAdvancedFeatures 演示高级功能
func demoAdvancedFeatures(plugin *storage.Plugin) {
	// 获取缓存统计
	stats := plugin.GetCacheStats()
	fmt.Printf("✓ 缓存统计:\n")
	fmt.Printf("  - 命中次数: %d\n", stats.Hits)
	fmt.Printf("  - 未命中次数: %d\n", stats.Misses)
	fmt.Printf("  - 命中率: %.2f%%\n", stats.HitRate()*100)
	fmt.Printf("  - 缓存大小: %d\n", stats.Size)

	// 获取本地存储后端的高级功能
	if localBackend, ok := plugin.GetBackend().(*storage.LocalStorageBackend); ok {
		// 获取数据库信息
		dbInfo, err := localBackend.GetDatabaseInfo()
		if err != nil {
			log.Printf("Failed to get database info: %v", err)
			return
		}

		fmt.Printf("✓ 数据库信息:\n")
		fmt.Printf("  - 数据库大小: %d 字节\n", dbInfo.Size)
		fmt.Printf("  - 页面数量: %d\n", dbInfo.PageCount)
		fmt.Printf("  - 页面大小: %d 字节\n", dbInfo.PageSize)
		fmt.Printf("  - 表数量: %d\n", dbInfo.TableCount)
		fmt.Printf("  - 索引数量: %d\n", dbInfo.IndexCount)
		fmt.Printf("  - 模式版本: %d\n", dbInfo.SchemaVersion)

		// 获取后端统计
		backendStats := localBackend.GetStats()
		fmt.Printf("✓ 后端统计:\n")
		fmt.Printf("  - 键数量: %d\n", backendStats.KeyCount)
		fmt.Printf("  - 读取次数: %d\n", backendStats.ReadCount)
		fmt.Printf("  - 写入次数: %d\n", backendStats.WriteCount)
		fmt.Printf("  - 删除次数: %d\n", backendStats.DeleteCount)

		// 执行数据库优化
		fmt.Println("✓ 执行数据库优化...")
		if err := localBackend.Analyze(); err != nil {
			log.Printf("Failed to analyze database: %v", err)
		} else {
			fmt.Println("  - 数据库分析完成")
		}

		if err := localBackend.Compact(); err != nil {
			log.Printf("Failed to compact database: %v", err)
		} else {
			fmt.Println("  - 数据库压缩完成")
		}
	}

	// 清理缓存
	err := plugin.ClearCache()
	if err != nil {
		log.Printf("Failed to clear cache: %v", err)
	} else {
		fmt.Println("✓ 缓存已清理")
	}
}