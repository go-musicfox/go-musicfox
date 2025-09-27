// Package loader 提供动态库加载器的工具函数
package loader

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	
	"github.com/ebitengine/purego"
)

// generatePluginID 生成插件唯一标识符
func (dl *DynamicLibraryLoader) generatePluginID(pluginPath string) string {
	// 从路径中提取文件名（不含扩展名）作为插件ID
	// 处理跨平台路径分隔符
	filename := filepath.Base(pluginPath)
	
	// 如果 filepath.Base 没有正确处理路径（比如在 Unix 系统上处理 Windows 路径）
	// 手动处理不同的路径分隔符
	if filename == pluginPath {
		// 尝试 Windows 路径分隔符
		if lastBackslash := strings.LastIndex(pluginPath, "\\"); lastBackslash != -1 {
			filename = pluginPath[lastBackslash+1:]
		} else if lastSlash := strings.LastIndex(pluginPath, "/"); lastSlash != -1 {
			// 尝试 Unix 路径分隔符
			filename = pluginPath[lastSlash+1:]
		}
	}
	
	ext := filepath.Ext(filename)
	if ext != "" {
		filename = filename[:len(filename)-len(ext)]
	}
	return filename
}

// validatePluginPath 验证插件路径
func (dl *DynamicLibraryLoader) validatePluginPath(pluginPath string) error {
	// 检查路径是否为空
	if pluginPath == "" {
		return fmt.Errorf("plugin path cannot be empty")
	}
	
	// 获取绝对路径
	absPath, err := filepath.Abs(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	
	// 如果配置了允许的路径列表，检查路径是否在列表中
	if len(dl.config.AllowedPaths) > 0 {
		allowed := false
		for _, allowedPath := range dl.config.AllowedPaths {
			allowedAbs, err := filepath.Abs(allowedPath)
			if err != nil {
				continue
			}
			
			// 检查是否在允许的目录下
			if strings.HasPrefix(absPath, allowedAbs) {
				allowed = true
				break
			}
		}
		
		if !allowed {
			return fmt.Errorf("plugin path not in allowed paths")
		}
	}
	
	return nil
}

// isValidPluginExtension 检查插件文件扩展名是否有效
func (dl *DynamicLibraryLoader) isValidPluginExtension(pluginPath string) bool {
	ext := strings.ToLower(filepath.Ext(pluginPath))
	
	switch runtime.GOOS {
	case "linux":
		return ext == ".so"
	case "darwin":
		return ext == ".so" || ext == ".dylib"
	case "windows":
		return ext == ".dll"
	default:
		return false
	}
}

// validatePluginInterface 验证插件接口
func (dl *DynamicLibraryLoader) validatePluginInterface(pluginInstance Plugin) error {
	if pluginInstance == nil {
		return fmt.Errorf("plugin instance is nil")
	}
	
	// 检查基本接口方法
	info := pluginInstance.GetInfo()
	if info == nil {
		return fmt.Errorf("plugin info is nil")
	}
	
	if info.Name == "" {
		return fmt.Errorf("plugin name is empty")
	}
	
	if info.Version == "" {
		return fmt.Errorf("plugin version is empty")
	}
	
	// 插件类型检查已通过动态库加载方式验证
	
	return nil
}

// cleanupLibrary 清理动态库资源
func (dl *DynamicLibraryLoader) cleanupLibrary(lib *LoadedLibrary) error {
	if lib == nil {
		return nil
	}
	
	// 清理符号表
	lib.Symbols = nil
	
	// 清理元数据
	lib.Metadata = nil
	
	// 更新状态
	lib.State = PluginStateStopped
	
	// 使用purego关闭动态库
	if lib.LibraryHandle != 0 {
		if err := purego.Dlclose(lib.LibraryHandle); err != nil {
			return fmt.Errorf("failed to close library handle: %w", err)
		}
		lib.LibraryHandle = 0
	}
	
	return nil
}

// GetStats 获取加载器统计信息
func (dl *DynamicLibraryLoader) GetStats() *LoaderStats {
	dl.mutex.RLock()
	defer dl.mutex.RUnlock()
	
	stats := &LoaderStats{
		CurrentLoaded: int64(len(dl.loadedLibs)),
	}
	
	// 计算统计信息
	var totalLoadTime time.Duration
	loadCount := int64(0)
	
	for _, lib := range dl.loadedLibs {
		if !lib.LoadTime.IsZero() {
			totalLoadTime += time.Since(lib.LoadTime)
			loadCount++
		}
	}
	
	if loadCount > 0 {
		stats.AverageLoadTime = totalLoadTime / time.Duration(loadCount)
	}
	
	return stats
}

// GetLibraryInfo 获取已加载库的详细信息
func (dl *DynamicLibraryLoader) GetLibraryInfo(pluginID string) (*LoadedLibrary, error) {
	dl.mutex.RLock()
	defer dl.mutex.RUnlock()
	
	lib, exists := dl.loadedLibs[pluginID]
	if !exists {
		return nil, fmt.Errorf("library '%s' not found", pluginID)
	}
	
	// 返回库信息的副本，避免外部修改
	return &LoadedLibrary{
		ID:             lib.ID,
		Path:           lib.Path,
		RefCount:       lib.RefCount,
		LoadTime:       lib.LoadTime,
		LastAccess:     lib.LastAccess,
		State:          lib.State,
		Metadata:       copyMetadata(lib.Metadata),
	}, nil
}

// copyMetadata 复制元数据映射
func copyMetadata(original map[string]interface{}) map[string]interface{} {
	if original == nil {
		return nil
	}
	
	copy := make(map[string]interface{})
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

// SetLibraryMetadata 设置库的元数据
func (dl *DynamicLibraryLoader) SetLibraryMetadata(pluginID string, key string, value interface{}) error {
	dl.mutex.Lock()
	defer dl.mutex.Unlock()
	
	lib, exists := dl.loadedLibs[pluginID]
	if !exists {
		return fmt.Errorf("library '%s' not found", pluginID)
	}
	
	if lib.Metadata == nil {
		lib.Metadata = make(map[string]interface{})
	}
	
	lib.Metadata[key] = value
	return nil
}

// GetLibraryMetadata 获取库的元数据
func (dl *DynamicLibraryLoader) GetLibraryMetadata(pluginID string, key string) (interface{}, error) {
	dl.mutex.RLock()
	defer dl.mutex.RUnlock()
	
	lib, exists := dl.loadedLibs[pluginID]
	if !exists {
		return nil, fmt.Errorf("library '%s' not found", pluginID)
	}
	
	if lib.Metadata == nil {
		return nil, fmt.Errorf("no metadata found")
	}
	
	value, exists := lib.Metadata[key]
	if !exists {
		return nil, fmt.Errorf("metadata key '%s' not found", key)
	}
	
	return value, nil
}

// UpdateLastAccess 更新库的最后访问时间
func (dl *DynamicLibraryLoader) UpdateLastAccess(pluginID string) {
	dl.mutex.Lock()
	defer dl.mutex.Unlock()
	
	if lib, exists := dl.loadedLibs[pluginID]; exists {
		lib.LastAccess = time.Now()
	}
}

// GetLoadedLibrariesByState 根据状态获取已加载的库
func (dl *DynamicLibraryLoader) GetLoadedLibrariesByState(state PluginState) []string {
	dl.mutex.RLock()
	defer dl.mutex.RUnlock()
	
	var libraries []string
	for id, lib := range dl.loadedLibs {
		if lib.State == state {
			libraries = append(libraries, id)
		}
	}
	
	return libraries
}

// GetOldestLibrary 获取最久未访问的库
func (dl *DynamicLibraryLoader) GetOldestLibrary() (string, *LoadedLibrary) {
	dl.mutex.RLock()
	defer dl.mutex.RUnlock()
	
	var oldestID string
	var oldestLib *LoadedLibrary
	var oldestTime time.Time
	
	for id, lib := range dl.loadedLibs {
		if oldestLib == nil || lib.LastAccess.Before(oldestTime) {
			oldestID = id
			oldestLib = lib
			oldestTime = lib.LastAccess
		}
	}
	
	return oldestID, oldestLib
}

// ForceUnloadOldest 强制卸载最久未访问的库
func (dl *DynamicLibraryLoader) ForceUnloadOldest() error {
	oldestID, _ := dl.GetOldestLibrary()
	if oldestID == "" {
		return fmt.Errorf("no libraries to unload")
	}
	
	return dl.UnloadPlugin(dl.ctx, oldestID)
}