package event

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EventBus 事件总线接口
type EventBus interface {
	// 订阅管理
	Subscribe(eventType EventType, handler EventHandler, options ...SubscribeOption) (*Subscription, error)
	SubscribeWithFilter(eventType EventType, handler EventHandler, filter EventFilter, options ...SubscribeOption) (*Subscription, error)
	Unsubscribe(subscriptionID string) error
	UnsubscribeAll(eventType EventType) error

	// 事件发布
	Publish(ctx context.Context, event Event) error
	PublishAsync(ctx context.Context, event Event) error
	PublishSync(ctx context.Context, event Event) error
	PublishWithPriority(ctx context.Context, event Event, priority EventPriority) error

	// 事件类型管理
	RegisterEventType(eventType EventType) error
	UnregisterEventType(eventType EventType) error
	GetRegisteredEventTypes() []EventType

	// 统计和监控
	GetSubscriptionCount(eventType EventType) int
	GetTotalSubscriptions() int
	GetEventStats() *EventStats
	GetStats() *EventStats

	// 生命周期管理
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool
}

// SubscribeOption 订阅选项
type SubscribeOption func(*SubscribeOptions)

// SubscribeOptions 订阅配置
type SubscribeOptions struct {
	Priority int
	Async    bool
	Timeout  time.Duration
	Group    string
	Filter   EventFilter
}

// WithPriority 设置订阅优先级
func WithPriority(priority int) SubscribeOption {
	return func(opts *SubscribeOptions) {
		opts.Priority = priority
	}
}

// WithAsync 设置异步处理
func WithAsync(async bool) SubscribeOption {
	return func(opts *SubscribeOptions) {
		opts.Async = async
	}
}

// WithTimeout 设置处理超时时间
func WithTimeout(timeout time.Duration) SubscribeOption {
	return func(opts *SubscribeOptions) {
		opts.Timeout = timeout
	}
}

// WithGroup 设置订阅者组
func WithGroup(group string) SubscribeOption {
	return func(opts *SubscribeOptions) {
		opts.Group = group
	}
}

// WithFilter 设置事件过滤器
func WithFilter(filter EventFilter) SubscribeOption {
	return func(opts *SubscribeOptions) {
		opts.Filter = filter
	}
}

// EventStats 事件统计信息
type EventStats struct {
	TotalEvents      int64            `json:"total_events"`
	TotalSubscribers int              `json:"total_subscribers"`
	EventCounts      map[EventType]int64 `json:"event_counts"`
	ErrorCounts      map[EventType]int64 `json:"error_counts"`
	LastEventTime    time.Time        `json:"last_event_time"`
}

// DefaultEventBus 默认事件总线实现
type DefaultEventBus struct {
	logger *slog.Logger

	// 订阅管理
	subscriptions map[EventType][]*Subscription
	subscriberMap map[string]*Subscription
	mutex         sync.RWMutex

	// 事件类型注册
	registeredTypes map[EventType]bool
	typesMutex      sync.RWMutex

	// 异步处理 - 优化版本
	eventQueue     chan *eventTask
	batchQueue     chan []*eventTask
	workerPool     chan struct{}
	maxWorkers     int
	minWorkers     int
	currentWorkers int
	queueSize      int
	batchSize      int
	batchTimeout   time.Duration

	// 动态扩缩容
	scaleUpThreshold   float64
	scaleDownThreshold float64
	lastScaleTime      time.Time
	scaleCooldown      time.Duration

	// 性能监控
	processingTimes []time.Duration
	timesMutex      sync.RWMutex

	// 统计信息
	stats      *EventStats
	statsMutex sync.RWMutex

	// 生命周期管理
	ctx    context.Context
	cancel context.CancelFunc
	running bool
	closed  bool
	wg     sync.WaitGroup

	// 批量处理
	batchProcessor *BatchProcessor
}

// eventTask 事件处理任务
type eventTask struct {
	ctx          context.Context
	event        Event
	subscription *Subscription
	priority     EventPriority
	startTime    time.Time
}

// BatchProcessor 批量处理器
type BatchProcessor struct {
	batchSize    int
	batchTimeout time.Duration
	pendingTasks []*eventTask
	lastFlush    time.Time
	mutex        sync.Mutex
}

