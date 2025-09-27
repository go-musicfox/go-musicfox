// Package loader 测试动态库加载器
package loader

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// TestNewDynamicLibraryLoader 测试创建动态库加载器
func TestNewDynamicLibraryLoader(t *testing.T) {
	tests := []struct {
		name   string
		config *DynamicLoaderConfig
		want   bool // 是否期望成功创建
	}{
		{
			name:   "with nil config",
			config: nil,
			want:   true,
		},
		{
			name: "with valid config",
			config: &DynamicLoaderConfig{
				MaxPlugins:        10,
				LoadTimeout:       30 * time.Second,
				UnloadTimeout:     15 * time.Second,
				EnableSymbolCache: true,
				ValidateSignature: false,
			},
			want: true,
		},
		{
			name: "with zero max plugins",
			config: &DynamicLoaderConfig{
				MaxPlugins: 0,
			},
			want: true, // 应该使用默认值
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewDynamicLibraryLoader(context.Background(), tt.config)
			
			if (loader != nil) != tt.want {
				t.Errorf("NewDynamicLibraryLoader() = %v, want %v", loader != nil, tt.want)
				return
			}
			
			if loader != nil {
				// 验证默认配置
				if loader.config == nil {
					t.Error("config should not be nil")
				}
				
				if loader.loadedLibs == nil {
				t.Error("loadedLibs should not be nil")
			}
			}
		})
	}
}

// TestLoadPlugin 测试加载插件
func TestLoadPlugin(t *testing.T) {
	loader := NewDynamicLibraryLoader(context.Background(), nil)
	ctx := context.Background()
	
	tests := []struct {
		name     string
		path     string
		config   map[string]interface{}
		wantErr  bool
		errorMsg string
	}{
		{
			name:     "empty path",
			path:     "",
			config:   nil,
			wantErr:  true,
			errorMsg: "plugin path cannot be empty",
		},
		{
			name:     "invalid extension",
			path:     "/path/to/plugin.txt",
			config:   nil,
			wantErr:  true,
			errorMsg: "failed to load dynamic library",
		},
		{
			name:     "non-existent file",
			path:     "/path/to/nonexistent.so",
			config:   nil,
			wantErr:  true,
			errorMsg: "failed to load dynamic library",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loader.LoadPlugin(ctx, tt.path)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadPlugin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.wantErr && err != nil {
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("LoadPlugin() error = %v, want error containing %v", err, tt.errorMsg)
				}
			}
		})
	}
}

// TestUnloadPlugin 测试卸载插件
func TestUnloadPlugin(t *testing.T) {
	loader := NewDynamicLibraryLoader(context.Background(), nil)
	ctx := context.Background()
	
	tests := []struct {
		name     string
		pluginID string
		wantErr  bool
		errorMsg string
	}{
		{
			name:     "empty plugin ID",
			pluginID: "",
			wantErr:  true,
			errorMsg: "plugin '' not found",
		},
		{
			name:     "non-existent plugin",
			pluginID: "nonexistent",
			wantErr:  true,
			errorMsg: "plugin 'nonexistent' not found",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := loader.UnloadPlugin(ctx, tt.pluginID)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("UnloadPlugin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.wantErr && err != nil {
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("UnloadPlugin() error = %v, want error containing %v", err, tt.errorMsg)
				}
			}
		})
	}
}

// TestGetLoadedPlugins 测试获取已加载插件列表
func TestGetLoadedPlugins(t *testing.T) {
	loader := NewDynamicLibraryLoader(context.Background(), nil)
	
	// 初始状态应该为空
	plugins := loader.GetLoadedPlugins()
	if len(plugins) != 0 {
		t.Errorf("GetLoadedPlugins() = %v, want empty slice", plugins)
	}
}

