package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// HotReloadConfig 热更新配置
type HotReloadConfig struct {
	Enabled          bool          `json:"enabled"`
	WatchDirectories []string      `json:"watch_directories"`
	WatchFiles       []string      `json:"watch_files"`
	IgnorePatterns   []string      `json:"ignore_patterns"`
	DebounceDelay    time.Duration `json:"debounce_delay"`
	MaxRetries       int           `json:"max_retries"`
	RetryDelay       time.Duration `json:"retry_delay"`
	BackupOnReload   bool          `json:"backup_on_reload"`
}

// DefaultHotReloadConfig 默认热更新配置
func DefaultHotReloadConfig() *HotReloadConfig {
	return &HotReloadConfig{
		Enabled:          false,
		WatchDirectories: []string{},
		WatchFiles:       []string{},
		IgnorePatterns:   []string{"*.tmp", "*.swp", "*~", ".#*"},
		DebounceDelay:    500 * time.Millisecond,
		MaxRetries:       3,
		RetryDelay:       1 * time.Second,
		BackupOnReload:   true,
	}
}

// HotReloadManager 热更新管理器
type HotReloadManager struct {
	config          *HotReloadConfig
	configManager   *AdvancedManager
	watcher         *fsnotify.Watcher
	ctx             context.Context
	cancel          context.CancelFunc
	eventChan       chan *ReloadEvent
	debounceTimer   *time.Timer
	pendingEvents   map[string]*ReloadEvent
	mutex           sync.RWMutex
	running         bool
	stats           *HotReloadStats
}

// ReloadEvent 重载事件
type ReloadEvent struct {
	Type      ReloadEventType `json:"type"`
	Path      string          `json:"path"`
	Operation string          `json:"operation"` // create, write, remove, rename
	Timestamp time.Time       `json:"timestamp"`
	Retries   int             `json:"retries"`
}

// ReloadEventType 重载事件类型
type ReloadEventType int

const (
	ReloadEventConfig ReloadEventType = iota
	ReloadEventTemplate
	ReloadEventPlugin
	ReloadEventDirectory
)

// HotReloadStats 热更新统计信息
type HotReloadStats struct {
	StartTime        time.Time `json:"start_time"`
	TotalEvents      int64     `json:"total_events"`
	SuccessfulReloads int64     `json:"successful_reloads"`
	FailedReloads    int64     `json:"failed_reloads"`
	LastReloadTime   time.Time `json:"last_reload_time"`
	LastError        string    `json:"last_error"`
	WatchedPaths     []string  `json:"watched_paths"`
}

// NewHotReloadManager 创建热更新管理器
func NewHotReloadManager(configManager *AdvancedManager, config *HotReloadConfig) *HotReloadManager {
	if config == nil {
		config = DefaultHotReloadConfig()
	}

	return &HotReloadManager{
		config:        config,
		configManager: configManager,
		eventChan:     make(chan *ReloadEvent, 100),
		pendingEvents: make(map[string]*ReloadEvent),
		stats: &HotReloadStats{
			WatchedPaths: make([]string, 0),
		},
	}
}

// Start 启动热更新管理器
func (hrm *HotReloadManager) Start(ctx context.Context) error {
	hrm.mutex.Lock()
	defer hrm.mutex.Unlock()

	if hrm.running {
		return fmt.Errorf("hot reload manager is already running")
	}

	if !hrm.config.Enabled {
		return fmt.Errorf("hot reload is disabled")
	}

	// 创建文件监控器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	hrm.watcher = watcher
	hrm.ctx, hrm.cancel = context.WithCancel(ctx)
	hrm.running = true
	hrm.stats.StartTime = time.Now()

	// 添加监控路径
	if err := hrm.addWatchPaths(); err != nil {
		hrm.Stop()
		return fmt.Errorf("failed to add watch paths: %w", err)
	}

	// 启动事件处理协程
	go hrm.handleEvents()
	go hrm.processReloadEvents()

	return nil
}

// Stop 停止热更新管理器
func (hrm *HotReloadManager) Stop() error {
	hrm.mutex.Lock()
	defer hrm.mutex.Unlock()

	if !hrm.running {
		return nil
	}

	if hrm.cancel != nil {
		hrm.cancel()
	}

	if hrm.watcher != nil {
		hrm.watcher.Close()
	}

	if hrm.debounceTimer != nil {
		hrm.debounceTimer.Stop()
	}

	hrm.running = false
	return nil
}

// IsRunning 检查热更新管理器是否运行中
func (hrm *HotReloadManager) IsRunning() bool {
	hrm.mutex.RLock()
	defer hrm.mutex.RUnlock()
	return hrm.running
}

