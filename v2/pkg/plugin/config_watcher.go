package plugin

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"log/slog"
)

// FileConfigWatcher 配置文件监控器实现
type FileConfigWatcher struct {
	logger     *slog.Logger
	configPath string
	pluginID   string
	watcher    *fsnotify.Watcher
	events     chan *ConfigChangeEvent
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.RWMutex
	isRunning  bool
	lastHash   string
	debounce   time.Duration
	lastChange time.Time
	backupDir  string
	maxBackups int
}

// ConfigWatcherOptions 配置监控器选项
type ConfigWatcherOptions struct {
	Debounce   time.Duration // 防抖动时间
	BackupDir  string        // 备份目录
	MaxBackups int           // 最大备份数量
}

// DefaultConfigWatcherOptions 默认配置监控器选项
func DefaultConfigWatcherOptions() *ConfigWatcherOptions {
	return &ConfigWatcherOptions{
		Debounce:   500 * time.Millisecond,
		BackupDir:  "",
		MaxBackups: 10,
	}
}

// NewConfigWatcher 创建配置监控器
func NewConfigWatcher(logger *slog.Logger, configPath, pluginID string) *FileConfigWatcher {
	return NewConfigWatcherWithOptions(logger, configPath, pluginID, DefaultConfigWatcherOptions())
}

// NewConfigWatcherWithOptions 使用选项创建配置监控器
func NewConfigWatcherWithOptions(logger *slog.Logger, configPath, pluginID string, options *ConfigWatcherOptions) *FileConfigWatcher {
	ctx, cancel := context.WithCancel(context.Background())

	backupDir := options.BackupDir
	if backupDir == "" {
		backupDir = filepath.Join(filepath.Dir(configPath), ".backups")
	}

	return &FileConfigWatcher{
		logger:     logger,
		configPath: configPath,
		pluginID:   pluginID,
		events:     make(chan *ConfigChangeEvent, 100),
		ctx:        ctx,
		cancel:     cancel,
		debounce:   options.Debounce,
		backupDir:  backupDir,
		maxBackups: options.MaxBackups,
	}
}

// Start 启动配置监控
func (cw *FileConfigWatcher) Start(ctx context.Context) error {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	if cw.isRunning {
		return fmt.Errorf("config watcher is already running")
	}

	// 创建文件系统监控器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	cw.watcher = watcher

	// 创建备份目录
	if err := os.MkdirAll(cw.backupDir, 0755); err != nil {
		cw.logger.Warn("Failed to create backup directory", "dir", cw.backupDir, "error", err)
	}

	// 计算初始文件哈希
	cw.lastHash, _ = cw.calculateFileHash(cw.configPath)

	// 监控配置文件
	if err := cw.watcher.Add(cw.configPath); err != nil {
		// 如果文件不存在，监控目录
		dir := filepath.Dir(cw.configPath)
		if err := cw.watcher.Add(dir); err != nil {
			cw.watcher.Close()
			return fmt.Errorf("failed to watch config file or directory: %w", err)
		}
	}

	cw.isRunning = true

	// 启动监控协程
	go cw.watchLoop()

	cw.logger.Info("Config watcher started", "plugin_id", cw.pluginID, "config_path", cw.configPath)
	return nil
}

// Stop 停止配置监控
func (cw *FileConfigWatcher) Stop() {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	if !cw.isRunning {
		return
	}

	cw.isRunning = false
	cw.cancel()

	if cw.watcher != nil {
		cw.watcher.Close()
	}

	close(cw.events)

	cw.logger.Info("Config watcher stopped", "plugin_id", cw.pluginID)
}

// Events 获取配置变更事件通道
func (cw *FileConfigWatcher) Events() <-chan *ConfigChangeEvent {
	return cw.events
}

// IsRunning 检查监控器是否运行中
func (cw *FileConfigWatcher) IsRunning() bool {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return cw.isRunning
}

// ForceReload 强制重新加载配置
func (cw *FileConfigWatcher) ForceReload() error {
	cw.mu.RLock()
	defer cw.mu.RUnlock()

	if !cw.isRunning {
		return fmt.Errorf("config watcher is not running")
	}

	// 发送重新加载事件
	event := &ConfigChangeEvent{
		PluginID:   cw.pluginID,
		ChangeType: ConfigChangeTypeReload,
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"forced": true,
		},
	}

	select {
	case cw.events <- event:
		return nil
	default:
		return fmt.Errorf("event channel is full")
	}
}

