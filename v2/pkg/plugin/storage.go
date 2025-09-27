package plugin

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
	"log/slog"
)

// StorageFormat 存储格式
type StorageFormat string

const (
	StorageFormatJSON StorageFormat = "json"
	StorageFormatYAML StorageFormat = "yaml"
	StorageFormatTOML StorageFormat = "toml"
)

// String 返回存储格式的字符串表示
func (sf StorageFormat) String() string {
	return string(sf)
}

// ConfigVersion 配置版本信息
type ConfigVersion struct {
	Version     int       `json:"version" yaml:"version"`
	Timestamp   time.Time `json:"timestamp" yaml:"timestamp"`
	Checksum    string    `json:"checksum" yaml:"checksum"`
	Description string    `json:"description,omitempty" yaml:"description,omitempty"`
	Author      string    `json:"author,omitempty" yaml:"author,omitempty"`
	Tags        []string  `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// ConfigMetadata 配置元数据
type ConfigMetadata struct {
	PluginID        string           `json:"plugin_id" yaml:"plugin_id"`
	CurrentVersion  int              `json:"current_version" yaml:"current_version"`
	Versions        []ConfigVersion  `json:"versions" yaml:"versions"`
	StorageFormat   StorageFormat    `json:"storage_format" yaml:"storage_format"`
	CreatedAt       time.Time        `json:"created_at" yaml:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at" yaml:"updated_at"`
	MaxVersions     int              `json:"max_versions" yaml:"max_versions"`
	CompressionType string           `json:"compression_type,omitempty" yaml:"compression_type,omitempty"`
	Encrypted       bool             `json:"encrypted,omitempty" yaml:"encrypted,omitempty"`
	CustomFields    map[string]interface{} `json:"custom_fields,omitempty" yaml:"custom_fields,omitempty"`
}

// Validate 验证配置元数据
func (cm *ConfigMetadata) Validate() error {
	if cm.PluginID == "" {
		return fmt.Errorf("plugin ID cannot be empty")
	}

	if cm.CurrentVersion < 0 {
		return fmt.Errorf("current version cannot be negative")
	}

	if cm.MaxVersions <= 0 {
		return fmt.Errorf("max versions must be positive")
	}

	validFormats := []StorageFormat{StorageFormatJSON, StorageFormatYAML, StorageFormatTOML}
	validFormat := false
	for _, format := range validFormats {
		if cm.StorageFormat == format {
			validFormat = true
			break
		}
	}
	if !validFormat {
		return fmt.Errorf("invalid storage format: %s", cm.StorageFormat)
	}

	return nil
}

// AddVersion 添加新版本
func (cm *ConfigMetadata) AddVersion(description, author string, tags []string, checksum string) {
	newVersion := ConfigVersion{
		Version:     cm.CurrentVersion + 1,
		Timestamp:   time.Now(),
		Checksum:    checksum,
		Description: description,
		Author:      author,
		Tags:        tags,
	}

	cm.Versions = append(cm.Versions, newVersion)
	cm.CurrentVersion = newVersion.Version
	cm.UpdatedAt = time.Now()

	// 清理旧版本
	if len(cm.Versions) > cm.MaxVersions {
		cm.Versions = cm.Versions[len(cm.Versions)-cm.MaxVersions:]
	}
}

// GetVersion 获取指定版本信息
func (cm *ConfigMetadata) GetVersion(version int) *ConfigVersion {
	for i := range cm.Versions {
		if cm.Versions[i].Version == version {
			return &cm.Versions[i]
		}
	}
	return nil
}

// GetLatestVersion 获取最新版本信息
func (cm *ConfigMetadata) GetLatestVersion() *ConfigVersion {
	if len(cm.Versions) == 0 {
		return nil
	}
	return &cm.Versions[len(cm.Versions)-1]
}

