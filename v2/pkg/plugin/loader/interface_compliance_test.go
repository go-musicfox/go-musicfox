package loader

import (
	"context"
	"log/slog"
	"testing"
)

// TestWASMPluginLoaderInterfaceCompliance 测试WASM插件加载器是否符合PluginLoader接口
func TestWASMPluginLoaderInterfaceCompliance(t *testing.T) {
	// 创建WASM插件加载器实例
	mockSecurityMgr := &MockSecurityManager{}
	loader := NewWASMPluginLoader(mockSecurityMgr, slog.Default())

	// 验证WASM加载器实现了PluginLoader接口
	var _ PluginLoader = loader

	// 验证加载器创建成功
	if loader == nil {
		t.Fatal("WASM plugin loader should not be nil")
	}

	// 验证加载器类型
	loaderType := loader.GetLoaderType()
	if loaderType != PluginTypeWebAssembly {
		t.Errorf("Expected loader type %v, got %v", PluginTypeWebAssembly, loaderType)
	}

	// 验证初始状态
	loadedPlugins := loader.GetLoadedPlugins()
	if len(loadedPlugins) != 0 {
		t.Errorf("Expected 0 loaded plugins initially, got %d", len(loadedPlugins))
	}

	// 验证插件未加载状态
	if loader.IsPluginLoaded("non-existent-plugin") {
		t.Error("Non-existent plugin should not be loaded")
	}

	// 验证获取不存在插件的信息会返回错误
	_, err := loader.GetPluginInfo("non-existent-plugin")
	if err == nil {
		t.Error("Getting info for non-existent plugin should return error")
	}

	// 验证清理功能
	err = loader.Cleanup()
	if err != nil {
		t.Errorf("Cleanup should not return error: %v", err)
	}
}

// TestHotReloadPluginLoaderInterfaceCompliance 测试热加载插件加载器是否符合PluginLoader接口
func TestHotReloadPluginLoaderInterfaceCompliance(t *testing.T) {
	// 创建热加载插件加载器实例
	mockEventBus := &MockEventBus{}
	loader := NewHotReloadPluginLoader(mockEventBus, slog.Default())
	defer loader.Cleanup()

	// 验证热加载加载器实现了PluginLoader接口
	var _ PluginLoader = loader

	// 验证加载器创建成功
	if loader == nil {
		t.Fatal("Hot reload plugin loader should not be nil")
	}

	// 验证加载器类型
	loaderType := loader.GetLoaderType()
	if loaderType != PluginTypeHotReload {
		t.Errorf("Expected loader type %v, got %v", PluginTypeHotReload, loaderType)
	}

	// 验证初始状态
	loadedPlugins := loader.GetLoadedPlugins()
	if len(loadedPlugins) != 0 {
		t.Errorf("Expected 0 loaded plugins initially, got %d", len(loadedPlugins))
	}

	// 验证插件未加载状态
	if loader.IsPluginLoaded("non-existent-plugin") {
		t.Error("Non-existent plugin should not be loaded")
	}

	// 验证获取不存在插件的信息会返回错误
	_, err := loader.GetPluginInfo("non-existent-plugin")
	if err == nil {
		t.Error("Getting info for non-existent plugin should return error")
	}

	// 验证清理功能
	err = loader.Cleanup()
	if err != nil {
		t.Errorf("Cleanup should not return error: %v", err)
	}
}

// TestPluginLoaderInterfaceDefinition 测试PluginLoader接口定义的完整性
func TestPluginLoaderInterfaceDefinition(t *testing.T) {
	// 验证PluginLoader接口包含所有必需的方法
	// 这个测试通过编译时检查来验证接口的完整性

	// 创建一个匿名函数来测试接口方法签名
	testInterface := func(loader PluginLoader) {
		// 这些方法调用不会执行，只是用于编译时检查
		if false {
			ctx := context.Background()
			_, _ = loader.LoadPlugin(ctx, "")
			_ = loader.UnloadPlugin(ctx, "")
			_ = loader.GetLoadedPlugins()
			_ = loader.IsPluginLoaded("")
			_, _ = loader.GetPluginInfo("")
			_ = loader.ReloadPlugin(ctx, "")
			_ = loader.ValidatePlugin("")
			_ = loader.GetLoaderType()
			_ = loader.Cleanup()
		}
	}

	// 验证WASM加载器符合接口
	mockSecurityMgr := &MockSecurityManager{}
	wasmLoader := NewWASMPluginLoader(mockSecurityMgr, slog.Default())
	testInterface(wasmLoader)

	// 验证热加载加载器符合接口
	mockEventBus := &MockEventBus{}
	hotReloadLoader := NewHotReloadPluginLoader(mockEventBus, slog.Default())
	defer hotReloadLoader.Cleanup()
	testInterface(hotReloadLoader)

	t.Log("PluginLoader interface compliance verified")
}