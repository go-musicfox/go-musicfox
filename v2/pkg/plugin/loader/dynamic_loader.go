// Package loader 实现动态链接库插件加载器
package loader

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"
	"unsafe"

	"github.com/ebitengine/purego"
)

// NewDynamicLibraryLoader 创建新的动态链接库加载器
func NewDynamicLibraryLoader(ctx context.Context, config *DynamicLoaderConfig) *DynamicLibraryLoader {
	loaderCtx, cancel := context.WithCancel(ctx)
	
	// 如果config为nil，使用默认配置
	if config == nil {
		config = DefaultDynamicLoaderConfig()
	}
	
	return &DynamicLibraryLoader{
		ctx:        loaderCtx,
		cancel:     cancel,
		loadedLibs: make(map[string]*LoadedLibrary),
		config:     config,

	}
}

// DefaultDynamicLoaderConfig 返回默认的动态加载器配置
func DefaultDynamicLoaderConfig() *DynamicLoaderConfig {
	return &DynamicLoaderConfig{
		MaxPlugins:        100,
		LoadTimeout:       30 * time.Second,
		UnloadTimeout:     10 * time.Second,
		EnableSymbolCache: true,
		ValidateSignature: false,
		AllowedPaths:      []string{},
		RequiredSymbols:   []string{"NewPlugin", "GetPluginInfo"},
		ResourceLimits: &ResourceLimits{
			MaxMemoryMB:     100,
			MaxCPUPercent:   50.0,
			MaxFileHandles:  50,
			MaxGoroutines:   100,
			MaxNetworkConn:  10,
			ExecutionTimeout: 30 * time.Second,
			Enabled:        true,
		},
	}
}

// LoadPlugin 加载插件
func (dl *DynamicLibraryLoader) LoadPlugin(ctx context.Context, pluginPath string) (Plugin, error) {
	// 创建带超时的上下文
	loadCtx, cancel := context.WithTimeout(ctx, dl.config.LoadTimeout)
	defer cancel()
	
	// 验证插件路径
	if err := dl.validatePluginPath(pluginPath); err != nil {
		return nil, fmt.Errorf("invalid plugin path: %w", err)
	}
	
	// 生成插件ID
	pluginID := dl.generatePluginID(pluginPath)
	
	// 检查是否已加载
	dl.mutex.RLock()
	if lib, exists := dl.loadedLibs[pluginID]; exists {
		lib.RefCount++
		lib.LastAccess = time.Now()
		dl.mutex.RUnlock()
		return &PluginWrapper{
			library: lib,
			loader:  dl,
			info:    lib.PluginInstance.GetInfo(),
			state:   lib.State,
		}, nil
	}
	dl.mutex.RUnlock()
	
	// 检查插件数量限制
	if len(dl.loadedLibs) >= dl.config.MaxPlugins {
		return nil, fmt.Errorf("maximum number of plugins (%d) reached", dl.config.MaxPlugins)
	}
	
	// 加载动态库
	loadedLib, err := dl.loadDynamicLibrary(loadCtx, pluginPath, pluginID)
	if err != nil {
		return nil, fmt.Errorf("failed to load dynamic library: %w", err)
	}
	
	// 创建插件包装器
	wrapper := &PluginWrapper{
		library: loadedLib,
		loader:  dl,
		info:    loadedLib.PluginInstance.GetInfo(),
		state:   PluginStateLoaded,
	}
	
	// 更新状态
	loadedLib.State = PluginStateLoaded
	loadedLib.LastAccess = time.Now()
	
	return wrapper, nil
}