// ConfigStorage 配置存储接口
type ConfigStorage interface {
	// 基础操作
	Save(ctx context.Context, pluginID string, config *EnhancedPluginConfig) error
	Load(ctx context.Context, pluginID string) (*EnhancedPluginConfig, error)
	Delete(ctx context.Context, pluginID string) error
	Exists(ctx context.Context, pluginID string) (bool, error)

	// 版本管理
	SaveVersion(ctx context.Context, pluginID string, config *EnhancedPluginConfig, description, author string, tags []string) error
	LoadVersion(ctx context.Context, pluginID string, version int) (*EnhancedPluginConfig, error)
	ListVersions(ctx context.Context, pluginID string) ([]ConfigVersion, error)
	DeleteVersion(ctx context.Context, pluginID string, version int) error

	// 备份和恢复
	Backup(ctx context.Context, pluginID string, backupPath string) error
	Restore(ctx context.Context, pluginID string, backupPath string) error
	ListBackups(ctx context.Context, pluginID string) ([]string, error)

	// 批量操作
	ListConfigs(ctx context.Context) ([]string, error)
	BatchSave(ctx context.Context, configs map[string]*EnhancedPluginConfig) error
	BatchLoad(ctx context.Context, pluginIDs []string) (map[string]*EnhancedPluginConfig, error)

	// 同步操作
	Sync(ctx context.Context) error
	GetMetadata(ctx context.Context, pluginID string) (*ConfigMetadata, error)
}

// FileConfigStorage 文件系统配置存储实现
type FileConfigStorage struct {
	logger      *slog.Logger
	basePath    string
	format      StorageFormat
	maxVersions int
	mu          sync.RWMutex
	metadata    map[string]*ConfigMetadata
	encryption  ConfigEncryption
	compression ConfigCompression
}

// FileConfigStorageOptions 文件配置存储选项
type FileConfigStorageOptions struct {
	BasePath    string
	Format      StorageFormat
	MaxVersions int
	Encryption  ConfigEncryption
	Compression ConfigCompression
}

// DefaultFileConfigStorageOptions 默认文件配置存储选项
func DefaultFileConfigStorageOptions() *FileConfigStorageOptions {
	return &FileConfigStorageOptions{
		BasePath:    "./configs",
		Format:      StorageFormatYAML,
		MaxVersions: 10,
		Encryption:  nil,
		Compression: nil,
	}
}

// NewFileConfigStorage 创建文件配置存储
func NewFileConfigStorage(logger *slog.Logger, options *FileConfigStorageOptions) *FileConfigStorage {
	if options == nil {
		options = DefaultFileConfigStorageOptions()
	}

	return &FileConfigStorage{
		logger:      logger,
		basePath:    options.BasePath,
		format:      options.Format,
		maxVersions: options.MaxVersions,
		metadata:    make(map[string]*ConfigMetadata),
		encryption:  options.Encryption,
		compression: options.Compression,
	}
}

// Initialize 初始化存储
func (fcs *FileConfigStorage) Initialize(ctx context.Context) error {
	// 创建基础目录
	if err := os.MkdirAll(fcs.basePath, 0755); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}

	// 创建子目录
	dirs := []string{"configs", "versions", "backups", "metadata"}
	for _, dir := range dirs {
		dirPath := filepath.Join(fcs.basePath, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// 加载元数据
	if err := fcs.loadAllMetadata(ctx); err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	fcs.logger.Info("File config storage initialized", "base_path", fcs.basePath)
	return nil
}

// Save 保存配置
func (fcs *FileConfigStorage) Save(ctx context.Context, pluginID string, config *EnhancedPluginConfig) error {
	fcs.mu.Lock()
	defer fcs.mu.Unlock()

	return fcs.saveConfig(ctx, pluginID, config, "", "", nil)
}

// Load 加载配置
func (fcs *FileConfigStorage) Load(ctx context.Context, pluginID string) (*EnhancedPluginConfig, error) {
	fcs.mu.RLock()
	defer fcs.mu.RUnlock()

	configPath := fcs.getConfigPath(pluginID)
	return fcs.loadConfigFromFile(configPath)
}

// Delete 删除配置
func (fcs *FileConfigStorage) Delete(ctx context.Context, pluginID string) error {
	fcs.mu.Lock()
	defer fcs.mu.Unlock()

	// 删除配置文件
	configPath := fcs.getConfigPath(pluginID)
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete config file: %w", err)
	}

	// 删除版本文件
	versionsDir := fcs.getVersionsDir(pluginID)
	if err := os.RemoveAll(versionsDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete versions directory: %w", err)
	}

	// 删除备份文件
	backupsDir := fcs.getBackupsDir(pluginID)
	if err := os.RemoveAll(backupsDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete backups directory: %w", err)
	}

	// 删除元数据
	metadataPath := fcs.getMetadataPath(pluginID)
	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete metadata file: %w", err)
	}

	delete(fcs.metadata, pluginID)

	fcs.logger.Info("Config deleted", "plugin_id", pluginID)
	return nil
}

