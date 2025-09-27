# 插件错误恢复策略系统

本包实现了一个完整的插件错误恢复策略系统，提供了多种恢复机制来确保插件系统的高可用性和故障自愈能力。

## 功能特性

### 1. 熔断器机制 (Circuit Breaker)
- **状态管理**: 支持关闭、开启、半开三种状态
- **自动状态转换**: 基于失败阈值和恢复时间的智能状态切换
- **可配置参数**: 失败阈值、成功阈值、超时时间、恢复时间等
- **实时监控**: 提供详细的指标和状态信息
- **回调支持**: 状态变化和请求拒绝的回调通知

### 2. 重试策略 (Retry Strategy)
- **多种重试策略**: 固定间隔、线性增长、指数退避、自定义策略
- **智能退避**: 支持抖动机制，避免惊群效应
- **错误分类**: 可配置的可重试错误类型
- **超时控制**: 单次操作和整体重试的超时管理
- **指标收集**: 详细的重试统计和成功率分析

### 3. 降级策略 (Fallback Strategy)
- **多种降级类型**: 服务降级、功能降级、缓存降级、默认值降级
- **缓存机制**: 内置缓存支持，提供快速降级响应
- **自定义降级**: 支持注册自定义降级函数
- **优雅降级**: 渐进式降级链，确保服务连续性
- **性能监控**: 降级成功率和缓存命中率统计

### 4. 自动恢复机制 (Auto Recovery)
- **健康检查**: 定期检查插件健康状态
- **多种恢复动作**: 重启、重新加载、故障转移、重置状态
- **智能触发**: 基于失败阈值和恢复尝试次数的智能恢复
- **并发控制**: 防止恢复操作过载
- **历史记录**: 完整的恢复历史和分析

### 5. 统一管理器 (Recovery Manager)
- **策略编排**: 统一管理和调度所有恢复策略
- **优先级控制**: 支持策略优先级和组合执行
- **并发限制**: 控制并发恢复操作数量
- **事件系统**: 完整的事件记录和回调机制
- **指标聚合**: 全局恢复指标和健康状态监控

### 6. 配置管理 (Configuration Management)
- **动态配置**: 支持运行时配置更新
- **配置验证**: 完整的配置验证和错误提示
- **版本管理**: 配置版本控制和兼容性检查
- **持久化**: JSON格式的配置序列化和反序列化
- **热更新**: 配置变更的实时通知和应用

## 快速开始

### 基本使用

```go
package main

import (
    "context"
    "log/slog"
    "os"
    "time"
    
    "your-project/pkg/plugin/recovery"
)

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
    
    // 1. 创建熔断器
    cbConfig := recovery.DefaultCircuitBreakerConfig()
    cbConfig.FailureThreshold = 5
    cbConfig.RecoveryTimeout = 30 * time.Second
    
    cb := recovery.NewCircuitBreaker("my-service-cb", cbConfig, logger)
    
    // 2. 使用熔断器保护操作
    ctx := context.Background()
    err := cb.Execute(ctx, func(ctx context.Context) error {
        // 你的业务逻辑
        return callExternalService()
    })
    
    if err != nil {
        logger.Error("Operation failed", "error", err)
    }
}
```

### 重试策略使用

```go
// 创建重试策略
retryConfig := recovery.DefaultRetryConfig()
retryConfig.MaxRetries = 3
retryConfig.Policy = recovery.RetryPolicyExponential
retryConfig.RetryableErrors = []string{"timeout", "connection"}

rs := recovery.NewRetryStrategy("my-retry", retryConfig, logger)

// 执行带重试的操作
err := rs.Execute(ctx, func(ctx context.Context) error {
    return unreliableOperation()
})
```

### 降级策略使用

```go
// 创建降级策略
fallbackConfig := recovery.DefaultFallbackConfig()
fallbackConfig.Type = recovery.FallbackTypeDefault
fallbackConfig.DefaultValue = "服务暂时不可用"

fs := recovery.NewFallbackStrategy("my-fallback", fallbackConfig, logger)

// 执行带降级的操作
result, err := fs.Execute(ctx, func(ctx context.Context) (interface{}, error) {
    return primaryService()
}, nil)
```

### 完整的恢复管理器使用

```go
// 1. 创建配置管理器
cm := recovery.NewConfigManager(logger)

// 2. 添加各种策略配置
cm.AddCircuitBreakerConfig("service-cb", cbConfig)
cm.AddRetryConfig("service-retry", retryConfig)
cm.AddFallbackConfig("service-fallback", fallbackConfig)

// 3. 创建恢复管理器
rmConfig := recovery.DefaultRecoveryManagerConfig()
rm := recovery.NewRecoveryManager(rmConfig, logger)

// 4. 注册策略
cb := recovery.NewCircuitBreaker("service-cb", cbConfig, logger)
rm.RegisterStrategy(&CircuitBreakerWrapper{cb: cb})

// 5. 启动管理器
ctx := context.Background()
rm.Start(ctx)
defer rm.Stop()

// 6. 执行恢复策略
result, err := rm.ExecuteRecovery(ctx, "my-plugin", 
    []string{"service-cb", "service-retry"}, 
    func(ctx context.Context) (interface{}, error) {
        return businessLogic()
    }, nil)
```

## 配置示例

### 熔断器配置

```json
{
  "failure_threshold": 5,
  "success_threshold": 3,
  "timeout": "30s",
  "recovery_timeout": "60s",
  "max_requests": 10,
  "monitoring_window": "5m"
}
```

### 重试策略配置

```json
{
  "max_retries": 3,
  "initial_delay": "100ms",
  "max_delay": "30s",
  "backoff_factor": 2.0,
  "jitter": true,
  "jitter_factor": 0.1,
  "policy": "exponential",
  "timeout": "30s",
  "retryable_errors": ["timeout", "connection", "temporary"]
}
```