// loadDynamicLibrary 加载动态库的具体实现
func (dl *DynamicLibraryLoader) loadDynamicLibrary(ctx context.Context, pluginPath, pluginID string) (*LoadedLibrary, error) {
	// 记录加载开始时间
	defer func() {

	}()
	
	// 根据操作系统选择加载方式
	switch runtime.GOOS {
	case "linux", "darwin":
		return dl.loadUnixPlugin(ctx, pluginPath, pluginID)
	case "windows":
		return dl.loadWindowsPlugin(ctx, pluginPath, pluginID)
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// loadUnixPlugin 加载Unix系统的动态库
func (dl *DynamicLibraryLoader) loadUnixPlugin(ctx context.Context, pluginPath, pluginID string) (*LoadedLibrary, error) {
	// 使用purego加载动态库
	libHandle, err := purego.Dlopen(pluginPath, purego.RTLD_NOW)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin: %w", err)
	}
	
	// 查找必需的符号
	symbols := make(map[string]unsafe.Pointer)
	for _, symbolName := range dl.config.RequiredSymbols {
		sym, err := purego.Dlsym(libHandle, symbolName)
		if err != nil {
			// 关闭库句柄
			purego.Dlclose(libHandle)
			return nil, fmt.Errorf("required symbol '%s' not found: %w", symbolName, err)
		}
		symbols[symbolName] = unsafe.Pointer(sym)
	}
	
	// 查找NewPlugin符号
	newPluginSymbol, err := purego.Dlsym(libHandle, "NewPlugin")
	if err != nil {
		purego.Dlclose(libHandle)
		return nil, fmt.Errorf("NewPlugin function not found: %w", err)
	}
	
	// 注册NewPlugin函数
	var newPluginFunc func() Plugin
	purego.RegisterLibFunc(&newPluginFunc, libHandle, "NewPlugin")
	
	// 创建插件实例
	pluginInstance := newPluginFunc()
	if pluginInstance == nil {
		purego.Dlclose(libHandle)
		return nil, fmt.Errorf("plugin instance creation failed")
	}
	
	// 验证插件接口
	if err := dl.validatePluginInterface(pluginInstance); err != nil {
		purego.Dlclose(libHandle)
		return nil, fmt.Errorf("plugin interface validation failed: %w", err)
	}
	
	// 创建LoadedLibrary实例
	loadedLib := &LoadedLibrary{
		ID:             pluginID,
		LibraryHandle:  libHandle,
		Path:           pluginPath,
		Handle:         unsafe.Pointer(newPluginSymbol),
		Symbols:        symbols,
		PluginInstance: pluginInstance,
		RefCount:       1,
		LoadTime:       time.Now(),
		LastAccess:     time.Now(),
		State:          PluginStateLoaded,
		Metadata:       make(map[string]interface{}),
	}
	
	// 添加到已加载列表
	dl.mutex.Lock()
	dl.loadedLibs[pluginID] = loadedLib
	dl.mutex.Unlock()
	
	return loadedLib, nil
}

// loadWindowsPlugin 加载Windows系统的动态库
func (dl *DynamicLibraryLoader) loadWindowsPlugin(ctx context.Context, pluginPath, pluginID string) (*LoadedLibrary, error) {
	// 使用purego加载Windows动态库
	libHandle, err := purego.Dlopen(pluginPath, purego.RTLD_NOW)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin: %w", err)
	}
	
	// 查找必需的符号
	symbols := make(map[string]unsafe.Pointer)
	for _, symbolName := range dl.config.RequiredSymbols {
		sym, err := purego.Dlsym(libHandle, symbolName)
		if err != nil {
			// 关闭库句柄
			purego.Dlclose(libHandle)
			return nil, fmt.Errorf("required symbol '%s' not found: %w", symbolName, err)
		}
		symbols[symbolName] = unsafe.Pointer(sym)
	}
	
	// 查找NewPlugin符号
	newPluginSymbol, err := purego.Dlsym(libHandle, "NewPlugin")
	if err != nil {
		purego.Dlclose(libHandle)
		return nil, fmt.Errorf("NewPlugin function not found: %w", err)
	}
	
	// 注册NewPlugin函数
	var newPluginFunc func() Plugin
	purego.RegisterLibFunc(&newPluginFunc, libHandle, "NewPlugin")
	
	// 创建插件实例
	pluginInstance := newPluginFunc()
	if pluginInstance == nil {
		purego.Dlclose(libHandle)
		return nil, fmt.Errorf("plugin instance creation failed")
	}
	
	// 验证插件接口
	if err := dl.validatePluginInterface(pluginInstance); err != nil {
		purego.Dlclose(libHandle)
		return nil, fmt.Errorf("plugin interface validation failed: %w", err)
	}
	
	// 创建LoadedLibrary实例
	loadedLib := &LoadedLibrary{
		ID:             pluginID,
		LibraryHandle:  libHandle,
		Path:           pluginPath,
		Handle:         unsafe.Pointer(newPluginSymbol),
		Symbols:        symbols,
		PluginInstance: pluginInstance,
		RefCount:       1,
		LoadTime:       time.Now(),
		LastAccess:     time.Now(),
		State:          PluginStateLoaded,
		Metadata:       make(map[string]interface{}),
	}
	
	// 添加到已加载列表
	dl.mutex.Lock()
	dl.loadedLibs[pluginID] = loadedLib
	dl.mutex.Unlock()
	
	return loadedLib, nil
}

