// Package plugin 实现生命周期管理器的辅助方法
package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/plugin/loader"
)

// 类型别名
type PluginState = loader.PluginState

// 常量别名
const (
	PluginStateUnknown = loader.PluginStateUnknown
)

// GetStateHistory 获取状态历史
func (dlm *DynamicLifecycleManager) GetStateHistory(pluginID string) ([]StateTransition, error) {
	dlm.mutex.RLock()
	defer dlm.mutex.RUnlock()
	
	history, exists := dlm.stateHistory[pluginID]
	if !exists {
		return nil, fmt.Errorf("no state history found for plugin: %s", pluginID)
	}
	
	// 返回历史记录的副本
	result := make([]StateTransition, len(history))
	copy(result, history)
	return result, nil
}

// RegisterStateListener 注册状态监听器
func (dlm *DynamicLifecycleManager) RegisterStateListener(pluginID string, listener StateListener) error {
	if listener == nil {
		return fmt.Errorf("listener cannot be nil")
	}
	
	dlm.mutex.Lock()
	defer dlm.mutex.Unlock()
	
	if dlm.stateListeners[pluginID] == nil {
		dlm.stateListeners[pluginID] = make([]StateListener, 0)
	}
	
	dlm.stateListeners[pluginID] = append(dlm.stateListeners[pluginID], listener)
	return nil
}

// UnregisterStateListener 注销状态监听器
func (dlm *DynamicLifecycleManager) UnregisterStateListener(pluginID string, listener StateListener) error {
	if listener == nil {
		return fmt.Errorf("listener cannot be nil")
	}
	
	dlm.mutex.Lock()
	defer dlm.mutex.Unlock()
	
	listeners, exists := dlm.stateListeners[pluginID]
	if !exists {
		return fmt.Errorf("no listeners found for plugin: %s", pluginID)
	}
	
	// 查找并移除监听器
	for i, l := range listeners {
		if l == listener {
			dlm.stateListeners[pluginID] = append(listeners[:i], listeners[i+1:]...)
			return nil
		}
	}
	
	return fmt.Errorf("listener not found for plugin: %s", pluginID)
}

// getPluginInstance 获取插件实例
func (dlm *DynamicLifecycleManager) getPluginInstance(pluginID string) (*loader.PluginWrapper, error) {
	if pluginID == "" {
		return nil, fmt.Errorf("plugin ID cannot be empty")
	}
	
	if dlm.loader == nil {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}
	
	// 通过加载器的公共方法获取插件
	plugins := dlm.loader.GetLoadedPlugins()
	plugin, exists := plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	if plugin == nil {
		return nil, fmt.Errorf("plugin instance is nil for: %s", pluginID)
	}

	// 创建PluginWrapper包装器
	wrapper := &loader.PluginWrapper{}
	return wrapper, nil
}

// recordStateTransition 记录状态转换
func (dlm *DynamicLifecycleManager) recordStateTransition(pluginID string, fromState, toState loader.PluginState, reason string) {
	if !dlm.config.EnableStateHistory {
		return
	}
	
	transition := StateTransition{
		FromState: fromState,
		ToState:   toState,
		Timestamp: time.Now(),
		Reason:    reason,
	}
	
	dlm.mutex.Lock()
	defer dlm.mutex.Unlock()
	
	// 初始化历史记录
	if dlm.stateHistory[pluginID] == nil {
		dlm.stateHistory[pluginID] = make([]StateTransition, 0)
	}
	
	// 添加新的转换记录
	dlm.stateHistory[pluginID] = append(dlm.stateHistory[pluginID], transition)
	
	// 限制历史记录大小
	if len(dlm.stateHistory[pluginID]) > dlm.config.MaxHistorySize {
		dlm.stateHistory[pluginID] = dlm.stateHistory[pluginID][1:]
	}
	
	// 通知监听器
	dlm.notifyStateListeners(pluginID, transition)
}

// notifyStateListeners 通知状态监听器
func (dlm *DynamicLifecycleManager) notifyStateListeners(pluginID string, transition StateTransition) {
	listeners, exists := dlm.stateListeners[pluginID]
	if !exists || len(listeners) == 0 {
		return
	}
	
	// 异步通知监听器，避免阻塞
	go func() {
		for _, listener := range listeners {
			func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("State listener panic for plugin %s: %v\n", pluginID, r)
					}
				}()
				listener.OnStateChanged(pluginID, transition)
			}()
		}
	}()
}

// executeWithRetry 带重试的执行函数
func (dlm *DynamicLifecycleManager) executeWithRetry(ctx context.Context, fn func() error) error {
	var lastErr error
	
	for i := 0; i <= dlm.config.MaxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		err := fn()
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		// 如果不是最后一次重试，等待一段时间
		if i < dlm.config.MaxRetries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(dlm.config.RetryDelay):
				// 继续重试
			}
		}
	}
	
	return fmt.Errorf("operation failed after %d retries: %w", dlm.config.MaxRetries, lastErr)
}

