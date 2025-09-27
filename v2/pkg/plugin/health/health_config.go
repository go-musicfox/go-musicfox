// pkg/plugin/health_config.go
package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ExtendedHealthCheckConfig 扩展的健康检查配置
type ExtendedHealthCheckConfig struct {
	// 基础配置
	Enabled         bool          `json:"enabled" yaml:"enabled"`
	CheckInterval   time.Duration `json:"check_interval" yaml:"check_interval"`
	Timeout         time.Duration `json:"timeout" yaml:"timeout"`
	RetryCount      int           `json:"retry_count" yaml:"retry_count"`
	RetryInterval   time.Duration `json:"retry_interval" yaml:"retry_interval"`
	
	// 阈值配置
	Thresholds      *ExtendedHealthThresholds `json:"thresholds" yaml:"thresholds"`
	
	// 策略配置
	Strategies      []string `json:"strategies" yaml:"strategies"`
	RecoveryEnabled bool     `json:"recovery_enabled" yaml:"recovery_enabled"`
	RecoveryStrategies []string `json:"recovery_strategies" yaml:"recovery_strategies"`
	
	// 通知配置
	Notifications   *NotificationConfig `json:"notifications" yaml:"notifications"`
	
	// 指标收集配置
	MetricsConfig   *MetricsConfig `json:"metrics" yaml:"metrics"`
}

// ExtendedHealthThresholds 扩展的健康检查阈值配置
type ExtendedHealthThresholds struct {
	// 内存阈值 (字节)
	MemoryWarning  int64 `json:"memory_warning" yaml:"memory_warning"`
	MemoryCritical int64 `json:"memory_critical" yaml:"memory_critical"`
	
	// CPU阈值 (百分比)
	CPUWarning  float64 `json:"cpu_warning" yaml:"cpu_warning"`
	CPUCritical float64 `json:"cpu_critical" yaml:"cpu_critical"`
	
	// 响应时间阈值
	ResponseTimeWarning  time.Duration `json:"response_time_warning" yaml:"response_time_warning"`
	ResponseTimeCritical time.Duration `json:"response_time_critical" yaml:"response_time_critical"`
	
	// 错误率阈值 (百分比)
	ErrorRateWarning  float64 `json:"error_rate_warning" yaml:"error_rate_warning"`
	ErrorRateCritical float64 `json:"error_rate_critical" yaml:"error_rate_critical"`
	
	// 协程数量阈值
	GoroutineWarning  int `json:"goroutine_warning" yaml:"goroutine_warning"`
	GoroutineCritical int `json:"goroutine_critical" yaml:"goroutine_critical"`
	
	// 堆使用率阈值 (百分比)
	HeapUsageWarning  float64 `json:"heap_usage_warning" yaml:"heap_usage_warning"`
	HeapUsageCritical float64 `json:"heap_usage_critical" yaml:"heap_usage_critical"`
}

// NotificationConfig 通知配置
type NotificationConfig struct {
	Enabled   bool     `json:"enabled" yaml:"enabled"`
	Channels  []string `json:"channels" yaml:"channels"`
	WebhookURL string  `json:"webhook_url" yaml:"webhook_url"`
	EmailTo   []string `json:"email_to" yaml:"email_to"`
	SlackChannel string `json:"slack_channel" yaml:"slack_channel"`
}

// MetricsConfig 指标收集配置
type MetricsConfig struct {
	Enabled         bool          `json:"enabled" yaml:"enabled"`
	CollectInterval time.Duration `json:"collect_interval" yaml:"collect_interval"`
	RetentionPeriod time.Duration `json:"retention_period" yaml:"retention_period"`
	Collectors      []string      `json:"collectors" yaml:"collectors"`
	ExportEnabled   bool          `json:"export_enabled" yaml:"export_enabled"`
	ExportFormat    string        `json:"export_format" yaml:"export_format"`
	ExportPath      string        `json:"export_path" yaml:"export_path"`
}

