package storage

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLocalStorageBackend_BasicOperations 测试本地存储后端基础操作
func TestLocalStorageBackend_BasicOperations(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "local_test.db")
	backupDir := filepath.Join(tempDir, "backups")

	config := &LocalStorageConfig{
		SQLiteConfig: &SQLiteBackendConfig{
			DatabasePath:    dbPath,
			MaxConnections:  5,
			ConnectionTTL:   30 * time.Minute,
			WALMode:         true,
			SyncMode:        "NORMAL",
			CacheSize:       1024,
			BusyTimeout:     10 * time.Second,
			JournalMode:     "WAL",
			AutoVacuum:      true,
			ForeignKeys:     true,
		},
		BackupDir:       backupDir,
		AutoMigrate:     true,
		AutoBackup:      false,
		BackupInterval:  24 * time.Hour,
		MaxBackups:      5,
		CompressionType: "gzip",
		EncryptionKey:   "test-key-123",
	}

	backend := NewLocalStorageBackend(config)
	require.NotNil(t, backend)

	// 初始化
	err := backend.Initialize()
	require.NoError(t, err)
	defer backend.Close()

	// 测试基础CRUD操作
	t.Run("CRUD Operations", func(t *testing.T) {
		key := "test_key"
		value := map[string]interface{}{
			"name": "test",
			"age":  25,
			"tags": []string{"tag1", "tag2"},
		}

		// Set
		err := backend.Set(key, value, 0)
		assert.NoError(t, err)

		// Get
		retrieved, err := backend.Get(key)
		assert.NoError(t, err)
		// JSON序列化会改变类型，所以我们检查具体字段
		retrievedMap, ok := retrieved.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "test", retrievedMap["name"])
		assert.Equal(t, float64(25), retrievedMap["age"]) // JSON反序列化数字为float64
		tags, ok := retrievedMap["tags"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, 2, len(tags))
		assert.Equal(t, "tag1", tags[0])
		assert.Equal(t, "tag2", tags[1])

		// Exists
		exists, err := backend.Exists(key)
		assert.NoError(t, err)
		assert.True(t, exists)

		// Delete
		err = backend.Delete(key)
		assert.NoError(t, err)

		// Verify deletion
		exists, err = backend.Exists(key)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	// 测试批量操作
	t.Run("Batch Operations", func(t *testing.T) {
		items := map[string]interface{}{
			"batch1": "value1",
			"batch2": map[string]interface{}{"nested": "value2"},
			"batch3": []int{1, 2, 3},
		}

		// Batch Set
		err := backend.SetBatch(items, 0)
		assert.NoError(t, err)

		// Batch Get
		keys := []string{"batch1", "batch2", "batch3"}
		retrieved, err := backend.GetBatch(keys)
		assert.NoError(t, err)
		assert.Equal(t, len(items), len(retrieved))

		// 验证每个键值对，考虑JSON序列化的类型转换
		assert.Equal(t, "value1", retrieved["batch1"])
		
		batch2, ok := retrieved["batch2"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "value2", batch2["nested"])
		
		batch3, ok := retrieved["batch3"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, 3, len(batch3))
		assert.Equal(t, float64(1), batch3[0]) // JSON反序列化数字为float64
		assert.Equal(t, float64(2), batch3[1])
		assert.Equal(t, float64(3), batch3[2])

		// Batch Delete
		err = backend.DeleteBatch(keys)
		assert.NoError(t, err)

		// Verify deletion
		for _, key := range keys {
			exists, err := backend.Exists(key)
			assert.NoError(t, err)
			assert.False(t, exists)
		}
	})

	// 测试查询操作
	t.Run("Query Operations", func(t *testing.T) {
		// 设置测试数据
		testData := map[string]interface{}{
			"user:alice": map[string]interface{}{"name": "Alice", "role": "admin"},
			"user:bob":   map[string]interface{}{"name": "Bob", "role": "user"},
			"user:charlie": map[string]interface{}{"name": "Charlie", "role": "user"},
			"config:theme": "dark",
			"config:lang":  "en",
		}

		err := backend.SetBatch(testData, 0)
		assert.NoError(t, err)

		// Find
		users, err := backend.Find("user:%", 0)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(users))

		// Count
		userCount, err := backend.Count("user:%")
		assert.NoError(t, err)
		assert.Equal(t, int64(3), userCount)

		// Keys
		userKeys, err := backend.Keys("user:%")
		assert.NoError(t, err)
		assert.Equal(t, 3, len(userKeys))
		assert.Contains(t, userKeys, "user:alice")
		assert.Contains(t, userKeys, "user:bob")
		assert.Contains(t, userKeys, "user:charlie")
	})
}

