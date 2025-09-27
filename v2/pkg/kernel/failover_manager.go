// Package kernel provides service failover and recovery functionality
package kernel

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// FailoverManager manages service failover and recovery
type FailoverManager struct {
	circuitBreakers map[string]*CircuitBreaker      // serviceID -> circuit breaker
	failoverConfigs map[string]*ServiceFailoverConfig // serviceID -> config
	retryStates    map[string]*RetryState           // serviceID -> retry state
	mutex          sync.RWMutex
	logger         Logger
	registry       ServiceRegistry
	metrics        *MetricsManager
}

// RetryState tracks retry attempts for a service
type RetryState struct {
	ServiceID     string        `json:"service_id"`
	Attempts      int           `json:"attempts"`
	LastAttempt   time.Time     `json:"last_attempt"`
	NextRetry     time.Time     `json:"next_retry"`
	BackoffDelay  time.Duration `json:"backoff_delay"`
	TotalFailures int64         `json:"total_failures"`
	mutex         sync.RWMutex  `json:"-"`
}

// FailoverResult represents the result of a failover operation
type FailoverResult struct {
	Success         bool                    `json:"success"`
	OriginalService *ServiceInstance        `json:"original_service"`
	FailoverService *ServiceInstance        `json:"failover_service"`
	Reason          string                  `json:"reason"`
	Timestamp       time.Time               `json:"timestamp"`
	RecoveryTime    *time.Duration          `json:"recovery_time,omitempty"`
	Metadata        map[string]interface{}  `json:"metadata,omitempty"`
}

// RecoveryStrategy defines how services should recover from failures
type RecoveryStrategy string

const (
	RecoveryStrategyImmediate   RecoveryStrategy = "immediate"
	RecoveryStrategyExponential RecoveryStrategy = "exponential"
	RecoveryStrategyLinear      RecoveryStrategy = "linear"
	RecoveryStrategyFixed       RecoveryStrategy = "fixed"
)

// FailoverEvent represents a failover event
type FailoverEvent struct {
	Type        string                 `json:"type"`
	ServiceID   string                 `json:"service_id"`
	ServiceName string                 `json:"service_name"`
	Timestamp   time.Time              `json:"timestamp"`
	Reason      string                 `json:"reason"`
	Details     map[string]interface{} `json:"details"`
	Severity    string                 `json:"severity"`
}

// Failover event types
const (
	FailoverEventCircuitOpened   = "circuit_opened"
	FailoverEventCircuitClosed   = "circuit_closed"
	FailoverEventCircuitHalfOpen = "circuit_half_open"
	FailoverEventServiceFailed   = "service_failed"
	FailoverEventServiceRecovered = "service_recovered"
	FailoverEventRetryExhausted  = "retry_exhausted"
	FailoverEventFailoverTriggered = "failover_triggered"
)

// NewFailoverManager creates a new failover manager
func NewFailoverManager(logger Logger, registry ServiceRegistry, metrics *MetricsManager) *FailoverManager {
	return &FailoverManager{
		circuitBreakers: make(map[string]*CircuitBreaker),
		failoverConfigs: make(map[string]*ServiceFailoverConfig),
		retryStates:    make(map[string]*RetryState),
		logger:         logger,
		registry:       registry,
		metrics:        metrics,
	}
}

// ConfigureFailover configures failover settings for a service
func (fm *FailoverManager) ConfigureFailover(serviceID string, config *ServiceFailoverConfig) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()
	
	fm.failoverConfigs[serviceID] = config
	
	// Initialize circuit breaker if enabled
	if config.CircuitBreakerEnabled {
		cb := &CircuitBreaker{
			ServiceID:        serviceID,
			State:           CircuitBreakerStateClosed,
			FailureCount:    0,
			FailureThreshold: config.FailureThreshold,
			Timeout:         config.RecoveryTimeout,
		}
		fm.circuitBreakers[serviceID] = cb
	}
	
	// Initialize retry state
	fm.retryStates[serviceID] = &RetryState{
		ServiceID:    serviceID,
		Attempts:     0,
		BackoffDelay: config.RetryDelay,
	}
	
	fm.logger.Info("Failover configured",
		"service_id", serviceID,
		"max_retries", config.MaxRetries,
		"circuit_breaker_enabled", config.CircuitBreakerEnabled)
	
	return nil
}

