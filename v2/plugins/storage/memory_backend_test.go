package storage

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryBackend_BasicOperations(t *testing.T) {
	mb := NewMemoryBackend()
	defer mb.Close()

	// 初始化
	err := mb.Initialize()
	require.NoError(t, err)

	// 测试Set和Get
	err = mb.Set("key1", "value1", 0)
	require.NoError(t, err)

	value, err := mb.Get("key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", value)

	// 测试Exists
	exists, err := mb.Exists("key1")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = mb.Exists("nonexistent")
	require.NoError(t, err)
	assert.False(t, exists)

	// 测试Delete
	err = mb.Delete("key1")
	require.NoError(t, err)

	_, err = mb.Get("key1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key not found")
}

func TestMemoryBackend_TTL(t *testing.T) {
	mb := NewMemoryBackend()
	defer mb.Close()

	err := mb.Initialize()
	require.NoError(t, err)

	// 设置带TTL的值
	err = mb.Set("ttl_key", "ttl_value", 100*time.Millisecond)
	require.NoError(t, err)

	// 立即获取应该成功
	value, err := mb.Get("ttl_key")
	require.NoError(t, err)
	assert.Equal(t, "ttl_value", value)

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	// 过期后获取应该失败
	_, err = mb.Get("ttl_key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key not found")
}

func TestMemoryBackend_BatchOperations(t *testing.T) {
	mb := NewMemoryBackend()
	defer mb.Close()

	err := mb.Initialize()
	require.NoError(t, err)

	// 测试SetBatch
	items := map[string]interface{}{
		"batch1": "value1",
		"batch2": "value2",
		"batch3": "value3",
	}
	err = mb.SetBatch(items, 0)
	require.NoError(t, err)

	// 测试GetBatch
	keys := []string{"batch1", "batch2", "batch3", "nonexistent"}
	result, err := mb.GetBatch(keys)
	require.NoError(t, err)

	assert.Equal(t, "value1", result["batch1"])
	assert.Equal(t, "value2", result["batch2"])
	assert.Equal(t, "value3", result["batch3"])
	_, exists := result["nonexistent"]
	assert.False(t, exists)

	// 测试DeleteBatch
	deleteKeys := []string{"batch1", "batch3"}
	err = mb.DeleteBatch(deleteKeys)
	require.NoError(t, err)

	// 验证删除结果
	_, err = mb.Get("batch1")
	assert.Error(t, err)

	value, err := mb.Get("batch2")
	require.NoError(t, err)
	assert.Equal(t, "value2", value)

	_, err = mb.Get("batch3")
	assert.Error(t, err)
}

func TestMemoryBackend_Find(t *testing.T) {
	mb := NewMemoryBackend()
	defer mb.Close()

	err := mb.Initialize()
	require.NoError(t, err)

	// 设置测试数据
	testData := map[string]interface{}{
		"user:1": "alice",
		"user:2": "bob",
		"user:3": "charlie",
		"config:timeout": 30,
		"config:retries": 3,
	}
	err = mb.SetBatch(testData, 0)
	require.NoError(t, err)

	// 测试Find - 前缀匹配
	result, err := mb.Find("user:*", 0)
	require.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, "alice", result["user:1"])
	assert.Equal(t, "bob", result["user:2"])
	assert.Equal(t, "charlie", result["user:3"])

	// 测试Find - 限制数量
	result, err = mb.Find("user:*", 2)
	require.NoError(t, err)
	assert.Len(t, result, 2)

	// 测试Find - 全匹配
	result, err = mb.Find("*", 0)
	require.NoError(t, err)
	assert.Len(t, result, 5)

	// 测试Count
	count, err := mb.Count("user:*")
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	count, err = mb.Count("config:*")
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// 测试Keys
	keys, err := mb.Keys("user:*")
	require.NoError(t, err)
	assert.Len(t, keys, 3)
	assert.Contains(t, keys, "user:1")
	assert.Contains(t, keys, "user:2")
	assert.Contains(t, keys, "user:3")
}

func TestMemoryBackend_Stats(t *testing.T) {
	mb := NewMemoryBackend()
	defer mb.Close()

	err := mb.Initialize()
	require.NoError(t, err)

	// 初始统计
	stats := mb.GetStats()
	assert.Equal(t, int64(0), stats.KeyCount)
	assert.Equal(t, int64(0), stats.ReadCount)
	assert.Equal(t, int64(0), stats.WriteCount)
	assert.Equal(t, int64(0), stats.DeleteCount)

	// 执行一些操作
	err = mb.Set("key1", "value1", 0)
	require.NoError(t, err)

	err = mb.Set("key2", "value2", 0)
	require.NoError(t, err)

	_, err = mb.Get("key1")
	require.NoError(t, err)

	_, err = mb.Get("key2")
	require.NoError(t, err)

	err = mb.Delete("key1")
	require.NoError(t, err)

	// 检查统计
	stats = mb.GetStats()
	assert.Equal(t, int64(1), stats.KeyCount) // key2还存在
	assert.Equal(t, int64(2), stats.ReadCount)
	assert.Equal(t, int64(2), stats.WriteCount)
	assert.Equal(t, int64(1), stats.DeleteCount)
	assert.Greater(t, stats.MemoryUsage, int64(0))
}

func TestMemoryBackend_Close(t *testing.T) {
	mb := NewMemoryBackend()

	err := mb.Initialize()
	require.NoError(t, err)

	// 设置一些数据
	err = mb.Set("key1", "value1", 0)
	require.NoError(t, err)

	// 关闭后端
	err = mb.Close()
	require.NoError(t, err)

	// 关闭后操作应该失败
	_, err = mb.Get("key1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backend is closed")

	err = mb.Set("key2", "value2", 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backend is closed")

	err = mb.Delete("key1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backend is closed")

	// 重复关闭应该不报错
	err = mb.Close()
	assert.NoError(t, err)
}

func TestMemoryBackend_ConcurrentAccess(t *testing.T) {
	mb := NewMemoryBackend()
	defer mb.Close()

	err := mb.Initialize()
	require.NoError(t, err)

	// 并发写入
	const numGoroutines = 10
	const numOperations = 100

	done := make(chan bool, numGoroutines)

	// 启动多个goroutine进行并发操作
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)

				// 写入
				err := mb.Set(key, value, 0)
				assert.NoError(t, err)

				// 读取
				readValue, err := mb.Get(key)
				assert.NoError(t, err)
				assert.Equal(t, value, readValue)

				// 删除
				err = mb.Delete(key)
				assert.NoError(t, err)
			}
			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// 验证最终状态
	stats := mb.GetStats()
	assert.Equal(t, int64(0), stats.KeyCount) // 所有键都被删除了
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		text    string
		pattern string
		expected bool
	}{
		{"hello", "hello", true},
		{"hello", "world", false},
		{"hello", "*", true},
		{"hello", "h*", true},
		{"hello", "*o", true},
		{"hello", "h*o", true},
		{"hello", "w*", false},
		{"hello", "*w", false},
		{"hello", "h*w", false},
		{"user:123", "user:*", true},
		{"user:123", "*:123", true},
		{"user:123", "user:*23", true},
		{"config:timeout", "config:*", true},
		{"config:timeout", "*:timeout", true},
	}

	for _, test := range tests {
		result := matchPattern(test.text, test.pattern)
		assert.Equal(t, test.expected, result, "matchPattern(%q, %q) = %v, expected %v", test.text, test.pattern, result, test.expected)
	}
}