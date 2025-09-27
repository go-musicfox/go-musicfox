// Package loader provides plugin loading implementations for the go-musicfox microkernel architecture.
// This file implements the hot-reload plugin loader that supports dynamic plugin reloading,
// version management, and seamless state migration.
package loader

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// PluginVersion represents a version of a plugin
type PluginVersion struct {
	Version   string    `json:"version"`
	Path      string    `json:"path"`
	Content   []byte    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Checksum  string    `json:"checksum"`
}

// HotReloadConfig contains configuration for hot reload functionality
type HotReloadConfig struct {
	WatchInterval    time.Duration `json:"watch_interval"`
	MaxVersions      int           `json:"max_versions"`
	AutoReload       bool          `json:"auto_reload"`
	BackupEnabled    bool          `json:"backup_enabled"`
	StatePreservation bool         `json:"state_preservation"`
}

// LoadedHotReloadPlugin represents a loaded hot-reload plugin with its metadata
type LoadedHotReloadPlugin struct {
	Plugin        Plugin
	Info          *PluginInfo
	Path          string
	Versions      []PluginVersion
	CurrentVersion string
	AutoReload    bool
	Watcher       *fsnotify.Watcher
	State         map[string]interface{}
	LoadTime      time.Time
	LastReload    time.Time
}

// HotReloadPluginLoader implements PluginLoader interface for hot-reload plugins
type HotReloadPluginLoader struct {
	logger        *slog.Logger
	eventBus      EventBus
	config        HotReloadConfig
	plugins       map[string]*LoadedHotReloadPlugin
	mutex         sync.RWMutex
	watcher       *fsnotify.Watcher
	ctx           context.Context
	cancel        context.CancelFunc
	stopChan      chan struct{}
}





// Hot reload event constants
const (
	EventPluginHotReloaded   = "plugin.hot_reloaded"
	EventPluginVersionAdded  = "plugin.version_added"
	EventPluginRolledBack    = "plugin.rolled_back"
	EventPluginStatePreserved = "plugin.state_preserved"
)

// NewHotReloadPluginLoader creates a new hot-reload plugin loader
func NewHotReloadPluginLoader(eventBus EventBus, logger *slog.Logger) *HotReloadPluginLoader {
	ctx, cancel := context.WithCancel(context.Background())
	
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error("Failed to create file watcher", "error", err)
		watcher = nil
	}

	loader := &HotReloadPluginLoader{
		logger:   logger,
		eventBus: eventBus,
		config: HotReloadConfig{
			WatchInterval:     time.Second * 2,
			MaxVersions:       10,
			AutoReload:        true,
			BackupEnabled:     true,
			StatePreservation: true,
		},
		plugins:  make(map[string]*LoadedHotReloadPlugin),
		watcher:  watcher,
		ctx:      ctx,
		cancel:   cancel,
		stopChan: make(chan struct{}),
	}

	// Start file monitoring service
	if watcher != nil {
		go loader.startFileMonitoring()
	}

	return loader
}

// NewHotReloadPluginLoaderWithSecurity 创建新的热重载插件加载器（适配manager.go的签名）
func NewHotReloadPluginLoaderWithSecurity(securityManager SecurityManager, logger *slog.Logger) *HotReloadPluginLoader {
	// 创建一个简单的事件总线实现
	eventBus := &SimpleEventBus{handlers: make(map[string][]func(interface{}))}
	return NewHotReloadPluginLoader(eventBus, logger)
}

// SimpleEventBus 简单的事件总线实现
type SimpleEventBus struct {
	handlers map[string][]func(interface{})
	mutex    sync.RWMutex
}

func (s *SimpleEventBus) Publish(event string, data interface{}) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	if handlers, exists := s.handlers[event]; exists {
		for _, handler := range handlers {
			handler(data)
		}
	}
	return nil
}

func (s *SimpleEventBus) Subscribe(event string, handler func(interface{})) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.handlers[event] = append(s.handlers[event], handler)
	return nil
}

