# 服务注册表 API 文档

服务注册表是微内核架构中的服务发现和管理组件，提供了插件间服务的注册、发现、调用和生命周期管理功能。

## 接口定义

### ServiceRegistry 接口

```go
type ServiceRegistry interface {
    // 服务注册
    RegisterService(service ServiceDescriptor) error
    RegisterServiceWithProvider(name string, provider ServiceProvider) error
    RegisterPlugin(plugin Plugin) error
    
    // 服务发现
    GetService(name string) (interface{}, error)
    GetServiceWithType(name string, serviceType reflect.Type) (interface{}, error)
    FindServices(criteria ServiceCriteria) ([]ServiceDescriptor, error)
    ListServices() ([]ServiceDescriptor, error)
    
    // 服务管理
    UnregisterService(name string) error
    UpdateService(name string, service ServiceDescriptor) error
    
    // 服务调用
    CallService(name string, method string, args ...interface{}) (interface{}, error)
    CallServiceAsync(name string, method string, args ...interface{}) (<-chan ServiceResult, error)
    
    // 健康检查
    CheckServiceHealth(name string) (*ServiceHealth, error)
    GetUnhealthyServices() ([]string, error)
    
    // 监控和统计
    GetServiceStats(name string) (*ServiceStats, error)
    GetRegistryStats() *RegistryStats
    
    // 事件订阅
    Subscribe(eventType ServiceEventType, handler ServiceEventHandler) error
    Unsubscribe(eventType ServiceEventType, handler ServiceEventHandler) error
    
    // 生命周期
    Start() error
    Stop() error
    HealthCheck() error
}
```

## 核心数据结构

### ServiceDescriptor 结构

```go
type ServiceDescriptor struct {
    Name         string                 `json:"name"`
    Type         string                 `json:"type"`
    Version      string                 `json:"version"`
    Description  string                 `json:"description"`
    Provider     string                 `json:"provider"`
    Interface    reflect.Type           `json:"-"`
    Instance     interface{}            `json:"-"`
    Metadata     map[string]interface{} `json:"metadata"`
    Tags         []string               `json:"tags"`
    Dependencies []string               `json:"dependencies"`
    Endpoints    []ServiceEndpoint      `json:"endpoints"`
    HealthCheck  HealthCheckConfig      `json:"health_check"`
    CreatedAt    time.Time              `json:"created_at"`
    UpdatedAt    time.Time              `json:"updated_at"`
}

type ServiceEndpoint struct {
    Name        string            `json:"name"`
    Method      string            `json:"method"`
    Path        string            `json:"path,omitempty"`
    Parameters  []ParameterInfo   `json:"parameters"`
    ReturnType  string            `json:"return_type"`
    Description string            `json:"description"`
    Metadata    map[string]string `json:"metadata"`
}

type ParameterInfo struct {
    Name        string `json:"name"`
    Type        string `json:"type"`
    Required    bool   `json:"required"`
    Description string `json:"description"`
}
```

### ServiceProvider 接口

```go
type ServiceProvider interface {
    GetService() (interface{}, error)
    GetServiceType() reflect.Type
    IsHealthy() bool
    Cleanup() error
}

// 单例服务提供者
type SingletonServiceProvider struct {
    instance    interface{}
    serviceType reflect.Type
    factory     func() (interface{}, error)
    once        sync.Once
}

// 工厂服务提供者
type FactoryServiceProvider struct {
    factory     func() (interface{}, error)
    serviceType reflect.Type
}

// 实例服务提供者
type InstanceServiceProvider struct {
    instance    interface{}
    serviceType reflect.Type
}
```

## 核心方法

### 服务注册

#### RegisterService(service ServiceDescriptor) error

注册服务描述符。

**参数:**
- `service`: 服务描述符

**返回值:**
- `error`: 注册失败时返回错误

**示例:**
```go
registry := kernel.GetServiceRegistry()

// 注册音频处理服务
audioService := ServiceDescriptor{
    Name:        "audio-processor",
    Type:        "AudioProcessor",
    Version:     "1.0.0",
    Description: "Audio processing service",
    Provider:    "audio-processor-plugin",
    Instance:    audioProcessorInstance,
    Tags:        []string{"audio", "processing"},
    Endpoints: []ServiceEndpoint{
        {
            Name:        "ProcessAudio",
            Method:      "ProcessAudio",
            Parameters: []ParameterInfo{
                {Name: "input", Type: "[]byte", Required: true},
                {Name: "sampleRate", Type: "int", Required: true},
                {Name: "channels", Type: "int", Required: true},
            },
            ReturnType:  "[]byte",
            Description: "Process audio data",
        },
    },
    HealthCheck: HealthCheckConfig{
        Enabled:  true,
        Interval: 30 * time.Second,
        Timeout:  5 * time.Second,
    },
}

if err := registry.RegisterService(audioService); err != nil {
    log.Printf("Failed to register audio service: %v", err)
}
```

