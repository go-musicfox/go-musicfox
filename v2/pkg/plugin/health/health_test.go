// pkg/plugin/health_test.go
package plugin

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	core "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
)

// MockPlugin 模拟插件用于测试
type MockPlugin struct {
	info        *core.PluginInfo
	healthError error
	isStarted   bool
	mutex       sync.RWMutex
}

func NewMockPlugin(name, version string) *MockPlugin {
	return &MockPlugin{
		info: &core.PluginInfo{
			Name:        name,
			Version:     version,
			Description: "Mock plugin for testing",
			Author:      "Test",
			LoadTime:    time.Now(),
		},
		isStarted: false,
	}
}

func (m *MockPlugin) GetInfo() *core.PluginInfo {
	return m.info
}

func (m *MockPlugin) Start() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.isStarted = true
	return nil
}

func (m *MockPlugin) Stop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.isStarted = false
	return nil
}

func (m *MockPlugin) Cleanup() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.isStarted = false
	return nil
}

func (m *MockPlugin) HealthCheck() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.healthError
}

func (m *MockPlugin) SetHealthError(err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.healthError = err
}

func (m *MockPlugin) IsStarted() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.isStarted
}

func (m *MockPlugin) GetCapabilities() []string {
	return []string{"health_check", "metrics"}
}

func (m *MockPlugin) GetDependencies() []string {
	return []string{}
}

func (m *MockPlugin) Initialize(ctx core.PluginContext) error {
	return nil
}

func (m *MockPlugin) ValidateConfig(config map[string]interface{}) error {
	return nil
}

func (m *MockPlugin) UpdateConfig(config map[string]interface{}) error {
	return nil
}

func (m *MockPlugin) GetMetrics() (*core.PluginMetrics, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	return &core.PluginMetrics{
		StartTime:     time.Now().Add(-time.Since(m.info.LoadTime)),
		Uptime:        time.Since(m.info.LoadTime),
		RequestCount:  100,
		ErrorCount:    0,
		MemoryUsage:   1024,
		CPUUsage:      0.1,
		CustomMetrics: map[string]interface{}{
			"goroutine_count":    5,
			"success_rate":       0.99,
			"avg_response_time":  time.Millisecond * 10,
			"throughput":         10.0,
		},
	}, nil
}

func (m *MockPlugin) HandleEvent(event interface{}) error {
	// Mock implementation - just return nil
	return nil
}

// TestDefaultHealthChecker 测试默认健康检查器
func TestDefaultHealthChecker(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	plugin := NewMockPlugin("test-plugin", "1.0.0")
	config := DefaultExtendedHealthCheckConfig()
	
	checker := NewDefaultHealthChecker(plugin, config, logger)
	
	if checker == nil {
		t.Fatal("Expected non-nil health checker")
	}
	
	if checker.plugin != plugin {
		t.Error("Expected plugin to be set correctly")
	}
	
	if checker.config != config {
		t.Error("Expected config to be set correctly")
	}
}

