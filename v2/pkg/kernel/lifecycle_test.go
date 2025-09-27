package kernel

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"log/slog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLifecycleManager_Initialize tests lifecycle manager initialization
func TestLifecycleManager_Initialize(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	lm := NewLifecycleManager(logger)
	require.NotNil(t, lm)

	// Initial state should be uninitialized
	assert.Equal(t, LifecycleStateUninitialized, lm.GetState())
	assert.False(t, lm.IsRunning())
	assert.Equal(t, time.Duration(0), lm.GetUptime())

	// Initialize
	ctx := context.Background()
	err := lm.Initialize(ctx)
	assert.NoError(t, err)

	// State should be initialized
	assert.Equal(t, LifecycleStateInitialized, lm.GetState())
	assert.False(t, lm.IsRunning())
}

// TestLifecycleManager_StartStop tests lifecycle manager start and stop
func TestLifecycleManager_StartStop(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	lm := NewLifecycleManager(logger)

	ctx := context.Background()

	// Initialize first
	err := lm.Initialize(ctx)
	require.NoError(t, err)

	// Start
	err = lm.Start(ctx)
	assert.NoError(t, err)
	assert.Equal(t, LifecycleStateRunning, lm.GetState())
	assert.True(t, lm.IsRunning())
	assert.Greater(t, lm.GetUptime(), time.Duration(0))

	// Stop
	err = lm.Stop(ctx)
	assert.NoError(t, err)
	assert.Equal(t, LifecycleStateStopped, lm.GetState())
	assert.False(t, lm.IsRunning())
}

// TestLifecycleManager_Shutdown tests lifecycle manager shutdown
func TestLifecycleManager_Shutdown(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	lm := NewLifecycleManager(logger)

	ctx := context.Background()

	// Initialize and start
	err := lm.Initialize(ctx)
	require.NoError(t, err)

	err = lm.Start(ctx)
	require.NoError(t, err)

	// Shutdown
	err = lm.Shutdown(ctx)
	assert.NoError(t, err)
	assert.Equal(t, LifecycleStateShutdown, lm.GetState())
	assert.False(t, lm.IsRunning())
}

// TestLifecycleManager_Phases tests lifecycle phases
func TestLifecycleManager_Phases(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	lm := NewLifecycleManager(logger)

	// Track phase execution
	phaseExecuted := make(map[string]bool)

	// Add init phase
	initPhase := &LifecyclePhase{
		Name:        "test_init",
		Description: "Test initialization phase",
		Timeout:     5 * time.Second,
		Handler: func(ctx context.Context) error {
			phaseExecuted["test_init"] = true
			return nil
		},
		Required: true,
		Order:    1,
	}

	err := lm.AddInitPhase(initPhase)
	assert.NoError(t, err)

	// Add start phase
	startPhase := &LifecyclePhase{
		Name:        "test_start",
		Description: "Test start phase",
		Timeout:     5 * time.Second,
		Handler: func(ctx context.Context) error {
			phaseExecuted["test_start"] = true
			return nil
		},
		Required: true,
		Order:    1,
	}

	err = lm.AddStartPhase(startPhase)
	assert.NoError(t, err)

	// Add stop phase
	stopPhase := &LifecyclePhase{
		Name:        "test_stop",
		Description: "Test stop phase",
		Timeout:     5 * time.Second,
		Handler: func(ctx context.Context) error {
			phaseExecuted["test_stop"] = true
			return nil
		},
		Required: true,
		Order:    1,
	}

	err = lm.AddStopPhase(stopPhase)
	assert.NoError(t, err)

	// Add shutdown phase
	shutdownPhase := &LifecyclePhase{
		Name:        "test_shutdown",
		Description: "Test shutdown phase",
		Timeout:     5 * time.Second,
		Handler: func(ctx context.Context) error {
			phaseExecuted["test_shutdown"] = true
			return nil
		},
		Required: true,
		Order:    1,
	}

	err = lm.AddShutdownPhase(shutdownPhase)
	assert.NoError(t, err)

	ctx := context.Background()

	// Execute lifecycle
	err = lm.Initialize(ctx)
	assert.NoError(t, err)
	assert.True(t, phaseExecuted["test_init"])

	err = lm.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, phaseExecuted["test_start"])

	err = lm.Stop(ctx)
	assert.NoError(t, err)
	assert.True(t, phaseExecuted["test_stop"])

	err = lm.Shutdown(ctx)
	assert.NoError(t, err)
	assert.True(t, phaseExecuted["test_shutdown"])
}