// TestLocalStorageBackend_Migration 测试数据迁移功能
func TestLocalStorageBackend_Migration(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "migration_test.db")

	config := DefaultLocalStorageConfig()
	config.SQLiteConfig.DatabasePath = dbPath
	config.AutoMigrate = true

	backend := NewLocalStorageBackend(config)
	require.NotNil(t, backend)

	err := backend.Initialize()
	require.NoError(t, err)
	defer backend.Close()

	// 测试获取迁移状态
	t.Run("Migration Status", func(t *testing.T) {
		status, err := backend.GetMigrationStatus()
		assert.NoError(t, err)
		assert.NotNil(t, status)
		assert.GreaterOrEqual(t, status.CurrentVersion, 1)
		assert.GreaterOrEqual(t, status.LatestVersion, status.CurrentVersion)
	})

	// 测试迁移管理器
	t.Run("Migration Manager", func(t *testing.T) {
		mgr := backend.GetMigrationManager()
		assert.NotNil(t, mgr)

		currentVersion, err := mgr.GetCurrentVersion()
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, currentVersion, 1)

		latestVersion := mgr.GetLatestVersion()
		assert.GreaterOrEqual(t, latestVersion, currentVersion)

		history, err := mgr.GetMigrationHistory()
		assert.NoError(t, err)
		assert.NotEmpty(t, history)
	})
}

// TestLocalStorageBackend_Backup 测试备份功能
func TestLocalStorageBackend_Backup(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "backup_test.db")
	backupDir := filepath.Join(tempDir, "backups")

	config := DefaultLocalStorageConfig()
	config.SQLiteConfig.DatabasePath = dbPath
	config.BackupDir = backupDir
	config.EncryptionKey = "test-encryption-key"

	backend := NewLocalStorageBackend(config)
	require.NotNil(t, backend)

	err := backend.Initialize()
	require.NoError(t, err)
	defer backend.Close()

	// 设置一些测试数据
	testData := map[string]interface{}{
		"backup_key1": "backup_value1",
		"backup_key2": map[string]interface{}{"nested": "backup_value2"},
		"backup_key3": []string{"item1", "item2", "item3"},
	}
	err = backend.SetBatch(testData, 0)
	require.NoError(t, err)

	// 测试创建备份
	t.Run("Create Backup", func(t *testing.T) {
		backupOptions := &BackupOptions{
			Name:        "test_backup",
			Description: "Test backup for unit testing",
			Type:        BackupTypeFull,
			Format:      BackupFormatJSON,
			Compress:    true,
			Encrypt:     true,
			Password:    "test-password",
		}

		backupInfo, err := backend.CreateBackup(backupOptions)
		assert.NoError(t, err)
		assert.NotNil(t, backupInfo)
		assert.Equal(t, "test_backup", backupInfo.Name)
		assert.Equal(t, BackupTypeFull, backupInfo.Type)
		assert.True(t, backupInfo.Compressed)
		assert.True(t, backupInfo.Encrypted)
		assert.Equal(t, "completed", backupInfo.Status)
		assert.Greater(t, backupInfo.FileSize, int64(0))
		assert.NotEmpty(t, backupInfo.Checksum)
	})

	// 测试列出备份
	t.Run("List Backups", func(t *testing.T) {
		backups, err := backend.ListBackups()
		assert.NoError(t, err)
		assert.NotEmpty(t, backups)

		// 应该至少有一个备份
		found := false
		for _, backup := range backups {
			if backup.Name == "test_backup" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find the test backup")
	})

	// 测试恢复备份
	t.Run("Restore Backup", func(t *testing.T) {
		// 先删除一些数据
		err := backend.Delete("backup_key1")
		assert.NoError(t, err)

		// 验证数据已删除
		exists, err := backend.Exists("backup_key1")
		assert.NoError(t, err)
		assert.False(t, exists)

		// 获取备份列表
		backups, err := backend.ListBackups()
		assert.NoError(t, err)
		assert.NotEmpty(t, backups)

		// 找到测试备份
		var testBackupID int64
		for _, backup := range backups {
			if backup.Name == "test_backup" {
				testBackupID = backup.ID
				break
			}
		}
		assert.NotZero(t, testBackupID)

		// 恢复备份
		err = backend.RestoreBackup(testBackupID, "test-password")
		assert.NoError(t, err)

		// 验证数据已恢复
		exists, err = backend.Exists("backup_key1")
		assert.NoError(t, err)
		assert.True(t, exists)

		value, err := backend.Get("backup_key1")
		assert.NoError(t, err)
		assert.Equal(t, "backup_value1", value)
	})

	// 测试删除备份
	t.Run("Delete Backup", func(t *testing.T) {
		// 获取备份列表
		backups, err := backend.ListBackups()
		assert.NoError(t, err)
		initialCount := len(backups)

		// 找到测试备份
		var testBackupID int64
		for _, backup := range backups {
			if backup.Name == "test_backup" {
				testBackupID = backup.ID
				break
			}
		}
		assert.NotZero(t, testBackupID)

		// 删除备份
		err = backend.DeleteBackup(testBackupID)
		assert.NoError(t, err)

		// 验证备份已删除
		backups, err = backend.ListBackups()
		assert.NoError(t, err)
		assert.Equal(t, initialCount-1, len(backups))

		// 确认测试备份不再存在
		for _, backup := range backups {
			assert.NotEqual(t, "test_backup", backup.Name)
		}
	})
}

