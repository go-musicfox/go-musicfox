package loader

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log/slog"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)



// RPC插件加载器
type RPCPluginLoader struct {
	processes map[string]*PluginProcess
	mutex     sync.RWMutex
	logger    *slog.Logger
	config    *RPCLoaderConfig
}

// RPC加载器配置
type RPCLoaderConfig struct {
	Network           string        `json:"network"`            // "tcp", "unix"
	Address           string        `json:"address"`            // 监听地址
	Timeout           time.Duration `json:"timeout"`            // 连接超时
	HeartbeatInterval time.Duration `json:"heartbeat_interval"` // 心跳间隔
	MaxRetries        int           `json:"max_retries"`        // 最大重试次数
}

// 插件进程信息
type PluginProcess struct {
	ID        string
	Path      string
	Process   *os.Process
	Client    *rpc.Client
	Config    *PluginProcessConfig
	State     ProcessState
	StartTime time.Time
	LastPing  time.Time
	PingCount int64
	cancel    context.CancelFunc
}

// 插件进程配置
type PluginProcessConfig struct {
	Executable string            `json:"executable"`
	Args       []string          `json:"args"`
	Env        map[string]string `json:"env"`
	WorkDir    string            `json:"work_dir"`
	Network    string            `json:"network"`
	Address    string            `json:"address"`
}

// 进程状态
type ProcessState int

const (
	ProcessStateStarting ProcessState = iota
	ProcessStateRunning
	ProcessStateStopping
	ProcessStateStopped
	ProcessStateError
)

// String returns the string representation of ProcessState
func (s ProcessState) String() string {
	switch s {
	case ProcessStateStarting:
		return "starting"
	case ProcessStateRunning:
		return "running"
	case ProcessStateStopping:
		return "stopping"
	case ProcessStateStopped:
		return "stopped"
	case ProcessStateError:
		return "error"
	default:
		return "unknown"
	}
}

// NewRPCPluginLoader 创建新的RPC插件加载器
func NewRPCPluginLoader(config *RPCLoaderConfig, logger *slog.Logger) *RPCPluginLoader {
	if config == nil {
		config = &RPCLoaderConfig{
			Network:           "tcp",
			Address:           "127.0.0.1:0", // 使用随机端口
			Timeout:           30 * time.Second,
			HeartbeatInterval: 10 * time.Second,
			MaxRetries:        3,
		}
	}
	
	return &RPCPluginLoader{
		processes: make(map[string]*PluginProcess),
		logger:    logger,
		config:    config,
	}
}

// NewRPCPluginLoaderWithSecurity 创建新的RPC插件加载器（适配manager.go的签名）
func NewRPCPluginLoaderWithSecurity(securityManager SecurityManager, logger *slog.Logger) *RPCPluginLoader {
	config := &RPCLoaderConfig{
		Network:           "tcp",
		Address:           "127.0.0.1:0", // 使用随机端口
		Timeout:           30 * time.Second,
		HeartbeatInterval: 10 * time.Second,
		MaxRetries:        3,
	}
	
	return &RPCPluginLoader{
		processes: make(map[string]*PluginProcess),
		logger:    logger,
		config:    config,
	}
}

// LoadPlugin 加载RPC插件（实现PluginLoader接口）
func (l *RPCPluginLoader) LoadPlugin(ctx context.Context, pluginPath string) (Plugin, error) {
	return l.loadPluginInternal(pluginPath)
}

// loadPluginInternal 内部加载RPC插件方法
func (l *RPCPluginLoader) loadPluginInternal(pluginPath string) (Plugin, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	// 检查是否已加载
	if process, exists := l.processes[pluginPath]; exists {
		if process.State == ProcessStateRunning {
			return l.createRPCPluginWrapper(process), nil
		}
	}
	
	// 启动插件进程
	process, err := l.startPluginProcess(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to start plugin process: %w", err)
	}
	
	// 建立RPC连接
	if err := l.establishRPCConnection(process); err != nil {
		l.stopPluginProcess(process)
		return nil, fmt.Errorf("failed to establish RPC connection: %w", err)
	}
	
	// 验证插件接口
	if err := l.validateRPCPluginInterface(process); err != nil {
		l.stopPluginProcess(process)
		return nil, fmt.Errorf("RPC plugin interface validation failed: %w", err)
	}
	
	process.State = ProcessStateRunning
	process.StartTime = time.Now()
	l.processes[pluginPath] = process
	
	// 启动心跳监控
	go l.monitorPluginProcess(process)
	
	l.logger.Info("RPC plugin loaded successfully", "path", pluginPath, "pid", process.Process.Pid)
	
	return l.createRPCPluginWrapper(process), nil
}



