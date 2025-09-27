package integration

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/kernel"
	"github.com/go-musicfox/go-musicfox/v2/test/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// PerformanceStabilityTestSuite 性能和稳定性测试套件
type PerformanceStabilityTestSuite struct {
	suite.Suite
	kernel        kernel.Kernel
	pluginManager kernel.PluginManager
	ctx           context.Context
	cancel        context.CancelFunc

	// 性能监控
	initialMemStats runtime.MemStats
	finalMemStats   runtime.MemStats
}

// SetupSuite 设置测试套件
func (suite *PerformanceStabilityTestSuite) SetupSuite() {
	suite.ctx, suite.cancel = context.WithTimeout(context.Background(), 60*time.Second) // 1分钟超时
}

// TearDownSuite 清理测试套件
func (suite *PerformanceStabilityTestSuite) TearDownSuite() {
	if suite.cancel != nil {
		suite.cancel()
	}
}

// SetupTest 设置每个测试
func (suite *PerformanceStabilityTestSuite) SetupTest() {
	// 记录初始内存状态
	runtime.GC()
	runtime.ReadMemStats(&suite.initialMemStats)

	// 创建并初始化微内核
	suite.kernel = kernel.NewMicroKernel()
	err := suite.kernel.Initialize(suite.ctx)
	suite.Require().NoError(err)
	err = suite.kernel.Start(suite.ctx)
	suite.Require().NoError(err)

	// 获取插件管理器
	suite.pluginManager = suite.kernel.GetPluginManager()
	suite.Require().NotNil(suite.pluginManager)
}

// TearDownTest 清理每个测试
func (suite *PerformanceStabilityTestSuite) TearDownTest() {
	if suite.kernel != nil {
		_ = suite.kernel.Shutdown(suite.ctx)
	}

	// 记录最终内存状态
	runtime.GC()
	runtime.ReadMemStats(&suite.finalMemStats)

	// 检查内存泄漏
	suite.checkMemoryLeak()
}

// checkMemoryLeak 检查内存泄漏
func (suite *PerformanceStabilityTestSuite) checkMemoryLeak() {
	memoryIncrease := suite.finalMemStats.Alloc - suite.initialMemStats.Alloc
	heapIncrease := suite.finalMemStats.HeapAlloc - suite.initialMemStats.HeapAlloc

	suite.T().Logf("Memory usage - Initial: %d bytes, Final: %d bytes, Increase: %d bytes",
		suite.initialMemStats.Alloc, suite.finalMemStats.Alloc, memoryIncrease)
	suite.T().Logf("Heap usage - Initial: %d bytes, Final: %d bytes, Increase: %d bytes",
		suite.initialMemStats.HeapAlloc, suite.finalMemStats.HeapAlloc, heapIncrease)

	// 允许一定的内存增长（10MB），但不应该有严重的内存泄漏
	const maxAllowedIncrease = 10 * 1024 * 1024 // 10MB
	if memoryIncrease > maxAllowedIncrease {
		suite.T().Logf("WARNING: Potential memory leak detected. Memory increased by %d bytes", memoryIncrease)
	}
}

// TestConcurrentPluginLoading 测试并发插件加载
func (suite *PerformanceStabilityTestSuite) TestConcurrentPluginLoading() {
	const numPlugins = 50
	const numGoroutines = 10

	suite.T().Logf("Testing concurrent loading of %d plugins with %d goroutines", numPlugins, numGoroutines)

	// 创建插件
	plugins := make([]*fixtures.MockPlugin, numPlugins)
	for i := 0; i < numPlugins; i++ {
		plugins[i] = fixtures.NewMockPlugin(fmt.Sprintf("concurrent-plugin-%d", i), "1.0.0")
	}

	// 并发加载插件
	start := time.Now()
	done := make(chan error, numGoroutines)
	pluginChan := make(chan *fixtures.MockPlugin, numPlugins)

	// 将插件放入通道
	for _, plugin := range plugins {
		pluginChan <- plugin
	}
	close(pluginChan)

	// 启动工作协程
	for i := 0; i < numGoroutines; i++ {
		go func(workerID int) {
			defer func() { done <- nil }()

			for plugin := range pluginChan {
				// 注册插件
				if err := suite.pluginManager.RegisterPlugin(plugin); err != nil {
					done <- fmt.Errorf("worker %d failed to register plugin %s: %w", workerID, plugin.GetInfo().Name, err)
					return
				}

				// 启动插件
				if err := suite.pluginManager.StartPlugin(plugin.GetInfo().Name); err != nil {
					done <- fmt.Errorf("worker %d failed to start plugin %s: %w", workerID, plugin.GetInfo().Name, err)
					return
				}
			}
		}(i)
	}

	// 等待所有工作协程完成
	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-done:
			assert.NoError(suite.T(), err)
		case <-time.After(30 * time.Second):
			suite.T().Fatal("Concurrent plugin loading timeout")
		}
	}

	loadingTime := time.Since(start)
	suite.T().Logf("Concurrent plugin loading completed in %v", loadingTime)

	// 验证所有插件都已加载
	loadedCount := suite.pluginManager.GetLoadedPluginCount()
	assert.Equal(suite.T(), numPlugins, loadedCount)

	// 验证所有插件都在运行
	for _, plugin := range plugins {
		assert.True(suite.T(), plugin.IsRunning())
		assert.True(suite.T(), plugin.IsStartCalled())
	}

	// 性能基准：加载时间不应超过合理范围
	expectedMaxTime := time.Duration(numPlugins) * 10 * time.Millisecond // 每个插件最多10ms
	if loadingTime > expectedMaxTime {
		suite.T().Logf("WARNING: Plugin loading took longer than expected: %v > %v", loadingTime, expectedMaxTime)
	}
}