func (s *SimpleEventBus) Unsubscribe(event string, handler func(interface{})) error {
	// 简化实现，实际应该移除特定的handler
	return nil
}

// LoadPlugin loads a hot-reload plugin from the specified path
func (h *HotReloadPluginLoader) LoadPlugin(ctx context.Context, pluginPath string) (Plugin, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Check if plugin is already loaded
	pluginID := h.getPluginID(pluginPath)
	if _, exists := h.plugins[pluginID]; exists {
		return nil, fmt.Errorf("plugin already loaded: %s", pluginID)
	}

	// Read plugin content
	content, err := os.ReadFile(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin file: %w", err)
	}

	// Create plugin instance based on file extension
	pluginInstance, err := h.createPluginInstance(pluginPath, content)
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin instance: %w", err)
	}

	// Create initial version
	initialVersion := PluginVersion{
		Version:   "1.0.0",
		Path:      pluginPath,
		Content:   content,
		Timestamp: time.Now(),
		Checksum:  h.calculateChecksum(content),
	}

	// Create loaded plugin metadata
	loadedPlugin := &LoadedHotReloadPlugin{
		Plugin:         pluginInstance,
		Info:           pluginInstance.GetInfo(),
		Path:           pluginPath,
		Versions:       []PluginVersion{initialVersion},
		CurrentVersion: "1.0.0",
		AutoReload:     h.config.AutoReload,
		State:          make(map[string]interface{}),
		LoadTime:       time.Now(),
		LastReload:     time.Now(),
	}

	// Setup file watcher if available
	if h.watcher != nil {
		if err := h.watcher.Add(pluginPath); err != nil {
			h.logger.Warn("Failed to add file watcher", "path", pluginPath, "error", err)
		}
	}

	h.plugins[pluginID] = loadedPlugin

	h.logger.Info("Hot-reload plugin loaded successfully",
		"plugin_id", pluginID,
		"path", pluginPath,
		"version", initialVersion.Version)

	return pluginInstance, nil
}

// UnloadPlugin unloads a hot-reload plugin
func (h *HotReloadPluginLoader) UnloadPlugin(ctx context.Context, pluginID string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	loadedPlugin, exists := h.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	// Stop plugin if it has a Stop method
	if err := loadedPlugin.Plugin.Stop(); err != nil {
		h.logger.Warn("Failed to stop plugin gracefully", "plugin_id", pluginID, "error", err)
	}

	// Cleanup plugin resources
	if err := loadedPlugin.Plugin.Cleanup(); err != nil {
		h.logger.Warn("Failed to cleanup plugin resources", "plugin_id", pluginID, "error", err)
	}

	// Remove file watcher
	if h.watcher != nil {
		if err := h.watcher.Remove(loadedPlugin.Path); err != nil {
			h.logger.Warn("Failed to remove file watcher", "path", loadedPlugin.Path, "error", err)
		}
	}

	delete(h.plugins, pluginID)

	h.logger.Info("Hot-reload plugin unloaded successfully", "plugin_id", pluginID)
	return nil
}

// GetLoadedPlugins 获取已加载的插件列表（实现PluginLoader接口）
func (h *HotReloadPluginLoader) GetLoadedPlugins() map[string]Plugin {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	result := make(map[string]Plugin)
	for pluginID, loadedPlugin := range h.plugins {
		result[pluginID] = loadedPlugin.Plugin
	}
	return result
}

// IsPluginLoaded checks if a plugin is currently loaded
func (h *HotReloadPluginLoader) IsPluginLoaded(pluginID string) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	_, exists := h.plugins[pluginID]
	return exists
}

// GetPluginInfo returns information about a loaded plugin
func (h *HotReloadPluginLoader) GetPluginInfo(pluginID string) (*PluginInfo, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	loadedPlugin, exists := h.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	return loadedPlugin.Info, nil
}

