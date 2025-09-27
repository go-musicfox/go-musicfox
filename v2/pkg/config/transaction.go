package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	koanfjson "github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// configTransaction 配置事务实现
type configTransaction struct {
	id          string
	manager     *AdvancedManager
	originalData map[string]interface{}
	changes     map[string]interface{}
	deletions   map[string]bool
	active      bool
	mutex       sync.RWMutex
	createdAt   time.Time
}

// BeginTransaction 开始配置事务
func (am *AdvancedManager) BeginTransaction() (ConfigTransaction, error) {
	am.transactionMutex.Lock()
	defer am.transactionMutex.Unlock()

	transactionID := generateID()

	// 创建事务快照
	originalData := make(map[string]interface{})
	for key, value := range am.k.All() {
		originalData[key] = value
	}

	transaction := &configTransaction{
		id:           transactionID,
		manager:      am,
		originalData: originalData,
		changes:      make(map[string]interface{}),
		deletions:    make(map[string]bool),
		active:       true,
		createdAt:    time.Now(),
	}

	am.activeTransactions[transactionID] = transaction

	return transaction, nil
}

// Set 在事务中设置配置值
func (ct *configTransaction) Set(key string, value interface{}) error {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	if !ct.active {
		return fmt.Errorf("transaction is not active")
	}

	ct.changes[key] = value
	delete(ct.deletions, key) // 如果之前标记为删除，现在取消删除

	return nil
}

// Delete 在事务中删除配置键
func (ct *configTransaction) Delete(key string) error {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	if !ct.active {
		return fmt.Errorf("transaction is not active")
	}

	ct.deletions[key] = true
	delete(ct.changes, key) // 如果之前有变更，现在删除变更记录

	return nil
}

// Commit 提交事务
func (ct *configTransaction) Commit() error {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	if !ct.active {
		return fmt.Errorf("transaction is not active")
	}

	// 应用所有变更
	for key, value := range ct.changes {
		oldValue := ct.manager.k.Get(key)
		ct.manager.k.Set(key, value)
		ct.manager.recordChange("transaction_set", key, oldValue, value, "system", fmt.Sprintf("transaction_%s", ct.id))
	}

	// 应用所有删除
	for key := range ct.deletions {
		oldValue := ct.manager.k.Get(key)
		ct.manager.k.Delete(key)
		ct.manager.recordChange("transaction_delete", key, oldValue, nil, "system", fmt.Sprintf("transaction_%s", ct.id))
	}

	// 标记事务为非活跃
	ct.active = false

	// 从活跃事务列表中移除
	ct.manager.transactionMutex.Lock()
	delete(ct.manager.activeTransactions, ct.id)
	ct.manager.transactionMutex.Unlock()

	// 触发配置变更回调
	for key, value := range ct.changes {
		change := &ConfigChange{
			ID:        generateID(),
			Timestamp: time.Now(),
			Operation: "transaction_commit",
			Key:       key,
			OldValue:  ct.originalData[key],
			NewValue:  value,
			User:      "system",
			Source:    fmt.Sprintf("transaction_%s", ct.id),
		}
		ct.manager.triggerCallbacks(change)
	}

	for key := range ct.deletions {
		change := &ConfigChange{
			ID:        generateID(),
			Timestamp: time.Now(),
			Operation: "transaction_commit",
			Key:       key,
			OldValue:  ct.originalData[key],
			NewValue:  nil,
			User:      "system",
			Source:    fmt.Sprintf("transaction_%s", ct.id),
		}
		ct.manager.triggerCallbacks(change)
	}

	return nil
}

// Rollback 回滚事务
func (ct *configTransaction) Rollback() error {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	if !ct.active {
		return fmt.Errorf("transaction is not active")
	}

	// 标记事务为非活跃
	ct.active = false

	// 从活跃事务列表中移除
	ct.manager.transactionMutex.Lock()
	delete(ct.manager.activeTransactions, ct.id)
	ct.manager.transactionMutex.Unlock()

	// 记录回滚操作
	ct.manager.recordChange("transaction_rollback", "*", nil, nil, "system", fmt.Sprintf("transaction_%s", ct.id))

	return nil
}

// IsActive 检查事务是否活跃
func (ct *configTransaction) IsActive() bool {
	ct.mutex.RLock()
	defer ct.mutex.RUnlock()
	return ct.active
}

// BatchUpdate 批量更新配置
func (am *AdvancedManager) BatchUpdate(updates map[string]interface{}) error {
	// 使用事务确保原子性
	tx, err := am.BeginTransaction()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// 应用所有更新
	for key, value := range updates {
		if err := tx.Set(key, value); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to set key %s: %w", key, err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit batch update: %w", err)
	}

	return nil
}

