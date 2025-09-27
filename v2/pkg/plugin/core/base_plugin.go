package plugin

import (
	"fmt"
	"sync"
	"time"
)

// BasePlugin 基础插件实现
// 提供插件接口的默认实现
type BasePlugin struct {
	info         *PluginInfo
	state        PluginState
	context      PluginContext
	capabilities []string
	dependencies []string
	config       map[string]interface{}
	metrics      *PluginMetrics
	mu           sync.RWMutex
	started      bool
	initialized  bool
}

// PluginMetrics 插件指标
type PluginMetrics struct {
	StartTime       time.Time              `json:"start_time"`       // 启动时间
	Uptime          time.Duration          `json:"uptime"`           // 运行时间
	RequestCount    int64                  `json:"request_count"`    // 请求数量
	ErrorCount      int64                  `json:"error_count"`      // 错误数量
	LastError       string                 `json:"last_error"`       // 最后错误
	LastErrorTime   time.Time              `json:"last_error_time"`  // 最后错误时间
	MemoryUsage     int64                  `json:"memory_usage"`     // 内存使用量
	CPUUsage        float64                `json:"cpu_usage"`        // CPU使用率
	CustomMetrics   map[string]interface{} `json:"custom_metrics"`   // 自定义指标
	mu              sync.RWMutex
}

// NewBasePlugin 创建基础插件
func NewBasePlugin(info *PluginInfo) *BasePlugin {
	if info == nil {
		info = &PluginInfo{
			Name:      "Unknown Plugin",
			Version:   "1.0.0",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	return &BasePlugin{
		info:         info,
		state:        PluginStateUnknown,
		capabilities: make([]string, 0),
		dependencies: make([]string, 0),
		config:       make(map[string]interface{}),
		metrics: &PluginMetrics{
			CustomMetrics: make(map[string]interface{}),
		},
	}
}

// GetInfo 获取插件信息
func (p *BasePlugin) GetInfo() *PluginInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// 返回信息副本
	info := *p.info
	return &info
}

// GetCapabilities 获取插件能力
func (p *BasePlugin) GetCapabilities() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	capabilities := make([]string, len(p.capabilities))
	copy(capabilities, p.capabilities)
	return capabilities
}

// GetDependencies 获取插件依赖
func (p *BasePlugin) GetDependencies() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	dependencies := make([]string, len(p.dependencies))
	copy(dependencies, p.dependencies)
	return dependencies
}

// Initialize 初始化插件
func (p *BasePlugin) Initialize(ctx PluginContext) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.initialized {
		return fmt.Errorf("plugin already initialized")
	}

	if ctx == nil {
		return fmt.Errorf("plugin context cannot be nil")
	}

	p.context = ctx
	p.state = PluginStateLoaded
	p.initialized = true

	// 初始化指标
	p.metrics.StartTime = time.Now()

	return nil
}

// Start 启动插件
func (p *BasePlugin) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.initialized {
		return fmt.Errorf("plugin not initialized")
	}

	if p.started {
		return fmt.Errorf("plugin already started")
	}

	if p.state != PluginStateLoaded && p.state != PluginStateStopped {
		return fmt.Errorf("invalid state for starting: %s", p.state.String())
	}

	p.state = PluginStateRunning
	p.started = true
	p.metrics.StartTime = time.Now()

	return nil
}

// Stop 停止插件
func (p *BasePlugin) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.started {
		return fmt.Errorf("plugin not started")
	}

	if p.state != PluginStateRunning && p.state != PluginStatePaused {
		return fmt.Errorf("invalid state for stopping: %s", p.state.String())
	}

	p.state = PluginStateStopped
	p.started = false

	return nil
}

// Cleanup 清理插件资源
func (p *BasePlugin) Cleanup() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 如果插件正在运行，先停止
	if p.started {
		p.state = PluginStateStopped
		p.started = false
	}

	// 清理资源
	p.context = nil
	p.config = make(map[string]interface{})
	p.state = PluginStateUnloaded
	p.initialized = false

	return nil
}

// HealthCheck 健康检查
func (p *BasePlugin) HealthCheck() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.initialized {
		return fmt.Errorf("plugin not initialized")
	}

	if p.state == PluginStateError || p.state == PluginStateCorrupted {
		return fmt.Errorf("plugin in error state: %s", p.state.String())
	}

	return nil
}

// ValidateConfig 验证配置
func (p *BasePlugin) ValidateConfig(config map[string]interface{}) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// 基础实现：接受所有配置
	return nil
}

// UpdateConfig 更新配置
func (p *BasePlugin) UpdateConfig(config map[string]interface{}) error {
	if err := p.ValidateConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// 更新配置
	for k, v := range config {
		p.config[k] = v
	}

	p.info.UpdatedAt = time.Now()

	return nil
}

