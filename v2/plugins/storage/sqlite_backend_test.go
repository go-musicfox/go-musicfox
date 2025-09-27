package storage

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSQLiteBackend_BasicOperations 测试SQLite后端基础操作
func TestSQLiteBackend_BasicOperations(t *testing.T) {
	// 创建临时数据库
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	config := &SQLiteBackendConfig{
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
	}

	backend := NewSQLiteBackend(config)
	require.NotNil(t, backend)

	// 初始化
	err := backend.Initialize()
	require.NoError(t, err)
	defer backend.Close()

	// 测试Set和Get
	t.Run("Set and Get", func(t *testing.T) {
		key := "test_key"
		value := map[string]interface{}{
			"name": "test",
			"age":  25,
		}

		err := backend.Set(key, value, 0)
		assert.NoError(t, err)

		retrieved, err := backend.Get(key)
		assert.NoError(t, err)
		// JSON序列化会改变类型，所以我们检查具体字段
		retrievedMap, ok := retrieved.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "test", retrievedMap["name"])
		assert.Equal(t, float64(25), retrievedMap["age"]) // JSON反序列化数字为float64
	})

	// 测试TTL
	t.Run("TTL Expiration", func(t *testing.T) {
		key := "ttl_key"
		value := "ttl_value"
		ttl := 500 * time.Millisecond // 增加TTL时间

		err := backend.Set(key, value, ttl)
		assert.NoError(t, err)

		// 确认键存在
		exists, err := backend.Exists(key)
		assert.NoError(t, err)
		assert.True(t, exists, "Key should exist after setting")

		// 立即获取应该成功
		retrieved, err := backend.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, value, retrieved)

		// 等待过期
		time.Sleep(600 * time.Millisecond) // 增加等待时间

		// 过期后获取应该失败
		_, err = backend.Get(key)
		assert.Error(t, err)

		// 确认键不再存在
		exists, err = backend.Exists(key)
		assert.NoError(t, err)
		assert.False(t, exists, "Key should not exist after expiration")
	})

	// 测试Delete
	t.Run("Delete", func(t *testing.T) {
		key := "delete_key"
		value := "delete_value"

		err := backend.Set(key, value, 0)
		assert.NoError(t, err)

		err = backend.Delete(key)
		assert.NoError(t, err)

		_, err = backend.Get(key)
		assert.Error(t, err)
	})

	// 测试Exists
	t.Run("Exists", func(t *testing.T) {
		key := "exists_key"
		value := "exists_value"

		// 键不存在
		exists, err := backend.Exists(key)
		assert.NoError(t, err)
		assert.False(t, exists)

		// 设置键
		err = backend.Set(key, value, 0)
		assert.NoError(t, err)

		// 键存在
		exists, err = backend.Exists(key)
		assert.NoError(t, err)
		assert.True(t, exists)
	})
}

// TestSQLiteBackend_BatchOperations 测试批量操作
func TestSQLiteBackend_BatchOperations(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "batch_test.db")

	config := DefaultSQLiteBackendConfig()
	config.DatabasePath = dbPath

	backend := NewSQLiteBackend(config)
	require.NotNil(t, backend)

	err := backend.Initialize()
	require.NoError(t, err)
	defer backend.Close()

	// 测试批量设置和获取
	t.Run("Batch Set and Get", func(t *testing.T) {
		items := map[string]interface{}{
			"key1": "value1",
			"key2": map[string]interface{}{"nested": "value2"},
			"key3": []string{"item1", "item2"},
		}

		err := backend.SetBatch(items, 0)
		assert.NoError(t, err)

		keys := []string{"key1", "key2", "key3"}
		retrieved, err := backend.GetBatch(keys)
		assert.NoError(t, err)
		assert.Equal(t, len(items), len(retrieved))

		// 验证每个键值对，考虑JSON序列化的类型转换
		assert.Equal(t, "value1", retrieved["key1"])
		
		key2Map, ok := retrieved["key2"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "value2", key2Map["nested"])
		
		key3Array, ok := retrieved["key3"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, 2, len(key3Array))
		assert.Equal(t, "item1", key3Array[0])
		assert.Equal(t, "item2", key3Array[1])
	})

	// 测试批量删除
	t.Run("Batch Delete", func(t *testing.T) {
		// 先设置一些键
		items := map[string]interface{}{
			"del1": "value1",
			"del2": "value2",
			"del3": "value3",
		}
		err := backend.SetBatch(items, 0)
		assert.NoError(t, err)

		// 批量删除
		keys := []string{"del1", "del2"}
		err = backend.DeleteBatch(keys)
		assert.NoError(t, err)

		// 验证删除结果
		_, err = backend.Get("del1")
		assert.Error(t, err)
		_, err = backend.Get("del2")
		assert.Error(t, err)

		// del3应该还存在
		value, err := backend.Get("del3")
		assert.NoError(t, err)
		assert.Equal(t, "value3", value)
	})
}