// WorkerPoolConfig 工作池配置
type WorkerPoolConfig struct {
	MinWorkers         int           `json:"min_workers"`
	MaxWorkers         int           `json:"max_workers"`
	QueueSize          int           `json:"queue_size"`
	BatchSize          int           `json:"batch_size"`
	BatchTimeout       time.Duration `json:"batch_timeout"`
	ScaleUpThreshold   float64       `json:"scale_up_threshold"`
	ScaleDownThreshold float64       `json:"scale_down_threshold"`
	ScaleCooldown      time.Duration `json:"scale_cooldown"`
}

// DefaultWorkerPoolConfig 默认工作池配置
func DefaultWorkerPoolConfig() *WorkerPoolConfig {
	return &WorkerPoolConfig{
		MinWorkers:         2,
		MaxWorkers:         20,
		QueueSize:          1000,
		BatchSize:          10,
		BatchTimeout:       time.Millisecond * 100,
		ScaleUpThreshold:   0.8,  // 80%队列使用率时扩容
		ScaleDownThreshold: 0.2,  // 20%队列使用率时缩容
		ScaleCooldown:      time.Second * 30,
	}
}

// NewDefaultEventBus 创建默认事件总线（便捷函数）
func NewDefaultEventBus() EventBus {
	return NewEventBus(slog.Default())
}

// NewEventBus 创建新的事件总线
func NewEventBus(logger *slog.Logger) EventBus {
	return NewEventBusWithConfig(logger, DefaultWorkerPoolConfig())
}

// NewEventBusWithConfig 使用配置创建事件总线
func NewEventBusWithConfig(logger *slog.Logger, config *WorkerPoolConfig) EventBus {
	ctx, cancel := context.WithCancel(context.Background())

	return &DefaultEventBus{
		logger:          logger,
		subscriptions:   make(map[EventType][]*Subscription),
		subscriberMap:   make(map[string]*Subscription),
		registeredTypes: make(map[EventType]bool),
		eventQueue:      make(chan *eventTask, config.QueueSize),
		batchQueue:      make(chan []*eventTask, config.QueueSize/config.BatchSize),
		workerPool:      make(chan struct{}, config.MaxWorkers),
		maxWorkers:      config.MaxWorkers,
		minWorkers:      config.MinWorkers,
		currentWorkers:  0,
		queueSize:       config.QueueSize,
		batchSize:       config.BatchSize,
		batchTimeout:    config.BatchTimeout,
		scaleUpThreshold:   config.ScaleUpThreshold,
		scaleDownThreshold: config.ScaleDownThreshold,
		scaleCooldown:      config.ScaleCooldown,
		processingTimes:    make([]time.Duration, 0, 1000),
		stats: &EventStats{
			EventCounts: make(map[EventType]int64),
			ErrorCounts: make(map[EventType]int64),
		},
		ctx:    ctx,
		cancel: cancel,
		batchProcessor: &BatchProcessor{
			batchSize:    config.BatchSize,
			batchTimeout: config.BatchTimeout,
			pendingTasks: make([]*eventTask, 0, config.BatchSize),
			lastFlush:    time.Now(),
		},
	}
}

// Start 启动事件总线
func (eb *DefaultEventBus) Start(ctx context.Context) error {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	if eb.running {
		return fmt.Errorf("event bus is already running")
	}

	// 如果之前已关闭，需要重新创建channel
	if eb.closed {
		eb.eventQueue = make(chan *eventTask, eb.queueSize)
		eb.batchQueue = make(chan []*eventTask, eb.queueSize/eb.batchSize)
		eb.ctx, eb.cancel = context.WithCancel(context.Background())
		eb.closed = false
	}

	// 启动最小数量的工作协程
	for i := 0; i < eb.minWorkers; i++ {
		eb.wg.Add(1)
		go eb.worker()
		eb.currentWorkers++
	}

	// 启动批量处理器
	eb.wg.Add(1)
	go eb.batchWorker()

	// 启动动态扩缩容管理器
	eb.wg.Add(1)
	go eb.scaleManager()

	// 启动性能监控器
	eb.wg.Add(1)
	go eb.performanceMonitor()

	eb.running = true
	eb.logger.Info("Event bus started", 
		"min_workers", eb.minWorkers,
		"max_workers", eb.maxWorkers, 
		"queue_size", eb.queueSize,
		"batch_size", eb.batchSize)
	return nil
}