// DefaultExtendedHealthCheckConfig 返回默认的扩展健康检查配置
func DefaultExtendedHealthCheckConfig() *ExtendedHealthCheckConfig {
	return &ExtendedHealthCheckConfig{
		Enabled:       true,
		CheckInterval: 30 * time.Second,
		Timeout:       10 * time.Second,
		RetryCount:    3,
		RetryInterval: 5 * time.Second,
		Thresholds: &ExtendedHealthThresholds{
			MemoryWarning:        50 * 1024 * 1024,  // 50MB
			MemoryCritical:       100 * 1024 * 1024, // 100MB
			CPUWarning:           70.0,
			CPUCritical:          90.0,
			ResponseTimeWarning:  1 * time.Second,
			ResponseTimeCritical: 5 * time.Second,
			ErrorRateWarning:     5.0,
			ErrorRateCritical:    15.0,
			GoroutineWarning:     500,
			GoroutineCritical:    1000,
			HeapUsageWarning:     75.0,
			HeapUsageCritical:    90.0,
		},
		Strategies:         []string{"basic", "performance", "resources"},
		RecoveryEnabled:    true,
		RecoveryStrategies: []string{"gc", "restart"},
		Notifications: &NotificationConfig{
			Enabled:  false,
			Channels: []string{"log"},
		},
		MetricsConfig: &MetricsConfig{
			Enabled:         true,
			CollectInterval: 10 * time.Second,
			RetentionPeriod: 24 * time.Hour,
			Collectors:      []string{"system", "plugin", "performance"},
			ExportEnabled:   false,
			ExportFormat:    "json",
			ExportPath:      "./metrics",
		},
	}
}

// HealthConfigManager 健康检查配置管理器
type HealthConfigManager struct {
	config     *ExtendedHealthCheckConfig
	configPath string
	mutex      sync.RWMutex
	watchers   []ExtendedConfigWatcher
}

// ExtendedConfigWatcher 扩展配置变更监听器
type ExtendedConfigWatcher interface {
	OnConfigChanged(config *ExtendedHealthCheckConfig) error
}

// NewHealthConfigManager 创建新的配置管理器
func NewHealthConfigManager(configPath string) *HealthConfigManager {
	return &HealthConfigManager{
		config:     DefaultExtendedHealthCheckConfig(),
		configPath: configPath,
		watchers:   make([]ExtendedConfigWatcher, 0),
	}
}

// LoadConfig 加载配置
func (m *HealthConfigManager) LoadConfig() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if m.configPath == "" {
		return nil // 使用默认配置
	}
	
	// 检查配置文件是否存在
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		// 配置文件不存在，创建默认配置文件
		return m.saveConfigLocked()
	}
	
	// 读取配置文件
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	// 解析配置
	var config ExtendedHealthCheckConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}
	
	// 验证配置
	if err := m.validateConfig(&config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	
	m.config = &config
	return nil
}

// SaveConfig 保存配置
func (m *HealthConfigManager) SaveConfig() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.saveConfigLocked()
}

func (m *HealthConfigManager) saveConfigLocked() error {
	if m.configPath == "" {
		return nil // 不保存到文件
	}
	
	// 确保目录存在
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// 序列化配置
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	// 写入文件
	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// GetConfig 获取当前配置
func (m *HealthConfigManager) GetConfig() *ExtendedHealthCheckConfig {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// 返回配置的副本
	configCopy := *m.config
	return &configCopy
}

// UpdateConfig 更新配置
func (m *HealthConfigManager) UpdateConfig(config *ExtendedHealthCheckConfig) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// 验证配置
	if err := m.validateConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	
	oldConfig := m.config
	m.config = config
	
	// 保存到文件
	if err := m.saveConfigLocked(); err != nil {
		m.config = oldConfig // 回滚
		return fmt.Errorf("failed to save config: %w", err)
	}
	
	// 通知监听器
	m.notifyWatchers(config)
	
	return nil
}

