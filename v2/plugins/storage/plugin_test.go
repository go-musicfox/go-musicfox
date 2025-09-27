package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	core "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
)

// mockPluginContext 测试用的插件上下文实现
type mockPluginContext struct{}

func (m *mockPluginContext) GetContext() context.Context {
	return context.Background()
}

func (m *mockPluginContext) GetContainer() core.ServiceRegistry {
	return nil
}

func (m *mockPluginContext) GetEventBus() core.EventBus {
	return nil
}

func (m *mockPluginContext) GetServiceRegistry() core.ServiceRegistry {
	return nil
}

func (m *mockPluginContext) GetLogger() core.Logger {
	return nil
}

func (m *mockPluginContext) GetPluginConfig() core.PluginConfig {
	return nil
}

func (m *mockPluginContext) UpdateConfig(config core.PluginConfig) error {
	return nil
}

func (m *mockPluginContext) GetDataDir() string {
	return "/tmp"
}

func (m *mockPluginContext) GetTempDir() string {
	return "/tmp"
}

func (m *mockPluginContext) SendMessage(topic string, data interface{}) error {
	return nil
}

func (m *mockPluginContext) Subscribe(topic string, handler core.EventHandler) error {
	return nil
}

func (m *mockPluginContext) Unsubscribe(topic string, handler core.EventHandler) error {
	return nil
}

func (m *mockPluginContext) BroadcastMessage(message interface{}) error {
	return nil
}

func (m *mockPluginContext) GetResourceMonitor() *core.ResourceMonitor {
	return nil
}

func (m *mockPluginContext) GetSecurityManager() *core.SecurityManager {
	return nil
}

func (m *mockPluginContext) GetIsolationGroup() *core.IsolationGroup {
	return nil
}

func (m *mockPluginContext) Shutdown() error {
	return nil
}

func TestNewPlugin(t *testing.T) {
	// 测试使用默认配置创建插件
	plugin, err := NewPlugin(nil)
	require.NoError(t, err)
	require.NotNil(t, plugin)

	info := plugin.GetInfo()
	assert.Equal(t, "storage", info.Name)
	assert.Equal(t, "1.0.0", info.Version)
	assert.Equal(t, "Storage plugin for go-musicfox", info.Description)

	// 测试使用自定义配置创建插件
	config := &StorageConfig{
		Backend:      "memory",
		CacheEnabled: true,
		CacheMaxSize: 1000,
		CacheTTL:     5 * time.Minute,
	}

	plugin2, err := NewPlugin(config)
	require.NoError(t, err)
	require.NotNil(t, plugin2)
}

