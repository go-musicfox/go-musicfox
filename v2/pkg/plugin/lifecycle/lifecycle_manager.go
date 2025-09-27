// Package plugin 实现插件生命周期管理
package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/plugin/loader"
)

// LifecycleManager 插件生命周期管理器接口
type LifecycleManager interface {
	// InitializePlugin 初始化插件
	InitializePlugin(ctx context.Context, pluginID string, config map[string]interface{}) error
	
	// StartPlugin 启动插件
	StartPlugin(ctx context.Context, pluginID string) error
	
	// StopPlugin 停止插件
	StopPlugin(ctx context.Context, pluginID string) error
	
	// RestartPlugin 重启插件
	RestartPlugin(ctx context.Context, pluginID string) error
	
	// CleanupPlugin 清理插件
	CleanupPlugin(ctx context.Context, pluginID string) error
	
	// GetPluginState 获取插件状态
	GetPluginState(pluginID string) (loader.PluginState, error)
	
	// SetPluginState 设置插件状态
	SetPluginState(pluginID string, state loader.PluginState) error
	
	// WaitForState 等待插件达到指定状态
	WaitForState(ctx context.Context, pluginID string, targetState loader.PluginState, timeout time.Duration) error
	
	// GetStateHistory 获取状态历史
	GetStateHistory(pluginID string) ([]StateTransition, error)
	
	// RegisterStateListener 注册状态监听器
	RegisterStateListener(pluginID string, listener StateListener) error
	
	// UnregisterStateListener 注销状态监听器
	UnregisterStateListener(pluginID string, listener StateListener) error
}

// DynamicLifecycleManager 动态插件生命周期管理器实现
type DynamicLifecycleManager struct {
	// loader 关联的动态库加载器
	loader *loader.DynamicLibraryLoader
	
	// stateHistory 状态历史记录
	stateHistory map[string][]StateTransition
	
	// stateListeners 状态监听器
	stateListeners map[string][]StateListener
	
	// mutex 读写锁
	mutex sync.RWMutex
	
	// config 生命周期管理配置
	config *LifecycleConfig
}

// StateTransition 状态转换记录
type StateTransition struct {
	// FromState 原状态
	FromState loader.PluginState
	
	// ToState 目标状态
	ToState loader.PluginState
	
	// Timestamp 转换时间
	Timestamp time.Time
	
	// Reason 转换原因
	Reason string
	
	// Error 错误信息（如果有）
	Error error
}

// StateListener 状态监听器接口
type StateListener interface {
	// OnStateChanged 状态变化回调
	OnStateChanged(pluginID string, transition StateTransition)
}

// LifecycleConfig 生命周期管理配置
type LifecycleConfig struct {
	// InitTimeout 初始化超时时间
	InitTimeout time.Duration
	
	// StartTimeout 启动超时时间
	StartTimeout time.Duration
	
	// StopTimeout 停止超时时间
	StopTimeout time.Duration
	
	// CleanupTimeout 清理超时时间
	CleanupTimeout time.Duration
	
	// MaxRetries 最大重试次数
	MaxRetries int
	
	// RetryDelay 重试延迟
	RetryDelay time.Duration
	
	// EnableStateHistory 是否启用状态历史
	EnableStateHistory bool
	
	// MaxHistorySize 最大历史记录数
	MaxHistorySize int
}

// NewDynamicLifecycleManager 创建新的动态生命周期管理器
func NewDynamicLifecycleManager(loader *loader.DynamicLibraryLoader, config *LifecycleConfig) *DynamicLifecycleManager {
	if config == nil {
		config = DefaultLifecycleConfig()
	}
	
	return &DynamicLifecycleManager{
		loader:          loader,
		stateHistory:    make(map[string][]StateTransition),
		stateListeners:  make(map[string][]StateListener),
		config:          config,
	}
}

// DefaultLifecycleConfig 返回默认的生命周期配置
func DefaultLifecycleConfig() *LifecycleConfig {
	return &LifecycleConfig{
		InitTimeout:        30 * time.Second,
		StartTimeout:       30 * time.Second,
		StopTimeout:        15 * time.Second,
		CleanupTimeout:     10 * time.Second,
		MaxRetries:         3,
		RetryDelay:         1 * time.Second,
		EnableStateHistory: true,
		MaxHistorySize:     100,
	}
}