// UpdateThresholds 更新阈值配置
func (m *HealthConfigManager) UpdateThresholds(thresholds *ExtendedHealthThresholds) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if err := m.validateThresholds(thresholds); err != nil {
		return fmt.Errorf("invalid thresholds: %w", err)
	}
	
	m.config.Thresholds = thresholds
	
	// 保存配置
	if err := m.saveConfigLocked(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	
	// 通知监听器
	m.notifyWatchers(m.config)
	
	return nil
}

// AddWatcher 添加配置监听器
func (m *HealthConfigManager) AddWatcher(watcher ExtendedConfigWatcher) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.watchers = append(m.watchers, watcher)
}

// RemoveWatcher 移除配置监听器
func (m *HealthConfigManager) RemoveWatcher(watcher ExtendedConfigWatcher) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	for i, w := range m.watchers {
		if w == watcher {
			m.watchers = append(m.watchers[:i], m.watchers[i+1:]...)
			break
		}
	}
}

// notifyWatchers 通知所有监听器
func (m *HealthConfigManager) notifyWatchers(config *ExtendedHealthCheckConfig) {
	for _, watcher := range m.watchers {
		go func(w ExtendedConfigWatcher) {
			if err := w.OnConfigChanged(config); err != nil {
				// 记录错误，但不影响其他监听器
				fmt.Printf("Config watcher error: %v\n", err)
			}
		}(watcher)
	}
}

// validateConfig 验证配置
func (m *HealthConfigManager) validateConfig(config *ExtendedHealthCheckConfig) error {
	if config.CheckInterval <= 0 {
		return fmt.Errorf("check_interval must be positive")
	}
	
	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	
	if config.RetryCount < 0 {
		return fmt.Errorf("retry_count must be non-negative")
	}
	
	if config.RetryInterval < 0 {
		return fmt.Errorf("retry_interval must be non-negative")
	}
	
	if config.Thresholds != nil {
		if err := m.validateThresholds(config.Thresholds); err != nil {
			return err
		}
	}
	
	if config.MetricsConfig != nil {
		if err := m.validateMetricsConfig(config.MetricsConfig); err != nil {
			return err
		}
	}
	
	return nil
}

// validateThresholds 验证阈值配置
func (m *HealthConfigManager) validateThresholds(thresholds *ExtendedHealthThresholds) error {
	if thresholds.MemoryWarning < 0 || thresholds.MemoryCritical < 0 {
		return fmt.Errorf("memory thresholds must be non-negative")
	}
	
	if thresholds.MemoryWarning >= thresholds.MemoryCritical {
		return fmt.Errorf("memory warning threshold must be less than critical threshold")
	}
	
	if thresholds.CPUWarning < 0 || thresholds.CPUWarning > 100 ||
	   thresholds.CPUCritical < 0 || thresholds.CPUCritical > 100 {
		return fmt.Errorf("CPU thresholds must be between 0 and 100")
	}
	
	if thresholds.CPUWarning >= thresholds.CPUCritical {
		return fmt.Errorf("CPU warning threshold must be less than critical threshold")
	}
	
	if thresholds.ResponseTimeWarning < 0 || thresholds.ResponseTimeCritical < 0 {
		return fmt.Errorf("response time thresholds must be non-negative")
	}
	
	if thresholds.ResponseTimeWarning >= thresholds.ResponseTimeCritical {
		return fmt.Errorf("response time warning threshold must be less than critical threshold")
	}
	
	if thresholds.ErrorRateWarning < 0 || thresholds.ErrorRateWarning > 100 ||
	   thresholds.ErrorRateCritical < 0 || thresholds.ErrorRateCritical > 100 {
		return fmt.Errorf("error rate thresholds must be between 0 and 100")
	}
	
	if thresholds.ErrorRateWarning >= thresholds.ErrorRateCritical {
		return fmt.Errorf("error rate warning threshold must be less than critical threshold")
	}
	
	if thresholds.GoroutineWarning < 0 || thresholds.GoroutineCritical < 0 {
		return fmt.Errorf("goroutine thresholds must be non-negative")
	}
	
	if thresholds.GoroutineWarning >= thresholds.GoroutineCritical {
		return fmt.Errorf("goroutine warning threshold must be less than critical threshold")
	}
	
	if thresholds.HeapUsageWarning < 0 || thresholds.HeapUsageWarning > 100 ||
	   thresholds.HeapUsageCritical < 0 || thresholds.HeapUsageCritical > 100 {
		return fmt.Errorf("heap usage thresholds must be between 0 and 100")
	}
	
	if thresholds.HeapUsageWarning >= thresholds.HeapUsageCritical {
		return fmt.Errorf("heap usage warning threshold must be less than critical threshold")
	}
	
	return nil
}