// Stop 停止事件总线
func (eb *DefaultEventBus) Stop(ctx context.Context) error {
	eb.mutex.Lock()
	if !eb.running {
		eb.mutex.Unlock()
		return fmt.Errorf("event bus is not running")
	}
	eb.running = false
	eb.mutex.Unlock()

	// 取消上下文，停止所有工作协程
	eb.cancel()

	// 安全关闭所有队列
	eb.mutex.Lock()
	if !eb.closed {
		close(eb.eventQueue)
		close(eb.batchQueue) // 也需要关闭batchQueue
		eb.closed = true
	}
	eb.mutex.Unlock()

	// 使用带超时的等待，避免无限等待
	done := make(chan struct{})
	go func() {
		eb.wg.Wait()
		close(done)
	}()

	// 减少超时时间，避免测试卡住
	timeout := 5 * time.Second
	if ctx != nil {
		if deadline, ok := ctx.Deadline(); ok {
			if remaining := time.Until(deadline); remaining < timeout {
				timeout = remaining - 100*time.Millisecond // 留一点缓冲时间
				if timeout <= 0 {
					timeout = 100 * time.Millisecond
				}
			}
		}
	}

	select {
	case <-done:
		eb.logger.Info("Event bus stopped")
		return nil
	case <-time.After(timeout):
		eb.logger.Warn("Event bus stop timeout, some goroutines may still be running")
		return fmt.Errorf("event bus stop timeout")
	case <-ctx.Done():
		eb.logger.Warn("Event bus stop cancelled by context")
		return ctx.Err()
	}
}

// IsRunning 检查事件总线是否运行中
func (eb *DefaultEventBus) IsRunning() bool {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()
	return eb.running
}

// worker 工作协程
func (eb *DefaultEventBus) worker() {
	defer eb.wg.Done()

	for {
		select {
		case task, ok := <-eb.eventQueue:
			if !ok {
				return // 队列已关闭
			}
			eb.processEventTask(task)
		case <-eb.ctx.Done():
			return // 上下文已取消
		}
	}
}

// processEventTask 处理事件任务
func (eb *DefaultEventBus) processEventTask(task *eventTask) {
	startTime := time.Now()
	defer func() {
		processingTime := time.Since(startTime)
		eb.recordProcessingTime(processingTime)
		
		if r := recover(); r != nil {
			eb.logger.Error("Event handler panicked",
				"event_type", task.event.GetType(),
				"subscription_id", task.subscription.ID,
				"panic", r,
				"processing_time", processingTime)
			eb.incrementErrorCount(task.event.GetType())
		}
		// 移除workerPool的阻塞读取，因为它会导致死锁
		// <-eb.workerPool // 释放工作协程槽位
	}()

	// 应用过滤器
	if task.subscription.Filter != nil && !task.subscription.Filter(task.event) {
		return
	}

	// 执行事件处理器
	if err := task.subscription.Handler(task.ctx, task.event); err != nil {
		eb.logger.Error("Event handler failed",
			"event_type", task.event.GetType(),
			"subscription_id", task.subscription.ID,
			"error", err,
			"processing_time", time.Since(startTime))
		eb.incrementErrorCount(task.event.GetType())
	}
}

// batchWorker 批量处理工作协程
func (eb *DefaultEventBus) batchWorker() {
	defer eb.wg.Done()

	ticker := time.NewTicker(eb.batchTimeout)
	defer ticker.Stop()

	for {
		select {
		case batch, ok := <-eb.batchQueue:
			if !ok {
				return // 队列已关闭
			}
			eb.processBatch(batch)
		case <-ticker.C:
			eb.flushPendingTasks()
		case <-eb.ctx.Done():
			return
		}
	}
}

// processBatch 处理批量任务
func (eb *DefaultEventBus) processBatch(batch []*eventTask) {
	if len(batch) == 0 {
		return
	}

	startTime := time.Now()
	eb.logger.Debug("Processing event batch", "batch_size", len(batch))

	// 并行处理批量任务
	var wg sync.WaitGroup
	for _, task := range batch {
		wg.Add(1)
		go func(t *eventTask) {
			defer wg.Done()
			eb.processEventTaskDirect(t)
		}(task)
	}
	wg.Wait()

	processingTime := time.Since(startTime)
	eb.logger.Debug("Batch processed", 
		"batch_size", len(batch),
		"processing_time", processingTime)
}