// TestLocalStorageBackend_AdvancedFeatures 测试高级功能
func TestLocalStorageBackend_AdvancedFeatures(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "advanced_test.db")

	config := DefaultLocalStorageConfig()
	config.SQLiteConfig.DatabasePath = dbPath

	backend := NewLocalStorageBackend(config)
	require.NotNil(t, backend)

	err := backend.Initialize()
	require.NoError(t, err)
	defer backend.Close()

	// 设置一些测试数据
	testData := map[string]interface{}{
		"advanced_key1": "advanced_value1",
		"advanced_key2": map[string]interface{}{"complex": "data"},
	}
	err = backend.SetBatch(testData, 0)
	require.NoError(t, err)

	// 测试数据库压缩
	t.Run("Database Compact", func(t *testing.T) {
		err := backend.Compact()
		assert.NoError(t, err)

		// 验证数据仍然存在
		value, err := backend.Get("advanced_key1")
		assert.NoError(t, err)
		assert.Equal(t, "advanced_value1", value)
	})

	// 测试数据库分析
	t.Run("Database Analyze", func(t *testing.T) {
		err := backend.Analyze()
		assert.NoError(t, err)
	})

	// 测试获取数据库信息
	t.Run("Database Info", func(t *testing.T) {
		info, err := backend.GetDatabaseInfo()
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Greater(t, info.Size, int64(0))
		assert.Greater(t, info.PageCount, int64(0))
		assert.Greater(t, info.PageSize, int64(0))
		assert.GreaterOrEqual(t, info.TableCount, int64(1))
		assert.GreaterOrEqual(t, info.SchemaVersion, 1)
	})

	// 测试统计信息
	t.Run("Backend Stats", func(t *testing.T) {
		stats := backend.GetStats()
		assert.GreaterOrEqual(t, stats.KeyCount, int64(2))
		assert.GreaterOrEqual(t, stats.WriteCount, int64(2))
	})
}

