// Package kernel provides the microkernel core implementation
package kernel

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Logger interface for dependency injection compatibility
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	With(args ...interface{}) Logger
}

// SlogAdapter adapts slog.Logger to our Logger interface
type SlogAdapter struct {
	logger *slog.Logger
}

func NewSlogAdapter(logger *slog.Logger) Logger {
	return &SlogAdapter{logger: logger}
}

func (s *SlogAdapter) Debug(msg string, args ...interface{}) {
	s.logger.Debug(msg, args...)
}

func (s *SlogAdapter) Info(msg string, args ...interface{}) {
	s.logger.Info(msg, args...)
}

func (s *SlogAdapter) Warn(msg string, args ...interface{}) {
	s.logger.Warn(msg, args...)
}

func (s *SlogAdapter) Error(msg string, args ...interface{}) {
	s.logger.Error(msg, args...)
}

func (s *SlogAdapter) With(args ...interface{}) Logger {
	return &SlogAdapter{logger: s.logger.With(args...)}
}

// ServiceState represents the state of a service
type ServiceState int

const (
	// ServiceStateUnknown represents unknown service state
	ServiceStateUnknown ServiceState = iota
	// ServiceStateRegistering represents service is being registered
	ServiceStateRegistering
	// ServiceStateRegistered represents service has been registered
	ServiceStateRegistered
	// ServiceStateHealthy represents service is healthy and available
	ServiceStateHealthy
	// ServiceStateUnhealthy represents service is unhealthy
	ServiceStateUnhealthy
	// ServiceStateDeregistering represents service is being deregistered
	ServiceStateDeregistering
	// ServiceStateDeregistered represents service has been deregistered
	ServiceStateDeregistered
)

// String returns the string representation of ServiceState
func (s ServiceState) String() string {
	switch s {
	case ServiceStateRegistering:
		return "registering"
	case ServiceStateRegistered:
		return "registered"
	case ServiceStateHealthy:
		return "healthy"
	case ServiceStateUnhealthy:
		return "unhealthy"
	case ServiceStateDeregistering:
		return "deregistering"
	case ServiceStateDeregistered:
		return "deregistered"
	default:
		return "unknown"
	}
}

// LoadBalanceStrategy represents load balancing strategy
type LoadBalanceStrategy int

const (
	// LoadBalanceRoundRobin represents round-robin load balancing
	LoadBalanceRoundRobin LoadBalanceStrategy = iota
	// LoadBalanceRandom represents random load balancing
	LoadBalanceRandom
	// LoadBalanceWeightedRoundRobin represents weighted round-robin load balancing
	LoadBalanceWeightedRoundRobin
	// LoadBalanceLeastConnections represents least connections load balancing
	LoadBalanceLeastConnections
)

// ServiceInfo represents service information
type ServiceInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Address     string            `json:"address"`
	Port        int               `json:"port"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
	HealthCheck *HealthCheckInfo  `json:"health_check"`
	Dependencies []string         `json:"dependencies"`
	Weight      int               `json:"weight"`
	RegisteredAt time.Time        `json:"registered_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// HealthCheckInfo represents health check configuration
type HealthCheckInfo struct {
	Enabled          bool          `json:"enabled"`
	Interval         time.Duration `json:"interval"`
	Timeout          time.Duration `json:"timeout"`
	Endpoint         string        `json:"endpoint"`
	Method           string        `json:"method"`
	FailureThreshold int           `json:"failure_threshold"`
}

// ServiceInstance represents a service instance with runtime information
type ServiceInstance struct {
	Info         *ServiceInfo  `json:"info"`
	State        ServiceState  `json:"state"`
	LastHealthy  time.Time     `json:"last_healthy"`
	LastSeen     time.Time     `json:"last_seen"`
	FailureCount int           `json:"failure_count"`
	Connections  int           `json:"connections"`
	mutex        sync.RWMutex  `json:"-"`
}

// SetState sets the service state in a thread-safe manner
func (si *ServiceInstance) SetState(state ServiceState) {
	si.mutex.Lock()
	defer si.mutex.Unlock()
	si.State = state
}

// GetState gets the service state in a thread-safe manner
func (si *ServiceInstance) GetState() ServiceState {
	si.mutex.RLock()
	defer si.mutex.RUnlock()
	return si.State
}

// IncrementConnections increments the connection count
func (si *ServiceInstance) IncrementConnections() {
	si.mutex.Lock()
	defer si.mutex.Unlock()
	si.Connections++
}

// DecrementConnections decrements the connection count
func (si *ServiceInstance) DecrementConnections() {
	si.mutex.Lock()
	defer si.mutex.Unlock()
	if si.Connections > 0 {
		si.Connections--
	}
}

// GetConnections gets the current connection count
func (si *ServiceInstance) GetConnections() int {
	si.mutex.RLock()
	defer si.mutex.RUnlock()
	return si.Connections
}

// IncrementFailureCount increments the failure count
func (si *ServiceInstance) IncrementFailureCount() {
	si.mutex.Lock()
	defer si.mutex.Unlock()
	si.FailureCount++
}

// ResetFailureCount resets the failure count
func (si *ServiceInstance) ResetFailureCount() {
	si.mutex.Lock()
	defer si.mutex.Unlock()
	si.FailureCount = 0
	si.LastHealthy = time.Now()
}