// Exists 检查配置是否存在
func (fcs *FileConfigStorage) Exists(ctx context.Context, pluginID string) (bool, error) {
	configPath := fcs.getConfigPath(pluginID)
	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check config existence: %w", err)
	}
	return true, nil
}

// SaveVersion 保存配置版本
func (fcs *FileConfigStorage) SaveVersion(ctx context.Context, pluginID string, config *EnhancedPluginConfig, description, author string, tags []string) error {
	fcs.mu.Lock()
	defer fcs.mu.Unlock()

	return fcs.saveConfig(ctx, pluginID, config, description, author, tags)
}

// LoadVersion 加载指定版本的配置
func (fcs *FileConfigStorage) LoadVersion(ctx context.Context, pluginID string, version int) (*EnhancedPluginConfig, error) {
	fcs.mu.RLock()
	defer fcs.mu.RUnlock()

	versionPath := fcs.getVersionPath(pluginID, version)
	return fcs.loadConfigFromFile(versionPath)
}

// ListVersions 列出所有版本
func (fcs *FileConfigStorage) ListVersions(ctx context.Context, pluginID string) ([]ConfigVersion, error) {
	fcs.mu.RLock()
	defer fcs.mu.RUnlock()

	metadata, exists := fcs.metadata[pluginID]
	if !exists {
		return nil, fmt.Errorf("metadata not found for plugin: %s", pluginID)
	}

	return metadata.Versions, nil
}

// DeleteVersion 删除指定版本
func (fcs *FileConfigStorage) DeleteVersion(ctx context.Context, pluginID string, version int) error {
	fcs.mu.Lock()
	defer fcs.mu.Unlock()

	// 删除版本文件
	versionPath := fcs.getVersionPath(pluginID, version)
	if err := os.Remove(versionPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete version file: %w", err)
	}

	// 更新元数据
	metadata, exists := fcs.metadata[pluginID]
	if exists {
		newVersions := make([]ConfigVersion, 0)
		for _, v := range metadata.Versions {
			if v.Version != version {
				newVersions = append(newVersions, v)
			}
		}
		metadata.Versions = newVersions
		metadata.UpdatedAt = time.Now()

		if err := fcs.saveMetadata(pluginID, metadata); err != nil {
			return fmt.Errorf("failed to update metadata: %w", err)
		}
	}

	fcs.logger.Info("Config version deleted", "plugin_id", pluginID, "version", version)
	return nil
}

// Backup 备份配置
func (fcs *FileConfigStorage) Backup(ctx context.Context, pluginID string, backupPath string) error {
	fcs.mu.RLock()
	defer fcs.mu.RUnlock()

	// 如果没有指定备份路径，使用默认路径
	if backupPath == "" {
		timestamp := time.Now().Format("20060102_150405")
		backupPath = filepath.Join(fcs.getBackupsDir(pluginID), fmt.Sprintf("backup_%s.tar.gz", timestamp))
	}

	// 创建备份目录
	backupDir := filepath.Dir(backupPath)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// 创建备份文件
	backupFile, err := os.Create(backupPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer backupFile.Close()

	// 备份配置文件
	configPath := fcs.getConfigPath(pluginID)
	if err := fcs.copyFile(configPath, backupFile); err != nil {
		return fmt.Errorf("failed to backup config file: %w", err)
	}

	// 备份元数据
	metadataPath := fcs.getMetadataPath(pluginID)
	if err := fcs.copyFile(metadataPath, backupFile); err != nil {
		fcs.logger.Warn("Failed to backup metadata", "plugin_id", pluginID, "error", err)
	}

	fcs.logger.Info("Config backed up", "plugin_id", pluginID, "backup_path", backupPath)
	return nil
}

// Restore 恢复配置
func (fcs *FileConfigStorage) Restore(ctx context.Context, pluginID string, backupPath string) error {
	fcs.mu.Lock()
	defer fcs.mu.Unlock()

	// 检查备份文件是否存在
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backupPath)
	}

	// 创建当前配置的备份
	currentBackupPath := filepath.Join(fcs.getBackupsDir(pluginID), fmt.Sprintf("pre_restore_%s.backup", time.Now().Format("20060102_150405")))
	if err := fcs.Backup(ctx, pluginID, currentBackupPath); err != nil {
		fcs.logger.Warn("Failed to create pre-restore backup", "plugin_id", pluginID, "error", err)
	}

	// 恢复配置文件
	configPath := fcs.getConfigPath(pluginID)
	if err := fcs.copyFileReverse(backupPath, configPath); err != nil {
		return fmt.Errorf("failed to restore config file: %w", err)
	}

	// 重新加载元数据
	if err := fcs.loadMetadata(pluginID); err != nil {
		fcs.logger.Warn("Failed to reload metadata after restore", "plugin_id", pluginID, "error", err)
	}

	fcs.logger.Info("Config restored", "plugin_id", pluginID, "backup_path", backupPath)
	return nil
}