// watchLoop 监控循环
func (cw *FileConfigWatcher) watchLoop() {
	defer func() {
		if r := recover(); r != nil {
			cw.logger.Error("Config watcher panic recovered", "plugin_id", cw.pluginID, "panic", r)
		}
	}()

	for {
		select {
		case <-cw.ctx.Done():
			return

		case event, ok := <-cw.watcher.Events:
			if !ok {
				return
			}

			// 只处理与配置文件相关的事件
			if !cw.isConfigFileEvent(event) {
				continue
			}

			cw.handleFileEvent(event)

		case err, ok := <-cw.watcher.Errors:
			if !ok {
				return
			}

			cw.logger.Error("Config watcher error", "plugin_id", cw.pluginID, "error", err)
		}
	}
}

// isConfigFileEvent 检查是否为配置文件事件
func (cw *FileConfigWatcher) isConfigFileEvent(event fsnotify.Event) bool {
	// 检查文件名是否匹配
	if event.Name == cw.configPath {
		return true
	}

	// 检查是否为同一目录下的同名文件（可能是临时文件重命名）
	dir := filepath.Dir(cw.configPath)
	base := filepath.Base(cw.configPath)
	eventDir := filepath.Dir(event.Name)
	eventBase := filepath.Base(event.Name)

	return dir == eventDir && (eventBase == base || cw.isTempFile(eventBase, base))
}

// isTempFile 检查是否为临时文件
func (cw *FileConfigWatcher) isTempFile(eventFile, configFile string) bool {
	// 常见的临时文件模式
	tempPatterns := []string{
		configFile + ".tmp",
		configFile + ".temp",
		"." + configFile + ".tmp",
		"." + configFile + ".swp",
		configFile + "~",
	}

	for _, pattern := range tempPatterns {
		if eventFile == pattern {
			return true
		}
	}

	return false
}

// handleFileEvent 处理文件事件
func (cw *FileConfigWatcher) handleFileEvent(event fsnotify.Event) {
	now := time.Now()

	// 防抖动：如果距离上次变更时间太短，则忽略
	if now.Sub(cw.lastChange) < cw.debounce {
		return
	}
	cw.lastChange = now

	// 延迟处理，等待文件写入完成
	go func() {
		time.Sleep(cw.debounce)
		cw.processFileChange(event)
	}()
}

// processFileChange 处理文件变更
func (cw *FileConfigWatcher) processFileChange(event fsnotify.Event) {
	// 计算新的文件哈希
	newHash, err := cw.calculateFileHash(cw.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件被删除
			cw.sendChangeEvent(ConfigChangeTypeDelete, nil, nil)
			return
		}
		cw.logger.Error("Failed to calculate file hash", "plugin_id", cw.pluginID, "error", err)
		return
	}

	// 检查文件是否真的发生了变更
	if newHash == cw.lastHash {
		return // 文件内容没有变更
	}

	// 创建备份
	if err := cw.createBackup(); err != nil {
		cw.logger.Warn("Failed to create backup", "plugin_id", cw.pluginID, "error", err)
	}

	// 确定变更类型
	changeType := ConfigChangeTypeUpdate
	if cw.lastHash == "" {
		changeType = ConfigChangeTypeCreate
	}

	cw.lastHash = newHash

	// 发送变更事件
	cw.sendChangeEvent(changeType, nil, nil)
}

// sendChangeEvent 发送配置变更事件
func (cw *FileConfigWatcher) sendChangeEvent(changeType ConfigChangeType, oldConfig, newConfig *EnhancedPluginConfig) {
	event := &ConfigChangeEvent{
		PluginID:   cw.pluginID,
		ChangeType: changeType,
		OldConfig:  oldConfig,
		NewConfig:  newConfig,
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"config_path": cw.configPath,
			"file_hash":   cw.lastHash,
		},
	}

	select {
	case cw.events <- event:
		cw.logger.Info("Config change event sent",
			"plugin_id", cw.pluginID,
			"change_type", changeType,
			"timestamp", event.Timestamp)
	default:
		cw.logger.Warn("Config change event dropped (channel full)", "plugin_id", cw.pluginID)
	}
}

