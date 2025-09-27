package config

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/patrickmn/go-cache"
)

// PerformanceConfig 性能配置
type PerformanceConfig struct {
	// 缓存配置
	CacheEnabled         bool          `json:"cache_enabled"`
	CacheDefaultTTL      time.Duration `json:"cache_default_ttl"`
	CacheCleanupInterval time.Duration `json:"cache_cleanup_interval"`
	CacheMaxSize         int           `json:"cache_max_size"`

	// 并发配置
	MaxConcurrentReads  int `json:"max_concurrent_reads"`
	MaxConcurrentWrites int `json:"max_concurrent_writes"`
	ReadTimeout         time.Duration `json:"read_timeout"`
	WriteTimeout        time.Duration `json:"write_timeout"`

	// 内存管理
	MemoryLimitMB       int64         `json:"memory_limit_mb"`
	GCInterval          time.Duration `json:"gc_interval"`
	MemoryCheckInterval time.Duration `json:"memory_check_interval"`

	// 压缩配置
	CompressionEnabled   bool    `json:"compression_enabled"`
	CompressionThreshold int     `json:"compression_threshold"`
	CompressionLevel     int     `json:"compression_level"`
	CompressionRatio     float64 `json:"compression_ratio"`
}

// DefaultPerformanceConfig 默认性能配置
func DefaultPerformanceConfig() *PerformanceConfig {
	return &PerformanceConfig{
		CacheEnabled:         true,
		CacheDefaultTTL:      5 * time.Minute,
		CacheCleanupInterval: 10 * time.Minute,
		CacheMaxSize:         1000,
		MaxConcurrentReads:   100,
		MaxConcurrentWrites:  10,
		ReadTimeout:          5 * time.Second,
		WriteTimeout:         10 * time.Second,
		MemoryLimitMB:        100,
		GCInterval:           30 * time.Second,
		MemoryCheckInterval:  10 * time.Second,
		CompressionEnabled:   true,
		CompressionThreshold: 1024, // 1KB
		CompressionLevel:     6,
		CompressionRatio:     0.7,
	}
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	// 访问控制
	EnableAccessControl    bool          `json:"enable_access_control"`
	MaxFailedAttempts      int           `json:"max_failed_attempts"`
	LockoutDuration        time.Duration `json:"lockout_duration"`
	SessionTimeout         time.Duration `json:"session_timeout"`

	// 加密配置
	EncryptionAlgorithm    string        `json:"encryption_algorithm"`
	KeyRotationInterval    time.Duration `json:"key_rotation_interval"`
	KeyDerivationIterations int          `json:"key_derivation_iterations"`

	// 审计配置
	EnableAuditLog         bool          `json:"enable_audit_log"`
	AuditLogMaxSize        int64         `json:"audit_log_max_size"`
	AuditLogRetention      time.Duration `json:"audit_log_retention"`

	// 安全检查
	EnableIntegrityCheck   bool          `json:"enable_integrity_check"`
	IntegrityCheckInterval time.Duration `json:"integrity_check_interval"`
	EnableAntiTampering    bool          `json:"enable_anti_tampering"`
}

// DefaultSecurityConfig 默认安全配置
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		EnableAccessControl:     true,
		MaxFailedAttempts:       5,
		LockoutDuration:         15 * time.Minute,
		SessionTimeout:          30 * time.Minute,
		EncryptionAlgorithm:     "AES-256-GCM",
		KeyRotationInterval:     24 * time.Hour,
		KeyDerivationIterations: 100000,
		EnableAuditLog:          true,
		AuditLogMaxSize:         100 * 1024 * 1024, // 100MB
		AuditLogRetention:       30 * 24 * time.Hour, // 30 days
		EnableIntegrityCheck:    true,
		IntegrityCheckInterval:  1 * time.Hour,
		EnableAntiTampering:     true,
	}
}

// OptimizedManager 优化的配置管理器
type OptimizedManager struct {
	*AdvancedManager

	// 性能优化
	performanceConfig *PerformanceConfig
	cache             *cache.Cache
	readSemaphore     chan struct{}
	writeSemaphore    chan struct{}
	memoryUsage       int64
	lastGC            time.Time

	// 安全加固
	securityConfig    *SecurityConfig
	failedAttempts    map[string]int
	lockedUsers       map[string]time.Time
	sessions          map[string]*Session
	integrityHash     []byte
	lastIntegrityCheck time.Time
	securityMutex     sync.RWMutex

	// 监控统计
	metrics           *PerformanceMetrics
	metricsMutex      sync.RWMutex

	// 生命周期管理
	optimizationCtx    context.Context
	optimizationCancel context.CancelFunc
	running            bool
}

