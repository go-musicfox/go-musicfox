package recovery

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// FallbackType 降级类型
type FallbackType int

const (
	// FallbackTypeService 服务降级
	FallbackTypeService FallbackType = iota
	// FallbackTypeFeature 功能降级
	FallbackTypeFeature
	// FallbackTypeCache 缓存降级
	FallbackTypeCache
	// FallbackTypeDefault 默认值降级
	FallbackTypeDefault
	// FallbackTypeCustom 自定义降级
	FallbackTypeCustom
)

// String 返回降级类型的字符串表示
func (ft FallbackType) String() string {
	switch ft {
	case FallbackTypeService:
		return "service"
	case FallbackTypeFeature:
		return "feature"
	case FallbackTypeCache:
		return "cache"
	case FallbackTypeDefault:
		return "default"
	case FallbackTypeCustom:
		return "custom"
	default:
		return "unknown"
	}
}

// FallbackConfig 降级配置
type FallbackConfig struct {
	// Type 降级类型
	Type FallbackType `json:"type"`
	// Enabled 是否启用降级
	Enabled bool `json:"enabled"`
	// Timeout 降级操作超时时间
	Timeout time.Duration `json:"timeout"`
	// MaxConcurrency 最大并发数
	MaxConcurrency int `json:"max_concurrency"`
	// CacheExpiry 缓存过期时间（缓存降级时使用）
	CacheExpiry time.Duration `json:"cache_expiry"`
	// DefaultValue 默认值（默认值降级时使用）
	DefaultValue interface{} `json:"default_value"`
	// ServiceEndpoint 备用服务端点（服务降级时使用）
	ServiceEndpoint string `json:"service_endpoint"`
	// FeatureFlags 功能开关（功能降级时使用）
	FeatureFlags map[string]bool `json:"feature_flags"`
	// Priority 降级优先级
	Priority int `json:"priority"`
}

// DefaultFallbackConfig 返回默认降级配置
func DefaultFallbackConfig() *FallbackConfig {
	return &FallbackConfig{
		Type:           FallbackTypeDefault,
		Enabled:        true,
		Timeout:        10 * time.Second,
		MaxConcurrency: 10,
		CacheExpiry:    5 * time.Minute,
		FeatureFlags:   make(map[string]bool),
		Priority:       1,
	}
}

// FallbackMetrics 降级指标
type FallbackMetrics struct {
	TotalFallbacks     int64     `json:"total_fallbacks"`
	SuccessFallbacks   int64     `json:"success_fallbacks"`
	FailedFallbacks    int64     `json:"failed_fallbacks"`
	CacheHits          int64     `json:"cache_hits"`
	CacheMisses        int64     `json:"cache_misses"`
	AverageLatency     time.Duration `json:"average_latency"`
	LastFallbackTime   time.Time `json:"last_fallback_time"`
	LastSuccessTime    time.Time `json:"last_success_time"`
}

// FallbackStrategy 降级策略实现
type FallbackStrategy struct {
	name     string
	config   *FallbackConfig
	logger   *slog.Logger
	metrics  *FallbackMetrics
	mutex    sync.RWMutex

	// 缓存存储
	cache    map[string]*CacheEntry
	cacheMutex sync.RWMutex

	// 降级函数
	fallbackFuncs map[FallbackType]func(ctx context.Context, args interface{}) (interface{}, error)

	// 回调函数
	onFallbackTriggered func(fallbackType FallbackType, reason string)
	onFallbackSuccess   func(fallbackType FallbackType, result interface{})
	onFallbackFailure   func(fallbackType FallbackType, err error)
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Value     interface{}
	Expiry    time.Time
	CreatedAt time.Time
}

// IsExpired 检查缓存是否过期
func (ce *CacheEntry) IsExpired() bool {
	return time.Now().After(ce.Expiry)
}

// NewFallbackStrategy 创建新的降级策略
func NewFallbackStrategy(name string, config *FallbackConfig, logger *slog.Logger) *FallbackStrategy {
	if config == nil {
		config = DefaultFallbackConfig()
	}

	fs := &FallbackStrategy{
		name:    name,
		config:  config,
		logger:  logger,
		metrics: &FallbackMetrics{
			LastSuccessTime: time.Now(),
		},
		cache:         make(map[string]*CacheEntry),
		fallbackFuncs: make(map[FallbackType]func(ctx context.Context, args interface{}) (interface{}, error)),
	}

	// 注册默认降级函数
	fs.registerDefaultFallbackFuncs()

	return fs
}

