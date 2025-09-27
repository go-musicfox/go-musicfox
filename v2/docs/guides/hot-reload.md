# 热重载插件开发指南

## 概述

热重载插件是 go-musicfox v2 微内核架构中的一种特殊插件类型，支持在运行时动态更新插件代码而无需重启应用程序。这种插件类型特别适用于开发阶段的快速迭代和生产环境的无缝更新。

### 核心特性

- **零停机更新**：在不中断服务的情况下更新插件
- **状态保持**：更新过程中保持插件状态
- **回滚机制**：支持快速回滚到之前版本
- **依赖管理**：智能处理插件间依赖关系
- **安全检查**：确保更新过程的安全性

### 技术原理

热重载插件基于以下技术实现：

1. **文件监控**：监控插件文件变化
2. **版本管理**：维护多个插件版本
3. **状态迁移**：在版本间迁移插件状态
4. **依赖解析**：处理插件依赖关系
5. **安全隔离**：确保更新过程不影响其他组件

## 开发环境准备

### 1. 安装依赖

```bash
# 安装 Go 开发环境
go version # 确保 Go 1.21+

# 安装文件监控工具
go install github.com/fsnotify/fsnotify@latest

# 安装热重载开发工具
go install github.com/go-musicfox/hot-reload-dev@latest
```

### 2. 项目结构

```
hot-reload-plugin/
├── plugin.json              # 插件配置
├── main.go                  # 插件主入口
├── state.go                 # 状态管理
├── migration.go             # 状态迁移
├── version.go               # 版本管理
├── hot_reload.go            # 热重载逻辑
├── build.sh                 # 构建脚本
├── test/
│   ├── integration_test.go  # 集成测试
│   └── hot_reload_test.go   # 热重载测试
└── examples/
    ├── basic/               # 基础示例
    └── advanced/            # 高级示例
```

## 插件开发

### 1. 基础插件结构

```go
// main.go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    "time"
    
    "github.com/go-musicfox/go-musicfox/v2/pkg/plugin"
    "github.com/go-musicfox/go-musicfox/v2/pkg/types"
)

// HotReloadPlugin 热重载插件实现
type HotReloadPlugin struct {
    // 基础字段
    id      string
    name    string
    version string
    
    // 状态管理
    state     *PluginState
    stateMux  sync.RWMutex
    
    // 热重载相关
    reloadChan chan ReloadSignal
    stopChan   chan struct{}
    
    // 依赖管理
    dependencies []string
    dependents   []string
    
    // 配置
    config *PluginConfig
    
    // 生命周期钩子
    beforeReload func() error
    afterReload  func() error
    
    // 日志
    logger plugin.Logger
}

// PluginState 插件状态
type PluginState struct {
    // 运行时状态
    IsRunning     bool                   `json:"is_running"`
    StartTime     time.Time              `json:"start_time"`
    LastReload    time.Time              `json:"last_reload"`
    ReloadCount   int                    `json:"reload_count"`
    
    // 业务状态
    ProcessedCount int64                 `json:"processed_count"`
    ErrorCount     int64                 `json:"error_count"`
    CustomData     map[string]interface{} `json:"custom_data"`
    
    // 版本信息
    CurrentVersion string                `json:"current_version"`
    PreviousVersion string               `json:"previous_version"`
}

// ReloadSignal 重载信号
type ReloadSignal struct {
    Type      ReloadType `json:"type"`
    NewPath   string     `json:"new_path"`
    OldPath   string     `json:"old_path"`
    Timestamp time.Time  `json:"timestamp"`
    Metadata  map[string]interface{} `json:"metadata"`
}

type ReloadType int

const (
    ReloadTypeUpdate ReloadType = iota
    ReloadTypeRollback
    ReloadTypeForce
)

// NewHotReloadPlugin 创建热重载插件实例
func NewHotReloadPlugin(config *PluginConfig) *HotReloadPlugin {
    return &HotReloadPlugin{
        id:         config.ID,
        name:       config.Name,
        version:    config.Version,
        config:     config,
        state:      &PluginState{
            CustomData: make(map[string]interface{}),
        },
        reloadChan: make(chan ReloadSignal, 10),
        stopChan:   make(chan struct{}),
        logger:     plugin.NewLogger(config.ID),
    }
}

// 实现 plugin.Plugin 接口
func (p *HotReloadPlugin) ID() string {
    return p.id
}

func (p *HotReloadPlugin) Name() string {
    return p.name
}

func (p *HotReloadPlugin) Version() string {
    return p.version
}

func (p *HotReloadPlugin) Type() plugin.Type {
    return plugin.TypeHotReload
}

func (p *HotReloadPlugin) Initialize(ctx context.Context) error {
    p.logger.Info("Initializing hot reload plugin", "id", p.id, "version", p.version)
    
    // 初始化状态
    p.stateMux.Lock()
    p.state.IsRunning = false
    p.state.StartTime = time.Now()
    p.state.CurrentVersion = p.version
    p.stateMux.Unlock()
    
    // 启动热重载监控
    go p.startReloadMonitor(ctx)
    
    return nil
}

func (p *HotReloadPlugin) Start(ctx context.Context) error {
    p.logger.Info("Starting hot reload plugin")
    
    p.stateMux.Lock()
    p.state.IsRunning = true
    p.stateMux.Unlock()
    
    // 启动主要业务逻辑
    go p.runMainLoop(ctx)
    
    return nil
}

func (p *HotReloadPlugin) Stop(ctx context.Context) error {
    p.logger.Info("Stopping hot reload plugin")
    
    // 停止热重载监控
    close(p.stopChan)
    
    p.stateMux.Lock()
    p.state.IsRunning = false
    p.stateMux.Unlock()
    
    return nil
}

func (p *HotReloadPlugin) Cleanup() error {
    p.logger.Info("Cleaning up hot reload plugin")
    
    // 清理资源
    close(p.reloadChan)
    
    return nil
}
```

### 2. 状态管理

