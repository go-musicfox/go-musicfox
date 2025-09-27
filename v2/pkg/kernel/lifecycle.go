package kernel

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// LifecycleState 生命周期状态
type LifecycleState int

const (
	LifecycleStateUninitialized LifecycleState = iota
	LifecycleStateInitializing
	LifecycleStateInitialized
	LifecycleStateStarting
	LifecycleStateRunning
	LifecycleStateStopping
	LifecycleStateStopped
	LifecycleStateShuttingDown
	LifecycleStateShutdown
	LifecycleStateError
)

// String returns the string representation of LifecycleState
func (ls LifecycleState) String() string {
	switch ls {
	case LifecycleStateUninitialized:
		return "uninitialized"
	case LifecycleStateInitializing:
		return "initializing"
	case LifecycleStateInitialized:
		return "initialized"
	case LifecycleStateStarting:
		return "starting"
	case LifecycleStateRunning:
		return "running"
	case LifecycleStateStopping:
		return "stopping"
	case LifecycleStateStopped:
		return "stopped"
	case LifecycleStateShuttingDown:
		return "shutting_down"
	case LifecycleStateShutdown:
		return "shutdown"
	case LifecycleStateError:
		return "error"
	default:
		return "unknown"
	}
}

// LifecyclePhase 生命周期阶段
type LifecyclePhase struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Timeout     time.Duration `json:"timeout"`
	Handler     func(ctx context.Context) error `json:"-"`
	Required    bool          `json:"required"`
	Order       int           `json:"order"`
}

// LifecycleManager 生命周期管理器接口
type LifecycleManager interface {
	// 生命周期控制
	Initialize(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Shutdown(ctx context.Context) error

	// 状态查询
	GetState() LifecycleState
	IsRunning() bool
	GetUptime() time.Duration

	// 阶段管理
	AddInitPhase(phase *LifecyclePhase) error
	AddStartPhase(phase *LifecyclePhase) error
	AddStopPhase(phase *LifecyclePhase) error
	AddShutdownPhase(phase *LifecyclePhase) error

	// 信号处理
	EnableSignalHandling() error
	DisableSignalHandling() error
	AddSignalHandler(sig os.Signal, handler func(os.Signal)) error

	// 事件回调
	OnStateChanged(callback func(oldState, newState LifecycleState))
	OnError(callback func(phase string, err error))
}

// KernelLifecycleManager 内核生命周期管理器实现
type KernelLifecycleManager struct {
	logger *slog.Logger

	// 状态管理
	state     LifecycleState
	startTime time.Time
	mutex     sync.RWMutex

	// 生命周期阶段
	initPhases     []*LifecyclePhase
	startPhases    []*LifecyclePhase
	stopPhases     []*LifecyclePhase
	shutdownPhases []*LifecyclePhase

	// 信号处理
	signalChan     chan os.Signal
	signalHandlers map[os.Signal]func(os.Signal)
	signalEnabled  bool

	// 事件回调
	stateChangeCallbacks []func(oldState, newState LifecycleState)
	errorCallbacks       []func(phase string, err error)

	// 上下文管理
	ctx    context.Context
	cancel context.CancelFunc
}

// NewLifecycleManager 创建新的生命周期管理器
func NewLifecycleManager(logger *slog.Logger) LifecycleManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &KernelLifecycleManager{
		logger:               logger,
		state:                LifecycleStateUninitialized,
		initPhases:           make([]*LifecyclePhase, 0),
		startPhases:          make([]*LifecyclePhase, 0),
		stopPhases:           make([]*LifecyclePhase, 0),
		shutdownPhases:       make([]*LifecyclePhase, 0),
		signalHandlers:       make(map[os.Signal]func(os.Signal)),
		stateChangeCallbacks: make([]func(oldState, newState LifecycleState), 0),
		errorCallbacks:       make([]func(phase string, err error), 0),
		ctx:                  ctx,
		cancel:               cancel,
	}
}

