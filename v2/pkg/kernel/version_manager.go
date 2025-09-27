// Package kernel provides service version management functionality
package kernel

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// VersionManager manages service versions and compatibility
type VersionManager struct {
	serviceVersions map[string][]*ServiceVersionInfo // serviceName -> versions
	compatibility   map[string]*ServiceCompatibility // serviceName -> compatibility rules
	deprecations    map[string]*DeprecationInfo      // serviceName:version -> deprecation info
	mutex          sync.RWMutex
	logger         Logger
}

// ServiceVersionInfo represents detailed version information
type ServiceVersionInfo struct {
	ServiceName    string           `json:"service_name"`
	Version        *ServiceVersion  `json:"version"`
	Instances      []*ServiceInstance `json:"instances"`
	RegisteredAt   time.Time        `json:"registered_at"`
	DeprecatedAt   *time.Time       `json:"deprecated_at,omitempty"`
	RemovedAt      *time.Time       `json:"removed_at,omitempty"`
	Compatibility  *ServiceCompatibility `json:"compatibility"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// DeprecationInfo represents service deprecation information
type DeprecationInfo struct {
	ServiceName    string    `json:"service_name"`
	Version        *ServiceVersion `json:"version"`
	DeprecatedAt   time.Time `json:"deprecated_at"`
	RemovalDate    *time.Time `json:"removal_date,omitempty"`
	Reason         string    `json:"reason"`
	MigrationGuide string    `json:"migration_guide,omitempty"`
	Replacement    *ServiceVersion `json:"replacement,omitempty"`
}

// NewVersionManager creates a new version manager
func NewVersionManager(logger Logger) *VersionManager {
	return &VersionManager{
		serviceVersions: make(map[string][]*ServiceVersionInfo),
		compatibility:   make(map[string]*ServiceCompatibility),
		deprecations:    make(map[string]*DeprecationInfo),
		logger:         logger,
	}
}

// ParseVersion parses a version string into ServiceVersion
func ParseVersion(versionStr string) (*ServiceVersion, error) {
	if versionStr == "" {
		return nil, fmt.Errorf("version string cannot be empty")
	}
	
	// Regular expression for semantic versioning
	// Supports: major.minor.patch[-pre][+build]
	re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(?:-(\w+(?:\.\w+)*))?(?:\+(\w+(?:\.\w+)*))?$`)
	matches := re.FindStringSubmatch(versionStr)
	
	if matches == nil {
		return nil, fmt.Errorf("invalid version format: %s", versionStr)
	}
	
	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])
	
	version := &ServiceVersion{
		Major: major,
		Minor: minor,
		Patch: patch,
		Pre:   matches[4],
		Build: matches[5],
	}
	
	return version, nil
}

// RegisterServiceVersion registers a service with version information
func (vm *VersionManager) RegisterServiceVersion(serviceName string, version *ServiceVersion, instance *ServiceInstance) error {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()
	
	// Find or create version info
	versionInfo := vm.findVersionInfo(serviceName, version)
	if versionInfo == nil {
		versionInfo = &ServiceVersionInfo{
			ServiceName:  serviceName,
			Version:      version,
			Instances:    make([]*ServiceInstance, 0),
			RegisteredAt: time.Now(),
			Metadata:     make(map[string]interface{}),
		}
		
		if vm.serviceVersions[serviceName] == nil {
			vm.serviceVersions[serviceName] = make([]*ServiceVersionInfo, 0)
		}
		vm.serviceVersions[serviceName] = append(vm.serviceVersions[serviceName], versionInfo)
		
		// Sort versions
		vm.sortVersions(serviceName)
	}
	
	// Add instance to version
	versionInfo.Instances = append(versionInfo.Instances, instance)
	
	vm.logger.Info("Service version registered",
		"service_name", serviceName,
		"version", version.String(),
		"instance_id", instance.Info.ID)
	
	return nil
}

// DeregisterServiceVersion deregisters a service instance from version
func (vm *VersionManager) DeregisterServiceVersion(serviceName string, version *ServiceVersion, instanceID string) error {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()
	
	versionInfo := vm.findVersionInfo(serviceName, version)
	if versionInfo == nil {
		return fmt.Errorf("version %s not found for service %s", version.String(), serviceName)
	}
	
	// Remove instance from version
	for i, instance := range versionInfo.Instances {
		if instance.Info.ID == instanceID {
			versionInfo.Instances = append(versionInfo.Instances[:i], versionInfo.Instances[i+1:]...)
			break
		}
	}
	
	// Remove version if no instances left
	if len(versionInfo.Instances) == 0 {
		vm.removeVersionInfo(serviceName, version)
	}
	
	vm.logger.Info("Service version deregistered",
		"service_name", serviceName,
		"version", version.String(),
		"instance_id", instanceID)
	
	return nil
}

