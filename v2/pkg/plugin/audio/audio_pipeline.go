package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AudioPipeline 音频处理管道
type AudioPipeline struct {
	stages    []PipelineStage
	config    *PipelineConfig
	metrics   *PipelineMetrics
	mu        sync.RWMutex
	running   bool
	cancelled bool
}

// PipelineStage 管道阶段接口
type PipelineStage interface {
	GetName() string
	GetDescription() string
	Process(ctx context.Context, buffer *AudioBuffer) (*AudioBuffer, error)
	IsEnabled() bool
	SetEnabled(enabled bool)
	GetMetrics() *StageMetrics
	Cleanup() error
}

// PipelineConfig 管道配置
type PipelineConfig struct {
	MaxConcurrency   int           `json:"max_concurrency"`   // 最大并发数
	Timeout          time.Duration `json:"timeout"`           // 处理超时时间
	BufferSize       int           `json:"buffer_size"`       // 缓冲区大小
	EnableMetrics    bool          `json:"enable_metrics"`    // 启用指标收集
	EnableOptimization bool        `json:"enable_optimization"` // 启用优化
	RetryAttempts    int           `json:"retry_attempts"`    // 重试次数
	RetryDelay       time.Duration `json:"retry_delay"`       // 重试延迟
}

// PipelineMetrics 管道指标
type PipelineMetrics struct {
	TotalProcessed   int64         `json:"total_processed"`   // 总处理数
	TotalErrors      int64         `json:"total_errors"`      // 总错误数
	AverageLatency   time.Duration `json:"average_latency"`   // 平均延迟
	Throughput       float64       `json:"throughput"`        // 吞吐量（samples/sec）
	LastProcessTime  time.Time     `json:"last_process_time"` // 最后处理时间
	StageMetrics     map[string]*StageMetrics `json:"stage_metrics"` // 阶段指标
	mu               sync.RWMutex
}

// StageMetrics 阶段指标
type StageMetrics struct {
	Processed       int64         `json:"processed"`        // 处理数
	Errors          int64         `json:"errors"`           // 错误数
	AverageLatency  time.Duration `json:"average_latency"`  // 平均延迟
	LastProcessTime time.Time     `json:"last_process_time"` // 最后处理时间
	Enabled         bool          `json:"enabled"`          // 是否启用
}

// AudioBuffer 音频缓冲区
type AudioBuffer struct {
	Data       []byte            `json:"data"`        // 音频数据
	Length     int               `json:"length"`      // 数据长度
	SampleRate int               `json:"sample_rate"` // 采样率
	Channels   int               `json:"channels"`    // 声道数
	Format     AudioFormat       `json:"format"`      // 音频格式
	Timestamp  time.Time         `json:"timestamp"`   // 时间戳
	Metadata   map[string]interface{} `json:"metadata"` // 元数据
}

// NewAudioPipeline 创建音频处理管道
func NewAudioPipeline() *AudioPipeline {
	return &AudioPipeline{
		stages: make([]PipelineStage, 0),
		config: &PipelineConfig{
			MaxConcurrency:     4,
			Timeout:            30 * time.Second,
			BufferSize:         4096,
			EnableMetrics:      true,
			EnableOptimization: true,
			RetryAttempts:      3,
			RetryDelay:         100 * time.Millisecond,
		},
		metrics: &PipelineMetrics{
			StageMetrics: make(map[string]*StageMetrics),
		},
	}
}

// AddStage 添加处理阶段
func (p *AudioPipeline) AddStage(stage PipelineStage) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if stage == nil {
		return fmt.Errorf("stage cannot be nil")
	}

	// 检查是否已存在同名阶段
	for _, existingStage := range p.stages {
		if existingStage.GetName() == stage.GetName() {
			return fmt.Errorf("stage with name '%s' already exists", stage.GetName())
		}
	}

	p.stages = append(p.stages, stage)

	// 初始化阶段指标
	if p.config.EnableMetrics {
		p.metrics.StageMetrics[stage.GetName()] = &StageMetrics{
			Enabled: stage.IsEnabled(),
		}
	}

	return nil
}

// RemoveStage 移除处理阶段
func (p *AudioPipeline) RemoveStage(stageName string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, stage := range p.stages {
		if stage.GetName() == stageName {
			// 清理阶段
			if err := stage.Cleanup(); err != nil {
				return fmt.Errorf("failed to cleanup stage '%s': %w", stageName, err)
			}

			// 从切片中移除
			p.stages = append(p.stages[:i], p.stages[i+1:]...)

			// 移除指标
			delete(p.metrics.StageMetrics, stageName)

			return nil
		}
	}

	return fmt.Errorf("stage '%s' not found", stageName)
}