// processEventTaskDirect 直接处理事件任务（不使用工作池）
func (eb *DefaultEventBus) processEventTaskDirect(task *eventTask) {
	startTime := time.Now()
	defer func() {
		processingTime := time.Since(startTime)
		eb.recordProcessingTime(processingTime)
		
		if r := recover(); r != nil {
			eb.logger.Error("Event handler panicked",
				"event_type", task.event.GetType(),
				"subscription_id", task.subscription.ID,
				"panic", r,
				"processing_time", processingTime)
			eb.incrementErrorCount(task.event.GetType())
		}
	}()

	// 应用过滤器
	if task.subscription.Filter != nil && !task.subscription.Filter(task.event) {
		return
	}

	// 执行事件处理器
	if err := task.subscription.Handler(task.ctx, task.event); err != nil {
		eb.logger.Error("Event handler failed",
			"event_type", task.event.GetType(),
			"subscription_id", task.subscription.ID,
			"error", err,
			"processing_time", time.Since(startTime))
		eb.incrementErrorCount(task.event.GetType())
	}
}

// flushPendingTasks 刷新待处理任务
func (eb *DefaultEventBus) flushPendingTasks() {
	eb.batchProcessor.mutex.Lock()
	defer eb.batchProcessor.mutex.Unlock()

	if len(eb.batchProcessor.pendingTasks) == 0 {
		return
	}

	// 检查EventBus是否已关闭
	eb.mutex.RLock()
	closed := eb.closed
	eb.mutex.RUnlock()

	if closed {
		// EventBus已关闭，直接处理剩余任务
		batch := make([]*eventTask, len(eb.batchProcessor.pendingTasks))
		copy(batch, eb.batchProcessor.pendingTasks)
		eb.batchProcessor.pendingTasks = eb.batchProcessor.pendingTasks[:0]
		go eb.processBatch(batch)
		return
	}

	// 复制待处理任务
	batch := make([]*eventTask, len(eb.batchProcessor.pendingTasks))
	copy(batch, eb.batchProcessor.pendingTasks)

	// 清空待处理任务
	eb.batchProcessor.pendingTasks = eb.batchProcessor.pendingTasks[:0]
	eb.batchProcessor.lastFlush = time.Now()

	// 发送到批量队列
	select {
	case eb.batchQueue <- batch:
		// 成功发送
	default:
		// 队列满，直接处理
		go eb.processBatch(batch)
	}
}

// scaleManager 动态扩缩容管理器
func (eb *DefaultEventBus) scaleManager() {
	defer eb.wg.Done()

	ticker := time.NewTicker(time.Second * 10) // 每10秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			eb.checkAndScale()
		case <-eb.ctx.Done():
			return
		}
	}
}

// checkAndScale 检查并执行扩缩容
func (eb *DefaultEventBus) checkAndScale() {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	// 检查冷却时间
	if time.Since(eb.lastScaleTime) < eb.scaleCooldown {
		return
	}

	// 计算队列使用率
	queueUtilization := float64(len(eb.eventQueue)) / float64(eb.queueSize)

	// 扩容检查
	if queueUtilization > eb.scaleUpThreshold && eb.currentWorkers < eb.maxWorkers {
		newWorkers := min(eb.maxWorkers-eb.currentWorkers, 2) // 每次最多增加2个工作协程
		for i := 0; i < newWorkers; i++ {
			eb.wg.Add(1)
			go eb.worker()
			eb.currentWorkers++
		}
		eb.lastScaleTime = time.Now()
		eb.logger.Info("Scaled up workers", 
			"new_workers", newWorkers,
			"current_workers", eb.currentWorkers,
			"queue_utilization", queueUtilization)
	}

	// 缩容检查
	if queueUtilization < eb.scaleDownThreshold && eb.currentWorkers > eb.minWorkers {
		// 缩容逻辑比较复杂，这里简化处理
		// 实际实现中可能需要优雅地停止工作协程
		eb.lastScaleTime = time.Now()
		eb.logger.Debug("Scale down condition met", 
			"current_workers", eb.currentWorkers,
			"queue_utilization", queueUtilization)
	}
}

// performanceMonitor 性能监控器
func (eb *DefaultEventBus) performanceMonitor() {
	defer eb.wg.Done()

	ticker := time.NewTicker(time.Minute) // 每分钟记录一次性能指标
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			eb.recordPerformanceMetrics()
		case <-eb.ctx.Done():
			return
		}
	}
}

