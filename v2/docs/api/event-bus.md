# 事件总线 API 文档

事件总线是微内核架构中的核心通信组件，提供了解耦的事件驱动通信机制，支持插件间的异步消息传递。

## 接口定义

### EventBus 接口

```go
type EventBus interface {
    // 事件发布
    Publish(eventType string, data interface{}) error
    PublishAsync(eventType string, data interface{}) error
    PublishWithContext(ctx context.Context, eventType string, data interface{}) error
    
    // 事件订阅
    Subscribe(eventType string, handler EventHandler) (SubscriptionID, error)
    SubscribeWithFilter(eventType string, filter EventFilter, handler EventHandler) (SubscriptionID, error)
    SubscribeOnce(eventType string, handler EventHandler) (SubscriptionID, error)
    
    // 订阅管理
    Unsubscribe(subscriptionID SubscriptionID) error
    UnsubscribeAll(eventType string) error
    
    // 事件查询
    GetSubscribers(eventType string) []SubscriptionInfo
    GetEventTypes() []string
    GetEventHistory(eventType string, limit int) ([]*Event, error)
    
    // 监控和统计
    GetStats() *EventBusStats
    GetEventStats(eventType string) (*EventStats, error)
    HealthCheck() error
    
    // 生命周期
    Start() error
    Stop() error
    Shutdown(ctx context.Context) error
}
```

## 核心数据结构

### Event 结构

```go
type Event struct {
    ID        string                 `json:"id"`
    Type      string                 `json:"type"`
    Source    string                 `json:"source"`
    Data      interface{}            `json:"data"`
    Metadata  map[string]interface{} `json:"metadata"`
    Timestamp time.Time              `json:"timestamp"`
    TraceID   string                 `json:"trace_id,omitempty"`
}
```

### EventHandler 类型

```go
type EventHandler func(event *Event) error

// 异步事件处理器
type AsyncEventHandler func(event *Event)

// 带上下文的事件处理器
type ContextEventHandler func(ctx context.Context, event *Event) error
```

### SubscriptionID 和订阅信息

```go
type SubscriptionID string

type SubscriptionInfo struct {
    ID          SubscriptionID `json:"id"`
    EventType   string         `json:"event_type"`
    Subscriber  string         `json:"subscriber"`
    Filter      EventFilter    `json:"filter,omitempty"`
    CreatedAt   time.Time      `json:"created_at"`
    CallCount   int64          `json:"call_count"`
    LastCalled  time.Time      `json:"last_called"`
    ErrorCount  int64          `json:"error_count"`
}
```

## 核心方法

### 事件发布

#### Publish(eventType string, data interface{}) error

同步发布事件，等待所有订阅者处理完成。

**参数:**
- `eventType`: 事件类型
- `data`: 事件数据

**返回值:**
- `error`: 发布失败或处理错误

**示例:**
```go
eventBus := kernel.GetEventBus()

// 发布插件加载事件
if err := eventBus.Publish("plugin.loaded", map[string]interface{}{
    "plugin_name": "audio-processor",
    "version":     "1.0.0",
    "load_time":   time.Now(),
}); err != nil {
    log.Printf("Failed to publish event: %v", err)
}

// 发布音乐播放事件
songData := &Song{
    ID:     "song123",
    Title:  "My Song",
    Artist: "Artist Name",
}

if err := eventBus.Publish("music.play", songData); err != nil {
    log.Printf("Failed to publish music play event: %v", err)
}
```

#### PublishAsync(eventType string, data interface{}) error

异步发布事件，不等待订阅者处理完成。

**参数:**
- `eventType`: 事件类型
- `data`: 事件数据

**返回值:**
- `error`: 发布失败时返回错误

**示例:**
```go
// 异步发布日志事件
if err := eventBus.PublishAsync("system.log", map[string]interface{}{
    "level":   "info",
    "message": "System operation completed",
    "module":  "plugin-manager",
}); err != nil {
    log.Printf("Failed to publish async event: %v", err)
}
```

#### PublishWithContext(ctx context.Context, eventType string, data interface{}) error

带上下文发布事件，支持超时和取消。

**参数:**
- `ctx`: 上下文对象
- `eventType`: 事件类型
- `data`: 事件数据