// Process 处理音频数据
func (p *AudioPipeline) Process(ctx context.Context, buffer *AudioBuffer) (*AudioBuffer, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if buffer == nil {
		return nil, fmt.Errorf("buffer cannot be nil")
	}

	if p.cancelled {
		return nil, fmt.Errorf("pipeline is cancelled")
	}

	p.running = true
	defer func() { p.running = false }()

	startTime := time.Now()
	currentBuffer := buffer

	// 创建处理上下文
	processCtx, cancel := context.WithTimeout(ctx, p.config.Timeout)
	defer cancel()

	// 逐个执行管道阶段
	for i, stage := range p.stages {
		if !stage.IsEnabled() {
			continue // 跳过禁用的阶段
		}

		// 检查上下文是否已取消
		select {
		case <-processCtx.Done():
			return nil, fmt.Errorf("pipeline processing timeout or cancelled")
		default:
		}

		// 执行阶段处理
		processedBuffer, err := p.processStageWithRetry(processCtx, stage, currentBuffer)
		if err != nil {
			p.updateErrorMetrics(stage.GetName())
			return nil, fmt.Errorf("stage '%s' (index %d) failed: %w", stage.GetName(), i, err)
		}

		currentBuffer = processedBuffer
		p.updateStageMetrics(stage.GetName(), time.Since(startTime))
	}

	// 更新总体指标
	p.updatePipelineMetrics(time.Since(startTime))

	return currentBuffer, nil
}

// processStageWithRetry 带重试的阶段处理
func (p *AudioPipeline) processStageWithRetry(ctx context.Context, stage PipelineStage, buffer *AudioBuffer) (*AudioBuffer, error) {
	var lastErr error

	for attempt := 0; attempt <= p.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			// 等待重试延迟
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(p.config.RetryDelay):
			}
		}

		result, err := stage.Process(ctx, buffer)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// 检查是否为不可重试的错误
	if !p.isRetryableError(err) {
		break
	}
	}

	return nil, fmt.Errorf("stage failed after %d attempts: %w", p.config.RetryAttempts+1, lastErr)
}

// isRetryableError 检查错误是否可重试
func (p *AudioPipeline) isRetryableError(err error) bool {
	// 简化实现：可以根据具体错误类型判断
	errorStr := err.Error()
	return !(pipelineContains(errorStr, "invalid") || pipelineContains(errorStr, "unsupported"))
}

// pipelineContains 检查字符串是否包含子串
func pipelineContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
			pipelineContainsInner(s[1:len(s)-1], substr))))
}

// pipelineContainsInner 内部包含检查
func pipelineContainsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetStages 获取所有阶段
func (p *AudioPipeline) GetStages() []PipelineStage {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stages := make([]PipelineStage, len(p.stages))
	copy(stages, p.stages)
	return stages
}

// GetStage 获取指定阶段
func (p *AudioPipeline) GetStage(name string) PipelineStage {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, stage := range p.stages {
		if stage.GetName() == name {
			return stage
		}
	}
	return nil
}

// EnableStage 启用阶段
func (p *AudioPipeline) EnableStage(name string) error {
	stage := p.GetStage(name)
	if stage == nil {
		return fmt.Errorf("stage '%s' not found", name)
	}

	stage.SetEnabled(true)
	return nil
}

// DisableStage 禁用阶段
func (p *AudioPipeline) DisableStage(name string) error {
	stage := p.GetStage(name)
	if stage == nil {
		return fmt.Errorf("stage '%s' not found", name)
	}

	stage.SetEnabled(false)
	return nil
}

// StageCount 获取阶段数量
func (p *AudioPipeline) StageCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.stages)
}

// IsRunning 检查管道是否正在运行
func (p *AudioPipeline) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.running
}

// Cancel 取消管道处理
func (p *AudioPipeline) Cancel() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cancelled = true
}

// Reset 重置管道状态
func (p *AudioPipeline) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cancelled = false
	p.running = false
}

// Clear 清理管道
func (p *AudioPipeline) Clear() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 清理所有阶段
	for _, stage := range p.stages {
		if err := stage.Cleanup(); err != nil {
			return fmt.Errorf("failed to cleanup stage '%s': %w", stage.GetName(), err)
		}
	}

	p.stages = p.stages[:0]
	p.metrics.StageMetrics = make(map[string]*StageMetrics)
	p.cancelled = false
	p.running = false

	return nil
}

// GetConfig 获取管道配置
func (p *AudioPipeline) GetConfig() *PipelineConfig {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// 返回配置副本
	config := *p.config
	return &config
}

// UpdateConfig 更新管道配置
func (p *AudioPipeline) UpdateConfig(config *PipelineConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// 验证配置
	if err := p.validateConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	p.config = config
	return nil
}

// validateConfig 验证配置
func (p *AudioPipeline) validateConfig(config *PipelineConfig) error {
	if config.MaxConcurrency <= 0 {
		return fmt.Errorf("max_concurrency must be positive")
	}

	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	if config.BufferSize <= 0 {
		return fmt.Errorf("buffer_size must be positive")
	}

	if config.RetryAttempts < 0 {
		return fmt.Errorf("retry_attempts cannot be negative")
	}

	if config.RetryDelay < 0 {
		return fmt.Errorf("retry_delay cannot be negative")
	}

	return nil
}

