package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteBackendConfig SQLite后端配置
type SQLiteBackendConfig struct {
	DatabasePath    string        `json:"database_path"`    // 数据库文件路径
	MaxConnections  int           `json:"max_connections"`  // 最大连接数
	ConnectionTTL   time.Duration `json:"connection_ttl"`   // 连接TTL
	WALMode         bool          `json:"wal_mode"`         // 是否启用WAL模式
	SyncMode        string        `json:"sync_mode"`        // 同步模式：OFF, NORMAL, FULL
	CacheSize       int           `json:"cache_size"`       // 缓存大小(KB)
	BusyTimeout     time.Duration `json:"busy_timeout"`     // 忙等超时
	JournalMode     string        `json:"journal_mode"`     // 日志模式：DELETE, WAL, MEMORY
	AutoVacuum      bool          `json:"auto_vacuum"`      // 自动清理
	ForeignKeys     bool          `json:"foreign_keys"`     // 外键约束
}

// DefaultSQLiteBackendConfig 默认SQLite后端配置
func DefaultSQLiteBackendConfig() *SQLiteBackendConfig {
	return &SQLiteBackendConfig{
		DatabasePath:    "./data/storage.db",
		MaxConnections:  10,
		ConnectionTTL:   30 * time.Minute,
		WALMode:         true,
		SyncMode:        "NORMAL",
		CacheSize:       8192, // 8MB
		BusyTimeout:     30 * time.Second,
		JournalMode:     "WAL",
		AutoVacuum:      true,
		ForeignKeys:     true,
	}
}

