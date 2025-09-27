// Package kernel provides extended load balancing functionality
package kernel

import (
	"context"
	"crypto/md5"
	"fmt"
	"hash/crc32"
	"math"
	"math/rand"
	"net"
	"sort"
	"sync"
	"time"
)

// ExtendedLoadBalancer provides advanced load balancing algorithms
type ExtendedLoadBalancer struct {
	// Consistent hashing
	consistentHashRings map[string]*ConsistentHashRing // serviceName -> hash ring
	
	// Response time tracking
	responseTimeTracker map[string]*ResponseTimeTracker // serviceID -> tracker
	
	// Geographic load balancing
	geographicZones map[string]*GeographicZone // zoneID -> zone
	
	// Load tracking
	loadTracker map[string]*LoadTracker // serviceID -> load tracker
	
	// Configuration
	configs map[string]*LoadBalancerConfig // serviceName -> config
	
	mutex  sync.RWMutex
	logger Logger
}

// ConsistentHashRing implements consistent hashing for load balancing
type ConsistentHashRing struct {
	nodes    map[uint32]*ServiceInstance // hash -> service instance
	sortedHashes []uint32               // sorted hash values
	virtualNodes int                    // number of virtual nodes per service
	mutex       sync.RWMutex
}

// ResponseTimeTracker tracks response times for services
type ResponseTimeTracker struct {
	ServiceID       string        `json:"service_id"`
	AverageTime     time.Duration `json:"average_time"`
	RecentTimes     []time.Duration `json:"recent_times"`
	MaxSamples      int           `json:"max_samples"`
	LastUpdate      time.Time     `json:"last_update"`
	mutex          sync.RWMutex  `json:"-"`
}

// GeographicZone represents a geographic zone for load balancing
type GeographicZone struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Region      string             `json:"region"`
	Latitude    float64            `json:"latitude"`
	Longitude   float64            `json:"longitude"`
	Instances   []*ServiceInstance `json:"instances"`
	Priority    int                `json:"priority"`
	Capacity    int                `json:"capacity"`
	CurrentLoad int                `json:"current_load"`
}

// LoadTracker tracks current load for services
type LoadTracker struct {
	ServiceID     string    `json:"service_id"`
	CurrentLoad   float64   `json:"current_load"`   // 0.0 to 1.0
	CPUUsage      float64   `json:"cpu_usage"`      // 0.0 to 100.0
	MemoryUsage   float64   `json:"memory_usage"`   // 0.0 to 100.0
	ActiveRequests int      `json:"active_requests"`
	LastUpdate    time.Time `json:"last_update"`
	mutex         sync.RWMutex `json:"-"`
}

// LoadBalancerConfig represents load balancer configuration
type LoadBalancerConfig struct {
	ServiceName          string                      `json:"service_name"`
	Strategy             ExtendedLoadBalanceStrategy `json:"strategy"`
	VirtualNodes         int                         `json:"virtual_nodes"`
	HealthCheckEnabled   bool                        `json:"health_check_enabled"`
	ResponseTimeWeight   float64                     `json:"response_time_weight"`
	LoadWeight           float64                     `json:"load_weight"`
	GeographicPreference []string                    `json:"geographic_preference"`
	StickySession        bool                        `json:"sticky_session"`
	SessionTimeout       time.Duration               `json:"session_timeout"`
	CustomWeights        map[string]float64          `json:"custom_weights"`
}

