package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	koanfjson "github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// LoadTemplate 加载配置模板
func (am *AdvancedManager) LoadTemplate(templatePath string) error {
	am.templateMutex.Lock()
	defer am.templateMutex.Unlock()

	// 检查模板文件是否存在
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("template file not found: %s", templatePath)
	}

	// 获取模板名称（文件名不含扩展名）
	templateName := strings.TrimSuffix(filepath.Base(templatePath), filepath.Ext(templatePath))

	// 创建新的koanf实例加载模板
	templateKoanf := koanf.New(".")

	// 根据文件扩展名选择解析器
	ext := strings.ToLower(filepath.Ext(templatePath))
	var parser koanf.Parser
	switch ext {
	 case ".json":
		parser = koanfjson.Parser()
	case ".yaml", ".yml":
		parser = yaml.Parser()
	default:
		return fmt.Errorf("unsupported template format: %s", ext)
	}

	// 加载模板文件
	if err := templateKoanf.Load(file.Provider(templatePath), parser); err != nil {
		return fmt.Errorf("failed to load template: %w", err)
	}

	// 存储模板
	am.templates[templateName] = templateKoanf

	return nil
}

// ApplyTemplate 应用配置模板
func (am *AdvancedManager) ApplyTemplate(templateName string, variables map[string]interface{}) error {
	am.templateMutex.RLock()
	template, exists := am.templates[templateName]
	am.templateMutex.RUnlock()

	if !exists {
		return fmt.Errorf("template %s not found", templateName)
	}

	// 获取模板数据
	templateData := template.All()

	// 处理变量替换
	processedData, err := am.processTemplateVariables(templateData, variables)
	if err != nil {
		return fmt.Errorf("failed to process template variables: %w", err)
	}

	// 合并到当前配置
	for key, value := range processedData {
		oldValue := am.k.Get(key)
		am.k.Set(key, value)
		am.recordChange("template_apply", key, oldValue, value, "system", fmt.Sprintf("template_%s", templateName))
	}

	return nil
}

// processTemplateVariables 处理模板变量替换
func (am *AdvancedManager) processTemplateVariables(data map[string]interface{}, variables map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for key, value := range data {
		processedValue, err := am.processValue(value, variables)
		if err != nil {
			return nil, err
		}
		result[key] = processedValue
	}

	return result, nil
}

// processValue 处理单个值的变量替换
func (am *AdvancedManager) processValue(value interface{}, variables map[string]interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		return am.replaceVariables(v, variables), nil
	case map[string]interface{}:
		return am.processTemplateVariables(v, variables)
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			processedItem, err := am.processValue(item, variables)
			if err != nil {
				return nil, err
			}
			result[i] = processedItem
		}
		return result, nil
	default:
		return value, nil
	}
}

// replaceVariables 替换字符串中的变量
func (am *AdvancedManager) replaceVariables(str string, variables map[string]interface{}) string {
	// 支持 ${variable} 和 {{variable}} 两种格式
	patterns := []string{
		`\$\{([^}]+)\}`,
		`\{\{([^}]+)\}\}`,
	}

	result := str
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllStringFunc(result, func(match string) string {
			// 提取变量名
			submatches := re.FindStringSubmatch(match)
			if len(submatches) < 2 {
				return match
			}

			varName := strings.TrimSpace(submatches[1])

			// 支持默认值语法: variable:default_value
			parts := strings.SplitN(varName, ":", 2)
			actualVarName := parts[0]
			defaultValue := ""
			if len(parts) > 1 {
				defaultValue = parts[1]
			}

			// 查找变量值
			if value, exists := variables[actualVarName]; exists {
				return fmt.Sprintf("%v", value)
			}

			// 尝试从环境变量获取
			if envValue := os.Getenv(actualVarName); envValue != "" {
				return envValue
			}

			// 返回默认值或原始匹配
			if defaultValue != "" {
				return defaultValue
			}

			return match
		})
	}

	return result
}

// SetInheritance 设置配置继承
func (am *AdvancedManager) SetInheritance(parentConfig string) error {
	am.templateMutex.Lock()
	defer am.templateMutex.Unlock()

	am.parentConfig = parentConfig

	// 构建继承树
	if am.inheritanceTree == nil {
		am.inheritanceTree = make(map[string][]string)
	}

	// 不需要在继承树中添加"current"，因为这会导致循环继承检测出现问题
	// 继承关系通过parentConfig字段维护即可

	return nil
}

