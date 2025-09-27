package storage

import (
	"fmt"
	"sync"
	"time"
)

// LocalStorageBackend 本地存储后端实现
// 基于SQLite数据库的本地存储后端，支持数据迁移、备份恢复等高级功能
type LocalStorageBackend struct {
	sqliteBackend   *SQLiteBackend    // SQLite后端
	migrationMgr    *MigrationManager // 迁移管理器
	backupMgr       *BackupManager    // 备份管理器
	config          *LocalStorageConfig // 配置
	mu              sync.RWMutex      // 读写锁
	closed          bool              // 是否已关闭
}

// LocalStorageConfig 本地存储配置
type LocalStorageConfig struct {
	SQLiteConfig    *SQLiteBackendConfig `json:"sqlite_config"`    // SQLite配置
	BackupDir       string               `json:"backup_dir"`       // 备份目录
	AutoMigrate     bool                 `json:"auto_migrate"`     // 自动迁移
	AutoBackup      bool                 `json:"auto_backup"`      // 自动备份
	BackupInterval  time.Duration        `json:"backup_interval"`  // 备份间隔
	MaxBackups      int                  `json:"max_backups"`      // 最大备份数量
	CompressionType string               `json:"compression_type"` // 压缩类型
	EncryptionKey   string               `json:"-"`                // 加密密钥（不序列化）
}

// DefaultLocalStorageConfig 默认本地存储配置
func DefaultLocalStorageConfig() *LocalStorageConfig {
	return &LocalStorageConfig{
		SQLiteConfig:    DefaultSQLiteBackendConfig(),
		BackupDir:       "./data/backups",
		AutoMigrate:     true,
		AutoBackup:      false,
		BackupInterval:  24 * time.Hour,
		MaxBackups:      7,
		CompressionType: "gzip",
	}
}

// NewLocalStorageBackend 创建本地存储后端
func NewLocalStorageBackend(config *LocalStorageConfig) *LocalStorageBackend {
	if config == nil {
		config = DefaultLocalStorageConfig()
	}

	// 创建SQLite后端
	sqliteBackend := NewSQLiteBackend(config.SQLiteConfig)

	return &LocalStorageBackend{
		sqliteBackend: sqliteBackend,
		config:        config,
	}
}

// Initialize 初始化后端
func (lsb *LocalStorageBackend) Initialize() error {
	lsb.mu.Lock()
	defer lsb.mu.Unlock()

	if lsb.closed {
		return fmt.Errorf("backend is closed")
	}

	// 初始化SQLite后端
	if err := lsb.sqliteBackend.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize SQLite backend: %w", err)
	}

	// 创建迁移管理器
	lsb.migrationMgr = NewMigrationManager(lsb.sqliteBackend.db)

	// 验证迁移
	if err := lsb.migrationMgr.ValidateMigrations(); err != nil {
		return fmt.Errorf("migration validation failed: %w", err)
	}

	// 自动迁移
	if lsb.config.AutoMigrate {
		if err := lsb.migrationMgr.MigrateToLatest(); err != nil {
			return fmt.Errorf("auto migration failed: %w", err)
		}
	}

	// 创建备份管理器
	lsb.backupMgr = NewBackupManager(lsb.sqliteBackend.db, lsb.config.BackupDir)

	// 启动自动备份
	if lsb.config.AutoBackup {
		go lsb.autoBackupWorker()
	}

	return nil
}

// Get 获取值
func (lsb *LocalStorageBackend) Get(key string) (interface{}, error) {
	lsb.mu.RLock()
	defer lsb.mu.RUnlock()

	if lsb.closed {
		return nil, fmt.Errorf("backend is closed")
	}

	return lsb.sqliteBackend.Get(key)
}

// Set 设置值
func (lsb *LocalStorageBackend) Set(key string, value interface{}, ttl time.Duration) error {
	lsb.mu.RLock()
	defer lsb.mu.RUnlock()

	if lsb.closed {
		return fmt.Errorf("backend is closed")
	}

	return lsb.sqliteBackend.Set(key, value, ttl)
}

