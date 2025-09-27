package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDefaultConfig(t *testing.T) {
	config := NewDefaultConfig()
	require.NotNil(t, config)

	// 检查默认值
	assert.Equal(t, "info", config.Get("log.level"))
	assert.Equal(t, "./logs/musicfox.log", config.Get("log.file"))
	assert.Equal(t, ":8080", config.Get("server.port"))
	assert.Equal(t, "localhost", config.Get("server.host"))
	assert.Equal(t, "./plugins", config.Get("plugins.directory"))
	assert.Equal(t, true, config.Get("plugins.auto_load"))
}

func TestConfigSetAndGet(t *testing.T) {
	config := NewDefaultConfig()

	// 测试设置和获取字符串值
	config.Set("test.string", "hello")
	assert.Equal(t, "hello", config.Get("test.string"))

	// 测试设置和获取数字值
	config.Set("test.number", 42)
	assert.Equal(t, 42, config.Get("test.number"))

	// 测试设置和获取布尔值
	config.Set("test.bool", true)
	assert.Equal(t, true, config.Get("test.bool"))

	// 测试获取不存在的键
	assert.Nil(t, config.Get("non.existent"))
}

func TestConfigGetString(t *testing.T) {
	config := NewDefaultConfig()

	// 测试获取存在的字符串值
	config.Set("test.string", "hello")
	value, ok := config.GetString("test.string")
	assert.True(t, ok)
	assert.Equal(t, "hello", value)

	// 测试获取不存在的键
	value, ok = config.GetString("non.existent")
	assert.False(t, ok)
	assert.Empty(t, value)

	// 测试获取非字符串类型的值
	config.Set("test.number", 42)
	value, ok = config.GetString("test.number")
	assert.False(t, ok)
	assert.Empty(t, value)
}

func TestConfigGetInt(t *testing.T) {
	config := NewDefaultConfig()

	// 测试获取存在的整数值
	config.Set("test.int", 42)
	value, ok := config.GetInt("test.int")
	assert.True(t, ok)
	assert.Equal(t, 42, value)

	// 测试获取不存在的键
	value, ok = config.GetInt("non.existent")
	assert.False(t, ok)
	assert.Equal(t, 0, value)

	// 测试获取非整数类型的值
	config.Set("test.string", "hello")
	value, ok = config.GetInt("test.string")
	assert.False(t, ok)
	assert.Equal(t, 0, value)
}

func TestConfigGetBool(t *testing.T) {
	config := NewDefaultConfig()

	// 测试获取存在的布尔值
	config.Set("test.bool", true)
	value, ok := config.GetBool("test.bool")
	assert.True(t, ok)
	assert.True(t, value)

	// 测试获取不存在的键
	value, ok = config.GetBool("non.existent")
	assert.False(t, ok)
	assert.False(t, value)

	// 测试获取非布尔类型的值
	config.Set("test.string", "hello")
	value, ok = config.GetBool("test.string")
	assert.False(t, ok)
	assert.False(t, value)
}

func TestConfigGetFloat64(t *testing.T) {
	config := NewDefaultConfig()

	// 测试获取存在的浮点数值
	config.Set("test.float", 3.14)
	value, ok := config.GetFloat64("test.float")
	assert.True(t, ok)
	assert.Equal(t, 3.14, value)

	// 测试获取不存在的键
	value, ok = config.GetFloat64("non.existent")
	assert.False(t, ok)
	assert.Equal(t, 0.0, value)

	// 测试获取非浮点数类型的值
	config.Set("test.string", "hello")
	value, ok = config.GetFloat64("test.string")
	assert.False(t, ok)
	assert.Equal(t, 0.0, value)
}

func TestConfigGetStringSlice(t *testing.T) {
	config := NewDefaultConfig()

	// 测试获取存在的字符串切片值
	expected := []string{"a", "b", "c"}
	config.Set("test.slice", expected)
	value, ok := config.GetStringSlice("test.slice")
	assert.True(t, ok)
	assert.Equal(t, expected, value)

	// 测试获取不存在的键
	value, ok = config.GetStringSlice("non.existent")
	assert.False(t, ok)
	assert.Nil(t, value)

	// 测试获取非字符串切片类型的值
	config.Set("test.string", "hello")
	value, ok = config.GetStringSlice("test.string")
	assert.False(t, ok)
	assert.Nil(t, value)
}

