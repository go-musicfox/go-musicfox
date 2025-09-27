// Package loader 实现插件加载器
package loader

import (
	"context"
	"sync"
	"time"
	"unsafe"
)



// DynamicLibraryLoader 动态链接库插件加载器
// 负责加载和管理动态链接库(.so, .dll, .dylib)插件
type DynamicLibraryLoader struct {
	// loadedLibs 已加载的动态库映射表
	loadedLibs map[string]*LoadedLibrary
	
	// mutex 读写锁，保护并发访问
	mutex sync.RWMutex
	
	// config 加载器配置
	config *DynamicLoaderConfig
	
	// ctx 上下文
	ctx context.Context
	
	// cancel 取消函数
	cancel context.CancelFunc
}





// SymbolInfo 符号信息
type SymbolInfo struct {
	// Name 符号名称
	Name string
	
	// Address 符号地址
	Address unsafe.Pointer
	
	// Type 符号类型
	Type SymbolType
	
	// Size 符号大小
	Size uintptr
	
	// Exported 是否导出
	Exported bool
}

// SymbolType 符号类型枚举
type SymbolType int

const (
	SymbolTypeFunction SymbolType = iota
	SymbolTypeVariable
	SymbolTypeConstant
	SymbolTypeType
)

// String 返回符号类型的字符串表示
func (s SymbolType) String() string {
	switch s {
	case SymbolTypeFunction:
		return "function"
	case SymbolTypeVariable:
		return "variable"
	case SymbolTypeConstant:
		return "constant"
	case SymbolTypeType:
		return "type"
	default:
		return "unknown"
	}
}

// PluginWrapper 插件包装器
// 为动态库插件提供统一的接口实现
type PluginWrapper struct {
	// library 关联的动态库
	library *LoadedLibrary
	
	// loader 加载器引用
	loader *DynamicLibraryLoader
	
	// info 插件信息
	info *PluginInfo
	
	// context 插件上下文
	context PluginContext
	
	// state 插件状态
	state PluginState
	
	// mutex 状态锁
	mutex sync.RWMutex
}

// GetInfo 获取插件信息
func (pw *PluginWrapper) GetInfo() *PluginInfo {
	pw.mutex.RLock()
	defer pw.mutex.RUnlock()
	return pw.info
}

// GetCapabilities 获取插件能力
func (pw *PluginWrapper) GetCapabilities() []string {
	if pw.library != nil && pw.library.PluginInstance != nil {
		return pw.library.PluginInstance.GetCapabilities()
	}
	return []string{}
}

// GetDependencies 获取插件依赖
func (pw *PluginWrapper) GetDependencies() []string {
	if pw.library != nil && pw.library.PluginInstance != nil {
		return pw.library.PluginInstance.GetDependencies()
	}
	return []string{}
}

// Initialize 初始化插件
func (pw *PluginWrapper) Initialize(ctx PluginContext) error {
	pw.mutex.Lock()
	defer pw.mutex.Unlock()
	pw.context = ctx
	if pw.library != nil && pw.library.PluginInstance != nil {
		return pw.library.PluginInstance.Initialize(ctx)
	}
	return nil
}

// Start 启动插件
func (pw *PluginWrapper) Start() error {
	pw.mutex.Lock()
	defer pw.mutex.Unlock()
	pw.state = PluginStateRunning
	if pw.library != nil && pw.library.PluginInstance != nil {
		return pw.library.PluginInstance.Start()
	}
	return nil
}

// Stop 停止插件
func (pw *PluginWrapper) Stop() error {
	pw.mutex.Lock()
	defer pw.mutex.Unlock()
	pw.state = PluginStateStopped
	if pw.library != nil && pw.library.PluginInstance != nil {
		return pw.library.PluginInstance.Stop()
	}
	return nil
}

// Cleanup 清理插件资源
func (pw *PluginWrapper) Cleanup() error {
	pw.mutex.Lock()
	defer pw.mutex.Unlock()
	pw.state = PluginStateUnloaded
	if pw.library != nil && pw.library.PluginInstance != nil {
		return pw.library.PluginInstance.Cleanup()
	}
	return nil
}

// GetMetrics 获取插件指标
func (pw *PluginWrapper) GetMetrics() (*PluginMetrics, error) {
	if pw.library != nil && pw.library.PluginInstance != nil {
		return pw.library.PluginInstance.GetMetrics()
	}
	return &PluginMetrics{}, nil
}

// GetState 获取插件状态
func (pw *PluginWrapper) GetState() PluginState {
	pw.mutex.RLock()
	defer pw.mutex.RUnlock()
	return pw.state
}

// SetConfig 设置插件配置
func (pw *PluginWrapper) SetConfig(config map[string]interface{}) error {
	pw.mutex.Lock()
	defer pw.mutex.Unlock()
	// 更新插件信息中的配置
	if pw.info != nil {
		for k, v := range config {
			pw.info.Config[k] = v
		}
	}
	return nil
}

// UpdateState 更新插件状态
func (pw *PluginWrapper) UpdateState(state PluginState) error {
	pw.mutex.Lock()
	defer pw.mutex.Unlock()
	pw.state = state
	if pw.library != nil {
		pw.library.State = state
	}
	return nil
}

// LoaderStats 加载器统计信息
type LoaderStats struct {
	// TotalLoaded 总加载数量
	TotalLoaded int64
	
	// TotalUnloaded 总卸载数量
	TotalUnloaded int64
	
	// CurrentLoaded 当前加载数量
	CurrentLoaded int64
	
	// LoadErrors 加载错误数量
	LoadErrors int64
	
	// UnloadErrors 卸载错误数量
	UnloadErrors int64
	
	// AverageLoadTime 平均加载时间
	AverageLoadTime time.Duration
	
	// AverageUnloadTime 平均卸载时间
	AverageUnloadTime time.Duration
	
	// LastLoadTime 最后加载时间
	LastLoadTime time.Time
	
	// LastUnloadTime 最后卸载时间
	LastUnloadTime time.Time
}