// recordPerformanceMetrics 记录性能指标
func (eb *DefaultEventBus) recordPerformanceMetrics() {
	eb.timesMutex.RLock()
	processingTimes := make([]time.Duration, len(eb.processingTimes))
	copy(processingTimes, eb.processingTimes)
	eb.timesMutex.RUnlock()

	if len(processingTimes) == 0 {
		return
	}

	// 计算平均处理时间
	var total time.Duration
	for _, t := range processingTimes {
		total += t
	}
	avgTime := total / time.Duration(len(processingTimes))

	// 计算队列使用率
	queueUtilization := float64(len(eb.eventQueue)) / float64(eb.queueSize)

	// 计算工作协程使用率
	workerUtilization := float64(eb.currentWorkers) / float64(eb.maxWorkers)

	eb.logger.Debug("Performance metrics",
		"avg_processing_time", avgTime,
		"queue_utilization", queueUtilization,
		"worker_utilization", workerUtilization,
		"current_workers", eb.currentWorkers,
		"processed_events", len(processingTimes))

	// 清理旧的处理时间记录
	eb.timesMutex.Lock()
	if len(eb.processingTimes) > 1000 {
		eb.processingTimes = eb.processingTimes[len(eb.processingTimes)-500:] // 保留最近500条记录
	}
	eb.timesMutex.Unlock()
}

// recordProcessingTime 记录处理时间
func (eb *DefaultEventBus) recordProcessingTime(duration time.Duration) {
	eb.timesMutex.Lock()
	defer eb.timesMutex.Unlock()
	eb.processingTimes = append(eb.processingTimes, duration)
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Subscribe 订阅事件
func (eb *DefaultEventBus) Subscribe(eventType EventType, handler EventHandler, options ...SubscribeOption) (*Subscription, error) {
	return eb.SubscribeWithFilter(eventType, handler, nil, options...)
}

// SubscribeWithFilter 带过滤器的事件订阅
func (eb *DefaultEventBus) SubscribeWithFilter(eventType EventType, handler EventHandler, filter EventFilter, options ...SubscribeOption) (*Subscription, error) {
	if handler == nil {
		return nil, fmt.Errorf("event handler cannot be nil")
	}

	// 应用订阅选项
	opts := &SubscribeOptions{
		Priority: int(PriorityNormal),
		Async:    true,
		Timeout:  30 * time.Second,
	}
	for _, option := range options {
		option(opts)
	}

	// 如果选项中有过滤器，使用选项中的过滤器
	if opts.Filter != nil {
		filter = opts.Filter
	}

	// 创建订阅
	subscription := &Subscription{
		ID:       uuid.New().String(),
		Type:     eventType,
		Handler:  handler,
		Filter:   filter,
		Priority: opts.Priority,
		Async:    opts.Async,
		CreatedAt: time.Now(),
	}

	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	// 添加到订阅列表
	eb.subscriptions[eventType] = append(eb.subscriptions[eventType], subscription)
	eb.subscriberMap[subscription.ID] = subscription

	// 按优先级排序
	eb.sortSubscriptions(eventType)

	eb.logger.Debug("Event subscription added",
		"event_type", eventType,
		"subscription_id", subscription.ID,
		"priority", subscription.Priority)

	return subscription, nil
}

// sortSubscriptions 按优先级排序订阅者
func (eb *DefaultEventBus) sortSubscriptions(eventType EventType) {
	subscriptions := eb.subscriptions[eventType]
	sort.Slice(subscriptions, func(i, j int) bool {
		return subscriptions[i].Priority > subscriptions[j].Priority // 高优先级在前
	})
}

// Unsubscribe 取消订阅
func (eb *DefaultEventBus) Unsubscribe(subscriptionID string) error {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	subscription, exists := eb.subscriberMap[subscriptionID]
	if !exists {
		return fmt.Errorf("subscription not found: %s", subscriptionID)
	}

	// 从订阅列表中移除
	subscriptions := eb.subscriptions[subscription.Type]
	for i, sub := range subscriptions {
		if sub.ID == subscriptionID {
			eb.subscriptions[subscription.Type] = append(subscriptions[:i], subscriptions[i+1:]...)
			break
		}
	}

	// 从订阅者映射中移除
	delete(eb.subscriberMap, subscriptionID)

	eb.logger.Debug("Event subscription removed",
		"event_type", subscription.Type,
		"subscription_id", subscriptionID)

	return nil
}

// UnsubscribeAll 取消指定事件类型的所有订阅
func (eb *DefaultEventBus) UnsubscribeAll(eventType EventType) error {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	subscriptions := eb.subscriptions[eventType]
	for _, subscription := range subscriptions {
		delete(eb.subscriberMap, subscription.ID)
	}

	delete(eb.subscriptions, eventType)

	eb.logger.Debug("All subscriptions removed for event type", "event_type", eventType)
	return nil
}

// Publish 同步发布事件
func (eb *DefaultEventBus) Publish(ctx context.Context, event Event) error {
	return eb.publishEvent(ctx, event, PriorityNormal, false)
}

// PublishAsync 异步发布事件
func (eb *DefaultEventBus) PublishAsync(ctx context.Context, event Event) error {
	return eb.publishEvent(ctx, event, PriorityNormal, true)
}

// PublishWithPriority 带优先级发布事件
func (eb *DefaultEventBus) PublishWithPriority(ctx context.Context, event Event, priority EventPriority) error {
	return eb.publishEvent(ctx, event, priority, true)
}

// publishEvent 发布事件的内部实现
func (eb *DefaultEventBus) publishEvent(ctx context.Context, event Event, priority EventPriority, async bool) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	// 检查事件总线是否正在运行
	if !eb.IsRunning() {
		return fmt.Errorf("event bus is not running")
	}

	subscriptions := eb.getSubscriptions(event.GetType())
	if len(subscriptions) == 0 {
		eb.logger.Debug("No subscribers for event type", "event_type", event.GetType())
		return nil
	}

	// 更新统计信息
	eb.incrementEventCount(event.GetType())

	// 处理订阅者
	for _, subscription := range subscriptions {
		if async || subscription.Async {
			eb.handleAsyncSubscription(ctx, event, subscription, priority)
		} else {
			if err := eb.handleSyncSubscription(ctx, event, subscription); err != nil {
				return err
			}
		}
	}

	return nil
}

