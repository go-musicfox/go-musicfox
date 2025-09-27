// Package thirdparty 实现沙箱安全机制
package thirdparty

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Sandbox 沙箱实现
type Sandbox struct {
	config     *SandboxConfig
	policy     *SecurityPolicy
	violations []SecurityViolation
	mu         sync.RWMutex
	active     bool
}

// SecurityViolation 安全违规记录
type SecurityViolation struct {
	Type        ViolationType `json:"type"`         // 违规类型
	Description string        `json:"description"`  // 描述
	Timestamp   time.Time     `json:"timestamp"`    // 时间戳
	Severity    Severity      `json:"severity"`     // 严重程度
	Blocked     bool          `json:"blocked"`      // 是否被阻止
	Details     map[string]interface{} `json:"details"` // 详细信息
}

// ViolationType 违规类型枚举
type ViolationType int

const (
	ViolationTypeFileAccess ViolationType = iota // 文件访问违规
	ViolationTypeNetworkAccess                   // 网络访问违规
	ViolationTypeSyscall                         // 系统调用违规
	ViolationTypeResourceLimit                   // 资源限制违规
	ViolationTypePermission                      // 权限违规
	ViolationTypeUnsafeOperation                 // 不安全操作违规
)

// String 返回违规类型的字符串表示
func (v ViolationType) String() string {
	switch v {
	case ViolationTypeFileAccess:
		return "file_access"
	case ViolationTypeNetworkAccess:
		return "network_access"
	case ViolationTypeSyscall:
		return "syscall"
	case ViolationTypeResourceLimit:
		return "resource_limit"
	case ViolationTypePermission:
		return "permission"
	case ViolationTypeUnsafeOperation:
		return "unsafe_operation"
	default:
		return "unknown"
	}
}

// Severity 严重程度枚举
type Severity int

const (
	SeverityLow Severity = iota    // 低
	SeverityMedium                 // 中
	SeverityHigh                   // 高
	SeverityCritical               // 严重
)

// String 返回严重程度的字符串表示
func (s Severity) String() string {
	switch s {
	case SeverityLow:
		return "low"
	case SeverityMedium:
		return "medium"
	case SeverityHigh:
		return "high"
	case SeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// NewSandbox 创建新的沙箱
func NewSandbox(config *SandboxConfig) (*Sandbox, error) {
	if config == nil {
		return nil, fmt.Errorf("sandbox config cannot be nil")
	}

	s := &Sandbox{
		config:     config,
		violations: make([]SecurityViolation, 0),
	}

	// 初始化默认安全策略
	s.initializeDefaultPolicy()

	return s, nil
}

// initializeDefaultPolicy 初始化默认安全策略
func (s *Sandbox) initializeDefaultPolicy() {
	s.policy = &SecurityPolicy{
		AllowUnsafeOperations: false,
		TrustedDomains:        []string{"localhost", "127.0.0.1"},
		BlockedDomains:        []string{},
		MaxRequestSize:        1024 * 1024, // 1MB
		RateLimitRPS:          10,           // 10 requests per second
	}
}

// Activate 激活沙箱
func (s *Sandbox) Activate() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.active {
		return fmt.Errorf("sandbox already active")
	}

	// 根据隔离级别设置沙箱
	switch s.config.IsolationLevel {
	case IsolationLevelNone:
		// 无隔离，不做任何限制
	case IsolationLevelBasic:
		// 基础隔离，限制文件和网络访问
		if err := s.setupBasicIsolation(); err != nil {
			return fmt.Errorf("failed to setup basic isolation: %w", err)
		}
	case IsolationLevelStrict:
		// 严格隔离，限制大部分系统调用
		if err := s.setupStrictIsolation(); err != nil {
			return fmt.Errorf("failed to setup strict isolation: %w", err)
		}
	case IsolationLevelComplete:
		// 完全隔离，最严格的限制
		if err := s.setupCompleteIsolation(); err != nil {
			return fmt.Errorf("failed to setup complete isolation: %w", err)
		}
	}

	s.active = true
	return nil
}

// Deactivate 停用沙箱
func (s *Sandbox) Deactivate() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active {
		return fmt.Errorf("sandbox not active")
	}

	// 清理隔离设置
	if err := s.cleanupIsolation(); err != nil {
		return fmt.Errorf("failed to cleanup isolation: %w", err)
	}

	s.active = false
	return nil
}

