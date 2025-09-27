# 服务注册表扩展功能

本文档描述了go-musicfox v2微内核架构中服务注册表的扩展功能实现。

## 概述

服务注册表扩展功能在原有基础服务注册表的基础上，增加了以下高级特性：

- **版本管理和兼容性检查**
- **服务监控和性能统计**
- **故障转移和恢复机制**
- **扩展负载均衡策略**
- **服务分组和标签管理**
- **告警和通知系统**

## 核心组件

### 1. 版本管理器 (VersionManager)

负责管理服务版本信息和兼容性检查。

```go
// 创建版本管理器
versionManager := NewVersionManager(logger)

// 解析版本
version, err := ParseVersion("1.2.3-alpha+build.1")

// 注册服务版本
instance := createServiceInstance("service-1", "my-service", "1.0.0")
err = versionManager.RegisterServiceVersion("my-service", version, instance)

// 获取兼容的服务
compatible, err := versionManager.GetCompatibleServices("my-service", requiredVersion)

// 弃用版本
err = versionManager.DeprecateServiceVersion("my-service", version, time.Now(), "Outdated", newVersion)
```

### 2. 指标管理器 (MetricsManager)

收集和管理服务性能指标。

```go
// 创建指标管理器
metricsManager := NewMetricsManager(logger)

// 记录服务调用
err = metricsManager.RecordServiceCall("service-1", 100*time.Millisecond, true, "")

// 获取服务指标
metrics, err := metricsManager.GetServiceMetrics("service-1")

// 设置告警阈值
thresholds := &AlertThresholds{
    MaxResponseTime: 200 * time.Millisecond,
    MaxErrorRate:    5.0,
    Enabled:         true,
}
err = metricsManager.SetAlertThresholds("service-1", thresholds)

// 创建告警
alert := &ServiceAlert{
    ServiceID: "service-1",
    Type:      ServiceAlertTypePerformance,
    Severity:  ServiceAlertSeverityHigh,
    Message:   "High response time detected",
}
err = metricsManager.CreateAlert(alert)
```

### 3. 故障转移管理器 (FailoverManager)

实现服务故障检测、熔断器和自动恢复。

```go
// 创建故障转移管理器
failoverManager := NewFailoverManager(logger, registry, metricsManager)

// 配置故障转移
config := &ServiceFailoverConfig{
    Enabled:               true,
    MaxRetries:           3,
    RetryDelay:           1 * time.Second,
    BackoffMultiplier:    2.0,
    CircuitBreakerEnabled: true,
    FailureThreshold:     5,
    RecoveryTimeout:      30 * time.Second,
}
err = failoverManager.ConfigureFailover("service-1", config)

// 检查是否可以调用服务
canCall, err := failoverManager.CanMakeCall("service-1")

// 记录服务调用结果
err = failoverManager.RecordServiceCall("service-1", true, 100*time.Millisecond)

// 尝试故障转移
result, err := failoverManager.AttemptFailover(ctx, "my-service", "failed-service-1")
```

### 4. 扩展负载均衡器 (ExtendedLoadBalancer)

提供多种高级负载均衡算法。

```go
// 创建扩展负载均衡器
loadBalancer := NewExtendedLoadBalancer(logger)

// 配置负载均衡
config := &LoadBalancerConfig{
    Strategy:     ExtendedLoadBalanceConsistentHash,
    VirtualNodes: 150,
}
err = loadBalancer.ConfigureLoadBalancing("my-service", config)

// 选择服务实例
context := map[string]interface{}{
    "session_id": "user123",
    "client_ip":  "192.168.1.100",
}
instance, err := loadBalancer.SelectServiceWithExtendedStrategy(
    ctx, "my-service", ExtendedLoadBalanceConsistentHash, context, instances)

// 记录响应时间
loadBalancer.RecordResponseTime("service-1", 100*time.Millisecond)

// 更新服务负载
loadBalancer.UpdateServiceLoad("service-1", 80.0, 60.0, 10)
```

## 扩展服务注册表

`ExtendedServiceRegistry` 整合了所有扩展功能：

```go
// 创建扩展服务注册表
config := DefaultExtendedRegistryConfig()
registry := NewExtendedServiceRegistry(logger, config)

// 注册带版本的服务
version, _ := ParseVersion("1.0.0")
service := &ServiceInfo{
    ID:      "service-1",
    Name:    "my-service",
    Version: "1.0.0",
    Address: "localhost",
    Port:    8080,
}
err = registry.RegisterServiceWithVersion(ctx, service, version)

// 记录服务调用
err = registry.RecordServiceCall(ctx, "service-1", 100*time.Millisecond, true)

// 使用扩展负载均衡选择服务
instance, err := registry.SelectServiceWithExtendedStrategy(
    ctx, "my-service", ExtendedLoadBalanceResponseTime, nil)

// 创建服务分组
err = registry.CreateServiceGroup(ctx, "web-services", []string{"service-1", "service-2"})

// 获取扩展统计信息
stats, err := registry.GetRegistryStatistics(ctx)

// 获取服务拓扑
topology, err := registry.GetServiceTopology(ctx)
```