// ServiceEvent represents a service registry event
type ServiceEvent struct {
	Type      interface{}      `json:"type"`
	Service   *ServiceInfo     `json:"service"`
	Timestamp time.Time        `json:"timestamp"`
	Message   string           `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// ServiceEventFilter represents a filter for service events
type ServiceEventFilter struct {
	ServiceName string `json:"service_name,omitempty"`
	EventType   string `json:"event_type,omitempty"`
}

// Service event types
const (
	ServiceEventRegistered        = "service_registered"
	ServiceEventDeregistered      = "service_deregistered"
	ServiceEventUpdated           = "service_updated"
	ServiceEventHealthy           = "service_healthy"
	ServiceEventUnhealthy         = "service_unhealthy"
	ServiceEventStateChanged      = "service_state_changed"
	ServiceEventHealthCheckPassed = "service_health_check_passed"
	ServiceEventHealthCheckFailed = "service_health_check_failed"
)



// ServiceQuery represents a service discovery query
type ServiceQuery struct {
	ServiceName string            `json:"service_name"`
	Name        string            `json:"name"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
	Healthy     *bool             `json:"healthy"`
	State       ServiceState      `json:"state"`
}

// ServiceStats represents statistics about the service registry
type ServiceStats struct {
	TotalServices     int                    `json:"total_services"`
	HealthyServices   int                    `json:"healthy_services"`
	UnhealthyServices int                    `json:"unhealthy_services"`
	RegistryUptime    time.Duration          `json:"registry_uptime"`
	ServicesByState   map[ServiceState]int   `json:"services_by_state"`
}

// HealthCheckConfig represents health check configuration
type HealthCheckConfig struct {
	Type             string `json:"type"`
	Path             string `json:"path,omitempty"`
	TimeoutSeconds   int    `json:"timeout_seconds"`
	FailureThreshold int    `json:"failure_threshold"`
}

