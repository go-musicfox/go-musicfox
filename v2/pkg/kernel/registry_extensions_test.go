// Package kernel provides tests for service registry extensions
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

// Test helper functions

func createTestLogger() Logger {
	return NewSlogAdapter(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))
}

func createTestServiceInfo(id, name, version string) *ServiceInfo {
	return &ServiceInfo{
		ID:      id,
		Name:    name,
		Version: version,
		Address: "localhost",
		Port:    8080,
		Tags:    []string{"test"},
		Metadata: map[string]string{
			"zone": "us-east-1",
		},
		Weight:       1,
		RegisteredAt: time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func createTestServiceInstance(id, name, version string) *ServiceInstance {
	return &ServiceInstance{
		Info:        createTestServiceInfo(id, name, version),
		State:       ServiceStateHealthy,
		LastHealthy: time.Now(),
		LastSeen:    time.Now(),
	}
}

// Version Manager Tests

func TestVersionManager_ParseVersion(t *testing.T) {
	tests := []struct {
		name        string
		versionStr  string
		expected    *ServiceVersion
		expectError bool
	}{
		{
			name:       "valid semantic version",
			versionStr: "1.2.3",
			expected:   &ServiceVersion{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:       "version with pre-release",
			versionStr: "1.2.3-alpha",
			expected:   &ServiceVersion{Major: 1, Minor: 2, Patch: 3, Pre: "alpha"},
		},
		{
			name:       "version with build metadata",
			versionStr: "1.2.3+build.1",
			expected:   &ServiceVersion{Major: 1, Minor: 2, Patch: 3, Build: "build.1"},
		},
		{
			name:       "full version",
			versionStr: "1.2.3-alpha+build.1",
			expected:   &ServiceVersion{Major: 1, Minor: 2, Patch: 3, Pre: "alpha", Build: "build.1"},
		},
		{
			name:        "invalid version",
			versionStr:  "invalid",
			expectError: true,
		},
		{
			name:        "empty version",
			versionStr:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := ParseVersion(tt.versionStr)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, version)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, version)
			}
		})
	}
}

func TestVersionManager_RegisterServiceVersion(t *testing.T) {
	logger := createTestLogger()
	vm := NewVersionManager(logger)
	
	version, err := ParseVersion("1.0.0")
	require.NoError(t, err)
	
	instance := createTestServiceInstance("service-1", "test-service", "1.0.0")
	
	err = vm.RegisterServiceVersion("test-service", version, instance)
	assert.NoError(t, err)
	
	// Verify version was registered
	versions, err := vm.ListServiceVersions("test-service")
	assert.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, version, versions[0])
}

func TestVersionManager_GetCompatibleServices(t *testing.T) {
	logger := createTestLogger()
	vm := NewVersionManager(logger)
	
	// Register multiple versions
	v100, _ := ParseVersion("1.0.0")
	v101, _ := ParseVersion("1.0.1")
	v200, _ := ParseVersion("2.0.0")
	
	instance100 := createTestServiceInstance("service-1", "test-service", "1.0.0")
	instance101 := createTestServiceInstance("service-2", "test-service", "1.0.1")
	instance200 := createTestServiceInstance("service-3", "test-service", "2.0.0")
	
	vm.RegisterServiceVersion("test-service", v100, instance100)
	vm.RegisterServiceVersion("test-service", v101, instance101)
	vm.RegisterServiceVersion("test-service", v200, instance200)
	
	// Test compatibility with 1.0.0 (should get 1.0.1 and 2.0.0)
	compatible, err := vm.GetCompatibleServices("test-service", v100)
	assert.NoError(t, err)
	assert.Len(t, compatible, 2) // 1.0.1 and 2.0.0 are compatible
}