// GetFailoverConfig returns failover configuration for a service
func (fm *FailoverManager) GetFailoverConfig(serviceID string) (*ServiceFailoverConfig, error) {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()
	
	config, exists := fm.failoverConfigs[serviceID]
	if !exists {
		return nil, fmt.Errorf("failover config not found for service %s", serviceID)
	}
	
	return config, nil
}

// GetCircuitBreaker returns circuit breaker for a service
func (fm *FailoverManager) GetCircuitBreaker(serviceID string) (*CircuitBreaker, error) {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()
	
	cb, exists := fm.circuitBreakers[serviceID]
	if !exists {
		return nil, fmt.Errorf("circuit breaker not found for service %s", serviceID)
	}
	
	return cb, nil
}

// UpdateCircuitBreakerState updates the state of a circuit breaker
func (fm *FailoverManager) UpdateCircuitBreakerState(serviceID string, state CircuitBreakerState) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()
	
	cb, exists := fm.circuitBreakers[serviceID]
	if !exists {
		return fmt.Errorf("circuit breaker not found for service %s", serviceID)
	}
	
	oldState := cb.State
	cb.State = state
	
	switch state {
	case CircuitBreakerStateOpen:
		cb.LastFailureTime = time.Now()
		cb.NextRetryTime = time.Now().Add(cb.Timeout)
	case CircuitBreakerStateClosed:
		cb.FailureCount = 0
	case CircuitBreakerStateHalfOpen:
		// Allow one test request
	}
	
	fm.logger.Info("Circuit breaker state changed",
		"service_id", serviceID,
		"old_state", oldState,
		"new_state", state)
	
	// Publish event
	fm.publishFailoverEvent(FailoverEventCircuitOpened, serviceID, "", 
		fmt.Sprintf("Circuit breaker state changed from %s to %s", oldState, state), nil)
	
	return nil
}

// RecordServiceCall records a service call result for circuit breaker logic
func (fm *FailoverManager) RecordServiceCall(serviceID string, success bool, duration time.Duration) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()
	
	cb, exists := fm.circuitBreakers[serviceID]
	if !exists {
		return nil // No circuit breaker configured
	}
	
	if success {
		fm.handleSuccessfulCall(cb)
	} else {
		fm.handleFailedCall(cb)
	}
	
	// Record metrics
	if fm.metrics != nil {
		fm.metrics.RecordServiceCall(serviceID, duration, success, "")
	}
	
	return nil
}

// CanMakeCall checks if a call can be made based on circuit breaker state
func (fm *FailoverManager) CanMakeCall(serviceID string) (bool, error) {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()
	
	cb, exists := fm.circuitBreakers[serviceID]
	if !exists {
		return true, nil // No circuit breaker, allow call
	}
	
	switch cb.State {
	case CircuitBreakerStateClosed:
		return true, nil
	case CircuitBreakerStateOpen:
		// Check if timeout has passed
		if time.Now().After(cb.NextRetryTime) {
			// Transition to half-open
			cb.State = CircuitBreakerStateHalfOpen
			fm.logger.Info("Circuit breaker transitioning to half-open", "service_id", serviceID)
			return true, nil
		}
		return false, fmt.Errorf("circuit breaker is open for service %s", serviceID)
	case CircuitBreakerStateHalfOpen:
		return true, nil // Allow test call
	default:
		return false, fmt.Errorf("unknown circuit breaker state: %s", cb.State)
	}
}