// GetMetrics 获取插件指标
func (p *BasePlugin) GetMetrics() (*PluginMetrics, error) {
	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()

	// 更新运行时间
	if p.started {
		p.metrics.Uptime = time.Since(p.metrics.StartTime)
	}

	// 返回指标副本
	metrics := &PluginMetrics{
		StartTime:     p.metrics.StartTime,
		Uptime:        p.metrics.Uptime,
		RequestCount:  p.metrics.RequestCount,
		ErrorCount:    p.metrics.ErrorCount,
		LastError:     p.metrics.LastError,
		LastErrorTime: p.metrics.LastErrorTime,
		MemoryUsage:   p.metrics.MemoryUsage,
		CPUUsage:      p.metrics.CPUUsage,
		CustomMetrics: make(map[string]interface{}),
	}

	// 复制自定义指标
	for k, v := range p.metrics.CustomMetrics {
		metrics.CustomMetrics[k] = v
	}

	return metrics, nil
}

// HandleEvent 处理事件
func (p *BasePlugin) HandleEvent(event interface{}) error {
	// 基础实现：忽略所有事件
	return nil
}

// GetState 获取插件状态
func (p *BasePlugin) GetState() PluginState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}

// SetState 设置插件状态
func (p *BasePlugin) SetState(state PluginState) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 检查状态转换是否有效
	if !p.state.IsValidStateTransition(state) {
		return fmt.Errorf("invalid state transition from %s to %s", p.state.String(), state.String())
	}

	p.state = state
	return nil
}

// IsInitialized 检查是否已初始化
func (p *BasePlugin) IsInitialized() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.initialized
}

// IsStarted 检查是否已启动
func (p *BasePlugin) IsStarted() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.started
}

// GetContext 获取插件上下文
func (p *BasePlugin) GetContext() PluginContext {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.context
}

// GetConfig 获取插件配置
func (p *BasePlugin) GetConfig() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// 返回配置副本
	config := make(map[string]interface{})
	for k, v := range p.config {
		config[k] = v
	}

	return config
}

// AddCapability 添加能力
func (p *BasePlugin) AddCapability(capability string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 检查是否已存在
	for _, cap := range p.capabilities {
		if cap == capability {
			return
		}
	}

	p.capabilities = append(p.capabilities, capability)
}

// RemoveCapability 移除能力
func (p *BasePlugin) RemoveCapability(capability string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, cap := range p.capabilities {
		if cap == capability {
			p.capabilities = append(p.capabilities[:i], p.capabilities[i+1:]...)
			return
		}
	}
}

// AddDependency 添加依赖
func (p *BasePlugin) AddDependency(dependency string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 检查是否已存在
	for _, dep := range p.dependencies {
		if dep == dependency {
			return
		}
	}

	p.dependencies = append(p.dependencies, dependency)
}

// RemoveDependency 移除依赖
func (p *BasePlugin) RemoveDependency(dependency string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, dep := range p.dependencies {
		if dep == dependency {
			p.dependencies = append(p.dependencies[:i], p.dependencies[i+1:]...)
			return
		}
	}
}

// IncrementRequestCount 增加请求计数
func (p *BasePlugin) IncrementRequestCount() {
	p.metrics.mu.Lock()
	defer p.metrics.mu.Unlock()
	p.metrics.RequestCount++
}

// IncrementErrorCount 增加错误计数
func (p *BasePlugin) IncrementErrorCount() {
	p.metrics.mu.Lock()
	defer p.metrics.mu.Unlock()
	p.metrics.ErrorCount++
}

// SetLastError 设置最后错误
func (p *BasePlugin) SetLastError(err string) {
	p.metrics.mu.Lock()
	defer p.metrics.mu.Unlock()
	p.metrics.LastError = err
	p.metrics.LastErrorTime = time.Now()
	p.metrics.ErrorCount++
}

// SetMemoryUsage 设置内存使用量
func (p *BasePlugin) SetMemoryUsage(usage int64) {
	p.metrics.mu.Lock()
	defer p.metrics.mu.Unlock()
	p.metrics.MemoryUsage = usage
}

// SetCPUUsage 设置CPU使用率
func (p *BasePlugin) SetCPUUsage(usage float64) {
	p.metrics.mu.Lock()
	defer p.metrics.mu.Unlock()
	p.metrics.CPUUsage = usage
}

// SetCustomMetric 设置自定义指标
func (p *BasePlugin) SetCustomMetric(key string, value interface{}) {
	p.metrics.mu.Lock()
	defer p.metrics.mu.Unlock()
	p.metrics.CustomMetrics[key] = value
}

// GetCustomMetric 获取自定义指标
func (p *BasePlugin) GetCustomMetric(key string) (interface{}, bool) {
	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()
	value, exists := p.metrics.CustomMetrics[key]
	return value, exists
}

// UpdateInfo 更新插件信息
func (p *BasePlugin) UpdateInfo(info *PluginInfo) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if info != nil {
		info.UpdatedAt = time.Now()
		p.info = info
	}
}

// GetVersion 获取插件版本
func (p *BasePlugin) GetVersion() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.info.Version
}

// String 返回插件的字符串表示
func (p *BasePlugin) String() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return fmt.Sprintf("Plugin{Name: %s, Version: %s, State: %s}",
		p.info.Name, p.info.Version, p.state.String())
}