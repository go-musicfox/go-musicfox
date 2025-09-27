package kernel

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMicroKernelIntegration tests the complete microkernel integration
func TestMicroKernelIntegration(t *testing.T) {
	// Create launcher with test configuration
	config := &LaunchConfig{
		Mode:            LaunchModeTest,
		LogLevel:        "info",
		LogFormat:       "text",
		StartTimeout:    10 * time.Second,
		StopTimeout:     5 * time.Second,
		ShutdownTimeout: 10 * time.Second,
		EnableSignals:   false, // Disable signals for testing
	}

	launcher := NewLauncherWithConfig(config)
	require.NotNil(t, launcher)

	// Initialize launcher
	err := launcher.Initialize()
	assert.NoError(t, err)

	// Verify kernel is created and accessible
	kernel := launcher.GetKernel()
	assert.NotNil(t, kernel)

	// Verify logger is created
	logger := launcher.GetLogger()
	assert.NotNil(t, logger)

	// Test kernel lifecycle
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Initialize kernel
	err = kernel.Initialize(ctx)
	assert.NoError(t, err)
	assert.Equal(t, KernelStateInitialized, kernel.GetStatus().State)

	// Start kernel
	err = kernel.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, kernel.IsRunning())
	assert.Equal(t, KernelStateRunning, kernel.GetStatus().State)

	// Verify core components are accessible
	assert.NotNil(t, kernel.GetPluginManager())
	assert.NotNil(t, kernel.GetEventBus())
	assert.NotNil(t, kernel.GetServiceRegistry())
	assert.NotNil(t, kernel.GetSecurityManager())
	assert.NotNil(t, kernel.GetConfig())
	assert.NotNil(t, kernel.GetLogger())
	assert.NotNil(t, kernel.GetContainer())

	// Stop kernel
	err = kernel.Stop(ctx)
	assert.NoError(t, err)
	assert.False(t, kernel.IsRunning())

	// Shutdown kernel
	err = kernel.Shutdown(ctx)
	assert.NoError(t, err)

	// Shutdown launcher
	err = launcher.Shutdown()
	assert.NoError(t, err)
}

// TestBootstrapIntegration tests bootstrap integration with kernel
func TestBootstrapIntegration(t *testing.T) {
	bootstrap := NewBootstrap(nil)
	require.NotNil(t, bootstrap)

	// Initialize bootstrap
	err := bootstrap.Initialize()
	assert.NoError(t, err)

	// Verify kernel is created
	kernel := bootstrap.GetKernel()
	assert.NotNil(t, kernel)

	// Verify logger is created
	logger := bootstrap.GetLogger()
	assert.NotNil(t, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test bootstrap lifecycle
	err = bootstrap.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, kernel.IsRunning())

	err = bootstrap.Stop(ctx)
	assert.NoError(t, err)
	assert.False(t, kernel.IsRunning())

	err = bootstrap.Shutdown(ctx)
	assert.NoError(t, err)
}

// TestConfigManagerIntegration tests config manager integration
func TestConfigManagerIntegration(t *testing.T) {
	bootstrap := NewBootstrap(nil)
	err := bootstrap.Initialize()
	require.NoError(t, err)

	kernel := bootstrap.GetKernel()
	config := kernel.GetConfig()
	assert.NotNil(t, config)

	// Verify default configuration is loaded
	assert.True(t, config.Exists("kernel.name"))
	assert.True(t, config.Exists("kernel.version"))
	assert.Equal(t, "go-musicfox", config.String("kernel.name"))

	// Test configuration access
	assert.True(t, config.Bool("plugins.enabled"))
	assert.True(t, config.Bool("security.enabled"))
	assert.True(t, config.Bool("registry.enabled"))
	assert.True(t, config.Bool("events.enabled"))
}