// RegisterEventType 注册事件类型
func (eb *DefaultEventBus) RegisterEventType(eventType EventType) error {
	eb.typesMutex.Lock()
	defer eb.typesMutex.Unlock()

	if eb.registeredTypes[eventType] {
		return fmt.Errorf("event type %s is already registered", eventType)
	}

	eb.registeredTypes[eventType] = true
	eb.logger.Debug("Event type registered", "event_type", eventType)
	return nil
}

// UnregisterEventType 注销事件类型
func (eb *DefaultEventBus) UnregisterEventType(eventType EventType) error {
	eb.typesMutex.Lock()
	defer eb.typesMutex.Unlock()

	delete(eb.registeredTypes, eventType)
	eb.logger.Debug("Event type unregistered", "event_type", eventType)
	return nil
}

// GetRegisteredEventTypes 获取已注册的事件类型
func (eb *DefaultEventBus) GetRegisteredEventTypes() []EventType {
	eb.typesMutex.RLock()
	defer eb.typesMutex.RUnlock()

	types := make([]EventType, 0, len(eb.registeredTypes))
	for eventType := range eb.registeredTypes {
		types = append(types, eventType)
	}
	return types
}

// GetSubscriptionCount 获取指定事件类型的订阅数量
func (eb *DefaultEventBus) GetSubscriptionCount(eventType EventType) int {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()
	return len(eb.subscriptions[eventType])
}

// GetTotalSubscriptions 获取总订阅数量
func (eb *DefaultEventBus) GetTotalSubscriptions() int {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()
	return len(eb.subscriberMap)
}

// GetEventStats 获取事件统计信息
func (eb *DefaultEventBus) GetEventStats() *EventStats {
	eb.statsMutex.RLock()
	defer eb.statsMutex.RUnlock()

	// 复制统计信息以避免并发修改
	stats := &EventStats{
		TotalEvents:      eb.stats.TotalEvents,
		TotalSubscribers: eb.GetTotalSubscriptions(),
		EventCounts:      make(map[EventType]int64),
		ErrorCounts:      make(map[EventType]int64),
		LastEventTime:    eb.stats.LastEventTime,
	}

	for k, v := range eb.stats.EventCounts {
		stats.EventCounts[k] = v
	}
	for k, v := range eb.stats.ErrorCounts {
		stats.ErrorCounts[k] = v
	}

	return stats
}

// incrementEventCount 增加事件计数
func (eb *DefaultEventBus) incrementEventCount(eventType EventType) {
	eb.statsMutex.Lock()
	defer eb.statsMutex.Unlock()

	eb.stats.TotalEvents++
	eb.stats.EventCounts[eventType]++
	eb.stats.LastEventTime = time.Now()
}

