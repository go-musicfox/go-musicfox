// pkg/kernel/security.go
package kernel

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/knadh/koanf/v2"
)

// SecurityManager 安全管理器接口
type SecurityManager interface {
	// 插件安全验证
	ValidatePlugin(pluginPath string, pluginType PluginType) error
	VerifySignature(pluginPath string, signature []byte) error
	CheckTrustedSource(source string) bool

	// 沙箱机制
	CreateSandbox(pluginID string, limits ResourceLimits) (Sandbox, error)
	DestroySandbox(sandboxID string) error
	GetSandbox(sandboxID string) (Sandbox, error)

	// 权限控制
	CheckPermission(pluginID string, resource string, action string) (bool, error)
	GrantPermission(pluginID string, resource string, actions []string) error
	RevokePermission(pluginID string, resource string, actions []string) error
	GetPermissions(pluginID string) map[string][]string

	// ACL控制
	AddACLRule(pluginID, resource, action string, allowed bool) error
RemoveACLRule(pluginID, resource, action string) error
	CheckACL(pluginID, resource, action string) (bool, error)

	// 资源监控
	MonitorResources(pluginID string) (*ResourceUsage, error)
	SetResourceLimits(pluginID string, limits *ResourceLimits) error
	GetResourceLimits(pluginID string) (*ResourceLimits, error)

	// 安全策略
	LoadSecurityPolicy(policyPath string) error
	UpdateSecurityPolicy(policy *SecurityPolicy) error
	GetSecurityPolicy() *SecurityPolicy
	IsPluginBlocked(pluginID string) bool

	// 安全审计
	LogSecurityEvent(event *SecurityEvent) error
	GetSecurityEvents(filter *SecurityEventFilter) ([]*SecurityEvent, error)
	ClearSecurityEvents(before time.Time) error

	// 生命周期管理
	Initialize(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// PluginType 插件类型
type PluginType string

const (
	PluginTypeDynamicLibrary PluginType = "dynamic_library"
	PluginTypeRPC           PluginType = "rpc"
	PluginTypeWebAssembly   PluginType = "webassembly"
	PluginTypeHotReload     PluginType = "hot_reload"
)

// ResourceLimits 资源限制
type ResourceLimits struct {
	MaxMemory             int64         `json:"max_memory"`             // 最大内存使用量（字节）
	MaxCPU                float64       `json:"max_cpu"`                // 最大CPU使用率（0-1）
	MaxCPUPercent         float64       `json:"max_cpu_percent"`        // 最大CPU使用百分比
	MaxDiskIO             int64         `json:"max_disk_io"`            // 最大磁盘IO（字节/秒）
	MaxNetworkIO          int64         `json:"max_network_io"`         // 最大网络IO（字节/秒）
	Timeout               time.Duration `json:"timeout"`                // 执行超时时间
	MaxGoroutines         int           `json:"max_goroutines"`         // 最大协程数
	MaxFileHandles        int           `json:"max_file_handles"`        // 最大文件句柄数
	MaxFileDescriptors    int           `json:"max_file_descriptors"`    // 最大文件描述符数
	MaxNetworkConnections int           `json:"max_network_connections"` // 最大网络连接数
}

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	MemoryUsed            int64     `json:"memory_used"`            // 当前内存使用量
	CPUUsage              float64   `json:"cpu_usage"`              // 当前CPU使用率
	CPUUsed               float64   `json:"cpu_used"`               // 当前CPU使用量
	DiskIORate            int64     `json:"disk_io_rate"`           // 当前磁盘IO速率
	NetworkIORate         int64     `json:"network_io_rate"`        // 当前网络IO速率
	GoroutineCount        int       `json:"goroutine_count"`        // 当前协程数
	FileHandleCount       int       `json:"file_handle_count"`      // 当前文件句柄数
	FileDescriptorsUsed   int       `json:"file_descriptors_used"`  // 当前文件描述符使用数
	NetworkConnections    int       `json:"network_connections"`    // 当前网络连接数
	LastUpdated           time.Time `json:"last_updated"`           // 最后更新时间
}

// Sandbox 沙箱接口
type Sandbox interface {
	GetID() string
	GetPluginID() string
	GetResourceLimits() ResourceLimits
	GetResourceUsage() (*ResourceUsage, error)
	Execute(ctx context.Context, fn func() error) error
	IsActive() bool
	Destroy() error
	GetStatus() SandboxStatus
	GetLimits() ResourceLimits
	UpdateResourceUsage() error
}

// SecurityLevel 安全级别
type SecurityLevel int

const (
	SecurityLevelLow SecurityLevel = iota
	SecurityLevelMedium
	SecurityLevelHigh
	SecurityLevelCritical
)

// SecurityPolicy 安全策略
type SecurityPolicy struct {
	Version           string                    `json:"version"`
	DefaultPermissions map[string][]string      `json:"default_permissions"`
	PluginPermissions  map[string]map[string][]string `json:"plugin_permissions"`
	TrustedSources    []string                  `json:"trusted_sources"`
	SignatureRequired bool                      `json:"signature_required"`
	SandboxEnabled    bool                      `json:"sandbox_enabled"`
	ResourceLimits    ResourceLimits            `json:"resource_limits"`
	AuditEnabled      bool                      `json:"audit_enabled"`
	EnforceSignatureVerification bool           `json:"enforce_signature_verification"`
	AllowUnsignedPlugins         bool           `json:"allow_unsigned_plugins"`
	EnableSandbox               bool           `json:"enable_sandbox"`
	BlockedPlugins              []string       `json:"blocked_plugins"`
	MaxPluginSize               int64          `json:"max_plugin_size"`
	AllowedFileExtensions       []string       `json:"allowed_file_extensions"`
	SecurityLevel               SecurityLevel  `json:"security_level"`
	CreatedAt         time.Time                 `json:"created_at"`
	UpdatedAt         time.Time                 `json:"updated_at"`
}

// SecurityEvent 安全事件
type SecurityEvent struct {
	ID          string                 `json:"id"`
	Type        SecurityEventType      `json:"type"`
	Level       SecurityEventLevel     `json:"level"`
	PluginID    string                 `json:"plugin_id"`
	Resource    string                 `json:"resource"`
	Action      string                 `json:"action"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details"`
	Timestamp   time.Time              `json:"timestamp"`
	SourceIP    string                 `json:"source_ip"`
	UserAgent   string                 `json:"user_agent"`
}

// SecurityEventType 安全事件类型
type SecurityEventType string

const (
	SecurityEventTypePluginLoad       SecurityEventType = "plugin_load"
	SecurityEventTypePluginUnload     SecurityEventType = "plugin_unload"
	SecurityEventTypePermissionDenied SecurityEventType = "permission_denied"
	SecurityEventTypeResourceExceeded SecurityEventType = "resource_exceeded"
	SecurityEventTypeSignatureInvalid SecurityEventType = "signature_invalid"
	SecurityEventTypeSandboxViolation SecurityEventType = "sandbox_violation"
	SecurityEventTypeUnauthorizedAccess SecurityEventType = "unauthorized_access"
	SecurityEventTypePolicyViolation  SecurityEventType = "policy_violation"
	SecurityEventTypeAudit            SecurityEventType = "audit"
	SecurityEventTypeSandboxCreate    SecurityEventType = "sandbox_create"
	SecurityEventTypeSandboxDestroy   SecurityEventType = "sandbox_destroy"
	SecurityEventTypeAccessDenied     SecurityEventType = "access_denied"
	SecurityEventTypeResourceLimit    SecurityEventType = "resource_limit"
	SecurityEventTypeACLCreate        SecurityEventType = "acl_create"
	SecurityEventTypeACLDelete        SecurityEventType = "acl_delete"
	SecurityEventTypeResourceMonitor  SecurityEventType = "resource_monitor"
	SecurityEventTypePolicy           SecurityEventType = "policy"
	SecurityEventTypeError            SecurityEventType = "error"
)

// SecurityEventLevel 安全事件级别
type SecurityEventLevel string

const (
	SecurityEventLevelInfo    SecurityEventLevel = "info"
	SecurityEventLevelWarning SecurityEventLevel = "warning"
	SecurityEventLevelError   SecurityEventLevel = "error"
	SecurityEventLevelCritical SecurityEventLevel = "critical"
)

// SandboxStatus 沙箱状态
type SandboxStatus string

const (
	SandboxStatusCreated   SandboxStatus = "created"
	SandboxStatusRunning   SandboxStatus = "running"
	SandboxStatusStopped   SandboxStatus = "stopped"
	SandboxStatusDestroyed SandboxStatus = "destroyed"
)

// 安全管理器错误定义
var (
	ErrSecurityManagerNotInitialized = errors.New("security manager not initialized")
	ErrInvalidPlugin                 = errors.New("invalid plugin")
	ErrPluginNotFound                = errors.New("plugin not found")
	ErrInvalidSignature              = errors.New("invalid signature")
	ErrUntrustedSource               = errors.New("untrusted source")
	ErrSandboxNotFound               = errors.New("sandbox not found")
	ErrSandboxCreationFailed         = errors.New("sandbox creation failed")
	ErrPermissionDenied              = errors.New("permission denied")
	ErrResourceLimitExceeded         = errors.New("resource limit exceeded")
	ErrInvalidSecurityPolicy         = errors.New("invalid security policy")
	ErrSecurityEventNotFound         = errors.New("security event not found")
	ErrAuditDisabled                 = errors.New("audit logging is disabled")
	ErrInvalidACLRule                = errors.New("invalid ACL rule")
	ErrACLRuleNotFound               = errors.New("ACL rule not found")
	ErrResourceMonitoringFailed      = errors.New("resource monitoring failed")
)

// SecurityEventFilter 安全事件过滤器
type SecurityEventFilter struct {
	Types     []SecurityEventType  `json:"types"`
	Levels    []SecurityEventLevel `json:"levels"`
	PluginIDs []string             `json:"plugin_ids"`
	Resource  string               `json:"resource"`
	Action    string               `json:"action"`
	StartTime *time.Time           `json:"start_time"`
	EndTime   *time.Time           `json:"end_time"`
	Limit     int                  `json:"limit"`
	Offset    int                  `json:"offset"`
}

// Permission 权限定义
type Permission struct {
	Resource string   `json:"resource"`
	Actions  []string `json:"actions"`
}

// SandboxConfig 沙箱配置
type SandboxConfig struct {
	Enabled         bool              `json:"enabled"`
	AllowedPaths    []string          `json:"allowed_paths"`
	BlockedPaths    []string          `json:"blocked_paths"`
	ResourceLimits  ResourceLimits    `json:"resource_limits"`
	NetworkAccess   bool              `json:"network_access"`
	EnvironmentVars map[string]string `json:"environment_vars"`
}



// ACLRule ACL规则
type ACLRule struct {
	ID        string    `json:"id"`
	Subject   string    `json:"subject"`
	Resource  string    `json:"resource"`
	Action    string    `json:"action"`
	Actions   []string  `json:"actions"`
	Subjects  []string  `json:"subjects"`
	Effect    string    `json:"effect"` // "allow" or "deny"
	Priority  int       `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
}

// SecurityStatistics 安全统计信息
type SecurityStatistics struct {
	TotalEvents        int                              `json:"total_events"`
	EventsByType       map[SecurityEventType]int        `json:"events_by_type"`
	EventsByLevel      map[SecurityEventLevel]int       `json:"events_by_level"`
	EventsByPlugin     map[string]int                   `json:"events_by_plugin"`
	TotalPlugins       int                              `json:"total_plugins"`
	ActivePlugins      int                              `json:"active_plugins"`
	SecurityViolations int                              `json:"security_violations"`
	PermissionDenials  int                              `json:"permission_denials"`
	SandboxViolations  int                              `json:"sandbox_violations"`
	SignatureFailures  int                              `json:"signature_failures"`
	GeneratedAt        time.Time                        `json:"generated_at"`
}

// ACLEntry 访问控制列表条目
type ACLEntry struct {
	PluginID    string       `json:"plugin_id"`
	Permissions []Permission `json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// SandboxImpl 沙箱实现
type SandboxImpl struct {
	ID        string         `json:"id"`
	PluginID  string         `json:"plugin_id"`
	Limits    ResourceLimits `json:"limits"`
	Resources *ResourceUsage `json:"resources"`
	Status    SandboxStatus  `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	mutex     sync.RWMutex
}

// GetID 获取沙箱ID
func (s *SandboxImpl) GetID() string {
	return s.ID
}

// GetPluginID 获取插件ID
func (s *SandboxImpl) GetPluginID() string {
	return s.PluginID
}

// GetResourceLimits 获取资源限制
func (s *SandboxImpl) GetResourceLimits() ResourceLimits {
	return s.Limits
}

// GetResourceUsage 获取资源使用情况
func (s *SandboxImpl) GetResourceUsage() (*ResourceUsage, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.Resources, nil
}

// Execute 在沙箱中执行函数
func (s *SandboxImpl) Execute(ctx context.Context, fn func() error) error {
	// 这里应该实现沙箱执行逻辑
	return fn()
}

// GetStatus 获取沙箱状态
func (s *SandboxImpl) GetStatus() SandboxStatus {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.Status
}

// GetLimits 获取资源限制
func (s *SandboxImpl) GetLimits() ResourceLimits {
	return s.Limits
}

// UpdateResourceUsage 更新资源使用情况
func (s *SandboxImpl) UpdateResourceUsage() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// 这里应该实现实际的资源使用情况更新逻辑
	s.Resources.LastUpdated = time.Now()
	return nil
}

// IsActive 检查沙箱是否活跃
func (s *SandboxImpl) IsActive() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.Status == SandboxStatusRunning
}

// Destroy 销毁沙箱
func (s *SandboxImpl) Destroy() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Status = SandboxStatusStopped
	return nil
}

// SecurityManagerImpl 安全管理器实现
type SecurityManagerImpl struct {
	config         *koanf.Koanf
	logger         *slog.Logger
	policy         *SecurityPolicy
	securityPolicy *SecurityPolicy
	acl            map[string]*ACLEntry
	aclRules       map[string]*ACLRule
	permissions    map[string][]Permission
	sandboxes      map[string]Sandbox
	securityEvents []*SecurityEvent
	resourceLimits map[string]*ResourceLimits
	resourceUsage  map[string]*ResourceUsage
	monitoringTickers map[string]*time.Ticker
	publicKey      *rsa.PublicKey

	mutex      sync.RWMutex
	eventMutex sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
}

// NewSecurityManager 创建新的安全管理器
func NewSecurityManager(config *koanf.Koanf, logger *slog.Logger) SecurityManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &SecurityManagerImpl{
		config:            config,
		logger:            logger,
		acl:               make(map[string]*ACLEntry),
		aclRules:          make(map[string]*ACLRule),
		permissions:       make(map[string][]Permission),
		sandboxes:         make(map[string]Sandbox),
		securityEvents:    make([]*SecurityEvent, 0),
		resourceLimits:    make(map[string]*ResourceLimits),
		resourceUsage:     make(map[string]*ResourceUsage),
		monitoringTickers: make(map[string]*time.Ticker),
		ctx:               ctx,
		cancel:            cancel,
	}
}

// Initialize 初始化安全管理器
func (sm *SecurityManagerImpl) Initialize(ctx context.Context) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 加载默认安全策略
	if err := sm.loadDefaultPolicy(); err != nil {
		return fmt.Errorf("failed to load default security policy: %w", err)
	}

	// 加载公钥用于签名验证
	if err := sm.loadPublicKey(); err != nil {
		sm.logger.Warn("Failed to load public key for signature verification", "error", err)
	}

	sm.logger.Info("Security manager initialized successfully")
	return nil
}

// Start 启动安全管理器
func (sm *SecurityManagerImpl) Start(ctx context.Context) error {
	sm.logger.Info("Security manager started")
	return nil
}

// Stop 停止安全管理器
func (sm *SecurityManagerImpl) Stop(ctx context.Context) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 销毁所有沙箱
	for _, sandbox := range sm.sandboxes {
		if err := sandbox.Destroy(); err != nil {
			sm.logger.Warn("Failed to destroy sandbox", "sandbox_id", sandbox.GetID(), "error", err)
		}
	}

	sm.logger.Info("Security manager stopped")
	return nil
}

// Shutdown 关闭安全管理器
func (sm *SecurityManagerImpl) Shutdown(ctx context.Context) error {
	if sm.cancel != nil {
		sm.cancel()
	}

	sm.logger.Info("Security manager shutdown")
	return nil
}

// loadDefaultPolicy 加载默认安全策略
func (sm *SecurityManagerImpl) loadDefaultPolicy() error {
	sm.policy = &SecurityPolicy{
		Version:           "1.0.0",
		DefaultPermissions: make(map[string][]string),
		PluginPermissions:  make(map[string]map[string][]string),
		TrustedSources:    []string{"localhost", "127.0.0.1"},
		SignatureRequired: false,
		SandboxEnabled:    true,
		ResourceLimits: ResourceLimits{
			MaxMemory:      100 * 1024 * 1024, // 100MB
			MaxCPU:        0.5,                // 50%
			MaxDiskIO:     10 * 1024 * 1024,   // 10MB/s
			MaxNetworkIO:  5 * 1024 * 1024,    // 5MB/s
			Timeout:       30 * time.Second,
			MaxGoroutines: 100,
			MaxFileHandles: 50,
		},
		AuditEnabled: true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 设置默认权限
	sm.policy.DefaultPermissions["file"] = []string{"read"}
	sm.policy.DefaultPermissions["network"] = []string{"connect"}
	sm.policy.DefaultPermissions["system"] = []string{"info"}

	return nil
}

// loadPublicKey 加载公钥
func (sm *SecurityManagerImpl) loadPublicKey() error {
	// 这里应该从配置文件或密钥存储中加载公钥
	// 暂时返回nil，表示未配置签名验证
	return nil
}

// generateEventID 生成事件ID
func (sm *SecurityManagerImpl) generateEventID() string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return fmt.Sprintf("sec_%d_%x", time.Now().UnixNano(), hash[:8])
}

// generateRandomString 生成随机字符串
func (sm *SecurityManagerImpl) generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// validatePluginPath 验证插件路径
func (sm *SecurityManagerImpl) validatePluginPath(pluginPath string) error {
	if pluginPath == "" {
		return fmt.Errorf("plugin path cannot be empty")
	}
	
	// 检查路径是否包含危险字符
	dangerousPatterns := []string{"..", "~", "$"}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(pluginPath, pattern) {
			return fmt.Errorf("plugin path contains dangerous pattern: %s", pattern)
		}
	}
	
	return nil
}

// validatePluginType 验证插件类型
func (sm *SecurityManagerImpl) validatePluginType(pluginType PluginType) error {
	validTypes := []PluginType{
		PluginTypeDynamicLibrary,
		PluginTypeRPC,
		PluginTypeWebAssembly,
		PluginTypeHotReload,
	}
	
	for _, validType := range validTypes {
		if pluginType == validType {
			return nil
		}
	}
	
	return fmt.Errorf("unsupported plugin type: %s", pluginType)
}

// validatePluginSignature 验证插件签名
func (sm *SecurityManagerImpl) validatePluginSignature(pluginPath string) error {
	if sm.publicKey == nil {
		return fmt.Errorf("public key not configured")
	}
	
	// 这里应该实现实际的签名验证逻辑
	// 暂时返回nil表示验证通过
	return nil
}

// readPluginFile 读取插件文件
func (sm *SecurityManagerImpl) readPluginFile(pluginPath string) ([]byte, error) {
	// 这里应该实现安全的文件读取逻辑
	// 暂时返回空数据
	return []byte{}, nil
}

// verifyRSASignature 验证RSA签名
func (sm *SecurityManagerImpl) verifyRSASignature(hash, signature []byte) error {
	if sm.publicKey == nil {
		return fmt.Errorf("public key not configured")
	}
	
	// 这里应该实现实际的RSA签名验证
	// 暂时返回nil表示验证通过
	return nil
}

// ValidatePlugin 验证插件安全性
func (sm *SecurityManagerImpl) ValidatePlugin(pluginPath string, pluginType PluginType) error {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// 检查插件路径是否安全
	if err := sm.validatePluginPath(pluginPath); err != nil {
		sm.logSecurityEvent(SecurityEventTypePluginLoad, SecurityEventLevelError, "", pluginPath, "validate", 
			fmt.Sprintf("Plugin path validation failed: %v", err), nil)
		return fmt.Errorf("plugin path validation failed: %w", err)
	}

	// 检查插件类型是否支持
	if err := sm.validatePluginType(pluginType); err != nil {
		sm.logSecurityEvent(SecurityEventTypePluginLoad, SecurityEventLevelError, "", pluginPath, "validate", 
			fmt.Sprintf("Plugin type validation failed: %v", err), nil)
		return fmt.Errorf("plugin type validation failed: %w", err)
	}

	// 如果启用了签名验证，检查插件签名
	if sm.policy.SignatureRequired {
		if err := sm.validatePluginSignature(pluginPath); err != nil {
			sm.logSecurityEvent(SecurityEventTypeSignatureInvalid, SecurityEventLevelError, "", pluginPath, "validate", 
				fmt.Sprintf("Plugin signature validation failed: %v", err), nil)
			return fmt.Errorf("plugin signature validation failed: %w", err)
		}
	}

	sm.logSecurityEvent(SecurityEventTypePluginLoad, SecurityEventLevelInfo, "", pluginPath, "validate", 
		"Plugin validation successful", map[string]interface{}{"type": string(pluginType)})

	return nil
}

// VerifySignature 验证插件数字签名
func (sm *SecurityManagerImpl) VerifySignature(pluginPath string, signature []byte) error {
	if sm.publicKey == nil {
		return fmt.Errorf("public key not configured")
	}

	// 读取插件文件内容
	pluginData, err := sm.readPluginFile(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to read plugin file: %w", err)
	}

	// 计算文件哈希
	hash := sha256.Sum256(pluginData)

	// 验证签名
	if err := sm.verifyRSASignature(hash[:], signature); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

// CheckTrustedSource 检查是否为可信源
func (sm *SecurityManagerImpl) CheckTrustedSource(source string) bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	for _, trustedSource := range sm.policy.TrustedSources {
		if source == trustedSource {
			return true
		}
	}

	return false
}

// fileExists 检查文件是否存在
func (sm *SecurityManagerImpl) fileExists(path string) bool {
	// 这里应该实现实际的文件存在检查
	// 暂时返回true
	return true
}



// CreateSandbox 创建插件沙箱环境
func (sm *SecurityManagerImpl) CreateSandbox(pluginID string, limits ResourceLimits) (Sandbox, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 检查插件是否已有沙箱
	if _, exists := sm.sandboxes[pluginID]; exists {
		return nil, fmt.Errorf("sandbox already exists for plugin: %s", pluginID)
	}

	// 创建沙箱实例
	sandbox := &SandboxImpl{
		ID:        fmt.Sprintf("sandbox_%s_%d", pluginID, time.Now().UnixNano()),
		PluginID:  pluginID,
		Limits:    limits,
		Status:    SandboxStatusCreated,
		CreatedAt: time.Now(),
		Resources: &ResourceUsage{
			MemoryUsed:         0,
			CPUUsage:          0,
			DiskIORate:        0,
			NetworkIORate:     0,
			GoroutineCount:    0,
			FileHandleCount:   0,
			LastUpdated:       time.Now(),
		},
	}

	// 保存沙箱
	sm.sandboxes[pluginID] = sandbox

	sm.logSecurityEvent(SecurityEventTypeSandboxCreate, SecurityEventLevelInfo, pluginID, 
		sandbox.ID, "create", "Sandbox created successfully", 
		map[string]interface{}{"sandbox_id": sandbox.ID, "limits": limits})

	return sandbox, nil
}

// DestroySandbox 销毁插件沙箱环境
func (sm *SecurityManagerImpl) DestroySandbox(sandboxID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 查找沙箱
	var foundSandbox Sandbox
	var foundPluginID string
	for pluginID, sandbox := range sm.sandboxes {
		if sandbox.GetID() == sandboxID {
			foundSandbox = sandbox
			foundPluginID = pluginID
			break
		}
	}

	if foundSandbox == nil {
		return fmt.Errorf("sandbox not found: %s", sandboxID)
	}

	// 销毁沙箱
	if err := foundSandbox.Destroy(); err != nil {
		sm.logSecurityEvent(SecurityEventTypeSandboxDestroy, SecurityEventLevelWarning, foundPluginID, 
			sandboxID, "destroy", fmt.Sprintf("Failed to destroy sandbox: %v", err), nil)
	}

	// 移除沙箱
	delete(sm.sandboxes, foundPluginID)

	sm.logSecurityEvent(SecurityEventTypeSandboxDestroy, SecurityEventLevelInfo, foundPluginID, 
		sandboxID, "destroy", "Sandbox destroyed successfully", nil)

	return nil
}

// GetSandbox 获取插件沙箱信息
func (sm *SecurityManagerImpl) GetSandbox(sandboxID string) (Sandbox, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// 查找沙箱
	for _, sandbox := range sm.sandboxes {
		if sandbox.GetID() == sandboxID {
			return sandbox, nil
		}
	}

	return nil, fmt.Errorf("sandbox not found: %s", sandboxID)
}

// EnforceSandbox 强制执行沙箱限制
func (sm *SecurityManagerImpl) EnforceSandbox(sandbox Sandbox) error {
	// 检查沙箱状态
	if sandbox.GetStatus() != SandboxStatusRunning {
		return fmt.Errorf("sandbox %s is not running", sandbox.GetID())
	}

	// 更新资源使用情况
	if err := sandbox.UpdateResourceUsage(); err != nil {
		sm.logSecurityEvent(SecurityEventTypeResourceLimit, SecurityEventLevelWarning, sandbox.GetPluginID(), 
			sandbox.GetID(), "enforce", fmt.Sprintf("Failed to update resource usage: %v", err), nil)
		return fmt.Errorf("failed to update resource usage: %w", err)
	}

	// 检查资源限制
	resources, err := sandbox.GetResourceUsage()
	if err != nil {
		return fmt.Errorf("failed to get resource usage: %w", err)
	}
	limits := sandbox.GetLimits()

	// 检查内存限制
	if resources.MemoryUsed > limits.MaxMemory {
		sm.logSecurityEvent(SecurityEventTypeResourceLimit, SecurityEventLevelWarning, sandbox.GetPluginID(), 
			sandbox.GetID(), "enforce", "Memory limit exceeded", 
			map[string]interface{}{"used": resources.MemoryUsed, "limit": limits.MaxMemory})
		return fmt.Errorf("memory limit exceeded: %d > %d", resources.MemoryUsed, limits.MaxMemory)
	}

	// 检查CPU限制
	if resources.CPUUsage > limits.MaxCPUPercent {
		sm.logSecurityEvent(SecurityEventTypeResourceLimit, SecurityEventLevelWarning, sandbox.GetPluginID(), 
			sandbox.GetID(), "enforce", "CPU limit exceeded", 
			map[string]interface{}{"used": resources.CPUUsage, "limit": limits.MaxCPUPercent})
		return fmt.Errorf("CPU limit exceeded: %.2f > %.2f", resources.CPUUsage, limits.MaxCPUPercent)
	}

	return nil
}



// CheckPermission 检查插件权限
func (sm *SecurityManagerImpl) CheckPermission(pluginID, resource, action string) (bool, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// 获取插件权限
	pluginPerms, exists := sm.permissions[pluginID]
	if !exists {
		return false, nil
	}

	// 检查是否有该权限
	for _, perm := range pluginPerms {
		if perm.Resource == resource {
			for _, allowedAction := range perm.Actions {
				if allowedAction == action {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// GrantPermission 授予插件权限
func (sm *SecurityManagerImpl) GrantPermission(pluginID, resource string, actions []string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 初始化插件权限列表
	if sm.permissions[pluginID] == nil {
		sm.permissions[pluginID] = make([]Permission, 0)
	}

	// 查找现有权限或创建新权限
	for i, perm := range sm.permissions[pluginID] {
		if perm.Resource == resource {
			// 添加新动作到现有权限
			for _, action := range actions {
				// 检查动作是否已存在
				alreadyExists := false
				for _, existingAction := range perm.Actions {
					if existingAction == action {
						alreadyExists = true
						break
					}
				}
				if !alreadyExists {
					sm.permissions[pluginID][i].Actions = append(sm.permissions[pluginID][i].Actions, action)
				}
			}
			return nil
		}
	}

	// 创建新权限
	newPermission := Permission{
		Resource: resource,
		Actions:  make([]string, len(actions)),
	}
	copy(newPermission.Actions, actions)
	sm.permissions[pluginID] = append(sm.permissions[pluginID], newPermission)

	return nil
}

// RevokePermission 撤销插件权限
func (sm *SecurityManagerImpl) RevokePermission(pluginID, resource string, actions []string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	pluginPerms, exists := sm.permissions[pluginID]
	if !exists {
		return fmt.Errorf("no permissions found for plugin %s", pluginID)
	}

	// 查找并移除权限
	for i, perm := range pluginPerms {
		if perm.Resource == resource {
			// 查找并移除特定动作
			for _, action := range actions {
				for j, existingAction := range perm.Actions {
					if existingAction == action {
						// 移除动作
						sm.permissions[pluginID][i].Actions = append(perm.Actions[:j], perm.Actions[j+1:]...)
						break
					}
				}
			}
			// 如果没有剩余动作，移除整个权限
			if len(sm.permissions[pluginID][i].Actions) == 0 {
				sm.permissions[pluginID] = append(pluginPerms[:i], pluginPerms[i+1:]...)
			}
			return nil
		}
	}

	return fmt.Errorf("permissions for resource %s not found for plugin %s", resource, pluginID)
}

// GetPermissions 获取插件权限列表
func (sm *SecurityManagerImpl) GetPermissions(pluginID string) map[string][]string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	pluginPerms, exists := sm.permissions[pluginID]
	if !exists {
		return make(map[string][]string)
	}

	// 转换为map格式
	result := make(map[string][]string)
	for _, perm := range pluginPerms {
		result[perm.Resource] = make([]string, len(perm.Actions))
		copy(result[perm.Resource], perm.Actions)
	}
	return result
}

// CreateACLRule 创建访问控制规则
func (sm *SecurityManagerImpl) CreateACLRule(rule *ACLRule) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 验证规则
	if err := sm.validateACLRule(rule); err != nil {
		return fmt.Errorf("invalid ACL rule: %w", err)
	}

	// 生成规则ID
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s_%s_%s", rule.Subject, rule.Resource, rule.Action)))
	rule.ID = fmt.Sprintf("acl_%d_%x", time.Now().UnixNano(), hash[:8])
	rule.CreatedAt = time.Now()

	// 添加规则
	sm.aclRules[rule.ID] = rule

	sm.logSecurityEvent(SecurityEventTypeACLCreate, SecurityEventLevelInfo, rule.Subject, 
		rule.Resource, rule.Action, "ACL rule created successfully", 
		map[string]interface{}{"rule_id": rule.ID, "effect": rule.Effect})

	return nil
}

// AddACLRule 添加访问控制规则
func (sm *SecurityManagerImpl) AddACLRule(pluginID, resource, action string, allowed bool) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 创建ACL规则
	rule := &ACLRule{
		Subject:  pluginID,
		Resource: resource,
		Action:   action,
		Effect:   "deny",
		Priority: 0,
	}

	if allowed {
		rule.Effect = "allow"
	}

	// 验证规则
	if err := sm.validateACLRule(rule); err != nil {
		return fmt.Errorf("invalid ACL rule: %w", err)
	}

	// 生成规则ID
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s_%s_%s", rule.Subject, rule.Resource, rule.Action)))
	rule.ID = fmt.Sprintf("acl_%d_%x", time.Now().UnixNano(), hash[:8])
	rule.CreatedAt = time.Now()

	// 添加规则
	sm.aclRules[rule.ID] = rule

	sm.logSecurityEvent(SecurityEventTypeACLCreate, SecurityEventLevelInfo, rule.Subject, 
		rule.Resource, rule.Action, "ACL rule created successfully", 
		map[string]interface{}{"rule_id": rule.ID, "effect": rule.Effect})

	return nil
}

// RemoveACLRule 移除访问控制规则 (DeleteACLRule的别名)
func (sm *SecurityManagerImpl) RemoveACLRule(pluginID, resource, action string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 查找匹配的规则
	var ruleToDelete *ACLRule
	var ruleIDToDelete string
	for ruleID, rule := range sm.aclRules {
		if rule.Subject == pluginID && rule.Resource == resource && rule.Action == action {
			ruleToDelete = rule
			ruleIDToDelete = ruleID
			break
		}
	}

	if ruleToDelete == nil {
		return fmt.Errorf("ACL rule not found for plugin %s, resource %s, action %s", pluginID, resource, action)
	}

	// 直接删除规则，避免调用DeleteACLRule造成死锁
	delete(sm.aclRules, ruleIDToDelete)
	sm.logSecurityEvent(SecurityEventTypeACLDelete, SecurityEventLevelInfo, ruleToDelete.Subject, 
		ruleToDelete.Resource, ruleToDelete.Action, "ACL rule deleted successfully", 
		map[string]interface{}{"rule_id": ruleIDToDelete})

	return nil
}

// DeleteACLRule 删除访问控制规则
func (sm *SecurityManagerImpl) DeleteACLRule(ruleID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	rule, exists := sm.aclRules[ruleID]
	if !exists {
		return fmt.Errorf("ACL rule not found: %s", ruleID)
	}

	// 移除规则
	delete(sm.aclRules, ruleID)
	sm.logSecurityEvent(SecurityEventTypeACLDelete, SecurityEventLevelInfo, rule.Subject, 
		rule.Resource, rule.Action, "ACL rule deleted successfully", 
		map[string]interface{}{"rule_id": ruleID})

	return nil
}

// CheckACL 检查访问控制列表
func (sm *SecurityManagerImpl) CheckACL(subject, resource, action string) (bool, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// 默认拒绝
	allowed := false

	// 遍历ACL规则
	for _, rule := range sm.aclRules {
		if sm.matchACLRule(rule, subject, resource, action) {
			if rule.Effect == "allow" {
				allowed = true
			} else if rule.Effect == "deny" {
				// 拒绝规则优先级更高
				return false, nil
			}
		}
	}

	return allowed, nil
}

// GetACLRules 获取访问控制规则列表
func (sm *SecurityManagerImpl) GetACLRules() []*ACLRule {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// 返回规则副本
	result := make([]*ACLRule, 0, len(sm.aclRules))
	for _, rule := range sm.aclRules {
		result = append(result, rule)
	}
	return result
}

// validateACLRule 验证ACL规则
func (sm *SecurityManagerImpl) validateACLRule(rule *ACLRule) error {
	if rule.Subject == "" {
		return fmt.Errorf("subject cannot be empty")
	}
	if rule.Resource == "" {
		return fmt.Errorf("resource cannot be empty")
	}
	if rule.Action == "" {
		return fmt.Errorf("action cannot be empty")
	}
	if rule.Effect != "allow" && rule.Effect != "deny" {
		return fmt.Errorf("invalid effect: %s", rule.Effect)
	}
	return nil
}

// matchACLRule 匹配ACL规则
func (sm *SecurityManagerImpl) matchACLRule(rule *ACLRule, subject, resource, action string) bool {
	// 简单的字符串匹配，可以扩展为支持通配符
	subjectMatch := rule.Subject == "*" || rule.Subject == subject
	resourceMatch := rule.Resource == "*" || rule.Resource == resource
	actionMatch := rule.Action == "*" || rule.Action == action

	return subjectMatch && resourceMatch && actionMatch
}

// SetResourceLimit 设置插件资源限制
func (sm *SecurityManagerImpl) SetResourceLimit(pluginID string, limits *ResourceLimits) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 验证资源限制
	if err := sm.validateResourceLimits(limits); err != nil {
		return fmt.Errorf("invalid resource limits: %w", err)
	}

	// 设置资源限制
	sm.resourceLimits[pluginID] = limits

	sm.logSecurityEvent(SecurityEventTypeResourceLimit, SecurityEventLevelInfo, pluginID, 
		"resource_limits", "set", "Resource limits set successfully", 
		map[string]interface{}{
			"max_memory": limits.MaxMemory,
			"max_cpu": limits.MaxCPUPercent,
			"max_file_descriptors": limits.MaxFileDescriptors,
			"max_network_connections": limits.MaxNetworkConnections,
		})

	return nil
}

// GetResourceLimit 获取插件资源限制
func (sm *SecurityManagerImpl) GetResourceLimit(pluginID string) (*ResourceLimits, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	limits, exists := sm.resourceLimits[pluginID]
	if !exists {
		return nil, fmt.Errorf("resource limits not found for plugin: %s", pluginID)
	}

	return limits, nil
}

// MonitorResourceUsage 监控插件资源使用情况
func (sm *SecurityManagerImpl) MonitorResourceUsage(pluginID string) (*ResourceUsage, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	usage, exists := sm.resourceUsage[pluginID]
	if !exists {
		return nil, fmt.Errorf("resource usage not found for plugin: %s", pluginID)
	}

	return usage, nil
}

// UpdateResourceUsage 更新插件资源使用情况
func (sm *SecurityManagerImpl) UpdateResourceUsage(pluginID string, usage *ResourceUsage) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 更新资源使用情况
	sm.resourceUsage[pluginID] = usage

	// 检查是否超出限制
	if err := sm.checkResourceUsageViolation(pluginID, usage); err != nil {
		sm.logSecurityEvent(SecurityEventTypeResourceLimit, SecurityEventLevelWarning, pluginID, 
			"resource_usage", "violation", fmt.Sprintf("Resource usage violation: %v", err), 
			map[string]interface{}{
				"memory_used": usage.MemoryUsed,
				"cpu_used": usage.CPUUsed,
				"file_descriptors_used": usage.FileDescriptorsUsed,
				"network_connections": usage.NetworkConnections,
			})
		return err
	}

	return nil
}

// GetResourceUsageReport 获取资源使用报告
func (sm *SecurityManagerImpl) GetResourceUsageReport() map[string]*ResourceUsage {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// 返回所有插件的资源使用情况副本
	report := make(map[string]*ResourceUsage)
	for pluginID, usage := range sm.resourceUsage {
		report[pluginID] = &ResourceUsage{
			MemoryUsed:          usage.MemoryUsed,
			CPUUsed:            usage.CPUUsed,
			FileDescriptorsUsed: usage.FileDescriptorsUsed,
			NetworkConnections:  usage.NetworkConnections,
			LastUpdated:        usage.LastUpdated,
		}
	}

	return report
}

// StartResourceMonitoring 启动资源监控
func (sm *SecurityManagerImpl) StartResourceMonitoring(pluginID string, interval time.Duration) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 检查是否已在监控
	if _, exists := sm.monitoringTickers[pluginID]; exists {
		return fmt.Errorf("resource monitoring already started for plugin: %s", pluginID)
	}

	// 创建定时器
	ticker := time.NewTicker(interval)
	sm.monitoringTickers[pluginID] = ticker

	// 启动监控协程
	go sm.resourceMonitoringLoop(pluginID, ticker)

	sm.logSecurityEvent(SecurityEventTypeResourceMonitor, SecurityEventLevelInfo, pluginID, 
		"monitoring", "start", "Resource monitoring started", 
		map[string]interface{}{"interval": interval.String()})

	return nil
}

// StopResourceMonitoring 停止资源监控
func (sm *SecurityManagerImpl) StopResourceMonitoring(pluginID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	ticker, exists := sm.monitoringTickers[pluginID]
	if !exists {
		return fmt.Errorf("resource monitoring not found for plugin: %s", pluginID)
	}

	// 停止定时器
	ticker.Stop()
	delete(sm.monitoringTickers, pluginID)

	sm.logSecurityEvent(SecurityEventTypeResourceMonitor, SecurityEventLevelInfo, pluginID, 
		"monitoring", "stop", "Resource monitoring stopped", nil)

	return nil
}

// validateResourceLimits 验证资源限制
func (sm *SecurityManagerImpl) validateResourceLimits(limits *ResourceLimits) error {
	if limits.MaxMemory < 0 {
		return fmt.Errorf("max memory cannot be negative")
	}
	if limits.MaxCPUPercent < 0 || limits.MaxCPUPercent > 100 {
		return fmt.Errorf("max CPU percent must be between 0 and 100")
	}
	if limits.MaxFileDescriptors < 0 {
		return fmt.Errorf("max file descriptors cannot be negative")
	}
	if limits.MaxNetworkConnections < 0 {
		return fmt.Errorf("max network connections cannot be negative")
	}
	return nil
}

// checkResourceUsageViolation 检查资源使用违规
func (sm *SecurityManagerImpl) checkResourceUsageViolation(pluginID string, usage *ResourceUsage) error {
	limits, exists := sm.resourceLimits[pluginID]
	if !exists {
		// 如果没有设置限制，使用默认限制
		limits = &ResourceLimits{
			MaxMemory:             100 * 1024 * 1024, // 100MB
			MaxCPUPercent:         50,                // 50%
			MaxFileDescriptors:    100,
			MaxNetworkConnections: 10,
		}
	}

	if usage.MemoryUsed > limits.MaxMemory {
		return fmt.Errorf("memory usage exceeded: %d > %d", usage.MemoryUsed, limits.MaxMemory)
	}
	if usage.CPUUsed > limits.MaxCPUPercent {
		return fmt.Errorf("CPU usage exceeded: %.2f > %.2f", usage.CPUUsed, limits.MaxCPUPercent)
	}
	if usage.FileDescriptorsUsed > limits.MaxFileDescriptors {
		return fmt.Errorf("file descriptor usage exceeded: %d > %d", 
			usage.FileDescriptorsUsed, limits.MaxFileDescriptors)
	}
	if usage.NetworkConnections > limits.MaxNetworkConnections {
		return fmt.Errorf("network connection usage exceeded: %d > %d", 
			usage.NetworkConnections, limits.MaxNetworkConnections)
	}

	return nil
}

// resourceMonitoringLoop 资源监控循环
func (sm *SecurityManagerImpl) resourceMonitoringLoop(pluginID string, ticker *time.Ticker) {
	for range ticker.C {
		// 获取当前资源使用情况
		usage := sm.collectResourceUsage(pluginID)
		if usage != nil {
			usage.LastUpdated = time.Now()
			// 更新资源使用情况
			if err := sm.UpdateResourceUsage(pluginID, usage); err != nil {
				// 如果资源使用超限，可以采取相应措施
				sm.handleResourceViolation(pluginID, err)
			}
		}
	}
}

// collectResourceUsage 收集资源使用情况
func (sm *SecurityManagerImpl) collectResourceUsage(pluginID string) *ResourceUsage {
	// 这里应该实现实际的资源使用情况收集逻辑
	// 暂时返回模拟数据
	return &ResourceUsage{
		MemoryUsed:          1024 * 1024, // 1MB
		CPUUsed:            10.0,         // 10%
		FileDescriptorsUsed: 5,
		NetworkConnections:  2,
		LastUpdated:        time.Now(),
	}
}

// handleResourceViolation 处理资源违规
func (sm *SecurityManagerImpl) handleResourceViolation(pluginID string, err error) {
	// 这里可以实现资源违规的处理逻辑
	// 例如：暂停插件、发送警告等
	sm.logSecurityEvent(SecurityEventTypeResourceLimit, SecurityEventLevelCritical, pluginID, 
		"resource_violation", "handle", fmt.Sprintf("Handling resource violation: %v", err), nil)
}

// SetSecurityPolicy 设置安全策略
func (sm *SecurityManagerImpl) SetSecurityPolicy(policy *SecurityPolicy) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 验证安全策略
	if err := sm.validateSecurityPolicy(policy); err != nil {
		return fmt.Errorf("invalid security policy: %w", err)
	}

	// 设置安全策略
	sm.securityPolicy = policy

	sm.logSecurityEvent(SecurityEventTypePolicy, SecurityEventLevelInfo, "", 
		"security_policy", "set", "Security policy updated successfully", 
		map[string]interface{}{
			"enforce_signature_verification": policy.EnforceSignatureVerification,
			"allow_unsigned_plugins": policy.AllowUnsignedPlugins,
			"enable_sandbox": policy.EnableSandbox,
			"default_permissions": policy.DefaultPermissions,
		})

	return nil
}

// GetSecurityPolicy 获取当前安全策略
func (sm *SecurityManagerImpl) GetSecurityPolicy() *SecurityPolicy {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if sm.securityPolicy == nil {
		return sm.getDefaultSecurityPolicy()
	}

	// 返回策略副本
	defaultPermsCopy := make(map[string][]string)
	for k, v := range sm.securityPolicy.DefaultPermissions {
		defaultPermsCopy[k] = append([]string{}, v...)
	}
	
	return &SecurityPolicy{
		EnforceSignatureVerification: sm.securityPolicy.EnforceSignatureVerification,
		AllowUnsignedPlugins:        sm.securityPolicy.AllowUnsignedPlugins,
		EnableSandbox:              sm.securityPolicy.EnableSandbox,
		DefaultPermissions:         defaultPermsCopy,
		TrustedSources:             append([]string{}, sm.securityPolicy.TrustedSources...),
		BlockedPlugins:             append([]string{}, sm.securityPolicy.BlockedPlugins...),
		MaxPluginSize:              sm.securityPolicy.MaxPluginSize,
		AllowedFileExtensions:      append([]string{}, sm.securityPolicy.AllowedFileExtensions...),
		SecurityLevel:              sm.securityPolicy.SecurityLevel,
		AuditEnabled:               sm.securityPolicy.AuditEnabled,
	}
}

// UpdateSecurityPolicy 更新安全策略的特定字段
func (sm *SecurityManagerImpl) UpdateSecurityPolicy(policy *SecurityPolicy) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if policy == nil {
		return fmt.Errorf("security policy cannot be nil")
	}

	// 验证策略
	if err := sm.validateSecurityPolicy(policy); err != nil {
		return fmt.Errorf("invalid policy: %w", err)
	}

	// 更新策略
	sm.securityPolicy = policy

	sm.logSecurityEvent(SecurityEventTypePolicy, SecurityEventLevelInfo, "", 
		"security_policy", "update", "Security policy updated", nil)

	return nil
}

// AddTrustedSource 添加可信源
func (sm *SecurityManagerImpl) AddTrustedSource(source string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.securityPolicy == nil {
		sm.securityPolicy = sm.getDefaultSecurityPolicy()
	}

	// 检查是否已存在
	for _, existing := range sm.securityPolicy.TrustedSources {
		if existing == source {
			return fmt.Errorf("trusted source already exists: %s", source)
		}
	}

	// 添加可信源
	sm.securityPolicy.TrustedSources = append(sm.securityPolicy.TrustedSources, source)

	sm.logSecurityEvent(SecurityEventTypePolicy, SecurityEventLevelInfo, "", 
		"trusted_source", "add", fmt.Sprintf("Added trusted source: %s", source), nil)

	return nil
}

// RemoveTrustedSource 移除可信源
func (sm *SecurityManagerImpl) RemoveTrustedSource(source string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.securityPolicy == nil {
		return fmt.Errorf("no security policy configured")
	}

	// 查找并移除
	for i, existing := range sm.securityPolicy.TrustedSources {
		if existing == source {
			sm.securityPolicy.TrustedSources = append(
				sm.securityPolicy.TrustedSources[:i],
				sm.securityPolicy.TrustedSources[i+1:]...)

			sm.logSecurityEvent(SecurityEventTypePolicy, SecurityEventLevelInfo, "", 
				"trusted_source", "remove", fmt.Sprintf("Removed trusted source: %s", source), nil)

			return nil
		}
	}

	return fmt.Errorf("trusted source not found: %s", source)
}

// AddBlockedPlugin 添加被阻止的插件
func (sm *SecurityManagerImpl) AddBlockedPlugin(pluginID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.securityPolicy == nil {
		sm.securityPolicy = sm.getDefaultSecurityPolicy()
	}

	// 检查是否已存在
	for _, existing := range sm.securityPolicy.BlockedPlugins {
		if existing == pluginID {
			return fmt.Errorf("plugin already blocked: %s", pluginID)
		}
	}

	// 添加到阻止列表
	sm.securityPolicy.BlockedPlugins = append(sm.securityPolicy.BlockedPlugins, pluginID)

	sm.logSecurityEvent(SecurityEventTypePolicy, SecurityEventLevelWarning, pluginID, 
		"blocked_plugin", "add", fmt.Sprintf("Added plugin to block list: %s", pluginID), nil)

	return nil
}

// RemoveBlockedPlugin 移除被阻止的插件
func (sm *SecurityManagerImpl) RemoveBlockedPlugin(pluginID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.securityPolicy == nil {
		return fmt.Errorf("no security policy configured")
	}

	// 查找并移除
	for i, existing := range sm.securityPolicy.BlockedPlugins {
		if existing == pluginID {
			sm.securityPolicy.BlockedPlugins = append(
				sm.securityPolicy.BlockedPlugins[:i],
				sm.securityPolicy.BlockedPlugins[i+1:]...)

			sm.logSecurityEvent(SecurityEventTypePolicy, SecurityEventLevelInfo, pluginID, 
				"blocked_plugin", "remove", fmt.Sprintf("Removed plugin from block list: %s", pluginID), nil)

			return nil
		}
	}

	return fmt.Errorf("blocked plugin not found: %s", pluginID)
}

// IsPluginBlocked 检查插件是否被阻止
func (sm *SecurityManagerImpl) IsPluginBlocked(pluginID string) bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if sm.securityPolicy == nil {
		return false
	}

	for _, blocked := range sm.securityPolicy.BlockedPlugins {
		if blocked == pluginID {
			return true
		}
	}

	return false
}

// validateSecurityPolicy 验证安全策略
func (sm *SecurityManagerImpl) validateSecurityPolicy(policy *SecurityPolicy) error {
	if policy == nil {
		return fmt.Errorf("security policy cannot be nil")
	}

	if policy.MaxPluginSize < 0 {
		return fmt.Errorf("max plugin size cannot be negative")
	}

	if policy.SecurityLevel < SecurityLevelLow || policy.SecurityLevel > SecurityLevelCritical {
		return fmt.Errorf("invalid security level: %d", policy.SecurityLevel)
	}

	// 验证文件扩展名格式
	for _, ext := range policy.AllowedFileExtensions {
		if !strings.HasPrefix(ext, ".") {
			return fmt.Errorf("file extension must start with dot: %s", ext)
		}
	}

	return nil
}

// getDefaultSecurityPolicy 获取默认安全策略
func (sm *SecurityManagerImpl) getDefaultSecurityPolicy() *SecurityPolicy {
	return &SecurityPolicy{
		EnforceSignatureVerification: true,
		AllowUnsignedPlugins:        false,
		EnableSandbox:              true,
		DefaultPermissions:         map[string][]string{"default": {"read"}},
		TrustedSources:             []string{},
		BlockedPlugins:             []string{},
		MaxPluginSize:              10 * 1024 * 1024, // 10MB
		AllowedFileExtensions:      []string{".so", ".dll", ".dylib"},
		SecurityLevel:              SecurityLevelMedium,
		AuditEnabled:               true,
	}
}

// GetSecurityEvents 获取安全事件列表
func (sm *SecurityManagerImpl) GetSecurityEvents(filter *SecurityEventFilter) ([]*SecurityEvent, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var filteredEvents []*SecurityEvent

	for _, event := range sm.securityEvents {
		if sm.matchesFilter(event, filter) {
			filteredEvents = append(filteredEvents, event)
		}
	}

	// 按时间倒序排列
	sort.Slice(filteredEvents, func(i, j int) bool {
		return filteredEvents[i].Timestamp.After(filteredEvents[j].Timestamp)
	})

	// 应用限制
	if filter != nil && filter.Limit > 0 && len(filteredEvents) > filter.Limit {
		filteredEvents = filteredEvents[:filter.Limit]
	}

	return filteredEvents, nil
}

// GetSecurityEventsByPlugin 获取特定插件的安全事件
func (sm *SecurityManagerImpl) GetSecurityEventsByPlugin(pluginID string, limit int) ([]*SecurityEvent, error) {
	filter := &SecurityEventFilter{
		PluginIDs: []string{pluginID},
		Limit:     limit,
	}
	return sm.GetSecurityEvents(filter)
}

// GetSecurityEventsByLevel 获取特定级别的安全事件
func (sm *SecurityManagerImpl) GetSecurityEventsByLevel(level SecurityEventLevel, limit int) ([]*SecurityEvent, error) {
	filter := &SecurityEventFilter{
		Levels: []SecurityEventLevel{level},
		Limit:  limit,
	}
	return sm.GetSecurityEvents(filter)
}

// GetSecurityEventsByTimeRange 获取时间范围内的安全事件
func (sm *SecurityManagerImpl) GetSecurityEventsByTimeRange(startTime, endTime time.Time, limit int) ([]*SecurityEvent, error) {
	filter := &SecurityEventFilter{
		StartTime: &startTime,
		EndTime:   &endTime,
		Limit:     limit,
	}
	return sm.GetSecurityEvents(filter)
}

// ClearSecurityEvents 清除安全事件（可选择性清除）
func (sm *SecurityManagerImpl) ClearSecurityEvents(before time.Time) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 清除指定时间之前的事件
	var remainingEvents []*SecurityEvent
	clearedCount := 0

	for _, event := range sm.securityEvents {
		if event.Timestamp.After(before) {
			remainingEvents = append(remainingEvents, event)
		} else {
			clearedCount++
		}
	}

	sm.securityEvents = remainingEvents
	sm.logSecurityEvent(SecurityEventTypeAudit, SecurityEventLevelInfo, "", 
		"audit", "clear_filtered", fmt.Sprintf("Cleared %d security events", clearedCount), 
		map[string]interface{}{"cleared_count": clearedCount})

	return nil
}

// ExportSecurityEvents 导出安全事件到文件
func (sm *SecurityManagerImpl) ExportSecurityEvents(filePath string, filter *SecurityEventFilter) error {
	events, err := sm.GetSecurityEvents(filter)
	if err != nil {
		return fmt.Errorf("failed to get security events: %w", err)
	}

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create export file: %w", err)
	}
	defer file.Close()

	// 写入JSON格式
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	exportData := map[string]interface{}{
		"export_time": time.Now(),
		"total_events": len(events),
		"events": events,
	}

	if err := encoder.Encode(exportData); err != nil {
		return fmt.Errorf("failed to encode events: %w", err)
	}

	sm.logSecurityEvent(SecurityEventTypeAudit, SecurityEventLevelInfo, "", 
		"audit", "export", fmt.Sprintf("Exported %d security events to %s", len(events), filePath), 
		map[string]interface{}{"file_path": filePath, "event_count": len(events)})

	return nil
}

// GetSecurityStatistics 获取安全统计信息
func (sm *SecurityManagerImpl) GetSecurityStatistics() *SecurityStatistics {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	stats := &SecurityStatistics{
		TotalEvents: len(sm.securityEvents),
		EventsByType: make(map[SecurityEventType]int),
		EventsByLevel: make(map[SecurityEventLevel]int),
		EventsByPlugin: make(map[string]int),
		GeneratedAt: time.Now(),
	}

	// 统计各类事件数量
	for _, event := range sm.securityEvents {
		stats.EventsByType[event.Type]++
		stats.EventsByLevel[event.Level]++
		if event.PluginID != "" {
			stats.EventsByPlugin[event.PluginID]++
		}
	}

	return stats
}

// EnableAuditLogging 启用审计日志
func (sm *SecurityManagerImpl) EnableAuditLogging() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.securityPolicy == nil {
		sm.securityPolicy = sm.getDefaultSecurityPolicy()
	}

	sm.securityPolicy.AuditEnabled = true
	sm.logSecurityEvent(SecurityEventTypeAudit, SecurityEventLevelInfo, "", 
		"audit", "enable", "Audit logging enabled", nil)

	return nil
}

// DisableAuditLogging 禁用审计日志
func (sm *SecurityManagerImpl) DisableAuditLogging() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.securityPolicy == nil {
		sm.securityPolicy = sm.getDefaultSecurityPolicy()
	}

	sm.securityPolicy.AuditEnabled = false
	sm.logSecurityEvent(SecurityEventTypeAudit, SecurityEventLevelInfo, "", 
		"audit", "disable", "Audit logging disabled", nil)

	return nil
}

// matchesFilter 检查事件是否匹配过滤器
func (sm *SecurityManagerImpl) matchesFilter(event *SecurityEvent, filter *SecurityEventFilter) bool {
	if filter == nil {
		return true
	}

	// 检查插件ID
	if len(filter.PluginIDs) > 0 {
		found := false
		for _, pluginID := range filter.PluginIDs {
			if event.PluginID == pluginID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 检查事件类型
	if len(filter.Types) > 0 {
		found := false
		for _, eventType := range filter.Types {
			if event.Type == eventType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 检查事件级别
	if len(filter.Levels) > 0 {
		found := false
		for _, level := range filter.Levels {
			if event.Level == level {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 检查时间范围
	if filter.StartTime != nil && event.Timestamp.Before(*filter.StartTime) {
		return false
	}
	if filter.EndTime != nil && event.Timestamp.After(*filter.EndTime) {
		return false
	}

	// 检查资源
	if filter.Resource != "" && event.Resource != filter.Resource {
		return false
	}

	// 检查操作
	if filter.Action != "" && event.Action != filter.Action {
		return false
	}

	return true
}

// logSecurityEvent 记录安全事件（内部方法）
func (sm *SecurityManagerImpl) logSecurityEvent(eventType SecurityEventType, level SecurityEventLevel, 
	pluginID, resource, action, message string, details map[string]interface{}) {
	
	// 检查是否启用审计
	if sm.securityPolicy != nil && !sm.securityPolicy.AuditEnabled {
		return
	}

	// 创建安全事件
	event := &SecurityEvent{
		ID:        fmt.Sprintf("%d-%s", time.Now().UnixNano(), pluginID),
		Type:      eventType,
		Level:     level,
		PluginID:  pluginID,
		Resource:  resource,
		Action:    action,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
	}

	// 添加到事件列表
	sm.securityEvents = append(sm.securityEvents, event)

	// 限制事件数量，避免内存溢出
	maxEvents := 10000
	if len(sm.securityEvents) > maxEvents {
		// 保留最新的事件
		sm.securityEvents = sm.securityEvents[len(sm.securityEvents)-maxEvents:]
	}

	// 根据事件级别进行不同处理
	switch level {
	case SecurityEventLevelCritical:
		// 关键事件，可能需要立即处理
		sm.handleCriticalSecurityEvent(event)
	case SecurityEventLevelError:
		// 错误事件，记录到错误日志
		sm.logToSystemLog("ERROR", event)
	case SecurityEventLevelWarning:
		// 警告事件，记录到警告日志
		sm.logToSystemLog("WARNING", event)
	case SecurityEventLevelInfo:
		// 信息事件，记录到信息日志
		sm.logToSystemLog("INFO", event)
	}
}

// handleCriticalSecurityEvent 处理关键安全事件
func (sm *SecurityManagerImpl) handleCriticalSecurityEvent(event *SecurityEvent) {
	// 关键事件处理逻辑
	// 例如：发送警报、暂停插件、记录到特殊日志等
	sm.logToSystemLog("CRITICAL", event)
	
	// 如果是插件相关的关键事件，可能需要暂停插件
	if event.PluginID != "" {
		// 这里可以添加暂停插件的逻辑
		// 例如：通知插件管理器暂停该插件
	}
}

// logToSystemLog 记录到系统日志
func (sm *SecurityManagerImpl) logToSystemLog(level string, event *SecurityEvent) {
	// 构建日志消息
	logMessage := fmt.Sprintf("[SECURITY-%s] %s: %s", level, event.Type, event.Message)
	if event.PluginID != "" {
		logMessage += fmt.Sprintf(" (Plugin: %s)", event.PluginID)
	}
	if event.Resource != "" {
		logMessage += fmt.Sprintf(" (Resource: %s)", event.Resource)
	}
	if event.Action != "" {
		logMessage += fmt.Sprintf(" (Action: %s)", event.Action)
	}

	// 这里可以使用标准库的log包或其他日志库
	// 暂时使用fmt.Printf作为示例
	fmt.Printf("[%s] %s\n", event.Timestamp.Format("2006-01-02 15:04:05"), logMessage)

	// 如果有详细信息，也记录下来
	if event.Details != nil && len(event.Details) > 0 {
		detailsJSON, _ := json.Marshal(event.Details)
		fmt.Printf("[%s] Details: %s\n", event.Timestamp.Format("2006-01-02 15:04:05"), string(detailsJSON))
	}
}

// RecoverFromError 从错误中恢复
func (sm *SecurityManagerImpl) RecoverFromError(err error, context string) error {
	if err == nil {
		return nil
	}

	// 记录错误事件
	sm.logSecurityEvent(
		SecurityEventTypeError,
		SecurityEventLevelError,
		"",
		"",
		"error_recovery",
		fmt.Sprintf("Error recovery in %s: %v", context, err),
		map[string]interface{}{
			"error":   err.Error(),
			"context": context,
		},
	)

	// 根据错误类型进行不同的恢复策略
	switch {
	case errors.Is(err, ErrSandboxCreationFailed):
		return sm.recoverFromSandboxError(err, context)
	case errors.Is(err, ErrResourceLimitExceeded):
		return sm.recoverFromResourceError(err, context)
	case errors.Is(err, ErrInvalidPlugin):
		return sm.recoverFromPluginError(err, context)
	case errors.Is(err, ErrPermissionDenied):
		return sm.recoverFromPermissionError(err, context)
	default:
		return sm.recoverFromGenericError(err, context)
	}
}

// recoverFromSandboxError 从沙箱错误中恢复
func (sm *SecurityManagerImpl) recoverFromSandboxError(err error, context string) error {
	// 尝试清理损坏的沙箱
	for pluginID, sandbox := range sm.sandboxes {
		if sandbox != nil && !sandbox.IsActive() {
			// 销毁损坏的沙箱
			delete(sm.sandboxes, pluginID)
			sm.logSecurityEvent(
				SecurityEventTypeSandboxViolation,
				SecurityEventLevelWarning,
				pluginID,
				"sandbox",
				"cleanup",
				"Cleaned up corrupted sandbox",
				nil,
			)
		}
	}
	return nil
}

// recoverFromResourceError 从资源错误中恢复
func (sm *SecurityManagerImpl) recoverFromResourceError(err error, context string) error {
	// 尝试释放资源
	for pluginID, usage := range sm.resourceUsage {
		if usage.MemoryUsed > 100*1024*1024 { // 超过100MB
			// 重置资源使用统计
			usage.MemoryUsed = 0
			usage.CPUUsage = 0
			usage.NetworkIORate = 0
			usage.LastUpdated = time.Now()
			
			sm.logSecurityEvent(
				SecurityEventTypeResourceExceeded,
				SecurityEventLevelWarning,
				pluginID,
				"resource",
				"reset",
				"Reset resource usage statistics",
				nil,
			)
		}
	}
	return nil
}

// recoverFromPluginError 从插件错误中恢复
func (sm *SecurityManagerImpl) recoverFromPluginError(err error, context string) error {
	// 可以尝试重新验证插件或清理插件状态
	sm.logSecurityEvent(
		SecurityEventTypePluginLoad,
		SecurityEventLevelWarning,
		"",
		"plugin",
		"recovery",
		"Attempting plugin error recovery",
		map[string]interface{}{
			"error":   err.Error(),
			"context": context,
		},
	)
	return nil
}

// recoverFromPermissionError 从权限错误中恢复
func (sm *SecurityManagerImpl) recoverFromPermissionError(err error, context string) error {
	// 记录权限错误，但不自动恢复（安全考虑）
	sm.logSecurityEvent(
		SecurityEventTypePermissionDenied,
		SecurityEventLevelError,
		"",
		"permission",
		"denied",
		"Permission denied - manual intervention required",
		map[string]interface{}{
			"error":   err.Error(),
			"context": context,
		},
	)
	return err // 权限错误不自动恢复
}

// recoverFromGenericError 从通用错误中恢复
func (sm *SecurityManagerImpl) recoverFromGenericError(err error, context string) error {
	// 通用错误恢复策略
	sm.logSecurityEvent(
		SecurityEventTypeError,
		SecurityEventLevelError,
		"",
		"",
		"generic_recovery",
		fmt.Sprintf("Generic error recovery attempted: %v", err),
		map[string]interface{}{
			"error":   err.Error(),
			"context": context,
		},
	)
	return err
}

// MonitorResources 监控插件资源使用情况
func (sm *SecurityManagerImpl) MonitorResources(pluginID string) (*ResourceUsage, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if pluginID == "" {
		return nil, fmt.Errorf("plugin ID cannot be empty")
	}

	// 检查是否存在该插件的资源使用记录
	if usage, exists := sm.resourceUsage[pluginID]; exists {
		return usage, nil
	}

	// 如果不存在，创建新的资源使用记录
	usage := &ResourceUsage{
		MemoryUsed:         0,
		CPUUsage:          0,
		CPUUsed:           0,
		DiskIORate:        0,
		NetworkIORate:     0,
		GoroutineCount:    0,
		FileHandleCount:   0,
		FileDescriptorsUsed: 0,
		NetworkConnections: 0,
		LastUpdated:       time.Now(),
	}

	sm.resourceUsage[pluginID] = usage
	return usage, nil
}

// LogSecurityEvent 记录安全事件
func (sm *SecurityManagerImpl) LogSecurityEvent(event *SecurityEvent) error {
	if event == nil {
		return fmt.Errorf("security event cannot be nil")
	}

	// 设置事件ID和时间戳（如果未设置）
	if event.ID == "" {
		event.ID = fmt.Sprintf("%d-%s", time.Now().UnixNano(), event.PluginID)
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 调用内部日志方法
	sm.logSecurityEvent(event.Type, event.Level, event.PluginID, 
		event.Resource, event.Action, event.Message, event.Details)

	return nil
}

// LoadSecurityPolicy 加载安全策略
func (sm *SecurityManagerImpl) LoadSecurityPolicy(policyPath string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 读取策略文件
	data, err := os.ReadFile(policyPath)
	if err != nil {
		return fmt.Errorf("failed to read security policy file: %w", err)
	}

	// 解析策略
	var policy SecurityPolicy
	if err := json.Unmarshal(data, &policy); err != nil {
		return fmt.Errorf("failed to parse security policy: %w", err)
	}

	// 验证策略
	if err := sm.validateSecurityPolicy(&policy); err != nil {
		return fmt.Errorf("invalid security policy: %w", err)
	}

	// 应用策略
	sm.securityPolicy = &policy
	sm.securityPolicy.UpdatedAt = time.Now()

	// 记录策略加载事件
	sm.logSecurityEvent(SecurityEventTypePolicy, SecurityEventLevelInfo, "", 
		"policy", "load", "Security policy loaded successfully", 
		map[string]interface{}{"policy_path": policyPath, "version": policy.Version})

	return nil
}

// HandlePanic 处理panic恢复
func (sm *SecurityManagerImpl) HandlePanic(pluginID string) {
	if r := recover(); r != nil {
		// 记录panic事件
		sm.logSecurityEvent(
			SecurityEventTypeError,
			SecurityEventLevelCritical,
			pluginID,
			"runtime",
			"panic",
			fmt.Sprintf("Plugin panic recovered: %v", r),
			map[string]interface{}{
				"panic_value": fmt.Sprintf("%v", r),
				"plugin_id":   pluginID,
			},
		)

		// 清理插件相关资源
		if pluginID != "" {
			// 销毁沙箱
			if err := sm.DestroySandbox(pluginID); err != nil {
				sm.logSecurityEvent(
					SecurityEventTypeError,
					SecurityEventLevelError,
					pluginID,
					"sandbox",
					"cleanup_failed",
					"Failed to cleanup sandbox after panic",
					map[string]interface{}{"error": err.Error()},
				)
			}

			// 清理资源使用统计
			delete(sm.resourceUsage, pluginID)
		}
	}
}

// ValidateSecurityState 验证安全管理器状态
func (sm *SecurityManagerImpl) ValidateSecurityState() error {
	if sm == nil {
		return ErrSecurityManagerNotInitialized
	}

	// 检查基本组件
	if sm.sandboxes == nil {
		return errors.New("sandboxes not initialized")
	}

	if sm.permissions == nil {
		return errors.New("permissions not initialized")
	}

	if sm.aclRules == nil {
		return errors.New("ACL rules not initialized")
	}

	if sm.resourceLimits == nil {
		return errors.New("resource limits not initialized")
	}

	if sm.resourceUsage == nil {
		return errors.New("resource usage not initialized")
	}

	if sm.securityEvents == nil {
		return errors.New("security events not initialized")
	}

	// 验证安全策略
	if sm.securityPolicy != nil {
		if err := sm.validateSecurityPolicy(sm.securityPolicy); err != nil {
			return fmt.Errorf("invalid security policy: %w", err)
		}
	}

	return nil
}

// SetResourceLimits 设置指定插件的资源限制
func (sm *SecurityManagerImpl) SetResourceLimits(pluginID string, limits *ResourceLimits) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if pluginID == "" {
		return fmt.Errorf("plugin ID cannot be empty")
	}

	// 验证资源限制的有效性
	if limits.MaxMemory < 0 {
		return fmt.Errorf("max memory cannot be negative")
	}
	if limits.MaxCPU < 0 || limits.MaxCPU > 1 {
		return fmt.Errorf("max CPU must be between 0 and 1")
	}
	if limits.MaxCPUPercent < 0 || limits.MaxCPUPercent > 100 {
		return fmt.Errorf("max CPU percent must be between 0 and 100")
	}

	// 设置资源限制
	sm.resourceLimits[pluginID] = limits

	// 记录事件
	sm.logSecurityEvent(SecurityEventTypeResourceLimit, SecurityEventLevelInfo, pluginID,
		"resource_limits", "set", "Resource limits updated for plugin",
		map[string]interface{}{
			"max_memory":              limits.MaxMemory,
			"max_cpu":                 limits.MaxCPU,
			"max_cpu_percent":         limits.MaxCPUPercent,
			"max_disk_io":             limits.MaxDiskIO,
			"max_network_io":          limits.MaxNetworkIO,
			"timeout":                 limits.Timeout.String(),
			"max_goroutines":          limits.MaxGoroutines,
			"max_file_handles":        limits.MaxFileHandles,
			"max_file_descriptors":    limits.MaxFileDescriptors,
			"max_network_connections": limits.MaxNetworkConnections,
		})

	return nil
}

// GetResourceLimits 获取指定插件的资源限制
func (sm *SecurityManagerImpl) GetResourceLimits(pluginID string) (*ResourceLimits, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	limit, exists := sm.resourceLimits[pluginID]
	if !exists {
		return nil, fmt.Errorf("resource limits not found for plugin: %s", pluginID)
	}

	// 返回资源限制的副本
	return &ResourceLimits{
		MaxMemory:             limit.MaxMemory,
		MaxCPU:                limit.MaxCPU,
		MaxCPUPercent:         limit.MaxCPUPercent,
		MaxDiskIO:             limit.MaxDiskIO,
		MaxNetworkIO:          limit.MaxNetworkIO,
		Timeout:               limit.Timeout,
		MaxGoroutines:         limit.MaxGoroutines,
		MaxFileHandles:        limit.MaxFileHandles,
		MaxFileDescriptors:    limit.MaxFileDescriptors,
		MaxNetworkConnections: limit.MaxNetworkConnections,
	}, nil
}

// SafeExecute 安全执行函数，带有错误恢复
func (sm *SecurityManagerImpl) SafeExecute(pluginID string, fn func() error) (err error) {
	// 设置panic恢复
	defer func() {
		if r := recover(); r != nil {
			sm.HandlePanic(pluginID)
			err = fmt.Errorf("panic recovered: %v", r)
		}
	}()

	// 验证安全状态
	if validationErr := sm.ValidateSecurityState(); validationErr != nil {
		return fmt.Errorf("security state validation failed: %w", validationErr)
	}

	// 执行函数
	err = fn()
	if err != nil {
		// 尝试从错误中恢复
		recoveryErr := sm.RecoverFromError(err, "safe_execute")
		if recoveryErr != nil {
			return fmt.Errorf("execution failed and recovery failed: %w (original: %v)", recoveryErr, err)
		}
	}

	return err
}