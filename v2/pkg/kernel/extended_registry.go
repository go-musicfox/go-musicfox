// Package kernel provides extended service registry implementation
package kernel

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// ExtendedServiceRegistry implements ServiceRegistryExtensions interface
type ExtendedServiceRegistry struct {
	*ServiceRegistryImpl // Embed base registry
	
	// Extended functionality managers
	versionManager  *VersionManager
	metricsManager  *MetricsManager
	failoverManager *FailoverManager
	loadBalancer    *ExtendedLoadBalancer
	
	// Service groups and tags
	serviceGroups map[string][]string // groupName -> serviceIDs
	groupMutex    sync.RWMutex
	
	// Extended configuration
	extendedConfig *ExtendedRegistryConfig
	
	// Cleanup ticker
	cleanupTicker *time.Ticker
	cleanupDone   chan struct{}
}

// ExtendedRegistryConfig represents extended registry configuration
type ExtendedRegistryConfig struct {
	EnableVersionManagement bool          `json:"enable_version_management"`
	EnableMetrics          bool          `json:"enable_metrics"`
	EnableFailover         bool          `json:"enable_failover"`
	EnableExtendedLB       bool          `json:"enable_extended_lb"`
	CleanupInterval        time.Duration `json:"cleanup_interval"`
	MetricsRetention       time.Duration `json:"metrics_retention"`
	MaxEventHistory        int           `json:"max_event_history"`
}

// DefaultExtendedRegistryConfig returns default configuration
func DefaultExtendedRegistryConfig() *ExtendedRegistryConfig {
	return &ExtendedRegistryConfig{
		EnableVersionManagement: true,
		EnableMetrics:          true,
		EnableFailover:         true,
		EnableExtendedLB:       true,
		CleanupInterval:        5 * time.Minute,
		MetricsRetention:       24 * time.Hour,
		MaxEventHistory:        1000,
	}
}

// NewExtendedServiceRegistry creates a new extended service registry
func NewExtendedServiceRegistry(logger *slog.Logger, config *ExtendedRegistryConfig) *ExtendedServiceRegistry {
	if config == nil {
		config = DefaultExtendedRegistryConfig()
	}
	
	baseRegistry := NewServiceRegistry(logger).(*ServiceRegistryImpl)
	adapter := NewSlogAdapter(logger)
	
	extendedRegistry := &ExtendedServiceRegistry{
		ServiceRegistryImpl: baseRegistry,
		serviceGroups:       make(map[string][]string),
		extendedConfig:      config,
		cleanupDone:         make(chan struct{}),
	}
	
	// Initialize managers based on configuration
	if config.EnableVersionManagement {
		extendedRegistry.versionManager = NewVersionManager(adapter)
	}
	
	if config.EnableMetrics {
		extendedRegistry.metricsManager = NewMetricsManager(adapter)
	}
	
	if config.EnableFailover {
		extendedRegistry.failoverManager = NewFailoverManager(adapter, baseRegistry, extendedRegistry.metricsManager)
	}
	
	if config.EnableExtendedLB {
		extendedRegistry.loadBalancer = NewExtendedLoadBalancer(adapter)
	}
	
	// Start cleanup routine
	if config.CleanupInterval > 0 {
		extendedRegistry.startCleanupRoutine()
	}
	
	return extendedRegistry
}

// Version Management Implementation

func (esr *ExtendedServiceRegistry) RegisterServiceWithVersion(ctx context.Context, service *ServiceInfo, version *ServiceVersion) error {
	if esr.versionManager == nil {
		return fmt.Errorf("version management not enabled")
	}
	
	// Register service in base registry first
	if err := esr.ServiceRegistryImpl.Register(ctx, service); err != nil {
		return err
	}
	
	// Get the registered instance
	instance, err := esr.ServiceRegistryImpl.GetService(ctx, service.ID)
	if err != nil {
		return err
	}
	
	// Register version
	return esr.versionManager.RegisterServiceVersion(service.Name, version, instance)
}

func (esr *ExtendedServiceRegistry) GetServicesByVersion(ctx context.Context, serviceName string, version *ServiceVersion) ([]*ServiceInstance, error) {
	if esr.versionManager == nil {
		return nil, fmt.Errorf("version management not enabled")
	}
	
	return esr.versionManager.GetServicesByVersion(serviceName, version)
}

