package plugin

import (
	"context"
	"os"
	"strings"
	"testing"
	// "time" // 暂时未使用

	"log/slog"
)

// TestRequiredFieldsRule 测试必填字段验证规则
func TestRequiredFieldsRule(t *testing.T) {
	rule := NewRequiredFieldsRule()
	ctx := context.Background()

	// 测试有效配置
	validConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:         "test-plugin",
			Name:       "Test Plugin",
			Version:    "1.0.0",
			PluginPath: "/path/to/plugin.so",
			Enabled:    true,
		},
	}

	result := rule.Validate(ctx, validConfig)
	if !result.Valid {
		t.Error("Valid config should pass required fields validation")
	}

	if len(result.Errors) > 0 {
		t.Errorf("Valid config should not have errors, got %d", len(result.Errors))
	}

	// 测试缺少必填字段的配置
	invalidConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			// 缺少ID
			Name:       "Test Plugin",
			Version:    "1.0.0",
			PluginPath: "/path/to/plugin.so",
			Enabled:    true,
		},
	}

	result = rule.Validate(ctx, invalidConfig)
	if result.Valid {
		t.Error("Config missing required fields should fail validation")
	}

	if len(result.Errors) == 0 {
		t.Error("Config missing required fields should have errors")
	}

	// 检查错误代码
	found := false
	for _, err := range result.Errors {
		if err.Code == "REQUIRED_FIELD_MISSING" && err.Field == "id" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should have REQUIRED_FIELD_MISSING error for id field")
	}
}

// TestFormatValidationRule 测试格式验证规则
func TestFormatValidationRule(t *testing.T) {
	rule := NewFormatValidationRule()
	ctx := context.Background()

	// 测试有效格式
	validConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:         "valid-plugin-id",
			Name:       "Valid Plugin",
			Version:    "1.2.3",
			PluginPath: "/absolute/path/to/plugin.so",
			Enabled:    true,
		},
		ConfigPath: "/absolute/path/to/config.yaml",
	}

	result := rule.Validate(ctx, validConfig)
	if !result.Valid {
		t.Error("Valid format config should pass validation")
	}

	// 测试无效ID格式
	invalidIDConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:         "invalid plugin id!", // 包含空格和特殊字符
			Name:       "Invalid Plugin",
			Version:    "1.0.0",
			PluginPath: "/path/to/plugin.so",
			Enabled:    true,
		},
	}

	result = rule.Validate(ctx, invalidIDConfig)
	if result.Valid {
		t.Error("Invalid ID format should fail validation")
	}

	// 测试无效版本格式
	invalidVersionConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:         "test-plugin",
			Name:       "Test Plugin",
			Version:    "invalid.version", // 无效的语义版本
			PluginPath: "/path/to/plugin.so",
			Enabled:    true,
		},
	}

	result = rule.Validate(ctx, invalidVersionConfig)
	if result.Valid {
		t.Error("Invalid version format should fail validation")
	}

	// 测试相对路径
	relativePathConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:         "test-plugin",
			Name:       "Test Plugin",
			Version:    "1.0.0",
			PluginPath: "relative/path/plugin.so", // 相对路径
			Enabled:    true,
		},
	}

	result = rule.Validate(ctx, relativePathConfig)
	if result.Valid {
		t.Error("Relative path should fail validation")
	}
}

// TestPathValidationRule 测试路径验证规则
func TestPathValidationRule(t *testing.T) {
	rule := NewPathValidationRule()
	ctx := context.Background()

	// 创建临时文件用于测试
	tempFile, err := os.CreateTemp("", "test-plugin-*.so")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// 测试存在的文件路径
	validConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:         "test-plugin",
			Name:       "Test Plugin",
			Version:    "1.0.0",
			PluginPath: tempFile.Name(),
			Enabled:    true,
		},
	}

	result := rule.Validate(ctx, validConfig)
	if !result.Valid {
		t.Error("Existing file path should pass validation")
	}

	// 测试不存在的文件路径
	invalidConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:         "test-plugin",
			Name:       "Test Plugin",
			Version:    "1.0.0",
			PluginPath: "/non/existent/path/plugin.so",
			Enabled:    true,
		},
	}

	result = rule.Validate(ctx, invalidConfig)
	if result.Valid {
		t.Error("Non-existent file path should fail validation")
	}

	// 检查错误代码
	found := false
	for _, err := range result.Errors {
		if err.Code == "PATH_NOT_FOUND" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should have PATH_NOT_FOUND error")
	}
}