#### RegisterServiceWithProvider(name string, provider ServiceProvider) error

使用服务提供者注册服务。

**参数:**
- `name`: 服务名称
- `provider`: 服务提供者

**返回值:**
- `error`: 注册失败时返回错误

**示例:**
```go
// 单例服务提供者
singletonProvider := &SingletonServiceProvider{
    serviceType: reflect.TypeOf((*MusicSourceService)(nil)).Elem(),
    factory: func() (interface{}, error) {
        return NewNeteaseCloudMusicService(), nil
    },
}

if err := registry.RegisterServiceWithProvider("netease-music", singletonProvider); err != nil {
    log.Printf("Failed to register music service: %v", err)
}

// 工厂服务提供者
factoryProvider := &FactoryServiceProvider{
    serviceType: reflect.TypeOf((*AudioEffect)(nil)).Elem(),
    factory: func() (interface{}, error) {
        return NewAudioEffect(), nil
    },
}

if err := registry.RegisterServiceWithProvider("audio-effect", factoryProvider); err != nil {
    log.Printf("Failed to register audio effect service: %v", err)
}
```

#### RegisterPlugin(plugin Plugin) error

注册插件提供的所有服务。

**参数:**
- `plugin`: 插件实例

**返回值:**
- `error`: 注册失败时返回错误

**示例:**
```go
// 插件自动注册其服务
if err := registry.RegisterPlugin(audioPlugin); err != nil {
    log.Printf("Failed to register plugin services: %v", err)
}
```

### 服务发现

#### GetService(name string) (interface{}, error)

获取服务实例。

**参数:**
- `name`: 服务名称

**返回值:**
- `interface{}`: 服务实例
- `error`: 服务不存在或获取失败时返回错误

**示例:**
```go
// 获取音频处理服务
service, err := registry.GetService("audio-processor")
if err != nil {
    log.Printf("Failed to get audio service: %v", err)
    return
}

// 类型断言
if audioProcessor, ok := service.(AudioProcessor); ok {
    result, err := audioProcessor.ProcessAudio(audioData, 44100, 2)
    if err != nil {
        log.Printf("Audio processing failed: %v", err)
    }
}
```

#### GetServiceWithType(name string, serviceType reflect.Type) (interface{}, error)

获取指定类型的服务实例。

**参数:**
- `name`: 服务名称
- `serviceType`: 期望的服务类型

**返回值:**
- `interface{}`: 服务实例
- `error`: 服务不存在或类型不匹配时返回错误

**示例:**
```go
// 获取特定类型的服务
musicServiceType := reflect.TypeOf((*MusicSourceService)(nil)).Elem()
service, err := registry.GetServiceWithType("netease-music", musicServiceType)
if err != nil {
    log.Printf("Failed to get music service: %v", err)
    return
}

musicService := service.(MusicSourceService)
songs, err := musicService.Search("周杰伦", SearchOptions{Limit: 10})
```

#### FindServices(criteria ServiceCriteria) ([]ServiceDescriptor, error)

根据条件查找服务。

**参数:**
- `criteria`: 查找条件

**返回值:**
- `[]ServiceDescriptor`: 匹配的服务描述符列表
- `error`: 查找失败时返回错误

**示例:**
```go
// 查找所有音频相关服务
criteria := ServiceCriteria{
    Tags:    []string{"audio"},
    Type:    "AudioProcessor",
    Healthy: true,
}

services, err := registry.FindServices(criteria)
if err != nil {
    log.Printf("Failed to find services: %v", err)
    return
}

for _, service := range services {
    log.Printf("Found audio service: %s v%s", service.Name, service.Version)
}
```

### 服务调用

#### CallService(name string, method string, args ...interface{}) (interface{}, error)

调用服务方法。

**参数:**
- `name`: 服务名称
- `method`: 方法名称
- `args`: 方法参数