// HealthChecker represents a health checker for services
type HealthChecker struct {
	registry *ServiceRegistryImpl
	interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
	logger   *slog.Logger
	running  bool
	mutex    sync.Mutex
	done     chan struct{}
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(registry *ServiceRegistryImpl, interval time.Duration, logger *slog.Logger) *HealthChecker {
	ctx, cancel := context.WithCancel(context.Background())
	return &HealthChecker{
		registry: registry,
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
		logger:   logger,
		running:  false,
		done:     make(chan struct{}),
	}
}

// Start starts the health checker
func (hc *HealthChecker) Start() {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	
	if hc.running {
		return // Already running
	}
	
	hc.running = true
	// Create new done channel for this run
	hc.done = make(chan struct{})
	
	go func() {
		defer func() {
			hc.mutex.Lock()
			hc.running = false
			hc.mutex.Unlock()
			select {
			case <-hc.done:
				// Channel already closed
			default:
				close(hc.done)
			}
		}()
		
		ticker := time.NewTicker(hc.interval)
		defer ticker.Stop()
		
		for {
			select {
			case <-hc.ctx.Done():
				return
			case <-ticker.C:
				hc.checkAllServices()
			}
		}
	}()
}

// Stop stops the health checker
func (hc *HealthChecker) Stop() {
	hc.mutex.Lock()
	if !hc.running {
		hc.mutex.Unlock()
		return // Not running
	}
	hc.mutex.Unlock()
	
	hc.cancel()
	
	// Wait for goroutine to finish with timeout
	select {
	case <-hc.done:
		// Goroutine finished normally
	case <-time.After(5 * time.Second):
		// Timeout waiting for goroutine to finish
		hc.logger.Warn("Health checker stop timeout")
	}
}

// checkAllServices checks the health of all registered services
func (hc *HealthChecker) checkAllServices() {
	services, err := hc.registry.ListServices(hc.ctx)
	if err != nil {
		hc.logger.Error("Failed to list services for health check", "error", err)
		return
	}
	
	for _, service := range services {
		if service.Info.HealthCheck != nil && service.Info.HealthCheck.Enabled {
			go func(serviceID string) {
				if err := hc.registry.CheckHealth(hc.ctx, serviceID); err != nil {
					hc.logger.Warn("Health check failed", "service_id", serviceID, "error", err)
				}
			}(service.Info.ID)
		}
	}
}

// ServiceRegistry defines the service registry interface
type ServiceRegistry interface {
	// Initialize initializes the service registry
	Initialize(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Shutdown(ctx context.Context) error
	
	// Register registers a service
	Register(ctx context.Context, service *ServiceInfo) error
	
	// Deregister deregisters a service
	Deregister(ctx context.Context, serviceID string) error
	
	// Discover discovers services by name
	Discover(ctx context.Context, serviceName string) ([]*ServiceInstance, error)
	
	// Query queries services with filters
	Query(ctx context.Context, query *ServiceQuery) ([]*ServiceInstance, error)
	
	// GetService gets a specific service by ID
	GetService(ctx context.Context, serviceID string) (*ServiceInstance, error)
	
	// ListServices lists all registered services
	ListServices(ctx context.Context) ([]*ServiceInstance, error)
	
	// UpdateService updates service information
	UpdateService(ctx context.Context, serviceID string, updates map[string]interface{}) error
	
	// SelectService selects a service instance using load balancing
	SelectService(ctx context.Context, serviceName string, strategy LoadBalanceStrategy) (*ServiceInstance, error)
	
	// CheckDependencies checks if service dependencies are satisfied
	CheckDependencies(ctx context.Context, serviceID string) error
	
	// Subscribe subscribes to service events
	Subscribe(ctx context.Context, callback func(*ServiceEvent)) error
	
	// Unsubscribe unsubscribes from service events
	Unsubscribe(ctx context.Context, callback func(*ServiceEvent)) error
	
	// StartHealthCheck starts health checking for all services
	StartHealthCheck(ctx context.Context) error
	
	// StopHealthCheck stops health checking
	StopHealthCheck(ctx context.Context) error
	
	// CheckHealth performs health check for a specific service
	CheckHealth(ctx context.Context, serviceID string) error
	
	// GetStats returns registry statistics
	GetStats(ctx context.Context) (*RegistryStats, error)
}

// RegistryStats represents service registry statistics
type RegistryStats struct {
	TotalServices   int                        `json:"total_services"`
	HealthyServices int                        `json:"healthy_services"`
	UnhealthyServices int                      `json:"unhealthy_services"`
	ServicesByState map[ServiceState]int       `json:"services_by_state"`
	ServicesByName  map[string]int             `json:"services_by_name"`
	Uptime          time.Duration              `json:"uptime"`
	LastUpdated     time.Time                  `json:"last_updated"`
}

// RegistryError represents a service registry error
type RegistryError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Cause     error                  `json:"cause,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	ServiceID string                 `json:"service_id,omitempty"`
}

// Error implements the error interface
func (e *RegistryError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("registry error [%s]: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("registry error [%s]: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause error
func (e *RegistryError) Unwrap() error {
	return e.Cause
}

// WithDetails adds details to the error
func (e *RegistryError) WithDetails(details map[string]interface{}) *RegistryError {
	e.Details = details
	return e
}

// WithServiceID adds service ID to the error
func (e *RegistryError) WithServiceID(serviceID string) *RegistryError {
	e.ServiceID = serviceID
	return e
}

// NewRegistryError creates a new registry error
func NewRegistryError(code, message string, cause error) *RegistryError {
	return &RegistryError{
		Code:      code,
		Message:   message,
		Cause:     cause,
		Timestamp: time.Now(),
	}
}

// Error codes
const (
	ErrCodeServiceNotFound     = "SERVICE_NOT_FOUND"
	ErrCodeServiceAlreadyExists = "SERVICE_ALREADY_EXISTS"
	ErrCodeInvalidServiceInfo  = "INVALID_SERVICE_INFO"
	ErrCodeHealthCheckFailed   = "HEALTH_CHECK_FAILED"
	ErrCodeDependencyNotMet    = "DEPENDENCY_NOT_MET"
	ErrCodeRegistryUnavailable = "REGISTRY_UNAVAILABLE"
	ErrCodeRegistryShutdown    = "REGISTRY_SHUTDOWN"
	ErrCodeInternalError       = "INTERNAL_ERROR"
	ErrCodeTimeout             = "TIMEOUT"
	ErrCodeNetworkError        = "NETWORK_ERROR"
	ErrCodeConfigurationError  = "CONFIGURATION_ERROR"
)



// String returns the string representation of LoadBalanceStrategy
func (s LoadBalanceStrategy) String() string {
	switch s {
	case LoadBalanceRoundRobin:
		return "round_robin"
	case LoadBalanceRandom:
		return "random"
	case LoadBalanceWeightedRoundRobin:
		return "weighted_round_robin"
	case LoadBalanceLeastConnections:
		return "least_connections"
	default:
		return "unknown"
	}
}



// ServiceRegistryImpl implements the ServiceRegistry interface
type ServiceRegistryImpl struct {
	// Core registry data
	services       map[string]*ServiceInstance
	servicesByName map[string][]*ServiceInstance
	mutex          sync.RWMutex

	// Event notification
	subscribers []func(*ServiceEvent)

	// Health checking
	healthChecker *HealthChecker

	// Load balancing state
	roundRobinCounters map[string]int
	weightedCounters   map[string]map[string]int
	random             *rand.Rand
	lbMutex            sync.RWMutex

	// Event history and filtering
	eventHistory   []*ServiceEvent
	maxEventHistory int

	// Dependency management
	dependencyGraph map[string][]string

	// Logging
	logger Logger

	// Shutdown flag
	shutdown bool

	// Additional fields from original
	startTime      time.Time
	subMutex       sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	eventMutex     sync.RWMutex
	depMutex       sync.RWMutex
}

// NewServiceRegistry creates a new service registry instance
func NewServiceRegistry(logger *slog.Logger) ServiceRegistry {
	ctx, cancel := context.WithCancel(context.Background())
	
	registry := &ServiceRegistryImpl{
		services:           make(map[string]*ServiceInstance),
		servicesByName:     make(map[string][]*ServiceInstance),
		subscribers:        make([]func(*ServiceEvent), 0),
		logger:             NewSlogAdapter(logger),
		startTime:          time.Now(),
		ctx:                ctx,
		cancel:             cancel,
		roundRobinCounters: make(map[string]int),
		weightedCounters:   make(map[string]map[string]int),
		random:             rand.New(rand.NewSource(time.Now().UnixNano())),
		eventHistory:       make([]*ServiceEvent, 0),
		maxEventHistory:    1000, // Keep last 1000 events
		dependencyGraph:    make(map[string][]string),
		shutdown:           false,
	}
	
	// Initialize health checker with 30 second interval but don't start it automatically
	registry.healthChecker = NewHealthChecker(registry, 30*time.Second, logger)
	// Note: Health checker is not started automatically to prevent goroutine leaks in tests
	// Call StartHealthCheck() explicitly when needed
	
	return registry
}

// Initialize initializes the service registry
func (r *ServiceRegistryImpl) Initialize(ctx context.Context) error {
	r.logger.Info("Initializing service registry")
	return nil
}

// Start starts the service registry
func (r *ServiceRegistryImpl) Start(ctx context.Context) error {
	r.logger.Info("Starting service registry")
	return nil
}

// Stop stops the service registry
func (r *ServiceRegistryImpl) Stop(ctx context.Context) error {
	r.logger.Info("Stopping service registry")
	return nil
}



// Register registers a new service
func (r *ServiceRegistryImpl) Register(ctx context.Context, info *ServiceInfo) error {
	// Check if registry is shutdown
	if r.shutdown {
		err := NewRegistryError(ErrCodeRegistryShutdown, "registry is shutdown", nil)
		r.logger.Error("Registration failed - registry shutdown", "service_id", info.ID)
		return err
	}

	// Validate service info with detailed logging
	if err := r.validateServiceInfo(info); err != nil {
		r.logger.Error("Service registration failed - validation error",
			"service_id", info.ID,
			"service_name", info.Name,
			"error", err.Error())
		return err
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if service already exists
	if _, exists := r.services[info.ID]; exists {
		err := NewRegistryError(ErrCodeServiceAlreadyExists, fmt.Sprintf("service with ID %s already exists", info.ID), nil)
		r.logger.Warn("Service registration failed - already exists",
			"service_id", info.ID,
			"service_name", info.Name)
		return err
	}

	// Create service instance with error handling
	instance := &ServiceInstance{
		Info:        info,
		State:       ServiceStateRegistering,
		Connections: 0,
		LastSeen:    time.Now(),
	}

	// Add to registry with panic recovery
	defer func() {
		if rec := recover(); rec != nil {
			r.logger.Error("Panic during service registration",
				"service_id", info.ID,
				"panic", rec)
		}
	}()

	r.services[info.ID] = instance
	if r.servicesByName[info.Name] == nil {
		r.servicesByName[info.Name] = make([]*ServiceInstance, 0)
	}
	r.servicesByName[info.Name] = append(r.servicesByName[info.Name], instance)

	// Set state to registered and healthy (for services without health checks)
	if info.HealthCheck == nil || !info.HealthCheck.Enabled {
		instance.SetState(ServiceStateHealthy)
	} else {
		instance.SetState(ServiceStateRegistered)
	}

	r.logger.Info("Service registered successfully",
		"service_id", info.ID,
		"service_name", info.Name,
		"address", fmt.Sprintf("%s:%d", info.Address, info.Port),
		"tags", strings.Join(info.Tags, ","),
		"weight", info.Weight)

	// Publish registration event with error handling
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				r.logger.Error("Panic during event publishing",
					"service_id", info.ID,
					"event_type", "registration",
					"panic", rec)
			}
		}()
		r.publishEvent(&ServiceEvent{
			Type:      ServiceEventRegistered,
			Service:   info,
			Timestamp: time.Now(),
			Message:   fmt.Sprintf("Service %s registered", info.Name),
			Details: map[string]interface{}{
				"address": fmt.Sprintf("%s:%d", info.Address, info.Port),
				"tags":    info.Tags,
				"weight":  info.Weight,
			},
		})
	}()

	return nil
}

// performHTTPHealthCheck performs HTTP health check
func (r *ServiceRegistryImpl) performHTTPHealthCheck(healthCheck *HealthCheckInfo) error {
	if healthCheck.Endpoint == "" {
		return fmt.Errorf("HTTP health check endpoint not configured")
	}
	
	client := &http.Client{
		Timeout: healthCheck.Timeout,
	}
	
	method := healthCheck.Method
	if method == "" {
		method = "GET"
	}
	
	req, err := http.NewRequest(method, healthCheck.Endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP health check failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP health check failed with status: %d", resp.StatusCode)
	}
	
	return nil
}

// performTCPHealthCheck performs TCP health check
func (r *ServiceRegistryImpl) performTCPHealthCheck(healthCheck *HealthCheckInfo) error {
	if healthCheck.Endpoint == "" {
		return fmt.Errorf("TCP health check endpoint not configured")
	}
	
	conn, err := net.DialTimeout("tcp", healthCheck.Endpoint, healthCheck.Timeout)
	if err != nil {
		return fmt.Errorf("TCP health check failed: %w", err)
	}
	defer conn.Close()
	
	return nil
}

// performGRPCHealthCheck performs gRPC health check
func (r *ServiceRegistryImpl) performGRPCHealthCheck(healthCheck *HealthCheckInfo) error {
	// For now, just perform a TCP check to the gRPC endpoint
	// In a real implementation, you would use the gRPC health checking protocol
	return r.performTCPHealthCheck(healthCheck)
}

// GetStats returns registry statistics
func (r *ServiceRegistryImpl) GetStats(ctx context.Context) (*RegistryStats, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	stats := &RegistryStats{
		TotalServices:     len(r.services),
		HealthyServices:   0,
		UnhealthyServices: 0,
		ServicesByState:   make(map[ServiceState]int),
		ServicesByName:    make(map[string]int),
		Uptime:            time.Since(r.startTime),
		LastUpdated:       time.Now(),
	}
	
	// Count services by state and name
	for _, instance := range r.services {
		state := instance.GetState()
		stats.ServicesByState[state]++
		
		if state == ServiceStateHealthy {
			stats.HealthyServices++
		} else {
			stats.UnhealthyServices++
		}
		
		stats.ServicesByName[instance.Info.Name]++
	}
	
	return stats, nil
}

// SelectService selects a service instance using load balancing
func (r *ServiceRegistryImpl) SelectService(ctx context.Context, serviceName string, strategy LoadBalanceStrategy) (*ServiceInstance, error) {
	return r.SelectInstance(ctx, serviceName, strategy)
}

// CheckHealth performs health check on a service
func (r *ServiceRegistryImpl) CheckHealth(ctx context.Context, serviceID string) error {
	r.mutex.RLock()
	instance, exists := r.services[serviceID]
	r.mutex.RUnlock()
	
	if !exists {
		return NewRegistryError(ErrCodeServiceNotFound, fmt.Sprintf("service with ID %s not found", serviceID), nil)
	}
	
	if instance.Info.HealthCheck == nil || !instance.Info.HealthCheck.Enabled {
		return nil // Health check not configured
	}
	
	healthCheck := instance.Info.HealthCheck
	
	// Perform health check based on type
	switch healthCheck.Method {
	case "http", "HTTP":
		return r.performHTTPHealthCheck(healthCheck)
	case "tcp", "TCP":
		return r.performTCPHealthCheck(healthCheck)
	case "grpc", "GRPC":
		return r.performGRPCHealthCheck(healthCheck)
	default:
		return fmt.Errorf("unsupported health check type: %s", healthCheck.Method)
	}
}

// StartHealthCheck starts health checking for all services
func (r *ServiceRegistryImpl) StartHealthCheck(ctx context.Context) error {
	if r.healthChecker != nil {
		r.healthChecker.Start()
	}
	return nil
}

// StopHealthCheck stops health checking
func (r *ServiceRegistryImpl) StopHealthCheck(ctx context.Context) error {
	if r.healthChecker != nil {
		r.healthChecker.Stop()
	}
	return nil
}

// Load balancing implementation

// SelectInstance selects a service instance based on the specified load balancing strategy
func (r *ServiceRegistryImpl) SelectInstance(ctx context.Context, serviceName string, strategy LoadBalanceStrategy) (*ServiceInstance, error) {
	r.mutex.RLock()
	instances, exists := r.servicesByName[serviceName]
	r.mutex.RUnlock()
	
	if !exists || len(instances) == 0 {
		return nil, NewRegistryError(ErrCodeServiceNotFound, fmt.Sprintf("no instances found for service %s", serviceName), nil)
	}
	
	// Filter healthy instances
	healthyInstances := make([]*ServiceInstance, 0)
	for _, instance := range instances {
		if instance.GetState() == ServiceStateHealthy {
			healthyInstances = append(healthyInstances, instance)
		}
	}
	
	if len(healthyInstances) == 0 {
		return nil, NewRegistryError(ErrCodeServiceNotFound, fmt.Sprintf("no healthy instances found for service %s", serviceName), nil)
	}
	
	switch strategy {
	case LoadBalanceRoundRobin:
		return r.selectRoundRobin(serviceName, healthyInstances), nil
	case LoadBalanceRandom:
		return r.selectRandom(healthyInstances), nil
	case LoadBalanceWeightedRoundRobin:
		return r.selectWeightedRoundRobin(serviceName, healthyInstances), nil
	case LoadBalanceLeastConnections:
		return r.selectLeastConnections(healthyInstances), nil
	default:
		return r.selectRoundRobin(serviceName, healthyInstances), nil
	}
}

// selectRoundRobin implements round-robin load balancing
func (r *ServiceRegistryImpl) selectRoundRobin(serviceName string, instances []*ServiceInstance) *ServiceInstance {
	r.lbMutex.Lock()
	defer r.lbMutex.Unlock()
	
	if r.roundRobinCounters == nil {
		r.roundRobinCounters = make(map[string]int)
	}
	
	counter := r.roundRobinCounters[serviceName]
	selected := instances[counter%len(instances)]
	r.roundRobinCounters[serviceName] = (counter + 1) % len(instances)
	
	return selected
}

// selectRandom implements random load balancing
func (r *ServiceRegistryImpl) selectRandom(instances []*ServiceInstance) *ServiceInstance {
	if len(instances) == 1 {
		return instances[0]
	}
	
	index := r.random.Intn(len(instances))
	return instances[index]
}

// selectWeightedRoundRobin implements weighted round-robin load balancing
func (r *ServiceRegistryImpl) selectWeightedRoundRobin(serviceName string, instances []*ServiceInstance) *ServiceInstance {
	r.lbMutex.Lock()
	defer r.lbMutex.Unlock()
	
	if r.weightedCounters == nil {
		r.weightedCounters = make(map[string]map[string]int)
	}
	
	if r.weightedCounters[serviceName] == nil {
		r.weightedCounters[serviceName] = make(map[string]int)
	}
	
	// Calculate total weight
	totalWeight := 0
	for _, instance := range instances {
		totalWeight += instance.Info.Weight
	}
	
	if totalWeight == 0 {
		// If no weights specified, fall back to round-robin
		return r.selectRoundRobin(serviceName, instances)
	}
	
	// Find the instance with the highest current weight
	var selected *ServiceInstance
	maxCurrentWeight := -1
	
	for _, instance := range instances {
		instanceID := instance.Info.ID
		currentWeight := r.weightedCounters[serviceName][instanceID]
		currentWeight += instance.Info.Weight
		r.weightedCounters[serviceName][instanceID] = currentWeight
		
		if currentWeight > maxCurrentWeight {
			maxCurrentWeight = currentWeight
			selected = instance
		}
	}
	
	// Reduce the selected instance's current weight by total weight
	if selected != nil {
		r.weightedCounters[serviceName][selected.Info.ID] -= totalWeight
	}
	
	return selected
}

// selectLeastConnections implements least connections load balancing
func (r *ServiceRegistryImpl) selectLeastConnections(instances []*ServiceInstance) *ServiceInstance {
	var selected *ServiceInstance
	minConnections := int(^uint(0) >> 1) // Max int
	
	for _, instance := range instances {
		connections := instance.GetConnections()
		if connections < minConnections {
			minConnections = connections
			selected = instance
		}
	}
	
	return selected
}

// GetLoadBalancingStats returns statistics about load balancing
func (r *ServiceRegistryImpl) GetLoadBalancingStats(ctx context.Context, serviceName string) (map[string]interface{}, error) {
	r.mutex.RLock()
	instances, exists := r.servicesByName[serviceName]
	r.mutex.RUnlock()
	
	if !exists {
		return nil, NewRegistryError(ErrCodeServiceNotFound, fmt.Sprintf("service %s not found", serviceName), nil)
	}
	
	stats := make(map[string]interface{})
	stats["total_instances"] = len(instances)
	
	healthyCount := 0
	totalConnections := 0
	instanceStats := make([]map[string]interface{}, 0)
	
	for _, instance := range instances {
		if instance.GetState() == ServiceStateHealthy {
			healthyCount++
		}
		
		connections := instance.GetConnections()
		totalConnections += connections
		
		instanceStats = append(instanceStats, map[string]interface{}{
			"id":          instance.Info.ID,
			"address":     fmt.Sprintf("%s:%d", instance.Info.Address, instance.Info.Port),
			"state":       instance.GetState().String(),
			"weight":      instance.Info.Weight,
			"connections": connections,
		})
	}
	
	stats["healthy_instances"] = healthyCount
	stats["total_connections"] = totalConnections
	stats["instances"] = instanceStats
	
	return stats, nil
}

// Subscribe subscribes to service events
func (r *ServiceRegistryImpl) Subscribe(ctx context.Context, callback func(*ServiceEvent)) error {
	r.subMutex.Lock()
	defer r.subMutex.Unlock()
	
	r.subscribers = append(r.subscribers, callback)
	return nil
}

// Unsubscribe unsubscribes from service events
func (r *ServiceRegistryImpl) Unsubscribe(ctx context.Context, callback func(*ServiceEvent)) error {
	r.subMutex.Lock()
	defer r.subMutex.Unlock()
	
	for i, sub := range r.subscribers {
		if &sub == &callback {
			r.subscribers = append(r.subscribers[:i], r.subscribers[i+1:]...)
			break
		}
	}
	return nil
}

// performHealthCheck performs health check on a service instance
func (r *ServiceRegistryImpl) performHealthCheck(instance *ServiceInstance) error {
	if instance.Info.HealthCheck == nil || !instance.Info.HealthCheck.Enabled {
		return nil
	}
	
	healthCheck := instance.Info.HealthCheck
	
	// Perform health check based on method
	switch strings.ToLower(healthCheck.Method) {
	case "http":
		return r.performHTTPHealthCheck(healthCheck)
	case "tcp":
		return r.performTCPHealthCheck(healthCheck)
	case "grpc":
		return r.performGRPCHealthCheck(healthCheck)
	default:
		return fmt.Errorf("unsupported health check method: %s", healthCheck.Method)
	}
}

// GetServiceStats returns statistics about the registry
func (r *ServiceRegistryImpl) GetServiceStats(ctx context.Context) (*ServiceStats, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	stats := &ServiceStats{
		TotalServices:    len(r.services),
		HealthyServices:  0,
		UnhealthyServices: 0,
		RegistryUptime:   time.Since(r.startTime),
		ServicesByState:  make(map[ServiceState]int),
	}
	
	for _, instance := range r.services {
		state := instance.GetState()
		stats.ServicesByState[state]++
		
		switch state {
		case ServiceStateHealthy:
			stats.HealthyServices++
		case ServiceStateUnhealthy:
			stats.UnhealthyServices++
		}
	}
	
	return stats, nil
}

// Shutdown gracefully shuts down the registry
func (r *ServiceRegistryImpl) Shutdown(ctx context.Context) error {
	// Stop health checker first
	if r.healthChecker != nil {
		r.healthChecker.Stop()
	}
	
	r.cancel()
	
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	// Deregister all services
	for serviceID := range r.services {
		r.Deregister(ctx, serviceID)
	}
	
	r.logger.Info("Service registry shutdown completed")
	return nil
}

// Helper methods

// validateServiceInfo validates service information
func (r *ServiceRegistryImpl) validateServiceInfo(service *ServiceInfo) error {
	if service.ID == "" {
		return fmt.Errorf("service ID cannot be empty")
	}
	if service.Name == "" {
		return fmt.Errorf("service name cannot be empty")
	}
	if service.Address == "" {
		return fmt.Errorf("service address cannot be empty")
	}
	if service.Port <= 0 || service.Port > 65535 {
		return fmt.Errorf("service port must be between 1 and 65535")
	}
	if service.Weight < 0 {
		return fmt.Errorf("service weight cannot be negative")
	}
	return nil
}

// matchesQuery checks if a service instance matches the query
func (r *ServiceRegistryImpl) matchesQuery(instance *ServiceInstance, query *ServiceQuery) bool {
	// Check service name (support both ServiceName and Name fields)
	if query.ServiceName != "" && instance.Info.Name != query.ServiceName {
		return false
	}
	if query.Name != "" && instance.Info.Name != query.Name {
		return false
	}
	
	// Check tags
	if len(query.Tags) > 0 {
		for _, tag := range query.Tags {
			found := false
			for _, serviceTag := range instance.Info.Tags {
				if serviceTag == tag {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	
	// Check metadata
	if len(query.Metadata) > 0 {
		for key, value := range query.Metadata {
			if instance.Info.Metadata[key] != value {
				return false
			}
		}
	}
	
	// Check healthy status
	if query.Healthy != nil {
		isHealthy := instance.GetState() == ServiceStateHealthy
		if *query.Healthy != isHealthy {
			return false
		}
	}
	
	return true
}

// publishEvent publishes an event to all subscribers
func (r *ServiceRegistryImpl) publishEvent(event *ServiceEvent) {
	r.subMutex.RLock()
	defer r.subMutex.RUnlock()
	
	for _, callback := range r.subscribers {
		go func(cb func(*ServiceEvent)) {
			defer func() {
				if rec := recover(); rec != nil {
					r.logger.Error("Event callback panic", "error", rec)
				}
			}()
			cb(event)
		}(callback)
	}
}

// NotifyServiceChange notifies all subscribers about a service change
func (r *ServiceRegistryImpl) NotifyServiceChange(ctx context.Context, serviceID string, eventType string, details map[string]interface{}) error {
	r.mutex.RLock()
	instance, exists := r.services[serviceID]
	r.mutex.RUnlock()
	
	if !exists {
		return NewRegistryError(ErrCodeServiceNotFound, fmt.Sprintf("service with ID %s not found", serviceID), nil)
	}
	
	event := &ServiceEvent{
		Type:      eventType,
		Service:   instance.Info,
		Timestamp: time.Now(),
		Details:   details,
	}
	
	r.publishEvent(event)
	return nil
}

// GetEventHistory returns the recent event history
func (r *ServiceRegistryImpl) GetEventHistory(ctx context.Context, limit int) ([]ServiceEvent, error) {
	if limit <= 0 {
		limit = 100 // Default limit
	}
	
	// Return a copy of recent events (this would require maintaining an event history)
	// For now, return empty slice as we haven't implemented event storage
	return []ServiceEvent{}, nil
}

// SubscribeWithFilter creates a subscription for service events with filtering
func (r *ServiceRegistryImpl) SubscribeWithFilter(ctx context.Context, filter ServiceEventFilter, callback func(*ServiceEvent)) error {
	wrappedCallback := func(event *ServiceEvent) {
		if r.matchesEventFilter(*event, filter) {
			callback(event)
		}
	}
	
	r.subMutex.Lock()
	r.subscribers = append(r.subscribers, wrappedCallback)
	r.subMutex.Unlock()
	
	return nil
}

// UnsubscribeAll removes all event subscribers
func (r *ServiceRegistryImpl) UnsubscribeAll(ctx context.Context) error {
	r.subMutex.Lock()
	r.subscribers = make([]func(*ServiceEvent), 0)
	r.subMutex.Unlock()
	
	return nil
}

// matchesEventFilter checks if an event matches the given filter
func (r *ServiceRegistryImpl) matchesEventFilter(event ServiceEvent, filter ServiceEventFilter) bool {
	if filter.ServiceName != "" && event.Service.Name != filter.ServiceName {
		return false
	}
	
	if filter.EventType != "" && event.Type != filter.EventType {
		return false
	}
	
	return true
}

// broadcastServiceStateChange broadcasts service state changes to subscribers
func (r *ServiceRegistryImpl) broadcastServiceStateChange(serviceID string, oldState, newState ServiceState) {
	r.mutex.RLock()
	instance, exists := r.services[serviceID]
	r.mutex.RUnlock()
	
	if !exists {
		return
	}
	
	event := &ServiceEvent{
		Type:      "service.state.changed",
		Service:   instance.Info,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"old_state": oldState.String(),
			"new_state": newState.String(),
			"service_id": serviceID,
		},
	}
	
	r.publishEvent(event)
}





// Dependency management methods

// CheckDependencies checks if all dependencies for a service are available and healthy
func (r *ServiceRegistryImpl) CheckDependencies(ctx context.Context, serviceID string) error {
	r.mutex.RLock()
	instance, exists := r.services[serviceID]
	r.mutex.RUnlock()
	
	if !exists {
		return NewRegistryError(ErrCodeServiceNotFound, fmt.Sprintf("service with ID %s not found", serviceID), nil)
	}
	
	if len(instance.Info.Dependencies) == 0 {
		return nil // No dependencies to check
	}
	
	for _, dep := range instance.Info.Dependencies {
		if err := r.checkSingleDependency(ctx, dep); err != nil {
			return NewRegistryError(ErrCodeDependencyNotMet, fmt.Sprintf("dependency %s not met for service %s", dep, serviceID), err)
		}
	}
	
	return nil
}

// ResolveDependencies resolves the dependency graph and returns services in dependency order
func (r *ServiceRegistryImpl) ResolveDependencies(ctx context.Context, serviceNames []string) ([]string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	// Build dependency graph
	graph := make(map[string][]string)
	inDegree := make(map[string]int)
	
	// Initialize graph
	for _, serviceName := range serviceNames {
		graph[serviceName] = make([]string, 0)
		inDegree[serviceName] = 0
	}
	
	// Build edges based on dependencies
	for _, serviceName := range serviceNames {
		instances := r.servicesByName[serviceName]
		if len(instances) == 0 {
			continue
		}
		
		// Use the first instance's dependencies
		for _, dep := range instances[0].Info.Dependencies {
			if _, exists := inDegree[dep]; exists {
				graph[dep] = append(graph[dep], serviceName)
				inDegree[serviceName]++
			}
		}
	}
	
	// Topological sort using Kahn's algorithm
	queue := make([]string, 0)
	result := make([]string, 0)
	
	// Find all nodes with no incoming edges
	for service, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, service)
		}
	}
	
	// Process queue
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)
		
		// Remove edges from current node
		for _, neighbor := range graph[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}
	
	// Check for circular dependencies
	if len(result) != len(serviceNames) {
		return nil, NewRegistryError(ErrCodeDependencyNotMet, "circular dependency detected", nil)
	}
	
	return result, nil
}

// GetDependencyGraph returns the dependency graph for visualization
func (r *ServiceRegistryImpl) GetDependencyGraph(ctx context.Context) (map[string][]string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	graph := make(map[string][]string)
	
	for _, instance := range r.services {
		serviceName := instance.Info.Name
		if graph[serviceName] == nil {
			graph[serviceName] = make([]string, 0)
		}
		
		for _, dep := range instance.Info.Dependencies {
			graph[serviceName] = append(graph[serviceName], dep)
		}
	}
	
	return graph, nil
}

// ValidateDependencyGraph validates the entire dependency graph for circular dependencies
func (r *ServiceRegistryImpl) ValidateDependencyGraph(ctx context.Context) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	// Get all unique service names
	serviceNames := make([]string, 0)
	seenNames := make(map[string]bool)
	
	for _, instance := range r.services {
		if !seenNames[instance.Info.Name] {
			serviceNames = append(serviceNames, instance.Info.Name)
			seenNames[instance.Info.Name] = true
		}
	}
	
	// Try to resolve dependencies
	_, err := r.ResolveDependencies(ctx, serviceNames)
	return err
}

// checkSingleDependency checks if a single dependency is available and healthy
func (r *ServiceRegistryImpl) checkSingleDependency(ctx context.Context, dependencyName string) error {
	instances, exists := r.servicesByName[dependencyName]
	if !exists || len(instances) == 0 {
		return fmt.Errorf("dependency service %s not found", dependencyName)
	}
	
	// Check if at least one instance is healthy
	healthyFound := false
	for _, instance := range instances {
		if instance.GetState() == ServiceStateHealthy {
			healthyFound = true
			break
		}
	}
	
	if !healthyFound {
		return fmt.Errorf("no healthy instances found for dependency service %s", dependencyName)
	}
	
	return nil
}

// Deregister deregisters a service
func (r *ServiceRegistryImpl) Deregister(ctx context.Context, serviceID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	instance, exists := r.services[serviceID]
	if !exists {
		return NewRegistryError(ErrCodeServiceNotFound, fmt.Sprintf("service with ID %s not found", serviceID), nil)
	}
	
	// Set state to deregistering
	instance.SetState(ServiceStateDeregistering)
	
	// Remove from services map
	delete(r.services, serviceID)
	
	// Remove from services by name map
	serviceName := instance.Info.Name
	if instances, exists := r.servicesByName[serviceName]; exists {
		for i, inst := range instances {
			if inst.Info.ID == serviceID {
				r.servicesByName[serviceName] = append(instances[:i], instances[i+1:]...)
				break
			}
		}
		// Remove empty slice
		if len(r.servicesByName[serviceName]) == 0 {
			delete(r.servicesByName, serviceName)
		}
	}
	
	// Set final state
	instance.SetState(ServiceStateDeregistered)
	
	r.logger.Info("Service deregistered successfully",
		"service_id", serviceID,
		"service_name", serviceName)
	
	// Publish deregistration event
	r.publishEvent(&ServiceEvent{
		Type:      ServiceEventDeregistered,
		Service:   instance.Info,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf("Service %s deregistered", serviceName),
	})
	
	return nil
}

// Discover discovers services by name
func (r *ServiceRegistryImpl) Discover(ctx context.Context, serviceName string) ([]*ServiceInstance, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	instances, exists := r.servicesByName[serviceName]
	if !exists {
		return []*ServiceInstance{}, nil
	}
	
	// Filter healthy services
	healthyInstances := make([]*ServiceInstance, 0)
	for _, instance := range instances {
		if instance.GetState() == ServiceStateHealthy {
			healthyInstances = append(healthyInstances, instance)
		}
	}
	
	return healthyInstances, nil
}

// Query queries services with filters
func (r *ServiceRegistryImpl) Query(ctx context.Context, query *ServiceQuery) ([]*ServiceInstance, error) {
	if query == nil {
		return r.ListServices(ctx)
	}
	
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	result := make([]*ServiceInstance, 0)
	
	for _, instance := range r.services {
		if r.matchesQuery(instance, query) {
			result = append(result, instance)
		}
	}
	
	return result, nil
}

// GetService gets a specific service by ID
func (r *ServiceRegistryImpl) GetService(ctx context.Context, serviceID string) (*ServiceInstance, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	instance, exists := r.services[serviceID]
	if !exists {
		return nil, NewRegistryError(ErrCodeServiceNotFound, fmt.Sprintf("service with ID %s not found", serviceID), nil)
	}
	
	return instance, nil
}

// ListServices lists all registered services
func (r *ServiceRegistryImpl) ListServices(ctx context.Context) ([]*ServiceInstance, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	instances := make([]*ServiceInstance, 0, len(r.services))
	for _, instance := range r.services {
		instances = append(instances, instance)
	}
	
	return instances, nil
}

// UpdateService updates service information
func (r *ServiceRegistryImpl) UpdateService(ctx context.Context, serviceID string, updates map[string]interface{}) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	instance, exists := r.services[serviceID]
	if !exists {
		return NewRegistryError(ErrCodeServiceNotFound, fmt.Sprintf("service with ID %s not found", serviceID), nil)
	}
	
	// Apply updates
	for key, value := range updates {
		switch key {
		case "address":
			if addr, ok := value.(string); ok {
				instance.Info.Address = addr
			}
		case "port":
			if port, ok := value.(int); ok {
				instance.Info.Port = port
			}
		case "tags":
			if tags, ok := value.([]string); ok {
				instance.Info.Tags = tags
			}
		case "metadata":
			if metadata, ok := value.(map[string]string); ok {
				instance.Info.Metadata = metadata
			}
		case "weight":
			if weight, ok := value.(int); ok {
				instance.Info.Weight = weight
			}
		}
	}
	
	instance.Info.UpdatedAt = time.Now()
	
	r.logger.Info("Service updated successfully",
		"service_id", serviceID,
		"updates", updates)
	
	// Publish update event
	r.publishEvent(&ServiceEvent{
		Type:      ServiceEventUpdated,
		Service:   instance.Info,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf("Service %s updated", instance.Info.Name),
	})
	
	return nil
}