// ReloadPlugin reloads a plugin with its current version
func (h *HotReloadPluginLoader) ReloadPlugin(ctx context.Context, pluginID string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	loadedPlugin, exists := h.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	// Read current plugin file
	content, err := os.ReadFile(loadedPlugin.Path)
	if err != nil {
		return fmt.Errorf("failed to read plugin file: %w", err)
	}

	// Check if content has changed
	newChecksum := h.calculateChecksum(content)
	currentVersion := h.getCurrentVersion(loadedPlugin)
	if currentVersion != nil && currentVersion.Checksum == newChecksum {
		h.logger.Debug("Plugin content unchanged, skipping reload", "plugin_id", pluginID)
		return nil
	}

	return h.performHotReload(pluginID, content)
}

// ValidatePlugin validates a plugin file before loading
func (h *HotReloadPluginLoader) ValidatePlugin(pluginPath string) error {
	// Check if file exists
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin file does not exist: %s", pluginPath)
	}

	// Check file extension
	ext := filepath.Ext(pluginPath)
	if !h.isSupportedExtension(ext) {
		return fmt.Errorf("unsupported plugin file extension: %s", ext)
	}

	// Read and validate content
	content, err := os.ReadFile(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to read plugin file: %w", err)
	}

	if len(content) == 0 {
		return fmt.Errorf("plugin file is empty: %s", pluginPath)
	}

	// Validate syntax based on file type
	return h.validatePluginSyntax(pluginPath, content)
}

// GetLoaderType 获取加载器类型（实现PluginLoader接口）
func (h *HotReloadPluginLoader) GetLoaderType() PluginType {
	return PluginTypeHotReload
}



// GetLoaderInfo 获取加载器信息（实现PluginLoader接口）
func (h *HotReloadPluginLoader) GetLoaderInfo() map[string]interface{} {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	return map[string]interface{}{
		"type":               "hotreload",
		"loaded_count":       len(h.plugins),
		"watch_interval":     h.config.WatchInterval,
		"max_versions":       h.config.MaxVersions,
		"auto_reload":        h.config.AutoReload,
		"backup_enabled":     h.config.BackupEnabled,
		"state_preservation": h.config.StatePreservation,
	}
}

// Shutdown 关闭加载器（实现PluginLoader接口）
func (h *HotReloadPluginLoader) Shutdown(ctx context.Context) error {
	return h.Cleanup()
}

// Cleanup cleans up the loader resources
func (h *HotReloadPluginLoader) Cleanup() error {
	// Stop file watcher
	h.mutex.Lock()
	if h.watcher != nil {
		if err := h.watcher.Close(); err != nil {
			h.logger.Warn("Failed to close file watcher", "error", err)
		}
		h.watcher = nil
	}
	h.mutex.Unlock()

	// Cancel context
	if h.cancel != nil {
		h.cancel()
		h.cancel = nil
	}

	// Close stop channel (only if not already closed)
	select {
	case <-h.stopChan:
		// Channel already closed
	default:
		close(h.stopChan)
	}

	// Cleanup all loaded plugins without holding the main lock
	h.mutex.Lock()
	pluginIDs := make([]string, 0, len(h.plugins))
	for pluginID := range h.plugins {
		pluginIDs = append(pluginIDs, pluginID)
	}
	h.mutex.Unlock()

	// Unload plugins one by one
	for _, pluginID := range pluginIDs {
		if err := h.UnloadPlugin(context.Background(), pluginID); err != nil {
			h.logger.Warn("Failed to unload plugin during cleanup", "plugin_id", pluginID, "error", err)
		}
	}

	h.logger.Info("Hot-reload plugin loader cleaned up successfully")
	return nil
}

// Hot-reload specific methods

// HotReload performs hot reload with new plugin version
func (h *HotReloadPluginLoader) HotReload(pluginID string, newVersion []byte) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	return h.performHotReload(pluginID, newVersion)
}

// EnableAutoReload enables automatic reloading for a plugin
func (h *HotReloadPluginLoader) EnableAutoReload(pluginID string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	loadedPlugin, exists := h.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	loadedPlugin.AutoReload = true
	h.logger.Info("Auto-reload enabled for plugin", "plugin_id", pluginID)
	return nil
}