**返回值:**
- `interface{}`: 方法返回值
- `error`: 调用失败时返回错误

**示例:**
```go
// 调用音频处理服务
result, err := registry.CallService("audio-processor", "ProcessAudio", audioData, 44100, 2)
if err != nil {
    log.Printf("Failed to call audio service: %v", err)
    return
}

processedAudio := result.([]byte)
```

#### CallServiceAsync(name string, method string, args ...interface{}) (<-chan ServiceResult, error)

异步调用服务方法。

**参数:**
- `name`: 服务名称
- `method`: 方法名称
- `args`: 方法参数

**返回值:**
- `<-chan ServiceResult`: 结果通道
- `error`: 调用失败时返回错误

**示例:**
```go
// 异步调用音乐搜索服务
resultChan, err := registry.CallServiceAsync("netease-music", "Search", "周杰伦", SearchOptions{Limit: 10})
if err != nil {
    log.Printf("Failed to call music service: %v", err)
    return
}

// 等待结果
select {
case result := <-resultChan:
    if result.Error != nil {
        log.Printf("Music search failed: %v", result.Error)
        return
    }
    songs := result.Value.([]*Song)
    log.Printf("Found %d songs", len(songs))
case <-time.After(10 * time.Second):
    log.Printf("Music search timeout")
}
```

## 服务查找条件

### ServiceCriteria 结构

```go
type ServiceCriteria struct {
    Name         string            `json:"name,omitempty"`
    Type         string            `json:"type,omitempty"`
    Version      string            `json:"version,omitempty"`
    Provider     string            `json:"provider,omitempty"`
    Tags         []string          `json:"tags,omitempty"`
    Metadata     map[string]string `json:"metadata,omitempty"`
    Healthy      bool              `json:"healthy,omitempty"`
    MinVersion   string            `json:"min_version,omitempty"`
    MaxVersion   string            `json:"max_version,omitempty"`
}

// 查找条件构建器
type CriteriaBuilder struct {
    criteria ServiceCriteria
}

func NewCriteriaBuilder() *CriteriaBuilder {
    return &CriteriaBuilder{}
}

func (cb *CriteriaBuilder) WithName(name string) *CriteriaBuilder {
    cb.criteria.Name = name
    return cb
}

func (cb *CriteriaBuilder) WithType(serviceType string) *CriteriaBuilder {
    cb.criteria.Type = serviceType
    return cb
}

func (cb *CriteriaBuilder) WithTags(tags ...string) *CriteriaBuilder {
    cb.criteria.Tags = append(cb.criteria.Tags, tags...)
    return cb
}

func (cb *CriteriaBuilder) WithMetadata(key, value string) *CriteriaBuilder {
    if cb.criteria.Metadata == nil {
        cb.criteria.Metadata = make(map[string]string)
    }
    cb.criteria.Metadata[key] = value
    return cb
}

func (cb *CriteriaBuilder) HealthyOnly() *CriteriaBuilder {
    cb.criteria.Healthy = true
    return cb
}

func (cb *CriteriaBuilder) Build() ServiceCriteria {
    return cb.criteria
}
```

### 使用示例

```go
// 使用构建器创建查找条件
criteria := NewCriteriaBuilder().
    WithType("MusicSource").
    WithTags("streaming", "cloud").
    WithMetadata("region", "china").
    HealthyOnly().
    Build()

services, err := registry.FindServices(criteria)
```

## 健康检查

### ServiceHealth 结构

```go
type ServiceHealth struct {
    ServiceName   string                 `json:"service_name"`
    Status        HealthStatus           `json:"status"`
    LastCheck     time.Time              `json:"last_check"`
    ResponseTime  time.Duration          `json:"response_time"`
    ErrorMessage  string                 `json:"error_message,omitempty"`
    Details       map[string]interface{} `json:"details,omitempty"`
    CheckCount    int64                  `json:"check_count"`
    FailureCount  int64                  `json:"failure_count"`
    SuccessRate   float64                `json:"success_rate"`
}

type HealthStatus int

const (
    HealthStatusUnknown HealthStatus = iota
    HealthStatusHealthy
    HealthStatusUnhealthy
    HealthStatusDegraded
    HealthStatusMaintenance
)

func (hs HealthStatus) String() string {
    switch hs {
    case HealthStatusHealthy:
        return "healthy"
    case HealthStatusUnhealthy:
        return "unhealthy"
    case HealthStatusDegraded:
        return "degraded"
    case HealthStatusMaintenance:
        return "maintenance"
    default:
        return "unknown"
    }
}
```

