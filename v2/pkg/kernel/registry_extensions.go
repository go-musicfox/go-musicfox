// Package kernel provides extended service registry functionality
package kernel

import (
	"context"
	"fmt"
	"time"
)

// ServiceVersion represents semantic version information
type ServiceVersion struct {
	Major int    `json:"major"`
	Minor int    `json:"minor"`
	Patch int    `json:"patch"`
	Pre   string `json:"pre,omitempty"`   // Pre-release version
	Build string `json:"build,omitempty"` // Build metadata
}

// String returns the string representation of the version
func (v *ServiceVersion) String() string {
	version := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Pre != "" {
		version += "-" + v.Pre
	}
	if v.Build != "" {
		version += "+" + v.Build
	}
	return version
}

// IsCompatible checks if this version is compatible with another version
func (v *ServiceVersion) IsCompatible(other *ServiceVersion) bool {
	// Major version must match for compatibility
	return v.Major == other.Major
}

// IsNewer checks if this version is newer than another version
func (v *ServiceVersion) IsNewer(other *ServiceVersion) bool {
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	return v.Patch > other.Patch
}

// ServiceCompatibility represents compatibility information
type ServiceCompatibility struct {
	MinVersion     *ServiceVersion `json:"min_version"`
	MaxVersion     *ServiceVersion `json:"max_version"`
	DeprecatedFrom *ServiceVersion `json:"deprecated_from,omitempty"`
	RemovedFrom    *ServiceVersion `json:"removed_from,omitempty"`
}

