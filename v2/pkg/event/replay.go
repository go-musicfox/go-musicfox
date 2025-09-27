package event

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// EventReplayer 事件重放器接口
type EventReplayer interface {
	// 重放控制
	Replay(ctx context.Context, options *ReplayOptions) error
	ReplayRange(ctx context.Context, start, end time.Time, options *ReplayOptions) error
	ReplayByType(ctx context.Context, eventType EventType, options *ReplayOptions) error
	ReplayBySource(ctx context.Context, source string, options *ReplayOptions) error

	// 重放状态
	GetStatus() *ReplayStatus
	Pause() error
	Resume() error
	Stop() error

	// 事件处理
	SetEventHandler(handler ReplayEventHandler)
	SetProgressHandler(handler ReplayProgressHandler)
	SetErrorHandler(handler ReplayErrorHandler)
}

// ReplayOptions 重放选项
type ReplayOptions struct {
	Speed         float64       `json:"speed"`          // 重放速度倍数，1.0为正常速度
	BatchSize     int           `json:"batch_size"`     // 批处理大小
	MaxEvents     int64         `json:"max_events"`     // 最大重放事件数量
	StartFrom     time.Time     `json:"start_from"`     // 开始时间
	StopAt        time.Time     `json:"stop_at"`        // 结束时间
	EventTypes    []EventType   `json:"event_types"`    // 过滤事件类型
	Sources       []string      `json:"sources"`        // 过滤事件源
	SkipErrors    bool          `json:"skip_errors"`    // 是否跳过错误
	RealTime      bool          `json:"real_time"`      // 是否按真实时间间隔重放
	Reverse       bool          `json:"reverse"`        // 是否反向重放
	LoopCount     int           `json:"loop_count"`     // 循环次数，0为无限循环，-1为不循环
}

// ReplayStatus 重放状态
type ReplayStatus struct {
	State         ReplayState   `json:"state"`
	TotalEvents   int64         `json:"total_events"`
	ProcessedEvents int64       `json:"processed_events"`
	CurrentEvent  *StoredEvent  `json:"current_event,omitempty"`
	StartTime     time.Time     `json:"start_time"`
	ElapsedTime   time.Duration `json:"elapsed_time"`
	EstimatedTime time.Duration `json:"estimated_time"`
	ErrorCount    int64         `json:"error_count"`
	LastError     string        `json:"last_error,omitempty"`
}

// ReplayState 重放状态枚举
type ReplayState int

const (
	ReplayStateIdle ReplayState = iota
	ReplayStateRunning
	ReplayStatePaused
	ReplayStateStopped
	ReplayStateCompleted
	ReplayStateError
)

// String 返回重放状态的字符串表示
func (s ReplayState) String() string {
	switch s {
	case ReplayStateIdle:
		return "idle"
	case ReplayStateRunning:
		return "running"
	case ReplayStatePaused:
		return "paused"
	case ReplayStateStopped:
		return "stopped"
	case ReplayStateCompleted:
		return "completed"
	case ReplayStateError:
		return "error"
	default:
		return "unknown"
	}
}

// 事件处理器类型
type ReplayEventHandler func(ctx context.Context, event Event) error
type ReplayProgressHandler func(status *ReplayStatus)
type ReplayErrorHandler func(err error, event Event)

// DefaultEventReplayer 默认事件重放器实现
type DefaultEventReplayer struct {
	logger *slog.Logger
	store  EventStore

	// 状态管理
	status *ReplayStatus
	mutex  sync.RWMutex

	// 控制信号
	ctx    context.Context
	cancel context.CancelFunc
	pauseCh chan struct{}
	resumeCh chan struct{}
	stopCh chan struct{}

	// 事件处理器
	eventHandler    ReplayEventHandler
	progressHandler ReplayProgressHandler
	errorHandler    ReplayErrorHandler
}

