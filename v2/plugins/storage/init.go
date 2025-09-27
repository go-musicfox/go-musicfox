package storage

import (
	"context"
	"fmt"
	"time"

	core "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
)

// PluginName 插件名称
const PluginName = "storage"

// PluginVersion 插件版本
const PluginVersion = "1.0.0"

// pluginInstance 插件实例
var pluginInstance *Plugin

// simplePluginContext 简单的插件上下文实现
type simplePluginContext struct{}

func (s *simplePluginContext) GetContext() context.Context {
	return context.Background()
}

func (s *simplePluginContext) GetContainer() core.ServiceRegistry {
	return nil
}

func (s *simplePluginContext) GetEventBus() core.EventBus {
	return nil
}

func (s *simplePluginContext) GetServiceRegistry() core.ServiceRegistry {
	return nil
}

func (s *simplePluginContext) GetLogger() core.Logger {
	return nil
}

func (s *simplePluginContext) GetPluginConfig() core.PluginConfig {
	return nil
}

func (s *simplePluginContext) UpdateConfig(config core.PluginConfig) error {
	return nil
}

func (s *simplePluginContext) GetDataDir() string {
	return "/tmp"
}

func (s *simplePluginContext) GetTempDir() string {
	return "/tmp"
}

func (s *simplePluginContext) SendMessage(topic string, data interface{}) error {
	return nil
}

func (s *simplePluginContext) Subscribe(topic string, handler core.EventHandler) error {
	return nil
}

func (s *simplePluginContext) Unsubscribe(topic string, handler core.EventHandler) error {
	return nil
}

func (s *simplePluginContext) BroadcastMessage(message interface{}) error {
	return nil
}

func (s *simplePluginContext) GetResourceMonitor() *core.ResourceMonitor {
	return nil
}

func (s *simplePluginContext) GetSecurityManager() *core.SecurityManager {
	return nil
}

func (s *simplePluginContext) GetIsolationGroup() *core.IsolationGroup {
	return nil
}

func (s *simplePluginContext) Shutdown() error {
	return nil
}

// init 自动注册插件
func init() {
	// 注册插件工厂函数
	// core.RegisterPlugin(PluginName, NewStoragePlugin)
}

// NewStoragePlugin 创建存储插件实例
func NewStoragePlugin() (core.Plugin, error) {
	if pluginInstance != nil {
		return pluginInstance, nil
	}

	// 使用默认配置创建插件
	config := DefaultStorageConfig()
	
	// 可以从环境变量或配置文件加载配置
	if err := loadConfigFromEnv(config); err != nil {
		return nil, fmt.Errorf("failed to load config from environment: %w", err)
	}

	plugin, err := NewPlugin(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage plugin: %w", err)
	}

	pluginInstance = plugin
	return plugin, nil
}

// GetStoragePlugin 获取存储插件实例
func GetStoragePlugin() (*Plugin, error) {
	if pluginInstance == nil {
		return nil, fmt.Errorf("storage plugin not initialized")
	}
	return pluginInstance, nil
}

// InitializeStoragePlugin 初始化存储插件
func InitializeStoragePlugin(ctx context.Context, config *StorageConfig) error {
	if pluginInstance != nil {
		return fmt.Errorf("storage plugin already initialized")
	}

	if config == nil {
		config = DefaultStorageConfig()
	}

	plugin, err := NewPlugin(config)
	if err != nil {
		return fmt.Errorf("failed to create storage plugin: %w", err)
	}

	// 创建一个简单的上下文用于初始化
	pluginCtx := &simplePluginContext{}
	if err := plugin.Initialize(pluginCtx); err != nil {
		return fmt.Errorf("failed to initialize storage plugin: %w", err)
	}

	if err := plugin.Start(); err != nil {
		return fmt.Errorf("failed to start storage plugin: %w", err)
	}

	pluginInstance = plugin
	return nil
}

// ShutdownStoragePlugin 关闭存储插件
func ShutdownStoragePlugin(ctx context.Context) error {
	if pluginInstance == nil {
		return nil
	}

	if err := pluginInstance.Stop(); err != nil {
		return fmt.Errorf("failed to stop storage plugin: %w", err)
	}

	if err := pluginInstance.Cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup storage plugin: %w", err)
	}

	pluginInstance = nil
	return nil
}