// Initialize 初始化生命周期管理器
func (lm *KernelLifecycleManager) Initialize(ctx context.Context) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	if lm.state != LifecycleStateUninitialized {
		return fmt.Errorf("lifecycle manager already initialized")
	}

	lm.setState(LifecycleStateInitializing)
	lm.logger.Info("Initializing lifecycle manager...")

	// 执行初始化阶段
	if err := lm.executePhases(ctx, lm.initPhases, "initialization"); err != nil {
		lm.setState(LifecycleStateError)
		return fmt.Errorf("initialization failed: %w", err)
	}

	lm.setState(LifecycleStateInitialized)
	lm.logger.Info("Lifecycle manager initialized successfully")
	return nil
}

// Start 启动生命周期管理器
func (lm *KernelLifecycleManager) Start(ctx context.Context) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	if lm.state != LifecycleStateInitialized {
		return fmt.Errorf("lifecycle manager not initialized")
	}

	lm.setState(LifecycleStateStarting)
	lm.logger.Info("Starting lifecycle manager...")

	// 执行启动阶段
	if err := lm.executePhases(ctx, lm.startPhases, "startup"); err != nil {
		lm.setState(LifecycleStateError)
		return fmt.Errorf("startup failed: %w", err)
	}

	// 启动信号处理（如果启用）
	if lm.signalEnabled {
		go lm.handleSignals()
	}

	lm.setState(LifecycleStateRunning)
	lm.startTime = time.Now()
	lm.logger.Info("Lifecycle manager started successfully")
	return nil
}

// Stop 停止生命周期管理器
func (lm *KernelLifecycleManager) Stop(ctx context.Context) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	if lm.state != LifecycleStateRunning {
		return fmt.Errorf("lifecycle manager not running")
	}

	lm.setState(LifecycleStateStopping)
	lm.logger.Info("Stopping lifecycle manager...")

	// 执行停止阶段（逆序执行）
	reversedStopPhases := make([]*LifecyclePhase, len(lm.stopPhases))
	for i, phase := range lm.stopPhases {
		reversedStopPhases[len(lm.stopPhases)-1-i] = phase
	}

	if err := lm.executePhases(ctx, reversedStopPhases, "shutdown"); err != nil {
		lm.setState(LifecycleStateError)
		return fmt.Errorf("shutdown failed: %w", err)
	}

	lm.setState(LifecycleStateStopped)
	lm.logger.Info("Lifecycle manager stopped successfully")
	return nil
}

// Shutdown 关闭生命周期管理器
func (lm *KernelLifecycleManager) Shutdown(ctx context.Context) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	lm.setState(LifecycleStateShuttingDown)
	lm.logger.Info("Shutting down lifecycle manager...")

	// 如果还在运行，先停止
	if lm.state == LifecycleStateRunning {
		lm.mutex.Unlock()
		if err := lm.Stop(ctx); err != nil {
			lm.logger.Error("Failed to stop during shutdown", "error", err)
		}
		lm.mutex.Lock()
	}

	// 执行关闭阶段（逆序执行）
	reversedShutdownPhases := make([]*LifecyclePhase, len(lm.shutdownPhases))
	for i, phase := range lm.shutdownPhases {
		reversedShutdownPhases[len(lm.shutdownPhases)-1-i] = phase
	}

	if err := lm.executePhases(ctx, reversedShutdownPhases, "shutdown"); err != nil {
		lm.logger.Error("Shutdown phase failed", "error", err)
	}

	// 停止信号处理
	if lm.signalChan != nil {
		signal.Stop(lm.signalChan)
		close(lm.signalChan)
		lm.signalChan = nil
	}

	// 取消上下文
	lm.cancel()

	lm.setState(LifecycleStateShutdown)
	lm.logger.Info("Lifecycle manager shutdown completed")
	return nil
}

// GetState 获取当前状态
func (lm *KernelLifecycleManager) GetState() LifecycleState {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()
	return lm.state
}