// TestHealthCheck 测试健康检查功能
func TestHealthCheck(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	plugin := NewMockPlugin("test-plugin", "1.0.0")
	config := DefaultExtendedHealthCheckConfig()
	
	checker := NewDefaultHealthChecker(plugin, config, logger)
	ctx := context.Background()
	
	// 测试健康的插件
	result, err := checker.CheckHealth(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result.Status != HealthStatusHealthy {
		t.Errorf("Expected healthy status, got %v", result.Status)
	}
	
	// 测试不健康的插件
	plugin.SetHealthError(errors.New("plugin error"))
	result, err = checker.CheckHealth(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result.Status == HealthStatusHealthy {
		t.Error("Expected unhealthy status")
	}
}

// TestHealthCheckWithStrategies 测试不同的健康检查策略
func TestHealthCheckWithStrategies(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	plugin := NewMockPlugin("test-plugin", "1.0.0")
	config := DefaultExtendedHealthCheckConfig()
	
	checker := NewDefaultHealthChecker(plugin, config, logger)
	ctx := context.Background()
	
	// 测试基础策略
	result, err := checker.CheckHealthWithStrategy(ctx, "basic")
	if err != nil {
		t.Fatalf("Expected no error for basic strategy, got %v", err)
	}
	
	if result.Status != HealthStatusHealthy {
		t.Errorf("Expected healthy status for basic strategy, got %v", result.Status)
	}
	
	// 测试性能策略
	result, err = checker.CheckHealthWithStrategy(ctx, "performance")
	if err != nil {
		t.Fatalf("Expected no error for performance strategy, got %v", err)
	}
	
	if result.Details == nil {
		t.Error("Expected performance details to be present")
	}
	
	// 测试资源策略
	result, err = checker.CheckHealthWithStrategy(ctx, "resources")
	if err != nil {
		t.Fatalf("Expected no error for resources strategy, got %v", err)
	}
	
	if result.Details == nil {
		t.Error("Expected resource details to be present")
	}
	
	// 测试不存在的策略
	_, err = checker.CheckHealthWithStrategy(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent strategy")
	}
}

// TestMetricsCollection 测试指标收集
func TestMetricsCollection(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	plugin := NewMockPlugin("test-plugin", "1.0.0")
	config := DefaultExtendedHealthCheckConfig()
	
	checker := NewDefaultHealthChecker(plugin, config, logger)
	ctx := context.Background()
	
	// 测试系统指标收集
	metrics, err := checker.CollectMetrics(ctx, "system")
	if err != nil {
		t.Fatalf("Expected no error for system metrics, got %v", err)
	}
	
	if metrics == nil {
		t.Error("Expected non-nil metrics")
	}
	
	if _, ok := metrics["memory_usage"]; !ok {
		t.Error("Expected memory_usage metric")
	}
	
	if _, ok := metrics["goroutines"]; !ok {
		t.Error("Expected goroutines metric")
	}
	
	// 测试插件指标收集
	metrics, err = checker.CollectMetrics(ctx, "plugin")
	if err != nil {
		t.Fatalf("Expected no error for plugin metrics, got %v", err)
	}
	
	if _, ok := metrics["plugin_name"]; !ok {
		t.Error("Expected plugin_name metric")
	}
	
	if _, ok := metrics["response_time"]; !ok {
		t.Error("Expected response_time metric")
	}
	
	// 测试不存在的收集器
	_, err = checker.CollectMetrics(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent collector")
	}
}

// TestRecoveryStrategies 测试恢复策略
func TestRecoveryStrategies(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	plugin := NewMockPlugin("test-plugin", "1.0.0")
	config := DefaultExtendedHealthCheckConfig()
	
	checker := NewDefaultHealthChecker(plugin, config, logger)
	ctx := context.Background()
	
	// 创建一个不健康的结果
	result := &HealthCheckResult{
		Status:  HealthStatusCritical,
		Message: "Plugin is critical",
		Details: map[string]interface{}{
			"memory_status": "critical",
		},
	}
	
	// 测试重启恢复策略
	plugin.Start() // 确保插件已启动
	err := checker.Recover(ctx, "restart", result)
	if err != nil {
		t.Fatalf("Expected no error for restart recovery, got %v", err)
	}
	
	if !plugin.IsStarted() {
		t.Error("Expected plugin to be restarted")
	}
	
	// 测试GC恢复策略 - 使用已有的result，它有memory_status为"critical"
	err = checker.Recover(ctx, "gc", result)
	if err != nil {
		t.Fatalf("Expected no error for gc recovery, got %v", err)
	}
	
	// 测试重置恢复策略 - 创建适合reset策略的结果
	resetResult := &HealthCheckResult{
		Status:  HealthStatusUnhealthy,
		Message: "Plugin is unhealthy",
		Details: map[string]interface{}{
			"plugin_health": "failed",
		},
	}
	err = checker.Recover(ctx, "reset", resetResult)
	if err != nil {
		t.Fatalf("Expected no error for reset recovery, got %v", err)
	}
	
	// 测试不存在的恢复策略
	err = checker.Recover(ctx, "nonexistent", result)
	if err == nil {
		t.Error("Expected error for nonexistent recovery strategy")
	}
}

// TestHealthThresholds 测试健康阈值
func TestHealthThresholds(t *testing.T) {
	thresholds := DefaultExtendedHealthCheckConfig().Thresholds
	
	// 测试内存阈值
	status := thresholds.IsHealthy("memory_usage", int64(25*1024*1024)) // 25MB
	if status != HealthStatusHealthy {
		t.Errorf("Expected healthy status for low memory usage, got %v", status)
	}
	
	status = thresholds.IsHealthy("memory_usage", int64(75*1024*1024)) // 75MB
	if status != HealthStatusDegraded {
		t.Errorf("Expected degraded status for medium memory usage, got %v", status)
	}
	
	status = thresholds.IsHealthy("memory_usage", int64(150*1024*1024)) // 150MB
	if status != HealthStatusCritical {
		t.Errorf("Expected critical status for high memory usage, got %v", status)
	}
	
	// 测试CPU阈值
	status = thresholds.IsHealthy("cpu_usage", 50.0)
	if status != HealthStatusHealthy {
		t.Errorf("Expected healthy status for low CPU usage, got %v", status)
	}
	
	status = thresholds.IsHealthy("cpu_usage", 80.0)
	if status != HealthStatusDegraded {
		t.Errorf("Expected degraded status for medium CPU usage, got %v", status)
	}
	
	status = thresholds.IsHealthy("cpu_usage", 95.0)
	if status != HealthStatusCritical {
		t.Errorf("Expected critical status for high CPU usage, got %v", status)
	}
	
	// 测试响应时间阈值
	status = thresholds.IsHealthy("response_time", 500*time.Millisecond)
	if status != HealthStatusHealthy {
		t.Errorf("Expected healthy status for fast response time, got %v", status)
	}
	
	status = thresholds.IsHealthy("response_time", 2*time.Second)
	if status != HealthStatusDegraded {
		t.Errorf("Expected degraded status for slow response time, got %v", status)
	}
	
	status = thresholds.IsHealthy("response_time", 10*time.Second)
	if status != HealthStatusCritical {
		t.Errorf("Expected critical status for very slow response time, got %v", status)
	}
	
	// 测试错误率阈值
	status = thresholds.IsHealthy("error_rate", 2.0)
	if status != HealthStatusHealthy {
		t.Errorf("Expected healthy status for low error rate, got %v", status)
	}
	
	status = thresholds.IsHealthy("error_rate", 10.0)
	if status != HealthStatusDegraded {
		t.Errorf("Expected degraded status for medium error rate, got %v", status)
	}
	
	status = thresholds.IsHealthy("error_rate", 20.0)
	if status != HealthStatusCritical {
		t.Errorf("Expected critical status for high error rate, got %v", status)
	}
	
	// 测试协程数量阈值
	status = thresholds.IsHealthy("goroutines", 100)
	if status != HealthStatusHealthy {
		t.Errorf("Expected healthy status for low goroutine count, got %v", status)
	}
	
	status = thresholds.IsHealthy("goroutines", 750)
	if status != HealthStatusDegraded {
		t.Errorf("Expected degraded status for medium goroutine count, got %v", status)
	}
	
	status = thresholds.IsHealthy("goroutines", 1500)
	if status != HealthStatusCritical {
		t.Errorf("Expected critical status for high goroutine count, got %v", status)
	}
	
	// 测试堆使用率阈值
	status = thresholds.IsHealthy("heap_usage", 50.0)
	if status != HealthStatusHealthy {
		t.Errorf("Expected healthy status for low heap usage, got %v", status)
	}
	
	status = thresholds.IsHealthy("heap_usage", 80.0)
	if status != HealthStatusDegraded {
		t.Errorf("Expected degraded status for medium heap usage, got %v", status)
	}
	
	status = thresholds.IsHealthy("heap_usage", 95.0)
	if status != HealthStatusCritical {
		t.Errorf("Expected critical status for high heap usage, got %v", status)
	}
}

// TestHealthConfigManager 测试健康检查配置管理器
func TestHealthConfigManager(t *testing.T) {
	// 创建临时配置文件
	tempFile := "/tmp/test_health_config.json"
	defer os.Remove(tempFile)
	
	manager := NewHealthConfigManager(tempFile)
	
	// 测试加载默认配置
	err := manager.LoadConfig()
	if err != nil {
		t.Fatalf("Expected no error loading config, got %v", err)
	}
	
	config := manager.GetConfig()
	if config == nil {
		t.Fatal("Expected non-nil config")
	}
	
	if !config.Enabled {
		t.Error("Expected config to be enabled by default")
	}
	
	// 测试更新配置
	newConfig := *config
	newConfig.CheckInterval = 60 * time.Second
	
	err = manager.UpdateConfig(&newConfig)
	if err != nil {
		t.Fatalf("Expected no error updating config, got %v", err)
	}
	
	updatedConfig := manager.GetConfig()
	if updatedConfig.CheckInterval != 60*time.Second {
		t.Error("Expected config to be updated")
	}
	
	// 测试更新阈值
	newThresholds := *config.Thresholds
	newThresholds.MemoryWarning = 75 * 1024 * 1024
	
	err = manager.UpdateThresholds(&newThresholds)
	if err != nil {
		t.Fatalf("Expected no error updating thresholds, got %v", err)
	}
	
	updatedConfig = manager.GetConfig()
	if updatedConfig.Thresholds.MemoryWarning != 75*1024*1024 {
		t.Error("Expected thresholds to be updated")
	}
}

// TestConfigValidation 测试配置验证
func TestConfigValidation(t *testing.T) {
	manager := NewHealthConfigManager("")
	
	// 测试无效的检查间隔
	invalidConfig := DefaultExtendedHealthCheckConfig()
	invalidConfig.CheckInterval = -1 * time.Second
	
	err := manager.UpdateConfig(invalidConfig)
	if err == nil {
		t.Error("Expected error for invalid check interval")
	}
	
	// 测试无效的超时时间
	invalidConfig = DefaultExtendedHealthCheckConfig()
	invalidConfig.Timeout = -1 * time.Second
	
	err = manager.UpdateConfig(invalidConfig)
	if err == nil {
		t.Error("Expected error for invalid timeout")
	}
	
	// 测试无效的重试次数
	invalidConfig = DefaultExtendedHealthCheckConfig()
	invalidConfig.RetryCount = -1
	
	err = manager.UpdateConfig(invalidConfig)
	if err == nil {
		t.Error("Expected error for invalid retry count")
	}
	
	// 测试无效的阈值
	invalidThresholds := &ExtendedHealthThresholds{
		MemoryWarning:  100 * 1024 * 1024,
		MemoryCritical: 50 * 1024 * 1024, // 警告阈值大于临界阈值
	}
	
	err = manager.UpdateThresholds(invalidThresholds)
	if err == nil {
		t.Error("Expected error for invalid thresholds")
	}
}

// TestConcurrentHealthChecks 测试并发健康检查
func TestConcurrentHealthChecks(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	plugin := NewMockPlugin("test-plugin", "1.0.0")
	config := DefaultExtendedHealthCheckConfig()
	
	checker := NewDefaultHealthChecker(plugin, config, logger)
	ctx := context.Background()
	
	// 并发执行多个健康检查
	var wg sync.WaitGroup
	results := make([]*HealthCheckResult, 10)
	errors := make([]error, 10)
	
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			result, err := checker.CheckHealth(ctx)
			results[index] = result
			errors[index] = err
		}(i)
	}
	
	wg.Wait()
	
	// 验证所有检查都成功完成
	for i, err := range errors {
		if err != nil {
			t.Errorf("Expected no error for concurrent check %d, got %v", i, err)
		}
		
		if results[i] == nil {
			t.Errorf("Expected non-nil result for concurrent check %d", i)
		}
	}
}

