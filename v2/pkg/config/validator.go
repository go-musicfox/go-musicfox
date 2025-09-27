package config

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/knadh/koanf/v2"
)

// Validator 配置验证器接口
type Validator interface {
	Validate(config interface{}) error
	ValidateField(fieldName string, value interface{}, rules []string) error
}

// ConfigValidator 配置验证器实现
type ConfigValidator struct {
	rules map[string][]string
}

// NewValidator 创建新的配置验证器
func NewValidator() *ConfigValidator {
	return &ConfigValidator{
		rules: make(map[string][]string),
	}
}

// AddRule 添加验证规则
func (v *ConfigValidator) AddRule(fieldName string, rules []string) {
	v.rules[fieldName] = rules
}

// Validate 验证配置对象
func (v *ConfigValidator) Validate(config interface{}) error {
	val := reflect.ValueOf(config)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("config must be a struct")
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// 获取字段的验证规则
		validateTag := field.Tag.Get("validate")
		if validateTag == "" {
			continue
		}

		rules := strings.Split(validateTag, ",")
		if err := v.ValidateField(field.Name, fieldValue.Interface(), rules); err != nil {
			return fmt.Errorf("field %s: %w", field.Name, err)
		}
	}

	return nil
}

// ValidateField 验证单个字段
func (v *ConfigValidator) ValidateField(fieldName string, value interface{}, rules []string) error {
	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if err := v.validateRule(fieldName, value, rule); err != nil {
			return err
		}
	}
	return nil
}

// validateRule 验证单个规则
func (v *ConfigValidator) validateRule(fieldName string, value interface{}, rule string) error {
	parts := strings.SplitN(rule, "=", 2)
	ruleName := parts[0]
	ruleValue := ""
	if len(parts) > 1 {
		ruleValue = parts[1]
	}

	switch ruleName {
	case "required":
		return v.validateRequired(fieldName, value)
	case "min":
		return v.validateMin(fieldName, value, ruleValue)
	case "max":
		return v.validateMax(fieldName, value, ruleValue)
	case "oneof":
		return v.validateOneOf(fieldName, value, ruleValue)
	case "regex":
		return v.validateRegex(fieldName, value, ruleValue)
	default:
		return fmt.Errorf("unknown validation rule: %s", ruleName)
	}
}

// validateRequired 验证必填字段
func (v *ConfigValidator) validateRequired(fieldName string, value interface{}) error {
	if value == nil {
		return fmt.Errorf("%s is required", fieldName)
	}

	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.String:
		if val.String() == "" {
			return fmt.Errorf("%s is required", fieldName)
		}
	case reflect.Slice, reflect.Map, reflect.Array:
		if val.Len() == 0 {
			return fmt.Errorf("%s is required", fieldName)
		}
	case reflect.Ptr, reflect.Interface:
		if val.IsNil() {
			return fmt.Errorf("%s is required", fieldName)
		}
	}

	return nil
}

// validateMin 验证最小值
func (v *ConfigValidator) validateMin(fieldName string, value interface{}, minStr string) error {
	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		min := parseInt64(minStr)
		if val.Int() < min {
			return fmt.Errorf("%s must be >= %d", fieldName, min)
		}
	case reflect.Float32, reflect.Float64:
		min := parseFloat64(minStr)
		if val.Float() < min {
			return fmt.Errorf("%s must be >= %f", fieldName, min)
		}
	case reflect.String:
		min := parseInt64(minStr)
		if int64(len(val.String())) < min {
			return fmt.Errorf("%s length must be >= %d", fieldName, min)
		}
	}
	return nil
}

// validateMax 验证最大值
func (v *ConfigValidator) validateMax(fieldName string, value interface{}, maxStr string) error {
	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		max := parseInt64(maxStr)
		if val.Int() > max {
			return fmt.Errorf("%s must be <= %d", fieldName, max)
		}
	case reflect.Float32, reflect.Float64:
		max := parseFloat64(maxStr)
		if val.Float() > max {
			return fmt.Errorf("%s must be <= %f", fieldName, max)
		}
	case reflect.String:
		max := parseInt64(maxStr)
		if int64(len(val.String())) > max {
			return fmt.Errorf("%s length must be <= %d", fieldName, max)
		}
	}
	return nil
}

