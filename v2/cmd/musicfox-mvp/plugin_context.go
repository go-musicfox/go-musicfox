package main

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Plugin 插件接口
type Plugin interface {
	// 基础信息
	GetInfo() *PluginInfo
	GetCapabilities() []string
	GetDependencies() []string

	// 生命周期管理
	Initialize() error
	Start() error
	Stop() error
	Cleanup() error

	// 健康检查
	HealthCheck() error
}

// PluginInfo 插件信息
type PluginInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	Type        string    `json:"type"`
	CreatedAt   time.Time `json:"created_at"`
}

// PluginManager 插件管理器
type PluginManager struct {
	logger   *slog.Logger
	plugins  map[string]Plugin
	running  bool
	mutex    sync.RWMutex
}

// NewPluginManager 创建插件管理器
func NewPluginManager(logger *slog.Logger) *PluginManager {
	return &PluginManager{
		logger:  logger,
		plugins: make(map[string]Plugin),
	}
}

// Start 启动插件管理器
func (pm *PluginManager) Start() error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if pm.running {
		return nil
	}

	pm.logger.Info("Starting plugin manager")
	pm.running = true
	pm.logger.Info("Plugin manager started successfully")
	return nil
}

// Stop 停止插件管理器
func (pm *PluginManager) Stop() error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if !pm.running {
		return nil
	}

	pm.logger.Info("Stopping plugin manager")
	pm.running = false
	pm.logger.Info("Plugin manager stopped successfully")
	return nil
}

// RegisterPlugin 注册插件
func (pm *PluginManager) RegisterPlugin(plugin Plugin) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	info := plugin.GetInfo()
	if info == nil {
		return fmt.Errorf("plugin info is nil")
	}

	if _, exists := pm.plugins[info.ID]; exists {
		return fmt.Errorf("plugin %s already registered", info.ID)
	}

	pm.plugins[info.ID] = plugin
	pm.logger.Info("Plugin registered", "id", info.ID, "name", info.Name)
	return nil
}

// GetPlugin 获取插件
func (pm *PluginManager) GetPlugin(id string) (Plugin, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	plugin, exists := pm.plugins[id]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", id)
	}

	return plugin, nil
}

// ListPlugins 列出所有插件
func (pm *PluginManager) ListPlugins() []Plugin {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	plugins := make([]Plugin, 0, len(pm.plugins))
	for _, plugin := range pm.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}