// GetServicesByVersion returns service instances for a specific version
func (vm *VersionManager) GetServicesByVersion(serviceName string, version *ServiceVersion) ([]*ServiceInstance, error) {
	vm.mutex.RLock()
	defer vm.mutex.RUnlock()
	
	versionInfo := vm.findVersionInfo(serviceName, version)
	if versionInfo == nil {
		return nil, fmt.Errorf("version %s not found for service %s", version.String(), serviceName)
	}
	
	// Filter healthy instances
	healthyInstances := make([]*ServiceInstance, 0)
	for _, instance := range versionInfo.Instances {
		if instance.GetState() == ServiceStateHealthy {
			healthyInstances = append(healthyInstances, instance)
		}
	}
	
	return healthyInstances, nil
}

// GetCompatibleServices returns services compatible with the required version
func (vm *VersionManager) GetCompatibleServices(serviceName string, requiredVersion *ServiceVersion) ([]*ServiceInstance, error) {
	vm.mutex.RLock()
	defer vm.mutex.RUnlock()
	
	versions := vm.serviceVersions[serviceName]
	if versions == nil {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}
	
	compatibleInstances := make([]*ServiceInstance, 0)
	
	for _, versionInfo := range versions {
		if vm.isVersionCompatible(versionInfo.Version, requiredVersion) {
			for _, instance := range versionInfo.Instances {
				if instance.GetState() == ServiceStateHealthy {
					compatibleInstances = append(compatibleInstances, instance)
				}
			}
		}
	}
	
	return compatibleInstances, nil
}

// CheckCompatibility checks compatibility between versions
func (vm *VersionManager) CheckCompatibility(serviceName string, requiredVersion *ServiceVersion) (*ServiceCompatibility, error) {
	vm.mutex.RLock()
	defer vm.mutex.RUnlock()
	
	compat := vm.compatibility[serviceName]
	if compat == nil {
		// Default compatibility: same major version
		compat = &ServiceCompatibility{
			MinVersion: &ServiceVersion{Major: requiredVersion.Major, Minor: 0, Patch: 0},
			MaxVersion: &ServiceVersion{Major: requiredVersion.Major + 1, Minor: 0, Patch: 0},
		}
	}
	
	return compat, nil
}

// ListServiceVersions returns all versions for a service
func (vm *VersionManager) ListServiceVersions(serviceName string) ([]*ServiceVersion, error) {
	vm.mutex.RLock()
	defer vm.mutex.RUnlock()
	
	versions := vm.serviceVersions[serviceName]
	if versions == nil {
		return []*ServiceVersion{}, nil
	}
	
	result := make([]*ServiceVersion, len(versions))
	for i, versionInfo := range versions {
		result[i] = versionInfo.Version
	}
	
	return result, nil
}

// DeprecateServiceVersion marks a service version as deprecated
func (vm *VersionManager) DeprecateServiceVersion(serviceName string, version *ServiceVersion, deprecationTime time.Time, reason string, replacement *ServiceVersion) error {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()
	
	versionInfo := vm.findVersionInfo(serviceName, version)
	if versionInfo == nil {
		return fmt.Errorf("version %s not found for service %s", version.String(), serviceName)
	}
	
	// Mark version as deprecated
	versionInfo.DeprecatedAt = &deprecationTime
	
	// Create deprecation info
	deprecationKey := fmt.Sprintf("%s:%s", serviceName, version.String())
	vm.deprecations[deprecationKey] = &DeprecationInfo{
		ServiceName:  serviceName,
		Version:      version,
		DeprecatedAt: deprecationTime,
		Reason:       reason,
		Replacement:  replacement,
	}
	
	vm.logger.Info("Service version deprecated",
		"service_name", serviceName,
		"version", version.String(),
		"reason", reason)
	
	return nil
}

// GetDeprecatedVersions returns all deprecated versions
func (vm *VersionManager) GetDeprecatedVersions() ([]*DeprecationInfo, error) {
	vm.mutex.RLock()
	defer vm.mutex.RUnlock()
	
	result := make([]*DeprecationInfo, 0, len(vm.deprecations))
	for _, deprecation := range vm.deprecations {
		result = append(result, deprecation)
	}
	
	return result, nil
}