// Delete 删除值
func (lsb *LocalStorageBackend) Delete(key string) error {
	lsb.mu.RLock()
	defer lsb.mu.RUnlock()

	if lsb.closed {
		return fmt.Errorf("backend is closed")
	}

	return lsb.sqliteBackend.Delete(key)
}

// Exists 检查键是否存在
func (lsb *LocalStorageBackend) Exists(key string) (bool, error) {
	lsb.mu.RLock()
	defer lsb.mu.RUnlock()

	if lsb.closed {
		return false, fmt.Errorf("backend is closed")
	}

	return lsb.sqliteBackend.Exists(key)
}

// GetBatch 批量获取值
func (lsb *LocalStorageBackend) GetBatch(keys []string) (map[string]interface{}, error) {
	lsb.mu.RLock()
	defer lsb.mu.RUnlock()

	if lsb.closed {
		return nil, fmt.Errorf("backend is closed")
	}

	return lsb.sqliteBackend.GetBatch(keys)
}

// SetBatch 批量设置值
func (lsb *LocalStorageBackend) SetBatch(items map[string]interface{}, ttl time.Duration) error {
	lsb.mu.RLock()
	defer lsb.mu.RUnlock()

	if lsb.closed {
		return fmt.Errorf("backend is closed")
	}

	return lsb.sqliteBackend.SetBatch(items, ttl)
}

// DeleteBatch 批量删除值
func (lsb *LocalStorageBackend) DeleteBatch(keys []string) error {
	lsb.mu.RLock()
	defer lsb.mu.RUnlock()

	if lsb.closed {
		return fmt.Errorf("backend is closed")
	}

	return lsb.sqliteBackend.DeleteBatch(keys)
}

// Find 查找匹配模式的键值对
func (lsb *LocalStorageBackend) Find(pattern string, limit int) (map[string]interface{}, error) {
	lsb.mu.RLock()
	defer lsb.mu.RUnlock()

	if lsb.closed {
		return nil, fmt.Errorf("backend is closed")
	}

	return lsb.sqliteBackend.Find(pattern, limit)
}

// Count 统计匹配模式的键数量
func (lsb *LocalStorageBackend) Count(pattern string) (int64, error) {
	lsb.mu.RLock()
	defer lsb.mu.RUnlock()

	if lsb.closed {
		return 0, fmt.Errorf("backend is closed")
	}

	return lsb.sqliteBackend.Count(pattern)
}

// Keys 获取匹配模式的所有键
func (lsb *LocalStorageBackend) Keys(pattern string) ([]string, error) {
	lsb.mu.RLock()
	defer lsb.mu.RUnlock()

	if lsb.closed {
		return nil, fmt.Errorf("backend is closed")
	}

	return lsb.sqliteBackend.Keys(pattern)
}

// Close 关闭后端
func (lsb *LocalStorageBackend) Close() error {
	lsb.mu.Lock()
	defer lsb.mu.Unlock()

	if lsb.closed {
		return nil
	}

	lsb.closed = true

	// 关闭SQLite后端
	if lsb.sqliteBackend != nil {
		return lsb.sqliteBackend.Close()
	}

	return nil
}

// GetStats 获取统计信息
func (lsb *LocalStorageBackend) GetStats() BackendStats {
	lsb.mu.RLock()
	defer lsb.mu.RUnlock()

	if lsb.closed || lsb.sqliteBackend == nil {
		return BackendStats{}
	}

	return lsb.sqliteBackend.GetStats()
}

// === 数据迁移相关方法 ===

// GetMigrationManager 获取迁移管理器
func (lsb *LocalStorageBackend) GetMigrationManager() *MigrationManager {
	lsb.mu.RLock()
	defer lsb.mu.RUnlock()
	return lsb.migrationMgr
}

// MigrateToVersion 迁移到指定版本
func (lsb *LocalStorageBackend) MigrateToVersion(version int) error {
	lsb.mu.RLock()
	defer lsb.mu.RUnlock()

	if lsb.closed || lsb.migrationMgr == nil {
		return fmt.Errorf("backend is closed or migration manager not initialized")
	}

	currentVersion, err := lsb.migrationMgr.GetCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if version > currentVersion {
		return lsb.migrationMgr.MigrateUp(version)
	} else if version < currentVersion {
		return lsb.migrationMgr.MigrateDown(version)
	}

	return nil // 已经是目标版本
}

