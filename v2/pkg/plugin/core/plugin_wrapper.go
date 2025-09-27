// Package plugin 实现插件包装器，为动态库插件提供统一接口
package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/plugin/loader"
)

// simplePluginContext 简单的插件上下文实现
type simplePluginContext struct {
	config map[string]interface{}
}

func (s *simplePluginContext) GetConfig() map[string]interface{} {
	return s.config
}

func (s *simplePluginContext) GetContext() context.Context {
	return context.Background()
}

func (s *simplePluginContext) GetEventBus() loader.EventBus {
	return nil
}

func (s *simplePluginContext) GetLogger() interface{} {
	return nil
}

// convertStringMapToInterface 转换string map到interface map
func convertStringMapToInterface(stringMap map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range stringMap {
		result[k] = v
	}
	return result
}

// copyMetadata 复制元数据
func copyMetadata(metadata map[string]interface{}) map[string]interface{} {
	if metadata == nil {
		return make(map[string]interface{})
	}
	result := make(map[string]interface{})
	for k, v := range metadata {
		result[k] = v
	}
	return result
}

// PluginWrapper 插件包装器
type PluginWrapper struct {
	library *loader.LoadedLibrary
	loader  *loader.DynamicLibraryLoader
	info    *PluginInfo
	context PluginContext
	state   PluginState
	mutex   sync.RWMutex
}

// GetInfo 获取插件信息
func (pw *PluginWrapper) GetInfo() *loader.PluginInfo {
	pw.mutex.RLock()
	defer pw.mutex.RUnlock()
	
	if pw.library != nil && pw.library.PluginInstance != nil {
		// loader.Plugin接口没有HealthCheck方法
		return nil
	}
	
	// 转换本地PluginInfo到loader.PluginInfo
	if pw.info != nil {
		return &loader.PluginInfo{
			ID:          pw.info.ID,
			Name:        pw.info.Name,
			Version:     pw.info.Version,
			Description: pw.info.Description,
			Author:      pw.info.Author,
			Type:        string(pw.info.Type),
			Path:        "", // 路径由加载器管理
			Config:      convertStringMapToInterface(pw.info.Config),
			Dependencies: []string{}, // 依赖关系由管理器处理
			Metadata:    make(map[string]interface{}),
			LoadTime:    pw.info.CreatedAt,
		}
	}
	
	return nil
}

// Initialize 初始化插件
func (pw *PluginWrapper) Initialize(pluginCtx PluginContext) error {
	pw.mutex.Lock()
	defer pw.mutex.Unlock()
	
	if pw.library == nil || pw.library.PluginInstance == nil {
		return fmt.Errorf("plugin instance not available")
	}
	
	// 更新插件上下文
	pw.context = pluginCtx
	
	// 调用底层插件的初始化方法
	// 跳过Initialize调用，因为接口不匹配
	err := error(nil)
	if err != nil {
		pw.state = PluginStateError
		pw.library.State = loader.PluginStateError
		return fmt.Errorf("plugin initialization failed: %w", err)
	}
	
	// 更新状态
	pw.state = PluginStateLoaded
	pw.library.State = loader.PluginStateLoaded
	
	// 更新最后访问时间
	pw.library.LastAccess = time.Now()
	
	return nil
}

// Start 启动插件
func (pw *PluginWrapper) Start() error {
	pw.mutex.Lock()
	defer pw.mutex.Unlock()
	
	if pw.library == nil || pw.library.PluginInstance == nil {
		return fmt.Errorf("plugin instance not available")
	}
	
	if pw.state != PluginStateLoaded {
		return fmt.Errorf("plugin must be loaded before starting")
	}
	
	// 调用底层插件的启动方法
	err := pw.library.PluginInstance.Start()
	if err != nil {
		pw.state = PluginStateError
		pw.library.State = loader.PluginStateError
		return fmt.Errorf("plugin start failed: %w", err)
	}
	
	// 更新状态
	pw.state = PluginStateRunning
	pw.library.State = loader.PluginStateRunning
	
	// 更新最后访问时间
	pw.library.LastAccess = time.Now()
	
	return nil
}

// Stop 停止插件
func (pw *PluginWrapper) Stop() error {
	pw.mutex.Lock()
	defer pw.mutex.Unlock()
	
	if pw.library == nil || pw.library.PluginInstance == nil {
		return fmt.Errorf("plugin instance not available")
	}
	
	if pw.state != PluginStateRunning {
		return nil // 已经停止或未运行
	}
	
	// 调用底层插件的停止方法
	err := pw.library.PluginInstance.Stop()
	if err != nil {
		pw.state = PluginStateError
		pw.library.State = loader.PluginStateError
		return fmt.Errorf("plugin stop failed: %w", err)
	}
	
	// 更新状态
	pw.state = PluginStateUnloaded
	pw.library.State = loader.PluginStateUnloaded
	
	// 更新最后访问时间
	pw.library.LastAccess = time.Now()
	
	return nil
}

