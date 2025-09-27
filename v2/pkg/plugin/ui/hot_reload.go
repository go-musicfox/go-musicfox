package ui

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// DefaultHotReloader 默认热重载器
type DefaultHotReloader struct {
	currentVersion string
	versionHistory []VersionEntry
	mutex          sync.RWMutex
	logger         *slog.Logger

	// 版本验证器
	versionValidator *VersionValidator

	// 重载处理器
	reloadHandlers map[string]ReloadHandler

	// 回滚管理
	rollbackManager *RollbackManager

	// 配置
	config *HotReloadConfig

	// 状态
	isReloading bool
	lastReload  time.Time
	maxHistory  int
}

// VersionEntry 版本条目
type VersionEntry struct {
	Version   string    `json:"version"`
	Hash      string    `json:"hash"`
	Data      []byte    `json:"data"`
	Timestamp time.Time `json:"timestamp"`
	Reason    string    `json:"reason"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
}

// VersionValidator 版本验证器
type VersionValidator struct {
	logger *slog.Logger
}

// ReloadHandler 重载处理器接口
type ReloadHandler interface {
	Handle(ctx context.Context, data []byte, version string) error
	GetType() string
	GetPriority() int
	CanHandle(data []byte) bool
}

// RollbackManager 回滚管理器
type RollbackManager struct {
	snapshots map[string]*Snapshot
	mutex     sync.RWMutex
	logger    *slog.Logger
}

// Snapshot 快照
type Snapshot struct {
	Version   string                 `json:"version"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Metadata  map[string]string      `json:"metadata"`
}

// HotReloadConfig 热重载配置
type HotReloadConfig struct {
	Enabled           bool          `json:"enabled"`
	MaxRetries        int           `json:"max_retries"`
	RetryDelay        time.Duration `json:"retry_delay"`
	ValidationTimeout time.Duration `json:"validation_timeout"`
	RollbackEnabled   bool          `json:"rollback_enabled"`
	MaxSnapshots      int           `json:"max_snapshots"`
	AllowedTypes      []string      `json:"allowed_types"`
}

// ReloadResult 重载结果
type ReloadResult struct {
	Success     bool          `json:"success"`
	Version     string        `json:"version"`
	OldVersion  string        `json:"old_version"`
	Duration    time.Duration `json:"duration"`
	Error       string        `json:"error,omitempty"`
	Changes     []string      `json:"changes"`
	RollbackID  string        `json:"rollback_id,omitempty"`
}

// NewDefaultHotReloader 创建默认热重载器
func NewDefaultHotReloader(logger *slog.Logger) *DefaultHotReloader {
	reloader := &DefaultHotReloader{
		currentVersion:   "1.0.0",
		versionHistory:   make([]VersionEntry, 0),
		logger:           logger,
		reloadHandlers:   make(map[string]ReloadHandler),
		maxHistory:       100,
		config: &HotReloadConfig{
			Enabled:           true,
			MaxRetries:        3,
			RetryDelay:        time.Second * 2,
			ValidationTimeout: time.Second * 10,
			RollbackEnabled:   true,
			MaxSnapshots:      10,
			AllowedTypes:      []string{"ui", "theme", "layout", "component"},
		},
	}

	// 初始化组件
	reloader.versionValidator = NewVersionValidator(logger)
	reloader.rollbackManager = NewRollbackManager(logger)

	// 注册默认处理器
	reloader.registerDefaultHandlers()

	return reloader
}

// registerDefaultHandlers 注册默认处理器
func (r *DefaultHotReloader) registerDefaultHandlers() {
	r.reloadHandlers["ui"] = NewUIReloadHandler(r.logger)
	r.reloadHandlers["theme"] = NewThemeReloadHandler(r.logger)
	r.reloadHandlers["layout"] = NewLayoutReloadHandler(r.logger)
	r.reloadHandlers["component"] = NewComponentReloadHandler(r.logger)
}