// Execute 在沙箱中执行函数
func (s *Sandbox) Execute(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.active {
		return nil, fmt.Errorf("sandbox not active")
	}

	// 创建执行上下文
	execCtx := &ExecutionContext{
		Timeout:     30 * time.Second,
		MemoryLimit: 64 * 1024 * 1024,
		GasLimit:    1000000,
		Metadata:    make(map[string]interface{}),
	}

	// 在受限环境中执行
	return s.executeWithLimits(ctx, execCtx, fn)
}

// executeWithLimits 在限制条件下执行函数
func (s *Sandbox) executeWithLimits(ctx context.Context, execCtx *ExecutionContext, fn func() (interface{}, error)) (interface{}, error) {
	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, execCtx.Timeout)
	defer cancel()

	// 执行结果通道
	resultChan := make(chan *ExecutionResult, 1)

	// 在goroutine中执行函数
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- &ExecutionResult{
					Error:   fmt.Sprintf("panic: %v", r),
					Success: false,
				}
			}
		}()

		startTime := time.Now()
		result, err := fn()
		duration := time.Since(startTime)

		execResult := &ExecutionResult{
			Value:    result,
			Duration: duration,
			Success:  err == nil,
		}

		if err != nil {
			execResult.Error = err.Error()
		}

		resultChan <- execResult
	}()

	// 等待执行完成或超时
	select {
	case result := <-resultChan:
		if !result.Success {
			return nil, fmt.Errorf("execution failed: %s", result.Error)
		}
		return result.Value, nil
	case <-timeoutCtx.Done():
		return nil, fmt.Errorf("execution timeout after %v", execCtx.Timeout)
	}
}

// ValidateFileAccess 验证文件访问
func (s *Sandbox) ValidateFileAccess(path string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.config.FileSystemAccess {
		violation := SecurityViolation{
			Type:        ViolationTypeFileAccess,
			Description: fmt.Sprintf("file system access denied: %s", path),
			Timestamp:   time.Now(),
			Severity:    SeverityHigh,
			Blocked:     true,
			Details:     map[string]interface{}{"path": path},
		}
		s.recordViolation(violation)
		return fmt.Errorf("file system access denied")
	}

	// 检查路径是否在允许列表中
	allowed := false
	for _, allowedPath := range s.config.AllowedPaths {
		if s.isPathAllowed(path, allowedPath) {
			allowed = true
			break
		}
	}

	if !allowed {
		violation := SecurityViolation{
			Type:        ViolationTypeFileAccess,
			Description: fmt.Sprintf("file access to unauthorized path: %s", path),
			Timestamp:   time.Now(),
			Severity:    SeverityMedium,
			Blocked:     true,
			Details:     map[string]interface{}{"path": path},
		}
		s.recordViolation(violation)
		return fmt.Errorf("access to path %s denied", path)
	}

	return nil
}

// ValidateNetworkAccess 验证网络访问
func (s *Sandbox) ValidateNetworkAccess(host string, port int) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.config.NetworkAccess {
		violation := SecurityViolation{
			Type:        ViolationTypeNetworkAccess,
			Description: fmt.Sprintf("network access denied: %s:%d", host, port),
			Timestamp:   time.Now(),
			Severity:    SeverityHigh,
			Blocked:     true,
			Details:     map[string]interface{}{"host": host, "port": port},
		}
		s.recordViolation(violation)
		return fmt.Errorf("network access denied")
	}

	// 检查主机是否在允许列表中
	allowed := false
	for _, allowedNetwork := range s.config.AllowedNetworks {
		if s.isNetworkAllowed(host, allowedNetwork) {
			allowed = true
			break
		}
	}

	if !allowed {
		violation := SecurityViolation{
			Type:        ViolationTypeNetworkAccess,
			Description: fmt.Sprintf("network access to unauthorized host: %s:%d", host, port),
			Timestamp:   time.Now(),
			Severity:    SeverityMedium,
			Blocked:     true,
			Details:     map[string]interface{}{"host": host, "port": port},
		}
		s.recordViolation(violation)
		return fmt.Errorf("access to %s:%d denied", host, port)
	}

	return nil
}