// addWatchPaths 添加监控路径
func (hrm *HotReloadManager) addWatchPaths() error {
	// 添加配置文件监控
	for _, file := range hrm.config.WatchFiles {
		if err := hrm.addWatchPath(file); err != nil {
			return fmt.Errorf("failed to watch file %s: %w", file, err)
		}
	}

	// 添加目录监控
	for _, dir := range hrm.config.WatchDirectories {
		if err := hrm.addWatchPath(dir); err != nil {
			return fmt.Errorf("failed to watch directory %s: %w", dir, err)
		}
	}

	// 添加配置管理器的配置文件
	if hrm.configManager.configFile != "" {
		if err := hrm.addWatchPath(hrm.configManager.configFile); err != nil {
			return fmt.Errorf("failed to watch config file %s: %w", hrm.configManager.configFile, err)
		}
	}

	// 添加配置目录
	if hrm.configManager.configDir != "" {
		if err := hrm.addWatchPath(hrm.configManager.configDir); err != nil {
			// 目录监控失败不是致命错误
			fmt.Printf("Warning: failed to watch config directory %s: %v\n", hrm.configManager.configDir, err)
		}
	}

	return nil
}

// addWatchPath 添加单个监控路径
func (hrm *HotReloadManager) addWatchPath(path string) error {
	// 检查路径是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	// 添加到监控器
	if err := hrm.watcher.Add(path); err != nil {
		return err
	}

	// 记录监控路径
	hrm.stats.WatchedPaths = append(hrm.stats.WatchedPaths, path)

	return nil
}

// handleEvents 处理文件系统事件
func (hrm *HotReloadManager) handleEvents() {
	for {
		select {
		case <-hrm.ctx.Done():
			return
		case event, ok := <-hrm.watcher.Events:
			if !ok {
				return
			}
			hrm.handleFileEvent(event)
		case err, ok := <-hrm.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("File watcher error: %v\n", err)
			hrm.stats.LastError = err.Error()
		}
	}
}

// handleFileEvent 处理单个文件事件
func (hrm *HotReloadManager) handleFileEvent(event fsnotify.Event) {
	// 检查是否应该忽略此事件
	if hrm.shouldIgnoreEvent(event) {
		return
	}

	// 创建重载事件
	reloadEvent := &ReloadEvent{
		Type:      hrm.getEventType(event.Name),
		Path:      event.Name,
		Operation: hrm.getOperation(event.Op),
		Timestamp: time.Now(),
		Retries:   0,
	}

	// 更新统计信息
	hrm.stats.TotalEvents++

	// 发送到事件通道
	select {
	case hrm.eventChan <- reloadEvent:
	default:
		fmt.Printf("Warning: reload event channel is full, dropping event for %s\n", event.Name)
	}
}

// shouldIgnoreEvent 检查是否应该忽略事件
func (hrm *HotReloadManager) shouldIgnoreEvent(event fsnotify.Event) bool {
	// 只处理写入和重命名事件
	if event.Op&fsnotify.Write == 0 && event.Op&fsnotify.Rename == 0 && event.Op&fsnotify.Create == 0 {
		return true
	}

	// 检查忽略模式
	filename := filepath.Base(event.Name)
	for _, pattern := range hrm.config.IgnorePatterns {
		if matched, _ := filepath.Match(pattern, filename); matched {
			return true
		}
	}

	return false
}

// getEventType 获取事件类型
func (hrm *HotReloadManager) getEventType(path string) ReloadEventType {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json", ".yaml", ".yml", ".toml":
		if strings.Contains(path, "template") {
			return ReloadEventTemplate
		}
		return ReloadEventConfig
	case ".so", ".dll", ".dylib":
		return ReloadEventPlugin
	default:
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return ReloadEventDirectory
		}
		return ReloadEventConfig
	}
}

// getOperation 获取操作类型
func (hrm *HotReloadManager) getOperation(op fsnotify.Op) string {
	switch {
	case op&fsnotify.Create == fsnotify.Create:
		return "create"
	case op&fsnotify.Write == fsnotify.Write:
		return "write"
	case op&fsnotify.Remove == fsnotify.Remove:
		return "remove"
	case op&fsnotify.Rename == fsnotify.Rename:
		return "rename"
	case op&fsnotify.Chmod == fsnotify.Chmod:
		return "chmod"
	default:
		return "unknown"
	}
}

// processReloadEvents 处理重载事件
func (hrm *HotReloadManager) processReloadEvents() {
	for {
		select {
		case <-hrm.ctx.Done():
			return
		case event := <-hrm.eventChan:
			hrm.debounceEvent(event)
		}
	}
}

// debounceEvent 防抖处理事件
func (hrm *HotReloadManager) debounceEvent(event *ReloadEvent) {
	hrm.mutex.Lock()
	defer hrm.mutex.Unlock()

	// 将事件添加到待处理列表
	hrm.pendingEvents[event.Path] = event

	// 重置防抖定时器
	if hrm.debounceTimer != nil {
		hrm.debounceTimer.Stop()
	}

	hrm.debounceTimer = time.AfterFunc(hrm.config.DebounceDelay, func() {
		hrm.processPendingEvents()
	})
}

// processPendingEvents 处理待处理的事件
func (hrm *HotReloadManager) processPendingEvents() {
	hrm.mutex.Lock()
	pendingEvents := make([]*ReloadEvent, 0, len(hrm.pendingEvents))
	for _, event := range hrm.pendingEvents {
		pendingEvents = append(pendingEvents, event)
	}
	hrm.pendingEvents = make(map[string]*ReloadEvent)
	hrm.mutex.Unlock()

	// 处理每个事件
	for _, event := range pendingEvents {
		hrm.processReloadEvent(event)
	}
}