### HealthCheckConfig 结构

```go
type HealthCheckConfig struct {
    Enabled         bool          `json:"enabled"`
    Interval        time.Duration `json:"interval"`
    Timeout         time.Duration `json:"timeout"`
    FailureThreshold int          `json:"failure_threshold"`
    SuccessThreshold int          `json:"success_threshold"`
    InitialDelay    time.Duration `json:"initial_delay"`
    Method          string        `json:"method,omitempty"`
    Path            string        `json:"path,omitempty"`
    ExpectedStatus  []int         `json:"expected_status,omitempty"`
}
```

### 健康检查示例

```go
// 检查特定服务健康状态
health, err := registry.CheckServiceHealth("netease-music")
if err != nil {
    log.Printf("Failed to check service health: %v", err)
    return
}

log.Printf("Service %s health: %s (%.2f%% success rate)",
    health.ServiceName, health.Status, health.SuccessRate*100)

if health.Status != HealthStatusHealthy {
    log.Printf("Service unhealthy: %s", health.ErrorMessage)
}

// 获取所有不健康的服务
unhealthyServices, err := registry.GetUnhealthyServices()
if err != nil {
    log.Printf("Failed to get unhealthy services: %v", err)
    return
}

for _, serviceName := range unhealthyServices {
    log.Printf("Unhealthy service: %s", serviceName)
}
```

## 统计和监控

### ServiceStats 结构

```go
type ServiceStats struct {
    ServiceName     string        `json:"service_name"`
    CallCount       int64         `json:"call_count"`
    ErrorCount      int64         `json:"error_count"`
    SuccessRate     float64       `json:"success_rate"`
    AverageLatency  time.Duration `json:"average_latency"`
    MaxLatency      time.Duration `json:"max_latency"`
    MinLatency      time.Duration `json:"min_latency"`
    LastCall        time.Time     `json:"last_call"`
    RegisteredAt    time.Time     `json:"registered_at"`
    Uptime          time.Duration `json:"uptime"`
    HealthStatus    HealthStatus  `json:"health_status"`
    
    // 方法级别统计
    MethodStats     map[string]*MethodStats `json:"method_stats"`
}

type MethodStats struct {
    MethodName      string        `json:"method_name"`
    CallCount       int64         `json:"call_count"`
    ErrorCount      int64         `json:"error_count"`
    AverageLatency  time.Duration `json:"average_latency"`
    LastCall        time.Time     `json:"last_call"`
}

type RegistryStats struct {
    TotalServices   int           `json:"total_services"`
    HealthyServices int           `json:"healthy_services"`
    TotalCalls      int64         `json:"total_calls"`
    TotalErrors     int64         `json:"total_errors"`
    AverageLatency  time.Duration `json:"average_latency"`
    Uptime          time.Duration `json:"uptime"`
    
    // 按类型分组的统计
    ServicesByType  map[string]int `json:"services_by_type"`
    
    // 按提供者分组的统计
    ServicesByProvider map[string]int `json:"services_by_provider"`
}
```

### 统计信息获取

```go
// 获取服务统计信息
stats, err := registry.GetServiceStats("audio-processor")
if err != nil {
    log.Printf("Failed to get service stats: %v", err)
    return
}

log.Printf("Service Stats: Calls=%d, Errors=%d, Success=%.2f%%, Latency=%v",
    stats.CallCount, stats.ErrorCount, stats.SuccessRate*100, stats.AverageLatency)

// 获取注册表统计信息
registryStats := registry.GetRegistryStats()
log.Printf("Registry Stats: Services=%d, Healthy=%d, Calls=%d, Errors=%d",
    registryStats.TotalServices, registryStats.HealthyServices,
    registryStats.TotalCalls, registryStats.TotalErrors)
```

## 事件系统

### 服务事件

```go
type ServiceEventType int

const (
    ServiceEventRegistered ServiceEventType = iota
    ServiceEventUnregistered
    ServiceEventHealthChanged
    ServiceEventCallStarted
    ServiceEventCallCompleted
    ServiceEventCallFailed
)

type ServiceEvent struct {
    Type        ServiceEventType       `json:"type"`
    ServiceName string                 `json:"service_name"`
    Timestamp   time.Time              `json:"timestamp"`
    Data        map[string]interface{} `json:"data,omitempty"`
}

type ServiceEventHandler func(event *ServiceEvent)
```

