package storage

import (
	"database/sql"
	"fmt"
	"sort"
	"time"
)

// Migration 数据库迁移结构
type Migration struct {
	Version     int                 `json:"version"`     // 版本号
	Description string              `json:"description"` // 描述
	UpSQL       string              `json:"up_sql"`      // 升级SQL
	DownSQL     string              `json:"down_sql"`    // 降级SQL
	UpFunc      func(*sql.DB) error `json:"-"`           // 升级函数
	DownFunc    func(*sql.DB) error `json:"-"`           // 降级函数
}

// MigrationManager 迁移管理器
type MigrationManager struct {
	db         *sql.DB
	migrations []Migration
}

// NewMigrationManager 创建迁移管理器
func NewMigrationManager(db *sql.DB) *MigrationManager {
	return &MigrationManager{
		db:         db,
		migrations: getBuiltinMigrations(),
	}
}

// getBuiltinMigrations 获取内置迁移
func getBuiltinMigrations() []Migration {
	return []Migration{
		{
			Version:     1,
			Description: "Initial schema",
			UpSQL: `
				CREATE TABLE IF NOT EXISTS storage_entries (
					key TEXT PRIMARY KEY,
					value TEXT NOT NULL,
					expire_at INTEGER,
					created_at INTEGER NOT NULL,
					updated_at INTEGER NOT NULL
				);
				
				CREATE INDEX IF NOT EXISTS idx_expire_at ON storage_entries(expire_at);
				CREATE INDEX IF NOT EXISTS idx_created_at ON storage_entries(created_at);
				CREATE INDEX IF NOT EXISTS idx_updated_at ON storage_entries(updated_at);
				
				CREATE TABLE IF NOT EXISTS schema_versions (
					version INTEGER PRIMARY KEY,
					description TEXT NOT NULL,
					applied_at INTEGER NOT NULL
				);
			`,
			DownSQL: `
				DROP TABLE IF EXISTS storage_entries;
				DROP TABLE IF EXISTS schema_versions;
			`,
		},
		{
			Version:     2,
			Description: "Add metadata table for enhanced storage features",
			UpSQL: `
				CREATE TABLE IF NOT EXISTS storage_metadata (
					key TEXT PRIMARY KEY,
					size INTEGER NOT NULL DEFAULT 0,
					type TEXT NOT NULL DEFAULT 'unknown',
					tags TEXT, -- JSON array of tags
					version INTEGER NOT NULL DEFAULT 1,
					checksum TEXT,
					created_at INTEGER NOT NULL,
					updated_at INTEGER NOT NULL,
					FOREIGN KEY (key) REFERENCES storage_entries(key) ON DELETE CASCADE
				);
				
				CREATE INDEX IF NOT EXISTS idx_metadata_type ON storage_metadata(type);
				CREATE INDEX IF NOT EXISTS idx_metadata_size ON storage_metadata(size);
				CREATE INDEX IF NOT EXISTS idx_metadata_version ON storage_metadata(version);
			`,
			DownSQL: `
				DROP TABLE IF EXISTS storage_metadata;
			`,
		},
		{
			Version:     3,
			Description: "Add backup and recovery tables",
			UpSQL: `
				CREATE TABLE IF NOT EXISTS storage_backups (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					name TEXT NOT NULL UNIQUE,
					description TEXT,
					type TEXT NOT NULL DEFAULT 'full', -- full, incremental
					file_path TEXT NOT NULL,
					file_size INTEGER NOT NULL DEFAULT 0,
					compressed BOOLEAN NOT NULL DEFAULT FALSE,
					encrypted BOOLEAN NOT NULL DEFAULT FALSE,
					checksum TEXT,
					entry_count INTEGER NOT NULL DEFAULT 0,
					created_at INTEGER NOT NULL,
					status TEXT NOT NULL DEFAULT 'completed' -- pending, completed, failed
				);
				
				CREATE INDEX IF NOT EXISTS idx_backups_type ON storage_backups(type);
				CREATE INDEX IF NOT EXISTS idx_backups_created_at ON storage_backups(created_at);
				CREATE INDEX IF NOT EXISTS idx_backups_status ON storage_backups(status);
			`,
			DownSQL: `
				DROP TABLE IF EXISTS storage_backups;
			`,
		},
	}
}