func TestNewPlugin_InvalidConfig(t *testing.T) {
	// 测试无效的后端类型
	config := &StorageConfig{
		Backend: "invalid_backend",
	}

	_, err := NewPlugin(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported backend")
}

func TestPlugin_Lifecycle(t *testing.T) {
	config := DefaultStorageConfig()
	plugin, err := NewPlugin(config)
	require.NoError(t, err)

	// 初始化
	err = plugin.Initialize(&mockPluginContext{})
	require.NoError(t, err)

	// 启动
	err = plugin.Start()
	require.NoError(t, err)

	// 测试基本操作
	err = plugin.Set("test_key", "test_value", 0)
	require.NoError(t, err)

	value, err := plugin.Get("test_key")
	require.NoError(t, err)
	assert.Equal(t, "test_value", value)

	// 停止
	err = plugin.Stop()
	require.NoError(t, err)

	// 清理
	err = plugin.Cleanup()
	require.NoError(t, err)

	// 清理后操作应该失败
	_, err = plugin.Get("test_key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin is closed")
}

func TestPlugin_BasicOperations(t *testing.T) {
	plugin := createTestPlugin(t)
	defer cleanupTestPlugin(t, plugin)

	// 测试Set和Get
	err := plugin.Set("key1", "value1", 0)
	require.NoError(t, err)

	value, err := plugin.Get("key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", value)

	// 测试Exists
	exists, err := plugin.Exists("key1")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = plugin.Exists("nonexistent")
	require.NoError(t, err)
	assert.False(t, exists)

	// 测试Delete
	err = plugin.Delete("key1")
	require.NoError(t, err)

	_, err = plugin.Get("key1")
	assert.Error(t, err)
}

func TestPlugin_BatchOperations(t *testing.T) {
	plugin := createTestPlugin(t)
	defer cleanupTestPlugin(t, plugin)

	// 测试SetBatch
	items := map[string]interface{}{
		"batch1": "value1",
		"batch2": "value2",
		"batch3": "value3",
	}
	err := plugin.SetBatch(items, 0)
	require.NoError(t, err)

	// 测试GetBatch
	keys := []string{"batch1", "batch2", "batch3", "nonexistent"}
	result, err := plugin.GetBatch(keys)
	require.NoError(t, err)

	assert.Equal(t, "value1", result["batch1"])
	assert.Equal(t, "value2", result["batch2"])
	assert.Equal(t, "value3", result["batch3"])
	_, exists := result["nonexistent"]
	assert.False(t, exists)

	// 测试DeleteBatch
	deleteKeys := []string{"batch1", "batch3"}
	err = plugin.DeleteBatch(deleteKeys)
	require.NoError(t, err)

	// 验证删除结果
	_, err = plugin.Get("batch1")
	assert.Error(t, err)

	value, err := plugin.Get("batch2")
	require.NoError(t, err)
	assert.Equal(t, "value2", value)
}

func TestPlugin_QueryOperations(t *testing.T) {
	plugin := createTestPlugin(t)
	defer cleanupTestPlugin(t, plugin)

	// 设置测试数据
	testData := map[string]interface{}{
		"user:1": "alice",
		"user:2": "bob",
		"user:3": "charlie",
		"config:timeout": 30,
		"config:retries": 3,
	}
	err := plugin.SetBatch(testData, 0)
	require.NoError(t, err)

	// 测试Find
	result, err := plugin.Find("user:*", 0)
	require.NoError(t, err)
	assert.Len(t, result, 3)

	// 测试Count
	count, err := plugin.Count("user:*")
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// 测试Keys
	keys, err := plugin.Keys("config:*")
	require.NoError(t, err)
	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "config:timeout")
	assert.Contains(t, keys, "config:retries")
}

func TestPlugin_Transaction(t *testing.T) {
	plugin := createTestPlugin(t)
	defer cleanupTestPlugin(t, plugin)

	// 设置初始数据
	err := plugin.Set("key1", "original_value", 0)
	require.NoError(t, err)

	// 开始事务
	tx, err := plugin.BeginTransaction()
	require.NoError(t, err)
	require.NotNil(t, tx)

	// 在事务中修改数据
	err = tx.Set("key1", "modified_value")
	require.NoError(t, err)

	err = tx.Set("key2", "new_value")
	require.NoError(t, err)

	// 插件中的数据应该还是原来的
	value, err := plugin.Get("key1")
	require.NoError(t, err)
	assert.Equal(t, "original_value", value)

	_, err = plugin.Get("key2")
	assert.Error(t, err)

	// 提交事务
	err = tx.Commit()
	require.NoError(t, err)

	// 现在插件中的数据应该被更新了
	value, err = plugin.Get("key1")
	require.NoError(t, err)
	assert.Equal(t, "modified_value", value)

	value, err = plugin.Get("key2")
	require.NoError(t, err)
	assert.Equal(t, "new_value", value)
}

