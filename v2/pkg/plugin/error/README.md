# 插件错误分类和处理系统

本模块实现了完整的插件错误分类和处理机制，符合设计文档中的规范要求。

## 核心功能

### 1. 错误分类系统 (`errors.go`)
- **PluginError接口**: 定义了插件错误的标准接口
- **BasePluginError**: 基础插件错误实现
- **ErrorCode枚举**: 包含60+种详细的错误代码分类
- **ErrorType枚举**: 错误类型分类（系统、插件、网络、IO等）
- **ErrorSeverity枚举**: 错误严重程度分级（trace到critical）
- **ErrorRecoveryStrategy**: 错误恢复策略配置

### 2. 错误处理器 (`error_handler.go`)
- **DynamicErrorHandler**: 动态错误处理器，支持多种处理策略
- **错误历史记录**: 维护插件错误历史
- **错误统计**: 实时统计错误信息
- **恢复策略**: 支持重试、重启、回退、优雅降级等策略
- **资源管理**: 集成资源限制和监控

### 3. 错误包装器 (`wrapper.go`)
- **ErrorWrapper接口**: 错误包装和解包功能
- **错误链管理**: 维护错误传播链
- **错误聚合器**: 按插件聚合错误信息
- **上下文增强**: 为错误添加丰富的上下文信息

### 4. 错误传播机制 (`propagation.go`)
- **ErrorPropagator**: 错误传播控制器
- **传播规则**: 基于条件的错误传播规则
- **传播动作**: 支持日志、通知、阻断等动作
- **传播链**: 维护错误传播路径

### 5. 错误日志系统 (`logger.go`)
- **ErrorLogger**: 结构化错误日志记录
- **多级别日志**: 支持trace到fatal的日志级别
- **格式化器**: 灵活的日志格式化
- **过滤器**: 基于条件的日志过滤
- **增强器**: 自动添加上下文信息

### 6. 错误监控系统 (`monitor.go`)
- **ErrorMonitor**: 实时错误监控
- **指标收集**: 收集错误统计指标
- **告警系统**: 基于阈值的自动告警
- **性能监控**: MTTR、MTBF等性能指标

### 7. 智能错误分类器 (`classifier.go`)
- **HybridErrorClassifier**: 混合错误分类器
- **规则引擎**: 基于规则的错误分类
- **机器学习**: 支持训练和预测
- **反馈学习**: 基于反馈优化分类准确性

## 测试覆盖

### 单元测试
- `error_handler_test.go`: 错误处理器测试
- `wrapper_test.go`: 错误包装器测试
- `monitor_test.go`: 错误监控器测试

### 集成测试
- `error_integration_test.go`: 完整的错误处理流程集成测试

## 设计特点

### 1. 模块化设计
- 每个组件职责单一，接口清晰
- 支持组件替换和扩展
- 松耦合架构

### 2. 高性能
- 并发安全设计
- 内存高效的数据结构
- 异步处理支持

### 3. 可扩展性
- 插件化的错误处理策略
- 可配置的分类规则
- 支持自定义错误类型

### 4. 可观测性
- 详细的错误统计
- 实时监控和告警
- 结构化日志记录

## 使用示例

```go
// 创建错误处理器
config := &ErrorHandlerConfig{
    MaxErrorHistory:     1000,
    AutoRecovery:        true,
    MaxRecoveryAttempts: 3,
    ErrorRateWindow:     5 * time.Minute,
}
handler := NewDynamicErrorHandler(config)

// 处理错误
ctx := context.Background()
err := errors.New("plugin operation failed")
handleErr := handler.HandleError(ctx, "my-plugin", err)

// 获取错误统计
stats, _ := handler.GetErrorStats("my-plugin")
fmt.Printf("Total errors: %d\n", stats.TotalErrors)
```

## 配置选项

### ErrorHandlerConfig
- `MaxErrorHistory`: 最大错误历史记录数
- `AutoRecovery`: 是否启用自动恢复
- `MaxRecoveryAttempts`: 最大恢复尝试次数
- `ErrorRateWindow`: 错误率计算窗口
- `EnableStackTrace`: 是否启用堆栈跟踪

### 恢复策略
- **重试策略**: 支持指数退避和抖动
- **重启策略**: 插件重启恢复
- **回退策略**: 降级到备用方案
- **优雅降级**: 部分功能禁用
- **熔断器**: 防止级联故障

## 性能指标

- **MTTR**: 平均恢复时间
- **MTBF**: 平均故障间隔时间
- **错误率**: 单位时间内错误发生率
- **分类准确率**: 错误分类的准确性

## 扩展点

1. **自定义错误类型**: 实现PluginError接口
2. **自定义恢复策略**: 实现RecoveryStrategy
3. **自定义分类规则**: 添加ClassificationRule
4. **自定义日志格式**: 实现ErrorFormatter
5. **自定义监控指标**: 扩展MetricsCollector

## 注意事项

1. 错误处理器是线程安全的，可以在多个goroutine中并发使用
2. 建议为不同类型的插件配置不同的恢复策略
3. 监控告警阈值需要根据实际业务场景调整
4. 错误历史记录会占用内存，建议定期清理
5. 分类器需要足够的训练数据才能达到较好的准确率

## 依赖

- Go 1.19+
- 标准库：context, sync, time, log/slog
- 无外部依赖

## 许可证

本模块遵循项目整体许可证。