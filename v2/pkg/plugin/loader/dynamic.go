package loader

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
	"unsafe"

	"github.com/ebitengine/purego"
)

// DynamicPluginLoader 动态库插件加载器
type DynamicPluginLoader struct {
	// loadedLibs 已加载的动态库映射表
	loadedLibs map[string]*LoadedLibrary
	
	// mutex 读写锁，保护并发访问
	mutex sync.RWMutex
	
	// config 加载器配置
	config *DynamicLoaderConfig
	
	// logger 日志记录器
	logger *slog.Logger
	
	// securityManager 安全管理器
	securityManager SecurityManager
	
	// ctx 上下文
	ctx context.Context
	
	// cancel 取消函数
	cancel context.CancelFunc
}

// LoadedLibrary 已加载的动态库信息
type LoadedLibrary struct {
	// ID 插件唯一标识符
	ID string
	
	// LibraryHandle purego动态库句柄
	LibraryHandle uintptr
	
	// Path 插件文件路径
	Path string
	
	// Handle 动态库句柄
	Handle unsafe.Pointer
	
	// Symbols 符号表映射
	Symbols map[string]unsafe.Pointer
	
	// PluginInstance 插件实例
	PluginInstance Plugin
	
	// RefCount 引用计数
	RefCount int
	
	// LoadTime 加载时间
	LoadTime time.Time
	
	// LastAccess 最后访问时间
	LastAccess time.Time
	
	// State 插件状态
	State PluginState
	
	// Metadata 元数据
	Metadata map[string]interface{}
}

// DynamicLoaderConfig 动态加载器配置
type DynamicLoaderConfig struct {
	// MaxPlugins 最大插件数量
	MaxPlugins int
	
	// LoadTimeout 加载超时时间
	LoadTimeout time.Duration
	
	// UnloadTimeout 卸载超时时间
	UnloadTimeout time.Duration
	
	// EnableSymbolCache 是否启用符号缓存
	EnableSymbolCache bool
	
	// ValidateSignature 是否验证插件签名
	ValidateSignature bool
	
	// AllowedPaths 允许的插件路径列表
	AllowedPaths []string
	
	// RequiredSymbols 必需的符号列表
	RequiredSymbols []string
	
	// ResourceLimits 资源限制
	ResourceLimits *ResourceLimits
}

// SecurityManager 安全管理器接口
type SecurityManager interface {
	// ValidatePlugin 验证插件安全性
	ValidatePlugin(pluginPath string) error
	
	// CheckPermissions 检查权限
	CheckPermissions(operation string, resource string) error
}

// NewDynamicPluginLoader 创建新的动态插件加载器
func NewDynamicPluginLoader(securityManager SecurityManager, logger *slog.Logger) *DynamicPluginLoader {
	if logger == nil {
		return nil
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	config := &DynamicLoaderConfig{
		MaxPlugins:        10,
		LoadTimeout:       30 * time.Second,
		UnloadTimeout:     10 * time.Second,
		EnableSymbolCache: true,
		ValidateSignature: true,
		AllowedPaths:      []string{},
		RequiredSymbols:   []string{"GetPlugin"},
		ResourceLimits:    &ResourceLimits{Enabled: false},
	}
	
	return &DynamicPluginLoader{
		loadedLibs:      make(map[string]*LoadedLibrary),
		config:          config,
		logger:          logger,
		securityManager: securityManager,
		ctx:             ctx,
		cancel:          cancel,
	}
}

// LoadPlugin 从指定路径加载插件
func (d *DynamicPluginLoader) LoadPlugin(ctx context.Context, pluginPath string) (Plugin, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	
	// 安全验证
	if d.securityManager != nil {
		if err := d.securityManager.ValidatePlugin(pluginPath); err != nil {
			return nil, fmt.Errorf("plugin validation failed: %w", err)
		}
	}
	
	// 使用purego加载动态库
	libHandle, err := purego.Dlopen(pluginPath, purego.RTLD_NOW)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin: %w", err)
	}

	// 查找GetPlugin符号
	getPluginSymbol, err := purego.Dlsym(libHandle, "GetPlugin")
	if err != nil {
		return nil, fmt.Errorf("GetPlugin symbol not found: %w", err)
	}

	// 注册函数并调用
	var getPlugin func() Plugin
	purego.RegisterLibFunc(&getPlugin, libHandle, "GetPlugin")

	// 创建插件实例
	pluginInstance := getPlugin()
	if pluginInstance == nil {
		return nil, fmt.Errorf("plugin instance is nil")
	}
	
	// 创建加载库信息
	pluginID := pluginInstance.GetInfo().ID
	loadedLib := &LoadedLibrary{
		ID:             pluginID,
		LibraryHandle:  libHandle,
		Path:           pluginPath,
		Handle:         unsafe.Pointer(getPluginSymbol),
		PluginInstance: pluginInstance,
		RefCount:       1,
		LoadTime:       time.Now(),
		LastAccess:     time.Now(),
		State:          PluginStateLoaded,
		Metadata:       make(map[string]interface{}),
		Symbols:        make(map[string]unsafe.Pointer),
	}
	
	d.loadedLibs[pluginID] = loadedLib
	
	d.logger.Info("Plugin loaded successfully", "id", pluginID, "path", pluginPath)
	return pluginInstance, nil
}