// InitializePlugin 初始化插件
func (dlm *DynamicLifecycleManager) InitializePlugin(ctx context.Context, pluginID string, config map[string]interface{}) error {
	// 创建带超时的上下文
	initCtx, cancel := context.WithTimeout(ctx, dlm.config.InitTimeout)
	defer cancel()
	
	// 获取插件实例
	plugin, err := dlm.getPluginInstance(pluginID)
	if err != nil {
		return err
	}
	
	// 检查当前状态
	currentState := plugin.GetState()
	if currentState != loader.PluginStateLoaded {
		return fmt.Errorf("plugin must be in loaded state to initialize, current state: %s", currentState)
	}
	
	// 记录状态转换
	dlm.recordStateTransition(pluginID, currentState, loader.PluginStateLoaded, "Starting initialization")
	
	// 设置配置
	if config != nil {
		if err := plugin.SetConfig(config); err != nil {
			dlm.recordStateTransition(pluginID, loader.PluginStateLoaded, loader.PluginStateError, fmt.Sprintf("Config setting failed: %v", err))
			return fmt.Errorf("failed to set plugin config: %w", err)
		}
	}
	
	// 创建插件上下文
	pluginCtx := dlm.createPluginContext(pluginID)
	
	// 执行初始化
	err = dlm.executeWithRetry(initCtx, func() error {
		return plugin.Initialize(pluginCtx)
	})
	
	if err != nil {
		dlm.recordStateTransition(pluginID, loader.PluginStateLoaded, loader.PluginStateError, fmt.Sprintf("Initialization failed: %v", err))
		return fmt.Errorf("plugin initialization failed: %w", err)
	}
	
	// 更新状态
	plugin.UpdateState(loader.PluginStateLoaded)
	dlm.recordStateTransition(pluginID, loader.PluginStateLoaded, loader.PluginStateLoaded, "Initialization completed")
	
	return nil
}

// StartPlugin 启动插件
func (dlm *DynamicLifecycleManager) StartPlugin(ctx context.Context, pluginID string) error {
	// 创建带超时的上下文
	startCtx, cancel := context.WithTimeout(ctx, dlm.config.StartTimeout)
	defer cancel()
	
	// 获取插件实例
	plugin, err := dlm.getPluginInstance(pluginID)
	if err != nil {
		return err
	}
	
	// 检查当前状态
	currentState := plugin.GetState()
	if currentState != loader.PluginStateLoaded {
		return fmt.Errorf("plugin must be initialized before starting, current state: %s", currentState)
	}
	
	// 记录状态转换
	dlm.recordStateTransition(pluginID, currentState, loader.PluginStateRunning, "Starting plugin")
	
	// 执行启动
	err = dlm.executeWithRetry(startCtx, func() error {
		return plugin.Start()
	})
	
	if err != nil {
		dlm.recordStateTransition(pluginID, loader.PluginStateRunning, loader.PluginStateError, fmt.Sprintf("Start failed: %v", err))
		return fmt.Errorf("plugin start failed: %w", err)
	}
	
	// 更新状态
	plugin.UpdateState(loader.PluginStateRunning)
	dlm.recordStateTransition(pluginID, loader.PluginStateRunning, loader.PluginStateRunning, "Plugin started successfully")
	
	return nil
}

// StopPlugin 停止插件
func (dlm *DynamicLifecycleManager) StopPlugin(ctx context.Context, pluginID string) error {
	// 创建带超时的上下文
	stopCtx, cancel := context.WithTimeout(ctx, dlm.config.StopTimeout)
	defer cancel()
	
	// 获取插件实例
	plugin, err := dlm.getPluginInstance(pluginID)
	if err != nil {
		return err
	}
	
	// 检查当前状态
	currentState := plugin.GetState()
	if currentState != loader.PluginStateRunning {
		return nil // 已经停止或未运行
	}
	
	// 记录状态转换
	dlm.recordStateTransition(pluginID, currentState, loader.PluginStateStopped, "Stopping plugin")
	
	// 执行停止
	err = dlm.executeWithRetry(stopCtx, func() error {
		return plugin.Stop()
	})
	
	if err != nil {
		dlm.recordStateTransition(pluginID, loader.PluginStateStopped, loader.PluginStateError, fmt.Sprintf("Stop failed: %v", err))
		return fmt.Errorf("plugin stop failed: %w", err)
	}
	
	// 更新状态
	plugin.UpdateState(loader.PluginStateStopped)
	dlm.recordStateTransition(pluginID, loader.PluginStateStopped, loader.PluginStateStopped, "Plugin stopped successfully")
	
	return nil
}

