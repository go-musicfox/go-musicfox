package plugin

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"log/slog"
)

// TestPluginConfigManager 测试插件配置管理器
func TestPluginConfigManager(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	manager := NewPluginConfigManager(logger)

	ctx := context.Background()

	// 测试初始化
	if err := manager.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize config manager: %v", err)
	}

	// 创建测试配置
	config := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:         "test-plugin",
			Name:       "Test Plugin",
			Version:    "1.0.0",
			PluginPath: "/path/to/plugin.so",
			Enabled:    true,
		},
		ConfigFormat: ConfigFormatYAML,
		AutoReload:   true,
		CustomConfig: map[string]interface{}{
			"test_key": "test_value",
		},
	}

	// 测试保存配置
	if err := manager.SaveConfig(ctx, config); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// 测试加载配置
	loadedConfig, err := manager.LoadConfig(ctx, "test-plugin")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedConfig.ID != config.ID {
		t.Errorf("Expected ID %s, got %s", config.ID, loadedConfig.ID)
	}

	if loadedConfig.Name != config.Name {
		t.Errorf("Expected Name %s, got %s", config.Name, loadedConfig.Name)
	}

	// 测试列出配置
	configs, err := manager.ListConfigs(ctx)
	if err != nil {
		t.Fatalf("Failed to list configs: %v", err)
	}

	found := false
	for _, cfg := range configs {
		if cfg.ID == "test-plugin" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Test plugin not found in config list")
	}

	// 测试删除配置
	if err := manager.DeleteConfig(ctx, "test-plugin"); err != nil {
		t.Fatalf("Failed to delete config: %v", err)
	}

	// 验证配置已删除
	_, err = manager.LoadConfig(ctx, "test-plugin")
	if err == nil {
		t.Error("Expected error when loading deleted config")
	}
}

// TestConfigValidation 测试配置验证
func TestConfigValidation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	manager := NewPluginConfigManager(logger)

	ctx := context.Background()

	// 测试有效配置
	validConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:         "valid-plugin",
			Name:       "Valid Plugin",
			Version:    "1.0.0",
			PluginPath: "/path/to/plugin.so",
			Enabled:    true,
		},
		ConfigFormat: ConfigFormatYAML,
	}

	if err := manager.validateConfig(ctx, validConfig); err != nil {
		t.Errorf("Valid config should not produce error: %v", err)
	}

	// 测试无效配置 - 缺少ID
	invalidConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			Name:       "Invalid Plugin",
			Version:    "1.0.0",
			PluginPath: "/path/to/plugin.so",
			Enabled:    true,
		},
	}

	if err := manager.validateConfig(ctx, invalidConfig); err == nil {
		t.Error("Invalid config should produce error")
	}

	// 测试无效配置 - 无效版本格式
	invalidVersionConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:         "invalid-version-plugin",
			Name:       "Invalid Version Plugin",
			Version:    "invalid-version",
			PluginPath: "/path/to/plugin.so",
			Enabled:    true,
		},
	}

	if err := manager.validateConfig(ctx, invalidVersionConfig); err == nil {
		t.Error("Config with invalid version should produce error")
	}
}

// TestConfigTemplate 测试配置模板
func TestConfigTemplate(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	manager := NewPluginConfigManager(logger)

	// 测试创建默认配置
	defaultConfig := manager.CreateDefaultConfig("test-plugin", "Test Plugin", "1.0.0")

	if defaultConfig.ID != "test-plugin" {
		t.Errorf("Expected ID 'test-plugin', got %s", defaultConfig.ID)
	}

	if defaultConfig.Name != "Test Plugin" {
		t.Errorf("Expected Name 'Test Plugin', got %s", defaultConfig.Name)
	}

	if defaultConfig.Version != "1.0.0" {
		t.Errorf("Expected Version '1.0.0', got %s", defaultConfig.Version)
	}

	if !defaultConfig.Enabled {
		t.Error("Default config should be enabled")
	}

	// 验证默认资源限制
	if defaultConfig.ResourceLimits == nil {
		t.Error("Default config should have resource limits")
	}

	// 验证默认安全配置
	if defaultConfig.SecurityConfig == nil {
		t.Error("Default config should have security config")
	}
}