### 事件订阅示例

```go
// 订阅服务注册事件
registry.Subscribe(ServiceEventRegistered, func(event *ServiceEvent) {
    log.Printf("Service registered: %s", event.ServiceName)
})

// 订阅服务健康状态变化事件
registry.Subscribe(ServiceEventHealthChanged, func(event *ServiceEvent) {
    oldStatus := event.Data["old_status"].(HealthStatus)
    newStatus := event.Data["new_status"].(HealthStatus)
    
    log.Printf("Service %s health changed: %s -> %s",
        event.ServiceName, oldStatus, newStatus)
    
    if newStatus == HealthStatusUnhealthy {
        // 处理服务不健康的情况
        handleUnhealthyService(event.ServiceName)
    }
})

// 订阅服务调用失败事件
registry.Subscribe(ServiceEventCallFailed, func(event *ServiceEvent) {
    method := event.Data["method"].(string)
    error := event.Data["error"].(string)
    
    log.Printf("Service call failed: %s.%s - %s",
        event.ServiceName, method, error)
})
```

## 实现类

### DefaultServiceRegistry

```go
type DefaultServiceRegistry struct {
    services      map[string]*registeredService
    providers     map[string]ServiceProvider
    healthChecks  map[string]*healthChecker
    stats         *RegistryStats
    eventHandlers map[ServiceEventType][]ServiceEventHandler
    
    // 配置
    config        *ServiceRegistryConfig
    
    // 并发控制
    mutex         sync.RWMutex
    
    // 生命周期
    ctx           context.Context
    cancel        context.CancelFunc
    wg            sync.WaitGroup
    
    // 日志
    logger        *slog.Logger
}

type registeredService struct {
    descriptor    ServiceDescriptor
    provider      ServiceProvider
    instance      interface{}
    stats         *ServiceStats
    health        *ServiceHealth
    createdAt     time.Time
}

type healthChecker struct {
    serviceName   string
    config        HealthCheckConfig
    health        *ServiceHealth
    ticker        *time.Ticker
    stopChan      chan struct{}
}
```

## 配置选项

### ServiceRegistryConfig 结构

```go
type ServiceRegistryConfig struct {
    // 健康检查配置
    DefaultHealthCheck HealthCheckConfig `yaml:"default_health_check"`
    HealthCheckEnabled bool              `yaml:"health_check_enabled"`
    
    // 统计配置
    StatsEnabled       bool              `yaml:"stats_enabled"`
    StatsInterval      time.Duration     `yaml:"stats_interval"`
    
    // 服务调用配置
    CallTimeout        time.Duration     `yaml:"call_timeout"`
    RetryCount         int               `yaml:"retry_count"`
    RetryDelay         time.Duration     `yaml:"retry_delay"`
    
    // 缓存配置
    CacheEnabled       bool              `yaml:"cache_enabled"`
    CacheSize          int               `yaml:"cache_size"`
    CacheTTL           time.Duration     `yaml:"cache_ttl"`
    
    // 事件配置
    EventsEnabled      bool              `yaml:"events_enabled"`
    EventQueueSize     int               `yaml:"event_queue_size"`
}

func DefaultServiceRegistryConfig() *ServiceRegistryConfig {
    return &ServiceRegistryConfig{
        DefaultHealthCheck: HealthCheckConfig{
            Enabled:          true,
            Interval:         30 * time.Second,
            Timeout:          5 * time.Second,
            FailureThreshold: 3,
            SuccessThreshold: 2,
            InitialDelay:     10 * time.Second,
        },
        HealthCheckEnabled: true,
        StatsEnabled:       true,
        StatsInterval:      10 * time.Second,
        CallTimeout:        30 * time.Second,
        RetryCount:         3,
        RetryDelay:         time.Second,
        CacheEnabled:       true,
        CacheSize:          1000,
        CacheTTL:           5 * time.Minute,
        EventsEnabled:      true,
        EventQueueSize:     1000,
    }
}
```

### YAML配置示例

