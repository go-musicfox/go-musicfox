package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/knadh/koanf/v2"
)

// EnableHotReload 启用配置热更新
func (am *AdvancedManager) EnableHotReload(ctx context.Context) error {
	if am.hotReloadCtx != nil {
		return fmt.Errorf("hot reload already enabled")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	am.watcher = watcher
	am.hotReloadCtx, am.hotReloadCancel = context.WithCancel(ctx)

	// 监听配置文件
	if am.configFile != "" {
		if err := am.watcher.Add(am.configFile); err != nil {
			am.watcher.Close()
			return fmt.Errorf("failed to watch config file: %w", err)
		}
	}

	// 监听配置目录
	if am.configDir != "" {
		if err := am.watcher.Add(am.configDir); err != nil {
			// 目录监听失败不是致命错误
			fmt.Printf("Warning: failed to watch config directory: %v\n", err)
		}
	}

	// 启动监听协程
	go am.watchConfigChanges()

	am.stats.HotReloadEnabled = true
	return nil
}

// DisableHotReload 禁用配置热更新
func (am *AdvancedManager) DisableHotReload() error {
	if am.hotReloadCancel != nil {
		am.hotReloadCancel()
		am.hotReloadCancel = nil
		am.hotReloadCtx = nil
	}

	if am.watcher != nil {
		am.watcher.Close()
		am.watcher = nil
	}

	am.stats.HotReloadEnabled = false
	return nil
}

// IsHotReloadEnabled 检查热更新是否启用
func (am *AdvancedManager) IsHotReloadEnabled() bool {
	return am.hotReloadCtx != nil
}

// OnConfigChanged 注册配置变更回调
func (am *AdvancedManager) OnConfigChanged(callback ConfigChangeCallback) error {
	am.callbackMutex.Lock()
	defer am.callbackMutex.Unlock()

	callbackID := generateID()
	am.callbacks[callbackID] = callback
	return nil
}

// RemoveConfigChangeCallback 移除配置变更回调
func (am *AdvancedManager) RemoveConfigChangeCallback(callbackID string) error {
	am.callbackMutex.Lock()
	defer am.callbackMutex.Unlock()

	delete(am.callbacks, callbackID)
	return nil
}

// watchConfigChanges 监听配置文件变化
func (am *AdvancedManager) watchConfigChanges() {
	for {
		select {
		case <-am.hotReloadCtx.Done():
			return
		case event, ok := <-am.watcher.Events:
			if !ok {
				return
			}

			// 只处理写入和重命名事件
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Rename == fsnotify.Rename {
				// 延迟处理，避免频繁重载
				time.Sleep(100 * time.Millisecond)
				am.handleConfigChange(event.Name)
			}
		case err, ok := <-am.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Config watcher error: %v\n", err)
		}
	}
}

// handleConfigChange 处理配置文件变化
func (am *AdvancedManager) handleConfigChange(filename string) {
	// 重新加载配置
	oldConfig := am.k.All()
	if err := am.LoadFromFile(filename); err != nil {
		fmt.Printf("Failed to reload config: %v\n", err)
		return
	}

	newConfig := am.k.All()

	// 检测变化并触发回调
	changes := am.detectChanges(oldConfig, newConfig)
	for _, change := range changes {
		am.recordChange(change.Operation, change.Key, change.OldValue, change.NewValue, "system", "hot_reload")
		am.triggerCallbacks(change)
	}
}

// detectChanges 检测配置变化
func (am *AdvancedManager) detectChanges(oldConfig, newConfig map[string]interface{}) []*ConfigChange {
	var changes []*ConfigChange

	// 检测新增和修改
	for key, newValue := range newConfig {
		if oldValue, exists := oldConfig[key]; exists {
			if !am.isEqual(oldValue, newValue) {
				changes = append(changes, &ConfigChange{
					ID:        generateID(),
					Timestamp: time.Now(),
					Operation: "set",
					Key:       key,
					OldValue:  oldValue,
					NewValue:  newValue,
					User:      "system",
					Source:    "hot_reload",
				})
			}
		} else {
			changes = append(changes, &ConfigChange{
				ID:        generateID(),
				Timestamp: time.Now(),
				Operation: "set",
				Key:       key,
				OldValue:  nil,
				NewValue:  newValue,
				User:      "system",
				Source:    "hot_reload",
			})
		}
	}

	// 检测删除
	for key, oldValue := range oldConfig {
		if _, exists := newConfig[key]; !exists {
			changes = append(changes, &ConfigChange{
				ID:        generateID(),
				Timestamp: time.Now(),
				Operation: "delete",
				Key:       key,
				OldValue:  oldValue,
				NewValue:  nil,
				User:      "system",
				Source:    "hot_reload",
			})
		}
	}

	return changes
}

// isEqual 比较两个值是否相等
func (am *AdvancedManager) isEqual(a, b interface{}) bool {
	// 简单的深度比较，可以根据需要优化
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return string(aJSON) == string(bJSON)
}

// triggerCallbacks 触发配置变更回调
func (am *AdvancedManager) triggerCallbacks(change *ConfigChange) {
	am.callbackMutex.RLock()
	defer am.callbackMutex.RUnlock()

	for _, callback := range am.callbacks {
		go func(cb ConfigChangeCallback) {
			if err := cb(change); err != nil {
				fmt.Printf("Config change callback error: %v\n", err)
			}
		}(callback)
	}
}

// GetVersion 获取当前配置版本
func (am *AdvancedManager) GetVersion() string {
	am.versionMutex.RLock()
	defer am.versionMutex.RUnlock()
	return am.currentVersion
}

// GetVersionHistory 获取版本历史
func (am *AdvancedManager) GetVersionHistory() ([]*ConfigVersion, error) {
	am.versionMutex.RLock()
	defer am.versionMutex.RUnlock()

	// 返回副本以避免并发修改
	history := make([]*ConfigVersion, len(am.versionHistory))
	copy(history, am.versionHistory)
	return history, nil
}

// CreateSnapshot 创建配置快照
func (am *AdvancedManager) CreateSnapshot(description string) (*ConfigVersion, error) {
	currentData := am.k.All()
	versionID := generateID()

	version := &ConfigVersion{
		ID:          versionID,
		Description: description,
		Timestamp:   time.Now(),
		Data:        currentData,
		Checksum:    calculateChecksum(currentData),
		CreatedBy:   "system",
		Tags:        []string{},
	}

	// 保存版本数据到文件
	versionFile := filepath.Join(am.versionDir, fmt.Sprintf("%s.json", versionID))
	data, err := json.MarshalIndent(version, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal version data: %w", err)
	}

	if err := os.WriteFile(versionFile, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to save version file: %w", err)
	}

	// 添加到历史记录
	am.versionMutex.Lock()
	am.versionHistory = append(am.versionHistory, version)
	am.currentVersion = versionID
	am.stats.VersionCount = len(am.versionHistory)
	am.versionMutex.Unlock()

	// 保存历史记录（不需要锁，因为saveVersionHistory内部会加锁）
	if err := am.saveVersionHistoryUnsafe(); err != nil {
		fmt.Printf("Warning: failed to save version history: %v\n", err)
	}

	return version, nil
}

// RollbackToVersion 回滚到指定版本
func (am *AdvancedManager) RollbackToVersion(versionID string) error {
	// 先查找版本（不持有锁）
	am.versionMutex.RLock()
	var targetVersion *ConfigVersion
	for _, version := range am.versionHistory {
		if version.ID == versionID {
			targetVersion = version
			break
		}
	}
	am.versionMutex.RUnlock()

	if targetVersion == nil {
		return fmt.Errorf("version %s not found", versionID)
	}

	// 不创建备份版本，避免影响版本历史计数
	// 如果需要备份，用户应该在回滚前手动创建快照

	// 现在获取锁进行回滚操作
	am.versionMutex.Lock()
	defer am.versionMutex.Unlock()

	// 应用目标版本的配置
	oldData := am.k.All()
	am.k = koanf.New(".")
	for key, value := range targetVersion.Data {
		am.k.Set(key, value)
	}

	// 记录回滚操作
	am.recordChange("rollback", "*", oldData, targetVersion.Data, "system", fmt.Sprintf("rollback_to_%s", versionID))
	am.currentVersion = versionID

	// 触发配置变更回调
	change := &ConfigChange{
		ID:        generateID(),
		Timestamp: time.Now(),
		Operation: "rollback",
		Key:       "*",
		OldValue:  oldData,
		NewValue:  targetVersion.Data,
		User:      "system",
		Source:    "rollback",
	}
	am.triggerCallbacks(change)

	// 保存配置到文件
	if am.configFile != "" {
		if err := am.SaveToFile(am.configFile); err != nil {
			return fmt.Errorf("failed to save config after rollback: %w", err)
		}
	}

	return nil
}

// CompareVersions 比较两个版本的差异
func (am *AdvancedManager) CompareVersions(version1, version2 string) (*ConfigDiff, error) {
	am.versionMutex.RLock()
	defer am.versionMutex.RUnlock()

	var v1Data, v2Data map[string]interface{}

	// 获取版本1的数据
	if version1 == "current" {
		v1Data = am.k.All()
	} else {
		for _, version := range am.versionHistory {
			if version.ID == version1 {
				v1Data = version.Data
				break
			}
		}
		if v1Data == nil {
			return nil, fmt.Errorf("version %s not found", version1)
		}
	}

	// 获取版本2的数据
	if version2 == "current" {
		v2Data = am.k.All()
	} else {
		for _, version := range am.versionHistory {
			if version.ID == version2 {
				v2Data = version.Data
				break
			}
		}
		if v2Data == nil {
			return nil, fmt.Errorf("version %s not found", version2)
		}
	}

	// 计算差异
	diff := &ConfigDiff{
		Added:    make(map[string]interface{}),
		Modified: make(map[string]DiffValue),
		Deleted:  make(map[string]interface{}),
	}

	// 检测新增和修改
	for key, v2Value := range v2Data {
		if v1Value, exists := v1Data[key]; exists {
			if !am.isEqual(v1Value, v2Value) {
				diff.Modified[key] = DiffValue{
					Old: v1Value,
					New: v2Value,
				}
			}
		} else {
			diff.Added[key] = v2Value
		}
	}

	// 检测删除
	for key, v1Value := range v1Data {
		if _, exists := v2Data[key]; !exists {
			diff.Deleted[key] = v1Value
		}
	}

	return diff, nil
}

// DeleteVersion 删除指定版本
func (am *AdvancedManager) DeleteVersion(versionID string) error {
	am.versionMutex.Lock()
	defer am.versionMutex.Unlock()

	// 不能删除当前版本
	if versionID == am.currentVersion {
		return fmt.Errorf("cannot delete current version")
	}

	// 从历史记录中移除
	for i, version := range am.versionHistory {
		if version.ID == versionID {
			// 删除版本文件
			versionFile := filepath.Join(am.versionDir, fmt.Sprintf("%s.json", versionID))
			if err := os.Remove(versionFile); err != nil && !os.IsNotExist(err) {
				fmt.Printf("Warning: failed to delete version file: %v\n", err)
			}

			// 从历史记录中移除
			am.versionHistory = append(am.versionHistory[:i], am.versionHistory[i+1:]...)
			am.stats.VersionCount = len(am.versionHistory)

			// 保存历史记录（使用不加锁版本，因为我们已经持有锁）
			if err := am.saveVersionHistoryUnsafe(); err != nil {
				fmt.Printf("Warning: failed to save version history: %v\n", err)
			}

			return nil
		}
	}

	return fmt.Errorf("version %s not found", versionID)
}