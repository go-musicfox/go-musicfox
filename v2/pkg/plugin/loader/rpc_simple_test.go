package loader

import (
	"log/slog"
	"os"
	"testing"
	"time"
)

// TestNewRPCPluginLoader_Simple 简单测试创建RPC插件加载器
func TestNewRPCPluginLoader_Simple(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	
	config := &RPCLoaderConfig{
		Network:           "tcp",
		Address:           "127.0.0.1:0",
		Timeout:           5 * time.Second,
		HeartbeatInterval: 1 * time.Second,
		MaxRetries:        1,
	}
	
	loader := NewRPCPluginLoader(config, logger)
	
	if loader == nil {
		t.Fatal("Expected non-nil loader")
	}
	
	if loader.config != config {
		t.Error("Config not set correctly")
	}
	
	if loader.logger != logger {
		t.Error("Logger not set correctly")
	}
	
	if loader.processes == nil {
		t.Error("Processes map not initialized")
	}
}

// TestRPCPluginLoader_GetProcesses 测试获取进程列表
func TestRPCPluginLoader_GetProcesses(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	
	config := &RPCLoaderConfig{
		Network:           "tcp",
		Address:           "127.0.0.1:0",
		Timeout:           5 * time.Second,
		HeartbeatInterval: 1 * time.Second,
		MaxRetries:        1,
	}
	
	loader := NewRPCPluginLoader(config, logger)
	
	// 测试空进程列表
	processes := loader.GetProcesses()
	if len(processes) != 0 {
		t.Errorf("Expected empty processes map, got %d", len(processes))
	}
	
	// 添加模拟进程
	process := &PluginProcess{
		ID:        "test-plugin",
		Path:      "/tmp/test_plugin",
		State:     ProcessStateRunning,
		StartTime: time.Now(),
		LastPing:  time.Now(),
		PingCount: 5,
		Config: &PluginProcessConfig{
			Network: "tcp",
			Address: "localhost:12345",
		},
	}
	
	loader.mutex.Lock()
	loader.processes["/tmp/test_plugin"] = process
	loader.mutex.Unlock()
	
	// 测试非空进程列表
	processes = loader.GetProcesses()
	if len(processes) != 1 {
		t.Errorf("Expected 1 process, got %d", len(processes))
	}
	
	if processes["/tmp/test_plugin"].ID != "test-plugin" {
		t.Errorf("Expected process ID 'test-plugin', got '%s'", processes["/tmp/test_plugin"].ID)
	}
}

// TestRPCPluginLoader_IsPluginLoaded_Simple 测试检查插件是否已加载
func TestRPCPluginLoader_IsPluginLoaded_Simple(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	
	config := &RPCLoaderConfig{
		Network:           "tcp",
		Address:           "127.0.0.1:0",
		Timeout:           5 * time.Second,
		HeartbeatInterval: 1 * time.Second,
		MaxRetries:        1,
	}
	
	loader := NewRPCPluginLoader(config, logger)
	
	pluginPath := "/tmp/test_plugin"
	
	// 测试未加载的插件
	if loader.IsPluginLoaded(pluginPath) {
		t.Error("Plugin should not be loaded")
	}
	
	// 添加模拟进程
	process := &PluginProcess{
		ID:        "test-plugin",
		Path:      pluginPath,
		State:     ProcessStateRunning,
		StartTime: time.Now(),
		LastPing:  time.Now(),
		PingCount: 5,
		Config: &PluginProcessConfig{
			Network: "tcp",
			Address: "localhost:12345",
		},
	}
	
	loader.mutex.Lock()
	loader.processes[pluginPath] = process
	loader.mutex.Unlock()
	
	// 测试已加载的插件
	if !loader.IsPluginLoaded("test-plugin") {
		t.Error("Plugin should be loaded")
	}
}

// TestRPCPluginWrapper_Basic 测试RPC插件包装器基本功能
func TestRPCPluginWrapper_Basic(t *testing.T) {
	process := &PluginProcess{
		ID:        "test-plugin",
		Path:      "/tmp/test_plugin",
		State:     ProcessStateRunning,
		StartTime: time.Now(),
		LastPing:  time.Now(),
		PingCount: 5,
		Config: &PluginProcessConfig{
			Network: "tcp",
			Address: "localhost:12345",
		},
	}
	
	wrapper := &RPCPluginWrapper{
		process: process,
	}
	
	// 测试基本属性
	if wrapper.process != process {
		t.Error("Process not set correctly")
	}
	
	// 测试基本属性访问
	if wrapper.process.ID != "test-plugin" {
		t.Errorf("Expected process ID 'test-plugin', got '%s'", wrapper.process.ID)
	}
	
	if wrapper.process.State != ProcessStateRunning {
		t.Errorf("Expected process state %v, got %v", ProcessStateRunning, wrapper.process.State)
	}
}