```go
// state.go
package main

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "sync"
    "time"
)

// StateManager 状态管理器
type StateManager struct {
    plugin     *HotReloadPlugin
    statePath  string
    backupPath string
    mutex      sync.RWMutex
    
    // 状态快照
    snapshots map[string]*StateSnapshot
    maxSnapshots int
}

// StateSnapshot 状态快照
type StateSnapshot struct {
    Version   string                 `json:"version"`
    Timestamp time.Time              `json:"timestamp"`
    State     *PluginState           `json:"state"`
    Metadata  map[string]interface{} `json:"metadata"`
}

// NewStateManager 创建状态管理器
func NewStateManager(plugin *HotReloadPlugin, dataDir string) *StateManager {
    return &StateManager{
        plugin:       plugin,
        statePath:    filepath.Join(dataDir, "state.json"),
        backupPath:   filepath.Join(dataDir, "state_backup.json"),
        snapshots:    make(map[string]*StateSnapshot),
        maxSnapshots: 10,
    }
}

// SaveState 保存状态
func (sm *StateManager) SaveState() error {
    sm.mutex.Lock()
    defer sm.mutex.Unlock()
    
    // 读取当前状态
    sm.plugin.stateMux.RLock()
    stateData, err := json.MarshalIndent(sm.plugin.state, "", "  ")
    sm.plugin.stateMux.RUnlock()
    
    if err != nil {
        return fmt.Errorf("failed to marshal state: %w", err)
    }
    
    // 备份当前状态文件
    if _, err := os.Stat(sm.statePath); err == nil {
        if err := os.Rename(sm.statePath, sm.backupPath); err != nil {
            sm.plugin.logger.Warn("Failed to backup state file", "error", err)
        }
    }
    
    // 写入新状态
    if err := os.WriteFile(sm.statePath, stateData, 0644); err != nil {
        // 恢复备份
        if _, backupErr := os.Stat(sm.backupPath); backupErr == nil {
            os.Rename(sm.backupPath, sm.statePath)
        }
        return fmt.Errorf("failed to write state file: %w", err)
    }
    
    return nil
}

// LoadState 加载状态
func (sm *StateManager) LoadState() error {
    sm.mutex.RLock()
    defer sm.mutex.RUnlock()
    
    // 尝试加载主状态文件
    stateData, err := os.ReadFile(sm.statePath)
    if err != nil {
        // 尝试加载备份文件
        if backupData, backupErr := os.ReadFile(sm.backupPath); backupErr == nil {
            stateData = backupData
            sm.plugin.logger.Warn("Loaded state from backup file")
        } else {
            return fmt.Errorf("failed to load state: %w", err)
        }
    }
    
    // 解析状态数据
    var state PluginState
    if err := json.Unmarshal(stateData, &state); err != nil {
        return fmt.Errorf("failed to unmarshal state: %w", err)
    }
    
    // 更新插件状态
    sm.plugin.stateMux.Lock()
    sm.plugin.state = &state
    sm.plugin.stateMux.Unlock()
    
    return nil
}

// CreateSnapshot 创建状态快照
func (sm *StateManager) CreateSnapshot(version string) error {
    sm.mutex.Lock()
    defer sm.mutex.Unlock()
    
    // 创建快照
    sm.plugin.stateMux.RLock()
    stateCopy := *sm.plugin.state // 浅拷贝
    // 深拷贝 CustomData
    stateCopy.CustomData = make(map[string]interface{})
    for k, v := range sm.plugin.state.CustomData {
        stateCopy.CustomData[k] = v
    }
    sm.plugin.stateMux.RUnlock()
    
    snapshot := &StateSnapshot{
        Version:   version,
        Timestamp: time.Now(),
        State:     &stateCopy,
        Metadata: map[string]interface{}{
            "plugin_id": sm.plugin.id,
            "created_by": "hot_reload",
        },
    }
    
    // 存储快照
    sm.snapshots[version] = snapshot
    
    // 清理旧快照
    if len(sm.snapshots) > sm.maxSnapshots {
        sm.cleanupOldSnapshots()
    }
    
    sm.plugin.logger.Info("Created state snapshot", "version", version)
    return nil
}

// RestoreSnapshot 恢复状态快照
func (sm *StateManager) RestoreSnapshot(version string) error {
    sm.mutex.RLock()
    snapshot, exists := sm.snapshots[version]
    sm.mutex.RUnlock()
    
    if !exists {
        return fmt.Errorf("snapshot not found: %s", version)
    }
    
    // 恢复状态
    sm.plugin.stateMux.Lock()
    sm.plugin.state = snapshot.State
    sm.plugin.state.LastReload = time.Now()
    sm.plugin.state.ReloadCount++
    sm.plugin.stateMux.Unlock()
    
    sm.plugin.logger.Info("Restored state snapshot", "version", version)
    return nil
}

// cleanupOldSnapshots 清理旧快照
func (sm *StateManager) cleanupOldSnapshots() {
    // 按时间排序，保留最新的快照
    type snapshotInfo struct {
        version   string
        timestamp time.Time
    }
    
    var snapshots []snapshotInfo
    for version, snapshot := range sm.snapshots {
        snapshots = append(snapshots, snapshotInfo{
            version:   version,
            timestamp: snapshot.Timestamp,
        })
    }
    
    // 简单排序（实际应该使用更高效的排序算法）
    for i := 0; i < len(snapshots)-1; i++ {
        for j := i + 1; j < len(snapshots); j++ {
            if snapshots[i].timestamp.Before(snapshots[j].timestamp) {
                snapshots[i], snapshots[j] = snapshots[j], snapshots[i]
            }
        }
    }
    
    // 删除多余的快照
    for i := sm.maxSnapshots; i < len(snapshots); i++ {
        delete(sm.snapshots, snapshots[i].version)
    }
}

// GetStateInfo 获取状态信息
func (sm *StateManager) GetStateInfo() map[string]interface{} {
    sm.mutex.RLock()
    defer sm.mutex.RUnlock()
    
    sm.plugin.stateMux.RLock()
    defer sm.plugin.stateMux.RUnlock()
    
    return map[string]interface{}{
        "current_state": sm.plugin.state,
        "snapshots_count": len(sm.snapshots),
        "available_snapshots": func() []string {
            var versions []string
            for version := range sm.snapshots {
                versions = append(versions, version)
            }
            return versions
        }(),
    }
}
```

### 3. 热重载逻辑