// AttemptFailover attempts to failover to an alternative service instance
func (fm *FailoverManager) AttemptFailover(ctx context.Context, serviceName string, failedServiceID string) (*FailoverResult, error) {
	// Get healthy instances of the same service
	healthyInstances, err := fm.registry.Discover(ctx, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to discover healthy instances: %w", err)
	}
	
	// Filter out the failed service
	alternatives := make([]*ServiceInstance, 0)
	for _, instance := range healthyInstances {
		if instance.Info.ID != failedServiceID {
			alternatives = append(alternatives, instance)
		}
	}
	
	if len(alternatives) == 0 {
		return &FailoverResult{
			Success:   false,
			Reason:    "No healthy alternative instances available",
			Timestamp: time.Now(),
		}, nil
	}
	
	// Select the best alternative (could use load balancing logic)
	failoverInstance := alternatives[0] // Simple selection for now
	
	// Get original service info
	originalService, err := fm.registry.GetService(ctx, failedServiceID)
	if err != nil {
		fm.logger.Warn("Could not get original service info", "service_id", failedServiceID, "error", err)
	}
	
	result := &FailoverResult{
		Success:         true,
		OriginalService: originalService,
		FailoverService: failoverInstance,
		Reason:          "Service failure detected, failed over to healthy instance",
		Timestamp:       time.Now(),
		Metadata: map[string]interface{}{
			"failed_service_id":    failedServiceID,
			"failover_service_id":  failoverInstance.Info.ID,
			"available_alternatives": len(alternatives),
		},
	}
	
	fm.logger.Info("Failover successful",
		"service_name", serviceName,
		"failed_service_id", failedServiceID,
		"failover_service_id", failoverInstance.Info.ID)
	
	// Publish failover event
	fm.publishFailoverEvent(FailoverEventFailoverTriggered, failedServiceID, serviceName,
		"Service failed over to healthy instance", result.Metadata)
	
	return result, nil
}

// AttemptRecovery attempts to recover a failed service
func (fm *FailoverManager) AttemptRecovery(ctx context.Context, serviceID string) (*FailoverResult, error) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()
	
	config, exists := fm.failoverConfigs[serviceID]
	if !exists {
		return nil, fmt.Errorf("no failover config for service %s", serviceID)
	}
	
	retryState := fm.retryStates[serviceID]
	if retryState == nil {
		return nil, fmt.Errorf("no retry state for service %s", serviceID)
	}
	
	// Check if we can retry
	if retryState.Attempts >= config.MaxRetries {
		return &FailoverResult{
			Success:   false,
			Reason:    "Maximum retry attempts exceeded",
			Timestamp: time.Now(),
		}, nil
	}
	
	// Check if enough time has passed since last attempt
	if time.Now().Before(retryState.NextRetry) {
		return &FailoverResult{
			Success:   false,
			Reason:    "Retry delay not yet elapsed",
			Timestamp: time.Now(),
		}, nil
	}
	
	// Attempt recovery by checking service health
	err := fm.registry.CheckHealth(ctx, serviceID)
	if err != nil {
		// Recovery failed, update retry state
		fm.updateRetryState(retryState, config, false)
		return &FailoverResult{
			Success:   false,
			Reason:    fmt.Sprintf("Health check failed: %v", err),
			Timestamp: time.Now(),
		}, nil
	}
	
	// Recovery successful
	fm.resetRetryState(retryState)
	
	// Reset circuit breaker if exists
	if cb, exists := fm.circuitBreakers[serviceID]; exists {
		cb.State = CircuitBreakerStateClosed
		cb.FailureCount = 0
	}
	
	recoveryTime := time.Since(retryState.LastAttempt)
	result := &FailoverResult{
		Success:      true,
		Reason:       "Service recovered successfully",
		Timestamp:    time.Now(),
		RecoveryTime: &recoveryTime,
		Metadata: map[string]interface{}{
			"attempts": retryState.Attempts,
			"total_failures": retryState.TotalFailures,
		},
	}
	
	fm.logger.Info("Service recovered",
		"service_id", serviceID,
		"attempts", retryState.Attempts,
		"recovery_time", recoveryTime)
	
	// Publish recovery event
	fm.publishFailoverEvent(FailoverEventServiceRecovered, serviceID, "",
		"Service recovered after failure", result.Metadata)
	
	return result, nil
}

// Helper methods

func (fm *FailoverManager) handleSuccessfulCall(cb *CircuitBreaker) {
	switch cb.State {
	case CircuitBreakerStateHalfOpen:
		// Successful call in half-open state, close the circuit
		cb.State = CircuitBreakerStateClosed
		cb.FailureCount = 0
		fm.logger.Info("Circuit breaker closed after successful call", "service_id", cb.ServiceID)
	case CircuitBreakerStateClosed:
		// Reset failure count on successful call
		cb.FailureCount = 0
	}
}

