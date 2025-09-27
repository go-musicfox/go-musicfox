// Package plugin 实现符号解析和函数调用功能
package plugin

import (
	"fmt"
	"reflect"
	"sync"
	"unsafe"

	"github.com/go-musicfox/go-musicfox/v2/pkg/plugin/loader"
)

// 类型别名
type SymbolType = loader.SymbolType

// 常量别名
const (
	SymbolTypeFunction = loader.SymbolTypeFunction
	SymbolTypeVariable = loader.SymbolTypeVariable
	SymbolTypeConstant = loader.SymbolTypeConstant
)

// SymbolResolver 符号解析器接口
type SymbolResolver interface {
	// ResolveSymbol 解析符号
	ResolveSymbol(pluginID, symbolName string) (interface{}, error)
	
	// ResolveFunction 解析函数符号
	ResolveFunction(pluginID, functionName string) (interface{}, error)
	
	// ResolveVariable 解析变量符号
	ResolveVariable(pluginID, variableName string) (interface{}, error)
	
	// CallFunction 调用函数
	CallFunction(pluginID, functionName string, args ...interface{}) ([]interface{}, error)
	
	// GetSymbolInfo 获取符号信息
	GetSymbolInfo(pluginID, symbolName string) (*loader.SymbolInfo, error)
	
	// ListSymbols 列出所有符号
	ListSymbols(pluginID string) ([]string, error)
	
	// CacheSymbol 缓存符号
	CacheSymbol(pluginID, symbolName string, symbol interface{}) error
	
	// ClearCache 清理符号缓存
	ClearCache(pluginID string) error
}

// 类型别名
type SymbolInfo = loader.SymbolInfo

// DynamicSymbolResolver 动态符号解析器实现
type DynamicSymbolResolver struct {
	// loader 关联的动态库加载器
	loader *loader.DynamicLibraryLoader
	
	// symbolCache 符号缓存
	symbolCache map[string]map[string]interface{}
	
	// mutex 读写锁
	mutex sync.RWMutex
	
	// enableCache 是否启用缓存
	enableCache bool
}

// NewDynamicSymbolResolver 创建新的动态符号解析器
func NewDynamicSymbolResolver(loader *loader.DynamicLibraryLoader, enableCache bool) *DynamicSymbolResolver {
	return &DynamicSymbolResolver{
		loader:      loader,
		symbolCache: make(map[string]map[string]interface{}),
		enableCache: enableCache,
	}
}

// ResolveSymbol 解析符号
func (dsr *DynamicSymbolResolver) ResolveSymbol(pluginID, symbolName string) (unsafe.Pointer, error) {
	if pluginID == "" {
		return nil, fmt.Errorf("library ID cannot be empty")
	}
	if symbolName == "" {
		return nil, fmt.Errorf("symbol name cannot be empty")
	}

	// 检查loader是否初始化
	if dsr.loader == nil {
		return nil, fmt.Errorf("loader is not initialized")
	}

	// 通过公共方法获取插件
	plugins := dsr.loader.GetLoadedPlugins()
	plugin, exists := plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	if plugin == nil {
		return nil, fmt.Errorf("plugin instance is nil for: %s", pluginID)
	}

	// 简化实现，返回nil指针
	// 实际实现需要根据具体的符号查找逻辑
	return nil, fmt.Errorf("symbol resolution not implemented: %s in plugin %s", symbolName, pluginID)
}

// ResolveFunction 解析函数符号
func (dsr *DynamicSymbolResolver) ResolveFunction(pluginID, functionName string) (interface{}, error) {
	// 参数验证
	if pluginID == "" {
		return nil, fmt.Errorf("library ID cannot be empty")
	}
	if functionName == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}
	
	symbol, err := dsr.ResolveSymbol(pluginID, functionName)
	if err != nil {
		return nil, err
	}
	
	// 检查是否为函数类型
	symbolValue := reflect.ValueOf(symbol)
	if symbolValue.Kind() != reflect.Func {
		return nil, fmt.Errorf("symbol '%s' is not a function", functionName)
	}
	
	return symbol, nil
}

// ResolveVariable 解析变量符号
func (dsr *DynamicSymbolResolver) ResolveVariable(pluginID, variableName string) (interface{}, error) {
	symbol, err := dsr.ResolveSymbol(pluginID, variableName)
	if err != nil {
		return nil, err
	}
	
	// 检查是否为变量类型（非函数）
	symbolValue := reflect.ValueOf(symbol)
	if symbolValue.Kind() == reflect.Func {
		return nil, fmt.Errorf("symbol '%s' is a function, not a variable", variableName)
	}
	
	return symbol, nil
}

// CallFunction 调用函数
func (dsr *DynamicSymbolResolver) CallFunction(pluginID, functionName string, args ...interface{}) ([]interface{}, error) {
	// 参数验证
	if pluginID == "" {
		return nil, fmt.Errorf("library ID cannot be empty")
	}
	if functionName == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}
	
	// 解析函数符号
	functionSymbol, err := dsr.ResolveFunction(pluginID, functionName)
	if err != nil {
		return nil, err
	}
	
	// 获取函数的反射值
	functionValue := reflect.ValueOf(functionSymbol)
	functionType := functionValue.Type()
	
	// 检查参数数量
	if len(args) != functionType.NumIn() {
		return nil, fmt.Errorf("function '%s' expects %d arguments, got %d",
			functionName, functionType.NumIn(), len(args))
	}
	
	// 准备参数
	inputArgs := make([]reflect.Value, len(args))
	for i, arg := range args {
		expectedType := functionType.In(i)
		argValue := reflect.ValueOf(arg)
		
		// 类型检查和转换
		if !argValue.Type().AssignableTo(expectedType) {
			// 尝试类型转换
			if argValue.Type().ConvertibleTo(expectedType) {
				argValue = argValue.Convert(expectedType)
			} else {
				return nil, fmt.Errorf("argument %d: cannot convert %s to %s",
					i, argValue.Type(), expectedType)
			}
		}
		
		inputArgs[i] = argValue
	}
	
	// 调用函数
	results := functionValue.Call(inputArgs)
	
	// 转换返回值
	outputResults := make([]interface{}, len(results))
	for i, result := range results {
		outputResults[i] = result.Interface()
	}
	
	// 注意：这里移除了UpdateLastAccess调用，因为loader接口可能不包含此方法

	return outputResults, nil
}