// calculateFileHash 计算文件哈希值
func (cw *FileConfigWatcher) calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// createBackup 创建配置文件备份
func (cw *FileConfigWatcher) createBackup() error {
	if _, err := os.Stat(cw.configPath); os.IsNotExist(err) {
		return nil // 文件不存在，无需备份
	}

	// 生成备份文件名
	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("%s_%s.backup", cw.pluginID, timestamp)
	backupPath := filepath.Join(cw.backupDir, backupName)

	// 读取原文件
	data, err := os.ReadFile(cw.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 写入备份文件
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	// 清理旧备份
	go cw.cleanupOldBackups()

	cw.logger.Debug("Config backup created", "plugin_id", cw.pluginID, "backup_path", backupPath)
	return nil
}

// cleanupOldBackups 清理旧备份文件
func (cw *FileConfigWatcher) cleanupOldBackups() {
	pattern := filepath.Join(cw.backupDir, cw.pluginID+"_*.backup")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		cw.logger.Warn("Failed to list backup files", "plugin_id", cw.pluginID, "error", err)
		return
	}

	// 如果备份文件数量超过限制，删除最旧的文件
	if len(matches) > cw.maxBackups {
		// 按修改时间排序
		type fileInfo struct {
			path    string
			modTime time.Time
		}

		files := make([]fileInfo, 0, len(matches))
		for _, match := range matches {
			if stat, err := os.Stat(match); err == nil {
				files = append(files, fileInfo{path: match, modTime: stat.ModTime()})
			}
		}

		// 简单排序（按修改时间）
		for i := 0; i < len(files)-1; i++ {
			for j := i + 1; j < len(files); j++ {
				if files[i].modTime.After(files[j].modTime) {
					files[i], files[j] = files[j], files[i]
				}
			}
		}

		// 删除最旧的文件
		for i := 0; i < len(files)-cw.maxBackups; i++ {
			if err := os.Remove(files[i].path); err != nil {
				cw.logger.Warn("Failed to remove old backup", "path", files[i].path, "error", err)
			}
		}
	}
}

// RestoreFromBackup 从备份恢复配置
func (cw *FileConfigWatcher) RestoreFromBackup(backupName string) error {
	backupPath := filepath.Join(cw.backupDir, backupName)

	// 检查备份文件是否存在
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backupName)
	}

	// 读取备份文件
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	// 创建当前配置的备份
	if err := cw.createBackup(); err != nil {
		cw.logger.Warn("Failed to backup current config before restore", "error", err)
	}

	// 恢复配置文件
	if err := os.WriteFile(cw.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to restore config file: %w", err)
	}

	cw.logger.Info("Config restored from backup", "plugin_id", cw.pluginID, "backup", backupName)
	return nil
}

// ListBackups 列出所有备份文件
func (cw *FileConfigWatcher) ListBackups() ([]string, error) {
	pattern := filepath.Join(cw.backupDir, cw.pluginID+"_*.backup")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list backup files: %w", err)
	}

	// 提取文件名
	backups := make([]string, 0, len(matches))
	for _, match := range matches {
		backups = append(backups, filepath.Base(match))
	}

	return backups, nil
}