**返回值:**
- `error`: 发布失败或超时错误

**示例:**
```go
// 带超时的事件发布
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := eventBus.PublishWithContext(ctx, "system.shutdown", nil); err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        log.Printf("Event publish timeout")
    } else {
        log.Printf("Failed to publish event: %v", err)
    }
}
```

### 事件订阅

#### Subscribe(eventType string, handler EventHandler) (SubscriptionID, error)

订阅指定类型的事件。

**参数:**
- `eventType`: 事件类型（支持通配符）
- `handler`: 事件处理函数

**返回值:**
- `SubscriptionID`: 订阅ID，用于取消订阅
- `error`: 订阅失败时返回错误

**示例:**
```go
// 订阅插件事件
subscriptionID, err := eventBus.Subscribe("plugin.*", func(event *Event) error {
    log.Printf("Plugin event: %s, Data: %v", event.Type, event.Data)
    return nil
})
if err != nil {
    log.Printf("Failed to subscribe: %v", err)
    return
}

// 订阅音乐播放事件
musicSubID, err := eventBus.Subscribe("music.play", func(event *Event) error {
    if song, ok := event.Data.(*Song); ok {
        log.Printf("Now playing: %s - %s", song.Artist, song.Title)
        
        // 更新UI显示
        return updateNowPlayingUI(song)
    }
    return nil
})
if err != nil {
    log.Printf("Failed to subscribe to music events: %v", err)
}
```

#### SubscribeWithFilter(eventType string, filter EventFilter, handler EventHandler) (SubscriptionID, error)

带过滤器的事件订阅。

**参数:**
- `eventType`: 事件类型
- `filter`: 事件过滤器
- `handler`: 事件处理函数

**返回值:**
- `SubscriptionID`: 订阅ID
- `error`: 订阅失败时返回错误

**示例:**
```go
// 只订阅特定插件的事件
filter := func(event *Event) bool {
    if data, ok := event.Data.(map[string]interface{}); ok {
        pluginName, exists := data["plugin_name"]
        return exists && pluginName == "audio-processor"
    }
    return false
}

subID, err := eventBus.SubscribeWithFilter("plugin.*", filter, func(event *Event) error {
    log.Printf("Audio processor event: %s", event.Type)
    return nil
})
```

#### SubscribeOnce(eventType string, handler EventHandler) (SubscriptionID, error)

一次性事件订阅，处理一次后自动取消订阅。

**参数:**
- `eventType`: 事件类型
- `handler`: 事件处理函数

**返回值:**
- `SubscriptionID`: 订阅ID
- `error`: 订阅失败时返回错误

**示例:**
```go
// 等待系统初始化完成
subID, err := eventBus.SubscribeOnce("system.initialized", func(event *Event) error {
    log.Printf("System initialization completed")
    
    // 启动应用逻辑
    return startApplication()
})
```

### 订阅管理

#### Unsubscribe(subscriptionID SubscriptionID) error

取消指定的订阅。

**参数:**
- `subscriptionID`: 订阅ID

**返回值:**
- `error`: 取消订阅失败时返回错误

**示例:**
```go
// 取消订阅
if err := eventBus.Unsubscribe(subscriptionID); err != nil {
    log.Printf("Failed to unsubscribe: %v", err)
}
```

#### UnsubscribeAll(eventType string) error

取消指定事件类型的所有订阅。

**参数:**
- `eventType`: 事件类型

**返回值:**
- `error`: 取消订阅失败时返回错误

## 事件过滤器

### EventFilter 接口

```go
type EventFilter func(event *Event) bool

// 预定义过滤器
func SourceFilter(source string) EventFilter {
    return func(event *Event) bool {
        return event.Source == source
    }
}

func DataFieldFilter(field string, value interface{}) EventFilter {
    return func(event *Event) bool {
        if data, ok := event.Data.(map[string]interface{}); ok {
            return data[field] == value
        }
        return false
    }
}

func TimeRangeFilter(start, end time.Time) EventFilter {
    return func(event *Event) bool {
        return event.Timestamp.After(start) && event.Timestamp.Before(end)
    }
}

// 组合过滤器
func AndFilter(filters ...EventFilter) EventFilter {
    return func(event *Event) bool {
        for _, filter := range filters {
            if !filter(event) {
                return false
            }
        }
        return true
    }
}

func OrFilter(filters ...EventFilter) EventFilter {
    return func(event *Event) bool {
        for _, filter := range filters {
            if filter(event) {
                return true
            }
        }
        return false
    }
}
```

