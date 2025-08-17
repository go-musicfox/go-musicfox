package composer

import (
	"bytes"
	"fmt"
	"maps"
	"sync"
	"text/template"
)

// TemplateManager 负责安全地管理和执行模板
type TemplateManager struct {
	mu        sync.RWMutex
	tpl       *template.Template
	templates map[string]string
}

// NewTemplateManager 创建一个新的模板管理器
func NewTemplateManager() *TemplateManager {
	return &TemplateManager{
		tpl:       template.New("root"),
		templates: make(map[string]string),
	}
}

// Register 动态注册一个新模板或覆盖现有模板
func (tm *TemplateManager) Register(name, content string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	_, err := tm.tpl.New(name).Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse template %q: %w", name, err)
	}
	tm.templates[name] = content
	return nil
}

// LoadFromMap 从一个 map 批量注册模板
func (tm *TemplateManager) LoadFromMap(templates map[string]string) error {
	for name, content := range templates {
		if err := tm.Register(name, content); err != nil {
			return err
		}
	}
	return nil
}

// GetTemplates 返回所有已注册模板的副本
func (tm *TemplateManager) GetTemplates() map[string]string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	clone := make(map[string]string, len(tm.templates))
	maps.Copy(clone, tm.templates)
	return clone
}

// Execute 安全地执行一个模板
func (tm *TemplateManager) Execute(name string, data any) (string, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var wr bytes.Buffer
	if t := tm.tpl.Lookup(name); t == nil {
		return "", fmt.Errorf("template %q not found", name)
	}

	err := tm.tpl.ExecuteTemplate(&wr, name, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template %q: %w", name, err)
	}
	return wr.String(), nil
}