### 降级策略配置

```json
{
  "type": "cache",
  "enabled": true,
  "timeout": "10s",
  "max_concurrency": 10,
  "cache_expiry": "5m",
  "priority": 1
}
```

### 自动恢复配置

```json
{
  "enabled": true,
  "health_check_interval": "30s",
  "health_check_timeout": "10s",
  "max_recovery_attempts": 3,
  "recovery_delay": "5s",
  "failure_threshold": 3,
  "recovery_actions": ["reset", "restart", "reload", "failover"]
}
```

## 监控和指标

### 熔断器指标

```go
metrics := cb.GetMetrics()
fmt.Printf("总请求数: %d\n", metrics.TotalRequests)
fmt.Printf("成功请求数: %d\n", metrics.SuccessRequests)
fmt.Printf("失败请求数: %d\n", metrics.FailureRequests)
fmt.Printf("拒绝请求数: %d\n", metrics.RejectedRequests)
fmt.Printf("状态变化次数: %d\n", metrics.StateChanges)
```

### 重试策略指标

```go
metrics := rs.GetMetrics()
fmt.Printf("总尝试次数: %d\n", metrics.TotalAttempts)
fmt.Printf("成功尝试次数: %d\n", metrics.SuccessAttempts)
fmt.Printf("总重试次数: %d\n", metrics.TotalRetries)
fmt.Printf("成功率: %.2f%%\n", rs.GetSuccessRate()*100)
```

### 管理器指标

```go
metrics := rm.GetMetrics()
fmt.Printf("总策略数: %d\n", metrics.TotalStrategies)
fmt.Printf("活跃策略数: %d\n", metrics.ActiveStrategies)
fmt.Printf("总恢复次数: %d\n", metrics.TotalRecoveries)
fmt.Printf("成功恢复次数: %d\n", metrics.SuccessfulRecoveries)
fmt.Printf("平均恢复时间: %v\n", metrics.AverageRecoveryTime)
```

## 最佳实践

### 1. 策略组合

建议按以下优先级组合使用策略：
1. **熔断器** - 快速失败，保护系统
2. **重试策略** - 处理临时性错误
3. **降级策略** - 提供备用方案
4. **自动恢复** - 长期健康管理

### 2. 配置调优

- **熔断器**: 根据服务的正常错误率设置失败阈值
- **重试**: 避免过度重试，设置合理的最大重试次数
- **降级**: 确保降级方案的可用性和性能
- **恢复**: 设置适当的健康检查间隔，避免过于频繁

### 3. 监控告警

- 监控熔断器状态变化
- 关注重试成功率下降
- 监控降级策略触发频率
- 设置恢复失败告警

### 4. 测试策略

- 单元测试覆盖所有策略类型
- 集成测试验证策略组合效果
- 压力测试验证并发性能
- 故障注入测试验证恢复能力

## 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                    Recovery Manager                          │
├─────────────────────────────────────────────────────────────┤
│  Strategy Registration │ Execution Orchestration │ Metrics  │
└─────────────────────────────────────────────────────────────┘
                                │
                ┌───────────────┼───────────────┐
                │               │               │
        ┌───────▼──────┐ ┌──────▼──────┐ ┌─────▼──────┐
        │Circuit Breaker│ │Retry Strategy│ │Fallback    │
        │               │ │              │ │Strategy    │
        └───────────────┘ └─────────────┘ └────────────┘
                                │
                        ┌───────▼──────┐
                        │Auto Recovery │
                        │Manager       │
                        └──────────────┘
                                │
                        ┌───────▼──────┐
                        │Config Manager│
                        └──────────────┘
```

## 扩展开发

### 自定义恢复策略

```go
type CustomStrategy struct {
    name string
    // 其他字段
}

func (cs *CustomStrategy) GetName() string {
    return cs.name
}

func (cs *CustomStrategy) GetType() recovery.StrategyType {
    return recovery.StrategyTypeCustom
}

func (cs *CustomStrategy) Execute(ctx context.Context, operation func(ctx context.Context) (interface{}, error), args interface{}) (interface{}, error) {
    // 自定义恢复逻辑
    return operation(ctx)
}

func (cs *CustomStrategy) Reset() {
    // 重置逻辑
}

func (cs *CustomStrategy) IsHealthy() bool {
    // 健康检查逻辑
    return true
}
```

### 自定义降级函数

```go
fs.RegisterFallbackFunc(recovery.FallbackTypeCustom, func(ctx context.Context, args interface{}) (interface{}, error) {
    // 自定义降级逻辑
    return "custom fallback result", nil
})
```

## 性能考虑

- **内存使用**: 合理设置历史记录和缓存大小
- **CPU开销**: 避免过于频繁的健康检查
- **网络影响**: 重试策略要考虑网络延迟
- **并发控制**: 使用适当的并发限制避免资源竞争

## 故障排查

### 常见问题

1. **熔断器一直开启**
   - 检查失败阈值设置是否合理
   - 确认底层服务是否真的有问题
   - 查看恢复超时时间设置

2. **重试策略不生效**
   - 确认错误类型是否在可重试列表中
   - 检查重试次数和延迟配置
   - 验证超时时间设置

3. **降级策略失败**
   - 检查降级函数是否正确注册
   - 确认降级配置类型匹配
   - 验证缓存过期时间设置

### 调试技巧

- 启用详细日志记录
- 使用指标监控策略状态
- 查看历史记录分析问题模式
- 使用回调函数进行实时监控

## 版本兼容性

- 当前版本: v1.0.0
- Go版本要求: 1.21+
- 配置版本: v1

## 贡献指南

1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。