// ImportConfig 导入配置
func (am *AdvancedManager) ImportConfig(source string, format string) error {
	// 检查源文件是否存在
	if _, err := os.Stat(source); os.IsNotExist(err) {
		return fmt.Errorf("source file not found: %s", source)
	}

	// 创建临时koanf实例加载源配置
	tempKoanf := koanf.New(".")

	// 根据格式选择解析器
	var parser koanf.Parser
	switch strings.ToLower(format) {
	case "json":
		parser = koanfjson.Parser()
	case "yaml", "yml":
		parser = yaml.Parser()
	case "toml":
		parser = toml.Parser()
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	// 加载源配置
	if err := tempKoanf.Load(file.Provider(source), parser); err != nil {
		return fmt.Errorf("failed to load source config: %w", err)
	}

	// 使用批量更新导入配置
	sourceData := tempKoanf.All()
	if err := am.BatchUpdate(sourceData); err != nil {
		return fmt.Errorf("failed to import config: %w", err)
	}

	// 记录导入操作
	am.recordChange("import", "*", nil, sourceData, "system", fmt.Sprintf("import_%s", source))

	return nil
}

// ExportConfig 导出配置
func (am *AdvancedManager) ExportConfig(target string, format string) error {
	// 确保目标目录存在
	dir := filepath.Dir(target)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// 获取当前配置数据
	configData := am.k.All()

	// 根据格式序列化数据
	var data []byte
	var err error

	switch strings.ToLower(format) {
	case "json":
		data, err = json.MarshalIndent(configData, "", "  ")
	case "yaml", "yml":
		// 使用简单的YAML序列化（可以根据需要使用更复杂的库）
		data, err = json.MarshalIndent(configData, "", "  ")
		if err == nil {
			// 这里应该使用YAML库，为了简化使用JSON格式
			data = []byte("# YAML export not fully implemented, using JSON format\n" + string(data))
		}
	case "toml":
		// 使用简单的TOML序列化（可以根据需要使用更复杂的库）
		data, err = json.MarshalIndent(configData, "", "  ")
		if err == nil {
			// 这里应该使用TOML库，为了简化使用JSON格式
			data = []byte("# TOML export not fully implemented, using JSON format\n" + string(data))
		}
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to serialize config data: %w", err)
	}

	// 写入目标文件
	if err := os.WriteFile(target, data, 0644); err != nil {
		return fmt.Errorf("failed to write target file: %w", err)
	}

	// 记录导出操作
	am.recordChange("export", "*", nil, target, "system", fmt.Sprintf("export_%s", target))

	return nil
}

// GetStats 获取配置统计信息
func (am *AdvancedManager) GetStats() *ConfigStats {
	am.statsMutex.RLock()
	defer am.statsMutex.RUnlock()

	// 更新统计信息
	am.stats.TotalKeys = len(am.k.Keys())

	encryptedCount := 0
	am.encryptionMutex.RLock()
	for _, encrypted := range am.encryptedKeys {
		if encrypted {
			encryptedCount++
		}
	}
	am.encryptionMutex.RUnlock()
	am.stats.EncryptedKeys = encryptedCount

	am.versionMutex.RLock()
	am.stats.VersionCount = len(am.versionHistory)
	am.versionMutex.RUnlock()

	// 计算内存使用量（简化计算）
	configData := am.k.All()
	if jsonData, err := json.Marshal(configData); err == nil {
		am.stats.MemoryUsage = int64(len(jsonData))
	}

	return am.stats
}

// GetChangeLog 获取配置变更日志
func (am *AdvancedManager) GetChangeLog() ([]*ConfigChange, error) {
	am.statsMutex.RLock()
	defer am.statsMutex.RUnlock()

	// 返回变更日志的副本
	changeLog := make([]*ConfigChange, len(am.changeLog))
	copy(changeLog, am.changeLog)

	return changeLog, nil
}

// ClearChangeLog 清空配置变更日志
func (am *AdvancedManager) ClearChangeLog() error {
	am.statsMutex.Lock()
	defer am.statsMutex.Unlock()

	am.changeLog = make([]*ConfigChange, 0)
	am.stats.ChangeCount = 0

	return nil
}

// GetActiveTransactions 获取活跃事务列表
func (am *AdvancedManager) GetActiveTransactions() []string {
	am.transactionMutex.RLock()
	defer am.transactionMutex.RUnlock()

	transactionIDs := make([]string, 0, len(am.activeTransactions))
	for id := range am.activeTransactions {
		transactionIDs = append(transactionIDs, id)
	}

	return transactionIDs
}

// AbortTransaction 中止指定事务
func (am *AdvancedManager) AbortTransaction(transactionID string) error {
	am.transactionMutex.RLock()
	transaction, exists := am.activeTransactions[transactionID]
	am.transactionMutex.RUnlock()

	if !exists {
		return fmt.Errorf("transaction %s not found", transactionID)
	}

	return transaction.Rollback()
}

// CleanupExpiredTransactions 清理过期事务
func (am *AdvancedManager) CleanupExpiredTransactions(maxAge time.Duration) error {
	am.transactionMutex.Lock()
	defer am.transactionMutex.Unlock()

	now := time.Now()
	expiredTransactions := make([]string, 0)

	for id, transaction := range am.activeTransactions {
		if now.Sub(transaction.createdAt) > maxAge {
			expiredTransactions = append(expiredTransactions, id)
		}
	}

	// 回滚过期事务
	for _, id := range expiredTransactions {
		if transaction, exists := am.activeTransactions[id]; exists {
			transaction.Rollback()
			fmt.Printf("Rolled back expired transaction: %s\n", id)
		}
	}

	return nil
}

// StartTransactionCleanup 启动事务清理协程
func (am *AdvancedManager) StartTransactionCleanup(ctx context.Context, interval time.Duration, maxAge time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := am.CleanupExpiredTransactions(maxAge); err != nil {
					fmt.Printf("Failed to cleanup expired transactions: %v\n", err)
				}
			}
		}
	}()
}