func (esr *ExtendedServiceRegistry) CheckCompatibility(ctx context.Context, serviceName string, requiredVersion *ServiceVersion) (*ServiceCompatibility, error) {
	if esr.versionManager == nil {
		return nil, fmt.Errorf("version management not enabled")
	}
	
	return esr.versionManager.CheckCompatibility(serviceName, requiredVersion)
}

func (esr *ExtendedServiceRegistry) ListServiceVersions(ctx context.Context, serviceName string) ([]*ServiceVersion, error) {
	if esr.versionManager == nil {
		return nil, fmt.Errorf("version management not enabled")
	}
	
	return esr.versionManager.ListServiceVersions(serviceName)
}

func (esr *ExtendedServiceRegistry) DeprecateServiceVersion(ctx context.Context, serviceName string, version *ServiceVersion, deprecationTime time.Time) error {
	if esr.versionManager == nil {
		return fmt.Errorf("version management not enabled")
	}
	
	return esr.versionManager.DeprecateServiceVersion(serviceName, version, deprecationTime, "Deprecated by administrator", nil)
}

// Performance Monitoring Implementation

func (esr *ExtendedServiceRegistry) RecordServiceCall(ctx context.Context, serviceID string, duration time.Duration, success bool) error {
	if esr.metricsManager == nil {
		return fmt.Errorf("metrics not enabled")
	}
	
	errorCode := ""
	if !success {
		errorCode = "unknown_error"
	}
	
	// Record in metrics manager
	if err := esr.metricsManager.RecordServiceCall(serviceID, duration, success, errorCode); err != nil {
		return err
	}
	
	// Record in load balancer for response time tracking
	if esr.loadBalancer != nil {
		esr.loadBalancer.RecordResponseTime(serviceID, duration)
	}
	
	// Record in failover manager for circuit breaker
	if esr.failoverManager != nil {
		return esr.failoverManager.RecordServiceCall(serviceID, success, duration)
	}
	
	return nil
}

func (esr *ExtendedServiceRegistry) GetServiceMetrics(ctx context.Context, serviceID string) (*ServiceMetrics, error) {
	if esr.metricsManager == nil {
		return nil, fmt.Errorf("metrics not enabled")
	}
	
	return esr.metricsManager.GetServiceMetrics(serviceID)
}

func (esr *ExtendedServiceRegistry) GetServiceMetricsByName(ctx context.Context, serviceName string) ([]*ServiceMetrics, error) {
	if esr.metricsManager == nil {
		return nil, fmt.Errorf("metrics not enabled")
	}
	
	return esr.metricsManager.GetServiceMetricsByName(serviceName, esr.ServiceRegistryImpl)
}

func (esr *ExtendedServiceRegistry) GetAggregatedMetrics(ctx context.Context, serviceName string, timeRange time.Duration) (*ServiceMetrics, error) {
	if esr.metricsManager == nil {
		return nil, fmt.Errorf("metrics not enabled")
	}
	
	return esr.metricsManager.GetAggregatedMetrics(serviceName, timeRange, esr.ServiceRegistryImpl)
}

// Alert Management Implementation

func (esr *ExtendedServiceRegistry) CreateAlert(ctx context.Context, alert *ServiceAlert) error {
	if esr.metricsManager == nil {
		return fmt.Errorf("metrics not enabled")
	}
	
	return esr.metricsManager.CreateAlert(alert)
}

func (esr *ExtendedServiceRegistry) ResolveAlert(ctx context.Context, alertID string) error {
	if esr.metricsManager == nil {
		return fmt.Errorf("metrics not enabled")
	}
	
	return esr.metricsManager.ResolveAlert(alertID)
}

func (esr *ExtendedServiceRegistry) GetActiveAlerts(ctx context.Context, serviceID string) ([]*ServiceAlert, error) {
	if esr.metricsManager == nil {
		return nil, fmt.Errorf("metrics not enabled")
	}
	
	return esr.metricsManager.GetActiveAlerts(serviceID)
}

func (esr *ExtendedServiceRegistry) GetAlertHistory(ctx context.Context, serviceID string, timeRange time.Duration) ([]*ServiceAlert, error) {
	if esr.metricsManager == nil {
		return nil, fmt.Errorf("metrics not enabled")
	}
	
	return esr.metricsManager.GetAlertHistory(serviceID, timeRange)
}

// Circuit Breaker and Failover Implementation

