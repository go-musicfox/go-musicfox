package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	manager := NewManager("/tmp/config", "config.yaml")

	assert.NotNil(t, manager.GetKoanf())
	assert.Equal(t, "/tmp/config", manager.GetConfigDir())
	assert.Equal(t, "config.yaml", manager.GetConfigFile())
}

func TestManager_LoadFromFile(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		filename string
		wantErr  bool
	}{
		{
			name:     "valid yaml config",
			content:  "plugins:\n  - name: test\n    version: 1.0.0",
			filename: "config.yaml",
			wantErr:  false,
		},
		{
			name:     "valid json config",
			content:  `{"plugins":[{"name":"test","version":"1.0.0"}]}`,
			filename: "config.json",
			wantErr:  false,
		},
		{
			name:     "invalid yaml",
			content:  "invalid: yaml: content:",
			filename: "config.yaml",
			wantErr:  true,
		},
		{
			name:     "non-existent file",
			filename: "non-existent.yaml",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager("/tmp/config", "config.yaml")
			
			var configPath string
			if tt.content != "" {
				// 创建临时配置文件
				tmpDir := t.TempDir()
				configPath = filepath.Join(tmpDir, tt.filename)
				err := os.WriteFile(configPath, []byte(tt.content), 0644)
				require.NoError(t, err)
			} else {
				configPath = tt.filename
			}

			err := manager.LoadFromFile(configPath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManager_LoadFromEnv(t *testing.T) {
	// 设置环境变量
	_ = os.Setenv("MUSICFOX_PLUGIN_NAME", "test-plugin")
	_ = os.Setenv("MUSICFOX_PLUGIN_VERSION", "1.0.0")
	defer func() {
		_ = os.Unsetenv("MUSICFOX_PLUGIN_NAME")
		_ = os.Unsetenv("MUSICFOX_PLUGIN_VERSION")
	}()

	manager := NewManager("/tmp/config", "config.yaml")
	err := manager.LoadFromEnv("MUSICFOX_")
	assert.NoError(t, err)

	// 验证环境变量是否正确加载
	name := manager.GetString("plugin.name")
	version := manager.GetString("plugin.version")
	assert.Equal(t, "test-plugin", name)
	assert.Equal(t, "1.0.0", version)
}

func TestManager_GetAndSet(t *testing.T) {
	manager := NewManager("/tmp/config", "config.yaml")

	// 测试Set和Get
	err := manager.Set("test.key", "test-value")
	assert.NoError(t, err)

	value := manager.GetString("test.key")
	assert.Equal(t, "test-value", value)

	// 测试不存在的key
	value = manager.GetString("non.existent")
	assert.Equal(t, "", value)

	// 测试默认值
	val := manager.Get("non.existent")
	if val == nil {
		value = "default"
	} else {
		value = val.(string)
	}
	assert.Equal(t, "default", value)

	// 测试获取不同类型的值
	floatVal := manager.Get("test.float")
	assert.Nil(t, floatVal)
	durationVal := manager.Get("test.duration")
	assert.Nil(t, durationVal)

	sliceVal := manager.Get("test.slice")
	assert.Nil(t, sliceVal)
}

func TestManager_GetTypedValues(t *testing.T) {
	manager := NewManager("/tmp/config", "config.yaml")

	// 设置不同类型的值
	_ = manager.Set("int.value", 42)
	_ = manager.Set("bool.value", true)
	_ = manager.Set("float.value", 3.14)
	_ = manager.Set("duration.value", "5m")
	_ = manager.Set("slice.value", []string{"a", "b", "c"})

	// 测试类型化获取
	assert.Equal(t, 42, manager.GetInt("int.value"))
	assert.Equal(t, true, manager.GetBool("bool.value"))
	
	floatVal := manager.Get("float.value")
	if floatVal != nil {
		assert.Equal(t, 3.14, floatVal.(float64))
	}
	
	durationVal := manager.Get("duration.value")
	if durationVal != nil {
		duration, err := time.ParseDuration(durationVal.(string))
		assert.NoError(t, err)
		assert.Equal(t, 5*time.Minute, duration)
	}

	sliceVal := manager.Get("slice.value")
	if sliceVal != nil {
		assert.Equal(t, []string{"a", "b", "c"}, sliceVal.([]string))
	}
}

func TestManager_SaveToFile(t *testing.T) {
	manager := NewManager("/tmp/config", "config.yaml")
	_ = manager.Set("test.key", "test-value")
	_ = manager.Set("test.number", 42)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err := manager.SaveToFile(configPath)
	assert.NoError(t, err)

	// 验证文件是否存在
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// 加载保存的配置验证内容
	newManager := NewManager("/tmp/config", "config.yaml")
	err = newManager.LoadFromFile(configPath)
	assert.NoError(t, err)

	assert.Equal(t, "test-value", newManager.GetString("test.key"))
	assert.Equal(t, 42, newManager.GetInt("test.number"))
}

func TestManager_Merge(t *testing.T) {
	manager1 := NewManager("/tmp/config", "config.yaml")
	_ = manager1.Set("base.key1", "value1")
	_ = manager1.Set("base.key2", "value2")

	manager2 := NewManager("/tmp/config", "config.yaml")
	_ = manager2.Set("base.key2", "override-value2")
	_ = manager2.Set("base.key3", "value3")

	err := manager1.Merge(manager2.GetKoanf())
	assert.NoError(t, err)

	// 验证合并结果
	assert.Equal(t, "value1", manager1.GetString("base.key1"))
	assert.Equal(t, "override-value2", manager1.GetString("base.key2")) // 应该被覆盖
	assert.Equal(t, "value3", manager1.GetString("base.key3"))
}

func TestManager_Validate(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func(*Manager)
		wantErr    bool
	}{
		{
			name: "valid plugin config",
			setupFunc: func(m *Manager) {
				config := &PluginConfig{
					Name:    "test-plugin",
					Type:    "dynamic_library",
					Path:    "/path/to/plugin.so",
					Enabled: true,
				}
				_ = m.SetPluginConfig("test", config)
			},
			wantErr: false,
		},
		{
			name: "no plugin configs",
			setupFunc: func(m *Manager) {
				// 不设置任何插件配置
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager("/tmp/config", "config.yaml")
			tt.setupFunc(manager)
			
			err := manager.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManager_Keys(t *testing.T) {
	manager := NewManager("/tmp/config", "config.yaml")
	_ = manager.Set("key1", "value1")
	_ = manager.Set("key2", "value2")
	_ = manager.Set("nested.key", "value3")

	// 测试键存在性
	assert.NotNil(t, manager.Get("key1"))
	assert.NotNil(t, manager.Get("key2"))
	assert.NotNil(t, manager.Get("nested.key"))
}

func TestManager_Exists(t *testing.T) {
	manager := NewManager("/tmp/config", "config.yaml")
	_ = manager.Set("existing.key", "value")

	assert.NotNil(t, manager.Get("existing.key"))
	assert.Nil(t, manager.Get("non.existing.key"))
}

func TestManager_Delete(t *testing.T) {
	manager := NewManager("/tmp/config", "config.yaml")
	_ = manager.Set("key.to.delete", "value")

	// 验证key存在
	assert.NotNil(t, manager.Get("key.to.delete"))

	// 删除key（通过设置为nil）
	err := manager.Set("key.to.delete", nil)
	assert.NoError(t, err)

	// 验证key已被删除
	assert.Nil(t, manager.Get("key.to.delete"))
}

func TestManager_CreateDefaultConfig(t *testing.T) {
	manager := NewManager("/tmp/config", "config.yaml")
	assert.NotNil(t, manager)

	// 验证默认配置的基本结构
	assert.NotEmpty(t, manager.GetString("app.name"))
	assert.NotEmpty(t, manager.GetString("app.version"))
	assert.True(t, manager.GetInt("server.port") > 0)
}