// incrementErrorCount 增加错误计数
func (eb *DefaultEventBus) incrementErrorCount(eventType EventType) {
	eb.statsMutex.Lock()
	defer eb.statsMutex.Unlock()

	eb.stats.ErrorCounts[eventType]++
}

// getSubscriptions 获取指定事件类型的订阅列表
func (eb *DefaultEventBus) getSubscriptions(eventType EventType) []*Subscription {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()
	return eb.subscriptions[eventType]
}

// handleAsyncSubscription 处理异步订阅
func (eb *DefaultEventBus) handleAsyncSubscription(ctx context.Context, event Event, subscription *Subscription, priority EventPriority) {
	task := &eventTask{
		ctx:          ctx,
		event:        event,
		subscription: subscription,
		priority:     priority,
		startTime:    time.Now(),
	}

	// 检查事件总线是否已关闭
	select {
	case <-eb.ctx.Done():
		return
	default:
	}

	// 尝试批量处理
	if eb.tryBatchProcess(task) {
		return
	}

	// 回退到单个任务处理
	select {
	case eb.workerPool <- struct{}{}: // 获取工作协程槽位
		eb.enqueueTask(ctx, task, event.GetType())
	default:
		eb.logger.Warn("All workers are busy, dropping event",
			"event_type", event.GetType())
	}
}

// tryBatchProcess 尝试批量处理
func (eb *DefaultEventBus) tryBatchProcess(task *eventTask) bool {
	eb.batchProcessor.mutex.Lock()
	defer eb.batchProcessor.mutex.Unlock()

	// 添加到待处理任务
	eb.batchProcessor.pendingTasks = append(eb.batchProcessor.pendingTasks, task)

	// 检查是否需要立即刷新
	if len(eb.batchProcessor.pendingTasks) >= eb.batchSize {
		// 复制待处理任务
		batch := make([]*eventTask, len(eb.batchProcessor.pendingTasks))
		copy(batch, eb.batchProcessor.pendingTasks)

		// 清空待处理任务
		eb.batchProcessor.pendingTasks = eb.batchProcessor.pendingTasks[:0]
		eb.batchProcessor.lastFlush = time.Now()

		// 异步发送到批量队列
		go func() {
			select {
			case eb.batchQueue <- batch:
				// 成功发送
			default:
				// 队列满，直接处理
				eb.processBatch(batch)
			}
		}()
	}

	return true
}

// enqueueTask 将任务加入队列
func (eb *DefaultEventBus) enqueueTask(ctx context.Context, task *eventTask, eventType EventType) {
	select {
	case eb.eventQueue <- task:
		// 任务已加入队列
	case <-ctx.Done():
		<-eb.workerPool // 释放槽位
	case <-eb.ctx.Done():
		<-eb.workerPool // 释放槽位
	default:
		<-eb.workerPool // 释放槽位
		eb.logger.Warn("Event queue is full, dropping event",
			"event_type", eventType)
	}
}

// handleSyncSubscription 处理同步订阅
func (eb *DefaultEventBus) handleSyncSubscription(ctx context.Context, event Event, subscription *Subscription) error {
	if subscription.Filter != nil && !subscription.Filter(event) {
		return nil
	}

	// 处理panic和错误
	var handlerErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				eb.logger.Error("Sync event handler panicked",
					"event_type", event.GetType(),
					"subscription_id", subscription.ID,
					"panic", r)
				eb.incrementErrorCount(event.GetType())
				handlerErr = fmt.Errorf("handler panic: %v", r)
			}
		}()

		handlerErr = subscription.Handler(ctx, event)
	}()

	if handlerErr != nil {
		eb.logger.Error("Sync event handler failed",
			"event_type", event.GetType(),
			"subscription_id", subscription.ID,
			"error", handlerErr)
		eb.incrementErrorCount(event.GetType())
		return handlerErr // 同步处理时返回错误
	}

	return nil
}

// PublishSync 同步发布事件（别名方法）
func (eb *DefaultEventBus) PublishSync(ctx context.Context, event Event) error {
	return eb.publishEvent(ctx, event, PriorityNormal, false)
}

// GetStats 获取统计信息（别名方法）
func (eb *DefaultEventBus) GetStats() *EventStats {
	return eb.GetEventStats()
}