// ResolveInheritance 解析配置继承
func (am *AdvancedManager) ResolveInheritance() error {
	if am.parentConfig == "" {
		return nil // 没有继承关系
	}

	// 检查是否存在循环继承
	if err := am.checkCircularInheritance("current", make(map[string]bool)); err != nil {
		return err
	}

	// 加载父配置
	parentKoanf, err := am.loadParentConfig(am.parentConfig)
	if err != nil {
		return fmt.Errorf("failed to load parent config: %w", err)
	}

	// 创建新的配置实例，先加载父配置，再加载当前配置
	mergedKoanf := koanf.New(".")

	// 合并父配置
	for key, value := range parentKoanf.All() {
		mergedKoanf.Set(key, value)
	}

	// 合并当前配置（覆盖父配置）
	for key, value := range am.k.All() {
		mergedKoanf.Set(key, value)
	}

	// 替换当前配置
	am.k = mergedKoanf

	return nil
}

// checkCircularInheritance 检查循环继承
func (am *AdvancedManager) checkCircularInheritance(configName string, visited map[string]bool) error {
	if visited[configName] {
		return fmt.Errorf("circular inheritance detected: %s", configName)
	}

	visited[configName] = true
	defer delete(visited, configName)

	// 检查父配置
	if configName == "current" && am.parentConfig != "" {
		// 如果父配置是文件路径，不需要进一步检查循环继承
		if strings.Contains(am.parentConfig, "/") || strings.Contains(am.parentConfig, "\\") {
			return nil
		}
		return am.checkCircularInheritance(am.parentConfig, visited)
	}

	// 检查其他继承关系
	if children, exists := am.inheritanceTree[configName]; exists {
		for _, child := range children {
			if err := am.checkCircularInheritance(child, visited); err != nil {
				return err
			}
		}
	}

	return nil
}

// loadParentConfig 加载父配置
func (am *AdvancedManager) loadParentConfig(parentConfig string) (*koanf.Koanf, error) {
	// 如果是模板，从模板中加载
	am.templateMutex.RLock()
	if template, exists := am.templates[parentConfig]; exists {
		am.templateMutex.RUnlock()
		return template, nil
	}
	am.templateMutex.RUnlock()

	// 如果是文件路径，从文件加载
	if strings.Contains(parentConfig, "/") || strings.Contains(parentConfig, "\\") {
		parentKoanf := koanf.New(".")
		ext := strings.ToLower(filepath.Ext(parentConfig))
		var parser koanf.Parser
		switch ext {
		case ".json":
			parser = koanfjson.Parser()
		case ".yaml", ".yml":
			parser = yaml.Parser()
		default:
			return nil, fmt.Errorf("unsupported parent config format: %s", ext)
		}

		if err := parentKoanf.Load(file.Provider(parentConfig), parser); err != nil {
			return nil, err
		}

		return parentKoanf, nil
	}

	// 如果是版本ID，从版本历史加载
	am.versionMutex.RLock()
	defer am.versionMutex.RUnlock()

	for _, version := range am.versionHistory {
		if version.ID == parentConfig {
			parentKoanf := koanf.New(".")
			for key, value := range version.Data {
				parentKoanf.Set(key, value)
			}
			return parentKoanf, nil
		}
	}

	return nil, fmt.Errorf("parent config not found: %s", parentConfig)
}

// GetEffectiveConfig 获取解析继承后的有效配置
func (am *AdvancedManager) GetEffectiveConfig() (*koanf.Koanf, error) {
	// 如果没有继承关系，直接返回当前配置
	if am.parentConfig == "" {
		return am.k, nil
	}

	// 创建临时配置管理器进行继承解析
	tempManager := &AdvancedManager{
		Manager:         am.Manager,
		parentConfig:    am.parentConfig,
		inheritanceTree: am.inheritanceTree,
		templates:       am.templates,
	}

	// 复制当前配置
	tempManager.k = koanf.New(".")
	for key, value := range am.k.All() {
		tempManager.k.Set(key, value)
	}

	// 解析继承
	if err := tempManager.ResolveInheritance(); err != nil {
		return nil, err
	}

	return tempManager.k, nil
}