// GetMigrationStatus 获取迁移状态
func (lsb *LocalStorageBackend) GetMigrationStatus() (*MigrationStatus, error) {
	lsb.mu.RLock()
	defer lsb.mu.RUnlock()

	if lsb.closed || lsb.migrationMgr == nil {
		return nil, fmt.Errorf("backend is closed or migration manager not initialized")
	}

	return lsb.migrationMgr.CheckMigrationStatus()
}

// === 备份恢复相关方法 ===

// GetBackupManager 获取备份管理器
func (lsb *LocalStorageBackend) GetBackupManager() *BackupManager {
	lsb.mu.RLock()
	defer lsb.mu.RUnlock()
	return lsb.backupMgr
}

// CreateBackup 创建备份
func (lsb *LocalStorageBackend) CreateBackup(options *BackupOptions) (*BackupInfo, error) {
	lsb.mu.RLock()
	defer lsb.mu.RUnlock()

	if lsb.closed || lsb.backupMgr == nil {
		return nil, fmt.Errorf("backend is closed or backup manager not initialized")
	}

	// 设置默认加密密钥
	if options.Encrypt && options.Password == "" && lsb.config.EncryptionKey != "" {
		options.Password = lsb.config.EncryptionKey
	}

	// 设置默认压缩
	if lsb.config.CompressionType != "" {
		options.Compress = true
	}

	return lsb.backupMgr.CreateBackup(options)
}

// RestoreBackup 恢复备份
func (lsb *LocalStorageBackend) RestoreBackup(backupID int64, password string) error {
	lsb.mu.RLock()
	defer lsb.mu.RUnlock()

	if lsb.closed || lsb.backupMgr == nil {
		return fmt.Errorf("backend is closed or backup manager not initialized")
	}

	// 使用默认加密密钥
	if password == "" && lsb.config.EncryptionKey != "" {
		password = lsb.config.EncryptionKey
	}

	return lsb.backupMgr.RestoreBackup(backupID, password)
}

// ListBackups 列出所有备份
func (lsb *LocalStorageBackend) ListBackups() ([]BackupInfo, error) {
	lsb.mu.RLock()
	defer lsb.mu.RUnlock()

	if lsb.closed || lsb.backupMgr == nil {
		return nil, fmt.Errorf("backend is closed or backup manager not initialized")
	}

	return lsb.backupMgr.ListBackups()
}

// DeleteBackup 删除备份
func (lsb *LocalStorageBackend) DeleteBackup(backupID int64) error {
	lsb.mu.RLock()
	defer lsb.mu.RUnlock()

	if lsb.closed || lsb.backupMgr == nil {
		return fmt.Errorf("backend is closed or backup manager not initialized")
	}

	return lsb.backupMgr.DeleteBackup(backupID)
}

// === 自动备份相关方法 ===

// autoBackupWorker 自动备份工作协程
func (lsb *LocalStorageBackend) autoBackupWorker() {
	ticker := time.NewTicker(lsb.config.BackupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			lsb.performAutoBackup()
		}

		// 检查是否已关闭
		lsb.mu.RLock()
		closed := lsb.closed
		lsb.mu.RUnlock()

		if closed {
			return
		}
	}
}

// performAutoBackup 执行自动备份
func (lsb *LocalStorageBackend) performAutoBackup() {
	if lsb.closed || lsb.backupMgr == nil {
		return
	}

	// 创建自动备份
	backupName := fmt.Sprintf("auto_backup_%s", time.Now().Format("20060102_150405"))
	options := &BackupOptions{
		Name:        backupName,
		Description: "Automatic backup",
		Type:        BackupTypeFull,
		Format:      BackupFormatJSON,
		Compress:    lsb.config.CompressionType != "",
		Encrypt:     lsb.config.EncryptionKey != "",
		Password:    lsb.config.EncryptionKey,
	}

	_, err := lsb.backupMgr.CreateBackup(options)
	if err != nil {
		fmt.Printf("Auto backup failed: %v\n", err)
		return
	}

	// 清理旧备份
	lsb.cleanupOldBackups()
}