// TestDependencyValidationRule 测试依赖验证规则
func TestDependencyValidationRule(t *testing.T) {
	rule := NewDependencyValidationRule()
	ctx := context.Background()

	// 设置可用插件
	availablePlugins := map[string]*EnhancedPluginConfig{
		"plugin-a": {
			BasePluginConfig: &BasePluginConfig{
				ID:      "plugin-a",
				Version: "1.0.0",
			},
		},
		"plugin-b": {
			BasePluginConfig: &BasePluginConfig{
				ID:           "plugin-b",
				Version:      "2.0.0",
				Dependencies: []string{"plugin-a"},
			},
		},
	}
	rule.SetAvailablePlugins(availablePlugins)

	// 测试有效依赖
	validConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:           "plugin-c",
			Version:      "1.0.0",
			Dependencies: []string{"plugin-a"},
		},
	}

	result := rule.Validate(ctx, validConfig)
	if !result.Valid {
		t.Error("Valid dependencies should pass validation")
	}

	// 测试不存在的依赖
	invalidConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:           "plugin-d",
			Version:      "1.0.0",
			Dependencies: []string{"non-existent-plugin"},
		},
	}

	result = rule.Validate(ctx, invalidConfig)
	if result.Valid {
		t.Error("Non-existent dependency should fail validation")
	}

	// 测试循环依赖
	cyclicConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:           "plugin-a", // 与已存在的plugin-a形成循环
			Version:      "1.0.0",
			Dependencies: []string{"plugin-b"}, // plugin-b依赖plugin-a
		},
	}

	result = rule.Validate(ctx, cyclicConfig)
	if result.Valid {
		t.Error("Cyclic dependency should fail validation")
	}
}

// TestResourceValidationRule 测试资源验证规则
func TestResourceValidationRule(t *testing.T) {
	rule := NewResourceValidationRule()
	ctx := context.Background()

	// 测试没有资源限制的配置
	noLimitsConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:      "test-plugin",
			Version: "1.0.0",
		},
	}

	result := rule.Validate(ctx, noLimitsConfig)
	if !result.Valid {
		t.Error("Config without resource limits should pass validation")
	}

	if len(result.Suggestions) == 0 {
		t.Error("Config without resource limits should have suggestions")
	}

	// TODO: 测试有效的资源限制
	// 由于EnhancedPluginConfig中没有ResourceLimits字段，需要重新设计测试
	validLimitsConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:             "test-plugin",
			Version:        "1.0.0",
			ResourceLimits: DefaultResourceLimits(),
		},
	}

	result = rule.Validate(ctx, validLimitsConfig)
	if !result.Valid {
		t.Error("Valid resource limits should pass validation")
	}

	// TODO: 测试过高的内存限制
	// 由于EnhancedPluginConfig中没有ResourceLimits字段，需要重新设计测试
	highMemoryConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:      "test-plugin",
			Version: "1.0.0",
			ResourceLimits: &ResourceLimits{
				MaxMemoryMB: 4096, // 4GB
				MaxCPUPercent: 100,
				MaxGoroutines: 1000,
			},
		},
	}

	result = rule.Validate(ctx, highMemoryConfig)
	if len(result.Warnings) == 0 {
		t.Error("High memory limit should generate warnings")
	}

	// TODO: 测试无效的CPU限制
	// 由于EnhancedPluginConfig中没有ResourceLimits字段，需要重新设计测试
	invalidCPUConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:      "test-plugin",
			Version: "1.0.0",
			ResourceLimits: &ResourceLimits{
				MaxMemoryMB: 512,
				MaxCPUPercent: 150, // 超过100%
				MaxGoroutines: 1000,
			},
		},
	}

	result = rule.Validate(ctx, invalidCPUConfig)
	if result.Valid {
		t.Error("Invalid CPU limit should fail validation")
	}
}

