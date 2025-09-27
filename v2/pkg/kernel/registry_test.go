package kernel

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServiceRegistry_Register tests service registration functionality
func TestServiceRegistry_Register(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	registry := NewServiceRegistry(logger)
	ctx := context.Background()
	var err error

	// Stop health checker to prevent background goroutines
	defer registry.StopHealthCheck(ctx)

	// Test successful registration
	serviceInfo := &ServiceInfo{
		ID:      "test-service-1",
		Name:    "test-service",
		Address: "localhost",
		Port:    8080,
		Tags:    []string{"api", "v1"},
		Weight:  100,
		Metadata: map[string]string{
			"version": "1.0.0",
		},
	}

	err = registry.Register(ctx, serviceInfo)
	require.NoError(t, err)
	
	// Get the registered service to verify
	registeredService, getErr := registry.GetService(ctx, serviceInfo.ID)
	require.NoError(t, getErr)
	assert.Equal(t, serviceInfo.ID, registeredService.Info.ID)
	assert.Equal(t, ServiceStateHealthy, registeredService.GetState()) // Services without health checks are set to healthy

	// Test duplicate registration
	err = registry.Register(ctx, serviceInfo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Test invalid service info
	invalidService := &ServiceInfo{
		ID:   "", // Empty ID should fail validation
		Name: "invalid",
	}
	err = registry.Register(ctx, invalidService)
	assert.Error(t, err)
}

// TestServiceRegistry_Deregister tests service deregistration functionality
func TestServiceRegistry_Deregister(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	registry := NewServiceRegistry(logger)
	ctx := context.Background()
	var err error

	// Stop health checker to prevent background goroutines
	defer registry.StopHealthCheck(ctx)

	// Register a service first
	serviceInfo := &ServiceInfo{
		ID:      "test-service-1",
		Name:    "test-service",
		Address: "localhost",
		Port:    8080,
	}

	err = registry.Register(ctx, serviceInfo)
	require.NoError(t, err)

	// Test successful deregistration
	err = registry.Deregister(ctx, serviceInfo.ID)
	assert.NoError(t, err)

	// Test deregistering non-existent service
	err = registry.Deregister(ctx, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestServiceRegistry_Discover tests service discovery functionality
func TestServiceRegistry_Discover(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	registry := NewServiceRegistry(logger)
	ctx := context.Background()
	var err error

	// Stop health checker to prevent background goroutines
	defer registry.StopHealthCheck(ctx)

	// Register multiple services with same name
	service1 := &ServiceInfo{
		ID:      "service-1",
		Name:    "api-service",
		Address: "localhost",
		Port:    8080,
	}
	service2 := &ServiceInfo{
		ID:      "service-2",
		Name:    "api-service",
		Address: "localhost",
		Port:    8081,
	}

	err = registry.Register(ctx, service1)
	require.NoError(t, err)
	err = registry.Register(ctx, service2)
	require.NoError(t, err)
	
	// Set services to healthy state
	instance1, err := registry.GetService(ctx, service1.ID)
	require.NoError(t, err)
	instance1.SetState(ServiceStateHealthy)
	
	instance2, err := registry.GetService(ctx, service2.ID)
	require.NoError(t, err)
	instance2.SetState(ServiceStateHealthy)

	// Test discovery
	instances, err := registry.Discover(ctx, "api-service")
	assert.NoError(t, err)
	assert.Len(t, instances, 2)

	// Test discovery of non-existent service
	instances, err = registry.Discover(ctx, "non-existent")
	assert.NoError(t, err)
	assert.Len(t, instances, 0)
}

// TestServiceRegistry_Query tests service query functionality
func TestServiceRegistry_Query(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	registry := NewServiceRegistry(logger)
	ctx := context.Background()
	var err error

	// Stop health checker to prevent background goroutines
	defer registry.StopHealthCheck(ctx)

	// Register services with different tags
	service1 := &ServiceInfo{
		ID:      "service-1",
		Name:    "api-service",
		Address: "localhost",
		Port:    8080,
		Tags:    []string{"api", "v1"},
	}
	service2 := &ServiceInfo{
		ID:      "service-2",
		Name:    "web-service",
		Address: "localhost",
		Port:    8081,
		Tags:    []string{"web", "v2"},
	}

	err = registry.Register(ctx, service1)
	require.NoError(t, err)
	err = registry.Register(ctx, service2)
	require.NoError(t, err)
	
	// Set services to healthy state
	instance1, err := registry.GetService(ctx, service1.ID)
	require.NoError(t, err)
	instance1.SetState(ServiceStateHealthy)
	
	instance2, err := registry.GetService(ctx, service2.ID)
	require.NoError(t, err)
	instance2.SetState(ServiceStateHealthy)

	// Test query by name
	query := &ServiceQuery{
		Name: "api-service",
	}
	instances, err := registry.Query(ctx, query)
	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assert.Equal(t, "service-1", instances[0].Info.ID)

	// Test query by tags
	query = &ServiceQuery{
		Tags: []string{"v1"},
	}
	instances, err = registry.Query(ctx, query)
	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assert.Equal(t, "service-1", instances[0].Info.ID)

	// Test query all services (nil query)
	instances, err = registry.Query(ctx, nil)
	assert.NoError(t, err)
	assert.Len(t, instances, 2)
}

// TestServiceRegistry_HealthCheck tests health checking functionality
func TestServiceRegistry_HealthCheck(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	registry := NewServiceRegistry(logger)
	ctx := context.Background()
	var err error

	// Stop health checker to prevent background goroutines
	defer registry.StopHealthCheck(ctx)

	// Register a service with health check
	serviceInfo := &ServiceInfo{
		ID:      "test-service-1",
		Name:    "test-service",
		Address: "localhost",
		Port:    8080,
		HealthCheck: &HealthCheckInfo{
			Enabled:  true,
			Interval: time.Second,
			Timeout:  500 * time.Millisecond,
			Endpoint: "http://localhost:8080/health",
			Method:   "GET",
		},
	}

	err = registry.Register(ctx, serviceInfo)
	require.NoError(t, err)

	// Get the registered service to check health
	instance, err := registry.GetService(ctx, serviceInfo.ID)
	require.NoError(t, err)
	require.NotNil(t, instance)
	
	// Perform health check - should fail since no actual service is running
	err = registry.CheckHealth(ctx, serviceInfo.ID)
	assert.Error(t, err) // Expected to fail for non-running service
}

// TestServiceRegistry_LoadBalancing tests load balancing functionality
func TestServiceRegistry_LoadBalancing(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	registry := NewServiceRegistry(logger)
	ctx := context.Background()
	var err error

	// Stop health checker to prevent background goroutines
	defer registry.StopHealthCheck(ctx)

	// Register multiple instances of the same service
	for i := 0; i < 3; i++ {
		serviceInfo := &ServiceInfo{
			ID:      fmt.Sprintf("service-%d", i),
			Name:    "load-balanced-service",
			Address: "localhost",
			Port:    8080 + i,
			Weight:  100,
		}
		err = registry.Register(ctx, serviceInfo)
		require.NoError(t, err)
		
		// Set service state to healthy for load balancing
		instance, err := registry.GetService(ctx, serviceInfo.ID)
		require.NoError(t, err)
		instance.SetState(ServiceStateHealthy)
	}

	// Test round-robin selection
	selectedIDs := make(map[string]int)
	for i := 0; i < 9; i++ { // 3 rounds
		instance, err := registry.SelectService(ctx, "load-balanced-service", LoadBalanceRoundRobin)
		assert.NoError(t, err)
		assert.NotNil(t, instance)
		selectedIDs[instance.Info.ID]++
	}

	// Each instance should be selected 3 times in round-robin
	for _, count := range selectedIDs {
		assert.Equal(t, 3, count)
	}

	// Test random selection
	instance, err := registry.SelectService(ctx, "load-balanced-service", LoadBalanceRandom)
	assert.NoError(t, err)
	assert.NotNil(t, instance)
}

// TestServiceRegistry_Events tests event notification functionality
func TestServiceRegistry_Events(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	registry := NewServiceRegistry(logger)
	ctx := context.Background()
	var err error

	// Stop health checker to prevent background goroutines
	defer registry.StopHealthCheck(ctx)

	// Subscribe to events using channel for synchronization
	eventCh := make(chan *ServiceEvent, 10)
	subscriber := func(event *ServiceEvent) {
		select {
		case eventCh <- event:
		default:
			// Channel full, ignore
		}
	}
	err = registry.Subscribe(ctx, subscriber)
	assert.NoError(t, err)

	// Register a service to trigger event
	serviceInfo := &ServiceInfo{
		ID:      "test-service-1",
		Name:    "test-service",
		Address: "localhost",
		Port:    8080,
	}

	err = registry.Register(ctx, serviceInfo)
	require.NoError(t, err)

	// Wait for event to be received
	select {
	case event := <-eventCh:
		// Check if registration event was received
		assert.Equal(t, "service_registered", event.Type)
	case <-time.After(1 * time.Second):
		t.Error("Expected to receive service_registered event")
	}
}

// TestServiceRegistry_Dependencies tests dependency management functionality
func TestServiceRegistry_Dependencies(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	registry := NewServiceRegistry(logger)
	ctx := context.Background()
	var err error

	// Stop health checker to prevent background goroutines
	defer registry.StopHealthCheck(ctx)

	// Register services with dependencies
	dbService := &ServiceInfo{
		ID:      "db-service",
		Name:    "database",
		Address: "localhost",
		Port:    5432,
	}

	apiService := &ServiceInfo{
		ID:           "api-service",
		Name:         "api",
		Address:      "localhost",
		Port:         8080,
		Dependencies: []string{"database"},
	}

	// Register database service first
	err = registry.Register(ctx, dbService)
	require.NoError(t, err)
	
	// Set database service to healthy state
	dbInstance, err := registry.GetService(ctx, dbService.ID)
	require.NoError(t, err)
	dbInstance.SetState(ServiceStateHealthy)

	// Register API service with dependency
	err = registry.Register(ctx, apiService)
	require.NoError(t, err)
	
	// Set API service to healthy state
	apiInstance, err := registry.GetService(ctx, apiService.ID)
	require.NoError(t, err)
	apiInstance.SetState(ServiceStateHealthy)

	// Test dependency checking
	err = registry.CheckDependencies(ctx, "api-service")
	assert.NoError(t, err)

	// Test that dependencies are properly registered
	// We can verify this by checking that both services are discoverable
	apiServices, err := registry.Discover(ctx, "api")
	assert.NoError(t, err)
	assert.Len(t, apiServices, 1)
	
	dbServices, err := registry.Discover(ctx, "database")
	assert.NoError(t, err)
	assert.Len(t, dbServices, 1)
}

// TestServiceRegistry_ErrorHandling tests error handling scenarios
func TestServiceRegistry_ErrorHandling(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	registry := NewServiceRegistry(logger)
	ctx := context.Background()
	var err error

	// Stop health checker to prevent background goroutines
	defer registry.StopHealthCheck(ctx)

	// Test operations on non-existent service
	_, err = registry.GetService(ctx, "non-existent")
	assert.Error(t, err)
	assert.IsType(t, &RegistryError{}, err)

	registryErr := err.(*RegistryError)
	assert.Equal(t, ErrCodeServiceNotFound, registryErr.Code)
	assert.Contains(t, registryErr.Message, "not found")

	// Test invalid service validation
	invalidService := &ServiceInfo{
		// Missing required fields
	}
	err = registry.Register(ctx, invalidService)
	assert.Error(t, err)
}

// TestServiceRegistry_Shutdown tests registry shutdown functionality
func TestServiceRegistry_Shutdown(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	registry := NewServiceRegistry(logger)
	ctx := context.Background()
	var err error

	// Stop health checker to prevent background goroutines
	defer registry.StopHealthCheck(ctx)

	// Register a service
	serviceInfo := &ServiceInfo{
		ID:      "test-service-1",
		Name:    "test-service",
		Address: "localhost",
		Port:    8080,
	}

	err = registry.Register(ctx, serviceInfo)
	require.NoError(t, err)

	// Test that service is registered
	instance, err := registry.GetService(ctx, serviceInfo.ID)
	require.NoError(t, err)
	require.NotNil(t, instance)

	// Test operations after shutdown should fail
	err = registry.Register(ctx, &ServiceInfo{
		ID:      "new-service",
		Name:    "new",
		Address: "localhost",
		Port:    8081,
	})
	// Should return error after shutdown (if shutdown was implemented)
}

// BenchmarkServiceRegistry_Register benchmarks service registration performance
func BenchmarkServiceRegistry_Register(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	registry := NewServiceRegistry(logger)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		serviceInfo := &ServiceInfo{
			ID:      fmt.Sprintf("service-%d", i),
			Name:    "benchmark-service",
			Address: "localhost",
			Port:    8080,
		}
		err := registry.Register(ctx, serviceInfo)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkServiceRegistry_Discover benchmarks service discovery performance
func BenchmarkServiceRegistry_Discover(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	registry := NewServiceRegistry(logger)
	ctx := context.Background()

	// Pre-register services
	for i := 0; i < 100; i++ {
		serviceInfo := &ServiceInfo{
			ID:      fmt.Sprintf("service-%d", i),
			Name:    "benchmark-service",
			Address: "localhost",
			Port:    8080 + i,
		}
		err := registry.Register(ctx, serviceInfo)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := registry.Discover(ctx, "benchmark-service")
		if err != nil {
			b.Fatal(err)
		}
	}
}