```go
// hot_reload.go
package main

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "plugin"
    "time"
    
    "github.com/fsnotify/fsnotify"
)

// ReloadManager 重载管理器
type ReloadManager struct {
    plugin      *HotReloadPlugin
    stateManager *StateManager
    watcher     *fsnotify.Watcher
    
    // 配置
    watchPaths   []string
    debounceTime time.Duration
    
    // 状态
    isReloading  bool
    lastReload   time.Time
    
    // 版本管理
    versionManager *VersionManager
}

// NewReloadManager 创建重载管理器
func NewReloadManager(p *HotReloadPlugin, stateManager *StateManager) (*ReloadManager, error) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, fmt.Errorf("failed to create file watcher: %w", err)
    }
    
    return &ReloadManager{
        plugin:       p,
        stateManager: stateManager,
        watcher:      watcher,
        debounceTime: 500 * time.Millisecond,
        versionManager: NewVersionManager(p),
    }, nil
}

// Start 启动重载监控
func (rm *ReloadManager) Start(ctx context.Context) error {
    // 添加监控路径
    for _, path := range rm.watchPaths {
        if err := rm.watcher.Add(path); err != nil {
            rm.plugin.logger.Warn("Failed to watch path", "path", path, "error", err)
        }
    }
    
    // 启动文件监控
    go rm.watchFiles(ctx)
    
    // 启动重载处理
    go rm.handleReloads(ctx)
    
    return nil
}

// Stop 停止重载监控
func (rm *ReloadManager) Stop() error {
    return rm.watcher.Close()
}

// watchFiles 监控文件变化
func (rm *ReloadManager) watchFiles(ctx context.Context) {
    debounceTimer := time.NewTimer(0)
    if !debounceTimer.Stop() {
        <-debounceTimer.C
    }
    
    var pendingEvents []fsnotify.Event
    
    for {
        select {
        case <-ctx.Done():
            return
            
        case event, ok := <-rm.watcher.Events:
            if !ok {
                return
            }
            
            // 过滤相关事件
            if rm.shouldProcessEvent(event) {
                pendingEvents = append(pendingEvents, event)
                
                // 重置防抖定时器
                if !debounceTimer.Stop() {
                    select {
                    case <-debounceTimer.C:
                    default:
                    }
                }
                debounceTimer.Reset(rm.debounceTime)
            }
            
        case <-debounceTimer.C:
            if len(pendingEvents) > 0 {
                rm.processFileEvents(pendingEvents)
                pendingEvents = nil
            }
            
        case err, ok := <-rm.watcher.Errors:
            if !ok {
                return
            }
            rm.plugin.logger.Error("File watcher error", "error", err)
        }
    }
}

// shouldProcessEvent 判断是否应该处理事件
func (rm *ReloadManager) shouldProcessEvent(event fsnotify.Event) bool {
    // 只处理写入和创建事件
    if event.Op&fsnotify.Write == 0 && event.Op&fsnotify.Create == 0 {
        return false
    }
    
    // 检查文件扩展名
    ext := filepath.Ext(event.Name)
    allowedExts := []string{".so", ".dll", ".dylib", ".go", ".json"}
    
    for _, allowedExt := range allowedExts {
        if ext == allowedExt {
            return true
        }
    }
    
    return false
}

// processFileEvents 处理文件事件
func (rm *ReloadManager) processFileEvents(events []fsnotify.Event) {
    rm.plugin.logger.Info("Processing file events", "count", len(events))
    
    // 分析变化的文件
    changedFiles := make(map[string]fsnotify.Event)
    for _, event := range events {
        changedFiles[event.Name] = event
    }
    
    // 创建重载信号
    signal := ReloadSignal{
        Type:      ReloadTypeUpdate,
        Timestamp: time.Now(),
        Metadata: map[string]interface{}{
            "changed_files": func() []string {
                var files []string
                for file := range changedFiles {
                    files = append(files, file)
                }
                return files
            }(),
            "event_count": len(events),
        },
    }
    
    // 发送重载信号
    select {
    case rm.plugin.reloadChan <- signal:
        rm.plugin.logger.Info("Reload signal sent")
    default:
        rm.plugin.logger.Warn("Reload channel full, skipping reload")
    }
}

// handleReloads 处理重载请求
func (rm *ReloadManager) handleReloads(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
            
        case signal := <-rm.plugin.reloadChan:
            if err := rm.performReload(signal); err != nil {
                rm.plugin.logger.Error("Reload failed", "error", err)
            }
        }
    }
}

// performReload 执行重载
func (rm *ReloadManager) performReload(signal ReloadSignal) error {
    if rm.isReloading {
        return fmt.Errorf("reload already in progress")
    }
    
    rm.isReloading = true
    defer func() {
        rm.isReloading = false
        rm.lastReload = time.Now()
    }()
    
    rm.plugin.logger.Info("Starting plugin reload", "type", signal.Type)
    
    // 1. 执行预重载钩子
    if rm.plugin.beforeReload != nil {
        if err := rm.plugin.beforeReload(); err != nil {
            return fmt.Errorf("before reload hook failed: %w", err)
        }
    }
    
    // 2. 创建状态快照
    oldVersion := rm.plugin.version
    if err := rm.stateManager.CreateSnapshot(oldVersion); err != nil {
        rm.plugin.logger.Warn("Failed to create state snapshot", "error", err)
    }
    
    // 3. 保存当前状态
    if err := rm.stateManager.SaveState(); err != nil {
        return fmt.Errorf("failed to save state: %w", err)
    }
    
    // 4. 执行重载逻辑
    switch signal.Type {
    case ReloadTypeUpdate:
        err := rm.performUpdate(signal)
        if err != nil {
            // 尝试回滚
            if rollbackErr := rm.performRollback(oldVersion); rollbackErr != nil {
                rm.plugin.logger.Error("Rollback failed", "error", rollbackErr)
            }
            return err
        }
        
    case ReloadTypeRollback:
        if err := rm.performRollback(signal.Metadata["target_version"].(string)); err != nil {
            return err
        }
        
    case ReloadTypeForce:
        if err := rm.performForceReload(signal); err != nil {
            return err
        }
    }
    
    // 5. 执行后重载钩子
    if rm.plugin.afterReload != nil {
        if err := rm.plugin.afterReload(); err != nil {
            rm.plugin.logger.Warn("After reload hook failed", "error", err)
        }
    }
    
    // 6. 更新状态
    rm.plugin.stateMux.Lock()
    rm.plugin.state.LastReload = time.Now()
    rm.plugin.state.ReloadCount++
    rm.plugin.state.PreviousVersion = oldVersion
    rm.plugin.stateMux.Unlock()
    
    rm.plugin.logger.Info("Plugin reload completed successfully")
    return nil
}

// performUpdate 执行更新
func (rm *ReloadManager) performUpdate(signal ReloadSignal) error {
    rm.plugin.logger.Info("Performing plugin update")
    
    // 检测新版本
    newVersion, err := rm.versionManager.DetectNewVersion()
    if err != nil {
        return fmt.Errorf("failed to detect new version: %w", err)
    }
    
    if newVersion == rm.plugin.version {
        rm.plugin.logger.Info("No version change detected")
        return nil
    }
    
    // 加载新插件
    newPluginPath := rm.versionManager.GetPluginPath(newVersion)
    newPlugin, err := rm.loadNewPlugin(newPluginPath)
    if err != nil {
        return fmt.Errorf("failed to load new plugin: %w", err)
    }
    
    // 迁移状态
    if err := rm.migrateState(rm.plugin, newPlugin); err != nil {
        return fmt.Errorf("failed to migrate state: %w", err)
    }
    
    // 替换插件实例
    rm.replacePlugin(newPlugin)
    
    rm.plugin.logger.Info("Plugin updated successfully", "old_version", rm.plugin.version, "new_version", newVersion)
    return nil
}

// performRollback 执行回滚
func (rm *ReloadManager) performRollback(targetVersion string) error {
    rm.plugin.logger.Info("Performing plugin rollback", "target_version", targetVersion)
    
    // 恢复状态快照
    if err := rm.stateManager.RestoreSnapshot(targetVersion); err != nil {
        return fmt.Errorf("failed to restore snapshot: %w", err)
    }
    
    // 加载目标版本插件
    targetPluginPath := rm.versionManager.GetPluginPath(targetVersion)
    targetPlugin, err := rm.loadNewPlugin(targetPluginPath)
    if err != nil {
        return fmt.Errorf("failed to load target plugin: %w", err)
    }
    
    // 替换插件实例
    rm.replacePlugin(targetPlugin)
    
    rm.plugin.logger.Info("Plugin rollback completed", "version", targetVersion)
    return nil
}

// performForceReload 执行强制重载
func (rm *ReloadManager) performForceReload(signal ReloadSignal) error {
    rm.plugin.logger.Info("Performing force reload")
    
    // 强制重载当前版本
    currentPluginPath := rm.versionManager.GetPluginPath(rm.plugin.version)
    newPlugin, err := rm.loadNewPlugin(currentPluginPath)
    if err != nil {
        return fmt.Errorf("failed to force reload plugin: %w", err)
    }
    
    // 保持当前状态
    newPlugin.state = rm.plugin.state
    
    // 替换插件实例
    rm.replacePlugin(newPlugin)
    
    rm.plugin.logger.Info("Force reload completed")
    return nil
}

// loadNewPlugin 加载新插件
func (rm *ReloadManager) loadNewPlugin(pluginPath string) (*HotReloadPlugin, error) {
    // 这里应该实现实际的插件加载逻辑
    // 可能涉及动态库加载、Go plugin 包等
    
    // 示例实现（简化）
    p, err := plugin.Open(pluginPath)
    if err != nil {
        return nil, fmt.Errorf("failed to open plugin: %w", err)
    }
    
    newPluginSymbol, err := p.Lookup("NewPlugin")
    if err != nil {
        return nil, fmt.Errorf("failed to lookup NewPlugin symbol: %w", err)
    }
    
    newPluginFunc, ok := newPluginSymbol.(func(*PluginConfig) *HotReloadPlugin)
    if !ok {
        return nil, fmt.Errorf("invalid NewPlugin function signature")
    }
    
    newPlugin := newPluginFunc(rm.plugin.config)
    return newPlugin, nil
}

// migrateState 迁移状态
func (rm *ReloadManager) migrateState(oldPlugin, newPlugin *HotReloadPlugin) error {
    // 实现状态迁移逻辑
    // 这里可能需要版本兼容性检查和数据转换
    
    oldPlugin.stateMux.RLock()
    oldState := *oldPlugin.state
    oldPlugin.stateMux.RUnlock()
    
    newPlugin.stateMux.Lock()
    newPlugin.state = &oldState
    newPlugin.stateMux.Unlock()
    
    return nil
}

// replacePlugin 替换插件实例
func (rm *ReloadManager) replacePlugin(newPlugin *HotReloadPlugin) {
    // 这里需要在更高层次的插件管理器中实现
    // 替换插件实例的逻辑
    
    // 更新引用
    oldPlugin := rm.plugin
    rm.plugin = newPlugin
    
    // 清理旧插件
    if err := oldPlugin.Cleanup(); err != nil {
        rm.plugin.logger.Warn("Failed to cleanup old plugin", "error", err)
    }
}
```