// TestSecurityValidationRule 测试安全验证规则
func TestSecurityValidationRule(t *testing.T) {
	rule := NewSecurityValidationRule()
	ctx := context.Background()

	// 测试没有安全配置的插件
	noSecurityConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:      "test-plugin",
			Version: "1.0.0",
		},
	}

	result := rule.Validate(ctx, noSecurityConfig)
	if !result.Valid {
		t.Error("Config without security should pass validation")
	}

	if len(result.Warnings) == 0 {
		t.Error("Config without security should have warnings")
	}

	// TODO: 测试有效的安全配置
	// 由于EnhancedPluginConfig中没有SecurityConfig字段，需要重新设计测试
	validSecurityConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:             "test-plugin",
			Version:        "1.0.0",
			SecurityConfig: DefaultSecurityConfig(),
		},
	}

	result = rule.Validate(ctx, validSecurityConfig)
	if !result.Valid {
		t.Error("Valid security config should pass validation")
	}

	// TODO: 测试危险权限
	// 由于EnhancedPluginConfig中没有SecurityConfig字段，需要重新设计测试
	dangerousPermConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:      "test-plugin",
			Version: "1.0.0",
			SecurityConfig: &SecurityConfig{
				EnableSeccomp:    false, // 第一个危险配置
				EnableNamespace:  false, // 第二个危险配置
				AllowedPaths:     []string{"/tmp"},
				BlockedPaths:     []string{"/etc"},
				AllowedHosts:     []string{"localhost"},
				BlockedHosts:     []string{},
				AllowedPorts:     []int{80, 443},
				BlockedPorts:     []int{22},
				AllowedSyscalls:  []string{"read", "write"},
				BlockedSyscalls:  []string{"execve"},
				MaxConnections:   10,
				ConnectionTimeout: 30,
			},
		},
	}

	result = rule.Validate(ctx, dangerousPermConfig)
	if len(result.Warnings) == 0 {
		t.Error("Dangerous permissions should generate warnings")
	}

	// 验证警告内容
	dangerousWarnings := 0
	for _, warning := range result.Warnings {
		if warning.Code == "DANGEROUS_PERMISSION" {
			dangerousWarnings++
		}
	}
	if dangerousWarnings != 2 {
		t.Errorf("Expected 2 dangerous permission warnings, got %d", dangerousWarnings)
	}
}

// TestNetworkValidationRule 测试网络验证规则
func TestNetworkValidationRule(t *testing.T) {
	rule := NewNetworkValidationRule()
	ctx := context.Background()

	// 测试有效的网络配置
	validNetworkConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:      "test-plugin",
			Version: "1.0.0",
		},
		CustomConfig: map[string]interface{}{
			"server_url":  "https://api.example.com",
			"server_port": 8080,
			"host":        "localhost",
		},
	}

	result := rule.Validate(ctx, validNetworkConfig)
	if !result.Valid {
		t.Error("Valid network config should pass validation")
	}

	// 测试无效的URL
	invalidURLConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:      "test-plugin",
			Version: "1.0.0",
		},
		CustomConfig: map[string]interface{}{
			"api_url": "invalid-url",
		},
	}

	result = rule.Validate(ctx, invalidURLConfig)
	if result.Valid {
		t.Error("Invalid URL should fail validation")
	}

	// 测试无效的端口
	invalidPortConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:      "test-plugin",
			Version: "1.0.0",
		},
		CustomConfig: map[string]interface{}{
			"port": 70000, // 超出有效范围
		},
	}

	result = rule.Validate(ctx, invalidPortConfig)
	if result.Valid {
		t.Error("Invalid port should fail validation")
	}

	// 测试无效的主机地址
	invalidHostConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:      "test-plugin",
			Version: "1.0.0",
		},
		CustomConfig: map[string]interface{}{
			"host": "invalid..host",
		},
	}

	result = rule.Validate(ctx, invalidHostConfig)
	if result.Valid {
		t.Error("Invalid host should fail validation")
	}
}

// TestCustomValidationRule 测试自定义验证规则
func TestCustomValidationRule(t *testing.T) {
	// 创建自定义验证规则
	customRule := NewCustomValidationRule(
		"custom_test_rule",
		"Test custom validation rule",
		ValidationSeverityWarning,
		func(ctx context.Context, config *EnhancedPluginConfig) *ValidationResult {
			result := &ValidationResult{Valid: true}

			// 自定义验证逻辑：检查插件名称是否包含"test"
			if !strings.Contains(strings.ToLower(config.Name), "test") {
				result.AddWarning("name", "Plugin name should contain 'test' for testing", "CUSTOM_NAME_CHECK", config.Name)
			}

			return result
		},
	)

	ctx := context.Background()

	// 测试符合自定义规则的配置
	validConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:      "test-plugin",
			Name:    "Test Plugin",
			Version: "1.0.0",
		},
	}

	result := customRule.Validate(ctx, validConfig)
	if !result.Valid {
		t.Error("Config matching custom rule should pass validation")
	}

	if len(result.Warnings) > 0 {
		t.Error("Config matching custom rule should not have warnings")
	}

	// 测试不符合自定义规则的配置
	invalidConfig := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:      "prod-plugin",
			Name:    "Production Plugin",
			Version: "1.0.0",
		},
	}

	result = customRule.Validate(ctx, invalidConfig)
	if !result.Valid {
		t.Error("Custom rule should not fail validation, only warn")
	}

	if len(result.Warnings) == 0 {
		t.Error("Config not matching custom rule should have warnings")
	}

	// 验证警告内容
	if result.Warnings[0].Code != "CUSTOM_NAME_CHECK" {
		t.Errorf("Expected warning code 'CUSTOM_NAME_CHECK', got %s", result.Warnings[0].Code)
	}
}