func (fm *FailoverManager) handleFailedCall(cb *CircuitBreaker) {
	switch cb.State {
	case CircuitBreakerStateClosed:
		cb.FailureCount++
		if cb.FailureCount >= cb.FailureThreshold {
			// Open the circuit
			cb.State = CircuitBreakerStateOpen
			cb.LastFailureTime = time.Now()
			cb.NextRetryTime = time.Now().Add(cb.Timeout)
			fm.logger.Warn("Circuit breaker opened due to failures",
				"service_id", cb.ServiceID,
				"failure_count", cb.FailureCount)
		}
	case CircuitBreakerStateHalfOpen:
		// Failed call in half-open state, reopen the circuit
		cb.State = CircuitBreakerStateOpen
		cb.FailureCount++
		cb.LastFailureTime = time.Now()
		cb.NextRetryTime = time.Now().Add(cb.Timeout)
		fm.logger.Warn("Circuit breaker reopened after failed test call", "service_id", cb.ServiceID)
	case CircuitBreakerStateOpen:
		// Already open, just increment failure count
		cb.FailureCount++
	}
}

func (fm *FailoverManager) updateRetryState(retryState *RetryState, config *ServiceFailoverConfig, success bool) {
	retryState.mutex.Lock()
	defer retryState.mutex.Unlock()
	
	retryState.LastAttempt = time.Now()
	
	if success {
		retryState.Attempts = 0
		retryState.BackoffDelay = config.RetryDelay
	} else {
		retryState.Attempts++
		retryState.TotalFailures++
		
		// Calculate next retry time with backoff
		backoff := fm.calculateBackoff(retryState.BackoffDelay, retryState.Attempts, config)
		retryState.NextRetry = time.Now().Add(backoff)
		retryState.BackoffDelay = backoff
	}
}

func (fm *FailoverManager) resetRetryState(retryState *RetryState) {
	retryState.mutex.Lock()
	defer retryState.mutex.Unlock()
	
	retryState.Attempts = 0
	retryState.NextRetry = time.Time{}
}

func (fm *FailoverManager) calculateBackoff(baseDelay time.Duration, attempts int, config *ServiceFailoverConfig) time.Duration {
	backoff := float64(baseDelay) * math.Pow(config.BackoffMultiplier, float64(attempts-1))
	if time.Duration(backoff) > config.MaxRetryDelay {
		return config.MaxRetryDelay
	}
	return time.Duration(backoff)
}

func (fm *FailoverManager) publishFailoverEvent(eventType, serviceID, serviceName, reason string, details map[string]interface{}) {
	event := &FailoverEvent{
		Type:        eventType,
		ServiceID:   serviceID,
		ServiceName: serviceName,
		Timestamp:   time.Now(),
		Reason:      reason,
		Details:     details,
		Severity:    "info",
	}
	
	// Log the event
	fm.logger.Info("Failover event",
		"type", eventType,
		"service_id", serviceID,
		"reason", reason)
	
	// Use the event variable to avoid unused variable error
	_ = event
	
	// Could publish to event bus here if available
}

// GetFailoverStatistics returns failover statistics
func (fm *FailoverManager) GetFailoverStatistics() map[string]interface{} {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()
	
	stats := make(map[string]interface{})
	stats["total_circuit_breakers"] = len(fm.circuitBreakers)
	stats["total_failover_configs"] = len(fm.failoverConfigs)
	
	openCircuits := 0
	closedCircuits := 0
	halfOpenCircuits := 0
	
	for _, cb := range fm.circuitBreakers {
		switch cb.State {
		case CircuitBreakerStateOpen:
			openCircuits++
		case CircuitBreakerStateClosed:
			closedCircuits++
		case CircuitBreakerStateHalfOpen:
			halfOpenCircuits++
		}
	}
	
	stats["open_circuits"] = openCircuits
	stats["closed_circuits"] = closedCircuits
	stats["half_open_circuits"] = halfOpenCircuits
	
	var totalFailures int64
	var totalAttempts int
	for _, retryState := range fm.retryStates {
		totalFailures += retryState.TotalFailures
		totalAttempts += retryState.Attempts
	}
	
	stats["total_failures"] = totalFailures
	stats["total_retry_attempts"] = totalAttempts
	
	return stats
}