// GetSymbolInfo 获取符号信息
func (dsr *DynamicSymbolResolver) GetSymbolInfo(pluginID, symbolName string) (*loader.SymbolInfo, error) {
	symbol, err := dsr.ResolveSymbol(pluginID, symbolName)
	if err != nil {
		return nil, err
	}
	
	symbolValue := reflect.ValueOf(symbol)
	symbolType := symbolValue.Type()
	
	// 确定符号类型
	var symType SymbolType
	switch symbolValue.Kind() {
	case reflect.Func:
		symType = SymbolTypeFunction
	case reflect.Ptr:
		if symbolValue.Elem().CanSet() {
			symType = SymbolTypeVariable
		} else {
			symType = SymbolTypeConstant
		}
	default:
		symType = SymbolTypeVariable
	}
	
	return &loader.SymbolInfo{
		Name:     symbolName,
		Address:  unsafe.Pointer(uintptr(symbolValue.Pointer())),
		Type:     symType,
		Size:     symbolType.Size(),
		Exported: true, // Go plugin中的符号都是导出的
	}, nil
}

// ListSymbols 列出所有符号
func (dsr *DynamicSymbolResolver) ListSymbols(pluginID string) ([]string, error) {
	// 检查loader是否初始化
	if dsr.loader == nil {
		return nil, fmt.Errorf("loader is not initialized")
	}

	// 通过公共方法检查插件是否已加载
	plugins := dsr.loader.GetLoadedPlugins()
	_, exists := plugins[pluginID]
	
	if !exists {
		return nil, fmt.Errorf("plugin '%s' not loaded", pluginID)
	}
	
	// 从缓存中获取符号列表
	dsr.mutex.RLock()
	cachedSymbols, hasCached := dsr.symbolCache[pluginID]
	dsr.mutex.RUnlock()
	
	if hasCached {
		symbols := make([]string, 0, len(cachedSymbols))
		for symbolName := range cachedSymbols {
			symbols = append(symbols, symbolName)
		}
		return symbols, nil
	}
	
	// 返回空列表
	return []string{}, nil
}

// CacheSymbol 缓存符号
func (dsr *DynamicSymbolResolver) CacheSymbol(pluginID, symbolName string, symbol interface{}) error {
	if !dsr.enableCache {
		return nil
	}
	
	dsr.cacheSymbol(pluginID, symbolName, symbol)
	return nil
}

// ClearCache 清理符号缓存
func (dsr *DynamicSymbolResolver) ClearCache(pluginID string) error {
	dsr.mutex.Lock()
	defer dsr.mutex.Unlock()
	
	if pluginID == "" {
		// 清理所有缓存
		dsr.symbolCache = make(map[string]map[string]interface{})
	} else {
		// 清理指定插件的缓存
		delete(dsr.symbolCache, pluginID)
	}
	
	return nil
}

// getCachedSymbol 获取缓存的符号
func (dsr *DynamicSymbolResolver) getCachedSymbol(pluginID, symbolName string) interface{} {
	dsr.mutex.RLock()
	defer dsr.mutex.RUnlock()
	
	pluginCache, exists := dsr.symbolCache[pluginID]
	if !exists {
		return nil
	}
	
	symbol, exists := pluginCache[symbolName]
	if !exists {
		return nil
	}
	
	return symbol
}

// cacheSymbol 缓存符号
func (dsr *DynamicSymbolResolver) cacheSymbol(pluginID, symbolName string, symbol interface{}) {
	dsr.mutex.Lock()
	defer dsr.mutex.Unlock()
	
	pluginCache, exists := dsr.symbolCache[pluginID]
	if !exists {
		pluginCache = make(map[string]interface{})
		dsr.symbolCache[pluginID] = pluginCache
	}
	
	pluginCache[symbolName] = symbol
}

// GetCacheStats 获取缓存统计信息
func (dsr *DynamicSymbolResolver) GetCacheStats() map[string]int {
	dsr.mutex.RLock()
	defer dsr.mutex.RUnlock()
	
	stats := make(map[string]int)
	for pluginID, pluginCache := range dsr.symbolCache {
		stats[pluginID] = len(pluginCache)
	}
	
	return stats
}

// ValidateSymbolSignature 验证符号签名
func (dsr *DynamicSymbolResolver) ValidateSymbolSignature(pluginID, symbolName string, expectedSignature reflect.Type) error {
	symbol, err := dsr.ResolveSymbol(pluginID, symbolName)
	if err != nil {
		return err
	}
	
	actualType := reflect.TypeOf(symbol)
	if actualType != expectedSignature {
		return fmt.Errorf("symbol '%s' signature mismatch: expected %s, got %s",
			symbolName, expectedSignature, actualType)
	}
	
	return nil
}

// CallFunctionSafe 安全调用函数（带错误恢复）
func (dsr *DynamicSymbolResolver) CallFunctionSafe(pluginID, functionName string, args ...interface{}) (results []interface{}, err error) {
	// 使用defer和recover来捕获panic
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("function call panicked: %v", r)
			results = nil
		}
	}()
	
	return dsr.CallFunction(pluginID, functionName, args...)
}