// TestSQLiteBackend_QueryOperations 测试查询操作
func TestSQLiteBackend_QueryOperations(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "query_test.db")

	config := DefaultSQLiteBackendConfig()
	config.DatabasePath = dbPath

	backend := NewSQLiteBackend(config)
	require.NotNil(t, backend)

	err := backend.Initialize()
	require.NoError(t, err)
	defer backend.Close()

	// 设置测试数据
	testData := map[string]interface{}{
		"user:1": map[string]interface{}{"name": "Alice", "age": 25},
		"user:2": map[string]interface{}{"name": "Bob", "age": 30},
		"user:3": map[string]interface{}{"name": "Charlie", "age": 35},
		"config:theme": "dark",
		"config:lang": "en",
	}

	err = backend.SetBatch(testData, 0)
	require.NoError(t, err)

	// 测试Find
	t.Run("Find with pattern", func(t *testing.T) {
		// 查找所有用户
		users, err := backend.Find("user:%", 0)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(users))

		// 查找配置项
		configs, err := backend.Find("config:%", 0)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(configs))

		// 限制结果数量
		limited, err := backend.Find("user:%", 2)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(limited))
	})

	// 测试Count
	t.Run("Count with pattern", func(t *testing.T) {
		userCount, err := backend.Count("user:%")
		assert.NoError(t, err)
		assert.Equal(t, int64(3), userCount)

		configCount, err := backend.Count("config:%")
		assert.NoError(t, err)
		assert.Equal(t, int64(2), configCount)

		allCount, err := backend.Count("%")
		assert.NoError(t, err)
		assert.Equal(t, int64(5), allCount)
	})

	// 测试Keys
	t.Run("Keys with pattern", func(t *testing.T) {
		userKeys, err := backend.Keys("user:%")
		assert.NoError(t, err)
		assert.Equal(t, 3, len(userKeys))
		assert.Contains(t, userKeys, "user:1")
		assert.Contains(t, userKeys, "user:2")
		assert.Contains(t, userKeys, "user:3")

		configKeys, err := backend.Keys("config:%")
		assert.NoError(t, err)
		assert.Equal(t, 2, len(configKeys))
		assert.Contains(t, configKeys, "config:theme")
		assert.Contains(t, configKeys, "config:lang")
	})
}

// TestSQLiteBackend_ConcurrentAccess 测试并发访问
func TestSQLiteBackend_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "concurrent_test.db")

	config := DefaultSQLiteBackendConfig()
	config.DatabasePath = dbPath
	config.MaxConnections = 10

	backend := NewSQLiteBackend(config)
	require.NotNil(t, backend)

	err := backend.Initialize()
	require.NoError(t, err)
	defer backend.Close()

	// 并发写入测试
	t.Run("Concurrent Writes", func(t *testing.T) {
		const numGoroutines = 10
		const numOperations = 100

		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()

				for j := 0; j < numOperations; j++ {
					key := fmt.Sprintf("goroutine_%d_key_%d", id, j)
					value := fmt.Sprintf("value_%d_%d", id, j)

					err := backend.Set(key, value, 0)
					assert.NoError(t, err)
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
				key := fmt.Sprintf("goroutine_%d_key_%d", i, j)
				expectedValue := fmt.Sprintf("value_%d_%d", i, j)

				actualValue, err := backend.Get(key)
				assert.NoError(t, err)
				assert.Equal(t, expectedValue, actualValue)
			}
		}
	})
}