func (esr *ExtendedServiceRegistry) GetCircuitBreaker(ctx context.Context, serviceID string) (*CircuitBreaker, error) {
	if esr.failoverManager == nil {
		return nil, fmt.Errorf("failover not enabled")
	}
	
	return esr.failoverManager.GetCircuitBreaker(serviceID)
}

func (esr *ExtendedServiceRegistry) UpdateCircuitBreakerState(ctx context.Context, serviceID string, state CircuitBreakerState) error {
	if esr.failoverManager == nil {
		return fmt.Errorf("failover not enabled")
	}
	
	return esr.failoverManager.UpdateCircuitBreakerState(serviceID, state)
}

func (esr *ExtendedServiceRegistry) ConfigureFailover(ctx context.Context, serviceID string, config *ServiceFailoverConfig) error {
	if esr.failoverManager == nil {
		return fmt.Errorf("failover not enabled")
	}
	
	return esr.failoverManager.ConfigureFailover(serviceID, config)
}

func (esr *ExtendedServiceRegistry) GetFailoverConfig(ctx context.Context, serviceID string) (*ServiceFailoverConfig, error) {
	if esr.failoverManager == nil {
		return nil, fmt.Errorf("failover not enabled")
	}
	
	return esr.failoverManager.GetFailoverConfig(serviceID)
}

// Extended Load Balancing Implementation

func (esr *ExtendedServiceRegistry) SelectServiceWithExtendedStrategy(ctx context.Context, serviceName string, strategy ExtendedLoadBalanceStrategy, context map[string]interface{}) (*ServiceInstance, error) {
	if esr.loadBalancer == nil {
		return nil, fmt.Errorf("extended load balancing not enabled")
	}
	
	// Get available instances
	instances, err := esr.ServiceRegistryImpl.Discover(ctx, serviceName)
	if err != nil {
		return nil, err
	}
	
	return esr.loadBalancer.SelectServiceWithExtendedStrategy(ctx, serviceName, strategy, context, instances)
}

func (esr *ExtendedServiceRegistry) ConfigureLoadBalancing(ctx context.Context, serviceName string, strategy ExtendedLoadBalanceStrategy, config map[string]interface{}) error {
	if esr.loadBalancer == nil {
		return fmt.Errorf("extended load balancing not enabled")
	}
	
	// Convert config to LoadBalancerConfig
	lbConfig := &LoadBalancerConfig{
		Strategy: strategy,
	}
	
	// Apply configuration from map
	if virtualNodes, ok := config["virtual_nodes"].(int); ok {
		lbConfig.VirtualNodes = virtualNodes
	} else {
		lbConfig.VirtualNodes = 150 // Default
	}
	
	if healthCheck, ok := config["health_check_enabled"].(bool); ok {
		lbConfig.HealthCheckEnabled = healthCheck
	}
	
	return esr.loadBalancer.ConfigureLoadBalancing(serviceName, lbConfig)
}

// Batch Operations Implementation

func (esr *ExtendedServiceRegistry) RegisterServicesBatch(ctx context.Context, services []*ServiceInfo) error {
	for _, service := range services {
		if err := esr.ServiceRegistryImpl.Register(ctx, service); err != nil {
			return fmt.Errorf("failed to register service %s: %w", service.ID, err)
		}
	}
	return nil
}

func (esr *ExtendedServiceRegistry) DeregisterServicesBatch(ctx context.Context, serviceIDs []string) error {
	for _, serviceID := range serviceIDs {
		if err := esr.ServiceRegistryImpl.Deregister(ctx, serviceID); err != nil {
			return fmt.Errorf("failed to deregister service %s: %w", serviceID, err)
		}
	}
	return nil
}

func (esr *ExtendedServiceRegistry) UpdateServicesBatch(ctx context.Context, updates map[string]map[string]interface{}) error {
	for serviceID, updateData := range updates {
		if err := esr.ServiceRegistryImpl.UpdateService(ctx, serviceID, updateData); err != nil {
			return fmt.Errorf("failed to update service %s: %w", serviceID, err)
		}
	}
	return nil
}

// Service Groups and Tags Implementation

func (esr *ExtendedServiceRegistry) CreateServiceGroup(ctx context.Context, groupName string, serviceIDs []string) error {
	esr.groupMutex.Lock()
	defer esr.groupMutex.Unlock()
	
	// Validate that all services exist
	for _, serviceID := range serviceIDs {
		if _, err := esr.ServiceRegistryImpl.GetService(ctx, serviceID); err != nil {
			return fmt.Errorf("service %s not found: %w", serviceID, err)
		}
	}
	
	esr.serviceGroups[groupName] = serviceIDs
	return nil
}