func TestVersionManager_DeprecateServiceVersion(t *testing.T) {
	logger := createTestLogger()
	vm := NewVersionManager(logger)
	
	version, _ := ParseVersion("1.0.0")
	replacement, _ := ParseVersion("2.0.0")
	instance := createTestServiceInstance("service-1", "test-service", "1.0.0")
	
	vm.RegisterServiceVersion("test-service", version, instance)
	
	deprecationTime := time.Now()
	err := vm.DeprecateServiceVersion("test-service", version, deprecationTime, "Outdated version", replacement)
	assert.NoError(t, err)
	
	// Verify deprecation
	deprecated, err := vm.GetDeprecatedVersions()
	assert.NoError(t, err)
	assert.Len(t, deprecated, 1)
	assert.Equal(t, "Outdated version", deprecated[0].Reason)
	assert.Equal(t, replacement, deprecated[0].Replacement)
}

// Metrics Manager Tests

func TestMetricsManager_RecordServiceCall(t *testing.T) {
	logger := createTestLogger()
	mm := NewMetricsManager(logger)
	
	serviceID := "test-service-1"
	duration := 100 * time.Millisecond
	
	// Record successful call
	err := mm.RecordServiceCall(serviceID, duration, true, "")
	assert.NoError(t, err)
	
	// Get metrics
	metrics, err := mm.GetServiceMetrics(serviceID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), metrics.TotalRequests)
	assert.Equal(t, int64(1), metrics.SuccessfulRequests)
	assert.Equal(t, int64(0), metrics.FailedRequests)
	assert.Equal(t, float64(0), metrics.ErrorRate)
	
	// Record failed call
	err = mm.RecordServiceCall(serviceID, duration, false, "timeout")
	assert.NoError(t, err)
	
	// Check updated metrics
	metrics, err = mm.GetServiceMetrics(serviceID)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), metrics.TotalRequests)
	assert.Equal(t, int64(1), metrics.SuccessfulRequests)
	assert.Equal(t, int64(1), metrics.FailedRequests)
	assert.Equal(t, float64(50), metrics.ErrorRate)
}

func TestMetricsManager_CreateAlert(t *testing.T) {
	logger := createTestLogger()
	mm := NewMetricsManager(logger)
	
	alert := &ServiceAlert{
		ServiceID:   "test-service-1",
		ServiceName: "test-service",
		Type:        ServiceAlertTypePerformance,
		Severity:    ServiceAlertSeverityHigh,
		Message:     "High response time detected",
		Details: map[string]interface{}{
			"response_time": "500ms",
		},
	}
	
	err := mm.CreateAlert(alert)
	assert.NoError(t, err)
	assert.NotEmpty(t, alert.ID)
	assert.Equal(t, ServiceAlertStatusActive, alert.Status)
	
	// Get active alerts
	activeAlerts, err := mm.GetActiveAlerts("test-service-1")
	assert.NoError(t, err)
	assert.Len(t, activeAlerts, 1)
	assert.Equal(t, alert.ID, activeAlerts[0].ID)
}

func TestMetricsManager_ResolveAlert(t *testing.T) {
	logger := createTestLogger()
	mm := NewMetricsManager(logger)
	
	alert := &ServiceAlert{
		ServiceID:   "test-service-1",
		ServiceName: "test-service",
		Type:        ServiceAlertTypePerformance,
		Severity:    ServiceAlertSeverityHigh,
		Message:     "High response time detected",
	}
	
	mm.CreateAlert(alert)
	
	// Resolve alert
	err := mm.ResolveAlert(alert.ID)
	assert.NoError(t, err)
	
	// Verify alert is resolved
	activeAlerts, err := mm.GetActiveAlerts("test-service-1")
	assert.NoError(t, err)
	assert.Len(t, activeAlerts, 0)
}

func TestMetricsManager_SetAlertThresholds(t *testing.T) {
	logger := createTestLogger()
	mm := NewMetricsManager(logger)
	
	thresholds := &AlertThresholds{
		MaxResponseTime: 200 * time.Millisecond,
		MaxErrorRate:    5.0,
		MinThroughput:   100.0,
		Enabled:         true,
	}
	
	err := mm.SetAlertThresholds("test-service-1", thresholds)
	assert.NoError(t, err)
	
	// Record a call that exceeds threshold
	err = mm.RecordServiceCall("test-service-1", 300*time.Millisecond, true, "")
	assert.NoError(t, err)
	
	// Should have created an alert
	time.Sleep(10 * time.Millisecond) // Allow goroutine to process
	activeAlerts, err := mm.GetActiveAlerts("test-service-1")
	assert.NoError(t, err)
	assert.Len(t, activeAlerts, 1)
	assert.Equal(t, ServiceAlertTypePerformance, activeAlerts[0].Type)
}