// TestHighFrequencyOperations 测试高频操作
func (suite *PerformanceStabilityTestSuite) TestHighFrequencyOperations() {
	const numOperations = 1000
	const numPlugins = 10

	suite.T().Logf("Testing %d high-frequency operations on %d plugins", numOperations, numPlugins)

	// 创建并注册插件
	plugins := make([]*fixtures.MockAudioProcessorPlugin, numPlugins)
	for i := 0; i < numPlugins; i++ {
		plugins[i] = fixtures.NewMockAudioProcessorPlugin()
		plugins[i].GetInfo().Name = fmt.Sprintf("audio-plugin-%d", i)

		err := suite.pluginManager.RegisterPlugin(plugins[i])
		assert.NoError(suite.T(), err)
		err = suite.pluginManager.StartPlugin(plugins[i].GetInfo().Name)
		assert.NoError(suite.T(), err)
	}

	// 高频音频处理操作
	start := time.Now()
	audioData := make([]byte, 1024) // 1KB音频数据
	for i := 0; i < len(audioData); i++ {
		audioData[i] = byte(i % 256)
	}

	var wg sync.WaitGroup
	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(opIndex int) {
			defer wg.Done()

			pluginIndex := opIndex % numPlugins
			plugin := plugins[pluginIndex]

			// 音频处理
			_, err := plugin.ProcessAudio(audioData, 44100, 2)
			assert.NoError(suite.T(), err)

			// 音量调节
			volume := float64(opIndex%100) / 100.0 // 0.0 到 0.99
			_, err = plugin.AdjustVolume(audioData, volume)
			assert.NoError(suite.T(), err)

			// 音效应用
			effects := []string{"reverb", "echo", "chorus"}
			effect := effects[opIndex%len(effects)]
			_, err = plugin.ApplyEffect(audioData, effect)
			assert.NoError(suite.T(), err)
		}(i)
	}

	wg.Wait()
	processingTime := time.Since(start)

	suite.T().Logf("High-frequency operations completed in %v", processingTime)
	suite.T().Logf("Average operation time: %v", processingTime/time.Duration(numOperations))

	// 验证操作统计
	totalProcessCount := 0
	totalVolumeAdjustments := 0
	totalEffectsApplied := 0

	for _, plugin := range plugins {
		totalProcessCount += plugin.GetProcessCount()
		totalVolumeAdjustments += plugin.GetVolumeAdjustments()
		totalEffectsApplied += plugin.GetEffectsApplied()
	}

	assert.Equal(suite.T(), numOperations, totalProcessCount)
	assert.Equal(suite.T(), numOperations, totalVolumeAdjustments)
	assert.Equal(suite.T(), numOperations, totalEffectsApplied)

	// 性能基准：平均操作时间不应超过1ms
	avgOpTime := processingTime / time.Duration(numOperations)
	if avgOpTime > time.Millisecond {
		suite.T().Logf("WARNING: Average operation time is high: %v", avgOpTime)
	}
}

