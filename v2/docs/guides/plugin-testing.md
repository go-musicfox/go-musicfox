# 插件测试指南

## 概述

本指南详细介绍 go-musicfox v2 插件的测试策略、测试方法和最佳实践，帮助开发者编写高质量的插件测试代码。

## 测试策略

### 测试金字塔

```
        /\        E2E 测试 (10%)
       /  \       - 端到端功能测试
      /    \      - 用户场景测试
     /______\     
    /        \    集成测试 (20%)
   /          \   - 插件与内核集成
  /            \  - 插件间交互测试
 /______________\ 
/                \ 单元测试 (70%)
\________________/ - 函数级别测试
                   - 组件隔离测试
```

### 测试分类

1. **单元测试**：测试插件的单个函数或方法
2. **集成测试**：测试插件与系统其他组件的交互
3. **性能测试**：测试插件的性能表现
4. **安全测试**：测试插件的安全性
5. **兼容性测试**：测试插件在不同环境下的兼容性

## 单元测试

### 基础测试结构

```go
// plugin_test.go
package main

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/mock"
)

func TestPluginInfo(t *testing.T) {
    plugin := &MyPlugin{}
    info := plugin.GetInfo()
    
    assert.Equal(t, "my-plugin", info.Name)
    assert.Equal(t, "1.0.0", info.Version)
    assert.NotEmpty(t, info.Description)
}

func TestPluginInitialization(t *testing.T) {
    plugin := &MyPlugin{}
    ctx := &MockPluginContext{}
    
    err := plugin.Initialize(ctx)
    require.NoError(t, err)
    
    // 验证初始化状态
    assert.True(t, plugin.isInitialized)
    assert.NotNil(t, plugin.logger)
}

func TestPluginLifecycle(t *testing.T) {
    plugin := &MyPlugin{}
    ctx := &MockPluginContext{}
    
    // 测试完整生命周期
    err := plugin.Initialize(ctx)
    require.NoError(t, err)
    
    err = plugin.Start()
    require.NoError(t, err)
    assert.True(t, plugin.isRunning)
    
    err = plugin.Stop()
    require.NoError(t, err)
    assert.False(t, plugin.isRunning)
    
    err = plugin.Cleanup()
    require.NoError(t, err)
}
```

### 模拟对象 (Mocks)

```go
// mocks.go
package main

import (
    "github.com/stretchr/testify/mock"
    "github.com/go-musicfox/go-musicfox/v2/pkg/plugin"
)

// MockPluginContext 模拟插件上下文
type MockPluginContext struct {
    mock.Mock
}

func (m *MockPluginContext) GetLogger() plugin.Logger {
    args := m.Called()
    return args.Get(0).(plugin.Logger)
}

func (m *MockPluginContext) GetEventBus() plugin.EventBus {
    args := m.Called()
    return args.Get(0).(plugin.EventBus)
}

func (m *MockPluginContext) GetServiceRegistry() plugin.ServiceRegistry {
    args := m.Called()
    return args.Get(0).(plugin.ServiceRegistry)
}

// MockEventBus 模拟事件总线
type MockEventBus struct {
    mock.Mock
}

func (m *MockEventBus) Publish(eventType string, data interface{}) error {
    args := m.Called(eventType, data)
    return args.Error(0)
}

func (m *MockEventBus) Subscribe(eventType string, handler plugin.EventHandler) error {
    args := m.Called(eventType, handler)
    return args.Error(0)
}

// MockLogger 模拟日志器
type MockLogger struct {
    mock.Mock
}

func (m *MockLogger) Info(msg string, args ...interface{}) {
    m.Called(append([]interface{}{msg}, args...)...)
}

func (m *MockLogger) Error(msg string, args ...interface{}) {
    m.Called(append([]interface{}{msg}, args...)...)
}
```

### 使用模拟对象的测试

```go
func TestPluginWithMocks(t *testing.T) {
    // 创建模拟对象
    mockCtx := &MockPluginContext{}
    mockLogger := &MockLogger{}
    mockEventBus := &MockEventBus{}
    
    // 设置期望行为
    mockCtx.On("GetLogger").Return(mockLogger)
    mockCtx.On("GetEventBus").Return(mockEventBus)
    mockLogger.On("Info", "Plugin initialized", mock.Anything)
    mockEventBus.On("Publish", "plugin.initialized", mock.Anything).Return(nil)
    
    // 执行测试
    plugin := &MyPlugin{}
    err := plugin.Initialize(mockCtx)
    
    // 验证结果
    require.NoError(t, err)
    
    // 验证模拟对象的调用
    mockCtx.AssertExpectations(t)
    mockLogger.AssertExpectations(t)
    mockEventBus.AssertExpectations(t)
}
```