### 4. 版本管理

```go
// version.go
package main

import (
    "fmt"
    "os"
    "path/filepath"
    "regexp"
    "sort"
    "strconv"
    "strings"
    "time"
)

// VersionManager 版本管理器
type VersionManager struct {
    plugin     *HotReloadPlugin
    pluginDir  string
    versionPattern *regexp.Regexp
    
    // 版本历史
    versions   []VersionInfo
    maxVersions int
}

// VersionInfo 版本信息
type VersionInfo struct {
    Version   string    `json:"version"`
    Path      string    `json:"path"`
    Timestamp time.Time `json:"timestamp"`
    Size      int64     `json:"size"`
    Checksum  string    `json:"checksum"`
}

// NewVersionManager 创建版本管理器
func NewVersionManager(plugin *HotReloadPlugin) *VersionManager {
    return &VersionManager{
        plugin:         plugin,
        pluginDir:      filepath.Dir(plugin.config.Path),
        versionPattern: regexp.MustCompile(`v(\d+)\.(\d+)\.(\d+)`),
        maxVersions:    20,
    }
}

// DetectNewVersion 检测新版本
func (vm *VersionManager) DetectNewVersion() (string, error) {
    // 扫描插件目录
    files, err := os.ReadDir(vm.pluginDir)
    if err != nil {
        return "", fmt.Errorf("failed to read plugin directory: %w", err)
    }
    
    var latestVersion string
    var latestTime time.Time
    
    for _, file := range files {
        if file.IsDir() {
            continue
        }
        
        // 检查文件名是否匹配版本模式
        matches := vm.versionPattern.FindStringSubmatch(file.Name())
        if len(matches) != 4 {
            continue
        }
        
        version := matches[0]
        
        // 获取文件信息
        filePath := filepath.Join(vm.pluginDir, file.Name())
        fileInfo, err := os.Stat(filePath)
        if err != nil {
            continue
        }
        
        // 比较修改时间
        if fileInfo.ModTime().After(latestTime) {
            latestTime = fileInfo.ModTime()
            latestVersion = version
        }
    }
    
    if latestVersion == "" {
        return vm.plugin.version, nil
    }
    
    return latestVersion, nil
}

// GetPluginPath 获取插件路径
func (vm *VersionManager) GetPluginPath(version string) string {
    // 构建版本特定的插件路径
    filename := fmt.Sprintf("%s_%s.so", vm.plugin.id, version)
    return filepath.Join(vm.pluginDir, filename)
}

// CompareVersions 比较版本
func (vm *VersionManager) CompareVersions(v1, v2 string) int {
    // 解析版本号
    parts1 := vm.parseVersion(v1)
    parts2 := vm.parseVersion(v2)
    
    // 比较各个部分
    for i := 0; i < len(parts1) && i < len(parts2); i++ {
        if parts1[i] < parts2[i] {
            return -1
        } else if parts1[i] > parts2[i] {
            return 1
        }
    }
    
    // 长度不同的情况
    if len(parts1) < len(parts2) {
        return -1
    } else if len(parts1) > len(parts2) {
        return 1
    }
    
    return 0
}

// parseVersion 解析版本号
func (vm *VersionManager) parseVersion(version string) []int {
    // 移除 'v' 前缀
    version = strings.TrimPrefix(version, "v")
    
    parts := strings.Split(version, ".")
    var result []int
    
    for _, part := range parts {
        if num, err := strconv.Atoi(part); err == nil {
            result = append(result, num)
        } else {
            result = append(result, 0)
        }
    }
    
    return result
}

// ListVersions 列出所有版本
func (vm *VersionManager) ListVersions() ([]VersionInfo, error) {
    files, err := os.ReadDir(vm.pluginDir)
    if err != nil {
        return nil, fmt.Errorf("failed to read plugin directory: %w", err)
    }
    
    var versions []VersionInfo
    
    for _, file := range files {
        if file.IsDir() {
            continue
        }
        
        matches := vm.versionPattern.FindStringSubmatch(file.Name())
        if len(matches) != 4 {
            continue
        }
        
        version := matches[0]
        filePath := filepath.Join(vm.pluginDir, file.Name())
        
        fileInfo, err := os.Stat(filePath)
        if err != nil {
            continue
        }
        
        versionInfo := VersionInfo{
            Version:   version,
            Path:      filePath,
            Timestamp: fileInfo.ModTime(),
            Size:      fileInfo.Size(),
        }
        
        versions = append(versions, versionInfo)
    }
    
    // 按版本排序
    sort.Slice(versions, func(i, j int) bool {
        return vm.CompareVersions(versions[i].Version, versions[j].Version) > 0
    })
    
    return versions, nil
}

// CleanupOldVersions 清理旧版本
func (vm *VersionManager) CleanupOldVersions() error {
    versions, err := vm.ListVersions()
    if err != nil {
        return err
    }
    
    if len(versions) <= vm.maxVersions {
        return nil
    }
    
    // 删除多余的旧版本
    for i := vm.maxVersions; i < len(versions); i++ {
        if err := os.Remove(versions[i].Path); err != nil {
            vm.plugin.logger.Warn("Failed to remove old version", "version", versions[i].Version, "error", err)
        } else {
            vm.plugin.logger.Info("Removed old version", "version", versions[i].Version)
        }
    }
    
    return nil
}
```