// SetCompatibilityRules sets compatibility rules for a service
func (vm *VersionManager) SetCompatibilityRules(serviceName string, compatibility *ServiceCompatibility) error {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()
	
	vm.compatibility[serviceName] = compatibility
	
	vm.logger.Info("Compatibility rules set",
		"service_name", serviceName,
		"min_version", compatibility.MinVersion.String(),
		"max_version", compatibility.MaxVersion.String())
	
	return nil
}

// GetLatestVersion returns the latest version of a service
func (vm *VersionManager) GetLatestVersion(serviceName string) (*ServiceVersion, error) {
	vm.mutex.RLock()
	defer vm.mutex.RUnlock()
	
	versions := vm.serviceVersions[serviceName]
	if versions == nil || len(versions) == 0 {
		return nil, fmt.Errorf("no versions found for service %s", serviceName)
	}
	
	// Versions are sorted, so the last one is the latest
	return versions[len(versions)-1].Version, nil
}

// Helper methods

func (vm *VersionManager) findVersionInfo(serviceName string, version *ServiceVersion) *ServiceVersionInfo {
	versions := vm.serviceVersions[serviceName]
	if versions == nil {
		return nil
	}
	
	for _, versionInfo := range versions {
		if vm.versionsEqual(versionInfo.Version, version) {
			return versionInfo
		}
	}
	
	return nil
}

func (vm *VersionManager) removeVersionInfo(serviceName string, version *ServiceVersion) {
	versions := vm.serviceVersions[serviceName]
	if versions == nil {
		return
	}
	
	for i, versionInfo := range versions {
		if vm.versionsEqual(versionInfo.Version, version) {
			vm.serviceVersions[serviceName] = append(versions[:i], versions[i+1:]...)
			break
		}
	}
	
	// Remove service entry if no versions left
	if len(vm.serviceVersions[serviceName]) == 0 {
		delete(vm.serviceVersions, serviceName)
	}
}

func (vm *VersionManager) sortVersions(serviceName string) {
	versions := vm.serviceVersions[serviceName]
	if versions == nil {
		return
	}
	
	sort.Slice(versions, func(i, j int) bool {
		return vm.compareVersions(versions[i].Version, versions[j].Version) < 0
	})
}

func (vm *VersionManager) compareVersions(v1, v2 *ServiceVersion) int {
	if v1.Major != v2.Major {
		return v1.Major - v2.Major
	}
	if v1.Minor != v2.Minor {
		return v1.Minor - v2.Minor
	}
	if v1.Patch != v2.Patch {
		return v1.Patch - v2.Patch
	}
	
	// Compare pre-release versions
	if v1.Pre == "" && v2.Pre != "" {
		return 1 // Release version is greater than pre-release
	}
	if v1.Pre != "" && v2.Pre == "" {
		return -1 // Pre-release is less than release
	}
	
	return strings.Compare(v1.Pre, v2.Pre)
}

func (vm *VersionManager) versionsEqual(v1, v2 *ServiceVersion) bool {
	return v1.Major == v2.Major &&
		v1.Minor == v2.Minor &&
		v1.Patch == v2.Patch &&
		v1.Pre == v2.Pre &&
		v1.Build == v2.Build
}

func (vm *VersionManager) isVersionCompatible(available, required *ServiceVersion) bool {
	// Basic compatibility: same major version
	if available.Major != required.Major {
		return false
	}
	
	// Available version should be >= required version
	return vm.compareVersions(available, required) >= 0
}

// GetVersionStatistics returns statistics about service versions
func (vm *VersionManager) GetVersionStatistics() map[string]interface{} {
	vm.mutex.RLock()
	defer vm.mutex.RUnlock()
	
	stats := make(map[string]interface{})
	stats["total_services"] = len(vm.serviceVersions)
	stats["total_versions"] = vm.getTotalVersionCount()
	stats["deprecated_versions"] = len(vm.deprecations)
	
	versionDistribution := make(map[string]int)
	for serviceName, versions := range vm.serviceVersions {
		versionDistribution[serviceName] = len(versions)
	}
	stats["version_distribution"] = versionDistribution
	
	return stats
}

func (vm *VersionManager) getTotalVersionCount() int {
	total := 0
	for _, versions := range vm.serviceVersions {
		total += len(versions)
	}
	return total
}