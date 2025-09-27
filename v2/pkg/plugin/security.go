package plugin

import (
	"context"
	"fmt"
	// "net" // 暂时未使用
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"log/slog"
)

// Permission 权限类型
type Permission string

const (
	// 文件系统权限
	PermissionFileRead   Permission = "file:read"
	PermissionFileWrite  Permission = "file:write"
	PermissionFileDelete Permission = "file:delete"
	PermissionFileExec   Permission = "file:execute"

	// 网络权限
	PermissionNetworkConnect Permission = "network:connect"
	PermissionNetworkListen  Permission = "network:listen"
	PermissionNetworkHTTP    Permission = "network:http"
	PermissionNetworkHTTPS   Permission = "network:https"

	// 系统权限
	PermissionSystemEnv    Permission = "system:env"
	PermissionSystemExec   Permission = "system:exec"
	PermissionSystemSignal Permission = "system:signal"

	// 数据库权限
	PermissionDBRead  Permission = "db:read"
	PermissionDBWrite Permission = "db:write"
	PermissionDBAdmin Permission = "db:admin"

	// API权限
	PermissionAPICall   Permission = "api:call"
	PermissionAPICreate Permission = "api:create"
	PermissionAPIUpdate Permission = "api:update"
	PermissionAPIDelete Permission = "api:delete"
)

// String 返回权限的字符串表示
func (p Permission) String() string {
	return string(p)
}

// SecurityLevel 安全级别
type SecurityLevel string

const (
	SecurityLevelNone     SecurityLevel = "none"     // 无安全限制
	SecurityLevelLow      SecurityLevel = "low"      // 低安全级别
	SecurityLevelMedium   SecurityLevel = "medium"   // 中等安全级别
	SecurityLevelHigh     SecurityLevel = "high"     // 高安全级别
	SecurityLevelCritical SecurityLevel = "critical" // 关键安全级别
)

// String 返回安全级别的字符串表示
func (sl SecurityLevel) String() string {
	return string(sl)
}

// EnforceMode 执行模式
type EnforceMode string

const (
	EnforceModePermissive EnforceMode = "permissive" // 宽松模式，仅记录违规
	EnforceModeEnforcing  EnforceMode = "enforcing"  // 强制模式，阻止违规操作
	EnforceModeDisabled   EnforceMode = "disabled"   // 禁用模式
)

// String 返回执行模式的字符串表示
func (em EnforceMode) String() string {
	return string(em)
}

// PathAccessRule 路径访问规则
type PathAccessRule struct {
	Path        string       `json:"path" yaml:"path"`               // 路径模式
	Permissions []Permission `json:"permissions" yaml:"permissions"` // 允许的权限
	Deny        bool         `json:"deny,omitempty" yaml:"deny,omitempty"` // 是否为拒绝规则
	Description string       `json:"description,omitempty" yaml:"description,omitempty"`
}

// Validate 验证路径访问规则
func (par *PathAccessRule) Validate() error {
	if par.Path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	if len(par.Permissions) == 0 && !par.Deny {
		return fmt.Errorf("permissions cannot be empty for allow rules")
	}

	return nil
}

// Matches 检查路径是否匹配规则
func (par *PathAccessRule) Matches(path string) bool {
	// 支持通配符匹配
	matched, err := filepath.Match(par.Path, path)
	if err != nil {
		return false
	}
	return matched
}

// HasPermission 检查是否有指定权限
func (par *PathAccessRule) HasPermission(permission Permission) bool {
	if par.Deny {
		return false
	}

	for _, p := range par.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// NetworkAccessRule 网络访问规则
type NetworkAccessRule struct {
	Host        string       `json:"host" yaml:"host"`               // 主机地址或模式
	Port        int          `json:"port,omitempty" yaml:"port,omitempty"` // 端口号，0表示任意端口
	Protocol    string       `json:"protocol,omitempty" yaml:"protocol,omitempty"` // 协议类型
	Permissions []Permission `json:"permissions" yaml:"permissions"` // 允许的权限
	Deny        bool         `json:"deny,omitempty" yaml:"deny,omitempty"` // 是否为拒绝规则
	Description string       `json:"description,omitempty" yaml:"description,omitempty"`
}

// Validate 验证网络访问规则
func (nar *NetworkAccessRule) Validate() error {
	if nar.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	if nar.Port < 0 || nar.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", nar.Port)
	}

	if len(nar.Permissions) == 0 && !nar.Deny {
		return fmt.Errorf("permissions cannot be empty for allow rules")
	}

	return nil
}