### 过滤器使用示例

```go
// 只订阅来自特定插件的错误事件
errorFilter := AndFilter(
    SourceFilter("audio-processor"),
    DataFieldFilter("level", "error"),
)

subID, err := eventBus.SubscribeWithFilter("plugin.log", errorFilter, func(event *Event) error {
    log.Printf("Audio processor error: %v", event.Data)
    return handleAudioProcessorError(event)
})
```

## 事件类型和常量

### 系统事件

```go
const (
    // 系统生命周期事件
    EventSystemStarting    = "system.starting"
    EventSystemStarted     = "system.started"
    EventSystemStopping    = "system.stopping"
    EventSystemStopped     = "system.stopped"
    EventSystemError       = "system.error"
    
    // 插件生命周期事件
    EventPluginLoading     = "plugin.loading"
    EventPluginLoaded      = "plugin.loaded"
    EventPluginStarting    = "plugin.starting"
    EventPluginStarted     = "plugin.started"
    EventPluginStopping    = "plugin.stopping"
    EventPluginStopped     = "plugin.stopped"
    EventPluginError       = "plugin.error"
    EventPluginReloaded    = "plugin.reloaded"
    
    // 音乐播放事件
    EventMusicPlay         = "music.play"
    EventMusicPause        = "music.pause"
    EventMusicStop         = "music.stop"
    EventMusicNext         = "music.next"
    EventMusicPrevious     = "music.previous"
    EventMusicSeek         = "music.seek"
    EventMusicVolumeChange = "music.volume_change"
    
    // UI事件
    EventUIShow            = "ui.show"
    EventUIHide            = "ui.hide"
    EventUIUpdate          = "ui.update"
    EventUIThemeChange     = "ui.theme_change"
    
    // 配置事件
    EventConfigChanged     = "config.changed"
    EventConfigReloaded    = "config.reloaded"
    
    // 网络事件
    EventNetworkConnected    = "network.connected"
    EventNetworkDisconnected = "network.disconnected"
    EventNetworkError        = "network.error"
)
```

## 统计和监控

### EventBusStats 结构

```go
type EventBusStats struct {
    TotalEvents       int64         `json:"total_events"`
    TotalSubscribers  int           `json:"total_subscribers"`
    EventTypes        int           `json:"event_types"`
    AverageLatency    time.Duration `json:"average_latency"`
    ErrorRate         float64       `json:"error_rate"`
    QueueSize         int           `json:"queue_size"`
    MaxQueueSize      int           `json:"max_queue_size"`
    DroppedEvents     int64         `json:"dropped_events"`
    Uptime            time.Duration `json:"uptime"`
}

// 获取事件总线统计
stats := eventBus.GetStats()
log.Printf("EventBus Stats: Events=%d, Subscribers=%d, Latency=%v, ErrorRate=%.2f%%",
    stats.TotalEvents, stats.TotalSubscribers, stats.AverageLatency, stats.ErrorRate*100)
```

### EventStats 结构

```go
type EventStats struct {
    EventType         string        `json:"event_type"`
    PublishCount      int64         `json:"publish_count"`
    SubscriberCount   int           `json:"subscriber_count"`
    AverageLatency    time.Duration `json:"average_latency"`
    MaxLatency        time.Duration `json:"max_latency"`
    MinLatency        time.Duration `json:"min_latency"`
    ErrorCount        int64         `json:"error_count"`
    LastPublished     time.Time     `json:"last_published"`
}

// 获取特定事件类型的统计
stats, err := eventBus.GetEventStats("music.play")
if err != nil {
    log.Printf("Failed to get event stats: %v", err)
    return
}

log.Printf("Music play events: Count=%d, Subscribers=%d, AvgLatency=%v",
    stats.PublishCount, stats.SubscriberCount, stats.AverageLatency)
```