// Failover Manager Tests

func TestFailoverManager_ConfigureFailover(t *testing.T) {
	logger := createTestLogger()
	registry := NewServiceRegistry(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	metrics := NewMetricsManager(logger)
	fm := NewFailoverManager(logger, registry, metrics)
	
	config := &ServiceFailoverConfig{
		Enabled:               true,
		MaxRetries:           3,
		RetryDelay:           1 * time.Second,
		BackoffMultiplier:    2.0,
		MaxRetryDelay:        10 * time.Second,
		CircuitBreakerEnabled: true,
		FailureThreshold:     5,
		RecoveryTimeout:      30 * time.Second,
	}
	
	err := fm.ConfigureFailover("test-service-1", config)
	assert.NoError(t, err)
	
	// Verify configuration
	retrievedConfig, err := fm.GetFailoverConfig("test-service-1")
	assert.NoError(t, err)
	assert.Equal(t, config.MaxRetries, retrievedConfig.MaxRetries)
	assert.Equal(t, config.CircuitBreakerEnabled, retrievedConfig.CircuitBreakerEnabled)
	
	// Verify circuit breaker was created
	cb, err := fm.GetCircuitBreaker("test-service-1")
	assert.NoError(t, err)
	assert.Equal(t, CircuitBreakerStateClosed, cb.State)
	assert.Equal(t, config.FailureThreshold, cb.FailureThreshold)
}

func TestFailoverManager_CircuitBreaker(t *testing.T) {
	logger := createTestLogger()
	registry := NewServiceRegistry(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	metrics := NewMetricsManager(logger)
	fm := NewFailoverManager(logger, registry, metrics)
	
	config := &ServiceFailoverConfig{
		Enabled:               true,
		CircuitBreakerEnabled: true,
		FailureThreshold:     3,
		RecoveryTimeout:      1 * time.Second,
	}
	
	fm.ConfigureFailover("test-service-1", config)
	
	// Initially should allow calls
	canCall, err := fm.CanMakeCall("test-service-1")
	assert.NoError(t, err)
	assert.True(t, canCall)
	
	// Record failures to trigger circuit breaker
	for i := 0; i < 3; i++ {
		err = fm.RecordServiceCall("test-service-1", false, 100*time.Millisecond)
		assert.NoError(t, err)
	}
	
	// Circuit should be open now
	cb, err := fm.GetCircuitBreaker("test-service-1")
	assert.NoError(t, err)
	assert.Equal(t, CircuitBreakerStateOpen, cb.State)
	
	// Should not allow calls
	canCall, err = fm.CanMakeCall("test-service-1")
	assert.Error(t, err)
	assert.False(t, canCall)
}

func TestFailoverManager_AttemptFailover(t *testing.T) {
	logger := createTestLogger()
	registry := NewServiceRegistry(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	metrics := NewMetricsManager(logger)
	fm := NewFailoverManager(logger, registry, metrics)
	
	ctx := context.Background()
	
	// Register multiple instances of the same service
	service1 := createTestServiceInfo("service-1", "test-service", "1.0.0")
	service2 := createTestServiceInfo("service-2", "test-service", "1.0.0")
	
	registry.Register(ctx, service1)
	registry.Register(ctx, service2)
	
	// Attempt failover from service-1
	result, err := fm.AttemptFailover(ctx, "test-service", "service-1")
	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "service-2", result.FailoverService.Info.ID)
	assert.NotNil(t, result.Metadata)
}

// Extended Load Balancer Tests

func TestExtendedLoadBalancer_ConsistentHash(t *testing.T) {
	logger := createTestLogger()
	elb := NewExtendedLoadBalancer(logger)
	
	config := &LoadBalancerConfig{
		Strategy:     ExtendedLoadBalanceConsistentHash,
		VirtualNodes: 150,
	}
	
	err := elb.ConfigureLoadBalancing("test-service", config)
	assert.NoError(t, err)
	
	instances := []*ServiceInstance{
		createTestServiceInstance("service-1", "test-service", "1.0.0"),
		createTestServiceInstance("service-2", "test-service", "1.0.0"),
		createTestServiceInstance("service-3", "test-service", "1.0.0"),
	}
	
	lbContext := map[string]interface{}{
			"session_id": "user123",
		}
	
		// Same session should always get the same instance
		instance1, err := elb.SelectServiceWithExtendedStrategy(context.Background(), "test-service", ExtendedLoadBalanceConsistentHash, lbContext, instances)
		assert.NoError(t, err)
	
		instance2, err := elb.SelectServiceWithExtendedStrategy(context.Background(), "test-service", ExtendedLoadBalanceConsistentHash, lbContext, instances)
	assert.NoError(t, err)
	
	assert.Equal(t, instance1.Info.ID, instance2.Info.ID)
}

func TestExtendedLoadBalancer_ResponseTime(t *testing.T) {
	logger := createTestLogger()
	elb := NewExtendedLoadBalancer(logger)
	
	instances := []*ServiceInstance{
		createTestServiceInstance("service-1", "test-service", "1.0.0"),
		createTestServiceInstance("service-2", "test-service", "1.0.0"),
	}
	
	// Record different response times
	elb.RecordResponseTime("service-1", 100*time.Millisecond)
	elb.RecordResponseTime("service-2", 200*time.Millisecond)
	
	// Should select service-1 (faster response time)
	instance, err := elb.SelectServiceWithExtendedStrategy(context.Background(), "test-service", ExtendedLoadBalanceResponseTime, nil, instances)
	assert.NoError(t, err)
	assert.Equal(t, "service-1", instance.Info.ID)
}

func TestExtendedLoadBalancer_LeastLoad(t *testing.T) {
	logger := createTestLogger()
	elb := NewExtendedLoadBalancer(logger)
	
	instances := []*ServiceInstance{
		createTestServiceInstance("service-1", "test-service", "1.0.0"),
		createTestServiceInstance("service-2", "test-service", "1.0.0"),
	}
	
	// Update load information
	elb.UpdateServiceLoad("service-1", 80.0, 60.0, 10) // High load
	elb.UpdateServiceLoad("service-2", 20.0, 30.0, 2)  // Low load
	
	// Should select service-2 (lower load)
	instance, err := elb.SelectServiceWithExtendedStrategy(context.Background(), "test-service", ExtendedLoadBalanceLeastLoad, nil, instances)
	assert.NoError(t, err)
	assert.Equal(t, "service-2", instance.Info.ID)
}

func TestExtendedLoadBalancer_Geographic(t *testing.T) {
	logger := createTestLogger()
	elb := NewExtendedLoadBalancer(logger)
	
	// Create instances in different zones
	instance1 := createTestServiceInstance("service-1", "test-service", "1.0.0")
	instance1.Info.Metadata["zone"] = "us-east-1"
	
	instance2 := createTestServiceInstance("service-2", "test-service", "1.0.0")
	instance2.Info.Metadata["zone"] = "us-west-1"
	
	instances := []*ServiceInstance{instance1, instance2}
	
	lbContext := map[string]interface{}{
			"client_zone": "us-east-1",
		}
	
		// Should prefer instance in the same zone
		instance, err := elb.SelectServiceWithExtendedStrategy(context.Background(), "test-service", ExtendedLoadBalanceGeographic, lbContext, instances)
	assert.NoError(t, err)
	assert.Equal(t, "service-1", instance.Info.ID)
}

// Integration Tests

func TestServiceRegistryExtensions_Integration(t *testing.T) {
	logger := createTestLogger()
	registry := NewServiceRegistry(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	vm := NewVersionManager(logger)
	mm := NewMetricsManager(logger)
	fm := NewFailoverManager(logger, registry, mm)
	elb := NewExtendedLoadBalancer(logger)
	
	ctx := context.Background()
	
	// Register services with versions
	service1 := createTestServiceInfo("service-1", "test-service", "1.0.0")
	service2 := createTestServiceInfo("service-2", "test-service", "1.1.0")
	
	registry.Register(ctx, service1)
	registry.Register(ctx, service2)
	
	// Register versions
	v100, _ := ParseVersion("1.0.0")
	v110, _ := ParseVersion("1.1.0")
	
	instance1, _ := registry.GetService(ctx, "service-1")
	instance2, _ := registry.GetService(ctx, "service-2")
	
	vm.RegisterServiceVersion("test-service", v100, instance1)
	vm.RegisterServiceVersion("test-service", v110, instance2)
	
	// Configure failover
	failoverConfig := &ServiceFailoverConfig{
		Enabled:               true,
		MaxRetries:           3,
		CircuitBreakerEnabled: true,
		FailureThreshold:     5,
	}
	
	fm.ConfigureFailover("service-1", failoverConfig)
	fm.ConfigureFailover("service-2", failoverConfig)
	
	// Configure load balancing
	lbConfig := &LoadBalancerConfig{
		Strategy:     ExtendedLoadBalanceResponseTime,
		VirtualNodes: 150,
	}
	
	elb.ConfigureLoadBalancing("test-service", lbConfig)
	
	// Record metrics
	mm.RecordServiceCall("service-1", 100*time.Millisecond, true, "")
	mm.RecordServiceCall("service-2", 200*time.Millisecond, true, "")
	
	elb.RecordResponseTime("service-1", 100*time.Millisecond)
	elb.RecordResponseTime("service-2", 200*time.Millisecond)
	
	// Test version compatibility
	compatible, err := vm.GetCompatibleServices("test-service", v100)
	assert.NoError(t, err)
	assert.Len(t, compatible, 2) // Both versions should be compatible
	
	// Test load balancing with response time
	instances := []*ServiceInstance{instance1, instance2}
	selected, err := elb.SelectServiceWithExtendedStrategy(ctx, "test-service", ExtendedLoadBalanceResponseTime, nil, instances)
	assert.NoError(t, err)
	assert.Equal(t, "service-1", selected.Info.ID) // Should select faster service
	
	// Test metrics
	metrics1, err := mm.GetServiceMetrics("service-1")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), metrics1.TotalRequests)
	
	// Test failover
	result, err := fm.AttemptFailover(ctx, "test-service", "service-1")
	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "service-2", result.FailoverService.Info.ID)
}