// Reload 热重载
func (r *DefaultHotReloader) Reload(ctx context.Context, data []byte) error {
	if !r.config.Enabled {
		return fmt.Errorf("hot reload is disabled")
	}

	r.mutex.Lock()
	if r.isReloading {
		r.mutex.Unlock()
		return fmt.Errorf("reload already in progress")
	}
	r.isReloading = true
	r.mutex.Unlock()

	defer func() {
		r.mutex.Lock()
		r.isReloading = false
		r.lastReload = time.Now()
		r.mutex.Unlock()
	}()

	start := time.Now()
	oldVersion := r.currentVersion
	newVersion := r.generateVersion(data)

	// 创建快照（如果启用回滚）
	var rollbackID string
	if r.config.RollbackEnabled {
		snapshot, err := r.createSnapshot(oldVersion)
		if err != nil {
			r.logger.Warn("Failed to create snapshot", "error", err)
		} else {
			rollbackID = snapshot.Version
		}
	}

	// 执行重载
	result := &ReloadResult{
		OldVersion: oldVersion,
		Version:    newVersion,
		RollbackID: rollbackID,
		Changes:    make([]string, 0),
	}

	err := r.performReload(ctx, data, newVersion, result)
	result.Duration = time.Since(start)
	result.Success = err == nil

	if err != nil {
		result.Error = err.Error()
		r.logger.Error("Hot reload failed", "error", err, "version", newVersion)

		// 尝试回滚
		if r.config.RollbackEnabled && rollbackID != "" {
			if rollbackErr := r.rollback(ctx, rollbackID); rollbackErr != nil {
				r.logger.Error("Rollback failed", "error", rollbackErr)
			} else {
				r.logger.Info("Rollback successful", "rollback_id", rollbackID)
			}
		}
	} else {
		r.currentVersion = newVersion
		r.logger.Info("Hot reload successful", "version", newVersion, "duration", result.Duration)
	}

	// 记录历史
	r.addToHistory(VersionEntry{
		Version:   newVersion,
		Hash:      r.calculateHash(data),
		Data:      data,
		Timestamp: time.Now(),
		Reason:    "hot_reload",
		Success:   result.Success,
		Error:     result.Error,
	})

	return err
}

// performReload 执行重载
func (r *DefaultHotReloader) performReload(ctx context.Context, data []byte, version string, result *ReloadResult) error {
	// 验证版本
	if err := r.versionValidator.Validate(version, data); err != nil {
		return fmt.Errorf("version validation failed: %w", err)
	}

	// 查找合适的处理器
	handler := r.findHandler(data)
	if handler == nil {
		return fmt.Errorf("no suitable handler found for data")
	}

	// 执行重载
	for attempt := 0; attempt < r.config.MaxRetries; attempt++ {
		if attempt > 0 {
			r.logger.Info("Retrying reload", "attempt", attempt+1, "max", r.config.MaxRetries)
			time.Sleep(r.config.RetryDelay)
		}

		// 创建超时上下文
		reloadCtx, cancel := context.WithTimeout(ctx, r.config.ValidationTimeout)
		err := handler.Handle(reloadCtx, data, version)
		cancel()

		if err == nil {
			result.Changes = append(result.Changes, fmt.Sprintf("Reloaded with %s handler", handler.GetType()))
			return nil
		}

		r.logger.Warn("Reload attempt failed", "attempt", attempt+1, "error", err)
		if attempt == r.config.MaxRetries-1 {
			return fmt.Errorf("reload failed after %d attempts: %w", r.config.MaxRetries, err)
		}
	}

	return fmt.Errorf("unexpected error in reload loop")
}

// findHandler 查找处理器
func (r *DefaultHotReloader) findHandler(data []byte) ReloadHandler {
	// 按优先级排序的处理器列表
	handlers := make([]ReloadHandler, 0, len(r.reloadHandlers))
	for _, handler := range r.reloadHandlers {
		if handler.CanHandle(data) {
			handlers = append(handlers, handler)
		}
	}

	// 按优先级排序
	for i := 0; i < len(handlers)-1; i++ {
		for j := i + 1; j < len(handlers); j++ {
			if handlers[i].GetPriority() < handlers[j].GetPriority() {
				handlers[i], handlers[j] = handlers[j], handlers[i]
			}
		}
	}

	if len(handlers) > 0 {
		return handlers[0]
	}

	return nil
}

