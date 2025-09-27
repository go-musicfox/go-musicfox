# 微内核 API 文档

微内核是 go-musicfox v2 架构的核心组件，负责系统的初始化、生命周期管理和核心服务协调。

## 接口定义

### Kernel 接口

```go
type Kernel interface {
    // 生命周期管理
    Initialize(ctx context.Context) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Shutdown(ctx context.Context) error
    
    // 核心组件访问
    GetPluginManager() PluginManager
    GetEventBus() EventBus
    GetServiceRegistry() ServiceRegistry
    GetSecurityManager() SecurityManager
    
    // 配置和工具
    GetConfig() *koanf.Koanf
    GetLogger() *slog.Logger
    GetContainer() *dig.Container
}
```

## 核心方法

### 生命周期管理

#### Initialize(ctx context.Context) error

初始化微内核及其所有核心组件。

**参数:**
- `ctx`: 上下文对象，用于控制初始化过程

**返回值:**
- `error`: 初始化失败时返回错误

**示例:**
```go
kernel := kernel.NewMicroKernel()
ctx := context.Background()

if err := kernel.Initialize(ctx); err != nil {
    log.Fatalf("Failed to initialize kernel: %v", err)
}
```

**初始化流程:**
1. 加载配置文件
2. 初始化日志系统
3. 设置依赖注入容器
4. 创建核心组件（事件总线、服务注册表、安全管理器、插件管理器）
5. 注册核心服务

#### Start(ctx context.Context) error

启动微内核，开始处理请求和事件。

**参数:**
- `ctx`: 上下文对象

**返回值:**
- `error`: 启动失败时返回错误

**示例:**
```go
if err := kernel.Start(ctx); err != nil {
    log.Fatalf("Failed to start kernel: %v", err)
}
```

#### Stop(ctx context.Context) error

停止微内核，优雅关闭所有服务。

**参数:**
- `ctx`: 上下文对象，包含超时控制

**返回值:**
- `error`: 停止过程中的错误

**示例:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := kernel.Stop(ctx); err != nil {
    log.Printf("Error during kernel stop: %v", err)
}
```

#### Shutdown(ctx context.Context) error

强制关闭微内核，释放所有资源。

**参数:**
- `ctx`: 上下文对象

**返回值:**
- `error`: 关闭过程中的错误

### 组件访问方法

#### GetPluginManager() PluginManager

获取插件管理器实例。

**返回值:**
- `PluginManager`: 插件管理器接口

**示例:**
```go
pm := kernel.GetPluginManager()
if err := pm.LoadPlugin("./my-plugin.so", plugin.PluginTypeDynamicLibrary); err != nil {
    log.Printf("Failed to load plugin: %v", err)
}
```

#### GetEventBus() EventBus

获取事件总线实例。

**返回值:**
- `EventBus`: 事件总线接口

**示例:**
```go
eventBus := kernel.GetEventBus()
eventBus.Subscribe("plugin.loaded", func(event *Event) {
    log.Printf("Plugin loaded: %s", event.Data)
})
```

#### GetServiceRegistry() ServiceRegistry

获取服务注册表实例。

**返回值:**
- `ServiceRegistry`: 服务注册表接口

#### GetSecurityManager() SecurityManager

获取安全管理器实例。

**返回值:**
- `SecurityManager`: 安全管理器接口

### 工具方法

#### GetConfig() *koanf.Koanf

获取配置管理器。

**返回值:**
- `*koanf.Koanf`: 配置管理器实例

**示例:**
```go
config := kernel.GetConfig()
logLevel := config.String("log.level")
if logLevel == "" {
    logLevel = "info"
}
```

#### GetLogger() *slog.Logger

获取结构化日志记录器。

**返回值:**
- `*slog.Logger`: 日志记录器实例

**示例:**
```go
logger := kernel.GetLogger()
logger.Info("Kernel operation completed", 
    "operation", "plugin_load",
    "duration", time.Since(start))
```

#### GetContainer() *dig.Container

获取依赖注入容器。

**返回值:**
- `*dig.Container`: 依赖注入容器

**示例:**
```go
container := kernel.GetContainer()
if err := container.Provide(func() MyService {
    return NewMyService()
}); err != nil {
    log.Printf("Failed to register service: %v", err)
}
```

## 实现类

### MicroKernel

```go
type MicroKernel struct {
    container       *dig.Container
    config          *koanf.Koanf
    logger          *slog.Logger
    pluginManager   PluginManager
    eventBus        EventBus
    serviceRegistry ServiceRegistry
    securityManager SecurityManager
    
    ctx    context.Context
    cancel context.CancelFunc
    
    // 内部状态
    state     KernelState
    stateMux  sync.RWMutex
    startTime time.Time
}
```

### 构造函数

#### NewMicroKernel() *MicroKernel

创建新的微内核实例。

**返回值:**
- `*MicroKernel`: 微内核实例

**示例:**
```go
kernel := kernel.NewMicroKernel()
```

#### NewMicroKernelWithConfig(config *koanf.Koanf) *MicroKernel

使用指定配置创建微内核实例。

**参数:**
- `config`: 配置对象

**返回值:**
- `*MicroKernel`: 微内核实例

## 配置选项

微内核支持以下配置选项：

```yaml
# kernel.yaml
kernel:
  # 内核配置
  name: "go-musicfox-kernel"
  version: "2.0.0"
  
  # 日志配置
  log:
    level: "info"  # debug, info, warn, error
    format: "json" # json, text
    output: "stdout" # stdout, stderr, file path
  
  # 插件配置
  plugins:
    # 插件目录
    directories:
      - "./plugins"
      - "/usr/local/lib/go-musicfox/plugins"
    
    # 自动加载插件
    auto_load: true
    
    # 插件加载超时
    load_timeout: "30s"
  
  # 安全配置
  security:
    # 启用沙箱
    sandbox_enabled: true
    
    # 允许的插件源
    trusted_sources:
      - "github.com/go-musicfox"
      - "plugins.go-musicfox.com"
  
  # 性能配置
  performance:
    # 最大并发插件数
    max_concurrent_plugins: 10
    
    # 事件队列大小
    event_queue_size: 1000
    
    # 垃圾回收间隔
    gc_interval: "5m"