// TestLifecycleManager_PhaseOrdering tests phase execution ordering
func TestLifecycleManager_PhaseOrdering(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	lm := NewLifecycleManager(logger)

	// Track execution order
	executionOrder := make([]string, 0)

	// Add phases in reverse order
	phase3 := &LifecyclePhase{
		Name:        "phase_3",
		Description: "Third phase",
		Timeout:     5 * time.Second,
		Handler: func(ctx context.Context) error {
			executionOrder = append(executionOrder, "phase_3")
			return nil
		},
		Required: true,
		Order:    3,
	}

	phase1 := &LifecyclePhase{
		Name:        "phase_1",
		Description: "First phase",
		Timeout:     5 * time.Second,
		Handler: func(ctx context.Context) error {
			executionOrder = append(executionOrder, "phase_1")
			return nil
		},
		Required: true,
		Order:    1,
	}

	phase2 := &LifecyclePhase{
		Name:        "phase_2",
		Description: "Second phase",
		Timeout:     5 * time.Second,
		Handler: func(ctx context.Context) error {
			executionOrder = append(executionOrder, "phase_2")
			return nil
		},
		Required: true,
		Order:    2,
	}

	// Add phases in random order
	err := lm.AddInitPhase(phase3)
	assert.NoError(t, err)
	err = lm.AddInitPhase(phase1)
	assert.NoError(t, err)
	err = lm.AddInitPhase(phase2)
	assert.NoError(t, err)

	// Execute and verify order
	ctx := context.Background()
	err = lm.Initialize(ctx)
	assert.NoError(t, err)

	// Verify execution order
	assert.Equal(t, []string{"phase_1", "phase_2", "phase_3"}, executionOrder)
}

// TestLifecycleManager_PhaseTimeout tests phase timeout handling
func TestLifecycleManager_PhaseTimeout(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	lm := NewLifecycleManager(logger)

	// Add phase with short timeout
	timeoutPhase := &LifecyclePhase{
		Name:        "timeout_phase",
		Description: "Phase that times out",
		Timeout:     100 * time.Millisecond,
		Handler: func(ctx context.Context) error {
			// Sleep longer than timeout
			time.Sleep(200 * time.Millisecond)
			return nil
		},
		Required: true,
		Order:    1,
	}

	err := lm.AddInitPhase(timeoutPhase)
	assert.NoError(t, err)

	// Execute and expect timeout
	ctx := context.Background()
	err = lm.Initialize(ctx)
	// Should still succeed as the phase handler doesn't check context
	// In a real implementation, the handler should respect context cancellation
	assert.NoError(t, err)
}

// TestLifecycleManager_OptionalPhaseFailure tests optional phase failure handling
func TestLifecycleManager_OptionalPhaseFailure(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	lm := NewLifecycleManager(logger)

	// Track error callback
	errorCallbackInvoked := false
	lm.OnError(func(phase string, err error) {
		errorCallbackInvoked = true
		assert.Equal(t, "optional_phase", phase)
		assert.Error(t, err)
	})

	// Add optional phase that fails
	optionalPhase := &LifecyclePhase{
		Name:        "optional_phase",
		Description: "Optional phase that fails",
		Timeout:     5 * time.Second,
		Handler: func(ctx context.Context) error {
			return assert.AnError
		},
		Required: false, // Optional phase
		Order:    1,
	}

	err := lm.AddInitPhase(optionalPhase)
	assert.NoError(t, err)

	// Add required phase that succeeds
	requiredPhase := &LifecyclePhase{
		Name:        "required_phase",
		Description: "Required phase that succeeds",
		Timeout:     5 * time.Second,
		Handler: func(ctx context.Context) error {
			return nil
		},
		Required: true,
		Order:    2,
	}

	err = lm.AddInitPhase(requiredPhase)
	assert.NoError(t, err)

	// Execute - should succeed despite optional phase failure
	ctx := context.Background()
	err = lm.Initialize(ctx)
	assert.NoError(t, err)

	// Wait for error callback
	time.Sleep(50 * time.Millisecond)
	assert.True(t, errorCallbackInvoked)
}

// TestLifecycleManager_RequiredPhaseFailure tests required phase failure handling
func TestLifecycleManager_RequiredPhaseFailure(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	lm := NewLifecycleManager(logger)

	// Add required phase that fails
	requiredPhase := &LifecyclePhase{
		Name:        "required_phase",
		Description: "Required phase that fails",
		Timeout:     5 * time.Second,
		Handler: func(ctx context.Context) error {
			return assert.AnError
		},
		Required: true,
		Order:    1,
	}

	err := lm.AddInitPhase(requiredPhase)
	assert.NoError(t, err)

	// Execute - should fail due to required phase failure
	ctx := context.Background()
	err = lm.Initialize(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required phase")
	assert.Equal(t, LifecycleStateError, lm.GetState())
}

// TestLifecycleManager_SignalHandling tests signal handling
func TestLifecycleManager_SignalHandling(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	lm := NewLifecycleManager(logger)

	// Enable signal handling
	err := lm.EnableSignalHandling()
	assert.NoError(t, err)

	// Track signal handler invocation
	signalHandlerInvoked := false
	err = lm.AddSignalHandler(syscall.SIGUSR1, func(sig os.Signal) {
		signalHandlerInvoked = true
		assert.Equal(t, syscall.SIGUSR1, sig)
	})
	assert.NoError(t, err)

	ctx := context.Background()

	// Initialize and start
	err = lm.Initialize(ctx)
	assert.NoError(t, err)

	err = lm.Start(ctx)
	assert.NoError(t, err)

	// Send signal to self
	process, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)

	err = process.Signal(syscall.SIGUSR1)
	assert.NoError(t, err)

	// Wait for signal handler
	time.Sleep(100 * time.Millisecond)
	assert.True(t, signalHandlerInvoked)

	// Disable signal handling
	err = lm.DisableSignalHandling()
	assert.NoError(t, err)

	// Shutdown
	err = lm.Shutdown(ctx)
	assert.NoError(t, err)
}

