package config

import (
	"testing"

	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	assert.NotNil(t, validator)
	assert.NotNil(t, validator.rules)
}

func TestValidator_AddRule(t *testing.T) {
	validator := NewValidator()
	rules := []string{"required", "min=1", "max=100"}
	
	validator.AddRule("test_field", rules)
	
	assert.Equal(t, rules, validator.rules["test_field"])
}

func TestValidator_ValidateField(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name      string
		fieldName string
		value     interface{}
		rules     []string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "required field with value",
			fieldName: "name",
			value:     "test",
			rules:     []string{"required"},
			wantErr:   false,
		},
		{
			name:      "required field empty",
			fieldName: "name",
			value:     "",
			rules:     []string{"required"},
			wantErr:   true,
			errMsg:    "name is required",
		},
		{
			name:      "min validation pass",
			fieldName: "age",
			value:     25,
			rules:     []string{"min=18"},
			wantErr:   false,
		},
		{
			name:      "min validation fail",
			fieldName: "age",
			value:     15,
			rules:     []string{"min=18"},
			wantErr:   true,
			errMsg:    "age must be >= 18",
		},
		{
			name:      "max validation pass",
			fieldName: "score",
			value:     85,
			rules:     []string{"max=100"},
			wantErr:   false,
		},
		{
			name:      "max validation fail",
			fieldName: "score",
			value:     105,
			rules:     []string{"max=100"},
			wantErr:   true,
			errMsg:    "score must be <= 100",
		},
		{
			name:      "oneof validation pass",
			fieldName: "status",
			value:     "active",
			rules:     []string{"oneof=active inactive pending"},
			wantErr:   false,
		},
		{
			name:      "oneof validation fail",
			fieldName: "status",
			value:     "unknown",
			rules:     []string{"oneof=active inactive pending"},
			wantErr:   true,
			errMsg:    "status must be one of: active inactive pending",
		},
		{
			name:      "regex validation pass",
			fieldName: "version",
			value:     "1.0.0",
			rules:     []string{"regex=^\\d+\\.\\d+\\.\\d+$"},
			wantErr:   false,
		},
		{
			name:      "regex validation fail",
			fieldName: "version",
			value:     "invalid-version",
			rules:     []string{"regex=^\\d+\\.\\d+\\.\\d+$"},
			wantErr:   true,
			errMsg:    "version does not match pattern",
		},
		{
			name:      "string length min validation",
			fieldName: "password",
			value:     "12345678",
			rules:     []string{"min=8"},
			wantErr:   false,
		},
		{
			name:      "string length min validation fail",
			fieldName: "password",
			value:     "123",
			rules:     []string{"min=8"},
			wantErr:   true,
			errMsg:    "password length must be >= 8",
		},
		{
			name:      "multiple rules pass",
			fieldName: "username",
			value:     "testuser",
			rules:     []string{"required", "min=3", "max=20"},
			wantErr:   false,
		},
		{
			name:      "unknown rule",
			fieldName: "field",
			value:     "value",
			rules:     []string{"unknown_rule"},
			wantErr:   true,
			errMsg:    "unknown validation rule: unknown_rule",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateField(tt.fieldName, tt.value, tt.rules)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" && err != nil {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_Validate(t *testing.T) {
	validator := NewValidator()

	// 定义测试结构体
	type TestStruct struct {
		Name    string `validate:"required,min=3"`
		Age     int    `validate:"min=0,max=150"`
		Email   string `validate:"required"`
		Status  string `validate:"oneof=active inactive"`
		Optional string // 没有验证标签
	}

	tests := []struct {
		name    string
		config  interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid struct",
			config: TestStruct{
				Name:   "John Doe",
				Age:    30,
				Email:  "john@example.com",
				Status: "active",
			},
			wantErr: false,
		},
		{
			name: "invalid name",
			config: TestStruct{
				Name:   "Jo", // 太短
				Age:    30,
				Email:  "john@example.com",
				Status: "active",
			},
			wantErr: true,
			errMsg:  "field Name",
		},
		{
			name: "invalid age",
			config: TestStruct{
				Name:   "John Doe",
				Age:    -5, // 负数
				Email:  "john@example.com",
				Status: "active",
			},
			wantErr: true,
			errMsg:  "field Age",
		},
		{
			name: "invalid email",
			config: TestStruct{
				Name:   "John Doe",
				Age:    30,
				Email:  "", // 空字符串触发required验证失败
				Status: "active",
			},
			wantErr: true,
			errMsg:  "field Email",
		},
		{
			name: "invalid status",
			config: TestStruct{
				Name:   "John Doe",
				Age:    30,
				Email:  "john@example.com",
				Status: "unknown",
			},
			wantErr: true,
			errMsg:  "field Status",
		},
		{
			name:    "non-struct input",
			config:  "not a struct",
			wantErr: true,
			errMsg:  "config must be a struct",
		},
	}

	for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validator.Validate(tt.config)
				if tt.wantErr {
					assert.Error(t, err)
					if tt.errMsg != "" && err != nil {
						assert.Contains(t, err.Error(), tt.errMsg)
					}
				} else {
					assert.NoError(t, err)
				}
			})
		}
}