// TestLocalStorageBackend_ErrorHandling 测试错误处理
func TestLocalStorageBackend_ErrorHandling(t *testing.T) {
	t.Run("Invalid Configuration", func(t *testing.T) {
		// 无效的数据库路径
		config := &LocalStorageConfig{
			SQLiteConfig: &SQLiteBackendConfig{
				DatabasePath: "/invalid/path/test.db",
			},
			BackupDir:   "/invalid/backup/path",
			AutoMigrate: true,
		}

		backend := NewLocalStorageBackend(config)
		err := backend.Initialize()
		assert.Error(t, err)
	})

	t.Run("Operations on Closed Backend", func(t *testing.T) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "closed_test.db")

		config := DefaultLocalStorageConfig()
		config.SQLiteConfig.DatabasePath = dbPath

		backend := NewLocalStorageBackend(config)
		err := backend.Initialize()
		require.NoError(t, err)

		// 关闭后端
		err = backend.Close()
		assert.NoError(t, err)

		// 尝试操作应该失败
		err = backend.Set("key", "value", 0)
		assert.Error(t, err)

		_, err = backend.Get("key")
		assert.Error(t, err)

		_, err = backend.GetMigrationStatus()
		assert.Error(t, err)

		_, err = backend.ListBackups()
		assert.Error(t, err)
	})
}

// TestLocalStorageBackend_ConcurrentAccess 测试并发访问
func TestLocalStorageBackend_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "concurrent_test.db")

	config := DefaultLocalStorageConfig()
	config.SQLiteConfig.DatabasePath = dbPath
	config.SQLiteConfig.MaxConnections = 20

	backend := NewLocalStorageBackend(config)
	require.NotNil(t, backend)

	err := backend.Initialize()
	require.NoError(t, err)
	defer backend.Close()

	const numGoroutines = 10
	const numOperations = 50

	done := make(chan bool, numGoroutines)

	// 并发读写测试
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("concurrent_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)

				// 写入
				err := backend.Set(key, value, 0)
				assert.NoError(t, err)

				// 读取
				retrieved, err := backend.Get(key)
				assert.NoError(t, err)
				assert.Equal(t, value, retrieved)

				// 检查存在性
				exists, err := backend.Exists(key)
				assert.NoError(t, err)
				assert.True(t, exists)
			}
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// 验证数据完整性
	for i := 0; i < numGoroutines; i++ {
		for j := 0; j < numOperations; j++ {
			key := fmt.Sprintf("concurrent_%d_%d", i, j)
			expectedValue := fmt.Sprintf("value_%d_%d", i, j)

			actualValue, err := backend.Get(key)
			assert.NoError(t, err)
			assert.Equal(t, expectedValue, actualValue)
		}
	}
}

// BenchmarkLocalStorageBackend_Set 基准测试Set操作
func BenchmarkLocalStorageBackend_Set(b *testing.B) {
	tempDir := b.TempDir()
	dbPath := filepath.Join(tempDir, "bench_set.db")

	config := DefaultLocalStorageConfig()
	config.SQLiteConfig.DatabasePath = dbPath

	backend := NewLocalStorageBackend(config)
	backend.Initialize()
	defer backend.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench_key_%d", i)
			value := fmt.Sprintf("bench_value_%d", i)
			backend.Set(key, value, 0)
			i++
		}
	})
}

// BenchmarkLocalStorageBackend_Get 基准测试Get操作
func BenchmarkLocalStorageBackend_Get(b *testing.B) {
	tempDir := b.TempDir()
	dbPath := filepath.Join(tempDir, "bench_get.db")

	config := DefaultLocalStorageConfig()
	config.SQLiteConfig.DatabasePath = dbPath

	backend := NewLocalStorageBackend(config)
	backend.Initialize()
	defer backend.Close()

	// 预设一些数据
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		value := fmt.Sprintf("bench_value_%d", i)
		backend.Set(key, value, 0)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench_key_%d", i%1000)
			backend.Get(key)
			i++
		}
	})
}

// BenchmarkLocalStorageBackend_BatchSet 基准测试批量Set操作
func BenchmarkLocalStorageBackend_BatchSet(b *testing.B) {
	tempDir := b.TempDir()
	dbPath := filepath.Join(tempDir, "bench_batch_set.db")

	config := DefaultLocalStorageConfig()
	config.SQLiteConfig.DatabasePath = dbPath

	backend := NewLocalStorageBackend(config)
	backend.Initialize()
	defer backend.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		items := make(map[string]interface{})
		for j := 0; j < 10; j++ {
			key := fmt.Sprintf("batch_key_%d_%d", i, j)
			value := fmt.Sprintf("batch_value_%d_%d", i, j)
			items[key] = value
		}
		backend.SetBatch(items, 0)
	}
}