// SQLiteEntry SQLite存储条目
type SQLiteEntry struct {
	Key       string     `json:"key"`
	Value     string     `json:"value"`     // JSON序列化的值
	ExpireAt  *time.Time `json:"expire_at"` // 过期时间
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// IsExpired 检查是否过期
func (se *SQLiteEntry) IsExpired() bool {
	if se.ExpireAt == nil {
		return false
	}
	return time.Now().After(*se.ExpireAt)
}

// SQLiteBackend SQLite存储后端
type SQLiteBackend struct {
	config    *SQLiteBackendConfig // 配置
	db        *sql.DB              // 数据库连接
	stats     BackendStats         // 统计信息
	statsMu   sync.RWMutex         // 统计信息锁
	closed    bool                 // 是否已关闭
	mu        sync.RWMutex         // 读写锁
	cleanupCh chan struct{}        // 清理通道
}

// NewSQLiteBackend 创建SQLite存储后端
func NewSQLiteBackend(config *SQLiteBackendConfig) *SQLiteBackend {
	if config == nil {
		config = DefaultSQLiteBackendConfig()
	}

	return &SQLiteBackend{
		config:    config,
		stats:     BackendStats{},
		cleanupCh: make(chan struct{}),
	}
}

// Initialize 初始化后端
func (sb *SQLiteBackend) Initialize() error {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if sb.closed {
		return fmt.Errorf("backend is closed")
	}

	// 创建数据库目录
	dbDir := filepath.Dir(sb.config.DatabasePath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// 打开数据库连接
	db, err := sql.Open("sqlite3", sb.buildConnectionString())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(sb.config.MaxConnections)
	db.SetMaxIdleConns(sb.config.MaxConnections / 2)
	db.SetConnMaxLifetime(sb.config.ConnectionTTL)

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	sb.db = db

	// 初始化数据库结构
	if err := sb.initializeSchema(); err != nil {
		sb.db.Close()
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	// 启动清理协程
	go sb.cleanupWorker()

	return nil
}

// buildConnectionString 构建连接字符串
func (sb *SQLiteBackend) buildConnectionString() string {
	params := []string{
		fmt.Sprintf("_busy_timeout=%d", int(sb.config.BusyTimeout.Milliseconds())),
		fmt.Sprintf("_cache_size=-%d", sb.config.CacheSize), // 负数表示KB
		fmt.Sprintf("_foreign_keys=%s", boolToString(sb.config.ForeignKeys)),
		fmt.Sprintf("_journal_mode=%s", sb.config.JournalMode),
		fmt.Sprintf("_synchronous=%s", sb.config.SyncMode),
	}

	if sb.config.AutoVacuum {
		params = append(params, "_auto_vacuum=FULL")
	}

	return fmt.Sprintf("%s?%s", sb.config.DatabasePath, strings.Join(params, "&"))
}

// boolToString 布尔值转字符串
func boolToString(b bool) string {
	if b {
		return "ON"
	}
	return "OFF"
}

// initializeSchema 初始化数据库结构
func (sb *SQLiteBackend) initializeSchema() error {
	// 创建存储表
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS storage_entries (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		expire_at INTEGER,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);
	`

	if _, err := sb.db.Exec(createTableSQL); err != nil {
		return fmt.Errorf("failed to create storage_entries table: %w", err)
	}

	// 创建索引
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_expire_at ON storage_entries(expire_at);",
		"CREATE INDEX IF NOT EXISTS idx_created_at ON storage_entries(created_at);",
		"CREATE INDEX IF NOT EXISTS idx_updated_at ON storage_entries(updated_at);",
	}

	for _, indexSQL := range indexes {
		if _, err := sb.db.Exec(indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// 创建版本管理表
	versionTableSQL := `
	CREATE TABLE IF NOT EXISTS schema_versions (
		version INTEGER PRIMARY KEY,
		description TEXT NOT NULL,
		applied_at INTEGER NOT NULL
	);
	`

	if _, err := sb.db.Exec(versionTableSQL); err != nil {
		return fmt.Errorf("failed to create schema_versions table: %w", err)
	}

	// 插入初始版本
	insertVersionSQL := `
	INSERT OR IGNORE INTO schema_versions (version, description, applied_at)
	VALUES (1, 'Initial schema', ?);
	`

	if _, err := sb.db.Exec(insertVersionSQL, time.Now().Unix()); err != nil {
		return fmt.Errorf("failed to insert initial version: %w", err)
	}

	return nil
}

// Get 获取值
func (sb *SQLiteBackend) Get(key string) (interface{}, error) {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	if sb.closed {
		return nil, fmt.Errorf("backend is closed")
	}

	// 更新统计信息
	sb.updateStats(func(stats *BackendStats) {
		stats.ReadCount++
	})

	query := `
	SELECT value, expire_at FROM storage_entries 
	WHERE key = ? AND (expire_at IS NULL OR expire_at > ?)
	`

	var valueStr string
	var expireAt *int64
	now := time.Now().Unix()

	err := sb.db.QueryRow(query, key, now).Scan(&valueStr, &expireAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("key not found: %s", key)
		}
		return nil, fmt.Errorf("failed to get key %s: %w", key, err)
	}

	// 反序列化值
	var value interface{}
	if err := json.Unmarshal([]byte(valueStr), &value); err != nil {
		return nil, fmt.Errorf("failed to unmarshal value for key %s: %w", key, err)
	}

	return value, nil
}

// Set 设置值
func (sb *SQLiteBackend) Set(key string, value interface{}, ttl time.Duration) error {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	if sb.closed {
		return fmt.Errorf("backend is closed")
	}

	// 更新统计信息
	sb.updateStats(func(stats *BackendStats) {
		stats.WriteCount++
	})

	// 序列化值
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
	}

	now := time.Now()
	nowUnix := now.Unix()
	var expireAt *int64

	// 设置过期时间
	if ttl > 0 {
		expire := now.Add(ttl).Unix()
		expireAt = &expire
	}

	// 使用UPSERT语句
	upsertSQL := `
	INSERT INTO storage_entries (key, value, expire_at, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(key) DO UPDATE SET
		value = excluded.value,
		expire_at = excluded.expire_at,
		updated_at = excluded.updated_at;
	`

	_, err = sb.db.Exec(upsertSQL, key, string(valueBytes), expireAt, nowUnix, nowUnix)
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}

// Delete 删除值
func (sb *SQLiteBackend) Delete(key string) error {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	if sb.closed {
		return fmt.Errorf("backend is closed")
	}

	// 更新统计信息
	sb.updateStats(func(stats *BackendStats) {
		stats.DeleteCount++
	})

	deleteSQL := "DELETE FROM storage_entries WHERE key = ?"
	_, err := sb.db.Exec(deleteSQL, key)
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}

	return nil
}

// Exists 检查键是否存在
func (sb *SQLiteBackend) Exists(key string) (bool, error) {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	if sb.closed {
		return false, fmt.Errorf("backend is closed")
	}

	query := `
	SELECT 1 FROM storage_entries 
	WHERE key = ? AND (expire_at IS NULL OR expire_at > ?)
	LIMIT 1
	`

	var exists int
	now := time.Now().Unix()
	err := sb.db.QueryRow(query, key, now).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check existence of key %s: %w", key, err)
	}

	return exists == 1, nil
}

// updateStats 更新统计信息
func (sb *SQLiteBackend) updateStats(updater func(*BackendStats)) {
	sb.statsMu.Lock()
	defer sb.statsMu.Unlock()
	updater(&sb.stats)
}

// cleanupWorker 清理过期数据的工作协程
func (sb *SQLiteBackend) cleanupWorker() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sb.cleanupExpiredEntries()
		case <-sb.cleanupCh:
			return
		}
	}
}

// cleanupExpiredEntries 清理过期条目
func (sb *SQLiteBackend) cleanupExpiredEntries() {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	if sb.closed || sb.db == nil {
		return
	}

	deleteSQL := "DELETE FROM storage_entries WHERE expire_at IS NOT NULL AND expire_at <= ?"
	now := time.Now().Unix()

	result, err := sb.db.Exec(deleteSQL, now)
	if err != nil {
		// 记录错误但不中断清理过程
		fmt.Printf("Failed to cleanup expired entries: %v\n", err)
		return
	}

	if rowsAffected, err := result.RowsAffected(); err == nil && rowsAffected > 0 {
		fmt.Printf("Cleaned up %d expired entries\n", rowsAffected)
	}
}

// Close 关闭后端
func (sb *SQLiteBackend) Close() error {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if sb.closed {
		return nil
	}

	sb.closed = true

	// 停止清理协程
	select {
	case <-sb.cleanupCh:
		// 通道已关闭
	default:
		close(sb.cleanupCh)
	}

	// 关闭数据库连接
	if sb.db != nil {
		return sb.db.Close()
	}

	return nil
}

// GetStats 获取统计信息
func (sb *SQLiteBackend) GetStats() BackendStats {
	sb.statsMu.RLock()
	defer sb.statsMu.RUnlock()

	// 获取键数量
	if sb.db != nil && !sb.closed {
		var count int64
		query := "SELECT COUNT(*) FROM storage_entries WHERE expire_at IS NULL OR expire_at > ?"
		now := time.Now().Unix()
		if err := sb.db.QueryRow(query, now).Scan(&count); err == nil {
			sb.stats.KeyCount = count
		}
	}

	return sb.stats
}