# API 文档索引

本目录包含 go-musicfox v2 微内核插件架构的所有核心 API 文档。

## 核心组件 API

### [微内核 API](kernel.md)
微内核是整个系统的核心，负责管理所有组件的生命周期和依赖关系。

**主要接口：**
- `Kernel` - 微内核主接口
- `MicroKernel` - 微内核实现
- 生命周期管理方法
- 组件访问方法

### [插件管理器 API](plugin-manager.md)
插件管理器负责插件的加载、卸载、生命周期管理和类型适配。

**主要接口：**
- `PluginManager` - 插件管理器接口
- `HybridPluginManager` - 混合插件管理器实现
- 插件加载和卸载方法
- 插件状态管理

### [事件总线 API](event-bus.md)
事件总线提供系统内异步事件通信机制，支持插件间的松耦合通信。

**主要接口：**
- `EventBus` - 事件总线接口
- 事件发布和订阅方法
- 事件过滤和路由
- 事件优先级管理

### [服务注册表 API](service-registry.md)
服务注册表管理系统中所有服务的注册、发现和健康检查。

**主要接口：**
- `ServiceRegistry` - 服务注册表接口
- 服务注册和发现方法
- 健康检查机制
- 服务依赖管理

### [安全管理器 API](security-manager.md)
安全管理器负责插件的安全验证、权限控制和沙箱管理。

**主要接口：**
- `SecurityManager` - 安全管理器接口
- 插件安全验证
- 权限控制机制
- 沙箱隔离管理

## 使用指南

### 快速开始

1. **初始化微内核**
   ```go
   kernel := kernel.NewMicroKernel()
   if err := kernel.Initialize(ctx); err != nil {
       log.Fatal(err)
   }
   ```

2. **获取核心组件**
   ```go
   pluginManager := kernel.GetPluginManager()
   eventBus := kernel.GetEventBus()
   serviceRegistry := kernel.GetServiceRegistry()
   ```

3. **加载插件**
   ```go
   err := pluginManager.LoadPlugin("/path/to/plugin.so", plugin.TypeDynamicLibrary)
   if err != nil {
       log.Printf("Failed to load plugin: %v", err)
   }
   ```

### API 设计原则

- **接口优先**：所有核心组件都定义了清晰的接口
- **依赖注入**：使用 dig 容器管理组件依赖
- **错误处理**：统一的错误类型和处理策略
- **并发安全**：所有 API 都是线程安全的
- **可测试性**：接口设计便于单元测试和模拟

### 错误处理

所有 API 都遵循 Go 的错误处理约定：

```go
result, err := someAPI.DoSomething()
if err != nil {
    // 处理错误
    return fmt.Errorf("operation failed: %w", err)
}
```

### 上下文管理

所有长时间运行的操作都支持 context.Context：

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := component.Start(ctx)
```

## 相关文档

- [插件开发指南](../guides/README.md)
- [架构设计文档](../architecture/README.md)
- [示例代码](../examples/README.md)
- [开发工具](../tools/README.md)

## 版本兼容性

当前 API 版本：`v2.0.0`

- **向后兼容**：在主版本内保持 API 兼容性
- **废弃通知**：废弃的 API 会提前一个版本通知
- **迁移指南**：重大变更会提供详细的迁移指南

## 贡献指南

如果您发现 API 文档有问题或需要改进，请：

1. 在 GitHub 上提交 Issue
2. 提供具体的改进建议
3. 如果可能，提交 Pull Request

---

更多信息请参考 [项目主页](../README.md) 或查看具体的 API 文档。