// SessionInfo represents sticky session information
type SessionInfo struct {
	SessionID   string           `json:"session_id"`
	ServiceID   string           `json:"service_id"`
	CreatedAt   time.Time        `json:"created_at"`
	LastAccess  time.Time        `json:"last_access"`
	ExpiresAt   time.Time        `json:"expires_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// NewExtendedLoadBalancer creates a new extended load balancer
func NewExtendedLoadBalancer(logger Logger) *ExtendedLoadBalancer {
	return &ExtendedLoadBalancer{
		consistentHashRings: make(map[string]*ConsistentHashRing),
		responseTimeTracker: make(map[string]*ResponseTimeTracker),
		geographicZones:     make(map[string]*GeographicZone),
		loadTracker:        make(map[string]*LoadTracker),
		configs:            make(map[string]*LoadBalancerConfig),
		logger:             logger,
	}
}

// ConfigureLoadBalancing configures load balancing for a service
func (elb *ExtendedLoadBalancer) ConfigureLoadBalancing(serviceName string, config *LoadBalancerConfig) error {
	elb.mutex.Lock()
	defer elb.mutex.Unlock()
	
	config.ServiceName = serviceName
	elb.configs[serviceName] = config
	
	// Initialize consistent hash ring if needed
	if config.Strategy == ExtendedLoadBalanceConsistentHash {
		elb.consistentHashRings[serviceName] = NewConsistentHashRing(config.VirtualNodes)
	}
	
	elb.logger.Info("Load balancing configured",
		"service_name", serviceName,
		"strategy", config.Strategy.String())
	
	return nil
}

// SelectServiceWithExtendedStrategy selects a service using extended load balancing
func (elb *ExtendedLoadBalancer) SelectServiceWithExtendedStrategy(ctx context.Context, serviceName string, strategy ExtendedLoadBalanceStrategy, context map[string]interface{}, instances []*ServiceInstance) (*ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("no instances available for service %s", serviceName)
	}
	
	if len(instances) == 1 {
		return instances[0], nil
	}
	
	switch strategy {
	case ExtendedLoadBalanceConsistentHash:
		return elb.selectConsistentHash(serviceName, context, instances)
	case ExtendedLoadBalanceIPHash:
		return elb.selectIPHash(context, instances)
	case ExtendedLoadBalanceResponseTime:
		return elb.selectByResponseTime(instances)
	case ExtendedLoadBalanceLeastLoad:
		return elb.selectLeastLoad(instances)
	case ExtendedLoadBalanceGeographic:
		return elb.selectGeographic(context, instances)
	default:
		// Fall back to round robin
		return instances[0], nil
	}
}

// Consistent Hash Load Balancing
func (elb *ExtendedLoadBalancer) selectConsistentHash(serviceName string, context map[string]interface{}, instances []*ServiceInstance) (*ServiceInstance, error) {
	elb.mutex.RLock()
	ring, exists := elb.consistentHashRings[serviceName]
	elb.mutex.RUnlock()
	
	if !exists {
		// Create ring on demand
		elb.mutex.Lock()
		ring = NewConsistentHashRing(150) // Default virtual nodes
		elb.consistentHashRings[serviceName] = ring
		elb.mutex.Unlock()
	}
	
	// Update ring with current instances
	ring.UpdateInstances(instances)
	
	// Get key for hashing (could be session ID, user ID, etc.)
	key := "default"
	if sessionID, ok := context["session_id"].(string); ok {
		key = sessionID
	} else if userID, ok := context["user_id"].(string); ok {
		key = userID
	} else if clientIP, ok := context["client_ip"].(string); ok {
		key = clientIP
	}
	
	return ring.GetInstance(key), nil
}

// IP Hash Load Balancing
func (elb *ExtendedLoadBalancer) selectIPHash(context map[string]interface{}, instances []*ServiceInstance) (*ServiceInstance, error) {
	clientIP, ok := context["client_ip"].(string)
	if !ok {
		// Fall back to random selection
		return instances[rand.Intn(len(instances))], nil
	}
	
	// Parse IP and use for hashing
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return instances[rand.Intn(len(instances))], nil
	}
	
	// Use CRC32 hash of IP
	hash := crc32.ChecksumIEEE(ip)
	index := int(hash) % len(instances)
	
	return instances[index], nil
}

// Response Time Based Load Balancing
func (elb *ExtendedLoadBalancer) selectByResponseTime(instances []*ServiceInstance) (*ServiceInstance, error) {
	elb.mutex.RLock()
	defer elb.mutex.RUnlock()
	
	var bestInstance *ServiceInstance
	var bestTime time.Duration = time.Duration(math.MaxInt64)
	
	for _, instance := range instances {
		tracker, exists := elb.responseTimeTracker[instance.Info.ID]
		if !exists {
			// No data, consider this instance
			if bestInstance == nil {
				bestInstance = instance
			}
			continue
		}
		
		if tracker.AverageTime < bestTime {
			bestTime = tracker.AverageTime
			bestInstance = instance
		}
	}
	
	if bestInstance == nil {
		return instances[0], nil
	}
	
	return bestInstance, nil
}

// Least Load Based Load Balancing
func (elb *ExtendedLoadBalancer) selectLeastLoad(instances []*ServiceInstance) (*ServiceInstance, error) {
	elb.mutex.RLock()
	defer elb.mutex.RUnlock()
	
	var bestInstance *ServiceInstance
	var leastLoad float64 = math.MaxFloat64
	
	for _, instance := range instances {
		tracker, exists := elb.loadTracker[instance.Info.ID]
		if !exists {
			// No load data, consider this instance as having zero load
			if bestInstance == nil {
				bestInstance = instance
				leastLoad = 0
			}
			continue
		}
		
		if tracker.CurrentLoad < leastLoad {
			leastLoad = tracker.CurrentLoad
			bestInstance = instance
		}
	}
	
	if bestInstance == nil {
		return instances[0], nil
	}
	
	return bestInstance, nil
}

// Geographic Load Balancing
func (elb *ExtendedLoadBalancer) selectGeographic(context map[string]interface{}, instances []*ServiceInstance) (*ServiceInstance, error) {
	clientZone, ok := context["client_zone"].(string)
	if !ok {
		// No zone info, fall back to random
		return instances[rand.Intn(len(instances))], nil
	}
	
	// Group instances by zone
	zoneInstances := make(map[string][]*ServiceInstance)
	for _, instance := range instances {
		zone := instance.Info.Metadata["zone"]
		if zone == "" {
			zone = "default"
		}
		zoneInstances[zone] = append(zoneInstances[zone], instance)
	}
	
	// Prefer instances in the same zone
	if sameZoneInstances, exists := zoneInstances[clientZone]; exists && len(sameZoneInstances) > 0 {
		return sameZoneInstances[rand.Intn(len(sameZoneInstances))], nil
	}
	
	// Fall back to any available instance
	return instances[rand.Intn(len(instances))], nil
}

// RecordResponseTime records response time for a service
func (elb *ExtendedLoadBalancer) RecordResponseTime(serviceID string, responseTime time.Duration) {
	elb.mutex.Lock()
	defer elb.mutex.Unlock()
	
	tracker, exists := elb.responseTimeTracker[serviceID]
	if !exists {
		tracker = &ResponseTimeTracker{
			ServiceID:   serviceID,
			RecentTimes: make([]time.Duration, 0),
			MaxSamples:  100, // Keep last 100 samples
		}
		elb.responseTimeTracker[serviceID] = tracker
	}
	
	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()
	
	// Add new sample
	tracker.RecentTimes = append(tracker.RecentTimes, responseTime)
	
	// Keep only recent samples
	if len(tracker.RecentTimes) > tracker.MaxSamples {
		tracker.RecentTimes = tracker.RecentTimes[1:]
	}
	
	// Calculate average
	var total time.Duration
	for _, t := range tracker.RecentTimes {
		total += t
	}
	tracker.AverageTime = total / time.Duration(len(tracker.RecentTimes))
	tracker.LastUpdate = time.Now()
}

// UpdateServiceLoad updates load information for a service
func (elb *ExtendedLoadBalancer) UpdateServiceLoad(serviceID string, cpuUsage, memoryUsage float64, activeRequests int) {
	elb.mutex.Lock()
	defer elb.mutex.Unlock()
	
	tracker, exists := elb.loadTracker[serviceID]
	if !exists {
		tracker = &LoadTracker{
			ServiceID: serviceID,
		}
		elb.loadTracker[serviceID] = tracker
	}
	
	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()
	
	tracker.CPUUsage = cpuUsage
	tracker.MemoryUsage = memoryUsage
	tracker.ActiveRequests = activeRequests
	
	// Calculate overall load (weighted average)
	tracker.CurrentLoad = (cpuUsage*0.4 + memoryUsage*0.3 + float64(activeRequests)*0.3) / 100.0
	tracker.LastUpdate = time.Now()
}

// Consistent Hash Ring Implementation

// NewConsistentHashRing creates a new consistent hash ring
func NewConsistentHashRing(virtualNodes int) *ConsistentHashRing {
	return &ConsistentHashRing{
		nodes:        make(map[uint32]*ServiceInstance),
		sortedHashes: make([]uint32, 0),
		virtualNodes: virtualNodes,
	}
}

// UpdateInstances updates the hash ring with current instances
func (chr *ConsistentHashRing) UpdateInstances(instances []*ServiceInstance) {
	chr.mutex.Lock()
	defer chr.mutex.Unlock()
	
	// Clear existing nodes
	chr.nodes = make(map[uint32]*ServiceInstance)
	chr.sortedHashes = make([]uint32, 0)
	
	// Add instances to ring
	for _, instance := range instances {
		chr.addInstance(instance)
	}
	
	// Sort hashes
	sort.Slice(chr.sortedHashes, func(i, j int) bool {
		return chr.sortedHashes[i] < chr.sortedHashes[j]
	})
}

// addInstance adds an instance to the hash ring
func (chr *ConsistentHashRing) addInstance(instance *ServiceInstance) {
	for i := 0; i < chr.virtualNodes; i++ {
		key := fmt.Sprintf("%s:%d", instance.Info.ID, i)
		hash := chr.hash(key)
		chr.nodes[hash] = instance
		chr.sortedHashes = append(chr.sortedHashes, hash)
	}
}

// GetInstance gets an instance from the hash ring
func (chr *ConsistentHashRing) GetInstance(key string) *ServiceInstance {
	chr.mutex.RLock()
	defer chr.mutex.RUnlock()
	
	if len(chr.sortedHashes) == 0 {
		return nil
	}
	
	hash := chr.hash(key)
	
	// Find the first hash >= target hash
	idx := sort.Search(len(chr.sortedHashes), func(i int) bool {
		return chr.sortedHashes[i] >= hash
	})
	
	// Wrap around if necessary
	if idx == len(chr.sortedHashes) {
		idx = 0
	}
	
	return chr.nodes[chr.sortedHashes[idx]]
}

// hash computes hash for a key
func (chr *ConsistentHashRing) hash(key string) uint32 {
	h := md5.Sum([]byte(key))
	return uint32(h[0])<<24 | uint32(h[1])<<16 | uint32(h[2])<<8 | uint32(h[3])
}

// GetLoadBalancingStatistics returns load balancing statistics
func (elb *ExtendedLoadBalancer) GetLoadBalancingStatistics(serviceName string) map[string]interface{} {
	elb.mutex.RLock()
	defer elb.mutex.RUnlock()
	
	stats := make(map[string]interface{})
	
	config, exists := elb.configs[serviceName]
	if exists {
		stats["strategy"] = config.Strategy.String()
		stats["virtual_nodes"] = config.VirtualNodes
	}
	
	// Response time statistics
	responseTimeStats := make(map[string]interface{})
	for serviceID, tracker := range elb.responseTimeTracker {
		responseTimeStats[serviceID] = map[string]interface{}{
			"average_time": tracker.AverageTime,
			"sample_count": len(tracker.RecentTimes),
			"last_update":  tracker.LastUpdate,
		}
	}
	stats["response_times"] = responseTimeStats
	
	// Load statistics
	loadStats := make(map[string]interface{})
	for serviceID, tracker := range elb.loadTracker {
		loadStats[serviceID] = map[string]interface{}{
			"current_load":    tracker.CurrentLoad,
			"cpu_usage":       tracker.CPUUsage,
			"memory_usage":    tracker.MemoryUsage,
			"active_requests": tracker.ActiveRequests,
			"last_update":     tracker.LastUpdate,
		}
	}
	stats["load_stats"] = loadStats
	
	return stats
}

// CleanupStaleData removes stale tracking data
func (elb *ExtendedLoadBalancer) CleanupStaleData(maxAge time.Duration) {
	elb.mutex.Lock()
	defer elb.mutex.Unlock()
	
	cutoff := time.Now().Add(-maxAge)
	
	// Clean up response time trackers
	for serviceID, tracker := range elb.responseTimeTracker {
		if tracker.LastUpdate.Before(cutoff) {
			delete(elb.responseTimeTracker, serviceID)
		}
	}
	
	// Clean up load trackers
	for serviceID, tracker := range elb.loadTracker {
		if tracker.LastUpdate.Before(cutoff) {
			delete(elb.loadTracker, serviceID)
		}
	}
	
	elb.logger.Debug("Cleaned up stale load balancing data", "cutoff", cutoff)
}