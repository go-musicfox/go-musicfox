// Package kernel provides service metrics and monitoring functionality
package kernel

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// MetricsManager manages service performance metrics and monitoring
type MetricsManager struct {
	serviceMetrics map[string]*ServiceMetrics // serviceID -> metrics
	alerts         map[string]*ServiceAlert   // alertID -> alert
	thresholds     map[string]*AlertThresholds // serviceID -> thresholds
	mutex          sync.RWMutex
	logger         Logger
	alertHandlers  []AlertHandler
	startTime      time.Time
}

// AlertThresholds defines alert thresholds for a service
type AlertThresholds struct {
	ServiceID           string        `json:"service_id"`
	MaxResponseTime     time.Duration `json:"max_response_time"`
	MaxErrorRate        float64       `json:"max_error_rate"`
	MinThroughput       float64       `json:"min_throughput"`
	MaxFailureCount     int64         `json:"max_failure_count"`
	HealthCheckTimeout  time.Duration `json:"health_check_timeout"`
	Enabled             bool          `json:"enabled"`
}

// AlertHandler handles service alerts
type AlertHandler interface {
	HandleAlert(ctx context.Context, alert *ServiceAlert) error
}

// MetricsSample represents a single metrics sample
type MetricsSample struct {
	Timestamp    time.Time     `json:"timestamp"`
	ResponseTime time.Duration `json:"response_time"`
	Success      bool          `json:"success"`
	ErrorCode    string        `json:"error_code,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ServiceMetricsHistory stores historical metrics data
type ServiceMetricsHistory struct {
	ServiceID string           `json:"service_id"`
	Samples   []*MetricsSample `json:"samples"`
	MaxSamples int             `json:"max_samples"`
	mutex     sync.RWMutex     `json:"-"`
}

// NewMetricsManager creates a new metrics manager
func NewMetricsManager(logger Logger) *MetricsManager {
	return &MetricsManager{
		serviceMetrics: make(map[string]*ServiceMetrics),
		alerts:         make(map[string]*ServiceAlert),
		thresholds:     make(map[string]*AlertThresholds),
		logger:         logger,
		alertHandlers:  make([]AlertHandler, 0),
		startTime:      time.Now(),
	}
}

// RecordServiceCall records a service call for metrics
func (mm *MetricsManager) RecordServiceCall(serviceID string, duration time.Duration, success bool, errorCode string) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	
	metrics := mm.getOrCreateMetrics(serviceID)
	
	// Update metrics
	metrics.TotalRequests++
	metrics.LastRequestTime = time.Now()
	
	if success {
		metrics.SuccessfulRequests++
	} else {
		metrics.FailedRequests++
	}
	
	// Update response time statistics
	mm.updateResponseTimeStats(metrics, duration)
	
	// Calculate error rate
	metrics.ErrorRate = float64(metrics.FailedRequests) / float64(metrics.TotalRequests) * 100
	
	// Calculate throughput (requests per second)
	uptime := time.Since(mm.startTime)
	if uptime.Seconds() > 0 {
		metrics.Throughput = float64(metrics.TotalRequests) / uptime.Seconds()
	}
	
	metrics.Uptime = uptime
	
	// Check for alerts
	go mm.checkAlerts(serviceID, metrics, duration, success, errorCode)
	
	mm.logger.Debug("Service call recorded",
		"service_id", serviceID,
		"duration", duration,
		"success", success,
		"error_rate", metrics.ErrorRate)
	
	return nil
}

// GetServiceMetrics returns metrics for a specific service
func (mm *MetricsManager) GetServiceMetrics(serviceID string) (*ServiceMetrics, error) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	
	metrics, exists := mm.serviceMetrics[serviceID]
	if !exists {
		return nil, fmt.Errorf("metrics not found for service %s", serviceID)
	}
	
	// Return a copy to avoid concurrent access issues
	return mm.copyMetrics(metrics), nil
}

// GetServiceMetricsByName returns metrics for all instances of a service by name
func (mm *MetricsManager) GetServiceMetricsByName(serviceName string, registry ServiceRegistry) ([]*ServiceMetrics, error) {
	ctx := context.Background()
	instances, err := registry.Discover(ctx, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to discover service instances: %w", err)
	}
	
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	
	result := make([]*ServiceMetrics, 0)
	for _, instance := range instances {
		if metrics, exists := mm.serviceMetrics[instance.Info.ID]; exists {
			result = append(result, mm.copyMetrics(metrics))
		}
	}
	
	return result, nil
}

// GetAggregatedMetrics returns aggregated metrics for a service over a time range
func (mm *MetricsManager) GetAggregatedMetrics(serviceName string, timeRange time.Duration, registry ServiceRegistry) (*ServiceMetrics, error) {
	metricsList, err := mm.GetServiceMetricsByName(serviceName, registry)
	if err != nil {
		return nil, err
	}
	
	if len(metricsList) == 0 {
		return nil, fmt.Errorf("no metrics found for service %s", serviceName)
	}
	
	// Aggregate metrics
	aggregated := &ServiceMetrics{
		ServiceName: serviceName,
	}
	
	var totalRequests, successfulRequests, failedRequests int64
	var totalResponseTime time.Duration
	var minResponseTime, maxResponseTime time.Duration = time.Duration(math.MaxInt64), 0
	var totalThroughput float64
	
	for _, metrics := range metricsList {
		totalRequests += metrics.TotalRequests
		successfulRequests += metrics.SuccessfulRequests
		failedRequests += metrics.FailedRequests
		totalResponseTime += metrics.AverageResponseTime
		totalThroughput += metrics.Throughput
		
		if metrics.MinResponseTime < minResponseTime {
			minResponseTime = metrics.MinResponseTime
		}
		if metrics.MaxResponseTime > maxResponseTime {
			maxResponseTime = metrics.MaxResponseTime
		}
	}
	
	aggregated.TotalRequests = totalRequests
	aggregated.SuccessfulRequests = successfulRequests
	aggregated.FailedRequests = failedRequests
	aggregated.MinResponseTime = minResponseTime
	aggregated.MaxResponseTime = maxResponseTime
	aggregated.Throughput = totalThroughput
	
	if totalRequests > 0 {
		aggregated.ErrorRate = float64(failedRequests) / float64(totalRequests) * 100
		aggregated.AverageResponseTime = totalResponseTime / time.Duration(len(metricsList))
	}
	
	return aggregated, nil
}

// CreateAlert creates a new service alert
func (mm *MetricsManager) CreateAlert(alert *ServiceAlert) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	
	if alert.ID == "" {
		alert.ID = fmt.Sprintf("%s-%d", alert.ServiceID, time.Now().UnixNano())
	}
	
	alert.TriggeredAt = time.Now()
	alert.Status = ServiceAlertStatusActive
	
	mm.alerts[alert.ID] = alert
	
	mm.logger.Warn("Service alert created",
		"alert_id", alert.ID,
		"service_id", alert.ServiceID,
		"type", alert.Type,
		"severity", alert.Severity,
		"message", alert.Message)
	
	// Notify alert handlers
	go mm.notifyAlertHandlers(alert)
	
	return nil
}

// ResolveAlert resolves an active alert
func (mm *MetricsManager) ResolveAlert(alertID string) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	
	alert, exists := mm.alerts[alertID]
	if !exists {
		return fmt.Errorf("alert %s not found", alertID)
	}
	
	if alert.Status != ServiceAlertStatusActive {
		return fmt.Errorf("alert %s is not active", alertID)
	}
	
	now := time.Now()
	alert.ResolvedAt = &now
	alert.Status = ServiceAlertStatusResolved
	
	mm.logger.Info("Service alert resolved",
		"alert_id", alertID,
		"service_id", alert.ServiceID,
		"duration", now.Sub(alert.TriggeredAt))
	
	return nil
}

// GetActiveAlerts returns active alerts for a service
func (mm *MetricsManager) GetActiveAlerts(serviceID string) ([]*ServiceAlert, error) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	
	result := make([]*ServiceAlert, 0)
	for _, alert := range mm.alerts {
		if alert.ServiceID == serviceID && alert.Status == ServiceAlertStatusActive {
			result = append(result, alert)
		}
	}
	
	return result, nil
}

// GetAlertHistory returns alert history for a service within a time range
func (mm *MetricsManager) GetAlertHistory(serviceID string, timeRange time.Duration) ([]*ServiceAlert, error) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	
	cutoff := time.Now().Add(-timeRange)
	result := make([]*ServiceAlert, 0)
	
	for _, alert := range mm.alerts {
		if alert.ServiceID == serviceID && alert.TriggeredAt.After(cutoff) {
			result = append(result, alert)
		}
	}
	
	return result, nil
}

// SetAlertThresholds sets alert thresholds for a service
func (mm *MetricsManager) SetAlertThresholds(serviceID string, thresholds *AlertThresholds) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	
	thresholds.ServiceID = serviceID
	mm.thresholds[serviceID] = thresholds
	
	mm.logger.Info("Alert thresholds set",
		"service_id", serviceID,
		"max_response_time", thresholds.MaxResponseTime,
		"max_error_rate", thresholds.MaxErrorRate)
	
	return nil
}

// AddAlertHandler adds an alert handler
func (mm *MetricsManager) AddAlertHandler(handler AlertHandler) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	
	mm.alertHandlers = append(mm.alertHandlers, handler)
}

// ExportMetrics exports metrics in the specified format
func (mm *MetricsManager) ExportMetrics(format string, timeRange time.Duration) ([]byte, error) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	
	switch format {
	case "json":
		return mm.exportJSON(timeRange)
	case "prometheus":
		return mm.exportPrometheus(timeRange)
	case "csv":
		return mm.exportCSV(timeRange)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// Helper methods

func (mm *MetricsManager) getOrCreateMetrics(serviceID string) *ServiceMetrics {
	metrics, exists := mm.serviceMetrics[serviceID]
	if !exists {
		metrics = &ServiceMetrics{
			ServiceID:           serviceID,
			TotalRequests:       0,
			SuccessfulRequests:  0,
			FailedRequests:      0,
			AverageResponseTime: 0,
			MinResponseTime:     time.Duration(math.MaxInt64),
			MaxResponseTime:     0,
			ErrorRate:           0,
			Throughput:          0,
		}
		mm.serviceMetrics[serviceID] = metrics
	}
	return metrics
}

func (mm *MetricsManager) updateResponseTimeStats(metrics *ServiceMetrics, duration time.Duration) {
	// Update min/max response times
	if duration < metrics.MinResponseTime {
		metrics.MinResponseTime = duration
	}
	if duration > metrics.MaxResponseTime {
		metrics.MaxResponseTime = duration
	}
	
	// Calculate average response time using exponential moving average
	if metrics.AverageResponseTime == 0 {
		metrics.AverageResponseTime = duration
	} else {
		// EMA with alpha = 0.1
		alpha := 0.1
		metrics.AverageResponseTime = time.Duration(float64(metrics.AverageResponseTime)*(1-alpha) + float64(duration)*alpha)
	}
}

func (mm *MetricsManager) copyMetrics(metrics *ServiceMetrics) *ServiceMetrics {
	return &ServiceMetrics{
		ServiceID:           metrics.ServiceID,
		ServiceName:         metrics.ServiceName,
		TotalRequests:       metrics.TotalRequests,
		SuccessfulRequests:  metrics.SuccessfulRequests,
		FailedRequests:      metrics.FailedRequests,
		AverageResponseTime: metrics.AverageResponseTime,
		MinResponseTime:     metrics.MinResponseTime,
		MaxResponseTime:     metrics.MaxResponseTime,
		LastRequestTime:     metrics.LastRequestTime,
		Uptime:              metrics.Uptime,
		ErrorRate:           metrics.ErrorRate,
		Throughput:          metrics.Throughput,
	}
}

func (mm *MetricsManager) checkAlerts(serviceID string, metrics *ServiceMetrics, duration time.Duration, success bool, errorCode string) {
	thresholds, exists := mm.thresholds[serviceID]
	if !exists || !thresholds.Enabled {
		return
	}
	
	// Check response time threshold
	if thresholds.MaxResponseTime > 0 && duration > thresholds.MaxResponseTime {
		alert := &ServiceAlert{
			ServiceID: serviceID,
			Type:      ServiceAlertTypePerformance,
			Severity:  ServiceAlertSeverityHigh,
			Message:   fmt.Sprintf("Response time %v exceeds threshold %v", duration, thresholds.MaxResponseTime),
			Details: map[string]interface{}{
				"response_time": duration,
				"threshold":     thresholds.MaxResponseTime,
			},
		}
		mm.CreateAlert(alert)
	}
	
	// Check error rate threshold
	if thresholds.MaxErrorRate > 0 && metrics.ErrorRate > thresholds.MaxErrorRate {
		alert := &ServiceAlert{
			ServiceID: serviceID,
			Type:      ServiceAlertTypePerformance,
			Severity:  ServiceAlertSeverityMedium,
			Message:   fmt.Sprintf("Error rate %.2f%% exceeds threshold %.2f%%", metrics.ErrorRate, thresholds.MaxErrorRate),
			Details: map[string]interface{}{
				"error_rate": metrics.ErrorRate,
				"threshold":  thresholds.MaxErrorRate,
			},
		}
		mm.CreateAlert(alert)
	}
	
	// Check throughput threshold
	if thresholds.MinThroughput > 0 && metrics.Throughput < thresholds.MinThroughput {
		alert := &ServiceAlert{
			ServiceID: serviceID,
			Type:      ServiceAlertTypePerformance,
			Severity:  ServiceAlertSeverityMedium,
			Message:   fmt.Sprintf("Throughput %.2f req/s below threshold %.2f req/s", metrics.Throughput, thresholds.MinThroughput),
			Details: map[string]interface{}{
				"throughput": metrics.Throughput,
				"threshold":  thresholds.MinThroughput,
			},
		}
		mm.CreateAlert(alert)
	}
}

func (mm *MetricsManager) notifyAlertHandlers(alert *ServiceAlert) {
	ctx := context.Background()
	for _, handler := range mm.alertHandlers {
		go func(h AlertHandler) {
			if err := h.HandleAlert(ctx, alert); err != nil {
				mm.logger.Error("Alert handler failed", "error", err, "alert_id", alert.ID)
			}
		}(handler)
	}
}

// Export methods

func (mm *MetricsManager) exportJSON(timeRange time.Duration) ([]byte, error) {
	// Implementation for JSON export
	return []byte("{}"), nil // Placeholder
}

func (mm *MetricsManager) exportPrometheus(timeRange time.Duration) ([]byte, error) {
	// Implementation for Prometheus format export
	return []byte(""), nil // Placeholder
}

func (mm *MetricsManager) exportCSV(timeRange time.Duration) ([]byte, error) {
	// Implementation for CSV export
	return []byte(""), nil // Placeholder
}

// GetMetricsStatistics returns overall metrics statistics
func (mm *MetricsManager) GetMetricsStatistics() map[string]interface{} {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	
	stats := make(map[string]interface{})
	stats["total_services"] = len(mm.serviceMetrics)
	stats["total_alerts"] = len(mm.alerts)
	
	activeAlerts := 0
	for _, alert := range mm.alerts {
		if alert.Status == ServiceAlertStatusActive {
			activeAlerts++
		}
	}
	stats["active_alerts"] = activeAlerts
	
	var totalRequests int64
	var totalErrors int64
	for _, metrics := range mm.serviceMetrics {
		totalRequests += metrics.TotalRequests
		totalErrors += metrics.FailedRequests
	}
	stats["total_requests"] = totalRequests
	stats["total_errors"] = totalErrors
	
	if totalRequests > 0 {
		stats["overall_error_rate"] = float64(totalErrors) / float64(totalRequests) * 100
	}
	
	return stats
}