// UpdateConfig 更新沙箱配置
func (s *Sandbox) UpdateConfig(config *SandboxConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if config == nil {
		return fmt.Errorf("sandbox config cannot be nil")
	}

	s.config = config
	return nil
}

// GetViolations 获取安全违规记录
func (s *Sandbox) GetViolations() []SecurityViolation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]SecurityViolation{}, s.violations...)
}

// ClearViolations 清除违规记录
func (s *Sandbox) ClearViolations() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.violations = make([]SecurityViolation, 0)
}

// Cleanup 清理沙箱资源
func (s *Sandbox) Cleanup() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.active {
		if err := s.cleanupIsolation(); err != nil {
			return fmt.Errorf("failed to cleanup isolation: %w", err)
		}
		s.active = false
	}

	// 清除违规记录
	s.violations = nil

	return nil
}

// recordViolation 记录安全违规
func (s *Sandbox) recordViolation(violation SecurityViolation) {
	// 限制违规记录数量，避免内存泄漏
	if len(s.violations) >= 1000 {
		// 移除最旧的记录
		s.violations = s.violations[1:]
	}
	s.violations = append(s.violations, violation)
}

// isPathAllowed 检查路径是否被允许
func (s *Sandbox) isPathAllowed(path, allowedPath string) bool {
	// 清理路径
	cleanPath := filepath.Clean(path)
	cleanAllowedPath := filepath.Clean(allowedPath)

	// 检查是否在允许的路径下
	rel, err := filepath.Rel(cleanAllowedPath, cleanPath)
	if err != nil {
		return false
	}

	// 如果相对路径不以".."开头，说明在允许的路径下
	return !strings.HasPrefix(rel, "..")
}

// isNetworkAllowed 检查网络访问是否被允许
func (s *Sandbox) isNetworkAllowed(host, allowedNetwork string) bool {
	// 简单的字符串匹配，实际实现中应该支持CIDR等格式
	return strings.Contains(host, allowedNetwork) || host == allowedNetwork
}

// setupBasicIsolation 设置基础隔离
func (s *Sandbox) setupBasicIsolation() error {
	// 在实际实现中，这里会设置系统级别的隔离
	// 例如使用 seccomp、namespaces 等技术
	// 这里只是一个简化的实现
	return nil
}

// setupStrictIsolation 设置严格隔离
func (s *Sandbox) setupStrictIsolation() error {
	// 设置更严格的隔离策略
	return nil
}

// setupCompleteIsolation 设置完全隔离
func (s *Sandbox) setupCompleteIsolation() error {
	// 设置最严格的隔离策略
	return nil
}

// cleanupIsolation 清理隔离设置
func (s *Sandbox) cleanupIsolation() error {
	// 清理系统级别的隔离设置
	return nil
}

// GetStats 获取沙箱统计信息
func (s *Sandbox) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["active"] = s.active
	stats["isolation_level"] = s.config.IsolationLevel.String()
	stats["violation_count"] = len(s.violations)
	stats["network_access"] = s.config.NetworkAccess
	stats["filesystem_access"] = s.config.FileSystemAccess
	stats["goroutines"] = runtime.NumGoroutine()

	// 统计违规类型
	violationTypes := make(map[string]int)
	for _, violation := range s.violations {
		violationTypes[violation.Type.String()]++
	}
	stats["violation_types"] = violationTypes

	return stats
}