func TestConfigGetStringWithDefault(t *testing.T) {
	config := NewDefaultConfig()

	// 测试获取存在的值
	config.Set("test.string", "hello")
	value := config.GetStringWithDefault("test.string", "default")
	assert.Equal(t, "hello", value)

	// 测试获取不存在的值，返回默认值
	value = config.GetStringWithDefault("non.existent", "default")
	assert.Equal(t, "default", value)

	// 测试获取非字符串类型的值，返回默认值
	config.Set("test.number", 42)
	value = config.GetStringWithDefault("test.number", "default")
	assert.Equal(t, "default", value)
}

func TestConfigGetIntWithDefault(t *testing.T) {
	config := NewDefaultConfig()

	// 测试获取存在的值
	config.Set("test.int", 42)
	value := config.GetIntWithDefault("test.int", 100)
	assert.Equal(t, 42, value)

	// 测试获取不存在的值，返回默认值
	value = config.GetIntWithDefault("non.existent", 100)
	assert.Equal(t, 100, value)

	// 测试获取非整数类型的值，返回默认值
	config.Set("test.string", "hello")
	value = config.GetIntWithDefault("test.string", 100)
	assert.Equal(t, 100, value)
}

func TestConfigGetBoolWithDefault(t *testing.T) {
	config := NewDefaultConfig()

	// 测试获取存在的值
	config.Set("test.bool", true)
	value := config.GetBoolWithDefault("test.bool", false)
	assert.True(t, value)

	// 测试获取不存在的值，返回默认值
	value = config.GetBoolWithDefault("non.existent", false)
	assert.False(t, value)

	// 测试获取非布尔类型的值，返回默认值
	config.Set("test.string", "hello")
	value = config.GetBoolWithDefault("test.string", false)
	assert.False(t, value)
}

func TestConfigHas(t *testing.T) {
	config := NewDefaultConfig()

	// 测试检查存在的键
	config.Set("test.key", "value")
	assert.True(t, config.Has("test.key"))

	// 测试检查不存在的键
	assert.False(t, config.Has("non.existent"))

	// 测试检查默认配置中的键
	assert.True(t, config.Has("log.level"))
}

func TestConfigDelete(t *testing.T) {
	config := NewDefaultConfig()

	// 设置一个键值对
	config.Set("test.key", "value")
	assert.True(t, config.Has("test.key"))

	// 删除键
	config.Delete("test.key")
	assert.False(t, config.Has("test.key"))

	// 删除不存在的键（应该不会出错）
	config.Delete("non.existent")
}

func TestConfigGetAllKeys(t *testing.T) {
	config := NewDefaultConfig()

	// 添加一些测试键
	config.Set("test.key1", "value1")
	config.Set("test.key2", "value2")
	config.Set("another.key", "value3")

	keys := config.GetAllKeys()

	// 检查是否包含默认键和新添加的键
	assert.Contains(t, keys, "log.level")
	assert.Contains(t, keys, "test.key1")
	assert.Contains(t, keys, "test.key2")
	assert.Contains(t, keys, "another.key")
}

func TestConfigLoadFromFile(t *testing.T) {
	// 创建临时配置文件
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.json")

	// 写入测试配置
	configContent := `{
		"log": {
			"level": "debug",
			"file": "/tmp/test.log"
		},
		"server": {
			"port": ":9090",
			"host": "0.0.0.0"
		},
		"test": {
			"value": "loaded_from_file"
		}
	}`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// 加载配置
	config := NewDefaultConfig()
	err = config.Load(configFile)
	assert.NoError(t, err)

	// 验证配置是否正确加载
	assert.Equal(t, "debug", config.Get("log.level"))
	assert.Equal(t, "/tmp/test.log", config.Get("log.file"))
	assert.Equal(t, ":9090", config.Get("server.port"))
	assert.Equal(t, "0.0.0.0", config.Get("server.host"))
	assert.Equal(t, "loaded_from_file", config.Get("test.value"))
}

func TestConfigLoadFromNonExistentFile(t *testing.T) {
	config := NewDefaultConfig()
	err := config.Load("/non/existent/file.json")
	assert.Error(t, err)
}

func TestConfigLoadFromInvalidJSON(t *testing.T) {
	// 创建临时配置文件
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid_config.json")

	// 写入无效的JSON
	invalidJSON := `{"log": {"level": "debug", "file": "/tmp/test.log"}` // 缺少闭合括号
	err := os.WriteFile(configFile, []byte(invalidJSON), 0644)
	require.NoError(t, err)

	// 尝试加载配置
	config := NewDefaultConfig()
	err = config.Load(configFile)
	assert.Error(t, err)
}