// TestSQLiteBackend_ErrorHandling 测试错误处理
func TestSQLiteBackend_ErrorHandling(t *testing.T) {
	t.Run("Invalid database path", func(t *testing.T) {
		config := &SQLiteBackendConfig{
			DatabasePath: "/invalid/path/test.db",
		}

		backend := NewSQLiteBackend(config)
		err := backend.Initialize()
		assert.Error(t, err)
	})

	t.Run("Operations on closed backend", func(t *testing.T) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "closed_test.db")

		config := DefaultSQLiteBackendConfig()
		config.DatabasePath = dbPath

		backend := NewSQLiteBackend(config)
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

		err = backend.Delete("key")
		assert.Error(t, err)
	})
}

// TestSQLiteBackend_Stats 测试统计信息
func TestSQLiteBackend_Stats(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "stats_test.db")

	config := DefaultSQLiteBackendConfig()
	config.DatabasePath = dbPath

	backend := NewSQLiteBackend(config)
	require.NotNil(t, backend)

	err := backend.Initialize()
	require.NoError(t, err)
	defer backend.Close()

	// 执行一些操作
	err = backend.Set("key1", "value1", 0)
	assert.NoError(t, err)

	err = backend.Set("key2", "value2", 0)
	assert.NoError(t, err)

	_, err = backend.Get("key1")
	assert.NoError(t, err)

	err = backend.Delete("key2")
	assert.NoError(t, err)

	// 获取统计信息
	stats := backend.GetStats()
	assert.Equal(t, int64(1), stats.ReadCount)
	assert.Equal(t, int64(2), stats.WriteCount)
	assert.Equal(t, int64(1), stats.DeleteCount)
	assert.Equal(t, int64(1), stats.KeyCount)
}

// TestSQLiteBackend_CleanupExpiredEntries 测试过期条目清理
func TestSQLiteBackend_CleanupExpiredEntries(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "cleanup_test.db")

	config := DefaultSQLiteBackendConfig()
	config.DatabasePath = dbPath

	backend := NewSQLiteBackend(config)
	require.NotNil(t, backend)

	err := backend.Initialize()
	require.NoError(t, err)
	defer backend.Close()

	// 设置一些带TTL的键
	err = backend.Set("temp1", "value1", 50*time.Millisecond)
	assert.NoError(t, err)

	err = backend.Set("temp2", "value2", 50*time.Millisecond)
	assert.NoError(t, err)

	err = backend.Set("permanent", "value3", 0) // 永久键
	assert.NoError(t, err)

	// 等待过期
	time.Sleep(100 * time.Millisecond)

	// 手动触发清理
	backend.cleanupExpiredEntries()

	// 验证过期键被清理
	_, err = backend.Get("temp1")
	assert.Error(t, err)

	_, err = backend.Get("temp2")
	assert.Error(t, err)

	// 永久键应该还存在
	value, err := backend.Get("permanent")
	assert.NoError(t, err)
	assert.Equal(t, "value3", value)
}

// BenchmarkSQLiteBackend_Set 基准测试Set操作
func BenchmarkSQLiteBackend_Set(b *testing.B) {
	tempDir := b.TempDir()
	dbPath := filepath.Join(tempDir, "bench_set.db")

	config := DefaultSQLiteBackendConfig()
	config.DatabasePath = dbPath

	backend := NewSQLiteBackend(config)
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

// BenchmarkSQLiteBackend_Get 基准测试Get操作
func BenchmarkSQLiteBackend_Get(b *testing.B) {
	tempDir := b.TempDir()
	dbPath := filepath.Join(tempDir, "bench_get.db")

	config := DefaultSQLiteBackendConfig()
	config.DatabasePath = dbPath

	backend := NewSQLiteBackend(config)
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