// DisableAutoReload disables automatic reloading for a plugin
func (h *HotReloadPluginLoader) DisableAutoReload(pluginID string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	loadedPlugin, exists := h.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	loadedPlugin.AutoReload = false
	h.logger.Info("Auto-reload disabled for plugin", "plugin_id", pluginID)
	return nil
}

// GetVersionHistory returns the version history of a plugin
func (h *HotReloadPluginLoader) GetVersionHistory(pluginID string) []PluginVersion {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	loadedPlugin, exists := h.plugins[pluginID]
	if !exists {
		return nil
	}

	// Return a copy of the versions slice
	versions := make([]PluginVersion, len(loadedPlugin.Versions))
	copy(versions, loadedPlugin.Versions)
	return versions
}

// RollbackToVersion rolls back a plugin to a specific version
func (h *HotReloadPluginLoader) RollbackToVersion(pluginID string, version string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	loadedPlugin, exists := h.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	// Find the target version
	var targetVersion *PluginVersion
	for i := range loadedPlugin.Versions {
		if loadedPlugin.Versions[i].Version == version {
			targetVersion = &loadedPlugin.Versions[i]
			break
		}
	}

	if targetVersion == nil {
		return fmt.Errorf("version not found: %s", version)
	}

	// Perform rollback
	if err := h.performHotReload(pluginID, targetVersion.Content); err != nil {
		return fmt.Errorf("failed to rollback to version %s: %w", version, err)
	}

	loadedPlugin.CurrentVersion = version

	// Publish rollback event
	if h.eventBus != nil {
		h.eventBus.Publish(EventPluginRolledBack, map[string]interface{}{
			"plugin_id": pluginID,
			"version":   version,
			"timestamp": time.Now(),
		})
	}

	h.logger.Info("Plugin rolled back successfully",
		"plugin_id", pluginID,
		"version", version)

	return nil
}

// Private helper methods

// getPluginID generates a unique plugin ID from the plugin path
func (h *HotReloadPluginLoader) getPluginID(pluginPath string) string {
	return filepath.Base(pluginPath)
}

// createPluginInstance creates a plugin instance based on file type
func (h *HotReloadPluginLoader) createPluginInstance(pluginPath string, content []byte) (Plugin, error) {
	ext := filepath.Ext(pluginPath)
	
	switch ext {
	case ".js", ".ts":
		return h.createJSPlugin(pluginPath, content)
	case ".go":
		return h.createGoPlugin(pluginPath, content)
	default:
		return nil, fmt.Errorf("unsupported plugin type: %s", ext)
	}
}

// createJSPlugin creates a JavaScript/TypeScript plugin instance
func (h *HotReloadPluginLoader) createJSPlugin(pluginPath string, content []byte) (Plugin, error) {
	// This is a simplified implementation
	// In a real implementation, you would use a JavaScript engine like V8 or Otto
	return &JSPlugin{
		path:    pluginPath,
		content: content,
		logger:  h.logger,
	}, nil
}

// createGoPlugin creates a Go plugin instance (for hot-reloadable Go plugins)
func (h *HotReloadPluginLoader) createGoPlugin(pluginPath string, content []byte) (Plugin, error) {
	// This would require compilation and dynamic loading
	// For now, return an error as this is complex to implement
	return nil, fmt.Errorf("Go plugin hot-reload not yet implemented")
}

// calculateChecksum calculates a checksum for plugin content
func (h *HotReloadPluginLoader) calculateChecksum(content []byte) string {
	// Simple checksum implementation
	// In production, use a proper hash function like SHA256
	sum := 0
	for _, b := range content {
		sum += int(b)
	}
	return fmt.Sprintf("%x", sum)
}