func TestConfigSaveToFile(t *testing.T) {
	config := NewDefaultConfig()

	// 修改一些配置
	config.Set("log.level", "debug")
	config.Set("test.value", "saved_to_file")
	config.Set("test.number", 123)
	config.Set("test.bool", true)

	// 保存到临时文件
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "saved_config.json")

	err := config.Save(configFile)
	assert.NoError(t, err)

	// 验证文件是否存在
	_, err = os.Stat(configFile)
	assert.NoError(t, err)

	// 重新加载配置验证内容
	newConfig := NewDefaultConfig()
	err = newConfig.Load(configFile)
	assert.NoError(t, err)

	assert.Equal(t, "debug", newConfig.Get("log.level"))
	assert.Equal(t, "saved_to_file", newConfig.Get("test.value"))
	assert.Equal(t, float64(123), newConfig.Get("test.number")) // JSON数字会被解析为float64
	assert.Equal(t, true, newConfig.Get("test.bool"))
}

func TestConfigSaveToInvalidPath(t *testing.T) {
	config := NewDefaultConfig()
	err := config.Save("/invalid/path/config.json")
	assert.Error(t, err)
}

func TestConfigReload(t *testing.T) {
	// 创建临时配置文件
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "reload_config.json")

	// 写入初始配置
	initialConfig := `{"test": {"value": "initial"}}`
	err := os.WriteFile(configFile, []byte(initialConfig), 0644)
	require.NoError(t, err)

	// 加载配置
	config := NewDefaultConfig()
	err = config.Load(configFile)
	require.NoError(t, err)
	assert.Equal(t, "initial", config.Get("test.value"))

	// 修改文件内容
	updatedConfig := `{"test": {"value": "updated"}}`
	err = os.WriteFile(configFile, []byte(updatedConfig), 0644)
	require.NoError(t, err)

	// 重新加载配置
	err = config.Reload()
	assert.NoError(t, err)
	assert.Equal(t, "updated", config.Get("test.value"))
}

func TestConfigReloadWithoutFile(t *testing.T) {
	config := NewDefaultConfig()
	// 没有加载过文件，重新加载应该返回错误
	err := config.Reload()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no file loaded")
}

func TestConfigClear(t *testing.T) {
	config := NewDefaultConfig()

	// 添加一些配置
	config.Set("test.key1", "value1")
	config.Set("test.key2", "value2")

	// 验证配置存在
	assert.True(t, config.Has("test.key1"))
	assert.True(t, config.Has("log.level")) // 默认配置

	// 清空配置
	config.Clear()

	// 验证所有配置都被清空
	assert.False(t, config.Has("test.key1"))
	assert.False(t, config.Has("test.key2"))
	assert.False(t, config.Has("log.level"))

	// 验证GetAllKeys返回空切片
	keys := config.GetAllKeys()
	assert.Empty(t, keys)
}

func TestConfigConcurrentAccess(t *testing.T) {
	config := NewDefaultConfig()

	// 并发读写测试
	done := make(chan bool, 20)

	// 启动多个写入goroutine
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("test.writer%d.key%d", id, j)
				value := fmt.Sprintf("value%d-%d", id, j)
				config.Set(key, value)
			}
		}(i)
	}

	// 启动多个读取goroutine
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			for j := 0; j < 100; j++ {
				// 读取操作
				_ = config.Get("log.level")
				_ = config.Has("log.level")
				_ = config.GetAllKeys()
				_, _ = config.GetString("log.level")
			}
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 20; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent access test")
		}
	}
}

// 基准测试
func BenchmarkConfigSet(b *testing.B) {
	config := NewDefaultConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("test.key%d", i)
		config.Set(key, "value")
	}
}

func BenchmarkConfigGet(b *testing.B) {
	config := NewDefaultConfig()
	// 预先设置一些键值对
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("test.key%d", i)
		config.Set(key, "value")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("test.key%d", i%1000)
		_ = config.Get(key)
	}
}

func BenchmarkConfigGetString(b *testing.B) {
	config := NewDefaultConfig()
	// 预先设置一些键值对
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("test.key%d", i)
		config.Set(key, "value")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("test.key%d", i%1000)
		_, _ = config.GetString(key)
	}
}