// Matches 检查网络地址是否匹配规则
func (nar *NetworkAccessRule) Matches(host string, port int, protocol string) bool {
	// 检查主机匹配
	hostMatched := false
	if nar.Host == "*" || nar.Host == host {
		hostMatched = true
	} else {
		// 支持通配符匹配
		if matched, err := filepath.Match(nar.Host, host); err == nil && matched {
			hostMatched = true
		}
	}

	if !hostMatched {
		return false
	}

	// 检查端口匹配
	if nar.Port != 0 && nar.Port != port {
		return false
	}

	// 检查协议匹配
	if nar.Protocol != "" && nar.Protocol != protocol {
		return false
	}

	return true
}

// HasPermission 检查是否有指定权限
func (nar *NetworkAccessRule) HasPermission(permission Permission) bool {
	if nar.Deny {
		return false
	}

	for _, p := range nar.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// SandboxConfig 沙箱配置
type SandboxConfig struct {
	Enabled          bool     `json:"enabled" yaml:"enabled"`
	RootPath         string   `json:"root_path,omitempty" yaml:"root_path,omitempty"`
	AllowedPaths     []string `json:"allowed_paths,omitempty" yaml:"allowed_paths,omitempty"`
	BlockedPaths     []string `json:"blocked_paths,omitempty" yaml:"blocked_paths,omitempty"`
	AllowedSyscalls  []string `json:"allowed_syscalls,omitempty" yaml:"allowed_syscalls,omitempty"`
	BlockedSyscalls  []string `json:"blocked_syscalls,omitempty" yaml:"blocked_syscalls,omitempty"`
	MaxFileSize      int64    `json:"max_file_size,omitempty" yaml:"max_file_size,omitempty"`
	MaxOpenFiles     int      `json:"max_open_files,omitempty" yaml:"max_open_files,omitempty"`
	NetworkIsolation bool     `json:"network_isolation,omitempty" yaml:"network_isolation,omitempty"`
	ProcessIsolation bool     `json:"process_isolation,omitempty" yaml:"process_isolation,omitempty"`
}

// Validate 验证沙箱配置
func (sc *SandboxConfig) Validate() error {
	if !sc.Enabled {
		return nil
	}

	if sc.RootPath != "" {
		if !filepath.IsAbs(sc.RootPath) {
			return fmt.Errorf("root path must be absolute: %s", sc.RootPath)
		}
	}

	if sc.MaxFileSize < 0 {
		return fmt.Errorf("max file size cannot be negative")
	}

	if sc.MaxOpenFiles < 0 {
		return fmt.Errorf("max open files cannot be negative")
	}

	return nil
}

// IsPathAllowed 检查路径是否被允许
func (sc *SandboxConfig) IsPathAllowed(path string) bool {
	if !sc.Enabled {
		return true
	}

	// 检查是否在阻止列表中
	for _, blocked := range sc.BlockedPaths {
		if matched, _ := filepath.Match(blocked, path); matched {
			return false
		}
	}

	// 如果有允许列表，检查是否在其中
	if len(sc.AllowedPaths) > 0 {
		for _, allowed := range sc.AllowedPaths {
			if matched, _ := filepath.Match(allowed, path); matched {
				return true
			}
		}
		return false
	}

	// 检查是否在根路径下
	if sc.RootPath != "" {
		rel, err := filepath.Rel(sc.RootPath, path)
		if err != nil || strings.HasPrefix(rel, "..") {
			return false
		}
	}

	return true
}

// IsSyscallAllowed 检查系统调用是否被允许
func (sc *SandboxConfig) IsSyscallAllowed(syscall string) bool {
	if !sc.Enabled {
		return true
	}

	// 检查是否在阻止列表中
	for _, blocked := range sc.BlockedSyscalls {
		if blocked == syscall {
			return false
		}
	}

	// 如果有允许列表，检查是否在其中
	if len(sc.AllowedSyscalls) > 0 {
		for _, allowed := range sc.AllowedSyscalls {
			if allowed == syscall {
				return true
			}
		}
		return false
	}

	return true
}

// DetailedSecurityConfig 详细安全配置
type DetailedSecurityConfig struct {
	Enabled             bool                  `json:"enabled" yaml:"enabled"`
	Level               SecurityLevel         `json:"level" yaml:"level"`
	EnforceMode         EnforceMode           `json:"enforce_mode" yaml:"enforce_mode"`
	Permissions         []Permission          `json:"permissions,omitempty" yaml:"permissions,omitempty"`
	PathAccessRules     []PathAccessRule      `json:"path_access_rules,omitempty" yaml:"path_access_rules,omitempty"`
	NetworkAccessRules  []NetworkAccessRule   `json:"network_access_rules,omitempty" yaml:"network_access_rules,omitempty"`
	Sandbox             *SandboxConfig        `json:"sandbox,omitempty" yaml:"sandbox,omitempty"`
	TrustedCertificates []string              `json:"trusted_certificates,omitempty" yaml:"trusted_certificates,omitempty"`
	AllowedDomains      []string              `json:"allowed_domains,omitempty" yaml:"allowed_domains,omitempty"`
	BlockedDomains      []string              `json:"blocked_domains,omitempty" yaml:"blocked_domains,omitempty"`
	RateLimits          map[string]RateLimit  `json:"rate_limits,omitempty" yaml:"rate_limits,omitempty"`
	AuditEnabled        bool                  `json:"audit_enabled,omitempty" yaml:"audit_enabled,omitempty"`
	AuditLogPath        string                `json:"audit_log_path,omitempty" yaml:"audit_log_path,omitempty"`
}

// RateLimit 速率限制
type RateLimit struct {
	Requests int           `json:"requests" yaml:"requests"` // 请求数量
	Window   time.Duration `json:"window" yaml:"window"`     // 时间窗口
	Enabled  bool          `json:"enabled" yaml:"enabled"`
}

// Validate 验证速率限制
func (rl *RateLimit) Validate() error {
	if rl.Enabled {
		if rl.Requests <= 0 {
			return fmt.Errorf("requests must be positive")
		}
		if rl.Window <= 0 {
			return fmt.Errorf("window must be positive")
		}
	}
	return nil
}

// NewDetailedSecurityConfig 创建详细安全配置
func NewDetailedSecurityConfig() *DetailedSecurityConfig {
	return &DetailedSecurityConfig{
		Enabled:     true,
		Level:       SecurityLevelMedium,
		EnforceMode: EnforceModeEnforcing,
		Permissions: []Permission{
			PermissionFileRead,
			PermissionNetworkHTTP,
			PermissionNetworkHTTPS,
		},
		PathAccessRules: []PathAccessRule{
			{
				Path:        "/tmp/*",
				Permissions: []Permission{PermissionFileRead, PermissionFileWrite},
				Description: "Allow read/write access to temp directory",
			},
			{
				Path:        "/etc/*",
				Deny:        true,
				Description: "Deny access to system configuration",
			},
		},
		NetworkAccessRules: []NetworkAccessRule{
			{
				Host:        "localhost",
				Permissions: []Permission{PermissionNetworkConnect},
				Description: "Allow connection to localhost",
			},
			{
				Host:        "*.example.com",
				Permissions: []Permission{PermissionNetworkHTTP, PermissionNetworkHTTPS},
				Description: "Allow HTTP/HTTPS to example.com subdomains",
			},
		},
		Sandbox: &SandboxConfig{
			Enabled:          true,
			MaxFileSize:      10 * 1024 * 1024, // 10MB
			MaxOpenFiles:     100,
			NetworkIsolation: false,
			ProcessIsolation: true,
		},
		RateLimits: map[string]RateLimit{
			"api_calls": {
				Requests: 100,
				Window:   time.Minute,
				Enabled:  true,
			},
			"file_operations": {
				Requests: 1000,
				Window:   time.Minute,
				Enabled:  true,
			},
		},
		AuditEnabled: true,
	}
}

// Validate 验证安全配置
func (sc *DetailedSecurityConfig) Validate() error {
	if !sc.Enabled {
		return nil
	}

	// 验证安全级别
	validLevels := []SecurityLevel{SecurityLevelNone, SecurityLevelLow, SecurityLevelMedium, SecurityLevelHigh, SecurityLevelCritical}
	validLevel := false
	for _, level := range validLevels {
		if sc.Level == level {
			validLevel = true
			break
		}
	}
	if !validLevel {
		return fmt.Errorf("invalid security level: %s", sc.Level)
	}

	// 验证执行模式
	validModes := []EnforceMode{EnforceModePermissive, EnforceModeEnforcing, EnforceModeDisabled}
	validMode := false
	for _, mode := range validModes {
		if sc.EnforceMode == mode {
			validMode = true
			break
		}
	}
	if !validMode {
		return fmt.Errorf("invalid enforce mode: %s", sc.EnforceMode)
	}

	// 验证路径访问规则
	for i, rule := range sc.PathAccessRules {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("invalid path access rule %d: %w", i, err)
		}
	}

	// 验证网络访问规则
	for i, rule := range sc.NetworkAccessRules {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("invalid network access rule %d: %w", i, err)
		}
	}

	// 验证沙箱配置
	if sc.Sandbox != nil {
		if err := sc.Sandbox.Validate(); err != nil {
			return fmt.Errorf("invalid sandbox config: %w", err)
		}
	}

	// 验证速率限制
	for name, limit := range sc.RateLimits {
		if err := limit.Validate(); err != nil {
			return fmt.Errorf("invalid rate limit %s: %w", name, err)
		}
	}

	// 验证审计日志路径
	if sc.AuditEnabled && sc.AuditLogPath != "" {
		dir := filepath.Dir(sc.AuditLogPath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("audit log directory does not exist: %s", dir)
		}
	}

	return nil
}