func TestPlugin_Cache(t *testing.T) {
	// 创建启用缓存的插件
	config := DefaultStorageConfig()
	config.CacheEnabled = true
	config.CacheMaxSize = 10

	plugin, err := NewPlugin(config)
	require.NoError(t, err)

	err = plugin.Initialize(&mockPluginContext{})
	require.NoError(t, err)
	err = plugin.Start()
	require.NoError(t, err)

	defer func() {
		plugin.Stop()
		plugin.Cleanup()
	}()

	// 第一次获取（缓存未命中，从后端获取）
	value, err := plugin.Get("cache_key")
	assert.Error(t, err) // 键不存在

	// 设置一些数据
	err = plugin.Set("cache_key", "cache_value", 0)
	require.NoError(t, err)

	// 第二次获取（从缓存）
	value, err = plugin.Get("cache_key")
	require.NoError(t, err)
	assert.Equal(t, "cache_value", value)

	// 第三次获取（从缓存）
	value, err = plugin.Get("cache_key")
	require.NoError(t, err)
	assert.Equal(t, "cache_value", value)

	// 检查缓存统计
	stats := plugin.GetCacheStats()
	t.Logf("Cache stats: Hits=%d, Misses=%d, Size=%d", stats.Hits, stats.Misses, stats.Size)
	assert.Greater(t, stats.Hits, int64(0))
	assert.Greater(t, stats.Misses, int64(0))
	assert.Equal(t, int64(1), stats.Size)

	// 清空缓存
	err = plugin.ClearCache()
	require.NoError(t, err)

	// 缓存统计应该被重置
	stats = plugin.GetCacheStats()
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(0), stats.Misses)
	assert.Equal(t, int64(0), stats.Size)
}

func TestPlugin_CacheDisabled(t *testing.T) {
	// 创建禁用缓存的插件
	config := DefaultStorageConfig()
	config.CacheEnabled = false

	plugin, err := NewPlugin(config)
	require.NoError(t, err)

	err = plugin.Initialize(&mockPluginContext{})
	require.NoError(t, err)
	err = plugin.Start()
	require.NoError(t, err)

	defer func() {
		plugin.Stop()
		plugin.Cleanup()
	}()

	// 清空缓存应该失败
	err = plugin.ClearCache()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cache is not enabled")
}

func TestPlugin_FileBackend(t *testing.T) {
	// 创建使用文件后端的插件
	config := DefaultStorageConfig()
	config.Backend = "file"
	config.FileConfig = DefaultFileBackendConfig()
	config.FileConfig.DataDir = t.TempDir() // 使用临时目录

	plugin, err := NewPlugin(config)
	require.NoError(t, err)

	err = plugin.Initialize(&mockPluginContext{})
	require.NoError(t, err)
	err = plugin.Start()
	require.NoError(t, err)

	defer func() {
		plugin.Stop()
		plugin.Cleanup()
	}()

	// 测试基本操作
	err = plugin.Set("file_key", "file_value", 0)
	require.NoError(t, err)

	value, err := plugin.Get("file_key")
	require.NoError(t, err)
	assert.Equal(t, "file_value", value)
}

func TestPlugin_ClosedOperations(t *testing.T) {
	plugin := createTestPlugin(t)

	// 关闭插件
	err := plugin.Stop()
	require.NoError(t, err)
	err = plugin.Cleanup()
	require.NoError(t, err)

	// 所有操作都应该失败
	_, err = plugin.Get("key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin is closed")

	err = plugin.Set("key", "value", 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin is closed")

	err = plugin.Delete("key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin is closed")

	_, err = plugin.Exists("key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin is closed")

	_, err = plugin.BeginTransaction()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin is closed")
}

// createTestPlugin 创建测试用的插件实例
func createTestPlugin(t *testing.T) *Plugin {
	config := DefaultStorageConfig()
	plugin, err := NewPlugin(config)
	require.NoError(t, err)

	err = plugin.Initialize(&mockPluginContext{})
	require.NoError(t, err)
	err = plugin.Start()
	require.NoError(t, err)

	return plugin
}

// cleanupTestPlugin 清理测试插件
func cleanupTestPlugin(t *testing.T, plugin *Plugin) {
	err := plugin.Stop()
	assert.NoError(t, err)
	err = plugin.Cleanup()
	assert.NoError(t, err)
}