// validateMetricsConfig 验证指标配置
func (m *HealthConfigManager) validateMetricsConfig(config *MetricsConfig) error {
	if config.CollectInterval <= 0 {
		return fmt.Errorf("collect_interval must be positive")
	}
	
	if config.RetentionPeriod <= 0 {
		return fmt.Errorf("retention_period must be positive")
	}
	
	if config.ExportEnabled {
		if config.ExportFormat == "" {
			return fmt.Errorf("export_format is required when export is enabled")
		}
		
		if config.ExportPath == "" {
			return fmt.Errorf("export_path is required when export is enabled")
		}
		
		validFormats := []string{"json", "csv", "prometheus"}
		validFormat := false
		for _, format := range validFormats {
			if config.ExportFormat == format {
				validFormat = true
				break
			}
		}
		
		if !validFormat {
			return fmt.Errorf("invalid export_format: %s, valid formats: %v", config.ExportFormat, validFormats)
		}
	}
	
	return nil
}

// IsHealthy 根据阈值判断指标是否健康
func (t *ExtendedHealthThresholds) IsHealthy(metricName string, value interface{}) HealthStatus {
	switch metricName {
	case "memory_usage":
		if memValue, ok := value.(int64); ok {
			if memValue >= t.MemoryCritical {
				return HealthStatusCritical
			} else if memValue >= t.MemoryWarning {
				return HealthStatusDegraded
			}
			return HealthStatusHealthy
		}
	
	case "cpu_usage":
		if cpuValue, ok := value.(float64); ok {
			if cpuValue >= t.CPUCritical {
				return HealthStatusCritical
			} else if cpuValue >= t.CPUWarning {
				return HealthStatusDegraded
			}
			return HealthStatusHealthy
		}
	
	case "response_time":
		if rtValue, ok := value.(time.Duration); ok {
			if rtValue >= t.ResponseTimeCritical {
				return HealthStatusCritical
			} else if rtValue >= t.ResponseTimeWarning {
				return HealthStatusDegraded
			}
			return HealthStatusHealthy
		}
	
	case "error_rate":
		if errValue, ok := value.(float64); ok {
			if errValue >= t.ErrorRateCritical {
				return HealthStatusCritical
			} else if errValue >= t.ErrorRateWarning {
				return HealthStatusDegraded
			}
			return HealthStatusHealthy
		}
	
	case "goroutines":
		if gorValue, ok := value.(int); ok {
			if gorValue >= t.GoroutineCritical {
				return HealthStatusCritical
			} else if gorValue >= t.GoroutineWarning {
				return HealthStatusDegraded
			}
			return HealthStatusHealthy
		}
	
	case "heap_usage":
		if heapValue, ok := value.(float64); ok {
			if heapValue >= t.HeapUsageCritical {
				return HealthStatusCritical
			} else if heapValue >= t.HeapUsageWarning {
				return HealthStatusDegraded
			}
			return HealthStatusHealthy
		}
	}
	
	return HealthStatusHealthy
}