## 实现类

### DefaultEventBus

```go
type DefaultEventBus struct {
    subscribers   map[string][]*subscription
    eventHistory  map[string][]*Event
    stats         *EventBusStats
    
    // 配置
    config        *EventBusConfig
    
    // 并发控制
    mutex         sync.RWMutex
    
    // 异步处理
    eventQueue    chan *eventTask
    workerPool    *WorkerPool
    
    // 生命周期
    ctx           context.Context
    cancel        context.CancelFunc
    wg            sync.WaitGroup
    
    // 日志
    logger        *slog.Logger
}

type subscription struct {
    id        SubscriptionID
    eventType string
    handler   EventHandler
    filter    EventFilter
    once      bool
    info      *SubscriptionInfo
}

type eventTask struct {
    event       *Event
    subscribers []*subscription
    resultChan  chan error
}
```

### 构造函数

```go
func NewEventBus(logger *slog.Logger) EventBus {
    return NewEventBusWithConfig(DefaultEventBusConfig(), logger)
}

func NewEventBusWithConfig(config *EventBusConfig, logger *slog.Logger) EventBus {
    ctx, cancel := context.WithCancel(context.Background())
    
    eb := &DefaultEventBus{
        subscribers:  make(map[string][]*subscription),
        eventHistory: make(map[string][]*Event),
        stats:        &EventBusStats{},
        config:       config,
        eventQueue:   make(chan *eventTask, config.QueueSize),
        ctx:          ctx,
        cancel:       cancel,
        logger:       logger,
    }
    
    // 创建工作池
    eb.workerPool = NewWorkerPool(config.WorkerCount, eb.processEventTask)
    
    return eb
}
```

## 配置选项

### EventBusConfig 结构

```go
type EventBusConfig struct {
    // 队列配置
    QueueSize       int           `yaml:"queue_size"`
    WorkerCount     int           `yaml:"worker_count"`
    
    // 超时配置
    PublishTimeout  time.Duration `yaml:"publish_timeout"`
    HandlerTimeout  time.Duration `yaml:"handler_timeout"`
    
    // 重试配置
    RetryCount      int           `yaml:"retry_count"`
    RetryDelay      time.Duration `yaml:"retry_delay"`
    
    // 历史记录
    HistoryEnabled  bool          `yaml:"history_enabled"`
    HistorySize     int           `yaml:"history_size"`
    
    // 监控配置
    MetricsEnabled  bool          `yaml:"metrics_enabled"`
    MetricsInterval time.Duration `yaml:"metrics_interval"`
    
    // 错误处理
    ErrorHandler    func(error)   `yaml:"-"`
    PanicRecovery   bool          `yaml:"panic_recovery"`
}

func DefaultEventBusConfig() *EventBusConfig {
    return &EventBusConfig{
        QueueSize:       1000,
        WorkerCount:     10,
        PublishTimeout:  30 * time.Second,
        HandlerTimeout:  10 * time.Second,
        RetryCount:      3,
        RetryDelay:      time.Second,
        HistoryEnabled:  true,
        HistorySize:     100,
        MetricsEnabled:  true,
        MetricsInterval: 10 * time.Second,
        PanicRecovery:   true,
    }
}
```

### YAML配置示例

```yaml
# event-bus.yaml
event_bus:
  # 队列配置
  queue_size: 1000
  worker_count: 10
  
  # 超时配置
  publish_timeout: "30s"
  handler_timeout: "10s"
  
  # 重试配置
  retry_count: 3
  retry_delay: "1s"
  
  # 历史记录
  history_enabled: true
  history_size: 100
  
  # 监控配置
  metrics_enabled: true
  metrics_interval: "10s"
  
  # 错误处理
  panic_recovery: true
  
  # 日志配置
  log_events: false
  log_level: "info"
```

## 错误处理

### 错误类型

```go
type EventBusError struct {
    Type      EventBusErrorType `json:"type"`
    Code      string            `json:"code"`
    Message   string            `json:"message"`
    EventType string            `json:"event_type,omitempty"`
    Cause     error             `json:"cause,omitempty"`
    Timestamp time.Time         `json:"timestamp"`
}

type EventBusErrorType int

const (
    EventBusErrorTypePublish EventBusErrorType = iota
    EventBusErrorTypeSubscribe
    EventBusErrorTypeHandler
    EventBusErrorTypeTimeout
    EventBusErrorTypeQueue
)
```

