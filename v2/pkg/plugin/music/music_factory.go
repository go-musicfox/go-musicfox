package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"

	core "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
)

// MusicSourceFactory 音乐源工厂
type MusicSourceFactory struct {
	registeredSources map[string]MusicSourceCreator
	activeSources     map[string]MusicSourcePlugin
	cacheManager      *CacheManager
	loadBalancer      *LoadBalancer
	healthChecker     *MusicSourceHealthChecker
	mu                sync.RWMutex
	config            *FactoryConfig
}

// MusicSourceCreator 音乐源创建函数类型
type MusicSourceCreator func(config map[string]interface{}) (MusicSourcePlugin, error)

// FactoryConfig 工厂配置
type FactoryConfig struct {
	DefaultSource     string                 `json:"default_source"`
	LoadBalancing     bool                   `json:"load_balancing"`
	HealthCheck       bool                   `json:"health_check"`
	HealthCheckInterval time.Duration        `json:"health_check_interval"`
	FailoverEnabled   bool                   `json:"failover_enabled"`
	MaxRetries        int                    `json:"max_retries"`
	SourceConfigs     map[string]interface{} `json:"source_configs"`
}

// LoadBalancer 负载均衡器
type LoadBalancer struct {
	strategy      LoadBalanceStrategy
	sources       []string
	currentIndex  int
	weights       map[string]int
	mu            sync.RWMutex
}

// LoadBalanceStrategy 负载均衡策略
type LoadBalanceStrategy int

const (
	LoadBalanceRoundRobin LoadBalanceStrategy = iota
	LoadBalanceWeighted
	LoadBalanceRandom
	LoadBalanceLeastConnections
)

// MusicSourceHealthChecker 音乐源健康检查器
type MusicSourceHealthChecker struct {
	factory       *MusicSourceFactory
	checkInterval time.Duration
	stopChan      chan struct{}
	mu            sync.RWMutex
	lastCheck     map[string]time.Time
	healthStatus  map[string]HealthStatus
}

// NewMusicSourceFactory 创建音乐源工厂
func NewMusicSourceFactory(config *FactoryConfig) *MusicSourceFactory {
	if config == nil {
		config = &FactoryConfig{
			LoadBalancing:       true,
			HealthCheck:         true,
			HealthCheckInterval: 30 * time.Second,
			FailoverEnabled:     true,
			MaxRetries:          3,
			SourceConfigs:       make(map[string]interface{}),
		}
	}

	factory := &MusicSourceFactory{
		registeredSources: make(map[string]MusicSourceCreator),
		activeSources:     make(map[string]MusicSourcePlugin),
		cacheManager:      NewCacheManager(),
		config:            config,
	}

	// 初始化负载均衡器
	if config.LoadBalancing {
		factory.loadBalancer = &LoadBalancer{
			strategy: LoadBalanceRoundRobin,
			sources:  make([]string, 0),
			weights:  make(map[string]int),
		}
	}

	// 初始化健康检查器
	if config.HealthCheck {
		factory.healthChecker = &MusicSourceHealthChecker{
			factory:       factory,
			checkInterval: config.HealthCheckInterval,
			stopChan:      make(chan struct{}),
			lastCheck:     make(map[string]time.Time),
			healthStatus:  make(map[string]HealthStatus),
		}
		go factory.healthChecker.start()
	}

	// 注册内置音乐源
	factory.registerBuiltinSources()

	return factory
}

// RegisterSource 注册音乐源
func (f *MusicSourceFactory) RegisterSource(name string, creator MusicSourceCreator) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.registeredSources[name]; exists {
		return fmt.Errorf("source %s already registered", name)
	}

	f.registeredSources[name] = creator

	// 添加到负载均衡器
	if f.loadBalancer != nil {
		f.loadBalancer.mu.Lock()
		f.loadBalancer.sources = append(f.loadBalancer.sources, name)
		f.loadBalancer.weights[name] = 1 // 默认权重
		f.loadBalancer.mu.Unlock()
	}

	return nil
}