// TestLongRunningStability 测试长时间运行稳定性
func (suite *PerformanceStabilityTestSuite) TestLongRunningStability() {
	const testDuration = 10 * time.Second // 10秒测试，避免超时
	const checkInterval = 2 * time.Second
	const numPlugins = 5

	suite.T().Logf("Testing long-running stability for %v with %d plugins", testDuration, numPlugins)

	// 创建并启动插件
	plugins := make([]*fixtures.MockPlugin, numPlugins)
	for i := 0; i < numPlugins; i++ {
		plugins[i] = fixtures.NewMockPlugin(fmt.Sprintf("stability-plugin-%d", i), "1.0.0")
		err := suite.pluginManager.RegisterPlugin(plugins[i])
		assert.NoError(suite.T(), err)
		err = suite.pluginManager.StartPlugin(plugins[i].GetInfo().Name)
		assert.NoError(suite.T(), err)
	}

	// 长时间运行测试
	start := time.Now()
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	healthCheckCount := 0
	configUpdateCount := 0

	for {
		select {
		case <-ticker.C:
			// 定期健康检查
			for _, plugin := range plugins {
				err := plugin.HealthCheck()
				assert.NoError(suite.T(), err)
				assert.True(suite.T(), plugin.IsRunning())
			}
			healthCheckCount++

			// 定期配置更新
			for i, plugin := range plugins {
				config := map[string]interface{}{
					"required_key": fmt.Sprintf("value-%d-%d", i, configUpdateCount),
					"timestamp":    time.Now().Unix(),
				}
				err := plugin.UpdateConfig(config)
				assert.NoError(suite.T(), err)
			}
			configUpdateCount++

			// 检查内核状态
			assert.True(suite.T(), suite.kernel.IsRunning())
			status := suite.kernel.GetStatus()
			assert.Equal(suite.T(), kernel.KernelStateRunning, status.State)

			// 记录运行时间
			elapsed := time.Since(start)
			suite.T().Logf("Stability test running for %v, health checks: %d, config updates: %d",
				elapsed, healthCheckCount, configUpdateCount)

			if elapsed >= testDuration {
				goto TestComplete
			}

		case <-suite.ctx.Done():
			suite.T().Fatal("Long-running stability test timeout")
		}
	}

TestComplete:
	totalRunTime := time.Since(start)
	suite.T().Logf("Long-running stability test completed successfully after %v", totalRunTime)

	// 验证最终状态
	for _, plugin := range plugins {
		assert.True(suite.T(), plugin.IsRunning())
		assert.True(suite.T(), plugin.GetHealthCheckCount() > 0)
		assert.True(suite.T(), plugin.GetConfigUpdates() > 0)
	}

	// 验证内核状态
	assert.True(suite.T(), suite.kernel.IsRunning())
	assert.Equal(suite.T(), numPlugins, suite.pluginManager.GetLoadedPluginCount())
}

// TestMemoryUsageUnderLoad 测试负载下的内存使用
func (suite *PerformanceStabilityTestSuite) TestMemoryUsageUnderLoad() {
	const numPlugins = 20
	const numOperations = 500
	const dataSize = 10 * 1024 // 10KB per operation

	suite.T().Logf("Testing memory usage under load: %d plugins, %d operations, %d bytes per operation",
		numPlugins, numOperations, dataSize)

	// 记录初始内存
	var initialMem runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&initialMem)

	// 创建插件
	plugins := make([]*fixtures.MockAudioProcessorPlugin, numPlugins)
	for i := 0; i < numPlugins; i++ {
		plugins[i] = fixtures.NewMockAudioProcessorPlugin()
		plugins[i].GetInfo().Name = fmt.Sprintf("memory-test-plugin-%d", i)

		err := suite.pluginManager.RegisterPlugin(plugins[i])
		assert.NoError(suite.T(), err)
		err = suite.pluginManager.StartPlugin(plugins[i].GetInfo().Name)
		assert.NoError(suite.T(), err)
	}

	// 记录插件加载后的内存
	var afterLoadMem runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&afterLoadMem)

	// 执行大量操作
	largeData := make([]byte, dataSize)
	for i := 0; i < len(largeData); i++ {
		largeData[i] = byte(i % 256)
	}

	var wg sync.WaitGroup
	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(opIndex int) {
			defer wg.Done()

			pluginIndex := opIndex % numPlugins
			plugin := plugins[pluginIndex]

			// 处理大量数据
			_, err := plugin.ProcessAudio(largeData, 44100, 2)
			assert.NoError(suite.T(), err)

			_, err = plugin.AdjustVolume(largeData, 0.5)
			assert.NoError(suite.T(), err)

			_, err = plugin.ApplyEffect(largeData, "reverb")
			assert.NoError(suite.T(), err)
		}(i)
	}

	wg.Wait()

	// 记录操作后的内存
	var afterOpsMem runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&afterOpsMem)

	// 分析内存使用
	loadMemIncrease := afterLoadMem.Alloc - initialMem.Alloc
	opsMemIncrease := afterOpsMem.Alloc - afterLoadMem.Alloc
	totalMemIncrease := afterOpsMem.Alloc - initialMem.Alloc

	suite.T().Logf("Memory usage analysis:")
	suite.T().Logf("  Initial: %d bytes", initialMem.Alloc)
	suite.T().Logf("  After loading plugins: %d bytes (+%d)", afterLoadMem.Alloc, loadMemIncrease)
	suite.T().Logf("  After operations: %d bytes (+%d)", afterOpsMem.Alloc, opsMemIncrease)
	suite.T().Logf("  Total increase: %d bytes", totalMemIncrease)
	suite.T().Logf("  Memory per plugin: %d bytes", loadMemIncrease/uint64(numPlugins))
	suite.T().Logf("  Memory per operation: %d bytes", opsMemIncrease/uint64(numOperations))

	// 内存使用合理性检查
	maxExpectedIncrease := uint64(numPlugins*1024*1024 + numOperations*1024) // 1MB per plugin + 1KB per operation
	if totalMemIncrease > maxExpectedIncrease {
		suite.T().Logf("WARNING: Memory usage higher than expected: %d > %d", totalMemIncrease, maxExpectedIncrease)
	}

	// 验证操作完成
	totalOps := 0
	for _, plugin := range plugins {
		totalOps += plugin.GetProcessCount()
	}
	assert.Equal(suite.T(), numOperations, totalOps)
}