// Benchmark Tests

func BenchmarkVersionManager_RegisterServiceVersion(b *testing.B) {
	logger := createTestLogger()
	vm := NewVersionManager(logger)
	version, _ := ParseVersion("1.0.0")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		instance := createTestServiceInstance(fmt.Sprintf("service-%d", i), "test-service", "1.0.0")
		vm.RegisterServiceVersion("test-service", version, instance)
	}
}

func BenchmarkMetricsManager_RecordServiceCall(b *testing.B) {
	logger := createTestLogger()
	mm := NewMetricsManager(logger)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mm.RecordServiceCall("test-service-1", 100*time.Millisecond, true, "")
	}
}

func BenchmarkExtendedLoadBalancer_ConsistentHash(b *testing.B) {
	logger := createTestLogger()
	elb := NewExtendedLoadBalancer(logger)
	
	config := &LoadBalancerConfig{
		Strategy:     ExtendedLoadBalanceConsistentHash,
		VirtualNodes: 150,
	}
	
	elb.ConfigureLoadBalancing("test-service", config)
	
	instances := make([]*ServiceInstance, 10)
	for i := 0; i < 10; i++ {
		instances[i] = createTestServiceInstance(fmt.Sprintf("service-%d", i), "test-service", "1.0.0")
	}
	
	lbContext := map[string]interface{}{
			"session_id": "user123",
		}
	
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			elb.SelectServiceWithExtendedStrategy(context.Background(), "test-service", ExtendedLoadBalanceConsistentHash, lbContext, instances)
		}
}