### 表格驱动测试

```go
func TestPluginConfigValidation(t *testing.T) {
    tests := []struct {
        name        string
        config      map[string]interface{}
        expectError bool
        errorMsg    string
    }{
        {
            name: "valid config",
            config: map[string]interface{}{
                "sample_rate": 44100,
                "buffer_size": 1024,
                "enabled":     true,
            },
            expectError: false,
        },
        {
            name: "missing sample_rate",
            config: map[string]interface{}{
                "buffer_size": 1024,
                "enabled":     true,
            },
            expectError: true,
            errorMsg:    "missing required config: sample_rate",
        },
        {
            name: "invalid sample_rate",
            config: map[string]interface{}{
                "sample_rate": 1000, // 太低
                "buffer_size": 1024,
                "enabled":     true,
            },
            expectError: true,
            errorMsg:    "invalid sample_rate",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            plugin := &MyPlugin{}
            err := plugin.ValidateConfig(tt.config)
            
            if tt.expectError {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorMsg)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

## 集成测试

### 插件与内核集成测试

```go
// integration_test.go
package main

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/go-musicfox/go-musicfox/v2/pkg/kernel"
    "github.com/go-musicfox/go-musicfox/v2/pkg/plugin"
)

func TestPluginIntegrationWithKernel(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // 创建测试内核
    k := kernel.NewMicroKernel()
    ctx := context.Background()
    
    // 初始化内核
    err := k.Initialize(ctx)
    require.NoError(t, err)
    
    // 启动内核
    err = k.Start(ctx)
    require.NoError(t, err)
    defer k.Stop(ctx)
    
    // 获取插件管理器
    pluginManager := k.GetPluginManager()
    
    // 加载插件
    err = pluginManager.LoadPlugin("./test-plugin.so", plugin.TypeDynamicLibrary)
    require.NoError(t, err)
    
    // 启动插件
    err = pluginManager.StartPlugin("test-plugin")
    require.NoError(t, err)
    
    // 验证插件状态
    pluginInfo, err := pluginManager.GetPlugin("test-plugin")
    require.NoError(t, err)
    assert.Equal(t, "test-plugin", pluginInfo.GetInfo().Name)
    
    // 测试插件功能
    // ...
    
    // 停止插件
    err = pluginManager.StopPlugin("test-plugin")
    require.NoError(t, err)
    
    // 卸载插件
    err = pluginManager.UnloadPlugin("test-plugin")
    require.NoError(t, err)
}
```

### 插件间通信测试

```go
func TestPluginCommunication(t *testing.T) {
    // 创建测试环境
    eventBus := event.NewEventBus()
    
    // 创建两个插件
    plugin1 := &ProducerPlugin{}
    plugin2 := &ConsumerPlugin{}
    
    // 初始化插件
    ctx1 := &TestPluginContext{eventBus: eventBus}
    ctx2 := &TestPluginContext{eventBus: eventBus}
    
    err := plugin1.Initialize(ctx1)
    require.NoError(t, err)
    
    err = plugin2.Initialize(ctx2)
    require.NoError(t, err)
    
    // 启动插件
    err = plugin1.Start()
    require.NoError(t, err)
    
    err = plugin2.Start()
    require.NoError(t, err)
    
    // 等待插件就绪
    time.Sleep(100 * time.Millisecond)
    
    // 测试事件通信
    testData := map[string]interface{}{
        "message": "hello world",
        "timestamp": time.Now(),
    }
    
    // 发布事件
    err = plugin1.PublishTestEvent(testData)
    require.NoError(t, err)
    
    // 等待事件处理
    time.Sleep(100 * time.Millisecond)
    
    // 验证事件接收
    receivedData := plugin2.GetLastReceivedData()
    assert.Equal(t, testData["message"], receivedData["message"])
}
```

## 性能测试

### 基准测试

```go
// benchmark_test.go
package main

import (
    "testing"
)

func BenchmarkPluginProcessAudio(b *testing.B) {
    plugin := &AudioPlugin{}
    ctx := &TestPluginContext{}
    
    err := plugin.Initialize(ctx)
    if err != nil {
        b.Fatal(err)
    }
    
    // 准备测试数据
    audioData := make([]byte, 1024*4) // 1024 samples, 4 bytes each
    for i := range audioData {
        audioData[i] = byte(i % 256)
    }
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _, err := plugin.ProcessAudio(audioData, 44100, 2)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkPluginMemoryAllocation(b *testing.B) {
    plugin := &AudioPlugin{}
    
    b.ReportAllocs()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        data := plugin.AllocateBuffer(1024)
        plugin.ReleaseBuffer(data)
    }
}

