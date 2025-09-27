package loader

import (
	"fmt"
	"log/slog"
	"net/rpc"
	"time"
)

// RPCPluginWrapper RPC插件包装器，实现Plugin接口
type RPCPluginWrapper struct {
	process *PluginProcess
	logger  *slog.Logger
	info    *PluginInfo
}



// GetInfo 获取插件信息
func (w *RPCPluginWrapper) GetInfo() *PluginInfo {
	if w.info != nil {
		return w.info
	}
	
	// 通过RPC调用获取插件信息
	var reply PluginInfo
	err := w.process.Client.Call("PluginRPCService.GetInfo", &struct{}{}, &reply)
	if err != nil {
		w.logger.Error("Failed to get plugin info via RPC", "error", err)
		// 返回默认信息
		return &PluginInfo{
			Name:        fmt.Sprintf("RPC Plugin %s", w.process.ID),
			Version:     "unknown",
			Description: "RPC Plugin",
			Author:      "unknown",
			LoadTime:    w.process.StartTime,
		}
	}
	
	w.info = &reply
	return w.info
}

// GetCapabilities 获取插件能力
func (w *RPCPluginWrapper) GetCapabilities() []string {
	var reply []string
	err := w.process.Client.Call("PluginRPCService.GetCapabilities", &struct{}{}, &reply)
	if err != nil {
		w.logger.Error("Failed to get plugin capabilities via RPC", "error", err)
		return []string{}
	}
	return reply
}

// GetDependencies 获取插件依赖
func (w *RPCPluginWrapper) GetDependencies() []string {
	var reply []string
	err := w.process.Client.Call("PluginRPCService.GetDependencies", &struct{}{}, &reply)
	if err != nil {
		w.logger.Error("Failed to get plugin dependencies via RPC", "error", err)
		return []string{}
	}
	return reply
}

// Initialize 初始化插件
func (w *RPCPluginWrapper) Initialize(ctx PluginContext) error {
	var reply struct{}
	err := w.process.Client.Call("PluginRPCService.Initialize", &ctx, &reply)
	if err != nil {
		w.logger.Error("Failed to initialize plugin via RPC", "error", err)
		return fmt.Errorf("RPC initialize failed: %w", err)
	}
	return nil
}

// Start 启动插件
func (w *RPCPluginWrapper) Start() error {
	var reply struct{}
	err := w.process.Client.Call("PluginRPCService.Start", &struct{}{}, &reply)
	if err != nil {
		w.logger.Error("Failed to start plugin via RPC", "error", err)
		return fmt.Errorf("RPC start failed: %w", err)
	}
	return nil
}

// Stop 停止插件
func (w *RPCPluginWrapper) Stop() error {
	var reply struct{}
	err := w.process.Client.Call("PluginRPCService.Stop", &struct{}{}, &reply)
	if err != nil {
		w.logger.Error("Failed to stop plugin via RPC", "error", err)
		return fmt.Errorf("RPC stop failed: %w", err)
	}
	return nil
}

// Cleanup 清理插件
func (w *RPCPluginWrapper) Cleanup() error {
	var reply struct{}
	err := w.process.Client.Call("PluginRPCService.Cleanup", &struct{}{}, &reply)
	if err != nil {
		w.logger.Error("Failed to cleanup plugin via RPC", "error", err)
		return fmt.Errorf("RPC cleanup failed: %w", err)
	}
	return nil
}

// GetMetrics 获取插件指标
func (w *RPCPluginWrapper) GetMetrics() (*PluginMetrics, error) {
	var reply PluginMetrics
	err := w.process.Client.Call("PluginRPCService.GetMetrics", &struct{}{}, &reply)
	if err != nil {
		w.logger.Error("Failed to get plugin metrics via RPC", "error", err)
		return nil, fmt.Errorf("RPC get metrics failed: %w", err)
	}
	return &reply, nil
}

// HealthCheck 健康检查
func (w *RPCPluginWrapper) HealthCheck() error {
	var reply struct{}
	err := w.process.Client.Call("PluginRPCService.HealthCheck", &struct{}{}, &reply)
	if err != nil {
		w.logger.Error("Plugin health check failed via RPC", "error", err)
		return fmt.Errorf("RPC health check failed: %w", err)
	}
	return nil
}

// ValidateConfig 验证配置
func (w *RPCPluginWrapper) ValidateConfig(config map[string]interface{}) error {
	var reply struct{}
	err := w.process.Client.Call("PluginRPCService.ValidateConfig", &config, &reply)
	if err != nil {
		w.logger.Error("Failed to validate config via RPC", "error", err)
		return fmt.Errorf("RPC validate config failed: %w", err)
	}
	return nil
}

// UpdateConfig 更新配置
func (w *RPCPluginWrapper) UpdateConfig(config map[string]interface{}) error {
	var reply struct{}
	err := w.process.Client.Call("PluginRPCService.UpdateConfig", &config, &reply)
	if err != nil {
		w.logger.Error("Failed to update config via RPC", "error", err)
		return fmt.Errorf("RPC update config failed: %w", err)
	}
	return nil
}

// GetProcess 获取关联的插件进程
func (w *RPCPluginWrapper) GetProcess() *PluginProcess {
	return w.process
}

// GetProcessID 获取进程ID
func (w *RPCPluginWrapper) GetProcessID() string {
	return w.process.ID
}

// GetProcessState 获取进程状态
func (w *RPCPluginWrapper) GetProcessState() ProcessState {
	return w.process.State
}

// GetLastPing 获取最后心跳时间
func (w *RPCPluginWrapper) GetLastPing() time.Time {
	return w.process.LastPing
}

// GetPingCount 获取心跳次数
func (w *RPCPluginWrapper) GetPingCount() int64 {
	return w.process.PingCount
}

// IsHealthy 检查插件是否健康
func (w *RPCPluginWrapper) IsHealthy() bool {
	if w.process.State != ProcessStateRunning {
		return false
	}
	
	// 检查最后心跳时间
	if time.Since(w.process.LastPing) > 30*time.Second {
		return false
	}
	
	return true
}

// CallRPCMethod 调用RPC方法
func (w *RPCPluginWrapper) CallRPCMethod(method string, args interface{}, reply interface{}) error {
	var client *rpc.Client = w.process.Client
	if client == nil {
		return fmt.Errorf("RPC client not available")
	}
	
	err := client.Call(method, args, reply)
	if err != nil {
		w.logger.Error("RPC method call failed", "method", method, "error", err)
		return fmt.Errorf("RPC call %s failed: %w", method, err)
	}
	
	return nil
}

// CallRPCMethodWithTimeout 带超时的RPC方法调用
func (w *RPCPluginWrapper) CallRPCMethodWithTimeout(method string, args interface{}, reply interface{}, timeout time.Duration) error {
	if w.process.Client == nil {
		return fmt.Errorf("RPC client not available")
	}
	
	done := make(chan error, 1)
	go func() {
		done <- w.process.Client.Call(method, args, reply)
	}()
	
	select {
	case err := <-done:
		if err != nil {
			w.logger.Error("RPC method call failed", "method", method, "error", err)
			return fmt.Errorf("RPC call %s failed: %w", method, err)
		}
		return nil
	case <-time.After(timeout):
		w.logger.Error("RPC method call timeout", "method", method, "timeout", timeout)
		return fmt.Errorf("RPC call %s timeout after %v", method, timeout)
	}
}