// createPluginContext 创建插件上下文
func (dlm *DynamicLifecycleManager) createPluginContext(pluginID string) loader.PluginContext {
	// 返回nil，实际实现需要根据具体需求
	return nil
}

// cleanupPluginData 清理插件数据
func (dlm *DynamicLifecycleManager) cleanupPluginData(pluginID string) {
	dlm.mutex.Lock()
	defer dlm.mutex.Unlock()
	
	// 清理状态历史
	delete(dlm.stateHistory, pluginID)
	
	// 清理状态监听器
	delete(dlm.stateListeners, pluginID)
}

// GetAllPluginStates 获取所有插件的状态
func (dlm *DynamicLifecycleManager) GetAllPluginStates() map[string]PluginState {
	if dlm.loader == nil {
		return make(map[string]PluginState)
	}
	
	// 通过公共方法获取已加载的插件
	plugins := dlm.loader.GetLoadedPlugins()
	states := make(map[string]PluginState)
	
	for pluginID := range plugins {
		// 默认状态为未知，实际实现需要根据具体情况获取状态
		states[pluginID] = PluginStateUnknown
	}
	
	return states
}

// GetPluginsByState 根据状态获取插件列表
func (dlm *DynamicLifecycleManager) GetPluginsByState(state PluginState) []string {
	var plugins []string
	
	allStates := dlm.GetAllPluginStates()
	for pluginID, pluginState := range allStates {
		if pluginState == state {
			plugins = append(plugins, pluginID)
		}
	}
	
	return plugins
}

// GetLifecycleStats 获取生命周期统计信息
func (dlm *DynamicLifecycleManager) GetLifecycleStats() *LifecycleStats {
	allStates := dlm.GetAllPluginStates()
	
	stats := &LifecycleStats{
		TotalPlugins: len(allStates),
		StateCounts:  make(map[PluginState]int),
	}
	
	for _, state := range allStates {
		stats.StateCounts[state]++
	}
	
	return stats
}

// LifecycleStats 生命周期统计信息
type LifecycleStats struct {
	// TotalPlugins 总插件数
	TotalPlugins int
	
	// StateCounts 各状态插件数量
	StateCounts map[PluginState]int
}

// BatchOperation 批量操作接口
type BatchOperation interface {
	// Execute 执行批量操作
	Execute(ctx context.Context, pluginIDs []string) map[string]error
}

// BatchInitializer 批量初始化器
type BatchInitializer struct {
	manager *DynamicLifecycleManager
	config  map[string]interface{}
}

// NewBatchInitializer 创建批量初始化器
func NewBatchInitializer(manager *DynamicLifecycleManager, config map[string]interface{}) *BatchInitializer {
	return &BatchInitializer{
		manager: manager,
		config:  config,
	}
}

// Execute 执行批量初始化
func (bi *BatchInitializer) Execute(ctx context.Context, pluginIDs []string) map[string]error {
	results := make(map[string]error)
	var wg sync.WaitGroup
	var mutex sync.Mutex
	
	for _, pluginID := range pluginIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			
			err := bi.manager.InitializePlugin(ctx, id, bi.config)
			
			mutex.Lock()
			results[id] = err
			mutex.Unlock()
		}(pluginID)
	}
	
	wg.Wait()
	return results
}

// BatchStarter 批量启动器
type BatchStarter struct {
	manager *DynamicLifecycleManager
}

// NewBatchStarter 创建批量启动器
func NewBatchStarter(manager *DynamicLifecycleManager) *BatchStarter {
	return &BatchStarter{
		manager: manager,
	}
}

// Execute 执行批量启动
func (bs *BatchStarter) Execute(ctx context.Context, pluginIDs []string) map[string]error {
	results := make(map[string]error)
	var wg sync.WaitGroup
	var mutex sync.Mutex
	
	for _, pluginID := range pluginIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			
			err := bs.manager.StartPlugin(ctx, id)
			
			mutex.Lock()
			results[id] = err
			mutex.Unlock()
		}(pluginID)
	}
	
	wg.Wait()
	return results
}

// BatchStopper 批量停止器
type BatchStopper struct {
	manager *DynamicLifecycleManager
}

// NewBatchStopper 创建批量停止器
func NewBatchStopper(manager *DynamicLifecycleManager) *BatchStopper {
	return &BatchStopper{
		manager: manager,
	}
}

// Execute 执行批量停止
func (bs *BatchStopper) Execute(ctx context.Context, pluginIDs []string) map[string]error {
	results := make(map[string]error)
	var wg sync.WaitGroup
	var mutex sync.Mutex
	
	for _, pluginID := range pluginIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			
			err := bs.manager.StopPlugin(ctx, id)
			
			mutex.Lock()
			results[id] = err
			mutex.Unlock()
		}(pluginID)
	}
	
	wg.Wait()
	return results
}