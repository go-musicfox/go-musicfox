package storage

import (
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BackupType 备份类型
type BackupType string

const (
	BackupTypeFull        BackupType = "full"        // 全量备份
	BackupTypeIncremental BackupType = "incremental" // 增量备份
)

// BackupFormat 备份格式
type BackupFormat string

const (
	BackupFormatJSON BackupFormat = "json" // JSON格式
	BackupFormatSQL  BackupFormat = "sql"  // SQL格式
	BackupFormatCSV  BackupFormat = "csv"  // CSV格式
)

// BackupOptions 备份选项
type BackupOptions struct {
	Name        string        `json:"name"`        // 备份名称
	Description string        `json:"description"` // 备份描述
	Type        BackupType    `json:"type"`        // 备份类型
	Format      BackupFormat  `json:"format"`      // 备份格式
	Compress    bool          `json:"compress"`    // 是否压缩
	Encrypt     bool          `json:"encrypt"`     // 是否加密
	Password    string        `json:"-"`           // 加密密码
	OutputPath  string        `json:"output_path"` // 输出路径
	Since       *time.Time    `json:"since"`       // 增量备份起始时间
}

// BackupInfo 备份信息
type BackupInfo struct {
	ID          int64         `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Type        BackupType    `json:"type"`
	FilePath    string        `json:"file_path"`
	FileSize    int64         `json:"file_size"`
	Compressed  bool          `json:"compressed"`
	Encrypted   bool          `json:"encrypted"`
	Checksum    string        `json:"checksum"`
	EntryCount  int64         `json:"entry_count"`
	CreatedAt   time.Time     `json:"created_at"`
	Status      string        `json:"status"`
}

// BackupManager 备份管理器
type BackupManager struct {
	db         *sql.DB
	backupDir  string
}

// NewBackupManager 创建备份管理器
func NewBackupManager(db *sql.DB, backupDir string) *BackupManager {
	return &BackupManager{
		db:        db,
		backupDir: backupDir,
	}
}

// CreateBackup 创建备份
func (bm *BackupManager) CreateBackup(options *BackupOptions) (*BackupInfo, error) {
	if options == nil {
		return nil, fmt.Errorf("backup options cannot be nil")
	}

	// 验证选项
	if err := bm.validateBackupOptions(options); err != nil {
		return nil, fmt.Errorf("invalid backup options: %w", err)
	}

	// 创建备份目录
	if err := os.MkdirAll(bm.backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// 生成备份文件路径
	filePath := bm.generateBackupFilePath(options)

	// 开始备份
	backupInfo := &BackupInfo{
		Name:        options.Name,
		Description: options.Description,
		Type:        options.Type,
		FilePath:    filePath,
		Compressed:  options.Compress,
		Encrypted:   options.Encrypt,
		CreatedAt:   time.Now(),
		Status:      "pending",
	}

	// 记录备份信息到数据库
	backupID, err := bm.recordBackupInfo(backupInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to record backup info: %w", err)
	}
	backupInfo.ID = backupID

	// 执行备份
	err = bm.performBackup(options, backupInfo)
	if err != nil {
		// 更新备份状态为失败
		bm.updateBackupStatus(backupID, "failed")
		return nil, fmt.Errorf("backup failed: %w", err)
	}

	// 计算文件大小和校验和
	if err := bm.updateBackupFileInfo(backupInfo); err != nil {
		return nil, fmt.Errorf("failed to update backup file info: %w", err)
	}

	// 更新备份状态为完成
	if err := bm.updateBackupStatus(backupID, "completed"); err != nil {
		return nil, fmt.Errorf("failed to update backup status: %w", err)
	}
	backupInfo.Status = "completed"

	return backupInfo, nil
}

// performBackup 执行备份
func (bm *BackupManager) performBackup(options *BackupOptions, info *BackupInfo) error {
	// 获取要备份的数据
	data, err := bm.getBackupData(options)
	if err != nil {
		return fmt.Errorf("failed to get backup data: %w", err)
	}

	info.EntryCount = int64(len(data))

	// 根据格式序列化数据
	var content []byte
	switch options.Format {
	case BackupFormatJSON:
		content, err = bm.serializeToJSON(data)
	case BackupFormatSQL:
		content, err = bm.serializeToSQL(data)
	case BackupFormatCSV:
		content, err = bm.serializeToCSV(data)
	default:
		return fmt.Errorf("unsupported backup format: %s", options.Format)
	}

	if err != nil {
		return fmt.Errorf("failed to serialize data: %w", err)
	}

	// 压缩数据
	if options.Compress {
		content, err = bm.compressData(content)
		if err != nil {
			return fmt.Errorf("failed to compress data: %w", err)
		}
	}

	// 加密数据
	if options.Encrypt {
		content, err = bm.encryptData(content, options.Password)
		if err != nil {
			return fmt.Errorf("failed to encrypt data: %w", err)
		}
	}

	// 写入文件
	if err := os.WriteFile(info.FilePath, content, 0644); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	return nil
}

// getBackupData 获取备份数据
func (bm *BackupManager) getBackupData(options *BackupOptions) ([]BackupEntry, error) {
	var query string
	var args []interface{}

	switch options.Type {
	case BackupTypeFull:
		query = `
			SELECT key, value, expire_at, created_at, updated_at 
			FROM storage_entries 
			WHERE expire_at IS NULL OR expire_at > ?
			ORDER BY key
		`
		args = []interface{}{time.Now().Unix()}

	case BackupTypeIncremental:
		if options.Since == nil {
			return nil, fmt.Errorf("incremental backup requires 'since' timestamp")
		}
		query = `
			SELECT key, value, expire_at, created_at, updated_at 
			FROM storage_entries 
			WHERE (expire_at IS NULL OR expire_at > ?) AND updated_at > ?
			ORDER BY key
		`
		args = []interface{}{time.Now().Unix(), options.Since.Unix()}

	default:
		return nil, fmt.Errorf("unsupported backup type: %s", options.Type)
	}

	rows, err := bm.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query backup data: %w", err)
	}
	defer rows.Close()

	var entries []BackupEntry
	for rows.Next() {
		var entry BackupEntry
		var expireAt *int64
		var createdAt, updatedAt int64

		err := rows.Scan(&entry.Key, &entry.Value, &expireAt, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan backup entry: %w", err)
		}

		if expireAt != nil {
			expireTime := time.Unix(*expireAt, 0)
			entry.ExpireAt = &expireTime
		}
		entry.CreatedAt = time.Unix(createdAt, 0)
		entry.UpdatedAt = time.Unix(updatedAt, 0)

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating backup data: %w", err)
	}

	return entries, nil
}

// BackupEntry 备份条目
type BackupEntry struct {
	Key       string     `json:"key"`
	Value     string     `json:"value"`
	ExpireAt  *time.Time `json:"expire_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// serializeToJSON 序列化为JSON格式
func (bm *BackupManager) serializeToJSON(entries []BackupEntry) ([]byte, error) {
	backupData := map[string]interface{}{
		"version":    1,
		"created_at": time.Now(),
		"entry_count": len(entries),
		"entries":    entries,
	}
	return json.MarshalIndent(backupData, "", "  ")
}

// serializeToSQL 序列化为SQL格式
func (bm *BackupManager) serializeToSQL(entries []BackupEntry) ([]byte, error) {
	var builder strings.Builder

	// 写入头部注释
	builder.WriteString("-- go-musicfox storage backup\n")
	builder.WriteString(fmt.Sprintf("-- Created at: %s\n", time.Now().Format(time.RFC3339)))
	builder.WriteString(fmt.Sprintf("-- Entry count: %d\n\n", len(entries)))

	// 写入表结构
	builder.WriteString("-- Table structure\n")
	builder.WriteString(`CREATE TABLE IF NOT EXISTS storage_entries (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL,
	expire_at INTEGER,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL
);

`)

	// 写入数据
	builder.WriteString("-- Data\n")
	builder.WriteString("BEGIN TRANSACTION;\n")

	for _, entry := range entries {
		var expireAt interface{}
		if entry.ExpireAt != nil {
			expireAt = entry.ExpireAt.Unix()
		} else {
			expireAt = "NULL"
		}

		builder.WriteString(fmt.Sprintf(
			"INSERT OR REPLACE INTO storage_entries (key, value, expire_at, created_at, updated_at) VALUES (%s, %s, %v, %d, %d);\n",
			bm.quoteSQLString(entry.Key),
			bm.quoteSQLString(entry.Value),
			expireAt,
			entry.CreatedAt.Unix(),
			entry.UpdatedAt.Unix(),
		))
	}

	builder.WriteString("COMMIT;\n")

	return []byte(builder.String()), nil
}

// serializeToCSV 序列化为CSV格式
func (bm *BackupManager) serializeToCSV(entries []BackupEntry) ([]byte, error) {
	var builder strings.Builder

	// 写入CSV头部
	builder.WriteString("key,value,expire_at,created_at,updated_at\n")

	// 写入数据行
	for _, entry := range entries {
		expireAt := ""
		if entry.ExpireAt != nil {
			expireAt = entry.ExpireAt.Format(time.RFC3339)
		}

		builder.WriteString(fmt.Sprintf(
			"%s,%s,%s,%s,%s\n",
			bm.escapeCSVField(entry.Key),
			bm.escapeCSVField(entry.Value),
			expireAt,
			entry.CreatedAt.Format(time.RFC3339),
			entry.UpdatedAt.Format(time.RFC3339),
		))
	}

	return []byte(builder.String()), nil
}

// quoteSQLString 转义SQL字符串
func (bm *BackupManager) quoteSQLString(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// escapeCSVField 转义CSV字段
func (bm *BackupManager) escapeCSVField(field string) string {
	if strings.ContainsAny(field, ",\"\n\r") {
		return "\"" + strings.ReplaceAll(field, "\"", "\"\"") + "\""
	}
	return field
}

// compressData 压缩数据
func (bm *BackupManager) compressData(data []byte) ([]byte, error) {
	var buf strings.Builder
	gzWriter := gzip.NewWriter(&buf)

	if _, err := gzWriter.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write compressed data: %w", err)
	}

	if err := gzWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return []byte(buf.String()), nil
}

// encryptData 加密数据
func (bm *BackupManager) encryptData(data []byte, password string) ([]byte, error) {
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty for encryption")
	}

	// 生成密钥
	key := sha256.Sum256([]byte(password))

	// 创建AES加密器
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// 生成随机IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	// 加密数据
	stream := cipher.NewCFBEncrypter(block, iv)
	encrypted := make([]byte, len(data))
	stream.XORKeyStream(encrypted, data)

	// 将IV和加密数据合并
	result := append(iv, encrypted...)
	return result, nil
}

// RestoreBackup 恢复备份
func (bm *BackupManager) RestoreBackup(backupID int64, password string) error {
	// 获取备份信息
	backupInfo, err := bm.GetBackupInfo(backupID)
	if err != nil {
		return fmt.Errorf("failed to get backup info: %w", err)
	}

	// 读取备份文件
	content, err := os.ReadFile(backupInfo.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	// 解密数据
	if backupInfo.Encrypted {
		content, err = bm.decryptData(content, password)
		if err != nil {
			return fmt.Errorf("failed to decrypt backup: %w", err)
		}
	}

	// 解压数据
	if backupInfo.Compressed {
		content, err = bm.decompressData(content)
		if err != nil {
			return fmt.Errorf("failed to decompress backup: %w", err)
		}
	}

	// 解析并恢复数据
	return bm.restoreFromContent(content, backupInfo)
}

// decryptData 解密数据
func (bm *BackupManager) decryptData(data []byte, password string) ([]byte, error) {
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty for decryption")
	}

	if len(data) < aes.BlockSize {
		return nil, fmt.Errorf("encrypted data too short")
	}

	// 生成密钥
	key := sha256.Sum256([]byte(password))

	// 创建AES解密器
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// 提取IV和加密数据
	iv := data[:aes.BlockSize]
	encrypted := data[aes.BlockSize:]

	// 解密数据
	stream := cipher.NewCFBDecrypter(block, iv)
	decrypted := make([]byte, len(encrypted))
	stream.XORKeyStream(decrypted, encrypted)

	return decrypted, nil
}

// decompressData 解压数据
func (bm *BackupManager) decompressData(data []byte) ([]byte, error) {
	reader := strings.NewReader(string(data))
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	decompressed, err := io.ReadAll(gzReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read decompressed data: %w", err)
	}

	return decompressed, nil
}

// restoreFromContent 从内容恢复数据
func (bm *BackupManager) restoreFromContent(content []byte, info *BackupInfo) error {
	// 解析JSON格式的备份
	var backupData map[string]interface{}
	if err := json.Unmarshal(content, &backupData); err != nil {
		return fmt.Errorf("failed to parse backup content: %w", err)
	}

	entries, ok := backupData["entries"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid backup format: missing entries")
	}

	// 开始事务
	tx, err := bm.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 恢复数据
	for _, entryData := range entries {
		entryMap, ok := entryData.(map[string]interface{})
		if !ok {
			continue
		}

		key, _ := entryMap["key"].(string)
		value, _ := entryMap["value"].(string)
		createdAtStr, _ := entryMap["created_at"].(string)
		updatedAtStr, _ := entryMap["updated_at"].(string)

		if key == "" || value == "" {
			continue
		}

		// 解析时间
		createdAt, _ := time.Parse(time.RFC3339, createdAtStr)
		updatedAt, _ := time.Parse(time.RFC3339, updatedAtStr)

		var expireAt *int64
		if expireAtStr, ok := entryMap["expire_at"].(string); ok && expireAtStr != "" {
			if expireTime, err := time.Parse(time.RFC3339, expireAtStr); err == nil {
				expire := expireTime.Unix()
				expireAt = &expire
			}
		}

		// 插入数据
		upsertSQL := `
			INSERT OR REPLACE INTO storage_entries (key, value, expire_at, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?)
		`
		_, err = tx.Exec(upsertSQL, key, value, expireAt, createdAt.Unix(), updatedAt.Unix())
		if err != nil {
			return fmt.Errorf("failed to restore entry %s: %w", key, err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit restore transaction: %w", err)
	}

	return nil
}

// 其他辅助方法...
func (bm *BackupManager) validateBackupOptions(options *BackupOptions) error {
	if options.Name == "" {
		return fmt.Errorf("backup name cannot be empty")
	}
	if options.Type != BackupTypeFull && options.Type != BackupTypeIncremental {
		return fmt.Errorf("invalid backup type: %s", options.Type)
	}
	if options.Format != BackupFormatJSON && options.Format != BackupFormatSQL && options.Format != BackupFormatCSV {
		return fmt.Errorf("invalid backup format: %s", options.Format)
	}
	if options.Encrypt && options.Password == "" {
		return fmt.Errorf("password required for encrypted backup")
	}
	return nil
}

func (bm *BackupManager) generateBackupFilePath(options *BackupOptions) string {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.%s", options.Name, timestamp, options.Format)

	if options.Compress {
		filename += ".gz"
	}
	if options.Encrypt {
		filename += ".enc"
	}

	if options.OutputPath != "" {
		return filepath.Join(options.OutputPath, filename)
	}
	return filepath.Join(bm.backupDir, filename)
}

func (bm *BackupManager) recordBackupInfo(info *BackupInfo) (int64, error) {
	insertSQL := `
		INSERT INTO storage_backups 
		(name, description, type, file_path, compressed, encrypted, entry_count, created_at, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := bm.db.Exec(insertSQL,
		info.Name, info.Description, info.Type, info.FilePath,
		info.Compressed, info.Encrypted, info.EntryCount,
		info.CreatedAt.Unix(), info.Status)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (bm *BackupManager) updateBackupStatus(id int64, status string) error {
	updateSQL := "UPDATE storage_backups SET status = ? WHERE id = ?"
	_, err := bm.db.Exec(updateSQL, status, id)
	return err
}

func (bm *BackupManager) updateBackupFileInfo(info *BackupInfo) error {
	// 获取文件大小
	fileInfo, err := os.Stat(info.FilePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	info.FileSize = fileInfo.Size()

	// 计算校验和
	checksum, err := bm.calculateChecksum(info.FilePath)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}
	info.Checksum = checksum

	// 更新数据库
	updateSQL := "UPDATE storage_backups SET file_size = ?, checksum = ? WHERE id = ?"
	_, err = bm.db.Exec(updateSQL, info.FileSize, info.Checksum, info.ID)
	return err
}

func (bm *BackupManager) calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// GetBackupInfo 获取备份信息
func (bm *BackupManager) GetBackupInfo(id int64) (*BackupInfo, error) {
	query := `
		SELECT id, name, description, type, file_path, file_size, 
		       compressed, encrypted, checksum, entry_count, created_at, status
		FROM storage_backups WHERE id = ?
	`

	var info BackupInfo
	var createdAt int64
	err := bm.db.QueryRow(query, id).Scan(
		&info.ID, &info.Name, &info.Description, &info.Type,
		&info.FilePath, &info.FileSize, &info.Compressed, &info.Encrypted,
		&info.Checksum, &info.EntryCount, &createdAt, &info.Status)
	if err != nil {
		return nil, err
	}

	info.CreatedAt = time.Unix(createdAt, 0)
	return &info, nil
}

// ListBackups 列出所有备份
func (bm *BackupManager) ListBackups() ([]BackupInfo, error) {
	query := `
		SELECT id, name, description, type, file_path, file_size, 
		       compressed, encrypted, checksum, entry_count, created_at, status
		FROM storage_backups ORDER BY created_at DESC
	`

	rows, err := bm.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var backups []BackupInfo
	for rows.Next() {
		var info BackupInfo
		var createdAt int64
		err := rows.Scan(
			&info.ID, &info.Name, &info.Description, &info.Type,
			&info.FilePath, &info.FileSize, &info.Compressed, &info.Encrypted,
			&info.Checksum, &info.EntryCount, &createdAt, &info.Status)
		if err != nil {
			return nil, err
		}
		info.CreatedAt = time.Unix(createdAt, 0)
		backups = append(backups, info)
	}

	return backups, rows.Err()
}

// DeleteBackup 删除备份
func (bm *BackupManager) DeleteBackup(id int64) error {
	// 获取备份信息
	info, err := bm.GetBackupInfo(id)
	if err != nil {
		return fmt.Errorf("failed to get backup info: %w", err)
	}

	// 删除文件
	if err := os.Remove(info.FilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete backup file: %w", err)
	}

	// 删除数据库记录
	deleteSQL := "DELETE FROM storage_backups WHERE id = ?"
	_, err = bm.db.Exec(deleteSQL, id)
	return err
}