### 5. 状态迁移

```go
// migration.go
package main

import (
    "encoding/json"
    "fmt"
    "reflect"
)

// StateMigrator 状态迁移器
type StateMigrator struct {
    migrations map[string]MigrationFunc
}

// MigrationFunc 迁移函数类型
type MigrationFunc func(oldState, newState *PluginState) error

// NewStateMigrator 创建状态迁移器
func NewStateMigrator() *StateMigrator {
    migrator := &StateMigrator{
        migrations: make(map[string]MigrationFunc),
    }
    
    // 注册默认迁移函数
    migrator.RegisterMigration("v1.0.0->v1.1.0", migrator.migrateV1_0_0ToV1_1_0)
    migrator.RegisterMigration("v1.1.0->v1.2.0", migrator.migrateV1_1_0ToV1_2_0)
    
    return migrator
}

// RegisterMigration 注册迁移函数
func (sm *StateMigrator) RegisterMigration(versionPair string, migrationFunc MigrationFunc) {
    sm.migrations[versionPair] = migrationFunc
}

// Migrate 执行状态迁移
func (sm *StateMigrator) Migrate(oldVersion, newVersion string, oldState, newState *PluginState) error {
    migrationKey := fmt.Sprintf("%s->%s", oldVersion, newVersion)
    
    if migrationFunc, exists := sm.migrations[migrationKey]; exists {
        return migrationFunc(oldState, newState)
    }
    
    // 如果没有特定的迁移函数，尝试通用迁移
    return sm.genericMigration(oldState, newState)
}

// genericMigration 通用迁移
func (sm *StateMigrator) genericMigration(oldState, newState *PluginState) error {
    // 使用反射进行字段匹配和复制
    oldValue := reflect.ValueOf(oldState).Elem()
    newValue := reflect.ValueOf(newState).Elem()
    
    oldType := oldValue.Type()
    newType := newValue.Type()
    
    // 遍历新状态的字段
    for i := 0; i < newType.NumField(); i++ {
        newField := newType.Field(i)
        newFieldValue := newValue.Field(i)
        
        // 查找旧状态中的对应字段
        if oldField, found := oldType.FieldByName(newField.Name); found {
            oldFieldValue := oldValue.FieldByName(newField.Name)
            
            // 类型兼容性检查
            if oldField.Type.AssignableTo(newField.Type) && newFieldValue.CanSet() {
                newFieldValue.Set(oldFieldValue)
            }
        }
    }
    
    return nil
}

// 具体的迁移函数示例
func (sm *StateMigrator) migrateV1_0_0ToV1_1_0(oldState, newState *PluginState) error {
    // 复制基础字段
    newState.IsRunning = oldState.IsRunning
    newState.StartTime = oldState.StartTime
    newState.ProcessedCount = oldState.ProcessedCount
    newState.ErrorCount = oldState.ErrorCount
    
    // v1.1.0 新增了 ReloadCount 字段
    newState.ReloadCount = 0
    
    // 迁移自定义数据
    if oldState.CustomData != nil {
        newState.CustomData = make(map[string]interface{})
        for k, v := range oldState.CustomData {
            newState.CustomData[k] = v
        }
    }
    
    return nil
}

func (sm *StateMigrator) migrateV1_1_0ToV1_2_0(oldState, newState *PluginState) error {
    // 复制所有字段
    *newState = *oldState
    
    // v1.2.0 新增了版本跟踪
    newState.CurrentVersion = "v1.2.0"
    newState.PreviousVersion = "v1.1.0"
    
    return nil
}
```

## 主要业务逻辑

### 1. 音频处理示例