// IsRunning 检查是否正在运行
func (lm *KernelLifecycleManager) IsRunning() bool {
	return lm.GetState() == LifecycleStateRunning
}

// GetUptime 获取运行时间
func (lm *KernelLifecycleManager) GetUptime() time.Duration {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	if lm.state == LifecycleStateRunning && !lm.startTime.IsZero() {
		return time.Since(lm.startTime)
	}
	return 0
}

// AddInitPhase 添加初始化阶段
func (lm *KernelLifecycleManager) AddInitPhase(phase *LifecyclePhase) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	if err := lm.validatePhase(phase); err != nil {
		return err
	}

	lm.initPhases = append(lm.initPhases, phase)
	lm.sortPhases(lm.initPhases)
	return nil
}

// AddStartPhase 添加启动阶段
func (lm *KernelLifecycleManager) AddStartPhase(phase *LifecyclePhase) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	if err := lm.validatePhase(phase); err != nil {
		return err
	}

	lm.startPhases = append(lm.startPhases, phase)
	lm.sortPhases(lm.startPhases)
	return nil
}

// AddStopPhase 添加停止阶段
func (lm *KernelLifecycleManager) AddStopPhase(phase *LifecyclePhase) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	if err := lm.validatePhase(phase); err != nil {
		return err
	}

	lm.stopPhases = append(lm.stopPhases, phase)
	lm.sortPhases(lm.stopPhases)
	return nil
}

// AddShutdownPhase 添加关闭阶段
func (lm *KernelLifecycleManager) AddShutdownPhase(phase *LifecyclePhase) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	if err := lm.validatePhase(phase); err != nil {
		return err
	}

	lm.shutdownPhases = append(lm.shutdownPhases, phase)
	lm.sortPhases(lm.shutdownPhases)
	return nil
}

// EnableSignalHandling 启用信号处理
func (lm *KernelLifecycleManager) EnableSignalHandling() error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	if lm.signalEnabled {
		return nil
	}

	lm.signalChan = make(chan os.Signal, 1)
	lm.signalEnabled = true

	// 注册默认信号处理器
	lm.addDefaultSignalHandlers()

	lm.logger.Info("Signal handling enabled")
	return nil
}

// DisableSignalHandling 禁用信号处理
func (lm *KernelLifecycleManager) DisableSignalHandling() error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	if !lm.signalEnabled {
		return nil
	}

	if lm.signalChan != nil {
		signal.Stop(lm.signalChan)
		close(lm.signalChan)
		lm.signalChan = nil
	}

	lm.signalEnabled = false
	lm.logger.Info("Signal handling disabled")
	return nil
}

// AddSignalHandler 添加信号处理器
func (lm *KernelLifecycleManager) AddSignalHandler(sig os.Signal, handler func(os.Signal)) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	lm.signalHandlers[sig] = handler

	if lm.signalEnabled && lm.signalChan != nil {
		signal.Notify(lm.signalChan, sig)
	}

	return nil
}

// OnStateChanged 注册状态变更回调
func (lm *KernelLifecycleManager) OnStateChanged(callback func(oldState, newState LifecycleState)) {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	lm.stateChangeCallbacks = append(lm.stateChangeCallbacks, callback)
}

// OnError 注册错误回调
func (lm *KernelLifecycleManager) OnError(callback func(phase string, err error)) {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	lm.errorCallbacks = append(lm.errorCallbacks, callback)
}

// setState 设置状态并触发回调
func (lm *KernelLifecycleManager) setState(newState LifecycleState) {
	oldState := lm.state
	lm.state = newState

	// 触发状态变更回调
	for _, callback := range lm.stateChangeCallbacks {
		go callback(oldState, newState)
	}

	lm.logger.Debug("Lifecycle state changed", "from", oldState.String(), "to", newState.String())
}

