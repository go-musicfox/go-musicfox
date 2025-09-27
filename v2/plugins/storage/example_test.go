package storage

import (
	"context"
	"fmt"
	"log"
	"time"
)

// ExampleStoragePlugin 展示如何使用存储插件
func ExampleStoragePlugin() {
	// 创建存储配置
	config := &StorageConfig{
		Backend:      "memory",
		CacheEnabled: true,
		CacheMaxSize: 100,
		CacheTTL:     5 * time.Minute,
	}

	// 创建存储插件
	plugin, err := NewPlugin(config)
	if err != nil {
		log.Fatalf("Failed to create plugin: %v", err)
	}

	// 初始化和启动插件
	if err := plugin.Initialize(&simplePluginContext{}); err != nil {
		log.Fatalf("Failed to initialize plugin: %v", err)
	}

	if err := plugin.Start(); err != nil {
		log.Fatalf("Failed to start plugin: %v", err)
	}

	// 基础操作
	err = plugin.Set("user:123", "Alice", 0)
	if err != nil {
		log.Printf("Failed to set value: %v", err)
		return
	}

	value, err := plugin.Get("user:123")
	if err != nil {
		log.Printf("Failed to get value: %v", err)
		return
	}
	fmt.Printf("Retrieved value: %v\n", value)

	// 批量操作
	items := map[string]interface{}{
		"user:124": "Bob",
		"user:125": "Charlie",
		"config:timeout": 30,
	}
	err = plugin.SetBatch(items, 0)
	if err != nil {
		log.Printf("Failed to set batch: %v", err)
		return
	}

	// 查询操作
	result, err := plugin.Find("user:*", 0)
	if err != nil {
		log.Printf("Failed to find: %v", err)
		return
	}
	fmt.Printf("Found %d users\n", len(result))

	// 事务操作
	tx, err := plugin.BeginTransaction()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return
	}

	err = tx.Set("temp:key", "temp_value")
	if err != nil {
		log.Printf("Failed to set in transaction: %v", err)
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return
	}

	// 缓存统计
	stats := plugin.GetCacheStats()
	fmt.Printf("Cache hit rate: %.2f%%\n", stats.HitRate()*100)

	// 清理
	if err := plugin.Stop(); err != nil {
		log.Printf("Failed to stop plugin: %v", err)
	}

	if err := plugin.Cleanup(); err != nil {
		log.Printf("Failed to cleanup plugin: %v", err)
	}

	fmt.Println("Storage plugin example completed successfully")

	// Output:
	// Retrieved value: Alice
	// Found 3 users
	// Cache hit rate: 100.00%
	// Storage plugin example completed successfully
}

// ExampleStorageAPI 展示如何使用存储API
func ExampleStorageAPI() {
	// 初始化存储插件
	config := DefaultStorageConfig()
	ctx := context.Background()

	err := InitializeStoragePlugin(ctx, config)
	if err != nil {
		log.Fatalf("Failed to initialize storage plugin: %v", err)
	}

	// 获取存储API
	api, err := GetStorageAPI()
	if err != nil {
		log.Fatalf("Failed to get storage API: %v", err)
	}

	// 使用API进行操作
	err = api.Set("example:key", "example_value")
	if err != nil {
		log.Printf("Failed to set value: %v", err)
		return
	}

	value, err := api.Get("example:key")
	if err != nil {
		log.Printf("Failed to get value: %v", err)
		return
	}

	fmt.Printf("API retrieved value: %v\n", value)

	// 清理
	err = ShutdownStoragePlugin(ctx)
	if err != nil {
		log.Printf("Failed to shutdown storage plugin: %v", err)
	}

	fmt.Println("Storage API example completed successfully")

	// Output:
	// API retrieved value: example_value
	// Storage API example completed successfully
}

// ExampleFileBackend 展示如何使用文件后端
func ExampleFileBackend() {
	// 创建文件后端配置
	fileConfig := &FileBackendConfig{
		DataDir:       "/tmp/storage_example",
		SyncMode:      false,
		Compression:   false,
		BackupCount:   3,
		FlushInterval: 5 * time.Second,
	}

	// 创建存储配置
	config := &StorageConfig{
		Backend:      "file",
		CacheEnabled: false,
		FileConfig:   fileConfig,
	}

	// 创建插件
	plugin, err := NewPlugin(config)
	if err != nil {
		log.Fatalf("Failed to create plugin: %v", err)
	}

	// 初始化和启动
	if err := plugin.Initialize(&simplePluginContext{}); err != nil {
		log.Fatalf("Failed to initialize plugin: %v", err)
	}

	if err := plugin.Start(); err != nil {
		log.Fatalf("Failed to start plugin: %v", err)
	}

	// 存储数据
	err = plugin.Set("persistent:key", "persistent_value", 0)
	if err != nil {
		log.Printf("Failed to set persistent value: %v", err)
		return
	}

	// 读取数据
	value, err := plugin.Get("persistent:key")
	if err != nil {
		log.Printf("Failed to get persistent value: %v", err)
		return
	}

	fmt.Printf("File backend retrieved value: %v\n", value)

	// 清理
	if err := plugin.Stop(); err != nil {
		log.Printf("Failed to stop plugin: %v", err)
	}

	if err := plugin.Cleanup(); err != nil {
		log.Printf("Failed to cleanup plugin: %v", err)
	}

	fmt.Println("File backend example completed successfully")

	// Output:
	// File backend retrieved value: persistent_value
	// File backend example completed successfully
}