// RestartPlugin 重启插件
func (dlm *DynamicLifecycleManager) RestartPlugin(ctx context.Context, pluginID string) error {
	// 先停止插件
	if err := dlm.StopPlugin(ctx, pluginID); err != nil {
		return fmt.Errorf("failed to stop plugin for restart: %w", err)
	}
	
	// 等待一小段时间
	time.Sleep(dlm.config.RetryDelay)
	
	// 重新启动插件
	if err := dlm.StartPlugin(ctx, pluginID); err != nil {
		return fmt.Errorf("failed to start plugin after restart: %w", err)
	}
	
	return nil
}

// CleanupPlugin 清理插件
func (dlm *DynamicLifecycleManager) CleanupPlugin(ctx context.Context, pluginID string) error {
	// 创建带超时的上下文
	cleanupCtx, cancel := context.WithTimeout(ctx, dlm.config.CleanupTimeout)
	defer cancel()
	
	// 获取插件实例
	plugin, err := dlm.getPluginInstance(pluginID)
	if err != nil {
		return err
	}
	
	// 获取当前状态
	currentState := plugin.GetState()
	
	// 如果插件还在运行，先停止它
	if currentState == loader.PluginStateRunning {
		if err := dlm.StopPlugin(ctx, pluginID); err != nil {
			// 记录错误但继续清理
			fmt.Printf("Warning: failed to stop plugin during cleanup: %v\n", err)
		}
	}
	
	// 记录状态转换
	dlm.recordStateTransition(pluginID, currentState, loader.PluginStateStopped, "Starting cleanup")
	
	// 执行清理
	err = dlm.executeWithRetry(cleanupCtx, func() error {
		return plugin.Cleanup()
	})
	
	if err != nil {
		dlm.recordStateTransition(pluginID, loader.PluginStateStopped, loader.PluginStateError, fmt.Sprintf("Cleanup failed: %v", err))
		return fmt.Errorf("plugin cleanup failed: %w", err)
	}
	
	// 更新状态
	plugin.UpdateState(loader.PluginStateStopped)
	dlm.recordStateTransition(pluginID, loader.PluginStateStopped, loader.PluginStateStopped, "Cleanup completed")
	
	// 清理状态历史和监听器
	dlm.cleanupPluginData(pluginID)
	
	return nil
}

// GetPluginState 获取插件状态
func (dlm *DynamicLifecycleManager) GetPluginState(pluginID string) (loader.PluginState, error) {
	plugin, err := dlm.getPluginInstance(pluginID)
	if err != nil {
		return loader.PluginStateUnknown, err
	}
	
	return plugin.GetState(), nil
}

// SetPluginState 设置插件状态
func (dlm *DynamicLifecycleManager) SetPluginState(pluginID string, state loader.PluginState) error {
	plugin, err := dlm.getPluginInstance(pluginID)
	if err != nil {
		return err
	}
	
	currentState := plugin.GetState()
	plugin.UpdateState(state)
	dlm.recordStateTransition(pluginID, currentState, state, "State manually set")
	
	return nil
}

// WaitForState 等待插件达到指定状态
func (dlm *DynamicLifecycleManager) WaitForState(ctx context.Context, pluginID string, targetState loader.PluginState, timeout time.Duration) error {
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("timeout waiting for plugin '%s' to reach state '%s'", pluginID, targetState)
		case <-ticker.C:
			currentState, err := dlm.GetPluginState(pluginID)
			if err != nil {
				return err
			}
			
			if currentState == targetState {
				return nil
			}
			
			// 如果状态是错误状态，立即返回
			if currentState == loader.PluginStateError {
				return fmt.Errorf("plugin '%s' is in error state", pluginID)
			}
		}
	}
}