// registerDefaultFallbackFuncs 注册默认降级函数
func (fs *FallbackStrategy) registerDefaultFallbackFuncs() {
	// 默认值降级
	fs.fallbackFuncs[FallbackTypeDefault] = func(ctx context.Context, args interface{}) (interface{}, error) {
		return fs.config.DefaultValue, nil
	}

	// 缓存降级
	fs.fallbackFuncs[FallbackTypeCache] = func(ctx context.Context, args interface{}) (interface{}, error) {
		cacheKey := fmt.Sprintf("%v", args)
		return fs.getFromCache(cacheKey)
	}

	// 功能降级
	fs.fallbackFuncs[FallbackTypeFeature] = func(ctx context.Context, args interface{}) (interface{}, error) {
		// 返回简化的功能实现
		return map[string]interface{}{
			"status":  "degraded",
			"message": "Feature is running in degraded mode",
			"data":    nil,
		}, nil
	}
}

// Execute 执行降级策略
func (fs *FallbackStrategy) Execute(ctx context.Context, primaryOperation func(ctx context.Context) (interface{}, error), args interface{}) (interface{}, error) {
	if !fs.config.Enabled {
		return primaryOperation(ctx)
	}

	// 尝试执行主要操作
	result, err := fs.tryPrimaryOperation(ctx, primaryOperation)
	if err == nil {
		// 主要操作成功，缓存结果（如果是缓存降级）
		if fs.config.Type == FallbackTypeCache {
			cacheKey := fmt.Sprintf("%v", args)
			fs.putToCache(cacheKey, result)
		}
		return result, nil
	}

	// 主要操作失败，执行降级
	return fs.executeFallback(ctx, args, err)
}

// tryPrimaryOperation 尝试执行主要操作
func (fs *FallbackStrategy) tryPrimaryOperation(ctx context.Context, operation func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, fs.config.Timeout)
	defer cancel()

	// 执行操作
	return operation(ctx)
}

// executeFallback 执行降级
func (fs *FallbackStrategy) executeFallback(ctx context.Context, args interface{}, primaryErr error) (interface{}, error) {
	fs.recordFallbackAttempt()

	// 触发降级回调
	if fs.onFallbackTriggered != nil {
		fs.onFallbackTriggered(fs.config.Type, primaryErr.Error())
	}

	if fs.logger != nil {
		fs.logger.Warn("Executing fallback strategy",
			"strategy", fs.name,
			"type", fs.config.Type.String(),
			"primary_error", primaryErr.Error())
	}

	// 获取降级函数
	fallbackFunc, exists := fs.fallbackFuncs[fs.config.Type]
	if !exists {
		fs.recordFallbackFailure(fmt.Errorf("no fallback function for type %s", fs.config.Type.String()))
		return nil, fmt.Errorf("no fallback function registered for type %s in strategy '%s'", fs.config.Type.String(), fs.name)
	}

	// 执行降级函数
	result, err := fallbackFunc(ctx, args)
	if err != nil {
		fs.recordFallbackFailure(err)
		if fs.onFallbackFailure != nil {
			fs.onFallbackFailure(fs.config.Type, err)
		}
		return nil, fmt.Errorf("fallback execution failed in strategy '%s': %w", fs.name, err)
	}

	// 降级成功
	fs.recordFallbackSuccess()
	if fs.onFallbackSuccess != nil {
		fs.onFallbackSuccess(fs.config.Type, result)
	}

	if fs.logger != nil {
		fs.logger.Info("Fallback strategy executed successfully",
			"strategy", fs.name,
			"type", fs.config.Type.String())
	}

	return result, nil
}

// getFromCache 从缓存获取数据
func (fs *FallbackStrategy) getFromCache(key string) (interface{}, error) {
	fs.cacheMutex.RLock()
	entry, exists := fs.cache[key]
	fs.cacheMutex.RUnlock()

	if !exists {
		fs.recordCacheMiss()
		return nil, fmt.Errorf("cache miss for key: %s", key)
	}

	if entry.IsExpired() {
		// 清理过期缓存
		fs.cacheMutex.Lock()
		delete(fs.cache, key)
		fs.cacheMutex.Unlock()

		fs.recordCacheMiss()
		return nil, fmt.Errorf("cache expired for key: %s", key)
	}

	fs.recordCacheHit()
	return entry.Value, nil
}

// putToCache 将数据放入缓存
func (fs *FallbackStrategy) putToCache(key string, value interface{}) {
	fs.cacheMutex.Lock()
	defer fs.cacheMutex.Unlock()

	fs.cache[key] = &CacheEntry{
		Value:     value,
		Expiry:    time.Now().Add(fs.config.CacheExpiry),
		CreatedAt: time.Now(),
	}
}