## 负载均衡策略

支持以下负载均衡策略：

1. **轮询 (Round Robin)** - 按顺序轮流选择服务
2. **随机 (Random)** - 随机选择服务
3. **加权轮询 (Weighted Round Robin)** - 根据权重轮询
4. **最少连接 (Least Connections)** - 选择连接数最少的服务
5. **一致性哈希 (Consistent Hash)** - 基于键的一致性哈希
6. **IP哈希 (IP Hash)** - 基于客户端IP的哈希
7. **响应时间 (Response Time)** - 选择响应时间最短的服务
8. **最少负载 (Least Load)** - 选择负载最低的服务
9. **地理位置 (Geographic)** - 基于地理位置的选择

## 监控指标

系统收集以下指标：

- **请求统计**: 总请求数、成功请求数、失败请求数
- **响应时间**: 平均响应时间、最小/最大响应时间
- **错误率**: 失败请求占总请求的百分比
- **吞吐量**: 每秒处理的请求数
- **可用性**: 服务健康状态和正常运行时间
- **资源使用**: CPU使用率、内存使用率、活跃连接数

## 告警类型

支持以下告警类型：

- **性能告警**: 响应时间过长、错误率过高、吞吐量过低
- **可用性告警**: 服务不健康、健康检查失败
- **资源告警**: CPU/内存使用率过高
- **依赖告警**: 依赖服务不可用

## 熔断器状态

熔断器有三种状态：

1. **关闭 (Closed)** - 正常状态，允许请求通过
2. **打开 (Open)** - 熔断状态，拒绝所有请求
3. **半开 (Half-Open)** - 测试状态，允许少量请求测试服务恢复

## 配置选项

```go
type ExtendedRegistryConfig struct {
    EnableVersionManagement bool          // 启用版本管理
    EnableMetrics          bool          // 启用指标收集
    EnableFailover         bool          // 启用故障转移
    EnableExtendedLB       bool          // 启用扩展负载均衡
    CleanupInterval        time.Duration // 清理间隔
    MetricsRetention       time.Duration // 指标保留时间
    MaxEventHistory        int           // 最大事件历史数
}
```

## 最佳实践

1. **版本管理**:
   - 使用语义化版本控制
   - 及时弃用过时版本
   - 提供迁移指南

2. **监控告警**:
   - 设置合理的告警阈值
   - 建立告警处理流程
   - 定期审查告警规则

3. **故障转移**:
   - 配置适当的重试策略
   - 设置合理的熔断阈值
   - 实现优雅降级

4. **负载均衡**:
   - 根据业务场景选择合适的策略
   - 定期监控负载分布
   - 考虑服务的地理位置

5. **性能优化**:
   - 定期清理过期数据
   - 监控系统资源使用
   - 优化服务响应时间

## 测试覆盖

扩展功能包含完整的单元测试和集成测试：

```bash
# 运行所有测试
go test -v ./pkg/kernel/...

# 运行基准测试
go test -bench=. ./pkg/kernel/...

# 查看测试覆盖率
go test -cover ./pkg/kernel/...
```

## 性能考虑

- **内存使用**: 定期清理过期的指标数据和事件历史
- **CPU开销**: 负载均衡算法的计算复杂度
- **网络延迟**: 健康检查和服务发现的频率
- **存储空间**: 指标数据和日志的存储策略

## 故障排除

常见问题和解决方案：

1. **服务发现失败**: 检查服务注册状态和网络连接
2. **负载不均衡**: 验证负载均衡策略配置和权重设置
3. **熔断器误触发**: 调整失败阈值和恢复超时时间
4. **指标数据丢失**: 检查存储配置和清理策略
5. **版本兼容性问题**: 验证版本规则和依赖关系

## 扩展开发

如需添加新的负载均衡策略或监控指标：

1. 实现相应的接口
2. 添加配置选项
3. 编写单元测试
4. 更新文档

## 总结

服务注册表扩展功能为go-musicfox v2提供了企业级的服务治理能力，包括版本管理、性能监控、故障转移和高级负载均衡等特性。通过合理配置和使用这些功能，可以显著提高系统的可靠性、可观测性和性能。