// ServiceMetrics represents service performance metrics
type ServiceMetrics struct {
	ServiceID        string        `json:"service_id"`
	ServiceName      string        `json:"service_name"`
	TotalRequests    int64         `json:"total_requests"`
	SuccessfulRequests int64       `json:"successful_requests"`
	FailedRequests   int64         `json:"failed_requests"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	MinResponseTime  time.Duration `json:"min_response_time"`
	MaxResponseTime  time.Duration `json:"max_response_time"`
	LastRequestTime  time.Time     `json:"last_request_time"`
	Uptime          time.Duration `json:"uptime"`
	ErrorRate       float64       `json:"error_rate"`
	Throughput      float64       `json:"throughput"` // requests per second
}

// ServiceAlert represents a service alert
type ServiceAlert struct {
	ID          string                 `json:"id"`
	ServiceID   string                 `json:"service_id"`
	ServiceName string                 `json:"service_name"`
	Type        ServiceAlertType       `json:"type"`
	Severity    ServiceAlertSeverity   `json:"severity"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details"`
	TriggeredAt time.Time              `json:"triggered_at"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	Status      ServiceAlertStatus     `json:"status"`
}

// ServiceAlertType represents the type of alert
type ServiceAlertType string

const (
	ServiceAlertTypeHealthCheck ServiceAlertType = "health_check"
	ServiceAlertTypePerformance ServiceAlertType = "performance"
	ServiceAlertTypeAvailability ServiceAlertType = "availability"
	ServiceAlertTypeDependency  ServiceAlertType = "dependency"
	ServiceAlertTypeResource    ServiceAlertType = "resource"
)

// ServiceAlertSeverity represents the severity of an alert
type ServiceAlertSeverity string

const (
	ServiceAlertSeverityLow      ServiceAlertSeverity = "low"
	ServiceAlertSeverityMedium   ServiceAlertSeverity = "medium"
	ServiceAlertSeverityHigh     ServiceAlertSeverity = "high"
	ServiceAlertSeverityCritical ServiceAlertSeverity = "critical"
)

// ServiceAlertStatus represents the status of an alert
type ServiceAlertStatus string

const (
	ServiceAlertStatusActive    ServiceAlertStatus = "active"
	ServiceAlertStatusResolved  ServiceAlertStatus = "resolved"
	ServiceAlertStatusSuppressed ServiceAlertStatus = "suppressed"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState string

const (
	CircuitBreakerStateClosed   CircuitBreakerState = "closed"
	CircuitBreakerStateOpen     CircuitBreakerState = "open"
	CircuitBreakerStateHalfOpen CircuitBreakerState = "half_open"
)

// CircuitBreaker represents a circuit breaker for service calls
type CircuitBreaker struct {
	ServiceID        string              `json:"service_id"`
	State           CircuitBreakerState `json:"state"`
	FailureCount    int                 `json:"failure_count"`
	FailureThreshold int                `json:"failure_threshold"`
	Timeout         time.Duration       `json:"timeout"`
	LastFailureTime time.Time           `json:"last_failure_time"`
	NextRetryTime   time.Time           `json:"next_retry_time"`
}

// ServiceFailoverConfig represents failover configuration
type ServiceFailoverConfig struct {
	Enabled              bool          `json:"enabled"`
	MaxRetries          int           `json:"max_retries"`
	RetryDelay          time.Duration `json:"retry_delay"`
	BackoffMultiplier   float64       `json:"backoff_multiplier"`
	MaxRetryDelay       time.Duration `json:"max_retry_delay"`
	CircuitBreakerEnabled bool         `json:"circuit_breaker_enabled"`
	FailureThreshold    int           `json:"failure_threshold"`
	RecoveryTimeout     time.Duration `json:"recovery_timeout"`
}

// ExtendedLoadBalanceStrategy represents extended load balancing strategies
type ExtendedLoadBalanceStrategy int

const (
	// Existing strategies from base implementation
	ExtendedLoadBalanceRoundRobin ExtendedLoadBalanceStrategy = iota
	ExtendedLoadBalanceRandom
	ExtendedLoadBalanceWeightedRoundRobin
	ExtendedLoadBalanceLeastConnections
	
	// New extended strategies
	ExtendedLoadBalanceConsistentHash
	ExtendedLoadBalanceIPHash
	ExtendedLoadBalanceResponseTime
	ExtendedLoadBalanceLeastLoad
	ExtendedLoadBalanceGeographic
)

// String returns the string representation of ExtendedLoadBalanceStrategy
func (s ExtendedLoadBalanceStrategy) String() string {
	switch s {
	case ExtendedLoadBalanceRoundRobin:
		return "round_robin"
	case ExtendedLoadBalanceRandom:
		return "random"
	case ExtendedLoadBalanceWeightedRoundRobin:
		return "weighted_round_robin"
	case ExtendedLoadBalanceLeastConnections:
		return "least_connections"
	case ExtendedLoadBalanceConsistentHash:
		return "consistent_hash"
	case ExtendedLoadBalanceIPHash:
		return "ip_hash"
	case ExtendedLoadBalanceResponseTime:
		return "response_time"
	case ExtendedLoadBalanceLeastLoad:
		return "least_load"
	case ExtendedLoadBalanceGeographic:
		return "geographic"
	default:
		return "unknown"
	}
}

// ServiceRegistryExtensions defines extended service registry interface
type ServiceRegistryExtensions interface {
	// Version Management
	RegisterServiceWithVersion(ctx context.Context, service *ServiceInfo, version *ServiceVersion) error
	GetServicesByVersion(ctx context.Context, serviceName string, version *ServiceVersion) ([]*ServiceInstance, error)
	CheckCompatibility(ctx context.Context, serviceName string, requiredVersion *ServiceVersion) (*ServiceCompatibility, error)
	ListServiceVersions(ctx context.Context, serviceName string) ([]*ServiceVersion, error)
	DeprecateServiceVersion(ctx context.Context, serviceName string, version *ServiceVersion, deprecationTime time.Time) error
	
	// Performance Monitoring
	RecordServiceCall(ctx context.Context, serviceID string, duration time.Duration, success bool) error
	GetServiceMetrics(ctx context.Context, serviceID string) (*ServiceMetrics, error)
	GetServiceMetricsByName(ctx context.Context, serviceName string) ([]*ServiceMetrics, error)
	GetAggregatedMetrics(ctx context.Context, serviceName string, timeRange time.Duration) (*ServiceMetrics, error)
	
	// Alert Management
	CreateAlert(ctx context.Context, alert *ServiceAlert) error
	ResolveAlert(ctx context.Context, alertID string) error
	GetActiveAlerts(ctx context.Context, serviceID string) ([]*ServiceAlert, error)
	GetAlertHistory(ctx context.Context, serviceID string, timeRange time.Duration) ([]*ServiceAlert, error)
	
	// Circuit Breaker and Failover
	GetCircuitBreaker(ctx context.Context, serviceID string) (*CircuitBreaker, error)
	UpdateCircuitBreakerState(ctx context.Context, serviceID string, state CircuitBreakerState) error
	ConfigureFailover(ctx context.Context, serviceID string, config *ServiceFailoverConfig) error
	GetFailoverConfig(ctx context.Context, serviceID string) (*ServiceFailoverConfig, error)
	
	// Extended Load Balancing
	SelectServiceWithExtendedStrategy(ctx context.Context, serviceName string, strategy ExtendedLoadBalanceStrategy, context map[string]interface{}) (*ServiceInstance, error)
	ConfigureLoadBalancing(ctx context.Context, serviceName string, strategy ExtendedLoadBalanceStrategy, config map[string]interface{}) error
	
	// Batch Operations
	RegisterServicesBatch(ctx context.Context, services []*ServiceInfo) error
	DeregisterServicesBatch(ctx context.Context, serviceIDs []string) error
	UpdateServicesBatch(ctx context.Context, updates map[string]map[string]interface{}) error
	
	// Service Groups and Tags
	CreateServiceGroup(ctx context.Context, groupName string, serviceIDs []string) error
	AddServiceToGroup(ctx context.Context, groupName string, serviceID string) error
	RemoveServiceFromGroup(ctx context.Context, groupName string, serviceID string) error
	GetServicesByGroup(ctx context.Context, groupName string) ([]*ServiceInstance, error)
	GetServicesByTags(ctx context.Context, tags []string, matchAll bool) ([]*ServiceInstance, error)
	
	// Advanced Statistics
	GetRegistryStatistics(ctx context.Context) (*ExtendedRegistryStats, error)
	GetServiceTopology(ctx context.Context) (*ServiceTopology, error)
	ExportMetrics(ctx context.Context, format string, timeRange time.Duration) ([]byte, error)
}

// ExtendedRegistryStats represents extended registry statistics
type ExtendedRegistryStats struct {
	*RegistryStats
	ServiceGroups        map[string][]string    `json:"service_groups"`
	VersionDistribution  map[string]int         `json:"version_distribution"`
	LoadBalancingStats   map[string]interface{} `json:"load_balancing_stats"`
	CircuitBreakerStats  map[string]interface{} `json:"circuit_breaker_stats"`
	AlertStats          map[string]int         `json:"alert_stats"`
	PerformanceStats    map[string]*ServiceMetrics `json:"performance_stats"`
}

// ServiceTopology represents the service dependency topology
type ServiceTopology struct {
	Services     map[string]*ServiceNode `json:"services"`
	Dependencies map[string][]string     `json:"dependencies"`
	Groups       map[string][]string     `json:"groups"`
	UpdatedAt    time.Time               `json:"updated_at"`
}

// ServiceNode represents a node in the service topology
type ServiceNode struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Version      *ServiceVersion        `json:"version"`
	State        ServiceState           `json:"state"`
	Metrics      *ServiceMetrics        `json:"metrics"`
	Dependencies []string               `json:"dependencies"`
	Dependents   []string               `json:"dependents"`
	Metadata     map[string]interface{} `json:"metadata"`
}