// startPluginProcess 启动插件进程
func (l *RPCPluginLoader) startPluginProcess(pluginPath string) (*PluginProcess, error) {
	// 读取插件配置
	config, err := l.loadPluginProcessConfig(pluginPath)
	if err != nil {
		return nil, err
	}
	
	// 准备环境变量
	env := os.Environ()
	for key, value := range config.Env {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	
	// 添加RPC通信配置
	env = append(env, fmt.Sprintf("PLUGIN_RPC_NETWORK=%s", config.Network))
	env = append(env, fmt.Sprintf("PLUGIN_RPC_ADDRESS=%s", config.Address))
	
	// 启动进程
	cmd := exec.Command(config.Executable, config.Args...)
	cmd.Env = env
	cmd.Dir = config.WorkDir
	
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start plugin process: %w", err)
	}
	
	_, cancel := context.WithCancel(context.Background())
	
	process := &PluginProcess{
		ID:      fmt.Sprintf("%s-%d", filepath.Base(pluginPath), cmd.Process.Pid),
		Path:    pluginPath,
		Process: cmd.Process,
		Config:  config,
		State:   ProcessStateStarting,
		cancel:  cancel,
	}
	
	// 监控进程退出
	go func() {
		cmd.Wait()
		l.mutex.Lock()
		defer l.mutex.Unlock()
		if process.State != ProcessStateStopping {
			process.State = ProcessStateError
			l.logger.Error("Plugin process exited unexpectedly", "path", pluginPath, "pid", process.Process.Pid)
		}
	}()
	
	return process, nil
}

// establishRPCConnection 建立RPC连接
func (l *RPCPluginLoader) establishRPCConnection(process *PluginProcess) error {
	// 等待插件进程启动RPC服务
	var conn net.Conn
	var err error
	
	for i := 0; i < l.config.MaxRetries; i++ {
		conn, err = net.DialTimeout(process.Config.Network, process.Config.Address, l.config.Timeout)
		if err == nil {
			break
		}
		
		l.logger.Debug("Waiting for plugin RPC service", "attempt", i+1, "error", err)
		time.Sleep(time.Second)
	}
	
	if err != nil {
		return fmt.Errorf("failed to connect to plugin RPC service: %w", err)
	}
	
	process.Client = rpc.NewClient(conn)
	return nil
}

// validateRPCPluginInterface 验证RPC插件接口
func (l *RPCPluginLoader) validateRPCPluginInterface(process *PluginProcess) error {
	// 测试基本RPC方法调用
	var reply struct{}
	err := process.Client.Call("PluginRPCService.HealthCheck", &struct{}{}, &reply)
	if err != nil {
		return fmt.Errorf("RPC health check failed: %w", err)
	}
	
	l.logger.Debug("RPC plugin interface validation successful", "path", process.Path)
	return nil
}

// stopPluginProcess 停止插件进程
func (l *RPCPluginLoader) stopPluginProcess(process *PluginProcess) error {
	if process == nil {
		return fmt.Errorf("process is nil")
	}
	
	process.State = ProcessStateStopping
	
	// 取消上下文
	if process.cancel != nil {
		process.cancel()
	}
	
	// 关闭RPC连接
	if process.Client != nil {
		process.Client.Close()
		process.Client = nil
	}
	
	// 终止进程
	if process.Process != nil {
		pid := process.Process.Pid
		if err := process.Process.Kill(); err != nil {
			l.logger.Warn("Failed to kill plugin process", "error", err, "pid", pid)
		}
		
		// 等待进程退出
		process.Process.Wait()
		l.logger.Info("Plugin process stopped", "path", process.Path, "pid", pid)
	} else {
		l.logger.Info("Plugin process stopped (no process handle)", "path", process.Path)
	}
	
	process.State = ProcessStateStopped
	return nil
}