// GetMetrics 获取管道指标
func (p *AudioPipeline) GetMetrics() *PipelineMetrics {
	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()

	// 返回指标副本
	metrics := &PipelineMetrics{
		TotalProcessed:  p.metrics.TotalProcessed,
		TotalErrors:     p.metrics.TotalErrors,
		AverageLatency:  p.metrics.AverageLatency,
		Throughput:      p.metrics.Throughput,
		LastProcessTime: p.metrics.LastProcessTime,
		StageMetrics:    make(map[string]*StageMetrics),
	}

	// 复制阶段指标
	for name, stageMetrics := range p.metrics.StageMetrics {
		metrics.StageMetrics[name] = &StageMetrics{
			Processed:       stageMetrics.Processed,
			Errors:          stageMetrics.Errors,
			AverageLatency:  stageMetrics.AverageLatency,
			LastProcessTime: stageMetrics.LastProcessTime,
			Enabled:         stageMetrics.Enabled,
		}
	}

	return metrics
}

// updatePipelineMetrics 更新管道指标
func (p *AudioPipeline) updatePipelineMetrics(latency time.Duration) {
	if !p.config.EnableMetrics {
		return
	}

	p.metrics.mu.Lock()
	defer p.metrics.mu.Unlock()

	p.metrics.TotalProcessed++
	p.metrics.LastProcessTime = time.Now()

	// 更新平均延迟
	if p.metrics.TotalProcessed == 1 {
		p.metrics.AverageLatency = latency
	} else {
		// 使用指数移动平均
		alpha := 0.1
		p.metrics.AverageLatency = time.Duration(float64(p.metrics.AverageLatency)*(1-alpha) + float64(latency)*alpha)
	}

	// 计算吞吐量（简化计算）
	if latency > 0 {
		p.metrics.Throughput = 1.0 / latency.Seconds()
	}
}

// updateStageMetrics 更新阶段指标
func (p *AudioPipeline) updateStageMetrics(stageName string, latency time.Duration) {
	if !p.config.EnableMetrics {
		return
	}

	p.metrics.mu.Lock()
	defer p.metrics.mu.Unlock()

	stageMetrics, exists := p.metrics.StageMetrics[stageName]
	if !exists {
		stageMetrics = &StageMetrics{}
		p.metrics.StageMetrics[stageName] = stageMetrics
	}

	stageMetrics.Processed++
	stageMetrics.LastProcessTime = time.Now()

	// 更新平均延迟
	if stageMetrics.Processed == 1 {
		stageMetrics.AverageLatency = latency
	} else {
		// 使用指数移动平均
		alpha := 0.1
		stageMetrics.AverageLatency = time.Duration(float64(stageMetrics.AverageLatency)*(1-alpha) + float64(latency)*alpha)
	}
}

// updateErrorMetrics 更新错误指标
func (p *AudioPipeline) updateErrorMetrics(stageName string) {
	if !p.config.EnableMetrics {
		return
	}

	p.metrics.mu.Lock()
	defer p.metrics.mu.Unlock()

	p.metrics.TotalErrors++

	stageMetrics, exists := p.metrics.StageMetrics[stageName]
	if !exists {
		stageMetrics = &StageMetrics{}
		p.metrics.StageMetrics[stageName] = stageMetrics
	}

	stageMetrics.Errors++
}

// AudioBufferPool 音频缓冲池
type AudioBufferPool struct {
	pool       sync.Pool
	bufferSize int
	mu         sync.RWMutex
	size       int
}

// NewAudioBufferPool 创建音频缓冲池
func NewAudioBufferPool(bufferSize int) *AudioBufferPool {
	pool := &AudioBufferPool{
		bufferSize: bufferSize,
	}

	pool.pool = sync.Pool{
		New: func() interface{} {
			return &AudioBuffer{
				Data:      make([]byte, bufferSize),
				Metadata:  make(map[string]interface{}),
				Timestamp: time.Now(),
			}
		},
	}

	return pool
}

// Get 获取音频缓冲区
func (p *AudioBufferPool) Get(size int) *AudioBuffer {
	buffer := p.pool.Get().(*AudioBuffer)

	// 如果需要更大的缓冲区，重新分配
	if len(buffer.Data) < size {
		buffer.Data = make([]byte, size)
	}

	// 重置缓冲区状态
	buffer.Length = 0
	buffer.SampleRate = 0
	buffer.Channels = 0
	buffer.Format = AudioFormatUnknown
	buffer.Timestamp = time.Now()

	// 清空元数据
	for k := range buffer.Metadata {
		delete(buffer.Metadata, k)
	}

	p.mu.Lock()
	p.size++
	p.mu.Unlock()

	return buffer
}

// Put 归还音频缓冲区
func (p *AudioBufferPool) Put(buffer *AudioBuffer) {
	if buffer == nil {
		return
	}

	p.pool.Put(buffer)

	p.mu.Lock()
	p.size--
	p.mu.Unlock()
}

// Size 获取缓冲池大小
func (p *AudioBufferPool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.size
}

// Clear 清空缓冲池
func (p *AudioBufferPool) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 创建新的池
	p.pool = sync.Pool{
		New: func() interface{} {
			return &AudioBuffer{
				Data:      make([]byte, p.bufferSize),
				Metadata:  make(map[string]interface{}),
				Timestamp: time.Now(),
			}
		},
	}

	p.size = 0
}