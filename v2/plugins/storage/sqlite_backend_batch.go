package storage

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// GetBatch 批量获取值
func (sb *SQLiteBackend) GetBatch(keys []string) (map[string]interface{}, error) {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	if sb.closed {
		return nil, fmt.Errorf("backend is closed")
	}

	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	// 更新统计信息
	sb.updateStats(func(stats *BackendStats) {
		stats.ReadCount += int64(len(keys))
	})

	// 构建IN查询
	placeholders := make([]string, len(keys))
	args := make([]interface{}, len(keys)+1)
	for i, key := range keys {
		placeholders[i] = "?"
		args[i] = key
	}
	args[len(keys)] = time.Now().Unix() // 用于过期时间检查

	query := fmt.Sprintf(`
		SELECT key, value FROM storage_entries 
		WHERE key IN (%s) AND (expire_at IS NULL OR expire_at > ?)
	`, strings.Join(placeholders, ","))

	rows, err := sb.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute batch get query: %w", err)
	}
	defer rows.Close()

	result := make(map[string]interface{})
	for rows.Next() {
		var key, valueStr string
		if err := rows.Scan(&key, &valueStr); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// 反序列化值
		var value interface{}
		if err := json.Unmarshal([]byte(valueStr), &value); err != nil {
			return nil, fmt.Errorf("failed to unmarshal value for key %s: %w", key, err)
		}

		result[key] = value
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return result, nil
}

// SetBatch 批量设置值
func (sb *SQLiteBackend) SetBatch(items map[string]interface{}, ttl time.Duration) error {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	if sb.closed {
		return fmt.Errorf("backend is closed")
	}

	if len(items) == 0 {
		return nil
	}

	// 更新统计信息
	sb.updateStats(func(stats *BackendStats) {
		stats.WriteCount += int64(len(items))
	})

	// 开始事务
	tx, err := sb.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().Unix()
	var expireAt *int64

	// 设置过期时间
	if ttl > 0 {
		expire := time.Now().Add(ttl).Unix()
		expireAt = &expire
	}

	// 准备批量插入语句
	upsertSQL := `
		INSERT INTO storage_entries (key, value, expire_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			expire_at = excluded.expire_at,
			updated_at = excluded.updated_at;
	`

	stmt, err := tx.Prepare(upsertSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// 批量执行插入
	for key, value := range items {
		// 序列化值
		valueBytes, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}

		_, err = stmt.Exec(key, string(valueBytes), expireAt, now, now)
		if err != nil {
			return fmt.Errorf("failed to execute statement for key %s: %w", key, err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteBatch 批量删除值
func (sb *SQLiteBackend) DeleteBatch(keys []string) error {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	if sb.closed {
		return fmt.Errorf("backend is closed")
	}

	if len(keys) == 0 {
		return nil
	}

	// 更新统计信息
	sb.updateStats(func(stats *BackendStats) {
		stats.DeleteCount += int64(len(keys))
	})

	// 构建IN查询
	placeholders := make([]string, len(keys))
	args := make([]interface{}, len(keys))
	for i, key := range keys {
		placeholders[i] = "?"
		args[i] = key
	}

	deleteSQL := fmt.Sprintf("DELETE FROM storage_entries WHERE key IN (%s)", strings.Join(placeholders, ","))
	_, err := sb.db.Exec(deleteSQL, args...)
	if err != nil {
		return fmt.Errorf("failed to execute batch delete: %w", err)
	}

	return nil
}

// Find 查找匹配模式的键值对
func (sb *SQLiteBackend) Find(pattern string, limit int) (map[string]interface{}, error) {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	if sb.closed {
		return nil, fmt.Errorf("backend is closed")
	}

	// 更新统计信息
	sb.updateStats(func(stats *BackendStats) {
		stats.ReadCount++
	})

	// 构建查询语句
	query := `
		SELECT key, value FROM storage_entries 
		WHERE key LIKE ? AND (expire_at IS NULL OR expire_at > ?)
	`

	args := []interface{}{pattern, time.Now().Unix()}

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := sb.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute find query: %w", err)
	}
	defer rows.Close()

	result := make(map[string]interface{})
	for rows.Next() {
		var key, valueStr string
		if err := rows.Scan(&key, &valueStr); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// 反序列化值
		var value interface{}
		if err := json.Unmarshal([]byte(valueStr), &value); err != nil {
			return nil, fmt.Errorf("failed to unmarshal value for key %s: %w", key, err)
		}

		result[key] = value
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return result, nil
}

// Count 统计匹配模式的键数量
func (sb *SQLiteBackend) Count(pattern string) (int64, error) {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	if sb.closed {
		return 0, fmt.Errorf("backend is closed")
	}

	query := `
		SELECT COUNT(*) FROM storage_entries 
		WHERE key LIKE ? AND (expire_at IS NULL OR expire_at > ?)
	`

	var count int64
	now := time.Now().Unix()
	err := sb.db.QueryRow(query, pattern, now).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count keys: %w", err)
	}

	return count, nil
}

// Keys 获取匹配模式的所有键
func (sb *SQLiteBackend) Keys(pattern string) ([]string, error) {
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
		SELECT key FROM storage_entries 
		WHERE key LIKE ? AND (expire_at IS NULL OR expire_at > ?)
		ORDER BY key
	`

	rows, err := sb.db.Query(query, pattern, time.Now().Unix())
	if err != nil {
		return nil, fmt.Errorf("failed to execute keys query: %w", err)
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, fmt.Errorf("failed to scan key: %w", err)
		}
		keys = append(keys, key)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return keys, nil
}