// UnloadPlugin 卸载插件
func (dl *DynamicLibraryLoader) UnloadPlugin(ctx context.Context, pluginID string) error {

	
	dl.mutex.Lock()
	defer dl.mutex.Unlock()
	
	lib, exists := dl.loadedLibs[pluginID]
	if !exists {
		return fmt.Errorf("plugin '%s' not found", pluginID)
	}
	
	// 减少引用计数
	lib.RefCount--
	if lib.RefCount > 0 {
		return nil // 还有其他引用，不卸载
	}
	
	// 停止插件
	if lib.PluginInstance != nil {
		if err := lib.PluginInstance.Stop(); err != nil {
			return fmt.Errorf("failed to stop plugin: %w", err)
		}
	}
	
	// 清理资源
	if err := dl.cleanupLibrary(lib); err != nil {
		return fmt.Errorf("failed to cleanup library: %w", err)
	}
	
	// 从已加载列表中移除
	delete(dl.loadedLibs, pluginID)
	
	return nil
}

// GetLoadedPlugins 获取已加载的插件列表
func (dl *DynamicLibraryLoader) GetLoadedPlugins() map[string]Plugin {
	dl.mutex.RLock()
	defer dl.mutex.RUnlock()
	
	plugins := make(map[string]Plugin)
	for id, lib := range dl.loadedLibs {
		if lib.PluginInstance != nil {
			plugins[id] = lib.PluginInstance
		}
	}
	
	return plugins
}

// IsPluginLoaded 检查插件是否已加载
func (dl *DynamicLibraryLoader) IsPluginLoaded(pluginID string) bool {
	dl.mutex.RLock()
	defer dl.mutex.RUnlock()
	
	_, exists := dl.loadedLibs[pluginID]
	return exists
}

// GetPluginInfo 获取插件信息
func (dl *DynamicLibraryLoader) GetPluginInfo(pluginID string) (*PluginInfo, error) {
	dl.mutex.RLock()
	defer dl.mutex.RUnlock()
	
	lib, exists := dl.loadedLibs[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin '%s' not found", pluginID)
	}
	
	if lib.PluginInstance == nil {
		return nil, fmt.Errorf("plugin instance is nil")
	}
	
	return lib.PluginInstance.GetInfo(), nil
}

// ReloadPlugin 重新加载插件
func (dl *DynamicLibraryLoader) ReloadPlugin(ctx context.Context, pluginID string) error {
	dl.mutex.RLock()
	lib, exists := dl.loadedLibs[pluginID]
	if !exists {
		dl.mutex.RUnlock()
		return fmt.Errorf("plugin '%s' not found", pluginID)
	}
	pluginPath := lib.Path
	dl.mutex.RUnlock()
	
	// 卸载现有插件
	if err := dl.UnloadPlugin(ctx, pluginID); err != nil {
		return fmt.Errorf("failed to unload plugin for reload: %w", err)
	}
	
	// 重新加载插件
	_, err := dl.LoadPlugin(ctx, pluginPath)
	if err != nil {
		return fmt.Errorf("failed to reload plugin: %w", err)
	}
	
	return nil
}

// ValidatePlugin 验证插件
func (dl *DynamicLibraryLoader) ValidatePlugin(pluginPath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(pluginPath); err != nil {
		return fmt.Errorf("plugin file not found: %w", err)
	}
	
	// 检查文件扩展名
	if !dl.isValidPluginExtension(pluginPath) {
		return fmt.Errorf("invalid plugin file extension")
	}
	
	// 检查路径是否在允许列表中
	if err := dl.validatePluginPath(pluginPath); err != nil {
		return err
	}
	
	return nil
}

// GetLoaderType 获取加载器类型
func (dl *DynamicLibraryLoader) GetLoaderType() PluginType {
	return PluginTypeDynamicLibrary
}

// Cleanup 清理资源
func (dl *DynamicLibraryLoader) Cleanup() error {
	dl.mutex.Lock()
	defer dl.mutex.Unlock()
	
	// 停止所有插件
	for id, lib := range dl.loadedLibs {
		if lib.PluginInstance != nil {
			if err := lib.PluginInstance.Stop(); err != nil {
				// 记录错误但继续清理其他插件
				fmt.Printf("Error stopping plugin %s: %v\n", id, err)
			}
		}
		
		if err := dl.cleanupLibrary(lib); err != nil {
			fmt.Printf("Error cleaning up library %s: %v\n", id, err)
		}
	}
	
	// 清空已加载列表
	dl.loadedLibs = make(map[string]*LoadedLibrary)
	
	// 取消上下文
	if dl.cancel != nil {
		dl.cancel()
	}
	
	return nil
}