// CreateSource 创建音乐源实例
func (f *MusicSourceFactory) CreateSource(name string, config map[string]interface{}) (MusicSourcePlugin, error) {
	f.mu.RLock()
	creator, exists := f.registeredSources[name]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("source %s not registered", name)
	}

	// 合并配置
	finalConfig := make(map[string]interface{})
	if sourceConfig, ok := f.config.SourceConfigs[name]; ok {
		if configMap, ok := sourceConfig.(map[string]interface{}); ok {
			for k, v := range configMap {
				finalConfig[k] = v
			}
		}
	}
	for k, v := range config {
		finalConfig[k] = v
	}

	// 创建音乐源
	source, err := creator(finalConfig)
	if err != nil {
		return nil, err
	}

	// 设置缓存
	if cache := f.cacheManager.GetDefaultCache(); cache != nil {
		if baseSrc, ok := source.(*BaseMusicSourcePlugin); ok {
			baseSrc.SetCache(cache)
		}
	}

	// 保存活跃源
	f.mu.Lock()
	f.activeSources[name] = source
	f.mu.Unlock()

	// 初始化健康状态
	if f.healthChecker != nil {
		f.healthChecker.mu.Lock()
		f.healthChecker.healthStatus[name] = HealthStatusHealthy()
		f.healthChecker.lastCheck[name] = time.Now()
		f.healthChecker.mu.Unlock()
	}

	return source, nil
}

// GetSource 获取音乐源实例
func (f *MusicSourceFactory) GetSource(name string) (MusicSourcePlugin, error) {
	f.mu.RLock()
	source, exists := f.activeSources[name]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("source %s not found or not active", name)
	}

	// 检查健康状态
	if f.healthChecker != nil {
		f.healthChecker.mu.RLock()
		status := f.healthChecker.healthStatus[name]
		f.healthChecker.mu.RUnlock()

		if !status.Healthy {
			return nil, fmt.Errorf("source %s is not healthy: %s", name, status.Message)
		}
	}

	return source, nil
}

// GetDefaultSource 获取默认音乐源
func (f *MusicSourceFactory) GetDefaultSource() (MusicSourcePlugin, error) {
	if f.config.DefaultSource != "" {
		return f.GetSource(f.config.DefaultSource)
	}

	// 使用负载均衡选择源
	if f.loadBalancer != nil {
		sourceName := f.loadBalancer.selectSource()
		if sourceName != "" {
			return f.GetSource(sourceName)
		}
	}

	// 返回第一个可用的源
	f.mu.RLock()
	defer f.mu.RUnlock()

	for name, source := range f.activeSources {
		if f.isSourceHealthy(name) {
			return source, nil
		}
	}

	return nil, fmt.Errorf("no healthy sources available")
}

// GetAvailableSources 获取可用的音乐源列表
func (f *MusicSourceFactory) GetAvailableSources() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	sources := make([]string, 0, len(f.activeSources))
	for name := range f.activeSources {
		if f.isSourceHealthy(name) {
			sources = append(sources, name)
		}
	}

	return sources
}

// SearchWithFailover 带故障转移的搜索
func (f *MusicSourceFactory) SearchWithFailover(ctx context.Context, query string, options SearchOptions) (*SearchResult, error) {
	availableSources := f.GetAvailableSources()
	if len(availableSources) == 0 {
		return nil, fmt.Errorf("no available sources")
	}

	var lastErr error
	for attempt := 0; attempt < f.config.MaxRetries; attempt++ {
		for _, sourceName := range availableSources {
			source, err := f.GetSource(sourceName)
			if err != nil {
				lastErr = err
				continue
			}

			result, err := source.Search(ctx, query, options)
			if err == nil {
				return result, nil
			}

			lastErr = err
			// 标记源为不健康
			f.markSourceUnhealthy(sourceName)
		}
	}

	return nil, fmt.Errorf("all sources failed, last error: %v", lastErr)
}