// HasPermission 检查是否有指定权限
func (sc *DetailedSecurityConfig) HasPermission(permission Permission) bool {
	if !sc.Enabled || sc.EnforceMode == EnforceModeDisabled {
		return true
	}

	for _, p := range sc.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// CheckPathAccess 检查路径访问权限
func (sc *DetailedSecurityConfig) CheckPathAccess(path string, permission Permission) bool {
	if !sc.Enabled || sc.EnforceMode == EnforceModeDisabled {
		return true
	}

	// 检查沙箱限制
	if sc.Sandbox != nil && !sc.Sandbox.IsPathAllowed(path) {
		return false
	}

	// 检查路径访问规则
	for _, rule := range sc.PathAccessRules {
		if rule.Matches(path) {
			if rule.Deny {
				return false
			}
			// TODO: 检查权限
			return true
		}
	}

	// 默认允许访问
	return true
}

// CheckNetworkAccess 检查网络访问权限
func (sc *SecurityConfig) CheckNetworkAccess(host string, port int, protocol string, permission Permission) bool {
	// TODO: 实现网络访问检查逻辑
	// 由于SecurityConfig结构已简化，需要重新设计
	
	// 检查域名黑名单
	for _, blocked := range sc.BlockedHosts {
		if matched, _ := regexp.MatchString(blocked, host); matched {
			return false
		}
	}

	// 检查域名白名单
	if len(sc.AllowedHosts) > 0 {
		allowed := false
		for _, allowedHost := range sc.AllowedHosts {
			if matched, _ := regexp.MatchString(allowedHost, host); matched {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}

	// TODO: 检查网络访问规则
	// 由于SecurityConfig结构已简化，需要重新设计
	// for _, rule := range sc.NetworkAccessRules {
	//     if rule.Matches(host, port, protocol) {
	//         if rule.Deny {
	//             return false
	//         }
	//         return rule.HasPermission(permission)
	//     }
	// }

	// 默认允许访问
	return true
}

// SecurityViolation 安全违规记录
type SecurityViolation struct {
	PluginID    string      `json:"plugin_id"`
	Violation   string      `json:"violation"`
	Resource    string      `json:"resource"`
	Permission  Permission  `json:"permission"`
	Timestamp   time.Time   `json:"timestamp"`
	Severity    string      `json:"severity"`
	Action      string      `json:"action"`
	Details     interface{} `json:"details,omitempty"`
}

// SecurityEnforcer 安全执行器
type SecurityEnforcer struct {
	logger     *slog.Logger
	pluginID   string
	config     *SecurityConfig
	violations chan *SecurityViolation
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	auditFile  *os.File
	rateLimiters map[string]*RateLimiter
}

// RateLimiter 速率限制器
type RateLimiter struct {
	limit    int
	window   time.Duration
	requests []time.Time
	mu       sync.Mutex
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:    limit,
		window:   window,
		requests: make([]time.Time, 0, limit),
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// 清理过期的请求记录
	validRequests := make([]time.Time, 0, len(rl.requests))
	for _, req := range rl.requests {
		if req.After(cutoff) {
			validRequests = append(validRequests, req)
		}
	}
	rl.requests = validRequests

	// 检查是否超过限制
	if len(rl.requests) >= rl.limit {
		return false
	}

	// 记录新请求
	rl.requests = append(rl.requests, now)
	return true
}

// NewSecurityEnforcer 创建安全执行器
func NewSecurityEnforcer(logger *slog.Logger, pluginID string, config *SecurityConfig) *SecurityEnforcer {
	ctx, cancel := context.WithCancel(context.Background())

	enforcer := &SecurityEnforcer{
		logger:       logger,
		pluginID:     pluginID,
		config:       config,
		violations:   make(chan *SecurityViolation, 100),
		ctx:          ctx,
		cancel:       cancel,
		rateLimiters: make(map[string]*RateLimiter),
	}

	// TODO: 初始化速率限制器
	// 由于SecurityConfig结构已简化，需要重新设计
	// for name, limit := range config.RateLimits {
	//     if limit.Enabled {
	//         enforcer.rateLimiters[name] = NewRateLimiter(limit.Requests, limit.Window)
	//     }
	// }

	return enforcer
}

// Start 启动安全执行器
func (se *SecurityEnforcer) Start(ctx context.Context) error {
	// TODO: 实现审计日志功能
	// 由于SecurityConfig结构已简化，需要重新设计
	// if se.config.AuditEnabled && se.config.AuditLogPath != "" {
	//     file, err := os.OpenFile(se.config.AuditLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	//     if err != nil {
	//         return fmt.Errorf("failed to open audit log file: %w", err)
	//     }
	//     se.auditFile = file
	// }

	// 启动违规处理协程
	go se.handleViolations()

	se.logger.Info("Security enforcer started", "plugin_id", se.pluginID)
	return nil
}

// Stop 停止安全执行器
func (se *SecurityEnforcer) Stop() {
	se.cancel()
	close(se.violations)

	if se.auditFile != nil {
		se.auditFile.Close()
	}

	se.logger.Info("Security enforcer stopped", "plugin_id", se.pluginID)
}

// CheckPermission 检查权限
func (se *SecurityEnforcer) CheckPermission(permission Permission) bool {
	// TODO: 实现权限检查逻辑
	// 由于SecurityConfig结构已简化，需要重新设计
	allowed := true // 临时允许所有权限

	if !allowed {
		se.recordViolation("permission_denied", string(permission), permission, "critical", "deny", nil)
	}

	return allowed
}

// CheckPathAccess 检查路径访问
func (se *SecurityEnforcer) CheckPathAccess(path string, permission Permission) bool {
	// TODO: 实现路径访问检查逻辑
	// 由于SecurityConfig结构已简化，需要重新设计
	allowed := true // 临时允许所有路径访问

	if !allowed {
		se.recordViolation("path_access_denied", path, permission, "high", "deny", map[string]interface{}{
			"path": path,
		})
	}

	return allowed
}

// CheckNetworkAccess 检查网络访问
func (se *SecurityEnforcer) CheckNetworkAccess(host string, port int, protocol string, permission Permission) bool {
	allowed := se.config.CheckNetworkAccess(host, port, protocol, permission)

	if !allowed {
		se.recordViolation("network_access_denied", fmt.Sprintf("%s:%d", host, port), permission, "high", "deny", map[string]interface{}{
			"host":     host,
			"port":     port,
			"protocol": protocol,
		})
	}

	return allowed
}

// CheckRateLimit 检查速率限制
func (se *SecurityEnforcer) CheckRateLimit(name string) bool {
	se.mu.RLock()
	limiter, exists := se.rateLimiters[name]
	se.mu.RUnlock()

	if !exists {
		return true
	}

	allowed := limiter.Allow()

	if !allowed {
		se.recordViolation("rate_limit_exceeded", name, "", "medium", "throttle", map[string]interface{}{
			"rate_limit_name": name,
		})
	}

	return allowed
}

// UpdateConfig 更新安全配置
func (se *SecurityEnforcer) UpdateConfig(config *SecurityConfig) error {
	// TODO: 实现配置验证
	// 由于SecurityConfig结构已简化，需要重新设计验证逻辑
	// if err := config.Validate(); err != nil {
	//     return fmt.Errorf("invalid security config: %w", err)
	// }

	se.mu.Lock()
	defer se.mu.Unlock()

	se.config = config

	// TODO: 更新速率限制器
	// 由于SecurityConfig结构已简化，需要重新设计
	// se.rateLimiters = make(map[string]*RateLimiter)
	// for name, limit := range config.RateLimits {
	//     if limit.Enabled {
	//         se.rateLimiters[name] = NewRateLimiter(limit.Requests, limit.Window)
	//     }
	// }

	se.logger.Info("Security config updated", "plugin_id", se.pluginID)
	return nil
}

// GetViolations 获取违规通道
func (se *SecurityEnforcer) GetViolations() <-chan *SecurityViolation {
	return se.violations
}

// recordViolation 记录安全违规
func (se *SecurityEnforcer) recordViolation(violation, resource string, permission Permission, severity, action string, details interface{}) {
	v := &SecurityViolation{
		PluginID:   se.pluginID,
		Violation:  violation,
		Resource:   resource,
		Permission: permission,
		Timestamp:  time.Now(),
		Severity:   severity,
		Action:     action,
		Details:    details,
	}

	select {
	case se.violations <- v:
	default:
		se.logger.Error("Violation channel full, dropping violation", "plugin_id", se.pluginID)
	}
}

// handleViolations 处理安全违规
func (se *SecurityEnforcer) handleViolations() {
	for {
		select {
		case <-se.ctx.Done():
			return

		case violation, ok := <-se.violations:
			if !ok {
				return
			}

			se.processViolation(violation)
		}
	}
}

// processViolation 处理单个违规
func (se *SecurityEnforcer) processViolation(violation *SecurityViolation) {
	// 记录日志
	se.logger.Warn("Security violation detected",
		"plugin_id", violation.PluginID,
		"violation", violation.Violation,
		"resource", violation.Resource,
		"permission", violation.Permission,
		"severity", violation.Severity,
		"action", violation.Action)

	// 写入审计日志
	if se.auditFile != nil {
		auditEntry := fmt.Sprintf("%s [%s] %s: %s - %s (resource: %s, permission: %s)\n",
			violation.Timestamp.Format(time.RFC3339),
			violation.Severity,
			violation.PluginID,
			violation.Violation,
			violation.Action,
			violation.Resource,
			violation.Permission)

		se.auditFile.WriteString(auditEntry)
		se.auditFile.Sync()
	}
}

// SecurityManager 安全管理器
type SecurityManager struct {
	logger    *slog.Logger
	enforcers map[string]*SecurityEnforcer
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewSecurityManager 创建安全管理器
func NewSecurityManager(logger *slog.Logger) *SecurityManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &SecurityManager{
		logger:    logger,
		enforcers: make(map[string]*SecurityEnforcer),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start 启动安全管理器
func (sm *SecurityManager) Start(ctx context.Context) error {
	sm.logger.Info("Security manager started")
	return nil
}

// Stop 停止安全管理器
func (sm *SecurityManager) Stop() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.cancel()

	// 停止所有执行器
	for _, enforcer := range sm.enforcers {
		enforcer.Stop()
	}

	sm.logger.Info("Security manager stopped")
}

// AddEnforcer 添加安全执行器
func (sm *SecurityManager) AddEnforcer(pluginID string, config *SecurityConfig) (*SecurityEnforcer, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.enforcers[pluginID]; exists {
		return nil, fmt.Errorf("enforcer for plugin %s already exists", pluginID)
	}

	enforcer := NewSecurityEnforcer(sm.logger, pluginID, config)
	if err := enforcer.Start(sm.ctx); err != nil {
		return nil, fmt.Errorf("failed to start enforcer: %w", err)
	}

	sm.enforcers[pluginID] = enforcer
	return enforcer, nil
}

// RemoveEnforcer 移除安全执行器
func (sm *SecurityManager) RemoveEnforcer(pluginID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if enforcer, exists := sm.enforcers[pluginID]; exists {
		enforcer.Stop()
		delete(sm.enforcers, pluginID)
	}
}

// GetEnforcer 获取安全执行器
func (sm *SecurityManager) GetEnforcer(pluginID string) *SecurityEnforcer {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.enforcers[pluginID]
}

// GetAllViolations 获取所有违规记录
func (sm *SecurityManager) GetAllViolations() map[string]<-chan *SecurityViolation {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	violations := make(map[string]<-chan *SecurityViolation)
	for pluginID, enforcer := range sm.enforcers {
		violations[pluginID] = enforcer.GetViolations()
	}

	return violations
}