// getCurrentVersion returns the current version of a plugin
func (h *HotReloadPluginLoader) getCurrentVersion(loadedPlugin *LoadedHotReloadPlugin) *PluginVersion {
	for i := range loadedPlugin.Versions {
		if loadedPlugin.Versions[i].Version == loadedPlugin.CurrentVersion {
			return &loadedPlugin.Versions[i]
		}
	}
	return nil
}

// isSupportedExtension checks if the file extension is supported
func (h *HotReloadPluginLoader) isSupportedExtension(ext string) bool {
	supportedExts := []string{".js", ".ts", ".go"}
	for _, supportedExt := range supportedExts {
		if ext == supportedExt {
			return true
		}
	}
	return false
}

// validatePluginSyntax validates plugin syntax based on file type
func (h *HotReloadPluginLoader) validatePluginSyntax(pluginPath string, content []byte) error {
	ext := filepath.Ext(pluginPath)
	
	switch ext {
	case ".js", ".ts":
		return h.validateJSSyntax(content)
	case ".go":
		return h.validateGoSyntax(content)
	default:
		return nil
	}
}

// validateJSSyntax validates JavaScript/TypeScript syntax
func (h *HotReloadPluginLoader) validateJSSyntax(content []byte) error {
	// Basic validation - check if it's valid JSON-like structure
	// In a real implementation, use a proper JS parser
	if len(content) < 10 {
		return fmt.Errorf("plugin content too short")
	}
	return nil
}

// validateGoSyntax validates Go syntax
func (h *HotReloadPluginLoader) validateGoSyntax(content []byte) error {
	// Basic validation - check for package declaration
	contentStr := string(content)
	if len(contentStr) < 10 || contentStr[:7] != "package" {
		return fmt.Errorf("invalid Go plugin: missing package declaration")
	}
	return nil
}

// performHotReload performs the actual hot reload operation
func (h *HotReloadPluginLoader) performHotReload(pluginID string, newContent []byte) error {
	loadedPlugin, exists := h.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	// Preserve current state if enabled
	var preservedState map[string]interface{}
	if h.config.StatePreservation {
		preservedState = h.preservePluginState(loadedPlugin.Plugin)
	}

	// Stop current plugin
	if err := loadedPlugin.Plugin.Stop(); err != nil {
		h.logger.Warn("Failed to stop plugin gracefully during hot reload", "plugin_id", pluginID, "error", err)
	}

	// Create new plugin instance
	newPlugin, err := h.createPluginInstance(loadedPlugin.Path, newContent)
	if err != nil {
		return fmt.Errorf("failed to create new plugin instance: %w", err)
	}

	// Initialize new plugin
	if err := newPlugin.Initialize(nil); err != nil {
		return fmt.Errorf("failed to initialize new plugin: %w", err)
	}

	// Restore state if preserved
	if h.config.StatePreservation && preservedState != nil {
		h.restorePluginState(newPlugin, preservedState)
	}

	// Start new plugin
	if err := newPlugin.Start(); err != nil {
		return fmt.Errorf("failed to start new plugin: %w", err)
	}

	// Create new version entry
	newVersion := PluginVersion{
		Version:   h.generateNextVersion(loadedPlugin),
		Path:      loadedPlugin.Path,
		Content:   newContent,
		Timestamp: time.Now(),
		Checksum:  h.calculateChecksum(newContent),
	}

	// Update loaded plugin
	loadedPlugin.Plugin = newPlugin
	loadedPlugin.Info = newPlugin.GetInfo()
	loadedPlugin.Versions = append(loadedPlugin.Versions, newVersion)
	loadedPlugin.CurrentVersion = newVersion.Version
	loadedPlugin.LastReload = time.Now()

	// Limit version history
	if len(loadedPlugin.Versions) > h.config.MaxVersions {
		loadedPlugin.Versions = loadedPlugin.Versions[1:]
	}

	// Publish hot reload event
	if h.eventBus != nil {
		h.eventBus.Publish(EventPluginHotReloaded, map[string]interface{}{
			"plugin_id": pluginID,
			"version":   newVersion.Version,
			"timestamp": time.Now(),
		})
	}

	h.logger.Info("Plugin hot reloaded successfully",
		"plugin_id", pluginID,
		"new_version", newVersion.Version)

	return nil
}