### 错误处理策略

```go
func setupEventBusErrorHandling(eventBus EventBus) {
    config := &EventBusConfig{
        ErrorHandler: func(err error) {
            if ebErr, ok := err.(*EventBusError); ok {
                switch ebErr.Type {
                case EventBusErrorTypeHandler:
                    log.Printf("Event handler error: %s - %v", ebErr.EventType, ebErr.Cause)
                    // 可以选择重试或跳过
                case EventBusErrorTypeTimeout:
                    log.Printf("Event processing timeout: %s", ebErr.EventType)
                    // 记录超时事件，可能需要调整超时时间
                case EventBusErrorTypeQueue:
                    log.Printf("Event queue error: %v", ebErr.Cause)
                    // 队列满或其他队列问题
                default:
                    log.Printf("EventBus error: %v", err)
                }
            }
        },
        PanicRecovery: true,
    }
}
```

## 最佳实践

### 1. 事件设计

```go
// 好的事件设计
type MusicPlayEvent struct {
    Song      *Song     `json:"song"`
    Playlist  *Playlist `json:"playlist,omitempty"`
    Position  int       `json:"position"`
    Timestamp time.Time `json:"timestamp"`
}

// 发布结构化事件
eventBus.Publish("music.play", &MusicPlayEvent{
    Song:      currentSong,
    Playlist:  currentPlaylist,
    Position:  playPosition,
    Timestamp: time.Now(),
})

// 订阅时进行类型检查
eventBus.Subscribe("music.play", func(event *Event) error {
    if playEvent, ok := event.Data.(*MusicPlayEvent); ok {
        return handleMusicPlay(playEvent)
    }
    return fmt.Errorf("invalid event data type")
})
```

### 2. 错误处理

```go
// 事件处理器中的错误处理
eventBus.Subscribe("plugin.error", func(event *Event) error {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Panic in event handler: %v", r)
        }
    }()
    
    // 处理事件
    if err := processPluginError(event); err != nil {
        // 记录错误但不阻止其他处理器
        log.Printf("Failed to process plugin error event: %v", err)
        return nil // 返回nil避免影响其他订阅者
    }
    
    return nil
})
```

### 3. 性能优化

```go
// 使用异步发布提高性能
func publishLogEvent(level, message string) {
    // 对于非关键事件使用异步发布
    eventBus.PublishAsync("system.log", map[string]interface{}{
        "level":     level,
        "message":   message,
        "timestamp": time.Now(),
    })
}

// 使用过滤器减少不必要的处理
highPriorityFilter := func(event *Event) bool {
    if data, ok := event.Data.(map[string]interface{}); ok {
        priority, exists := data["priority"]
        return exists && priority == "high"
    }
    return false
}

eventBus.SubscribeWithFilter("system.*", highPriorityFilter, handleHighPriorityEvent)
```

### 4. 监控和调试

```go
// 定期监控事件总线状态
func monitorEventBus(eventBus EventBus) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        stats := eventBus.GetStats()
        
        // 检查队列积压
        if float64(stats.QueueSize)/float64(stats.MaxQueueSize) > 0.8 {
            log.Printf("EventBus queue is %d%% full", 
                int(float64(stats.QueueSize)/float64(stats.MaxQueueSize)*100))
        }
        
        // 检查错误率
        if stats.ErrorRate > 0.05 { // 5%
            log.Printf("EventBus error rate is high: %.2f%%", stats.ErrorRate*100)
        }
        
        // 检查延迟
        if stats.AverageLatency > 100*time.Millisecond {
            log.Printf("EventBus average latency is high: %v", stats.AverageLatency)
        }
    }
}
```

## 相关文档

- [微内核 API](kernel.md)
- [插件管理器 API](plugin-manager.md)
- [服务注册表 API](service-registry.md)
- [事件驱动架构](../architecture/event-driven.md)
- [性能优化指南](../architecture/performance.md)