// TestPerformanceMetricsCollector 测试性能指标收集器
func TestPerformanceMetricsCollector(t *testing.T) {
	plugin := NewMockPlugin("test-plugin", "1.0.0")
	collector := &PerformanceMetricsCollector{
		plugin:    plugin,
		startTime: time.Now(),
		stats:     make(map[string]*PerformanceStats),
	}
	
	// 记录一些请求
	collector.RecordRequest("test_operation", 100*time.Millisecond, true)
	collector.RecordRequest("test_operation", 200*time.Millisecond, true)
	collector.RecordRequest("test_operation", 150*time.Millisecond, false)
	
	// 收集指标
	ctx := context.Background()
	metrics, err := collector.Collect(ctx)
	if err != nil {
		t.Fatalf("Expected no error collecting metrics, got %v", err)
	}
	
	if metrics == nil {
		t.Fatal("Expected non-nil metrics")
	}
	
	// 验证指标
	if totalRequests, ok := metrics["total_requests"].(int64); !ok || totalRequests != 3 {
		t.Errorf("Expected 3 total requests, got %v", totalRequests)
	}
	
	if successRequests, ok := metrics["success_requests"].(int64); !ok || successRequests != 2 {
		t.Errorf("Expected 2 success requests, got %v", successRequests)
	}
	
	if errorRequests, ok := metrics["error_requests"].(int64); !ok || errorRequests != 1 {
		t.Errorf("Expected 1 error request, got %v", errorRequests)
	}
	
	if errorRate, ok := metrics["error_rate"].(float64); !ok || errorRate < 0.33 || errorRate > 0.34 {
		t.Errorf("Expected error rate around 0.33, got %v", errorRate)
	}
}

// BenchmarkHealthCheck 健康检查性能基准测试
func BenchmarkHealthCheck(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	plugin := NewMockPlugin("test-plugin", "1.0.0")
	config := DefaultExtendedHealthCheckConfig()
	
	checker := NewDefaultHealthChecker(plugin, config, logger)
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := checker.CheckHealth(ctx)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

// BenchmarkMetricsCollection 指标收集性能基准测试
func BenchmarkMetricsCollection(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	plugin := NewMockPlugin("test-plugin", "1.0.0")
	config := DefaultExtendedHealthCheckConfig()
	
	checker := NewDefaultHealthChecker(plugin, config, logger)
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := checker.CollectMetrics(ctx, "system")
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}