// clearExpiredCache 清理过期缓存
func (fs *FallbackStrategy) clearExpiredCache() {
	fs.cacheMutex.Lock()
	defer fs.cacheMutex.Unlock()

	now := time.Now()
	for key, entry := range fs.cache {
		if now.After(entry.Expiry) {
			delete(fs.cache, key)
		}
	}
}

// RegisterFallbackFunc 注册降级函数
func (fs *FallbackStrategy) RegisterFallbackFunc(fallbackType FallbackType, fn func(ctx context.Context, args interface{}) (interface{}, error)) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	fs.fallbackFuncs[fallbackType] = fn

	if fs.logger != nil {
		fs.logger.Info("Fallback function registered",
			"strategy", fs.name,
			"type", fallbackType.String())
	}
}

// recordFallbackAttempt 记录降级尝试
func (fs *FallbackStrategy) recordFallbackAttempt() {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	fs.metrics.TotalFallbacks++
	fs.metrics.LastFallbackTime = time.Now()
}

// recordFallbackSuccess 记录降级成功
func (fs *FallbackStrategy) recordFallbackSuccess() {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	fs.metrics.SuccessFallbacks++
	fs.metrics.LastSuccessTime = time.Now()
}

// recordFallbackFailure 记录降级失败
func (fs *FallbackStrategy) recordFallbackFailure(err error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	fs.metrics.FailedFallbacks++

	if fs.logger != nil {
		fs.logger.Error("Fallback strategy failed",
			"strategy", fs.name,
			"error", err.Error())
	}
}

// recordCacheHit 记录缓存命中
func (fs *FallbackStrategy) recordCacheHit() {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	fs.metrics.CacheHits++
}

// recordCacheMiss 记录缓存未命中
func (fs *FallbackStrategy) recordCacheMiss() {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	fs.metrics.CacheMisses++
}

// GetMetrics 获取指标
func (fs *FallbackStrategy) GetMetrics() *FallbackMetrics {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	// 返回指标的副本
	metrics := *fs.metrics
	return &metrics
}

// Reset 重置降级策略
func (fs *FallbackStrategy) Reset() {
	fs.mutex.Lock()
	fs.cacheMutex.Lock()
	defer fs.mutex.Unlock()
	defer fs.cacheMutex.Unlock()

	fs.metrics = &FallbackMetrics{
		LastSuccessTime: time.Now(),
	}
	fs.cache = make(map[string]*CacheEntry)

	if fs.logger != nil {
		fs.logger.Info("Fallback strategy reset", "strategy", fs.name)
	}
}

// SetOnFallbackTriggered 设置降级触发回调
func (fs *FallbackStrategy) SetOnFallbackTriggered(callback func(fallbackType FallbackType, reason string)) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	fs.onFallbackTriggered = callback
}

// SetOnFallbackSuccess 设置降级成功回调
func (fs *FallbackStrategy) SetOnFallbackSuccess(callback func(fallbackType FallbackType, result interface{})) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	fs.onFallbackSuccess = callback
}

// SetOnFallbackFailure 设置降级失败回调
func (fs *FallbackStrategy) SetOnFallbackFailure(callback func(fallbackType FallbackType, err error)) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	fs.onFallbackFailure = callback
}

// GetName 获取策略名称
func (fs *FallbackStrategy) GetName() string {
	return fs.name
}

// GetConfig 获取配置
func (fs *FallbackStrategy) GetConfig() *FallbackConfig {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	// 返回配置的副本
	config := *fs.config
	return &config
}

// UpdateConfig 更新配置
func (fs *FallbackStrategy) UpdateConfig(config *FallbackConfig) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	fs.config = config

	if fs.logger != nil {
		fs.logger.Info("Fallback strategy config updated",
			"strategy", fs.name,
			"type", config.Type.String(),
			"enabled", config.Enabled)
	}
}

// GetSuccessRate 获取成功率
func (fs *FallbackStrategy) GetSuccessRate() float64 {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	if fs.metrics.TotalFallbacks == 0 {
		return 0.0
	}

	return float64(fs.metrics.SuccessFallbacks) / float64(fs.metrics.TotalFallbacks)
}

// GetCacheHitRate 获取缓存命中率
func (fs *FallbackStrategy) GetCacheHitRate() float64 {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	totalCacheRequests := fs.metrics.CacheHits + fs.metrics.CacheMisses
	if totalCacheRequests == 0 {
		return 0.0
	}

	return float64(fs.metrics.CacheHits) / float64(totalCacheRequests)
}

// StartCacheCleanup 启动缓存清理任务
func (fs *FallbackStrategy) StartCacheCleanup(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fs.clearExpiredCache()
			}
		}
	}()
}