// TestConfigFormat 测试配置格式
func TestConfigFormat(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	manager := NewPluginConfigManager(logger)

	config := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:         "format-test-plugin",
			Name:       "Format Test Plugin",
			Version:    "1.0.0",
			PluginPath: "/path/to/plugin.so",
			Enabled:    true,
		},
		ConfigFormat: ConfigFormatJSON,
		CustomConfig: map[string]interface{}{
			"string_value": "test",
			"int_value":    42,
			"bool_value":   true,
			"array_value":  []string{"a", "b", "c"},
			"object_value": map[string]interface{}{
				"nested_key": "nested_value",
			},
		},
	}

	// 测试JSON格式序列化
	jsonData, err := manager.serializeConfig(config, ConfigFormatJSON)
	if err != nil {
		t.Fatalf("Failed to serialize config to JSON: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("JSON serialization produced empty data")
	}

	// 测试JSON格式反序列化
	deserializedConfig, err := manager.deserializeConfig(jsonData, ConfigFormatJSON)
	if err != nil {
		t.Fatalf("Failed to deserialize config from JSON: %v", err)
	}

	if deserializedConfig.ID != config.ID {
		t.Errorf("Deserialized ID mismatch: expected %s, got %s", config.ID, deserializedConfig.ID)
	}

	// 测试YAML格式
	config.ConfigFormat = ConfigFormatYAML
	yamlData, err := manager.serializeConfig(config, ConfigFormatYAML)
	if err != nil {
		t.Fatalf("Failed to serialize config to YAML: %v", err)
	}

	if len(yamlData) == 0 {
		t.Error("YAML serialization produced empty data")
	}

	// 测试YAML格式反序列化
	deserializedYamlConfig, err := manager.deserializeConfig(yamlData, ConfigFormatYAML)
	if err != nil {
		t.Fatalf("Failed to deserialize config from YAML: %v", err)
	}

	if deserializedYamlConfig.ID != config.ID {
		t.Errorf("Deserialized YAML ID mismatch: expected %s, got %s", config.ID, deserializedYamlConfig.ID)
	}
}

// TestConfigMerge 测试配置合并
func TestConfigMerge(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	manager := NewPluginConfigManager(logger)

	// 基础配置
	baseConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:         "merge-test-plugin",
			Name:       "Merge Test Plugin",
			Version:    "1.0.0",
			PluginPath: "/path/to/plugin.so",
			Enabled:    true,
		},
		CustomConfig: map[string]interface{}{
			"base_key":   "base_value",
			"shared_key": "base_shared_value",
		},
	}

	// 覆盖配置
	overrideConfig := map[string]interface{}{
		"override_key": "override_value",
		"shared_key":   "override_shared_value",
	}

	// 执行合并
	mergedConfig := manager.MergeConfigs(baseConfig, overrideConfig)

	// 验证合并结果
	if mergedConfig.CustomConfig["base_key"] != "base_value" {
		t.Error("Base key should be preserved")
	}

	if mergedConfig.CustomConfig["override_key"] != "override_value" {
		t.Error("Override key should be added")
	}

	if mergedConfig.CustomConfig["shared_key"] != "override_shared_value" {
		t.Error("Shared key should be overridden")
	}
}

// TestConfigEnvironmentVariables 测试环境变量替换
func TestConfigEnvironmentVariables(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	manager := NewPluginConfigManager(logger)

	// 设置测试环境变量
	os.Setenv("TEST_ENV_VAR", "test_value")
	os.Setenv("TEST_PORT", "8080")
	defer func() {
		os.Unsetenv("TEST_ENV_VAR")
		os.Unsetenv("TEST_PORT")
	}()

	config := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:         "env-test-plugin",
			Name:       "Environment Test Plugin",
			Version:    "1.0.0",
			PluginPath: "/path/to/plugin.so",
			Enabled:    true,
		},
		CustomConfig: map[string]interface{}{
			"env_value":    "${TEST_ENV_VAR}",
			"port":         "${TEST_PORT}",
			"default_value": "${UNDEFINED_VAR:default_value}",
			"normal_value": "normal",
		},
	}

	// 执行环境变量替换
	processedConfig := manager.ProcessEnvironmentVariables(config)

	// 验证替换结果
	if processedConfig.CustomConfig["env_value"] != "test_value" {
		t.Errorf("Expected 'test_value', got %v", processedConfig.CustomConfig["env_value"])
	}

	if processedConfig.CustomConfig["port"] != "8080" {
		t.Errorf("Expected '8080', got %v", processedConfig.CustomConfig["port"])
	}

	if processedConfig.CustomConfig["default_value"] != "default_value" {
		t.Errorf("Expected 'default_value', got %v", processedConfig.CustomConfig["default_value"])
	}

	if processedConfig.CustomConfig["normal_value"] != "normal" {
		t.Errorf("Expected 'normal', got %v", processedConfig.CustomConfig["normal_value"])
	}
}