// ListBackups 列出备份文件
func (fcs *FileConfigStorage) ListBackups(ctx context.Context, pluginID string) ([]string, error) {
	backupsDir := fcs.getBackupsDir(pluginID)

	files, err := os.ReadDir(backupsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read backups directory: %w", err)
	}

	backups := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {
			backups = append(backups, file.Name())
		}
	}

	sort.Strings(backups)
	return backups, nil
}

// ListConfigs 列出所有配置
func (fcs *FileConfigStorage) ListConfigs(ctx context.Context) ([]string, error) {
	configsDir := filepath.Join(fcs.basePath, "configs")

	files, err := os.ReadDir(configsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read configs directory: %w", err)
	}

	configs := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {
			name := file.Name()
			// 移除文件扩展名
			if ext := filepath.Ext(name); ext != "" {
				name = strings.TrimSuffix(name, ext)
			}
			configs = append(configs, name)
		}
	}

	sort.Strings(configs)
	return configs, nil
}

// BatchSave 批量保存配置
func (fcs *FileConfigStorage) BatchSave(ctx context.Context, configs map[string]*EnhancedPluginConfig) error {
	fcs.mu.Lock()
	defer fcs.mu.Unlock()

	for pluginID, config := range configs {
		if err := fcs.saveConfig(ctx, pluginID, config, "", "", nil); err != nil {
			return fmt.Errorf("failed to save config for plugin %s: %w", pluginID, err)
		}
	}

	fcs.logger.Info("Batch save completed", "count", len(configs))
	return nil
}

// BatchLoad 批量加载配置
func (fcs *FileConfigStorage) BatchLoad(ctx context.Context, pluginIDs []string) (map[string]*EnhancedPluginConfig, error) {
	fcs.mu.RLock()
	defer fcs.mu.RUnlock()

	configs := make(map[string]*EnhancedPluginConfig)

	for _, pluginID := range pluginIDs {
		configPath := fcs.getConfigPath(pluginID)
		config, err := fcs.loadConfigFromFile(configPath)
		if err != nil {
			fcs.logger.Warn("Failed to load config in batch", "plugin_id", pluginID, "error", err)
			continue
		}
		configs[pluginID] = config
	}

	fcs.logger.Info("Batch load completed", "requested", len(pluginIDs), "loaded", len(configs))
	return configs, nil
}

// Sync 同步操作
func (fcs *FileConfigStorage) Sync(ctx context.Context) error {
	fcs.mu.Lock()
	defer fcs.mu.Unlock()

	// 重新加载所有元数据
	if err := fcs.loadAllMetadata(ctx); err != nil {
		return fmt.Errorf("failed to reload metadata: %w", err)
	}

	fcs.logger.Info("Storage synced")
	return nil
}

// GetMetadata 获取元数据
func (fcs *FileConfigStorage) GetMetadata(ctx context.Context, pluginID string) (*ConfigMetadata, error) {
	fcs.mu.RLock()
	defer fcs.mu.RUnlock()

	metadata, exists := fcs.metadata[pluginID]
	if !exists {
		return nil, fmt.Errorf("metadata not found for plugin: %s", pluginID)
	}

	// 返回副本
	copy := *metadata
	return &copy, nil
}