// processReloadEvent 处理单个重载事件
func (hrm *HotReloadManager) processReloadEvent(event *ReloadEvent) {
	// 创建备份（如果启用）
	if hrm.config.BackupOnReload {
		if _, err := hrm.configManager.CreateSnapshot(fmt.Sprintf("Backup before hot reload: %s", event.Path)); err != nil {
			fmt.Printf("Warning: failed to create backup before reload: %v\n", err)
		}
	}

	// 根据事件类型处理
	var err error
	switch event.Type {
	case ReloadEventConfig:
		err = hrm.reloadConfig(event)
	case ReloadEventTemplate:
		err = hrm.reloadTemplate(event)
	case ReloadEventPlugin:
		err = hrm.reloadPlugin(event)
	case ReloadEventDirectory:
		err = hrm.reloadDirectory(event)
	default:
		err = fmt.Errorf("unknown event type: %v", event.Type)
	}

	// 处理结果
	if err != nil {
		hrm.handleReloadError(event, err)
	} else {
		hrm.handleReloadSuccess(event)
	}
}

// reloadConfig 重载配置文件
func (hrm *HotReloadManager) reloadConfig(event *ReloadEvent) error {
	// 检查文件是否存在
	if _, err := os.Stat(event.Path); os.IsNotExist(err) {
		return fmt.Errorf("config file does not exist: %s", event.Path)
	}

	// 重新加载配置文件
	if err := hrm.configManager.LoadFromFile(event.Path); err != nil {
		return fmt.Errorf("failed to reload config file: %w", err)
	}

	return nil
}

// reloadTemplate 重载模板文件
func (hrm *HotReloadManager) reloadTemplate(event *ReloadEvent) error {
	// 检查文件是否存在
	if _, err := os.Stat(event.Path); os.IsNotExist(err) {
		return fmt.Errorf("template file does not exist: %s", event.Path)
	}

	// 重新加载模板
	if err := hrm.configManager.LoadTemplate(event.Path); err != nil {
		return fmt.Errorf("failed to reload template: %w", err)
	}

	return nil
}

// reloadPlugin 重载插件
func (hrm *HotReloadManager) reloadPlugin(event *ReloadEvent) error {
	// 插件重载需要与插件管理器集成
	// 这里只是一个占位符实现
	fmt.Printf("Plugin reload not implemented: %s\n", event.Path)
	return nil
}

// reloadDirectory 重载目录
func (hrm *HotReloadManager) reloadDirectory(event *ReloadEvent) error {
	// 目录变化可能需要重新扫描配置文件
	// 这里只是一个占位符实现
	fmt.Printf("Directory reload not implemented: %s\n", event.Path)
	return nil
}

// handleReloadError 处理重载错误
func (hrm *HotReloadManager) handleReloadError(event *ReloadEvent, err error) {
	hrm.stats.FailedReloads++
	hrm.stats.LastError = err.Error()

	fmt.Printf("Hot reload failed for %s: %v\n", event.Path, err)

	// 重试机制
	if event.Retries < hrm.config.MaxRetries {
		event.Retries++
		time.AfterFunc(hrm.config.RetryDelay, func() {
			select {
			case hrm.eventChan <- event:
			default:
				fmt.Printf("Failed to retry reload for %s: event channel full\n", event.Path)
			}
		})
	}
}

// handleReloadSuccess 处理重载成功
func (hrm *HotReloadManager) handleReloadSuccess(event *ReloadEvent) {
	hrm.stats.SuccessfulReloads++
	hrm.stats.LastReloadTime = time.Now()

	fmt.Printf("Hot reload successful for %s\n", event.Path)
}

// GetStats 获取热更新统计信息
func (hrm *HotReloadManager) GetStats() *HotReloadStats {
	hrm.mutex.RLock()
	defer hrm.mutex.RUnlock()

	// 返回统计信息的副本
	stats := *hrm.stats
	stats.WatchedPaths = make([]string, len(hrm.stats.WatchedPaths))
	copy(stats.WatchedPaths, hrm.stats.WatchedPaths)

	return &stats
}

// UpdateConfig 更新热更新配置
func (hrm *HotReloadManager) UpdateConfig(config *HotReloadConfig) error {
	hrm.mutex.Lock()
	defer hrm.mutex.Unlock()

	wasRunning := hrm.running

	// 如果正在运行，先停止
	if wasRunning {
		if err := hrm.Stop(); err != nil {
			return fmt.Errorf("failed to stop hot reload manager: %w", err)
		}
	}

	// 更新配置
	hrm.config = config

	// 如果之前在运行，重新启动
	if wasRunning && config.Enabled {
		if err := hrm.Start(hrm.ctx); err != nil {
			return fmt.Errorf("failed to restart hot reload manager: %w", err)
		}
	}

	return nil
}