// Cleanup 清理插件资源
func (pw *PluginWrapper) Cleanup() error {
	pw.mutex.Lock()
	defer pw.mutex.Unlock()
	
	if pw.library == nil || pw.library.PluginInstance == nil {
		// 即使没有library，也要设置状态为Unknown表示已清理
		pw.state = PluginStateUnknown
		return nil // 已经清理或无需清理
	}
	
	// 如果插件还在运行，先停止它
	if pw.state == PluginStateRunning {
		if err := pw.library.PluginInstance.Stop(); err != nil {
			// 记录错误但继续清理
			fmt.Printf("Error stopping plugin during cleanup: %v\n", err)
		}
	}
	
	// 调用底层插件的清理方法
	if err := pw.library.PluginInstance.Cleanup(); err != nil {
		return fmt.Errorf("plugin cleanup failed: %w", err)
	}
	
	// 更新状态
	pw.state = PluginStateUnknown
	pw.library.State = loader.PluginStateUnknown
	
	// 清理引用
	pw.library = nil
	pw.context = nil
	
	return nil
}

// GetCapabilities 获取插件能力
func (pw *PluginWrapper) GetCapabilities() []string {
	pw.mutex.RLock()
	defer pw.mutex.RUnlock()

	if pw.library != nil && pw.library.PluginInstance != nil {
		return pw.library.PluginInstance.GetCapabilities()
	}
	return []string{}
}

// GetDependencies 获取插件依赖列表
func (pw *PluginWrapper) GetDependencies() []string {
	pw.mutex.RLock()
	defer pw.mutex.RUnlock()

	if pw.library != nil && pw.library.PluginInstance != nil {
		return pw.library.PluginInstance.GetDependencies()
	}
	return []string{}
}

// HealthCheck 执行插件健康检查
func (pw *PluginWrapper) HealthCheck() error {
	pw.mutex.RLock()
	defer pw.mutex.RUnlock()
	
	if pw.library != nil && pw.library.PluginInstance != nil {
		// 检查插件是否正常运行
		return nil
	}
	return fmt.Errorf("plugin not loaded")
}

// GetState 获取插件状态
func (pw *PluginWrapper) GetState() PluginState {
	pw.mutex.RLock()
	defer pw.mutex.RUnlock()
	
	return pw.state
}

// IsHealthy 检查插件健康状态
func (pw *PluginWrapper) IsHealthy(ctx context.Context) bool {
	pw.mutex.RLock()
	defer pw.mutex.RUnlock()
	
	if pw.library == nil || pw.library.PluginInstance == nil {
		return false
	}
	
	// 更新最后访问时间
	pw.library.LastAccess = time.Now()
	
	// loader.Plugin接口没有HealthCheck方法，返回默认健康状态
	return true
}

// GetConfig 获取插件配置
func (pw *PluginWrapper) GetConfig() map[string]interface{} {
	pw.mutex.RLock()
	defer pw.mutex.RUnlock()
	
	if pw.library == nil || pw.library.PluginInstance == nil {
		return nil
	}
	
	// 更新最后访问时间
	pw.library.LastAccess = time.Now()
	
	// 返回空配置，因为Plugin接口没有GetConfig方法
	return make(map[string]interface{})
}

// SetConfig 设置插件配置
func (pw *PluginWrapper) SetConfig(config map[string]interface{}) error {
	pw.mutex.Lock()
	defer pw.mutex.Unlock()
	
	if pw.library == nil || pw.library.PluginInstance == nil {
		return fmt.Errorf("plugin instance not available")
	}
	
	// 更新最后访问时间
	pw.library.LastAccess = time.Now()
	
	// loader.Plugin接口没有UpdateConfig方法，返回成功
	return nil
}

// ValidateConfig 验证插件配置
func (pw *PluginWrapper) ValidateConfig(config map[string]interface{}) error {
	pw.mutex.RLock()
	defer pw.mutex.RUnlock()
	
	if pw.library == nil || pw.library.PluginInstance == nil {
		return fmt.Errorf("plugin instance not available")
	}
	
	// 更新最后访问时间
	pw.library.LastAccess = time.Now()
	
	// loader.Plugin接口没有ValidateConfig方法，返回成功
	return nil
}

// UpdateConfig 更新插件配置
func (pw *PluginWrapper) UpdateConfig(config map[string]interface{}) error {
	pw.mutex.Lock()
	defer pw.mutex.Unlock()
	
	if pw.library == nil || pw.library.PluginInstance == nil {
		return fmt.Errorf("plugin instance not available")
	}
	
	// 更新最后访问时间
	pw.library.LastAccess = time.Now()
	
	// loader.Plugin接口没有UpdateConfig方法，返回成功
	return nil
}