// saveConfig 保存配置（内部方法）
func (fcs *FileConfigStorage) saveConfig(ctx context.Context, pluginID string, config *EnhancedPluginConfig, description, author string, tags []string) error {
	// 序列化配置
	data, err := fcs.serializeConfig(config)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	// 应用压缩
	if fcs.compression != nil {
		data, err = fcs.compression.Compress(data)
		if err != nil {
			return fmt.Errorf("failed to compress config: %w", err)
		}
	}

	// 应用加密
	if fcs.encryption != nil {
		data, err = fcs.encryption.Encrypt(data)
		if err != nil {
			return fmt.Errorf("failed to encrypt config: %w", err)
		}
	}

	// 计算校验和
	checksum := fcs.calculateChecksum(data)

	// 保存到主配置文件
	configPath := fcs.getConfigPath(pluginID)
	if err := fcs.writeFile(configPath, data); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// 更新元数据
	metadata := fcs.getOrCreateMetadata(pluginID)
	if description != "" || author != "" || len(tags) > 0 {
		// 保存版本
		versionPath := fcs.getVersionPath(pluginID, metadata.CurrentVersion+1)
		if err := fcs.writeFile(versionPath, data); err != nil {
			return fmt.Errorf("failed to write version file: %w", err)
		}

		metadata.AddVersion(description, author, tags, checksum)
	} else {
		metadata.UpdatedAt = time.Now()
	}

	// 保存元数据
	if err := fcs.saveMetadata(pluginID, metadata); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	fcs.logger.Debug("Config saved", "plugin_id", pluginID, "checksum", checksum)
	return nil
}

// loadConfigFromFile 从文件加载配置
func (fcs *FileConfigStorage) loadConfigFromFile(filePath string) (*EnhancedPluginConfig, error) {
	// 读取文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 应用解密
	if fcs.encryption != nil {
		data, err = fcs.encryption.Decrypt(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt config: %w", err)
		}
	}

	// 应用解压缩
	if fcs.compression != nil {
		data, err = fcs.compression.Decompress(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress config: %w", err)
		}
	}

	// 反序列化配置
	config, err := fcs.deserializeConfig(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize config: %w", err)
	}

	return config, nil
}

// serializeConfig 序列化配置
func (fcs *FileConfigStorage) serializeConfig(config *EnhancedPluginConfig) ([]byte, error) {
	switch fcs.format {
	case StorageFormatJSON:
		return json.MarshalIndent(config, "", "  ")
	case StorageFormatYAML:
		return yaml.Marshal(config)
	case StorageFormatTOML:
		// 这里需要TOML库的支持
		return nil, fmt.Errorf("TOML format not implemented")
	default:
		return nil, fmt.Errorf("unsupported format: %s", fcs.format)
	}
}

// deserializeConfig 反序列化配置
func (fcs *FileConfigStorage) deserializeConfig(data []byte) (*EnhancedPluginConfig, error) {
	var config EnhancedPluginConfig

	switch fcs.format {
	case StorageFormatJSON:
		err := json.Unmarshal(data, &config)
		return &config, err
	case StorageFormatYAML:
		err := yaml.Unmarshal(data, &config)
		return &config, err
	case StorageFormatTOML:
		// 这里需要TOML库的支持
		return nil, fmt.Errorf("TOML format not implemented")
	default:
		return nil, fmt.Errorf("unsupported format: %s", fcs.format)
	}
}

// getOrCreateMetadata 获取或创建元数据
func (fcs *FileConfigStorage) getOrCreateMetadata(pluginID string) *ConfigMetadata {
	if metadata, exists := fcs.metadata[pluginID]; exists {
		return metadata
	}

	metadata := &ConfigMetadata{
		PluginID:        pluginID,
		CurrentVersion:  0,
		Versions:        make([]ConfigVersion, 0),
		StorageFormat:   fcs.format,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		MaxVersions:     fcs.maxVersions,
		CompressionType: fcs.getCompressionType(),
		Encrypted:       fcs.encryption != nil,
		CustomFields:    make(map[string]interface{}),
	}

	fcs.metadata[pluginID] = metadata
	return metadata
}

// saveMetadata 保存元数据
func (fcs *FileConfigStorage) saveMetadata(pluginID string, metadata *ConfigMetadata) error {
	metadataPath := fcs.getMetadataPath(pluginID)

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return fcs.writeFile(metadataPath, data)
}

// loadMetadata 加载元数据
func (fcs *FileConfigStorage) loadMetadata(pluginID string) error {
	metadataPath := fcs.getMetadataPath(pluginID)

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 元数据文件不存在，创建新的
			fcs.getOrCreateMetadata(pluginID)
			return nil
		}
		return fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata ConfigMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	fcs.metadata[pluginID] = &metadata
	return nil
}