```yaml
# service-registry.yaml
service_registry:
  # 健康检查配置
  health_check_enabled: true
  default_health_check:
    enabled: true
    interval: "30s"
    timeout: "5s"
    failure_threshold: 3
    success_threshold: 2
    initial_delay: "10s"
  
  # 统计配置
  stats_enabled: true
  stats_interval: "10s"
  
  # 服务调用配置
  call_timeout: "30s"
  retry_count: 3
  retry_delay: "1s"
  
  # 缓存配置
  cache_enabled: true
  cache_size: 1000
  cache_ttl: "5m"
  
  # 事件配置
  events_enabled: true
  event_queue_size: 1000
```

## 最佳实践

### 1. 服务设计

```go
// 定义清晰的服务接口
type MusicSourceService interface {
    Search(query string, options SearchOptions) ([]*Song, error)
    GetSong(id string) (*Song, error)
    GetPlaylist(id string) (*Playlist, error)
    GetLyrics(songID string) (*Lyrics, error)
}

// 实现服务接口
type NeteaseCloudMusicService struct {
    client *http.Client
    config *NeteaseConfig
}

func (n *NeteaseCloudMusicService) Search(query string, options SearchOptions) ([]*Song, error) {
    // 实现搜索逻辑
    return nil, nil
}

// 注册服务时提供完整的描述信息
serviceDesc := ServiceDescriptor{
    Name:        "netease-music",
    Type:        "MusicSourceService",
    Version:     "1.0.0",
    Description: "NetEase Cloud Music service for searching and retrieving music",
    Provider:    "netease-plugin",
    Interface:   reflect.TypeOf((*MusicSourceService)(nil)).Elem(),
    Instance:    musicService,
    Tags:        []string{"music", "streaming", "china"},
    Metadata: map[string]interface{}{
        "region":     "china",
        "rate_limit": 1000,
        "api_version": "v1",
    },
}
```

### 2. 错误处理和重试

```go
// 带重试的服务调用
func callServiceWithRetry(registry ServiceRegistry, serviceName, method string, args ...interface{}) (interface{}, error) {
    var lastErr error
    
    for i := 0; i < 3; i++ {
        result, err := registry.CallService(serviceName, method, args...)
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        
        // 检查是否是可重试的错误
        if !isRetryableError(err) {
            break
        }
        
        // 等待后重试
        time.Sleep(time.Duration(i+1) * time.Second)
    }
    
    return nil, fmt.Errorf("service call failed after retries: %w", lastErr)
}

func isRetryableError(err error) bool {
    // 判断错误是否可重试
    if strings.Contains(err.Error(), "timeout") {
        return true
    }
    if strings.Contains(err.Error(), "connection refused") {
        return true
    }
    return false
}
```

### 3. 服务监控

```go
// 监控服务健康状态
func monitorServiceHealth(registry ServiceRegistry) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        services := registry.ListServices()
        
        for _, service := range services {
            health, err := registry.CheckServiceHealth(service.Name)
            if err != nil {
                log.Printf("Failed to check health for %s: %v", service.Name, err)
                continue
            }
            
            if health.Status != HealthStatusHealthy {
                log.Printf("Service %s is unhealthy: %s", service.Name, health.ErrorMessage)
                
                // 尝试重启服务
                if health.FailureCount > 5 {
                    log.Printf("Attempting to restart service %s", service.Name)
                    // 实现服务重启逻辑
                }
            }
        }
    }
}
```

### 4. 服务发现优化

```go
// 缓存服务实例以提高性能
type CachedServiceRegistry struct {
    registry ServiceRegistry
    cache    map[string]interface{}
    cacheTTL map[string]time.Time
    mutex    sync.RWMutex
}

func (c *CachedServiceRegistry) GetService(name string) (interface{}, error) {
    c.mutex.RLock()
    if service, exists := c.cache[name]; exists {
        if ttl, ok := c.cacheTTL[name]; ok && time.Now().Before(ttl) {
            c.mutex.RUnlock()
            return service, nil
        }
    }
    c.mutex.RUnlock()
    
    // 缓存未命中，从注册表获取
    service, err := c.registry.GetService(name)
    if err != nil {
        return nil, err
    }
    
    // 更新缓存
    c.mutex.Lock()
    c.cache[name] = service
    c.cacheTTL[name] = time.Now().Add(5 * time.Minute)
    c.mutex.Unlock()
    
    return service, nil
}
```

## 相关文档

- [微内核 API](kernel.md)
- [插件管理器 API](plugin-manager.md)
- [事件总线 API](event-bus.md)
- [安全管理器 API](security-manager.md)
- [服务发现架构](../architecture/service-discovery.md)
- [依赖注入指南](../guides/dependency-injection.md)