```go
// audio_processor.go
package main

import (
    "context"
    "fmt"
    "sync/atomic"
    "time"
)

// AudioProcessor 音频处理器
type AudioProcessor struct {
    plugin *HotReloadPlugin
    
    // 处理参数
    sampleRate int
    channels   int
    bufferSize int
    
    // 统计信息
    processedSamples int64
    processingTime   int64
}

// NewAudioProcessor 创建音频处理器
func NewAudioProcessor(plugin *HotReloadPlugin) *AudioProcessor {
    return &AudioProcessor{
        plugin:     plugin,
        sampleRate: 44100,
        channels:   2,
        bufferSize: 1024,
    }
}

// ProcessAudio 处理音频数据
func (ap *AudioProcessor) ProcessAudio(ctx context.Context, input []float32) ([]float32, error) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        atomic.AddInt64(&ap.processingTime, duration.Nanoseconds())
        atomic.AddInt64(&ap.processedSamples, int64(len(input)))
        
        // 更新插件状态
        ap.plugin.stateMux.Lock()
        ap.plugin.state.ProcessedCount++
        ap.plugin.stateMux.Unlock()
    }()
    
    // 检查插件是否正在重载
    if ap.plugin.reloadManager != nil && ap.plugin.reloadManager.isReloading {
        // 在重载期间，可能需要特殊处理
        return ap.processAudioDuringReload(input)
    }
    
    // 正常音频处理逻辑
    output := make([]float32, len(input))
    
    // 示例：简单的音量调节
    volume := ap.getVolumeFromConfig()
    for i, sample := range input {
        output[i] = sample * volume
    }
    
    return output, nil
}

// processAudioDuringReload 重载期间的音频处理
func (ap *AudioProcessor) processAudioDuringReload(input []float32) ([]float32, error) {
    // 在重载期间，可以选择：
    // 1. 直接传递输入（bypass）
    // 2. 使用缓存的处理结果
    // 3. 应用简化的处理逻辑
    
    ap.plugin.logger.Debug("Processing audio during reload, using bypass mode")
    
    // 简单的 bypass 模式
    output := make([]float32, len(input))
    copy(output, input)
    
    return output, nil
}

// getVolumeFromConfig 从配置获取音量
func (ap *AudioProcessor) getVolumeFromConfig() float32 {
    ap.plugin.stateMux.RLock()
    defer ap.plugin.stateMux.RUnlock()
    
    if volumeInterface, exists := ap.plugin.state.CustomData["volume"]; exists {
        if volume, ok := volumeInterface.(float32); ok {
            return volume
        }
    }
    
    return 1.0 // 默认音量
}

// GetStats 获取处理统计
func (ap *AudioProcessor) GetStats() map[string]interface{} {
    processedSamples := atomic.LoadInt64(&ap.processedSamples)
    processingTime := atomic.LoadInt64(&ap.processingTime)
    
    var avgProcessingTime float64
    if processedSamples > 0 {
        avgProcessingTime = float64(processingTime) / float64(processedSamples)
    }
    
    return map[string]interface{}{
        "processed_samples":     processedSamples,
        "total_processing_time": time.Duration(processingTime),
        "avg_processing_time":   time.Duration(avgProcessingTime),
        "sample_rate":           ap.sampleRate,
        "channels":              ap.channels,
        "buffer_size":           ap.bufferSize,
    }
}
```

### 2. 主循环实现

```go
// 在 main.go 中添加主循环方法

// runMainLoop 运行主循环
func (p *HotReloadPlugin) runMainLoop(ctx context.Context) {
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()
    
    audioProcessor := NewAudioProcessor(p)
    
    for {
        select {
        case <-ctx.Done():
            return
            
        case <-p.stopChan:
            return
            
        case <-ticker.C:
            // 定期处理任务
            p.performPeriodicTasks()
            
        default:
            // 处理音频数据（示例）
            if p.hasAudioData() {
                audioData := p.getAudioData()
                if processedData, err := audioProcessor.ProcessAudio(ctx, audioData); err == nil {
                    p.outputAudioData(processedData)
                } else {
                    p.logger.Error("Audio processing failed", "error", err)
                    p.stateMux.Lock()
                    p.state.ErrorCount++
                    p.stateMux.Unlock()
                }
            }
            
            // 避免 CPU 占用过高
            time.Sleep(1 * time.Millisecond)
        }
    }
}

// performPeriodicTasks 执行定期任务
func (p *HotReloadPlugin) performPeriodicTasks() {
    // 更新统计信息
    p.updateStats()
    
    // 检查内存使用
    p.checkMemoryUsage()
    
    // 清理过期数据
    p.cleanupExpiredData()
}

// startReloadMonitor 启动重载监控
func (p *HotReloadPlugin) startReloadMonitor(ctx context.Context) {
    stateManager := NewStateManager(p, "/tmp/plugin_state")
    reloadManager, err := NewReloadManager(p, stateManager)
    if err != nil {
        p.logger.Error("Failed to create reload manager", "error", err)
        return
    }
    
    p.reloadManager = reloadManager
    
    // 设置监控路径
    reloadManager.watchPaths = []string{
        "./plugins/",
        "./config/",
    }
    
    if err := reloadManager.Start(ctx); err != nil {
        p.logger.Error("Failed to start reload manager", "error", err)
    }
}

// 辅助方法（示例实现）
func (p *HotReloadPlugin) hasAudioData() bool {
    // 检查是否有音频数据需要处理
    return false // 示例返回
}

func (p *HotReloadPlugin) getAudioData() []float32 {
    // 获取音频数据
    return make([]float32, 1024) // 示例返回
}

func (p *HotReloadPlugin) outputAudioData(data []float32) {
    // 输出处理后的音频数据
}

func (p *HotReloadPlugin) updateStats() {
    // 更新统计信息
}

func (p *HotReloadPlugin) checkMemoryUsage() {
    // 检查内存使用情况
}

func (p *HotReloadPlugin) cleanupExpiredData() {
    // 清理过期数据
}
```

## 构建和部署

### 1. 构建脚本

```bash
#!/bin/bash
# build.sh

set -e

PLUGIN_NAME="hot-reload-plugin"
VERSION="v1.0.0"
BUILD_DIR="build"
DIST_DIR="dist"

echo "Building hot reload plugin..."

# 创建构建目录
mkdir -p "$BUILD_DIR"
mkdir -p "$DIST_DIR"

# 构建插件
echo "Building plugin binary..."
go build -buildmode=plugin -o "$BUILD_DIR/${PLUGIN_NAME}_${VERSION}.so" .

# 复制配置文件
cp plugin.json "$BUILD_DIR/"

# 创建版本信息文件
cat > "$BUILD_DIR/version.json" << EOF
{
  "version": "$VERSION",
  "build_time": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "git_commit": "$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')",
  "go_version": "$(go version | cut -d' ' -f3)"
}
EOF

# 运行测试
echo "Running tests..."
go test -v ./...

# 创建发布包
echo "Creating distribution package..."
tar -czf "$DIST_DIR/${PLUGIN_NAME}_${VERSION}.tar.gz" -C "$BUILD_DIR" .

echo "Build completed successfully!"
echo "Plugin binary: $BUILD_DIR/${PLUGIN_NAME}_${VERSION}.so"
echo "Distribution package: $DIST_DIR/${PLUGIN_NAME}_${VERSION}.tar.gz"
```