// GetTrackURLWithFailover 带故障转移的获取音轨URL
func (f *MusicSourceFactory) GetTrackURLWithFailover(ctx context.Context, trackID string, quality AudioQuality) (string, error) {
	availableSources := f.GetAvailableSources()
	if len(availableSources) == 0 {
		return "", fmt.Errorf("no available sources")
	}

	var lastErr error
	for _, sourceName := range availableSources {
		source, err := f.GetSource(sourceName)
		if err != nil {
			lastErr = err
			continue
		}

		url, err := source.GetTrackURL(ctx, trackID, quality)
		if err == nil {
			return url, nil
		}

		lastErr = err
	}

	return "", fmt.Errorf("all sources failed, last error: %v", lastErr)
}

// SetLoadBalanceStrategy 设置负载均衡策略
func (f *MusicSourceFactory) SetLoadBalanceStrategy(strategy LoadBalanceStrategy) {
	if f.loadBalancer != nil {
		f.loadBalancer.mu.Lock()
		f.loadBalancer.strategy = strategy
		f.loadBalancer.mu.Unlock()
	}
}

// SetSourceWeight 设置源权重
func (f *MusicSourceFactory) SetSourceWeight(sourceName string, weight int) {
	if f.loadBalancer != nil {
		f.loadBalancer.mu.Lock()
		f.loadBalancer.weights[sourceName] = weight
		f.loadBalancer.mu.Unlock()
	}
}

// registerBuiltinSources 注册内置音乐源
func (f *MusicSourceFactory) registerBuiltinSources() {
	// 注册网易云音乐
	f.RegisterSource("netease", func(config map[string]interface{}) (MusicSourcePlugin, error) {
		apiBase, ok := config["api_base"].(string)
		if !ok {
			apiBase = "https://music.163.com/api"
		}
		return NewNeteaseAdapter(apiBase), nil
	})

	// 注册Spotify
	f.RegisterSource("spotify", func(config map[string]interface{}) (MusicSourcePlugin, error) {
		clientID, ok := config["client_id"].(string)
		if !ok {
			return nil, fmt.Errorf("spotify client_id required")
		}
		clientSecret, ok := config["client_secret"].(string)
		if !ok {
			return nil, fmt.Errorf("spotify client_secret required")
		}
		return NewSpotifyAdapter(clientID, clientSecret), nil
	})

	// 注册本地音乐
	f.RegisterSource("local", func(config map[string]interface{}) (MusicSourcePlugin, error) {
		musicDirsInterface, ok := config["music_dirs"]
		if !ok {
			return nil, fmt.Errorf("local music_dirs required")
		}

		musicDirs := make([]string, 0)
		if dirs, ok := musicDirsInterface.([]interface{}); ok {
			for _, dir := range dirs {
				if dirStr, ok := dir.(string); ok {
					musicDirs = append(musicDirs, dirStr)
				}
			}
		}

		return NewLocalMusicAdapter(musicDirs), nil
	})
}

// isSourceHealthy 检查源是否健康
func (f *MusicSourceFactory) isSourceHealthy(sourceName string) bool {
	if f.healthChecker == nil {
		return true
	}

	f.healthChecker.mu.RLock()
	status := f.healthChecker.healthStatus[sourceName]
	f.healthChecker.mu.RUnlock()

	return status.Healthy
}

// markSourceUnhealthy 标记源为不健康
func (f *MusicSourceFactory) markSourceUnhealthy(sourceName string) {
	if f.healthChecker != nil {
		f.healthChecker.mu.Lock()
		f.healthChecker.healthStatus[sourceName] = HealthStatus{
			Healthy:   false,
			Message:   "Source marked as unhealthy",
			Timestamp: time.Now(),
		}
		f.healthChecker.mu.Unlock()
	}
}

// LoadBalancer 方法