// CanReload 是否可以重载
func (r *DefaultHotReloader) CanReload() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.config.Enabled && !r.isReloading
}

// GetVersion 获取版本
func (r *DefaultHotReloader) GetVersion() string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.currentVersion
}

// ValidateVersion 验证版本
func (r *DefaultHotReloader) ValidateVersion(version string) error {
	return r.versionValidator.Validate(version, nil)
}

// generateVersion 生成版本
func (r *DefaultHotReloader) generateVersion(data []byte) string {
	hash := r.calculateHash(data)
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%d-%s", timestamp, hash[:8])
}

// calculateHash 计算哈希
func (r *DefaultHotReloader) calculateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// createSnapshot 创建快照
func (r *DefaultHotReloader) createSnapshot(version string) (*Snapshot, error) {
	snapshot := &Snapshot{
		Version:   version,
		Timestamp: time.Now(),
		Data:      make(map[string]interface{}),
		Metadata:  make(map[string]string),
	}

	// 保存当前状态
	snapshot.Data["version"] = r.currentVersion
	snapshot.Metadata["type"] = "hot_reload_snapshot"

	return r.rollbackManager.SaveSnapshot(snapshot)
}

// rollback 回滚
func (r *DefaultHotReloader) rollback(ctx context.Context, rollbackID string) error {
	snapshot, err := r.rollbackManager.GetSnapshot(rollbackID)
	if err != nil {
		return fmt.Errorf("failed to get snapshot: %w", err)
	}

	// 恢复版本
	if version, ok := snapshot.Data["version"].(string); ok {
		r.currentVersion = version
	}

	r.logger.Info("Rollback completed", "rollback_id", rollbackID, "version", r.currentVersion)
	return nil
}

// addToHistory 添加到历史
func (r *DefaultHotReloader) addToHistory(entry VersionEntry) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.versionHistory = append(r.versionHistory, entry)

	// 限制历史长度
	if len(r.versionHistory) > r.maxHistory {
		r.versionHistory = r.versionHistory[1:]
	}
}

// GetVersionHistory 获取版本历史
func (r *DefaultHotReloader) GetVersionHistory() []VersionEntry {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	history := make([]VersionEntry, len(r.versionHistory))
	copy(history, r.versionHistory)
	return history
}

// RegisterHandler 注册处理器
func (r *DefaultHotReloader) RegisterHandler(handler ReloadHandler) {
	r.reloadHandlers[handler.GetType()] = handler
	r.logger.Info("Reload handler registered", "type", handler.GetType(), "priority", handler.GetPriority())
}

// NewVersionValidator 创建版本验证器
func NewVersionValidator(logger *slog.Logger) *VersionValidator {
	return &VersionValidator{
		logger: logger,
	}
}

// Validate 验证版本
func (v *VersionValidator) Validate(version string, data []byte) error {
	if version == "" {
		return fmt.Errorf("version cannot be empty")
	}

	// 基本格式验证
	if len(version) < 3 {
		return fmt.Errorf("version too short: %s", version)
	}

	// 如果有数据，验证数据完整性
	if data != nil && len(data) == 0 {
		return fmt.Errorf("data cannot be empty")
	}

	return nil
}

// NewRollbackManager 创建回滚管理器
func NewRollbackManager(logger *slog.Logger) *RollbackManager {
	return &RollbackManager{
		snapshots: make(map[string]*Snapshot),
		logger:    logger,
	}
}

// SaveSnapshot 保存快照
func (r *RollbackManager) SaveSnapshot(snapshot *Snapshot) (*Snapshot, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.snapshots[snapshot.Version] = snapshot
	r.logger.Info("Snapshot saved", "version", snapshot.Version)
	return snapshot, nil
}

// GetSnapshot 获取快照
func (r *RollbackManager) GetSnapshot(version string) (*Snapshot, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	snapshot, exists := r.snapshots[version]
	if !exists {
		return nil, fmt.Errorf("snapshot not found: %s", version)
	}

	return snapshot, nil
}

// ListSnapshots 列出快照
func (r *RollbackManager) ListSnapshots() []*Snapshot {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	snapshots := make([]*Snapshot, 0, len(r.snapshots))
	for _, snapshot := range r.snapshots {
		snapshots = append(snapshots, snapshot)
	}

	return snapshots
}