// TestGoroutineLeakDetection 测试协程泄漏检测
func (suite *PerformanceStabilityTestSuite) TestGoroutineLeakDetection() {
	// 记录初始协程数量
	initialGoroutines := runtime.NumGoroutine()
	suite.T().Logf("Initial goroutines: %d", initialGoroutines)

	// 创建并启动多个插件
	const numPlugins = 10
	plugins := make([]*fixtures.MockPlugin, numPlugins)

	for i := 0; i < numPlugins; i++ {
		plugins[i] = fixtures.NewMockPlugin(fmt.Sprintf("goroutine-test-plugin-%d", i), "1.0.0")
		err := suite.pluginManager.RegisterPlugin(plugins[i])
		assert.NoError(suite.T(), err)
		err = suite.pluginManager.StartPlugin(plugins[i].GetInfo().Name)
		assert.NoError(suite.T(), err)
	}

	// 记录插件启动后的协程数量
	afterStartGoroutines := runtime.NumGoroutine()
	suite.T().Logf("Goroutines after starting plugins: %d (+%d)", afterStartGoroutines, afterStartGoroutines-initialGoroutines)

	// 执行一些操作
	for i := 0; i < 100; i++ {
		for _, plugin := range plugins {
			err := plugin.HealthCheck()
			assert.NoError(suite.T(), err)
		}
	}

	// 停止所有插件
	for _, plugin := range plugins {
		err := suite.pluginManager.StopPlugin(plugin.GetInfo().Name)
		assert.NoError(suite.T(), err)
	}

	// 注销所有插件
	for _, plugin := range plugins {
		err := suite.pluginManager.UnregisterPlugin(plugin.GetInfo().Name)
		assert.NoError(suite.T(), err)
	}

	// 等待协程清理
	time.Sleep(100 * time.Millisecond)
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	// 记录最终协程数量
	finalGoroutines := runtime.NumGoroutine()
	suite.T().Logf("Final goroutines: %d", finalGoroutines)

	// 检查协程泄漏
	goroutineIncrease := finalGoroutines - initialGoroutines
	if goroutineIncrease > 5 { // 允许少量协程增长
		suite.T().Logf("WARNING: Potential goroutine leak detected. Goroutines increased by %d", goroutineIncrease)
		
		// 打印协程堆栈信息用于调试
		buf := make([]byte, 1<<16)
		stackSize := runtime.Stack(buf, true)
		suite.T().Logf("Goroutine stack trace:\n%s", buf[:stackSize])
	} else {
		suite.T().Logf("No significant goroutine leak detected. Increase: %d goroutines", goroutineIncrease)
	}
}

// TestPerformanceStability 运行性能和稳定性测试
func TestPerformanceStability(t *testing.T) {
	suite.Run(t, new(PerformanceStabilityTestSuite))
}