// preservePluginState preserves the current state of a plugin
func (h *HotReloadPluginLoader) preservePluginState(plugin Plugin) map[string]interface{} {
	// This is a simplified implementation
	// In a real implementation, you would need to define a state preservation interface
	state := make(map[string]interface{})
	
	// Try to get state if plugin supports it
	if stateProvider, ok := plugin.(interface{ GetState() map[string]interface{} }); ok {
		state = stateProvider.GetState()
	}
	
	return state
}

// restorePluginState restores the preserved state to a plugin
func (h *HotReloadPluginLoader) restorePluginState(plugin Plugin, state map[string]interface{}) {
	// Try to restore state if plugin supports it
	if stateReceiver, ok := plugin.(interface{ SetState(map[string]interface{}) error }); ok {
		if err := stateReceiver.SetState(state); err != nil {
			h.logger.Warn("Failed to restore plugin state", "error", err)
		}
	}
}

// generateNextVersion generates the next version number
func (h *HotReloadPluginLoader) generateNextVersion(loadedPlugin *LoadedHotReloadPlugin) string {
	// Simple version increment
	versionNum := len(loadedPlugin.Versions) + 1
	return fmt.Sprintf("1.%d.0", versionNum)
}

// startFileMonitoring starts the file monitoring service
func (h *HotReloadPluginLoader) startFileMonitoring() {
	if h.logger != nil {
		h.logger.Info("Starting file monitoring service")
	}

	// Check if watcher is available
	h.mutex.RLock()
	watcher := h.watcher
	h.mutex.RUnlock()
	
	if watcher == nil {
		if h.logger != nil {
			h.logger.Warn("File watcher not available, monitoring disabled")
		}
		return
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				if h.logger != nil {
					h.logger.Info("File watcher events channel closed")
				}
				return
			}
			h.handleFileEvent(event)

		case err, ok := <-watcher.Errors:
			if !ok {
				if h.logger != nil {
					h.logger.Info("File watcher errors channel closed")
				}
				return
			}
			if h.logger != nil {
				h.logger.Error("File watcher error", "error", err)
			}

		case <-h.ctx.Done():
			if h.logger != nil {
				h.logger.Info("File monitoring service stopped")
			}
			return

		case <-h.stopChan:
			if h.logger != nil {
				h.logger.Info("File monitoring service stopped via stop channel")
			}
			return
		}
	}
}

// handleFileEvent handles file system events
func (h *HotReloadPluginLoader) handleFileEvent(event fsnotify.Event) {
	h.logger.Debug("File event received", "event", event.String())

	// Only handle write and create events
	if !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) {
		return
	}

	// Find the plugin associated with this file
	pluginID := h.getPluginID(event.Name)
	
	h.mutex.RLock()
	loadedPlugin, exists := h.plugins[pluginID]
	h.mutex.RUnlock()

	if !exists {
		h.logger.Debug("File event for unloaded plugin", "path", event.Name)
		return
	}

	// Check if auto-reload is enabled for this plugin
	if !loadedPlugin.AutoReload {
		h.logger.Debug("Auto-reload disabled for plugin", "plugin_id", pluginID)
		return
	}

	// Add a small delay to avoid multiple rapid events
	time.Sleep(h.config.WatchInterval)

	// Perform auto-reload
	if err := h.ReloadPlugin(h.ctx, pluginID); err != nil {
		h.logger.Error("Auto-reload failed", "plugin_id", pluginID, "error", err)
		
		// Publish auto-reload failure event
		if h.eventBus != nil {
			h.eventBus.Publish("plugin.auto_reload_failed", map[string]interface{}{
				"plugin_id": pluginID,
				"error":     err.Error(),
				"timestamp": time.Now(),
			})
		}
	} else {
		h.logger.Info("Auto-reload successful", "plugin_id", pluginID)
		
		// Publish auto-reload success event
		if h.eventBus != nil {
			h.eventBus.Publish("plugin.auto_reload_success", map[string]interface{}{
				"plugin_id": pluginID,
				"timestamp": time.Now(),
			})
		}
	}
}