// 重载处理器实现

// UIReloadHandler UI重载处理器
type UIReloadHandler struct {
	logger *slog.Logger
}

// NewUIReloadHandler 创建UI重载处理器
func NewUIReloadHandler(logger *slog.Logger) *UIReloadHandler {
	return &UIReloadHandler{
		logger: logger,
	}
}

// Handle 处理重载
func (h *UIReloadHandler) Handle(ctx context.Context, data []byte, version string) error {
	h.logger.Info("Handling UI reload", "version", version, "size", len(data))
	// 实现UI重载逻辑
	return nil
}

// GetType 获取类型
func (h *UIReloadHandler) GetType() string {
	return "ui"
}

// GetPriority 获取优先级
func (h *UIReloadHandler) GetPriority() int {
	return 100
}

// CanHandle 是否可以处理
func (h *UIReloadHandler) CanHandle(data []byte) bool {
	// 简单检查是否包含UI相关内容
	content := string(data)
	return len(content) > 0 && (len(content) < 1024*1024) // 限制大小
}

// ThemeReloadHandler 主题重载处理器
type ThemeReloadHandler struct {
	logger *slog.Logger
}

// NewThemeReloadHandler 创建主题重载处理器
func NewThemeReloadHandler(logger *slog.Logger) *ThemeReloadHandler {
	return &ThemeReloadHandler{
		logger: logger,
	}
}

// Handle 处理重载
func (h *ThemeReloadHandler) Handle(ctx context.Context, data []byte, version string) error {
	h.logger.Info("Handling theme reload", "version", version, "size", len(data))
	// 实现主题重载逻辑
	return nil
}

// GetType 获取类型
func (h *ThemeReloadHandler) GetType() string {
	return "theme"
}

// GetPriority 获取优先级
func (h *ThemeReloadHandler) GetPriority() int {
	return 90
}

// CanHandle 是否可以处理
func (h *ThemeReloadHandler) CanHandle(data []byte) bool {
	content := string(data)
	return len(content) > 0 && (len(content) < 512*1024)
}

// LayoutReloadHandler 布局重载处理器
type LayoutReloadHandler struct {
	logger *slog.Logger
}

// NewLayoutReloadHandler 创建布局重载处理器
func NewLayoutReloadHandler(logger *slog.Logger) *LayoutReloadHandler {
	return &LayoutReloadHandler{
		logger: logger,
	}
}

// Handle 处理重载
func (h *LayoutReloadHandler) Handle(ctx context.Context, data []byte, version string) error {
	h.logger.Info("Handling layout reload", "version", version, "size", len(data))
	// 实现布局重载逻辑
	return nil
}

// GetType 获取类型
func (h *LayoutReloadHandler) GetType() string {
	return "layout"
}

// GetPriority 获取优先级
func (h *LayoutReloadHandler) GetPriority() int {
	return 80
}

// CanHandle 是否可以处理
func (h *LayoutReloadHandler) CanHandle(data []byte) bool {
	content := string(data)
	return len(content) > 0 && (len(content) < 256*1024)
}

// ComponentReloadHandler 组件重载处理器
type ComponentReloadHandler struct {
	logger *slog.Logger
}

// NewComponentReloadHandler 创建组件重载处理器
func NewComponentReloadHandler(logger *slog.Logger) *ComponentReloadHandler {
	return &ComponentReloadHandler{
		logger: logger,
	}
}

// Handle 处理重载
func (h *ComponentReloadHandler) Handle(ctx context.Context, data []byte, version string) error {
	h.logger.Info("Handling component reload", "version", version, "size", len(data))
	// 实现组件重载逻辑
	return nil
}

// GetType 获取类型
func (h *ComponentReloadHandler) GetType() string {
	return "component"
}

// GetPriority 获取优先级
func (h *ComponentReloadHandler) GetPriority() int {
	return 70
}

// CanHandle 是否可以处理
func (h *ComponentReloadHandler) CanHandle(data []byte) bool {
	content := string(data)
	return len(content) > 0 && (len(content) < 128*1024)
}