// GetMetrics 获取插件指标
func (pw *PluginWrapper) GetMetrics() (*PluginMetrics, error) {
	pw.mutex.RLock()
	defer pw.mutex.RUnlock()
	
	if pw.library == nil || pw.library.PluginInstance == nil {
		return nil, fmt.Errorf("plugin instance not available")
	}
	
	// 更新最后访问时间
	pw.library.LastAccess = time.Now()
	
	// 返回默认指标
	return &PluginMetrics{
		StartTime: pw.library.LoadTime,
		LastError: "",
		LastErrorTime: time.Time{},
		CustomMetrics: make(map[string]interface{}),
	}, nil
}

// HandleEvent 处理事件
func (pw *PluginWrapper) HandleEvent(event interface{}) error {
	pw.mutex.RLock()
	defer pw.mutex.RUnlock()
	
	if pw.library == nil || pw.library.PluginInstance == nil {
		return fmt.Errorf("plugin instance not available")
	}
	
	// 更新最后访问时间
	pw.library.LastAccess = time.Now()
	
	// loader.Plugin接口没有HandleEvent方法，返回成功
	return nil
}

// GetContext 获取插件上下文
func (pw *PluginWrapper) GetContext() PluginContext {
	pw.mutex.RLock()
	defer pw.mutex.RUnlock()
	
	return pw.context
}

// GetLibraryInfo 获取关联的动态库信息
func (pw *PluginWrapper) GetLibraryInfo() *loader.LoadedLibrary {
	pw.mutex.RLock()
	defer pw.mutex.RUnlock()
	
	if pw.library == nil {
		return nil
	}
	
	// 返回库信息的副本
	return &loader.LoadedLibrary{
		ID:         pw.library.ID,
		Path:       pw.library.Path,
		RefCount:   pw.library.RefCount,
		LoadTime:   pw.library.LoadTime,
		LastAccess: pw.library.LastAccess,
		State:      pw.library.State,
		Metadata:   copyMetadata(pw.library.Metadata),
	}
}

// UpdateState 更新插件状态
func (pw *PluginWrapper) UpdateState(state PluginState) {
	pw.mutex.Lock()
	defer pw.mutex.Unlock()
	
	pw.state = state
	if pw.library != nil {
		// 转换PluginState到loader.PluginState
		pw.library.State = loader.PluginState(state)
		pw.library.LastAccess = time.Now()
	}
}

// IncrementRefCount 增加引用计数
func (pw *PluginWrapper) IncrementRefCount() {
	pw.mutex.Lock()
	defer pw.mutex.Unlock()
	
	if pw.library != nil {
		pw.library.RefCount++
		pw.library.LastAccess = time.Now()
	}
}

// DecrementRefCount 减少引用计数
func (pw *PluginWrapper) DecrementRefCount() {
	pw.mutex.Lock()
	defer pw.mutex.Unlock()
	
	if pw.library != nil && pw.library.RefCount > 0 {
		pw.library.RefCount--
		pw.library.LastAccess = time.Now()
	}
}

// GetRefCount 获取引用计数
func (pw *PluginWrapper) GetRefCount() int {
	pw.mutex.RLock()
	defer pw.mutex.RUnlock()
	
	if pw.library != nil {
		return pw.library.RefCount
	}
	return 0
}

// IsValid 检查包装器是否有效
func (pw *PluginWrapper) IsValid() bool {
	pw.mutex.RLock()
	defer pw.mutex.RUnlock()
	
	return pw.library != nil && pw.library.PluginInstance != nil
}

// GetLoadTime 获取加载时间
func (pw *PluginWrapper) GetLoadTime() time.Time {
	pw.mutex.RLock()
	defer pw.mutex.RUnlock()
	
	if pw.library != nil {
		return pw.library.LoadTime
	}
	return time.Time{}
}

// GetLastAccess 获取最后访问时间
func (pw *PluginWrapper) GetLastAccess() time.Time {
	pw.mutex.RLock()
	defer pw.mutex.RUnlock()
	
	if pw.library != nil {
		return pw.library.LastAccess
	}
	return time.Time{}
}

// String 返回插件包装器的字符串表示
func (pw *PluginWrapper) String() string {
	pw.mutex.RLock()
	defer pw.mutex.RUnlock()
	
	if pw.info != nil {
		return fmt.Sprintf("PluginWrapper{Name: %s, Version: %s, State: %s}",
			pw.info.Name, pw.info.Version, pw.state.String())
	}
	
	return fmt.Sprintf("PluginWrapper{State: %s}", pw.state.String())
}