// cleanupOldBackups 清理旧备份
func (lsb *LocalStorageBackend) cleanupOldBackups() {
	if lsb.config.MaxBackups <= 0 {
		return
	}

	backups, err := lsb.backupMgr.ListBackups()
	if err != nil {
		fmt.Printf("Failed to list backups for cleanup: %v\n", err)
		return
	}

	// 只保留自动备份
	var autoBackups []BackupInfo
	for _, backup := range backups {
		if backup.Name[:11] == "auto_backup" {
			autoBackups = append(autoBackups, backup)
		}
	}

	// 删除超出数量限制的备份
	if len(autoBackups) > lsb.config.MaxBackups {
		for i := lsb.config.MaxBackups; i < len(autoBackups); i++ {
			if err := lsb.backupMgr.DeleteBackup(autoBackups[i].ID); err != nil {
				fmt.Printf("Failed to delete old backup %d: %v\n", autoBackups[i].ID, err)
			}
		}
	}
}

// === 高级功能方法 ===

// Compact 压缩数据库
func (lsb *LocalStorageBackend) Compact() error {
	lsb.mu.RLock()
	defer lsb.mu.RUnlock()

	if lsb.closed || lsb.sqliteBackend == nil || lsb.sqliteBackend.db == nil {
		return fmt.Errorf("backend is closed or database not available")
	}

	// 执行VACUUM命令压缩数据库
	_, err := lsb.sqliteBackend.db.Exec("VACUUM")
	if err != nil {
		return fmt.Errorf("failed to compact database: %w", err)
	}

	return nil
}

// Analyze 分析数据库统计信息
func (lsb *LocalStorageBackend) Analyze() error {
	lsb.mu.RLock()
	defer lsb.mu.RUnlock()

	if lsb.closed || lsb.sqliteBackend == nil || lsb.sqliteBackend.db == nil {
		return fmt.Errorf("backend is closed or database not available")
	}

	// 执行ANALYZE命令更新统计信息
	_, err := lsb.sqliteBackend.db.Exec("ANALYZE")
	if err != nil {
		return fmt.Errorf("failed to analyze database: %w", err)
	}

	return nil
}

// GetDatabaseInfo 获取数据库信息
func (lsb *LocalStorageBackend) GetDatabaseInfo() (*DatabaseInfo, error) {
	lsb.mu.RLock()
	defer lsb.mu.RUnlock()

	if lsb.closed || lsb.sqliteBackend == nil || lsb.sqliteBackend.db == nil {
		return nil, fmt.Errorf("backend is closed or database not available")
	}

	info := &DatabaseInfo{}

	// 获取数据库大小
	var pageCount, pageSize int64
	err := lsb.sqliteBackend.db.QueryRow("PRAGMA page_count").Scan(&pageCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get page count: %w", err)
	}

	err = lsb.sqliteBackend.db.QueryRow("PRAGMA page_size").Scan(&pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get page size: %w", err)
	}

	info.Size = pageCount * pageSize
	info.PageCount = pageCount
	info.PageSize = pageSize

	// 获取表信息
	var tableCount int64
	err = lsb.sqliteBackend.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&tableCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get table count: %w", err)
	}
	info.TableCount = tableCount

	// 获取索引信息
	var indexCount int64
	err = lsb.sqliteBackend.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index'").Scan(&indexCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get index count: %w", err)
	}
	info.IndexCount = indexCount

	// 获取版本信息
	if lsb.migrationMgr != nil {
		version, err := lsb.migrationMgr.GetCurrentVersion()
		if err == nil {
			info.SchemaVersion = version
		}
	}

	return info, nil
}

// DatabaseInfo 数据库信息
type DatabaseInfo struct {
	Size          int64 `json:"size"`           // 数据库大小（字节）
	PageCount     int64 `json:"page_count"`     // 页面数量
	PageSize      int64 `json:"page_size"`      // 页面大小
	TableCount    int64 `json:"table_count"`    // 表数量
	IndexCount    int64 `json:"index_count"`    // 索引数量
	SchemaVersion int   `json:"schema_version"`  // 模式版本
}