// loadPluginProcessConfig 加载插件进程配置
func (l *RPCPluginLoader) loadPluginProcessConfig(pluginPath string) (*PluginProcessConfig, error) {
	// 查找配置文件
	configPath := pluginPath + ".config.json"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 使用默认配置
		return &PluginProcessConfig{
			Executable: pluginPath,
			Args:       []string{},
			Env:        make(map[string]string),
			WorkDir:    filepath.Dir(pluginPath),
			Network:    l.config.Network,
			Address:    l.config.Address,
		}, nil
	}
	
	// 读取配置文件
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin config: %w", err)
	}
	
	var config PluginProcessConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse plugin config: %w", err)
	}
	
	// 设置默认值
	if config.Network == "" {
		config.Network = l.config.Network
	}
	if config.Address == "" {
		config.Address = l.config.Address
	}
	if config.WorkDir == "" {
		config.WorkDir = filepath.Dir(pluginPath)
	}
	if config.Env == nil {
		config.Env = make(map[string]string)
	}
	
	return &config, nil
}

// monitorPluginProcess 监控插件进程
func (l *RPCPluginLoader) monitorPluginProcess(process *PluginProcess) {
	ticker := time.NewTicker(l.config.HeartbeatInterval)
	defer ticker.Stop()
	
	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	process.cancel = cancel
	defer cancel()
	
	for {
		select {
		case <-ctx.Done():
			l.logger.Info("Plugin monitoring stopped", "path", process.Path)
			return
		case <-ticker.C:
			// 检查进程状态
			if process.State != ProcessStateRunning {
				l.logger.Info("Plugin process not running, stopping monitor", "path", process.Path, "state", process.State)
				return
			}
			
			if err := l.pingPlugin(process); err != nil {
				l.logger.Error("Plugin ping failed", "path", process.Path, "error", err)
				process.State = ProcessStateError
				return
			}
			process.LastPing = time.Now()
			process.PingCount++
		}
	}
}

// pingPlugin 向插件发送心跳
func (l *RPCPluginLoader) pingPlugin(process *PluginProcess) error {
	if process.Client == nil {
		return fmt.Errorf("RPC client not available")
	}
	
	var reply struct{}
	err := process.Client.Call("PluginRPCService.HealthCheck", &struct{}{}, &reply)
	if err != nil {
		return fmt.Errorf("RPC ping failed: %w", err)
	}
	
	return nil
}

// createRPCPluginWrapper 创建RPC插件包装器
func (l *RPCPluginLoader) createRPCPluginWrapper(process *PluginProcess) Plugin {
	return &RPCPluginWrapper{
		process: process,
		logger:  l.logger,
	}
}

// GetProcesses 获取所有插件进程信息
func (l *RPCPluginLoader) GetProcesses() map[string]*PluginProcess {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	
	processes := make(map[string]*PluginProcess)
	for path, process := range l.processes {
		processes[path] = process
	}
	return processes
}

// GetProcessByPath 根据路径获取插件进程
func (l *RPCPluginLoader) GetProcessByPath(pluginPath string) (*PluginProcess, bool) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	
	process, exists := l.processes[pluginPath]
	return process, exists
}



// IsPluginLoaded 检查插件是否已加载
func (l *RPCPluginLoader) IsPluginLoaded(pluginID string) bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	for _, process := range l.processes {
		if process.ID == pluginID && process.State == ProcessStateRunning {
			return true
		}
	}
	return false
}

// Close 关闭RPC插件加载器
func (l *RPCPluginLoader) Close() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// 停止所有插件进程
	for _, process := range l.processes {
		if err := l.stopPluginProcess(process); err != nil {
			l.logger.Error("Failed to stop plugin process during cleanup", "error", err, "path", process.Path)
		}
	}

	// 清空进程列表
	l.processes = make(map[string]*PluginProcess)

	l.logger.Info("RPC plugin loader closed")
	return nil
}

