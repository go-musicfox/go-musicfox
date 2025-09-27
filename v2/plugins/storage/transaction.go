package storage

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// TransactionState 事务状态
type TransactionState int

const (
	TransactionStateActive TransactionState = iota // 活跃状态
	TransactionStateCommitted                       // 已提交
	TransactionStateRolledBack                      // 已回滚
	TransactionStateAborted                         // 已中止
)

// String 返回事务状态的字符串表示
func (ts TransactionState) String() string {
	switch ts {
	case TransactionStateActive:
		return "active"
	case TransactionStateCommitted:
		return "committed"
	case TransactionStateRolledBack:
		return "rolled_back"
	case TransactionStateAborted:
		return "aborted"
	default:
		return "unknown"
	}
}

// TransactionOperation 事务操作类型
type TransactionOperation int

const (
	TransactionOpSet TransactionOperation = iota // 设置操作
	TransactionOpDelete                           // 删除操作
)

// TransactionEntry 事务条目
type TransactionEntry struct {
	Operation TransactionOperation `json:"operation"` // 操作类型
	Key       string              `json:"key"`       // 键
	Value     interface{}         `json:"value"`     // 值（仅用于Set操作）
	Timestamp time.Time           `json:"timestamp"` // 时间戳
}

// MemoryTransaction 内存事务实现
type MemoryTransaction struct {
	id        string                       // 事务ID
	state     TransactionState             // 事务状态
	operations []TransactionEntry          // 操作日志
	backend   StorageBackend              // 存储后端
	mu        sync.RWMutex                // 读写锁
	createdAt time.Time                   // 创建时间
	timeout   time.Duration               // 超时时间
}

// NewMemoryTransaction 创建内存事务
func NewMemoryTransaction(backend StorageBackend) *MemoryTransaction {
	return &MemoryTransaction{
		id:         uuid.New().String(),
		state:      TransactionStateActive,
		operations: make([]TransactionEntry, 0),
		backend:    backend,
		createdAt:  time.Now(),
		timeout:    30 * time.Second, // 默认30秒超时
	}
}



// GetID 获取事务ID
func (mt *MemoryTransaction) GetID() string {
	mt.mu.RLock()
	defer mt.mu.RUnlock()
	return mt.id
}

// IsActive 检查事务是否活跃
func (mt *MemoryTransaction) IsActive() bool {
	mt.mu.RLock()
	
	// 检查超时
	if time.Since(mt.createdAt) > mt.timeout {
		// 需要写锁来修改状态
		mt.mu.RUnlock()
		mt.mu.Lock()
		if mt.state == TransactionStateActive {
			mt.state = TransactionStateAborted
		}
		mt.mu.Unlock()
		return false
	}
	
	isActive := mt.state == TransactionStateActive
	mt.mu.RUnlock()
	return isActive
}

// Get 获取值
func (mt *MemoryTransaction) Get(key string) (interface{}, error) {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	// 检查事务状态（避免重复加锁）
	if mt.state != TransactionStateActive {
		return nil, fmt.Errorf("transaction is not active: %s", mt.state.String())
	}

	// 从操作日志中查找最新的值
	for i := len(mt.operations) - 1; i >= 0; i-- {
		entry := mt.operations[i]
		if entry.Key == key {
			switch entry.Operation {
			case TransactionOpSet:
				return entry.Value, nil
			case TransactionOpDelete:
				return nil, fmt.Errorf("key not found: %s", key)
			}
		}
	}

	// 从后端存储获取
	return mt.backend.Get(key)
}

// Set 设置值
func (mt *MemoryTransaction) Set(key string, value interface{}) error {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	// 检查事务状态（避免重复加锁）
	if mt.state != TransactionStateActive {
		return fmt.Errorf("transaction is not active: %s", mt.state.String())
	}

	// 添加到操作日志
	entry := TransactionEntry{
		Operation: TransactionOpSet,
		Key:       key,
		Value:     value,
		Timestamp: time.Now(),
	}
	mt.operations = append(mt.operations, entry)

	return nil
}

// Delete 删除值
func (mt *MemoryTransaction) Delete(key string) error {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	// 检查事务状态（避免重复加锁）
	if mt.state != TransactionStateActive {
		return fmt.Errorf("transaction is not active: %s", mt.state.String())
	}

	// 添加到操作日志
	entry := TransactionEntry{
		Operation: TransactionOpDelete,
		Key:       key,
		Timestamp: time.Now(),
	}
	mt.operations = append(mt.operations, entry)

	return nil
}

// Commit 提交事务
func (mt *MemoryTransaction) Commit() error {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	// 检查事务状态（避免重复加锁）
	if mt.state != TransactionStateActive {
		return fmt.Errorf("transaction is not active: %s", mt.state.String())
	}

	// 应用所有操作到后端存储
	for _, entry := range mt.operations {
		switch entry.Operation {
		case TransactionOpSet:
			if err := mt.backend.Set(entry.Key, entry.Value, 0); err != nil {
				mt.state = TransactionStateAborted
				return fmt.Errorf("failed to commit set operation for key %s: %w", entry.Key, err)
			}
		case TransactionOpDelete:
			if err := mt.backend.Delete(entry.Key); err != nil {
				mt.state = TransactionStateAborted
				return fmt.Errorf("failed to commit delete operation for key %s: %w", entry.Key, err)
			}
		}
	}

	mt.state = TransactionStateCommitted
	return nil
}

// Rollback 回滚事务
func (mt *MemoryTransaction) Rollback() error {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	if mt.state != TransactionStateActive {
		return fmt.Errorf("transaction is not active: %s", mt.state.String())
	}

	// 清空操作日志
	mt.operations = mt.operations[:0]
	mt.state = TransactionStateRolledBack

	return nil
}

// GetOperationCount 获取操作数量
func (mt *MemoryTransaction) GetOperationCount() int {
	mt.mu.RLock()
	defer mt.mu.RUnlock()
	return len(mt.operations)
}

// GetOperations 获取操作列表（只读）
func (mt *MemoryTransaction) GetOperations() []TransactionEntry {
	mt.mu.RLock()
	defer mt.mu.RUnlock()
	
	operations := make([]TransactionEntry, len(mt.operations))
	copy(operations, mt.operations)
	return operations
}

// GetState 获取事务状态
func (mt *MemoryTransaction) GetState() TransactionState {
	mt.mu.RLock()
	defer mt.mu.RUnlock()
	return mt.state
}

// GetCreatedAt 获取创建时间
func (mt *MemoryTransaction) GetCreatedAt() time.Time {
	mt.mu.RLock()
	defer mt.mu.RUnlock()
	return mt.createdAt
}

// SetTimeout 设置超时时间
func (mt *MemoryTransaction) SetTimeout(timeout time.Duration) {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	mt.timeout = timeout
}

// GetTimeout 获取超时时间
func (mt *MemoryTransaction) GetTimeout() time.Duration {
	mt.mu.RLock()
	defer mt.mu.RUnlock()
	return mt.timeout
}