// ConfigHotReloader 配置热重载器
type ConfigHotReloader struct {
	logger        *slog.Logger
	configManager *PluginConfigManager
	// TODO: pluginManager类型待定义
	// pluginManager PluginManager
	pluginManager interface{}
	watchers      map[string]ConfigWatcher
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewConfigHotReloader 创建配置热重载器
func NewConfigHotReloader(logger *slog.Logger, configManager *PluginConfigManager, pluginManager interface{}) *ConfigHotReloader {
	ctx, cancel := context.WithCancel(context.Background())
	return &ConfigHotReloader{
		logger:        logger,
		configManager: configManager,
		pluginManager: pluginManager,
		watchers:      make(map[string]ConfigWatcher),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start 启动热重载器
func (hr *ConfigHotReloader) Start(ctx context.Context) error {
	// 为所有已配置的插件启动监控
	configs, err := hr.configManager.ListConfigs(ctx)
	if err != nil {
		return fmt.Errorf("failed to list configs: %w", err)
	}

	for _, config := range configs {
		if config.AutoReload {
			if err := hr.StartWatching(ctx, config.ID); err != nil {
				hr.logger.Warn("Failed to start watching config", "plugin_id", config.ID, "error", err)
			}
		}
	}

	hr.logger.Info("Config hot reloader started")
	return nil
}

// Stop 停止热重载器
func (hr *ConfigHotReloader) Stop() {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	hr.cancel()

	// 停止所有监控器
	for _, watcher := range hr.watchers {
		watcher.Stop()
	}

	hr.logger.Info("Config hot reloader stopped")
}

// StartWatching 开始监控指定插件的配置
func (hr *ConfigHotReloader) StartWatching(ctx context.Context, pluginID string) error {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	if _, exists := hr.watchers[pluginID]; exists {
		return fmt.Errorf("already watching plugin %s", pluginID)
	}

	config, err := hr.configManager.LoadConfig(ctx, pluginID)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	watcher := NewConfigWatcher(hr.logger, config.ConfigPath, pluginID)
	if err := watcher.Start(ctx); err != nil {
		return fmt.Errorf("failed to start watcher: %w", err)
	}

	hr.watchers[pluginID] = ConfigWatcher(watcher)

	// 启动事件处理协程
	go hr.handleConfigChanges(pluginID, watcher.Events())

	return nil
}

// StopWatching 停止监控指定插件的配置
func (hr *ConfigHotReloader) StopWatching(pluginID string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	if watcher, exists := hr.watchers[pluginID]; exists {
		watcher.Stop()
		delete(hr.watchers, pluginID)
	}
}

// handleConfigChanges 处理配置变更事件
func (hr *ConfigHotReloader) handleConfigChanges(pluginID string, events <-chan *ConfigChangeEvent) {
	for {
		select {
		case <-hr.ctx.Done():
			return

		case event, ok := <-events:
			if !ok {
				return
			}

			hr.processConfigChange(event)
		}
	}
}

// processConfigChange 处理配置变更
func (hr *ConfigHotReloader) processConfigChange(event *ConfigChangeEvent) {
	hr.logger.Info("Processing config change",
		"plugin_id", event.PluginID,
		"change_type", event.ChangeType,
		"timestamp", event.Timestamp)

	switch event.ChangeType {
	case ConfigChangeTypeCreate, ConfigChangeTypeUpdate, ConfigChangeTypeReload:
		if err := hr.reloadPluginConfig(event.PluginID); err != nil {
			hr.logger.Error("Failed to reload plugin config",
				"plugin_id", event.PluginID,
				"error", err)
		}

	case ConfigChangeTypeDelete:
		hr.logger.Info("Plugin config deleted", "plugin_id", event.PluginID)
		// 可以选择停止插件或使用默认配置
	}
}

// reloadPluginConfig 重新加载插件配置
func (hr *ConfigHotReloader) reloadPluginConfig(pluginID string) error {
	// 重新加载配置
	newConfig, err := hr.configManager.LoadConfig(context.Background(), pluginID)
	if err != nil {
		return fmt.Errorf("failed to load new config: %w", err)
	}

	// 验证新配置
	if err := hr.configManager.validateConfig(context.Background(), newConfig); err != nil {
		return fmt.Errorf("new config validation failed: %w", err)
	}

	// 应用新配置到插件
	if hr.pluginManager != nil {
		if err := hr.applyConfigToPlugin(pluginID, newConfig); err != nil {
			return fmt.Errorf("failed to apply config to plugin: %w", err)
		}
	}

	hr.logger.Info("Plugin config reloaded successfully", "plugin_id", pluginID)
	return nil
}

// applyConfigToPlugin 将配置应用到插件
func (hr *ConfigHotReloader) applyConfigToPlugin(pluginID string, config *EnhancedPluginConfig) error {
	// TODO: 实现插件配置应用逻辑
	// 由于插件现在是interface{}类型，需要重新设计
	hr.logger.Info("Plugin config application not implemented", "plugin_id", pluginID)
	return nil
}