func (esr *ExtendedServiceRegistry) AddServiceToGroup(ctx context.Context, groupName string, serviceID string) error {
	esr.groupMutex.Lock()
	defer esr.groupMutex.Unlock()
	
	// Validate service exists
	if _, err := esr.ServiceRegistryImpl.GetService(ctx, serviceID); err != nil {
		return fmt.Errorf("service %s not found: %w", serviceID, err)
	}
	
	if esr.serviceGroups[groupName] == nil {
		esr.serviceGroups[groupName] = make([]string, 0)
	}
	
	// Check if already in group
	for _, id := range esr.serviceGroups[groupName] {
		if id == serviceID {
			return nil // Already in group
		}
	}
	
	esr.serviceGroups[groupName] = append(esr.serviceGroups[groupName], serviceID)
	return nil
}

func (esr *ExtendedServiceRegistry) RemoveServiceFromGroup(ctx context.Context, groupName string, serviceID string) error {
	esr.groupMutex.Lock()
	defer esr.groupMutex.Unlock()
	
	serviceIDs, exists := esr.serviceGroups[groupName]
	if !exists {
		return fmt.Errorf("group %s not found", groupName)
	}
	
	for i, id := range serviceIDs {
		if id == serviceID {
			esr.serviceGroups[groupName] = append(serviceIDs[:i], serviceIDs[i+1:]...)
			break
		}
	}
	
	// Remove group if empty
	if len(esr.serviceGroups[groupName]) == 0 {
		delete(esr.serviceGroups, groupName)
	}
	
	return nil
}

func (esr *ExtendedServiceRegistry) GetServicesByGroup(ctx context.Context, groupName string) ([]*ServiceInstance, error) {
	esr.groupMutex.RLock()
	serviceIDs, exists := esr.serviceGroups[groupName]
	esr.groupMutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("group %s not found", groupName)
	}
	
	instances := make([]*ServiceInstance, 0, len(serviceIDs))
	for _, serviceID := range serviceIDs {
		if instance, err := esr.ServiceRegistryImpl.GetService(ctx, serviceID); err == nil {
			instances = append(instances, instance)
		}
	}
	
	return instances, nil
}

func (esr *ExtendedServiceRegistry) GetServicesByTags(ctx context.Context, tags []string, matchAll bool) ([]*ServiceInstance, error) {
	allServices, err := esr.ServiceRegistryImpl.ListServices(ctx)
	if err != nil {
		return nil, err
	}
	
	matching := make([]*ServiceInstance, 0)
	
	for _, service := range allServices {
		if esr.matchesTags(service.Info.Tags, tags, matchAll) {
			matching = append(matching, service)
		}
	}
	
	return matching, nil
}

// Advanced Statistics Implementation

func (esr *ExtendedServiceRegistry) GetRegistryStatistics(ctx context.Context) (*ExtendedRegistryStats, error) {
	// Get base stats
	baseStats, err := esr.ServiceRegistryImpl.GetStats(ctx)
	if err != nil {
		return nil, err
	}
	
	extendedStats := &ExtendedRegistryStats{
		RegistryStats: baseStats,
	}
	
	// Add service groups
	esr.groupMutex.RLock()
	extendedStats.ServiceGroups = make(map[string][]string)
	for groupName, serviceIDs := range esr.serviceGroups {
		extendedStats.ServiceGroups[groupName] = serviceIDs
	}
	esr.groupMutex.RUnlock()
	
	// Add version distribution
	if esr.versionManager != nil {
		versionStats := esr.versionManager.GetVersionStatistics()
		extendedStats.VersionDistribution = make(map[string]int)
		if versionDist, ok := versionStats["version_distribution"].(map[string]int); ok {
			extendedStats.VersionDistribution = versionDist
		}
	}
	
	// Add load balancing stats
	if esr.loadBalancer != nil {
		extendedStats.LoadBalancingStats = esr.loadBalancer.GetLoadBalancingStatistics("")
	}
	
	// Add circuit breaker stats
	if esr.failoverManager != nil {
		extendedStats.CircuitBreakerStats = esr.failoverManager.GetFailoverStatistics()
	}
	
	// Add alert stats
	if esr.metricsManager != nil {
		alertStats := esr.metricsManager.GetMetricsStatistics()
		extendedStats.AlertStats = make(map[string]int)
		if activeAlerts, ok := alertStats["active_alerts"].(int); ok {
			extendedStats.AlertStats["active_alerts"] = activeAlerts
		}
		if totalAlerts, ok := alertStats["total_alerts"].(int); ok {
			extendedStats.AlertStats["total_alerts"] = totalAlerts
		}
	}
	
	return extendedStats, nil
}

