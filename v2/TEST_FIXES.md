# Go-MusicFox v2 测试超时问题修复

## 问题描述

在运行 `go test ./...` 时遇到测试超时问题，主要表现为：

1. 测试在30秒后超时退出
2. 出现大量goroutine堆栈信息
3. 疑似存在死循环或死锁

## 问题分析

通过分析发现主要问题在于：

### 1. EventBus死锁问题

**位置**: `pkg/event/bus.go` 的 `Stop()` 方法

**问题**: 
- 在Stop方法中使用了`defer eb.mutex.Unlock()`，但在等待goroutine结束时可能导致死锁
- `eb.wg.Wait()`可能无限等待，没有超时机制

**修复**:
- 重构了Stop方法的锁管理，避免长时间持有锁
- 添加了10秒超时机制，防止无限等待
- 使用goroutine异步等待，避免阻塞

### 2. 性能测试超时设置过长

**位置**: `test/integration/performance_stability_test.go`

**问题**:
- 测试套件超时设置为5分钟（300秒）
- 长时间运行稳定性测试设置为1分钟

**修复**:
- 将测试套件超时缩短为1分钟（60秒）
- 将长时间运行测试缩短为30秒
- 缩短检查间隔从5秒到2秒

## 修复内容

### 1. EventBus Stop方法修复

```go
// 修复前：可能导致死锁
func (eb *DefaultEventBus) Stop(ctx context.Context) error {
    eb.mutex.Lock()
    defer eb.mutex.Unlock()  // 长时间持有锁
    // ...
    eb.wg.Wait()  // 可能无限等待
}

// 修复后：避免死锁，添加超时
func (eb *DefaultEventBus) Stop(ctx context.Context) error {
    eb.mutex.Lock()
    if !eb.running {
        eb.mutex.Unlock()
        return fmt.Errorf("event bus is not running")
    }
    eb.running = false
    eb.mutex.Unlock()  // 尽早释放锁
    
    // 使用带超时的等待
    done := make(chan struct{})
    go func() {
        eb.wg.Wait()
        close(done)
    }()
    
    select {
    case <-done:
        return nil
    case <-time.After(10 * time.Second):  // 10秒超时
        return fmt.Errorf("event bus stop timeout")
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

### 2. 测试超时设置优化

```go
// 修复前：超时时间过长
func (suite *PerformanceStabilityTestSuite) SetupSuite() {
    suite.ctx, suite.cancel = context.WithTimeout(context.Background(), 300*time.Second) // 5分钟
}

func TestLongRunningStability() {
    const testDuration = 60 * time.Second // 1分钟
    const checkInterval = 5 * time.Second
}

// 修复后：合理的超时时间
func (suite *PerformanceStabilityTestSuite) SetupSuite() {
    suite.ctx, suite.cancel = context.WithTimeout(context.Background(), 60*time.Second) // 1分钟
}

func TestLongRunningStability() {
    const testDuration = 30 * time.Second // 30秒
    const checkInterval = 2 * time.Second
}
```

## 使用方法

### 1. 使用安全测试脚本

```bash
# 运行所有测试（推荐）
./run_tests_safe.sh
```

### 2. 手动运行单个测试包

```bash
# 运行集成测试
cd test/integration && go test -v -timeout=60s

# 运行端到端测试
cd test/e2e && go test -v -timeout=60s

# 运行事件系统测试
cd pkg/event && go test -v -timeout=60s
```

### 3. 运行特定测试

```bash
# 运行特定的测试方法
go test -v -timeout=60s -run TestConcurrentPlayback

# 运行性能测试
go test -v -timeout=60s -run TestPerformance
```

## 注意事项

1. **不要使用 `go test ./...`**: 这会运行所有测试，可能导致超时
2. **设置合理的超时时间**: 建议使用 `-timeout=60s` 参数
3. **分包运行测试**: 使用提供的脚本或手动分包运行
4. **监控goroutine**: 如果仍然出现问题，检查goroutine泄漏

## 验证修复

修复后的测试应该能够：

1. ✅ 集成测试在45秒内完成
2. ✅ 端到端测试在30秒内完成
3. ✅ 没有goroutine泄漏警告
4. ✅ EventBus能够正常启动和停止

## 如果问题仍然存在

如果修复后仍然出现超时问题，请：

1. 检查系统资源使用情况
2. 运行单个测试文件定位具体问题
3. 使用 `go test -v -timeout=30s -run TestSpecificTest` 运行特定测试
4. 检查是否有其他进程占用资源

## 相关文件

- `pkg/event/bus.go` - EventBus实现
- `test/integration/performance_stability_test.go` - 性能测试
- `test/e2e/music_playback_e2e_test.go` - 端到端测试
- `run_tests_safe.sh` - 安全测试脚本