// NewDefaultEventReplayer 创建默认事件重放器
func NewDefaultEventReplayer(logger *slog.Logger, store EventStore) *DefaultEventReplayer {
	return &DefaultEventReplayer{
		logger: logger,
		store:  store,
		status: &ReplayStatus{
			State: ReplayStateIdle,
		},
		pauseCh:  make(chan struct{}, 1),
		resumeCh: make(chan struct{}, 1),
		stopCh:   make(chan struct{}, 1),
	}
}

// SetEventHandler 设置事件处理器
func (r *DefaultEventReplayer) SetEventHandler(handler ReplayEventHandler) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.eventHandler = handler
}

// SetProgressHandler 设置进度处理器
func (r *DefaultEventReplayer) SetProgressHandler(handler ReplayProgressHandler) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.progressHandler = handler
}

// SetErrorHandler 设置错误处理器
func (r *DefaultEventReplayer) SetErrorHandler(handler ReplayErrorHandler) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.errorHandler = handler
}

// GetStatus 获取重放状态
func (r *DefaultEventReplayer) GetStatus() *ReplayStatus {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// 复制状态以避免并发修改
	status := *r.status
	return &status
}

// Replay 重放所有事件
func (r *DefaultEventReplayer) Replay(ctx context.Context, options *ReplayOptions) error {
	if options == nil {
		options = &ReplayOptions{
			Speed:     1.0,
			BatchSize: 100,
			LoopCount: -1,
		}
	}

	// 获取所有事件
	totalCount, err := r.store.Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to get event count: %w", err)
	}

	if totalCount == 0 {
		return fmt.Errorf("no events to replay")
	}

	// 分批获取事件
	var allEvents []Event
	offset := 0
	for {
		events, err := r.store.GetByTimeRange(ctx, time.Time{}, time.Now(), options.BatchSize, offset)
		if err != nil {
			return fmt.Errorf("failed to get events: %w", err)
		}

		if len(events) == 0 {
			break
		}

		allEvents = append(allEvents, events...)
		offset += len(events)

		if options.MaxEvents > 0 && int64(len(allEvents)) >= options.MaxEvents {
			allEvents = allEvents[:options.MaxEvents]
			break
		}
	}

	return r.replayEvents(ctx, allEvents, options)
}

// ReplayRange 重放指定时间范围的事件
func (r *DefaultEventReplayer) ReplayRange(ctx context.Context, start, end time.Time, options *ReplayOptions) error {
	if options == nil {
		options = &ReplayOptions{
			Speed:     1.0,
			BatchSize: 100,
			LoopCount: -1,
		}
	}

	// 获取时间范围内的事件
	var allEvents []Event
	offset := 0
	for {
		events, err := r.store.GetByTimeRange(ctx, start, end, options.BatchSize, offset)
		if err != nil {
			return fmt.Errorf("failed to get events by time range: %w", err)
		}

		if len(events) == 0 {
			break
		}

		allEvents = append(allEvents, events...)
		offset += len(events)

		if options.MaxEvents > 0 && int64(len(allEvents)) >= options.MaxEvents {
			allEvents = allEvents[:options.MaxEvents]
			break
		}
	}

	return r.replayEvents(ctx, allEvents, options)
}

// ReplayByType 按事件类型重放
func (r *DefaultEventReplayer) ReplayByType(ctx context.Context, eventType EventType, options *ReplayOptions) error {
	if options == nil {
		options = &ReplayOptions{
			Speed:     1.0,
			BatchSize: 100,
			LoopCount: -1,
		}
	}

	// 获取指定类型的事件
	var allEvents []Event
	offset := 0
	for {
		events, err := r.store.GetByType(ctx, eventType, options.BatchSize, offset)
		if err != nil {
			return fmt.Errorf("failed to get events by type: %w", err)
		}

		if len(events) == 0 {
			break
		}

		allEvents = append(allEvents, events...)
		offset += len(events)

		if options.MaxEvents > 0 && int64(len(allEvents)) >= options.MaxEvents {
			allEvents = allEvents[:options.MaxEvents]
			break
		}
	}

	return r.replayEvents(ctx, allEvents, options)
}