// UnloadPlugin 卸载指定的插件
func (d *DynamicPluginLoader) UnloadPlugin(ctx context.Context, pluginID string) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	
	loadedLib, exists := d.loadedLibs[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}
	
	// 清理插件资源
	if err := loadedLib.PluginInstance.Cleanup(); err != nil {
		d.logger.Warn("Plugin cleanup failed", "id", pluginID, "error", err)
	}

	// 关闭动态库
	if err := purego.Dlclose(loadedLib.LibraryHandle); err != nil {
		d.logger.Warn("Failed to close dynamic library", "id", pluginID, "error", err)
	}

	// 从映射中删除
	delete(d.loadedLibs, pluginID)
	
	d.logger.Info("Plugin unloaded successfully", "id", pluginID)
	return nil
}

// GetLoadedPlugins 获取已加载的插件列表
func (d *DynamicPluginLoader) GetLoadedPlugins() map[string]Plugin {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	
	result := make(map[string]Plugin)
	for id, lib := range d.loadedLibs {
		result[id] = lib.PluginInstance
	}
	return result
}

// GetPluginInfo 获取插件信息
func (d *DynamicPluginLoader) GetPluginInfo(pluginID string) (*PluginInfo, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	
	loadedLib, exists := d.loadedLibs[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}
	
	return loadedLib.PluginInstance.GetInfo(), nil
}

// ValidatePlugin 验证插件是否有效
func (d *DynamicPluginLoader) ValidatePlugin(pluginPath string) error {
	return d.securityManager.ValidatePlugin(pluginPath)
}

// GetLoaderType 获取加载器类型
func (d *DynamicPluginLoader) GetLoaderType() PluginType {
	return PluginTypeDynamicLibrary
}

// GetLoaderInfo 获取加载器信息
func (d *DynamicPluginLoader) GetLoaderInfo() map[string]interface{} {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	
	return map[string]interface{}{
		"type":           "dynamic",
		"loaded_count":   len(d.loadedLibs),
		"max_plugins":    d.config.MaxPlugins,
		"load_timeout":   d.config.LoadTimeout,
		"unload_timeout": d.config.UnloadTimeout,
	}
}

// Shutdown 关闭加载器
func (d *DynamicPluginLoader) Shutdown(ctx context.Context) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	
	// 卸载所有插件
	for pluginID := range d.loadedLibs {
		if err := d.UnloadPlugin(ctx, pluginID); err != nil {
			d.logger.Warn("Failed to unload plugin during shutdown", "id", pluginID, "error", err)
		}
	}
	
	// 取消上下文
	d.cancel()
	
	d.logger.Info("Dynamic plugin loader shutdown completed")
	return nil
}

// ReloadPlugin 重新加载插件
func (d *DynamicPluginLoader) ReloadPlugin(ctx context.Context, pluginID string) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	
	loadedLib, exists := d.loadedLibs[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}
	
	pluginPath := loadedLib.Path
	
	// 先卸载
	if err := d.UnloadPlugin(ctx, pluginID); err != nil {
		return fmt.Errorf("failed to unload plugin for reload: %w", err)
	}
	
	// 重新加载
	_, err := d.LoadPlugin(ctx, pluginPath)
	if err != nil {
		return fmt.Errorf("failed to reload plugin: %w", err)
	}
	
	d.logger.Info("Plugin reloaded successfully", "id", pluginID)
	return nil
}

// IsPluginLoaded 检查插件是否已加载
func (d *DynamicPluginLoader) IsPluginLoaded(pluginID string) bool {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	
	_, exists := d.loadedLibs[pluginID]
	return exists
}

// Cleanup 清理资源
func (d *DynamicPluginLoader) Cleanup() error {
	return d.Shutdown(context.Background())
}