func TestNewMerger(t *testing.T) {
	merger := NewMerger(MergeStrategyOverride)
	assert.NotNil(t, merger)
	assert.Equal(t, MergeStrategyOverride, merger.strategy)
}

func TestConfigMerger_Merge(t *testing.T) {
	tests := []struct {
		name     string
		strategy MergeStrategy
		base     map[string]interface{}
		override map[string]interface{}
		expected map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "override strategy",
			strategy: MergeStrategyOverride,
			base: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			override: map[string]interface{}{
				"key2": "override_value2",
				"key3": "value3",
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": "override_value2",
				"key3": "value3",
			},
			wantErr: false,
		},
		{
			name:     "append strategy with arrays",
			strategy: MergeStrategyAppend,
			base: map[string]interface{}{
				"items": []string{"item1", "item2"},
				"key1":  "value1",
			},
			override: map[string]interface{}{
				"items": []string{"item3", "item4"},
				"key2":  "value2",
			},
			expected: map[string]interface{}{
				"items": []string{"item1", "item2", "item3", "item4"},
				"key1":  "value1",
				"key2":  "value2",
			},
			wantErr: false,
		},
		{
			name:     "deep merge strategy",
			strategy: MergeStrategyMerge,
			base: map[string]interface{}{
				"nested": map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			},
			override: map[string]interface{}{
				"nested": map[string]interface{}{
					"key2": "override_value2",
					"key3": "value3",
				},
			},
			expected: map[string]interface{}{
				"nested": map[string]interface{}{
					"key1": "value1",
					"key2": "override_value2",
					"key3": "value3",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merger := NewMerger(tt.strategy)

			// 创建koanf实例
			baseKoanf := koanf.New(".")
			overrideKoanf := koanf.New(".")

			// 加载测试数据
			for k, v := range tt.base {
				_ = baseKoanf.Set(k, v)
			}
			for k, v := range tt.override {
				_ = overrideKoanf.Set(k, v)
			}

			// 执行合并
			result, err := merger.Merge(baseKoanf, overrideKoanf)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)

				// 验证合并结果
				for k, expectedValue := range tt.expected {
					actualValue := result.Get(k)
					assert.Equal(t, expectedValue, actualValue, "key: %s", k)
				}
			}
		})
	}
}

func TestMergeStrategy_UnsupportedStrategy(t *testing.T) {
	merger := &ConfigMerger{strategy: MergeStrategy(999)} // 不支持的策略

	baseKoanf := koanf.New(".")
	overrideKoanf := koanf.New(".")

	_, err := merger.Merge(baseKoanf, overrideKoanf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported merge strategy")
}

func TestParseHelperFunctions(t *testing.T) {
	// 测试parseInt64
	assert.Equal(t, int64(42), parseInt64("42"))
	assert.Equal(t, int64(0), parseInt64(""))
	assert.Equal(t, int64(0), parseInt64("invalid"))

	// 测试parseFloat64
	assert.Equal(t, 3.14, parseFloat64("3.14"))
	assert.Equal(t, 0.0, parseFloat64(""))
	assert.Equal(t, 0.0, parseFloat64("invalid"))

	// 测试parseDuration
	duration, err := parseDuration("5m")
	assert.NoError(t, err)
	assert.Equal(t, "5m0s", duration.String())

	_, err = parseDuration("invalid")
	assert.Error(t, err)
}