// ReplayBySource 按事件源重放
func (r *DefaultEventReplayer) ReplayBySource(ctx context.Context, source string, options *ReplayOptions) error {
	if options == nil {
		options = &ReplayOptions{
			Speed:     1.0,
			BatchSize: 100,
			LoopCount: -1,
		}
	}

	// 获取指定源的事件
	var allEvents []Event
	offset := 0
	for {
		events, err := r.store.GetBySource(ctx, source, options.BatchSize, offset)
		if err != nil {
			return fmt.Errorf("failed to get events by source: %w", err)
		}

		if len(events) == 0 {
			break
		}

		allEvents = append(allEvents, events...)
		offset += len(events)

		if options.MaxEvents > 0 && int64(len(allEvents)) >= options.MaxEvents {
			allEvents = allEvents[:options.MaxEvents]
			break
		}
	}

	return r.replayEvents(ctx, allEvents, options)
}

// Pause 暂停重放
func (r *DefaultEventReplayer) Pause() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.status.State != ReplayStateRunning {
		return fmt.Errorf("replay is not running")
	}

	select {
	case r.pauseCh <- struct{}{}:
		r.status.State = ReplayStatePaused
		r.logger.Info("Event replay paused")
		return nil
	default:
		return fmt.Errorf("pause signal already sent")
	}
}

// Resume 恢复重放
func (r *DefaultEventReplayer) Resume() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.status.State != ReplayStatePaused {
		return fmt.Errorf("replay is not paused")
	}

	select {
	case r.resumeCh <- struct{}{}:
		r.status.State = ReplayStateRunning
		r.logger.Info("Event replay resumed")
		return nil
	default:
		return fmt.Errorf("resume signal already sent")
	}
}

// Stop 停止重放
func (r *DefaultEventReplayer) Stop() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.status.State == ReplayStateIdle || r.status.State == ReplayStateStopped {
		return fmt.Errorf("replay is not running")
	}

	select {
	case r.stopCh <- struct{}{}:
		r.status.State = ReplayStateStopped
		r.logger.Info("Event replay stopped")
		return nil
	default:
		return fmt.Errorf("stop signal already sent")
	}
}

// replayEvents 重放事件列表
func (r *DefaultEventReplayer) replayEvents(ctx context.Context, events []Event, options *ReplayOptions) error {
	if len(events) == 0 {
		return fmt.Errorf("no events to replay")
	}

	// 初始化重放状态
	r.mutex.Lock()
	if r.status.State == ReplayStateRunning {
		r.mutex.Unlock()
		return fmt.Errorf("replay is already running")
	}

	r.ctx, r.cancel = context.WithCancel(ctx)
	r.status = &ReplayStatus{
		State:         ReplayStateRunning,
		TotalEvents:   int64(len(events)),
		ProcessedEvents: 0,
		StartTime:     time.Now(),
		ErrorCount:    0,
	}
	r.mutex.Unlock()

	defer func() {
		r.mutex.Lock()
		if r.status.State == ReplayStateRunning {
			r.status.State = ReplayStateCompleted
		}
		r.mutex.Unlock()
		r.cancel()
	}()

	// 过滤事件
	filteredEvents := r.filterEvents(events, options)
	if len(filteredEvents) == 0 {
		return fmt.Errorf("no events match the filter criteria")
	}

	// 反向重放
	if options.Reverse {
		for i, j := 0, len(filteredEvents)-1; i < j; i, j = i+1, j-1 {
			filteredEvents[i], filteredEvents[j] = filteredEvents[j], filteredEvents[i]
		}
	}

	// 循环重放
	loopCount := options.LoopCount
	if loopCount == 0 {
		loopCount = -1 // 无限循环
	}

	for loop := 0; loopCount < 0 || loop < loopCount; loop++ {
		if err := r.replayEventLoop(filteredEvents, options); err != nil {
			return err
		}

		// 检查是否被停止
		select {
		case <-r.ctx.Done():
			return r.ctx.Err()
		case <-r.stopCh:
			return nil
		default:
		}
	}

	return nil
}