// loadConfigFromEnv 从环境变量加载配置
func loadConfigFromEnv(config *StorageConfig) error {
	// 这里可以实现从环境变量加载配置的逻辑
	// 例如：
	// if backend := os.Getenv("STORAGE_BACKEND"); backend != "" {
	//     config.Backend = backend
	// }
	// if cacheEnabled := os.Getenv("STORAGE_CACHE_ENABLED"); cacheEnabled != "" {
	//     config.CacheEnabled = cacheEnabled == "true"
	// }
	
	// 目前返回nil，表示使用默认配置
	return nil
}

// StorageAPI 存储API接口，提供给其他插件使用
type StorageAPI interface {
	// 基础存储操作
	Get(key string) (interface{}, error)
	Set(key string, value interface{}) error
	SetWithTTL(key string, value interface{}, ttl time.Duration) error
	Delete(key string) error
	Exists(key string) (bool, error)

	// 批量操作
	GetBatch(keys []string) (map[string]interface{}, error)
	SetBatch(items map[string]interface{}) error
	DeleteBatch(keys []string) error

	// 查询操作
	Find(pattern string, limit int) (map[string]interface{}, error)
	Count(pattern string) (int64, error)
	Keys(pattern string) ([]string, error)

	// 事务支持
	BeginTransaction() (Transaction, error)

	// 缓存管理
	ClearCache() error
	GetCacheStats() CacheStats
}

// storageAPIImpl 存储API实现
type storageAPIImpl struct {
	plugin *Plugin
}

// GetStorageAPI 获取存储API
func GetStorageAPI() (StorageAPI, error) {
	plugin, err := GetStoragePlugin()
	if err != nil {
		return nil, err
	}

	return &storageAPIImpl{plugin: plugin}, nil
}

// Get 获取值
func (api *storageAPIImpl) Get(key string) (interface{}, error) {
	return api.plugin.Get(key)
}

// Set 设置值
func (api *storageAPIImpl) Set(key string, value interface{}) error {
	return api.plugin.Set(key, value, 0)
}

// SetWithTTL 设置值（带TTL）
func (api *storageAPIImpl) SetWithTTL(key string, value interface{}, ttl time.Duration) error {
	return api.plugin.Set(key, value, ttl)
}

// Delete 删除值
func (api *storageAPIImpl) Delete(key string) error {
	return api.plugin.Delete(key)
}

// Exists 检查键是否存在
func (api *storageAPIImpl) Exists(key string) (bool, error) {
	return api.plugin.Exists(key)
}

// GetBatch 批量获取值
func (api *storageAPIImpl) GetBatch(keys []string) (map[string]interface{}, error) {
	return api.plugin.GetBatch(keys)
}

// SetBatch 批量设置值
func (api *storageAPIImpl) SetBatch(items map[string]interface{}) error {
	return api.plugin.SetBatch(items, 0)
}

// DeleteBatch 批量删除值
func (api *storageAPIImpl) DeleteBatch(keys []string) error {
	return api.plugin.DeleteBatch(keys)
}

// Find 查找匹配的键值对
func (api *storageAPIImpl) Find(pattern string, limit int) (map[string]interface{}, error) {
	return api.plugin.Find(pattern, limit)
}

// Count 统计匹配的键数量
func (api *storageAPIImpl) Count(pattern string) (int64, error) {
	return api.plugin.Count(pattern)
}

// Keys 获取所有匹配的键
func (api *storageAPIImpl) Keys(pattern string) ([]string, error) {
	return api.plugin.Keys(pattern)
}

// BeginTransaction 开始事务
func (api *storageAPIImpl) BeginTransaction() (Transaction, error) {
	return api.plugin.BeginTransaction()
}

// ClearCache 清空缓存
func (api *storageAPIImpl) ClearCache() error {
	return api.plugin.ClearCache()
}

// GetCacheStats 获取缓存统计
func (api *storageAPIImpl) GetCacheStats() CacheStats {
	return api.plugin.GetCacheStats()
}