func BenchmarkPluginConcurrency(b *testing.B) {
    plugin := &AudioPlugin{}
    ctx := &TestPluginContext{}
    
    err := plugin.Initialize(ctx)
    if err != nil {
        b.Fatal(err)
    }
    
    audioData := make([]byte, 1024*4)
    
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            _, err := plugin.ProcessAudio(audioData, 44100, 2)
            if err != nil {
                b.Fatal(err)
            }
        }
    })
}
```

### 性能分析

```go
func TestPluginPerformanceProfile(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping performance test in short mode")
    }
    
    plugin := &AudioPlugin{}
    ctx := &TestPluginContext{}
    
    err := plugin.Initialize(ctx)
    require.NoError(t, err)
    
    // 性能测试参数
    const (
        sampleRate = 44100
        channels   = 2
        duration   = 10 * time.Second
        bufferSize = 1024
    )
    
    audioData := make([]byte, bufferSize*channels*4)
    
    start := time.Now()
    processed := 0
    
    for time.Since(start) < duration {
        _, err := plugin.ProcessAudio(audioData, sampleRate, channels)
        require.NoError(t, err)
        processed++
    }
    
    elapsed := time.Since(start)
    throughput := float64(processed) / elapsed.Seconds()
    
    t.Logf("Processed %d buffers in %v", processed, elapsed)
    t.Logf("Throughput: %.2f buffers/second", throughput)
    
    // 验证性能要求
    minThroughput := 1000.0 // 最少每秒处理1000个缓冲区
    assert.Greater(t, throughput, minThroughput, "Plugin throughput below minimum requirement")
}
```

### 内存泄漏测试

```go
func TestPluginMemoryLeak(t *testing.T) {
    plugin := &AudioPlugin{}
    ctx := &TestPluginContext{}
    
    err := plugin.Initialize(ctx)
    require.NoError(t, err)
    
    // 获取初始内存使用
    var m1, m2 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // 执行大量操作
    audioData := make([]byte, 1024*4)
    for i := 0; i < 10000; i++ {
        _, err := plugin.ProcessAudio(audioData, 44100, 2)
        require.NoError(t, err)
    }
    
    // 强制垃圾回收
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    // 检查内存增长
    memGrowth := m2.Alloc - m1.Alloc
    t.Logf("Memory growth: %d bytes", memGrowth)
    
    // 验证内存增长在合理范围内
    maxGrowth := uint64(1024 * 1024) // 1MB
    assert.Less(t, memGrowth, maxGrowth, "Potential memory leak detected")
}
```

## 安全测试

### 权限测试

```go
func TestPluginPermissions(t *testing.T) {
    plugin := &MyPlugin{}
    securityManager := &MockSecurityManager{}
    
    // 测试无权限访问
    securityManager.On("CheckPermission", "my-plugin", "file-access").Return(errors.New("permission denied"))
    
    err := plugin.AccessFile("/etc/passwd")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "permission denied")
    
    // 测试有权限访问
    securityManager.ExpectedCalls = nil
    securityManager.On("CheckPermission", "my-plugin", "file-access").Return(nil)
    
    err = plugin.AccessFile("/tmp/test.txt")
    assert.NoError(t, err)
}
```

### 沙箱测试

```go
func TestPluginSandbox(t *testing.T) {
    plugin := &MyPlugin{}
    
    // 测试沙箱限制
    tests := []struct {
        name        string
        operation   func() error
        expectError bool
    }{
        {
            name:        "allowed file access",
            operation:   func() error { return plugin.ReadFile("./data/test.txt") },
            expectError: false,
        },
        {
            name:        "forbidden file access",
            operation:   func() error { return plugin.ReadFile("/etc/passwd") },
            expectError: true,
        },
        {
            name:        "allowed network access",
            operation:   func() error { return plugin.HTTPRequest("https://api.example.com") },
            expectError: false,
        },
        {
            name:        "forbidden network access",
            operation:   func() error { return plugin.HTTPRequest("https://malicious.com") },
            expectError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.operation()
            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## 兼容性测试

### 版本兼容性测试

```go
func TestPluginVersionCompatibility(t *testing.T) {
    tests := []struct {
        pluginVersion string
        kernelVersion string
        compatible    bool
    }{
        {"1.0.0", "2.0.0", true},
        {"1.5.0", "2.0.0", true},
        {"2.0.0", "1.9.0", false},
        {"2.1.0", "2.0.0", false},
    }
    
    for _, tt := range tests {
        t.Run(fmt.Sprintf("%s-%s", tt.pluginVersion, tt.kernelVersion), func(t *testing.T) {
            compatible := checkVersionCompatibility(tt.pluginVersion, tt.kernelVersion)
            assert.Equal(t, tt.compatible, compatible)
        })
    }
}
```

### 平台兼容性测试

```go
// +build integration

func TestPluginCrossPlatform(t *testing.T) {
    platforms := []struct {
        goos   string
        goarch string
    }{
        {"linux", "amd64"},
        {"linux", "arm64"},
        {"darwin", "amd64"},
        {"darwin", "arm64"},
        {"windows", "amd64"},
    }
    
    for _, platform := range platforms {
        t.Run(fmt.Sprintf("%s-%s", platform.goos, platform.goarch), func(t *testing.T) {
            // 构建插件
            cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", "plugin.so")
            cmd.Env = append(os.Environ(),
                fmt.Sprintf("GOOS=%s", platform.goos),
                fmt.Sprintf("GOARCH=%s", platform.goarch),
            )
            
            err := cmd.Run()
            if platform.goos == "windows" {
                // Windows 不支持 plugin 模式
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## 测试工具和辅助函数

### 测试辅助函数

```go
// test_helpers.go
package main

import (
    "os"
    "path/filepath"
    "testing"
    "time"
)

// CreateTempPlugin 创建临时插件文件
func CreateTempPlugin(t *testing.T, content []byte) string {
    tempDir := t.TempDir()
    pluginPath := filepath.Join(tempDir, "plugin.so")
    
    err := os.WriteFile(pluginPath, content, 0644)
    if err != nil {
        t.Fatal(err)
    }
    
    return pluginPath
}

// WaitForCondition 等待条件满足
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, message string) {
    deadline := time.Now().Add(timeout)
    
    for time.Now().Before(deadline) {
        if condition() {
            return
        }
        time.Sleep(10 * time.Millisecond)
    }
    
    t.Fatal(message)
}

// AssertEventually 断言条件最终满足
func AssertEventually(t *testing.T, condition func() bool, timeout time.Duration, message string) {
    WaitForCondition(t, condition, timeout, message)
}

// CreateTestAudioData 创建测试音频数据
func CreateTestAudioData(samples int, channels int) []byte {
    data := make([]byte, samples*channels*4) // 32-bit float
    
    for i := 0; i < len(data); i += 4 {
        // 生成正弦波
        sample := float32(0.5 * math.Sin(2*math.Pi*440*float64(i/4)/44100))
        binary.LittleEndian.PutUint32(data[i:], math.Float32bits(sample))
    }
    
    return data
}
```

### 测试配置

```go
// test_config.go
package main

import (
    "github.com/go-musicfox/go-musicfox/v2/pkg/plugin"
)

// TestPluginContext 测试用插件上下文
type TestPluginContext struct {
    logger          plugin.Logger
    eventBus        plugin.EventBus
    serviceRegistry plugin.ServiceRegistry
    config          map[string]interface{}
}

func NewTestPluginContext() *TestPluginContext {
    return &TestPluginContext{
        logger:          &TestLogger{},
        eventBus:        &TestEventBus{},
        serviceRegistry: &TestServiceRegistry{},
        config:          make(map[string]interface{}),
    }
}

func (ctx *TestPluginContext) GetLogger() plugin.Logger {
    return ctx.logger
}

func (ctx *TestPluginContext) GetEventBus() plugin.EventBus {
    return ctx.eventBus
}

func (ctx *TestPluginContext) GetServiceRegistry() plugin.ServiceRegistry {
    return ctx.serviceRegistry
}

func (ctx *TestPluginContext) GetConfig(key string) interface{} {
    return ctx.config[key]
}

func (ctx *TestPluginContext) SetConfig(key string, value interface{}) {
    ctx.config[key] = value
}
```

## 测试自动化

### Makefile 测试目标

```makefile
# Makefile
.PHONY: test test-unit test-integration test-benchmark test-coverage

# 运行所有测试
test: test-unit test-integration

# 单元测试
test-unit:
	@echo "Running unit tests..."
	go test -v -short ./...

# 集成测试
test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration ./...

# 基准测试
test-benchmark:
	@echo "Running benchmark tests..."
	go test -v -bench=. -benchmem ./...

# 测试覆盖率
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 竞态条件检测
test-race:
	@echo "Running tests with race detection..."
	go test -v -race ./...

# 内存泄漏检测
test-leak:
	@echo "Running tests with leak detection..."
	go test -v -tags=leak ./...
```

### GitHub Actions 配置

```yaml
# .github/workflows/test.yml
name: Test

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: [1.21, 1.22]
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: ${{ matrix.go-version }}
    
    - name: Cache dependencies
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run unit tests
      run: make test-unit
    
    - name: Run integration tests
      run: make test-integration
    
    - name: Run benchmark tests
      run: make test-benchmark
    
    - name: Generate coverage report
      run: make test-coverage
    
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
```

## 测试最佳实践

### 1. 测试命名规范

```go
// 好的测试命名
func TestPluginInitialize_WithValidConfig_ShouldSucceed(t *testing.T) {}
func TestPluginProcessAudio_WithInvalidData_ShouldReturnError(t *testing.T) {}
func TestPluginHealthCheck_WhenDisconnected_ShouldFail(t *testing.T) {}

// 避免的命名
func TestPlugin(t *testing.T) {}
func TestFunction1(t *testing.T) {}
func TestStuff(t *testing.T) {}
```

### 2. 测试组织

```go
// 按功能组织测试
func TestPlugin_Initialization(t *testing.T) {
    t.Run("with valid config", func(t *testing.T) {
        // 测试逻辑
    })
    
    t.Run("with invalid config", func(t *testing.T) {
        // 测试逻辑
    })
    
    t.Run("with missing config", func(t *testing.T) {
        // 测试逻辑
    })
}
```

### 3. 测试数据管理

```go
// 使用测试数据文件
func loadTestData(t *testing.T, filename string) []byte {
    data, err := os.ReadFile(filepath.Join("testdata", filename))
    require.NoError(t, err)
    return data
}

// 使用构建器模式创建测试数据
type PluginConfigBuilder struct {
    config map[string]interface{}
}

func NewPluginConfigBuilder() *PluginConfigBuilder {
    return &PluginConfigBuilder{
        config: make(map[string]interface{}),
    }
}

func (b *PluginConfigBuilder) WithSampleRate(rate int) *PluginConfigBuilder {
    b.config["sample_rate"] = rate
    return b
}

func (b *PluginConfigBuilder) WithBufferSize(size int) *PluginConfigBuilder {
    b.config["buffer_size"] = size
    return b
}

func (b *PluginConfigBuilder) Build() map[string]interface{} {
    return b.config
}
```

### 4. 错误测试

```go
func TestPluginErrorHandling(t *testing.T) {
    plugin := &MyPlugin{}
    
    // 测试特定错误类型
    err := plugin.ProcessInvalidData(nil)
    require.Error(t, err)
    
    var pluginErr *PluginError
    assert.True(t, errors.As(err, &pluginErr))
    assert.Equal(t, PluginErrorTypeInvalidData, pluginErr.Type)
    
    // 测试错误消息
    assert.Contains(t, err.Error(), "invalid data")
}
```

### 5. 并发测试

```go
func TestPluginConcurrency(t *testing.T) {
    plugin := &MyPlugin{}
    ctx := NewTestPluginContext()
    
    err := plugin.Initialize(ctx)
    require.NoError(t, err)
    
    const numGoroutines = 100
    const numOperations = 1000
    
    var wg sync.WaitGroup
    errors := make(chan error, numGoroutines)
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            
            for j := 0; j < numOperations; j++ {
                if err := plugin.DoSomething(); err != nil {
                    errors <- err
                    return
                }
            }
        }()
    }
    
    wg.Wait()
    close(errors)
    
    // 检查是否有错误
    for err := range errors {
        t.Errorf("Concurrent operation failed: %v", err)
    }
}
```

## 总结

通过遵循本指南中的测试策略和最佳实践，您可以：

1. **提高代码质量**：通过全面的测试覆盖发现和修复问题
2. **确保稳定性**：通过集成测试验证插件与系统的兼容性
3. **优化性能**：通过基准测试识别性能瓶颈
4. **保证安全性**：通过安全测试验证权限控制和沙箱机制
5. **维护兼容性**：通过兼容性测试确保跨版本和跨平台支持

记住，好的测试不仅能发现问题，还能作为代码的文档，帮助其他开发者理解插件的预期行为。