// replayEventLoop 重放事件循环
func (r *DefaultEventReplayer) replayEventLoop(events []Event, options *ReplayOptions) error {
	var lastEventTime time.Time

	for i, event := range events {
		// 检查控制信号
		select {
		case <-r.ctx.Done():
			return r.ctx.Err()
		case <-r.stopCh:
			return nil
		case <-r.pauseCh:
			// 等待恢复信号
			select {
			case <-r.ctx.Done():
				return r.ctx.Err()
			case <-r.stopCh:
				return nil
			case <-r.resumeCh:
				// 继续执行
			}
		default:
		}

		// 实时重放：按真实时间间隔等待
		if options.RealTime && !lastEventTime.IsZero() {
			currentEventTime := event.GetTimestamp()
			interval := currentEventTime.Sub(lastEventTime)
			if interval > 0 {
				// 根据速度调整间隔
				adjustedInterval := time.Duration(float64(interval) / options.Speed)
				time.Sleep(adjustedInterval)
			}
		}
		lastEventTime = event.GetTimestamp()

		// 更新当前事件
		r.mutex.Lock()
		if storedEvent, ok := event.(*StoredEvent); ok {
			r.status.CurrentEvent = storedEvent
		}
		r.status.ProcessedEvents = int64(i + 1)
		r.status.ElapsedTime = time.Since(r.status.StartTime)
		if r.status.ProcessedEvents > 0 {
			avgTimePerEvent := r.status.ElapsedTime / time.Duration(r.status.ProcessedEvents)
			remainingEvents := r.status.TotalEvents - r.status.ProcessedEvents
			r.status.EstimatedTime = avgTimePerEvent * time.Duration(remainingEvents)
		}
		r.mutex.Unlock()

		// 调用进度处理器
		if r.progressHandler != nil {
			r.progressHandler(r.GetStatus())
		}

		// 处理事件
		if r.eventHandler != nil {
			if err := r.eventHandler(r.ctx, event); err != nil {
				r.mutex.Lock()
				r.status.ErrorCount++
				r.status.LastError = err.Error()
				r.mutex.Unlock()

				// 调用错误处理器
				if r.errorHandler != nil {
					r.errorHandler(err, event)
				}

				// 如果不跳过错误，则停止重放
				if !options.SkipErrors {
					r.mutex.Lock()
					r.status.State = ReplayStateError
					r.mutex.Unlock()
					return fmt.Errorf("event processing failed: %w", err)
				}
			}
		}

		// 根据速度调整处理间隔
		if options.Speed > 0 && options.Speed != 1.0 {
			baseInterval := time.Millisecond * 10 // 基础间隔
			adjustedInterval := time.Duration(float64(baseInterval) / options.Speed)
			time.Sleep(adjustedInterval)
		}
	}

	return nil
}

// filterEvents 过滤事件
func (r *DefaultEventReplayer) filterEvents(events []Event, options *ReplayOptions) []Event {
	var filtered []Event

	for _, event := range events {
		// 时间范围过滤
		if !options.StartFrom.IsZero() && event.GetTimestamp().Before(options.StartFrom) {
			continue
		}
		if !options.StopAt.IsZero() && event.GetTimestamp().After(options.StopAt) {
			continue
		}

		// 事件类型过滤
		if len(options.EventTypes) > 0 {
			matched := false
			for _, eventType := range options.EventTypes {
				if event.GetType() == eventType {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		// 事件源过滤
		if len(options.Sources) > 0 {
			matched := false
			for _, source := range options.Sources {
				if event.GetSource() == source {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		filtered = append(filtered, event)
	}

	return filtered
}