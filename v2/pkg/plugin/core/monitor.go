package plugin

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/plugin/loader"
)

// PluginMonitor 插件监控器
type PluginMonitor struct {
	logger   *slog.Logger
	interval time.Duration
	plugins  map[string]*ManagedPlugin
	mutex    sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	ticker   *time.Ticker
	metrics  *MonitorMetrics
}

// MonitorMetrics 监控指标
type MonitorMetrics struct {
	TotalPlugins    int
	RunningPlugins  int
	ErrorPlugins    int
	MemoryUsage     uint64
	CPUUsage        float64
	LastUpdateTime  time.Time
	mutex           sync.RWMutex
}

// NewPluginMonitor 创建新的插件监控器
func NewPluginMonitor(logger *slog.Logger, interval time.Duration) *PluginMonitor {
	if logger == nil || interval <= 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	m := &PluginMonitor{
		logger:   logger,
		interval: interval,
		plugins:  make(map[string]*ManagedPlugin),
		ctx:      ctx,
		cancel:   cancel,
		ticker:   time.NewTicker(interval),
		metrics:  &MonitorMetrics{},
	}

	// 启动监控协程
	go m.run()

	return m
}

// AddPlugin 添加插件到监控
func (m *PluginMonitor) AddPlugin(plugin *ManagedPlugin) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.plugins[plugin.ID] = plugin
	m.logger.Debug("Plugin added to monitor", slog.String("plugin_id", plugin.ID))
}

// RemovePlugin 从监控中移除插件
func (m *PluginMonitor) RemovePlugin(pluginID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.plugins, pluginID)
	m.logger.Debug("Plugin removed from monitor", slog.String("plugin_id", pluginID))
}

// GetMetrics 获取监控指标
func (m *PluginMonitor) GetMetrics() *MonitorMetrics {
	m.metrics.mutex.RLock()
	defer m.metrics.mutex.RUnlock()

	// 返回指标副本
	return &MonitorMetrics{
		TotalPlugins:   m.metrics.TotalPlugins,
		RunningPlugins: m.metrics.RunningPlugins,
		ErrorPlugins:   m.metrics.ErrorPlugins,
		MemoryUsage:    m.metrics.MemoryUsage,
		CPUUsage:       m.metrics.CPUUsage,
		LastUpdateTime: m.metrics.LastUpdateTime,
	}
}

// Stop 停止监控
func (m *PluginMonitor) Stop() {
	m.cancel()
	if m.ticker != nil {
		m.ticker.Stop()
	}
	m.logger.Info("Plugin monitor stopped")
}

// run 运行监控循环
func (m *PluginMonitor) run() {
	m.logger.Info("Plugin monitor started")

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-m.ticker.C:
			m.collectMetrics()
		}
	}
}

// collectMetrics 收集监控指标
func (m *PluginMonitor) collectMetrics() {
	m.mutex.RLock()
	plugins := make([]*ManagedPlugin, 0, len(m.plugins))
	for _, plugin := range m.plugins {
		plugins = append(plugins, plugin)
	}
	m.mutex.RUnlock()

	// 统计插件状态
	totalPlugins := len(plugins)
	runningPlugins := 0
	errorPlugins := 0
	var totalMemory uint64
	var totalCPU float64

	for _, plugin := range plugins {
		plugin.mutex.RLock()
		state := plugin.State
		plugin.mutex.RUnlock()

		switch state {
		case loader.PluginStateRunning:
			runningPlugins++
		case loader.PluginStateError:
			errorPlugins++
		}

		// 收集插件指标
		if metrics, err := m.getPluginMetrics(plugin); err == nil && metrics != nil {
			totalMemory += uint64(metrics.MemoryUsage)
			// CPU使用率计算（简化实现）
			totalCPU += metrics.CPUUsage
		}
	}

	// 更新监控指标
	m.metrics.mutex.Lock()
	m.metrics.TotalPlugins = totalPlugins
	m.metrics.RunningPlugins = runningPlugins
	m.metrics.ErrorPlugins = errorPlugins
	m.metrics.MemoryUsage = totalMemory
	m.metrics.CPUUsage = totalCPU
	m.metrics.LastUpdateTime = time.Now()
	m.metrics.mutex.Unlock()

	if m.logger != nil {
		m.logger.Debug("Metrics collected",
			slog.Int("total_plugins", totalPlugins),
			slog.Int("running_plugins", runningPlugins),
			slog.Int("error_plugins", errorPlugins),
			slog.Uint64("memory_usage", totalMemory),
			slog.Float64("cpu_usage", totalCPU),
		)
	}
}

// getPluginMetrics 获取插件指标
func (m *PluginMonitor) getPluginMetrics(plugin *ManagedPlugin) (*loader.PluginMetrics, error) {
	if plugin.Plugin == nil {
		return nil, fmt.Errorf("plugin is nil")
	}

	metrics, err := plugin.Plugin.GetMetrics()
	if err != nil {
		return nil, err
	}
	if metrics == nil {
		return &loader.PluginMetrics{
			PluginID:     plugin.ID,
			Uptime:       time.Since(plugin.LoadTime),
			MemoryUsage:  0,
			CPUUsage:     0.0,
			RequestCount: 0,
			ErrorCount:   0,
			Timestamp:    time.Now(),
		}, nil
	}

	return metrics, nil
}



// HealthResult 健康检查结果
type HealthResult struct {
	PluginID    string
	Healthy     bool
	LastCheck   time.Time
	ErrorCount  int
	LastError   string
	ResponseTime time.Duration
}