// selectSource 选择音乐源
func (lb *LoadBalancer) selectSource() string {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if len(lb.sources) == 0 {
		return ""
	}

	switch lb.strategy {
	case LoadBalanceRoundRobin:
		source := lb.sources[lb.currentIndex]
		lb.currentIndex = (lb.currentIndex + 1) % len(lb.sources)
		return source

	case LoadBalanceWeighted:
		return lb.selectWeightedSource()

	case LoadBalanceRandom:
		// 简化的随机选择
		return lb.sources[time.Now().Unix()%int64(len(lb.sources))]

	default:
		return lb.sources[0]
	}
}

// selectWeightedSource 加权选择
func (lb *LoadBalancer) selectWeightedSource() string {
	totalWeight := 0
	for _, source := range lb.sources {
		totalWeight += lb.weights[source]
	}

	if totalWeight == 0 {
		return lb.sources[0]
	}

	// 简化的加权选择逻辑
	target := int(time.Now().Unix()) % totalWeight
	current := 0

	for _, source := range lb.sources {
		current += lb.weights[source]
		if current > target {
			return source
		}
	}

	return lb.sources[0]
}

// MusicSourceHealthChecker 方法

// start 启动健康检查
func (hc *MusicSourceHealthChecker) start() {
	ticker := time.NewTicker(hc.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hc.performHealthCheck()
		case <-hc.stopChan:
			return
		}
	}
}

// performHealthCheck 执行健康检查
func (hc *MusicSourceHealthChecker) performHealthCheck() {
	hc.factory.mu.RLock()
	sources := make(map[string]MusicSourcePlugin)
	for name, source := range hc.factory.activeSources {
		sources[name] = source
	}
	hc.factory.mu.RUnlock()

	for name, source := range sources {
		status := hc.checkSourceHealth(source)

		hc.mu.Lock()
		hc.healthStatus[name] = status
		hc.lastCheck[name] = time.Now()
		hc.mu.Unlock()
	}
}

// checkSourceHealth 检查单个源的健康状态
func (hc *MusicSourceHealthChecker) checkSourceHealth(source MusicSourcePlugin) HealthStatus {
	// 简单的健康检查：尝试获取服务信息
	serviceInfo := source.GetServiceInfo()
	if serviceInfo == nil {
		return HealthStatus{
			Healthy:   false,
			Message:   "Service info unavailable",
			Timestamp: time.Now(),
		}
	}

	// 检查服务状态
	if serviceInfo.Status == core.ServiceStatusRunning {
		return HealthStatus{
			Healthy:   true,
			Message:   "Service running",
			Timestamp: time.Now(),
		}
	}

	return HealthStatus{
		Healthy:   false,
		Message:   "Service not running",
		Timestamp: time.Now(),
	}
}

// stop 停止健康检查
func (hc *MusicSourceHealthChecker) stop() {
	close(hc.stopChan)
}

// GetFactoryStats 获取工厂统计信息
func (f *MusicSourceFactory) GetFactoryStats() map[string]interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()

	stats := map[string]interface{}{
		"registered_sources": len(f.registeredSources),
		"active_sources":     len(f.activeSources),
		"load_balancing":     f.config.LoadBalancing,
		"health_check":       f.config.HealthCheck,
		"failover_enabled":   f.config.FailoverEnabled,
	}

	// 添加健康状态统计
	if f.healthChecker != nil {
		f.healthChecker.mu.RLock()
		healthyCount := 0
		for _, status := range f.healthChecker.healthStatus {
			if status.Healthy {
				healthyCount++
			}
		}
		stats["healthy_sources"] = healthyCount
		stats["health_status"] = f.healthChecker.healthStatus
		f.healthChecker.mu.RUnlock()
	}

	return stats
}

// Close 关闭工厂
func (f *MusicSourceFactory) Close() error {
	// 停止健康检查
	if f.healthChecker != nil {
		f.healthChecker.stop()
	}

	// 关闭所有活跃源
	f.mu.Lock()
	for name, source := range f.activeSources {
		if err := source.Stop(); err != nil {
			fmt.Printf("Error stopping source %s: %v\n", name, err)
		}
	}
	f.activeSources = make(map[string]MusicSourcePlugin)
	f.mu.Unlock()

	return nil
}