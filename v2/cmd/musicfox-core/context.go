package main

import (
	"context"

	"github.com/go-musicfox/go-musicfox/v2/pkg/kernel"
	plugin "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
)

// PluginContext 插件上下文实现
type PluginContext struct {
	serviceRegistry kernel.ServiceRegistry
	pluginConfig    plugin.PluginConfig
}

// GetServiceRegistry 获取服务注册表
func (pc *PluginContext) GetServiceRegistry() kernel.ServiceRegistry {
	return pc.serviceRegistry
}

// GetPluginConfig 获取插件配置
func (pc *PluginContext) GetPluginConfig() plugin.PluginConfig {
	return pc.pluginConfig
}

// GetLogger 获取日志器
func (pc *PluginContext) GetLogger() interface{} {
	// 从服务注册表获取日志器
	if logger, err := pc.serviceRegistry.GetService(context.Background(), "logger"); err == nil {
		return logger
	}
	return nil
}

// GetEventBus 获取事件总线
func (pc *PluginContext) GetEventBus() interface{} {
	// 从服务注册表获取事件总线
	if eventBus, err := pc.serviceRegistry.GetService(context.Background(), "event_bus"); err == nil {
		return eventBus
	}
	return nil
}