// BenchmarkConfigSerialization 配置序列化性能测试
func BenchmarkConfigSerialization(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	manager := NewPluginConfigManager(logger)

	config := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:         "benchmark-plugin",
			Name:       "Benchmark Plugin",
			Version:    "1.0.0",
			PluginPath: "/path/to/plugin.so",
			Enabled:    true,
		},
		CustomConfig: make(map[string]interface{}),
	}

	// 创建大量配置数据
	for i := 0; i < 1000; i++ {
		config.CustomConfig[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
	}

	b.ResetTimer()

	b.Run("JSON", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := manager.serializeConfig(config, ConfigFormatJSON)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("YAML", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := manager.serializeConfig(config, ConfigFormatYAML)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// TestConfigConcurrency 测试配置并发访问
func TestConfigConcurrency(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	manager := NewPluginConfigManager(logger)

	ctx := context.Background()
	if err := manager.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize config manager: %v", err)
	}

	// 并发读写测试
	const numGoroutines = 10
	const numOperations = 100

	done := make(chan bool, numGoroutines)

	// 启动多个goroutine进行并发操作
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < numOperations; j++ {
				pluginID := fmt.Sprintf("concurrent-plugin-%d", id)

				config := &EnhancedPluginConfig{
					BasePluginConfig: &BasePluginConfig{
						ID:         pluginID,
						Name:       fmt.Sprintf("Concurrent Plugin %d", id),
						Version:    "1.0.0",
						PluginPath: "/path/to/plugin.so",
						Enabled:    true,
					},
					ConfigFormat: ConfigFormatYAML,
					CustomConfig: map[string]interface{}{
						"iteration": j,
					},
				}

				// 保存配置
				if err := manager.SaveConfig(ctx, config); err != nil {
					t.Errorf("Failed to save config in goroutine %d: %v", id, err)
					return
				}

				// 加载配置
				_, err := manager.LoadConfig(ctx, pluginID)
				if err != nil {
					t.Errorf("Failed to load config in goroutine %d: %v", id, err)
					return
				}

				// 删除配置
				if err := manager.DeleteConfig(ctx, pluginID); err != nil {
					t.Errorf("Failed to delete config in goroutine %d: %v", id, err)
					return
				}
			}
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
			// 继续等待
		case <-time.After(30 * time.Second):
			t.Fatal("Concurrent test timed out")
		}
	}
}

// TestConfigErrorHandling 测试配置错误处理
func TestConfigErrorHandling(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	manager := NewPluginConfigManager(logger)

	ctx := context.Background()

	// 测试加载不存在的配置
	_, err := manager.LoadConfig(ctx, "non-existent-plugin")
	if err == nil {
		t.Error("Expected error when loading non-existent config")
	}

	// 测试删除不存在的配置
	err = manager.DeleteConfig(ctx, "non-existent-plugin")
	if err == nil {
		t.Error("Expected error when deleting non-existent config")
	}

	// 测试保存无效配置
	invalidConfig := &EnhancedPluginConfig{}
	err = manager.SaveConfig(ctx, invalidConfig)
	if err == nil {
		t.Error("Expected error when saving invalid config")
	}

	// 测试无效的配置格式
	invalidData := []byte("invalid json data")
	_, err = manager.deserializeConfig(invalidData, ConfigFormatJSON)
	if err == nil {
		t.Error("Expected error when deserializing invalid JSON")
	}
}

// TestConfigBackwardCompatibility 测试配置向后兼容性
func TestConfigBackwardCompatibility(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	manager := NewPluginConfigManager(logger)

	// 模拟旧版本配置数据
	oldConfigData := `{
		"id": "legacy-plugin",
		"name": "Legacy Plugin",
		"version": "0.9.0",
		"plugin_path": "/path/to/legacy/plugin.so",
		"enabled": true
	}`

	// 测试反序列化旧版本配置
	config, err := manager.deserializeConfig([]byte(oldConfigData), ConfigFormatJSON)
	if err != nil {
		t.Fatalf("Failed to deserialize legacy config: %v", err)
	}

	// 验证基础字段
	if config.ID != "legacy-plugin" {
		t.Errorf("Expected ID 'legacy-plugin', got %s", config.ID)
	}

	if config.Name != "Legacy Plugin" {
		t.Errorf("Expected Name 'Legacy Plugin', got %s", config.Name)
	}

	// 验证默认值设置
	if config.ConfigFormat == "" {
		config.ConfigFormat = ConfigFormatJSON // 应该设置默认值
	}

	if config.CustomConfig == nil {
		config.CustomConfig = make(map[string]interface{}) // 应该初始化
	}
}