### 2. 配置文件

```json
{
  "id": "hot-reload-plugin",
  "name": "Hot Reload Audio Plugin",
  "version": "v1.0.0",
  "description": "A hot-reloadable audio processing plugin with state persistence",
  "author": "Your Name",
  "license": "MIT",
  "type": "hot_reload",
  "category": "audio-effect",
  "tags": ["audio", "hot-reload", "real-time"],
  "api_version": "v2.0.0",
  "min_kernel_version": "2.0.0",
  "hot_reload": {
    "enabled": true,
    "watch_paths": [
      "./plugins/",
      "./config/"
    ],
    "debounce_time": "500ms",
    "max_versions": 20,
    "state_persistence": true,
    "rollback_enabled": true,
    "auto_reload": true
  },
  "permissions": [
    {
      "id": "audio-processing",
      "description": "Process audio data in real-time",
      "required": true
    },
    {
      "id": "file-watch",
      "description": "Monitor file system changes",
      "required": true
    },
    {
      "id": "state-persistence",
      "description": "Save and restore plugin state",
      "required": true
    }
  ],
  "config_schema": {
    "type": "object",
    "properties": {
      "audio_params": {
        "type": "object",
        "properties": {
          "sample_rate": {
            "type": "integer",
            "minimum": 8000,
            "maximum": 192000,
            "default": 44100
          },
          "channels": {
            "type": "integer",
            "minimum": 1,
            "maximum": 8,
            "default": 2
          },
          "buffer_size": {
            "type": "integer",
            "minimum": 64,
            "maximum": 8192,
            "default": 1024
          },
          "volume": {
            "type": "number",
            "minimum": 0.0,
            "maximum": 2.0,
            "default": 1.0
          }
        }
      },
      "reload_settings": {
        "type": "object",
        "properties": {
          "auto_reload": {
            "type": "boolean",
            "default": true
          },
          "debounce_time": {
            "type": "string",
            "pattern": "^\\d+[ms|s|m|h]$",
            "default": "500ms"
          },
          "max_versions": {
            "type": "integer",
            "minimum": 1,
            "maximum": 100,
            "default": 20
          }
        }
      }
    }
  },
  "dependencies": {
    "kernel": ">=2.0.0",
    "plugin-manager": ">=2.0.0",
    "event-bus": ">=2.0.0"
  },
  "build": {
    "go_version": ">=1.21",
    "build_mode": "plugin",
    "cgo_enabled": false,
    "ldflags": "-s -w"
  }
}
```

## 测试

### 1. 单元测试

```go
// test/hot_reload_test.go
package test

import (
    "context"
    "os"
    "path/filepath"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestHotReloadPlugin(t *testing.T) {
    // 创建临时目录
    tempDir, err := os.MkdirTemp("", "hot_reload_test")
    require.NoError(t, err)
    defer os.RemoveAll(tempDir)
    
    // 创建插件配置
    config := &PluginConfig{
        ID:      "test-hot-reload",
        Name:    "Test Hot Reload Plugin",
        Version: "v1.0.0",
        Path:    filepath.Join(tempDir, "plugin.so"),
    }
    
    // 创建插件实例
    plugin := NewHotReloadPlugin(config)
    require.NotNil(t, plugin)
    
    // 测试初始化
    ctx := context.Background()
    err = plugin.Initialize(ctx)
    assert.NoError(t, err)
    
    // 测试启动
    err = plugin.Start(ctx)
    assert.NoError(t, err)
    
    // 验证状态
    assert.True(t, plugin.state.IsRunning)
    assert.Equal(t, "v1.0.0", plugin.state.CurrentVersion)
    
    // 测试停止
    err = plugin.Stop(ctx)
    assert.NoError(t, err)
    
    // 测试清理
    err = plugin.Cleanup()
    assert.NoError(t, err)
}

func TestStateManager(t *testing.T) {
    tempDir, err := os.MkdirTemp("", "state_test")
    require.NoError(t, err)
    defer os.RemoveAll(tempDir)
    
    config := &PluginConfig{
        ID:      "test-state",
        Name:    "Test State Plugin",
        Version: "v1.0.0",
    }
    
    plugin := NewHotReloadPlugin(config)
    stateManager := NewStateManager(plugin, tempDir)
    
    // 测试状态保存
    plugin.state.ProcessedCount = 100
    plugin.state.ErrorCount = 5
    plugin.state.CustomData["test_key"] = "test_value"
    
    err = stateManager.SaveState()
    assert.NoError(t, err)
    
    // 修改状态
    plugin.state.ProcessedCount = 200
    plugin.state.CustomData["test_key"] = "modified_value"
    
    // 测试状态加载
    err = stateManager.LoadState()
    assert.NoError(t, err)
    
    // 验证状态恢复
    assert.Equal(t, int64(100), plugin.state.ProcessedCount)
    assert.Equal(t, int64(5), plugin.state.ErrorCount)
    assert.Equal(t, "test_value", plugin.state.CustomData["test_key"])
}

func TestVersionManager(t *testing.T) {
    config := &PluginConfig{
        ID:      "test-version",
        Name:    "Test Version Plugin",
        Version: "v1.0.0",
    }
    
    plugin := NewHotReloadPlugin(config)
    versionManager := NewVersionManager(plugin)
    
    // 测试版本比较
    assert.Equal(t, -1, versionManager.CompareVersions("v1.0.0", "v1.1.0"))
    assert.Equal(t, 1, versionManager.CompareVersions("v1.1.0", "v1.0.0"))
    assert.Equal(t, 0, versionManager.CompareVersions("v1.0.0", "v1.0.0"))
    
    // 测试版本解析
    parts := versionManager.parseVersion("v1.2.3")
    assert.Equal(t, []int{1, 2, 3}, parts)
}

func TestReloadManager(t *testing.T) {
    tempDir, err := os.MkdirTemp("", "reload_test")
    require.NoError(t, err)
    defer os.RemoveAll(tempDir)
    
    config := &PluginConfig{
        ID:      "test-reload",
        Name:    "Test Reload Plugin",
        Version: "v1.0.0",
        Path:    filepath.Join(tempDir, "plugin.so"),
    }
    
    plugin := NewHotReloadPlugin(config)
    stateManager := NewStateManager(plugin, tempDir)
    
    reloadManager, err := NewReloadManager(plugin, stateManager)
    require.NoError(t, err)
    
    // 测试文件事件过滤
    event1 := fsnotify.Event{Name: "test.so", Op: fsnotify.Write}
    event2 := fsnotify.Event{Name: "test.txt", Op: fsnotify.Write}
    event3 := fsnotify.Event{Name: "test.so", Op: fsnotify.Remove}
    
    assert.True(t, reloadManager.shouldProcessEvent(event1))
    assert.False(t, reloadManager.shouldProcessEvent(event2))
    assert.False(t, reloadManager.shouldProcessEvent(event3))
}
```