// validateOneOf 验证枚举值
func (v *ConfigValidator) validateOneOf(fieldName string, value interface{}, options string) error {
	valStr := fmt.Sprintf("%v", value)
	optionList := strings.Split(options, " ")

	for _, option := range optionList {
		if valStr == option {
			return nil
		}
	}

	return fmt.Errorf("%s must be one of: %s", fieldName, options)
}

// validateRegex 验证正则表达式
func (v *ConfigValidator) validateRegex(fieldName string, value interface{}, pattern string) error {
	valStr := fmt.Sprintf("%v", value)
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %s", pattern)
	}

	if !regex.MatchString(valStr) {
		return fmt.Errorf("%s does not match pattern: %s", fieldName, pattern)
	}

	return nil
}

// ConfigMerger 配置合并器
type ConfigMerger struct {
	strategy MergeStrategy
}

// MergeStrategy 合并策略
type MergeStrategy int

const (
	MergeStrategyOverride MergeStrategy = iota // 覆盖策略
	MergeStrategyAppend                        // 追加策略
	MergeStrategyMerge                         // 深度合并策略
)

// NewMerger 创建新的配置合并器
func NewMerger(strategy MergeStrategy) *ConfigMerger {
	return &ConfigMerger{
		strategy: strategy,
	}
}

// Merge 合并两个koanf配置
func (m *ConfigMerger) Merge(base, override *koanf.Koanf) (*koanf.Koanf, error) {
	result := koanf.New(".")

	// 先加载基础配置
	if err := result.Merge(base); err != nil {
		return nil, fmt.Errorf("failed to load base config: %w", err)
	}

	// 根据策略合并覆盖配置
	switch m.strategy {
	case MergeStrategyOverride:
		if err := result.Merge(override); err != nil {
			return nil, fmt.Errorf("failed to merge override config: %w", err)
		}
	case MergeStrategyAppend:
		return m.mergeAppend(result, override)
	case MergeStrategyMerge:
		return m.mergeDeep(result, override)
	default:
		return nil, fmt.Errorf("unsupported merge strategy: %d", m.strategy)
	}

	return result, nil
}

// mergeAppend 追加合并策略
func (m *ConfigMerger) mergeAppend(base, override *koanf.Koanf) (*koanf.Koanf, error) {
	// 对于数组类型的配置，进行追加而不是覆盖
	for _, key := range override.Keys() {
		overrideVal := override.Get(key)
		baseVal := base.Get(key)

		if baseVal != nil && reflect.TypeOf(baseVal).Kind() == reflect.Slice &&
			reflect.TypeOf(overrideVal).Kind() == reflect.Slice {
			// 合并数组
			baseSlice := reflect.ValueOf(baseVal)
			overrideSlice := reflect.ValueOf(overrideVal)
			mergedSlice := reflect.MakeSlice(baseSlice.Type(), 0, baseSlice.Len()+overrideSlice.Len())

			// 添加基础数组元素
			for i := 0; i < baseSlice.Len(); i++ {
				mergedSlice = reflect.Append(mergedSlice, baseSlice.Index(i))
			}

			// 添加覆盖数组元素
			for i := 0; i < overrideSlice.Len(); i++ {
				mergedSlice = reflect.Append(mergedSlice, overrideSlice.Index(i))
			}

			if err := base.Set(key, mergedSlice.Interface()); err != nil {
				return nil, err
			}
		} else {
			// 非数组类型直接覆盖
			if err := base.Set(key, overrideVal); err != nil {
				return nil, err
			}
		}
	}

	return base, nil
}

// mergeDeep 深度合并策略
func (m *ConfigMerger) mergeDeep(base, override *koanf.Koanf) (*koanf.Koanf, error) {
	// 深度合并，递归处理嵌套对象
	return base, base.Merge(override)
}

// 辅助函数
func parseInt64(s string) int64 {
	if s == "" {
		return 0
	}
	// 简化实现，实际应该处理错误
	var result int64
	_, _ = fmt.Sscanf(s, "%d", &result)
	return result
}

func parseFloat64(s string) float64 {
	if s == "" {
		return 0
	}
	// 简化实现，实际应该处理错误
	var result float64
	_, _ = fmt.Sscanf(s, "%f", &result)
	return result
}

// parseDuration 解析时间间隔
func parseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}