// SetConfig updates the hot-reload configuration
func (h *HotReloadPluginLoader) SetConfig(config HotReloadConfig) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	h.config = config
	h.logger.Info("Hot-reload configuration updated", "config", config)
}

// GetConfig returns the current hot-reload configuration
func (h *HotReloadPluginLoader) GetConfig() HotReloadConfig {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	return h.config
}

// GetPluginStats returns statistics for a loaded plugin
func (h *HotReloadPluginLoader) GetPluginStats(pluginID string) (*PluginStats, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	loadedPlugin, exists := h.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	return &PluginStats{
		PluginID:       pluginID,
		LoadTime:       loadedPlugin.LoadTime,
		LastReload:     loadedPlugin.LastReload,
		VersionCount:   len(loadedPlugin.Versions),
		CurrentVersion: loadedPlugin.CurrentVersion,
		AutoReload:     loadedPlugin.AutoReload,
		Path:           loadedPlugin.Path,
	}, nil
}

// PluginStats contains statistics about a loaded plugin
type PluginStats struct {
	PluginID       string    `json:"plugin_id"`
	LoadTime       time.Time `json:"load_time"`
	LastReload     time.Time `json:"last_reload"`
	VersionCount   int       `json:"version_count"`
	CurrentVersion string    `json:"current_version"`
	AutoReload     bool      `json:"auto_reload"`
	Path           string    `json:"path"`
}

// JSPlugin represents a JavaScript/TypeScript plugin
type JSPlugin struct {
	path    string
	content []byte
	logger  *slog.Logger
	info    *PluginInfo
}

// Implement Plugin interface for JSPlugin
func (js *JSPlugin) GetInfo() *PluginInfo {
	if js.info == nil {
		js.info = &PluginInfo{
			Name:        filepath.Base(js.path),
			Version:     "1.0.0",
			Description: "JavaScript/TypeScript Hot-Reload Plugin",
			Author:      "Hot-Reload System",
			LoadTime:    time.Now(),
		}
	}
	return js.info
}

func (js *JSPlugin) GetCapabilities() []string {
	return []string{"hot-reload", "javascript", "typescript"}
}

func (js *JSPlugin) GetDependencies() []string {
	return []string{}
}

func (js *JSPlugin) Initialize(ctx PluginContext) error {
	js.logger.Info("Initializing JS plugin", "path", js.path)
	return nil
}

func (js *JSPlugin) Start() error {
	js.logger.Info("Starting JS plugin", "path", js.path)
	return nil
}

func (js *JSPlugin) Stop() error {
	js.logger.Info("Stopping JS plugin", "path", js.path)
	return nil
}

func (js *JSPlugin) Cleanup() error {
	js.logger.Info("Cleaning up JS plugin", "path", js.path)
	return nil
}

func (js *JSPlugin) HealthCheck() error {
	return nil
}

func (js *JSPlugin) ValidateConfig(config map[string]interface{}) error {
	return nil
}

func (js *JSPlugin) UpdateConfig(config map[string]interface{}) error {
	return nil
}

func (js *JSPlugin) HandleEvent(event interface{}) error {
	js.logger.Debug("JS plugin handling event", "event", event)
	return nil
}

func (js *JSPlugin) GetMetrics() (*PluginMetrics, error) {
	return &PluginMetrics{
		PluginID:      filepath.Base(js.path),
		Uptime:        time.Since(time.Now()),
		MemoryUsage:   int64(len(js.content)),
		CPUUsage:      0.0,
		RequestCount:  0,
		ErrorCount:    0,
		SuccessRate:   100.0,
		CustomMetrics: map[string]interface{}{
			"type": "javascript",
			"path": js.path,
			"size": len(js.content),
		},
		Timestamp: time.Now(),
	}, nil
}