### 2. 集成测试

```go
// test/integration_test.go
package test

import (
    "context"
    "os"
    "path/filepath"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestHotReloadIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    tempDir, err := os.MkdirTemp("", "integration_test")
    require.NoError(t, err)
    defer os.RemoveAll(tempDir)
    
    // 创建插件配置
    config := &PluginConfig{
        ID:      "integration-test",
        Name:    "Integration Test Plugin",
        Version: "v1.0.0",
        Path:    filepath.Join(tempDir, "plugin_v1.0.0.so"),
    }
    
    // 创建插件实例
    plugin := NewHotReloadPlugin(config)
    
    // 初始化和启动插件
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    err = plugin.Initialize(ctx)
    require.NoError(t, err)
    
    err = plugin.Start(ctx)
    require.NoError(t, err)
    
    // 等待插件启动
    time.Sleep(100 * time.Millisecond)
    
    // 验证初始状态
    assert.True(t, plugin.state.IsRunning)
    assert.Equal(t, "v1.0.0", plugin.state.CurrentVersion)
    
    // 模拟文件变化触发重载
    newPluginPath := filepath.Join(tempDir, "plugin_v1.1.0.so")
    
    // 创建新版本文件（模拟）
    err = os.WriteFile(newPluginPath, []byte("mock plugin content"), 0644)
    require.NoError(t, err)
    
    // 发送重载信号
    reloadSignal := ReloadSignal{
        Type:      ReloadTypeUpdate,
        NewPath:   newPluginPath,
        OldPath:   config.Path,
        Timestamp: time.Now(),
        Metadata: map[string]interface{}{
            "trigger": "file_change",
        },
    }
    
    select {
    case plugin.reloadChan <- reloadSignal:
        // 信号发送成功
    case <-time.After(1 * time.Second):
        t.Fatal("Failed to send reload signal")
    }
    
    // 等待重载完成
    time.Sleep(2 * time.Second)
    
    // 验证重载后状态
    assert.True(t, plugin.state.IsRunning)
    assert.Greater(t, plugin.state.ReloadCount, 0)
    
    // 停止插件
    err = plugin.Stop(ctx)
    assert.NoError(t, err)
    
    err = plugin.Cleanup()
    assert.NoError(t, err)
}

func TestStatePeristenceIntegration(t *testing.T) {
    tempDir, err := os.MkdirTemp("", "persistence_test")
    require.NoError(t, err)
    defer os.RemoveAll(tempDir)
    
    config := &PluginConfig{
        ID:      "persistence-test",
        Name:    "Persistence Test Plugin",
        Version: "v1.0.0",
    }
    
    // 第一个插件实例
    plugin1 := NewHotReloadPlugin(config)
    stateManager1 := NewStateManager(plugin1, tempDir)
    
    // 设置状态
    plugin1.state.ProcessedCount = 1000
    plugin1.state.ErrorCount = 10
    plugin1.state.CustomData["session_id"] = "test-session-123"
    plugin1.state.CustomData["user_preferences"] = map[string]interface{}{
        "volume": 0.8,
        "eq_enabled": true,
    }
    
    // 保存状态
    err = stateManager1.SaveState()
    require.NoError(t, err)
    
    // 创建状态快照
    err = stateManager1.CreateSnapshot("v1.0.0")
    require.NoError(t, err)
    
    // 第二个插件实例（模拟重启）
    plugin2 := NewHotReloadPlugin(config)
    stateManager2 := NewStateManager(plugin2, tempDir)
    
    // 加载状态
    err = stateManager2.LoadState()
    require.NoError(t, err)
    
    // 验证状态恢复
    assert.Equal(t, int64(1000), plugin2.state.ProcessedCount)
    assert.Equal(t, int64(10), plugin2.state.ErrorCount)
    assert.Equal(t, "test-session-123", plugin2.state.CustomData["session_id"])
    
    userPrefs, ok := plugin2.state.CustomData["user_preferences"].(map[string]interface{})
    require.True(t, ok)
    assert.Equal(t, 0.8, userPrefs["volume"])
    assert.Equal(t, true, userPrefs["eq_enabled"])
    
    // 测试快照恢复
    plugin2.state.ProcessedCount = 2000 // 修改状态
    
    err = stateManager2.RestoreSnapshot("v1.0.0")
    require.NoError(t, err)
    
    // 验证快照恢复
    assert.Equal(t, int64(1000), plugin2.state.ProcessedCount)
}
```

## 性能优化

### 1. 内存优化

- 使用对象池减少内存分配
- 及时释放不再使用的资源
- 监控内存使用情况
- 实现内存泄漏检测

### 2. CPU 优化

- 使用协程池处理并发任务
- 避免在热路径中进行重载检查
- 优化状态同步机制
- 使用无锁数据结构

### 3. I/O 优化

- 批量处理文件事件
- 使用异步 I/O 操作
- 实现智能缓存机制
- 优化状态序列化性能

## 最佳实践

### 1. 安全性

- 验证插件签名
- 限制文件访问权限
- 实现安全的状态迁移
- 防止恶意插件注入

### 2. 可靠性

- 实现完善的错误处理
- 提供回滚机制
- 监控插件健康状态
- 实现自动恢复机制

### 3. 可维护性

- 提供详细的日志记录
- 实现调试接口
- 编写完整的测试用例
- 提供性能监控工具

### 4. 兼容性

- 保持 API 向后兼容
- 实现渐进式升级
- 支持多版本共存
- 提供迁移工具

## 常见问题

### Q: 热重载会影响音频处理的实时性吗？

A: 通过合理的设计，热重载对实时性的影响可以降到最低。在重载期间可以使用 bypass 模式或缓存结果。

### Q: 如何处理插件依赖关系？

A: 实现依赖图分析，按照依赖顺序进行重载，确保依赖的插件先完成重载。

### Q: 状态迁移失败怎么办？

A: 提供回滚机制，在迁移失败时自动恢复到之前的版本和状态。

### Q: 如何调试热重载问题？

A: 启用详细日志记录，使用调试模式，监控重载过程中的各个步骤。

## 相关文档

- [插件开发快速入门](plugin-quickstart.md)
- [动态库插件开发指南](dynamic-library.md)
- [RPC 插件开发指南](rpc-plugin.md)
- [WebAssembly 插件开发指南](webassembly.md)
- [插件测试指南](plugin-testing.md)
- [API 文档](../api/README.md)
```