// executePhases 执行生命周期阶段
func (lm *KernelLifecycleManager) executePhases(ctx context.Context, phases []*LifecyclePhase, phaseName string) error {
	for _, phase := range phases {
		lm.logger.Info("Executing lifecycle phase", "phase", phase.Name, "type", phaseName)

		// 创建带超时的上下文
		phaseCtx := ctx
		var cancel context.CancelFunc
		if phase.Timeout > 0 {
			phaseCtx, cancel = context.WithTimeout(ctx, phase.Timeout)
			defer cancel()
		}

		// 执行阶段处理器
		if phase.Handler != nil {
			if err := phase.Handler(phaseCtx); err != nil {
				lm.logger.Error("Lifecycle phase failed", "phase", phase.Name, "error", err)

				// 触发错误回调
				for _, callback := range lm.errorCallbacks {
					go callback(phase.Name, err)
				}

				// 如果是必需阶段，返回错误
				if phase.Required {
					return fmt.Errorf("required phase '%s' failed: %w", phase.Name, err)
				}

				lm.logger.Warn("Optional phase failed, continuing", "phase", phase.Name)
			}
		}

		lm.logger.Debug("Lifecycle phase completed", "phase", phase.Name)
	}

	return nil
}

// validatePhase 验证生命周期阶段
func (lm *KernelLifecycleManager) validatePhase(phase *LifecyclePhase) error {
	if phase == nil {
		return fmt.Errorf("phase cannot be nil")
	}

	if phase.Name == "" {
		return fmt.Errorf("phase name cannot be empty")
	}

	if phase.Handler == nil {
		return fmt.Errorf("phase handler cannot be nil")
	}

	if phase.Timeout < 0 {
		return fmt.Errorf("phase timeout cannot be negative")
	}

	return nil
}

// sortPhases 按顺序排序阶段
func (lm *KernelLifecycleManager) sortPhases(phases []*LifecyclePhase) {
	// 简单的冒泡排序
	for i := 0; i < len(phases)-1; i++ {
		for j := 0; j < len(phases)-i-1; j++ {
			if phases[j].Order > phases[j+1].Order {
				phases[j], phases[j+1] = phases[j+1], phases[j]
			}
		}
	}
}

// addDefaultSignalHandlers 添加默认信号处理器
func (lm *KernelLifecycleManager) addDefaultSignalHandlers() {
	// SIGINT (Ctrl+C) - 优雅停止
	lm.signalHandlers[syscall.SIGINT] = func(sig os.Signal) {
		lm.logger.Info("Received SIGINT, initiating graceful shutdown...")
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := lm.Stop(ctx); err != nil {
				lm.logger.Error("Failed to stop gracefully", "error", err)
				os.Exit(1)
			}
			os.Exit(0)
		}()
	}

	// SIGTERM - 优雅停止
	lm.signalHandlers[syscall.SIGTERM] = func(sig os.Signal) {
		lm.logger.Info("Received SIGTERM, initiating graceful shutdown...")
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := lm.Stop(ctx); err != nil {
				lm.logger.Error("Failed to stop gracefully", "error", err)
				os.Exit(1)
			}
			os.Exit(0)
		}()
	}

	// 注册信号
	if lm.signalChan != nil {
		signal.Notify(lm.signalChan, syscall.SIGINT, syscall.SIGTERM)
	}
}

// handleSignals 处理信号
func (lm *KernelLifecycleManager) handleSignals() {
	for {
		select {
		case sig, ok := <-lm.signalChan:
			if !ok {
				return
			}

			lm.logger.Info("Received signal", "signal", sig.String())

			// 查找并执行信号处理器
			lm.mutex.RLock()
			handler, exists := lm.signalHandlers[sig]
			lm.mutex.RUnlock()

			if exists && handler != nil {
				handler(sig)
			} else {
				lm.logger.Warn("No handler for signal", "signal", sig.String())
			}

		case <-lm.ctx.Done():
			return
		}
	}
}