```

## 错误处理

### 错误类型

```go
type KernelError struct {
    Type      KernelErrorType `json:"type"`
    Code      string          `json:"code"`
    Message   string          `json:"message"`
    Cause     error           `json:"cause,omitempty"`
    Timestamp time.Time       `json:"timestamp"`
}

type KernelErrorType int

const (
    KernelErrorTypeInitialization KernelErrorType = iota
    KernelErrorTypeStartup
    KernelErrorTypeShutdown
    KernelErrorTypeConfiguration
    KernelErrorTypeComponent
)
```

### 常见错误

| 错误代码 | 描述 | 解决方案 |
|---------|------|----------|
| `KERNEL_INIT_FAILED` | 内核初始化失败 | 检查配置文件和依赖 |
| `KERNEL_START_TIMEOUT` | 启动超时 | 增加启动超时时间 |
| `KERNEL_CONFIG_INVALID` | 配置无效 | 验证配置文件格式 |
| `KERNEL_COMPONENT_ERROR` | 组件错误 | 检查组件依赖和状态 |

## 最佳实践

### 1. 生命周期管理

```go
func main() {
    kernel := kernel.NewMicroKernel()
    
    // 设置优雅关闭
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // 监听系统信号
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        log.Println("Shutting down kernel...")
        
        shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer shutdownCancel()
        
        if err := kernel.Stop(shutdownCtx); err != nil {
            log.Printf("Error during shutdown: %v", err)
        }
        cancel()
    }()
    
    // 初始化和启动
    if err := kernel.Initialize(ctx); err != nil {
        log.Fatalf("Failed to initialize: %v", err)
    }
    
    if err := kernel.Start(ctx); err != nil {
        log.Fatalf("Failed to start: %v", err)
    }
    
    // 等待关闭信号
    <-ctx.Done()
    log.Println("Kernel stopped")
}
```

### 2. 错误处理

```go
func handleKernelError(err error) {
    if kernelErr, ok := err.(*kernel.KernelError); ok {
        switch kernelErr.Type {
        case kernel.KernelErrorTypeInitialization:
            log.Printf("Initialization error: %s", kernelErr.Message)
            // 重试初始化或退出
        case kernel.KernelErrorTypeStartup:
            log.Printf("Startup error: %s", kernelErr.Message)
            // 尝试恢复或降级启动
        default:
            log.Printf("Kernel error: %s", kernelErr.Message)
        }
    } else {
        log.Printf("Unknown error: %v", err)
    }
}
```

### 3. 配置管理

```go
func loadKernelConfig() *koanf.Koanf {
    k := koanf.New(".")
    
    // 加载默认配置
    if err := k.Load(confmap.Provider(defaultConfig, "."), nil); err != nil {
        log.Printf("Failed to load default config: %v", err)
    }
    
    // 加载配置文件
    if err := k.Load(file.Provider("kernel.yaml"), yaml.Parser()); err != nil {
        log.Printf("Failed to load config file: %v", err)
    }
    
    // 加载环境变量
    if err := k.Load(env.Provider("MUSICFOX_", ".", func(s string) string {
        return strings.Replace(strings.ToLower(
            strings.TrimPrefix(s, "MUSICFOX_")), "_", ".", -1)
    }), nil); err != nil {
        log.Printf("Failed to load env vars: %v", err)
    }
    
    return k
}
```

## 监控和诊断

### 内核状态

```go
type KernelState int

const (
    KernelStateUninitialized KernelState = iota
    KernelStateInitializing
    KernelStateInitialized
    KernelStateStarting
    KernelStateRunning
    KernelStateStopping
    KernelStateStopped
    KernelStateError
)

// 获取内核状态
func (k *MicroKernel) GetState() KernelState {
    k.stateMux.RLock()
    defer k.stateMux.RUnlock()
    return k.state
}

// 获取运行时统计
func (k *MicroKernel) GetStats() *KernelStats {
    return &KernelStats{
        State:       k.GetState(),
        StartTime:   k.startTime,
        Uptime:      time.Since(k.startTime),
        PluginCount: k.pluginManager.GetPluginCount(),
        EventCount:  k.eventBus.GetEventCount(),
    }
}
```

### 健康检查

```go
func (k *MicroKernel) HealthCheck() error {
    if k.GetState() != KernelStateRunning {
        return fmt.Errorf("kernel not running")
    }
    
    // 检查核心组件
    if err := k.pluginManager.HealthCheck(); err != nil {
        return fmt.Errorf("plugin manager unhealthy: %w", err)
    }
    
    if err := k.eventBus.HealthCheck(); err != nil {
        return fmt.Errorf("event bus unhealthy: %w", err)
    }
    
    return nil
}
```

## 相关文档

- [插件管理器 API](plugin-manager.md)
- [事件总线 API](event-bus.md)
- [服务注册表 API](service-registry.md)
- [安全管理器 API](security-manager.md)
- [微内核架构设计](../architecture/microkernel.md)