// TestIsPluginLoaded 测试检查插件是否已加载
func TestIsPluginLoaded(t *testing.T) {
	loader := NewDynamicLibraryLoader(context.Background(), nil)
	
	tests := []struct {
		name     string
		pluginID string
		want     bool
	}{
		{
			name:     "empty plugin ID",
			pluginID: "",
			want:     false,
		},
		{
			name:     "non-existent plugin",
			pluginID: "nonexistent",
			want:     false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := loader.IsPluginLoaded(tt.pluginID); got != tt.want {
				t.Errorf("IsPluginLoaded() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetStats 测试获取统计信息
func TestGetStats(t *testing.T) {
	loader := NewDynamicLibraryLoader(context.Background(), nil)
	
	stats := loader.GetStats()
	if stats == nil {
		t.Error("GetStats() should not return nil")
		return
	}
	
	// 验证初始统计信息
	if stats.TotalLoaded != 0 {
		t.Errorf("TotalLoaded = %d, want 0", stats.TotalLoaded)
	}
	
	if stats.TotalUnloaded != 0 {
		t.Errorf("TotalUnloaded = %d, want 0", stats.TotalUnloaded)
	}
	
	if stats.LoadErrors != 0 {
		t.Errorf("LoadErrors = %d, want 0", stats.LoadErrors)
	}
	
	if stats.UnloadErrors != 0 {
		t.Errorf("UnloadErrors = %d, want 0", stats.UnloadErrors)
	}
}

// TestValidatePluginPath 测试插件路径验证
func TestValidatePluginPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "relative path",
			path:    "plugin.so",
			wantErr: false, // 相对路径应该被接受
		},
		{
			name:    "absolute path",
			path:    "/usr/lib/plugin.so",
			wantErr: false,
		},
		{
			name:    "invalid extension",
			path:    "/path/to/plugin.txt",
			wantErr: false, // validatePluginPath不检查扩展名
		},
		{
			name:    "valid .so extension",
			path:    "/path/to/plugin.so",
			wantErr: false,
		},
		{
			name:    "valid .dylib extension",
			path:    "/path/to/plugin.dylib",
			wantErr: false,
		},
		{
			name:    "valid .dll extension",
			path:    "C:\\path\\to\\plugin.dll",
			wantErr: false,
		},
	}
	
	loader := NewDynamicLibraryLoader(context.Background(), nil)
		for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := loader.validatePluginPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePluginPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestIsValidPluginExtension 测试插件扩展名验证
func TestIsValidPluginExtension(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "empty path",
			path: "",
			want: false,
		},
		{
			name: "no extension",
			path: "plugin",
			want: false,
		},
		{
			name: ".so extension",
			path: "plugin.so",
			want: true,
		},
		{
			name: ".dylib extension",
			path: "plugin.dylib",
			want: runtime.GOOS == "darwin", // 只在macOS上支持.dylib
		},
		{
			name: ".dll extension",
			path: "plugin.dll",
			want: runtime.GOOS == "windows", // 只在Windows上支持.dll
		},
		{
			name: "invalid extension",
			path: "plugin.txt",
			want: false,
		},
		{
			name: "case sensitive",
			path: "plugin.SO",
			want: true, // 大写.SO会被转换为小写.so，在所有支持.so的平台上都被接受
		},
	}
	
	loader := NewDynamicLibraryLoader(context.Background(), nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := loader.isValidPluginExtension(tt.path); got != tt.want {
				t.Errorf("isValidPluginExtension() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGeneratePluginID 测试插件ID生成
func TestGeneratePluginID(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "simple path",
			path: "/path/to/plugin.so",
			want: "plugin",
		},
		{
			name: "complex path",
			path: "/usr/local/lib/plugins/audio_processor.so",
			want: "audio_processor",
		},
		{
			name: "windows path",
			path: "C:\\plugins\\video_codec.dll",
			want: "video_codec",
		},
		{
			name: "relative path",
			path: "./plugins/network.dylib",
			want: "network",
		},
	}
	
	loader := NewDynamicLibraryLoader(context.Background(), nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := loader.generatePluginID(tt.path); got != tt.want {
				t.Errorf("generatePluginID() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestConcurrentOperations 测试并发操作
func TestConcurrentOperations(t *testing.T) {
	loader := NewDynamicLibraryLoader(context.Background(), nil)
	
	// 测试并发获取统计信息
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			stats := loader.GetStats()
			if stats == nil {
				t.Error("GetStats() should not return nil")
			}
			done <- true
		}()
	}
	
	// 等待所有协程完成
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// 成功
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}
}

// TestDefaultConfig 测试默认配置
func TestDefaultConfig(t *testing.T) {
	config := DefaultDynamicLoaderConfig()
	
	if config == nil {
		t.Fatal("DefaultDynamicLoaderConfig() should not return nil")
	}
	
	// 验证默认值
	if config.MaxPlugins <= 0 {
		t.Errorf("MaxPlugins = %d, want > 0", config.MaxPlugins)
	}
	
	if config.LoadTimeout <= 0 {
		t.Errorf("LoadTimeout = %v, want > 0", config.LoadTimeout)
	}
	
	if config.UnloadTimeout <= 0 {
		t.Errorf("UnloadTimeout = %v, want > 0", config.UnloadTimeout)
	}
}

// createTempPlugin 创建临时插件文件用于测试
func createTempPlugin(t testing.TB, content string) string {
	tempDir := t.TempDir()
	pluginPath := filepath.Join(tempDir, "test_plugin.so")
	
	if err := os.WriteFile(pluginPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp plugin: %v", err)
	}
	
	return pluginPath
}

// BenchmarkLoadPlugin 基准测试插件加载
func BenchmarkLoadPlugin(b *testing.B) {
	loader := NewDynamicLibraryLoader(context.Background(), nil)
	ctx := context.Background()
	
	// 创建临时插件文件
	tempPlugin := createTempPlugin(b, "dummy plugin content")
	defer os.Remove(tempPlugin)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 注意：这里会失败，因为不是真正的插件文件
		// 但可以测试路径验证等前期处理的性能
		_, _ = loader.LoadPlugin(ctx, tempPlugin)
	}
}

// BenchmarkGetStats 基准测试获取统计信息
func BenchmarkGetStats(b *testing.B) {
	loader := NewDynamicLibraryLoader(context.Background(), nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = loader.GetStats()
	}
}