// Session 会话信息
type Session struct {
	ID        string    `json:"id"`
	User      string    `json:"user"`
	CreatedAt time.Time `json:"created_at"`
	LastAccess time.Time `json:"last_access"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	// 操作统计
	ReadCount       int64 `json:"read_count"`
	WriteCount      int64 `json:"write_count"`
	CacheHitCount   int64 `json:"cache_hit_count"`
	CacheMissCount  int64 `json:"cache_miss_count"`

	// 性能统计
	AvgReadTime     time.Duration `json:"avg_read_time"`
	AvgWriteTime    time.Duration `json:"avg_write_time"`
	MaxReadTime     time.Duration `json:"max_read_time"`
	MaxWriteTime    time.Duration `json:"max_write_time"`

	// 资源统计
	MemoryUsage     int64 `json:"memory_usage"`
	MaxMemoryUsage  int64 `json:"max_memory_usage"`
	GCCount         int64 `json:"gc_count"`
	LastGCTime      time.Time `json:"last_gc_time"`

	// 安全统计
	FailedAttempts  int64 `json:"failed_attempts"`
	BlockedAttempts int64 `json:"blocked_attempts"`
	ActiveSessions  int   `json:"active_sessions"`
}

// NewOptimizedManager 创建优化的配置管理器
func NewOptimizedManager(configDir, configFile string, perfConfig *PerformanceConfig, secConfig *SecurityConfig) *OptimizedManager {
	if perfConfig == nil {
		perfConfig = DefaultPerformanceConfig()
	}
	if secConfig == nil {
		secConfig = DefaultSecurityConfig()
	}

	baseManager := NewAdvancedManager(configDir, configFile)

	om := &OptimizedManager{
		AdvancedManager:   baseManager,
		performanceConfig: perfConfig,
		securityConfig:    secConfig,
		failedAttempts:    make(map[string]int),
		lockedUsers:       make(map[string]time.Time),
		sessions:          make(map[string]*Session),
		metrics:           &PerformanceMetrics{},
	}

	// 初始化缓存
	if perfConfig.CacheEnabled {
		om.cache = cache.New(perfConfig.CacheDefaultTTL, perfConfig.CacheCleanupInterval)
	}

	// 初始化信号量
	om.readSemaphore = make(chan struct{}, perfConfig.MaxConcurrentReads)
	om.writeSemaphore = make(chan struct{}, perfConfig.MaxConcurrentWrites)

	// 填充信号量
	for i := 0; i < perfConfig.MaxConcurrentReads; i++ {
		om.readSemaphore <- struct{}{}
	}
	for i := 0; i < perfConfig.MaxConcurrentWrites; i++ {
		om.writeSemaphore <- struct{}{}
	}

	return om
}

// Start 启动优化管理器
func (om *OptimizedManager) Start(ctx context.Context) error {
	if om.running {
		return fmt.Errorf("optimized manager is already running")
	}

	om.optimizationCtx, om.optimizationCancel = context.WithCancel(ctx)
	om.running = true

	// 启动后台任务
	go om.memoryMonitor()
	go om.securityMonitor()
	go om.performanceMonitor()
	go om.integrityChecker()

	return nil
}

// Stop 停止优化管理器
func (om *OptimizedManager) Stop() error {
	if !om.running {
		return nil
	}

	if om.optimizationCancel != nil {
		om.optimizationCancel()
	}

	om.running = false
	return nil
}

// OptimizedGet 优化的配置获取
func (om *OptimizedManager) OptimizedGet(key string, user string) (interface{}, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		om.updateReadMetrics(duration)
	}()

	// 安全检查
	if err := om.checkAccess("read", key, user); err != nil {
		return nil, err
	}

	// 获取读取信号量
	select {
	case <-om.readSemaphore:
		defer func() { om.readSemaphore <- struct{}{} }()
	case <-time.After(om.performanceConfig.ReadTimeout):
		return nil, fmt.Errorf("read timeout")
	}

	// 尝试从缓存获取
	if om.cache != nil {
		if cached, found := om.cache.Get(key); found {
			atomic.AddInt64(&om.metrics.CacheHitCount, 1)
			return cached, nil
		}
		atomic.AddInt64(&om.metrics.CacheMissCount, 1)
	}

	// 从配置获取
	value := om.AdvancedManager.Get(key)

	// 缓存结果
	if om.cache != nil && value != nil {
		om.cache.Set(key, value, cache.DefaultExpiration)
	}

	return value, nil
}

// OptimizedSet 优化的配置设置
func (om *OptimizedManager) OptimizedSet(key string, value interface{}, user string) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		om.updateWriteMetrics(duration)
	}()

	// 安全检查
	if err := om.checkAccess("write", key, user); err != nil {
		return err
	}

	// 获取写入信号量
	select {
	case <-om.writeSemaphore:
		defer func() { om.writeSemaphore <- struct{}{} }()
	case <-time.After(om.performanceConfig.WriteTimeout):
		return fmt.Errorf("write timeout")
	}

	// 设置配置
	err := om.AdvancedManager.Set(key, value)
	if err != nil {
		return err
	}

	// 更新缓存
	if om.cache != nil {
		om.cache.Set(key, value, cache.DefaultExpiration)
	}

	// 记录审计日志
	if om.securityConfig.EnableAuditLog {
		om.logAuditEvent("config_set", user, key, value)
	}

	return nil
}

// checkAccess 检查访问权限（增强版）
func (om *OptimizedManager) checkAccess(operation, key, user string) error {
	om.securityMutex.RLock()
	defer om.securityMutex.RUnlock()

	// 检查用户是否被锁定
	if lockTime, locked := om.lockedUsers[user]; locked {
		if time.Since(lockTime) < om.securityConfig.LockoutDuration {
			atomic.AddInt64(&om.metrics.BlockedAttempts, 1)
			return fmt.Errorf("user %s is locked out", user)
		}
		// 锁定时间已过，解除锁定
		delete(om.lockedUsers, user)
		delete(om.failedAttempts, user)
	}

	// 检查会话有效性
	if om.securityConfig.EnableAccessControl {
		if !om.isValidSession(user) {
			return fmt.Errorf("invalid or expired session for user %s", user)
		}
	}

	// 使用基础访问控制检查
	if !om.AdvancedManager.CheckAccess(operation, key, user) {
		// 记录失败尝试
		om.recordFailedAttempt(user)
		return fmt.Errorf("access denied for user %s on key %s", user, key)
	}

	return nil
}

// recordFailedAttempt 记录失败尝试
func (om *OptimizedManager) recordFailedAttempt(user string) {
	om.securityMutex.Lock()
	defer om.securityMutex.Unlock()

	om.failedAttempts[user]++
	atomic.AddInt64(&om.metrics.FailedAttempts, 1)

	// 检查是否需要锁定用户
	if om.failedAttempts[user] >= om.securityConfig.MaxFailedAttempts {
		om.lockedUsers[user] = time.Now()
		fmt.Printf("User %s locked out due to %d failed attempts\n", user, om.failedAttempts[user])
	}
}

// isValidSession 检查会话有效性
func (om *OptimizedManager) isValidSession(user string) bool {
	session, exists := om.sessions[user]
	if !exists {
		return false
	}

	// 检查会话是否过期
	if time.Since(session.LastAccess) > om.securityConfig.SessionTimeout {
		delete(om.sessions, user)
		return false
	}

	// 更新最后访问时间
	session.LastAccess = time.Now()
	return true
}

// CreateSession 创建会话
func (om *OptimizedManager) CreateSession(user, ipAddress, userAgent string) (*Session, error) {
	om.securityMutex.Lock()
	defer om.securityMutex.Unlock()

	// 生成会话ID
	sessionID, err := om.generateSecureID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	session := &Session{
		ID:         sessionID,
		User:       user,
		CreatedAt:  time.Now(),
		LastAccess: time.Now(),
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}

	om.sessions[user] = session
	return session, nil
}

// generateSecureID 生成安全ID
func (om *OptimizedManager) generateSecureID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", bytes), nil
}

// memoryMonitor 内存监控
func (om *OptimizedManager) memoryMonitor() {
	ticker := time.NewTicker(om.performanceConfig.MemoryCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-om.optimizationCtx.Done():
			return
		case <-ticker.C:
			om.checkMemoryUsage()
		}
	}
}

// checkMemoryUsage 检查内存使用情况
func (om *OptimizedManager) checkMemoryUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	currentUsage := int64(m.Alloc / 1024 / 1024) // MB
	atomic.StoreInt64(&om.memoryUsage, currentUsage)

	om.metricsMutex.Lock()
	om.metrics.MemoryUsage = currentUsage
	if currentUsage > om.metrics.MaxMemoryUsage {
		om.metrics.MaxMemoryUsage = currentUsage
	}
	om.metricsMutex.Unlock()

	// 检查是否超过内存限制
	if currentUsage > om.performanceConfig.MemoryLimitMB {
		fmt.Printf("Warning: Memory usage (%d MB) exceeds limit (%d MB)\n", currentUsage, om.performanceConfig.MemoryLimitMB)
		om.triggerGC()
	}

	// 定期垃圾回收
	if time.Since(om.lastGC) > om.performanceConfig.GCInterval {
		om.triggerGC()
	}
}

// triggerGC 触发垃圾回收
func (om *OptimizedManager) triggerGC() {
	runtime.GC()
	om.lastGC = time.Now()

	om.metricsMutex.Lock()
	om.metrics.GCCount++
	om.metrics.LastGCTime = om.lastGC
	om.metricsMutex.Unlock()

	// 清理过期缓存
	if om.cache != nil {
		om.cache.DeleteExpired()
	}
}

// securityMonitor 安全监控
func (om *OptimizedManager) securityMonitor() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-om.optimizationCtx.Done():
			return
		case <-ticker.C:
			om.cleanupExpiredSessions()
			om.checkKeyRotation()
		}
	}
}

// cleanupExpiredSessions 清理过期会话
func (om *OptimizedManager) cleanupExpiredSessions() {
	om.securityMutex.Lock()
	defer om.securityMutex.Unlock()

	now := time.Now()
	for user, session := range om.sessions {
		if now.Sub(session.LastAccess) > om.securityConfig.SessionTimeout {
			delete(om.sessions, user)
		}
	}

	om.metricsMutex.Lock()
	om.metrics.ActiveSessions = len(om.sessions)
	om.metricsMutex.Unlock()
}

// checkKeyRotation 检查密钥轮换
func (om *OptimizedManager) checkKeyRotation() {
	// 这里应该实现密钥轮换逻辑
	// 为了简化，只是记录检查时间
	if om.securityConfig.KeyRotationInterval > 0 {
		// 实际实现中应该检查上次轮换时间并执行轮换
		fmt.Printf("Key rotation check at %s\n", time.Now().Format(time.RFC3339))
	}
}

// performanceMonitor 性能监控
func (om *OptimizedManager) performanceMonitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-om.optimizationCtx.Done():
			return
		case <-ticker.C:
			om.reportPerformanceMetrics()
		}
	}
}

// reportPerformanceMetrics 报告性能指标
func (om *OptimizedManager) reportPerformanceMetrics() {
	om.metricsMutex.RLock()
	metrics := *om.metrics
	om.metricsMutex.RUnlock()

	// 计算缓存命中率
	totalCacheRequests := metrics.CacheHitCount + metrics.CacheMissCount
	cacheHitRate := float64(0)
	if totalCacheRequests > 0 {
		cacheHitRate = float64(metrics.CacheHitCount) / float64(totalCacheRequests) * 100
	}

	fmt.Printf("Performance Metrics: Reads=%d, Writes=%d, Cache Hit Rate=%.2f%%, Memory=%dMB\n",
		metrics.ReadCount, metrics.WriteCount, cacheHitRate, metrics.MemoryUsage)
}

// integrityChecker 完整性检查
func (om *OptimizedManager) integrityChecker() {
	ticker := time.NewTicker(om.securityConfig.IntegrityCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-om.optimizationCtx.Done():
			return
		case <-ticker.C:
			if om.securityConfig.EnableIntegrityCheck {
				om.performIntegrityCheck()
			}
		}
	}
}

// performIntegrityCheck 执行完整性检查
func (om *OptimizedManager) performIntegrityCheck() {
	// 计算当前配置的哈希值
	currentHash := om.calculateConfigHash()

	// 如果是第一次检查，保存哈希值
	if om.integrityHash == nil {
		om.integrityHash = currentHash
		om.lastIntegrityCheck = time.Now()
		return
	}

	// 比较哈希值
	if !om.compareHashes(om.integrityHash, currentHash) {
		fmt.Printf("Warning: Configuration integrity check failed at %s\n", time.Now().Format(time.RFC3339))
		// 这里可以触发安全事件或恢复操作
	}

	om.integrityHash = currentHash
	om.lastIntegrityCheck = time.Now()
}

// calculateConfigHash 计算配置哈希值
func (om *OptimizedManager) calculateConfigHash() []byte {
	configData := om.AdvancedManager.k.All()
	return []byte(calculateChecksum(configData))
}

// compareHashes 安全比较哈希值
func (om *OptimizedManager) compareHashes(hash1, hash2 []byte) bool {
	return subtle.ConstantTimeCompare(hash1, hash2) == 1
}

// updateReadMetrics 更新读取指标
func (om *OptimizedManager) updateReadMetrics(duration time.Duration) {
	om.metricsMutex.Lock()
	defer om.metricsMutex.Unlock()

	om.metrics.ReadCount++
	if duration > om.metrics.MaxReadTime {
		om.metrics.MaxReadTime = duration
	}

	// 计算平均时间（简化实现）
	if om.metrics.ReadCount == 1 {
		om.metrics.AvgReadTime = duration
	} else {
		om.metrics.AvgReadTime = (om.metrics.AvgReadTime + duration) / 2
	}
}

// updateWriteMetrics 更新写入指标
func (om *OptimizedManager) updateWriteMetrics(duration time.Duration) {
	om.metricsMutex.Lock()
	defer om.metricsMutex.Unlock()

	om.metrics.WriteCount++
	if duration > om.metrics.MaxWriteTime {
		om.metrics.MaxWriteTime = duration
	}

	// 计算平均时间（简化实现）
	if om.metrics.WriteCount == 1 {
		om.metrics.AvgWriteTime = duration
	} else {
		om.metrics.AvgWriteTime = (om.metrics.AvgWriteTime + duration) / 2
	}
}

// logAuditEvent 记录审计事件
func (om *OptimizedManager) logAuditEvent(action, user, key string, value interface{}) {
	// 简化的审计日志实现
	fmt.Printf("AUDIT: %s - User: %s, Action: %s, Key: %s, Time: %s\n",
		time.Now().Format(time.RFC3339), user, action, key, time.Now().Format(time.RFC3339))
}

// GetPerformanceMetrics 获取性能指标
func (om *OptimizedManager) GetPerformanceMetrics() *PerformanceMetrics {
	om.metricsMutex.RLock()
	defer om.metricsMutex.RUnlock()

	// 返回指标副本
	metrics := *om.metrics
	return &metrics
}

// GetSecurityStatus 获取安全状态
func (om *OptimizedManager) GetSecurityStatus() map[string]interface{} {
	om.securityMutex.RLock()
	defer om.securityMutex.RUnlock()

	return map[string]interface{}{
		"active_sessions":    len(om.sessions),
		"locked_users":       len(om.lockedUsers),
		"failed_attempts":    len(om.failedAttempts),
		"last_integrity_check": om.lastIntegrityCheck,
		"encryption_enabled": len(om.AdvancedManager.encryptionKey) > 0,
		"access_control_enabled": om.securityConfig.EnableAccessControl,
	}
}

// ClearCache 清空缓存
func (om *OptimizedManager) ClearCache() {
	if om.cache != nil {
		om.cache.Flush()
	}
}

// GetCacheStats 获取缓存统计
func (om *OptimizedManager) GetCacheStats() map[string]interface{} {
	if om.cache == nil {
		return map[string]interface{}{"enabled": false}
	}

	return map[string]interface{}{
		"enabled":    true,
		"item_count": om.cache.ItemCount(),
		"hit_count":  atomic.LoadInt64(&om.metrics.CacheHitCount),
		"miss_count": atomic.LoadInt64(&om.metrics.CacheMissCount),
	}
}