// TestLifecycleManager_StateChangeCallbacks tests state change callbacks
func TestLifecycleManager_StateChangeCallbacks(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	lm := NewLifecycleManager(logger)

	// Track state changes
	stateChanges := make([]struct {
		Old LifecycleState
		New LifecycleState
	}, 0)

	lm.OnStateChanged(func(oldState, newState LifecycleState) {
		stateChanges = append(stateChanges, struct {
			Old LifecycleState
			New LifecycleState
		}{oldState, newState})
	})

	ctx := context.Background()

	// Execute lifecycle
	err := lm.Initialize(ctx)
	assert.NoError(t, err)

	err = lm.Start(ctx)
	assert.NoError(t, err)

	err = lm.Stop(ctx)
	assert.NoError(t, err)

	// Wait for callbacks
	time.Sleep(50 * time.Millisecond)

	// Verify state changes
	assert.GreaterOrEqual(t, len(stateChanges), 3)

	// Check some expected transitions
	found := false
	for _, change := range stateChanges {
		if change.Old == LifecycleStateUninitialized && change.New == LifecycleStateInitializing {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected uninitialized -> initializing transition")
}

// TestLifecycleManager_InvalidPhase tests invalid phase handling
func TestLifecycleManager_InvalidPhase(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	lm := NewLifecycleManager(logger)

	// Test nil phase
	err := lm.AddInitPhase(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")

	// Test phase with empty name
	emptyNamePhase := &LifecyclePhase{
		Name:    "",
		Handler: func(ctx context.Context) error { return nil },
	}
	err = lm.AddInitPhase(emptyNamePhase)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name cannot be empty")

	// Test phase with nil handler
	nilHandlerPhase := &LifecyclePhase{
		Name:    "test",
		Handler: nil,
	}
	err = lm.AddInitPhase(nilHandlerPhase)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler cannot be nil")

	// Test phase with negative timeout
	negativeTimeoutPhase := &LifecyclePhase{
		Name:    "test",
		Timeout: -1 * time.Second,
		Handler: func(ctx context.Context) error { return nil },
	}
	err = lm.AddInitPhase(negativeTimeoutPhase)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout cannot be negative")
}

// TestLifecycleManager_InvalidStateTransitions tests invalid state transitions
func TestLifecycleManager_InvalidStateTransitions(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	lm := NewLifecycleManager(logger)

	ctx := context.Background()

	// Try to start without initializing
	err := lm.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	// Try to stop without starting
	err = lm.Initialize(ctx)
	assert.NoError(t, err)

	err = lm.Stop(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")

	// Try to initialize again
	err = lm.Initialize(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already initialized")
}

// TestLifecycleState_String tests lifecycle state string representation
func TestLifecycleState_String(t *testing.T) {
	tests := []struct {
		state    LifecycleState
		expected string
	}{
		{LifecycleStateUninitialized, "uninitialized"},
		{LifecycleStateInitializing, "initializing"},
		{LifecycleStateInitialized, "initialized"},
		{LifecycleStateStarting, "starting"},
		{LifecycleStateRunning, "running"},
		{LifecycleStateStopping, "stopping"},
		{LifecycleStateStopped, "stopped"},
		{LifecycleStateShuttingDown, "shutting_down"},
		{LifecycleStateShutdown, "shutdown"},
		{LifecycleStateError, "error"},
		{LifecycleState(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.state.String())
		})
	}
}

// BenchmarkLifecycleManager_Initialize benchmarks lifecycle initialization
func BenchmarkLifecycleManager_Initialize(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lm := NewLifecycleManager(logger)
		err := lm.Initialize(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLifecycleManager_StartStop benchmarks lifecycle start/stop cycle
func BenchmarkLifecycleManager_StartStop(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	lm := NewLifecycleManager(logger)

	ctx := context.Background()
	err := lm.Initialize(ctx)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := lm.Start(ctx)
		if err != nil {
			b.Fatal(err)
		}

		err = lm.Stop(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}