// loadAllMetadata 加载所有元数据
func (fcs *FileConfigStorage) loadAllMetadata(ctx context.Context) error {
	metadataDir := filepath.Join(fcs.basePath, "metadata")

	files, err := os.ReadDir(metadataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read metadata directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}

		pluginID := strings.TrimSuffix(name, ".json")
		if err := fcs.loadMetadata(pluginID); err != nil {
			fcs.logger.Warn("Failed to load metadata", "plugin_id", pluginID, "error", err)
		}
	}

	return nil
}

// 路径辅助方法
func (fcs *FileConfigStorage) getConfigPath(pluginID string) string {
	filename := fmt.Sprintf("%s.%s", pluginID, fcs.format)
	return filepath.Join(fcs.basePath, "configs", filename)
}

func (fcs *FileConfigStorage) getVersionPath(pluginID string, version int) string {
	filename := fmt.Sprintf("%s_v%d.%s", pluginID, version, fcs.format)
	return filepath.Join(fcs.getVersionsDir(pluginID), filename)
}

func (fcs *FileConfigStorage) getVersionsDir(pluginID string) string {
	return filepath.Join(fcs.basePath, "versions", pluginID)
}

func (fcs *FileConfigStorage) getBackupsDir(pluginID string) string {
	return filepath.Join(fcs.basePath, "backups", pluginID)
}

func (fcs *FileConfigStorage) getMetadataPath(pluginID string) string {
	return filepath.Join(fcs.basePath, "metadata", pluginID+".json")
}

// 工具方法
func (fcs *FileConfigStorage) writeFile(filePath string, data []byte) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(filePath, data, 0644)
}

func (fcs *FileConfigStorage) copyFile(src string, dst io.Writer) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	_, err = io.Copy(dst, srcFile)
	return err
}

func (fcs *FileConfigStorage) copyFileReverse(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func (fcs *FileConfigStorage) calculateChecksum(data []byte) string {
	hash := md5.Sum(data)
	return fmt.Sprintf("%x", hash)
}

func (fcs *FileConfigStorage) getCompressionType() string {
	if fcs.compression != nil {
		return fcs.compression.Type()
	}
	return ""
}

// ConfigEncryption 配置加密接口
type ConfigEncryption interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
	Type() string
}

// ConfigCompression 配置压缩接口
type ConfigCompression interface {
	Compress(data []byte) ([]byte, error)
	Decompress(data []byte) ([]byte, error)
	Type() string
}

// ConfigStorageManager 配置存储管理器
type ConfigStorageManager struct {
	logger  *slog.Logger
	storage ConfigStorage
	mu      sync.RWMutex
}

// NewConfigStorageManager 创建配置存储管理器
func NewConfigStorageManager(logger *slog.Logger, storage ConfigStorage) *ConfigStorageManager {
	return &ConfigStorageManager{
		logger:  logger,
		storage: storage,
	}
}

// Initialize 初始化存储管理器
func (csm *ConfigStorageManager) Initialize(ctx context.Context) error {
	if initializer, ok := csm.storage.(*FileConfigStorage); ok {
		return initializer.Initialize(ctx)
	}
	return nil
}

// GetStorage 获取存储实例
func (csm *ConfigStorageManager) GetStorage() ConfigStorage {
	csm.mu.RLock()
	defer csm.mu.RUnlock()
	return csm.storage
}

// SetStorage 设置存储实例
func (csm *ConfigStorageManager) SetStorage(storage ConfigStorage) {
	csm.mu.Lock()
	defer csm.mu.Unlock()
	csm.storage = storage
}

// MigrateStorage 迁移存储
func (csm *ConfigStorageManager) MigrateStorage(ctx context.Context, newStorage ConfigStorage) error {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	// 获取所有配置
	configIDs, err := csm.storage.ListConfigs(ctx)
	if err != nil {
		return fmt.Errorf("failed to list configs: %w", err)
	}

	// 批量加载配置
	configs, err := csm.storage.BatchLoad(ctx, configIDs)
	if err != nil {
		return fmt.Errorf("failed to load configs: %w", err)
	}

	// 批量保存到新存储
	if err := newStorage.BatchSave(ctx, configs); err != nil {
		return fmt.Errorf("failed to save configs to new storage: %w", err)
	}

	// 切换存储
	csm.storage = newStorage

	csm.logger.Info("Storage migration completed", "migrated_configs", len(configs))
	return nil
}