// TestServiceRegistryIntegration tests service registry integration
func TestServiceRegistryIntegration(t *testing.T) {
	bootstrap := NewBootstrap(nil)
	err := bootstrap.Initialize()
	require.NoError(t, err)

	kernel := bootstrap.GetKernel()
	ctx := context.Background()

	// Initialize kernel to set up components
	err = kernel.Initialize(ctx)
	assert.NoError(t, err)

	registry := kernel.GetServiceRegistry()
	assert.NotNil(t, registry)

	// Test service registration
	serviceInfo := &ServiceInfo{
		ID:      "test-service-integration",
		Name:    "integration-test",
		Address: "localhost",
		Port:    8080,
		Tags:    []string{"test", "integration"},
	}

	err = registry.Register(ctx, serviceInfo)
	assert.NoError(t, err)

	// Test service discovery
	services, err := registry.Discover(ctx, "integration-test")
	assert.NoError(t, err)
	assert.Len(t, services, 1)
	if len(services) > 0 {
		assert.Equal(t, "test-service-integration", services[0].Info.ID)
	}

	// Test service deregistration
	err = registry.Deregister(ctx, "test-service-integration")
	assert.NoError(t, err)

	// Verify service is removed
	services, err = registry.Discover(ctx, "integration-test")
	assert.NoError(t, err)
	assert.Len(t, services, 0)

	// Cleanup
	err = kernel.Shutdown(ctx)
	assert.NoError(t, err)
}

// TestEventBusIntegration tests event bus integration
func TestEventBusIntegration(t *testing.T) {
	bootstrap := NewBootstrap(nil)
	err := bootstrap.Initialize()
	require.NoError(t, err)

	kernel := bootstrap.GetKernel()
	ctx := context.Background()

	// Initialize kernel to set up components
	err = kernel.Initialize(ctx)
	assert.NoError(t, err)

	eventBus := kernel.GetEventBus()
	assert.NotNil(t, eventBus)

	// Test basic event bus functionality
	// Note: Actual event bus implementation may vary
	// This is a simplified test for integration purposes
	assert.NotNil(t, eventBus)
	
	// Skip detailed event testing as it requires specific event implementation
	// The important part is that the event bus is accessible from the kernel

	// Cleanup
	err = kernel.Shutdown(ctx)
	assert.NoError(t, err)
}

// TestSecurityManagerIntegration tests security manager integration
func TestSecurityManagerIntegration(t *testing.T) {
	bootstrap := NewBootstrap(nil)
	err := bootstrap.Initialize()
	require.NoError(t, err)

	kernel := bootstrap.GetKernel()
	ctx := context.Background()

	// Initialize kernel to set up components
	err = kernel.Initialize(ctx)
	assert.NoError(t, err)

	securityManager := kernel.GetSecurityManager()
	assert.NotNil(t, securityManager)

	// Test security validation
	pluginID := "test-integration-plugin"
	resource := "test-resource"
	actions := []string{"read"}

	// Test permission check (should be denied by default)
	allowed, err := securityManager.CheckPermission(pluginID, resource, actions[0])
	assert.NoError(t, err)
	assert.False(t, allowed)

	// Grant permission
	err = securityManager.GrantPermission(pluginID, resource, actions)
	assert.NoError(t, err)

	// Test permission check again (should be allowed now)
	allowed, err = securityManager.CheckPermission(pluginID, resource, actions[0])
	assert.NoError(t, err)
	assert.True(t, allowed)

	// Revoke permission
	err = securityManager.RevokePermission(pluginID, resource, actions)
	assert.NoError(t, err)

	// Test permission check (should be denied again)
	allowed, err = securityManager.CheckPermission(pluginID, resource, actions[0])
	assert.NoError(t, err)
	assert.False(t, allowed)

	// Cleanup
	err = kernel.Shutdown(ctx)
	assert.NoError(t, err)
}

// TestPluginManagerIntegration tests plugin manager integration
func TestPluginManagerIntegration(t *testing.T) {
	bootstrap := NewBootstrap(nil)
	err := bootstrap.Initialize()
	require.NoError(t, err)

	kernel := bootstrap.GetKernel()
	ctx := context.Background()

	// Initialize kernel to set up components
	err = kernel.Initialize(ctx)
	assert.NoError(t, err)

	pluginManager := kernel.GetPluginManager()
	assert.NotNil(t, pluginManager)

	// Test plugin manager basic functionality
	assert.Equal(t, 0, pluginManager.GetLoadedPluginCount())

	loadedPlugins := pluginManager.GetLoadedPlugins()
	assert.NotNil(t, loadedPlugins)
	assert.Len(t, loadedPlugins, 0)

	// Cleanup
	err = kernel.Shutdown(ctx)
	assert.NoError(t, err)
}