// Shutdown 关闭RPC插件加载器
func (l *RPCPluginLoader) Shutdown(ctx context.Context) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	var errors []error
	
	for path, process := range l.processes {
		if err := l.stopPluginProcess(process); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop process %s: %w", path, err))
		}
	}
	
	l.processes = make(map[string]*PluginProcess)
	
	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}
	
	l.logger.Info("RPC plugin loader shutdown completed")
	return nil
}

// UnloadPlugin 卸载指定的插件（实现PluginLoader接口）
func (l *RPCPluginLoader) UnloadPlugin(ctx context.Context, pluginID string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	// 查找对应的进程
	var targetProcess *PluginProcess
	var targetPath string
	
	for path, process := range l.processes {
		if process.ID == pluginID {
			targetProcess = process
			targetPath = path
			break
		}
	}
	
	if targetProcess == nil {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}
	
	// 停止插件进程
	if err := l.stopPluginProcess(targetProcess); err != nil {
		return fmt.Errorf("failed to stop plugin process: %w", err)
	}
	
	// 从进程列表中移除
	delete(l.processes, targetPath)
	
	l.logger.Info("RPC plugin unloaded successfully", "id", pluginID, "path", targetPath)
	return nil
}

// GetLoadedPlugins 获取已加载的插件列表（实现PluginLoader接口）
func (l *RPCPluginLoader) GetLoadedPlugins() map[string]Plugin {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	
	result := make(map[string]Plugin)
	for _, process := range l.processes {
		if process.State == ProcessStateRunning {
			result[process.ID] = l.createRPCPluginWrapper(process)
		}
	}
	return result
}

// GetPluginInfo 获取插件信息（实现PluginLoader接口）
func (l *RPCPluginLoader) GetPluginInfo(pluginID string) (*PluginInfo, error) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	
	for _, process := range l.processes {
		if process.ID == pluginID {
			wrapper := l.createRPCPluginWrapper(process)
			return wrapper.GetInfo(), nil
		}
	}
	
	return nil, fmt.Errorf("plugin not found: %s", pluginID)
}

// ValidatePlugin 验证插件是否有效（实现PluginLoader接口）
func (l *RPCPluginLoader) ValidatePlugin(pluginPath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin file does not exist: %s", pluginPath)
	}
	
	// 检查是否为可执行文件
	fileInfo, err := os.Stat(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	
	if fileInfo.Mode()&0111 == 0 {
		return fmt.Errorf("plugin file is not executable: %s", pluginPath)
	}
	
	return nil
}

// GetLoaderType 获取加载器类型（实现PluginLoader接口）
func (l *RPCPluginLoader) GetLoaderType() PluginType {
	return PluginTypeRPC
}

// GetLoaderInfo 获取加载器信息（实现PluginLoader接口）
func (l *RPCPluginLoader) GetLoaderInfo() map[string]interface{} {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	
	return map[string]interface{}{
		"type":               "rpc",
		"loaded_count":       len(l.processes),
		"network":            l.config.Network,
		"address":            l.config.Address,
		"timeout":            l.config.Timeout,
		"heartbeat_interval": l.config.HeartbeatInterval,
		"max_retries":        l.config.MaxRetries,
	}
}

// ReloadPlugin 重新加载插件（实现PluginLoader接口）
func (l *RPCPluginLoader) ReloadPlugin(ctx context.Context, pluginID string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	// 查找对应的进程
	var targetProcess *PluginProcess
	var targetPath string
	
	for path, process := range l.processes {
		if process.ID == pluginID {
			targetProcess = process
			targetPath = path
			break
		}
	}
	
	if targetProcess == nil {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}
	
	pluginPath := targetPath
	
	// 先停止
	if err := l.stopPluginProcess(targetProcess); err != nil {
		return fmt.Errorf("failed to stop plugin for reload: %w", err)
	}
	
	// 从进程列表中移除
	delete(l.processes, targetPath)
	
	// 重新加载
	_, err := l.loadPluginInternal(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to reload plugin: %w", err)
	}
	
	l.logger.Info("RPC plugin reloaded successfully", "id", pluginID)
	return nil
}

// Cleanup 清理资源
func (l *RPCPluginLoader) Cleanup() error {
	return l.Close()
}