// TestConfigValidator 测试配置验证器
func TestConfigValidator(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// 测试默认验证器
	validator := NewConfigValidator(logger, nil)

	if len(validator.GetRules()) == 0 {
		t.Error("Default validator should have built-in rules")
	}

	// 测试添加自定义规则
	customRule := NewCustomValidationRule(
		"test_rule",
		"Test rule",
		ValidationSeverityInfo,
		func(ctx context.Context, config *EnhancedPluginConfig) *ValidationResult {
			return &ValidationResult{Valid: true}
		},
	)

	validator.AddRule(customRule)

	// 验证规则已添加
	rules := validator.GetRules()
	found := false
	for _, rule := range rules {
		if rule.GetName() == "test_rule" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Custom rule should be added to validator")
	}

	// 测试移除规则
	validator.RemoveRule("test_rule")
	rules = validator.GetRules()
	for _, rule := range rules {
		if rule.GetName() == "test_rule" {
			t.Error("Custom rule should be removed from validator")
		}
	}
}

// TestConfigValidatorLevels 测试验证器级别
func TestConfigValidatorLevels(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// 测试基础级别
	basicOptions := &ConfigValidatorOptions{
		Level: ValidationLevelBasic,
	}
	basicValidator := NewConfigValidator(logger, basicOptions)
	basicRules := basicValidator.GetRules()

	// 测试标准级别
	standardOptions := &ConfigValidatorOptions{
		Level: ValidationLevelStandard,
	}
	standardValidator := NewConfigValidator(logger, standardOptions)
	standardRules := standardValidator.GetRules()

	// 测试严格级别
	strictOptions := &ConfigValidatorOptions{
		Level: ValidationLevelStrict,
	}
	strictValidator := NewConfigValidator(logger, strictOptions)
	strictRules := strictValidator.GetRules()

	// 验证规则数量递增
	if len(standardRules) <= len(basicRules) {
		t.Error("Standard level should have more rules than basic level")
	}

	if len(strictRules) <= len(standardRules) {
		t.Error("Strict level should have more rules than standard level")
	}

	// 测试动态设置级别
	basicValidator.SetLevel(ValidationLevelStrict)
	newRules := basicValidator.GetRules()
	if len(newRules) != len(strictRules) {
		t.Error("Setting level should update rules")
	}
}

// TestValidationResult 测试验证结果
func TestValidationResult(t *testing.T) {
	result := &ValidationResult{Valid: true}

	// 测试添加错误
	result.AddError("test_field", "Test error", "TEST_ERROR", "test_value")

	if result.Valid {
		t.Error("Result should be invalid after adding error")
	}

	if !result.HasErrors() {
		t.Error("Result should have errors")
	}

	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}

	// 测试添加警告
	result.AddWarning("test_field", "Test warning", "TEST_WARNING", "test_value")

	if !result.HasWarnings() {
		t.Error("Result should have warnings")
	}

	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}

	// 测试添加建议
	result.AddSuggestion("test_field", "Test suggestion", "Use better value")

	if len(result.Suggestions) != 1 {
		t.Errorf("Expected 1 suggestion, got %d", len(result.Suggestions))
	}

	// 测试合并结果
	other := &ValidationResult{
		Valid: false,
		Errors: []ValidationError{
			{Field: "other_field", Message: "Other error", Severity: ValidationSeverityError},
		},
		Metadata: map[string]interface{}{
			"test_key": "test_value",
		},
	}

	result.Merge(other)

	if len(result.Errors) != 2 {
		t.Errorf("Expected 2 errors after merge, got %d", len(result.Errors))
	}

	if result.Metadata["test_key"] != "test_value" {
		t.Error("Metadata should be merged")
	}
}

// BenchmarkConfigValidator 验证器性能测试
func BenchmarkConfigValidator(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	validator := NewConfigValidator(logger, nil)

	// TODO: 由于EnhancedPluginConfig中没有ResourceLimits和SecurityConfig字段，需要重新设计测试
	config := &EnhancedPluginConfig{
		BasePluginConfig: &BasePluginConfig{
			ID:             "benchmark-plugin",
			Name:           "Benchmark Plugin",
			Version:        "1.0.0",
			PluginPath:     "/path/to/plugin.so",
			Enabled:        true,
			ResourceLimits: DefaultResourceLimits(),
			SecurityConfig: DefaultSecurityConfig(),
		},
		CustomConfig: map[string]interface{}{
			"server_url":  "https://api.example.com",
			"server_port": 8080,
			"host":        "localhost",
		},
	}

	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validator.Validate(ctx, config)
	}
}