func (esr *ExtendedServiceRegistry) GetServiceTopology(ctx context.Context) (*ServiceTopology, error) {
	allServices, err := esr.ServiceRegistryImpl.ListServices(ctx)
	if err != nil {
		return nil, err
	}
	
	topology := &ServiceTopology{
		Services:     make(map[string]*ServiceNode),
		Dependencies: make(map[string][]string),
		Groups:       make(map[string][]string),
		UpdatedAt:    time.Now(),
	}
	
	// Build service nodes
	for _, service := range allServices {
		var version *ServiceVersion
		if esr.versionManager != nil {
			if v, err := ParseVersion(service.Info.Version); err == nil {
				version = v
			}
		}
		
		var metrics *ServiceMetrics
		if esr.metricsManager != nil {
			if m, err := esr.metricsManager.GetServiceMetrics(service.Info.ID); err == nil {
				metrics = m
			}
		}
		
		node := &ServiceNode{
			ID:           service.Info.ID,
			Name:         service.Info.Name,
			Version:      version,
			State:        service.State,
			Metrics:      metrics,
			Dependencies: service.Info.Dependencies,
			Dependents:   make([]string, 0),
			Metadata:     make(map[string]interface{}),
		}
		
		// Copy metadata
		for k, v := range service.Info.Metadata {
			node.Metadata[k] = v
		}
		
		topology.Services[service.Info.ID] = node
		topology.Dependencies[service.Info.ID] = service.Info.Dependencies
	}
	
	// Build dependents
	for serviceID, deps := range topology.Dependencies {
		for _, dep := range deps {
			if depNode, exists := topology.Services[dep]; exists {
				depNode.Dependents = append(depNode.Dependents, serviceID)
			}
		}
	}
	
	// Add groups
	esr.groupMutex.RLock()
	for groupName, serviceIDs := range esr.serviceGroups {
		topology.Groups[groupName] = serviceIDs
	}
	esr.groupMutex.RUnlock()
	
	return topology, nil
}

func (esr *ExtendedServiceRegistry) ExportMetrics(ctx context.Context, format string, timeRange time.Duration) ([]byte, error) {
	if esr.metricsManager == nil {
		return nil, fmt.Errorf("metrics not enabled")
	}
	
	return esr.metricsManager.ExportMetrics(format, timeRange)
}

// Helper methods

func (esr *ExtendedServiceRegistry) matchesTags(serviceTags, requiredTags []string, matchAll bool) bool {
	if len(requiredTags) == 0 {
		return true
	}
	
	matches := 0
	for _, requiredTag := range requiredTags {
		for _, serviceTag := range serviceTags {
			if serviceTag == requiredTag {
				matches++
				break
			}
		}
	}
	
	if matchAll {
		return matches == len(requiredTags)
	}
	return matches > 0
}

func (esr *ExtendedServiceRegistry) startCleanupRoutine() {
	esr.cleanupTicker = time.NewTicker(esr.extendedConfig.CleanupInterval)
	
	go func() {
		for {
			select {
			case <-esr.cleanupTicker.C:
				esr.performCleanup()
			case <-esr.cleanupDone:
				return
			}
		}
	}()
}

func (esr *ExtendedServiceRegistry) performCleanup() {
	// Clean up stale load balancing data
	if esr.loadBalancer != nil {
		esr.loadBalancer.CleanupStaleData(esr.extendedConfig.MetricsRetention)
	}
	
	// Additional cleanup tasks can be added here
	esr.logger.Debug("Performed registry cleanup")
}

// Shutdown gracefully shuts down the extended registry
func (esr *ExtendedServiceRegistry) Shutdown(ctx context.Context) error {
	// Stop cleanup routine
	if esr.cleanupTicker != nil {
		esr.cleanupTicker.Stop()
		close(esr.cleanupDone)
	}
	
	// Shutdown base registry
	return esr.ServiceRegistryImpl.Shutdown(ctx)
}