// AddMigration 添加自定义迁移
func (mm *MigrationManager) AddMigration(migration Migration) {
	mm.migrations = append(mm.migrations, migration)
	// 按版本号排序
	sort.Slice(mm.migrations, func(i, j int) bool {
		return mm.migrations[i].Version < mm.migrations[j].Version
	})
}

// GetCurrentVersion 获取当前数据库版本
func (mm *MigrationManager) GetCurrentVersion() (int, error) {
	query := "SELECT COALESCE(MAX(version), 0) FROM schema_versions"
	var version int
	err := mm.db.QueryRow(query).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("failed to get current version: %w", err)
	}
	return version, nil
}

// GetLatestVersion 获取最新可用版本
func (mm *MigrationManager) GetLatestVersion() int {
	if len(mm.migrations) == 0 {
		return 0
	}
	return mm.migrations[len(mm.migrations)-1].Version
}

// MigrateUp 向上迁移到指定版本
func (mm *MigrationManager) MigrateUp(targetVersion int) error {
	currentVersion, err := mm.GetCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if targetVersion <= currentVersion {
		return fmt.Errorf("target version %d is not greater than current version %d", targetVersion, currentVersion)
	}

	// 开始事务
	tx, err := mm.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 执行迁移
	for _, migration := range mm.migrations {
		if migration.Version <= currentVersion || migration.Version > targetVersion {
			continue
		}

		fmt.Printf("Applying migration %d: %s\n", migration.Version, migration.Description)

		// 执行SQL迁移
		if migration.UpSQL != "" {
			if _, err := tx.Exec(migration.UpSQL); err != nil {
				return fmt.Errorf("failed to execute up SQL for version %d: %w", migration.Version, err)
			}
		}

		// 执行函数迁移
		if migration.UpFunc != nil {
			if err := migration.UpFunc(mm.db); err != nil {
				return fmt.Errorf("failed to execute up function for version %d: %w", migration.Version, err)
			}
		}

		// 记录版本
		insertVersionSQL := "INSERT INTO schema_versions (version, description, applied_at) VALUES (?, ?, ?)"
		if _, err := tx.Exec(insertVersionSQL, migration.Version, migration.Description, time.Now().Unix()); err != nil {
			return fmt.Errorf("failed to record version %d: %w", migration.Version, err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration transaction: %w", err)
	}

	return nil
}

// MigrateDown 向下迁移到指定版本
func (mm *MigrationManager) MigrateDown(targetVersion int) error {
	currentVersion, err := mm.GetCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if targetVersion >= currentVersion {
		return fmt.Errorf("target version %d is not less than current version %d", targetVersion, currentVersion)
	}

	// 开始事务
	tx, err := mm.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 反向执行迁移
	for i := len(mm.migrations) - 1; i >= 0; i-- {
		migration := mm.migrations[i]
		if migration.Version <= targetVersion || migration.Version > currentVersion {
			continue
		}

		fmt.Printf("Reverting migration %d: %s\n", migration.Version, migration.Description)

		// 执行SQL回滚
		if migration.DownSQL != "" {
			if _, err := tx.Exec(migration.DownSQL); err != nil {
				return fmt.Errorf("failed to execute down SQL for version %d: %w", migration.Version, err)
			}
		}

		// 执行函数回滚
		if migration.DownFunc != nil {
			if err := migration.DownFunc(mm.db); err != nil {
				return fmt.Errorf("failed to execute down function for version %d: %w", migration.Version, err)
			}
		}

		// 删除版本记录
		deleteVersionSQL := "DELETE FROM schema_versions WHERE version = ?"
		if _, err := tx.Exec(deleteVersionSQL, migration.Version); err != nil {
			return fmt.Errorf("failed to delete version record %d: %w", migration.Version, err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback transaction: %w", err)
	}

	return nil
}

// MigrateToLatest 迁移到最新版本
func (mm *MigrationManager) MigrateToLatest() error {
	latestVersion := mm.GetLatestVersion()
	if latestVersion == 0 {
		return fmt.Errorf("no migrations available")
	}

	currentVersion, err := mm.GetCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if currentVersion >= latestVersion {
		fmt.Printf("Database is already at the latest version %d\n", currentVersion)
		return nil
	}

	return mm.MigrateUp(latestVersion)
}

// GetMigrationHistory 获取迁移历史
func (mm *MigrationManager) GetMigrationHistory() ([]MigrationRecord, error) {
	query := `
		SELECT version, description, applied_at 
		FROM schema_versions 
		ORDER BY version ASC
	`

	rows, err := mm.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query migration history: %w", err)
	}
	defer rows.Close()

	var history []MigrationRecord
	for rows.Next() {
		var record MigrationRecord
		var appliedAt int64
		if err := rows.Scan(&record.Version, &record.Description, &appliedAt); err != nil {
			return nil, fmt.Errorf("failed to scan migration record: %w", err)
		}
		record.AppliedAt = time.Unix(appliedAt, 0)
		history = append(history, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating migration history: %w", err)
	}

	return history, nil
}

// ValidateMigrations 验证迁移完整性
func (mm *MigrationManager) ValidateMigrations() error {
	// 检查版本号连续性
	for i := 1; i < len(mm.migrations); i++ {
		prev := mm.migrations[i-1]
		curr := mm.migrations[i]
		if curr.Version <= prev.Version {
			return fmt.Errorf("migration version %d is not greater than previous version %d", curr.Version, prev.Version)
		}
	}

	// 检查必要字段
	for _, migration := range mm.migrations {
		if migration.Version <= 0 {
			return fmt.Errorf("migration version must be positive: %d", migration.Version)
		}
		if migration.Description == "" {
			return fmt.Errorf("migration description cannot be empty for version %d", migration.Version)
		}
		if migration.UpSQL == "" && migration.UpFunc == nil {
			return fmt.Errorf("migration must have either UpSQL or UpFunc for version %d", migration.Version)
		}
	}

	return nil
}

// MigrationRecord 迁移记录
type MigrationRecord struct {
	Version     int       `json:"version"`
	Description string    `json:"description"`
	AppliedAt   time.Time `json:"applied_at"`
}

// GetPendingMigrations 获取待执行的迁移
func (mm *MigrationManager) GetPendingMigrations() ([]Migration, error) {
	currentVersion, err := mm.GetCurrentVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get current version: %w", err)
	}

	var pending []Migration
	for _, migration := range mm.migrations {
		if migration.Version > currentVersion {
			pending = append(pending, migration)
		}
	}

	return pending, nil
}

// CheckMigrationStatus 检查迁移状态
func (mm *MigrationManager) CheckMigrationStatus() (*MigrationStatus, error) {
	currentVersion, err := mm.GetCurrentVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get current version: %w", err)
	}

	latestVersion := mm.GetLatestVersion()
	pending, err := mm.GetPendingMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to get pending migrations: %w", err)
	}

	status := &MigrationStatus{
		CurrentVersion:    currentVersion,
		LatestVersion:     latestVersion,
		PendingCount:      len(pending),
		NeedsMigration:    currentVersion < latestVersion,
		PendingMigrations: pending,
	}

	return status, nil
}

// MigrationStatus 迁移状态
type MigrationStatus struct {
	CurrentVersion    int         `json:"current_version"`
	LatestVersion     int         `json:"latest_version"`
	PendingCount      int         `json:"pending_count"`
	NeedsMigration    bool        `json:"needs_migration"`
	PendingMigrations []Migration `json:"pending_migrations"`
}